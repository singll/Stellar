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
	"gopkg.in/yaml.v3"
)

// NucleiPOCExecutor Nuclei模板执行器
type NucleiPOCExecutor struct {
	name           string
	supportedTypes []string
	nucleiPath     string
	tempDir        string
	maxTemplateSize int64
	timeout        time.Duration
}

// NucleiTemplate Nuclei模板结构
type NucleiTemplate struct {
	ID   string `yaml:"id"`
	Info struct {
		Name        string   `yaml:"name"`
		Author      string   `yaml:"author"`
		Severity    string   `yaml:"severity"`
		Description string   `yaml:"description"`
		Reference   []string `yaml:"reference"`
		Tags        []string `yaml:"tags"`
	} `yaml:"info"`
	Requests []struct {
		Method  string            `yaml:"method"`
		Path    []string          `yaml:"path"`
		Headers map[string]string `yaml:"headers"`
		Body    string            `yaml:"body"`
		Matchers []struct {
			Type      string   `yaml:"type"`
			Status    []int    `yaml:"status"`
			Words     []string `yaml:"words"`
			Regex     []string `yaml:"regex"`
			Condition string   `yaml:"condition"`
			Part      string   `yaml:"part"`
		} `yaml:"matchers"`
		Extractors []struct {
			Type  string   `yaml:"type"`
			Part  string   `yaml:"part"`
			Group int      `yaml:"group"`
			Regex []string `yaml:"regex"`
		} `yaml:"extractors"`
	} `yaml:"requests"`
	Variables map[string]string `yaml:"variables"`
}

