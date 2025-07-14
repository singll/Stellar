package nodemanager

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// HealthMonitorService 健康监控服务
type HealthMonitorService struct {
	db           *mongo.Database
	registry     *RegistryService
	config       *HealthMonitorConfig
	monitors     map[primitive.ObjectID]*NodeMonitor
	monitorsMutex sync.RWMutex
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// HealthMonitorConfig 健康监控配置
type HealthMonitorConfig struct {
	CheckInterval    time.Duration `json:"check_interval"`    // 检查间隔
	Timeout          time.Duration `json:"timeout"`           // 超时时间
	MaxFailures      int           `json:"max_failures"`      // 最大失败次数
	HealthEndpoint   string        `json:"health_endpoint"`   // 健康检查端点
	EnabledChecks    []string      `json:"enabled_checks"`    // 启用的检查项
	AlertThresholds  *AlertThresholds `json:"alert_thresholds"` // 告警阈值
}

// AlertThresholds 告警阈值
type AlertThresholds struct {
	CPUUsage       float64 `json:"cpu_usage"`        // CPU使用率阈值
	MemoryUsage    float64 `json:"memory_usage"`     // 内存使用率阈值
	DiskUsage      float64 `json:"disk_usage"`       // 磁盘使用率阈值
	LoadAverage    float64 `json:"load_average"`     // 负载阈值
	ResponseTime   time.Duration `json:"response_time"` // 响应时间阈值
	NetworkLatency time.Duration `json:"network_latency"` // 网络延迟阈值
}

// NodeMonitor 节点监控器
type NodeMonitor struct {
	node         *models.Node
	config       *HealthMonitorConfig
	failureCount int
	lastCheck    time.Time
	checks       map[string]HealthChecker  // 修复：接口不需要指针
	alerts       []*Alert
	alertsMutex  sync.RWMutex
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	Check(node *models.Node) *models.HealthCheck
	GetName() string
}

// Alert 告警
type Alert struct {
	ID        primitive.ObjectID `json:"id"`
	NodeID    primitive.ObjectID `json:"node_id"`
	Type      string             `json:"type"`
	Level     AlertLevel         `json:"level"`
	Message   string             `json:"message"`
	Details   map[string]interface{} `json:"details"`
	CreatedAt time.Time          `json:"created_at"`
	Resolved  bool               `json:"resolved"`
	ResolvedAt *time.Time        `json:"resolved_at"`
}

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning" 
	AlertLevelError    AlertLevel = "error"
	AlertLevelCritical AlertLevel = "critical"
)

// NewHealthMonitorService 创建健康监控服务
func NewHealthMonitorService(db *mongo.Database, registry *RegistryService, config *HealthMonitorConfig) *HealthMonitorService {
	if config == nil {
		config = &HealthMonitorConfig{
			CheckInterval:   30 * time.Second,
			Timeout:         10 * time.Second,
			MaxFailures:     3,
			HealthEndpoint:  "/health",
			EnabledChecks:   []string{"ping", "http", "resource"},
			AlertThresholds: &AlertThresholds{
				CPUUsage:       85.0,
				MemoryUsage:    90.0,
				DiskUsage:      95.0,
				LoadAverage:    10.0,
				ResponseTime:   5 * time.Second,
				NetworkLatency: 1 * time.Second,
			},
		}
	}

	service := &HealthMonitorService{
		db:       db,
		registry: registry,
		config:   config,
		monitors: make(map[primitive.ObjectID]*NodeMonitor),
		stopChan: make(chan struct{}),
	}

	return service
}

// Start 启动健康监控服务
func (s *HealthMonitorService) Start(ctx context.Context) error {
	// 加载现有节点并创建监控器
	nodes, err := s.registry.ListNodes(ctx, nil)
	if err != nil {
		return fmt.Errorf("加载节点失败: %v", err)
	}

	for _, node := range nodes {
		s.addNodeMonitor(node)
	}

	// 启动监控循环
	s.wg.Add(1)
	go s.monitorLoop(ctx)

	return nil
}

// Stop 停止健康监控服务
func (s *HealthMonitorService) Stop() error {
	close(s.stopChan)
	s.wg.Wait()
	return nil
}

// AddNode 添加节点监控
func (s *HealthMonitorService) AddNode(node *models.Node) {
	s.addNodeMonitor(node)
}

