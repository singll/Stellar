package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AssetType 定义资产类型
type AssetType string

const (
	AssetTypeDomain    AssetType = "domain"    // 域名
	AssetTypeSubdomain AssetType = "subdomain" // 子域名
	AssetTypeIP        AssetType = "ip"        // IP地址
	AssetTypePort      AssetType = "port"      // 端口
	AssetTypeURL       AssetType = "url"       // URL
	AssetTypeApp       AssetType = "app"       // 应用
	AssetTypeMiniApp   AssetType = "miniapp"   // 小程序
	AssetTypeHTTP      AssetType = "http"      // HTTP服务
	AssetTypeOther     AssetType = "other"     // 其他服务
)

// BaseAsset 定义所有资产类型共有的基础字段
type BaseAsset struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
	LastScanTime  time.Time          `bson:"lastScanTime" json:"lastScanTime"`
	Type          AssetType          `bson:"type" json:"type"`
	ProjectID     primitive.ObjectID `bson:"projectId" json:"projectId"`
	Tags          []string           `bson:"tags" json:"tags"`
	TaskName      string             `bson:"taskName" json:"taskName"`
	RootDomain    string             `bson:"rootDomain" json:"rootDomain"`
	ChangeHistory []AssetChange      `bson:"changeHistory" json:"changeHistory"`
}

// AssetChange 记录资产变更历史
type AssetChange struct {
	Time       time.Time              `bson:"time" json:"time"`
	FieldName  string                 `bson:"fieldName" json:"fieldName"`
	OldValue   interface{}            `bson:"oldValue" json:"oldValue"`
	NewValue   interface{}            `bson:"newValue" json:"newValue"`
	ChangeType string                 `bson:"changeType" json:"changeType"` // add, update, delete
	Metadata   map[string]interface{} `bson:"metadata" json:"metadata"`
}

// DomainAsset 定义域名资产
type DomainAsset struct {
	BaseAsset `bson:",inline"`
	Domain    string   `bson:"domain" json:"domain"`
	IPs       []string `bson:"ips" json:"ips"`
	Whois     string   `bson:"whois" json:"whois"`
	ICPInfo   *ICPInfo `bson:"icpInfo" json:"icpInfo"`
}

// ICPInfo 定义ICP备案信息
type ICPInfo struct {
	ICPNo       string `bson:"icpNo" json:"icpNo"`             // 备案号
	CompanyName string `bson:"companyName" json:"companyName"` // 公司名称
	CompanyType string `bson:"companyType" json:"companyType"` // 公司类型
	UpdateTime  string `bson:"updateTime" json:"updateTime"`   // 更新时间
}

// SubdomainAsset 定义子域名资产
type SubdomainAsset struct {
	BaseAsset `bson:",inline"`
	Host      string   `bson:"host" json:"host"`
	IPs       []string `bson:"ips" json:"ips"`
	CNAME     string   `bson:"cname" json:"cname"`
	Type      string   `bson:"dnsType" json:"dnsType"`   // A, AAAA, CNAME, etc.
	Value     []string `bson:"value" json:"value"`       // DNS解析值
	TakeOver  bool     `bson:"takeOver" json:"takeOver"` // 是否可能被接管
}

// IPAsset 定义IP资产
type IPAsset struct {
	BaseAsset   `bson:",inline"`
	IP          string                 `bson:"ip" json:"ip"`
	Location    *IPLocation            `bson:"location" json:"location"`
	ASN         string                 `bson:"asn" json:"asn"`
	ISP         string                 `bson:"isp" json:"isp"`
	Fingerprint map[string]interface{} `bson:"fingerprint" json:"fingerprint"`
}

// IPLocation 定义IP地理位置信息
type IPLocation struct {
	Country     string  `bson:"country" json:"country"`
	CountryCode string  `bson:"countryCode" json:"countryCode"`
	Region      string  `bson:"region" json:"region"`
	City        string  `bson:"city" json:"city"`
	Latitude    float64 `bson:"latitude" json:"latitude"`
	Longitude   float64 `bson:"longitude" json:"longitude"`
}

// PortAsset 定义端口资产
type PortAsset struct {
	BaseAsset `bson:",inline"`
	IP        string `bson:"ip" json:"ip"`
	Host      string `bson:"host" json:"host"`
	Port      int    `bson:"port" json:"port"`
	Service   string `bson:"service" json:"service"`
	Protocol  string `bson:"protocol" json:"protocol"` // TCP, UDP
	Version   string `bson:"version" json:"version"`
	Banner    string `bson:"banner" json:"banner"`
	TLS       bool   `bson:"tls" json:"tls"`
	Transport string `bson:"transport" json:"transport"`
	Status    string `bson:"status" json:"status"` // open, closed, filtered
}

