package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NodeRepository 节点仓库接口
type NodeRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, node *Node) error
	GetByID(ctx context.Context, id string) (*Node, error)
	GetByName(ctx context.Context, name string) (*Node, error)
	Update(ctx context.Context, node *Node) error
	Delete(ctx context.Context, id string) error

	// 查询操作
	List(ctx context.Context, params NodeQueryParams) ([]*Node, int64, error)
	GetByStatus(ctx context.Context, status string) ([]*Node, error)
	GetByRole(ctx context.Context, role string) ([]*Node, error)
	GetByTags(ctx context.Context, tags []string) ([]*Node, error)

	// 状态操作
	UpdateStatus(ctx context.Context, id string, status string) error
	UpdateLastHeartbeat(ctx context.Context, id string, timestamp time.Time) error
	UpdateNodeStatus(ctx context.Context, id string, nodeStatus NodeStatus) error
	UpdateConfig(ctx context.Context, id string, config NodeConfig) error

	// 统计操作
	GetStats(ctx context.Context) (*NodeStats, error)
	GetNodeTaskStats(ctx context.Context, nodeId string) (*NodeTaskStats, error)
	UpdateTaskStats(ctx context.Context, nodeId string, stats NodeTaskStats) error

	// 批量操作
	BatchUpdateStatus(ctx context.Context, ids []string, status string) error
	BatchDelete(ctx context.Context, ids []string) error

	// 清理操作
	CleanupOfflineNodes(ctx context.Context, timeout time.Duration) (int64, error)
}

