package vulnscan

import (
	"context"
	"encoding/json"
	"fmt"

	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/StellarServer/internal/models"
	pkgerrors "github.com/StellarServer/internal/pkg/errors"
	"github.com/StellarServer/internal/pkg/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GoPOCExecutor Go POC执行器
type GoPOCExecutor struct {
	name           string
	supportedTypes []string
	goPath         string
	tempDir        string
	maxScriptSize  int64
	timeout        time.Duration
	buildCache     map[string]string // 编译缓存
}

// NewGoPOCExecutor 创建Go POC执行器
func NewGoPOCExecutor() *GoPOCExecutor {
	executor := &GoPOCExecutor{
		name:           "go",
		supportedTypes: []string{"go", "golang"},
		maxScriptSize:  2 * 1024 * 1024,  // 2MB
		timeout:        60 * time.Second, // Go编译需要更长时间
		buildCache:     make(map[string]string),
	}

	// 查找Go编译器
	executor.goPath = executor.findGoPath()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "stellar_go_poc_*")
	if err != nil {
		tempDir = "/tmp"
	}
	executor.tempDir = tempDir

	return executor
}

// findGoPath 查找Go编译器路径
func (e *GoPOCExecutor) findGoPath() string {
	if path, err := exec.LookPath("go"); err == nil {
		return path
	}
	return "go" // 默认值
}

// Execute 执行Go POC
func (e *GoPOCExecutor) Execute(ctx context.Context, poc *models.POC, target POCTarget) (*models.POCResult, error) {
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
		result.Error = fmt.Sprintf("Go脚本过大，超过限制 %d 字节", e.maxScriptSize)
		return result, nil
	}

	// 检查缓存
	scriptHash := fmt.Sprintf("%x", []byte(poc.Script))
	var binaryPath string
	var err error

	if cachedBinary, exists := e.buildCache[scriptHash]; exists {
		if _, err := os.Stat(cachedBinary); err == nil {
			binaryPath = cachedBinary
		} else {
			delete(e.buildCache, scriptHash)
		}
	}

	// 编译Go程序
	if binaryPath == "" {
		binaryPath, err = e.compileGoScript(poc.Script, target, scriptHash)
		if err != nil {
			result.Error = fmt.Sprintf("编译Go脚本失败: %v", err)
			return result, nil
		}
		e.buildCache[scriptHash] = binaryPath
	}

	// 设置执行上下文
	execCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// 执行编译后的程序
	cmd := exec.CommandContext(execCtx, binaryPath)
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

	// 执行程序
	output, err := cmd.CombinedOutput()
	result.ExecutionTime = time.Since(startTime).Milliseconds()

	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			result.Error = "Go程序执行超时"
		} else {
			result.Error = fmt.Sprintf("Go程序执行失败: %v", err)
		}
		result.Output = string(output)
		return result, nil
	}

	// 解析输出结果
	if err := e.parseOutput(string(output), result); err != nil {
		result.Error = fmt.Sprintf("解析程序输出失败: %v", err)
		result.Output = string(output)
		return result, nil
	}

	return result, nil
}

// compileGoScript 编译Go脚本
func (e *GoPOCExecutor) compileGoScript(script string, target POCTarget, hash string) (string, error) {
	// 创建项目目录
	projectDir := filepath.Join(e.tempDir, fmt.Sprintf("poc_%s", hash))
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		logger.Error("compileGoScript create project dir failed", map[string]interface{}{"projectDir": projectDir, "error": err})
		return "", pkgerrors.WrapFileError(err, "创建Go项目目录")
	}

	// 创建go.mod文件
	goModContent := `module stellar-poc

go 1.19

require (
	github.com/go-resty/resty/v2 v2.7.0
)
`
	goModPath := filepath.Join(projectDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		logger.Error("compileGoScript write go.mod failed", map[string]interface{}{"goModPath": goModPath, "error": err})
		return "", pkgerrors.WrapFileError(err, "创建go.mod文件")
	}

	// 构建完整的Go程序
	fullScript := e.buildGoScript(script, target)

	// 创建main.go文件
	mainGoPath := filepath.Join(projectDir, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(fullScript), 0644); err != nil {
		logger.Error("compileGoScript write main.go failed", map[string]interface{}{"mainGoPath": mainGoPath, "error": err})
		return "", pkgerrors.WrapFileError(err, "创建main.go文件")
	}

	// 编译程序
	binaryPath := filepath.Join(projectDir, "poc")
	buildCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(buildCtx, e.goPath, "build", "-o", binaryPath, ".")
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	if output, err := cmd.CombinedOutput(); err != nil {
		logger.Error("compileGoScript build failed", map[string]interface{}{"projectDir": projectDir, "output": string(output), "error": err})
		return "", pkgerrors.WrapError(fmt.Errorf("编译失败: %v\n%s", err, string(output)), pkgerrors.CodePluginError, "编译Go脚本失败", 500)
	}

	return binaryPath, nil
}

