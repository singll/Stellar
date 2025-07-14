package plugin

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	goplugin "plugin"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Loader 插件加载器接口
type Loader interface {
	// Load 加载插件
	Load(path string) (Plugin, error)

	// SupportedExtensions 返回支持的文件扩展名
	SupportedExtensions() []string
}

// GoLoader Go插件加载器
type GoLoader struct{}

// Load 加载Go插件
func (l *GoLoader) Load(path string) (Plugin, error) {
	// 打开Go插件
	p, err := goplugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开Go插件失败: %v", err)
	}

	// 查找插件符号
	sym, err := p.Lookup("Plugin")
	if err != nil {
		return nil, fmt.Errorf("查找插件符号失败: %v", err)
	}

	// 类型断言
	plugin, ok := sym.(Plugin)
	if !ok {
		return nil, fmt.Errorf("插件符号类型错误: %T", sym)
	}

	return plugin, nil
}

// SupportedExtensions 返回支持的文件扩展名
func (l *GoLoader) SupportedExtensions() []string {
	return []string{".so"}
}

// ScriptLoader 脚本插件加载器
type ScriptLoader struct {
	interpreters map[string]string
}

// NewScriptLoader 创建脚本插件加载器
func NewScriptLoader() *ScriptLoader {
	return &ScriptLoader{
		interpreters: map[string]string{
			".py":  "python",
			".js":  "node",
			".rb":  "ruby",
			".sh":  "bash",
			".lua": "lua",
		},
	}
}

// Load 加载脚本插件
func (l *ScriptLoader) Load(path string) (Plugin, error) {
	ext := filepath.Ext(path)
	interpreter, ok := l.interpreters[ext]
	if !ok {
		return nil, fmt.Errorf("不支持的脚本类型: %s", ext)
	}

	// 创建脚本插件包装器
	return NewScriptPlugin(path, interpreter)
}

// SupportedExtensions 返回支持的文件扩展名
func (l *ScriptLoader) SupportedExtensions() []string {
	exts := make([]string, 0, len(l.interpreters))
	for ext := range l.interpreters {
		exts = append(exts, ext)
	}
	return exts
}

// YAMLLoader YAML插件加载器
type YAMLLoader struct{}

// YAMLPluginConfig YAML插件配置
type YAMLPluginConfig struct {
	ID          string                 `yaml:"id"`          // 插件ID
	Name        string                 `yaml:"name"`        // 插件名称
	Version     string                 `yaml:"version"`     // 版本
	Author      string                 `yaml:"author"`      // 作者
	Description string                 `yaml:"description"` // 描述
	Type        string                 `yaml:"type"`        // 插件类型
	Category    string                 `yaml:"category"`    // 分类
	Tags        []string               `yaml:"tags"`        // 标签
	Dependencies []string              `yaml:"dependencies"` // 依赖
	Config      map[string]interface{} `yaml:"config"`      // 配置
	Script      ScriptConfig           `yaml:"script"`      // 脚本配置
}

// ScriptConfig 脚本配置
type ScriptConfig struct {
	Language string   `yaml:"language"` // 脚本语言：python, javascript, shell
	Content  string   `yaml:"content"`  // 脚本内容
	Entry    string   `yaml:"entry"`    // 入口函数名
	Args     []string `yaml:"args"`     // 参数
}

// validateConfig 验证YAML配置
func (l *YAMLLoader) validateConfig(config *YAMLPluginConfig) error {
	if config.ID == "" {
		return fmt.Errorf("插件ID不能为空")
	}
	if config.Name == "" {
		return fmt.Errorf("插件名称不能为空")
	}
	if config.Version == "" {
		return fmt.Errorf("插件版本不能为空")
	}
	if config.Type == "" {
		return fmt.Errorf("插件类型不能为空")
	}
	if config.Script.Language == "" {
		return fmt.Errorf("脚本语言不能为空")
	}
	if config.Script.Content == "" {
		return fmt.Errorf("脚本内容不能为空")
	}

	// 验证支持的脚本语言
	supportedLanguages := []string{"python", "javascript", "shell", "lua"}
	langSupported := false
	for _, lang := range supportedLanguages {
		if config.Script.Language == lang {
			langSupported = true
			break
		}
	}
	if !langSupported {
		return fmt.Errorf("不支持的脚本语言: %s", config.Script.Language)
	}

	return nil
}

