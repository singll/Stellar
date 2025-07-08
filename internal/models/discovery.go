package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AssetDiscoveryTask 资产发现任务
type AssetDiscoveryTask struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID     primitive.ObjectID `bson:"projectId" json:"projectId"`
	TaskName      string             `bson:"taskName" json:"taskName"`
	DiscoveryType string             `bson:"discoveryType" json:"discoveryType"` // network, service, web
	Targets       []string           `bson:"targets" json:"targets"`             // IP段、域名等
	Status        string             `bson:"status" json:"status"`               // pending, running, completed, failed, stopped
	Progress      float64            `bson:"progress" json:"progress"`           // 0-100
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	StartedAt     time.Time          `bson:"startedAt" json:"startedAt"`
	CompletedAt   time.Time          `bson:"completedAt" json:"completedAt"`
	Config        DiscoveryConfig    `bson:"config" json:"config"`
	ResultSummary DiscoverySummary   `bson:"resultSummary" json:"resultSummary"`
	Error         string             `bson:"error" json:"error"`
	NodeID        string             `bson:"nodeId" json:"nodeId"` // 执行任务的节点ID
	Tags          []string           `bson:"tags" json:"tags"`
}

// DiscoveryConfig 资产发现配置
type DiscoveryConfig struct {
	Concurrency    int               `bson:"concurrency" json:"concurrency"`       // 并发数
	Timeout        int               `bson:"timeout" json:"timeout"`               // 超时时间(秒)
	RetryCount     int               `bson:"retryCount" json:"retryCount"`         // 重试次数
	RateLimit      int               `bson:"rateLimit" json:"rateLimit"`           // 请求速率限制(每秒)
	FollowRedirect bool              `bson:"followRedirect" json:"followRedirect"` // 是否跟随重定向
	CustomHeaders  map[string]string `bson:"customHeaders" json:"customHeaders"`   // 自定义请求头
	Cookies        string            `bson:"cookies" json:"cookies"`               // Cookie
	Proxy          string            `bson:"proxy" json:"proxy"`                   // 代理
	ScanDepth      int               `bson:"scanDepth" json:"scanDepth"`           // 扫描深度
	PortRanges     []string          `bson:"portRanges" json:"portRanges"`         // 端口范围
	ExcludeIPs     []string          `bson:"excludeIPs" json:"excludeIPs"`         // 排除的IP
	OnlyAliveHosts bool              `bson:"onlyAliveHosts" json:"onlyAliveHosts"` // 只扫描活跃主机
	ServiceDetect  bool              `bson:"serviceDetect" json:"serviceDetect"`   // 是否进行服务检测
	OSDetect       bool              `bson:"osDetect" json:"osDetect"`             // 是否进行操作系统检测
}

// DiscoverySummary 资产发现结果摘要
type DiscoverySummary struct {
	TotalTargets     int            `bson:"totalTargets" json:"totalTargets"`         // 目标总数
	ScannedTargets   int            `bson:"scannedTargets" json:"scannedTargets"`     // 已扫描目标数
	DiscoveredAssets int            `bson:"discoveredAssets" json:"discoveredAssets"` // 发现的资产数
	HostCount        int            `bson:"hostCount" json:"hostCount"`               // 主机数
	ServiceCount     int            `bson:"serviceCount" json:"serviceCount"`         // 服务数
	WebAppCount      int            `bson:"webAppCount" json:"webAppCount"`           // Web应用数
	OSDistribution   map[string]int `bson:"osDistribution" json:"osDistribution"`     // 操作系统分布
	ServiceTypes     map[string]int `bson:"serviceTypes" json:"serviceTypes"`         // 服务类型分布
	Domains          []string       `bson:"domains" json:"domains"`                   // 发现的域名
}

// DiscoveryResult 资产发现结果
type DiscoveryResult struct {
	ID           primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	TaskID       primitive.ObjectID     `bson:"taskId" json:"taskId"`
	Target       string                 `bson:"target" json:"target"`
	AssetType    string                 `bson:"assetType" json:"assetType"` // host, service, webapp
	AssetID      primitive.ObjectID     `bson:"assetId" json:"assetId"`     // 关联的资产ID
	IP           string                 `bson:"ip" json:"ip"`
	Hostname     string                 `bson:"hostname" json:"hostname"`
	OS           string                 `bson:"os" json:"os"`
	OSVersion    string                 `bson:"osVersion" json:"osVersion"`
	Ports        []PortInfo             `bson:"ports" json:"ports"`
	Services     []ServiceInfo          `bson:"services" json:"services"`
	WebApps      []WebAppInfo           `bson:"webApps" json:"webApps"`
	IsAlive      bool                   `bson:"isAlive" json:"isAlive"`
	FirstSeen    time.Time              `bson:"firstSeen" json:"firstSeen"`
	LastSeen     time.Time              `bson:"lastSeen" json:"lastSeen"`
	Tags         []string               `bson:"tags" json:"tags"`
	Screenshot   string                 `bson:"screenshot" json:"screenshot"`
	Notes        string                 `bson:"notes" json:"notes"`
	CustomFields map[string]interface{} `bson:"customFields" json:"customFields"`
}

// PortInfo 端口信息
type PortInfo struct {
	Port     int    `bson:"port" json:"port"`
	Protocol string `bson:"protocol" json:"protocol"` // tcp, udp
	State    string `bson:"state" json:"state"`       // open, closed, filtered
	Service  string `bson:"service" json:"service"`
	Version  string `bson:"version" json:"version"`
	Banner   string `bson:"banner" json:"banner"`
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name        string `bson:"name" json:"name"`
	Port        int    `bson:"port" json:"port"`
	Protocol    string `bson:"protocol" json:"protocol"`
	Version     string `bson:"version" json:"version"`
	Product     string `bson:"product" json:"product"`
	ExtraInfo   string `bson:"extraInfo" json:"extraInfo"`
	Fingerprint string `bson:"fingerprint" json:"fingerprint"`
}

// WebAppInfo Web应用信息
type WebAppInfo struct {
	URL          string            `bson:"url" json:"url"`
	Title        string            `bson:"title" json:"title"`
	StatusCode   int               `bson:"statusCode" json:"statusCode"`
	Server       string            `bson:"server" json:"server"`
	Technologies []string          `bson:"technologies" json:"technologies"`
	Headers      map[string]string `bson:"headers" json:"headers"`
	Cookies      []string          `bson:"cookies" json:"cookies"`
	Screenshot   string            `bson:"screenshot" json:"screenshot"`
	Favicon      string            `bson:"favicon" json:"favicon"`
	FaviconHash  string            `bson:"faviconHash" json:"faviconHash"`
}
