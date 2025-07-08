package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Sandbox 插件沙箱
type Sandbox struct {
	workDir    string
	tempDir    string
	timeoutSec int
	resources  map[string]int
}

// SandboxConfig 沙箱配置
type SandboxConfig struct {
	WorkDir    string         // 工作目录
	TempDir    string         // 临时目录
	TimeoutSec int            // 超时时间（秒）
	Resources  map[string]int // 资源限制
}

// NewSandbox 创建沙箱
func NewSandbox(config SandboxConfig) (*Sandbox, error) {
	// 设置默认值
	if config.WorkDir == "" {
		config.WorkDir = os.TempDir()
	}
	if config.TempDir == "" {
		config.TempDir = os.TempDir()
	}
	if config.TimeoutSec <= 0 {
		config.TimeoutSec = 30
	}
	if config.Resources == nil {
		config.Resources = make(map[string]int)
	}

	// 创建沙箱
	sandbox := &Sandbox{
		workDir:    config.WorkDir,
		tempDir:    config.TempDir,
		timeoutSec: config.TimeoutSec,
		resources:  config.Resources,
	}

	return sandbox, nil
}

// RunCommand 在沙箱中运行命令
func (s *Sandbox) RunCommand(ctx context.Context, cmd *exec.Cmd) ([]byte, error) {
	// 设置超时上下文
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(s.timeoutSec)*time.Second)
		defer cancel()
	}

	// 设置工作目录
	if cmd.Dir == "" {
		cmd.Dir = s.workDir
	}

	// 设置环境变量
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}

	// 添加资源限制
	// 注意：这里的实现是平台相关的，可能需要根据操作系统进行调整
	if s.resources != nil {
		// 在Linux上可以使用cgroups或ulimit
		// 在Windows上可以使用Job Objects
		// 这里只是一个示例，实际实现需要根据平台进行调整
	}

	// 创建管道
	outputPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("创建输出管道失败: %v", err)
	}
	errorPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("创建错误管道失败: %v", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动命令失败: %v", err)
	}

	// 创建通道接收命令完成信号
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// 读取输出
	output := make([]byte, 0)
	buffer := make([]byte, 1024)
	for {
		n, err := outputPipe.Read(buffer)
		if n > 0 {
			output = append(output, buffer[:n]...)
		}
		if err != nil {
			break
		}
	}

	// 读取错误
	errorOutput := make([]byte, 0)
	buffer = make([]byte, 1024)
	for {
		n, err := errorPipe.Read(buffer)
		if n > 0 {
			errorOutput = append(errorOutput, buffer[:n]...)
		}
		if err != nil {
			break
		}
	}

	// 等待命令完成或超时
	select {
	case err := <-done:
		if err != nil {
			if len(errorOutput) > 0 {
				return nil, fmt.Errorf("命令执行失败: %v\n%s", err, string(errorOutput))
			}
			return nil, fmt.Errorf("命令执行失败: %v", err)
		}
		return output, nil
	case <-ctx.Done():
		// 命令超时，强制终止
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return nil, fmt.Errorf("命令执行超时")
	}
}

// CreateTempFile 在沙箱中创建临时文件
func (s *Sandbox) CreateTempFile(prefix string, content []byte) (string, error) {
	// 创建临时文件
	tempFile, err := os.CreateTemp(s.tempDir, prefix)
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer tempFile.Close()

	// 写入内容
	if _, err := tempFile.Write(content); err != nil {
		return "", fmt.Errorf("写入临时文件失败: %v", err)
	}

	return tempFile.Name(), nil
}

// CleanupTempFile 清理临时文件
func (s *Sandbox) CleanupTempFile(path string) error {
	// 检查路径是否在临时目录中
	if !strings.HasPrefix(path, s.tempDir) {
		return fmt.Errorf("路径不在临时目录中: %s", path)
	}

	// 删除文件
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("删除临时文件失败: %v", err)
	}

	return nil
}

// GetTempDir 获取临时目录
func (s *Sandbox) GetTempDir() string {
	return s.tempDir
}

// GetWorkDir 获取工作目录
func (s *Sandbox) GetWorkDir() string {
	return s.workDir
}

// SetTimeout 设置超时时间
func (s *Sandbox) SetTimeout(timeoutSec int) {
	if timeoutSec > 0 {
		s.timeoutSec = timeoutSec
	}
}

// SetResource 设置资源限制
func (s *Sandbox) SetResource(name string, value int) {
	s.resources[name] = value
}

// GetResource 获取资源限制
func (s *Sandbox) GetResource(name string) int {
	return s.resources[name]
}
