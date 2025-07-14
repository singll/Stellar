package models

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Node 分布式节点
type Node struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name         string             `json:"name"`
	IP           string             `json:"ip"`
	Port         int                `json:"port"`
	Type         NodeType           `json:"type"`
	Status       NodeStatus         `json:"status"`
	Version      string             `json:"version"`
	Capabilities []string           `json:"capabilities"`
	Metadata     map[string]string  `json:"metadata"`
	
	// 性能指标
	CPU          float64 `json:"cpu"`           // CPU使用率 (0-100)
	Memory       float64 `json:"memory"`        // 内存使用率 (0-100)
	Disk         float64 `json:"disk"`          // 磁盘使用率 (0-100)
	Network      float64 `json:"network"`       // 网络使用率 (0-100)
	Load         float64 `json:"load"`          // 系统负载
	
	// 任务统计
	ActiveTasks    int64 `json:"active_tasks"`    // 活跃任务数
	CompletedTasks int64 `json:"completed_tasks"` // 完成任务数
	FailedTasks    int64 `json:"failed_tasks"`    // 失败任务数
	
	// 时间戳
	RegisteredAt time.Time  `json:"registered_at"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	LastUpdate   time.Time  `json:"last_update"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	
	// 区域和分组
	Region      string   `json:"region"`
	Zone        string   `json:"zone"`
	Tags        []string `json:"tags"`
	Group       string   `json:"group"`
	
	// 安全配置
	Secret      string `json:"-"`        // 节点密钥，不返回给客户端
	CertHash    string `json:"cert_hash"` // 证书哈希
	Trusted     bool   `json:"trusted"`   // 是否可信节点
	
	// TODO: 以下字段根据DEV_PLAN 0.3版本需要完善（向后兼容）
	ApiKey            string           `json:"api_key" bson:"api_key"`                         // API密钥
	RegisterTime      time.Time        `json:"register_time" bson:"register_time"`             // 注册时间
	LastHeartbeatTime time.Time        `json:"last_heartbeat_time" bson:"last_heartbeat_time"` // 最后心跳时间
	Config            NodeConfig       `json:"config" bson:"config"`                           // 节点配置
	StatusInfo        NodeStatusInfo   `json:"status_info" bson:"status_info"`                 // 状态信息
	TaskStats         NodeTaskStats    `json:"task_stats" bson:"task_stats"`                   // 任务统计
}

// NodeType 节点类型
type NodeType string

const (
	NodeTypeWorker     NodeType = "worker"     // 工作节点
	NodeTypeManager    NodeType = "manager"    // 管理节点
	NodeTypeScheduler  NodeType = "scheduler"  // 调度节点
	NodeTypeMonitor    NodeType = "monitor"    // 监控节点
	NodeTypeMixed      NodeType = "mixed"      // 混合节点
)

// NodeStatus 节点状态 
type NodeStatus string

const (
	NodeStatusOnline      NodeStatus = "online"      // 在线
	NodeStatusOffline     NodeStatus = "offline"     // 离线
	NodeStatusMaintenance NodeStatus = "maintenance" // 维护中
	NodeStatusDraining    NodeStatus = "draining"    // 排空中
	NodeStatusFailed      NodeStatus = "failed"      // 失败
	NodeStatusUnknown     NodeStatus = "unknown"     // 未知
	NodeStatusDisabled    NodeStatus = "disabled"    // 禁用
	NodeStatusRegisting   NodeStatus = "registing"   // 注册中
)

// NodeStatusInfo 节点状态详细信息
type NodeStatusInfo struct {
	CpuUsage       float64   `json:"cpuUsage"`
	MemoryUsage    int64     `json:"memoryUsage"`
	DiskUsage      int64     `json:"diskUsage"`
	LoadAverage    []float64 `json:"loadAverage"`
	RunningTasks   int       `json:"runningTasks"`
	QueuedTasks    int       `json:"queuedTasks"`
	NetworkIn      int64     `json:"networkIn"`
	NetworkOut     int64     `json:"networkOut"`
	UptimeSeconds  int64     `json:"uptimeSeconds"`
	LastUpdateTime time.Time `json:"lastUpdateTime"`
}

// 向后兼容的常量定义
const (
	NodeStatusMaintain  = string(NodeStatusMaintenance) // 维护中 - 保持兼容
	NodeRoleMaster     = string(NodeTypeManager)        // 主节点 - 保持兼容
	NodeRoleSlave      = string(NodeTypeWorker)         // 从节点 - 保持兼容
	NodeRoleWorker     = string(NodeTypeWorker)         // 工作节点 - 保持兼容
)

