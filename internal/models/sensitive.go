package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SensitiveRule 敏感信息检测规则
type SensitiveRule struct {
	ID                    primitive.ObjectID `bson:"_id" json:"id"`
	Name                  string             `bson:"name" json:"name"`
	Description           string             `bson:"description" json:"description"`
	Type                  string             `bson:"type" json:"type"`                                   // 规则类型：regex, keyword, etc.
	Pattern               string             `bson:"pattern" json:"pattern"`                             // 正则表达式或关键词
	Category              string             `bson:"category" json:"category"`                           // 分类：password, api_key, etc.
	RiskLevel             string             `bson:"riskLevel" json:"riskLevel"`                         // 风险等级：high, medium, low
	Tags                  []string           `bson:"tags" json:"tags"`                                   // 标签
	Enabled               bool               `bson:"enabled" json:"enabled"`                             // 是否启用
	Context               int                `bson:"context" json:"context"`                             // 上下文行数
	Examples              []string           `bson:"examples" json:"examples"`                           // 示例
	FalsePositivePatterns []string           `bson:"falsePositivePatterns" json:"falsePositivePatterns"` // 误报模式
	CreatedAt             time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt             time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// SensitiveRuleGroup 敏感信息规则组
type SensitiveRuleGroup struct {
	ID          primitive.ObjectID   `bson:"_id" json:"id"`
	Name        string               `bson:"name" json:"name"`
	Description string               `bson:"description" json:"description"`
	Rules       []primitive.ObjectID `bson:"rules" json:"rules"` // 规则ID列表
	Enabled     bool                 `bson:"enabled" json:"enabled"`
	CreatedAt   time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time            `bson:"updatedAt" json:"updatedAt"`
}

// SensitiveWhitelist 敏感信息白名单
type SensitiveWhitelist struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Type        string             `bson:"type" json:"type"`           // 白名单类型：target, pattern
	Value       string             `bson:"value" json:"value"`         // 目标URL或正则表达式
	ExpiresAt   time.Time          `bson:"expiresAt" json:"expiresAt"` // 过期时间
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// SensitiveDetectionConfig 敏感信息检测配置
type SensitiveDetectionConfig struct {
	Concurrency    int    `bson:"concurrency" json:"concurrency"`       // 并发数
	Timeout        int    `bson:"timeout" json:"timeout"`               // 超时时间（秒）
	MaxDepth       int    `bson:"maxDepth" json:"maxDepth"`             // 最大爬取深度
	ContextLines   int    `bson:"contextLines" json:"contextLines"`     // 上下文行数
	FollowLinks    bool   `bson:"followLinks" json:"followLinks"`       // 是否跟踪链接
	UserAgent      string `bson:"userAgent" json:"userAgent"`           // User-Agent
	IgnoreRobots   bool   `bson:"ignoreRobots" json:"ignoreRobots"`     // 是否忽略robots.txt
	MaxFileSize    int    `bson:"maxFileSize" json:"maxFileSize"`       // 最大文件大小（KB）
	FileTypes      string `bson:"fileTypes" json:"fileTypes"`           // 文件类型
	ExcludeURLs    string `bson:"excludeUrls" json:"excludeUrls"`       // 排除的URL
	IncludeURLs    string `bson:"includeUrls" json:"includeUrls"`       // 包含的URL
	Authentication string `bson:"authentication" json:"authentication"` // 认证信息（JSON格式）
	
	// 文件检测相关配置
	FilePatterns     []string `bson:"filePatterns" json:"filePatterns"`         // 文件匹配模式，如 "*.txt", "**/*.log"
	ExcludePatterns  []string `bson:"excludePatterns" json:"excludePatterns"`   // 排除的文件模式
	RecursiveSearch  bool     `bson:"recursiveSearch" json:"recursiveSearch"`   // 是否递归搜索子目录
	FollowSymlinks   bool     `bson:"followSymlinks" json:"followSymlinks"`     // 是否跟踪符号链接
	MaxFileSizeBytes int64    `bson:"maxFileSizeBytes" json:"maxFileSizeBytes"` // 最大文件大小（字节）
	ScanArchives     bool     `bson:"scanArchives" json:"scanArchives"`         // 是否扫描压缩文件
}

// SensitiveDetectionRequest 敏感信息检测请求
type SensitiveDetectionRequest struct {
	ProjectID   primitive.ObjectID       `json:"projectId"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Targets     []string                 `json:"targets"`
	RuleGroups  []primitive.ObjectID     `json:"ruleGroups"`
	Rules       []primitive.ObjectID     `json:"rules"`
	Config      SensitiveDetectionConfig `json:"config"`
}

// SensitiveDetectionStatus 敏感信息检测状态
type SensitiveDetectionStatus string

const (
	SensitiveDetectionStatusPending   SensitiveDetectionStatus = "pending"
	SensitiveDetectionStatusRunning   SensitiveDetectionStatus = "running"
	SensitiveDetectionStatusCompleted SensitiveDetectionStatus = "completed"
	SensitiveDetectionStatusFailed    SensitiveDetectionStatus = "failed"
	SensitiveDetectionStatusCancelled SensitiveDetectionStatus = "cancelled"
)

// SensitiveFinding 敏感信息发现
type SensitiveFinding struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Target      string             `bson:"target" json:"target"`           // 目标URL或文件路径
	TargetType  string             `bson:"targetType" json:"targetType"`   // 目标类型：url, file, directory
	Rule        primitive.ObjectID `bson:"rule" json:"rule"`               // 规则ID
	RuleName    string             `bson:"ruleName" json:"ruleName"`       // 规则名称
	Category    string             `bson:"category" json:"category"`       // 分类
	RiskLevel   string             `bson:"riskLevel" json:"riskLevel"`     // 风险等级
	Pattern     string             `bson:"pattern" json:"pattern"`         // 匹配模式
	MatchedText string             `bson:"matchedText" json:"matchedText"` // 匹配文本
	Context     string             `bson:"context" json:"context"`         // 上下文
	LineNumber  int                `bson:"lineNumber" json:"lineNumber"`   // 行号（对于文件）
	FilePath    string             `bson:"filePath" json:"filePath"`       // 文件路径（相对路径）
	FileSize    int64              `bson:"fileSize" json:"fileSize"`       // 文件大小（字节）
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}

// SensitiveDetectionSummary 敏感信息检测摘要
type SensitiveDetectionSummary struct {
	TotalFindings  int            `bson:"totalFindings" json:"totalFindings"`   // 总发现数
	RiskLevelCount map[string]int `bson:"riskLevelCount" json:"riskLevelCount"` // 各风险等级数量
	CategoryCount  map[string]int `bson:"categoryCount" json:"categoryCount"`   // 各分类数量
}

// SensitiveDetectionResult 敏感信息检测结果
type SensitiveDetectionResult struct {
	ID          primitive.ObjectID        `bson:"_id" json:"id"`
	ProjectID   primitive.ObjectID        `bson:"projectId" json:"projectId"`
	Name        string                    `bson:"name" json:"name"`
	Targets     []string                  `bson:"targets" json:"targets"`
	Status      SensitiveDetectionStatus  `bson:"status" json:"status"`
	StartTime   time.Time                 `bson:"startTime" json:"startTime"`
	EndTime     time.Time                 `bson:"endTime" json:"endTime"`
	Progress    float64                   `bson:"progress" json:"progress"`
	Config      SensitiveDetectionConfig  `bson:"config" json:"config"`
	Findings    []*SensitiveFinding       `bson:"findings" json:"findings"`
	Summary     SensitiveDetectionSummary `bson:"summary" json:"summary"`
	CreatedAt   time.Time                 `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time                 `bson:"updatedAt" json:"updatedAt"`
	TotalCount  int                       `bson:"totalCount" json:"totalCount"`
	FinishCount int                       `bson:"finishCount" json:"finishCount"`
}

// SensitiveRuleCreateRequest 创建敏感规则请求
type SensitiveRuleCreateRequest struct {
	Name                  string   `json:"name" binding:"required"`
	Description           string   `json:"description"`
	Type                  string   `json:"type" binding:"required"`
	Pattern               string   `json:"pattern" binding:"required"`
	Category              string   `json:"category" binding:"required"`
	RiskLevel             string   `json:"riskLevel" binding:"required"`
	Tags                  []string `json:"tags"`
	Enabled               bool     `json:"enabled"`
	Context               int      `json:"context"`
	Examples              []string `json:"examples"`
	FalsePositivePatterns []string `json:"falsePositivePatterns"`
}

// SensitiveRuleUpdateRequest 更新敏感规则请求
type SensitiveRuleUpdateRequest struct {
	Name                  string   `json:"name"`
	Description           string   `json:"description"`
	Pattern               string   `json:"pattern"`
	Category              string   `json:"category"`
	RiskLevel             string   `json:"riskLevel"`
	Tags                  []string `json:"tags"`
	Enabled               bool     `json:"enabled"`
	Context               int      `json:"context"`
	Examples              []string `json:"examples"`
	FalsePositivePatterns []string `json:"falsePositivePatterns"`
}

// SensitiveRuleGroupCreateRequest 创建敏感规则组请求
type SensitiveRuleGroupCreateRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Rules       []string `json:"rules" binding:"required"`
	Enabled     bool     `json:"enabled"`
}

