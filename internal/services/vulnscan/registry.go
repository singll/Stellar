package vulnscan

import (
	"fmt"
	"sync"

	"github.com/StellarServer/internal/models"
)

// VulnPlugin 漏洞扫描插件接口
type VulnPlugin interface {
	// Info 获取插件信息
	Info() models.POC
	// Scan 执行扫描
	Scan(target VulnTarget, options map[string]interface{}) (VulnScanResult, error)
}

// VulnTarget 扫描目标
type VulnTarget struct {
	URL      string
	Host     string
	Port     int
	Protocol string
	Path     string
	Params   map[string]string
}

// VulnScanResult 扫描结果
type VulnScanResult struct {
	Vulnerable bool
	Details    string
	Payload    string
	Request    string
	Response   string
	Screenshot string
}

// VulnPluginRegistry 插件注册表
type VulnPluginRegistry interface {
	// RegisterPlugin 注册插件
	RegisterPlugin(plugin VulnPlugin) error
	// GetPlugin 获取插件
	GetPlugin(id string) (VulnPlugin, error)
	// ListPlugins 列出所有插件
	ListPlugins() []VulnPlugin
	// ListPluginsByCategory 按分类列出插件
	ListPluginsByCategory(category string) []VulnPlugin
	// Count 获取插件数量
	Count() int
	// LoadPlugins 加载插件
	LoadPlugins(path string) error
	// UnloadPlugin 卸载插件
	UnloadPlugin(id string) error
}

// Registry 插件注册表实现
type Registry struct {
	plugins map[string]VulnPlugin
	mutex   sync.RWMutex
}

// NewRegistry 创建插件注册表
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]VulnPlugin),
	}
}

// RegisterPlugin 注册插件
func (r *Registry) RegisterPlugin(plugin VulnPlugin) error {
	// 获取插件信息
	info := plugin.Info()

	// 检查插件ID是否为空
	if info.ID.IsZero() {
		return fmt.Errorf("插件ID不能为空")
	}

	// 检查插件是否已存在
	idStr := info.ID.Hex()
	r.mutex.RLock()
	_, exists := r.plugins[idStr]
	r.mutex.RUnlock()
	if exists {
		return fmt.Errorf("插件已存在: %s", idStr)
	}

	// 注册插件
	r.mutex.Lock()
	r.plugins[idStr] = plugin
	r.mutex.Unlock()

	return nil
}

// Register 注册插件 (PluginRegistry 接口实现)
func (r *Registry) Register(plugin Plugin) error {
	vulnPlugin, ok := plugin.(VulnPlugin)
	if !ok {
		return fmt.Errorf("插件类型不匹配")
	}
	return r.RegisterPlugin(vulnPlugin)
}

// GetPlugin 获取插件
func (r *Registry) GetPlugin(id string) (VulnPlugin, error) {
	r.mutex.RLock()
	plugin, exists := r.plugins[id]
	r.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("插件不存在: %s", id)
	}

	return plugin, nil
}

// ListPlugins 列出所有插件
func (r *Registry) ListPlugins() []VulnPlugin {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	plugins := make([]VulnPlugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// ListPluginsByCategory 按分类列出插件
func (r *Registry) ListPluginsByCategory(category string) []VulnPlugin {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var plugins []VulnPlugin
	for _, plugin := range r.plugins {
		info := plugin.Info()
		if info.Category == category {
			plugins = append(plugins, plugin)
		}
	}

	return plugins
}

// Count 获取插件数量
func (r *Registry) Count() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.plugins)
}

// LoadPlugins 加载插件
func (r *Registry) LoadPlugins(path string) error {
	// 实际实现中，这里应该扫描目录并加载插件
	// 这里只是一个占位实现
	return nil
}

// UnloadPlugin 卸载插件
func (r *Registry) UnloadPlugin(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.plugins[id]; !exists {
		return fmt.Errorf("插件不存在: %s", id)
	}

	delete(r.plugins, id)
	return nil
}
