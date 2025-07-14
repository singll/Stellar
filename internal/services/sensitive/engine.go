package sensitive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ScanURL 扫描URL
func (e *DetectionEngine) ScanURL(url string) (*DetectionResult, error) {
	startTime := time.Now()
	
	result := &DetectionResult{
		ID:        primitive.NewObjectID(),
		URL:       url,
		Matches:   []*SensitiveMatch{},
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	// 获取内容
	content, headers, statusCode, err := e.fetchContent(url)
	if err != nil {
		return nil, fmt.Errorf("获取内容失败: %v", err)
	}

	result.StatusCode = statusCode
	result.ContentType = headers.Get("Content-Type")
	result.ContentSize = len(content)
	result.Title = e.extractTitle(content)

	// 检测敏感信息
	matches := e.detectSensitiveInfo(content)
	result.Matches = matches

	// 生成摘要
	result.Summary = e.generateSummary(matches)
	result.ScanTime = time.Since(startTime)

	// 保存结果
	if err := e.saveResult(result); err != nil {
		fmt.Printf("保存检测结果失败: %v", err)
	}

	return result, nil
}

// fetchContent 获取内容
func (e *DetectionEngine) fetchContent(url string) (string, http.Header, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", nil, 0, err
	}

	req.Header.Set("User-Agent", e.config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := e.client.Do(req)
	if err != nil {
		return "", nil, 0, err
	}
	defer resp.Body.Close()

	// 限制读取大小
	limitedReader := io.LimitReader(resp.Body, int64(e.config.MaxContentSize))
	
	contentBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", nil, resp.StatusCode, err
	}

	return string(contentBytes), resp.Header, resp.StatusCode, nil
}

// extractTitle 提取标题
func (e *DetectionEngine) extractTitle(content string) string {
	titleRegex := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	matches := titleRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// detectSensitiveInfo 检测敏感信息
func (e *DetectionEngine) detectSensitiveInfo(content string) []*SensitiveMatch {
	var matches []*SensitiveMatch

	lines := strings.Split(content, "\n")
	
	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}

		// 检查类别是否启用
		if !e.isCategoryEnabled(rule.Category) {
			continue
		}

		// 检查严重级别
		if !e.isSeverityEnabled(rule.Severity) {
			continue
		}

		// 使用正则表达式匹配
		if rule.Regex != nil {
			ruleMatches := e.findRegexMatches(rule, content, lines)
			matches = append(matches, ruleMatches...)
		}

		// 使用关键词匹配
		if len(rule.Keywords) > 0 {
			keywordMatches := e.findKeywordMatches(rule, content, lines)
			matches = append(matches, keywordMatches...)
		}
	}

	// 去重和排序
	matches = e.deduplicateMatches(matches)
	sort.Slice(matches, func(i, j int) bool {
		return e.getSeverityWeight(matches[i].Severity) > e.getSeverityWeight(matches[j].Severity)
	})

	return matches
}

// findRegexMatches 查找正则匹配
func (e *DetectionEngine) findRegexMatches(rule *DetectionRule, content string, lines []string) []*SensitiveMatch {
	var matches []*SensitiveMatch

	allMatches := rule.Regex.FindAllStringSubmatchIndex(content, -1)
	
	for _, matchIndex := range allMatches {
		if len(matchIndex) < 2 {
			continue
		}

		start := matchIndex[0]
		end := matchIndex[1]
		matchText := content[start:end]

		// 计算行号和列号
		position := e.calculatePosition(content, start)
		
		// 提取上下文
		context := e.extractContext(lines, position.Line, e.config.ContextLines)

		// 计算置信度
		confidence := e.calculateConfidence(rule, matchText, context)

		match := &SensitiveMatch{
			RuleID:      rule.ID,
			RuleName:    rule.Name,
			Category:    rule.Category,
			Severity:    rule.Severity,
			Description: rule.Description,
			Match:       e.maskSensitiveData(matchText, rule.Category),
			Context:     context,
			Position:    position,
			Confidence:  confidence,
			Metadata:    make(map[string]interface{}),
		}

		// 添加元数据
		match.Metadata["rule_type"] = "regex"
		match.Metadata["pattern"] = rule.Pattern
		match.Metadata["full_match"] = matchText

		matches = append(matches, match)
	}

	return matches
}

