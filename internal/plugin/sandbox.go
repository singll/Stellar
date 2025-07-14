package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Sandbox 插件沙箱
type Sandbox struct {
	workDir    string
	tempDir    string
	timeoutSec int
	resources  map[string]int
	
	// 新增的安全特性
	limits SandboxLimits
	accessControl AccessControl
	monitor *ResourceMonitor
	mutex sync.RWMutex
}

// SandboxLimits 沙盒资源限制
type SandboxLimits struct {
	// 最大内存使用（字节）
	MaxMemory int64 `json:"max_memory"`
	// 最大CPU时间（秒）
	MaxCPUTime time.Duration `json:"max_cpu_time"`
	// 最大执行时间（秒）
	MaxWallTime time.Duration `json:"max_wall_time"`
	// 最大文件大小（字节）
	MaxFileSize int64 `json:"max_file_size"`
	// 最大进程数
	MaxProcesses int `json:"max_processes"`
	// 最大文件描述符数
	MaxFileDescriptors int `json:"max_file_descriptors"`
	// 网络访问权限
	AllowNetwork bool `json:"allow_network"`
	// 文件系统访问权限
	AllowFileSystem bool `json:"allow_file_system"`
}

// AccessControl 访问控制
type AccessControl struct {
	// 允许的系统调用
	AllowedSyscalls []string `json:"allowed_syscalls"`
	// 禁止的系统调用
	ForbiddenSyscalls []string `json:"forbidden_syscalls"`
	// 允许访问的目录
	AllowedDirectories []string `json:"allowed_directories"`
	// 禁止访问的目录
	ForbiddenDirectories []string `json:"forbidden_directories"`
	// 允许的网络端口
	AllowedPorts []int `json:"allowed_ports"`
	// 环境变量白名单
	AllowedEnvVars []string `json:"allowed_env_vars"`
}

// ResourceMonitor 资源监控器
type ResourceMonitor struct {
	// 当前内存使用
	CurrentMemory int64 `json:"current_memory"`
	// 峰值内存使用
	PeakMemory int64 `json:"peak_memory"`
	// CPU使用时间
	CPUTime time.Duration `json:"cpu_time"`
	// 墙上时间
	WallTime time.Duration `json:"wall_time"`
	// 读取字节数
	ReadBytes int64 `json:"read_bytes"`
	// 写入字节数
	WriteBytes int64 `json:"write_bytes"`
	// 文件操作数
	FileOperations int64 `json:"file_operations"`
	// 网络操作数
	NetworkOperations int64 `json:"network_operations"`
	// 系统调用数
	SyscallCount int64 `json:"syscall_count"`
}

// SandboxConfig 沙箱配置
type SandboxConfig struct {
	WorkDir    string         // 工作目录
	TempDir    string         // 临时目录
	TimeoutSec int            // 超时时间（秒）
	Resources  map[string]int // 资源限制
}

// NewSandbox 创建沙箱（为执行引擎使用的增强版本）
func NewSandbox() *Sandbox {
	return &Sandbox{
		workDir:    os.TempDir(),
		tempDir:    os.TempDir(),
		timeoutSec: 30,
		resources:  make(map[string]int),
		limits: SandboxLimits{
			MaxMemory:          256 * 1024 * 1024, // 256MB
			MaxCPUTime:         30 * time.Second,   // 30秒
			MaxWallTime:        60 * time.Second,   // 60秒
			MaxFileSize:        10 * 1024 * 1024,   // 10MB
			MaxProcesses:       10,                 // 10个进程
			MaxFileDescriptors: 100,                // 100个文件描述符
			AllowNetwork:       false,              // 默认禁止网络访问
			AllowFileSystem:    true,               // 允许文件系统访问（限制在临时目录）
		},
		accessControl: AccessControl{
			AllowedSyscalls: []string{
				"read", "write", "open", "close", "stat",
				"mmap", "munmap", "brk", "rt_sigaction",
				"rt_sigprocmask", "getpid", "exit_group",
			},
			ForbiddenSyscalls: []string{
				"execve", "fork", "clone", "socket", "connect",
				"bind", "listen", "accept", "sendto", "recvfrom",
				"mount", "umount", "chroot", "setuid", "setgid",
			},
			AllowedDirectories: []string{
				"/tmp",
				"/var/tmp",
			},
			ForbiddenDirectories: []string{
				"/etc",
				"/bin",
				"/sbin",
				"/usr/bin",
				"/usr/sbin",
				"/root",
				"/home",
			},
			AllowedPorts:   []int{},
			AllowedEnvVars: []string{"PATH", "HOME", "USER", "LANG"},
		},
		monitor: &ResourceMonitor{},
	}
}

