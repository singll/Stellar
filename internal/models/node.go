package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NodeStatus 节点状态
const (
	NodeStatusOnline    = "online"    // 在线
	NodeStatusOffline   = "offline"   // 离线
	NodeStatusDisabled  = "disabled"  // 禁用
	NodeStatusMaintain  = "maintain"  // 维护中
	NodeStatusRegisting = "registing" // 注册中
)

// NodeRole 节点角色
const (
	NodeRoleMaster = "master" // 主节点
	NodeRoleSlave  = "slave"  // 从节点
	NodeRoleWorker = "worker" // 工作节点
)

// Node 节点数据模型
type Node struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name              string             `bson:"name" json:"name"`                           // 节点名称
	Role              string             `bson:"role" json:"role"`                           // 节点角色
	Status            string             `bson:"status" json:"status"`                       // 节点状态
	IP                string             `bson:"ip" json:"ip"`                               // 节点IP
	Port              int                `bson:"port" json:"port"`                           // 节点端口
	ApiKey            string             `bson:"apiKey" json:"apiKey"`                       // API密钥
	RegisterTime      time.Time          `bson:"registerTime" json:"registerTime"`           // 注册时间
	LastHeartbeatTime time.Time          `bson:"lastHeartbeatTime" json:"lastHeartbeatTime"` // 最后心跳时间
	Tags              []string           `bson:"tags" json:"tags"`                           // 标签
	Config            NodeConfig         `bson:"config" json:"config"`                       // 节点配置
	StatusInfo        NodeStatus         `bson:"nodeStatus" json:"nodeStatus"`               // 节点状态信息
	TaskStats         NodeTaskStats      `bson:"taskStats" json:"taskStats"`                 // 任务统计
}

// NodeConfig 节点配置
type NodeConfig struct {
	MaxConcurrentTasks int      `bson:"maxConcurrentTasks" json:"maxConcurrentTasks"` // 最大并发任务数
	MaxMemoryUsage     int      `bson:"maxMemoryUsage" json:"maxMemoryUsage"`         // 最大内存使用量(MB)
	MaxCpuUsage        int      `bson:"maxCpuUsage" json:"maxCpuUsage"`               // 最大CPU使用率(%)
	HeartbeatInterval  int      `bson:"heartbeatInterval" json:"heartbeatInterval"`   // 心跳间隔(秒)
	TaskTimeout        int      `bson:"taskTimeout" json:"taskTimeout"`               // 任务超时时间(秒)
	EnabledTaskTypes   []string `bson:"enabledTaskTypes" json:"enabledTaskTypes"`     // 启用的任务类型
	LogLevel           string   `bson:"logLevel" json:"logLevel"`                     // 日志级别
	AutoUpdate         bool     `bson:"autoUpdate" json:"autoUpdate"`                 // 是否自动更新
}

// NodeStatus 节点状态信息
type NodeStatus struct {
	CpuUsage       float64   `bson:"cpuUsage" json:"cpuUsage"`             // CPU使用率
	MemoryUsage    int64     `bson:"memoryUsage" json:"memoryUsage"`       // 内存使用量(MB)
	DiskUsage      int64     `bson:"diskUsage" json:"diskUsage"`           // 磁盘使用量(MB)
	LoadAverage    []float64 `bson:"loadAverage" json:"loadAverage"`       // 负载平均值
	RunningTasks   int       `bson:"runningTasks" json:"runningTasks"`     // 运行中的任务数
	QueuedTasks    int       `bson:"queuedTasks" json:"queuedTasks"`       // 队列中的任务数
	NetworkIn      int64     `bson:"networkIn" json:"networkIn"`           // 网络入流量(KB/s)
	NetworkOut     int64     `bson:"networkOut" json:"networkOut"`         // 网络出流量(KB/s)
	UptimeSeconds  int64     `bson:"uptimeSeconds" json:"uptimeSeconds"`   // 正常运行时间(秒)
	LastUpdateTime time.Time `bson:"lastUpdateTime" json:"lastUpdateTime"` // 最后更新时间
}

// NodeTaskStats 节点任务统计
type NodeTaskStats struct {
	TotalTasks     int            `bson:"totalTasks" json:"totalTasks"`         // 总任务数
	SuccessTasks   int            `bson:"successTasks" json:"successTasks"`     // 成功任务数
	FailedTasks    int            `bson:"failedTasks" json:"failedTasks"`       // 失败任务数
	TaskTypeStats  map[string]int `bson:"taskTypeStats" json:"taskTypeStats"`   // 任务类型统计
	AvgExecuteTime int            `bson:"avgExecuteTime" json:"avgExecuteTime"` // 平均执行时间(秒)
	LastTaskTime   time.Time      `bson:"lastTaskTime" json:"lastTaskTime"`     // 最后任务时间
}

// NodeHeartbeat 节点心跳
type NodeHeartbeat struct {
	NodeID       primitive.ObjectID `bson:"nodeId" json:"nodeId"`             // 节点ID
	Timestamp    time.Time          `bson:"timestamp" json:"timestamp"`       // 时间戳
	Status       string             `bson:"status" json:"status"`             // 状态
	CpuUsage     float64            `bson:"cpuUsage" json:"cpuUsage"`         // CPU使用率
	MemoryUsage  int64              `bson:"memoryUsage" json:"memoryUsage"`   // 内存使用量(MB)
	RunningTasks int                `bson:"runningTasks" json:"runningTasks"` // 运行中的任务数
	QueuedTasks  int                `bson:"queuedTasks" json:"queuedTasks"`   // 队列中的任务数
	Version      string             `bson:"version" json:"version"`           // 版本
}

// NodeRegistrationRequest 节点注册请求
type NodeRegistrationRequest struct {
	Name   string     `json:"name"`   // 节点名称
	IP     string     `json:"ip"`     // 节点IP
	Port   int        `json:"port"`   // 节点端口
	Role   string     `json:"role"`   // 节点角色
	Tags   []string   `json:"tags"`   // 标签
	Config NodeConfig `json:"config"` // 节点配置
}

// NodeRegistrationResponse 节点注册响应
type NodeRegistrationResponse struct {
	NodeID  string `json:"nodeId"`  // 节点ID
	ApiKey  string `json:"apiKey"`  // API密钥
	Status  string `json:"status"`  // 状态
	Message string `json:"message"` // 消息
}

// NodeConfigUpdateRequest 节点配置更新请求
type NodeConfigUpdateRequest struct {
	NodeID string     `json:"nodeId"` // 节点ID
	Config NodeConfig `json:"config"` // 节点配置
}

// NodeStatusResponse 节点状态响应
type NodeStatusResponse struct {
	NodeID     string        `json:"nodeId"`     // 节点ID
	Name       string        `json:"name"`       // 节点名称
	Status     string        `json:"status"`     // 状态
	Role       string        `json:"role"`       // 角色
	LastSeen   time.Time     `json:"lastSeen"`   // 最后在线时间
	StatusInfo NodeStatus    `json:"nodeStatus"` // 节点状态信息
	TaskStats  NodeTaskStats `json:"taskStats"`  // 任务统计
}
