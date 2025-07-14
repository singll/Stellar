package taskmanager

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TaskDispatcher 任务分发器
type TaskDispatcher struct {
	db           *mongo.Database
	redisClient  *redis.Client
	nodeManager  NodeManager
	queueManager *QueueManager
	running      bool
	workerCount  int
	taskChan     chan *models.Task
	ctx          context.Context
	cancel       context.CancelFunc
	mutex        sync.RWMutex
}

// NewTaskDispatcher 创建任务分发器
func NewTaskDispatcher(db *mongo.Database, redisClient *redis.Client, nodeManager NodeManager, queueManager *QueueManager, workerCount int) *TaskDispatcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskDispatcher{
		db:           db,
		redisClient:  redisClient,
		nodeManager:  nodeManager,
		queueManager: queueManager,
		workerCount:  workerCount,
		taskChan:     make(chan *models.Task, workerCount*2),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start 启动任务分发器
func (td *TaskDispatcher) Start() error {
	td.mutex.Lock()
	defer td.mutex.Unlock()

	if td.running {
		return errors.New("任务分发器已经在运行")
	}

	// 启动工作线程
	for i := 0; i < td.workerCount; i++ {
		go td.worker()
	}

	// 启动任务调度器
	go td.scheduler()

	td.running = true
	return nil
}

// Stop 停止任务分发器
func (td *TaskDispatcher) Stop() {
	td.mutex.Lock()
	defer td.mutex.Unlock()

	if !td.running {
		return
	}

	td.cancel()
	close(td.taskChan)
	td.running = false
}

// SubmitTask 提交任务
func (td *TaskDispatcher) SubmitTask(task *models.Task) error {
	// 验证任务
	if task.ID.IsZero() {
		task.ID = primitive.NewObjectID()
	}
	if task.Status == "" {
		task.Status = string(models.TaskStatusPending)
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	if task.Priority == 0 {
		task.Priority = models.TaskPriorityNormal
	}

	// 保存任务到数据库
	_, err := td.db.Collection("tasks").InsertOne(td.ctx, task)
	if err != nil {
		return err
	}

	// 将任务加入队列
	queueName := "default"
	switch task.Type {
	case models.TaskTypeSubdomainEnum:
		queueName = "subdomain"
	case models.TaskTypePortScan:
		queueName = "portscan"
	case models.TaskTypeVulnScan:
		queueName = "vulnscan"
	case models.TaskTypeAssetDiscovery:
		queueName = "discovery"
	}

	// 检查队列是否存在，不存在则创建
	_, err = td.queueManager.GetQueue(queueName)
	if err != nil {
		_, err = td.queueManager.CreateQueue(queueName, task.Type, task.Priority, 1000)
		if err != nil {
			return err
		}
	}

	// 将任务加入队列
	return td.queueManager.EnqueueTask(queueName, task)
}

// CancelTask 取消任务
func (td *TaskDispatcher) CancelTask(taskID string) error {
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return err
	}

	// 更新任务状态
	_, err = td.db.Collection("tasks").UpdateOne(
		td.ctx,
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"status": models.TaskStatusCancelled,
			},
		},
	)
	if err != nil {
		return err
	}

	// 发送取消信号到节点
	var task models.Task
	err = td.db.Collection("tasks").FindOne(td.ctx, bson.M{"_id": objID}).Decode(&task)
	if err != nil {
		return err
	}

	if task.NodeID != "" {
		// 通过Redis发送取消信号
		err = td.redisClient.Publish(td.ctx, "task_cancel:"+task.NodeID, taskID).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetTaskStatus 获取任务状态
func (td *TaskDispatcher) GetTaskStatus(taskID string) (string, error) {
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return "", err
	}

	var task models.Task
	err = td.db.Collection("tasks").FindOne(td.ctx, bson.M{"_id": objID}).Decode(&task)
	if err != nil {
		return "", err
	}

	return task.Status, nil
}

