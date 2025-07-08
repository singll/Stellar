package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SubdomainEnumTask 子域名枚举任务
type SubdomainEnumTask struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	ProjectID     primitive.ObjectID   `bson:"projectId" json:"projectId"`
	RootDomain    string               `bson:"rootDomain" json:"rootDomain"`
	TaskName      string               `bson:"taskName" json:"taskName"`
	Status        string               `bson:"status" json:"status"` // pending, running, completed, failed
	CreatedAt     time.Time            `bson:"createdAt" json:"createdAt"`
	StartedAt     time.Time            `bson:"startedAt" json:"startedAt"`
	CompletedAt   time.Time            `bson:"completedAt" json:"completedAt"`
	Progress      float64              `bson:"progress" json:"progress"` // 0-100
	Config        SubdomainEnumConfig  `bson:"config" json:"config"`
	ResultSummary SubdomainEnumSummary `bson:"resultSummary" json:"resultSummary"`
	Error         string               `bson:"error" json:"error"`
	NodeID        string               `bson:"nodeId" json:"nodeId"` // 执行任务的节点ID
	Tags          []string             `bson:"tags" json:"tags"`
}

// SubdomainEnumConfig 子域名枚举配置
type SubdomainEnumConfig struct {
	Methods           []string `bson:"methods" json:"methods"`                     // 枚举方法: dns_brute, dns_zone_transfer, search_engines, certificate_transparency, etc.
	DictionaryPath    string   `bson:"dictionaryPath" json:"dictionaryPath"`       // 字典路径
	Concurrency       int      `bson:"concurrency" json:"concurrency"`             // 并发数
	Timeout           int      `bson:"timeout" json:"timeout"`                     // 超时时间(秒)
	RetryCount        int      `bson:"retryCount" json:"retryCount"`               // 重试次数
	ResolverServers   []string `bson:"resolverServers" json:"resolverServers"`     // DNS解析服务器
	RateLimit         int      `bson:"rateLimit" json:"rateLimit"`                 // 请求速率限制(每秒)
	IncludeWildcard   bool     `bson:"includeWildcard" json:"includeWildcard"`     // 是否包含泛解析域名
	VerifySubdomains  bool     `bson:"verifySubdomains" json:"verifySubdomains"`   // 是否验证子域名
	SaveToDB          bool     `bson:"saveToDB" json:"saveToDB"`                   // 是否保存到数据库
	RecursiveSearch   bool     `bson:"recursiveSearch" json:"recursiveSearch"`     // 是否递归搜索
	RecursiveDepth    int      `bson:"recursiveDepth" json:"recursiveDepth"`       // 递归深度
	IncludeThirdLevel bool     `bson:"includeThirdLevel" json:"includeThirdLevel"` // 是否包含三级域名
}

// SubdomainEnumSummary 子域名枚举结果摘要
type SubdomainEnumSummary struct {
	TotalFound       int            `bson:"totalFound" json:"totalFound"`             // 发现的子域名总数
	NewFound         int            `bson:"newFound" json:"newFound"`                 // 新发现的子域名数
	ResolvedCount    int            `bson:"resolvedCount" json:"resolvedCount"`       // 可解析的子域名数
	UnresolvedCount  int            `bson:"unresolvedCount" json:"unresolvedCount"`   // 不可解析的子域名数
	WildcardCount    int            `bson:"wildcardCount" json:"wildcardCount"`       // 泛解析域名数
	TakeOverCount    int            `bson:"takeOverCount" json:"takeOverCount"`       // 可能被接管的子域名数
	MethodStats      map[string]int `bson:"methodStats" json:"methodStats"`           // 各方法发现的子域名数
	ProcessedRecords int            `bson:"processedRecords" json:"processedRecords"` // 处理的记录数
}

// SubdomainResult 子域名枚举结果
type SubdomainResult struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TaskID       primitive.ObjectID `bson:"taskId" json:"taskId"`
	ProjectID    primitive.ObjectID `bson:"projectId" json:"projectId"`
	RootDomain   string             `bson:"rootDomain" json:"rootDomain"`
	Subdomain    string             `bson:"subdomain" json:"subdomain"`
	IPs          []string           `bson:"ips" json:"ips"`
	CNAME        string             `bson:"cname" json:"cname"`
	Type         string             `bson:"type" json:"type"` // A, AAAA, CNAME, etc.
	Records      []DNSRecord        `bson:"records" json:"records"`
	IsWildcard   bool               `bson:"isWildcard" json:"isWildcard"`
	IsResolved   bool               `bson:"isResolved" json:"isResolved"`
	IsTakeOver   bool               `bson:"isTakeOver" json:"isTakeOver"`
	TakeOverType string             `bson:"takeOverType" json:"takeOverType"`
	Source       string             `bson:"source" json:"source"` // 发现方法
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
	AssetID      primitive.ObjectID `bson:"assetId" json:"assetId"` // 关联的资产ID
}

// DNSRecord DNS记录
type DNSRecord struct {
	Type  string `bson:"type" json:"type"`   // A, AAAA, CNAME, TXT, MX, NS, etc.
	Value string `bson:"value" json:"value"` // 记录值
	TTL   int    `bson:"ttl" json:"ttl"`     // 生存时间
}

// SubdomainDictionary 子域名字典
type SubdomainDictionary struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Count       int                `bson:"count" json:"count"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
	FilePath    string             `bson:"filePath" json:"filePath"`
	IsDefault   bool               `bson:"isDefault" json:"isDefault"`
	Tags        []string           `bson:"tags" json:"tags"`
}

// SubdomainTakeOverRule 子域名接管规则
type SubdomainTakeOverRule struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Service      string             `bson:"service" json:"service"`           // 服务名称
	Fingerprint  string             `bson:"fingerprint" json:"fingerprint"`   // 指纹
	CNAMEPattern string             `bson:"cnamePattern" json:"cnamePattern"` // CNAME模式
	Status       string             `bson:"status" json:"status"`             // 状态: active, inactive
	Description  string             `bson:"description" json:"description"`   // 描述
	Reference    string             `bson:"reference" json:"reference"`       // 参考链接
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}
