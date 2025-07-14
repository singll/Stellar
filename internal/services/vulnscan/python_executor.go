package vulnscan

import (
	
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PythonPOCExecutor Python POC执行器
type PythonPOCExecutor struct {
	name           string
	supportedTypes []string
	pythonPath     string
	tempDir        string
	maxScriptSize  int64
	timeout        time.Duration
}

// NewPythonPOCExecutor 创建Python POC执行器
func NewPythonPOCExecutor() *PythonPOCExecutor {
	executor := &PythonPOCExecutor{
		name:           "python",
		supportedTypes: []string{"python", "py"},
		maxScriptSize:  1024 * 1024, // 1MB
		timeout:        30 * time.Second,
	}
	
	// 查找Python解释器
	executor.pythonPath = executor.findPythonPath()
	
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "stellar_python_poc_*")
	if err != nil {
		tempDir = "/tmp"
	}
	executor.tempDir = tempDir
	
	return executor
}

// findPythonPath 查找Python解释器路径
func (e *PythonPOCExecutor) findPythonPath() string {
	// 按优先级查找Python解释器
	candidates := []string{"python3", "python", "python3.9", "python3.8", "python3.7"}
	
	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return path
		}
	}
	
	return "python3" // 默认值
}

// Execute 执行Python POC
func (e *PythonPOCExecutor) Execute(ctx context.Context, poc *models.POC, target POCTarget) (*models.POCResult, error) {
	startTime := time.Now()
	
	result := &models.POCResult{
		ID:            primitive.NewObjectID(),
		POCID:         poc.ID,
		Target:        target.URL,
		Success:       false,
		CreatedAt:     startTime,
		ExecutionTime: 0,
		Params:        make(map[string]string),
	}
	
	// 验证脚本大小
	if int64(len(poc.Script)) > e.maxScriptSize {
		result.Error = fmt.Sprintf("Python脚本过大，超过限制 %d 字节", e.maxScriptSize)
		return result, nil
	}
	
	// 创建临时脚本文件
	scriptFile, err := e.createTempScript(poc.Script, target)
	if err != nil {
		result.Error = fmt.Sprintf("创建临时脚本失败: %v", err)
		return result, nil
	}
	defer os.Remove(scriptFile)
	
	// 设置执行上下文
	execCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()
	
	// 构建Python执行命令
	cmd := exec.CommandContext(execCtx, e.pythonPath, scriptFile)
	cmd.Dir = e.tempDir
	
	// 设置环境变量
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TARGET_URL=%s", target.URL),
		fmt.Sprintf("TARGET_HOST=%s", target.Host),
		fmt.Sprintf("TARGET_PORT=%d", target.Port),
		fmt.Sprintf("TARGET_SCHEME=%s", target.Scheme),
		fmt.Sprintf("TARGET_PATH=%s", target.Path),
	)
	
	// 添加自定义参数
	for key, value := range target.Extra {
		cmd.Env = append(cmd.Env, fmt.Sprintf("CUSTOM_%s=%s", strings.ToUpper(key), value))
		result.Params[key] = value
	}
	
	// 执行脚本
	output, err := cmd.CombinedOutput()
	result.ExecutionTime = time.Since(startTime).Milliseconds()
	
	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			result.Error = "Python脚本执行超时"
		} else {
			result.Error = fmt.Sprintf("Python脚本执行失败: %v", err)
		}
		result.Output = string(output)
		return result, nil
	}
	
	// 解析输出结果
	if err := e.parseOutput(string(output), result); err != nil {
		result.Error = fmt.Sprintf("解析脚本输出失败: %v", err)
		result.Output = string(output)
		return result, nil
	}
	
	return result, nil
}

// createTempScript 创建临时Python脚本文件
func (e *PythonPOCExecutor) createTempScript(script string, target POCTarget) (string, error) {
	// 创建唯一的临时文件名
	filename := fmt.Sprintf("poc_%d_%s.py", time.Now().Unix(), target.Hash())
	scriptPath := filepath.Join(e.tempDir, filename)
	
	// 构建完整的Python脚本
	fullScript := e.buildPythonScript(script, target)
	
	// 写入文件
	return scriptPath, os.WriteFile(scriptPath, []byte(fullScript), 0600)
}