// GetTaskProgress 获取任务进度
func (td *TaskDispatcher) GetTaskProgress(taskID string) (float64, error) {
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return 0, err
	}

	var task models.Task
	err = td.db.Collection("tasks").FindOne(td.ctx, bson.M{"_id": objID}).Decode(&task)
	if err != nil {
		return 0, err
	}

	return task.Progress, nil
}

// GetTaskResult 获取任务结果
func (td *TaskDispatcher) GetTaskResult(taskID string) (*models.TaskResult, error) {
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, err
	}

	var task models.Task
	err = td.db.Collection("tasks").FindOne(td.ctx, bson.M{"_id": objID}).Decode(&task)
	if err != nil {
		return nil, err
	}

	if task.ResultID.IsZero() {
		return nil, errors.New("任务结果不存在")
	}

	var result models.TaskResult
	err = td.db.Collection("task_results").FindOne(td.ctx, bson.M{"_id": task.ResultID}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateTaskStatus 更新任务状态
func (td *TaskDispatcher) UpdateTaskStatus(taskID string, status string, progress float64) error {
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":   status,
			"progress": progress,
		},
	}

	if status == string(models.TaskStatusRunning) && progress == 0 {
		update["$set"].(bson.M)["startedAt"] = time.Now()
	} else if status == string(models.TaskStatusCompleted) || status == string(models.TaskStatusFailed) {
		update["$set"].(bson.M)["completedAt"] = time.Now()
	}

	_, err = td.db.Collection("tasks").UpdateOne(td.ctx, bson.M{"_id": objID}, update)
	return err
}

// SaveTaskResult 保存任务结果
func (td *TaskDispatcher) SaveTaskResult(taskID string, result *models.TaskResult) error {
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return err
	}

	// 保存结果
	if result.ID.IsZero() {
		result.ID = primitive.NewObjectID()
	}
	result.TaskID = objID
	result.CreatedAt = time.Now()
	result.EndTime = time.Now()
	result.UpdatedAt = time.Now()

	_, err = td.db.Collection("task_results").InsertOne(td.ctx, result)
	if err != nil {
		return err
	}

	// 更新任务的结果ID
	_, err = td.db.Collection("tasks").UpdateOne(
		td.ctx,
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"resultId": result.ID,
			},
		},
	)
	if err != nil {
		return err
	}

	// 处理任务结果
	go td.handleTaskResult(taskID, result)

	return nil
}

// handleTaskResult 处理任务结果
func (td *TaskDispatcher) handleTaskResult(taskID string, result *models.TaskResult) {
	// 获取任务信息
	task, err := td.getTaskByID(taskID)
	if err != nil {
		td.logError("获取任务信息失败: " + err.Error())
		return
	}

	// 根据任务类型处理结果
	switch task.Type {
	case models.TaskTypeSubdomainEnum:
		td.processSubdomainResult(task, result)
	case models.TaskTypePortScan:
		td.processPortScanResult(task, result)
	case models.TaskTypeVulnScan:
		td.processVulnScanResult(task, result)
	case models.TaskTypeAssetDiscovery:
		td.processAssetDiscoveryResult(task, result)
	}

	// 触发任务回调
	if task.CallbackURL != "" {
		td.triggerTaskCallback(task, result)
	}

	// 检查依赖于此任务的其他任务，并尝试启动它们
	td.checkDependentTasks(taskID)
}

