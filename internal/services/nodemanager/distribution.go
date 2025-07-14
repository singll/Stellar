package nodemanager

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TaskDistributionService 任务分发服务
type TaskDistributionService struct {
	db            *mongo.Database
	registry      *RegistryService
	loadBalancer  LoadBalancer
	taskQueue     *TaskQueue
	config        *DistributionConfig
	metrics       *DistributionMetrics
	metricsMutex  sync.RWMutex
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// DistributionConfig 分发配置
type DistributionConfig struct {
	Strategy           LoadBalancingStrategy `json:"strategy"`            // 负载均衡策略
	MaxTasksPerNode    int                   `json:"max_tasks_per_node"`  // 每个节点最大任务数
	TaskTimeout        time.Duration         `json:"task_timeout"`        // 任务超时时间
	RetryPolicy        *RetryPolicy          `json:"retry_policy"`        // 重试策略
	PriorityEnabled    bool                  `json:"priority_enabled"`    // 是否启用优先级
	AffinityRules      []*AffinityRule       `json:"affinity_rules"`      // 亲和性规则
	ResourceRequirements *ResourceRequirements `json:"resource_requirements"` // 资源需求
}

// LoadBalancingStrategy 负载均衡策略
type LoadBalancingStrategy string

const (
	StrategyRoundRobin    LoadBalancingStrategy = "round_robin"    // 轮询
	StrategyLeastLoaded   LoadBalancingStrategy = "least_loaded"   // 最少负载
	StrategyWeighted      LoadBalancingStrategy = "weighted"       // 加权
	StrategyRandom        LoadBalancingStrategy = "random"         // 随机
	StrategyHash          LoadBalancingStrategy = "hash"           // 哈希
	StrategyCapability    LoadBalancingStrategy = "capability"     // 能力匹配
)

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`     // 最大重试次数
	InitialDelay  time.Duration `json:"initial_delay"`   // 初始延迟
	MaxDelay      time.Duration `json:"max_delay"`       // 最大延迟
	BackoffFactor float64       `json:"backoff_factor"`  // 退避因子
}

// AffinityRule 亲和性规则
type AffinityRule struct {
	TaskType     string            `json:"task_type"`     // 任务类型
	NodeSelector map[string]string `json:"node_selector"` // 节点选择器
	Weight       int               `json:"weight"`        // 权重
	Required     bool              `json:"required"`      // 是否必需
}

// ResourceRequirements 资源需求
type ResourceRequirements struct {
	MinCPU    float64 `json:"min_cpu"`    // 最小CPU
	MinMemory float64 `json:"min_memory"` // 最小内存
	MinDisk   float64 `json:"min_disk"`   // 最小磁盘
}

// TaskQueue 任务队列
type TaskQueue struct {
	tasks        []*models.NodeTask
	priorityTasks []*models.NodeTask
	mutex        sync.RWMutex
	maxSize      int
	condition    *sync.Cond
}

// DistributionMetrics 分发指标
type DistributionMetrics struct {
	TotalTasks       int64            `json:"total_tasks"`
	CompletedTasks   int64            `json:"completed_tasks"`
	FailedTasks      int64            `json:"failed_tasks"`
	RetiredTasks     int64            `json:"retried_tasks"`
	AverageWaitTime  time.Duration    `json:"average_wait_time"`
	NodeDistribution map[string]int64 `json:"node_distribution"`
	StrategyStats    map[string]int64 `json:"strategy_stats"`
}

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	SelectNode(nodes []*models.Node, task *models.NodeTask) (*models.Node, error)
	GetName() string
	UpdateMetrics(nodeID primitive.ObjectID, success bool, duration time.Duration)
}

// NewTaskDistributionService 创建任务分发服务
func NewTaskDistributionService(db *mongo.Database, registry *RegistryService, config *DistributionConfig) *TaskDistributionService {
	if config == nil {
		config = &DistributionConfig{
			Strategy:        StrategyLeastLoaded,
			MaxTasksPerNode: 10,
			TaskTimeout:     5 * time.Minute,
			RetryPolicy: &RetryPolicy{
				MaxRetries:    3,
				InitialDelay:  1 * time.Second,
				MaxDelay:      30 * time.Second,
				BackoffFactor: 2.0,
			},
			PriorityEnabled: true,
		}
	}

	service := &TaskDistributionService{
		db:       db,
		registry: registry,
		config:   config,
		taskQueue: NewTaskQueue(1000),
		metrics: &DistributionMetrics{
			NodeDistribution: make(map[string]int64),
			StrategyStats:    make(map[string]int64),
		},
		stopChan: make(chan struct{}),
	}

	// 初始化负载均衡器
	service.loadBalancer = service.createLoadBalancer()

	return service
}

// Start 启动任务分发服务
func (s *TaskDistributionService) Start(ctx context.Context) error {
	// 启动任务分发循环
	s.wg.Add(1)
	go s.distributionLoop(ctx)

	// 启动任务监控循环
	s.wg.Add(1)
	go s.monitoringLoop(ctx)

	return nil
}

// Stop 停止任务分发服务
func (s *TaskDistributionService) Stop() error {
	close(s.stopChan)
	s.wg.Wait()
	return nil
}

// SubmitTask 提交任务
func (s *TaskDistributionService) SubmitTask(task *models.NodeTask) error {
	if task.Priority > 0 && s.config.PriorityEnabled {
		return s.taskQueue.EnqueuePriority(task)
	}
	return s.taskQueue.Enqueue(task)
}

// DistributeTask 分发任务
func (s *TaskDistributionService) DistributeTask(ctx context.Context, task *models.NodeTask) error {
	// 获取可用节点
	nodes, err := s.getAvailableNodes(ctx, task)
	if err != nil {
		return fmt.Errorf("获取可用节点失败: %v", err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("没有可用节点")
	}

	// 使用负载均衡器选择节点
	selectedNode, err := s.loadBalancer.SelectNode(nodes, task)
	if err != nil {
		return fmt.Errorf("选择节点失败: %v", err)
	}

	// 分配任务到节点
	err = s.assignTaskToNode(ctx, task, selectedNode)
	if err != nil {
		return fmt.Errorf("分配任务失败: %v", err)
	}

	// 更新指标
	s.updateMetrics(selectedNode.ID, task)

	return nil
}

// getAvailableNodes 获取可用节点
func (s *TaskDistributionService) getAvailableNodes(ctx context.Context, task *models.NodeTask) ([]*models.Node, error) {
	// 获取健康节点
	allNodes, err := s.registry.GetHealthyNodes(ctx)
	if err != nil {
		return nil, err
	}

	var availableNodes []*models.Node

	for _, node := range allNodes {
		// 检查节点是否可以接受任务
		if !node.CanAcceptTasks() {
			continue
		}

		// 检查任务数限制
		if node.ActiveTasks >= int64(s.config.MaxTasksPerNode) {
			continue
		}

		// 检查能力匹配
		if !s.checkCapabilityMatch(node, task) {
			continue
		}

		// 检查资源需求
		if !s.checkResourceRequirements(node, task) {
			continue
		}

		// 检查亲和性规则
		if !s.checkAffinityRules(node, task) {
			continue
		}

		availableNodes = append(availableNodes, node)
	}

	return availableNodes, nil
}

// checkCapabilityMatch 检查能力匹配
func (s *TaskDistributionService) checkCapabilityMatch(node *models.Node, task *models.NodeTask) bool {
	// 如果任务没有特定能力要求，任何节点都可以处理
	requiredCapability, exists := task.Payload["required_capability"]
	if !exists {
		return true
	}

	capability, ok := requiredCapability.(string)
	if !ok {
		return true
	}

	return node.HasCapability(capability)
}

// checkResourceRequirements 检查资源需求
func (s *TaskDistributionService) checkResourceRequirements(node *models.Node, task *models.NodeTask) bool {
	if s.config.ResourceRequirements == nil {
		return true
	}

	req := s.config.ResourceRequirements

	// 检查CPU
	if req.MinCPU > 0 && node.CPU > (100-req.MinCPU) {
		return false
	}

	// 检查内存
	if req.MinMemory > 0 && node.Memory > (100-req.MinMemory) {
		return false
	}

	// 检查磁盘
	if req.MinDisk > 0 && node.Disk > (100-req.MinDisk) {
		return false
	}

	return true
}

// checkAffinityRules 检查亲和性规则
func (s *TaskDistributionService) checkAffinityRules(node *models.Node, task *models.NodeTask) bool {
	if len(s.config.AffinityRules) == 0 {
		return true
	}

	// 查找匹配的亲和性规则
	var matchingRules []*AffinityRule
	for _, rule := range s.config.AffinityRules {
		if rule.TaskType == "" || rule.TaskType == task.TaskType {
			matchingRules = append(matchingRules, rule)
		}
	}

	if len(matchingRules) == 0 {
		return true
	}

	// 检查必需规则
	for _, rule := range matchingRules {
		if rule.Required {
			if !s.matchNodeSelector(node, rule.NodeSelector) {
				return false
			}
		}
	}

	return true
}

// matchNodeSelector 匹配节点选择器
func (s *TaskDistributionService) matchNodeSelector(node *models.Node, selector map[string]string) bool {
	for key, value := range selector {
		switch key {
		case "region":
			if node.Region != value {
				return false
			}
		case "zone":
			if node.Zone != value {
				return false
			}
		case "type":
			if string(node.Type) != value {
				return false
			}
		default:
			// 检查元数据
			if nodeValue, exists := node.Metadata[key]; !exists || nodeValue != value {
				return false
			}
		}
	}
	return true
}

// assignTaskToNode 分配任务到节点
func (s *TaskDistributionService) assignTaskToNode(ctx context.Context, task *models.NodeTask, node *models.Node) error {
	// 设置任务分配信息
	task.NodeID = node.ID
	task.Status = models.TaskStatusPending
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	
	if s.config.TaskTimeout > 0 {
		timeout := time.Now().Add(s.config.TaskTimeout)
		task.TimeoutAt = &timeout
	}

	// 保存到数据库
	collection := s.db.Collection("node_tasks")
	_, err := collection.InsertOne(ctx, task)
	if err != nil {
		return fmt.Errorf("保存任务失败: %v", err)
	}

	// 更新节点活跃任务数
	update := bson.M{
		"$inc": bson.M{"active_tasks": 1},
		"$set": bson.M{"updated_at": time.Now()},
	}
	
	nodeCollection := s.db.Collection("nodes")
	_, err = nodeCollection.UpdateOne(ctx, bson.M{"_id": node.ID}, update)
	if err != nil {
		return fmt.Errorf("更新节点状态失败: %v", err)
	}

	return nil
}

// createLoadBalancer 创建负载均衡器
func (s *TaskDistributionService) createLoadBalancer() LoadBalancer {
	switch s.config.Strategy {
	case StrategyRoundRobin:
		return NewRoundRobinBalancer()
	case StrategyLeastLoaded:
		return NewLeastLoadedBalancer()
	case StrategyWeighted:
		return NewWeightedBalancer()
	case StrategyRandom:
		return NewRandomBalancer()
	case StrategyHash:
		return NewHashBalancer()
	case StrategyCapability:
		return NewCapabilityBalancer()
	default:
		return NewLeastLoadedBalancer()
	}
}

// updateMetrics 更新指标
func (s *TaskDistributionService) updateMetrics(nodeID primitive.ObjectID, task *models.NodeTask) {
	s.metricsMutex.Lock()
	defer s.metricsMutex.Unlock()

	s.metrics.TotalTasks++
	s.metrics.NodeDistribution[nodeID.Hex()]++
	s.metrics.StrategyStats[string(s.config.Strategy)]++
}

// distributionLoop 分发循环
func (s *TaskDistributionService) distributionLoop(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		default:
			// 获取待分发的任务
			task := s.taskQueue.Dequeue()
			if task == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 分发任务
			err := s.DistributeTask(ctx, task)
			if err != nil {
				fmt.Printf("分发任务失败: %v\n", err)
				
				// 重试逻辑
				if task.RetryCount < s.config.RetryPolicy.MaxRetries {
					task.RetryCount++
					
					// 计算退避延迟
					delay := s.calculateBackoffDelay(task.RetryCount)
					time.Sleep(delay)
					
					// 重新入队
					s.taskQueue.Enqueue(task)
					
					s.metricsMutex.Lock()
					s.metrics.RetiredTasks++
					s.metricsMutex.Unlock()
				} else {
					// 标记为失败
					task.Status = models.TaskStatusFailed
					task.Error = err.Error()
					
					s.metricsMutex.Lock()
					s.metrics.FailedTasks++
					s.metricsMutex.Unlock()
				}
			}
		}
	}
}

// monitoringLoop 监控循环
func (s *TaskDistributionService) monitoringLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkTimeoutTasks(ctx)
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// checkTimeoutTasks 检查超时任务
func (s *TaskDistributionService) checkTimeoutTasks(ctx context.Context) {
	collection := s.db.Collection("node_tasks")
	now := time.Now()

	// 查找超时任务
	filter := bson.M{
		"status": models.TaskStatusRunning,
		"timeout_at": bson.M{"$lt": now},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var task models.NodeTask
		if err := cursor.Decode(&task); err != nil {
			continue
		}

		// 标记为超时
		update := bson.M{
			"$set": bson.M{
				"status":      models.TaskStatusTimeout,
				"updated_at":  now,
				"completed_at": now,
			},
		}

		collection.UpdateOne(ctx, bson.M{"_id": task.ID}, update)

		// 减少节点活跃任务数
		nodeCollection := s.db.Collection("nodes")
		nodeUpdate := bson.M{
			"$inc": bson.M{"active_tasks": -1},
			"$set": bson.M{"updated_at": now},
		}
		nodeCollection.UpdateOne(ctx, bson.M{"_id": task.NodeID}, nodeUpdate)
	}
}

// calculateBackoffDelay 计算退避延迟
func (s *TaskDistributionService) calculateBackoffDelay(retryCount int) time.Duration {
	policy := s.config.RetryPolicy
	delay := policy.InitialDelay
	
	for i := 1; i < retryCount; i++ {
		delay = time.Duration(float64(delay) * policy.BackoffFactor)
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
			break
		}
	}
	
	return delay
}

// GetMetrics 获取分发指标
func (s *TaskDistributionService) GetMetrics() *DistributionMetrics {
	s.metricsMutex.RLock()
	defer s.metricsMutex.RUnlock()

	// 深拷贝指标
	metrics := &DistributionMetrics{
		TotalTasks:       s.metrics.TotalTasks,
		CompletedTasks:   s.metrics.CompletedTasks,
		FailedTasks:      s.metrics.FailedTasks,
		RetiredTasks:     s.metrics.RetiredTasks,
		AverageWaitTime:  s.metrics.AverageWaitTime,
		NodeDistribution: make(map[string]int64),
		StrategyStats:    make(map[string]int64),
	}

	for k, v := range s.metrics.NodeDistribution {
		metrics.NodeDistribution[k] = v
	}

	for k, v := range s.metrics.StrategyStats {
		metrics.StrategyStats[k] = v
	}

	return metrics
}

// NewTaskQueue 创建任务队列
func NewTaskQueue(maxSize int) *TaskQueue {
	tq := &TaskQueue{
		tasks:        make([]*models.NodeTask, 0),
		priorityTasks: make([]*models.NodeTask, 0),
		maxSize:      maxSize,
	}
	tq.condition = sync.NewCond(&tq.mutex)
	return tq
}

// Enqueue 入队普通任务
func (tq *TaskQueue) Enqueue(task *models.NodeTask) error {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	if len(tq.tasks)+len(tq.priorityTasks) >= tq.maxSize {
		return fmt.Errorf("任务队列已满")
	}

	tq.tasks = append(tq.tasks, task)
	tq.condition.Signal()
	return nil
}

// EnqueuePriority 入队优先任务
func (tq *TaskQueue) EnqueuePriority(task *models.NodeTask) error {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	if len(tq.tasks)+len(tq.priorityTasks) >= tq.maxSize {
		return fmt.Errorf("任务队列已满")
	}

	// 按优先级插入
	inserted := false
	for i, existingTask := range tq.priorityTasks {
		if task.Priority > existingTask.Priority {
			tq.priorityTasks = append(tq.priorityTasks[:i], 
				append([]*models.NodeTask{task}, tq.priorityTasks[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		tq.priorityTasks = append(tq.priorityTasks, task)
	}

	tq.condition.Signal()
	return nil
}

// Dequeue 出队任务
func (tq *TaskQueue) Dequeue() *models.NodeTask {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	// 优先处理高优先级任务
	if len(tq.priorityTasks) > 0 {
		task := tq.priorityTasks[0]
		tq.priorityTasks = tq.priorityTasks[1:]
		return task
	}

	// 处理普通任务
	if len(tq.tasks) > 0 {
		task := tq.tasks[0]
		tq.tasks = tq.tasks[1:]
		return task
	}

	return nil
}

// Size 获取队列大小
func (tq *TaskQueue) Size() int {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()
	return len(tq.tasks) + len(tq.priorityTasks)
}

// 负载均衡器实现

// RoundRobinBalancer 轮询负载均衡器
type RoundRobinBalancer struct {
	counter int64
	mutex   sync.Mutex
}

func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{}
}

func (b *RoundRobinBalancer) SelectNode(nodes []*models.Node, task *models.NodeTask) (*models.Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("没有可用节点")
	}

	b.mutex.Lock()
	index := int(b.counter % int64(len(nodes)))
	b.counter++
	b.mutex.Unlock()

	return nodes[index], nil
}

func (b *RoundRobinBalancer) GetName() string {
	return "round_robin"
}

func (b *RoundRobinBalancer) UpdateMetrics(nodeID primitive.ObjectID, success bool, duration time.Duration) {
	// 轮询策略不需要更新指标
}

// LeastLoadedBalancer 最少负载均衡器
type LeastLoadedBalancer struct{}

func NewLeastLoadedBalancer() *LeastLoadedBalancer {
	return &LeastLoadedBalancer{}
}

func (b *LeastLoadedBalancer) SelectNode(nodes []*models.Node, task *models.NodeTask) (*models.Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("没有可用节点")
	}

	// 按负载分数排序
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].GetLoadScore() < nodes[j].GetLoadScore()
	})

	return nodes[0], nil
}

func (b *LeastLoadedBalancer) GetName() string {
	return "least_loaded"
}

func (b *LeastLoadedBalancer) UpdateMetrics(nodeID primitive.ObjectID, success bool, duration time.Duration) {
	// 最少负载策略会实时计算，不需要存储历史指标
}

// RandomBalancer 随机负载均衡器
type RandomBalancer struct{}

func NewRandomBalancer() *RandomBalancer {
	return &RandomBalancer{}
}

func (b *RandomBalancer) SelectNode(nodes []*models.Node, task *models.NodeTask) (*models.Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("没有可用节点")
	}

	index := rand.Intn(len(nodes))
	return nodes[index], nil
}

func (b *RandomBalancer) GetName() string {
	return "random"
}

func (b *RandomBalancer) UpdateMetrics(nodeID primitive.ObjectID, success bool, duration time.Duration) {
	// 随机策略不需要更新指标
}

// WeightedBalancer 加权负载均衡器
type WeightedBalancer struct {
	weights map[string]int
	mutex   sync.RWMutex
}

func NewWeightedBalancer() *WeightedBalancer {
	return &WeightedBalancer{
		weights: make(map[string]int),
	}
}

func (b *WeightedBalancer) SelectNode(nodes []*models.Node, task *models.NodeTask) (*models.Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("没有可用节点")
	}

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	// 计算总权重
	totalWeight := 0
	for _, node := range nodes {
		weight := b.weights[node.ID.Hex()]
		if weight <= 0 {
			weight = 1 // 默认权重
		}
		totalWeight += weight
	}

	// 随机选择
	target := rand.Intn(totalWeight)
	current := 0

	for _, node := range nodes {
		weight := b.weights[node.ID.Hex()]
		if weight <= 0 {
			weight = 1
		}
		current += weight
		if current >= target {
			return node, nil
		}
	}

	// 兜底返回第一个节点
	return nodes[0], nil
}

func (b *WeightedBalancer) GetName() string {
	return "weighted"
}

func (b *WeightedBalancer) UpdateMetrics(nodeID primitive.ObjectID, success bool, duration time.Duration) {
	// 根据成功率动态调整权重
	b.mutex.Lock()
	defer b.mutex.Unlock()

	key := nodeID.Hex()
	if success {
		b.weights[key]++
	} else {
		if b.weights[key] > 1 {
			b.weights[key]--
		}
	}
}

// HashBalancer 哈希负载均衡器
type HashBalancer struct{}

func NewHashBalancer() *HashBalancer {
	return &HashBalancer{}
}

func (b *HashBalancer) SelectNode(nodes []*models.Node, task *models.NodeTask) (*models.Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("没有可用节点")
	}

	// 使用任务ID的哈希值选择节点
	hash := int(task.ID.Timestamp().Unix()) % len(nodes)
	return nodes[hash], nil
}

func (b *HashBalancer) GetName() string {
	return "hash"
}

func (b *HashBalancer) UpdateMetrics(nodeID primitive.ObjectID, success bool, duration time.Duration) {
	// 哈希策略不需要更新指标
}

// CapabilityBalancer 能力匹配负载均衡器
type CapabilityBalancer struct{}

func NewCapabilityBalancer() *CapabilityBalancer {
	return &CapabilityBalancer{}
}

func (b *CapabilityBalancer) SelectNode(nodes []*models.Node, task *models.NodeTask) (*models.Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("没有可用节点")
	}

	// 检查是否有特定能力要求
	if requiredCap, exists := task.Payload["required_capability"]; exists {
		if capability, ok := requiredCap.(string); ok {
			// 优先选择具有特定能力的节点
			for _, node := range nodes {
				if node.HasCapability(capability) {
					return node, nil
				}
			}
		}
	}

	// 如果没有特定要求，选择能力最多的节点
	sort.Slice(nodes, func(i, j int) bool {
		return len(nodes[i].Capabilities) > len(nodes[j].Capabilities)
	})

	return nodes[0], nil
}

func (b *CapabilityBalancer) GetName() string {
	return "capability"
}

func (b *CapabilityBalancer) UpdateMetrics(nodeID primitive.ObjectID, success bool, duration time.Duration) {
	// 能力匹配策略主要基于静态能力，不需要更新指标
}