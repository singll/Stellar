package config

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"io/ioutil"

	"gopkg.in/yaml.v3"
)

const (
	VERSION         = "1.0"
	UPDATEURL       = "http://update.stellar-server.top"
	REMOTE_REPO_URL = "https://github.com/StellarServer/StellarServer.git"
)

var (
	SECRET_KEY       = "StellarServer-15847412364125411"
	MONGODB_IP       string
	MONGODB_PORT     int
	MONGODB_DATABASE string
	MONGODB_USER     string
	MONGODB_PASSWORD string
	REDIS_IP         string
	REDIS_PORT       string
	REDIS_PASSWORD   string
	TIMEZONE         string = "Asia/Shanghai"
	LOG_INFO                = make(map[string][]string)
	GET_LOG_NAME            = []string{}
	NODE_TIMEOUT            = 50
	TOTAL_LOGS              = 1000
	APP                     = make(map[string]interface{})
	Project_List            = make(map[string]string)
	PLUGINKEY        string
)

// SetTimezone 设置时区
func SetTimezone(t string) {
	TIMEZONE = t
}

// GetTimezone 获取时区
func GetTimezone() string {
	return TIMEZONE
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// Config 应用程序配置
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`  // 新增统一数据库配置
	MongoDB   MongoDBConfig   `yaml:"mongodb"`   // 保持向后兼容
	Redis     RedisConfig     `yaml:"redis"`
	Auth      AuthConfig      `yaml:"auth"`
	Subdomain SubdomainConfig `yaml:"subdomain"`
	PortScan  PortScanConfig  `yaml:"portscan"`
	VulnScan  VulnScanConfig  `yaml:"vulnscan"`
	Discovery DiscoveryConfig `yaml:"discovery"`
	Node      NodeConfig      `yaml:"node"`
	Task      TaskConfig      `yaml:"task"`
	Logs      LogsConfig      `yaml:"logs"`
	System    SystemConfig    `yaml:"system"`
}

// DatabaseConfig 统一数据库配置
type DatabaseConfig struct {
	Type     string `yaml:"type"`     // mysql, postgres, sqlite, mongodb
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
	Path     string `yaml:"path"` // for sqlite
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	Mode         string        `yaml:"mode"`
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
}

// MongoDBConfig MongoDB配置
type MongoDBConfig struct {
	URI           string `yaml:"uri"`
	Database      string `yaml:"database"`
	MaxPoolSize   uint64 `yaml:"maxPoolSize"`
	MinPoolSize   uint64 `yaml:"minPoolSize"`
	MaxIdleTimeMS int    `yaml:"maxIdleTimeMS"`
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr         string `yaml:"addr"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"poolSize"`
	MinIdleConns int    `yaml:"minIdleConns"`
	MaxConnAgeMS int    `yaml:"maxConnAgeMS"`
	Port         string `yaml:"port"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret   string `yaml:"jwtSecret"`
	TokenExpiry int    `yaml:"tokenExpiry"`
}

// SubdomainConfig 子域名枚举配置
type SubdomainConfig struct {
	Methods          []string `yaml:"methods"`          // 枚举方法
	DictionaryPath   string   `yaml:"dictionaryPath"`   // 字典路径
	Concurrency      int      `yaml:"concurrency"`      // 最大并发数
	Timeout          int      `yaml:"timeout"`          // 单位: 秒
	RetryCount       int      `yaml:"retryCount"`       // 重试次数
	RateLimit        int      `yaml:"rateLimit"`        // 每秒请求数
	ResolverServers  []string `yaml:"resolverServers"`  // DNS解析服务器
	VerifySubdomains bool     `yaml:"verifySubdomains"` // 是否验证子域名
	CheckTakeover    bool     `yaml:"checkTakeover"`    // 是否检查子域名接管
	RecursiveSearch  bool     `yaml:"recursiveSearch"`  // 是否递归搜索
}

// PortScanConfig 端口扫描配置
type PortScanConfig struct {
	Ports       string   `yaml:"ports"`       // 端口范围
	Concurrency int      `yaml:"concurrency"` // 并发数
	Timeout     int      `yaml:"timeout"`     // 单位: 秒
	RetryCount  int      `yaml:"retryCount"`  // 重试次数
	RateLimit   int      `yaml:"rateLimit"`   // 每秒请求数
	ScanType    string   `yaml:"scanType"`    // 扫描类型
	ExcludeIPs  []string `yaml:"excludeIPs"`  // 排除的IP地址
}

// VulnScanConfig 漏洞扫描配置
type VulnScanConfig struct {
	Timeout        int    `yaml:"timeout"`        // 单位: 秒
	MaxConcurrency int    `yaml:"maxConcurrency"` // 最大并发数
	PluginPath     string `yaml:"pluginPath"`     // 插件路径
}

// DiscoveryConfig 资产发现配置
type DiscoveryConfig struct {
	Timeout        int `yaml:"timeout"`        // 单位: 秒
	MaxConcurrency int `yaml:"maxConcurrency"` // 最大并发数
	ScanInterval   int `yaml:"scanInterval"`   // 扫描间隔(小时)
}

// NodeConfig 节点管理配置
type NodeConfig struct {
	HeartbeatInterval int    `yaml:"heartbeatInterval"` // 心跳间隔(秒)
	HeartbeatTimeout  int    `yaml:"heartbeatTimeout"`  // 心跳超时(秒)
	EnableAutoRemove  bool   `yaml:"enableAutoRemove"`  // 是否自动移除离线节点
	AutoRemoveAfter   int    `yaml:"autoRemoveAfter"`   // 自动移除离线节点的时间(秒)
	MasterNodeName    string `yaml:"masterNodeName"`    // 主节点名称
}

