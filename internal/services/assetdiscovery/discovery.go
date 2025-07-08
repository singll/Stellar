package assetdiscovery

import (
	"context"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/time/rate"
)

// DiscoveryService 资产发现服务
type DiscoveryService struct {
	db            *mongo.Database
	limiter       *rate.Limiter
	taskMap       map[string]*DiscoveryTask
	taskMutex     sync.RWMutex
	resultHandler ResultHandler
}

// DiscoveryTask 资产发现任务
type DiscoveryTask struct {
	ID         string
	Task       *models.AssetDiscoveryTask
	Context    context.Context
	CancelFunc context.CancelFunc
	Progress   float64
	Status     string
	StartTime  time.Time
	EndTime    time.Time
	Results    []*models.DiscoveryResult
	Error      error
	Mutex      sync.Mutex
}

// DiscoveryOptions 资产发现选项
type DiscoveryOptions struct {
	Concurrency    int
	Timeout        time.Duration
	RetryCount     int
	RateLimit      int
	FollowRedirect bool
	CustomHeaders  map[string]string
	Cookies        string
	Proxy          string
	ScanDepth      int
}

// ResultHandler 结果处理器接口
type ResultHandler interface {
	HandleDiscoveryResult(task *DiscoveryTask, result *models.DiscoveryResult) error
	UpdateTaskStatus(taskID string, status string, progress float64) error
	SaveDiscoveryResult(result *models.DiscoveryResult) error
	GetDiscoveryResults(taskID string) ([]*models.DiscoveryResult, error)
}

// NewDiscoveryService 创建资产发现服务
func NewDiscoveryService(db *mongo.Database, handler ResultHandler) *DiscoveryService {
	return &DiscoveryService{
		db:            db,
		limiter:       rate.NewLimiter(rate.Limit(10), 1), // 默认限制为每秒10个请求
		taskMap:       make(map[string]*DiscoveryTask),
		resultHandler: handler,
	}
}

// CreateTask 创建资产发现任务
func (s *DiscoveryService) CreateTask(task *models.AssetDiscoveryTask) (string, error) {
	// 生成任务ID
	taskID := primitive.NewObjectID().Hex()

	// 设置任务初始状态
	task.ID = primitive.NewObjectID()
	task.Status = "pending"
	task.CreatedAt = time.Now()
	task.Progress = 0

	// 保存任务到数据库
	_, err := s.db.Collection("asset_discovery_tasks").InsertOne(context.Background(), task)
	if err != nil {
		return "", err
	}

	return taskID, nil
}

// StartTask 启动资产发现任务
func (s *DiscoveryService) StartTask(taskID string) error {
	// 查询任务信息
	var task models.AssetDiscoveryTask
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return err
	}

	err = s.db.Collection("asset_discovery_tasks").FindOne(context.Background(), map[string]interface{}{
		"_id": objID,
	}).Decode(&task)
	if err != nil {
		return err
	}

	// 创建上下文，用于取消任务
	ctx, cancel := context.WithCancel(context.Background())

	// 创建任务
	discoveryTask := &DiscoveryTask{
		ID:         taskID,
		Task:       &task,
		Context:    ctx,
		CancelFunc: cancel,
		Progress:   0,
		Status:     "running",
		StartTime:  time.Now(),
		Results:    make([]*models.DiscoveryResult, 0),
	}

	// 添加到任务映射
	s.taskMutex.Lock()
	s.taskMap[taskID] = discoveryTask
	s.taskMutex.Unlock()

	// 更新任务状态
	task.Status = "running"
	task.StartedAt = time.Now()
	_, err = s.db.Collection("asset_discovery_tasks").UpdateOne(context.Background(), map[string]interface{}{
		"_id": task.ID,
	}, map[string]interface{}{
		"$set": map[string]interface{}{
			"status":    "running",
			"startedAt": task.StartedAt,
		},
	})
	if err != nil {
		cancel()
		return err
	}

	// 启动任务执行
	go s.runTask(discoveryTask)

	return nil
}

// StopTask 停止资产发现任务
func (s *DiscoveryService) StopTask(taskID string) error {
	s.taskMutex.RLock()
	task, exists := s.taskMap[taskID]
	s.taskMutex.RUnlock()

	if !exists {
		return nil // 任务不存在或已完成
	}

	// 取消任务
	task.CancelFunc()

	// 更新任务状态
	task.Mutex.Lock()
	task.Status = "stopped"
	task.EndTime = time.Now()
	task.Mutex.Unlock()

	// 更新数据库中的任务状态
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return err
	}

	_, err = s.db.Collection("asset_discovery_tasks").UpdateOne(context.Background(), map[string]interface{}{
		"_id": objID,
	}, map[string]interface{}{
		"$set": map[string]interface{}{
			"status":      "stopped",
			"completedAt": time.Now(),
		},
	})

	return err
}