// buildGoScript 构建完整的Go程序
func (e *GoPOCExecutor) buildGoScript(userScript string, target POCTarget) string {
	// Go程序模板
	template := `package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// StellarResult POC执行结果
type StellarResult struct {
	Success     bool              ` + "`json:\"success\"`" + `
	Vulnerable  bool              ` + "`json:\"vulnerable\"`" + `
	Payload     string            ` + "`json:\"payload\"`" + `
	Request     string            ` + "`json:\"request\"`" + `
	Response    string            ` + "`json:\"response\"`" + `
	Output      string            ` + "`json:\"output\"`" + `
	Error       string            ` + "`json:\"error\"`" + `
	Data        map[string]string ` + "`json:\"data\"`" + `
}

// SetSuccess 设置成功状态
func (r *StellarResult) SetSuccess(success bool) {
	r.Success = success
	r.Vulnerable = success
}

// SetPayload 设置payload
func (r *StellarResult) SetPayload(payload string) {
	r.Payload = payload
}

// SetRequest 设置请求信息
func (r *StellarResult) SetRequest(request string) {
	r.Request = request
}

// SetResponse 设置响应信息
func (r *StellarResult) SetResponse(response string) {
	r.Response = response
}

// SetOutput 设置输出信息
func (r *StellarResult) SetOutput(output string) {
	r.Output = output
}

// SetError 设置错误信息
func (r *StellarResult) SetError(error string) {
	r.Error = error
}

// SetData 设置自定义数据
func (r *StellarResult) SetData(key, value string) {
	if r.Data == nil {
		r.Data = make(map[string]string)
	}
	r.Data[key] = value
}

// Log 记录日志
func (r *StellarResult) Log(message string) {
	r.Output += message + "\n"
}

// 全局变量
var result = &StellarResult{
	Data: make(map[string]string),
}

var (
	targetURL    string
	targetHost   string
	targetPort   int
	targetScheme string
	targetPath   string
	client       *resty.Client
)

// MakeRequest 发送HTTP请求的辅助函数
func MakeRequest(method, url string, headers map[string]string, body interface{}) (*resty.Response, error) {
	req := client.R()
	
	// 设置默认头部
	req.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	
	// 设置自定义头部
	for k, v := range headers {
		req.SetHeader(k, v)
	}
	
	// 设置请求体
	if body != nil {
		switch v := body.(type) {
		case string:
			req.SetBody(v)
		case []byte:
			req.SetBody(v)
		default:
			req.SetBody(v)
		}
	}
	
	// 发送请求
	var resp *resty.Response
	var err error
	
	switch strings.ToUpper(method) {
	case "GET":
		resp, err = req.Get(url)
	case "POST":
		resp, err = req.Post(url)
	case "PUT":
		resp, err = req.Put(url)
	case "DELETE":
		resp, err = req.Delete(url)
	case "HEAD":
		resp, err = req.Head(url)
	case "OPTIONS":
		resp, err = req.Options(url)
	case "PATCH":
		resp, err = req.Patch(url)
	default:
		return nil, fmt.Errorf("不支持的HTTP方法: %s", method)
	}
	
	if err == nil && resp != nil {
		// 记录请求和响应
		result.SetRequest(fmt.Sprintf("%s %s", method, url))
		result.SetResponse(fmt.Sprintf("Status: %d\nHeaders: %v\nBody: %s", 
			resp.StatusCode(), resp.Header(), string(resp.Body())[:min(1000, len(resp.Body()))]))
	}
	
	return resp, err
}

// CheckResponse 检查响应是否匹配漏洞特征
func CheckResponse(resp *resty.Response, patterns []string, statusCodes []int) bool {
	if resp == nil {
		return false
	}
	
	// 检查状态码
	if len(statusCodes) > 0 {
		for _, code := range statusCodes {
			if resp.StatusCode() == code {
				return true
			}
		}
	}
	
	// 检查响应内容
	if len(patterns) > 0 {
		content := strings.ToLower(string(resp.Body()))
		for _, pattern := range patterns {
			if strings.Contains(content, strings.ToLower(pattern)) {
				result.SetPayload(pattern)
				return true
			}
		}
	}
	
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	// 初始化目标信息
	targetURL = os.Getenv("TARGET_URL")
	targetHost = os.Getenv("TARGET_HOST")
	targetScheme = os.Getenv("TARGET_SCHEME")
	targetPath = os.Getenv("TARGET_PATH")
	
	if portStr := os.Getenv("TARGET_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			targetPort = port
		}
	}
	
	// 解析URL
	if targetURL != "" {
		if parsed, err := url.Parse(targetURL); err == nil {
			if targetHost == "" {
				targetHost = parsed.Hostname()
			}
			if targetPort == 0 {
				if parsed.Port() != "" {
					if port, err := strconv.Atoi(parsed.Port()); err == nil {
						targetPort = port
					}
				} else {
					if parsed.Scheme == "https" {
						targetPort = 443
					} else {
						targetPort = 80
					}
				}
			}
			if targetScheme == "" {
				targetScheme = parsed.Scheme
			}
			if targetPath == "" {
				targetPath = parsed.Path
				if targetPath == "" {
					targetPath = "/"
				}
			}
		}
	}
	
	// 初始化HTTP客户端
	client = resty.New()
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(3)
	client.SetRetryWaitTime(1 * time.Second)
	client.SetTLSClientConfig(&resty.TLSClientConfig{InsecureSkipVerify: true})
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			result.SetError(fmt.Sprintf("程序panic: %v", r))
		}
		
		// 输出结果
		logger.Info("STELLAR_RESULT_START")
		if jsonData, err := json.MarshalIndent(result, "", "  "); err == nil {
			logger.Info(string(jsonData))
		} else {
			logger.Error("JSON序列化失败", map[string]interface{}{"error": err})
		}
		logger.Info("STELLAR_RESULT_END")
	}()
	
	// 用户POC代码开始
%s
	// 用户POC代码结束
}
`

	// 缩进用户脚本
	lines := strings.Split(userScript, "\n")
	indentedLines := make([]string, len(lines))
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			indentedLines[i] = "\t" + line
		} else {
			indentedLines[i] = ""
		}
	}
	indentedScript := strings.Join(indentedLines, "\n")

	return fmt.Sprintf(template, indentedScript)
}

