package taskmanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TaskExecutor 任务执行器接口
type TaskExecutor interface {
	Execute(ctx context.Context, task *models.Task) (*models.TaskResult, error)
	GetSupportedTypes() []string
	GetExecutorInfo() models.ExecutorInfo
}

// ExecutionEngine 任务执行引擎
type ExecutionEngine struct {
	db             *mongo.Database
	redisClient    *redis.Client
	executors      map[string]TaskExecutor
	executorsMutex sync.RWMutex
	runningTasks   map[string]*ExecutionContext
	tasksMutex     sync.RWMutex
	config         ExecutionConfig
	ctx            context.Context
	cancel         context.CancelFunc
	eventChan      chan ExecutionEvent
}

// ExecutionConfig 执行引擎配置
type ExecutionConfig struct {
	MaxConcurrentTasks int           // 最大并发执行任务数
	DefaultTimeout     time.Duration // 默认任务超时时间
	EnableRetry        bool          // 是否启用重试
	MaxRetries         int           // 最大重试次数
	RetryInterval      time.Duration // 重试间隔
	ResultBufferSize   int           // 结果缓冲区大小
}

// ExecutionContext 执行上下文
type ExecutionContext struct {
	TaskID     string
	Executor   TaskExecutor
	Context    context.Context
	Cancel     context.CancelFunc
	StartTime  time.Time
	Progress   float64
	Status     string
	LastUpdate time.Time
	RetryCount int
	mutex      sync.Mutex
}

