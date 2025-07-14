package sensitive

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SimpleLogger 简单日志接口 - TODO: 根据DEV_PLAN 0.6版本需要完善
type SimpleLogger interface {
	Error(msg string, args ...interface{})
	Info(msg string, args ...interface{})
}

// DefaultLogger 默认日志实现
type DefaultLogger struct{}

func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] %s %v", msg, args)
}

func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] %s %v", msg, args)
}

// DetectionEngine 敏感信息检测引擎
type DetectionEngine struct {
	db       *mongo.Database
	rules    []*DetectionRule
	config   *DetectionConfig
	ruleMap  map[string]*DetectionRule
	client   *http.Client
}

// DetectionConfig 检测配置
type DetectionConfig struct {
	// 检测配置
	MaxContentSize     int           `json:"max_content_size"`     // 最大内容大小（字节）
	Timeout           time.Duration `json:"timeout"`              // 请求超时时间
	UserAgent         string        `json:"user_agent"`           // User-Agent
	FollowRedirects   bool          `json:"follow_redirects"`     // 是否跟随重定向
	MaxRedirects      int           `json:"max_redirects"`        // 最大重定向次数
	
	// 规则配置
	EnabledCategories []string      `json:"enabled_categories"`   // 启用的检测类别
	Severity          string        `json:"severity"`             // 最低检测级别
	CustomRules       []*DetectionRule `json:"custom_rules"`      // 自定义规则
	
	// 输出配置
	OutputFormats     []string      `json:"output_formats"`       // 输出格式
	IncludeContext    bool          `json:"include_context"`      // 是否包含上下文
	ContextLines      int           `json:"context_lines"`        // 上下文行数
	
	// 过滤配置
	WhitelistDomains  []string      `json:"whitelist_domains"`    // 白名单域名
	BlacklistPatterns []string      `json:"blacklist_patterns"`   // 黑名单模式
}

// DetectionRule 检测规则
type DetectionRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Severity    SeverityLevel     `json:"severity"`
	Enabled     bool              `json:"enabled"`
	Pattern     string            `json:"pattern"`
	Regex       *regexp.Regexp    `json:"-"`
	Keywords    []string          `json:"keywords"`
	FileTypes   []string          `json:"file_types"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// SeverityLevel 严重级别
type SeverityLevel string

const (
	SeverityInfo     SeverityLevel = "info"
	SeverityLow      SeverityLevel = "low"
	SeverityMedium   SeverityLevel = "medium"
	SeverityHigh     SeverityLevel = "high"
	SeverityCritical SeverityLevel = "critical"
)

// DetectionResult 检测结果
type DetectionResult struct {
	ID          primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	URL         string              `json:"url"`
	Title       string              `json:"title"`
	StatusCode  int                 `json:"status_code"`
	ContentType string              `json:"content_type"`
	ContentSize int                 `json:"content_size"`
	Matches     []*SensitiveMatch   `json:"matches"`
	Summary     *DetectionSummary   `json:"summary"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time           `json:"created_at"`
	ScanTime    time.Duration       `json:"scan_time"`
}

