package sensitive

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// RuleManager 规则管理器
type RuleManager struct {
	db           *mongo.Database
	engine       *DetectionEngine
	builtinRules []*DetectionRule
	customRules  []*DetectionRule
	rulesets     map[string]*RuleSet
}

// RuleSet 规则集
type RuleSet struct {
	ID          string           `json:"id" bson:"_id,omitempty"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Category    string           `json:"category"`
	Language    string           `json:"language"`
	Rules       []*DetectionRule `json:"rules"`
	Enabled     bool             `json:"enabled"`
	Version     string           `json:"version"`
	Author      string           `json:"author"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// RuleTemplate 规则模板
type RuleTemplate struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Language    string            `json:"language"`
	Template    string            `json:"template"`
	Variables   map[string]string `json:"variables"`
	Examples    []string          `json:"examples"`
}

// RuleValidationResult 规则验证结果
type RuleValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
	Suggestions []string `json:"suggestions"`
}

// RuleStatistics 规则统计
type RuleStatistics struct {
	TotalRules       int                       `json:"total_rules"`
	EnabledRules     int                       `json:"enabled_rules"`
	DisabledRules    int                       `json:"disabled_rules"`
	RulesByCategory  map[string]int            `json:"rules_by_category"`
	RulesBySeverity  map[SeverityLevel]int     `json:"rules_by_severity"`
	RulesByLanguage  map[string]int            `json:"rules_by_language"`
	CustomRules      int                       `json:"custom_rules"`
	BuiltinRules     int                       `json:"builtin_rules"`
	LastUpdated      time.Time                 `json:"last_updated"`
}

// NewRuleManager 创建规则管理器
func NewRuleManager(db *mongo.Database, engine *DetectionEngine) *RuleManager {
	manager := &RuleManager{
		db:       db,
		engine:   engine,
		rulesets: make(map[string]*RuleSet),
	}
	
	// 初始化内置规则
	manager.initBuiltinRules()
	
	// 加载自定义规则
	manager.loadCustomRules()
	
	return manager
}

// initBuiltinRules 初始化内置规则
func (rm *RuleManager) initBuiltinRules() {
	// 加载内置规则集
	builtinRulesets := []*RuleSet{
		rm.createCredentialsRuleSet(),
		rm.createAPIKeysRuleSet(),
		rm.createPIIRuleSet(),
		rm.createFinancialRuleSet(),
		rm.createInfrastructureRuleSet(),
		rm.createSecurityRuleSet(),
		rm.createDatabaseRuleSet(),
		rm.createCloudRuleSet(),
	}
	
	for _, ruleset := range builtinRulesets {
		rm.rulesets[ruleset.ID] = ruleset
		rm.builtinRules = append(rm.builtinRules, ruleset.Rules...)
	}
}