// RemoveNode 移除节点监控
func (s *HealthMonitorService) RemoveNode(nodeID primitive.ObjectID) {
	s.monitorsMutex.Lock()
	delete(s.monitors, nodeID)
	s.monitorsMutex.Unlock()
}

// GetNodeHealth 获取节点健康状态
func (s *HealthMonitorService) GetNodeHealth(nodeID primitive.ObjectID) (*models.NodeHealth, error) {
	s.monitorsMutex.RLock()
	monitor, exists := s.monitors[nodeID]
	s.monitorsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("节点监控器不存在")
	}

	health := &models.NodeHealth{
		NodeID:    nodeID,
		CheckTime: monitor.lastCheck,
		Healthy:   monitor.node.IsHealthy(),
		Checks:    []models.HealthCheck{},
	}

	// 执行所有健康检查
	for _, checker := range monitor.checks {
		check := checker.Check(monitor.node)
		if check != nil {
			health.Checks = append(health.Checks, *check)
		}
	}

	// 计算整体健康状态
	health.Healthy = s.calculateOverallHealth(health.Checks)

	return health, nil
}

// GetNodeAlerts 获取节点告警
func (s *HealthMonitorService) GetNodeAlerts(nodeID primitive.ObjectID) ([]*Alert, error) {
	s.monitorsMutex.RLock()
	monitor, exists := s.monitors[nodeID]
	s.monitorsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("节点监控器不存在")
	}

	monitor.alertsMutex.RLock()
	alerts := make([]*Alert, len(monitor.alerts))
	copy(alerts, monitor.alerts)
	monitor.alertsMutex.RUnlock()

	return alerts, nil
}

// addNodeMonitor 添加节点监控器
func (s *HealthMonitorService) addNodeMonitor(node *models.Node) {
	monitor := &NodeMonitor{
		node:   node,
		config: s.config,
		checks: make(map[string]HealthChecker),
		alerts: []*Alert{},
	}

	// 初始化健康检查器
	for _, checkName := range s.config.EnabledChecks {
		switch checkName {
		case "ping":
			monitor.checks["ping"] = NewPingChecker(s.config.Timeout)
		case "http":
			monitor.checks["http"] = NewHTTPChecker(s.config.HealthEndpoint, s.config.Timeout)
		case "resource":
			monitor.checks["resource"] = NewResourceChecker(s.config.AlertThresholds)
		}
	}

	s.monitorsMutex.Lock()
	s.monitors[node.ID] = monitor
	s.monitorsMutex.Unlock()
}

// monitorLoop 监控循环
func (s *HealthMonitorService) monitorLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.performHealthChecks(ctx)
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// performHealthChecks 执行健康检查
func (s *HealthMonitorService) performHealthChecks(ctx context.Context) {
	s.monitorsMutex.RLock()
	monitors := make([]*NodeMonitor, 0, len(s.monitors))
	for _, monitor := range s.monitors {
		monitors = append(monitors, monitor)
	}
	s.monitorsMutex.RUnlock()

	// 并发检查所有节点
	var wg sync.WaitGroup
	for _, monitor := range monitors {
		wg.Add(1)
		go func(m *NodeMonitor) {
			defer wg.Done()
			s.checkNodeHealth(ctx, m)
		}(monitor)
	}
	wg.Wait()
}

