package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TODO: 完善页面监控相关模型的实现

// MonitoringTaskStatus 监控任务状态
type MonitoringTaskStatus string

const (
	MonitoringTaskStatusActive  MonitoringTaskStatus = "active"  // 活跃
	MonitoringTaskStatusPaused  MonitoringTaskStatus = "paused"  // 暂停
	MonitoringTaskStatusFailed  MonitoringTaskStatus = "failed"  // 失败
	MonitoringTaskStatusRunning MonitoringTaskStatus = "running" // 运行中
)

// PageMonitoringStatus 页面监控状态
type PageMonitoringStatus string

const (
	PageMonitoringStatusActive  PageMonitoringStatus = "active"  // 活跃
	PageMonitoringStatusPaused  PageMonitoringStatus = "paused"  // 暂停
	PageMonitoringStatusFailed  PageMonitoringStatus = "failed"  // 失败
	PageMonitoringStatusRunning PageMonitoringStatus = "running" // 运行中
)

// MonitoringTask 监控任务
type MonitoringTask struct {
	ID          primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Name        string               `json:"name" bson:"name"`
	URL         string               `json:"url" bson:"url"`
	Status      MonitoringTaskStatus `json:"status" bson:"status"`
	Config      MonitoringConfig     `json:"config" bson:"config"`         // 监控配置
	Interval    time.Duration        `json:"interval" bson:"interval"`
	CreatedAt   time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at" bson:"updated_at"`
	// TODO: 以下字段根据DEV_PLAN 0.6版本需要完善
	Message     string               `json:"message" bson:"message"`       // 消息信息
	LastRunAt   time.Time            `json:"last_run_at" bson:"last_run_at"` // 最后运行时间
}

// PageMonitoringTask 用于与PageSnapshot关联的任务模型
type PageMonitoringTask struct {
	MonitoringTask // 嵌入基础监控任务
	// 这里可以添加页面监控特有的字段
}

// AuthenticationConfig 认证配置
type AuthenticationConfig struct {
	Type     string `json:"type" bson:"type"`         // 认证类型：basic, cookie
	Username string `json:"username" bson:"username"` // 用户名
	Password string `json:"password" bson:"password"` // 密码
	Cookie   string `json:"cookie" bson:"cookie"`     // Cookie值
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	// TODO: 以下字段根据DEV_PLAN 0.6版本需要完善
	Headers            map[string]string     `json:"headers" bson:"headers"`               // HTTP头信息
	Authentication     *AuthenticationConfig `json:"authentication" bson:"authentication"` // 认证信息
	Timeout            int                   `json:"timeout" bson:"timeout"`               // 超时设置
	MaxDepth           int                   `json:"max_depth" bson:"max_depth"`           // 最大深度
	IgnoreNumbers      bool                  `json:"ignore_numbers" bson:"ignore_numbers"` // 忽略数字变化
	IgnorePatterns     []string              `json:"ignore_patterns" bson:"ignore_patterns"` // 忽略的正则模式
	CompareMethod      string                `json:"compare_method" bson:"compare_method"` // 比较方法: text, hash, html
	SimilarityThreshold float64              `json:"similarity_threshold" bson:"similarity_threshold"` // 相似度阈值
	NotifyOnChange     bool                  `json:"notify_on_change" bson:"notify_on_change"` // 变更时通知
	NotifyMethods      []string              `json:"notify_methods" bson:"notify_methods"`     // 通知方式
	NotifyConfig       map[string]interface{} `json:"notify_config" bson:"notify_config"`      // 通知配置
}

// PageSnapshot 页面快照
type PageSnapshot struct {
	// TODO: 以下字段根据DEV_PLAN 0.6版本需要完善
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`      // 快照ID
	TaskID       primitive.ObjectID `json:"task_id" bson:"task_id"`       // 任务ID
	MonitoringID primitive.ObjectID `json:"monitoring_id" bson:"monitoring_id"` // 监控任务ID
	URL          string             `json:"url" bson:"url"`               // 页面URL
	Headers      map[string]string  `json:"headers" bson:"headers"`       // HTTP头信息
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"` // 创建时间
	LoadTime     int64              `json:"load_time" bson:"load_time"`   // 加载时间(毫秒)
	Content      string             `json:"content" bson:"content"`       // 页面内容
	Hash         string             `json:"hash" bson:"hash"`             // 内容哈希
	Title        string             `json:"title" bson:"title"`           // 页面标题
	Size         int64              `json:"size" bson:"size"`             // 页面大小
	StatusCode   int                `json:"status_code" bson:"status_code"` // HTTP状态码
	HTML         string             `json:"html" bson:"html"`             // HTML内容
	Text         string             `json:"text" bson:"text"`             // 文本内容
	ContentHash  string             `json:"content_hash" bson:"content_hash"` // 内容哈希
}

