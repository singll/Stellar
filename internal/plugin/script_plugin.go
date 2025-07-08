package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ScriptPlugin 脚本插件实现
type ScriptPlugin struct {
	path        string                 // 脚本路径
	interpreter string                 // 解释器
	info        PluginInfo             // 插件信息
	config      map[string]interface{} // 配置
}

// NewScriptPlugin 创建脚本插件
func NewScriptPlugin(path string, interpreter string) (*ScriptPlugin, error) {
	plugin := &ScriptPlugin{
		path:        path,
		interpreter: interpreter,
		config:      make(map[string]interface{}),
	}

	// 加载插件信息
	if err := plugin.loadInfo(); err != nil {
		return nil, err
	}

	return plugin, nil
}

// loadInfo 加载插件信息
func (p *ScriptPlugin) loadInfo() error {
	// 调用脚本的info命令获取插件信息
	cmd := exec.Command(p.interpreter, p.path, "info")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("获取插件信息失败: %v", err)
	}

	// 解析JSON输出
	if err := json.Unmarshal(output, &p.info); err != nil {
		return fmt.Errorf("解析插件信息失败: %v", err)
	}

	return nil
}

// Info 返回插件信息
func (p *ScriptPlugin) Info() PluginInfo {
	return p.info
}

// Init 初始化插件
func (p *ScriptPlugin) Init(config map[string]interface{}) error {
	p.config = config

	// 将配置转换为JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 调用脚本的init命令初始化插件
	cmd := exec.Command(p.interpreter, p.path, "init")
	cmd.Stdin = strings.NewReader(string(configJSON))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("初始化插件失败: %v", err)
	}

	return nil
}

// Execute 执行插件
func (p *ScriptPlugin) Execute(ctx PluginContext) (PluginResult, error) {
	var result PluginResult

	// 将参数转换为JSON
	paramsJSON, err := json.Marshal(ctx.Params)
	if err != nil {
		return result, fmt.Errorf("序列化参数失败: %v", err)
	}

	// 创建临时文件存储环境变量
	envFile, err := os.CreateTemp("", "plugin_env_*.json")
	if err != nil {
		return result, fmt.Errorf("创建环境变量文件失败: %v", err)
	}
	defer os.Remove(envFile.Name())

	// 写入环境变量
	envJSON, err := json.Marshal(ctx.Environment)
	if err != nil {
		return result, fmt.Errorf("序列化环境变量失败: %v", err)
	}
	if _, err := envFile.Write(envJSON); err != nil {
		return result, fmt.Errorf("写入环境变量失败: %v", err)
	}
	envFile.Close()

	// 准备命令
	cmd := exec.CommandContext(ctx.Context, p.interpreter, p.path, "execute")
	cmd.Stdin = strings.NewReader(string(paramsJSON))
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PLUGIN_ENV_FILE=%s", envFile.Name()),
		fmt.Sprintf("PLUGIN_WORK_DIR=%s", ctx.WorkDir),
		fmt.Sprintf("PLUGIN_TIMEOUT=%d", int(ctx.Timeout.Seconds())),
	)

	// 设置工作目录
	if ctx.WorkDir != "" {
		cmd.Dir = ctx.WorkDir
	} else {
		cmd.Dir = filepath.Dir(p.path)
	}

	// 执行命令
	startTime := time.Now()
	output, err := cmd.Output()
	executionTime := time.Since(startTime)

	// 处理执行结果
	if err != nil {
		// 尝试解析错误输出
		var errResult PluginResult
		if err := json.Unmarshal(output, &errResult); err == nil {
			errResult.Success = false
			errResult.ExecutionTime = executionTime
			return errResult, nil
		}

		// 如果无法解析，创建一个默认的错误结果
		return PluginResult{
			Success:       false,
			Error:         fmt.Sprintf("执行插件失败: %v", err),
			ExecutionTime: executionTime,
		}, nil
	}

	// 解析执行结果
	if err := json.Unmarshal(output, &result); err != nil {
		return PluginResult{
			Success:       false,
			Error:         fmt.Sprintf("解析执行结果失败: %v", err),
			ExecutionTime: executionTime,
		}, nil
	}

	// 确保执行时间被设置
	result.ExecutionTime = executionTime
	return result, nil
}

// Validate 验证参数
func (p *ScriptPlugin) Validate(params map[string]interface{}) error {
	// 将参数转换为JSON
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("序列化参数失败: %v", err)
	}

	// 调用脚本的validate命令验证参数
	cmd := exec.Command(p.interpreter, p.path, "validate")
	cmd.Stdin = strings.NewReader(string(paramsJSON))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("验证参数失败: %v", err)
	}

	return nil
}

// Cleanup 清理资源
func (p *ScriptPlugin) Cleanup() error {
	// 调用脚本的cleanup命令清理资源
	cmd := exec.Command(p.interpreter, p.path, "cleanup")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("清理资源失败: %v", err)
	}

	return nil
}