// getTaskByID 根据ID获取任务
func (td *TaskDispatcher) getTaskByID(taskID string) (*models.Task, error) {
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, err
	}

	var task models.Task
	err = td.db.Collection("tasks").FindOne(td.ctx, bson.M{"_id": objID}).Decode(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

// processSubdomainResult 处理子域名枚举结果
func (td *TaskDispatcher) processSubdomainResult(task *models.Task, result *models.TaskResult) {
	// 如果结果处理失败，记录错误但不影响任务完成状态
	defer func() {
		if r := recover(); r != nil {
			td.logError(fmt.Sprintf("处理子域名枚举结果时发生错误: %v", r))
		}
	}()

	td.logInfo(fmt.Sprintf("处理子域名枚举结果: 任务ID=%s", task.ID.Hex()))

	// 提取子域名结果数据
	subdomains, ok := result.Data["subdomains"].([]interface{})
	if !ok {
		td.logError("子域名结果数据格式不正确")
		return
	}

	// 处理每个子域名
	for _, item := range subdomains {
		if subdomain, ok := item.(map[string]interface{}); ok {
			// 创建或更新子域名资产
			td.createOrUpdateSubdomainAsset(task.ProjectID, subdomain)
		}
	}
}

// processPortScanResult 处理端口扫描结果
func (td *TaskDispatcher) processPortScanResult(task *models.Task, result *models.TaskResult) {
	// 如果结果处理失败，记录错误但不影响任务完成状态
	defer func() {
		if r := recover(); r != nil {
			td.logError(fmt.Sprintf("处理端口扫描结果时发生错误: %v", r))
		}
	}()

	td.logInfo(fmt.Sprintf("处理端口扫描结果: 任务ID=%s", task.ID.Hex()))

	// 提取端口扫描结果数据
	portResults, ok := result.Data["open_ports"].([]interface{})
	if !ok {
		td.logError("端口扫描结果数据格式不正确")
		return
	}

	// 处理每个端口扫描结果
	for _, item := range portResults {
		if portResult, ok := item.(map[string]interface{}); ok {
			// 创建或更新端口扫描资产
			td.createOrUpdatePortScanAsset(task.ProjectID, portResult)
		}
	}
}

// processVulnScanResult 处理漏洞扫描结果
func (td *TaskDispatcher) processVulnScanResult(task *models.Task, result *models.TaskResult) {
	// 如果结果处理失败，记录错误但不影响任务完成状态
	defer func() {
		if r := recover(); r != nil {
			td.logError(fmt.Sprintf("处理漏洞扫描结果时发生错误: %v", r))
		}
	}()

	td.logInfo(fmt.Sprintf("处理漏洞扫描结果: 任务ID=%s", task.ID.Hex()))

	// 提取漏洞扫描结果数据
	vulnResults, ok := result.Data["vulnerabilities"].([]interface{})
	if !ok {
		td.logError("漏洞扫描结果数据格式不正确")
		return
	}

	// 处理每个漏洞结果
	for _, item := range vulnResults {
		if vulnResult, ok := item.(map[string]interface{}); ok {
			// 创建或更新漏洞资产
			td.createOrUpdateVulnerability(task.ProjectID, vulnResult)
		}
	}
}

// processAssetDiscoveryResult 处理资产发现结果
func (td *TaskDispatcher) processAssetDiscoveryResult(task *models.Task, result *models.TaskResult) {
	// 如果结果处理失败，记录错误但不影响任务完成状态
	defer func() {
		if r := recover(); r != nil {
			td.logError(fmt.Sprintf("处理资产发现结果时发生错误: %v", r))
		}
	}()

	td.logInfo(fmt.Sprintf("处理资产发现结果: 任务ID=%s", task.ID.Hex()))

	// 提取资产发现结果数据
	discoveryResult, ok := result.Data["discovery_result"].(map[string]interface{})
	if !ok {
		td.logError("资产发现结果数据格式不正确")
		return
	}

	// 提取创建的资产
	assets, ok := discoveryResult["created_assets"].([]interface{})
	if !ok {
		td.logError("创建的资产数据格式不正确")
		return
	}

	// 处理每个资产
	for _, item := range assets {
		if asset, ok := item.(map[string]interface{}); ok {
			// 根据资产类型创建或更新资产
			assetType, _ := asset["type"].(string)
			switch assetType {
			case "domain":
				td.createOrUpdateSubdomainAsset(task.ProjectID, asset)
			case "host":
				td.createOrUpdateHostAsset(task.ProjectID, asset)
			case "service":
				td.createOrUpdateServiceAsset(task.ProjectID, asset)
			case "url":
				td.createOrUpdateURLAsset(task.ProjectID, asset)
			}
		}
	}
}

// createOrUpdateSubdomainAsset 创建或更新子域名资产
func (td *TaskDispatcher) createOrUpdateSubdomainAsset(projectID primitive.ObjectID, data map[string]interface{}) {
	// 实现子域名资产创建或更新逻辑
	// 这里只是一个简单的示例
	td.logInfo(fmt.Sprintf("创建或更新子域名资产: %v", data))
}

// createOrUpdatePortScanAsset 创建或更新端口扫描资产
func (td *TaskDispatcher) createOrUpdatePortScanAsset(projectID primitive.ObjectID, data map[string]interface{}) {
	// 实现端口扫描资产创建或更新逻辑
	// 这里只是一个简单的示例
	td.logInfo(fmt.Sprintf("创建或更新端口扫描资产: %v", data))
}

// createOrUpdateVulnerability 创建或更新漏洞
func (td *TaskDispatcher) createOrUpdateVulnerability(projectID primitive.ObjectID, data map[string]interface{}) {
	// 实现漏洞创建或更新逻辑
	// 这里只是一个简单的示例
	td.logInfo(fmt.Sprintf("创建或更新漏洞: %v", data))
}

// createOrUpdateHostAsset 创建或更新主机资产
func (td *TaskDispatcher) createOrUpdateHostAsset(projectID primitive.ObjectID, data map[string]interface{}) {
	// 实现主机资产创建或更新逻辑
	// 这里只是一个简单的示例
	td.logInfo(fmt.Sprintf("创建或更新主机资产: %v", data))
}

// createOrUpdateServiceAsset 创建或更新服务资产
func (td *TaskDispatcher) createOrUpdateServiceAsset(projectID primitive.ObjectID, data map[string]interface{}) {
	// 实现服务资产创建或更新逻辑
	// 这里只是一个简单的示例
	td.logInfo(fmt.Sprintf("创建或更新服务资产: %v", data))
}

// createOrUpdateURLAsset 创建或更新URL资产
func (td *TaskDispatcher) createOrUpdateURLAsset(projectID primitive.ObjectID, data map[string]interface{}) {
	// 实现URL资产创建或更新逻辑
	// 这里只是一个简单的示例
	td.logInfo(fmt.Sprintf("创建或更新URL资产: %v", data))
}

// triggerTaskCallback 触发任务回调
func (td *TaskDispatcher) triggerTaskCallback(task *models.Task, result *models.TaskResult) {
	// 实现任务回调逻辑
	// 这里只是一个简单的示例
	td.logInfo(fmt.Sprintf("触发任务回调: URL=%s, TaskID=%s", task.CallbackURL, task.ID.Hex()))
}

// checkDependentTasks 检查依赖于此任务的其他任务
func (td *TaskDispatcher) checkDependentTasks(taskID string) {
	// 查找依赖于此任务的其他任务
	cursor, err := td.db.Collection("tasks").Find(
		td.ctx,
		bson.M{
			"dependsOn": taskID,
			"status":    string(models.TaskStatusPending),
		},
	)
	if err != nil {
		td.logError("查找依赖任务失败: " + err.Error())
		return
	}
	defer cursor.Close(td.ctx)

	var dependentTasks []models.Task
	if err := cursor.All(td.ctx, &dependentTasks); err != nil {
		td.logError("解析依赖任务失败: " + err.Error())
		return
	}

	// 尝试启动依赖任务
	for _, depTask := range dependentTasks {
		// 检查所有依赖是否满足
		if td.checkTaskDependencies(&depTask) {
			// 将任务状态更新为已入队
			_ = td.UpdateTaskStatus(depTask.ID.Hex(), string(models.TaskStatusPending), 0)

			// 将任务加入队列
			queueName := "default"
			switch depTask.Type {
			case models.TaskTypeSubdomainEnum:
				queueName = "subdomain"
			case models.TaskTypePortScan:
				queueName = "portscan"
			case models.TaskTypeVulnScan:
				queueName = "vulnscan"
			case models.TaskTypeAssetDiscovery:
				queueName = "discovery"
			}

			// 将任务加入队列
			_ = td.queueManager.EnqueueTask(queueName, &depTask)
		}
	}
}

// worker 工作线程
func (td *TaskDispatcher) worker() {
	for {
		select {
		case <-td.ctx.Done():
			return
		case task, ok := <-td.taskChan:
			if !ok {
				return
			}

			// 处理任务
			td.processTask(task)
		}
	}
}

// scheduler 任务调度器
func (td *TaskDispatcher) scheduler() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-td.ctx.Done():
			return
		case <-ticker.C:
			// 检查是否有待处理的任务
			task, queueName, err := td.queueManager.GetNextTask()
			if err != nil {
				continue
			}

			// 检查任务依赖是否满足
			if !td.checkTaskDependencies(task) {
				// 如果依赖未满足，将任务重新加入队列
				_ = td.queueManager.EnqueueTask(queueName, task)
				continue
			}

			// 分配任务到工作线程
			td.taskChan <- task
		}
	}
}

