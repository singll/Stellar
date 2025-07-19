package pagemonitoring

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	pkgerrors "github.com/StellarServer/internal/pkg/errors"
	"github.com/StellarServer/internal/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PageMonitoringService 页面监控服务
type PageMonitoringService struct {
	db          *mongo.Database
	redisClient *redis.Client
	ctx         context.Context
	cancel      context.CancelFunc
	mutex       sync.RWMutex
}

// NewPageMonitoringService 创建页面监控服务
func NewPageMonitoringService(db *mongo.Database, redisClient *redis.Client) *PageMonitoringService {
	ctx, cancel := context.WithCancel(context.Background())
	return &PageMonitoringService{
		db:          db,
		redisClient: redisClient,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动服务
func (s *PageMonitoringService) Start() error {
	// 初始化数据库索引
	if err := s.initIndexes(); err != nil {
		return err
	}

	// 启动定时任务
	go s.scheduleMonitoringTasks()

	return nil
}

// Stop 停止服务
func (s *PageMonitoringService) Stop() {
	s.cancel()
}

// initIndexes 初始化数据库索引
func (s *PageMonitoringService) initIndexes() error {
	// 为PageMonitoring集合创建URL唯一索引
	_, err := s.db.Collection("page_monitoring").Indexes().CreateOne(
		s.ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "url", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		logger.Error("initIndexes create page_monitoring index failed", map[string]interface{}{"error": err})
		return pkgerrors.WrapDatabaseError(err, "创建页面监控索引")
	}

	// 为PageSnapshot集合创建索引
	_, err = s.db.Collection("page_snapshots").Indexes().CreateOne(
		s.ctx,
		mongo.IndexModel{
			Keys: bson.D{{Key: "monitoringId", Value: 1}, {Key: "createdAt", Value: -1}},
		},
	)
	if err != nil {
		logger.Error("initIndexes create page_snapshots index failed", map[string]interface{}{"error": err})
		return pkgerrors.WrapDatabaseError(err, "创建页面快照索引")
	}

	// 为PageChange集合创建索引
	_, err = s.db.Collection("page_changes").Indexes().CreateOne(
		s.ctx,
		mongo.IndexModel{
			Keys: bson.D{{Key: "monitoringId", Value: 1}, {Key: "changedAt", Value: -1}},
		},
	)
	if err != nil {
		logger.Error("initIndexes create page_changes index failed", map[string]interface{}{"error": err})
		return pkgerrors.WrapDatabaseError(err, "创建页面变更索引")
	}

	return nil
}

// scheduleMonitoringTasks 调度监控任务
func (s *PageMonitoringService) scheduleMonitoringTasks() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkAndScheduleTasks()
		}
	}
}

// checkAndScheduleTasks 检查并调度任务
func (s *PageMonitoringService) checkAndScheduleTasks() {
	// 查找需要执行的监控任务
	now := time.Now()
	filter := bson.M{
		"status":      models.PageMonitoringStatusActive,
		"nextCheckAt": bson.M{"$lte": now},
	}

	cursor, err := s.db.Collection("page_monitoring").Find(s.ctx, filter)
	if err != nil {
		logger.Error("checkAndScheduleTasks find tasks failed", map[string]interface{}{"error": err})
		return
	}
	defer cursor.Close(s.ctx)

	var tasks []models.PageMonitoring
	if err := cursor.All(s.ctx, &tasks); err != nil {
		logger.Error("checkAndScheduleTasks decode tasks failed", map[string]interface{}{"error": err})
		return
	}

	// 将任务分发到节点
	for _, task := range tasks {
		s.dispatchMonitoringTask(&task)
	}
}

// dispatchMonitoringTask 分发监控任务
func (s *PageMonitoringService) dispatchMonitoringTask(task *models.PageMonitoring) {
	// 更新下次检查时间
	nextCheckAt := time.Now().Add(time.Duration(task.Interval) * time.Hour)
	_, err := s.db.Collection("page_monitoring").UpdateOne(
		s.ctx,
		bson.M{"_id": task.ID},
		bson.M{
			"$set": bson.M{
				"lastCheckAt": time.Now(),
				"nextCheckAt": nextCheckAt,
			},
		},
	)
	if err != nil {
		logger.Error("dispatchMonitoringTask update task failed", map[string]interface{}{"taskID": task.ID.Hex(), "error": err})
		return
	}

	// 将任务添加到Redis队列
	taskData := map[string]interface{}{
		"id":     task.ID.Hex(),
		"url":    task.URL,
		"config": task.Config,
		"type":   "page_monitoring",
	}

	// 序列化任务数据
	taskBytes, err := bson.Marshal(taskData)
	if err != nil {
		logger.Error("dispatchMonitoringTask marshal task failed", map[string]interface{}{"taskID": task.ID.Hex(), "error": err})
		return
	}

	// 添加到任务队列
	err = s.redisClient.LPush(s.ctx, "task_queue:page_monitoring", string(taskBytes)).Err()
	if err != nil {
		logger.Error("dispatchMonitoringTask add to redis failed", map[string]interface{}{"taskID": task.ID.Hex(), "error": err})
		return
	}
}