// NodeQueryParams 节点查询参数
type NodeQueryParams struct {
	Status     string   `json:"status,omitempty"`
	Role       string   `json:"role,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Search     string   `json:"search,omitempty"`
	Page       int      `json:"page,omitempty"`
	PageSize   int      `json:"pageSize,omitempty"`
	SortBy     string   `json:"sortBy,omitempty"`
	SortDesc   bool     `json:"sortDesc,omitempty"`
	OnlineOnly bool     `json:"onlineOnly,omitempty"`
}

// NodeStats 节点统计信息
type NodeStats struct {
	Total            int                    `json:"total"`
	Online           int                    `json:"online"`
	Offline          int                    `json:"offline"`
	Disabled         int                    `json:"disabled"`
	Maintaining      int                    `json:"maintaining"`
	ByRole           map[string]int         `json:"byRole"`
	ByStatus         map[string]int         `json:"byStatus"`
	TotalTasks       int                    `json:"totalTasks"`
	RunningTasks     int                    `json:"runningTasks"`
	QueuedTasks      int                    `json:"queuedTasks"`
	AvgCpuUsage      float64                `json:"avgCpuUsage"`
	AvgMemoryUsage   float64                `json:"avgMemoryUsage"`
	ResourceUsage    NodeResourceUsageStats `json:"resourceUsage"`
	LastUpdateTime   time.Time              `json:"lastUpdateTime"`
}

// NodeResourceUsageStats 节点资源使用统计
type NodeResourceUsageStats struct {
	TotalMemory    int64   `json:"totalMemory"`    // 总内存 (MB)
	UsedMemory     int64   `json:"usedMemory"`     // 已用内存 (MB)
	TotalDisk      int64   `json:"totalDisk"`      // 总磁盘 (MB)
	UsedDisk       int64   `json:"usedDisk"`       // 已用磁盘 (MB)
	AvgNetworkIn   int64   `json:"avgNetworkIn"`   // 平均网络入流量 (KB/s)
	AvgNetworkOut  int64   `json:"avgNetworkOut"`  // 平均网络出流量 (KB/s)
	HighCpuNodes   int     `json:"highCpuNodes"`   // 高CPU使用率节点数
	HighMemNodes   int     `json:"highMemNodes"`   // 高内存使用率节点数
	OverloadedNodes int    `json:"overloadedNodes"` // 过载节点数
}

// nodeRepository 节点仓库实现
type nodeRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewNodeRepository 创建节点仓库
func NewNodeRepository(db *mongo.Database) NodeRepository {
	return &nodeRepository{
		db:         db,
		collection: db.Collection("nodes"),
	}
}

// Create 创建节点
func (r *nodeRepository) Create(ctx context.Context, node *Node) error {
	if node.ID.IsZero() {
		node.ID = primitive.NewObjectID()
	}
	
	now := time.Now()
	if node.RegisteredAt.IsZero() {
		node.RegisteredAt = now
	}
	if node.LastHeartbeat.IsZero() {
		node.LastHeartbeat = now
	}

	_, err := r.collection.InsertOne(ctx, node)
	return err
}

// GetByID 根据ID获取节点
func (r *nodeRepository) GetByID(ctx context.Context, id string) (*Node, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("无效的节点ID: %w", err)
	}

	var node Node
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&node)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("节点不存在")
		}
		return nil, err
	}

	return &node, nil
}

// GetByName 根据名称获取节点
func (r *nodeRepository) GetByName(ctx context.Context, name string) (*Node, error) {
	var node Node
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&node)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("节点不存在")
		}
		return nil, err
	}

	return &node, nil
}

// Update 更新节点
func (r *nodeRepository) Update(ctx context.Context, node *Node) error {
	objID, err := primitive.ObjectIDFromHex(node.ID.Hex())
	if err != nil {
		return fmt.Errorf("无效的节点ID: %w", err)
	}

	_, err = r.collection.ReplaceOne(ctx, bson.M{"_id": objID}, node)
	return err
}

// Delete 删除节点
func (r *nodeRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("无效的节点ID: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("节点不存在")
	}

	return nil
}

// List 分页查询节点列表
func (r *nodeRepository) List(ctx context.Context, params NodeQueryParams) ([]*Node, int64, error) {
	// 构建过滤条件
	filter := bson.M{}

	if params.Status != "" {
		filter["status"] = params.Status
	}

	if params.Role != "" {
		filter["role"] = params.Role
	}

	if len(params.Tags) > 0 {
		filter["tags"] = bson.M{"$in": params.Tags}
	}

	if params.Search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": params.Search, "$options": "i"}},
			{"ip": bson.M{"$regex": params.Search, "$options": "i"}},
		}
	}

	if params.OnlineOnly {
		filter["status"] = NodeStatusOnline
	}

	// 设置分页
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	skip := int64((params.Page - 1) * params.PageSize)
	limit := int64(params.PageSize)

	// 设置排序
	sort := bson.D{{Key: "registerTime", Value: -1}}
	if params.SortBy != "" {
		sortValue := 1
		if params.SortDesc {
			sortValue = -1
		}
		sort = bson.D{{Key: params.SortBy, Value: sortValue}}
	}

	// 查询总数
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(sort)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var nodes []*Node
	for cursor.Next(ctx) {
		var node Node
		if err := cursor.Decode(&node); err != nil {
			return nil, 0, err
		}
		nodes = append(nodes, &node)
	}

	return nodes, total, nil
}

// GetByStatus 根据状态获取节点
func (r *nodeRepository) GetByStatus(ctx context.Context, status string) ([]*Node, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": status})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var nodes []*Node
	for cursor.Next(ctx) {
		var node Node
		if err := cursor.Decode(&node); err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}

	return nodes, nil
}

// GetByRole 根据角色获取节点
func (r *nodeRepository) GetByRole(ctx context.Context, role string) ([]*Node, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"role": role})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var nodes []*Node
	for cursor.Next(ctx) {
		var node Node
		if err := cursor.Decode(&node); err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}

	return nodes, nil
}

// GetByTags 根据标签获取节点
func (r *nodeRepository) GetByTags(ctx context.Context, tags []string) ([]*Node, error) {
	filter := bson.M{"tags": bson.M{"$in": tags}}
	
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var nodes []*Node
	for cursor.Next(ctx) {
		var node Node
		if err := cursor.Decode(&node); err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}

	return nodes, nil
}

// UpdateStatus 更新节点状态
func (r *nodeRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("无效的节点ID: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("节点不存在")
	}

	return nil
}

// UpdateLastHeartbeat 更新最后心跳时间
func (r *nodeRepository) UpdateLastHeartbeat(ctx context.Context, id string, timestamp time.Time) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("无效的节点ID: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"lastHeartbeatTime": timestamp,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("节点不存在")
	}

	return nil
}

// UpdateNodeStatus 更新节点状态信息
func (r *nodeRepository) UpdateNodeStatus(ctx context.Context, id string, nodeStatus NodeStatus) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("无效的节点ID: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"nodeStatus": nodeStatus,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("节点不存在")
	}

	return nil
}

// UpdateConfig 更新节点配置
func (r *nodeRepository) UpdateConfig(ctx context.Context, id string, config NodeConfig) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("无效的节点ID: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"config": config,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("节点不存在")
	}

	return nil
}

// GetStats 获取节点统计信息
func (r *nodeRepository) GetStats(ctx context.Context) (*NodeStats, error) {
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": nil,
				"total": bson.M{"$sum": 1},
				"online": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$status", NodeStatusOnline}},
							"then": 1,
							"else": 0,
						},
					},
				},
				"offline": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$status", NodeStatusOffline}},
							"then": 1,
							"else": 0,
						},
					},
				},
				"disabled": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$status", NodeStatusDisabled}},
							"then": 1,
							"else": 0,
						},
					},
				},
				"maintaining": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$status", NodeStatusMaintain}},
							"then": 1,
							"else": 0,
						},
					},
				},
				"totalTasks":    bson.M{"$sum": "$nodeStatus.runningTasks"},
				"runningTasks":  bson.M{"$sum": "$nodeStatus.runningTasks"},
				"queuedTasks":   bson.M{"$sum": "$nodeStatus.queuedTasks"},
				"avgCpuUsage":   bson.M{"$avg": "$nodeStatus.cpuUsage"},
				"avgMemoryUsage": bson.M{"$avg": "$nodeStatus.memoryUsage"},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result struct {
		Total         int     `bson:"total"`
		Online        int     `bson:"online"`
		Offline       int     `bson:"offline"`
		Disabled      int     `bson:"disabled"`
		Maintaining   int     `bson:"maintaining"`
		TotalTasks    int     `bson:"totalTasks"`
		RunningTasks  int     `bson:"runningTasks"`
		QueuedTasks   int     `bson:"queuedTasks"`
		AvgCpuUsage   float64 `bson:"avgCpuUsage"`
		AvgMemoryUsage float64 `bson:"avgMemoryUsage"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
	}

	// 获取按角色和状态的统计
	byRole, err := r.getStatsByField(ctx, "role")
	if err != nil {
		return nil, err
	}

	byStatus, err := r.getStatsByField(ctx, "status")
	if err != nil {
		return nil, err
	}

	stats := &NodeStats{
		Total:          result.Total,
		Online:         result.Online,
		Offline:        result.Offline,
		Disabled:       result.Disabled,
		Maintaining:    result.Maintaining,
		ByRole:         byRole,
		ByStatus:       byStatus,
		TotalTasks:     result.TotalTasks,
		RunningTasks:   result.RunningTasks,
		QueuedTasks:    result.QueuedTasks,
		AvgCpuUsage:    result.AvgCpuUsage,
		AvgMemoryUsage: result.AvgMemoryUsage,
		LastUpdateTime: time.Now(),
	}

	return stats, nil
}