// parseOutput 解析Go程序输出
func (e *GoPOCExecutor) parseOutput(output string, result *models.POCResult) error {
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
	jsonStr := output[startIdx+len(startMarker) : endIdx]
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
func (e *GoPOCExecutor) GetSupportedTypes() []string {
	return e.supportedTypes
}

// Validate 验证Go POC脚本
func (e *GoPOCExecutor) Validate(poc *models.POC) error {
	// 检查脚本大小
	if int64(len(poc.Script)) > e.maxScriptSize {
		return fmt.Errorf("Go脚本过大，超过限制 %d 字节", e.maxScriptSize)
	}

	// 检查脚本内容
	if strings.TrimSpace(poc.Script) == "" {
		return fmt.Errorf("Go脚本内容不能为空")
	}

	// 检查危险函数调用
	dangerousPatterns := []string{
		`os\.Remove`,
		`os\.RemoveAll`,
		`os\.Rename`,
		`os\.Create`,
		`os\.OpenFile`,
		`exec\.Command`,
		`exec\.CommandContext`,
		`ioutil\.WriteFile`,
		`os\.WriteFile`,
		`syscall\.`,
		`unsafe\.`,
	}

	for _, pattern := range dangerousPatterns {
		if matched, _ := regexp.MatchString(pattern, poc.Script); matched {
			return fmt.Errorf("Go脚本包含潜在危险的函数调用: %s", pattern)
		}
	}

	// Go语法检查
	if err := e.goSyntaxCheck(poc.Script); err != nil {
		return fmt.Errorf("Go脚本语法错误: %v", err)
	}

	return nil
}

// goSyntaxCheck Go语法检查
func (e *GoPOCExecutor) goSyntaxCheck(script string) error {
	// 构建一个完整的Go程序进行语法检查
	fullScript := `package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

func main() {
` + script + `
}`

	// 使用Go的AST解析器检查语法
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, "poc.go", fullScript, parser.ParseComments)
	if err != nil {
		return err
	}

	return nil
}

// GetName 获取执行器名称
func (e *GoPOCExecutor) GetName() string {
	return e.name
}

// Cleanup 清理资源
func (e *GoPOCExecutor) Cleanup() {
	if e.tempDir != "" && e.tempDir != "/tmp" {
		os.RemoveAll(e.tempDir)
	}
}

// SetGoPath 设置Go编译器路径
func (e *GoPOCExecutor) SetGoPath(path string) {
	e.goPath = path
}

// SetTimeout 设置执行超时时间
func (e *GoPOCExecutor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// SetMaxScriptSize 设置最大脚本大小
func (e *GoPOCExecutor) SetMaxScriptSize(size int64) {
	e.maxScriptSize = size
}

// ClearBuildCache 清理编译缓存
func (e *GoPOCExecutor) ClearBuildCache() {
	for _, binaryPath := range e.buildCache {
		os.Remove(binaryPath)
	}
	e.buildCache = make(map[string]string)
}
