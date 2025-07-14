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

// JavaScriptPOCExecutor JavaScript POC执行器
type JavaScriptPOCExecutor struct {
	name           string
	supportedTypes []string
	nodePath       string
	tempDir        string
	maxScriptSize  int64
	timeout        time.Duration
}

// NewJavaScriptPOCExecutor 创建JavaScript POC执行器
func NewJavaScriptPOCExecutor() *JavaScriptPOCExecutor {
	executor := &JavaScriptPOCExecutor{
		name:           "javascript",
		supportedTypes: []string{"javascript", "js", "node"},
		maxScriptSize:  1024 * 1024, // 1MB
		timeout:        30 * time.Second,
	}
	
	// 查找Node.js
	executor.nodePath = executor.findNodePath()
	
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "stellar_js_poc_*")
	if err != nil {
		tempDir = "/tmp"
	}
	executor.tempDir = tempDir
	
	return executor
}

// findNodePath 查找Node.js路径
func (e *JavaScriptPOCExecutor) findNodePath() string {
	candidates := []string{"node", "nodejs", "/usr/bin/node", "/usr/local/bin/node"}
	
	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return path
		}
	}
	
	return "node" // 默认值
}

// Execute 执行JavaScript POC
func (e *JavaScriptPOCExecutor) Execute(ctx context.Context, poc *models.POC, target POCTarget) (*models.POCResult, error) {
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
		result.Error = fmt.Sprintf("JavaScript脚本过大，超过限制 %d 字节", e.maxScriptSize)
		return result, nil
	}
	
	// 检查Node.js是否可用
	if err := e.checkNodeAvailable(); err != nil {
		result.Error = fmt.Sprintf("Node.js不可用: %v", err)
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
	
	// 构建Node.js执行命令
	cmd := exec.CommandContext(execCtx, e.nodePath, scriptFile)
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
			result.Error = "JavaScript脚本执行超时"
		} else {
			result.Error = fmt.Sprintf("JavaScript脚本执行失败: %v", err)
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

// checkNodeAvailable 检查Node.js是否可用
func (e *JavaScriptPOCExecutor) checkNodeAvailable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, e.nodePath, "--version")
	_, err := cmd.Output()
	return err
}

// createTempScript 创建临时JavaScript脚本文件
func (e *JavaScriptPOCExecutor) createTempScript(script string, target POCTarget) (string, error) {
	// 创建唯一的临时文件名
	filename := fmt.Sprintf("poc_%d_%s.js", time.Now().Unix(), target.Hash())
	scriptPath := filepath.Join(e.tempDir, filename)
	
	// 构建完整的JavaScript脚本
	fullScript := e.buildJavaScriptScript(script, target)
	
	// 写入文件
	return scriptPath, os.WriteFile(scriptPath, []byte(fullScript), 0600)
}

