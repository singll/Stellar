package main

import (
	"fmt"
	"time"

	"github.com/StellarServer/internal/plugin"
)

// HelloPlugin 示例插件
type HelloPlugin struct {
	info   plugin.PluginInfo
	config map[string]interface{}
}

// Plugin 导出的插件实例
var Plugin = &HelloPlugin{
	info: plugin.PluginInfo{
		ID:          "hello_plugin",
		Name:        "Hello Plugin",
		Version:     "1.0.0",
		Type:        plugin.TypeUtility,
		Author:      "StellarServer Team",
		Description: "A simple example plugin",
		Category:    "example",
		Tags:        []string{"example", "hello", "utility"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Params: []plugin.PluginParam{
			{
				Name:        "name",
				Type:        "string",
				Description: "Your name",
				Required:    true,
				Default:     "World",
			},
			{
				Name:        "greeting",
				Type:        "string",
				Description: "Greeting message",
				Required:    false,
				Default:     "Hello",
				Options:     []string{"Hello", "Hi", "Hey", "Greetings"},
			},
		},
		Language: "Go",
	},
	config: make(map[string]interface{}),
}

// Info 返回插件信息
func (p *HelloPlugin) Info() plugin.PluginInfo {
	return p.info
}

// Init 初始化插件
func (p *HelloPlugin) Init(config map[string]interface{}) error {
	p.config = config
	return nil
}

// Execute 执行插件
func (p *HelloPlugin) Execute(ctx plugin.PluginContext) (plugin.PluginResult, error) {
	// 获取参数
	name := "World"
	if nameParam, ok := ctx.Params["name"]; ok && nameParam != nil {
		if nameStr, ok := nameParam.(string); ok {
			name = nameStr
		}
	}

	greeting := "Hello"
	if greetingParam, ok := ctx.Params["greeting"]; ok && greetingParam != nil {
		if greetingStr, ok := greetingParam.(string); ok {
			greeting = greetingStr
		}
	}

	// 构建消息
	message := greeting + ", " + name + "!"

	// 记录日志
	if ctx.Logger != nil {
		ctx.Logger.Info("Executing HelloPlugin with name: %s, greeting: %s", name, greeting)
	}

	// 返回结果
	return plugin.PluginResult{
		Success: true,
		Data:    message,
		Message: "Plugin executed successfully",
		Metadata: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}, nil
}

// Validate 验证参数
func (p *HelloPlugin) Validate(params map[string]interface{}) error {
	// 检查必需参数
	if _, ok := params["name"]; !ok {
		return fmt.Errorf("missing required parameter: name")
	}
	return nil
}

// Cleanup 清理资源
func (p *HelloPlugin) Cleanup() error {
	// 无需清理资源
	return nil
}

// 插件入口点
func main() {
	// 这个函数在使用Go插件时不会被调用
	// 但对于构建独立可执行文件很有用
}
