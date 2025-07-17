package router

import (
	"github.com/StellarServer/internal/api"
	"github.com/StellarServer/internal/models"
	"github.com/gin-gonic/gin"
)

// 处理器工厂函数 - 使用现有的Handler实现

// NewAuthHandler 创建认证处理器
func NewAuthHandler(deps *AppDependencies) RouteHandler {
	return api.NewAuthHandler(deps.MongoDB)
}

// NewProjectHandler 创建项目处理器
func NewProjectHandler(deps *AppDependencies) RouteHandler {
	return api.NewProjectHandler(deps.MongoDB)
}

// NewAssetHandler 创建资产处理器
func NewAssetHandler(deps *AppDependencies) RouteHandler {
	return api.NewAssetHandler(deps.MongoDB)
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(deps *AppDependencies) RouteHandler {
	return api.NewTaskHandler(deps.TaskManager)
}

// NewNodeHandler 创建节点处理器
func NewNodeHandler(deps *AppDependencies) RouteHandler {
	return api.NewNodeHandler(deps.NodeManager, models.NewNodeRepository(deps.MongoDB))
}

// NewVulnerabilityHandler 创建漏洞处理器
func NewVulnerabilityHandler(deps *AppDependencies) RouteHandler {
	return api.NewVulnerabilityAPI(deps.VulnEngine, deps.VulnHandler, deps.VulnRegistry)
}

// NewPortScanHandler 创建端口扫描处理器
func NewPortScanHandler(deps *AppDependencies) RouteHandler {
	return api.NewPortScanAPI(deps.PortScanTaskManager)
}

// NewSubdomainHandler 创建子域名处理器
func NewSubdomainHandler(deps *AppDependencies) RouteHandler {
	return api.NewSubdomainHandler(deps.MongoDB)
}

// NewDiscoveryHandler 创建资产发现处理器
func NewDiscoveryHandler(deps *AppDependencies) RouteHandler {
	return &EmptyHandler{name: "discovery"}
}

// NewSensitiveHandler 创建敏感信息处理器
func NewSensitiveHandler(deps *AppDependencies) RouteHandler {
	return &EmptyHandler{name: "sensitive"}
}

// NewMonitoringHandler 创建监控处理器
func NewMonitoringHandler(deps *AppDependencies) RouteHandler {
	return api.NewMonitoringHandler(deps.MonitoringService)
}

// NewPluginHandler 创建插件处理器
func NewPluginHandler(deps *AppDependencies) RouteHandler {
	return api.NewPluginHandler(deps.PluginManager, deps.PluginStore)
}

// NewVulnDBHandler 创建漏洞数据库处理器
func NewVulnDBHandler(deps *AppDependencies) RouteHandler {
	return &EmptyHandler{name: "vulndb"}
}

// NewStatisticsHandler 创建统计处理器
func NewStatisticsHandler(deps *AppDependencies) RouteHandler {
	return api.NewStatisticsAPI(deps.MongoDB)
}

// EmptyHandler 空处理器，用于暂时替代未完成的处理器
type EmptyHandler struct {
	name string
}

func (h *EmptyHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": h.name + " 接口暂未实现"})
	})
}
