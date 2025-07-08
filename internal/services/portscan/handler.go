package portscan

import (
	"context"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Handler 端口扫描处理器
type Handler struct {
	db      *mongo.Database
	manager *Manager
}

// NewHandler 创建端口扫描处理器
func NewHandler(db *mongo.Database, manager *Manager) *Handler {
	return &Handler{
		db:      db,
		manager: manager,
	}
}

// CreateScanTask 创建扫描任务
func (h *Handler) CreateScanTask(task *models.PortScanTask) (string, error) {
	return h.manager.CreateTask(task)
}

// StartScanTask 启动扫描任务
func (h *Handler) StartScanTask(taskID string) error {
	return h.manager.StartTask(taskID)
}

// StopScanTask 停止扫描任务
func (h *Handler) StopScanTask(taskID string) error {
	return h.manager.StopTask(taskID)
}

// GetScanTaskStatus 获取扫描任务状态
func (h *Handler) GetScanTaskStatus(taskID string) (string, error) {
	return h.manager.GetTaskStatus(taskID)
}

// GetScanTaskProgress 获取扫描任务进度
func (h *Handler) GetScanTaskProgress(taskID string) (float64, error) {
	return h.manager.GetTaskProgress(taskID)
}

// GetScanTasks 获取扫描任务列表
func (h *Handler) GetScanTasks(filter map[string]interface{}, page, pageSize int) ([]models.PortScanTask, int, error) {
	// 构建查询条件
	query := bson.M{}
	for key, value := range filter {
		query[key] = value
	}

	// 计算总数
	total, err := h.db.Collection("port_scan_tasks").CountDocuments(context.Background(), query)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	skip := (page - 1) * pageSize
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(pageSize)).SetSort(bson.D{{"createdAt", -1}})
	cursor, err := h.db.Collection("port_scan_tasks").Find(context.Background(), query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.Background())

	// 解析结果
	var tasks []models.PortScanTask
	if err = cursor.All(context.Background(), &tasks); err != nil {
		return nil, 0, err
	}

	return tasks, int(total), nil
}

// GetScanTask 获取扫描任务
func (h *Handler) GetScanTask(taskID string) (*models.PortScanTask, error) {
	// 解析任务ID
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, err
	}

	// 查询任务
	var task models.PortScanTask
	err = h.db.Collection("port_scan_tasks").FindOne(context.Background(), bson.M{
		"_id": objID,
	}).Decode(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

// DeleteScanTask 删除扫描任务
func (h *Handler) DeleteScanTask(taskID string) error {
	// 解析任务ID
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return err
	}

	// 检查任务状态
	var task models.PortScanTask
	err = h.db.Collection("port_scan_tasks").FindOne(context.Background(), bson.M{
		"_id": objID,
	}).Decode(&task)
	if err != nil {
		return err
	}

	// 如果任务正在运行，先停止任务
	if task.Status == "running" {
		err = h.manager.StopTask(taskID)
		if err != nil {
			return err
		}
	}

	// 删除任务
	_, err = h.db.Collection("port_scan_tasks").DeleteOne(context.Background(), bson.M{
		"_id": objID,
	})
	return err
}

// GetScanResults 获取扫描结果
func (h *Handler) GetScanResults(taskID string) ([]models.PortScanResult, error) {
	// 解析任务ID
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, err
	}

	// 查询结果
	cursor, err := h.db.Collection("port_scan_results").Find(context.Background(), bson.M{
		"taskId": objID,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// 解析结果
	var results []models.PortScanResult
	if err = cursor.All(context.Background(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

// SaveScanResult 保存扫描结果
func (h *Handler) SaveScanResult(result *models.PortScanResult) error {
	// 设置创建时间
	if result.CreatedAt.IsZero() {
		result.CreatedAt = time.Now()
	}

	// 保存结果
	_, err := h.db.Collection("port_scan_results").InsertOne(context.Background(), result)
	return err
}

// UpdateTaskProgress 更新任务进度
func (h *Handler) UpdateTaskProgress(taskID string, progress float64) error {
	// 解析任务ID
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return err
	}

	// 更新进度
	_, err = h.db.Collection("port_scan_tasks").UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"progress": progress}},
	)
	return err
}

