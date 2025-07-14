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

	"github.com/StellarServer/internal/api"
	"github.com/StellarServer/internal/config"
	"github.com/StellarServer/internal/database"
	"github.com/StellarServer/internal/models"
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
)

func banner() {
	fmt.Println(`
   _____ _       _ _              _____
  / ____| |     | | |            / ____|
 | (___ | |_ ___| | | __ _ _ __ | (___   ___ _ ____   _____ _ __
  \___ \| __/ _ \ | |/ _' | '__|  \___ \ / _ \ '__\ \ / / _ \ '__|
  ____) | ||  __/ | | (_| | |     ____) |  __/ |   \ V /  __/ |
 |_____/ \__\___|_|_|\__,_|_|    |_____/ \___|_|    \_/ \___|_|

 StellarServer - 安全资产管理平台
 版本: ` + VERSION + `
	`)
}

func main() {
	fmt.Println("StellarServer 正在启动...")
	fmt.Println("正在解析命令行参数...")

	flag.Parse()

	banner()

	fmt.Println("初始化日志系统...")
	// 初始化日志系统
	logConfig := utils.DefaultLogConfig()
	switch *logLevel {
	case "debug":
		logConfig.Level = utils.DebugLevel
	case "info":
		logConfig.Level = utils.InfoLevel
	case "warn":
		logConfig.Level = utils.WarnLevel
	case "error":
		logConfig.Level = utils.ErrorLevel
	default:
		logConfig.Level = utils.InfoLevel
	}

	if err := utils.InitGlobalLogger(logConfig); err != nil {
		fmt.Printf("初始化日志系统失败: %v\n", err)
		log.Fatalf("初始化日志系统失败: %v", err)
	}

	utils.Info("StellarServer启动中", "version", VERSION)

	// 初始化性能监控
	if *enableMonitor {
		utils.InitGlobalMonitor(10*time.Second, 100)
		defer utils.StopGlobalMonitor()
		utils.Info("性能监控已启用")
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

	// 加载配置
	fmt.Println("正在加载配置文件:", *configFile)
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		utils.Fatal("加载配置失败", err)
	}

	utils.InitJWTConfig(cfg.Auth.JWTSecret, cfg.Auth.TokenExpiry)

	utils.Info("配置加载成功", "config", *configFile)
	fmt.Println("配置加载成功")

	// 连接数据库
	fmt.Println("正在连接MongoDB...")
	utils.Info("正在连接MongoDB...")
	mongoClient, err := database.ConnectMongoDB(cfg.MongoDB)
	if err != nil {
		fmt.Printf("连接MongoDB失败: %v\n", err)
		utils.Fatal("连接MongoDB失败", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			utils.Error("断开MongoDB连接失败", err)
		}
	}()
	utils.Info("MongoDB连接成功")
	fmt.Println("MongoDB连接成功")

	// 连接Redis
	fmt.Println("正在连接Redis...")
	utils.Info("正在连接Redis...")
	redisClient, err := database.ConnectRedis(cfg.Redis)
	if err != nil {
		fmt.Printf("连接Redis失败: %v\n", err)
		utils.Fatal("连接Redis失败", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			utils.Error("关闭Redis连接失败", err)
		}
	}()
	utils.Info("Redis连接成功")
	fmt.Println("Redis连接成功")

	// 初始化数据库
	db := mongoClient.Database(cfg.MongoDB.Database)
	utils.Info("数据库初始化成功", "database", cfg.MongoDB.Database)
	fmt.Println("数据库初始化成功:", cfg.MongoDB.Database)

	// 初始化服务
	fmt.Println("正在初始化服务...")
	// 子域名枚举服务
	subdomainResolver := subdomain.NewResolver()
	// 转换配置
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
	_ = subdomain.NewEnumerator(db, subdomainResolver, subdomainEnumConfig)
	utils.Info("子域名枚举服务初始化成功")
	fmt.Println("子域名枚举服务初始化成功")

	// 端口扫描服务
	// portScanConfig := models.PortScanConfig{
	// 	Ports:        cfg.PortScan.Ports,
	// 	Concurrency:  cfg.PortScan.Concurrency,
	// 	Timeout:      cfg.PortScan.Timeout,
	// 	RetryCount:   cfg.PortScan.RetryCount,
	// 	RateLimit:    cfg.PortScan.RateLimit,
	// 	ScanType:     cfg.PortScan.ScanType,
	// 	ExcludeHosts: cfg.PortScan.ExcludeIPs,
	// }
	// portScanner := portscan.NewScanner(portScanConfig)
	// portScanManager := portscan.NewManager(db, portScanner)
	portScanResultHandler := portscan.NewMongoResultHandler(db)
	portScanTaskManager := portscan.NewTaskManager(db, portScanResultHandler)
	// portScanHandler := portscan.NewHandler(db, portScanManager) // This handler is not used
	utils.Info("端口扫描服务初始化成功")
	fmt.Println("端口扫描服务初始化成功")

	// 漏洞扫描服务
	vulnHandler := vulnscan.NewVulnHandler(db)
	utils.Info("漏洞扫描服务初始化成功")
	fmt.Println("漏洞扫描服务初始化成功")

	// 插件服务
	vulnRegistry := vulnscan.NewRegistry()
	pluginRegistry := plugin.NewRegistry()
	pluginStore := plugin.NewMongoMetadataStore(db, "plugin_metadata")
	pluginManager := plugin.NewManager(pluginRegistry, pluginStore)

	// 漏洞扫描引擎和注册表
	vulnEngine := vulnscan.NewEngine(vulnRegistry, db, vulnHandler)

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
	vulndbService := vulndb.NewService(db, vulndbConfig) // 暂时保留，可能被其他地方引用
	_ = vulndb.NewScheduler(vulndbService, vulndbConfig) // vulndbScheduler暂时未使用
	_ = vulndbService                                     // 暂时标记为未使用

	// 资产发现服务
	redisResultHandler := assetdiscovery.NewRedisResultHandler(redisClient)
	discoveryService := assetdiscovery.NewDiscoveryService(db, redisResultHandler)
	discoveryHandler := assetdiscovery.NewHandler(db, discoveryService)
	utils.Info("资产发现服务初始化成功")
	fmt.Println("资产发现服务初始化成功")

	// 页面监控服务
	monitoringService := pagemonitoring.NewPageMonitoringService(db, redisClient)
	if err := monitoringService.Start(); err != nil {
		utils.Fatal("启动页面监控服务失败", err)
	}
	defer monitoringService.Stop()
	utils.Info("页面监控服务启动成功")
	fmt.Println("页面监控服务启动成功")

	// 敏感信息检测服务
	sensitiveService := sensitive.NewService(db)
	utils.Info("敏感信息检测服务初始化成功")
	fmt.Println("敏感信息检测服务初始化成功")

	// 节点管理服务
	nodeManagerConfig := nodemanager.NodeManagerConfig{
		HeartbeatInterval: cfg.Node.HeartbeatInterval,
		HeartbeatTimeout:  cfg.Node.HeartbeatTimeout,
		EnableAutoRemove:  cfg.Node.EnableAutoRemove,
		AutoRemoveAfter:   cfg.Node.AutoRemoveAfter,
	}
	nodeManager := nodemanager.NewNodeManager(db, redisClient, nodeManagerConfig)
	if err := nodeManager.Start(); err != nil {
		utils.Fatal("启动节点管理服务失败", err)
	}
	defer nodeManager.Stop()
	utils.Info("节点管理服务启动成功")
	fmt.Println("节点管理服务启动成功")

	// 任务管理服务
	taskManagerConfig := taskmanager.TaskManagerConfig{
		MaxConcurrentTasks: cfg.Task.MaxConcurrentTasks,
		TaskTimeout:        cfg.Task.TaskTimeout,
		EnableRetry:        cfg.Task.EnableRetry,
		MaxRetries:         cfg.Task.MaxRetries,
		RetryInterval:      cfg.Task.RetryInterval,
	}
	taskManager := taskmanager.NewTaskManager(db, redisClient, nodeManager, taskManagerConfig)
	if err := taskManager.Start(); err != nil {
		utils.Fatal("启动任务管理服务失败", err)
	}
	defer taskManager.Stop()
	utils.Info("任务管理服务启动成功")
	fmt.Println("任务管理服务启动成功")

	// 设置Gin模式
	switch cfg.Server.Mode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
	utils.Info("Gin模式设置", "mode", gin.Mode())
	fmt.Println("Gin模式设置为:", gin.Mode())

	// 初始化Gin引擎
	router := gin.Default()
	utils.Info("Gin引擎初始化成功")
	fmt.Println("Gin引擎初始化成功")

	// 注册API路由
	sensitiveHandler := api.NewSensitiveHandler(sensitiveService)
	// vulndbHandler := api.NewVulnDatabaseHandler(vulndbService, vulndbScheduler) // 暂时注释，未在routes中使用
	discoveryHandlerForAPI := api.NewDiscoveryHandler(db, discoveryHandler)
	
	// 创建插件市场实例 (临时占位符)
	marketplaceConfig := plugin.MarketplaceConfig{
		// TODO: 配置市场参数
	}
	pluginMarketplace := plugin.NewMarketplace(marketplaceConfig)
	
	api.RegisterAPIRoutes(
		router,
		db,
		redisClient,
		cfg,
		nodeManager,
		taskManager,
		vulnEngine,
		vulnHandler,
		vulnRegistry,
		portScanTaskManager,
		pluginManager,
		pluginStore,
		pluginMarketplace,
		monitoringService,
		discoveryHandlerForAPI,
		sensitiveHandler,
	)
	utils.Info("API路由注册成功")
	fmt.Println("API路由注册成功")

	// 注册WebSocket路由
	router.GET("/ws", handleWebSocket)
	utils.Info("WebSocket路由注册成功", "path", "/ws")
	fmt.Println("WebSocket路由注册成功: /ws")

	// 显示所有路由
	if *showRoutes {
		printAllRoutes(router)
	}

	// 保存路由到文件
	if *saveRoutesFile != "" {
		if err := saveRoutesToFile(router, *saveRoutesFile); err != nil {
			utils.Error("保存路由到文件失败", err)
		} else {
			utils.Info("路由信息已保存", "file", *saveRoutesFile)
		}
	}

	// 创建HTTP服务器
	serverAddress := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         serverAddress,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 启动服务器
	go func() {
		utils.Info("服务器正在启动", "address", serverAddress)
		fmt.Println("服务器正在启动，监听地址:", serverAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Fatal("启动服务器失败", err)
		}
	}()

	// 优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	utils.Info("收到关闭信号，服务器正在关闭...")
	fmt.Println("收到关闭信号，服务器正在关闭...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		utils.Fatal("服务器关闭失败", err)
	}
	utils.Info("服务器已成功关闭")
	fmt.Println("服务器已成功关闭")
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()
	log.Println("WebSocket连接已建立")

	// 简单地回显所有收到的消息
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("读取WebSocket消息失败: %v", err)
			break
		}

		log.Printf("收到消息: %s", p)
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Printf("发送WebSocket消息失败: %v", err)
			break
		}
	}
	log.Println("WebSocket连接已关闭")
}

func printAllRoutes(engine *gin.Engine) {
	content, err := getRoutesContent(engine)
	if err != nil {
		fmt.Printf("获取路由信息失败: %v\n", err)
		utils.Error("获取路由信息失败", err)
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
