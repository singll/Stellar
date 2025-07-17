package router

import (
	"github.com/StellarServer/internal/config"
	"github.com/StellarServer/internal/database"
	"github.com/StellarServer/internal/plugin"
	"github.com/StellarServer/internal/services/nodemanager"
	"github.com/StellarServer/internal/services/pagemonitoring"
	"github.com/StellarServer/internal/services/portscan"
	"github.com/StellarServer/internal/services/taskmanager"
	"github.com/StellarServer/internal/services/vulnscan"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

// AppDependencies 应用依赖容器
type AppDependencies struct {
	DB                  *database.DB
	MongoDB             *mongo.Database
	RedisClient         *redis.Client
	Config              *config.Config
	NodeManager         *nodemanager.NodeManager
	TaskManager         *taskmanager.TaskManager
	VulnEngine          *vulnscan.Engine
	VulnHandler         *vulnscan.VulnHandler
	VulnRegistry        vulnscan.PluginRegistry
	PortScanTaskManager *portscan.TaskManager
	PluginManager       *plugin.Manager
	PluginStore         plugin.MetadataStore
	PluginMarketplace   *plugin.Marketplace
	MonitoringService   *pagemonitoring.PageMonitoringService
}

// SetupAllRoutes 设置所有路由
func SetupAllRoutes(engine *gin.Engine, deps *AppDependencies) *RouteManager {
	// 创建路由管理器
	rm := NewRouteManager(engine)

	// 应用全局中间件
	rm.ApplyGlobalMiddleware()

	// 注册顶级路由（健康检查等）
	setupTopLevelRoutes(rm, deps)

	// 注册公开路由（无需认证）
	setupPublicRoutes(rm, deps)

	// 注册认证路由（需要认证）
	setupAuthenticatedRoutes(rm, deps)

	return rm
}

// setupTopLevelRoutes 设置顶级路由
func setupTopLevelRoutes(rm *RouteManager, deps *AppDependencies) {
	// 健康检查
	rm.RegisterTopLevel("GET", "/health", func(c *gin.Context) {
		health := deps.DB.Health()
		status := "healthy"
		statusCode := 200

		for _, err := range health {
			if err != nil {
				status = "unhealthy"
				statusCode = 503
				break
			}
		}

		c.JSON(statusCode, gin.H{
			"status":   status,
			"version":  "1.0",
			"services": health,
		})
	})

	// WebSocket
	rm.RegisterTopLevel("GET", "/ws", handleWebSocket)
}

// setupPublicRoutes 设置公开路由
func setupPublicRoutes(rm *RouteManager, deps *AppDependencies) {
	// 认证相关路由（登录、注册等）
	authHandler := NewAuthHandler(deps)
	rm.RegisterPublicGroup("auth", "/auth", authHandler)
}

// setupAuthenticatedRoutes 设置需要认证的路由
func setupAuthenticatedRoutes(rm *RouteManager, deps *AppDependencies) {
	// 项目管理
	projectHandler := NewProjectHandler(deps)
	rm.RegisterAuthGroup("projects", "/projects", projectHandler)

	// 资产管理
	assetHandler := NewAssetHandler(deps)
	rm.RegisterAuthGroup("assets", "/assets", assetHandler)

	// 任务管理
	taskHandler := NewTaskHandler(deps)
	rm.RegisterAuthGroup("tasks", "/tasks", taskHandler)

	// 节点管理
	nodeHandler := NewNodeHandler(deps)
	rm.RegisterAuthGroup("nodes", "/nodes", nodeHandler)

	// 漏洞管理
	vulnHandler := NewVulnerabilityHandler(deps)
	rm.RegisterAuthGroup("vulnerabilities", "/vulnerabilities", vulnHandler)

	// 端口扫描
	portscanHandler := NewPortScanHandler(deps)
	rm.RegisterAuthGroup("portscan", "/portscan", portscanHandler)

	// 子域名枚举
	subdomainHandler := NewSubdomainHandler(deps)
	rm.RegisterAuthGroup("subdomains", "/subdomains", subdomainHandler)

	// 资产发现
	discoveryHandler := NewDiscoveryHandler(deps)
	rm.RegisterAuthGroup("discovery", "/discovery", discoveryHandler)

	// 敏感信息检测
	sensitiveHandler := NewSensitiveHandler(deps)
	rm.RegisterAuthGroup("sensitive", "/sensitive", sensitiveHandler)

	// 页面监控
	monitoringHandler := NewMonitoringHandler(deps)
	rm.RegisterAuthGroup("monitoring", "/monitoring", monitoringHandler)

	// 插件管理
	pluginHandler := NewPluginHandler(deps)
	rm.RegisterAuthGroup("plugins", "/plugins", pluginHandler)

	// 漏洞数据库
	vulndbHandler := NewVulnDBHandler(deps)
	rm.RegisterAuthGroup("vulndb", "/vulndb", vulndbHandler)

	// 统计分析
	statisticsHandler := NewStatisticsHandler(deps)
	rm.RegisterAuthGroup("statistics", "/statistics", statisticsHandler)
}

// WebSocket处理器（临时）
func handleWebSocket(c *gin.Context) {
	c.JSON(200, gin.H{"message": "WebSocket endpoint"})
}