// URLAsset 定义URL资产
type URLAsset struct {
	BaseAsset     `bson:",inline"`
	URL           string                 `bson:"url" json:"url"`
	Host          string                 `bson:"host" json:"host"`
	Path          string                 `bson:"path" json:"path"`
	Query         string                 `bson:"query" json:"query"`
	Fragment      string                 `bson:"fragment" json:"fragment"`
	StatusCode    int                    `bson:"statusCode" json:"statusCode"`
	Title         string                 `bson:"title" json:"title"`
	ContentType   string                 `bson:"contentType" json:"contentType"`
	ContentLength int                    `bson:"contentLength" json:"contentLength"`
	Hash          string                 `bson:"hash" json:"hash"`                 // 页面内容hash
	Screenshot    string                 `bson:"screenshot" json:"screenshot"`     // 截图路径
	Technologies  []string               `bson:"technologies" json:"technologies"` // 使用的技术栈
	Headers       map[string]string      `bson:"headers" json:"headers"`           // HTTP响应头
	Favicon       *FaviconInfo           `bson:"favicon" json:"favicon"`           // favicon信息
	Metadata      map[string]interface{} `bson:"metadata" json:"metadata"`
}

// FaviconInfo 定义favicon信息
type FaviconInfo struct {
	Path    string `bson:"path" json:"path"`
	MMH3    string `bson:"mmh3" json:"mmh3"` // MurmurHash3哈希值
	Content string `bson:"content" json:"content"`
}

// HTTPAsset 定义HTTP服务资产
type HTTPAsset struct {
	BaseAsset     `bson:",inline"`
	Host          string                 `bson:"host" json:"host"`
	IP            string                 `bson:"ip" json:"ip"`
	Port          int                    `bson:"port" json:"port"`
	URL           string                 `bson:"url" json:"url"`
	Title         string                 `bson:"title" json:"title"`
	StatusCode    int                    `bson:"statusCode" json:"statusCode"`
	ContentType   string                 `bson:"contentType" json:"contentType"`
	ContentLength int                    `bson:"contentLength" json:"contentLength"`
	WebServer     string                 `bson:"webServer" json:"webServer"`
	TLS           bool                   `bson:"tls" json:"tls"`
	Hash          string                 `bson:"hash" json:"hash"`
	CDNName       string                 `bson:"cdnName" json:"cdnName"`
	CDN           bool                   `bson:"cdn" json:"cdn"`
	Screenshot    string                 `bson:"screenshot" json:"screenshot"`
	Technologies  []string               `bson:"technologies" json:"technologies"`
	Headers       map[string]string      `bson:"headers" json:"headers"`
	Favicon       *FaviconInfo           `bson:"favicon" json:"favicon"`
	JARM          string                 `bson:"jarm" json:"jarm"` // TLS指纹
	Metadata      map[string]interface{} `bson:"metadata" json:"metadata"`
}

// AppAsset 定义移动应用资产
type AppAsset struct {
	BaseAsset   `bson:",inline"`
	AppName     string                 `bson:"appName" json:"appName"`
	PackageName string                 `bson:"packageName" json:"packageName"`
	Platform    string                 `bson:"platform" json:"platform"` // iOS, Android
	Version     string                 `bson:"version" json:"version"`
	Developer   string                 `bson:"developer" json:"developer"`
	DownloadURL string                 `bson:"downloadUrl" json:"downloadUrl"`
	Description string                 `bson:"description" json:"description"`
	Permissions []string               `bson:"permissions" json:"permissions"`
	SHA256      string                 `bson:"sha256" json:"sha256"`
	IconURL     string                 `bson:"iconUrl" json:"iconUrl"`
	Metadata    map[string]interface{} `bson:"metadata" json:"metadata"`
}

// MiniAppAsset 定义小程序资产
type MiniAppAsset struct {
	BaseAsset   `bson:",inline"`
	AppName     string                 `bson:"appName" json:"appName"`
	AppID       string                 `bson:"appId" json:"appId"`
	Platform    string                 `bson:"platform" json:"platform"` // 微信, 支付宝, 百度等
	Developer   string                 `bson:"developer" json:"developer"`
	Description string                 `bson:"description" json:"description"`
	IconURL     string                 `bson:"iconUrl" json:"iconUrl"`
	QRCodeURL   string                 `bson:"qrCodeUrl" json:"qrCodeUrl"`
	Metadata    map[string]interface{} `bson:"metadata" json:"metadata"`
}

// AssetRelation 定义资产之间的关系
type AssetRelation struct {
	ID            primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	SourceAssetID primitive.ObjectID     `bson:"sourceAssetId" json:"sourceAssetId"`
	TargetAssetID primitive.ObjectID     `bson:"targetAssetId" json:"targetAssetId"`
	RelationType  string                 `bson:"relationType" json:"relationType"` // contains, resolves_to, redirects_to, etc.
	CreatedAt     time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time              `bson:"updatedAt" json:"updatedAt"`
	ProjectID     primitive.ObjectID     `bson:"projectId" json:"projectId"`
	Metadata      map[string]interface{} `bson:"metadata" json:"metadata"`
}

// AssetCollection 返回资产集合名称
func AssetCollection(assetType AssetType) string {
	switch assetType {
	case AssetTypeDomain:
		return "domain_assets"
	case AssetTypeSubdomain:
		return "subdomain_assets"
	case AssetTypeIP:
		return "ip_assets"
	case AssetTypePort:
		return "port_assets"
	case AssetTypeURL:
		return "url_assets"
	case AssetTypeHTTP:
		return "http_assets"
	case AssetTypeApp:
		return "app_assets"
	case AssetTypeMiniApp:
		return "miniapp_assets"
	default:
		return "other_assets"
	}
}