// processTask 处理任务
func (td *TaskDispatcher) processTask(task *models.Task) {
	// 更新任务状态为运行中
	_ = td.UpdateTaskStatus(task.ID.Hex(), string(models.TaskStatusRunning), 0)

	// 选择合适的节点执行任务
	node, err := td.selectNode(task)
	if err != nil {
		// 如果没有合适的节点，更新任务状态为失败
		_ = td.UpdateTaskStatus(task.ID.Hex(), string(models.TaskStatusFailed), 0)
		_ = td.SaveTaskResult(task.ID.Hex(), &models.TaskResult{
			Status: "failed",
			Error:  "无法找到合适的节点执行任务: " + err.Error(),
		})
		return
	}

	// 将任务分配给节点
	task.NodeID = node.ID.Hex()
	_ = td.UpdateTaskStatus(task.ID.Hex(), string(models.TaskStatusRunning), 0)

	// 通过Redis发送任务到节点
	taskData, _ := bson.Marshal(task)
	err = td.redisClient.Publish(td.ctx, "task_assign:"+node.ID.Hex(), string(taskData)).Err()
	if err != nil {
		// 如果发送失败，更新任务状态为失败
		_ = td.UpdateTaskStatus(task.ID.Hex(), string(models.TaskStatusFailed), 0)
		_ = td.SaveTaskResult(task.ID.Hex(), &models.TaskResult{
			Status: "failed",
			Error:  "无法将任务发送到节点: " + err.Error(),
		})
		return
	}

	// 设置任务超时
	if task.Timeout > 0 {
		go func() {
			timer := time.NewTimer(time.Duration(task.Timeout) * time.Second)
			defer timer.Stop()

			select {
			case <-td.ctx.Done():
				return
			case <-timer.C:
				// 检查任务状态
				status, err := td.GetTaskStatus(task.ID.Hex())
				if err != nil {
					return
				}

				// 如果任务仍在运行，则标记为超时
				if status == string(models.TaskStatusRunning) {
					_ = td.UpdateTaskStatus(task.ID.Hex(), string(models.TaskStatusTimeout), task.Progress)
					_ = td.SaveTaskResult(task.ID.Hex(), &models.TaskResult{
						Status: "timeout",
						Error:  "任务执行超时",
					})

					// 通过Redis发送取消信号
					_ = td.redisClient.Publish(td.ctx, "task_cancel:"+node.ID.Hex(), task.ID.Hex()).Err()
				}
			}
		}()
	}
}

