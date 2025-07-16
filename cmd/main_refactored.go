package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/StellarServer/internal/api/router"
	"github.com/StellarServer/internal/app"
	"github.com/StellarServer/internal/config"
	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/pkg/logger"
	"github.com/StellarServer/internal/plugin"
	"github.com/StellarServer/internal/services/assetdiscovery"
	"github.com/StellarServer/internal/services/nodemanager"
	"github.com/StellarServer/internal/services/pagemonitoring"
	"github.com/StellarServer/internal/services/portscan"
	"github.com/StellarServer/internal/services/sensitive"
	"github.com/StellarServer/internal/services/subdomain"
	"github.com/StellarServer/internal/services/taskmanager"
	"github.com/StellarServer/internal/services/vulnscan"
	"github.com/StellarServer/internal/services/vulndb"
	"github.com/StellarServer/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	VERSION = "1.0"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有跨域WebSocket连接
	},
}

var (
	configFile     = flag.String("config", "config.yaml", "配置文件路径")
	logLevel       = flag.String("log-level", "info", "日志级别: debug, info, warn, error")
	enableMonitor  = flag.Bool("monitor", true, "启用性能监控")
	showRoutes     = flag.Bool("show-routes", true, "启动时显示所有路由")
	saveRoutesFile = flag.String("save-routes", "", "保存路由信息到指定文件")
	env            = flag.String("env", "", "运行环境: development, production, test")
)

func banner() {
	fmt.Println(`
   _____ _       _ _              _____
  / ____| |     | | |            / ____|
 | (___ | |_ ___| | | __ _ _ __ | (___   ___ _ ____   _____ _ __
  \___ \| __/ _ \ | |/ _' | '__|  \___ \ / _ \ '__\ \ / / _ \ '__|
  ____) | ||  __/ | | (_| | |     ____) |  __/ |   \ V /  __/ |
 |_____/ \__\___|_|_|\__,_|_|    |_____/ \___|_|    \_/ \___|_|

 StellarServer - 安全资产管理平台 (重构版)
 版本: ` + VERSION + `
	`)
}