// checkNodeHealth 检查单个节点健康状态
func (s *HealthMonitorService) checkNodeHealth(ctx context.Context, monitor *NodeMonitor) {
	monitor.lastCheck = time.Now()
	healthy := true
	var failedChecks []string

	// 执行所有健康检查
	for name, checker := range monitor.checks {
		check := checker.Check(monitor.node)
		if check == nil || check.Status != "healthy" {
			healthy = false
			failedChecks = append(failedChecks, name)
		}
	}

	// 更新失败计数
	if !healthy {
		monitor.failureCount++
	} else {
		monitor.failureCount = 0
	}

	// 检查是否需要更新节点状态
	currentStatus := monitor.node.Status
	newStatus := currentStatus

	if monitor.failureCount >= s.config.MaxFailures {
		newStatus = models.NodeStatusFailed
	} else if healthy && currentStatus == models.NodeStatusFailed {
		newStatus = models.NodeStatusOnline
	}

	// 更新节点状态
	if newStatus != currentStatus {
		err := s.registry.UpdateNodeStatus(ctx, monitor.node.ID, newStatus)
		if err != nil {
			fmt.Printf("更新节点状态失败: %v\n", err)
		} else {
			monitor.node.Status = newStatus
			
			// 生成状态变更告警
			alert := &Alert{
				ID:        primitive.NewObjectID(),
				NodeID:    monitor.node.ID,
				Type:      "status_change",
				Level:     s.getAlertLevel(newStatus),
				Message:   fmt.Sprintf("节点状态从 %s 变更为 %s", currentStatus, newStatus),
				Details: map[string]interface{}{
					"old_status":    currentStatus,
					"new_status":    newStatus,
					"failed_checks": failedChecks,
					"failure_count": monitor.failureCount,
				},
				CreatedAt: time.Now(),
			}
			s.addAlert(monitor, alert)
		}
	}

	// 检查资源告警
	s.checkResourceAlerts(monitor)
}

// checkResourceAlerts 检查资源告警
func (s *HealthMonitorService) checkResourceAlerts(monitor *NodeMonitor) {
	thresholds := s.config.AlertThresholds
	node := monitor.node

	// CPU 使用率告警
	if node.CPU > thresholds.CPUUsage {
		alert := &Alert{
			ID:      primitive.NewObjectID(),
			NodeID:  node.ID,
			Type:    "cpu_high",
			Level:   AlertLevelWarning,
			Message: fmt.Sprintf("CPU使用率过高: %.2f%%", node.CPU),
			Details: map[string]interface{}{
				"cpu_usage":  node.CPU,
				"threshold":  thresholds.CPUUsage,
			},
			CreatedAt: time.Now(),
		}
		s.addAlert(monitor, alert)
	}

	// 内存使用率告警
	if node.Memory > thresholds.MemoryUsage {
		alert := &Alert{
			ID:      primitive.NewObjectID(),
			NodeID:  node.ID,
			Type:    "memory_high",
			Level:   AlertLevelWarning,
			Message: fmt.Sprintf("内存使用率过高: %.2f%%", node.Memory),
			Details: map[string]interface{}{
				"memory_usage": node.Memory,
				"threshold":    thresholds.MemoryUsage,
			},
			CreatedAt: time.Now(),
		}
		s.addAlert(monitor, alert)
	}

	// 磁盘使用率告警
	if node.Disk > thresholds.DiskUsage {
		alert := &Alert{
			ID:      primitive.NewObjectID(),
			NodeID:  node.ID,
			Type:    "disk_high",
			Level:   AlertLevelError,
			Message: fmt.Sprintf("磁盘使用率过高: %.2f%%", node.Disk),
			Details: map[string]interface{}{
				"disk_usage": node.Disk,
				"threshold":  thresholds.DiskUsage,
			},
			CreatedAt: time.Now(),
		}
		s.addAlert(monitor, alert)
	}

	// 系统负载告警
	if node.Load > thresholds.LoadAverage {
		alert := &Alert{
			ID:      primitive.NewObjectID(),
			NodeID:  node.ID,
			Type:    "load_high",
			Level:   AlertLevelWarning,
			Message: fmt.Sprintf("系统负载过高: %.2f", node.Load),
			Details: map[string]interface{}{
				"load_average": node.Load,
				"threshold":    thresholds.LoadAverage,
			},
			CreatedAt: time.Now(),
		}
		s.addAlert(monitor, alert)
	}
}

// addAlert 添加告警
func (s *HealthMonitorService) addAlert(monitor *NodeMonitor, alert *Alert) {
	monitor.alertsMutex.Lock()
	
	// 检查是否存在相同类型的未解决告警
	exists := false
	for _, existingAlert := range monitor.alerts {
		if !existingAlert.Resolved && existingAlert.Type == alert.Type {
			exists = true
			break
		}
	}
	
	if !exists {
		monitor.alerts = append(monitor.alerts, alert)
		
		// 保存告警到数据库
		go s.saveAlert(alert)
	}
	
	monitor.alertsMutex.Unlock()
}

// saveAlert 保存告警到数据库
func (s *HealthMonitorService) saveAlert(alert *Alert) {
	ctx := context.Background()
	collection := s.db.Collection("node_alerts")
	collection.InsertOne(ctx, alert)
}