// createCredentialsRuleSet 创建凭据规则集
func (rm *RuleManager) createCredentialsRuleSet() *RuleSet {
	rules := []*DetectionRule{
		{
			ID:          "cred_password_assignment",
			Name:        "密码赋值",
			Description: "检测代码中的密码赋值",
			Category:    "credentials",
			Severity:    SeverityMedium,
			Enabled:     true,
			Pattern:     `(?i)(password|pwd|pass|secret|key)\s*[=:]\s*["\']([^"\']{6,})["\']`,
			Keywords:    []string{"password", "pwd", "pass", "secret", "key"},
			FileTypes:   []string{".py", ".js", ".go", ".java", ".php", ".rb", ".sh"},
		},
		{
			ID:          "cred_private_key",
			Name:        "私钥文件",
			Description: "检测PEM格式的私钥",
			Category:    "credentials",
			Severity:    SeverityCritical,
			Enabled:     true,
			Pattern:     `-----BEGIN\s+(RSA\s+|EC\s+|DSA\s+|OPENSSH\s+)?PRIVATE\s+KEY-----`,
			Keywords:    []string{"BEGIN", "PRIVATE", "KEY"},
			FileTypes:   []string{".pem", ".key", ".txt", ".log"},
		},
		{
			ID:          "cred_certificate",
			Name:        "证书文件",
			Description: "检测PEM格式的证书",
			Category:    "credentials",
			Severity:    SeverityLow,
			Enabled:     true,
			Pattern:     `-----BEGIN\s+CERTIFICATE-----`,
			Keywords:    []string{"BEGIN", "CERTIFICATE"},
			FileTypes:   []string{".pem", ".crt", ".cer", ".txt"},
		},
	}
	
	return &RuleSet{
		ID:          "builtin_credentials",
		Name:        "凭据检测规则",
		Description: "检测各种类型的凭据信息",
		Category:    "credentials",
		Language:    "regex",
		Rules:       rules,
		Enabled:     true,
		Version:     "1.0.0",
		Author:      "Stellar Team",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// createAPIKeysRuleSet 创建API密钥规则集
func (rm *RuleManager) createAPIKeysRuleSet() *RuleSet {
	rules := []*DetectionRule{
		{
			ID:          "api_aws_access_key",
			Name:        "AWS访问密钥",
			Description: "检测AWS访问密钥",
			Category:    "api_keys",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `AKIA[0-9A-Z]{16}`,
			Keywords:    []string{"AKIA", "aws", "access", "key"},
		},
		{
			ID:          "api_aws_secret_key",
			Name:        "AWS密钥",
			Description: "检测AWS密钥",
			Category:    "api_keys",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `(?i)aws.{0,20}['\"][0-9a-zA-Z/+]{40}['\"]`,
			Keywords:    []string{"aws", "secret", "key"},
		},
		{
			ID:          "api_github_token",
			Name:        "GitHub令牌",
			Description: "检测GitHub个人访问令牌",
			Category:    "api_keys",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `gh[pousr]_[A-Za-z0-9_]{36}`,
			Keywords:    []string{"github", "token", "ghp_", "gho_", "ghu_", "ghs_", "ghr_"},
		},
		{
			ID:          "api_slack_token",
			Name:        "Slack令牌",
			Description: "检测Slack API令牌",
			Category:    "api_keys",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `xox[baprs]-[0-9a-zA-Z\-]{10,48}`,
			Keywords:    []string{"slack", "token", "xoxb", "xoxa", "xoxp", "xoxr", "xoxs"},
		},
		{
			ID:          "api_google_api_key",
			Name:        "Google API密钥",
			Description: "检测Google API密钥",
			Category:    "api_keys",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `AIza[0-9A-Za-z\-_]{35}`,
			Keywords:    []string{"google", "api", "key", "AIza"},
		},
	}
	
	return &RuleSet{
		ID:          "builtin_api_keys",
		Name:        "API密钥检测规则",
		Description: "检测各种云服务和平台的API密钥",
		Category:    "api_keys",
		Language:    "regex",
		Rules:       rules,
		Enabled:     true,
		Version:     "1.0.0",
		Author:      "Stellar Team",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// createPIIRuleSet 创建个人信息规则集
func (rm *RuleManager) createPIIRuleSet() *RuleSet {
	rules := []*DetectionRule{
		{
			ID:          "pii_email_address",
			Name:        "电子邮件地址",
			Description: "检测电子邮件地址",
			Category:    "pii",
			Severity:    SeverityLow,
			Enabled:     true,
			Pattern:     `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
			Keywords:    []string{"@", "email", "mail"},
		},
		{
			ID:          "pii_phone_number",
			Name:        "电话号码",
			Description: "检测电话号码",
			Category:    "pii",
			Severity:    SeverityLow,
			Enabled:     true,
			Pattern:     `(\+\d{1,3}\s?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}`,
			Keywords:    []string{"phone", "tel", "mobile", "cell"},
		},
		{
			ID:          "pii_ssn",
			Name:        "社会保障号码",
			Description: "检测美国社会保障号码",
			Category:    "pii",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `\b\d{3}-\d{2}-\d{4}\b`,
			Keywords:    []string{"ssn", "social", "security"},
		},
		{
			ID:          "pii_chinese_id",
			Name:        "中国身份证号",
			Description: "检测中国居民身份证号码",
			Category:    "pii",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `\b[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(1[0-2]))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]\b`,
			Keywords:    []string{"身份证", "id_card", "identity"},
		},
	}
	
	return &RuleSet{
		ID:          "builtin_pii",
		Name:        "个人信息检测规则",
		Description: "检测各种个人身份信息",
		Category:    "pii",
		Language:    "regex",
		Rules:       rules,
		Enabled:     true,
		Version:     "1.0.0",
		Author:      "Stellar Team",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// createFinancialRuleSet 创建金融信息规则集
func (rm *RuleManager) createFinancialRuleSet() *RuleSet {
	rules := []*DetectionRule{
		{
			ID:          "fin_credit_card",
			Name:        "信用卡号",
			Description: "检测信用卡号码",
			Category:    "financial",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `(?:\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}|\d{13,19})`,
			Keywords:    []string{"card", "credit", "visa", "mastercard", "amex"},
		},
		{
			ID:          "fin_bank_account",
			Name:        "银行账号",
			Description: "检测银行账号",
			Category:    "financial",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `(?i)(account|bank).{0,20}\d{8,20}`,
			Keywords:    []string{"account", "bank", "routing"},
		},
		{
			ID:          "fin_iban",
			Name:        "国际银行账号",
			Description: "检测IBAN银行账号",
			Category:    "financial",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `[A-Z]{2}\d{2}[A-Z0-9]{4}\d{7}([A-Z0-9]?){0,16}`,
			Keywords:    []string{"iban", "bank", "account"},
		},
	}
	
	return &RuleSet{
		ID:          "builtin_financial",
		Name:        "金融信息检测规则",
		Description: "检测金融相关的敏感信息",
		Category:    "financial",
		Language:    "regex",
		Rules:       rules,
		Enabled:     true,
		Version:     "1.0.0",
		Author:      "Stellar Team",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// createInfrastructureRuleSet 创建基础设施规则集
func (rm *RuleManager) createInfrastructureRuleSet() *RuleSet {
	rules := []*DetectionRule{
		{
			ID:          "infra_ip_address",
			Name:        "IP地址",
			Description: "检测IPv4地址",
			Category:    "infrastructure",
			Severity:    SeverityLow,
			Enabled:     true,
			Pattern:     `\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`,
			Keywords:    []string{"ip", "address", "host"},
		},
		{
			ID:          "infra_ipv6_address",
			Name:        "IPv6地址",
			Description: "检测IPv6地址",
			Category:    "infrastructure",
			Severity:    SeverityLow,
			Enabled:     true,
			Pattern:     `([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}`,
			Keywords:    []string{"ipv6", "address"},
		},
		{
			ID:          "infra_domain",
			Name:        "域名",
			Description: "检测域名",
			Category:    "infrastructure",
			Severity:    SeverityInfo,
			Enabled:     true,
			Pattern:     `[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*\.[a-zA-Z]{2,}`,
			Keywords:    []string{"domain", "host", "url"},
		},
	}
	
	return &RuleSet{
		ID:          "builtin_infrastructure",
		Name:        "基础设施信息检测规则",
		Description: "检测网络和基础设施相关信息",
		Category:    "infrastructure",
		Language:    "regex",
		Rules:       rules,
		Enabled:     true,
		Version:     "1.0.0",
		Author:      "Stellar Team",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// createSecurityRuleSet 创建安全相关规则集
func (rm *RuleManager) createSecurityRuleSet() *RuleSet {
	rules := []*DetectionRule{
		{
			ID:          "sec_jwt_token",
			Name:        "JWT令牌",
			Description: "检测JSON Web Token",
			Category:    "credentials",
			Severity:    SeverityMedium,
			Enabled:     true,
			Pattern:     `eyJ[A-Za-z0-9_-]*\.eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*`,
			Keywords:    []string{"jwt", "token", "eyJ"},
		},
		{
			ID:          "sec_bearer_token",
			Name:        "Bearer令牌",
			Description: "检测Bearer授权令牌",
			Category:    "credentials",
			Severity:    SeverityMedium,
			Enabled:     true,
			Pattern:     `(?i)bearer\s+[a-zA-Z0-9\-._~+/]+=*`,
			Keywords:    []string{"bearer", "authorization", "token"},
		},
		{
			ID:          "sec_api_key_header",
			Name:        "API密钥头",
			Description: "检测HTTP头中的API密钥",
			Category:    "api_keys",
			Severity:    SeverityMedium,
			Enabled:     true,
			Pattern:     `(?i)(x-api-key|api-key|apikey)\s*:\s*[a-zA-Z0-9\-._~+/]{16,}`,
			Keywords:    []string{"api-key", "x-api-key", "apikey"},
		},
	}
	
	return &RuleSet{
		ID:          "builtin_security",
		Name:        "安全令牌检测规则",
		Description: "检测各种安全令牌和认证信息",
		Category:    "credentials",
		Language:    "regex",
		Rules:       rules,
		Enabled:     true,
		Version:     "1.0.0",
		Author:      "Stellar Team",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// createDatabaseRuleSet 创建数据库规则集
func (rm *RuleManager) createDatabaseRuleSet() *RuleSet {
	rules := []*DetectionRule{
		{
			ID:          "db_connection_string",
			Name:        "数据库连接字符串",
			Description: "检测数据库连接字符串",
			Category:    "credentials",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `(?i)(mongodb|mysql|postgres|redis|oracle|mssql)://[^:\s]+:[^@\s]+@[^\s]+`,
			Keywords:    []string{"mongodb://", "mysql://", "postgres://", "redis://", "oracle://"},
		},
		{
			ID:          "db_username_password",
			Name:        "数据库用户名密码",
			Description: "检测数据库用户名和密码配置",
			Category:    "credentials",
			Severity:    SeverityMedium,
			Enabled:     true,
			Pattern:     `(?i)(db_user|db_pass|database_user|database_password)\s*[=:]\s*["\']([^"\']{3,})["\']`,
			Keywords:    []string{"db_user", "db_pass", "database_user", "database_password"},
		},
	}
	
	return &RuleSet{
		ID:          "builtin_database",
		Name:        "数据库信息检测规则",
		Description: "检测数据库连接和认证信息",
		Category:    "credentials",
		Language:    "regex",
		Rules:       rules,
		Enabled:     true,
		Version:     "1.0.0",
		Author:      "Stellar Team",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// createCloudRuleSet 创建云服务规则集
func (rm *RuleManager) createCloudRuleSet() *RuleSet {
	rules := []*DetectionRule{
		{
			ID:          "cloud_azure_key",
			Name:        "Azure密钥",
			Description: "检测Azure服务密钥",
			Category:    "api_keys",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `[a-zA-Z0-9+/]{86}==`,
			Keywords:    []string{"azure", "subscription", "key"},
		},
		{
			ID:          "cloud_gcp_key",
			Name:        "Google Cloud密钥",
			Description: "检测Google Cloud Platform密钥",
			Category:    "api_keys",
			Severity:    SeverityHigh,
			Enabled:     true,
			Pattern:     `"type":\s*"service_account"`,
			Keywords:    []string{"service_account", "private_key", "gcp"},
		},
		{
			ID:          "cloud_docker_auth",
			Name:        "Docker认证",
			Description: "检测Docker认证配置",
			Category:    "credentials",
			Severity:    SeverityMedium,
			Enabled:     true,
			Pattern:     `"auths":\s*{[^}]*"auth":\s*"[^"]+"}`,
			Keywords:    []string{"docker", "auths", "auth"},
		},
	}
	
	return &RuleSet{
		ID:          "builtin_cloud",
		Name:        "云服务检测规则",
		Description: "检测各种云服务的认证信息",
		Category:    "api_keys",
		Language:    "regex",
		Rules:       rules,
		Enabled:     true,
		Version:     "1.0.0",
		Author:      "Stellar Team",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// loadCustomRules 加载自定义规则
func (rm *RuleManager) loadCustomRules() {
	ctx := context.Background()
	collection := rm.db.Collection("sensitive_rulesets")
	
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return
	}
	defer cursor.Close(ctx)
	
	for cursor.Next(ctx) {
		var ruleset RuleSet
		if err := cursor.Decode(&ruleset); err != nil {
			continue
		}
		
		rm.rulesets[ruleset.ID] = &ruleset
		rm.customRules = append(rm.customRules, ruleset.Rules...)
	}
}

// CreateRuleSet 创建规则集
func (rm *RuleManager) CreateRuleSet(ruleset *RuleSet) error {
	// 验证规则集
	if err := rm.validateRuleSet(ruleset); err != nil {
		return fmt.Errorf("规则集验证失败: %v", err)
	}
	
	// 设置时间戳
	now := time.Now()
	ruleset.CreatedAt = now
	ruleset.UpdatedAt = now
	
	// 保存到数据库
	ctx := context.Background()
	collection := rm.db.Collection("sensitive_rulesets")
	
	_, err := collection.InsertOne(ctx, ruleset)
	if err != nil {
		return fmt.Errorf("保存规则集失败: %v", err)
	}
	
	// 添加到内存
	rm.rulesets[ruleset.ID] = ruleset
	rm.customRules = append(rm.customRules, ruleset.Rules...)
	
	// 更新检测引擎
	rm.updateEngine()
	
	return nil
}

// UpdateRuleSet 更新规则集
func (rm *RuleManager) UpdateRuleSet(rulesetID string, updates map[string]interface{}) error {
	ruleset, exists := rm.rulesets[rulesetID]
	if !exists {
		return fmt.Errorf("规则集不存在: %s", rulesetID)
	}
	
	// 应用更新
	if name, ok := updates["name"].(string); ok {
		ruleset.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		ruleset.Description = description
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		ruleset.Enabled = enabled
	}
	if rules, ok := updates["rules"].([]*DetectionRule); ok {
		ruleset.Rules = rules
	}
	
	ruleset.UpdatedAt = time.Now()
	
	// 验证更新后的规则集
	if err := rm.validateRuleSet(ruleset); err != nil {
		return fmt.Errorf("规则集验证失败: %v", err)
	}
	
	// 更新数据库
	ctx := context.Background()
	collection := rm.db.Collection("sensitive_rulesets")
	
	filter := bson.M{"_id": rulesetID}
	update := bson.M{"$set": ruleset}
	
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("更新规则集失败: %v", err)
	}
	
	// 重新加载规则
	rm.reloadRules()
	
	return nil
}

// DeleteRuleSet 删除规则集
func (rm *RuleManager) DeleteRuleSet(rulesetID string) error {
	// 检查是否为内置规则集
	if strings.HasPrefix(rulesetID, "builtin_") {
		return fmt.Errorf("不能删除内置规则集")
	}
	
	// 从数据库删除
	ctx := context.Background()
	collection := rm.db.Collection("sensitive_rulesets")
	
	_, err := collection.DeleteOne(ctx, bson.M{"_id": rulesetID})
	if err != nil {
		return fmt.Errorf("删除规则集失败: %v", err)
	}
	
	// 从内存删除
	delete(rm.rulesets, rulesetID)
	
	// 重新加载规则
	rm.reloadRules()
	
	return nil
}

// GetRuleSet 获取规则集
func (rm *RuleManager) GetRuleSet(rulesetID string) (*RuleSet, error) {
	ruleset, exists := rm.rulesets[rulesetID]
	if !exists {
		return nil, fmt.Errorf("规则集不存在: %s", rulesetID)
	}
	
	return ruleset, nil
}

// ListRuleSets 列出所有规则集
func (rm *RuleManager) ListRuleSets() []*RuleSet {
	var rulesets []*RuleSet
	for _, ruleset := range rm.rulesets {
		rulesets = append(rulesets, ruleset)
	}
	
	// 按名称排序
	sort.Slice(rulesets, func(i, j int) bool {
		return rulesets[i].Name < rulesets[j].Name
	})
	
	return rulesets
}

// ValidateRule 验证单个规则
func (rm *RuleManager) ValidateRule(rule *DetectionRule) *RuleValidationResult {
	result := &RuleValidationResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		Suggestions: []string{},
	}
	
	// 基本字段验证
	if rule.ID == "" {
		result.Errors = append(result.Errors, "规则ID不能为空")
		result.Valid = false
	}
	
	if rule.Name == "" {
		result.Errors = append(result.Errors, "规则名称不能为空")
		result.Valid = false
	}
	
	if rule.Category == "" {
		result.Warnings = append(result.Warnings, "建议设置规则类别")
	}
	
	// 正则表达式验证
	if rule.Pattern != "" {
		_, err := regexp.Compile(rule.Pattern)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("正则表达式无效: %v", err))
			result.Valid = false
		}
	}
	
	// 关键词验证
	if len(rule.Keywords) == 0 && rule.Pattern == "" {
		result.Errors = append(result.Errors, "必须设置正则表达式或关键词")
		result.Valid = false
	}
	
	// 性能建议
	if rule.Pattern != "" && len(rule.Pattern) > 200 {
		result.Warnings = append(result.Warnings, "正则表达式过长，可能影响性能")
	}
	
	if len(rule.Keywords) > 20 {
		result.Warnings = append(result.Warnings, "关键词过多，可能影响性能")
	}
	
	return result
}

// validateRuleSet 验证规则集
func (rm *RuleManager) validateRuleSet(ruleset *RuleSet) error {
	if ruleset.ID == "" {
		return fmt.Errorf("规则集ID不能为空")
	}
	
	if ruleset.Name == "" {
		return fmt.Errorf("规则集名称不能为空")
	}
	
	if len(ruleset.Rules) == 0 {
		return fmt.Errorf("规则集不能为空")
	}
	
	// 验证每个规则
	for i, rule := range ruleset.Rules {
		result := rm.ValidateRule(rule)
		if !result.Valid {
			return fmt.Errorf("规则 %d 验证失败: %s", i+1, strings.Join(result.Errors, "; "))
		}
	}
	
	return nil
}

// GetStatistics 获取规则统计
func (rm *RuleManager) GetStatistics() *RuleStatistics {
	stats := &RuleStatistics{
		RulesByCategory: make(map[string]int),
		RulesBySeverity: make(map[SeverityLevel]int),
		RulesByLanguage: make(map[string]int),
		LastUpdated:     time.Now(),
	}
	
	// 统计所有规则
	allRules := append(rm.builtinRules, rm.customRules...)
	stats.TotalRules = len(allRules)
	stats.BuiltinRules = len(rm.builtinRules)
	stats.CustomRules = len(rm.customRules)
	
	for _, rule := range allRules {
		if rule.Enabled {
			stats.EnabledRules++
		} else {
			stats.DisabledRules++
		}
		
		stats.RulesByCategory[rule.Category]++
		stats.RulesBySeverity[rule.Severity]++
	}
	
	// 统计规则集语言
	for _, ruleset := range rm.rulesets {
		if ruleset.Language != "" {
			stats.RulesByLanguage[ruleset.Language]++
		}
	}
	
	return stats
}

// ExportRuleSet 导出规则集
func (rm *RuleManager) ExportRuleSet(rulesetID string) ([]byte, error) {
	ruleset, exists := rm.rulesets[rulesetID]
	if !exists {
		return nil, fmt.Errorf("规则集不存在: %s", rulesetID)
	}
	
	data, err := json.MarshalIndent(ruleset, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化规则集失败: %v", err)
	}
	
	return data, nil
}

// ImportRuleSet 导入规则集
func (rm *RuleManager) ImportRuleSet(data []byte) error {
	var ruleset RuleSet
	if err := json.Unmarshal(data, &ruleset); err != nil {
		return fmt.Errorf("解析规则集失败: %v", err)
	}
	
	// 检查ID冲突
	if _, exists := rm.rulesets[ruleset.ID]; exists {
		return fmt.Errorf("规则集ID已存在: %s", ruleset.ID)
	}
	
	return rm.CreateRuleSet(&ruleset)
}

// ImportRuleSetFromFile 从文件导入规则集
func (rm *RuleManager) ImportRuleSetFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}
	
	return rm.ImportRuleSet(data)
}

// ExportRuleSetToFile 导出规则集到文件
func (rm *RuleManager) ExportRuleSetToFile(rulesetID, filePath string) error {
	data, err := rm.ExportRuleSet(rulesetID)
	if err != nil {
		return err
	}
	
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}
	
	return os.WriteFile(filePath, data, 0644)
}

