package plugin

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
)

// Manager 插件管理器
type Manager struct {
	registry      Registry
	pluginDirs    []string
	metadataStore MetadataStore
	mutex         sync.RWMutex
	hooks         map[string][]PluginHook
}

// PluginHook 插件钩子
type PluginHook func(event PluginEvent)

// PluginEvent 插件事件
type PluginEvent struct {
	Type      string
	PluginID  string
	Timestamp time.Time
	Data      map[string]interface{}
}

// MetadataStore 插件元数据存储接口
type MetadataStore interface {
	// GetPluginMetadata 获取插件元数据
	GetPluginMetadata(id string) (*models.PluginMetadata, error)

	// SavePluginMetadata 保存插件元数据
	SavePluginMetadata(metadata *models.PluginMetadata) error

	// ListPluginMetadata 列出插件元数据
	ListPluginMetadata() ([]*models.PluginMetadata, error)

	// DeletePluginMetadata 删除插件元数据
	DeletePluginMetadata(id string) error

	// ListPlugins 列出所有插件元数据
	ListPlugins() ([]*models.PluginMetadata, error)

	// ListPluginsByType 按类型列出插件元数据
	ListPluginsByType(pluginType string) ([]*models.PluginMetadata, error)

	// ListPluginsByCategory 按分类列出插件元数据
	ListPluginsByCategory(category string) ([]*models.PluginMetadata, error)

	// ListPluginsByTag 按标签列出插件元数据
	ListPluginsByTag(tag string) ([]*models.PluginMetadata, error)
}

// NewManager 创建插件管理器
func NewManager(registry Registry, metadataStore MetadataStore) *Manager {
	return &Manager{
		registry:      registry,
		metadataStore: metadataStore,
		pluginDirs:    make([]string, 0),
		hooks:         make(map[string][]PluginHook),
	}
}

// AddPluginDirectory 添加插件目录
func (m *Manager) AddPluginDirectory(dir string) error {
	// 检查目录是否存在
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("访问插件目录失败: %v", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("指定的路径不是目录: %s", dir)
	}

	// 添加到插件目录列表
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查是否已存在
	for _, existingDir := range m.pluginDirs {
		if existingDir == dir {
			return nil
		}
	}

	m.pluginDirs = append(m.pluginDirs, dir)
	return nil
}

// LoadPlugins 加载所有插件
func (m *Manager) LoadPlugins() error {
	m.mutex.RLock()
	dirs := make([]string, len(m.pluginDirs))
	copy(dirs, m.pluginDirs)
	m.mutex.RUnlock()

	// 加载每个目录中的插件
	for _, dir := range dirs {
		if err := m.registry.LoadPlugins(dir); err != nil {
			return fmt.Errorf("加载插件目录 %s 失败: %v", dir, err)
		}
	}

	// 处理插件加载完成事件
	m.triggerEvent("plugins_loaded", "", map[string]interface{}{
		"count": len(m.registry.ListPlugins()),
	})

	return nil
}

// GetPlugin 获取插件
func (m *Manager) GetPlugin(id string) (Plugin, error) {
	return m.registry.GetPlugin(id)
}

// ListPlugins 列出所有插件
func (m *Manager) ListPlugins() []Plugin {
	return m.registry.ListPlugins()
}

// ListPluginsByType 按类型列出插件
func (m *Manager) ListPluginsByType(pluginType PluginType) []Plugin {
	return m.registry.ListPluginsByType(pluginType)
}

// InstallPlugin 安装插件
func (m *Manager) InstallPlugin(sourcePath string, destDir string) (string, error) {
	// 检查源文件是否存在
	info, err := os.Stat(sourcePath)
	if err != nil {
		return "", fmt.Errorf("访问插件文件失败: %v", err)
	}

	// 创建目标目录
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("创建插件目录失败: %v", err)
	}

	// 目标文件路径
	destPath := filepath.Join(destDir, filepath.Base(sourcePath))

	// 如果是目录，则需要递归复制
	if info.IsDir() {
		return "", fmt.Errorf("不支持安装目录插件: %s", sourcePath)
	}

	// 复制文件
	if err := copyFile(sourcePath, destPath); err != nil {
		return "", fmt.Errorf("复制插件文件失败: %v", err)
	}

	// 加载插件
	plugin, err := loadSinglePlugin(destPath, m.registry)
	if err != nil {
		// 如果加载失败，删除文件
		os.Remove(destPath)
		return "", fmt.Errorf("加载插件失败: %v", err)
	}

	// 创建插件元数据
	pluginInfo := plugin.Info()
	metadata := &models.PluginMetadata{
		ID:          pluginInfo.ID,
		Name:        pluginInfo.Name,
		Version:     pluginInfo.Version,
		Type:        string(pluginInfo.Type),
		Author:      pluginInfo.Author,
		Description: pluginInfo.Description,
		Path:        destPath,
		InstallTime: time.Now(),
		Enabled:     true,
	}

	// 保存元数据
	if err := m.metadataStore.SavePluginMetadata(metadata); err != nil {
		return "", fmt.Errorf("保存插件元数据失败: %v", err)
	}

	// 触发插件安装事件
	m.triggerEvent("plugin_installed", pluginInfo.ID, map[string]interface{}{
		"path":    destPath,
		"version": pluginInfo.Version,
	})

	return pluginInfo.ID, nil
}

