package plugin

import (
	"errors"
	"fmt"
	"sync"
)

// Registry 插件注册表接口
type Registry interface {
	// Register 注册插件
	Register(plugin Plugin) error

	// GetPlugin 获取插件
	GetPlugin(id string) (Plugin, error)

	// ListPlugins 列出所有插件
	ListPlugins() []Plugin

	// ListPluginsByType 按类型列出插件
	ListPluginsByType(pluginType PluginType) []Plugin

	// ListPluginsByCategory 按分类列出插件
	ListPluginsByCategory(category string) []Plugin

	// ListPluginsByTag 按标签列出插件
	ListPluginsByTag(tag string) []Plugin

	// UnregisterPlugin 注销插件
	UnregisterPlugin(id string) error

	// LoadPlugins 从目录加载插件
	LoadPlugins(path string) error

	// RegisterLoader 注册插件加载器
	RegisterLoader(loader Loader)
}

// RegistryImpl 插件注册表实现
type RegistryImpl struct {
	plugins    map[string]Plugin
	loaders    map[string]Loader
	pluginLock sync.RWMutex
	loaderLock sync.RWMutex
}

// NewRegistry 创建插件注册表
func NewRegistry() *RegistryImpl {
	return &RegistryImpl{
		plugins: make(map[string]Plugin),
		loaders: make(map[string]Loader),
	}
}

// Register 注册插件
func (r *RegistryImpl) Register(plugin Plugin) error {
	r.pluginLock.Lock()
	defer r.pluginLock.Unlock()

	info := plugin.Info()
	if info.ID == "" {
		return errors.New("插件ID不能为空")
	}

	if _, exists := r.plugins[info.ID]; exists {
		return fmt.Errorf("插件ID已存在: %s", info.ID)
	}

	r.plugins[info.ID] = plugin
	return nil
}

// GetPlugin 获取插件
func (r *RegistryImpl) GetPlugin(id string) (Plugin, error) {
	r.pluginLock.RLock()
	defer r.pluginLock.RUnlock()

	plugin, exists := r.plugins[id]
	if !exists {
		return nil, fmt.Errorf("插件不存在: %s", id)
	}

	return plugin, nil
}

// ListPlugins 列出所有插件
func (r *RegistryImpl) ListPlugins() []Plugin {
	r.pluginLock.RLock()
	defer r.pluginLock.RUnlock()

	plugins := make([]Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// ListPluginsByType 按类型列出插件
func (r *RegistryImpl) ListPluginsByType(pluginType PluginType) []Plugin {
	r.pluginLock.RLock()
	defer r.pluginLock.RUnlock()

	var plugins []Plugin
	for _, plugin := range r.plugins {
		if plugin.Info().Type == pluginType {
			plugins = append(plugins, plugin)
		}
	}

	return plugins
}

// ListPluginsByCategory 按分类列出插件
func (r *RegistryImpl) ListPluginsByCategory(category string) []Plugin {
	r.pluginLock.RLock()
	defer r.pluginLock.RUnlock()

	var plugins []Plugin
	for _, plugin := range r.plugins {
		if plugin.Info().Category == category {
			plugins = append(plugins, plugin)
		}
	}

	return plugins
}

// ListPluginsByTag 按标签列出插件
func (r *RegistryImpl) ListPluginsByTag(tag string) []Plugin {
	r.pluginLock.RLock()
	defer r.pluginLock.RUnlock()

	var plugins []Plugin
	for _, plugin := range r.plugins {
		for _, t := range plugin.Info().Tags {
			if t == tag {
				plugins = append(plugins, plugin)
				break
			}
		}
	}

	return plugins
}

// UnregisterPlugin 注销插件
func (r *RegistryImpl) UnregisterPlugin(id string) error {
	r.pluginLock.Lock()
	defer r.pluginLock.Unlock()

	if _, exists := r.plugins[id]; !exists {
		return fmt.Errorf("插件不存在: %s", id)
	}

	delete(r.plugins, id)
	return nil
}

// RegisterLoader 注册插件加载器
func (r *RegistryImpl) RegisterLoader(loader Loader) {
	r.loaderLock.Lock()
	defer r.loaderLock.Unlock()

	for _, ext := range loader.SupportedExtensions() {
		r.loaders[ext] = loader
	}
}

// LoadPlugins 从目录加载插件
func (r *RegistryImpl) LoadPlugins(path string) error {
	// 调用LoadPluginsFromDirectory方法
	return r.LoadPluginsFromDirectory(path)
}
