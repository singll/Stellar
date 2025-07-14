package portscan

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TaskManager 端口扫描任务管理器
type TaskManager struct {
	db            *mongo.Database
	activeTasks   map[string]*TaskRunner
	taskMutex     sync.RWMutex
	resultHandler ResultHandler
}

// TaskRunner 任务运行器
type TaskRunner struct {
	Task       *models.PortScanTask
	Scanner    *PortScanner
	Context    context.Context
	CancelFunc context.CancelFunc
	Done       chan struct{}
}

// ResultHandler 结果处理器接口
type ResultHandler interface {
	HandleResult(result *models.PortScanResult) error
	SaveTask(task *models.PortScanTask) error
	UpdateTaskProgress(taskID string, progress float64) error
	UpdateTaskStatus(taskID string, status string) error
	FinishTask(task *models.PortScanTask) error
}

// DefaultResultHandler 默认结果处理器实现
type DefaultResultHandler struct {
	db *mongo.Database
}

// NewDefaultResultHandler 创建默认结果处理器
func NewDefaultResultHandler(db *mongo.Database) *DefaultResultHandler {
	return &DefaultResultHandler{
		db: db,
	}
}

// HandleResult 处理扫描结果
func (h *DefaultResultHandler) HandleResult(result *models.PortScanResult) error {
	_, err := h.db.Collection("port_scan_results").InsertOne(context.Background(), result)
	return err
}

// SaveTask 保存任务
func (h *DefaultResultHandler) SaveTask(task *models.PortScanTask) error {
	_, err := h.db.Collection("port_scan_tasks").InsertOne(context.Background(), task)
	return err
}

// UpdateTaskProgress 更新任务进度
func (h *DefaultResultHandler) UpdateTaskProgress(taskID string, progress float64) error {
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return fmt.Errorf("无效的任务ID: %v", err)
	}

	_, err = h.db.Collection("port_scan_tasks").UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{
			"progress":  progress,
			"updatedAt": time.Now(),
		}},
	)
	return err
}

// UpdateTaskStatus 更新任务状态
func (h *DefaultResultHandler) UpdateTaskStatus(taskID string, status string) error {
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return fmt.Errorf("无效的任务ID: %v", err)
	}

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	if status == "running" {
		update["$set"].(bson.M)["startTime"] = time.Now()
	} else if status == "completed" || status == "failed" || status == "stopped" {
		update["$set"].(bson.M)["completedAt"] = time.Now()
	}

	_, err = h.db.Collection("port_scan_tasks").UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		update,
	)
	return err
}

// FinishTask 完成任务
func (h *DefaultResultHandler) FinishTask(task *models.PortScanTask) error {
	return h.UpdateTaskStatus(task.ID.Hex(), "completed")
}

// NewTaskManager 创建任务管理器
func NewTaskManager(db *mongo.Database, resultHandler ResultHandler) *TaskManager {
	if resultHandler == nil {
		resultHandler = NewDefaultResultHandler(db)
	}
	return &TaskManager{
		db:            db,
		activeTasks:   make(map[string]*TaskRunner),
		resultHandler: resultHandler,
	}
}