// CreateMonitoring 创建监控任务
func (s *PageMonitoringService) CreateMonitoring(req *models.PageMonitoringCreateRequest) (*models.PageMonitoring, error) {
	// 验证URL
	if req.URL == "" {
		logger.Error("CreateMonitoring empty URL", map[string]interface{}{"request": req})
		return nil, pkgerrors.NewAppError(pkgerrors.CodeValidationFailed, "URL不能为空", 400)
	}

	// 验证项目ID
	projectID := req.ProjectID
	if projectID.IsZero() {
		logger.Error("CreateMonitoring invalid projectID", map[string]interface{}{"request": req})
		return nil, pkgerrors.NewAppError(pkgerrors.CodeValidationFailed, "无效的项目ID", 400)
	}

	// 设置默认值
	if req.Name == "" {
		req.Name = req.URL
	}
	if req.Interval <= 0 {
		req.Interval = 24 // 默认24小时
	}

	// 创建监控任务
	now := time.Now()
	monitoring := &models.PageMonitoring{
		ID:          primitive.NewObjectID(),
		URL:         req.URL,
		Name:        req.Name,
		Status:      string(models.PageMonitoringStatusActive),
		ProjectID:   projectID,
		Interval:    req.Interval,
		LastCheckAt: time.Time{},
		NextCheckAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
		Tags:        req.Tags,
		Config:      req.Config,
		ChangeCount: 0,
	}

	// 设置默认配置
	if monitoring.Config.CompareMethod == "" {
		monitoring.Config.CompareMethod = "html"
	}
	if monitoring.Config.SimilarityThreshold == 0 {
		monitoring.Config.SimilarityThreshold = 0.9
	}
	if monitoring.Config.Timeout == 0 {
		monitoring.Config.Timeout = 30
	}

	// 保存到数据库
	_, err := s.db.Collection("page_monitoring").InsertOne(s.ctx, monitoring)
	if err != nil {
		logger.Error("CreateMonitoring insert failed", map[string]interface{}{"monitoring": monitoring, "error": err})
		return nil, pkgerrors.WrapDatabaseError(err, "创建页面监控任务")
	}

	// 立即调度一次任务
	go s.dispatchMonitoringTask(monitoring)

	return monitoring, nil
}

// UpdateMonitoring 更新监控任务
func (s *PageMonitoringService) UpdateMonitoring(id string, req *models.PageMonitoringUpdateRequest) (*models.PageMonitoring, error) {
	// 验证ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("无效的监控ID")
	}

	// 构建更新数据
	update := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	// 更新名称
	if req.Name != "" {
		update["$set"].(bson.M)["name"] = req.Name
	}

	// 更新状态
	if req.Status != "" {
		update["$set"].(bson.M)["status"] = req.Status
	}

	// 更新间隔
	if req.Interval > 0 {
		update["$set"].(bson.M)["interval"] = req.Interval
	}

	// 更新标签
	if req.Tags != nil {
		update["$set"].(bson.M)["tags"] = req.Tags
	}

	// 更新配置
	if req.Config.CompareMethod != "" {
		update["$set"].(bson.M)["config"] = req.Config
	}

	// 执行更新
	result := s.db.Collection("page_monitoring").FindOneAndUpdate(
		s.ctx,
		bson.M{"_id": objID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	// 检查结果
	var monitoring models.PageMonitoring
	if err := result.Decode(&monitoring); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("监控任务不存在")
		}
		return nil, err
	}

	return &monitoring, nil
}

// GetMonitoring 获取监控任务
func (s *PageMonitoringService) GetMonitoring(id string) (*models.PageMonitoring, error) {
	// 验证ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("无效的监控ID")
	}

	// 查询数据库
	var monitoring models.PageMonitoring
	err = s.db.Collection("page_monitoring").FindOne(
		s.ctx,
		bson.M{"_id": objID},
	).Decode(&monitoring)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("监控任务不存在")
		}
		return nil, err
	}

	return &monitoring, nil
}