// UninstallPlugin 卸载插件
func (m *Manager) UninstallPlugin(id string) error {
	// 获取插件元数据
	metadata, err := m.metadataStore.GetPluginMetadata(id)
	if err != nil {
		return fmt.Errorf("获取插件元数据失败: %v", err)
	}

	// 从注册表中移除插件
	if err := m.registry.UnregisterPlugin(id); err != nil {
		return fmt.Errorf("注销插件失败: %v", err)
	}

	// 删除插件文件
	if err := os.Remove(metadata.Path); err != nil {
		return fmt.Errorf("删除插件文件失败: %v", err)
	}

	// 删除元数据
	if err := m.metadataStore.DeletePluginMetadata(id); err != nil {
		return fmt.Errorf("删除插件元数据失败: %v", err)
	}

	// 触发插件卸载事件
	m.triggerEvent("plugin_uninstalled", id, nil)

	return nil
}

// EnablePlugin 启用插件
func (m *Manager) EnablePlugin(id string) error {
	// 获取插件元数据
	metadata, err := m.metadataStore.GetPluginMetadata(id)
	if err != nil {
		return fmt.Errorf("获取插件元数据失败: %v", err)
	}

	// 检查插件是否已启用
	if metadata.Enabled {
		return nil
	}

	// 更新元数据
	metadata.Enabled = true
	metadata.UpdateTime = time.Now()
	if err := m.metadataStore.SavePluginMetadata(metadata); err != nil {
		return fmt.Errorf("保存插件元数据失败: %v", err)
	}

	// 触发插件启用事件
	m.triggerEvent("plugin_enabled", id, nil)

	return nil
}

// DisablePlugin 禁用插件
func (m *Manager) DisablePlugin(id string) error {
	// 获取插件元数据
	metadata, err := m.metadataStore.GetPluginMetadata(id)
	if err != nil {
		return fmt.Errorf("获取插件元数据失败: %v", err)
	}

	// 检查插件是否已禁用
	if !metadata.Enabled {
		return nil
	}

	// 更新元数据
	metadata.Enabled = false
	metadata.UpdateTime = time.Now()
	if err := m.metadataStore.SavePluginMetadata(metadata); err != nil {
		return fmt.Errorf("保存插件元数据失败: %v", err)
	}

	// 触发插件禁用事件
	m.triggerEvent("plugin_disabled", id, nil)

	return nil
}

// AddHook 添加插件事件钩子
func (m *Manager) AddHook(eventType string, hook PluginHook) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.hooks[eventType]; !exists {
		m.hooks[eventType] = make([]PluginHook, 0)
	}

	m.hooks[eventType] = append(m.hooks[eventType], hook)
}

// triggerEvent 触发插件事件
func (m *Manager) triggerEvent(eventType string, pluginID string, data map[string]interface{}) {
	m.mutex.RLock()
	hooks, exists := m.hooks[eventType]
	m.mutex.RUnlock()

	if !exists {
		return
	}

	event := PluginEvent{
		Type:      eventType,
		PluginID:  pluginID,
		Timestamp: time.Now(),
		Data:      data,
	}

	for _, hook := range hooks {
		go hook(event)
	}
}

// loadSinglePlugin 加载单个插件
func loadSinglePlugin(path string, registry Registry) (Plugin, error) {
	ext := filepath.Ext(path)
	if ext == "" {
		return nil, fmt.Errorf("无法确定插件类型: %s", path)
	}

	// 获取加载器
	var loader Loader
	switch ext {
	case ".so":
		loader = &GoLoader{}
	case ".py":
		loader = NewScriptLoader()
	case ".yaml", ".yml":
		loader = &YAMLLoader{}
	default:
		return nil, fmt.Errorf("不支持的插件类型: %s", ext)
	}

	// 加载插件
	plugin, err := loader.Load(path)
	if err != nil {
		return nil, err
	}

	// 注册插件
	if err := registry.Register(plugin); err != nil {
		return nil, err
	}

	return plugin, nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	// 读取源文件
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// 写入目标文件
	return os.WriteFile(dst, data, 0644)
}

