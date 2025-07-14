package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// PluginSDK 插件开发SDK
type PluginSDK struct {
	context PluginContext
	logger  *Logger
	config  *Config
	api     *APIClient
}

// PluginContext 插件上下文
type PluginContext struct {
	PluginID    string                 `json:"plugin_id"`
	Version     string                 `json:"version"`
	Params      map[string]interface{} `json:"params"`
	Environment string                 `json:"environment"` // development, testing, production
	WorkDir     string                 `json:"work_dir"`
	TempDir     string                 `json:"temp_dir"`
	DataDir     string                 `json:"data_dir"`
	LogLevel    string                 `json:"log_level"`
}

// Config 插件配置
type Config struct {
	Host      string            `json:"host"`
	Port      int               `json:"port"`
	APIKey    string            `json:"api_key"`
	Timeout   time.Duration     `json:"timeout"`
	Headers   map[string]string `json:"headers"`
	EnableSSL bool              `json:"enable_ssl"`
}

// Logger 插件日志器
type Logger struct {
	level     LogLevel
	output    *os.File
	prefix    string
	enableBot bool
}

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// APIClient API客户端
type APIClient struct {
	baseURL string
	apiKey  string
	headers map[string]string
	timeout time.Duration
}

// Result 插件执行结果
type Result struct {
	Success   bool                   `json:"success"`
	Data      interface{}            `json:"data"`
	Message   string                 `json:"message"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewPluginSDK 创建插件SDK实例
func NewPluginSDK() *PluginSDK {
	// 从环境变量读取配置
	context := PluginContext{
		PluginID:    getEnvDefault("PLUGIN_ID", "unknown"),
		Version:     getEnvDefault("PLUGIN_VERSION", "1.0.0"),
		Environment: getEnvDefault("PLUGIN_ENV", "development"),
		WorkDir:     getEnvDefault("PLUGIN_WORK_DIR", "/tmp"),
		TempDir:     getEnvDefault("PLUGIN_TEMP_DIR", "/tmp"),
		DataDir:     getEnvDefault("PLUGIN_DATA_DIR", "/tmp/data"),
		LogLevel:    getEnvDefault("PLUGIN_LOG_LEVEL", "INFO"),
	}

	// 解析参数
	if paramsStr := os.Getenv("PLUGIN_PARAMS"); paramsStr != "" {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(paramsStr), &params); err == nil {
			context.Params = params
		}
	}

	// 创建配置
	config := &Config{
		Host:      getEnvDefault("STELLAR_HOST", "localhost"),
		Port:      getEnvIntDefault("STELLAR_PORT", 8090),
		APIKey:    getEnvDefault("STELLAR_API_KEY", ""),
		Timeout:   time.Duration(getEnvIntDefault("STELLAR_TIMEOUT", 30)) * time.Second,
		EnableSSL: getEnvBoolDefault("STELLAR_SSL", false),
		Headers:   make(map[string]string),
	}

	// 创建日志器
	logger := &Logger{
		level:     parseLogLevel(context.LogLevel),
		prefix:    fmt.Sprintf("[%s] ", context.PluginID),
		enableBot: getEnvBoolDefault("PLUGIN_ENABLE_BOT_LOG", false),
	}

	// 设置日志输出
	if logFile := os.Getenv("PLUGIN_LOG_FILE"); logFile != "" {
		if file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			logger.output = file
		}
	}
	if logger.output == nil {
		logger.output = os.Stderr
	}

	// 创建API客户端
	baseURL := fmt.Sprintf("http://%s:%d", config.Host, config.Port)
	if config.EnableSSL {
		baseURL = fmt.Sprintf("https://%s:%d", config.Host, config.Port)
	}

	apiClient := &APIClient{
		baseURL: baseURL,
		apiKey:  config.APIKey,
		headers: config.Headers,
		timeout: config.Timeout,
	}

	return &PluginSDK{
		context: context,
		logger:  logger,
		config:  config,
		api:     apiClient,
	}
}

// GetContext 获取插件上下文
func (sdk *PluginSDK) GetContext() PluginContext {
	return sdk.context
}

// GetParam 获取插件参数
func (sdk *PluginSDK) GetParam(key string) interface{} {
	if sdk.context.Params == nil {
		return nil
	}
	return sdk.context.Params[key]
}

// GetParamString 获取字符串参数
func (sdk *PluginSDK) GetParamString(key string) string {
	if value := sdk.GetParam(key); value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// GetParamInt 获取整数参数
func (sdk *PluginSDK) GetParamInt(key string) int {
	if value := sdk.GetParam(key); value != nil {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return 0
}

// GetParamBool 获取布尔参数
func (sdk *PluginSDK) GetParamBool(key string) bool {
	if value := sdk.GetParam(key); value != nil {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

// SetResult 设置执行结果
func (sdk *PluginSDK) SetResult(success bool, data interface{}, message string) *Result {
	result := &Result{
		Success:   success,
		Data:      data,
		Message:   message,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// 输出结果标记（用于执行引擎解析）
	fmt.Println("PLUGIN_RESULT_START")
	if jsonData, err := json.Marshal(result); err == nil {
		fmt.Println(string(jsonData))
	}
	fmt.Println("PLUGIN_RESULT_END")

	return result
}

// SetError 设置错误结果
func (sdk *PluginSDK) SetError(err error, message string) *Result {
	result := &Result{
		Success:   false,
		Data:      nil,
		Message:   message,
		Error:     err.Error(),
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// 输出结果标记
	fmt.Println("PLUGIN_RESULT_START")
	if jsonData, err := json.Marshal(result); err == nil {
		fmt.Println(string(jsonData))
	}
	fmt.Println("PLUGIN_RESULT_END")

	return result
}

// Log 记录日志
func (sdk *PluginSDK) Log(level LogLevel, message string, args ...interface{}) {
	sdk.logger.Log(level, message, args...)
}

// Debug 记录调试日志
func (sdk *PluginSDK) Debug(message string, args ...interface{}) {
	sdk.logger.Log(DEBUG, message, args...)
}

// Info 记录信息日志
func (sdk *PluginSDK) Info(message string, args ...interface{}) {
	sdk.logger.Log(INFO, message, args...)
}

// Warn 记录警告日志
func (sdk *PluginSDK) Warn(message string, args ...interface{}) {
	sdk.logger.Log(WARN, message, args...)
}

// Error 记录错误日志
func (sdk *PluginSDK) Error(message string, args ...interface{}) {
	sdk.logger.Log(ERROR, message, args...)
}

// Fatal 记录致命错误日志并退出
func (sdk *PluginSDK) Fatal(message string, args ...interface{}) {
	sdk.logger.Log(FATAL, message, args...)
	os.Exit(1)
}

// CreateTempFile 创建临时文件
func (sdk *PluginSDK) CreateTempFile(prefix string, content []byte) (string, error) {
	// 确保临时目录存在
	if err := os.MkdirAll(sdk.context.TempDir, 0755); err != nil {
		return "", fmt.Errorf("创建临时目录失败: %v", err)
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp(sdk.context.TempDir, prefix+"_*.tmp")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer tempFile.Close()

	// 写入内容
	if content != nil {
		if _, err := tempFile.Write(content); err != nil {
			return "", fmt.Errorf("写入临时文件失败: %v", err)
		}
	}

	return tempFile.Name(), nil
}

// SaveData 保存数据到数据目录
func (sdk *PluginSDK) SaveData(filename string, data []byte) error {
	// 确保数据目录存在
	if err := os.MkdirAll(sdk.context.DataDir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %v", err)
	}

	// 保存文件
	filePath := filepath.Join(sdk.context.DataDir, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("保存数据文件失败: %v", err)
	}

	sdk.Debug("数据已保存到: %s", filePath)
	return nil
}

// LoadData 从数据目录加载数据
func (sdk *PluginSDK) LoadData(filename string) ([]byte, error) {
	filePath := filepath.Join(sdk.context.DataDir, filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("加载数据文件失败: %v", err)
	}
	
	sdk.Debug("数据已从 %s 加载", filePath)
	return data, nil
}

// HTTPGet 发送GET请求
func (sdk *PluginSDK) HTTPGet(url string, headers map[string]string) ([]byte, error) {
	return sdk.api.Get(url, headers)
}

// HTTPPost 发送POST请求
func (sdk *PluginSDK) HTTPPost(url string, data []byte, headers map[string]string) ([]byte, error) {
	return sdk.api.Post(url, data, headers)
}

// GetAPI 获取API客户端
func (sdk *PluginSDK) GetAPI() *APIClient {
	return sdk.api
}

// IsProductionMode 检查是否为生产模式
func (sdk *PluginSDK) IsProductionMode() bool {
	return sdk.context.Environment == "production"
}

// IsDevelopmentMode 检查是否为开发模式
func (sdk *PluginSDK) IsDevelopmentMode() bool {
	return sdk.context.Environment == "development"
}

// ValidateParam 验证参数
func (sdk *PluginSDK) ValidateParam(key string, required bool) error {
	value := sdk.GetParam(key)
	if required && value == nil {
		return fmt.Errorf("必需参数 %s 不能为空", key)
	}
	return nil
}

// ValidateParams 批量验证参数
func (sdk *PluginSDK) ValidateParams(requiredParams []string) error {
	for _, param := range requiredParams {
		if err := sdk.ValidateParam(param, true); err != nil {
			return err
		}
	}
	return nil
}

// Log 方法实现
func (l *Logger) Log(level LogLevel, message string, args ...interface{}) {
	if level < l.level {
		return
	}

	levelStr := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[level]
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	
	formattedMessage := fmt.Sprintf(message, args...)
	logLine := fmt.Sprintf("%s %s%s %s\n", timestamp, l.prefix, levelStr, formattedMessage)
	
	l.output.WriteString(logLine)

	// 如果启用了机器人日志，也输出到stderr（用于日志收集）
	if l.enableBot && l.output != os.Stderr {
		fmt.Fprintf(os.Stderr, "[LOG] %s", formattedMessage)
	}
}

// 工具函数
func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBoolDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

func parseLogLevel(level string) LogLevel {
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}