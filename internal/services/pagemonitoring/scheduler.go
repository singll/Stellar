package pagemonitoring

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MonitoringScheduler 监控调度器
type MonitoringScheduler struct {
	db            *mongo.Database
	crawler       *PageCrawler
	notifier      *NotificationService
	running       bool
	stopChan      chan struct{}
	wg            sync.WaitGroup
	mutex         sync.RWMutex
	activeTasks   map[primitive.ObjectID]*ScheduledTask
	maxWorkers    int
	workerPool    chan struct{}
	taskQueue     chan *models.MonitoringTask
}

// ScheduledTask 调度任务
type ScheduledTask struct {
	Task        *models.MonitoringTask
	NextRun     time.Time
	LastRun     time.Time
	RunCount    int64
	FailCount   int64
	Status      TaskStatus
	Worker      *TaskWorker
	Context     context.Context
	CancelFunc  context.CancelFunc
}

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusIdle    TaskStatus = "idle"    // 空闲
	TaskStatusRunning TaskStatus = "running" // 运行中
	TaskStatusPaused  TaskStatus = "paused"  // 暂停
	TaskStatusFailed  TaskStatus = "failed"  // 失败
	TaskStatusStopped TaskStatus = "stopped" // 已停止
)

// TaskWorker 任务工作器
type TaskWorker struct {
	ID        int
	Task      *models.MonitoringTask
	Scheduler *MonitoringScheduler
	Context   context.Context
	CancelFunc context.CancelFunc
}

// NewMonitoringScheduler 创建监控调度器
func NewMonitoringScheduler(db *mongo.Database, maxWorkers int) *MonitoringScheduler {
	if maxWorkers <= 0 {
		maxWorkers = 10
	}

	return &MonitoringScheduler{
		db:          db,
		crawler:     NewPageCrawler(30),
		notifier:    NewNotificationService(),
		stopChan:    make(chan struct{}),
		activeTasks: make(map[primitive.ObjectID]*ScheduledTask),
		maxWorkers:  maxWorkers,
		workerPool:  make(chan struct{}, maxWorkers),
		taskQueue:   make(chan *models.MonitoringTask, 100),
	}
}

// Start 启动调度器
func (s *MonitoringScheduler) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return fmt.Errorf("调度器已在运行")
	}

	s.running = true
	log.Println("启动页面监控调度器...")

	// 加载现有的监控任务
	if err := s.loadTasks(); err != nil {
		return fmt.Errorf("加载监控任务失败: %v", err)
	}

	// 启动主调度循环
	s.wg.Add(1)
	go s.scheduleLoop()

	// 启动任务队列处理器
	s.wg.Add(1)
	go s.taskQueueProcessor()

	// 启动工作器池
	for i := 0; i < s.maxWorkers; i++ {
		s.wg.Add(1)
		go s.workerLoop(i)
	}

	log.Printf("页面监控调度器已启动，最大工作器数: %d", s.maxWorkers)
	return nil
}

// Stop 停止调度器
func (s *MonitoringScheduler) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return nil
	}

	log.Println("停止页面监控调度器...")
	s.running = false

	// 取消所有活动任务
	for _, task := range s.activeTasks {
		if task.CancelFunc != nil {
			task.CancelFunc()
		}
	}

	// 发送停止信号
	close(s.stopChan)

	// 等待所有goroutine结束
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	// 设置超时
	select {
	case <-done:
		log.Println("页面监控调度器已停止")
	case <-time.After(30 * time.Second):
		log.Println("页面监控调度器停止超时")
	}

	return nil
}

// loadTasks 加载监控任务
func (s *MonitoringScheduler) loadTasks() error {
	collection := s.db.Collection("monitoring_tasks")
	
	// 查询所有启用的监控任务
	filter := bson.M{
		"enabled": true,
		"$or": []bson.M{
			{"status": models.MonitoringTaskStatusActive},
			{"status": models.MonitoringTaskStatusPaused},
		},
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	var tasks []*models.MonitoringTask
	if err := cursor.All(context.Background(), &tasks); err != nil {
		return err
	}

	log.Printf("加载了 %d 个监控任务", len(tasks))

	// 初始化调度任务
	for _, task := range tasks {
		scheduledTask := &ScheduledTask{
			Task:     task,
			NextRun:  s.calculateNextRun(task),
			Status:   TaskStatusIdle,
			RunCount: 0,
		}

		if task.Status == models.MonitoringTaskStatusPaused {
			scheduledTask.Status = TaskStatusPaused
		}

		s.activeTasks[task.ID] = scheduledTask
	}

	return nil
}

// calculateNextRun 计算下次运行时间
func (s *MonitoringScheduler) calculateNextRun(task *models.MonitoringTask) time.Time {
	now := time.Now()
	
	// 如果任务从未运行过，立即运行
	if task.LastRunAt.IsZero() {
		return now
	}

	// 根据监控间隔计算下次运行时间
	interval := task.Interval
	nextRun := task.LastRunAt.Add(interval)

	// 如果下次运行时间已过，立即运行
	if nextRun.Before(now) {
		return now
	}

	return nextRun
}

// scheduleLoop 调度循环
func (s *MonitoringScheduler) scheduleLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(10 * time.Second) // 每10秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAndScheduleTasks()
		}
	}
}

