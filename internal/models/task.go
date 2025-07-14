package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ExecutorInfo 执行器信息
type ExecutorInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
}

// TaskPriority 任务优先级
const (
	TaskPriorityLow    = 1 // 低优先级
	TaskPriorityNormal = 2 // 正常优先级
	TaskPriorityHigh   = 3 // 高优先级
	TaskPriorityCrit   = 4 // 关键优先级
)

// TaskType 任务类型
const (
	TaskTypeSubdomainEnum  = "subdomain_enum"  // 子域名枚举
	TaskTypePortScan       = "port_scan"       // 端口扫描
	TaskTypeVulnScan       = "vuln_scan"       // 漏洞扫描
	TaskTypeAssetDiscovery = "asset_discovery" // 资产发现
)

// Task 任务基础模型
type Task struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name"`               // 任务名称
	Description string                 `bson:"description" json:"description"` // 任务描述
	Type        string                 `bson:"type" json:"type"`               // 任务类型
	Status      string                 `bson:"status" json:"status"`           // 任务状态
	Priority    int                    `bson:"priority" json:"priority"`       // 优先级
	ProjectID   primitive.ObjectID     `bson:"projectId" json:"projectId"`     // 所属项目
	CreatedBy   primitive.ObjectID     `bson:"createdBy" json:"createdBy"`     // 创建者
	CreatedAt   time.Time              `bson:"createdAt" json:"createdAt"`     // 创建时间
	UpdatedAt   time.Time              `bson:"updatedAt" json:"updatedAt"`     // 更新时间
	StartedAt   time.Time              `bson:"startedAt" json:"startedAt"`     // 开始时间
	CompletedAt time.Time              `bson:"completedAt" json:"completedAt"` // 完成时间
	Timeout     int                    `bson:"timeout" json:"timeout"`         // 超时时间(秒)
	RetryCount  int                    `bson:"retryCount" json:"retryCount"`   // 重试次数
	MaxRetries  int                    `bson:"maxRetries" json:"maxRetries"`   // 最大重试次数
	Progress    float64                `bson:"progress" json:"progress"`       // 进度(0-100)
	NodeID      string                 `bson:"nodeId" json:"nodeId"`           // 执行节点ID
	DependsOn   []string               `bson:"dependsOn" json:"dependsOn"`     // 依赖任务ID
	Tags        []string               `bson:"tags" json:"tags"`               // 标签
	Error       string                 `bson:"error" json:"error"`             // 错误信息
	ResultID    primitive.ObjectID     `bson:"resultId" json:"resultId"`       // 结果ID
	Params      interface{}            `bson:"params" json:"params"`           // 任务参数
	Config      map[string]interface{} `bson:"config,omitempty" json:"config,omitempty"` // 任务配置
	CallbackURL string                 `bson:"callbackUrl" json:"callbackUrl"` // 回调URL
}

// TaskResult 任务结果
type TaskResult struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	TaskID    primitive.ObjectID     `bson:"taskId" json:"taskId"`           // 任务ID
	Status    string                 `bson:"status" json:"status"`           // 结果状态
	Data      map[string]interface{} `bson:"data" json:"data"`               // 结果数据
	Summary   string                 `bson:"summary" json:"summary"`         // 结果摘要
	StartTime time.Time              `bson:"startTime" json:"startTime"`     // 开始时间
	EndTime   time.Time              `bson:"endTime" json:"endTime"`         // 结束时间
	CreatedAt time.Time              `bson:"createdAt" json:"createdAt"`     // 创建时间
	UpdatedAt time.Time              `bson:"updatedAt" json:"updatedAt"`     // 更新时间
	Error     string                 `bson:"error" json:"error"`             // 错误信息
}

// TaskQueue 任务队列
type TaskQueue struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`           // 队列名称
	Type      string             `bson:"type" json:"type"`           // 队列类型
	Priority  int                `bson:"priority" json:"priority"`   // 队列优先级
	MaxSize   int                `bson:"maxSize" json:"maxSize"`     // 最大队列大小
	TaskCount int                `bson:"taskCount" json:"taskCount"` // 任务数量
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"` // 创建时间
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"` // 更新时间
}

// TaskEvent 任务事件
type TaskEvent struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TaskID    primitive.ObjectID `bson:"taskId" json:"taskId"`       // 任务ID
	Type      string             `bson:"type" json:"type"`           // 事件类型
	Status    string             `bson:"status" json:"status"`       // 任务状态
	Message   string             `bson:"message" json:"message"`     // 事件消息
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"` // 事件时间
	NodeID    string             `bson:"nodeId" json:"nodeId"`       // 节点ID
}