// checkTaskDependencies 检查任务依赖是否满足
func (td *TaskDispatcher) checkTaskDependencies(task *models.Task) bool {
	if len(task.DependsOn) == 0 {
		return true
	}

	// 获取所有依赖任务的状态
	dependencyMet := true
	dependencyStatuses := make(map[string]string)
	dependencyResults := make(map[string]*models.TaskResult)

	for _, depID := range task.DependsOn {
		objID, err := primitive.ObjectIDFromHex(depID)
		if err != nil {
			// 如果依赖ID无效，标记依赖未满足
			dependencyMet = false
			continue
		}

		var depTask models.Task
		err = td.db.Collection("tasks").FindOne(td.ctx, bson.M{"_id": objID}).Decode(&depTask)
		if err != nil {
			// 如果依赖任务不存在，标记依赖未满足
			dependencyMet = false
			continue
		}

		// 记录依赖任务状态
		dependencyStatuses[depID] = depTask.Status

		// 检查依赖任务是否已完成
		if depTask.Status != string(models.TaskStatusCompleted) {
			dependencyMet = false
			continue
		}

		// 获取依赖任务的结果
		if !depTask.ResultID.IsZero() {
			var result models.TaskResult
			err = td.db.Collection("task_results").FindOne(td.ctx, bson.M{"_id": depTask.ResultID}).Decode(&result)
			if err == nil {
				dependencyResults[depID] = &result
			}
		}
	}

	// 如果所有依赖都满足，处理依赖任务的结果数据
	if dependencyMet {
		// 将依赖任务的结果合并到当前任务的参数中
		if task.Params == nil {
			task.Params = make(map[string]interface{})
		}

		paramsMap, ok := task.Params.(map[string]interface{})
		if !ok {
			paramsMap = make(map[string]interface{})
		}

		// 添加依赖结果数据
		dependencyData := make(map[string]interface{})
		for depID, result := range dependencyResults {
			dependencyData[depID] = result.Data
		}

		// 如果已有依赖数据字段，合并而不是覆盖
		if existingData, ok := paramsMap["dependencyData"]; ok {
			if existingMap, ok := existingData.(map[string]interface{}); ok {
				for k, v := range dependencyData {
					existingMap[k] = v
				}
				paramsMap["dependencyData"] = existingMap
			} else {
				paramsMap["dependencyData"] = dependencyData
			}
		} else {
			paramsMap["dependencyData"] = dependencyData
		}

		task.Params = paramsMap

		// 更新任务参数
		_, err := td.db.Collection("tasks").UpdateOne(
			td.ctx,
			bson.M{"_id": task.ID},
			bson.M{
				"$set": bson.M{
					"params": task.Params,
				},
			},
		)
		if err != nil {
			// 如果更新失败，记录错误但继续执行任务
			td.logError("更新任务参数失败: " + err.Error())
		}

		return true
	}

	// 记录依赖状态，便于调试
	td.logInfo(fmt.Sprintf("任务 %s 依赖未满足: %v", task.ID.Hex(), dependencyStatuses))

	return false
}