// checkAndScheduleTasks 检查并调度任务
func (s *MonitoringScheduler) checkAndScheduleTasks() {
	s.mutex.RLock()
	tasks := make([]*ScheduledTask, 0, len(s.activeTasks))
	for _, task := range s.activeTasks {
		tasks = append(tasks, task)
	}
	s.mutex.RUnlock()

	now := time.Now()
	for _, scheduledTask := range tasks {
		// 跳过暂停或运行中的任务
		if scheduledTask.Status == TaskStatusPaused || 
		   scheduledTask.Status == TaskStatusRunning {
			continue
		}

		// 检查是否到了运行时间
		if now.After(scheduledTask.NextRun) || now.Equal(scheduledTask.NextRun) {
			select {
			case s.taskQueue <- scheduledTask.Task:
				log.Printf("任务 %s 已添加到执行队列", scheduledTask.Task.Name)
				s.mutex.Lock()
				scheduledTask.Status = TaskStatusRunning
				scheduledTask.LastRun = now
				s.mutex.Unlock()
			default:
				log.Printf("任务队列已满，跳过任务 %s", scheduledTask.Task.Name)
			}
		}
	}
}

// taskQueueProcessor 任务队列处理器
func (s *MonitoringScheduler) taskQueueProcessor() {
	defer s.wg.Done()

	for {
		select {
		case <-s.stopChan:
			return
		case task := <-s.taskQueue:
			// 获取工作器
			select {
			case s.workerPool <- struct{}{}:
				// 创建工作器
				ctx, cancel := context.WithTimeout(context.Background(), 
					time.Duration(task.Config.Timeout)*time.Second)
				
				worker := &TaskWorker{
					Task:       task,
					Scheduler:  s,
					Context:    ctx,
					CancelFunc: cancel,
				}

				// 异步执行任务
				go func() {
					defer func() {
						<-s.workerPool // 释放工作器
						cancel()
					}()
					worker.Execute()
				}()
			case <-s.stopChan:
				return
			}
		}
	}
}

// workerLoop 工作器循环
func (s *MonitoringScheduler) workerLoop(workerID int) {
	defer s.wg.Done()
	
	log.Printf("工作器 %d 已启动", workerID)
	
	for {
		select {
		case <-s.stopChan:
			log.Printf("工作器 %d 已停止", workerID)
			return
		default:
			// 工作器在这里等待任务分配
			time.Sleep(1 * time.Second)
		}
	}
}

// Execute 执行监控任务
func (w *TaskWorker) Execute() {
	startTime := time.Now()
	task := w.Task
	
	log.Printf("开始执行监控任务: %s (URL: %s)", task.Name, task.URL)

	// 更新任务状态
	w.updateTaskStatus(models.MonitoringTaskStatusRunning, "")

	// 抓取页面
	snapshot, err := w.Scheduler.crawler.FetchPage(task.URL, task.Config)
	if err != nil {
		w.handleError(fmt.Errorf("抓取页面失败: %v", err))
		return
	}

	// 保存快照
	if err := w.saveSnapshot(snapshot); err != nil {
		w.handleError(fmt.Errorf("保存快照失败: %v", err))
		return
	}

	// 获取上一次快照进行比较
	lastSnapshot, err := w.getLastSnapshot(task.URL)
	if err != nil {
		log.Printf("获取上次快照失败: %v", err)
		// 如果是第一次运行，这是正常的
	}

	var changeDetected bool
	var similarity float64
	var diff string

	// 如果有上次快照，进行比较
	if lastSnapshot != nil {
		change, sim, d := CompareSnapshots(lastSnapshot, snapshot, task.Config)
		similarity = sim
		diff = d

		// 保存变更记录
		if change.Status == models.PageChangeStatusChanged {
			changeDetected = true
			if err := w.savePageChange(change); err != nil {
				log.Printf("保存变更记录失败: %v", err)
			}
		}
	}

	// 发送通知
	if changeDetected {
		w.sendChangeNotification(snapshot, similarity, diff)
	}

	// 更新任务状态和统计
	executionTime := time.Since(startTime)
	w.updateTaskCompletion(executionTime, changeDetected, similarity)

	log.Printf("监控任务 %s 执行完成，耗时: %v", task.Name, executionTime)
}

