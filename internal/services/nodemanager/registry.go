package nodemanager

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// RegistryService 节点注册服务
type RegistryService struct {
	db           *mongo.Database
	nodes        map[primitive.ObjectID]*models.Node // 内存中的节点缓存
	nodesMutex   sync.RWMutex
	eventChan    chan *models.NodeEvent
	stopChan     chan struct{}
	cleanupTimer *time.Timer
}

// NewRegistryService 创建节点注册服务
func NewRegistryService(db *mongo.Database) *RegistryService {
	service := &RegistryService{
		db:        db,
		nodes:     make(map[primitive.ObjectID]*models.Node),
		eventChan: make(chan *models.NodeEvent, 100),
		stopChan:  make(chan struct{}),
	}

	// 启动后台任务
	go service.backgroundTasks()
	
	// 从数据库加载现有节点
	service.loadExistingNodes()

	return service
}

// RegisterNode 注册新节点
func (s *RegistryService) RegisterNode(ctx context.Context, registration *models.NodeRegistration) (*models.Node, error) {
	// 验证注册信息
	if err := registration.ValidateRegistration(); err != nil {
		return nil, fmt.Errorf("节点注册验证失败: %v", err)
	}

	// 检查节点是否已存在
	existing, err := s.findNodeByAddress(ctx, registration.IP, registration.Port)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("检查现有节点失败: %v", err)
	}

	if existing != nil {
		return nil, fmt.Errorf("节点已存在: %s:%d", registration.IP, registration.Port)
	}

	// 生成节点密钥
	secret, err := s.generateSecret()
	if err != nil {
		return nil, fmt.Errorf("生成节点密钥失败: %v", err)
	}

	// 创建新节点
	now := time.Now()
	node := &models.Node{
		ID:           primitive.NewObjectID(),
		Name:         registration.Name,
		IP:           registration.IP,
		Port:         registration.Port,
		Type:         registration.Type,
		Status:       models.NodeStatusOnline,
		Version:      registration.Version,
		Capabilities: registration.Capabilities,
		Metadata:     registration.Metadata,
		Region:       registration.Region,
		Zone:         registration.Zone,
		Tags:         registration.Tags,
		Group:        registration.Group,
		Secret:       secret,
		RegisteredAt: now,
		LastHeartbeat: now,
		LastUpdate:   now,
		CreatedAt:    now,
		UpdatedAt:    now,
		Trusted:      false, // 新节点默认不受信任
	}

	// 保存到数据库
	collection := s.db.Collection("nodes")
	_, err = collection.InsertOne(ctx, node)
	if err != nil {
		return nil, fmt.Errorf("保存节点到数据库失败: %v", err)
	}

	// 添加到内存缓存
	s.nodesMutex.Lock()
	s.nodes[node.ID] = node
	s.nodesMutex.Unlock()

	// 发送注册事件
	s.emitEvent(&models.NodeEvent{
		NodeID:    node.ID,
		Type:      "node_registered",
		Message:   fmt.Sprintf("节点 %s 已注册", node.Name),
		Level:     "info",
		CreatedAt: time.Now(),
	})

	return node, nil
}

// UnregisterNode 注销节点
func (s *RegistryService) UnregisterNode(ctx context.Context, nodeID primitive.ObjectID) error {
	// 获取节点信息
	node, err := s.GetNode(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("获取节点信息失败: %v", err)
	}

	// 从数据库删除
	collection := s.db.Collection("nodes")
	_, err = collection.DeleteOne(ctx, bson.M{"_id": nodeID})
	if err != nil {
		return fmt.Errorf("从数据库删除节点失败: %v", err)
	}

	// 从内存缓存删除
	s.nodesMutex.Lock()
	delete(s.nodes, nodeID)
	s.nodesMutex.Unlock()

	// 发送注销事件
	s.emitEvent(&models.NodeEvent{
		NodeID:    nodeID,
		Type:      "node_unregistered",
		Message:   fmt.Sprintf("节点 %s 已注销", node.Name),
		Level:     "info",
		CreatedAt: time.Now(),
	})

	return nil
}