// buildJavaScriptScript 构建完整的JavaScript脚本
func (e *JavaScriptPOCExecutor) buildJavaScriptScript(userScript string, target POCTarget) string {
	// JavaScript脚本模板
	template := `#!/usr/bin/env node
/**
 * Stellar POC JavaScript执行器
 * 自动生成的POC脚本包装器
 */

const fs = require('fs');
const https = require('https');
const http = require('http');
const url = require('url');

// Stellar POC执行结果类
class StellarResult {
    constructor() {
        this.success = false;
        this.vulnerable = false;
        this.payload = "";
        this.request = "";
        this.response = "";
        this.output = "";
        this.error = "";
        this.data = {};
    }
    
    setSuccess(success = true) {
        this.success = success;
        this.vulnerable = success;
    }
    
    setPayload(payload) {
        this.payload = String(payload);
    }
    
    setRequest(request) {
        this.request = String(request);
    }
    
    setResponse(response) {
        this.response = String(response);
    }
    
    setOutput(output) {
        this.output = String(output);
    }
    
    setError(error) {
        this.error = String(error);
    }
    
    setData(key, value) {
        this.data[key] = value;
    }
    
    toJSON() {
        return {
            success: this.success,
            vulnerable: this.vulnerable,
            payload: this.payload,
            request: this.request,
            response: this.response,
            output: this.output,
            error: this.error,
            data: this.data
        };
    }
}

// 全局结果对象
const result = new StellarResult();

// 目标信息
const targetUrl = process.env.TARGET_URL || "";
const targetHost = process.env.TARGET_HOST || "";
const targetPort = parseInt(process.env.TARGET_PORT || "0");
const targetScheme = process.env.TARGET_SCHEME || "";
const targetPath = process.env.TARGET_PATH || "";

// 解析URL
let parsedUrl = {};
if (targetUrl) {
    try {
        parsedUrl = url.parse(targetUrl);
    } catch (e) {
        result.setError('URL解析失败: ' + e.message);
    }
}

// 辅助函数 - 发送HTTP请求
async function makeRequest(requestUrl, options = {}) {
    return new Promise((resolve, reject) => {
        try {
            const parsedUrl = url.parse(requestUrl);
            const isHttps = parsedUrl.protocol === 'https:';
            const httpModule = isHttps ? https : http;
            
            const requestOptions = {
                hostname: parsedUrl.hostname,
                port: parsedUrl.port || (isHttps ? 443 : 80),
                path: parsedUrl.path,
                method: options.method || 'GET',
                headers: {
                    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
                    ...options.headers
                },
                timeout: options.timeout || 10000,
                rejectUnauthorized: false // 忽略SSL证书错误
            };
            
            const req = httpModule.request(requestOptions, (res) => {
                let data = '';
                
                res.on('data', (chunk) => {
                    data += chunk;
                });
                
                res.on('end', () => {
                    const response = {
                        statusCode: res.statusCode,
                        headers: res.headers,
                        body: data,
                        url: requestUrl
                    };
                    
                    // 记录请求和响应
                    result.setRequest(requestOptions.method + ' ' + requestUrl);
                    result.setResponse('Status: ' + res.statusCode + 
                        '\\nHeaders: ' + JSON.stringify(res.headers) + 
                        '\\nBody: ' + data.substring(0, 1000));
                    
                    resolve(response);
                });
            });
            
            req.on('error', (error) => {
                result.setError('请求失败: ' + error.message);
                reject(error);
            });
            
            req.on('timeout', () => {
                req.destroy();
                const timeoutError = new Error('请求超时');
                result.setError('请求超时');
                reject(timeoutError);
            });
            
            // 发送请求体
            if (options.body) {
                req.write(options.body);
            }
            
            req.end();
        } catch (error) {
            result.setError('请求创建失败: ' + error.message);
            reject(error);
        }
    });
}

// 辅助函数 - 检查响应是否匹配漏洞特征
function checkResponse(response, patterns = [], statusCodes = []) {
    if (!response) {
        return false;
    }
    
    // 检查状态码
    if (statusCodes.length > 0 && statusCodes.includes(response.statusCode)) {
        return true;
    }
    
    // 检查响应内容
    if (patterns.length > 0) {
        const content = response.body.toLowerCase();
        for (const pattern of patterns) {
            if (content.includes(pattern.toLowerCase())) {
                result.setPayload(pattern);
                return true;
            }
        }
    }
    
    return false;
}

// 辅助函数 - 记录日志
function log(message) {
    result.setOutput(result.output + String(message) + '\\n');
}

// 包装执行用户POC脚本
async function executePOC() {
    try {
        // 用户POC脚本开始
%s
        // 用户POC脚本结束
    } catch (error) {
        result.setError('POC执行异常: ' + error.message + '\\n' + error.stack);
    }
}

// 主执行函数
async function main() {
    await executePOC();
    
    // 输出结果
    console.log('STELLAR_RESULT_START');
    console.log(JSON.stringify(result.toJSON(), null, 2));
    console.log('STELLAR_RESULT_END');
}

// 错误处理
process.on('uncaughtException', (error) => {
    result.setError('未捕获异常: ' + error.message);
    console.log('STELLAR_RESULT_START');
    console.log(JSON.stringify(result.toJSON(), null, 2));
    console.log('STELLAR_RESULT_END');
    process.exit(1);
});

process.on('unhandledRejection', (reason, promise) => {
    result.setError('未处理的Promise拒绝: ' + reason);
    console.log('STELLAR_RESULT_START');
    console.log(JSON.stringify(result.toJSON(), null, 2));
    console.log('STELLAR_RESULT_END');
    process.exit(1);
});

// 运行主函数
main().catch((error) => {
    result.setError('主函数执行失败: ' + error.message);
    console.log('STELLAR_RESULT_START');
    console.log(JSON.stringify(result.toJSON(), null, 2));
    console.log('STELLAR_RESULT_END');
    process.exit(1);
});
`
	
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
	
	return fmt.Sprintf(template, indentedScript)
}