// TaskConfig 任务管理配置
type TaskConfig struct {
	MaxConcurrentTasks int  `yaml:"maxConcurrentTasks"` // 最大并发任务数
	TaskTimeout        int  `yaml:"taskTimeout"`        // 任务超时时间(秒)
	EnableRetry        bool `yaml:"enableRetry"`        // 是否启用重试
	MaxRetries         int  `yaml:"maxRetries"`         // 最大重试次数
	RetryInterval      int  `yaml:"retryInterval"`      // 重试间隔(秒)
	QueueCapacity      int  `yaml:"queueCapacity"`      // 队列容量
}

// LogsConfig 日志配置
type LogsConfig struct {
	TotalLogs int `yaml:"totalLogs"` // 日志总数
}

// SystemConfig 系统配置
type SystemConfig struct {
	Timezone string `yaml:"timezone"` // 时区
}

// SetConfig 设置配置
func SetConfig() {
	SECRET_KEY = GenerateRandomString(16)

	// 读取或生成插件密钥
	if _, err := os.Stat("PLUGINKEY"); err == nil {
		data, err := os.ReadFile("PLUGINKEY")
		if err == nil {
			PLUGINKEY = string(data)
		}
	} else {
		PLUGINKEY = GenerateRandomString(6)
		err := os.WriteFile("PLUGINKEY", []byte(PLUGINKEY), 0644)
		if err != nil {
			panic(err)
		}
	}

	// 读取配置文件
	configFilePath := "config.yaml"
	if _, err := os.Stat(configFilePath); err == nil {
		data, err := os.ReadFile(configFilePath)
		if err != nil {
			panic(err)
		}

		var cfg Config
		err = yaml.Unmarshal(data, &cfg)
		if err != nil {
			panic(err)
		}

		MONGODB_IP = cfg.MongoDB.URI
		MONGODB_PORT, _ = strconv.Atoi(cfg.MongoDB.Database)
		MONGODB_DATABASE = cfg.MongoDB.Database
		MONGODB_USER = cfg.MongoDB.Username
		MONGODB_PASSWORD = cfg.MongoDB.Password
		REDIS_IP = cfg.Redis.Addr
		REDIS_PORT = cfg.Redis.Port
		REDIS_PASSWORD = cfg.Redis.Password
		TOTAL_LOGS = cfg.Logs.TotalLogs
		TIMEZONE = cfg.System.Timezone

		// 检查环境变量是否覆盖配置
		if envUser, exists := os.LookupEnv("MONGODB_USER"); exists && envUser != MONGODB_USER {
			MONGODB_USER = envUser
		}
		if envPass, exists := os.LookupEnv("MONGODB_PASSWORD"); exists && envPass != MONGODB_PASSWORD {
			MONGODB_PASSWORD = envPass
		}
		if envRedisPass, exists := os.LookupEnv("REDIS_PASSWORD"); exists && envRedisPass != REDIS_PASSWORD {
			REDIS_PASSWORD = envRedisPass
		}
	} else {
		// 使用环境变量或默认值
		TIMEZONE = getEnv("TIMEZONE", "Asia/Shanghai")
		MONGODB_IP = getEnv("MONGODB_IP", "127.0.0.1")
		MONGODB_PORT, _ = strconv.Atoi(getEnv("MONGODB_PORT", "27017"))
		MONGODB_DATABASE = getEnv("MONGODB_DATABASE", "StellarServer")
		MONGODB_USER = getEnv("MONGODB_USER", "root")
		MONGODB_PASSWORD = getEnv("MONGODB_PASSWORD", "QckSdkg5CKvtxfec")
		REDIS_IP = getEnv("REDIS_IP", "127.0.0.1")
		REDIS_PORT = getEnv("REDIS_PORT", "6379")
		REDIS_PASSWORD = getEnv("REDIS_PASSWORD", "StellarServer")
		TOTAL_LOGS = 1000

		// 创建新的配置文件
		cfg := Config{}
		cfg.System.Timezone = TIMEZONE
		cfg.MongoDB.URI = MONGODB_IP
		cfg.MongoDB.Database = strconv.Itoa(MONGODB_PORT)
		cfg.MongoDB.Username = MONGODB_USER
		cfg.MongoDB.Password = MONGODB_PASSWORD
		cfg.Redis.Addr = REDIS_IP
		cfg.Redis.Port = REDIS_PORT
		cfg.Redis.Password = REDIS_PASSWORD
		cfg.Logs.TotalLogs = TOTAL_LOGS

		data, err := yaml.Marshal(&cfg)
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(configFilePath, data, 0644)
		if err != nil {
			panic(err)
		}
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// applyEnvironmentOverrides 应用环境变量覆盖配置
func applyEnvironmentOverrides(config *Config) {
	// MongoDB配置
	if uri := os.Getenv("MONGODB_URI"); uri != "" {
		config.MongoDB.URI = uri
	}
	if db := os.Getenv("MONGODB_DATABASE"); db != "" {
		config.MongoDB.Database = db
	}

	// Redis配置
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		config.Redis.Addr = addr
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		config.Redis.Password = password
	}

	// 认证配置
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		config.Auth.JWTSecret = secret
	}

	// 服务器配置
	if port := os.Getenv("SERVER_PORT"); port != "" {
		var p int
		if _, err := fmt.Sscanf(port, "%d", &p); err == nil && p > 0 {
			config.Server.Port = p
		}
	}
}