// CreateTask 创建新任务
func (m *TaskManager) CreateTask(task *models.PortScanTask) (string, error) {
	// 设置默认值
	if task.ID.IsZero() {
		task.ID = primitive.NewObjectID()
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	task.Status = "pending"
	task.Progress = 0

	// 保存任务到数据库
	err := m.resultHandler.SaveTask(task)
	if err != nil {
		return "", fmt.Errorf("保存任务失败: %v", err)
	}

	return task.ID.Hex(), nil
}

// StartTask 启动任务
func (m *TaskManager) StartTask(taskID string) error {
	// 检查任务是否已经在运行
	m.taskMutex.RLock()
	_, exists := m.activeTasks[taskID]
	m.taskMutex.RUnlock()
	if exists {
		return fmt.Errorf("任务已在运行中")
	}

	// 从数据库获取任务
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return fmt.Errorf("无效的任务ID: %v", err)
	}

	var task models.PortScanTask
	err = m.db.Collection("port_scan_tasks").FindOne(context.Background(), bson.M{"_id": objID}).Decode(&task)
	if err != nil {
		return fmt.Errorf("获取任务失败: %v", err)
	}

	// 检查任务状态
	if task.Status == "running" || task.Status == "completed" {
		return fmt.Errorf("任务状态不允许启动: %s", task.Status)
	}

	// 更新任务状态
	task.Status = "running"
	task.StartTime = time.Now()
	err = m.resultHandler.UpdateTaskStatus(taskID, "running")
	if err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	// 创建扫描器
	scanner := NewPortScanner(task.Config, taskID, task.ProjectID.Hex())

	// 创建上下文和取消函数
	ctx, cancel := context.WithCancel(context.Background())

	// 创建任务运行器
	runner := &TaskRunner{
		Task:       &task,
		Scanner:    scanner,
		Context:    ctx,
		CancelFunc: cancel,
		Done:       make(chan struct{}),
	}

	// 添加到活动任务列表
	m.taskMutex.Lock()
	m.activeTasks[taskID] = runner
	m.taskMutex.Unlock()

	// 启动任务
	go m.runTask(runner)

	return nil
}

// StopTask 停止任务
func (m *TaskManager) StopTask(taskID string) error {
	// 检查任务是否在运行
	m.taskMutex.RLock()
	runner, exists := m.activeTasks[taskID]
	m.taskMutex.RUnlock()
	if !exists {
		return fmt.Errorf("任务未在运行中")
	}

	// 取消任务
	runner.CancelFunc()

	// 等待任务结束
	<-runner.Done

	// 从活动任务列表中移除
	m.taskMutex.Lock()
	delete(m.activeTasks, taskID)
	m.taskMutex.Unlock()

	// 更新任务状态
	err := m.resultHandler.UpdateTaskStatus(taskID, "stopped")
	if err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	return nil
}

// GetTaskStatus 获取任务状态
func (m *TaskManager) GetTaskStatus(taskID string) (string, error) {
	// 检查任务是否在运行
	m.taskMutex.RLock()
	runner, exists := m.activeTasks[taskID]
	m.taskMutex.RUnlock()
	if exists {
		return string(runner.Task.Status), nil
	}

	// 从数据库获取任务
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return "", fmt.Errorf("无效的任务ID: %v", err)
	}

	var task models.PortScanTask
	err = m.db.Collection("port_scan_tasks").FindOne(context.Background(), bson.M{"_id": objID}).Decode(&task)
	if err != nil {
		return "", fmt.Errorf("获取任务失败: %v", err)
	}

	return string(task.Status), nil
}

// GetTaskProgress 获取任务进度
func (m *TaskManager) GetTaskProgress(taskID string) (float64, error) {
	// 检查任务是否在运行
	m.taskMutex.RLock()
	runner, exists := m.activeTasks[taskID]
	m.taskMutex.RUnlock()
	if exists {
		return runner.Task.Progress, nil
	}

	// 从数据库获取任务
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return 0, fmt.Errorf("无效的任务ID: %v", err)
	}

	var task models.PortScanTask
	err = m.db.Collection("port_scan_tasks").FindOne(context.Background(), bson.M{"_id": objID}).Decode(&task)
	if err != nil {
		return 0, fmt.Errorf("获取任务失败: %v", err)
	}

	return task.Progress, nil
}

// GetTaskResults 获取任务结果
func (m *TaskManager) GetTaskResults(taskID string, limit, skip int) ([]*models.PortScanResult, error) {
	// 从数据库获取结果
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, fmt.Errorf("无效的任务ID: %v", err)
	}

	// 设置默认值
	if limit <= 0 {
		limit = 100
	}

	// 转换为int64类型
	limit64 := int64(limit)
	skip64 := int64(skip)

	// 查询结果
	cursor, err := m.db.Collection("port_scan_results").Find(
		context.Background(),
		bson.M{"taskId": objID},
		&options.FindOptions{
			Limit: &limit64,
			Skip:  &skip64,
			Sort:  bson.D{{"createdAt", -1}},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("查询结果失败: %v", err)
	}
	defer cursor.Close(context.Background())

	// 解析结果
	var results []*models.PortScanResult
	err = cursor.All(context.Background(), &results)
	if err != nil {
		return nil, fmt.Errorf("解析结果失败: %v", err)
	}

	return results, nil
}