func main() {
	fmt.Println("StellarServer 正在启动...")
	fmt.Println("正在解析命令行参数...")

	flag.Parse()

	banner()

	// 获取环境
	environment := *env
	if environment == "" {
		environment = os.Getenv("APP_ENV")
		if environment == "" {
			environment = "development"
		}
	}

	// 初始化日志系统
	fmt.Println("初始化日志系统...")
	logConfig := logger.Config{
		Level:  *logLevel,
		Format: "console",
		Output: "stdout",
	}
	
	if err := logger.Init(logConfig); err != nil {
		fmt.Printf("初始化日志系统失败: %v\n", err)
		log.Fatalf("初始化日志系统失败: %v", err)
	}

	logger.Info("StellarServer启动中", map[string]interface{}{
		"version":     VERSION,
		"environment": environment,
	})

	// 初始化性能监控（保持向后兼容）
	if *enableMonitor {
		utils.InitGlobalMonitor(10*time.Second, 100)
		defer utils.StopGlobalMonitor()
		logger.Info("性能监控已启用", nil)
		fmt.Println("性能监控已启用")

		// 定期记录性能指标
		go func() {
			ticker := time.NewTicker(1 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				utils.LogPerformance()
			}
		}()
	}

	// 创建配置管理器
	configManager := config.NewManager(environment)
	
	// 根据命令行参数加载配置
	var err error
	if *configFile != "" && *configFile != "config.yaml" {
		// 如果指定了具体的配置文件，直接加载
		err = configManager.LoadFile(*configFile)
		if err != nil {
			logger.Fatal("无法加载指定的配置文件", map[string]interface{}{
				"file":  *configFile,
				"error": err,
			})
		}
		logger.Info("配置加载成功", map[string]interface{}{
			"file": *configFile,
			"env":  environment,
		})
		fmt.Printf("配置加载成功: %s\n", *configFile)
	} else {
		// 尝试从不同路径加载配置
		configPaths := []string{"./configs", "."}
		var configLoaded bool
		
		for _, path := range configPaths {
			err := configManager.Load(path)
			if err == nil {
				configLoaded = true
				logger.Info("配置加载成功", map[string]interface{}{
					"path": path,
					"env":  environment,
				})
				fmt.Printf("配置加载成功: %s\n", path)
				break
			}
		}
		
		if !configLoaded {
			logger.Fatal("无法加载配置文件", map[string]interface{}{
				"paths":       configPaths,
				"config_file": *configFile,
			})
		}
	}

	cfg := configManager.Get()

	// 初始化JWT配置（保持向后兼容）
	utils.InitJWTConfig(cfg.Auth.JWTSecret, cfg.Auth.TokenExpiry)

	// 初始化应用程序
	fmt.Println("正在初始化应用程序...")
	application, err := app.NewApplication(cfg)
	if err != nil {
		logger.Fatal("应用程序初始化失败", map[string]interface{}{
			"error": err,
		})
	}
	defer application.Shutdown()

	logger.Info("应用程序初始化成功", nil)
	fmt.Println("应用程序初始化成功")

	// 初始化业务服务
	fmt.Println("正在初始化业务服务...")
	
	// 获取数据库连接
	mongoDB := application.DB.MongoDB.GetDatabase()
	redisClient := application.DB.Redis.GetClient()

	// 子域名枚举服务
	subdomainResolver := subdomain.NewResolver()
	subdomainEnumConfig := models.SubdomainEnumConfig{
		Methods:          cfg.Subdomain.Methods,
		DictionaryPath:   cfg.Subdomain.DictionaryPath,
		Concurrency:      cfg.Subdomain.Concurrency,
		Timeout:          cfg.Subdomain.Timeout,
		RetryCount:       cfg.Subdomain.RetryCount,
		RateLimit:        cfg.Subdomain.RateLimit,
		ResolverServers:  cfg.Subdomain.ResolverServers,
		VerifySubdomains: cfg.Subdomain.VerifySubdomains,
		RecursiveSearch:  cfg.Subdomain.RecursiveSearch,
		SaveToDB:         true,
	}
	_ = subdomain.NewEnumerator(mongoDB, subdomainResolver, subdomainEnumConfig)
	logger.Info("子域名枚举服务初始化成功", nil)
	fmt.Println("子域名枚举服务初始化成功")

	// 端口扫描服务
	portScanResultHandler := portscan.NewMongoResultHandler(mongoDB)
	portScanTaskManager := portscan.NewTaskManager(mongoDB, portScanResultHandler)
	logger.Info("端口扫描服务初始化成功", nil)
	fmt.Println("端口扫描服务初始化成功")

	// 漏洞扫描服务
	vulnHandler := vulnscan.NewVulnHandler(mongoDB)
	logger.Info("漏洞扫描服务初始化成功", nil)
	fmt.Println("漏洞扫描服务初始化成功")

	// 插件服务
	vulnRegistry := vulnscan.NewRegistry()
	pluginRegistry := plugin.NewRegistry()
	pluginStore := plugin.NewMongoMetadataStore(mongoDB, "plugin_metadata")
	pluginManager := plugin.NewManager(pluginRegistry, pluginStore)

	// 漏洞扫描引擎和注册表
	vulnEngine := vulnscan.NewEngine(vulnRegistry, mongoDB, vulnHandler)

	// 漏洞数据库服务
	vulndbConfig := vulndb.Config{
		UpdateInterval: 24 * time.Hour,
		CVEConfig: vulndb.CVEConfig{
			APIURL:    "https://services.nvd.nist.gov/rest/json",
			Timeout:   30 * time.Second,
			BatchSize: 100,
		},
		CWEConfig: vulndb.CWEConfig{
			XMLURL:  "https://cwe.mitre.org/data/xml/cwec_latest.xml.zip",
			Timeout: 60 * time.Second,
		},
		CNVDConfig: vulndb.CNVDConfig{
			APIURL:    "https://www.cnvd.org.cn/flaw/list",
			Timeout:   30 * time.Second,
			BatchSize: 50,
		},
	}
	vulndbService := vulndb.NewService(mongoDB, vulndbConfig)
	_ = vulndb.NewScheduler(vulndbService, vulndbConfig)
	_ = vulndbService // 暂时标记为未使用

	// 资产发现服务
	redisResultHandler := assetdiscovery.NewRedisResultHandler(redisClient)
	discoveryService := assetdiscovery.NewDiscoveryService(mongoDB, redisResultHandler)
	_ = assetdiscovery.NewHandler(mongoDB, discoveryService) // 保留变量但标记为未使用，将来可能用于依赖注入
	logger.Info("资产发现服务初始化成功", nil)
	fmt.Println("资产发现服务初始化成功")

	// 页面监控服务
	monitoringService := pagemonitoring.NewPageMonitoringService(mongoDB, redisClient)
	if err := monitoringService.Start(); err != nil {
		logger.Fatal("启动页面监控服务失败", map[string]interface{}{
			"error": err,
		})
	}
	defer monitoringService.Stop()
	logger.Info("页面监控服务启动成功", nil)
	fmt.Println("页面监控服务启动成功")

	// 敏感信息检测服务
	_ = sensitive.NewService(mongoDB) // 保留变量但标记为未使用，将来可能用于依赖注入
	logger.Info("敏感信息检测服务初始化成功", nil)
	fmt.Println("敏感信息检测服务初始化成功")

	// 节点管理服务
	nodeManagerConfig := nodemanager.NodeManagerConfig{
		HeartbeatInterval: cfg.Node.HeartbeatInterval,
		HeartbeatTimeout:  cfg.Node.HeartbeatTimeout,
		EnableAutoRemove:  cfg.Node.EnableAutoRemove,
		AutoRemoveAfter:   cfg.Node.AutoRemoveAfter,
	}
	nodeManager := nodemanager.NewNodeManager(mongoDB, redisClient, nodeManagerConfig)
	if err := nodeManager.Start(); err != nil {
		logger.Fatal("启动节点管理服务失败", map[string]interface{}{
			"error": err,
		})
	}
	defer nodeManager.Stop()
	logger.Info("节点管理服务启动成功", nil)
	fmt.Println("节点管理服务启动成功")

	// 任务管理服务
	taskManagerConfig := taskmanager.TaskManagerConfig{
		MaxConcurrentTasks: cfg.Task.MaxConcurrentTasks,
		TaskTimeout:        cfg.Task.TaskTimeout,
		EnableRetry:        cfg.Task.EnableRetry,
		MaxRetries:         cfg.Task.MaxRetries,
		RetryInterval:      cfg.Task.RetryInterval,
	}
	taskManager := taskmanager.NewTaskManager(mongoDB, redisClient, nodeManager, taskManagerConfig)
	if err := taskManager.Start(); err != nil {
		logger.Fatal("启动任务管理服务失败", map[string]interface{}{
			"error": err,
		})
	}
	defer taskManager.Stop()
	logger.Info("任务管理服务启动成功", nil)
	fmt.Println("任务管理服务启动成功")

	logger.Info("所有业务服务初始化完成", nil)
	fmt.Println("所有业务服务初始化完成")

	// 设置Gin模式
	switch cfg.Server.Mode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		if configManager.IsProduction() {
			gin.SetMode(gin.ReleaseMode)
		} else {
			gin.SetMode(gin.DebugMode)
		}
	}

	logger.Info("Gin模式设置", map[string]interface{}{
		"mode": gin.Mode(),
	})
	fmt.Println("Gin模式设置为:", gin.Mode())

	// 自动迁移数据库模型
	if application.DB.GORM != nil {
		fmt.Println("正在进行数据库迁移...")
		err = application.DB.AutoMigrate(
			&models.UserSQL{},
			// 添加其他模型...
		)
		if err != nil {
			logger.Error("数据库迁移失败", map[string]interface{}{
				"error": err,
			})
		} else {
			logger.Info("数据库迁移完成", nil)
			fmt.Println("数据库迁移完成")
		}
	}

	// 使用统一优雅的路由注册系统
	fmt.Println("正在注册路由...")
	
	// 创建插件市场实例 (临时占位符)
	marketplaceConfig := plugin.MarketplaceConfig{
		// TODO: 配置市场参数
	}
	pluginMarketplace := plugin.NewMarketplace(marketplaceConfig)
	
	// 创建应用依赖容器
	deps := &router.AppDependencies{
		DB:                    application.DB,
		MongoDB:               mongoDB,
		RedisClient:           redisClient,
		Config:                cfg,
		NodeManager:           nodeManager,
		TaskManager:           taskManager,
		VulnEngine:            vulnEngine,
		VulnHandler:           vulnHandler,
		VulnRegistry:          vulnRegistry,
		PortScanTaskManager:   portScanTaskManager,
		PluginManager:         pluginManager,
		PluginStore:           pluginStore,
		PluginMarketplace:     pluginMarketplace,
		MonitoringService:     monitoringService,
	}
	
	// 创建Gin引擎并设置所有路由
	ginRouter := gin.New()
	routeManager := router.SetupAllRoutes(ginRouter, deps)
	
	logger.Info("统一路由系统注册成功", nil)
	fmt.Println("统一路由系统注册成功")

	// 显示所有路由
	if *showRoutes {
		printAllRoutes(ginRouter)
	}

	// 保存路由到文件
	if *saveRoutesFile != "" {
		if err := saveRoutesToFile(ginRouter, *saveRoutesFile); err != nil {
			logger.Error("保存路由到文件失败", map[string]interface{}{
				"error": err,
				"file":  *saveRoutesFile,
			})
		} else {
			logger.Info("路由信息已保存", map[string]interface{}{
				"file": *saveRoutesFile,
			})
		}
	}

	// 创建HTTP服务器
	serverAddress := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         serverAddress,
		Handler:      ginRouter,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 避免未使用变量警告
	_ = routeManager

	// 启动服务器
	go func() {
		logger.Info("服务器正在启动", map[string]interface{}{
			"address":     serverAddress,
			"environment": environment,
		})
		fmt.Println("服务器正在启动，监听地址:", serverAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("启动服务器失败", map[string]interface{}{
				"error": err,
			})
		}
	}()

	// 优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("收到关闭信号，服务器正在关闭...", nil)
	fmt.Println("收到关闭信号，服务器正在关闭...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("服务器关闭失败", map[string]interface{}{
			"error": err,
		})
	}
	
	logger.Info("服务器已成功关闭", nil)
	fmt.Println("服务器已成功关闭")
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket升级失败", map[string]interface{}{
			"error": err,
		})
		return
	}
	defer conn.Close()
	
	logger.Info("WebSocket连接已建立", map[string]interface{}{
		"remote_addr": c.Request.RemoteAddr,
	})

	// 简单地回显所有收到的消息
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			logger.Error("读取WebSocket消息失败", map[string]interface{}{
				"error": err,
			})
			break
		}

		logger.Debug("收到WebSocket消息", map[string]interface{}{
			"message": string(p),
		})
		
		if err := conn.WriteMessage(messageType, p); err != nil {
			logger.Error("发送WebSocket消息失败", map[string]interface{}{
				"error": err,
			})
			break
		}
	}
	
	logger.Info("WebSocket连接已关闭", nil)
}

func printAllRoutes(engine *gin.Engine) {
	content, err := getRoutesContent(engine)
	if err != nil {
		fmt.Printf("获取路由信息失败: %v\n", err)
		logger.Error("获取路由信息失败", map[string]interface{}{
			"error": err,
		})
		return
	}
	fmt.Println(content)
}

func saveRoutesToFile(engine *gin.Engine, filename string) error {
	content, err := getRoutesContent(engine)
	if err != nil {
		return fmt.Errorf("获取路由内容失败: %w", err)
	}

	err = os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("写入路由文件失败: %w", err)
	}

	fmt.Printf("所有路由信息已保存到 %s\n", filename)
	return nil
}

func getRoutesContent(engine *gin.Engine) (string, error) {
	var sb strings.Builder
	routes := engine.Routes()
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Path < routes[j].Path
	})

	sb.WriteString("----------------- All Routes -----------------\n")
	for _, route := range routes {
		sb.WriteString(fmt.Sprintf("%-6s %-25s --> %s\n", route.Method, route.Path, route.Handler))
	}
	sb.WriteString("----------------------------------------------\n")

	return sb.String(), nil
}