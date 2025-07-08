package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PortScanStatus 定义端口扫描状态
type PortScanStatus string

const (
	// PortScanStatusPending 等待扫描
	PortScanStatusPending PortScanStatus = "pending"
	// PortScanStatusRunning 扫描中
	PortScanStatusRunning PortScanStatus = "running"
	// PortScanStatusCompleted 扫描完成
	PortScanStatusCompleted PortScanStatus = "completed"
	// PortScanStatusFailed 扫描失败
	PortScanStatusFailed PortScanStatus = "failed"
)

// PortScanTask 定义端口扫描任务
type PortScanTask struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID     primitive.ObjectID `bson:"projectId" json:"projectId"`
	Name          string             `bson:"name" json:"name"`
	Description   string             `bson:"description" json:"description"`
	Targets       []string           `bson:"targets" json:"targets"`
	Ports         []int              `bson:"ports" json:"ports"`
	Protocol      string             `bson:"protocol" json:"protocol"` // TCP, UDP
	RateLimit     int                `bson:"rateLimit" json:"rateLimit"`
	Timeout       int                `bson:"timeout" json:"timeout"`
	Status        PortScanStatus     `bson:"status" json:"status"`
	Progress      float64            `bson:"progress" json:"progress"`
	StartTime     time.Time          `bson:"startTime" json:"startTime"`
	EndTime       time.Time          `bson:"endTime" json:"endTime"`
	CompletedAt   time.Time          `bson:"completedAt" json:"completedAt"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
	CreatedBy     string             `bson:"createdBy" json:"createdBy"`
	NodeID        string             `bson:"nodeId" json:"nodeId"`
	Error         string             `bson:"error" json:"error"`
	ResultSummary PortScanSummary    `bson:"resultSummary" json:"resultSummary"`
	Config        PortScanConfig     `bson:"config" json:"config"`
}

// PortScanResult 定义端口扫描结果
type PortScanResult struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TaskID    primitive.ObjectID `bson:"taskId" json:"taskId"`
	ProjectID primitive.ObjectID `bson:"projectId" json:"projectId"`
	Host      string             `bson:"host" json:"host"`
	IP        string             `bson:"ip" json:"ip"`
	Port      int                `bson:"port" json:"port"`
	Protocol  string             `bson:"protocol" json:"protocol"` // TCP, UDP
	Status    string             `bson:"status" json:"status"`     // open, closed, filtered
	Service   string             `bson:"service" json:"service"`
	Product   string             `bson:"product" json:"product"`
	Banner    string             `bson:"banner" json:"banner"`
	Version   string             `bson:"version" json:"version"`
	ExtraInfo string             `bson:"extraInfo" json:"extraInfo"`
	TTL       int                `bson:"ttl" json:"ttl"`
	Reason    string             `bson:"reason" json:"reason"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// PortScanConfig 定义端口扫描配置
type PortScanConfig struct {
	DefaultPorts     []int    `bson:"defaultPorts" json:"defaultPorts"`
	DefaultProtocol  string   `bson:"defaultProtocol" json:"defaultProtocol"`
	DefaultRateLimit int      `bson:"defaultRateLimit" json:"defaultRateLimit"`
	DefaultTimeout   int      `bson:"defaultTimeout" json:"defaultTimeout"`
	MaxTargets       int      `bson:"maxTargets" json:"maxTargets"`
	MaxPorts         int      `bson:"maxPorts" json:"maxPorts"`
	EnableBanner     bool     `bson:"enableBanner" json:"enableBanner"`
	EnableService    bool     `bson:"enableService" json:"enableService"`
	EnableVersion    bool     `bson:"enableVersion" json:"enableVersion"`
	RateLimit        int      `bson:"rateLimit" json:"rateLimit"`
	Ports            string   `bson:"ports" json:"ports"`
	ScanType         string   `bson:"scanType" json:"scanType"`
	Concurrency      int      `bson:"concurrency" json:"concurrency"`
	Timeout          int      `bson:"timeout" json:"timeout"`
	ScanMethod       string   `bson:"scanMethod" json:"scanMethod"`
	ServiceDetection bool     `bson:"serviceDetection" json:"serviceDetection"`
	ExcludeHosts     []string `bson:"excludeHosts" json:"excludeHosts"`
	SaveToDB         bool     `bson:"saveToDB" json:"saveToDB"`
	RetryCount       int      `bson:"retryCount" json:"retryCount"`
}

// PortScanSummary 定义端口扫描摘要
type PortScanSummary struct {
	TotalHosts     int            `bson:"totalHosts" json:"totalHosts"`
	ProcessedHosts int            `bson:"processedHosts" json:"processedHosts"`
	OpenPorts      int            `bson:"openPorts" json:"openPorts"`
	ServiceStats   map[string]int `bson:"serviceStats" json:"serviceStats"`
}