// NodeRegistration 节点注册请求
type NodeRegistration struct {
	Name         string            `json:"name"`
	IP           string            `json:"ip"`
	Port         int               `json:"port"`
	Type         NodeType          `json:"type"`
	Version      string            `json:"version"`
	Capabilities []string          `json:"capabilities"`
	Metadata     map[string]string `json:"metadata"`
	Region       string            `json:"region"`
	Zone         string            `json:"zone"`
	Tags         []string          `json:"tags"`
	Group        string            `json:"group"`
	Secret       string            `json:"secret"`
}

// NodeHeartbeat 节点心跳
type NodeHeartbeat struct {
	NodeID       primitive.ObjectID `json:"node_id"`
	Status       NodeStatus         `json:"status"`
	CPU          float64            `json:"cpu"`
	Memory       float64            `json:"memory"`
	Disk         float64            `json:"disk"`
	Network      float64            `json:"network"`
	Load         float64            `json:"load"`
	ActiveTasks  int64              `json:"active_tasks"`
	Metadata     map[string]string  `json:"metadata"`
	Timestamp    time.Time          `json:"timestamp"`
	// TODO: 以下字段根据DEV_PLAN 0.3版本需要完善（向后兼容）
	CpuUsage      float64 `json:"cpuUsage" bson:"cpuUsage"`           // CPU使用率
	MemoryUsage   float64 `json:"memoryUsage" bson:"memoryUsage"`     // 内存使用率
	RunningTasks  int     `json:"runningTasks" bson:"runningTasks"`   // 运行中任务数
	QueuedTasks   int     `json:"queuedTasks" bson:"queuedTasks"`     // 排队任务数
}

// NodeTask 节点任务
type NodeTask struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	NodeID       primitive.ObjectID `json:"node_id"`
	TaskType     string             `json:"task_type"`
	TaskID       primitive.ObjectID `json:"task_id"`
	Status       TaskStatus         `json:"status"`
	Priority     int                `json:"priority"`
	RetryCount   int                `json:"retry_count"`
	MaxRetries   int                `json:"max_retries"`
	Payload      map[string]interface{} `json:"payload"`
	Result       map[string]interface{} `json:"result"`
	Error        string             `json:"error"`
	StartedAt    *time.Time         `json:"started_at"`
	CompletedAt  *time.Time         `json:"completed_at"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	TimeoutAt    *time.Time         `json:"timeout_at"`
}

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"    // 待执行
	TaskStatusQueued     TaskStatus = "queued"     // 已入队
	TaskStatusRunning    TaskStatus = "running"    // 执行中
	TaskStatusCompleted  TaskStatus = "completed"  // 已完成
	TaskStatusFailed     TaskStatus = "failed"     // 失败
	TaskStatusTimeout    TaskStatus = "timeout"    // 超时
	TaskStatusCancelled  TaskStatus = "cancelled"  // 已取消
	TaskStatusCanceled   TaskStatus = "canceled"   // 已取消 (美式拼写)
	TaskStatusPaused     TaskStatus = "paused"     // 暂停
)

// NodeCluster 节点集群
type NodeCluster struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	NodeIDs     []primitive.ObjectID `json:"node_ids"`
	Config      ClusterConfig      `json:"config"`
	Status      ClusterStatus      `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// ClusterConfig 集群配置
type ClusterConfig struct {
	LoadBalancing    string            `json:"load_balancing"`    // round_robin, least_loaded, hash
	FailoverEnabled  bool              `json:"failover_enabled"`
	HealthCheckURL   string            `json:"health_check_url"`
	HeartbeatInterval time.Duration    `json:"heartbeat_interval"`
	Metadata         map[string]string `json:"metadata"`
}

// ClusterStatus 集群状态
type ClusterStatus string

const (
	ClusterStatusActive   ClusterStatus = "active"   // 活跃
	ClusterStatusInactive ClusterStatus = "inactive" // 非活跃
	ClusterStatusDegraded ClusterStatus = "degraded" // 降级
	ClusterStatusFailed   ClusterStatus = "failed"   // 失败
)

// NodeHealth 节点健康状态
type NodeHealth struct {
	NodeID      primitive.ObjectID `json:"node_id"`
	Healthy     bool               `json:"healthy"`
	CheckTime   time.Time          `json:"check_time"`
	ResponseTime time.Duration     `json:"response_time"`
	Checks      []HealthCheck      `json:"checks"`
}

