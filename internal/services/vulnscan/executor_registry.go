package vulnscan

import (
	"fmt"
	"sync"

	"github.com/StellarServer/internal/models"
)

// POCExecutorRegistry POC执行器注册表
type POCExecutorRegistry struct {
	executors map[string]POCExecutor
	mutex     sync.RWMutex
}

// ExecutorInfo 执行器信息
type ExecutorInfo struct {
	Name           string   `json:"name"`
	SupportedTypes []string `json:"supported_types"`
	Description    string   `json:"description"`
	Version        string   `json:"version"`
	Author         string   `json:"author"`
	Status         string   `json:"status"` // active, inactive, error
}

// NewPOCExecutorRegistry 创建执行器注册表
func NewPOCExecutorRegistry() *POCExecutorRegistry {
	registry := &POCExecutorRegistry{
		executors: make(map[string]POCExecutor),
	}
	
	// 注册默认执行器
	registry.registerDefaultExecutors()
	
	return registry
}

// registerDefaultExecutors 注册默认执行器
func (r *POCExecutorRegistry) registerDefaultExecutors() {
	// 注册Python执行器
	pythonExecutor := NewPythonPOCExecutor()
	r.Register(pythonExecutor)
	
	// 注册Go执行器
	goExecutor := NewGoPOCExecutor()
	r.Register(goExecutor)
	
	// 注册Nuclei执行器
	nucleiExecutor := NewNucleiPOCExecutor()
	r.Register(nucleiExecutor)
	
	// 注册JavaScript执行器
	jsExecutor := NewJavaScriptPOCExecutor()
	r.Register(jsExecutor)
}

// Register 注册执行器
func (r *POCExecutorRegistry) Register(executor POCExecutor) error {
	if executor == nil {
		return fmt.Errorf("执行器不能为空")
	}
	
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	name := executor.GetName()
	if name == "" {
		return fmt.Errorf("执行器名称不能为空")
	}
	
	// 检查是否已存在
	if _, exists := r.executors[name]; exists {
		return fmt.Errorf("执行器 %s 已存在", name)
	}
	
	r.executors[name] = executor
	return nil
}

// Unregister 注销执行器
func (r *POCExecutorRegistry) Unregister(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.executors[name]; !exists {
		return fmt.Errorf("执行器 %s 不存在", name)
	}
	
	delete(r.executors, name)
	return nil
}

// Get 获取执行器
func (r *POCExecutorRegistry) Get(name string) (POCExecutor, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	executor, exists := r.executors[name]
	if !exists {
		return nil, fmt.Errorf("执行器 %s 不存在", name)
	}
	
	return executor, nil
}

// GetByType 根据脚本类型获取执行器
func (r *POCExecutorRegistry) GetByType(scriptType string) (POCExecutor, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	for _, executor := range r.executors {
		for _, supportedType := range executor.GetSupportedTypes() {
			if supportedType == scriptType {
				return executor, nil
			}
		}
	}
	
	return nil, fmt.Errorf("不支持的脚本类型: %s", scriptType)
}

// List 列出所有执行器
func (r *POCExecutorRegistry) List() []ExecutorInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var executors []ExecutorInfo
	for name, executor := range r.executors {
		info := ExecutorInfo{
			Name:           name,
			SupportedTypes: executor.GetSupportedTypes(),
			Status:         "active",
		}
		
		// 根据执行器类型设置描述信息
		switch name {
		case "python":
			info.Description = "Python脚本执行器，支持requests库进行HTTP请求"
			info.Version = "1.0.0"
			info.Author = "Stellar Team"
		case "go":
			info.Description = "Go程序执行器，支持编译和执行Go代码"
			info.Version = "1.0.0"
			info.Author = "Stellar Team"
		case "nuclei":
			info.Description = "Nuclei模板执行器，支持YAML格式的漏洞检测模板"
			info.Version = "1.0.0"
			info.Author = "Stellar Team"
		case "javascript":
			info.Description = "JavaScript脚本执行器，基于Node.js运行环境"
			info.Version = "1.0.0"
			info.Author = "Stellar Team"
		default:
			info.Description = "自定义POC执行器"
			info.Version = "unknown"
			info.Author = "unknown"
		}
		
		executors = append(executors, info)
	}
	
	return executors
}

// GetSupportedTypes 获取所有支持的脚本类型
func (r *POCExecutorRegistry) GetSupportedTypes() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var types []string
	typeSet := make(map[string]bool)
	
	for _, executor := range r.executors {
		for _, scriptType := range executor.GetSupportedTypes() {
			if !typeSet[scriptType] {
				types = append(types, scriptType)
				typeSet[scriptType] = true
			}
		}
	}
	
	return types
}

// ValidatePOC 验证POC脚本
func (r *POCExecutorRegistry) ValidatePOC(poc *models.POC) error {
	executor, err := r.GetByType(poc.ScriptType)
	if err != nil {
		return err
	}
	
	return executor.Validate(poc)
}

// GetExecutorStats 获取执行器统计信息
func (r *POCExecutorRegistry) GetExecutorStats() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	stats := map[string]interface{}{
		"total_executors":   len(r.executors),
		"supported_types":   len(r.GetSupportedTypes()),
		"executor_details":  r.List(),
	}
	
	return stats
}

// CheckExecutorHealth 检查执行器健康状态
func (r *POCExecutorRegistry) CheckExecutorHealth() map[string]string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	health := make(map[string]string)
	
	for name, executor := range r.executors {
		// 尝试验证一个空的POC来检查执行器是否正常
		testPOC := &models.POC{
			Name:       "health_check",
			Script:     "console.log('test');",
			ScriptType: executor.GetSupportedTypes()[0],
		}
		
		err := executor.Validate(testPOC)
		if err != nil {
			health[name] = "error: " + err.Error()
		} else {
			health[name] = "healthy"
		}
	}
	
	return health
}

// 全局执行器注册表实例
var GlobalExecutorRegistry *POCExecutorRegistry

// init 初始化全局注册表
func init() {
	GlobalExecutorRegistry = NewPOCExecutorRegistry()
}

// RegisterGlobalExecutor 注册全局执行器
func RegisterGlobalExecutor(executor POCExecutor) error {
	return GlobalExecutorRegistry.Register(executor)
}

// GetGlobalExecutor 获取全局执行器
func GetGlobalExecutor(name string) (POCExecutor, error) {
	return GlobalExecutorRegistry.Get(name)
}

// GetGlobalExecutorByType 根据类型获取全局执行器
func GetGlobalExecutorByType(scriptType string) (POCExecutor, error) {
	return GlobalExecutorRegistry.GetByType(scriptType)
}

// ListGlobalExecutors 列出所有全局执行器
func ListGlobalExecutors() []ExecutorInfo {
	return GlobalExecutorRegistry.List()
}

// GetGlobalSupportedTypes 获取全局支持的脚本类型
func GetGlobalSupportedTypes() []string {
	return GlobalExecutorRegistry.GetSupportedTypes()
}