// ListMonitorings 列出监控任务
func (s *PageMonitoringService) ListMonitorings(req *models.PageMonitoringQueryRequest) ([]*models.PageMonitoring, int64, error) {
	// 构建查询条件
	filter := bson.M{}

	// 项目ID过滤
	if req.ProjectID != "" {
		projectID, err := primitive.ObjectIDFromHex(req.ProjectID)
		if err == nil {
			filter["projectId"] = projectID
		}
	}

	// 状态过滤
	if req.Status != "" {
		filter["status"] = req.Status
	}

	// 标签过滤
	if len(req.Tags) > 0 {
		filter["tags"] = bson.M{"$in": req.Tags}
	}

	// URL关键字过滤
	if req.URL != "" {
		filter["url"] = bson.M{"$regex": req.URL, "$options": "i"}
	}

	// 名称关键字过滤
	if req.Name != "" {
		filter["name"] = bson.M{"$regex": req.Name, "$options": "i"}
	}

	// 设置分页
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// 设置排序
	sortOpt := options.Find()
	if req.SortBy != "" {
		sortDir := 1
		if req.SortOrder == "desc" {
			sortDir = -1
		}
		sortOpt.SetSort(bson.D{{Key: req.SortBy, Value: sortDir}})
	} else {
		sortOpt.SetSort(bson.D{{Key: "createdAt", Value: -1}})
	}

	// 设置分页
	sortOpt.SetSkip(int64(req.Offset)).SetLimit(int64(req.Limit))

	// 查询总数
	total, err := s.db.Collection("page_monitoring").CountDocuments(s.ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	cursor, err := s.db.Collection("page_monitoring").Find(s.ctx, filter, sortOpt)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(s.ctx)

	// 解析结果
	var monitorings []*models.PageMonitoring
	if err := cursor.All(s.ctx, &monitorings); err != nil {
		return nil, 0, err
	}

	return monitorings, total, nil
}

// DeleteMonitoring 删除监控任务
func (s *PageMonitoringService) DeleteMonitoring(id string) error {
	// 验证ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("无效的监控ID")
	}

	// 删除监控任务
	result, err := s.db.Collection("page_monitoring").DeleteOne(s.ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("监控任务不存在")
	}

	// 删除相关的快照和变更记录
	_, _ = s.db.Collection("page_snapshots").DeleteMany(s.ctx, bson.M{"monitoringId": objID})
	_, _ = s.db.Collection("page_changes").DeleteMany(s.ctx, bson.M{"monitoringId": objID})

	return nil
}

// GetSnapshots 获取页面快照
func (s *PageMonitoringService) GetSnapshots(monitoringID string, limit, offset int) ([]*models.PageSnapshot, int64, error) {
	// 验证ID
	objID, err := primitive.ObjectIDFromHex(monitoringID)
	if err != nil {
		return nil, 0, errors.New("无效的监控ID")
	}

	// 设置分页
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// 查询总数
	filter := bson.M{"monitoringId": objID}
	total, err := s.db.Collection("page_snapshots").CountDocuments(s.ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := s.db.Collection("page_snapshots").Find(s.ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(s.ctx)

	// 解析结果
	var snapshots []*models.PageSnapshot
	if err := cursor.All(s.ctx, &snapshots); err != nil {
		return nil, 0, err
	}

	return snapshots, total, nil
}

// GetChanges 获取页面变更记录
func (s *PageMonitoringService) GetChanges(monitoringID string, limit, offset int) ([]*models.PageChange, int64, error) {
	// 验证ID
	objID, err := primitive.ObjectIDFromHex(monitoringID)
	if err != nil {
		return nil, 0, errors.New("无效的监控ID")
	}

	// 设置分页
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// 查询总数
	filter := bson.M{"monitoringId": objID}
	total, err := s.db.Collection("page_changes").CountDocuments(s.ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	opts := options.Find().
		SetSort(bson.D{{Key: "changedAt", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := s.db.Collection("page_changes").Find(s.ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(s.ctx)

	// 解析结果
	var changes []*models.PageChange
	if err := cursor.All(s.ctx, &changes); err != nil {
		return nil, 0, err
	}

	return changes, total, nil
}

// CalculateContentHash 计算内容哈希
func (s *PageMonitoringService) CalculateContentHash(content string) string {
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}