// findKeywordMatches 查找关键词匹配
func (e *DetectionEngine) findKeywordMatches(rule *DetectionRule, content string, lines []string) []*SensitiveMatch {
	var matches []*SensitiveMatch

	lowerContent := strings.ToLower(content)
	
	for _, keyword := range rule.Keywords {
		lowerKeyword := strings.ToLower(keyword)
		
		// 查找所有关键词出现位置
		start := 0
		for {
			index := strings.Index(lowerContent[start:], lowerKeyword)
			if index == -1 {
				break
			}
			
			actualStart := start + index
			actualEnd := actualStart + len(keyword)
			
			// 检查单词边界
			if !e.isWordBoundary(content, actualStart, actualEnd) {
				start = actualStart + 1
				continue
			}

			matchText := content[actualStart:actualEnd]
			position := e.calculatePosition(content, actualStart)
			context := e.extractContext(lines, position.Line, e.config.ContextLines)
			confidence := e.calculateConfidence(rule, matchText, context)

			match := &SensitiveMatch{
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				Category:    rule.Category,
				Severity:    rule.Severity,
				Description: rule.Description,
				Match:       matchText,
				Context:     context,
				Position:    position,
				Confidence:  confidence,
				Metadata:    make(map[string]interface{}),
			}

			match.Metadata["rule_type"] = "keyword"
			match.Metadata["keyword"] = keyword

			matches = append(matches, match)
			start = actualStart + 1
		}
	}

	return matches
}

// calculatePosition 计算位置
func (e *DetectionEngine) calculatePosition(content string, start int) MatchPosition {
	beforeMatch := content[:start]
	lines := strings.Split(beforeMatch, "\n")
	
	line := len(lines)
	column := 1
	if line > 0 {
		column = len(lines[line-1]) + 1
	}

	return MatchPosition{
		Line:   line,
		Column: column,
		Start:  start,
		End:    start,
	}
}

// extractContext 提取上下文
func (e *DetectionEngine) extractContext(lines []string, lineNum, contextLines int) string {
	start := lineNum - contextLines - 1
	end := lineNum + contextLines

	if start < 0 {
		start = 0
	}
	if end >= len(lines) {
		end = len(lines)
	}

	contextLines_ := lines[start:end]
	return strings.Join(contextLines_, "\n")
}

// calculateConfidence 计算置信度
func (e *DetectionEngine) calculateConfidence(rule *DetectionRule, match, context string) float64 {
	confidence := 0.5 // 基础置信度

	// 根据规则类型调整
	if rule.Regex != nil {
		confidence += 0.3 // 正则匹配置信度较高
	}

	// 根据匹配长度调整
	if len(match) > 20 {
		confidence += 0.1
	}

	// 根据上下文调整
	lowerContext := strings.ToLower(context)
	for _, keyword := range rule.Keywords {
		if strings.Contains(lowerContext, strings.ToLower(keyword)) {
			confidence += 0.1
			break
		}
	}

	// 根据类别调整
	switch rule.Category {
	case "credentials", "api_keys":
		confidence += 0.1
	case "financial":
		confidence += 0.15
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// maskSensitiveData 掩码敏感数据
func (e *DetectionEngine) maskSensitiveData(data, category string) string {
	switch category {
	case "credentials", "api_keys":
		if len(data) <= 8 {
			return strings.Repeat("*", len(data))
		}
		return data[:4] + strings.Repeat("*", len(data)-8) + data[len(data)-4:]
	
	case "financial":
		// 信用卡号显示前4位和后4位
		if len(data) >= 8 {
			return data[:4] + strings.Repeat("*", len(data)-8) + data[len(data)-4:]
		}
		return strings.Repeat("*", len(data))
	
	case "pii":
		// PII数据显示前2位和后2位
		if len(data) >= 4 {
			return data[:2] + strings.Repeat("*", len(data)-4) + data[len(data)-2:]
		}
		return strings.Repeat("*", len(data))
	
	default:
		return data
	}
}

// isWordBoundary 检查单词边界
func (e *DetectionEngine) isWordBoundary(content string, start, end int) bool {
	if start > 0 {
		prevChar := content[start-1]
		if (prevChar >= 'a' && prevChar <= 'z') || 
		   (prevChar >= 'A' && prevChar <= 'Z') || 
		   (prevChar >= '0' && prevChar <= '9') || 
		   prevChar == '_' {
			return false
		}
	}

	if end < len(content) {
		nextChar := content[end]
		if (nextChar >= 'a' && nextChar <= 'z') || 
		   (nextChar >= 'A' && nextChar <= 'Z') || 
		   (nextChar >= '0' && nextChar <= '9') || 
		   nextChar == '_' {
			return false
		}
	}

	return true
}

// isCategoryEnabled 检查类别是否启用
func (e *DetectionEngine) isCategoryEnabled(category string) bool {
	if len(e.config.EnabledCategories) == 0 {
		return true
	}

	for _, enabled := range e.config.EnabledCategories {
		if enabled == category {
			return true
		}
	}
	return false
}

// isSeverityEnabled 检查严重级别是否启用
func (e *DetectionEngine) isSeverityEnabled(severity SeverityLevel) bool {
	configSeverityWeight := e.getSeverityWeight(SeverityLevel(e.config.Severity))
	ruleSeverityWeight := e.getSeverityWeight(severity)
	
	return ruleSeverityWeight >= configSeverityWeight
}

// getSeverityWeight 获取严重级别权重
func (e *DetectionEngine) getSeverityWeight(severity SeverityLevel) int {
	switch severity {
	case SeverityInfo:
		return 1
	case SeverityLow:
		return 2
	case SeverityMedium:
		return 3
	case SeverityHigh:
		return 4
	case SeverityCritical:
		return 5
	default:
		return 0
	}
}

// deduplicateMatches 去重匹配结果
func (e *DetectionEngine) deduplicateMatches(matches []*SensitiveMatch) []*SensitiveMatch {
	seen := make(map[string]bool)
	var unique []*SensitiveMatch

	for _, match := range matches {
		// 使用位置和规则ID作为唯一标识
		key := fmt.Sprintf("%s:%d:%d", match.RuleID, match.Position.Start, match.Position.End)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, match)
		}
	}

	return unique
}

