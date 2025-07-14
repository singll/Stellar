package nodemanager

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// Redis键前缀
	nodeHeartbeatPrefix = "node:heartbeat:"
	nodeStatusPrefix    = "node:status:"
	nodeConfigPrefix    = "node:config:"

	// 默认配置
	defaultHeartbeatInterval = 30 // 默认心跳间隔(秒)
	defaultHeartbeatTimeout  = 90 // 默认心跳超时(秒)
)

// NodeManager 节点管理器
type NodeManager struct {
	db          *mongo.Database
	redisClient *redis.Client
	nodeRepo    models.NodeRepository
	nodes       map[string]*models.Node
	nodesMutex  sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	eventChan   chan NodeEvent
	config      NodeManagerConfig
}

// NodeManagerConfig 节点管理器配置
type NodeManagerConfig struct {
	HeartbeatInterval int  // 心跳间隔(秒)
	HeartbeatTimeout  int  // 心跳超时(秒)
	EnableAutoRemove  bool // 是否自动移除离线节点
	AutoRemoveAfter   int  // 自动移除离线节点的时间(秒)
}

// NodeEvent 节点事件
type NodeEvent struct {
	Type      string      // 事件类型
	NodeID    string      // 节点ID
	Timestamp time.Time   // 时间戳
	Data      interface{} // 事件数据
}

// NodeEventType 节点事件类型
const (
	NodeEventRegister     = "register"      // 节点注册
	NodeEventHeartbeat    = "heartbeat"     // 节点心跳
	NodeEventOffline      = "offline"       // 节点离线
	NodeEventOnline       = "online"        // 节点上线
	NodeEventRemoved      = "removed"       // 节点移除
	NodeEventConfigUpdate = "config_update" // 节点配置更新
)

// NewNodeManager 创建节点管理器
func NewNodeManager(db *mongo.Database, redisClient *redis.Client, config NodeManagerConfig) *NodeManager {
	// 设置默认配置
	if config.HeartbeatInterval <= 0 {
		config.HeartbeatInterval = defaultHeartbeatInterval
	}
	if config.HeartbeatTimeout <= 0 {
		config.HeartbeatTimeout = defaultHeartbeatTimeout
	}
	if config.AutoRemoveAfter <= 0 {
		config.AutoRemoveAfter = 86400 // 默认1天
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	// 创建节点仓库
	nodeRepo := models.NewNodeRepository(db)

	return &NodeManager{
		db:          db,
		redisClient: redisClient,
		nodeRepo:    nodeRepo,
		nodes:       make(map[string]*models.Node),
		ctx:         ctx,
		cancel:      cancel,
		eventChan:   make(chan NodeEvent, 100),
		config:      config,
	}
}

// Start 启动节点管理器
func (m *NodeManager) Start() error {
	// 加载所有节点
	if err := m.loadNodes(); err != nil {
		return err
	}

	// 启动心跳检测
	go m.heartbeatChecker()

	// 启动事件处理
	go m.eventHandler()

	return nil
}

// Stop 停止节点管理器
func (m *NodeManager) Stop() {
	m.cancel()
	close(m.eventChan)
}

// RegisterNode 注册节点
func (m *NodeManager) RegisterNode(req models.NodeRegistrationRequest) (*models.NodeRegistrationResponse, error) {
	// 验证请求
	if req.Name == "" {
		return nil, errors.New("节点名称不能为空")
	}
	if req.IP == "" {
		return nil, errors.New("节点IP不能为空")
	}
	if req.Port <= 0 {
		return nil, errors.New("节点端口无效")
	}
	if req.Role == "" {
		req.Role = models.NodeRoleWorker // 默认为工作节点
	}

	// 检查节点名称是否已存在
	_, err := m.nodeRepo.GetByName(context.Background(), req.Name)
	if err == nil {
		return nil, errors.New("节点名称已存在")
	}

	// 生成API密钥
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, err
	}

	// 创建节点
	node := models.Node{
		Name:              req.Name,
		Type:              models.NodeTypeWorker, // 根据req.Role设置，这里暂时用默认值
		Status:            models.NodeStatusRegisting,
		IP:                req.IP,
		Port:              req.Port,
		ApiKey:            apiKey,
		RegisterTime:      time.Now(),
		LastHeartbeatTime: time.Now(),
		Tags:              req.Tags,
	}

	// 设置默认配置
	if req.Config.MaxConcurrentTasks <= 0 {
		req.Config.MaxConcurrentTasks = 10
	}
	if req.Config.HeartbeatInterval <= 0 {
		req.Config.HeartbeatInterval = m.config.HeartbeatInterval
	}
	node.Config = req.Config

	// 保存节点到数据库
	err = m.nodeRepo.Create(context.Background(), &node)
	if err != nil {
		return nil, err
	}

	// 添加到内存中
	m.nodesMutex.Lock()
	m.nodes[node.ID.Hex()] = &node
	m.nodesMutex.Unlock()

	// 发送事件
	m.eventChan <- NodeEvent{
		Type:      NodeEventRegister,
		NodeID:    node.ID.Hex(),
		Timestamp: time.Now(),
		Data:      node,
	}

	// 返回注册响应
	return &models.NodeRegistrationResponse{
		NodeID:  node.ID.Hex(),
		ApiKey:  apiKey,
		Status:  "success",
		Message: "节点注册成功",
	}, nil
}

