package plugin

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	goplugin "plugin"
	"strings"
	"sync"
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

// Load 加载YAML插件
func (l *YAMLLoader) Load(path string) (Plugin, error) {
	// 读取YAML文件并解析
	// 这里只是一个占位实现，实际实现需要根据YAML格式定义
	return nil, fmt.Errorf("YAML插件加载器尚未实现")
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