// saveSnapshot 保存快照
func (w *TaskWorker) saveSnapshot(snapshot *models.PageSnapshot) error {
	collection := w.Scheduler.db.Collection("page_snapshots")
	
	snapshot.TaskID = w.Task.ID
	_, err := collection.InsertOne(w.Context, snapshot)
	return err
}

// getLastSnapshot 获取最后一次快照
func (w *TaskWorker) getLastSnapshot(url string) (*models.PageSnapshot, error) {
	collection := w.Scheduler.db.Collection("page_snapshots")
	
	filter := bson.M{
		"url":     url,
		"task_id": w.Task.ID,
	}
	
	var snapshot models.PageSnapshot
	err := collection.FindOne(w.Context, filter, 
		&options.FindOneOptions{
			Sort: bson.M{"created_at": -1},
		}).Decode(&snapshot)
	
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	
	return &snapshot, err
}

// savePageChange 保存页面变更
func (w *TaskWorker) savePageChange(change *models.PageChange) error {
	collection := w.Scheduler.db.Collection("page_changes")
	
	change.TaskID = w.Task.ID
	_, err := collection.InsertOne(w.Context, change)
	return err
}

// sendChangeNotification 发送变更通知
func (w *TaskWorker) sendChangeNotification(snapshot *models.PageSnapshot, similarity float64, diff string) {
	notification := &NotificationMessage{
		Type:      NotificationTypePageChange,
		Title:     fmt.Sprintf("页面变更检测: %s", w.Task.Name),
		Message:   fmt.Sprintf("检测到页面 %s 发生变更", w.Task.URL),
		Level:     NotificationLevelInfo,
		Data: map[string]interface{}{
			"task_id":    w.Task.ID,
			"task_name":  w.Task.Name,
			"url":        w.Task.URL,
			"similarity": similarity,
			"diff":       diff,
			"timestamp":  snapshot.CreatedAt,
		},
		CreatedAt: time.Now(),
	}

	// 根据相似度调整通知级别
	if similarity < 0.3 {
		notification.Level = NotificationLevelCritical
		notification.Title = fmt.Sprintf("重大页面变更: %s", w.Task.Name)
	} else if similarity < 0.7 {
		notification.Level = NotificationLevelWarning
		notification.Title = fmt.Sprintf("页面变更警告: %s", w.Task.Name)
	}

	if err := w.Scheduler.notifier.Send(notification); err != nil {
		log.Printf("发送通知失败: %v", err)
	}
}

// updateTaskStatus 更新任务状态
func (w *TaskWorker) updateTaskStatus(status models.MonitoringTaskStatus, message string) {
	collection := w.Scheduler.db.Collection("monitoring_tasks")
	
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"message":    message,
			"updated_at": time.Now(),
		},
	}
	
	_, err := collection.UpdateOne(w.Context, bson.M{"_id": w.Task.ID}, update)
	if err != nil {
		log.Printf("更新任务状态失败: %v", err)
	}
}

// updateTaskCompletion 更新任务完成状态
func (w *TaskWorker) updateTaskCompletion(executionTime time.Duration, changeDetected bool, similarity float64) {
	collection := w.Scheduler.db.Collection("monitoring_tasks")
	
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":           models.MonitoringTaskStatusActive,
			"last_run_at":      &now,
			"last_execution_time": int(executionTime.Milliseconds()),
			"message":          "执行成功",
			"updated_at":       now,
		},
		"$inc": bson.M{
			"run_count": 1,
		},
	}

	if changeDetected {
		update["$inc"].(bson.M)["change_count"] = 1
		update["$set"].(bson.M)["last_change_at"] = &now
		update["$set"].(bson.M)["last_similarity"] = similarity
	}

	_, err := collection.UpdateOne(w.Context, bson.M{"_id": w.Task.ID}, update)
	if err != nil {
		log.Printf("更新任务完成状态失败: %v", err)
	}

	// 更新调度任务的下次运行时间
	w.Scheduler.mutex.Lock()
	if scheduledTask, exists := w.Scheduler.activeTasks[w.Task.ID]; exists {
		scheduledTask.Status = TaskStatusIdle
		scheduledTask.RunCount++
		scheduledTask.NextRun = w.Scheduler.calculateNextRun(w.Task)
	}
	w.Scheduler.mutex.Unlock()
}