// parseOutput 解析JavaScript脚本输出
func (e *JavaScriptPOCExecutor) parseOutput(output string, result *models.POCResult) error {
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
func (e *JavaScriptPOCExecutor) GetSupportedTypes() []string {
	return e.supportedTypes
}

// Validate 验证JavaScript POC脚本
func (e *JavaScriptPOCExecutor) Validate(poc *models.POC) error {
	// 检查脚本大小
	if int64(len(poc.Script)) > e.maxScriptSize {
		return fmt.Errorf("JavaScript脚本过大，超过限制 %d 字节", e.maxScriptSize)
	}
	
	// 检查脚本内容
	if strings.TrimSpace(poc.Script) == "" {
		return fmt.Errorf("JavaScript脚本内容不能为空")
	}
	
	// 检查危险函数调用
	dangerousPatterns := []string{
		`require\s*\(\s*['"']child_process['"]`,
		`require\s*\(\s*['"']fs['"].*\.write`,
		`eval\s*\(`,
		`Function\s*\(`,
		`setTimeout\s*\(.*eval`,
		`setInterval\s*\(.*eval`,
		`process\.exit\s*\(`,
		`process\.kill\s*\(`,
		`\.unlink\s*\(`,
		`\.rmdir\s*\(`,
		`\.mkdir\s*\(`,
	}
	
	for _, pattern := range dangerousPatterns {
		if matched, _ := regexp.MatchString(pattern, poc.Script); matched {
			return fmt.Errorf("JavaScript脚本包含潜在危险的函数调用: %s", pattern)
		}
	}
	
	// 基本语法检查
	if err := e.basicSyntaxCheck(poc.Script); err != nil {
		return fmt.Errorf("JavaScript脚本语法错误: %v", err)
	}
	
	return nil
}

// basicSyntaxCheck 基本语法检查
func (e *JavaScriptPOCExecutor) basicSyntaxCheck(script string) error {
	// 创建临时文件
	tempFile, err := os.CreateTemp(e.tempDir, "syntax_check_*.js")
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
	
	cmd := exec.CommandContext(ctx, e.nodePath, "--check", tempFile.Name())
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("语法检查失败: %s", string(output))
	}
	
	return nil
}

// GetName 获取执行器名称
func (e *JavaScriptPOCExecutor) GetName() string {
	return e.name
}

// Cleanup 清理资源
func (e *JavaScriptPOCExecutor) Cleanup() {
	if e.tempDir != "" && e.tempDir != "/tmp" {
		os.RemoveAll(e.tempDir)
	}
}

// SetNodePath 设置Node.js路径
func (e *JavaScriptPOCExecutor) SetNodePath(path string) {
	e.nodePath = path
}

// SetTimeout 设置执行超时时间
func (e *JavaScriptPOCExecutor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// SetMaxScriptSize 设置最大脚本大小
func (e *JavaScriptPOCExecutor) SetMaxScriptSize(size int64) {
	e.maxScriptSize = size
}

// InstallNode 安装Node.js（如果未安装）
func (e *JavaScriptPOCExecutor) InstallNode() error {
	// 检查是否已安装
	if e.checkNodeAvailable() == nil {
		return nil
	}
	
	return fmt.Errorf("Node.js未安装，请手动安装 Node.js")
}