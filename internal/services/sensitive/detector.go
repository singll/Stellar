package sensitive

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Detector 敏感信息检测器
type Detector struct {
	rules     []*models.SensitiveRule
	whitelist []*models.SensitiveWhitelist
	config    models.SensitiveDetectionConfig
	logger    *utils.Logger
}

// NewDetector 创建敏感信息检测器
func NewDetector(rules []*models.SensitiveRule, whitelist []*models.SensitiveWhitelist, config models.SensitiveDetectionConfig, logger *utils.Logger) *Detector {
	return &Detector{
		rules:     rules,
		whitelist: whitelist,
		config:    config,
		logger:    logger,
	}
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
	var findings []*models.SensitiveFinding

	// 获取目标内容
	content, err := d.fetchContent(ctx, target, config)
	if err != nil {
		return nil, err
	}

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

			// 提取上下文
			context := d.extractContext(content, matchedText, config.ContextLines)

			// 创建发现
			finding := &models.SensitiveFinding{
				ID:          primitive.NewObjectID(),
				Target:      target,
				Rule:        rule.ID,
				RuleName:    rule.Name,
				Category:    rule.Category,
				RiskLevel:   rule.RiskLevel,
				Pattern:     rule.Pattern,
				MatchedText: matchedText,
				Context:     context,
				CreatedAt:   time.Now(),
			}

			findings = append(findings, finding)
		}
	}

	return findings, nil
}

// fetchContent 获取目标内容
func (d *Detector) fetchContent(ctx context.Context, target string, config models.SensitiveDetectionConfig) (string, error) {
	// 检查目标类型
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		// 处理网页内容
		return d.fetchWebContent(ctx, target, config)
	} else {
		// 处理文件内容 (这里需要根据实际情况实现)
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