// SensitiveRuleGroupUpdateRequest 更新敏感规则组请求
type SensitiveRuleGroupUpdateRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Rules       []string `json:"rules"`
	Enabled     bool     `json:"enabled"`
}

// SensitiveWhitelistCreateRequest 创建敏感信息白名单请求
type SensitiveWhitelistCreateRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Type        string    `json:"type" binding:"required"`
	Value       string    `json:"value" binding:"required"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// SensitiveWhitelistUpdateRequest 更新敏感信息白名单请求
type SensitiveWhitelistUpdateRequest struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Value       string    `json:"value"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// TODO: 以下模型根据DEV_PLAN 0.6版本需要完善

// SensitiveInfoDetectionRequest 敏感信息检测请求（新增用于service）
type SensitiveInfoDetectionRequest struct {
	ProjectID   primitive.ObjectID       `json:"projectId"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Targets     []string                 `json:"targets"`
	Paths       []string                 `json:"paths"` // 文件路径列表
	RuleGroups  []primitive.ObjectID     `json:"ruleGroups"`
	Rules       []primitive.ObjectID     `json:"rules"`
	Config      SensitiveDetectionConfig `json:"config"`
}

// SensitiveInfoDetectionResponse 敏感信息检测响应
type SensitiveInfoDetectionResponse struct {
	Status         string    `json:"status"`
	Message        string    `json:"message"`
	DetectionID    string    `json:"detectionId"`
	EstimatedTime  int       `json:"estimatedTime"`
	CreatedAt      time.Time `json:"createdAt"`
}

// SensitiveInfoScanRequest 敏感信息扫描请求（用于替代现有的）
type SensitiveInfoScanRequest struct {
	ProjectID   primitive.ObjectID       `json:"projectId"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Targets     []string                 `json:"targets"`
	Paths       []string                 `json:"paths"` // 文件路径列表
	RuleGroups  []primitive.ObjectID     `json:"ruleGroups"`
	Rules       []primitive.ObjectID     `json:"rules"`
	Config      SensitiveDetectionConfig `json:"config"`
}