// handleError 处理错误
func (w *TaskWorker) handleError(err error) {
	log.Printf("监控任务 %s 执行失败: %v", w.Task.Name, err)

	// 更新任务状态
	w.updateTaskStatus(models.MonitoringTaskStatusFailed, err.Error())

	// 更新调度任务状态
	w.Scheduler.mutex.Lock()
	if scheduledTask, exists := w.Scheduler.activeTasks[w.Task.ID]; exists {
		scheduledTask.Status = TaskStatusFailed
		scheduledTask.FailCount++
		// 失败后延迟重试
		scheduledTask.NextRun = time.Now().Add(w.Task.Interval * 2)
	}
	w.Scheduler.mutex.Unlock()

	// 发送错误通知
	notification := &NotificationMessage{
		Type:    NotificationTypeTaskError,
		Title:   fmt.Sprintf("监控任务失败: %s", w.Task.Name),
		Message: fmt.Sprintf("任务 %s 执行失败: %v", w.Task.Name, err),
		Level:   NotificationLevelError,
		Data: map[string]interface{}{
			"task_id":   w.Task.ID,
			"task_name": w.Task.Name,
			"url":       w.Task.URL,
			"error":     err.Error(),
		},
		CreatedAt: time.Now(),
	}

	if err := w.Scheduler.notifier.Send(notification); err != nil {
		log.Printf("发送错误通知失败: %v", err)
	}
}

// AddTask 添加监控任务
func (s *MonitoringScheduler) AddTask(task *models.MonitoringTask) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 创建调度任务
	scheduledTask := &ScheduledTask{
		Task:    task,
		NextRun: s.calculateNextRun(task),
		Status:  TaskStatusIdle,
	}

	if task.Status == models.MonitoringTaskStatusPaused {
		scheduledTask.Status = TaskStatusPaused
	}

	s.activeTasks[task.ID] = scheduledTask
	log.Printf("已添加监控任务: %s", task.Name)

	return nil
}

// RemoveTask 移除监控任务
func (s *MonitoringScheduler) RemoveTask(taskID primitive.ObjectID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if scheduledTask, exists := s.activeTasks[taskID]; exists {
		// 取消正在运行的任务
		if scheduledTask.CancelFunc != nil {
			scheduledTask.CancelFunc()
		}
		
		delete(s.activeTasks, taskID)
		log.Printf("已移除监控任务: %s", taskID.Hex())
	}

	return nil
}

// PauseTask 暂停监控任务
func (s *MonitoringScheduler) PauseTask(taskID primitive.ObjectID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if scheduledTask, exists := s.activeTasks[taskID]; exists {
		scheduledTask.Status = TaskStatusPaused
		log.Printf("已暂停监控任务: %s", taskID.Hex())
	}

	return nil
}

// ResumeTask 恢复监控任务
func (s *MonitoringScheduler) ResumeTask(taskID primitive.ObjectID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if scheduledTask, exists := s.activeTasks[taskID]; exists {
		scheduledTask.Status = TaskStatusIdle
		scheduledTask.NextRun = time.Now() // 立即运行
		log.Printf("已恢复监控任务: %s", taskID.Hex())
	}

	return nil
}

// GetTaskStatus 获取任务状态
func (s *MonitoringScheduler) GetTaskStatus(taskID primitive.ObjectID) (*ScheduledTask, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	task, exists := s.activeTasks[taskID]
	return task, exists
}

// GetAllTaskStatus 获取所有任务状态
func (s *MonitoringScheduler) GetAllTaskStatus() map[primitive.ObjectID]*ScheduledTask {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make(map[primitive.ObjectID]*ScheduledTask)
	for id, task := range s.activeTasks {
		result[id] = task
	}

	return result
}

// GetStatistics 获取调度器统计信息
func (s *MonitoringScheduler) GetStatistics() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stats := map[string]interface{}{
		"running":     s.running,
		"total_tasks": len(s.activeTasks),
		"max_workers": s.maxWorkers,
	}

	// 按状态统计任务
	statusCount := make(map[TaskStatus]int)
	for _, task := range s.activeTasks {
		statusCount[task.Status]++
	}
	stats["task_status"] = statusCount

	return stats
}