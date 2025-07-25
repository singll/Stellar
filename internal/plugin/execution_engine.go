package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ExecutionEngine 插件执行引擎
type ExecutionEngine struct {
	// 沙盒配置
	sandbox *Sandbox
	// 执行器映射
	executors map[string]Executor
	// 并发控制
	semaphore chan struct{}
	// 统计信息
	stats ExecutionStats
	mutex sync.RWMutex
}

// Executor 脚本执行器接口
type Executor interface {
	Execute(ctx context.Context, script string, params map[string]interface{}) (*ExecutionResult, error)
	Validate(script string) error
	GetLanguage() string
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Success      bool                   `json:"success"`
	Data         interface{}            `json:"data"`
	Output       string                 `json:"output"`
	Error        string                 `json:"error"`
	ExitCode     int                    `json:"exit_code"`
	ExecutionTime time.Duration         `json:"execution_time"`
	ResourceUsage *ResourceUsage        `json:"resource_usage"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	CPUTime    time.Duration `json:"cpu_time"`
	Memory     int64         `json:"memory"`      // bytes
	MaxMemory  int64         `json:"max_memory"`  // bytes
	ReadBytes  int64         `json:"read_bytes"`
	WriteBytes int64         `json:"write_bytes"`
}

// ExecutionStats 执行统计
type ExecutionStats struct {
	TotalExecutions   int64         `json:"total_executions"`
	SuccessfulExecs   int64         `json:"successful_execs"`
	FailedExecs       int64         `json:"failed_execs"`
	AvgExecutionTime  time.Duration `json:"avg_execution_time"`
	LastExecutionTime time.Time     `json:"last_execution_time"`
}

// NewExecutionEngine 创建执行引擎
func NewExecutionEngine(maxConcurrency int) *ExecutionEngine {
	engine := &ExecutionEngine{
		sandbox:   NewSandbox(),
		executors: make(map[string]Executor),
		semaphore: make(chan struct{}, maxConcurrency),
	}
	
	// 注册默认执行器
	engine.registerDefaultExecutors()
	
	return engine
}

// registerDefaultExecutors 注册默认执行器
func (e *ExecutionEngine) registerDefaultExecutors() {
	e.executors["python"] = NewPythonExecutor()
	e.executors["javascript"] = NewJavaScriptExecutor()
	e.executors["shell"] = NewShellExecutor()
	e.executors["lua"] = NewLuaExecutor()
}

// Execute 执行插件脚本
func (e *ExecutionEngine) Execute(ctx context.Context, language, script string, params map[string]interface{}) (*ExecutionResult, error) {
	// 并发控制
	e.semaphore <- struct{}{}
	defer func() { <-e.semaphore }()
	
	e.mutex.Lock()
	e.stats.TotalExecutions++
	e.stats.LastExecutionTime = time.Now()
	e.mutex.Unlock()
	
	// 获取执行器
	executor, exists := e.executors[language]
	if !exists {
		e.mutex.Lock()
		e.stats.FailedExecs++
		e.mutex.Unlock()
		return nil, fmt.Errorf("不支持的脚本语言: %s", language)
	}
	
	// 验证脚本
	if err := executor.Validate(script); err != nil {
		e.mutex.Lock()
		e.stats.FailedExecs++
		e.mutex.Unlock()
		return nil, fmt.Errorf("脚本验证失败: %v", err)
	}
	
	// 在沙盒中执行
	result, err := e.sandbox.Execute(ctx, func() (*ExecutionResult, error) {
		return executor.Execute(ctx, script, params)
	})
	
	// 更新统计信息
	e.mutex.Lock()
	if err != nil {
		e.stats.FailedExecs++
	} else {
		e.stats.SuccessfulExecs++
		if result != nil {
			// 更新平均执行时间
			totalTime := e.stats.AvgExecutionTime * time.Duration(e.stats.SuccessfulExecs-1)
			e.stats.AvgExecutionTime = (totalTime + result.ExecutionTime) / time.Duration(e.stats.SuccessfulExecs)
		}
	}
	e.mutex.Unlock()
	
	return result, err
}

// GetStats 获取执行统计信息
func (e *ExecutionEngine) GetStats() ExecutionStats {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.stats
}

// PythonExecutor Python执行器
type PythonExecutor struct {
	pythonPath string
	tempDir    string
}

// NewPythonExecutor 创建Python执行器
func NewPythonExecutor() *PythonExecutor {
	return &PythonExecutor{
		pythonPath: "python3",
		tempDir:    "/tmp",
	}
}

// Execute 执行Python脚本
func (p *PythonExecutor) Execute(ctx context.Context, script string, params map[string]interface{}) (*ExecutionResult, error) {
	startTime := time.Now()
	
	// 构建完整的Python脚本
	fullScript := p.buildScript(script, params)
	
	// 创建临时文件
	tempFile := filepath.Join(p.tempDir, fmt.Sprintf("plugin_%d.py", time.Now().UnixNano()))
	if err := os.WriteFile(tempFile, []byte(fullScript), 0600); err != nil {
		return nil, fmt.Errorf("创建临时脚本文件失败: %v", err)
	}
	defer os.Remove(tempFile)
	
	// 执行脚本
	cmd := exec.CommandContext(ctx, p.pythonPath, tempFile)
	output, err := cmd.CombinedOutput()
	
	executionTime := time.Since(startTime)
	
	result := &ExecutionResult{
		Success:       err == nil,
		Output:        string(output),
		ExecutionTime: executionTime,
		Metadata:      make(map[string]interface{}),
	}
	
	if err != nil {
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
	} else {
		// 尝试解析JSON输出
		if data, parseErr := p.parseOutput(string(output)); parseErr == nil {
			result.Data = data
		}
	}
	
	return result, nil
}

// buildScript 构建Python脚本
func (p *PythonExecutor) buildScript(userScript string, params map[string]interface{}) string {
	template := `#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import json
import sys
import os
from datetime import datetime

# 插件参数
PLUGIN_PARAMS = %s

# 全局结果对象
result = {
    "success": False,
    "data": None,
    "message": "",
    "timestamp": datetime.now().isoformat()
}

def set_result(success, data=None, message=""):
    global result
    result["success"] = success
    result["data"] = data
    result["message"] = message

def log(message):
    print(f"[LOG] {message}", file=sys.stderr)

try:
    # 用户脚本开始
%s
    # 用户脚本结束
    
    if not result["success"] and result["data"] is None:
        result["success"] = True
        result["message"] = "执行完成"
        
except Exception as e:
    result["success"] = False
    result["error"] = str(e)
    result["message"] = f"执行异常: {e}"

# 输出结果
print("PLUGIN_RESULT_START")
print(json.dumps(result, ensure_ascii=False, default=str))
print("PLUGIN_RESULT_END")
`
	
	// 序列化参数
	paramsJSON, _ := json.Marshal(params)
	
	// 缩进用户脚本
	lines := strings.Split(userScript, "\n")
	indentedLines := make([]string, len(lines))
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			indentedLines[i] = "    " + line
		} else {
			indentedLines[i] = ""
		}
	}
	indentedScript := strings.Join(indentedLines, "\n")
	
	return fmt.Sprintf(template, string(paramsJSON), indentedScript)
}

// parseOutput 解析输出
func (p *PythonExecutor) parseOutput(output string) (interface{}, error) {
	startMarker := "PLUGIN_RESULT_START"
	endMarker := "PLUGIN_RESULT_END"
	
	startIdx := strings.Index(output, startMarker)
	endIdx := strings.Index(output, endMarker)
	
	if startIdx == -1 || endIdx == -1 {
		return nil, fmt.Errorf("未找到结果标记")
	}
	
	jsonStr := output[startIdx+len(startMarker):endIdx]
	jsonStr = strings.TrimSpace(jsonStr)
	
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("解析JSON结果失败: %v", err)
	}
	
	return result, nil
}

// Validate 验证Python脚本
func (p *PythonExecutor) Validate(script string) error {
	// 检查危险函数
	dangerousPatterns := []string{
		"os.system", "subprocess", "eval", "exec", "__import__",
		"open(", "file(", "input(", "raw_input(",
	}
	
	lowerScript := strings.ToLower(script)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerScript, pattern) {
			return fmt.Errorf("脚本包含潜在危险的函数调用: %s", pattern)
		}
	}
	
	return nil
}

// GetLanguage 获取语言名称
func (p *PythonExecutor) GetLanguage() string {
	return "python"
}

// JavaScriptExecutor JavaScript执行器
type JavaScriptExecutor struct {
	nodePath string
	tempDir  string
}

// NewJavaScriptExecutor 创建JavaScript执行器
func NewJavaScriptExecutor() *JavaScriptExecutor {
	return &JavaScriptExecutor{
		nodePath: "node",
		tempDir:  "/tmp",
	}
}

// Execute 执行JavaScript脚本
func (j *JavaScriptExecutor) Execute(ctx context.Context, script string, params map[string]interface{}) (*ExecutionResult, error) {
	startTime := time.Now()
	
	// 构建完整的JavaScript脚本
	fullScript := j.buildScript(script, params)
	
	// 创建临时文件
	tempFile := filepath.Join(j.tempDir, fmt.Sprintf("plugin_%d.js", time.Now().UnixNano()))
	if err := os.WriteFile(tempFile, []byte(fullScript), 0600); err != nil {
		return nil, fmt.Errorf("创建临时脚本文件失败: %v", err)
	}
	defer os.Remove(tempFile)
	
	// 执行脚本
	cmd := exec.CommandContext(ctx, j.nodePath, tempFile)
	output, err := cmd.CombinedOutput()
	
	executionTime := time.Since(startTime)
	
	result := &ExecutionResult{
		Success:       err == nil,
		Output:        string(output),
		ExecutionTime: executionTime,
		Metadata:      make(map[string]interface{}),
	}
	
	if err != nil {
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
	} else {
		// 尝试解析JSON输出
		if data, parseErr := j.parseOutput(string(output)); parseErr == nil {
			result.Data = data
		}
	}
	
	return result, nil
}

// buildScript 构建JavaScript脚本
func (j *JavaScriptExecutor) buildScript(userScript string, params map[string]interface{}) string {
	template := `#!/usr/bin/env node

// 插件参数
const PLUGIN_PARAMS = %s;

// 全局结果对象
const result = {
    success: false,
    data: null,
    message: "",
    timestamp: new Date().toISOString()
};

function setResult(success, data = null, message = "") {
    result.success = success;
    result.data = data;
    result.message = message;
}

function log(message) {
    console.error('[LOG]', message);
}

// 用户脚本包装
(async function() {
    try {
        // 用户脚本开始
%s
        // 用户脚本结束
        
        if (!result.success && result.data === null) {
            result.success = true;
            result.message = "执行完成";
        }
        
    } catch (error) {
        result.success = false;
        result.error = error.message;
        result.message = "执行异常: " + error.message;
    }
    
    // 输出结果
    console.log("PLUGIN_RESULT_START");
    console.log(JSON.stringify(result, null, 2));
    console.log("PLUGIN_RESULT_END");
})();
`
	
	// 序列化参数
	paramsJSON, _ := json.Marshal(params)
	
	// 缩进用户脚本
	lines := strings.Split(userScript, "\n")
	indentedLines := make([]string, len(lines))
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			indentedLines[i] = "        " + line
		} else {
			indentedLines[i] = ""
		}
	}
	indentedScript := strings.Join(indentedLines, "\n")
	
	return fmt.Sprintf(template, string(paramsJSON), indentedScript)
}

// parseOutput 解析输出（与Python相同）
func (j *JavaScriptExecutor) parseOutput(output string) (interface{}, error) {
	startMarker := "PLUGIN_RESULT_START"
	endMarker := "PLUGIN_RESULT_END"
	
	startIdx := strings.Index(output, startMarker)
	endIdx := strings.Index(output, endMarker)
	
	if startIdx == -1 || endIdx == -1 {
		return nil, fmt.Errorf("未找到结果标记")
	}
	
	jsonStr := output[startIdx+len(startMarker):endIdx]
	jsonStr = strings.TrimSpace(jsonStr)
	
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("解析JSON结果失败: %v", err)
	}
	
	return result, nil
}

// Validate 验证JavaScript脚本
func (j *JavaScriptExecutor) Validate(script string) error {
	// 检查危险函数
	dangerousPatterns := []string{
		"require('child_process')", "require(\"child_process\")",
		"eval(", "Function(", "setTimeout(", "setInterval(",
		"process.exit", "process.kill",
	}
	
	for _, pattern := range dangerousPatterns {
		if strings.Contains(script, pattern) {
			return fmt.Errorf("脚本包含潜在危险的函数调用: %s", pattern)
		}
	}
	
	return nil
}

// GetLanguage 获取语言名称
func (j *JavaScriptExecutor) GetLanguage() string {
	return "javascript"
}

// ShellExecutor Shell执行器
type ShellExecutor struct {
	shellPath string
	tempDir   string
}

// NewShellExecutor 创建Shell执行器
func NewShellExecutor() *ShellExecutor {
	return &ShellExecutor{
		shellPath: "/bin/bash",
		tempDir:   "/tmp",
	}
}

// Execute 执行Shell脚本
func (s *ShellExecutor) Execute(ctx context.Context, script string, params map[string]interface{}) (*ExecutionResult, error) {
	startTime := time.Now()
	
	// 构建完整的Shell脚本
	fullScript := s.buildScript(script, params)
	
	// 创建临时文件
	tempFile := filepath.Join(s.tempDir, fmt.Sprintf("plugin_%d.sh", time.Now().UnixNano()))
	if err := os.WriteFile(tempFile, []byte(fullScript), 0700); err != nil {
		return nil, fmt.Errorf("创建临时脚本文件失败: %v", err)
	}
	defer os.Remove(tempFile)
	
	// 执行脚本
	cmd := exec.CommandContext(ctx, s.shellPath, tempFile)
	output, err := cmd.CombinedOutput()
	
	executionTime := time.Since(startTime)
	
	result := &ExecutionResult{
		Success:       err == nil,
		Output:        string(output),
		ExecutionTime: executionTime,
		Metadata:      make(map[string]interface{}),
	}
	
	if err != nil {
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
	}
	
	return result, nil
}

// buildScript 构建Shell脚本
func (s *ShellExecutor) buildScript(userScript string, params map[string]interface{}) string {
	template := `#!/bin/bash
set -e

# 插件参数（通过环境变量传递）
%s

# 用户脚本开始
%s
# 用户脚本结束
`
	
	// 构建环境变量
	var envVars []string
	for key, value := range params {
		envVars = append(envVars, fmt.Sprintf("export PLUGIN_%s='%v'", strings.ToUpper(key), value))
	}
	envString := strings.Join(envVars, "\n")
	
	return fmt.Sprintf(template, envString, userScript)
}

// Validate 验证Shell脚本
func (s *ShellExecutor) Validate(script string) error {
	// 检查危险命令
	dangerousCommands := []string{
		"rm -rf", "dd if=", "mkfs", "fdisk",
		"wget", "curl", "nc ", "netcat",
		"chmod 777", "chown root",
	}
	
	lowerScript := strings.ToLower(script)
	for _, cmd := range dangerousCommands {
		if strings.Contains(lowerScript, cmd) {
			return fmt.Errorf("脚本包含潜在危险的命令: %s", cmd)
		}
	}
	
	return nil
}

// GetLanguage 获取语言名称
func (s *ShellExecutor) GetLanguage() string {
	return "shell"
}

// LuaExecutor Lua执行器
type LuaExecutor struct {
	luaPath string
	tempDir string
}

// NewLuaExecutor 创建Lua执行器
func NewLuaExecutor() *LuaExecutor {
	return &LuaExecutor{
		luaPath: "lua",
		tempDir: "/tmp",
	}
}

// Execute 执行Lua脚本
func (l *LuaExecutor) Execute(ctx context.Context, script string, params map[string]interface{}) (*ExecutionResult, error) {
	startTime := time.Now()
	
	// 构建完整的Lua脚本
	fullScript := l.buildScript(script, params)
	
	// 创建临时文件
	tempFile := filepath.Join(l.tempDir, fmt.Sprintf("plugin_%d.lua", time.Now().UnixNano()))
	if err := os.WriteFile(tempFile, []byte(fullScript), 0600); err != nil {
		return nil, fmt.Errorf("创建临时脚本文件失败: %v", err)
	}
	defer os.Remove(tempFile)
	
	// 执行脚本
	cmd := exec.CommandContext(ctx, l.luaPath, tempFile)
	output, err := cmd.CombinedOutput()
	
	executionTime := time.Since(startTime)
	
	result := &ExecutionResult{
		Success:       err == nil,
		Output:        string(output),
		ExecutionTime: executionTime,
		Metadata:      make(map[string]interface{}),
	}
	
	if err != nil {
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
	}
	
	return result, nil
}

// buildScript 构建Lua脚本
func (l *LuaExecutor) buildScript(userScript string, params map[string]interface{}) string {
	template := `#!/usr/bin/env lua

-- JSON库（简化实现）
local json = {}
function json.encode(obj)
    if type(obj) == "table" then
        local items = {}
        for k, v in pairs(obj) do
            table.insert(items, string.format('"%s":%s', k, json.encode(v)))
        end
        return "{" .. table.concat(items, ",") .. "}"
    elseif type(obj) == "string" then
        return '"' .. obj .. '"'
    elseif type(obj) == "boolean" then
        return obj and "true" or "false"
    else
        return tostring(obj)
    end
end

-- 插件参数
local PLUGIN_PARAMS = %s

-- 全局结果对象
local result = {
    success = false,
    data = nil,
    message = "",
    timestamp = os.date("%%Y-%%m-%%dT%%H:%%M:%%S")
}

function set_result(success, data, message)
    result.success = success or false
    result.data = data
    result.message = message or ""
end

function log(message)
    io.stderr:write("[LOG] " .. tostring(message) .. "\n")
end

-- 用户脚本开始
%s
-- 用户脚本结束

if not result.success and result.data == nil then
    result.success = true
    result.message = "执行完成"
end

-- 输出结果
print("PLUGIN_RESULT_START")
print(json.encode(result))
print("PLUGIN_RESULT_END")
`
	
	// 序列化参数（简化处理）
	paramsStr := "{}"
	
	return fmt.Sprintf(template, paramsStr, userScript)
}

// Validate 验证Lua脚本
func (l *LuaExecutor) Validate(script string) error {
	// 检查危险函数
	dangerousPatterns := []string{
		"os.execute", "io.popen", "load(", "loadfile(",
		"dofile(", "require(", "package.loadlib(",
	}
	
	lowerScript := strings.ToLower(script)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerScript, pattern) {
			return fmt.Errorf("脚本包含潜在危险的函数调用: %s", pattern)
		}
	}
	
	return nil
}

// GetLanguage 获取语言名称
func (l *LuaExecutor) GetLanguage() string {
	return "lua"
}