// GetTaskStatus 获取任务状态
func (s *DiscoveryService) GetTaskStatus(taskID string) (string, error) {
	s.taskMutex.RLock()
	task, exists := s.taskMap[taskID]
	s.taskMutex.RUnlock()

	if exists {
		task.Mutex.Lock()
		status := task.Status
		task.Mutex.Unlock()
		return status, nil
	}

	// 任务不在内存中，从数据库查询
	var dbTask models.AssetDiscoveryTask
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return "", err
	}

	err = s.db.Collection("asset_discovery_tasks").FindOne(context.Background(), map[string]interface{}{
		"_id": objID,
	}).Decode(&dbTask)
	if err != nil {
		return "", err
	}

	return dbTask.Status, nil
}

// GetTaskProgress 获取任务进度
func (s *DiscoveryService) GetTaskProgress(taskID string) (float64, error) {
	s.taskMutex.RLock()
	task, exists := s.taskMap[taskID]
	s.taskMutex.RUnlock()

	if exists {
		task.Mutex.Lock()
		progress := task.Progress
		task.Mutex.Unlock()
		return progress, nil
	}

	// 任务不在内存中，从数据库查询
	var dbTask models.AssetDiscoveryTask
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return 0, err
	}

	err = s.db.Collection("asset_discovery_tasks").FindOne(context.Background(), map[string]interface{}{
		"_id": objID,
	}).Decode(&dbTask)
	if err != nil {
		return 0, err
	}

	return dbTask.Progress, nil
}

// runTask 执行资产发现任务
func (s *DiscoveryService) runTask(task *DiscoveryTask) {
	defer func() {
		// 任务完成，从映射中删除
		s.taskMutex.Lock()
		delete(s.taskMap, task.ID)
		s.taskMutex.Unlock()

		// 关闭上下文
		task.CancelFunc()
	}()

	// 更新任务状态为运行中
	task.Mutex.Lock()
	task.Status = "running"
	task.Mutex.Unlock()

	// 根据任务类型执行不同的发现逻辑
	var err error
	switch task.Task.DiscoveryType {
	case "network":
		err = s.runNetworkDiscovery(task)
	case "service":
		err = s.runServiceDiscovery(task)
	case "web":
		err = s.runWebDiscovery(task)
	default:
		err = s.runNetworkDiscovery(task) // 默认为网络发现
	}

	// 检查任务是否被取消
	select {
	case <-task.Context.Done():
		// 任务已被取消
		task.Mutex.Lock()
		task.Status = "stopped"
		task.EndTime = time.Now()
		task.Mutex.Unlock()

		// 更新数据库中的任务状态
		s.resultHandler.UpdateTaskStatus(task.ID, "stopped", task.Progress)
		return
	default:
		// 任务正常完成
	}

	// 处理任务完成
	task.Mutex.Lock()
	if err != nil {
		task.Status = "failed"
		task.Error = err
	} else {
		task.Status = "completed"
		task.Progress = 100
	}
	task.EndTime = time.Now()
	task.Mutex.Unlock()

	// 更新数据库中的任务状态
	status := "completed"
	if err != nil {
		status = "failed"
	}

	s.resultHandler.UpdateTaskStatus(task.ID, status, task.Progress)

	// 更新任务完成时间
	objID, _ := primitive.ObjectIDFromHex(task.ID)
	s.db.Collection("asset_discovery_tasks").UpdateOne(context.Background(), map[string]interface{}{
		"_id": objID,
	}, map[string]interface{}{
		"$set": map[string]interface{}{
			"completedAt": time.Now(),
		},
	})
}

// runNetworkDiscovery 执行网络资产发现
func (s *DiscoveryService) runNetworkDiscovery(task *DiscoveryTask) error {
	// 创建网络扫描器
	networkScanner := NewNetworkScanner(
		task.Task.Config.Concurrency,
		time.Duration(task.Task.Config.Timeout)*time.Second,
		task.Task.Config.RetryCount,
		s.resultHandler,
	)

	// 调用网络扫描器的ScanNetwork方法
	return networkScanner.ScanNetwork(task)
}

// runServiceDiscovery 执行服务资产发现
func (s *DiscoveryService) runServiceDiscovery(task *DiscoveryTask) error {
	// 创建服务扫描器
	serviceScanner := NewServiceScanner(
		task.Task.Config.Concurrency,
		time.Duration(task.Task.Config.Timeout)*time.Second,
		task.Task.Config.RetryCount,
		s.resultHandler,
	)

	// 调用服务扫描器的ScanServices方法
	return serviceScanner.ScanServices(task)
}

// Web资产发现
func (s *DiscoveryService) runWebDiscovery(task *DiscoveryTask) error {
	// 这里实现Web资产发现的逻辑
	// 例如网站爬取、指纹识别等

	// TODO: 实现具体的Web发现逻辑

	return nil
}
