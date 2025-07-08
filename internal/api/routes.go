package api

import (
	"github.com/StellarServer/internal/config"
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

// RegisterAPIRoutes registers all API routes
func RegisterAPIRoutes(
	router *gin.Engine,
	db *mongo.Database,
	redisClient *redis.Client,
	cfg *config.Config,
	nodeManager *nodemanager.NodeManager,
	taskManager *taskmanager.TaskManager,
	vulnEngine *vulnscan.Engine,
	vulnHandler *vulnscan.VulnHandler,
	vulnRegistry vulnscan.PluginRegistry,
	portScanTaskManager *portscan.TaskManager,
	pluginManager *plugin.Manager,
	pluginStore plugin.MetadataStore,
	monitoringService *pagemonitoring.PageMonitoringService,
	discoveryHandler *DiscoveryHandler,
	sensitiveHandler *SensitiveHandler,
) {
	// API v1 group
	apiV1 := router.Group("/api/v1")

	// 认证相关路由（公开）
	authHandler := NewAuthHandler(db)
	authHandler.RegisterRoutes(apiV1)

	// 需要认证的业务路由
	projectsGroup := apiV1.Group("/projects")
	projectsGroup.Use(AuthMiddleware())
	projectHandler := NewProjectHandler(db)
	projectHandler.RegisterRoutes(projectsGroup)

	assetsGroup := apiV1.Group("/assets")
	assetsGroup.Use(AuthMiddleware())
	assetHandler := NewAssetHandler(db)
	assetHandler.RegisterRoutes(assetsGroup)

	nodesGroup := apiV1.Group("/nodes")
	nodesGroup.Use(AuthMiddleware())
	nodeHandler := NewNodeHandler(nodeManager)
	nodeHandler.RegisterRoutes(nodesGroup)

	taskAPI := NewTaskAPI(taskManager)
	taskAPI.RegisterRoutes(router) // some task routes might be top-level

	vulnerabilitiesGroup := apiV1.Group("/vulnerabilities")
	vulnerabilitiesGroup.Use(AuthMiddleware())
	vulnAPI := NewVulnerabilityAPI(vulnEngine, vulnHandler, vulnRegistry)
	vulnAPI.RegisterRoutes(vulnerabilitiesGroup)

	portscanGroup := apiV1.Group("/portscan")
	portscanGroup.Use(AuthMiddleware())
	portScanAPI := NewPortScanAPI(portScanTaskManager)
	portScanAPI.RegisterRoutes(portscanGroup)

	subdomainsGroup := apiV1.Group("/subdomains")
	subdomainsGroup.Use(AuthMiddleware())
	subdomainHandler := NewSubdomainHandler(db)
	subdomainHandler.RegisterRoutes(subdomainsGroup)

	pluginsGroup := apiV1.Group("/plugins")
	pluginsGroup.Use(AuthMiddleware())
	pluginHandler := NewPluginHandler(pluginManager, pluginStore)
	pluginHandler.RegisterRoutes(pluginsGroup)

	monitoringGroup := apiV1.Group("/monitoring")
	monitoringGroup.Use(AuthMiddleware())
	monitoringHandler := NewMonitoringHandler(monitoringService)
	monitoringHandler.RegisterRoutes(monitoringGroup)

	discoveryGroup := apiV1.Group("/discovery")
	discoveryGroup.Use(AuthMiddleware())
	discoveryHandler.RegisterRoutes(discoveryGroup)

	sensitiveGroup := apiV1.Group("/sensitive")
	sensitiveGroup.Use(AuthMiddleware())
	sensitiveHandler.RegisterRoutes(sensitiveGroup)

	statisticsGroup := apiV1.Group("/statistics")
	statisticsGroup.Use(AuthMiddleware())
	statsAPI := NewStatisticsAPI(db)
	statsAPI.RegisterRoutes(statisticsGroup)
}