// buildPythonScript 构建完整的Python脚本
func (e *PythonPOCExecutor) buildPythonScript(userScript string, target POCTarget) string {
	// Python脚本模板
	template := `#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Stellar POC Python执行器
自动生成的POC脚本包装器
"""

import sys
import os
import json
import time
import traceback
from urllib.parse import urlparse

# 导入常用库
try:
    import requests
    import urllib3
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
except ImportError:
    requests = None

# Stellar POC执行结果类
class StellarResult:
    def __init__(self):
        self.success = False
        self.vulnerable = False
        self.payload = ""
        self.request = ""
        self.response = ""
        self.output = ""
        self.error = ""
        self.data = {}
    
    def set_success(self, success=True):
        self.success = success
        self.vulnerable = success
    
    def set_payload(self, payload):
        self.payload = str(payload)
    
    def set_request(self, request):
        self.request = str(request)
    
    def set_response(self, response):
        self.response = str(response)
    
    def set_output(self, output):
        self.output = str(output)
    
    def set_error(self, error):
        self.error = str(error)
    
    def set_data(self, key, value):
        self.data[key] = value
    
    def to_json(self):
        return {
            "success": self.success,
            "vulnerable": self.vulnerable,
            "payload": self.payload,
            "request": self.request,
            "response": self.response,
            "output": self.output,
            "error": self.error,
            "data": self.data
        }

# 全局结果对象
result = StellarResult()

# 目标信息
target_url = os.getenv("TARGET_URL", "")
target_host = os.getenv("TARGET_HOST", "")
target_port = int(os.getenv("TARGET_PORT", "0"))
target_scheme = os.getenv("TARGET_SCHEME", "")
target_path = os.getenv("TARGET_PATH", "")

# 解析URL
if target_url:
    parsed = urlparse(target_url)
    if not target_host:
        target_host = parsed.hostname or ""
    if not target_port:
        target_port = parsed.port or (443 if parsed.scheme == "https" else 80)
    if not target_scheme:
        target_scheme = parsed.scheme or "http"
    if not target_path:
        target_path = parsed.path or "/"

# 辅助函数
def make_request(url, method="GET", headers=None, data=None, timeout=10, verify=False, **kwargs):
    """发送HTTP请求的辅助函数"""
    if not requests:
        result.set_error("requests库未安装")
        return None
    
    try:
        if headers is None:
            headers = {
                "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
            }
        
        response = requests.request(
            method=method.upper(),
            url=url,
            headers=headers,
            data=data,
            timeout=timeout,
            verify=verify,
            allow_redirects=True,
            **kwargs
        )
        
        # 记录请求和响应
        result.set_request(f"{method.upper()} {url}")
        result.set_response(f"Status: {response.status_code}\nHeaders: {dict(response.headers)}\nBody: {response.text[:1000]}")
        
        return response
    except Exception as e:
        result.set_error(f"请求失败: {str(e)}")
        return None

def check_response(response, patterns=None, status_codes=None):
    """检查响应是否匹配漏洞特征"""
    if not response:
        return False
    
    # 检查状态码
    if status_codes and response.status_code in status_codes:
        return True
    
    # 检查响应内容
    if patterns:
        content = response.text.lower()
        for pattern in patterns:
            if pattern.lower() in content:
                result.set_payload(pattern)
                return True
    
    return False

def log(message):
    """记录日志"""
    result.set_output(result.output + str(message) + "\n")

# 用户POC脚本开始
try:
%s
except Exception as e:
    result.set_error(f"POC执行异常: {str(e)}\n{traceback.format_exc()}")

# 输出结果
print("STELLAR_RESULT_START")
print(json.dumps(result.to_json(), ensure_ascii=False, indent=2))
print("STELLAR_RESULT_END")
`
	
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
	
	return fmt.Sprintf(template, indentedScript)
}