// NucleiResult Nuclei执行结果
type NucleiResult struct {
	TemplateID   string `json:"template-id"`
	TemplatePath string `json:"template-path"`
	Info         struct {
		Name        string   `json:"name"`
		Author      []string `json:"author"`
		Severity    string   `json:"severity"`
		Description string   `json:"description"`
		Reference   []string `json:"reference"`
		Tags        []string `json:"tags"`
	} `json:"info"`
	Type      string `json:"type"`
	Host      string `json:"host"`
	MatchedAt string `json:"matched-at"`
	Request   string `json:"request,omitempty"`
	Response  string `json:"response,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// NewNucleiPOCExecutor 创建Nuclei执行器
func NewNucleiPOCExecutor() *NucleiPOCExecutor {
	executor := &NucleiPOCExecutor{
		name:            "nuclei",
		supportedTypes:  []string{"nuclei", "template"},
		maxTemplateSize: 512 * 1024, // 512KB
		timeout:         60 * time.Second,
	}
	
	// 查找Nuclei
	executor.nucleiPath = executor.findNucleiPath()
	
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "stellar_nuclei_poc_*")
	if err != nil {
		tempDir = "/tmp"
	}
	executor.tempDir = tempDir
	
	return executor
}

// findNucleiPath 查找Nuclei路径
func (e *NucleiPOCExecutor) findNucleiPath() string {
	candidates := []string{"nuclei", "/usr/local/bin/nuclei", "/usr/bin/nuclei"}
	
	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return path
		}
	}
	
	return "nuclei" // 默认值
}

// Execute 执行Nuclei模板
func (e *NucleiPOCExecutor) Execute(ctx context.Context, poc *models.POC, target POCTarget) (*models.POCResult, error) {
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
	
	// 验证模板大小
	if int64(len(poc.Script)) > e.maxTemplateSize {
		result.Error = fmt.Sprintf("Nuclei模板过大，超过限制 %d 字节", e.maxTemplateSize)
		return result, nil
	}
	
	// 检查Nuclei是否可用
	if err := e.checkNucleiAvailable(); err != nil {
		result.Error = fmt.Sprintf("Nuclei不可用: %v", err)
		return result, nil
	}
	
	// 创建临时模板文件
	templateFile, err := e.createTempTemplate(poc.Script, poc.ID.Hex())
	if err != nil {
		result.Error = fmt.Sprintf("创建临时模板失败: %v", err)
		return result, nil
	}
	defer os.Remove(templateFile)
	
	// 设置执行上下文
	execCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()
	
	// 构建Nuclei命令
	args := []string{
		"-t", templateFile,
		"-target", target.URL,
		"-json",
		"-silent",
		"-no-color",
		"-disable-update-check",
	}
	
	// 添加自定义头部
	if headers := target.Extra["headers"]; headers != "" {
		args = append(args, "-H", headers)
	}
	
	// 添加代理
	if proxy := target.Extra["proxy"]; proxy != "" {
		args = append(args, "-proxy", proxy)
	}
	
	// 添加超时设置
	args = append(args, "-timeout", "30")
	
	cmd := exec.CommandContext(execCtx, e.nucleiPath, args...)
	cmd.Dir = e.tempDir
	
	// 记录参数
	for key, value := range target.Extra {
		result.Params[key] = value
	}
	
	// 执行Nuclei
	output, err := cmd.CombinedOutput()
	result.ExecutionTime = time.Since(startTime).Milliseconds()
	
	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			result.Error = "Nuclei执行超时"
		} else {
			result.Error = fmt.Sprintf("Nuclei执行失败: %v", err)
		}
		result.Output = string(output)
		return result, nil
	}
	
	// 解析Nuclei输出
	if err := e.parseNucleiOutput(string(output), result); err != nil {
		result.Error = fmt.Sprintf("解析Nuclei输出失败: %v", err)
		result.Output = string(output)
		return result, nil
	}
	
	return result, nil
}

// checkNucleiAvailable 检查Nuclei是否可用
func (e *NucleiPOCExecutor) checkNucleiAvailable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, e.nucleiPath, "-version")
	_, err := cmd.Output()
	return err
}

// createTempTemplate 创建临时模板文件
func (e *NucleiPOCExecutor) createTempTemplate(template, pocID string) (string, error) {
	// 创建唯一的临时文件名
	filename := fmt.Sprintf("nuclei_template_%s_%d.yaml", pocID, time.Now().Unix())
	templatePath := filepath.Join(e.tempDir, filename)
	
	// 验证和规范化模板
	normalizedTemplate, err := e.normalizeTemplate(template)
	if err != nil {
		return "", fmt.Errorf("模板规范化失败: %v", err)
	}
	
	// 写入文件
	return templatePath, os.WriteFile(templatePath, []byte(normalizedTemplate), 0600)
}

// normalizeTemplate 规范化Nuclei模板
func (e *NucleiPOCExecutor) normalizeTemplate(template string) (string, error) {
	// 解析YAML模板
	var nucleiTemplate NucleiTemplate
	if err := yaml.Unmarshal([]byte(template), &nucleiTemplate); err != nil {
		return "", fmt.Errorf("YAML解析失败: %v", err)
	}
	
	// 确保必要字段存在
	if nucleiTemplate.ID == "" {
		nucleiTemplate.ID = fmt.Sprintf("stellar-poc-%d", time.Now().Unix())
	}
	
	if nucleiTemplate.Info.Name == "" {
		nucleiTemplate.Info.Name = "Stellar POC Template"
	}
	
	if nucleiTemplate.Info.Severity == "" {
		nucleiTemplate.Info.Severity = "info"
	}
	
	// 规范化severity
	nucleiTemplate.Info.Severity = strings.ToLower(nucleiTemplate.Info.Severity)
	validSeverities := map[string]bool{
		"info": true, "low": true, "medium": true, "high": true, "critical": true,
	}
	if !validSeverities[nucleiTemplate.Info.Severity] {
		nucleiTemplate.Info.Severity = "info"
	}
	
	// 重新序列化
	normalizedBytes, err := yaml.Marshal(&nucleiTemplate)
	if err != nil {
		return "", fmt.Errorf("YAML序列化失败: %v", err)
	}
	
	return string(normalizedBytes), nil
}

// parseNucleiOutput 解析Nuclei输出
func (e *NucleiPOCExecutor) parseNucleiOutput(output string, result *models.POCResult) error {
	result.Output = output
	
	if strings.TrimSpace(output) == "" {
		// 没有输出，表示没有匹配到漏洞
		result.Success = false
		return nil
	}
	
	// 尝试解析JSON格式的输出
	lines := strings.Split(strings.TrimSpace(output), "\n")
	foundVulnerability := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		var nucleiResult NucleiResult
		if err := json.Unmarshal([]byte(line), &nucleiResult); err != nil {
			// 如果不是JSON格式，检查是否包含漏洞关键词
			lowerLine := strings.ToLower(line)
			if strings.Contains(lowerLine, "found") || strings.Contains(lowerLine, "detected") || 
			   strings.Contains(lowerLine, "vulnerable") || strings.Contains(lowerLine, "[") {
				foundVulnerability = true
				result.Output = line
			}
			continue
		}
		
		// 解析JSON结果
		foundVulnerability = true
		result.Success = true
		result.Payload = nucleiResult.TemplateID
		
		// 构建请求信息
		if nucleiResult.Request != "" {
			result.Request = nucleiResult.Request
		} else {
			result.Request = fmt.Sprintf("Template: %s\nTarget: %s", nucleiResult.TemplateID, nucleiResult.Host)
		}
		
		// 构建响应信息
		if nucleiResult.Response != "" {
			result.Response = nucleiResult.Response
		} else {
			result.Response = fmt.Sprintf("Match: %s\nSeverity: %s\nDescription: %s", 
				nucleiResult.MatchedAt, nucleiResult.Info.Severity, nucleiResult.Info.Description)
		}
		
		// 只处理第一个结果
		break
	}
	
	result.Success = foundVulnerability
	return nil
}

// GetSupportedTypes 获取支持的模板类型
func (e *NucleiPOCExecutor) GetSupportedTypes() []string {
	return e.supportedTypes
}

// Validate 验证Nuclei模板
func (e *NucleiPOCExecutor) Validate(poc *models.POC) error {
	// 检查模板大小
	if int64(len(poc.Script)) > e.maxTemplateSize {
		return fmt.Errorf("Nuclei模板过大，超过限制 %d 字节", e.maxTemplateSize)
	}
	
	// 检查模板内容
	if strings.TrimSpace(poc.Script) == "" {
		return fmt.Errorf("Nuclei模板内容不能为空")
	}
	
	// 验证YAML格式
	var template NucleiTemplate
	if err := yaml.Unmarshal([]byte(poc.Script), &template); err != nil {
		return fmt.Errorf("Nuclei模板YAML格式错误: %v", err)
	}
	
	// 验证必要字段
	if template.Info.Name == "" {
		return fmt.Errorf("Nuclei模板缺少info.name字段")
	}
	
	if len(template.Requests) == 0 {
		return fmt.Errorf("Nuclei模板缺少requests字段")
	}
	
	// 验证severity
	validSeverities := []string{"info", "low", "medium", "high", "critical"}
	severityValid := false
	for _, valid := range validSeverities {
		if strings.ToLower(template.Info.Severity) == valid {
			severityValid = true
			break
		}
	}
	if !severityValid {
		return fmt.Errorf("Nuclei模板severity字段无效，必须是: %s", strings.Join(validSeverities, ", "))
	}
	
	// 检查危险操作
	dangerousPatterns := []string{
		`file://`,
		`\$\{.*\}`,  // 表达式注入
		`<%.*%>`,    // 模板注入
		`{{.*}}`,    // 另一种模板注入
	}
	
	for _, pattern := range dangerousPatterns {
		if matched, _ := regexp.MatchString(pattern, poc.Script); matched {
			return fmt.Errorf("Nuclei模板包含潜在危险的内容: %s", pattern)
		}
	}
	
	// 验证每个request
	for i, req := range template.Requests {
		if len(req.Path) == 0 {
			return fmt.Errorf("Nuclei模板第%d个request缺少path字段", i+1)
		}
		
		if len(req.Matchers) == 0 {
			return fmt.Errorf("Nuclei模板第%d个request缺少matchers字段", i+1)
		}
		
		// 验证每个matcher
		for j, matcher := range req.Matchers {
			if matcher.Type == "" {
				return fmt.Errorf("Nuclei模板第%d个request第%d个matcher缺少type字段", i+1, j+1)
			}
			
			validMatcherTypes := []string{"status", "word", "regex", "binary", "size", "dsl"}
			typeValid := false
			for _, valid := range validMatcherTypes {
				if matcher.Type == valid {
					typeValid = true
					break
				}
			}
			if !typeValid {
				return fmt.Errorf("Nuclei模板第%d个request第%d个matcher的type字段无效", i+1, j+1)
			}
		}
	}
	
	return nil
}