// logInfo 记录信息日志
func (td *TaskDispatcher) logInfo(message string) {
	// 这里可以实现日志记录逻辑
	// 简单实现，将日志写入MongoDB
	logEntry := bson.M{
		"level":     "info",
		"message":   message,
		"timestamp": time.Now(),
		"component": "TaskDispatcher",
	}
	_, _ = td.db.Collection("system_logs").InsertOne(td.ctx, logEntry)
}

// logError 记录错误日志
func (td *TaskDispatcher) logError(message string) {
	// 这里可以实现日志记录逻辑
	// 简单实现，将日志写入MongoDB
	logEntry := bson.M{
		"level":     "error",
		"message":   message,
		"timestamp": time.Now(),
		"component": "TaskDispatcher",
	}
	_, _ = td.db.Collection("system_logs").InsertOne(td.ctx, logEntry)
}

// selectNode 选择合适的节点执行任务
func (td *TaskDispatcher) selectNode(task *models.Task) (*models.Node, error) {
	// 获取所有在线节点
	nodes := td.nodeManager.GetNodesByStatus(string(models.NodeStatusOnline))
	if len(nodes) == 0 {
		return nil, errors.New("没有在线节点可用")
	}

	// 根据任务类型过滤节点
	var candidates []*models.Node
	for _, node := range nodes {
		// 检查节点是否支持该任务类型
		if len(node.Config.EnabledTaskTypes) > 0 {
			supported := false
			for _, t := range node.Config.EnabledTaskTypes {
				if t == task.Type {
					supported = true
					break
				}
			}
			if !supported {
				continue
			}
		}

		// 检查节点当前运行的任务数是否达到上限
		if node.Config.MaxConcurrentTasks > 0 && node.StatusInfo.RunningTasks >= node.Config.MaxConcurrentTasks {
			continue
		}

		candidates = append(candidates, node)
	}

	if len(candidates) == 0 {
		return nil, errors.New("没有合适的节点可执行该任务")
	}

	// 高级负载均衡算法：考虑多种因素
	type nodeScore struct {
		node  *models.Node
		score float64
	}

	var scores []nodeScore
	for _, node := range candidates {
		// 基础分数 - 反比于运行任务数
		baseScore := 100.0
		if node.StatusInfo.RunningTasks > 0 {
			baseScore = 100.0 / float64(node.StatusInfo.RunningTasks+1)
		}

		// CPU使用率因子 (CPU使用率越低，分数越高)
		cpuFactor := 1.0
		if node.StatusInfo.CpuUsage > 0 {
			cpuFactor = 1.0 - (node.StatusInfo.CpuUsage/100.0)*0.8 // CPU使用率影响80%
		}

		// 内存使用率因子 (内存使用率越低，分数越高)
		memFactor := 1.0
		if node.StatusInfo.MemoryUsage > 0 {
			// 将内存使用量转换为百分比 (假设最大内存为节点配置的MaxMemoryUsage)
			memUsagePercent := 0.0
			if node.Config.MaxMemoryUsage > 0 {
				memUsagePercent = float64(node.StatusInfo.MemoryUsage) / float64(node.Config.MaxMemoryUsage) * 100.0
			} else {
				// 如果没有配置最大内存，假设使用率为50%
				memUsagePercent = 50.0
			}
			memFactor = 1.0 - (memUsagePercent/100.0)*0.6 // 内存使用率影响60%
		}

		// 网络因子 (网络流量越低，分数越高)
		netFactor := 1.0
		if node.StatusInfo.NetworkIn > 0 || node.StatusInfo.NetworkOut > 0 {
			// 假设最大网络流量为100MB/s (102400KB/s)
			netUsage := float64(node.StatusInfo.NetworkIn+node.StatusInfo.NetworkOut) / 102400.0
			netFactor = 1.0 - math.Min(netUsage, 1.0)*0.4 // 网络使用率影响40%
		}

		// 任务亲和性 (如果节点之前处理过同类任务，给予额外分数)
		affinityFactor := 1.0
		if node.TaskStats.TaskTypeStats != nil {
			if count, ok := node.TaskStats.TaskTypeStats[task.Type]; ok && count > 0 {
				affinityFactor = 1.2 // 提高20%的分数
			}
		}

		// 计算最终分数
		finalScore := baseScore * cpuFactor * memFactor * netFactor * affinityFactor

		// 优先级调整 (高优先级任务更倾向于选择性能更好的节点)
		if task.Priority >= models.TaskPriorityHigh {
			// 对于高优先级任务，更重视CPU和内存因素
			finalScore = finalScore * (cpuFactor * 1.5) * (memFactor * 1.3)
		}

		scores = append(scores, nodeScore{
			node:  node,
			score: finalScore,
		})
	}

	// 按分数排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// 选择得分最高的节点
	return scores[0].node, nil
}