// UpdateNodeHeartbeat 更新节点心跳
func (s *RegistryService) UpdateNodeHeartbeat(ctx context.Context, heartbeat *models.NodeHeartbeat) error {
	// 验证心跳数据
	if err := heartbeat.Validate(); err != nil {
		return fmt.Errorf("心跳数据验证失败: %v", err)
	}

	// 获取节点
	node, err := s.GetNode(ctx, heartbeat.NodeID)
	if err != nil {
		return fmt.Errorf("获取节点信息失败: %v", err)
	}

	// 检查节点状态变化
	statusChanged := node.Status != heartbeat.Status

	// 更新节点信息
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":         heartbeat.Status,
			"cpu":            heartbeat.CPU,
			"memory":         heartbeat.Memory,
			"disk":           heartbeat.Disk,
			"network":        heartbeat.Network,
			"load":           heartbeat.Load,
			"active_tasks":   heartbeat.ActiveTasks,
			"last_heartbeat": now,
			"last_update":    now,
			"updated_at":     now,
		},
	}

	// 如果有元数据更新
	if heartbeat.Metadata != nil {
		for key, value := range heartbeat.Metadata {
			update["$set"].(bson.M)["metadata."+key] = value
		}
	}

	// 更新数据库
	collection := s.db.Collection("nodes")
	_, err = collection.UpdateOne(ctx, bson.M{"_id": heartbeat.NodeID}, update)
	if err != nil {
		return fmt.Errorf("更新数据库失败: %v", err)
	}

	// 更新内存缓存
	s.nodesMutex.Lock()
	if cachedNode, exists := s.nodes[heartbeat.NodeID]; exists {
		cachedNode.Status = heartbeat.Status
		cachedNode.CPU = heartbeat.CPU
		cachedNode.Memory = heartbeat.Memory
		cachedNode.Disk = heartbeat.Disk
		cachedNode.Network = heartbeat.Network
		cachedNode.Load = heartbeat.Load
		cachedNode.ActiveTasks = heartbeat.ActiveTasks
		cachedNode.LastHeartbeat = now
		cachedNode.LastUpdate = now
		cachedNode.UpdatedAt = now
		
		if heartbeat.Metadata != nil {
			if cachedNode.Metadata == nil {
				cachedNode.Metadata = make(map[string]string)
			}
			for key, value := range heartbeat.Metadata {
				cachedNode.Metadata[key] = value
			}
		}
	}
	s.nodesMutex.Unlock()

	// 如果状态发生变化，发送事件
	if statusChanged {
		s.emitEvent(&models.NodeEvent{
			NodeID:    heartbeat.NodeID,
			Type:      "node_status_changed",
			Message:   fmt.Sprintf("节点 %s 状态变更为 %s", node.Name, heartbeat.Status),
			Level:     s.getEventLevel(heartbeat.Status),
			CreatedAt: time.Now(),
		})
	}

	return nil
}

// GetNode 获取节点信息
func (s *RegistryService) GetNode(ctx context.Context, nodeID primitive.ObjectID) (*models.Node, error) {
	// 先从内存缓存查找
	s.nodesMutex.RLock()
	if node, exists := s.nodes[nodeID]; exists {
		s.nodesMutex.RUnlock()
		return node, nil
	}
	s.nodesMutex.RUnlock()

	// 从数据库查找
	collection := s.db.Collection("nodes")
	var node models.Node
	err := collection.FindOne(ctx, bson.M{"_id": nodeID}).Decode(&node)
	if err != nil {
		return nil, err
	}

	// 添加到内存缓存
	s.nodesMutex.Lock()
	s.nodes[nodeID] = &node
	s.nodesMutex.Unlock()

	return &node, nil
}

// ListNodes 列出所有节点
func (s *RegistryService) ListNodes(ctx context.Context, filters map[string]interface{}) ([]*models.Node, error) {
	// 构建查询条件
	query := bson.M{}
	if filters != nil {
		for key, value := range filters {
			query[key] = value
		}
	}

	collection := s.db.Collection("nodes")
	cursor, err := collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var nodes []*models.Node
	for cursor.Next(ctx) {
		var node models.Node
		if err := cursor.Decode(&node); err != nil {
			continue
		}
		nodes = append(nodes, &node)
	}

	return nodes, nil
}

// GetHealthyNodes 获取健康的节点
func (s *RegistryService) GetHealthyNodes(ctx context.Context) ([]*models.Node, error) {
	s.nodesMutex.RLock()
	defer s.nodesMutex.RUnlock()

	var healthyNodes []*models.Node
	for _, node := range s.nodes {
		if node.IsHealthy() {
			healthyNodes = append(healthyNodes, node)
		}
	}

	return healthyNodes, nil
}

// GetNodesByCapability 根据能力获取节点
func (s *RegistryService) GetNodesByCapability(ctx context.Context, capability string) ([]*models.Node, error) {
	s.nodesMutex.RLock()
	defer s.nodesMutex.RUnlock()

	var matchingNodes []*models.Node
	for _, node := range s.nodes {
		if node.HasCapability(capability) && node.CanAcceptTasks() {
			matchingNodes = append(matchingNodes, node)
		}
	}

	return matchingNodes, nil
}

// GetNodesByTag 根据标签获取节点
func (s *RegistryService) GetNodesByTag(ctx context.Context, tag string) ([]*models.Node, error) {
	filters := map[string]interface{}{
		"tags": bson.M{"$in": []string{tag}},
	}
	return s.ListNodes(ctx, filters)
}