// SensitiveInfoScanResult 敏感信息扫描结果（扩展）
type SensitiveInfoScanResult struct {
	ID             primitive.ObjectID        `bson:"_id" json:"id"`
	ProjectID      primitive.ObjectID        `bson:"projectId" json:"projectId"`
	Name           string                    `bson:"name" json:"name"`
	Targets        []string                  `bson:"targets" json:"targets"`
	Status         string                    `bson:"status" json:"status"`
	Message        string                    `bson:"message" json:"message"`
	StartTime      time.Time                 `bson:"startTime" json:"startTime"`
	EndTime        time.Time                 `bson:"endTime" json:"endTime"`
	Progress       float64                   `bson:"progress" json:"progress"`
	Config         SensitiveDetectionConfig  `bson:"config" json:"config"`
	Findings       []*SensitiveFinding       `bson:"findings" json:"findings"`
	Summary        SensitiveDetectionSummary `bson:"summary" json:"summary"`
	TotalFiles     int                       `bson:"totalFiles" json:"totalFiles"`
	ProcessedFiles int                       `bson:"processedFiles" json:"processedFiles"`
	ScanResults    map[string]interface{}    `bson:"scanResults" json:"scanResults"`
	CreatedAt      time.Time                 `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time                 `bson:"updatedAt" json:"updatedAt"`
}
