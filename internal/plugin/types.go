package plugin

import (
	"context"
	"time"
)

// PluginType 定义插件类型
type PluginType string

// 插件类型常量
const (
	TypeVulnScan       PluginType = "vulnscan"       // 漏洞扫描插件
	TypeAssetDiscovery PluginType = "assetdiscovery" // 资产发现插件
	TypePortScan       PluginType = "portscan"       // 端口扫描插件
	TypeSubdomainEnum  PluginType = "subdomain"      // 子域名枚举插件
	TypeSensitiveInfo  PluginType = "sensitive"      // 敏感信息检测插件
	TypeMonitoring     PluginType = "monitoring"     // 监控插件
	TypeUtility        PluginType = "utility"        // 工具类插件
)

// PluginInfo 插件信息
type PluginInfo struct {
	ID          string            // 插件ID，必须唯一
	Name        string            // 插件名称
	Version     string            // 插件版本
	Type        PluginType        // 插件类型
	Author      string            // 作者
	Description string            // 描述
	Category    string            // 分类
	Tags        []string          // 标签
	Website     string            // 网站
	License     string            // 许可证
	References  []string          // 参考资料
	CreatedAt   time.Time         // 创建时间
	UpdatedAt   time.Time         // 更新时间
	Params      []PluginParam     // 参数定义
	Requires    map[string]string // 依赖要求，key为依赖名，value为版本要求
	Language    string            // 插件语言，如Go、Python等
}

// PluginParam 插件参数定义
type PluginParam struct {
	Name        string      // 参数名
	Type        string      // 参数类型，如string、int、bool等
	Description string      // 参数描述
	Required    bool        // 是否必需
	Default     interface{} // 默认值
	Options     []string    // 可选值列表，如果为空则表示无限制
}

// PluginContext 插件执行上下文
type PluginContext struct {
	Context     context.Context        // 上下文
	Timeout     time.Duration          // 超时时间
	Config      map[string]interface{} // 配置
	Params      map[string]interface{} // 参数
	Environment map[string]string      // 环境变量
	WorkDir     string                 // 工作目录
	Logger      Logger                 // 日志记录器
}

// PluginResult 插件执行结果
type PluginResult struct {
	Success       bool                   // 是否成功
	Data          interface{}            // 结果数据
	Message       string                 // 消息
	Error         string                 // 错误信息
	ExecutionTime time.Duration          // 执行时间
	Metadata      map[string]interface{} // 元数据
}

// Logger 日志接口
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// Plugin 插件接口
type Plugin interface {
	// Info 返回插件信息
	Info() PluginInfo

	// Init 初始化插件
	Init(config map[string]interface{}) error

	// Execute 执行插件
	Execute(ctx PluginContext) (PluginResult, error)

	// Validate 验证参数
	Validate(params map[string]interface{}) error

	// Cleanup 清理资源
	Cleanup() error
}