// SensitiveMatch 敏感信息匹配
type SensitiveMatch struct {
	RuleID      string        `json:"rule_id"`
	RuleName    string        `json:"rule_name"`
	Category    string        `json:"category"`
	Severity    SeverityLevel `json:"severity"`
	Description string        `json:"description"`
	Match       string        `json:"match"`
	Context     string        `json:"context"`
	Position    MatchPosition `json:"position"`
	Confidence  float64       `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// MatchPosition 匹配位置
type MatchPosition struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	Start  int `json:"start"`
	End    int `json:"end"`
}

// DetectionSummary 检测摘要
type DetectionSummary struct {
	TotalMatches     int                          `json:"total_matches"`
	MatchesByCategory map[string]int              `json:"matches_by_category"`
	MatchesBySeverity map[SeverityLevel]int       `json:"matches_by_severity"`
	HighestSeverity  SeverityLevel               `json:"highest_severity"`
	Categories       []string                    `json:"categories"`
	RiskScore        float64                     `json:"risk_score"`
}

// TODO: 完善敏感信息检测器的实现
// Detector 敏感信息检测器
type Detector struct {
	// TODO: 添加具体的字段实现
	db        *mongo.Database
	ruleMap   map[string]*models.SensitiveRule
	mu        sync.RWMutex
	logger    SimpleLogger // TODO: 替换为实际的logger类型（已完善接口）
	rules     []*models.SensitiveRule // TODO: 完善规则管理
	whitelist []*models.SensitiveWhitelist    // TODO: 完善白名单管理（修复为正确的类型）
}

// NewSimpleLogger 创建简单日志器
func NewSimpleLogger() SimpleLogger {
	return &simpleLogger{}
}

type simpleLogger struct{}

func (l *simpleLogger) Error(msg string, args ...interface{}) {
	log.Printf("ERROR: "+msg, args...)
}

func (l *simpleLogger) Info(msg string, args ...interface{}) {
	log.Printf("INFO: "+msg, args...)
}

// NewDetector 创建检测器
func NewDetector(db *mongo.Database, rules []*DetectionRule) *Detector {
	return &Detector{
		db:        db,
		rules:     convertDetectionRules(rules),
		ruleMap:   make(map[string]*models.SensitiveRule),
		whitelist: []*models.SensitiveWhitelist{},
		logger:    NewSimpleLogger(),
	}
}

// convertDetectionRules 转换规则类型
func convertDetectionRules(rules []*DetectionRule) []*models.SensitiveRule {
	var result []*models.SensitiveRule
	for _, rule := range rules {
		// 将字符串ID转换为ObjectID，如果失败则创建新的ObjectID
		var objID primitive.ObjectID
		if parsed, err := primitive.ObjectIDFromHex(rule.ID); err == nil {
			objID = parsed
		} else {
			objID = primitive.NewObjectID()
		}
		
		result = append(result, &models.SensitiveRule{
			ID:          objID,
			Name:        rule.Name,
			Description: rule.Description,
			Pattern:     rule.Pattern,
			RiskLevel:   string(rule.Severity), // 使用RiskLevel字段映射Severity
			Enabled:     rule.Enabled,
		})
	}
	return result
}

// NewDetectionEngine 创建检测引擎
func NewDetectionEngine(db *mongo.Database) *DetectionEngine {
	engine := &DetectionEngine{
		db:      db,
		rules:   []*DetectionRule{},
		ruleMap: make(map[string]*DetectionRule),
		config: &DetectionConfig{
			MaxContentSize:    10 * 1024 * 1024, // 10MB
			Timeout:          30 * time.Second,
			UserAgent:        "Stellar-SensitiveDetector/1.0",
			FollowRedirects:  true,
			MaxRedirects:     5,
			EnabledCategories: []string{"credentials", "pii", "financial", "api_keys", "infrastructure"},
			Severity:         "low",
			OutputFormats:    []string{"json"},
			IncludeContext:   true,
			ContextLines:     2,
		},
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// 初始化默认规则
	engine.initDefaultRules()
	
	return engine
}

// initDefaultRules 初始化默认检测规则
func (e *DetectionEngine) initDefaultRules() {
	defaultRules := []*DetectionRule{
		// API 密钥类
		{
			ID:          "api_key_aws",
			Name:        "AWS Access Key",
			Description: "AWS 访问密钥",
			Category:    "api_keys",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `AKIA[0-9A-Z]{16}`,
			Keywords:    []string{"AKIA", "aws", "access", "key"},
		},
		{
			ID:          "api_key_github",
			Name:        "GitHub Token",
			Description: "GitHub 个人访问令牌",
			Category:    "api_keys",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `gh[pousr]_[A-Za-z0-9_]{36}`,
			Keywords:    []string{"github", "token", "ghp_", "gho_", "ghu_", "ghs_", "ghr_"},
		},
		{
			ID:          "api_key_slack",
			Name:        "Slack Token",
			Description: "Slack API 令牌",
			Category:    "api_keys",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `xox[baprs]-[0-9a-zA-Z\-]{10,48}`,
			Keywords:    []string{"slack", "token", "xoxb", "xoxa", "xoxp", "xoxr", "xoxs"},
		},
		
		// 凭据类
		{
			ID:          "credential_password",
			Name:        "Password in Code",
			Description: "代码中的密码",
			Category:    "credentials",
			Severity:    SeverityMedium,
			Enabled:     true,
			Pattern:     `(?i)(password|pwd|pass)\s*[=:]\s*[\"']([^\"']{8,})[\"']`,
			Keywords:    []string{"password", "pwd", "pass"},
		},
		{
			ID:          "credential_private_key",
			Name:        "Private Key",
			Description: "私钥文件",
			Category:    "credentials",
			Severity:    SeverityCritical,
			Enabled:     true,
			Pattern:     `-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----`,
			Keywords:    []string{"BEGIN", "PRIVATE", "KEY"},
		},
		{
			ID:          "credential_database",
			Name:        "Database Connection",
			Description: "数据库连接字符串",
			Category:    "credentials",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `(?i)(mongodb|mysql|postgres|redis|oracle)://[^:\s]+:[^@\s]+@[^\s]+`,
			Keywords:    []string{"mongodb://", "mysql://", "postgres://", "redis://", "oracle://"},
		},
		
		// 个人身份信息 (PII)
		{
			ID:          "pii_credit_card",
			Name:        "Credit Card Number",
			Description: "信用卡号码",
			Category:    "financial",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `(?:\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}|\d{13,19})`,
			Keywords:    []string{"card", "credit", "visa", "mastercard"},
		},
		{
			ID:          "pii_ssn",
			Name:        "Social Security Number",
			Description: "社会保障号码",
			Category:    "pii",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `\b\d{3}-\d{2}-\d{4}\b`,
			Keywords:    []string{"ssn", "social", "security"},
		},
		{
			ID:          "pii_email",
			Name:        "Email Address",
			Description: "电子邮件地址",
			Category:    "pii",
			Severity:    SeverityLow,
			Enabled:     true,
			Pattern:     `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
			Keywords:    []string{"@", "email", "mail"},
		},
		{
			ID:          "pii_phone",
			Name:        "Phone Number",
			Description: "电话号码",
			Category:    "pii",
			Severity:    SeverityLow,
			Enabled:     true,
			Pattern:     `(\+\d{1,3}\s?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}`,
			Keywords:    []string{"phone", "tel", "mobile"},
		},
		
		// 基础设施信息
		{
			ID:          "infra_ip_address",
			Name:        "IP Address",
			Description: "IP地址",
			Category:    "infrastructure",
			Severity:    SeverityLow,
			Enabled:     true,
			Pattern:     `\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`,
			Keywords:    []string{"ip", "address", "host"},
		},
		{
			ID:          "infra_jwt_token",
			Name:        "JWT Token",
			Description: "JSON Web Token",
			Category:    "credentials",
			Severity:    SeverityMedium,
			Enabled:     true,
			Pattern:     `eyJ[A-Za-z0-9_-]*\.eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*`,
			Keywords:    []string{"jwt", "token", "eyJ"},
		},
	}

	for _, rule := range defaultRules {
		if err := e.addRule(rule); err != nil {
			log.Printf("添加默认规则失败 %s: %v", rule.ID, err)
		}
	}
}

// addRule 添加规则
func (e *DetectionEngine) addRule(rule *DetectionRule) error {
	// 编译正则表达式
	if rule.Pattern != "" {
		regex, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return fmt.Errorf("编译正则表达式失败: %v", err)
		}
		rule.Regex = regex
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	e.rules = append(e.rules, rule)
	e.ruleMap[rule.ID] = rule

	return nil
}

// Detect 执行敏感信息检测
func (d *Detector) Detect(ctx context.Context, req models.SensitiveDetectionRequest) (*models.SensitiveDetectionResult, error) {
	// 创建检测结果
	result := &models.SensitiveDetectionResult{
		ID:          primitive.NewObjectID(),
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Targets:     req.Targets,
		Status:      models.SensitiveDetectionStatusRunning,
		StartTime:   time.Now(),
		EndTime:     time.Time{},
		Progress:    0,
		Config:      req.Config,
		Findings:    []*models.SensitiveFinding{},
		Summary:     models.SensitiveDetectionSummary{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		TotalCount:  0,
		FinishCount: 0,
	}

	// 启动检测协程
	go d.runDetection(ctx, req, result)

	return result, nil
}

// runDetection 运行检测任务
func (d *Detector) runDetection(ctx context.Context, req models.SensitiveDetectionRequest, result *models.SensitiveDetectionResult) {
	// 初始化计数器
	totalTargets := len(req.Targets)
	result.TotalCount = totalTargets
	result.FinishCount = 0

	// 创建工作通道
	targetChan := make(chan string, totalTargets)
	resultChan := make(chan *models.SensitiveFinding, totalTargets*5) // 每个目标可能有多个发现

	// 启动工作协程
	var wg sync.WaitGroup
	concurrency := req.Config.Concurrency
	if concurrency <= 0 {
		concurrency = 10 // 默认并发数
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range targetChan {
				// 检查上下文是否取消
				select {
				case <-ctx.Done():
					return
				default:
				}

				// 处理单个目标
				findings, err := d.detectTarget(ctx, target, req.Config)
				if err != nil {
					d.logger.Error("检测目标失败", err, "target", target)
					continue
				}

				// 发送发现结果
				for _, finding := range findings {
					resultChan <- finding
				}

				// 更新进度
				result.FinishCount++
				result.Progress = float64(result.FinishCount) / float64(totalTargets) * 100
				result.UpdatedAt = time.Now()
				// 这里应该有更新数据库中结果的代码
			}
		}()
	}

	// 发送目标到通道
	for _, target := range req.Targets {
		targetChan <- target
	}
	close(targetChan)

	// 处理结果的协程
	go func() {
		findings := []*models.SensitiveFinding{}
		riskLevelCount := map[string]int{}
		categoryCount := map[string]int{}

		for finding := range resultChan {
			findings = append(findings, finding)
			riskLevelCount[finding.RiskLevel]++
			categoryCount[finding.Category]++
		}

		// 更新结果
		result.Findings = findings
		result.Summary.TotalFindings = len(findings)
		result.Summary.RiskLevelCount = riskLevelCount
		result.Summary.CategoryCount = categoryCount
		result.Status = models.SensitiveDetectionStatusCompleted
		result.EndTime = time.Now()
		result.UpdatedAt = time.Now()
		// 这里应该有更新数据库中结果的代码
	}()

	// 等待所有工作协程完成
	wg.Wait()
	close(resultChan)
}

// detectTarget 检测单个目标
func (d *Detector) detectTarget(ctx context.Context, target string, config models.SensitiveDetectionConfig) ([]*models.SensitiveFinding, error) {
	// 检查目标类型并处理
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		// 处理URL目标
		return d.detectURLTarget(ctx, target, config)
	} else if strings.Contains(target, "*") || strings.Contains(target, "?") || strings.Contains(target, "**") {
		// 处理文件模式目标
		return d.detectFilesByPattern(ctx, target, config)
	} else if d.isDirectory(target) {
		// 处理目录目标
		return d.detectDirectory(ctx, target, config, config.RecursiveSearch)
	} else {
		// 处理单个文件目标
		return d.detectSingleFile(ctx, target, config)
	}
}

// detectURLTarget 检测URL目标
func (d *Detector) detectURLTarget(ctx context.Context, target string, config models.SensitiveDetectionConfig) ([]*models.SensitiveFinding, error) {
	// 获取内容
	content, err := d.fetchWebContent(ctx, target, config)
	if err != nil {
		return nil, err
	}

	// 应用规则检测
	return d.applyRules(ctx, target, content, config)
}

// detectSingleFile 检测单个文件
func (d *Detector) detectSingleFile(ctx context.Context, target string, config models.SensitiveDetectionConfig) ([]*models.SensitiveFinding, error) {
	// 获取文件内容
	content, err := d.fetchFileContent(ctx, target, config)
	if err != nil {
		return nil, err
	}

	// 应用规则检测
	return d.applyRules(ctx, target, content, config)
}

// applyRules 应用规则检测内容
func (d *Detector) applyRules(ctx context.Context, target string, content string, config models.SensitiveDetectionConfig) ([]*models.SensitiveFinding, error) {
	var findings []*models.SensitiveFinding

	// 应用每个规则
	for _, rule := range d.rules {
		// 跳过禁用的规则
		if !rule.Enabled {
			continue
		}

		// 编译正则表达式
		regex, err := regexp.Compile(rule.Pattern)
		if err != nil {
			d.logger.Error("规则正则表达式无效", err, "rule", rule.Name, "pattern", rule.Pattern)
			continue
		}

		// 查找匹配
		matches := regex.FindAllStringSubmatch(content, -1)
		if len(matches) == 0 {
			continue
		}

		// 处理每个匹配
		for _, match := range matches {
			if len(match) == 0 {
				continue
			}

			// 提取匹配内容
			matchedText := match[0]

			// 检查白名单
			if d.isWhitelisted(target, matchedText) {
				continue
			}

			// 检查误报模式
			if d.isFalsePositive(rule, matchedText) {
				continue
			}

			// 提取上下文和行号
			context, lineNumber := d.extractContextWithLineNumber(content, matchedText, config.ContextLines)

			// 确定目标类型
			targetType := "url"
			if !strings.HasPrefix(target, "http") {
				targetType = "file"
			}

			// 获取文件信息
			var fileSize int64
			if targetType == "file" {
				if info, err := os.Stat(target); err == nil {
					fileSize = info.Size()
				}
			}

			// 创建发现
			finding := &models.SensitiveFinding{
				ID:          primitive.NewObjectID(),
				Target:      target,
				TargetType:  targetType,
				Rule:        rule.ID,
				RuleName:    rule.Name,
				Category:    rule.Category,
				RiskLevel:   rule.RiskLevel,
				Pattern:     rule.Pattern,
				MatchedText: matchedText,
				Context:     context,
				LineNumber:  lineNumber,
				FilePath:    target, // 对于文件，存储完整路径
				FileSize:    fileSize,
				CreatedAt:   time.Now(),
			}

			findings = append(findings, finding)
		}
	}

	return findings, nil
}

// isDirectory 检查路径是否为目录
func (d *Detector) isDirectory(path string) bool {
	// 移除 file:// 前缀
	if strings.HasPrefix(path, "file://") {
		path = strings.TrimPrefix(path, "file://")
	}

	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// fetchContent 获取目标内容
func (d *Detector) fetchContent(ctx context.Context, target string, config models.SensitiveDetectionConfig) (string, error) {
	// 检查目标类型
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		// 处理网页内容
		return d.fetchWebContent(ctx, target, config)
	} else if strings.HasPrefix(target, "file://") || !strings.Contains(target, "://") {
		// 处理文件内容
		return d.fetchFileContent(ctx, target, config)
	} else {
		return "", fmt.Errorf("不支持的目标类型: %s", target)
	}
}

// fetchWebContent 获取网页内容
func (d *Detector) fetchWebContent(ctx context.Context, url string, config models.SensitiveDetectionConfig) (string, error) {
	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", "StellarServer/1.0")

	// 设置超时
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// isWhitelisted 检查是否在白名单中
func (d *Detector) isWhitelisted(target string, text string) bool {
	for _, item := range d.whitelist {
		switch item.Type {
		case "target":
			if item.Value == target {
				return true
			}
		case "pattern":
			regex, err := regexp.Compile(item.Value)
			if err != nil {
				d.logger.Error("白名单正则表达式无效", err, "value", item.Value)
				continue
			}
			if regex.MatchString(text) {
				return true
			}
		}
	}
	return false
}

// isFalsePositive 检查是否为误报
func (d *Detector) isFalsePositive(rule *models.SensitiveRule, text string) bool {
	for _, pattern := range rule.FalsePositivePatterns {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			d.logger.Error("误报模式正则表达式无效", err, "pattern", pattern)
			continue
		}
		if regex.MatchString(text) {
			return true
		}
	}
	return false
}

// extractContext 提取上下文
func (d *Detector) extractContext(content, matchedText string, contextLines int) string {
	// 查找匹配文本在内容中的位置
	index := strings.Index(content, matchedText)
	if index == -1 {
		return ""
	}

	// 计算上下文的起始和结束位置
	start := index
	end := index + len(matchedText)

	// 向前查找contextLines行
	for i := 0; i < contextLines; i++ {
		newStart := strings.LastIndex(content[:start], "\n")
		if newStart == -1 {
			break
		}
		start = newStart + 1
	}

	// 向后查找contextLines行
	for i := 0; i < contextLines; i++ {
		newEnd := strings.Index(content[end:], "\n")
		if newEnd == -1 {
			break
		}
		end += newEnd + 1
	}

	// 提取上下文
	context := content[start:end]

	// 如果上下文太长，进行截断
	maxContextLength := 1000
	if len(context) > maxContextLength {
		// 确保匹配文本在截断后的上下文中
		matchStart := index - start

		if matchStart+maxContextLength/2 < len(context) {
			// 将匹配文本放在上下文中间
			contextStart := matchStart - maxContextLength/2
			if contextStart < 0 {
				contextStart = 0
			}
			contextEnd := contextStart + maxContextLength
			if contextEnd > len(context) {
				contextEnd = len(context)
			}
			context = context[contextStart:contextEnd]
		} else {
			// 取最后maxContextLength个字符
			context = context[len(context)-maxContextLength:]
		}
	}

	return context
}

// extractContextWithLineNumber 提取上下文和行号
func (d *Detector) extractContextWithLineNumber(content, matchedText string, contextLines int) (string, int) {
	// 分割内容为行
	lines := strings.Split(content, "\n")
	
	// 查找匹配文本所在的行
	var matchLine = -1
	for i, line := range lines {
		if strings.Contains(line, matchedText) {
			matchLine = i
			break
		}
	}
	
	if matchLine == -1 {
		// 如果没有找到行，使用原来的方法
		context := d.extractContext(content, matchedText, contextLines)
		return context, 0
	}
	
	// 计算上下文的起始和结束行
	startLine := matchLine - contextLines
	endLine := matchLine + contextLines + 1
	
	if startLine < 0 {
		startLine = 0
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	
	// 提取上下文
	contextLines_slice := lines[startLine:endLine]
	context := strings.Join(contextLines_slice, "\n")
	
	// 返回上下文和行号（从1开始计数）
	return context, matchLine + 1
}

// fetchFileContent 获取文件内容
func (d *Detector) fetchFileContent(ctx context.Context, target string, config models.SensitiveDetectionConfig) (string, error) {
	// 移除 file:// 前缀（如果存在）
	filePath := target
	if strings.HasPrefix(target, "file://") {
		filePath = strings.TrimPrefix(target, "file://")
	}

	// 检查文件扩展名，决定是否需要处理
	if !d.isTextFile(filePath) {
		return "", fmt.Errorf("不支持的文件类型: %s", filePath)
	}

	// 检查文件大小，避免处理过大的文件
	info, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("无法获取文件信息: %v", err)
	}

	// 限制文件大小为10MB
	maxFileSize := int64(10 * 1024 * 1024) // 10MB
	if info.Size() > maxFileSize {
		return "", fmt.Errorf("文件过大，超过限制 (%d bytes)", maxFileSize)
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	// 尝试检测文件编码并转换为UTF-8
	return d.convertToUTF8(content)
}

// isTextFile 检查是否为文本文件
func (d *Detector) isTextFile(filePath string) bool {
	// 支持的文本文件扩展名
	textExtensions := map[string]bool{
		".txt":        true,
		".log":        true,
		".json":       true,
		".xml":        true,
		".html":       true,
		".htm":        true,
		".css":        true,
		".js":         true,
		".ts":         true,
		".java":       true,
		".py":         true,
		".go":         true,
		".c":          true,
		".cpp":        true,
		".h":          true,
		".hpp":        true,
		".php":        true,
		".rb":         true,
		".sh":         true,
		".bat":        true,
		".ps1":        true,
		".yml":        true,
		".yaml":       true,
		".ini":        true,
		".conf":       true,
		".config":     true,
		".properties": true,
		".env":        true,
		".sql":        true,
		".md":         true,
		".markdown":   true,
		".rst":        true,
		".csv":        true,
		".tsv":        true,
		".dockerfile": true,
		".gitignore":  true,
		".gitattributes": true,
		"":            true, // 无扩展名的文件也尝试处理
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(filePath))
	return textExtensions[ext]
}

// convertToUTF8 转换内容为UTF-8编码
func (d *Detector) convertToUTF8(content []byte) (string, error) {
	// 检测内容是否已经是有效的UTF-8
	if utf8.Valid(content) {
		return string(content), nil
	}

	// 尝试常见的编码转换
	encodings := []encoding.Encoding{
		charmap.Windows1252,  // Windows-1252
		charmap.ISO8859_1,    // ISO-8859-1
		simplifiedchinese.GBK, // GBK
		traditionalchinese.Big5, // Big5
	}

	for _, enc := range encodings {
		decoder := enc.NewDecoder()
		decoded, err := decoder.Bytes(content)
		if err == nil && utf8.Valid(decoded) {
			return string(decoded), nil
		}
	}

	// 如果所有编码都失败，尝试移除无效字符
	return strings.ToValidUTF8(string(content), ""), nil
}

// detectFilesByPattern 按文件模式检测
func (d *Detector) detectFilesByPattern(ctx context.Context, pattern string, config models.SensitiveDetectionConfig) ([]*models.SensitiveFinding, error) {
	var findings []*models.SensitiveFinding

	// 使用 filepath.Glob 查找匹配的文件
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("文件模式匹配失败: %v", err)
	}

	// 检测每个文件
	for _, file := range files {
		select {
		case <-ctx.Done():
			return findings, ctx.Err()
		default:
		}

		fileFindings, err := d.detectTarget(ctx, file, config)
		if err != nil {
			d.logger.Error("检测文件失败", err, "file", file)
			continue
		}

		findings = append(findings, fileFindings...)
	}

	return findings, nil
}

// detectDirectory 检测目录中的文件
func (d *Detector) detectDirectory(ctx context.Context, dirPath string, config models.SensitiveDetectionConfig, recursive bool) ([]*models.SensitiveFinding, error) {
	var findings []*models.SensitiveFinding

	// 遍历目录
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查上下文是否取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 跳过目录
		if info.IsDir() {
			// 如果不是递归模式，跳过子目录
			if !recursive && path != dirPath {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查文件是否为文本文件
		if !d.isTextFile(path) {
			return nil
		}

		// 检测文件
		fileFindings, err := d.detectTarget(ctx, path, config)
		if err != nil {
			d.logger.Error("检测文件失败", err, "file", path)
			return nil // 继续处理其他文件
		}

		findings = append(findings, fileFindings...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历目录失败: %v", err)
	}

	return findings, nil
}

// GetResults 获取检测结果
func (d *Detector) GetResults(projectID primitive.ObjectID, limit int) ([]*models.SensitiveDetectionResult, error) {
	collection := d.db.Collection("sensitive_detection_results")
	
	filter := bson.M{}
	if !projectID.IsZero() {
		filter["project_id"] = projectID
	}
	
	findOptions := &options.FindOptions{}
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}
	findOptions.SetSort(bson.D{{"created_at", -1}})
	
	cursor, err := collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var results []*models.SensitiveDetectionResult
	for cursor.Next(context.Background()) {
		var result models.SensitiveDetectionResult
		if err := cursor.Decode(&result); err == nil {
			results = append(results, &result)
		}
	}
	
	return results, nil
}

// GetResult 获取单个检测结果
func (d *Detector) GetResult(resultID primitive.ObjectID) (*models.SensitiveDetectionResult, error) {
	collection := d.db.Collection("sensitive_detection_results")
	
	var result models.SensitiveDetectionResult
	err := collection.FindOne(context.Background(), bson.M{"_id": resultID}).Decode(&result)
	if err != nil {
		return nil, err
	}
	
	return &result, nil
}

// UpdateResult 更新检测结果
func (d *Detector) UpdateResult(resultID primitive.ObjectID, updates map[string]interface{}) error {
	collection := d.db.Collection("sensitive_detection_results")
	
	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": resultID},
		bson.M{"$set": updates},
	)
	
	return err
}

// DeleteResult 删除检测结果
func (d *Detector) DeleteResult(resultID primitive.ObjectID) error {
	collection := d.db.Collection("sensitive_detection_results")
	
	_, err := collection.DeleteOne(context.Background(), bson.M{"_id": resultID})
	return err
}

// GetResultsByTimeRange 根据时间范围获取结果
func (d *Detector) GetResultsByTimeRange(startTime, endTime time.Time) ([]*models.SensitiveDetectionResult, error) {
	collection := d.db.Collection("sensitive_detection_results")
	
	filter := bson.M{
		"created_at": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}
	
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var results []*models.SensitiveDetectionResult
	for cursor.Next(context.Background()) {
		var result models.SensitiveDetectionResult
		if err := cursor.Decode(&result); err == nil {
			results = append(results, &result)
		}
	}
	
	return results, nil
}

// GetStatistics 获取统计信息
func (d *Detector) GetStatistics() (map[string]interface{}, error) {
	collection := d.db.Collection("sensitive_detection_results")
	
	totalCount, _ := collection.CountDocuments(context.Background(), bson.M{})
	
	stats := map[string]interface{}{
		"total_results": totalCount,
		"active_rules":  len(d.rules),
	}
	
	return stats, nil
}
