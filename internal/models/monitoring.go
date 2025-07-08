package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PageMonitoringStatus 页面监控状态
const (
	PageMonitoringStatusActive   = "active"   // 活跃
	PageMonitoringStatusInactive = "inactive" // 非活跃
	PageMonitoringStatusError    = "error"    // 错误
)

// PageChangeStatus 页面变更状态
const (
	PageChangeStatusNew       = "new"       // 新增
	PageChangeStatusChanged   = "changed"   // 变更
	PageChangeStatusRemoved   = "removed"   // 移除
	PageChangeStatusUnchanged = "unchanged" // 未变更
)

// PageMonitoring 页面监控模型
type PageMonitoring struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	URL              string             `bson:"url" json:"url"`                           // 监控URL
	Name             string             `bson:"name" json:"name"`                         // 监控名称
	Status           string             `bson:"status" json:"status"`                     // 监控状态
	ProjectID        primitive.ObjectID `bson:"projectId" json:"projectId"`               // 所属项目
	Interval         int                `bson:"interval" json:"interval"`                 // 监控间隔(小时)
	LastCheckAt      time.Time          `bson:"lastCheckAt" json:"lastCheckAt"`           // 最后检查时间
	NextCheckAt      time.Time          `bson:"nextCheckAt" json:"nextCheckAt"`           // 下次检查时间
	CreatedAt        time.Time          `bson:"createdAt" json:"createdAt"`               // 创建时间
	UpdatedAt        time.Time          `bson:"updatedAt" json:"updatedAt"`               // 更新时间
	Tags             []string           `bson:"tags" json:"tags"`                         // 标签
	Config           MonitoringConfig   `bson:"config" json:"config"`                     // 监控配置
	LatestSnapshot   *PageSnapshot      `bson:"latestSnapshot" json:"latestSnapshot"`     // 最新快照
	PreviousSnapshot *PageSnapshot      `bson:"previousSnapshot" json:"previousSnapshot"` // 上一次快照
	ChangeCount      int                `bson:"changeCount" json:"changeCount"`           // 变更次数
	HasChanged       bool               `bson:"hasChanged" json:"hasChanged"`             // 是否有变更
	Similarity       float64            `bson:"similarity" json:"similarity"`             // 相似度
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	IgnoreCSS           bool              `bson:"ignoreCSS" json:"ignoreCSS"`                     // 忽略CSS变化
	IgnoreJS            bool              `bson:"ignoreJS" json:"ignoreJS"`                       // 忽略JS变化
	IgnoreImages        bool              `bson:"ignoreImages" json:"ignoreImages"`               // 忽略图片变化
	IgnoreNumbers       bool              `bson:"ignoreNumbers" json:"ignoreNumbers"`             // 忽略数字变化
	IgnorePatterns      []string          `bson:"ignorePatterns" json:"ignorePatterns"`           // 忽略的正则表达式模式
	SimilarityThreshold float64           `bson:"similarityThreshold" json:"similarityThreshold"` // 相似度阈值
	Timeout             int               `bson:"timeout" json:"timeout"`                         // 请求超时时间(秒)
	Headers             map[string]string `bson:"headers" json:"headers"`                         // 自定义请求头
	Authentication      AuthConfig        `bson:"authentication" json:"authentication"`           // 认证配置
	Selector            string            `bson:"selector" json:"selector"`                       // CSS选择器，用于只监控页面的特定部分
	CompareMethod       string            `bson:"compareMethod" json:"compareMethod"`             // 比较方法: text, html, visual, hash
	NotifyOnChange      bool              `bson:"notifyOnChange" json:"notifyOnChange"`           // 变更时通知
	NotifyMethods       []string          `bson:"notifyMethods" json:"notifyMethods"`             // 通知方式: email, webhook, sms
	NotifyConfig        map[string]string `bson:"notifyConfig" json:"notifyConfig"`               // 通知配置
}

// AuthConfig 认证配置
type AuthConfig struct {
	Type     string `bson:"type" json:"type"`         // 认证类型: none, basic, form, cookie
	Username string `bson:"username" json:"username"` // 用户名
	Password string `bson:"password" json:"password"` // 密码
	Cookie   string `bson:"cookie" json:"cookie"`     // Cookie
}

// PageSnapshot 页面快照
type PageSnapshot struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MonitoringID primitive.ObjectID `bson:"monitoringId" json:"monitoringId"` // 监控ID
	URL          string             `bson:"url" json:"url"`                   // URL
	StatusCode   int                `bson:"statusCode" json:"statusCode"`     // HTTP状态码
	Headers      map[string]string  `bson:"headers" json:"headers"`           // 响应头
	HTML         string             `bson:"html" json:"html"`                 // HTML内容
	Text         string             `bson:"text" json:"text"`                 // 文本内容
	ContentHash  string             `bson:"contentHash" json:"contentHash"`   // 内容哈希
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`       // 创建时间
	Size         int                `bson:"size" json:"size"`                 // 内容大小(字节)
	LoadTime     int                `bson:"loadTime" json:"loadTime"`         // 加载时间(毫秒)
}

// PageChange 页面变更记录
type PageChange struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MonitoringID  primitive.ObjectID `bson:"monitoringId" json:"monitoringId"`   // 监控ID
	URL           string             `bson:"url" json:"url"`                     // URL
	Status        string             `bson:"status" json:"status"`               // 变更状态
	OldSnapshotID primitive.ObjectID `bson:"oldSnapshotId" json:"oldSnapshotId"` // 旧快照ID
	NewSnapshotID primitive.ObjectID `bson:"newSnapshotId" json:"newSnapshotId"` // 新快照ID
	Similarity    float64            `bson:"similarity" json:"similarity"`       // 相似度
	ChangedAt     time.Time          `bson:"changedAt" json:"changedAt"`         // 变更时间
	Diff          string             `bson:"diff" json:"diff"`                   // 差异内容
	DiffType      string             `bson:"diffType" json:"diffType"`           // 差异类型: text, html, visual
}

// PageMonitoringCreateRequest 创建页面监控请求
type PageMonitoringCreateRequest struct {
	URL       string           `json:"url"`       // 监控URL
	Name      string           `json:"name"`      // 监控名称
	ProjectID string           `json:"projectId"` // 所属项目
	Interval  int              `json:"interval"`  // 监控间隔(小时)
	Tags      []string         `json:"tags"`      // 标签
	Config    MonitoringConfig `json:"config"`    // 监控配置
}

// PageMonitoringUpdateRequest 更新页面监控请求
type PageMonitoringUpdateRequest struct {
	Name     string           `json:"name"`     // 监控名称
	Status   string           `json:"status"`   // 监控状态
	Interval int              `json:"interval"` // 监控间隔(小时)
	Tags     []string         `json:"tags"`     // 标签
	Config   MonitoringConfig `json:"config"`   // 监控配置
}

// PageMonitoringQueryRequest 查询页面监控请求
type PageMonitoringQueryRequest struct {
	ProjectID string   `json:"projectId"` // 所属项目
	Status    string   `json:"status"`    // 监控状态
	Tags      []string `json:"tags"`      // 标签
	URL       string   `json:"url"`       // URL关键字
	Name      string   `json:"name"`      // 名称关键字
	Limit     int      `json:"limit"`     // 限制数量
	Offset    int      `json:"offset"`    // 偏移量
	SortBy    string   `json:"sortBy"`    // 排序字段
	SortOrder string   `json:"sortOrder"` // 排序方向
}
