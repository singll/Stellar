package taskmanager

import (
	"context"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/services/taskmanager/executors"
	"github.com/StellarServer/internal/services/vulnscan"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TaskManager 任务管理器
type TaskManager struct {
	db            *mongo.Database
	redisClient   *redis.Client
	nodeManager   NodeManager
	tasks         map[string]*TaskInfo
	tasksMutex    sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	eventChan     chan TaskEvent
	config        TaskManagerConfig
	queues        map[string]*TaskQueue
	queuesMutex   sync.RWMutex
	dispatcher    *TaskDispatcher
	queueManager  *QueueManager
	executor      *ExecutionEngine
	scheduler     *TaskScheduler
}

// TaskManagerConfig 任务管理器配置
type TaskManagerConfig struct {
	MaxConcurrentTasks int  // 最大并发任务数
	TaskTimeout        int  // 任务超时时间(秒)
	EnableRetry        bool // 是否启用重试
	MaxRetries         int  // 最大重试次数
	RetryInterval      int  // 重试间隔(秒)
}

// TaskInfo 任务信息
type TaskInfo struct {
	Task       *models.Task
	Status     string
	Progress   float64
	StartTime  time.Time
	EndTime    time.Time
	RetryCount int
	NodeID     string
	Mutex      sync.Mutex
}

// TaskQueue 任务队列
type TaskQueue struct {
	Name      string
	Type      string
	Priority  int
	Tasks     []*models.Task
	MaxSize   int
	TaskCount int
	Mutex     sync.Mutex
}

// TaskEvent 任务事件
type TaskEvent struct {
	Type      string      // 事件类型
	TaskID    string      // 任务ID
	Timestamp time.Time   // 时间戳
	Data      interface{} // 事件数据
}

// NodeManager 节点管理器接口
type NodeManager interface {
	GetAllNodes() []*models.Node
	GetNodesByRole(role string) []*models.Node
	GetNodesByStatus(status string) []*models.Node
	GetNode(nodeID string) (*models.Node, error)
}

// NewTaskManager 创建任务管理器
func NewTaskManager(db *mongo.Database, redisClient *redis.Client, nodeManager NodeManager, config TaskManagerConfig) *TaskManager {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建队列管理器
	queueManager := NewQueueManager(db, redisClient)

	// 创建任务分发器
	dispatcher := NewTaskDispatcher(db, redisClient, nodeManager, queueManager, config.MaxConcurrentTasks)

	// 创建执行引擎
	executorConfig := ExecutionConfig{
		MaxConcurrentTasks: config.MaxConcurrentTasks,
		DefaultTimeout:     time.Duration(config.TaskTimeout) * time.Second,
		EnableRetry:        config.EnableRetry,
		MaxRetries:         config.MaxRetries,
		RetryInterval:      time.Duration(config.RetryInterval) * time.Second,
		ResultBufferSize:   100,
	}
	executor := NewExecutionEngine(db, redisClient, executorConfig)

	// 创建任务调度器
	scheduler := NewTaskScheduler(db, nil) // 稍后会设置taskManager引用

	tm := &TaskManager{
		db:           db,
		redisClient:  redisClient,
		nodeManager:  nodeManager,
		dispatcher:   dispatcher,
		queueManager: queueManager,
		executor:     executor,
		scheduler:    scheduler,
		config:       config,
		ctx:          ctx,
		cancel:       cancel,
		tasks:        make(map[string]*TaskInfo),
		queues:       make(map[string]*TaskQueue),
		eventChan:    make(chan TaskEvent, 100),
	}

	// 设置调度器的taskManager引用
	scheduler.taskManager = tm

	// 注册默认执行器
	tm.registerDefaultExecutors()

	return tm
}

// registerDefaultExecutors 注册默认执行器
func (tm *TaskManager) registerDefaultExecutors() {
	// 注册子域名枚举执行器
	subdomainExecutor := executors.NewSubdomainExecutor(executors.SubdomainConfig{
		MaxWorkers:     50,
		Timeout:        5 * time.Second,
		DNSServers:     []string{"8.8.8.8", "1.1.1.1", "114.114.114.114"},
		EnableWildcard: true,
		MaxRetries:     3,
	})
	tm.executor.RegisterExecutor("subdomain_enum", subdomainExecutor)

	// 注册端口扫描执行器
	portScanExecutor := executors.NewPortScanExecutor(executors.PortScanConfig{
		MaxWorkers:     100,
		Timeout:        30 * time.Second,
		ConnectTimeout: 3 * time.Second,
		EnableBanner:   true,
		BannerTimeout:  5 * time.Second,
		MaxRetries:     2,
	})
	tm.executor.RegisterExecutor("port_scan", portScanExecutor)

	// 注册资产发现执行器
	assetDiscoveryExecutor := executors.NewAssetDiscoveryExecutor(tm.db, executors.AssetDiscoveryConfig{
		EnableDomainAssets:    true,
		EnableSubdomainAssets: true,
		EnableIPAssets:        true,
		EnablePortAssets:      true,
		EnableURLAssets:       true,
		AutoCreateAssets:      true,
	})
	tm.executor.RegisterExecutor("asset_discovery", assetDiscoveryExecutor)

	// 注册漏洞扫描执行器
	vulnHandler := vulnscan.NewVulnHandler(tm.db)
	pocEngineConfig := vulnscan.POCEngineConfig{
		MaxConcurrency: 10,
		Timeout:        30 * time.Second,
		RateLimit:      10.0,
		EnableCache:    true,
		CacheTTL:       time.Hour,
		EnableSandbox:  true,
		MaxMemoryMB:    512,
		MaxScriptSize:  1024 * 1024, // 1MB
	}
	pocEngine := vulnscan.NewPOCEngine(pocEngineConfig)
	vulnScanExecutor := executors.NewVulnScanExecutor(tm.db, pocEngine, vulnHandler)
	tm.executor.RegisterExecutor("vuln_scan", vulnScanExecutor)
}

// Start 启动任务管理器
func (tm *TaskManager) Start() error {
	// 启动队列管理器
	if err := tm.queueManager.Start(); err != nil {
		return err
	}

	// 启动任务分发器
	if err := tm.dispatcher.Start(); err != nil {
		return err
	}

	// 启动任务调度器
	if err := tm.scheduler.Start(); err != nil {
		return err
	}

	// 启动任务监控
	go tm.monitorTasks()

	return nil
}

// Stop 停止任务管理器
func (tm *TaskManager) Stop() {
	tm.cancel()
	tm.scheduler.Stop()
	tm.dispatcher.Stop()
	tm.executor.Shutdown()
}

// SubmitTask 提交任务
func (tm *TaskManager) SubmitTask(task *models.Task) error {
	// 先保存任务到数据库
	if err := tm.saveTask(task); err != nil {
		return err
	}

	// 直接使用执行引擎执行任务
	return tm.executor.ExecuteTask(task)
}

// saveTask 保存任务到数据库
func (tm *TaskManager) saveTask(task *models.Task) error {
	collection := tm.db.Collection("tasks")
	
	if task.ID.IsZero() {
		task.ID = primitive.NewObjectID()
	}
	
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	
	task.UpdatedAt = time.Now()
	
	_, err := collection.InsertOne(tm.ctx, task)
	return err
}

// CancelTask 取消任务
func (tm *TaskManager) CancelTask(taskID string) error {
	return tm.executor.CancelTask(taskID)
}

// GetTask 获取任务
func (tm *TaskManager) GetTask(taskID string) (*models.Task, error) {
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, err
	}

	var task models.Task
	err = tm.db.Collection("tasks").FindOne(tm.ctx, bson.M{"_id": objID}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &task, nil
}

// ListTasks 列出任务
func (tm *TaskManager) ListTasks(projectID, status, taskType string, limit, offset int) ([]*models.Task, int64, error) {
	filter := bson.M{}

	// 添加过滤条件
	if projectID != "" {
		objID, err := primitive.ObjectIDFromHex(projectID)
		if err == nil {
			filter["projectId"] = objID
		}
	}

	if status != "" {
		filter["status"] = status
	}

	if taskType != "" {
		filter["type"] = taskType
	}

	// 设置分页
	if limit <= 0 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	// 查询总数
	total, err := tm.db.Collection("tasks").CountDocuments(tm.ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 查询任务列表
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := tm.db.Collection("tasks").Find(tm.ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(tm.ctx)

	var tasks []*models.Task
	if err := cursor.All(tm.ctx, &tasks); err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// UpdateTaskStatus 更新任务状态
func (tm *TaskManager) UpdateTaskStatus(taskID string, status string, progress float64) error {
	return tm.dispatcher.UpdateTaskStatus(taskID, status, progress)
}

// GetTaskResult 获取任务结果
func (tm *TaskManager) GetTaskResult(taskID string) (*models.TaskResult, error) {
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, err
	}

	var result models.TaskResult
	err = tm.db.Collection("task_results").FindOne(tm.ctx, bson.M{"task_id": objID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &result, nil
}

// GetRunningTasks 获取运行中的任务
func (tm *TaskManager) GetRunningTasks() []string {
	return tm.executor.GetRunningTasks()
}

// GetExecutors 获取已注册的执行器
func (tm *TaskManager) GetExecutors() map[string]models.ExecutorInfo {
	return tm.executor.GetExecutors()
}

// RegisterExecutor 注册执行器
func (tm *TaskManager) RegisterExecutor(taskType string, executor TaskExecutor) error {
	return tm.executor.RegisterExecutor(taskType, executor)
}

// UnregisterExecutor 注销执行器
func (tm *TaskManager) UnregisterExecutor(taskType string) error {
	return tm.executor.UnregisterExecutor(taskType)
}

// SaveTaskResult 保存任务结果
func (tm *TaskManager) SaveTaskResult(taskID string, result *models.TaskResult) error {
	return tm.dispatcher.SaveTaskResult(taskID, result)
}

// monitorTasks 监控任务
func (tm *TaskManager) monitorTasks() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-tm.ctx.Done():
			return
		case <-ticker.C:
			tm.checkStuckTasks()
		}
	}
}

// checkStuckTasks 检查卡住的任务
func (tm *TaskManager) checkStuckTasks() {
	// 查找运行时间超过超时时间的任务
	timeout := time.Now().Add(-time.Duration(tm.config.TaskTimeout) * time.Second)

	filter := bson.M{
		"status":    models.TaskStatusRunning,
		"startedAt": bson.M{"$lt": timeout},
	}

	cursor, err := tm.db.Collection("tasks").Find(tm.ctx, filter)
	if err != nil {
		return
	}
	defer cursor.Close(tm.ctx)

	var stuckTasks []models.Task
	if err := cursor.All(tm.ctx, &stuckTasks); err != nil {
		return
	}

	for _, task := range stuckTasks {
		// 更新任务状态为超时
		_ = tm.UpdateTaskStatus(task.ID.Hex(), string(models.TaskStatusTimeout), task.Progress)

		// 保存任务结果
		_ = tm.SaveTaskResult(task.ID.Hex(), &models.TaskResult{
			Status: "timeout",
			Error:  "任务执行超时",
		})

		// 如果任务配置了重试，则重新提交任务
		if tm.config.EnableRetry && task.RetryCount < tm.config.MaxRetries {
			task.RetryCount++
			task.Status = string(models.TaskStatusPending)
			_ = tm.SubmitTask(&task)
		}
	}
}

// ==================== 任务调度相关方法 ====================

// CreateScheduleRule 创建调度规则
func (tm *TaskManager) CreateScheduleRule(rule *models.TaskScheduleRule) error {
	return tm.scheduler.CreateScheduleRule(rule)
}

// UpdateScheduleRule 更新调度规则
func (tm *TaskManager) UpdateScheduleRule(ruleID string, updates *models.TaskScheduleRule) error {
	return tm.scheduler.UpdateScheduleRule(ruleID, updates)
}

// DeleteScheduleRule 删除调度规则
func (tm *TaskManager) DeleteScheduleRule(ruleID string) error {
	return tm.scheduler.DeleteScheduleRule(ruleID)
}

// ToggleScheduleRule 切换调度规则状态
func (tm *TaskManager) ToggleScheduleRule(ruleID string, enabled bool) error {
	return tm.scheduler.ToggleScheduleRule(ruleID, enabled)
}

// TriggerScheduleRule 手动触发调度规则
func (tm *TaskManager) TriggerScheduleRule(ruleID string) (*models.Task, error) {
	return tm.scheduler.TriggerScheduleRule(ruleID)
}

// GetScheduleRules 获取调度规则列表
func (tm *TaskManager) GetScheduleRules(projectID string) ([]*models.TaskScheduleRule, error) {
	return tm.scheduler.GetScheduleRules(projectID)
}

// GetSchedulerStats 获取调度器统计信息
func (tm *TaskManager) GetSchedulerStats() map[string]interface{} {
	return tm.scheduler.GetStats()
}