// PageChangeStatus 页面变化状态
type PageChangeStatus string

const (
	PageChangeStatusUnchanged PageChangeStatus = "unchanged" // 未改变
	PageChangeStatusChanged   PageChangeStatus = "changed"   // 已改变
	PageChangeStatusFailed    PageChangeStatus = "failed"    // 失败
)

// PageChange 页面变化
type PageChange struct {
	// TODO: 以下字段根据DEV_PLAN 0.6版本需要完善
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TaskID        primitive.ObjectID `json:"task_id" bson:"task_id"`       // 任务ID
	URL           string             `json:"url" bson:"url"`
	OldSnapshotID primitive.ObjectID `json:"old_snapshot_id" bson:"old_snapshot_id"`
	NewSnapshotID primitive.ObjectID `json:"new_snapshot_id" bson:"new_snapshot_id"`
	ChangeType    string             `json:"change_type" bson:"change_type"` // 变化类型
	OldValue      string             `json:"old_value" bson:"old_value"`     // 旧值
	NewValue      string             `json:"new_value" bson:"new_value"`     // 新值
	Confidence    float64            `json:"confidence" bson:"confidence"`   // 置信度
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`   // 创建时间
	ChangedAt     time.Time          `json:"changed_at" bson:"changed_at"`   // 变化时间
	Similarity    float64            `json:"similarity" bson:"similarity"`   // 相似度
	Status        PageChangeStatus   `json:"status" bson:"status"`           // 状态
	Diff          string             `json:"diff" bson:"diff"`               // 差异信息
	DiffType      string             `json:"diff_type" bson:"diff_type"`     // 差异类型
}

// PageMonitoring 页面监控
type PageMonitoring struct {
	// TODO: 添加字段实现
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name           string             `json:"name" bson:"name"`                     // 监控任务名称
	URL            string             `json:"url" bson:"url"`                       // 监控的URL
	ProjectID      primitive.ObjectID `json:"project_id" bson:"project_id"`        // 项目ID
	Config           MonitoringConfig   `json:"config" bson:"config"`                     // 监控配置
	Interval         time.Duration      `json:"interval" bson:"interval"`                 // 监控间隔
	LatestSnapshot   *PageSnapshot      `json:"latest_snapshot" bson:"latest_snapshot"`   // 最新快照
	PreviousSnapshot *PageSnapshot      `json:"previous_snapshot" bson:"previous_snapshot"` // 上一次快照
	Status           string             `json:"status" bson:"status"`                     // 监控状态
	Tags           []string           `json:"tags" bson:"tags"`                     // 标签
	ChangeCount    int64              `json:"change_count" bson:"change_count"`     // 变更次数
	NextCheckAt    time.Time          `json:"next_check_at" bson:"next_check_at"`   // 下次检查时间
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`         // 创建时间
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`         // 更新时间
	LastCheckAt    time.Time          `json:"last_check_at" bson:"last_check_at"`   // 最后检查时间
}

// TODO: 完善页面监控相关请求结构体 - 对应DEV_PLAN 0.6版本
// PageMonitoringCreateRequest 创建页面监控请求
type PageMonitoringCreateRequest struct {
	// TODO: 添加字段实现
	Name      string            `json:"name"`       // 监控名称
	URL       string            `json:"url"`        // 监控URL
	ProjectID primitive.ObjectID `json:"project_id"` // 项目ID
	Interval  time.Duration     `json:"interval"`   // 监控间隔
	Tags      []string          `json:"tags"`       // 标签
	Config    MonitoringConfig  `json:"config"`     // 监控配置
}

// PageMonitoringUpdateRequest 更新页面监控请求  
type PageMonitoringUpdateRequest struct {
	Name      string            `json:"name,omitempty"`       // 监控名称
	Status    string            `json:"status,omitempty"`     // 监控状态
	Interval  time.Duration     `json:"interval,omitempty"`   // 监控间隔
	Tags      []string          `json:"tags,omitempty"`       // 标签
	Config    MonitoringConfig  `json:"config,omitempty"`     // 监控配置
}

// PageMonitoringQueryRequest 查询页面监控请求
type PageMonitoringQueryRequest struct {
	ProjectID string   `json:"project_id,omitempty"` // 项目ID
	Status    string   `json:"status,omitempty"`     // 状态过滤
	Tags      []string `json:"tags,omitempty"`       // 标签过滤
	URL       string   `json:"url,omitempty"`        // URL关键字
	Name      string   `json:"name,omitempty"`       // 名称关键字
	Limit     int      `json:"limit,omitempty"`      // 分页大小
	Offset    int      `json:"offset,omitempty"`     // 分页偏移
	SortBy    string   `json:"sort_by,omitempty"`    // 排序字段
	SortOrder string   `json:"sort_order,omitempty"` // 排序方向
}