// MongoResultHandler MongoDB结果处理器
type MongoResultHandler struct {
	db *mongo.Database
}

// NewMongoResultHandler 创建MongoDB结果处理器
func NewMongoResultHandler(db *mongo.Database) *MongoResultHandler {
	return &MongoResultHandler{
		db: db,
	}
}

// HandleResult 处理扫描结果
func (h *MongoResultHandler) HandleResult(result *models.PortScanResult) error {
	// 设置创建时间
	if result.CreatedAt.IsZero() {
		result.CreatedAt = time.Now()
	}
	result.UpdatedAt = time.Now()

	// 检查是否已存在相同结果
	filter := bson.M{
		"host":     result.Host,
		"port":     result.Port,
		"protocol": result.Protocol,
		"taskId":   result.TaskID,
	}

	// 使用upsert操作，如果存在则更新，不存在则插入
	update := bson.M{
		"$set": bson.M{
			"status":    result.Status,
			"service":   result.Service,
			"product":   result.Product,
			"version":   result.Version,
			"extraInfo": result.ExtraInfo,
			"banner":    result.Banner,
			"updatedAt": result.UpdatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := h.db.Collection("port_scan_results").UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		return err
	}

	// 创建或更新端口资产
	err = h.createOrUpdatePortAsset(result)
	if err != nil {
		return err
	}

	// 更新任务摘要
	err = h.updateTaskSummary(result.TaskID)
	if err != nil {
		return err
	}

	return nil
}

// SaveTask 保存任务
func (h *MongoResultHandler) SaveTask(task *models.PortScanTask) error {
	// 设置创建时间
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	// 如果ID为空，生成新ID
	if task.ID.IsZero() {
		task.ID = primitive.NewObjectID()
	}

	// 初始化结果摘要
	task.ResultSummary = models.PortScanSummary{
		TotalHosts:     len(task.Targets),
		ServiceStats:   make(map[string]int),
		ProcessedHosts: 0,
	}

	// 保存任务
	_, err := h.db.Collection("port_scan_tasks").InsertOne(context.Background(), task)
	return err
}

// UpdateTaskStatus 更新任务状态
func (h *MongoResultHandler) UpdateTaskStatus(taskID string, status string) error {
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	// 如果状态是running，设置开始时间
	if status == "running" {
		update["$set"].(bson.M)["startedAt"] = time.Now()
	}

	// 如果状态是completed或failed，设置完成时间
	if status == "completed" || status == "failed" {
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
func (h *MongoResultHandler) FinishTask(task *models.PortScanTask) error {
	// 更新任务状态
	task.Status = "completed"
	task.CompletedAt = time.Now()
	task.Progress = 100

	update := bson.M{
		"$set": bson.M{
			"status":      task.Status,
			"completedAt": task.CompletedAt,
			"progress":    task.Progress,
		},
	}

	_, err := h.db.Collection("port_scan_tasks").UpdateOne(
		context.Background(),
		bson.M{"_id": task.ID},
		update,
	)
	return err
}

// createOrUpdatePortAsset 创建或更新端口资产
func (h *MongoResultHandler) createOrUpdatePortAsset(result *models.PortScanResult) error {
	// 只处理开放的端口
	if result.Status != "open" {
		return nil
	}

	// 查询项目ID
	var task models.PortScanTask
	err := h.db.Collection("port_scan_tasks").FindOne(
		context.Background(),
		bson.M{"_id": result.TaskID},
	).Decode(&task)
	if err != nil {
		return err
	}

	// 检查是否已存在相同资产
	filter := bson.M{
		"host":      result.Host,
		"port":      result.Port,
		"protocol":  result.Protocol,
		"projectId": task.ProjectID,
		"type":      "port",
	}

	// 准备资产数据
	now := time.Now()
	// 不使用asset变量，避免未使用变量错误
	_ = models.PortAsset{
		BaseAsset: models.BaseAsset{
			ID:        primitive.NewObjectID(),
			ProjectID: task.ProjectID,
			Type:      "port",
			CreatedAt: now,
			UpdatedAt: now,
			Tags:      []string{result.Protocol, result.Service},
		},
		Host:     result.Host,
		Port:     result.Port,
		Protocol: result.Protocol,
		Service:  result.Service,
		Version:  result.Version,
		Banner:   result.Banner,
	}

	// 使用upsert操作，如果存在则更新，不存在则插入
	update := bson.M{
		"$set": bson.M{
			"service":   result.Service,
			"version":   result.Version,
			"banner":    result.Banner,
			"updatedAt": now,
		},
		"$addToSet": bson.M{
			"tags": bson.M{
				"$each": []string{result.Protocol, result.Service},
			},
		},
	}

	opts := options.Update().SetUpsert(true)
	updateResult, err := h.db.Collection("assets").UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		return err
	}

	// 如果是新创建的资产，保存资产ID到结果中
	if updateResult.UpsertedID != nil {
		if objID, ok := updateResult.UpsertedID.(primitive.ObjectID); ok {
			// 更新结果中的assetId字段
			_, err = h.db.Collection("port_scan_results").UpdateOne(
				context.Background(),
				bson.M{"_id": result.ID},
				bson.M{"$set": bson.M{"assetId": objID}},
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// updateTaskSummary 更新任务摘要
func (h *MongoResultHandler) updateTaskSummary(taskID primitive.ObjectID) error {
	// 统计开放的端口数
	openPortsCount, err := h.db.Collection("port_scan_results").CountDocuments(
		context.Background(),
		bson.M{
			"taskId": taskID,
			"status": "open",
		},
	)
	if err != nil {
		return err
	}

	// 统计扫描的主机数
	var hosts []string
	cursor, err := h.db.Collection("port_scan_results").Distinct(
		context.Background(),
		"host",
		bson.M{"taskId": taskID},
	)
	if err != nil {
		return err
	}
	for _, host := range cursor {
		if hostStr, ok := host.(string); ok {
			hosts = append(hosts, hostStr)
		}
	}

	// 统计服务类型
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"taskId": taskID,
				"status": "open",
			},
		},
		{
			"$group": bson.M{
				"_id": "$service",
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
	}

	// 修复类型错误，正确处理Aggregate返回的*mongo.Cursor
	cursor2, err := h.db.Collection("port_scan_results").Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cursor2.Close(context.Background())

	serviceStats := make(map[string]int)
	for cursor2.Next(context.Background()) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor2.Decode(&result); err != nil {
			return err
		}
		serviceStats[result.ID] = result.Count
	}

	// 更新任务摘要
	update := bson.M{
		"$set": bson.M{
			"resultSummary.upHosts":        len(hosts),
			"resultSummary.openPorts":      openPortsCount,
			"resultSummary.serviceStats":   serviceStats,
			"resultSummary.processedHosts": len(hosts),
		},
	}

	_, err = h.db.Collection("port_scan_tasks").UpdateOne(
		context.Background(),
		bson.M{"_id": taskID},
		update,
	)
	return err
}

// ListTasks 列出任务
func (h *MongoResultHandler) ListTasks(query map[string]interface{}, limit, skip int) ([]*models.PortScanTask, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.D{{"createdAt", -1}})

	cursor, err := h.db.Collection("port_scan_tasks").Find(context.Background(), query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var tasks []*models.PortScanTask
	err = cursor.All(context.Background(), &tasks)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetTask 获取任务
func (h *MongoResultHandler) GetTask(taskID string) (*models.PortScanTask, error) {
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, err
	}

	var task models.PortScanTask
	err = h.db.Collection("port_scan_tasks").FindOne(
		context.Background(),
		bson.M{"_id": objID},
	).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &task, nil
}

// UpdateTaskProgress updates the progress of a task
func (h *MongoResultHandler) UpdateTaskProgress(taskID string, progress float64) error {
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"progress": progress}}

	_, err = h.db.Collection("port_scan_tasks").UpdateOne(context.Background(), filter, update)
	return err
}