// NewSandboxWithConfig 创建沙箱（保持与现有代码兼容）
func NewSandboxWithConfig(config SandboxConfig) (*Sandbox, error) {
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
		limits: SandboxLimits{
			MaxMemory:          256 * 1024 * 1024, // 256MB
			MaxCPUTime:         30 * time.Second,   // 30秒
			MaxWallTime:        time.Duration(config.TimeoutSec) * time.Second,
			MaxFileSize:        10 * 1024 * 1024,   // 10MB
			MaxProcesses:       10,
			MaxFileDescriptors: 100,
			AllowNetwork:       false,
			AllowFileSystem:    true,
		},
		accessControl: AccessControl{
			AllowedDirectories: []string{config.TempDir, config.WorkDir},
			AllowedEnvVars:     []string{"PATH", "HOME", "USER", "LANG"},
		},
		monitor: &ResourceMonitor{},
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

// Execute 在沙盒中执行函数（为执行引擎专用）
func (s *Sandbox) Execute(ctx context.Context, fn func() (*ExecutionResult, error)) (*ExecutionResult, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 重置监控器
	s.monitor = &ResourceMonitor{}

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, s.limits.MaxWallTime)
	defer cancel()

	// 创建资源监控goroutine
	monitorDone := make(chan bool)
	go s.monitorResources(timeoutCtx, monitorDone)

	// 创建执行通道
	resultChan := make(chan *ExecutionResult, 1)
	errorChan := make(chan error, 1)

	// 启动执行
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errorChan <- fmt.Errorf("沙盒执行恐慌: %v", r)
			}
		}()

		result, err := fn()
		if err != nil {
			errorChan <- err
		} else {
			// 添加资源使用信息
			if result != nil {
				result.ResourceUsage = &ResourceUsage{
					CPUTime:    s.monitor.CPUTime,
					Memory:     s.monitor.CurrentMemory,
					MaxMemory:  s.monitor.PeakMemory,
					ReadBytes:  s.monitor.ReadBytes,
					WriteBytes: s.monitor.WriteBytes,
				}
			}
			resultChan <- result
		}
	}()

	// 等待执行完成或超时
	select {
	case result := <-resultChan:
		monitorDone <- true
		return result, nil
	case err := <-errorChan:
		monitorDone <- true
		return nil, err
	case <-timeoutCtx.Done():
		monitorDone <- true
		return nil, fmt.Errorf("沙盒执行超时: %v", s.limits.MaxWallTime)
	}
}

// monitorResources 监控资源使用
func (s *Sandbox) monitorResources(ctx context.Context, done <-chan bool) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-done:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 更新监控信息
			s.updateResourceUsage()

			// 检查资源限制
			if err := s.checkResourceLimits(); err != nil {
				// 资源超限，需要终止执行
				return
			}

			// 更新墙上时间
			s.monitor.WallTime = time.Since(startTime)
		}
	}
}

// updateResourceUsage 更新资源使用情况
func (s *Sandbox) updateResourceUsage() {
	// 获取当前进程内存使用
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	s.monitor.CurrentMemory = int64(memStats.Alloc)
	if s.monitor.CurrentMemory > s.monitor.PeakMemory {
		s.monitor.PeakMemory = s.monitor.CurrentMemory
	}
}