// GetName 获取执行器名称
func (e *NucleiPOCExecutor) GetName() string {
	return e.name
}

// Cleanup 清理资源
func (e *NucleiPOCExecutor) Cleanup() {
	if e.tempDir != "" && e.tempDir != "/tmp" {
		os.RemoveAll(e.tempDir)
	}
}

// SetNucleiPath 设置Nuclei路径
func (e *NucleiPOCExecutor) SetNucleiPath(path string) {
	e.nucleiPath = path
}

// SetTimeout 设置执行超时时间
func (e *NucleiPOCExecutor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// SetMaxTemplateSize 设置最大模板大小
func (e *NucleiPOCExecutor) SetMaxTemplateSize(size int64) {
	e.maxTemplateSize = size
}

// InstallNuclei 安装Nuclei（如果未安装）
func (e *NucleiPOCExecutor) InstallNuclei() error {
	// 检查是否已安装
	if e.checkNucleiAvailable() == nil {
		return nil
	}
	
	// 尝试使用go install安装
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "go", "install", "-v", "github.com/projectdiscovery/nuclei/v2/cmd/nuclei@latest")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("安装Nuclei失败: %v\n%s", err, string(output))
	}
	
	// 重新查找Nuclei路径
	e.nucleiPath = e.findNucleiPath()
	return nil
}