// UpdateNodeStatus 更新节点状态
func (s *RegistryService) UpdateNodeStatus(ctx context.Context, nodeID primitive.ObjectID, status models.NodeStatus) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	collection := s.db.Collection("nodes")
	_, err := collection.UpdateOne(ctx, bson.M{"_id": nodeID}, update)
	if err != nil {
		return err
	}

	// 更新内存缓存
	s.nodesMutex.Lock()
	if node, exists := s.nodes[nodeID]; exists {
		node.Status = status
		node.UpdatedAt = time.Now()
	}
	s.nodesMutex.Unlock()

	return nil
}

// findNodeByAddress 根据地址查找节点
func (s *RegistryService) findNodeByAddress(ctx context.Context, ip string, port int) (*models.Node, error) {
	collection := s.db.Collection("nodes")
	var node models.Node
	err := collection.FindOne(ctx, bson.M{"ip": ip, "port": port}).Decode(&node)
	return &node, err
}

// generateSecret 生成节点密钥
func (s *RegistryService) generateSecret() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:]), nil
}

// loadExistingNodes 从数据库加载现有节点
func (s *RegistryService) loadExistingNodes() {
	ctx := context.Background()
	nodes, err := s.ListNodes(ctx, nil)
	if err != nil {
		return
	}

	s.nodesMutex.Lock()
	defer s.nodesMutex.Unlock()

	for _, node := range nodes {
		s.nodes[node.ID] = node
	}
}

// backgroundTasks 后台任务
func (s *RegistryService) backgroundTasks() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒执行一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupOfflineNodes()
		case event := <-s.eventChan:
			s.handleEvent(event)
		case <-s.stopChan:
			return
		}
	}
}

// cleanupOfflineNodes 清理离线节点
func (s *RegistryService) cleanupOfflineNodes() {
	ctx := context.Background()
	now := time.Now()
	offlineThreshold := 10 * time.Minute // 10分钟没有心跳视为离线

	s.nodesMutex.Lock()
	defer s.nodesMutex.Unlock()

	for nodeID, node := range s.nodes {
		if node.Status != models.NodeStatusOffline && 
		   now.Sub(node.LastHeartbeat) > offlineThreshold {
			
			// 标记为离线
			node.Status = models.NodeStatusOffline
			node.UpdatedAt = now

			// 更新数据库
			collection := s.db.Collection("nodes")
			update := bson.M{
				"$set": bson.M{
					"status":     models.NodeStatusOffline,
					"updated_at": now,
				},
			}
			collection.UpdateOne(ctx, bson.M{"_id": nodeID}, update)

			// 发送离线事件
			s.emitEvent(&models.NodeEvent{
				NodeID:    nodeID,
				Type:      "node_offline",
				Message:   fmt.Sprintf("节点 %s 已离线", node.Name),
				Level:     "warn",
				CreatedAt: time.Now(),
			})
		}
	}
}

// emitEvent 发送事件
func (s *RegistryService) emitEvent(event *models.NodeEvent) {
	event.ID = primitive.NewObjectID()
	select {
	case s.eventChan <- event:
	default:
		// 事件队列满时丢弃事件
	}
}

// handleEvent 处理事件
func (s *RegistryService) handleEvent(event *models.NodeEvent) {
	// 保存事件到数据库
	ctx := context.Background()
	collection := s.db.Collection("node_events")
	collection.InsertOne(ctx, event)
}

// getEventLevel 根据状态获取事件级别
func (s *RegistryService) getEventLevel(status models.NodeStatus) string {
	switch status {
	case models.NodeStatusOnline:
		return "info"
	case models.NodeStatusOffline, models.NodeStatusFailed:
		return "error"
	case models.NodeStatusMaintenance, models.NodeStatusDraining:
		return "warn"
	default:
		return "info"
	}
}

// GetNodeMetrics 获取节点指标
func (s *RegistryService) GetNodeMetrics(ctx context.Context, nodeID primitive.ObjectID) (*models.NodeMetrics, error) {
	node, err := s.GetNode(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	// 创建指标数据
	metrics := &models.NodeMetrics{
		NodeID:    nodeID,
		Timestamp: time.Now(),
		CPU: models.CPUMetrics{
			Usage:    node.CPU,
			LoadAvg1: node.Load,
		},
		Memory: models.MemoryMetrics{
			Usage: node.Memory,
		},
		Disk: models.DiskMetrics{
			Usage: node.Disk,
		},
		Network: models.NetworkMetrics{
			Usage: node.Network,
		},
		Tasks: models.TaskMetrics{
			Total:   node.ActiveTasks + node.CompletedTasks + node.FailedTasks,
			Running: node.ActiveTasks,
			Completed: node.CompletedTasks,
			Failed:  node.FailedTasks,
		},
	}

	// 计算成功率
	if metrics.Tasks.Total > 0 {
		metrics.Tasks.SuccessRate = float64(metrics.Tasks.Completed) / float64(metrics.Tasks.Total) * 100
	}

	return metrics, nil
}

// Close 关闭服务
func (s *RegistryService) Close() {
	close(s.stopChan)
	if s.cleanupTimer != nil {
		s.cleanupTimer.Stop()
	}
}