// UpdateNodeStatus 更新节点状态
func (m *NodeManager) UpdateNodeStatus(nodeID string, heartbeat models.NodeHeartbeat) error {
	// 获取节点
	m.nodesMutex.RLock()
	node, exists := m.nodes[nodeID]
	m.nodesMutex.RUnlock()

	if !exists {
		return errors.New("节点不存在")
	}

	// 更新节点状态
	now := time.Now()

	// 更新Redis中的心跳记录
	heartbeatKey := nodeHeartbeatPrefix + nodeID
	heartbeatData, _ := bson.Marshal(heartbeat)
	err := m.redisClient.Set(m.ctx, heartbeatKey, heartbeatData, time.Duration(m.config.HeartbeatTimeout)*time.Second).Err()
	if err != nil {
		return err
	}

	// 更新节点状态
	wasOffline := node.Status == models.NodeStatusOffline

	// 更新节点状态信息
	nodeStatus := models.NodeStatusInfo{
		CpuUsage:       heartbeat.CpuUsage,
		MemoryUsage:    int64(heartbeat.MemoryUsage),
		RunningTasks:   heartbeat.RunningTasks,
		QueuedTasks:    heartbeat.QueuedTasks,
		LastUpdateTime: now,
	}

	// 更新数据库
	err = m.nodeRepo.UpdateStatus(context.Background(), nodeID, string(models.NodeStatusOnline))
	if err != nil {
		return err
	}

	err = m.nodeRepo.UpdateLastHeartbeat(context.Background(), nodeID, now)
	if err != nil {
		return err
	}

	err = m.nodeRepo.UpdateNodeStatus(context.Background(), nodeID, models.NodeStatusOnline)
	if err != nil {
		return err
	}

	// 更新内存中的节点信息
	m.nodesMutex.Lock()
	node.Status = models.NodeStatusOnline
	node.LastHeartbeatTime = now
	node.StatusInfo = nodeStatus
	m.nodesMutex.Unlock()

	// 如果节点之前是离线状态，现在是在线状态，发送上线事件
	if wasOffline {
		m.eventChan <- NodeEvent{
			Type:      NodeEventOnline,
			NodeID:    nodeID,
			Timestamp: now,
			Data:      node,
		}
	}

	// 发送心跳事件
	m.eventChan <- NodeEvent{
		Type:      NodeEventHeartbeat,
		NodeID:    nodeID,
		Timestamp: now,
		Data:      heartbeat,
	}

	return nil
}

// UpdateNodeConfig 更新节点配置
func (m *NodeManager) UpdateNodeConfig(nodeID string, config models.NodeConfig) error {
	// 获取节点
	m.nodesMutex.RLock()
	node, exists := m.nodes[nodeID]
	m.nodesMutex.RUnlock()

	if !exists {
		return errors.New("节点不存在")
	}

	// 更新节点配置
	m.nodesMutex.Lock()
	node.Config = config
	m.nodesMutex.Unlock()

	// 更新数据库
	err := m.nodeRepo.UpdateConfig(context.Background(), nodeID, config)
	if err != nil {
		return err
	}

	// 更新Redis中的配置
	configKey := nodeConfigPrefix + nodeID
	configData, _ := bson.Marshal(config)
	err = m.redisClient.Set(m.ctx, configKey, configData, 0).Err()
	if err != nil {
		return err
	}

	// 发送配置更新事件
	m.eventChan <- NodeEvent{
		Type:      NodeEventConfigUpdate,
		NodeID:    nodeID,
		Timestamp: time.Now(),
		Data:      config,
	}

	return nil
}

