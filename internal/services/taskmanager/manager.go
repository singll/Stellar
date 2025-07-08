package taskmanager

import (
	"context"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TaskManager 任务管理器
type TaskManager struct {
	db           *mongo.Database
	redisClient  *redis.Client
	nodeManager  NodeManager
	tasks        map[string]*TaskInfo
	tasksMutex   sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	eventChan    chan TaskEvent
	config       TaskManagerConfig
	queues       map[string]*TaskQueue
	queuesMutex  sync.RWMutex
	dispatcher   *TaskDispatcher
	queueManager *QueueManager
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

	return &TaskManager{
		db:           db,
		redisClient:  redisClient,
		nodeManager:  nodeManager,
		dispatcher:   dispatcher,
		queueManager: queueManager,
		config:       config,
		ctx:          ctx,
		cancel:       cancel,
	}
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

	// 启动任务监控
	go tm.monitorTasks()

	return nil
}

// Stop 停止任务管理器
func (tm *TaskManager) Stop() {
	tm.cancel()
	tm.dispatcher.Stop()
}

// SubmitTask 提交任务
func (tm *TaskManager) SubmitTask(task *models.Task) error {
	return tm.dispatcher.SubmitTask(task)
}

// CancelTask 取消任务
func (tm *TaskManager) CancelTask(taskID string) error {
	return tm.dispatcher.CancelTask(taskID)
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
	return tm.dispatcher.GetTaskResult(taskID)
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
		_ = tm.UpdateTaskStatus(task.ID.Hex(), models.TaskStatusTimeout, task.Progress)

		// 保存任务结果
		_ = tm.SaveTaskResult(task.ID.Hex(), &models.TaskResult{
			Status: "timeout",
			Error:  "任务执行超时",
		})

		// 如果任务配置了重试，则重新提交任务
		if tm.config.EnableRetry && task.RetryCount < tm.config.MaxRetries {
			task.RetryCount++
			task.Status = models.TaskStatusPending
			_ = tm.SubmitTask(&task)
		}
	}
}