// TaskEventType 任务事件类型
const (
	TaskEventCreated   = "created"   // 任务创建
	TaskEventQueued    = "queued"    // 任务入队
	TaskEventStarted   = "started"   // 任务开始
	TaskEventProgress  = "progress"  // 任务进度
	TaskEventCompleted = "completed" // 任务完成
	TaskEventFailed    = "failed"    // 任务失败
	TaskEventCanceled  = "canceled"  // 任务取消
	TaskEventTimeout   = "timeout"   // 任务超时
	TaskEventRetry     = "retry"     // 任务重试
	TaskEventAssigned  = "assigned"  // 任务分配
)

// TaskCreateRequest 创建任务请求
type TaskCreateRequest struct {
	Name        string      `json:"name"`        // 任务名称
	Description string      `json:"description"` // 任务描述
	Type        string      `json:"type"`        // 任务类型
	Priority    int         `json:"priority"`    // 优先级
	ProjectID   string      `json:"projectId"`   // 所属项目
	Timeout     int         `json:"timeout"`     // 超时时间(秒)
	MaxRetries  int         `json:"maxRetries"`  // 最大重试次数
	DependsOn   []string    `json:"dependsOn"`   // 依赖任务ID
	Tags        []string    `json:"tags"`        // 标签
	Params      interface{} `json:"params"`      // 任务参数
	CallbackURL string      `json:"callbackUrl"` // 回调URL
}

// TaskUpdateRequest 更新任务请求
type TaskUpdateRequest struct {
	Status     string      `json:"status"`     // 任务状态
	Progress   float64     `json:"progress"`   // 进度
	Error      string      `json:"error"`      // 错误信息
	ResultData interface{} `json:"resultData"` // 结果数据
}

// TaskTemplate 任务模板
type TaskTemplate struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name"`               // 模板名称
	Description string                 `bson:"description" json:"description"` // 模板描述
	Type        string                 `bson:"type" json:"type"`               // 任务类型
	Priority    int                    `bson:"priority" json:"priority"`       // 优先级
	Config      map[string]interface{} `bson:"config" json:"config"`           // 模板配置
	Timeout     int                    `bson:"timeout" json:"timeout"`         // 超时时间(秒)
	MaxRetries  int                    `bson:"maxRetries" json:"maxRetries"`   // 最大重试次数
	Tags        []string               `bson:"tags" json:"tags"`               // 标签
	IsPublic    bool                   `bson:"isPublic" json:"isPublic"`       // 是否公开
	CreatedBy   primitive.ObjectID     `bson:"createdBy" json:"createdBy"`     // 创建者
	CreatedAt   time.Time              `bson:"createdAt" json:"createdAt"`     // 创建时间
	UpdatedAt   time.Time              `bson:"updatedAt" json:"updatedAt"`     // 更新时间
	UsageCount  int                    `bson:"usageCount" json:"usageCount"`   // 使用次数
	Metadata    map[string]interface{} `bson:"metadata" json:"metadata"`       // 元数据
}

// TaskScheduleRule 任务调度规则
type TaskScheduleRule struct {
	ID             primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name           string                 `bson:"name" json:"name"`                     // 规则名称
	Description    string                 `bson:"description" json:"description"`       // 规则描述
	TemplateID     string                 `bson:"templateId" json:"templateId"`         // 任务模板ID
	CronExpression string                 `bson:"cronExpression" json:"cronExpression"` // Cron表达式
	Timezone       string                 `bson:"timezone" json:"timezone"`             // 时区
	Enabled        bool                   `bson:"enabled" json:"enabled"`               // 是否启用
	ProjectID      string                 `bson:"projectId" json:"projectId"`           // 所属项目
	NextRunTime    time.Time              `bson:"nextRunTime" json:"nextRunTime"`       // 下次运行时间
	LastRunTime    time.Time              `bson:"lastRunTime" json:"lastRunTime"`       // 上次运行时间
	RunCount       int                    `bson:"runCount" json:"runCount"`             // 运行次数
	MaxRuns        *int                   `bson:"maxRuns" json:"maxRuns"`               // 最大运行次数
	CreatedBy      string                 `bson:"createdBy" json:"createdBy"`           // 创建者
	CreatedAt      time.Time              `bson:"createdAt" json:"createdAt"`           // 创建时间
	UpdatedAt      time.Time              `bson:"updatedAt" json:"updatedAt"`           // 更新时间
	Metadata       map[string]interface{} `bson:"metadata" json:"metadata"`             // 元数据
}