// getStatsByField 根据字段获取统计信息
func (r *nodeRepository) getStatsByField(ctx context.Context, field string) (map[string]int, error) {
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$" + field,
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]int)
	for cursor.Next(ctx) {
		var item struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		result[item.ID] = item.Count
	}

	return result, nil
}

// GetNodeTaskStats 获取节点任务统计
func (r *nodeRepository) GetNodeTaskStats(ctx context.Context, nodeId string) (*NodeTaskStats, error) {
	objID, err := primitive.ObjectIDFromHex(nodeId)
	if err != nil {
		return nil, fmt.Errorf("无效的节点ID: %w", err)
	}

	var node Node
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&node)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("节点不存在")
		}
		return nil, err
	}

	// 构建TaskStats
	taskStats := NodeTaskStats{
		TotalTasks:     int(node.ActiveTasks + node.CompletedTasks + node.FailedTasks),
		SuccessTasks:   int(node.CompletedTasks),
		FailedTasks:    int(node.FailedTasks),
		TaskTypeStats:  map[string]int{}, // 简化实现
		AvgExecuteTime: 0,                // 需要计算
		LastTaskTime:   node.LastUpdate,
	}

	return &taskStats, nil
}

// UpdateTaskStats 更新节点任务统计
func (r *nodeRepository) UpdateTaskStats(ctx context.Context, nodeId string, stats NodeTaskStats) error {
	objID, err := primitive.ObjectIDFromHex(nodeId)
	if err != nil {
		return fmt.Errorf("无效的节点ID: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"taskStats": stats,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("节点不存在")
	}

	return nil
}

// BatchUpdateStatus 批量更新节点状态
func (r *nodeRepository) BatchUpdateStatus(ctx context.Context, ids []string, status string) error {
	objIDs := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		objIDs = append(objIDs, objID)
	}

	if len(objIDs) == 0 {
		return errors.New("没有有效的节点ID")
	}

	filter := bson.M{"_id": bson.M{"$in": objIDs}}
	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}

// BatchDelete 批量删除节点
func (r *nodeRepository) BatchDelete(ctx context.Context, ids []string) error {
	objIDs := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		objIDs = append(objIDs, objID)
	}

	if len(objIDs) == 0 {
		return errors.New("没有有效的节点ID")
	}

	filter := bson.M{"_id": bson.M{"$in": objIDs}}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

// CleanupOfflineNodes 清理离线节点
func (r *nodeRepository) CleanupOfflineNodes(ctx context.Context, timeout time.Duration) (int64, error) {
	cutoff := time.Now().Add(-timeout)
	filter := bson.M{
		"status": NodeStatusOffline,
		"lastHeartbeatTime": bson.M{"$lt": cutoff},
	}

	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}