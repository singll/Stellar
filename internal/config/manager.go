package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manager 配置管理器
type Manager struct {
	config *Config
	env    string
}

// NewManager 创建配置管理器
func NewManager(env string) *Manager {
	if env == "" {
		env = getEnv("APP_ENV", "development")
	}

	return &Manager{
		env: env,
	}
}

// LoadFile 加载指定的配置文件
func (m *Manager) LoadFile(configFile string) error {
	// 尝试读取指定文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", configFile, err)
	}

	// 应用环境变量覆盖
	m.applyEnvironmentOverrides(&config)

	// 验证配置
	err = m.validateConfig(&config)
	if err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	m.config = &config
	return nil
}
func (m *Manager) Load(configDir string) error {
	// 配置文件优先级：
	// 1. config.{env}.yaml
	// 2. config.yaml
	// 3. 默认配置

	configFiles := []string{
		filepath.Join(configDir, fmt.Sprintf("config.%s.yaml", m.env)),
		filepath.Join(configDir, "config.yaml"),
	}

	var config Config
	loaded := false

	// 尝试加载配置文件
	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); err == nil {
			data, err := os.ReadFile(configFile)
			if err != nil {
				return fmt.Errorf("failed to read config file %s: %w", configFile, err)
			}

			err = yaml.Unmarshal(data, &config)
			if err != nil {
				return fmt.Errorf("failed to parse config file %s: %w", configFile, err)
			}

			loaded = true
			break
		}
	}

	// 如果没有找到配置文件，使用默认配置
	if !loaded {
		config = m.getDefaultConfig()
	}

	// 应用环境变量覆盖
	m.applyEnvironmentOverrides(&config)

	// 验证配置
	err := m.validateConfig(&config)
	if err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	m.config = &config
	return nil
}

// Get 获取配置
func (m *Manager) Get() *Config {
	return m.config
}

// GetEnv 获取当前环境
func (m *Manager) GetEnv() string {
	return m.env
}

// IsDevelopment 是否为开发环境
func (m *Manager) IsDevelopment() bool {
	return m.env == "development" || m.env == "dev"
}

// IsProduction 是否为生产环境
func (m *Manager) IsProduction() bool {
	return m.env == "production" || m.env == "prod"
}

// IsTest 是否为测试环境
func (m *Manager) IsTest() bool {
	return m.env == "test" || m.env == "testing"
}

// getDefaultConfig 获取默认配置
func (m *Manager) getDefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
			Mode: "debug",
		},
		Database: DatabaseConfig{
			Type:     "sqlite",
			Path:     "./stellar.db",
			Host:     "localhost",
			Port:     5432,
			Database: "stellar",
			Username: "",
			Password: "",
			SSLMode:  "disable",
		},
		MongoDB: MongoDBConfig{
			URI:           "mongodb://localhost:27017",
			Database:      "stellar",
			MaxPoolSize:   100,
			MinPoolSize:   10,
			MaxIdleTimeMS: 30000,
		},
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
		Auth: AuthConfig{
			JWTSecret:   "stellar-server-secret",
			TokenExpiry: 24,
		},
		System: SystemConfig{
			Timezone: "UTC",
		},
	}
}

// applyEnvironmentOverrides 应用环境变量覆盖
func (m *Manager) applyEnvironmentOverrides(config *Config) {
	// 服务器配置
	if host := os.Getenv("SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		var p int
		if _, err := fmt.Sscanf(port, "%d", &p); err == nil && p > 0 {
			config.Server.Port = p
		}
	}
	if mode := os.Getenv("SERVER_MODE"); mode != "" {
		config.Server.Mode = mode
	}

	// 数据库配置
	if dbType := os.Getenv("DB_TYPE"); dbType != "" {
		config.Database.Type = dbType
	}
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		var p int
		if _, err := fmt.Sscanf(dbPort, "%d", &p); err == nil && p > 0 {
			config.Database.Port = p
		}
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		config.Database.Database = dbName
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		config.Database.Username = dbUser
	}
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		config.Database.Password = dbPass
	}
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		config.Database.Path = dbPath
	}

	// MongoDB配置
	if mongoURI := os.Getenv("MONGODB_URI"); mongoURI != "" {
		config.MongoDB.URI = mongoURI
	}
	if mongoDB := os.Getenv("MONGODB_DATABASE"); mongoDB != "" {
		config.MongoDB.Database = mongoDB
	}

	// Redis配置
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		config.Redis.Addr = redisAddr
	}
	if redisPass := os.Getenv("REDIS_PASSWORD"); redisPass != "" {
		config.Redis.Password = redisPass
	}

	// 认证配置
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.Auth.JWTSecret = jwtSecret
	}
}

// validateConfig 验证配置
func (m *Manager) validateConfig(config *Config) error {
	// 验证服务器配置
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// 验证数据库配置
	validDBTypes := []string{"mysql", "postgres", "sqlite", "mongodb"}

	if !contains(validDBTypes, config.Database.Type) {
		return fmt.Errorf("invalid database type: %s", config.Database.Type)
	}

	if config.Database.Type == "sqlite" && config.Database.Path == "" {
		return fmt.Errorf("sqlite database path is required")
	}

	// 验证认证配置
	if config.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if config.Auth.TokenExpiry <= 0 {
		return fmt.Errorf("invalid token expiry: %d", config.Auth.TokenExpiry)
	}

	return nil
}

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}