// ListTasks 列出任务
func (m *TaskManager) ListTasks(query map[string]interface{}, limit, skip int) ([]*models.PortScanTask, error) {
	// 设置默认值
	if limit <= 0 {
		limit = 10
	}

	// 转换为int64类型
	limit64 := int64(limit)
	skip64 := int64(skip)

	// 查询任务
	cursor, err := m.db.Collection("port_scan_tasks").Find(
		context.Background(),
		query,
		&options.FindOptions{
			Limit: &limit64,
			Skip:  &skip64,
			Sort:  bson.D{{"createdAt", -1}},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %v", err)
	}
	defer cursor.Close(context.Background())

	// 解析结果
	var tasks []*models.PortScanTask
	err = cursor.All(context.Background(), &tasks)
	if err != nil {
		return nil, fmt.Errorf("解析任务失败: %v", err)
	}

	return tasks, nil
}

// GetTask 获取任务详情
func (m *TaskManager) GetTask(taskID string) (*models.PortScanTask, error) {
	// 检查任务是否在运行
	m.taskMutex.RLock()
	runner, exists := m.activeTasks[taskID]
	m.taskMutex.RUnlock()
	if exists {
		return runner.Task, nil
	}

	// 从数据库获取任务
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, fmt.Errorf("无效的任务ID: %v", err)
	}

	var task models.PortScanTask
	err = m.db.Collection("port_scan_tasks").FindOne(context.Background(), bson.M{"_id": objID}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // 任务不存在
		}
		return nil, fmt.Errorf("获取任务失败: %v", err)
	}

	return &task, nil
}

// runTask 执行任务
func (m *TaskManager) runTask(runner *TaskRunner) {
	defer close(runner.Done)
	defer runner.CancelFunc()

	// 获取任务信息
	task := runner.Task
	scanner := runner.Scanner
	ctx := runner.Context

	// 启动结果处理协程
	go m.handleResults(ctx, scanner, task)

	// 启动进度更新协程
	go func() {
		// 监控进度并更新
		for {
			select {
			case <-ctx.Done():
				return
			case progress := <-scanner.ProgressChan:
				// 更新任务进度
				m.resultHandler.UpdateTaskProgress(task.ID.Hex(), progress)
				task.Progress = progress
			}
		}
	}()

	// 开始扫描
	err := scanner.Start(ctx, task.Targets)
	if err != nil {
		// 更新任务状态为失败
		task.Status = "failed"
		task.Error = err.Error()
		m.resultHandler.UpdateTaskStatus(task.ID.Hex(), "failed")
		return
	}

	// 检查是否被取消
	select {
	case <-ctx.Done():
		// 任务被取消
		task.Status = "stopped"
		m.resultHandler.UpdateTaskStatus(task.ID.Hex(), "stopped")
	default:
		// 任务完成
		task.Status = "completed"
		task.CompletedAt = time.Now()
		task.Progress = 100
		m.resultHandler.FinishTask(task)
	}

	// 从活动任务列表中移除
	m.taskMutex.Lock()
	delete(m.activeTasks, task.ID.Hex())
	m.taskMutex.Unlock()
}

// handleResults 处理扫描结果
func (m *TaskManager) handleResults(ctx context.Context, scanner *PortScanner, task *models.PortScanTask) {
	for {
		select {
		case <-ctx.Done():
			return
		case result, ok := <-scanner.ResultChan:
			if !ok {
				// 通道已关闭
				return
			}
			// 处理结果
			m.resultHandler.HandleResult(result)
		}
	}
}