// generateSummary 生成摘要
func (e *DetectionEngine) generateSummary(matches []*SensitiveMatch) *DetectionSummary {
	summary := &DetectionSummary{
		TotalMatches:      len(matches),
		MatchesByCategory: make(map[string]int),
		MatchesBySeverity: make(map[SeverityLevel]int),
		HighestSeverity:   SeverityInfo,
		Categories:        []string{},
	}

	categoryMap := make(map[string]bool)
	highestWeight := 0

	for _, match := range matches {
		// 统计类别
		summary.MatchesByCategory[match.Category]++
		categoryMap[match.Category] = true

		// 统计严重级别
		summary.MatchesBySeverity[match.Severity]++

		// 更新最高严重级别
		weight := e.getSeverityWeight(match.Severity)
		if weight > highestWeight {
			highestWeight = weight
			summary.HighestSeverity = match.Severity
		}
	}

	// 提取类别列表
	for category := range categoryMap {
		summary.Categories = append(summary.Categories, category)
	}
	sort.Strings(summary.Categories)

	// 计算风险分数
	summary.RiskScore = e.calculateRiskScore(matches)

	return summary
}

// calculateRiskScore 计算风险分数
func (e *DetectionEngine) calculateRiskScore(matches []*SensitiveMatch) float64 {
	if len(matches) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, match := range matches {
		// 基础分数按严重级别
		baseScore := float64(e.getSeverityWeight(match.Severity)) * 20.0
		
		// 置信度调整
		confidenceAdjustment := match.Confidence
		
		// 类别权重
		categoryWeight := 1.0
		switch match.Category {
		case "credentials", "api_keys":
			categoryWeight = 1.5
		case "financial":
			categoryWeight = 1.3
		case "pii":
			categoryWeight = 1.2
		}

		score := baseScore * confidenceAdjustment * categoryWeight
		totalScore += score
	}

	// 归一化到0-100
	avgScore := totalScore / float64(len(matches))
	normalizedScore := avgScore / 100.0 * 100.0

	if normalizedScore > 100.0 {
		normalizedScore = 100.0
	}

	return normalizedScore
}

// saveResult 保存结果
func (e *DetectionEngine) saveResult(result *DetectionResult) error {
	collection := e.db.Collection("sensitive_results")
	
	_, err := collection.InsertOne(context.Background(), result)
	if err != nil {
		return fmt.Errorf("保存检测结果失败: %v", err)
	}

	return nil
}

// ScanContent 扫描内容
func (e *DetectionEngine) ScanContent(content string, source string) (*DetectionResult, error) {
	startTime := time.Now()
	
	result := &DetectionResult{
		ID:          primitive.NewObjectID(),
		URL:         source,
		StatusCode:  200,
		ContentType: "text/plain",
		ContentSize: len(content),
		Matches:     []*SensitiveMatch{},
		Metadata:    make(map[string]interface{}),
		CreatedAt:   time.Now(),
	}

	// 检测敏感信息
	matches := e.detectSensitiveInfo(content)
	result.Matches = matches

	// 生成摘要
	result.Summary = e.generateSummary(matches)
	result.ScanTime = time.Since(startTime)

	return result, nil
}

// GetRules 获取所有规则
func (e *DetectionEngine) GetRules() []*DetectionRule {
	return e.rules
}

// GetRule 获取指定规则
func (e *DetectionEngine) GetRule(ruleID string) (*DetectionRule, bool) {
	rule, exists := e.ruleMap[ruleID]
	return rule, exists
}

