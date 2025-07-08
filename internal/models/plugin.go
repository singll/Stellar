package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PluginMetadata 插件元数据
type PluginMetadata struct {
	ID           string                 `bson:"_id" json:"id"`                    // 插件ID
	Name         string                 `bson:"name" json:"name"`                 // 插件名称
	Version      string                 `bson:"version" json:"version"`           // 插件版本
	Type         string                 `bson:"type" json:"type"`                 // 插件类型
	Author       string                 `bson:"author" json:"author"`             // 作者
	Description  string                 `bson:"description" json:"description"`   // 描述
	Category     string                 `bson:"category" json:"category"`         // 分类
	Tags         []string               `bson:"tags" json:"tags"`                 // 标签
	Path         string                 `bson:"path" json:"path"`                 // 插件路径
	Config       map[string]interface{} `bson:"config" json:"config"`             // 配置
	Enabled      bool                   `bson:"enabled" json:"enabled"`           // 是否启用
	InstallTime  time.Time              `bson:"install_time" json:"installTime"`  // 安装时间
	UpdateTime   time.Time              `bson:"update_time" json:"updateTime"`    // 更新时间
	LastRunTime  time.Time              `bson:"last_run_time" json:"lastRunTime"` // 最后运行时间
	RunCount     int                    `bson:"run_count" json:"runCount"`        // 运行次数
	ErrorCount   int                    `bson:"error_count" json:"errorCount"`    // 错误次数
	AvgRuntime   float64                `bson:"avg_runtime" json:"avgRuntime"`    // 平均运行时间
	Dependencies []PluginDependency     `bson:"dependencies" json:"dependencies"` // 依赖
}

// PluginDependency 插件依赖
type PluginDependency struct {
	ID      string `bson:"id" json:"id"`           // 依赖ID
	Version string `bson:"version" json:"version"` // 依赖版本
}

// PluginRunRecord 插件运行记录
type PluginRunRecord struct {
	ID        primitive.ObjectID     `bson:"_id" json:"id"`               // 记录ID
	PluginID  string                 `bson:"plugin_id" json:"pluginId"`   // 插件ID
	StartTime time.Time              `bson:"start_time" json:"startTime"` // 开始时间
	EndTime   time.Time              `bson:"end_time" json:"endTime"`     // 结束时间
	Duration  time.Duration          `bson:"duration" json:"duration"`    // 持续时间
	Success   bool                   `bson:"success" json:"success"`      // 是否成功
	Error     string                 `bson:"error" json:"error"`          // 错误信息
	Params    map[string]interface{} `bson:"params" json:"params"`        // 参数
	Result    interface{}            `bson:"result" json:"result"`        // 结果
	TaskID    string                 `bson:"task_id" json:"taskId"`       // 关联的任务ID
	UserID    string                 `bson:"user_id" json:"userId"`       // 用户ID
}

// PluginConfig 插件配置
type PluginConfig struct {
	ID          string                 `bson:"_id" json:"id"`                  // 配置ID
	PluginID    string                 `bson:"plugin_id" json:"pluginId"`      // 插件ID
	Name        string                 `bson:"name" json:"name"`               // 配置名称
	Description string                 `bson:"description" json:"description"` // 描述
	Config      map[string]interface{} `bson:"config" json:"config"`           // 配置内容
	IsDefault   bool                   `bson:"is_default" json:"isDefault"`    // 是否为默认配置
	CreatedAt   time.Time              `bson:"created_at" json:"createdAt"`    // 创建时间
	UpdatedAt   time.Time              `bson:"updated_at" json:"updatedAt"`    // 更新时间
	CreatedBy   string                 `bson:"created_by" json:"createdBy"`    // 创建者
}

// PluginMarketItem 插件市场项
type PluginMarketItem struct {
	ID           string             `bson:"_id" json:"id"`                    // 插件ID
	Name         string             `bson:"name" json:"name"`                 // 插件名称
	Version      string             `bson:"version" json:"version"`           // 插件版本
	Type         string             `bson:"type" json:"type"`                 // 插件类型
	Author       string             `bson:"author" json:"author"`             // 作者
	Description  string             `bson:"description" json:"description"`   // 描述
	Category     string             `bson:"category" json:"category"`         // 分类
	Tags         []string           `bson:"tags" json:"tags"`                 // 标签
	DownloadURL  string             `bson:"download_url" json:"downloadUrl"`  // 下载URL
	Homepage     string             `bson:"homepage" json:"homepage"`         // 主页
	License      string             `bson:"license" json:"license"`           // 许可证
	Stars        int                `bson:"stars" json:"stars"`               // 星级
	Downloads    int                `bson:"downloads" json:"downloads"`       // 下载次数
	PublishTime  time.Time          `bson:"publish_time" json:"publishTime"`  // 发布时间
	UpdateTime   time.Time          `bson:"update_time" json:"updateTime"`    // 更新时间
	Verified     bool               `bson:"verified" json:"verified"`         // 是否已验证
	Screenshots  []string           `bson:"screenshots" json:"screenshots"`   // 截图
	Dependencies []PluginDependency `bson:"dependencies" json:"dependencies"` // 依赖
}