// HealthCheck 健康检查项
type HealthCheck struct {
	Name    string        `json:"name"`
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Latency time.Duration `json:"latency"`
}

// NodeMetrics 节点指标
type NodeMetrics struct {
	NodeID    primitive.ObjectID `json:"node_id"`
	Timestamp time.Time          `json:"timestamp"`
	CPU       CPUMetrics         `json:"cpu"`
	Memory    MemoryMetrics      `json:"memory"`
	Disk      DiskMetrics        `json:"disk"`
	Network   NetworkMetrics     `json:"network"`
	Tasks     TaskMetrics        `json:"tasks"`
}

// CPUMetrics CPU指标
type CPUMetrics struct {
	Usage     float64 `json:"usage"`      // 使用率 (0-100)
	UserTime  float64 `json:"user_time"`  // 用户时间
	SystemTime float64 `json:"system_time"` // 系统时间
	IdleTime  float64 `json:"idle_time"`  // 空闲时间
	LoadAvg1  float64 `json:"load_avg_1"` // 1分钟平均负载
	LoadAvg5  float64 `json:"load_avg_5"` // 5分钟平均负载
	LoadAvg15 float64 `json:"load_avg_15"` // 15分钟平均负载
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	Total     uint64  `json:"total"`     // 总内存 (bytes)
	Used      uint64  `json:"used"`      // 已用内存 (bytes)
	Free      uint64  `json:"free"`      // 空闲内存 (bytes)
	Available uint64  `json:"available"` // 可用内存 (bytes)
	Usage     float64 `json:"usage"`     // 使用率 (0-100)
	Cached    uint64  `json:"cached"`    // 缓存 (bytes)
	Buffers   uint64  `json:"buffers"`   // 缓冲区 (bytes)
}

// DiskMetrics 磁盘指标
type DiskMetrics struct {
	Total   uint64  `json:"total"`   // 总空间 (bytes)
	Used    uint64  `json:"used"`    // 已用空间 (bytes)
	Free    uint64  `json:"free"`    // 空闲空间 (bytes)
	Usage   float64 `json:"usage"`   // 使用率 (0-100)
	Inodes  uint64  `json:"inodes"`  // 总inode数
	IUsed   uint64  `json:"iused"`   // 已用inode数
	IFree   uint64  `json:"ifree"`   // 空闲inode数
	IUsage  float64 `json:"iusage"`  // inode使用率 (0-100)
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	BytesIn  uint64  `json:"bytes_in"`  // 入流量 (bytes)
	BytesOut uint64  `json:"bytes_out"` // 出流量 (bytes)
	PacketsIn uint64 `json:"packets_in"` // 入包数
	PacketsOut uint64 `json:"packets_out"` // 出包数
	ErrorsIn  uint64  `json:"errors_in"`  // 入错误数
	ErrorsOut uint64  `json:"errors_out"` // 出错误数
	DropsIn   uint64  `json:"drops_in"`   // 入丢包数
	DropsOut  uint64  `json:"drops_out"`  // 出丢包数
	Usage     float64 `json:"usage"`      // 网络使用率 (0-100)
}

// TaskMetrics 任务指标
type TaskMetrics struct {
	Total     int64   `json:"total"`     // 总任务数
	Running   int64   `json:"running"`   // 运行中任务数
	Completed int64   `json:"completed"` // 完成任务数
	Failed    int64   `json:"failed"`    // 失败任务数
	Pending   int64   `json:"pending"`   // 待执行任务数
	SuccessRate float64 `json:"success_rate"` // 成功率
	AvgDuration float64 `json:"avg_duration"` // 平均执行时间
}

// NodeEvent 节点事件
type NodeEvent struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	NodeID    primitive.ObjectID `json:"node_id"`
	Type      string             `json:"type"`
	Message   string             `json:"message"`
	Level     string             `json:"level"`    // info, warn, error
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time          `json:"created_at"`
}

// 向后兼容的数据结构
type NodeConfig struct {
	MaxConcurrentTasks int      `bson:"maxConcurrentTasks" json:"maxConcurrentTasks"`
	MaxMemoryUsage     int      `bson:"maxMemoryUsage" json:"maxMemoryUsage"`
	MaxCpuUsage        int      `bson:"maxCpuUsage" json:"maxCpuUsage"`
	HeartbeatInterval  int      `bson:"heartbeatInterval" json:"heartbeatInterval"`
	TaskTimeout        int      `bson:"taskTimeout" json:"taskTimeout"`
	EnabledTaskTypes   []string `bson:"enabledTaskTypes" json:"enabledTaskTypes"`
	LogLevel           string   `bson:"logLevel" json:"logLevel"`
	AutoUpdate         bool     `bson:"autoUpdate" json:"autoUpdate"`
}