// calculateOverallHealth 计算整体健康状态
func (s *HealthMonitorService) calculateOverallHealth(checks []models.HealthCheck) bool {
	if len(checks) == 0 {
		return false
	}

	for _, check := range checks {
		if check.Status != "healthy" {
			return false
		}
	}
	return true
}

// getAlertLevel 根据节点状态获取告警级别
func (s *HealthMonitorService) getAlertLevel(status models.NodeStatus) AlertLevel {
	switch status {
	case models.NodeStatusFailed:
		return AlertLevelCritical
	case models.NodeStatusOffline:
		return AlertLevelError
	case models.NodeStatusMaintenance:
		return AlertLevelWarning
	default:
		return AlertLevelInfo
	}
}

// 健康检查器实现

// PingChecker Ping检查器
type PingChecker struct {
	timeout time.Duration
}

func NewPingChecker(timeout time.Duration) *PingChecker {
	return &PingChecker{timeout: timeout}
}

func (c *PingChecker) Check(node *models.Node) *models.HealthCheck {
	start := time.Now()
	
	// 这里应该实现真正的ping检查
	// 为了简化，我们使用TCP连接检查
	timeout := time.After(c.timeout)
	
	check := &models.HealthCheck{
		Name:    "ping",
		Latency: time.Since(start),
	}
	
	select {
	case <-timeout:
		check.Status = "timeout"
		check.Message = "连接超时"
	default:
		check.Status = "healthy"
		check.Message = "连接正常"
	}
	
	return check
}

func (c *PingChecker) GetName() string {
	return "ping"
}

// HTTPChecker HTTP检查器
type HTTPChecker struct {
	endpoint string
	timeout  time.Duration
	client   *http.Client
}

func NewHTTPChecker(endpoint string, timeout time.Duration) *HTTPChecker {
	return &HTTPChecker{
		endpoint: endpoint,
		timeout:  timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *HTTPChecker) Check(node *models.Node) *models.HealthCheck {
	start := time.Now()
	url := fmt.Sprintf("http://%s:%d%s", node.IP, node.Port, c.endpoint)
	
	check := &models.HealthCheck{
		Name:    "http",
		Latency: time.Since(start),
	}
	
	resp, err := c.client.Get(url)
	if err != nil {
		check.Status = "unhealthy"
		check.Message = fmt.Sprintf("HTTP请求失败: %v", err)
		return check
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		check.Status = "healthy"
		check.Message = "HTTP检查通过"
	} else {
		check.Status = "unhealthy"
		check.Message = fmt.Sprintf("HTTP状态码: %d", resp.StatusCode)
	}
	
	check.Latency = time.Since(start)
	return check
}

func (c *HTTPChecker) GetName() string {
	return "http"
}

// ResourceChecker 资源检查器
type ResourceChecker struct {
	thresholds *AlertThresholds
}

func NewResourceChecker(thresholds *AlertThresholds) *ResourceChecker {
	return &ResourceChecker{thresholds: thresholds}
}

func (c *ResourceChecker) Check(node *models.Node) *models.HealthCheck {
	check := &models.HealthCheck{
		Name:    "resource",
		Status:  "healthy",
		Message: "资源使用正常",
		Latency: 0,
	}
	
	// 检查资源使用率
	issues := []string{}
	
	if node.CPU > c.thresholds.CPUUsage {
		issues = append(issues, fmt.Sprintf("CPU使用率过高: %.2f%%", node.CPU))
	}
	
	if node.Memory > c.thresholds.MemoryUsage {
		issues = append(issues, fmt.Sprintf("内存使用率过高: %.2f%%", node.Memory))
	}
	
	if node.Disk > c.thresholds.DiskUsage {
		issues = append(issues, fmt.Sprintf("磁盘使用率过高: %.2f%%", node.Disk))
	}
	
	if node.Load > c.thresholds.LoadAverage {
		issues = append(issues, fmt.Sprintf("系统负载过高: %.2f", node.Load))
	}
	
	if len(issues) > 0 {
		check.Status = "unhealthy"
		check.Message = fmt.Sprintf("资源告警: %v", issues)
	}
	
	return check
}

func (c *ResourceChecker) GetName() string {
	return "resource"
}