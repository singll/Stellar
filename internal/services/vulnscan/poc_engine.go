package vulnscan

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"golang.org/x/time/rate"
)

// POCEngine POC执行引擎
type POCEngine struct {
	// POC执行器映射
	executors map[string]POCExecutor
	// 执行沙盒
	sandbox *ExecutionSandbox
	// 结果缓存
	cache *POCResultCache
	// 速率限制器
	rateLimiter *rate.Limiter
	// 配置
	config POCEngineConfig
	// 统计信息
	stats POCEngineStats
	// 互斥锁
	mutex sync.RWMutex
}

// POCEngineConfig POC引擎配置
type POCEngineConfig struct {
	MaxConcurrency  int           `json:"max_concurrency"`  // 最大并发数
	Timeout         time.Duration `json:"timeout"`          // 执行超时时间
	RateLimit       float64       `json:"rate_limit"`       // 速率限制(每秒)
	EnableCache     bool          `json:"enable_cache"`     // 是否启用缓存
	CacheTTL        time.Duration `json:"cache_ttl"`        // 缓存过期时间
	EnableSandbox   bool          `json:"enable_sandbox"`   // 是否启用沙盒
	MaxMemoryMB     int           `json:"max_memory_mb"`    // 最大内存使用(MB)
	MaxScriptSize   int           `json:"max_script_size"`  // 最大脚本大小(字节)
}