// Load 加载YAML插件
func (l *YAMLLoader) Load(path string) (Plugin, error) {
	// 读取YAML文件
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取YAML插件文件失败: %v", err)
	}

	// 解析YAML内容
	var yamlConfig YAMLPluginConfig
	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		return nil, fmt.Errorf("解析YAML插件文件失败: %v", err)
	}

	// 验证YAML配置
	if err := l.validateConfig(&yamlConfig); err != nil {
		return nil, fmt.Errorf("YAML插件配置验证失败: %v", err)
	}

	// 创建YAML插件实例
	return NewYAMLPlugin(&yamlConfig, path)
}

// NewYAMLPlugin 创建YAML插件实例
func NewYAMLPlugin(config *YAMLPluginConfig, path string) (Plugin, error) {
	// 创建YAML插件实例
	yamlPlugin := &YAMLPlugin{
		config: config,
		path:   path,
	}

	// 验证脚本执行环境
	if err := yamlPlugin.validateEnvironment(); err != nil {
		return nil, fmt.Errorf("YAML插件环境验证失败: %v", err)
	}

	return yamlPlugin, nil
}

// YAMLPlugin YAML插件实现
type YAMLPlugin struct {
	config *YAMLPluginConfig
	path   string
}

// validateEnvironment 验证执行环境
func (p *YAMLPlugin) validateEnvironment() error {
	// 根据脚本语言检查执行环境
	switch p.config.Script.Language {
	case "python":
		return p.checkPythonEnvironment()
	case "javascript":
		return p.checkNodeEnvironment()
	case "shell":
		return p.checkShellEnvironment()
	case "lua":
		return p.checkLuaEnvironment()
	default:
		return fmt.Errorf("不支持的脚本语言: %s", p.config.Script.Language)
	}
}

// checkPythonEnvironment 检查Python环境
func (p *YAMLPlugin) checkPythonEnvironment() error {
	// 这里可以检查Python是否安装，版本是否符合要求等
	// 简化实现，假设Python环境可用
	return nil
}

// checkNodeEnvironment 检查Node.js环境
func (p *YAMLPlugin) checkNodeEnvironment() error {
	// 检查Node.js环境
	return nil
}

// checkShellEnvironment 检查Shell环境
func (p *YAMLPlugin) checkShellEnvironment() error {
	// 检查Shell环境
	return nil
}

// checkLuaEnvironment 检查Lua环境
func (p *YAMLPlugin) checkLuaEnvironment() error {
	// 检查Lua环境
	return nil
}