type NodeTaskStats struct {
	TotalTasks     int            `bson:"totalTasks" json:"totalTasks"`
	SuccessTasks   int            `bson:"successTasks" json:"successTasks"`
	FailedTasks    int            `bson:"failedTasks" json:"failedTasks"`
	TaskTypeStats  map[string]int `bson:"taskTypeStats" json:"taskTypeStats"`
	AvgExecuteTime int            `bson:"avgExecuteTime" json:"avgExecuteTime"`
	LastTaskTime   time.Time      `bson:"lastTaskTime" json:"lastTaskTime"`
}

// 向后兼容的请求响应结构
type NodeRegistrationRequest struct {
	Name   string     `json:"name"`
	IP     string     `json:"ip"`
	Port   int        `json:"port"`
	Role   string     `json:"role"`   // 映射到 Type
	Tags   []string   `json:"tags"`
	Config NodeConfig `json:"config"`
}

type NodeRegistrationResponse struct {
	NodeID  string `json:"nodeId"`
	ApiKey  string `json:"apiKey"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type NodeConfigUpdateRequest struct {
	NodeID string     `json:"nodeId"`
	Config NodeConfig `json:"config"`
}

type NodeStatusResponse struct {
	NodeID     string           `json:"nodeId"`
	Name       string           `json:"name"`
	Status     string           `json:"status"`
	Role       string           `json:"role"`
	LastSeen   time.Time        `json:"lastSeen"`
	StatusInfo NodeStatusInfo   `json:"nodeStatus"`  // 使用正确的结构体类型
	TaskStats  NodeTaskStats    `json:"taskStats"`
}

// 方法实现

// GetNodeAddress 获取节点地址
func (n *Node) GetNodeAddress() string {
	return fmt.Sprintf("%s:%d", n.IP, n.Port)
}

// IsOnline 检查节点是否在线
func (n *Node) IsOnline() bool {
	return n.Status == NodeStatusOnline
}

// IsHealthy 检查节点是否健康
func (n *Node) IsHealthy() bool {
	if !n.IsOnline() {
		return false
	}
	
	// 检查心跳超时（5分钟内必须有心跳）
	if time.Since(n.LastHeartbeat) > 5*time.Minute {
		return false
	}
	
	return true
}

// CanAcceptTasks 检查节点是否可以接受新任务
func (n *Node) CanAcceptTasks() bool {
	if !n.IsHealthy() {
		return false
	}
	
	// 检查节点状态
	if n.Status == NodeStatusDraining || n.Status == NodeStatusMaintenance {
		return false
	}
	
	// 检查资源使用率
	if n.CPU > 90 || n.Memory > 90 || n.Disk > 95 {
		return false
	}
	
	return true
}

// GetLoadScore 获取节点负载分数（用于负载均衡）
func (n *Node) GetLoadScore() float64 {
	if !n.CanAcceptTasks() {
		return 1000.0 // 高分数表示不可用
	}
	
	// 计算综合负载分数
	score := (n.CPU * 0.4) + (n.Memory * 0.3) + (n.Load * 0.2) + (float64(n.ActiveTasks) * 0.1)
	return score
}

// HasCapability 检查节点是否具有指定能力
func (n *Node) HasCapability(capability string) bool {
	for _, cap := range n.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// ValidateRegistration 验证节点注册信息
func (nr *NodeRegistration) ValidateRegistration() error {
	if nr.Name == "" {
		return fmt.Errorf("节点名称不能为空")
	}
	
	if nr.IP == "" {
		return fmt.Errorf("节点IP地址不能为空")
	}
	
	// 验证IP地址格式
	if net.ParseIP(nr.IP) == nil {
		return fmt.Errorf("无效的IP地址: %s", nr.IP)
	}
	
	if nr.Port <= 0 || nr.Port > 65535 {
		return fmt.Errorf("节点端口必须在1-65535范围内")
	}
	
	if nr.Type == "" {
		return fmt.Errorf("节点类型不能为空")
	}
	
	validTypes := map[NodeType]bool{
		NodeTypeWorker:    true,
		NodeTypeManager:   true,
		NodeTypeScheduler: true,
		NodeTypeMonitor:   true,
		NodeTypeMixed:     true,
	}
	
	if !validTypes[nr.Type] {
		return fmt.Errorf("无效的节点类型: %s", nr.Type)
	}
	
	if nr.Secret == "" {
		return fmt.Errorf("节点密钥不能为空")
	}
	
	return nil
}

// Validate 验证心跳数据
func (nh *NodeHeartbeat) Validate() error {
	if nh.NodeID.IsZero() {
		return fmt.Errorf("节点ID不能为空")
	}
	
	if nh.CPU < 0 || nh.CPU > 100 {
		return fmt.Errorf("CPU使用率必须在0-100范围内")
	}
	
	if nh.Memory < 0 || nh.Memory > 100 {
		return fmt.Errorf("内存使用率必须在0-100范围内")
	}
	
	if nh.Disk < 0 || nh.Disk > 100 {
		return fmt.Errorf("磁盘使用率必须在0-100范围内")
	}
	
	if nh.ActiveTasks < 0 {
		return fmt.Errorf("活跃任务数不能为负数")
	}
	
	return nil
}

// ToResponse 将节点转换为响应格式（向后兼容）
func (n *Node) ToResponse() *NodeStatusResponse {
	// 状态映射
	status := string(n.Status)
	
	// 创建兼容的NodeStatusInfo结构
	nodeStatus := NodeStatusInfo{
		CpuUsage:       n.CPU,
		MemoryUsage:    int64(n.Memory * 1024), // 转换为MB
		DiskUsage:      int64(n.Disk * 1024),   // 转换为MB
		LoadAverage:    []float64{n.Load, n.Load, n.Load}, // 简化实现
		RunningTasks:   int(n.ActiveTasks),
		QueuedTasks:    0, // 简化实现
		NetworkIn:      0, // 需要从NetworkMetrics获取
		NetworkOut:     0, // 需要从NetworkMetrics获取
		UptimeSeconds:  int64(time.Since(n.RegisteredAt).Seconds()),
		LastUpdateTime: n.LastUpdate,
	}
	
	// 创建兼容的TaskStats结构
	taskStats := NodeTaskStats{
		TotalTasks:     int(n.CompletedTasks + n.FailedTasks + n.ActiveTasks),
		SuccessTasks:   int(n.CompletedTasks),
		FailedTasks:    int(n.FailedTasks),
		TaskTypeStats:  make(map[string]int),
		AvgExecuteTime: 0, // 需要从详细统计计算
		LastTaskTime:   n.LastUpdate,
	}
	
	return &NodeStatusResponse{
		NodeID:     n.ID.Hex(),
		Name:       n.Name,
		Status:     status,
		Role:       string(n.Type),
		LastSeen:   n.LastHeartbeat,
		StatusInfo: nodeStatus,
		TaskStats:  taskStats,
	}
}

// FromRegistrationRequest 从注册请求创建节点（向后兼容）
func (nr *NodeRegistrationRequest) ToNodeRegistration() *NodeRegistration {
	// 角色映射到类型
	var nodeType NodeType
	switch nr.Role {
	case NodeRoleMaster:
		nodeType = NodeTypeManager
	case NodeRoleWorker:
		nodeType = NodeTypeWorker
	case "scheduler":
		nodeType = NodeTypeScheduler
	case "monitor":
		nodeType = NodeTypeMonitor
	default:
		nodeType = NodeTypeWorker
	}
	
	// 从config中提取capabilities
	capabilities := nr.Config.EnabledTaskTypes
	if capabilities == nil {
		capabilities = []string{}
	}
	
	return &NodeRegistration{
		Name:         nr.Name,
		IP:           nr.IP,
		Port:         nr.Port,
		Type:         nodeType,
		Version:      "1.0.0", // 默认版本
		Capabilities: capabilities,
		Metadata:     map[string]string{
			"max_concurrent_tasks": strconv.Itoa(nr.Config.MaxConcurrentTasks),
			"max_memory_usage":     strconv.Itoa(nr.Config.MaxMemoryUsage),
			"max_cpu_usage":        strconv.Itoa(nr.Config.MaxCpuUsage),
			"heartbeat_interval":   strconv.Itoa(nr.Config.HeartbeatInterval),
			"task_timeout":         strconv.Itoa(nr.Config.TaskTimeout),
			"log_level":            nr.Config.LogLevel,
		},
		Tags:   nr.Tags,
		Secret: "", // 需要外部生成
	}
}