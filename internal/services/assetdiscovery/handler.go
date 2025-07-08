package assetdiscovery

import (
	"context"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Handler 资产发现处理器
type Handler struct {
	db      *mongo.Database
	service *DiscoveryService
}

// NewHandler 创建资产发现处理器
func NewHandler(db *mongo.Database, service *DiscoveryService) *Handler {
	return &Handler{
		db:      db,
		service: service,
	}
}

// CreateDiscoveryTask 创建资产发现任务
func (h *Handler) CreateDiscoveryTask(task *models.AssetDiscoveryTask) (string, error) {
	return h.service.CreateTask(task)
}

// StartDiscoveryTask 启动资产发现任务
func (h *Handler) StartDiscoveryTask(taskID string) error {
	return h.service.StartTask(taskID)
}

// StopDiscoveryTask 停止资产发现任务
func (h *Handler) StopDiscoveryTask(taskID string) error {
	return h.service.StopTask(taskID)
}

// GetDiscoveryTaskStatus 获取资产发现任务状态
func (h *Handler) GetDiscoveryTaskStatus(taskID string) (string, error) {
	return h.service.GetTaskStatus(taskID)
}

// GetDiscoveryTaskProgress 获取资产发现任务进度
func (h *Handler) GetDiscoveryTaskProgress(taskID string) (float64, error) {
	return h.service.GetTaskProgress(taskID)
}

// GetDiscoveryTasks 获取资产发现任务列表
func (h *Handler) GetDiscoveryTasks(filter map[string]interface{}, page, pageSize int) ([]models.AssetDiscoveryTask, int, error) {
	// 构建查询条件
	query := bson.M{}
	for key, value := range filter {
		query[key] = value
	}

	// 计算总数
	total, err := h.db.Collection("asset_discovery_tasks").CountDocuments(context.Background(), query)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	skip := (page - 1) * pageSize
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(pageSize)).SetSort(bson.D{{"createdAt", -1}})
	cursor, err := h.db.Collection("asset_discovery_tasks").Find(context.Background(), query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.Background())

	// 解析结果
	var tasks []models.AssetDiscoveryTask
	if err = cursor.All(context.Background(), &tasks); err != nil {
		return nil, 0, err
	}

	return tasks, int(total), nil
}

// GetDiscoveryTask 获取资产发现任务
func (h *Handler) GetDiscoveryTask(taskID string) (*models.AssetDiscoveryTask, error) {
	// 解析任务ID
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, err
	}

	// 查询任务
	var task models.AssetDiscoveryTask
	err = h.db.Collection("asset_discovery_tasks").FindOne(context.Background(), bson.M{
		"_id": objID,
	}).Decode(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

// DeleteDiscoveryTask 删除资产发现任务
func (h *Handler) DeleteDiscoveryTask(taskID string) error {
	// 解析任务ID
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return err
	}

	// 检查任务状态
	var task models.AssetDiscoveryTask
	err = h.db.Collection("asset_discovery_tasks").FindOne(context.Background(), bson.M{
		"_id": objID,
	}).Decode(&task)
	if err != nil {
		return err
	}

	// 如果任务正在运行，先停止任务
	if task.Status == "running" {
		err = h.service.StopTask(taskID)
		if err != nil {
			return err
		}
	}

	// 删除任务
	_, err = h.db.Collection("asset_discovery_tasks").DeleteOne(context.Background(), bson.M{
		"_id": objID,
	})
	return err
}

// GetDiscoveryResults 获取资产发现结果
func (h *Handler) GetDiscoveryResults(taskID string) ([]*models.DiscoveryResult, error) {
	// 解析任务ID
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, err
	}

	// 查询结果
	cursor, err := h.db.Collection("discovery_results").Find(context.Background(), bson.M{
		"taskId": objID,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// 解析结果
	var results []*models.DiscoveryResult
	if err = cursor.All(context.Background(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

// SaveDiscoveryResult 保存资产发现结果
func (h *Handler) SaveDiscoveryResult(result *models.DiscoveryResult) error {
	// 设置创建时间
	if result.FirstSeen.IsZero() {
		result.FirstSeen = time.Now()
	}

	// 设置最后一次见到的时间
	result.LastSeen = time.Now()

	// 保存结果
	_, err := h.db.Collection("discovery_results").InsertOne(context.Background(), result)
	return err
}

// HandleDiscoveryResult 处理资产发现结果
func (h *Handler) HandleDiscoveryResult(task *DiscoveryTask, result *models.DiscoveryResult) error {
	// 保存结果
	return h.SaveDiscoveryResult(result)
}

// UpdateTaskStatus 更新任务状态
func (h *Handler) UpdateTaskStatus(taskID string, status string, progress float64) error {
	// 解析任务ID
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return err
	}

	// 更新状态
	update := bson.M{
		"$set": bson.M{
			"status":   status,
			"progress": progress,
		},
	}

	// 如果任务完成或失败，设置完成时间
	if status == "completed" || status == "failed" || status == "stopped" {
		update["$set"].(bson.M)["completedAt"] = time.Now()
	}

	// 如果任务开始运行，设置开始时间
	if status == "running" {
		update["$set"].(bson.M)["startedAt"] = time.Now()
	}

	_, err = h.db.Collection("asset_discovery_tasks").UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		update,
	)
	return err
}