// ExecutionEvent 执行事件
type ExecutionEvent struct {
	Type      string      `json:"type"`
	TaskID    string      `json:"task_id"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
	Message   string      `json:"message"`
}

// NewExecutionEngine 创建任务执行引擎
func NewExecutionEngine(db *mongo.Database, redisClient *redis.Client, config ExecutionConfig) *ExecutionEngine {
	ctx, cancel := context.WithCancel(context.Background())

	engine := &ExecutionEngine{
		db:           db,
		redisClient:  redisClient,
		executors:    make(map[string]TaskExecutor),
		runningTasks: make(map[string]*ExecutionContext),
		config:       config,
		ctx:          ctx,
		cancel:       cancel,
		eventChan:    make(chan ExecutionEvent, config.ResultBufferSize),
	}

	// 启动事件处理协程
	go engine.handleEvents()

	return engine
}

// RegisterExecutor 注册任务执行器
func (e *ExecutionEngine) RegisterExecutor(taskType string, executor TaskExecutor) error {
	e.executorsMutex.Lock()
	defer e.executorsMutex.Unlock()

	if _, exists := e.executors[taskType]; exists {
		return fmt.Errorf("executor for task type %s already registered", taskType)
	}

	e.executors[taskType] = executor

	// 发送注册事件
	e.sendEvent(ExecutionEvent{
		Type:      "executor_registered",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"task_type": taskType,
			"executor":  executor.GetExecutorInfo(),
		},
		Message: fmt.Sprintf("Executor registered for task type: %s", taskType),
	})

	return nil
}

// UnregisterExecutor 注销任务执行器
func (e *ExecutionEngine) UnregisterExecutor(taskType string) error {
	e.executorsMutex.Lock()
	defer e.executorsMutex.Unlock()

	if _, exists := e.executors[taskType]; !exists {
		return fmt.Errorf("executor for task type %s not found", taskType)
	}

	delete(e.executors, taskType)

	// 发送注销事件
	e.sendEvent(ExecutionEvent{
		Type:      "executor_unregistered",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"task_type": taskType},
		Message:   fmt.Sprintf("Executor unregistered for task type: %s", taskType),
	})

	return nil
}

// ExecuteTask 执行任务
func (e *ExecutionEngine) ExecuteTask(task *models.Task) error {
	// 检查是否支持该任务类型
	e.executorsMutex.RLock()
	executor, exists := e.executors[task.Type]
	e.executorsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("no executor found for task type: %s", task.Type)
	}

	// 检查并发限制
	e.tasksMutex.RLock()
	runningCount := len(e.runningTasks)
	e.tasksMutex.RUnlock()

	if runningCount >= e.config.MaxConcurrentTasks {
		return fmt.Errorf("maximum concurrent tasks reached: %d", e.config.MaxConcurrentTasks)
	}

	// 创建执行上下文
	taskCtx, taskCancel := context.WithTimeout(e.ctx, e.config.DefaultTimeout)
	if task.Config != nil {
		if timeout, ok := task.Config["timeout"].(int64); ok && timeout > 0 {
			taskCtx, taskCancel = context.WithTimeout(e.ctx, time.Duration(timeout)*time.Second)
		}
	}

	execCtx := &ExecutionContext{
		TaskID:    task.ID.Hex(),
		Executor:  executor,
		Context:   taskCtx,
		Cancel:    taskCancel,
		StartTime: time.Now(),
		Status:    "running",
		Progress:  0.0,
	}

	// 注册运行中的任务
	e.tasksMutex.Lock()
	e.runningTasks[task.ID.Hex()] = execCtx
	e.tasksMutex.Unlock()

	// 更新任务状态为运行中
	if err := e.updateTaskStatus(task.ID, "running", 0.0); err != nil {
		// 记录错误但继续执行
		e.sendEvent(ExecutionEvent{
			Type:      "task_status_update_failed",
			TaskID:    task.ID.Hex(),
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"error": err.Error()},
			Message:   "Failed to update task status to running",
		})
	}

	// 发送任务开始事件
	e.sendEvent(ExecutionEvent{
		Type:      "task_started",
		TaskID:    task.ID.Hex(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"task_type": task.Type,
			"task_name": task.Name,
		},
		Message: "Task execution started",
	})

	// 异步执行任务
	go e.executeTaskAsync(task, execCtx)

	return nil
}

// executeTaskAsync 异步执行任务
func (e *ExecutionEngine) executeTaskAsync(task *models.Task, execCtx *ExecutionContext) {
	defer func() {
		// 清理运行中的任务
		e.tasksMutex.Lock()
		delete(e.runningTasks, task.ID.Hex())
		e.tasksMutex.Unlock()

		// 取消上下文
		execCtx.Cancel()

		// 记录执行完成
		execCtx.mutex.Lock()
		execCtx.Status = "completed"
		execCtx.mutex.Unlock()
	}()

	var result *models.TaskResult
	var err error

	// 实际执行任务
	result, err = execCtx.Executor.Execute(execCtx.Context, task)

	// 处理执行结果
	if err != nil {
		// 执行失败，检查是否需要重试
		if e.config.EnableRetry && execCtx.RetryCount < e.config.MaxRetries {
			e.scheduleRetry(task, execCtx, err)
			return
		}

		// 更新任务状态为失败
		e.updateTaskStatus(task.ID, "failed", 100.0)

		// 保存错误结果
		errorResult := &models.TaskResult{
			ID:        primitive.NewObjectID(),
			TaskID:    task.ID,
			Status:    "failed",
			Error:     err.Error(),
			StartTime: execCtx.StartTime,
			EndTime:   time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		e.saveTaskResult(errorResult)

		// 发送失败事件
		e.sendEvent(ExecutionEvent{
			Type:      "task_failed",
			TaskID:    task.ID.Hex(),
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"error":       err.Error(),
				"retry_count": execCtx.RetryCount,
			},
			Message: "Task execution failed",
		})

		return
	}

	// 执行成功
	if result != nil {
		result.EndTime = time.Now()
		result.UpdatedAt = time.Now()

		// 保存任务结果
		if saveErr := e.saveTaskResult(result); saveErr != nil {
			e.sendEvent(ExecutionEvent{
				Type:      "task_result_save_failed",
				TaskID:    task.ID.Hex(),
				Timestamp: time.Now(),
				Data:      map[string]interface{}{"error": saveErr.Error()},
				Message:   "Failed to save task result",
			})
		}
	}

	// 更新任务状态为完成
	e.updateTaskStatus(task.ID, "completed", 100.0)

	// 发送成功事件
	e.sendEvent(ExecutionEvent{
		Type:      "task_completed",
		TaskID:    task.ID.Hex(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"duration": time.Since(execCtx.StartTime).Seconds(),
		},
		Message: "Task execution completed successfully",
	})
}

// scheduleRetry 安排重试
func (e *ExecutionEngine) scheduleRetry(task *models.Task, execCtx *ExecutionContext, lastErr error) {
	execCtx.mutex.Lock()
	execCtx.RetryCount++
	retryCount := execCtx.RetryCount
	execCtx.mutex.Unlock()

	// 发送重试事件
	e.sendEvent(ExecutionEvent{
		Type:      "task_retry_scheduled",
		TaskID:    task.ID.Hex(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"retry_count": retryCount,
			"last_error":  lastErr.Error(),
			"retry_in":    e.config.RetryInterval.Seconds(),
		},
		Message: fmt.Sprintf("Task retry scheduled (attempt %d/%d)", retryCount, e.config.MaxRetries),
	})

	// 安排重试
	go func() {
		time.Sleep(e.config.RetryInterval)

		// 重新执行任务
		if err := e.ExecuteTask(task); err != nil {
			e.sendEvent(ExecutionEvent{
				Type:      "task_retry_failed",
				TaskID:    task.ID.Hex(),
				Timestamp: time.Now(),
				Data:      map[string]interface{}{"error": err.Error()},
				Message:   "Failed to schedule task retry",
			})
		}
	}()
}

// CancelTask 取消任务
func (e *ExecutionEngine) CancelTask(taskID string) error {
	e.tasksMutex.Lock()
	defer e.tasksMutex.Unlock()

	execCtx, exists := e.runningTasks[taskID]
	if !exists {
		return fmt.Errorf("task %s is not running", taskID)
	}

	// 取消执行上下文
	execCtx.Cancel()

	// 更新状态
	execCtx.mutex.Lock()
	execCtx.Status = "canceled"
	execCtx.mutex.Unlock()

	// 更新数据库中的任务状态
	taskObjID, _ := primitive.ObjectIDFromHex(taskID)
	e.updateTaskStatus(taskObjID, "canceled", execCtx.Progress)

	// 发送取消事件
	e.sendEvent(ExecutionEvent{
		Type:      "task_canceled",
		TaskID:    taskID,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"progress": execCtx.Progress},
		Message:   "Task execution canceled",
	})

	return nil
}

// GetRunningTasks 获取运行中的任务
func (e *ExecutionEngine) GetRunningTasks() []string {
	e.tasksMutex.RLock()
	defer e.tasksMutex.RUnlock()

	taskIDs := make([]string, 0, len(e.runningTasks))
	for taskID := range e.runningTasks {
		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs
}

// GetTaskStatus 获取任务状态
func (e *ExecutionEngine) GetTaskStatus(taskID string) (*ExecutionContext, error) {
	e.tasksMutex.RLock()
	defer e.tasksMutex.RUnlock()

	execCtx, exists := e.runningTasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s is not running", taskID)
	}

	return execCtx, nil
}

// GetExecutors 获取已注册的执行器
func (e *ExecutionEngine) GetExecutors() map[string]models.ExecutorInfo {
	e.executorsMutex.RLock()
	defer e.executorsMutex.RUnlock()

	executors := make(map[string]models.ExecutorInfo)
	for taskType, executor := range e.executors {
		executors[taskType] = executor.GetExecutorInfo()
	}

	return executors
}

// Shutdown 关闭执行引擎
func (e *ExecutionEngine) Shutdown() {
	// 取消所有运行中的任务
	e.tasksMutex.Lock()
	for _, execCtx := range e.runningTasks {
		execCtx.Cancel()
	}
	e.tasksMutex.Unlock()

	// 关闭事件通道
	close(e.eventChan)

	// 取消上下文
	e.cancel()
}

// updateTaskStatus 更新任务状态
func (e *ExecutionEngine) updateTaskStatus(taskID primitive.ObjectID, status string, progress float64) error {
	collection := e.db.Collection("tasks")

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"progress":   progress,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(e.ctx, bson.M{"_id": taskID}, update)
	return err
}

// saveTaskResult 保存任务结果
func (e *ExecutionEngine) saveTaskResult(result *models.TaskResult) error {
	collection := e.db.Collection("task_results")

	if result.ID.IsZero() {
		result.ID = primitive.NewObjectID()
	}

	if result.CreatedAt.IsZero() {
		result.CreatedAt = time.Now()
	}

	result.UpdatedAt = time.Now()

	_, err := collection.InsertOne(e.ctx, result)
	return err
}

// sendEvent 发送事件
func (e *ExecutionEngine) sendEvent(event ExecutionEvent) {
	select {
	case e.eventChan <- event:
	default:
		// 如果事件通道满了，丢弃事件避免阻塞
	}
}

// handleEvents 处理事件
func (e *ExecutionEngine) handleEvents() {
	for {
		select {
		case event, ok := <-e.eventChan:
			if !ok {
				return
			}

			// 这里可以实现事件的具体处理逻辑
			// 比如写入日志、发送通知、更新监控指标等
			e.processEvent(event)

		case <-e.ctx.Done():
			return
		}
	}
}

// processEvent 处理具体事件
func (e *ExecutionEngine) processEvent(event ExecutionEvent) {
	// 保存事件到Redis（可选）
	if e.redisClient != nil {
		eventKey := fmt.Sprintf("task_events:%s", event.TaskID)
		eventData := map[string]interface{}{
			"type":      event.Type,
			"timestamp": event.Timestamp.Unix(),
			"data":      event.Data,
			"message":   event.Message,
		}

		// 使用Redis的列表存储事件历史
		e.redisClient.LPush(e.ctx, eventKey, eventData)
		// 限制事件历史长度
		e.redisClient.LTrim(e.ctx, eventKey, 0, 99) // 保留最近100个事件
		// 设置过期时间
		e.redisClient.Expire(e.ctx, eventKey, 24*time.Hour)
	}
}