// AddCustomRule 添加自定义规则
func (e *DetectionEngine) AddCustomRule(rule *DetectionRule) error {
	return e.addRule(rule)
}

// UpdateRule 更新规则
func (e *DetectionEngine) UpdateRule(ruleID string, updates map[string]interface{}) error {
	rule, exists := e.ruleMap[ruleID]
	if !exists {
		return fmt.Errorf("规则不存在: %s", ruleID)
	}

	// 更新规则字段
	if name, ok := updates["name"].(string); ok {
		rule.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		rule.Description = description
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		rule.Enabled = enabled
	}
	if pattern, ok := updates["pattern"].(string); ok {
		if pattern != rule.Pattern {
			regex, err := regexp.Compile(pattern)
			if err != nil {
				return fmt.Errorf("编译正则表达式失败: %v", err)
			}
			rule.Pattern = pattern
			rule.Regex = regex
		}
	}

	rule.UpdatedAt = time.Now()
	return nil
}

// DeleteRule 删除规则
func (e *DetectionEngine) DeleteRule(ruleID string) error {
	_, exists := e.ruleMap[ruleID]
	if !exists {
		return fmt.Errorf("规则不存在: %s", ruleID)
	}

	delete(e.ruleMap, ruleID)

	// 从切片中移除
	for i, rule := range e.rules {
		if rule.ID == ruleID {
			e.rules = append(e.rules[:i], e.rules[i+1:]...)
			break
		}
	}

	return nil
}

// LoadConfig 加载配置
func (e *DetectionEngine) LoadConfig(config *DetectionConfig) {
	e.config = config
	
	// 更新HTTP客户端
	e.client.Timeout = config.Timeout
}

// ExportRules 导出规则
func (e *DetectionEngine) ExportRules() ([]byte, error) {
	data, err := json.MarshalIndent(e.rules, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化规则失败: %v", err)
	}
	return data, nil
}

// ImportRules 导入规则
func (e *DetectionEngine) ImportRules(data []byte) error {
	var rules []*DetectionRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return fmt.Errorf("解析规则失败: %v", err)
	}

	for _, rule := range rules {
		if err := e.addRule(rule); err != nil {
			fmt.Printf("导入规则失败 %s: %v", rule.ID, err)
		}
	}

	return nil
}

// ScanFile 扫描文件
func (e *DetectionEngine) ScanFile(filePath string) (*DetectionResult, error) {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	// 检查文件大小
	if len(content) > e.config.MaxContentSize {
		return nil, fmt.Errorf("文件过大，超过限制")
	}

	// 扫描内容
	return e.ScanContent(string(content), "file://"+filePath)
}

// ScanDirectory 扫描目录
func (e *DetectionEngine) ScanDirectory(dirPath string, recursive bool) ([]*DetectionResult, error) {
	var results []*DetectionResult

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			if !recursive && path != dirPath {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查文件类型
		if !e.isTextFile(path) {
			return nil
		}

		// 扫描文件
		result, err := e.ScanFile(path)
		if err != nil {
			fmt.Printf("扫描文件失败 %s: %v\n", path, err)
			return nil
		}

		if len(result.Matches) > 0 {
			results = append(results, result)
		}

		return nil
	})

	return results, err
}

// isTextFile 检查是否为文本文件
func (e *DetectionEngine) isTextFile(filePath string) bool {
	textExtensions := map[string]bool{
		".txt": true, ".log": true, ".json": true, ".xml": true,
		".html": true, ".htm": true, ".css": true, ".js": true,
		".ts": true, ".java": true, ".py": true, ".go": true,
		".c": true, ".cpp": true, ".h": true, ".hpp": true,
		".php": true, ".rb": true, ".sh": true, ".bat": true,
		".yml": true, ".yaml": true, ".ini": true, ".conf": true,
		".env": true, ".sql": true, ".md": true, ".csv": true,
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	return textExtensions[ext]
}

// GetResults 获取检测结果
func (e *DetectionEngine) GetResults(filter bson.M, limit int64) ([]*DetectionResult, error) {
	collection := e.db.Collection("sensitive_results")
	
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []*DetectionResult
	if err := cursor.All(context.Background(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

// GetResultByID 根据ID获取检测结果
func (e *DetectionEngine) GetResultByID(id primitive.ObjectID) (*DetectionResult, error) {
	collection := e.db.Collection("sensitive_results")
	
	var result DetectionResult
	err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteResult 删除检测结果
func (e *DetectionEngine) DeleteResult(id primitive.ObjectID) error {
	collection := e.db.Collection("sensitive_results")
	
	_, err := collection.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}