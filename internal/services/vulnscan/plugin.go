package vulnscan

import (
	"context"
	"time"

	"github.com/StellarServer/internal/models"
)

// Target 扫描目标
type Target struct {
	URL       string                 // 目标URL
	Host      string                 // 目标主机
	Port      int                    // 目标端口
	Protocol  string                 // 协议
	Path      string                 // 路径
	Params    map[string]string      // 参数
	Headers   map[string]string      // 请求头
	Cookies   string                 // Cookie
	AssetID   string                 // 关联的资产ID
	AssetType string                 // 资产类型
	Metadata  map[string]interface{} // 元数据
}

// ScanContext 扫描上下文
type ScanContext struct {
	Context     context.Context        // 上下文
	Timeout     time.Duration          // 超时时间
	Proxy       string                 // 代理
	UserAgent   string                 // User-Agent
	Concurrency int                    // 并发数
	RetryCount  int                    // 重试次数
	RateLimit   int                    // 速率限制
	Cookies     string                 // Cookie
	Headers     map[string]string      // 请求头
	Metadata    map[string]interface{} // 元数据
}

// ScanResult 扫描结果
type ScanResult struct {
	Success        bool                   // 是否成功
	Vulnerability  *models.Vulnerability  // 漏洞信息
	Output         string                 // 输出
	Error          string                 // 错误
	ExecutionTime  time.Duration          // 执行时间
	Request        string                 // 请求
	Response       string                 // 响应
	Payload        string                 // Payload
	Screenshot     string                 // 截图
	AdditionalInfo map[string]interface{} // 额外信息
}

// Plugin 插件接口
type Plugin interface {
	// Info 获取插件信息
	Info() PluginInfo

	// Check 检查目标是否存在漏洞
	Check(ctx context.Context, target Target, params map[string]string) (ScanResult, error)

	// Init 初始化插件
	Init(config map[string]interface{}) error

	// Validate 验证参数
	Validate(params map[string]string) error
}

// PluginInfo 插件信息
type PluginInfo struct {
	ID             string                       // 插件ID
	Name           string                       // 插件名称
	Description    string                       // 描述
	Author         string                       // 作者
	References     []string                     // 参考资料
	CVEID          string                       // CVE ID
	CWEID          string                       // CWE ID
	Severity       models.VulnerabilitySeverity // 严重性
	Type           models.VulnerabilityType     // 类型
	Category       string                       // 分类
	Tags           []string                     // 标签
	RequiredParams []string                     // 必需参数
	DefaultParams  map[string]string            // 默认参数
}

// PluginRegistry 插件注册表
type PluginRegistry interface {
	// Register 注册插件
	Register(plugin Plugin) error

	// GetPlugin 获取插件
	GetPlugin(id string) (VulnPlugin, error)

	// ListPlugins 列出插件
	ListPlugins() []VulnPlugin

	// ListPluginsByCategory 按分类列出插件
	ListPluginsByCategory(category string) []VulnPlugin

	// ListPluginsByType 按类型列出插件
	ListPluginsByType(typeName models.VulnerabilityType) []VulnPlugin

	// ListPluginsBySeverity 按严重性列出插件
	ListPluginsBySeverity(severity models.VulnerabilitySeverity) []VulnPlugin

	// LoadPlugins 加载插件
	LoadPlugins(path string) error

	// UnloadPlugin 卸载插件
	UnloadPlugin(id string) error
}

// PluginLoader 插件加载器
type PluginLoader interface {
	// LoadPlugin 加载插件
	LoadPlugin(path string) (Plugin, error)

	// LoadPlugins 加载多个插件
	LoadPlugins(path string) ([]Plugin, error)

	// SupportedTypes 支持的插件类型
	SupportedTypes() []string
}

// GoPluginLoader Go插件加载器
type GoPluginLoader struct{}

// LoadPlugin 加载Go插件
func (l *GoPluginLoader) LoadPlugin(path string) (Plugin, error) {
	// 实现Go插件加载逻辑
	return nil, nil
}

// LoadPlugins 加载多个Go插件
func (l *GoPluginLoader) LoadPlugins(path string) ([]Plugin, error) {
	// 实现多个Go插件加载逻辑
	return nil, nil
}

// SupportedTypes 支持的插件类型
func (l *GoPluginLoader) SupportedTypes() []string {
	return []string{".so"}
}

// YAMLPluginLoader YAML插件加载器
type YAMLPluginLoader struct{}

// LoadPlugin 加载YAML插件
func (l *YAMLPluginLoader) LoadPlugin(path string) (Plugin, error) {
	// 实现YAML插件加载逻辑
	return nil, nil
}

// LoadPlugins 加载多个YAML插件
func (l *YAMLPluginLoader) LoadPlugins(path string) ([]Plugin, error) {
	// 实现多个YAML插件加载逻辑
	return nil, nil
}

// SupportedTypes 支持的插件类型
func (l *YAMLPluginLoader) SupportedTypes() []string {
	return []string{".yaml", ".yml"}
}

// PythonPluginLoader Python插件加载器
type PythonPluginLoader struct{}

// LoadPlugin 加载Python插件
func (l *PythonPluginLoader) LoadPlugin(path string) (Plugin, error) {
	// 实现Python插件加载逻辑
	return nil, nil
}

// LoadPlugins 加载多个Python插件
func (l *PythonPluginLoader) LoadPlugins(path string) ([]Plugin, error) {
	// 实现多个Python插件加载逻辑
	return nil, nil
}

// SupportedTypes 支持的插件类型
func (l *PythonPluginLoader) SupportedTypes() []string {
	return []string{".py"}
}

// ListPluginsBySeverity 按严重性列出插件
func (r *Registry) ListPluginsBySeverity(severity models.VulnerabilitySeverity) []VulnPlugin {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var plugins []VulnPlugin
	for _, plugin := range r.plugins {
		info := plugin.Info()
		if info.Severity == severity {
			plugins = append(plugins, plugin)
		}
	}

	return plugins
}

// ListPluginsByType 按类型列出插件
func (r *Registry) ListPluginsByType(typeName models.VulnerabilityType) []VulnPlugin {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var plugins []VulnPlugin
	for _, plugin := range r.plugins {
		info := plugin.Info()
		if info.Type == typeName {
			plugins = append(plugins, plugin)
		}
	}

	return plugins
}