// Info 获取插件信息
func (p *YAMLPlugin) Info() PluginInfo {
	return PluginInfo{
		ID:          p.config.ID,
		Name:        p.config.Name,
		Version:     p.config.Version,
		Author:      p.config.Author,
		Description: p.config.Description,
		Type:        PluginType(p.config.Type),
		Category:    p.config.Category,
		Tags:        p.config.Tags,
		Language:    p.config.Script.Language,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Init 初始化插件
func (p *YAMLPlugin) Init(config map[string]interface{}) error {
	p.config.Config = config
	return nil
}

// Execute 执行插件
func (p *YAMLPlugin) Execute(ctx PluginContext) (PluginResult, error) {
	startTime := time.Now()
	
	// 根据脚本语言执行相应的脚本
	var result interface{}
	var err error
	
	switch p.config.Script.Language {
	case "python":
		result, err = p.executePython(ctx.Params)
	case "javascript":
		result, err = p.executeJavaScript(ctx.Params)
	case "shell":
		result, err = p.executeShell(ctx.Params)
	case "lua":
		result, err = p.executeLua(ctx.Params)
	default:
		err = fmt.Errorf("不支持的脚本语言: %s", p.config.Script.Language)
	}
	
	executionTime := time.Since(startTime)
	
	if err != nil {
		return PluginResult{
			Success:       false,
			Data:          nil,
			Message:       "",
			Error:         err.Error(),
			ExecutionTime: executionTime,
			Metadata:      make(map[string]interface{}),
		}, err
	}
	
	return PluginResult{
		Success:       true,
		Data:          result,
		Message:       "执行成功",
		Error:         "",
		ExecutionTime: executionTime,
		Metadata:      make(map[string]interface{}),
	}, nil
}

// Validate 验证参数
func (p *YAMLPlugin) Validate(params map[string]interface{}) error {
	// TODO: 根据配置验证参数
	return nil
}

// Cleanup 清理资源
func (p *YAMLPlugin) Cleanup() error {
	// 清理插件资源
	return nil
}

// executePython 执行Python脚本
func (p *YAMLPlugin) executePython(params map[string]interface{}) (interface{}, error) {
	// TODO: 实现Python脚本执行
	// 这里可以使用exec.Command调用Python解释器
	// 或者使用Python的C API进行嵌入式执行
	return nil, fmt.Errorf("Python脚本执行尚未实现")
}

// executeJavaScript 执行JavaScript脚本
func (p *YAMLPlugin) executeJavaScript(params map[string]interface{}) (interface{}, error) {
	// TODO: 实现JavaScript脚本执行
	// 可以使用Node.js或者嵌入式JS引擎如otto
	return nil, fmt.Errorf("JavaScript脚本执行尚未实现")
}

// executeShell 执行Shell脚本
func (p *YAMLPlugin) executeShell(params map[string]interface{}) (interface{}, error) {
	// TODO: 实现Shell脚本执行
	// 使用exec.Command执行shell命令
	return nil, fmt.Errorf("Shell脚本执行尚未实现")
}

// executeLua 执行Lua脚本
func (p *YAMLPlugin) executeLua(params map[string]interface{}) (interface{}, error) {
	// TODO: 实现Lua脚本执行
	// 可以使用gopher-lua库
	return nil, fmt.Errorf("Lua脚本执行尚未实现")
}

// Config 获取插件配置
func (p *YAMLPlugin) Config() map[string]interface{} {
	return p.config.Config
}

// SetConfig 设置插件配置  
func (p *YAMLPlugin) SetConfig(config map[string]interface{}) {
	p.config.Config = config
}

// Initialize 初始化插件
func (p *YAMLPlugin) Initialize() error {
	// 执行插件初始化逻辑
	return nil
}

// SupportedExtensions 返回支持的文件扩展名
func (l *YAMLLoader) SupportedExtensions() []string {
	return []string{".yaml", ".yml"}
}

// LoadPluginsFromDirectory 从目录加载插件
func (r *RegistryImpl) LoadPluginsFromDirectory(path string) error {
	r.loaderLock.RLock()
	defer r.loaderLock.RUnlock()

	// 检查目录是否存在
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("访问插件目录失败: %v", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("指定的路径不是目录: %s", path)
	}

	// 读取目录下的所有文件
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("读取插件目录失败: %v", err)
	}

	// 并发加载插件
	var wg sync.WaitGroup
	errors := make(chan error, len(files))
	plugins := make(chan Plugin, len(files))

	for _, file := range files {
		// 跳过目录和隐藏文件
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}

		// 获取文件扩展名
		ext := filepath.Ext(file.Name())
		if ext == "" {
			continue
		}

		// 查找对应的加载器
		loader, exists := r.loaders[ext]
		if !exists {
			continue
		}

		// 并发加载插件
		wg.Add(1)
		go func(fileName string, loader Loader) {
			defer wg.Done()

			filePath := filepath.Join(path, fileName)
			plugin, err := loader.Load(filePath)
			if err != nil {
				errors <- fmt.Errorf("加载插件 %s 失败: %v", fileName, err)
				return
			}

			plugins <- plugin
		}(file.Name(), loader)
	}

	// 等待所有加载完成
	go func() {
		wg.Wait()
		close(plugins)
		close(errors)
	}()

	// 处理错误
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	// 注册插件
	for plugin := range plugins {
		if err := r.Register(plugin); err != nil {
			errs = append(errs, err)
		}
	}

	// 如果有错误，返回第一个错误
	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}