// reloadRules 重新加载规则
func (rm *RuleManager) reloadRules() {
	// 清空当前规则
	rm.customRules = []*DetectionRule{}
	
	// 重新加载自定义规则
	for _, ruleset := range rm.rulesets {
		if !strings.HasPrefix(ruleset.ID, "builtin_") && ruleset.Enabled {
			rm.customRules = append(rm.customRules, ruleset.Rules...)
		}
	}
	
	// 更新检测引擎
	rm.updateEngine()
}

// updateEngine 更新检测引擎
func (rm *RuleManager) updateEngine() {
	if rm.engine != nil {
		// 清空引擎规则
		rm.engine.rules = []*DetectionRule{}
		rm.engine.ruleMap = make(map[string]*DetectionRule)
		
		// 添加启用的规则
		allRules := append(rm.builtinRules, rm.customRules...)
		for _, rule := range allRules {
			if rule.Enabled {
				rm.engine.addRule(rule)
			}
		}
	}
}

// GetRuleTemplates 获取规则模板
func (rm *RuleManager) GetRuleTemplates() []*RuleTemplate {
	return []*RuleTemplate{
		{
			ID:          "password_template",
			Name:        "密码检测模板",
			Description: "用于检测密码的通用模板",
			Category:    "credentials",
			Language:    "regex",
			Template:    `(?i){{.field}}\s*[=:]\s*["\']([^"\']{{{.min_length}},})["\']`,
			Variables: map[string]string{
				"field":      "password|pwd|pass|secret",
				"min_length": "6",
			},
			Examples: []string{
				`password = "mypassword123"`,
				`secret: "topsecret"`,
				`pwd="admin123"`,
			},
		},
		{
			ID:          "api_key_template",
			Name:        "API密钥检测模板",
			Description: "用于检测API密钥的通用模板",
			Category:    "api_keys",
			Language:    "regex",
			Template:    `{{.prefix}}[{{.charset}}]{{{.length}}}`,
			Variables: map[string]string{
				"prefix":  "AKIA",
				"charset": "0-9A-Z",
				"length":  "16",
			},
			Examples: []string{
				`AKIA1234567890ABCDEF`,
				`AKIAIOSFODNN7EXAMPLE`,
			},
		},
	}
}