// InstallPluginFromFile 从本地文件安装插件
func (m *Manager) InstallPluginFromFile(sourcePath string, config map[string]interface{}, forceUpdate bool) (string, error) {
	// 检查源文件是否存在
	_, err := os.Stat(sourcePath)
	if err != nil {
		return "", fmt.Errorf("访问插件文件失败: %v", err)
	}

	// 获取默认插件目录
	if len(m.pluginDirs) == 0 {
		return "", fmt.Errorf("未配置插件目录")
	}
	destDir := m.pluginDirs[0]

	// 安装插件
	pluginID, err := m.InstallPlugin(sourcePath, destDir)
	if err != nil {
		return "", err
	}

	// 如果提供了配置，更新插件配置
	if config != nil {
		metadata, err := m.metadataStore.GetPluginMetadata(pluginID)
		if err != nil {
			return "", fmt.Errorf("获取插件元数据失败: %v", err)
		}

		metadata.Config = config
		metadata.UpdateTime = time.Now()
		if err := m.metadataStore.SavePluginMetadata(metadata); err != nil {
			return "", fmt.Errorf("保存插件配置失败: %v", err)
		}
	}

	return pluginID, nil
}

// InstallPluginFromURL 从URL安装插件
func (m *Manager) InstallPluginFromURL(url string, config map[string]interface{}, forceUpdate bool) (string, error) {
	// 获取默认插件目录
	if len(m.pluginDirs) == 0 {
		return "", fmt.Errorf("未配置插件目录")
	}
	destDir := m.pluginDirs[0]

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "plugin-*.tmp")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// 下载文件
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("下载插件失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载插件失败: 状态码 %d", resp.StatusCode)
	}

	// 写入临时文件
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("保存插件失败: %v", err)
	}

	// 安装插件
	pluginID, err := m.InstallPlugin(tempFile.Name(), destDir)
	if err != nil {
		return "", err
	}

	// 如果提供了配置，更新插件配置
	if config != nil {
		metadata, err := m.metadataStore.GetPluginMetadata(pluginID)
		if err != nil {
			return "", fmt.Errorf("获取插件元数据失败: %v", err)
		}

		metadata.Config = config
		metadata.UpdateTime = time.Now()
		if err := m.metadataStore.SavePluginMetadata(metadata); err != nil {
			return "", fmt.Errorf("保存插件配置失败: %v", err)
		}
	}

	return pluginID, nil
}

// InstallPluginFromMarket 从插件市场安装插件
func (m *Manager) InstallPluginFromMarket(pluginID string, config map[string]interface{}, forceUpdate bool) (string, error) {
	// 这里应该实现从插件市场获取插件的逻辑
	// 目前先使用硬编码的URL映射
	marketURLs := map[string]string{
		"com.stellar.plugin.subdomain-finder": "https://plugins.stellarserver.io/subdomain-finder-1.0.0.zip",
		"com.stellar.plugin.port-scanner":     "https://plugins.stellarserver.io/port-scanner-1.0.0.zip",
	}

	url, ok := marketURLs[pluginID]
	if !ok {
		return "", fmt.Errorf("插件市场中不存在该插件: %s", pluginID)
	}

	// 从URL安装插件
	return m.InstallPluginFromURL(url, config, forceUpdate)
}

// ReloadPlugin 重新加载插件
func (m *Manager) ReloadPlugin(id string) error {
	// 获取插件元数据
	metadata, err := m.metadataStore.GetPluginMetadata(id)
	if err != nil {
		return fmt.Errorf("获取插件元数据失败: %v", err)
	}

	// 检查插件是否已启用
	if !metadata.Enabled {
		return fmt.Errorf("插件未启用，无法重新加载")
	}

	// 从注册表中移除插件
	if err := m.registry.UnregisterPlugin(id); err != nil {
		return fmt.Errorf("注销插件失败: %v", err)
	}

	// 重新加载插件
	plugin, err := loadSinglePlugin(metadata.Path, m.registry)
	if err != nil {
		return fmt.Errorf("重新加载插件失败: %v", err)
	}

	// 触发插件重新加载事件
	m.triggerEvent("plugin_reloaded", id, map[string]interface{}{
		"path":    metadata.Path,
		"version": plugin.Info().Version,
	})

	return nil
}