// POCEngineStats POC引擎统计信息
type POCEngineStats struct {
	TotalExecutions  int64         `json:"total_executions"`
	SuccessfulExecs  int64         `json:"successful_execs"`
	FailedExecs      int64         `json:"failed_execs"`
	CacheHits        int64         `json:"cache_hits"`
	CacheMisses      int64         `json:"cache_misses"`
	AvgExecutionTime time.Duration `json:"avg_execution_time"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// POCExecutor POC执行器接口
type POCExecutor interface {
	// Execute 执行POC
	Execute(ctx context.Context, poc *models.POC, target POCTarget) (*models.POCResult, error)
	// GetSupportedTypes 获取支持的脚本类型
	GetSupportedTypes() []string
	// Validate 验证POC脚本
	Validate(poc *models.POC) error
	// GetName 获取执行器名称
	GetName() string
}

// POCTarget POC扫描目标 (重命名避免与plugin.go中的Target冲突)
type POCTarget struct {
	URL    string            `json:"url"`
	Host   string            `json:"host"`
	Port   int               `json:"port"`
	Scheme string            `json:"scheme"`
	Path   string            `json:"path"`
	Query  string            `json:"query"`
	Extra  map[string]string `json:"extra"`
}

// Hash 计算目标哈希值
func (t POCTarget) Hash() string {
	data := fmt.Sprintf("%s:%s:%d:%s:%s:%s", 
		t.URL, t.Host, t.Port, t.Scheme, t.Path, t.Query)
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

// String 返回目标字符串表示
func (t POCTarget) String() string {
	if t.URL != "" {
		return t.URL
	}
	if t.Host != "" {
		if t.Port > 0 && t.Port != 80 && t.Port != 443 {
			return fmt.Sprintf("%s:%d", t.Host, t.Port)
		}
		return t.Host
	}
	return ""
}

// POCResultCache POC结果缓存
type POCResultCache struct {
	cache   map[string]*CacheEntry
	mutex   sync.RWMutex
	ttl     time.Duration
	maxSize int
	stats   CacheStats
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Result    *models.POCResult
	ExpiresAt time.Time
	CreatedAt time.Time
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Hits     int64 `json:"hits"`
	Misses   int64 `json:"misses"`
	Size     int   `json:"size"`
	MaxSize  int   `json:"max_size"`
	HitRatio float64 `json:"hit_ratio"`
}

// ExecutionSandbox 执行沙盒
type ExecutionSandbox struct {
	config SandboxConfig
	limits ResourceLimits
	policy SecurityPolicy
}

// SandboxConfig 沙盒配置
type SandboxConfig struct {
	EnableContainerization bool          `json:"enable_containerization"`
	MaxExecutionTime       time.Duration `json:"max_execution_time"`
	EnableNetworkAccess    bool          `json:"enable_network_access"`
	AllowedHosts           []string      `json:"allowed_hosts"`
	EnableFileAccess       bool          `json:"enable_file_access"`
	AllowedPaths           []string      `json:"allowed_paths"`
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	MaxMemoryMB  int           `json:"max_memory_mb"`
	MaxCPUPercent float64      `json:"max_cpu_percent"`
	MaxDuration  time.Duration `json:"max_duration"`
	MaxFileSize  int64         `json:"max_file_size"`
	MaxOpenFiles int           `json:"max_open_files"`
}

// SecurityPolicy 安全策略
type SecurityPolicy struct {
	BlockedFunctions []string `json:"blocked_functions"`
	BlockedModules   []string `json:"blocked_modules"`
	BlockedPatterns  []string `json:"blocked_patterns"`
	RequireAuth      bool     `json:"require_auth"`
}

// NewPOCEngine 创建POC引擎
func NewPOCEngine(config POCEngineConfig) *POCEngine {
	engine := &POCEngine{
		executors: make(map[string]POCExecutor),
		config:    config,
		stats:     POCEngineStats{LastUpdated: time.Now()},
	}

	// 创建速率限制器
	if config.RateLimit > 0 {
		engine.rateLimiter = rate.NewLimiter(rate.Limit(config.RateLimit), int(config.RateLimit))
	}

	// 创建缓存
	if config.EnableCache {
		engine.cache = NewPOCResultCache(config.CacheTTL, 10000) // 默认最大10000个缓存项
	}

	// 创建沙盒
	if config.EnableSandbox {
		sandboxConfig := SandboxConfig{
			EnableContainerization: true,
			MaxExecutionTime:       config.Timeout,
			EnableNetworkAccess:    true,
			EnableFileAccess:       false,
		}
		engine.sandbox = NewExecutionSandbox(sandboxConfig)
	}

	// 注册默认执行器
	engine.registerDefaultExecutors()

	return engine
}

// registerDefaultExecutors 注册默认执行器
func (e *POCEngine) registerDefaultExecutors() {
	// 使用全局注册表中的执行器
	for _, executorInfo := range ListGlobalExecutors() {
		if executor, err := GetGlobalExecutor(executorInfo.Name); err == nil {
			e.RegisterExecutor(executor)
		}
	}
}

// RegisterExecutor 注册POC执行器
func (e *POCEngine) RegisterExecutor(executor POCExecutor) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for _, scriptType := range executor.GetSupportedTypes() {
		e.executors[scriptType] = executor
	}
}

// UnregisterExecutor 注销POC执行器
func (e *POCEngine) UnregisterExecutor(scriptType string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	delete(e.executors, scriptType)
}

// GetExecutor 获取POC执行器
func (e *POCEngine) GetExecutor(scriptType string) (POCExecutor, bool) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	executor, exists := e.executors[scriptType]
	return executor, exists
}

// ListExecutors 列出所有执行器
func (e *POCEngine) ListExecutors() map[string]POCExecutor {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	result := make(map[string]POCExecutor)
	for k, v := range e.executors {
		result[k] = v
	}
	return result
}

// ExecutePOC 执行POC
func (e *POCEngine) ExecutePOC(ctx context.Context, poc *models.POC, target POCTarget) (*models.POCResult, error) {
	startTime := time.Now()
	
	// 更新统计信息
	defer func() {
		e.updateStats(time.Since(startTime))
	}()

	// 验证POC
	if err := e.validatePOC(poc); err != nil {
		e.stats.FailedExecs++
		return nil, fmt.Errorf("POC验证失败: %v", err)
	}

	// 检查缓存
	if e.cache != nil {
		if cached := e.cache.Get(poc.ID.Hex(), target.Hash()); cached != nil {
			e.stats.SuccessfulExecs++
			return cached, nil
		}
	}

	// 速率限制
	if e.rateLimiter != nil {
		if err := e.rateLimiter.Wait(ctx); err != nil {
			e.stats.FailedExecs++
			return nil, fmt.Errorf("速率限制: %v", err)
		}
	}

	// 获取执行器
	executor, exists := e.GetExecutor(poc.ScriptType)
	if !exists {
		e.stats.FailedExecs++
		return nil, fmt.Errorf("不支持的POC类型: %s", poc.ScriptType)
	}

	// 执行POC
	var result *models.POCResult
	var err error

	if e.sandbox != nil {
		// 沙盒执行
		result, err = e.sandbox.Execute(ctx, func() (*models.POCResult, error) {
			return executor.Execute(ctx, poc, target)
		})
	} else {
		// 直接执行
		result, err = executor.Execute(ctx, poc, target)
	}

	if err != nil {
		e.stats.FailedExecs++
		return nil, err
	}

	// 缓存结果
	if e.cache != nil && result != nil {
		e.cache.Set(poc.ID.Hex(), target.Hash(), result)
	}

	e.stats.SuccessfulExecs++
	return result, nil
}

// ValidatePOC 验证POC脚本
func (e *POCEngine) ValidatePOC(poc *models.POC) error {
	// 基本验证
	if err := e.validatePOC(poc); err != nil {
		return err
	}

	// 获取执行器进行详细验证
	executor, exists := e.GetExecutor(poc.ScriptType)
	if !exists {
		return fmt.Errorf("不支持的POC类型: %s", poc.ScriptType)
	}

	return executor.Validate(poc)
}

// validatePOC 内部POC验证
func (e *POCEngine) validatePOC(poc *models.POC) error {
	if poc == nil {
		return fmt.Errorf("POC为空")
	}

	if poc.Script == "" {
		return fmt.Errorf("POC脚本为空")
	}

	if poc.ScriptType == "" {
		return fmt.Errorf("POC脚本类型为空")
	}

	// 检查脚本大小
	if e.config.MaxScriptSize > 0 && len(poc.Script) > e.config.MaxScriptSize {
		return fmt.Errorf("POC脚本过大: %d > %d", len(poc.Script), e.config.MaxScriptSize)
	}

	// 安全检查
	if err := e.securityCheck(poc); err != nil {
		return fmt.Errorf("安全检查失败: %v", err)
	}

	return nil
}

// securityCheck 安全检查
func (e *POCEngine) securityCheck(poc *models.POC) error {
	// 检查危险函数调用
	dangerousFunctions := []string{
		"eval", "exec", "system", "shell_exec", "passthru",
		"file_get_contents", "file_put_contents", "fwrite",
		"chmod", "chown", "unlink", "rmdir", "mkdir",
		"os.system", "subprocess", "commands",
	}

	for _, dangerous := range dangerousFunctions {
		if contains(poc.Script, dangerous) {
			return fmt.Errorf("包含危险函数: %s", dangerous)
		}
	}

	// 检查敏感路径访问
	sensitivePaths := []string{
		"/etc/passwd", "/etc/shadow", "/etc/hosts",
		"C:\\Windows\\System32", "C:\\Users",
		"../", "./", "~",
	}

	for _, path := range sensitivePaths {
		if contains(poc.Script, path) {
			return fmt.Errorf("包含敏感路径访问: %s", path)
		}
	}

	return nil
}

// GetStats 获取统计信息
func (e *POCEngine) GetStats() POCEngineStats {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	stats := e.stats
	if e.cache != nil {
		cacheStats := e.cache.GetStats()
		stats.CacheHits = cacheStats.Hits
		stats.CacheMisses = cacheStats.Misses
	}

	return stats
}

// updateStats 更新统计信息
func (e *POCEngine) updateStats(executionTime time.Duration) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.stats.TotalExecutions++
	
	// 更新平均执行时间
	if e.stats.TotalExecutions == 1 {
		e.stats.AvgExecutionTime = executionTime
	} else {
		// 计算移动平均值
		e.stats.AvgExecutionTime = time.Duration(
			(int64(e.stats.AvgExecutionTime)*int64(e.stats.TotalExecutions-1) + int64(executionTime)) / int64(e.stats.TotalExecutions),
		)
	}
	
	e.stats.LastUpdated = time.Now()
}

// Shutdown 关闭POC引擎
func (e *POCEngine) Shutdown(ctx context.Context) error {
	// 清理缓存
	if e.cache != nil {
		e.cache.Clear()
	}

	// 关闭沙盒
	if e.sandbox != nil {
		return e.sandbox.Shutdown(ctx)
	}

	return nil
}

// NewPOCResultCache 创建POC结果缓存
func NewPOCResultCache(ttl time.Duration, maxSize int) *POCResultCache {
	return &POCResultCache{
		cache:   make(map[string]*CacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
		stats:   CacheStats{MaxSize: maxSize},
	}
}

// Get 获取缓存项
func (c *POCResultCache) Get(pocID, targetHash string) *models.POCResult {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	key := pocID + ":" + targetHash
	entry, exists := c.cache[key]
	if !exists {
		c.stats.Misses++
		return nil
	}

	// 检查是否过期
	if time.Now().After(entry.ExpiresAt) {
		delete(c.cache, key)
		c.stats.Misses++
		return nil
	}

	c.stats.Hits++
	c.updateHitRatio()
	return entry.Result
}

// Set 设置缓存项
func (c *POCResultCache) Set(pocID, targetHash string, result *models.POCResult) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := pocID + ":" + targetHash

	// 检查缓存大小限制
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	c.cache[key] = &CacheEntry{
		Result:    result,
		ExpiresAt: time.Now().Add(c.ttl),
		CreatedAt: time.Now(),
	}

	c.stats.Size = len(c.cache)
}

// evictOldest 淘汰最旧的缓存项
func (c *POCResultCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.cache {
		if oldestKey == "" || entry.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
	}
}

// Clear 清理缓存
func (c *POCResultCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]*CacheEntry)
	c.stats.Size = 0
}

// GetStats 获取缓存统计信息
func (c *POCResultCache) GetStats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.stats
}

// updateHitRatio 更新命中率
func (c *POCResultCache) updateHitRatio() {
	total := c.stats.Hits + c.stats.Misses
	if total > 0 {
		c.stats.HitRatio = float64(c.stats.Hits) / float64(total)
	}
}

// NewExecutionSandbox 创建执行沙盒
func NewExecutionSandbox(config SandboxConfig) *ExecutionSandbox {
	return &ExecutionSandbox{
		config: config,
		limits: ResourceLimits{
			MaxMemoryMB:   512,
			MaxCPUPercent: 50,
			MaxDuration:   config.MaxExecutionTime,
			MaxFileSize:   1024 * 1024, // 1MB
			MaxOpenFiles:  10,
		},
		policy: SecurityPolicy{
			BlockedFunctions: []string{"eval", "exec", "system"},
			BlockedModules:   []string{"os", "subprocess"},
			RequireAuth:      false,
		},
	}
}

// Execute 在沙盒中执行函数
func (s *ExecutionSandbox) Execute(ctx context.Context, fn func() (*models.POCResult, error)) (*models.POCResult, error) {
	// 创建带超时的上下文
	execCtx, cancel := context.WithTimeout(ctx, s.limits.MaxDuration)
	defer cancel()

	// 在goroutine中执行
	resultChan := make(chan *models.POCResult, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errorChan <- fmt.Errorf("POC执行panic: %v", r)
			}
		}()

		result, err := fn()
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- result
		}
	}()

	// 等待结果或超时
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return nil, err
	case <-execCtx.Done():
		return nil, fmt.Errorf("POC执行超时")
	}
}

// Shutdown 关闭沙盒
func (s *ExecutionSandbox) Shutdown(ctx context.Context) error {
	// 清理资源
	return nil
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr || 
			 findSubstring(s, substr))))
}

// findSubstring 查找子串
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}