// GetNode 获取节点
func (m *NodeManager) GetNode(nodeID string) (*models.Node, error) {
	m.nodesMutex.RLock()
	node, exists := m.nodes[nodeID]
	m.nodesMutex.RUnlock()

	if !exists {
		// 尝试从数据库获取
		dbNode, err := m.nodeRepo.GetByID(context.Background(), nodeID)
		if err != nil {
			return nil, err
		}
		
		// 添加到内存中
		m.nodesMutex.Lock()
		m.nodes[nodeID] = dbNode
		m.nodesMutex.Unlock()
		
		return dbNode, nil
	}

	return node, nil
}

// GetAllNodes 获取所有节点
func (m *NodeManager) GetAllNodes() []*models.Node {
	m.nodesMutex.RLock()
	defer m.nodesMutex.RUnlock()

	nodes := make([]*models.Node, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodes = append(nodes, node)
	}

	return nodes
}

// GetNodesByRole 根据角色获取节点
func (m *NodeManager) GetNodesByRole(role string) []*models.Node {
	m.nodesMutex.RLock()
	defer m.nodesMutex.RUnlock()

	var nodes []*models.Node
	for _, node := range m.nodes {
		if node.Type == models.NodeType(role) {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// GetNodesByStatus 根据状态获取节点
func (m *NodeManager) GetNodesByStatus(status string) []*models.Node {
	m.nodesMutex.RLock()
	defer m.nodesMutex.RUnlock()

	var nodes []*models.Node
	for _, node := range m.nodes {
		if node.Status == models.NodeStatus(status) {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// RemoveNode 移除节点
func (m *NodeManager) RemoveNode(nodeID string) error {
	// 从数据库中删除节点
	err := m.nodeRepo.Delete(context.Background(), nodeID)
	if err != nil {
		return err
	}

	// 从内存中删除节点
	m.nodesMutex.Lock()
	delete(m.nodes, nodeID)
	m.nodesMutex.Unlock()

	// 从Redis中删除节点相关数据
	heartbeatKey := nodeHeartbeatPrefix + nodeID
	configKey := nodeConfigPrefix + nodeID
	statusKey := nodeStatusPrefix + nodeID

	pipeline := m.redisClient.Pipeline()
	pipeline.Del(m.ctx, heartbeatKey)
	pipeline.Del(m.ctx, configKey)
	pipeline.Del(m.ctx, statusKey)
	_, err = pipeline.Exec(m.ctx)
	if err != nil {
		return err
	}

	// 发送节点移除事件
	m.eventChan <- NodeEvent{
		Type:      NodeEventRemoved,
		NodeID:    nodeID,
		Timestamp: time.Now(),
	}

	return nil
}

// loadNodes 从数据库加载所有节点
func (m *NodeManager) loadNodes() error {
	// 查询所有节点
	params := models.NodeQueryParams{
		Page:     1,
		PageSize: 1000, // 加载所有节点
	}
	
	nodes, _, err := m.nodeRepo.List(context.Background(), params)
	if err != nil {
		return err
	}

	// 清空当前节点
	m.nodesMutex.Lock()
	m.nodes = make(map[string]*models.Node)
	
	// 加载节点到内存
	for _, node := range nodes {
		m.nodes[node.ID.Hex()] = node
	}
	m.nodesMutex.Unlock()

	return nil
}

// heartbeatChecker 心跳检测器
func (m *NodeManager) heartbeatChecker() {
	ticker := time.NewTicker(time.Duration(m.config.HeartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkNodeHeartbeats()
		}
	}
}

// checkNodeHeartbeats 检查节点心跳
func (m *NodeManager) checkNodeHeartbeats() {
	now := time.Now()
	timeout := time.Duration(m.config.HeartbeatTimeout) * time.Second
	autoRemoveTime := time.Duration(m.config.AutoRemoveAfter) * time.Second

	m.nodesMutex.Lock()
	defer m.nodesMutex.Unlock()

	for id, node := range m.nodes {
		// 检查心跳超时
		if node.Status == models.NodeStatusOnline && now.Sub(node.LastHeartbeatTime) > timeout {
			// 更新节点状态为离线
			node.Status = models.NodeStatusOffline

			// 更新数据库
			err := m.nodeRepo.UpdateStatus(context.Background(), id, string(models.NodeStatusOffline))
			if err != nil {
				continue
			}

			// 发送离线事件
			m.eventChan <- NodeEvent{
				Type:      NodeEventOffline,
				NodeID:    id,
				Timestamp: now,
				Data:      node,
			}
		}

		// 自动移除长时间离线的节点
		if m.config.EnableAutoRemove && node.Status == models.NodeStatusOffline && now.Sub(node.LastHeartbeatTime) > autoRemoveTime {
			// 从数据库中删除节点
			err := m.nodeRepo.Delete(context.Background(), id)
			if err != nil {
				continue
			}

			// 从内存中删除节点
			delete(m.nodes, id)

			// 从Redis中删除节点相关数据
			heartbeatKey := nodeHeartbeatPrefix + id
			configKey := nodeConfigPrefix + id
			statusKey := nodeStatusPrefix + id

			pipeline := m.redisClient.Pipeline()
			pipeline.Del(m.ctx, heartbeatKey)
			pipeline.Del(m.ctx, configKey)
			pipeline.Del(m.ctx, statusKey)
			_, _ = pipeline.Exec(m.ctx)

			// 发送节点移除事件
			m.eventChan <- NodeEvent{
				Type:      NodeEventRemoved,
				NodeID:    id,
				Timestamp: now,
			}
		}
	}
}

// eventHandler 事件处理器
func (m *NodeManager) eventHandler() {
	for event := range m.eventChan {
		// 处理事件
		switch event.Type {
		case NodeEventRegister:
			// 处理节点注册事件
			m.handleRegisterEvent(event)
		case NodeEventHeartbeat:
			// 处理节点心跳事件
			m.handleHeartbeatEvent(event)
		case NodeEventOffline:
			// 处理节点离线事件
			m.handleOfflineEvent(event)
		case NodeEventOnline:
			// 处理节点上线事件
			m.handleOnlineEvent(event)
		case NodeEventRemoved:
			// 处理节点移除事件
			m.handleRemovedEvent(event)
		case NodeEventConfigUpdate:
			// 处理节点配置更新事件
			m.handleConfigUpdateEvent(event)
		}
	}
}

// handleRegisterEvent 处理节点注册事件
func (m *NodeManager) handleRegisterEvent(event NodeEvent) {
	// 可以在这里添加额外的处理逻辑，如通知其他系统
	fmt.Printf("节点注册: %s\n", event.NodeID)
}

// handleHeartbeatEvent 处理节点心跳事件
func (m *NodeManager) handleHeartbeatEvent(event NodeEvent) {
	// 可以在这里添加额外的处理逻辑，如更新监控系统
}

// handleOfflineEvent 处理节点离线事件
func (m *NodeManager) handleOfflineEvent(event NodeEvent) {
	fmt.Printf("节点离线: %s\n", event.NodeID)
	// 可以在这里添加额外的处理逻辑，如通知管理员
}

// handleOnlineEvent 处理节点上线事件
func (m *NodeManager) handleOnlineEvent(event NodeEvent) {
	fmt.Printf("节点上线: %s\n", event.NodeID)
	// 可以在这里添加额外的处理逻辑，如重新分配任务
}

// handleRemovedEvent 处理节点移除事件
func (m *NodeManager) handleRemovedEvent(event NodeEvent) {
	fmt.Printf("节点移除: %s\n", event.NodeID)
	// 可以在这里添加额外的处理逻辑，如清理相关资源
}

// handleConfigUpdateEvent 处理节点配置更新事件
func (m *NodeManager) handleConfigUpdateEvent(event NodeEvent) {
	fmt.Printf("节点配置更新: %s\n", event.NodeID)
	// 可以在这里添加额外的处理逻辑，如通知节点更新配置
}

// generateAPIKey 生成API密钥
func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