// parseOutput 解析Python脚本输出
func (e *PythonPOCExecutor) parseOutput(output string, result *models.POCResult) error {
	// 查找结果标记
	startMarker := "STELLAR_RESULT_START"
	endMarker := "STELLAR_RESULT_END"
	
	startIdx := strings.Index(output, startMarker)
	endIdx := strings.Index(output, endMarker)
	
	if startIdx == -1 || endIdx == -1 {
		// 没有找到标记，尝试从输出中提取有用信息
		result.Output = output
		
		// 简单启发式检查是否发现漏洞
		lowerOutput := strings.ToLower(output)
		vulnerableKeywords := []string{
			"vulnerable", "exploit", "success", "found", "detected",
			"漏洞", "成功", "发现", "检测到", "存在",
		}
		
		for _, keyword := range vulnerableKeywords {
			if strings.Contains(lowerOutput, keyword) {
				result.Success = true
				result.Output = output
				break
			}
		}
		
		return nil
	}
	
	// 提取JSON结果
	jsonStr := output[startIdx+len(startMarker):endIdx]
	jsonStr = strings.TrimSpace(jsonStr)
	
	var scriptResult map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &scriptResult); err != nil {
		return fmt.Errorf("解析JSON结果失败: %v", err)
	}
	
	// 解析结果字段
	if success, ok := scriptResult["success"].(bool); ok {
		result.Success = success
	}
	
	if payload, ok := scriptResult["payload"].(string); ok {
		result.Payload = payload
	}
	
	if request, ok := scriptResult["request"].(string); ok {
		result.Request = request
	}
	
	if response, ok := scriptResult["response"].(string); ok {
		result.Response = response
	}
	
	if output, ok := scriptResult["output"].(string); ok {
		result.Output = output
	}
	
	if error, ok := scriptResult["error"].(string); ok && error != "" {
		result.Error = error
		result.Success = false
	}
	
	return nil
}

// GetSupportedTypes 获取支持的脚本类型
func (e *PythonPOCExecutor) GetSupportedTypes() []string {
	return e.supportedTypes
}

// Validate 验证Python POC脚本
func (e *PythonPOCExecutor) Validate(poc *models.POC) error {
	// 检查脚本大小
	if int64(len(poc.Script)) > e.maxScriptSize {
		return fmt.Errorf("Python脚本过大，超过限制 %d 字节", e.maxScriptSize)
	}
	
	// 检查脚本内容
	if strings.TrimSpace(poc.Script) == "" {
		return fmt.Errorf("Python脚本内容不能为空")
	}
	
	// 检查危险函数调用
	dangerousPatterns := []string{
		`os\.system\s*\(`,
		`subprocess\.call\s*\(`,
		`eval\s*\(`,
		`exec\s*\(`,
		`__import__\s*\(`,
		`open\s*\(.+[,\s]+['"]['"]?w`,
		`file\s*\(`,
		`input\s*\(`,
		`raw_input\s*\(`,
	}
	
	for _, pattern := range dangerousPatterns {
		if matched, _ := regexp.MatchString(pattern, poc.Script); matched {
			return fmt.Errorf("Python脚本包含潜在危险的函数调用: %s", pattern)
		}
	}
	
	// 基本语法检查
	if err := e.basicSyntaxCheck(poc.Script); err != nil {
		return fmt.Errorf("Python脚本语法错误: %v", err)
	}
	
	return nil
}

// basicSyntaxCheck 基本语法检查
func (e *PythonPOCExecutor) basicSyntaxCheck(script string) error {
	// 创建临时文件
	tempFile, err := os.CreateTemp(e.tempDir, "syntax_check_*.py")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	
	// 写入脚本
	if _, err := tempFile.WriteString(script); err != nil {
		return err
	}
	tempFile.Close()
	
	// 运行语法检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, e.pythonPath, "-m", "py_compile", tempFile.Name())
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("语法检查失败: %s", string(output))
	}
	
	return nil
}

// GetName 获取执行器名称
func (e *PythonPOCExecutor) GetName() string {
	return e.name
}

// Cleanup 清理资源
func (e *PythonPOCExecutor) Cleanup() {
	if e.tempDir != "" && e.tempDir != "/tmp" {
		os.RemoveAll(e.tempDir)
	}
}

// SetPythonPath 设置Python解释器路径
func (e *PythonPOCExecutor) SetPythonPath(path string) {
	e.pythonPath = path
}

// SetTimeout 设置执行超时时间
func (e *PythonPOCExecutor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// SetMaxScriptSize 设置最大脚本大小
func (e *PythonPOCExecutor) SetMaxScriptSize(size int64) {
	e.maxScriptSize = size
}