// checkResourceLimits 检查资源限制
func (s *Sandbox) checkResourceLimits() error {
	// 检查内存限制
	if s.monitor.PeakMemory > s.limits.MaxMemory {
		return fmt.Errorf("内存使用超限: %d > %d", s.monitor.PeakMemory, s.limits.MaxMemory)
	}

	// 检查CPU时间限制
	if s.monitor.CPUTime > s.limits.MaxCPUTime {
		return fmt.Errorf("CPU时间超限: %v > %v", s.monitor.CPUTime, s.limits.MaxCPUTime)
	}

	// 检查墙上时间限制
	if s.monitor.WallTime > s.limits.MaxWallTime {
		return fmt.Errorf("执行时间超限: %v > %v", s.monitor.WallTime, s.limits.MaxWallTime)
	}

	return nil
}

// CreateRestrictedCommand 创建受限制的命令
func (s *Sandbox) CreateRestrictedCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)

	// 设置环境变量
	cmd.Env = s.buildRestrictedEnv()

	// 设置工作目录
	cmd.Dir = s.tempDir

	// 设置资源限制（Linux/Unix特有）
	if runtime.GOOS == "linux" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			// 创建新的进程组
			Setpgid: true,
		}
	}

	return cmd
}

// buildRestrictedEnv 构建受限的环境变量
func (s *Sandbox) buildRestrictedEnv() []string {
	env := []string{}
	
	// 只添加允许的环境变量
	for _, envVar := range s.accessControl.AllowedEnvVars {
		if value := os.Getenv(envVar); value != "" {
			env = append(env, fmt.Sprintf("%s=%s", envVar, value))
		}
	}

	// 添加沙盒特定的环境变量
	env = append(env, "SANDBOX_MODE=1")
	env = append(env, "PLUGIN_EXECUTION=1")

	return env
}

// ValidateFileAccess 验证文件访问权限
func (s *Sandbox) ValidateFileAccess(path string, operation string) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 检查是否允许文件系统访问
	if !s.limits.AllowFileSystem {
		return fmt.Errorf("文件系统访问被禁止")
	}

	// 检查禁止访问的目录
	for _, forbidden := range s.accessControl.ForbiddenDirectories {
		if len(path) >= len(forbidden) && path[:len(forbidden)] == forbidden {
			return fmt.Errorf("访问被禁止的目录: %s", forbidden)
		}
	}

	// 检查是否在允许的目录中
	allowed := false
	for _, allowedDir := range s.accessControl.AllowedDirectories {
		if len(path) >= len(allowedDir) && path[:len(allowedDir)] == allowedDir {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("文件路径不在允许的目录中: %s", path)
	}

	return nil
}

// ValidateNetworkAccess 验证网络访问权限
func (s *Sandbox) ValidateNetworkAccess(host string, port int) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 检查是否允许网络访问
	if !s.limits.AllowNetwork {
		return fmt.Errorf("网络访问被禁止")
	}

	// 检查端口是否在允许列表中
	if len(s.accessControl.AllowedPorts) > 0 {
		allowed := false
		for _, allowedPort := range s.accessControl.AllowedPorts {
			if port == allowedPort {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("端口 %d 不在允许列表中", port)
		}
	}

	return nil
}

// SetLimits 设置沙盒限制
func (s *Sandbox) SetLimits(limits SandboxLimits) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.limits = limits
}

// GetLimits 获取沙盒限制
func (s *Sandbox) GetLimits() SandboxLimits {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.limits
}

// SetAccessControl 设置访问控制
func (s *Sandbox) SetAccessControl(control AccessControl) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.accessControl = control
}

// GetAccessControl 获取访问控制
func (s *Sandbox) GetAccessControl() AccessControl {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.accessControl
}

// GetResourceUsage 获取资源使用情况
func (s *Sandbox) GetResourceUsage() ResourceMonitor {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return *s.monitor
}

// Reset 重置沙盒状态
func (s *Sandbox) Reset() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.monitor = &ResourceMonitor{}
}

// GetStats 获取沙盒统计信息
func (s *Sandbox) GetStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return map[string]interface{}{
		"limits":         s.limits,
		"access_control": s.accessControl,
		"resource_usage": *s.monitor,
	}
}
