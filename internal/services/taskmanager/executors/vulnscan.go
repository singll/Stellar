package executors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/services/vulnscan"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// VulnScanExecutor 漏洞扫描执行器
type VulnScanExecutor struct {
	db        *mongo.Database
	pocEngine *vulnscan.POCEngine
	handler   *vulnscan.VulnHandler
}

// VulnScanConfig 漏洞扫描配置
type VulnScanConfig struct {
	POCIDs          []string              `json:"pocIds"`          // POC ID列表
	POCCategories   []string              `json:"pocCategories"`   // POC分类
	Targets         []string              `json:"targets"`         // 扫描目标
	Concurrency     int                   `json:"concurrency"`     // 并发数
	Timeout         int                   `json:"timeout"`         // 超时时间(秒)
	RetryCount      int                   `json:"retryCount"`      // 重试次数
	RateLimit       int                   `json:"rateLimit"`       // 请求速率限制(每秒)
	FollowRedirect  bool                  `json:"followRedirect"`  // 是否跟随重定向
	CustomHeaders   map[string]string     `json:"customHeaders"`   // 自定义请求头
	Cookies         string                `json:"cookies"`         // Cookie
	Proxy           string                `json:"proxy"`           // 代理
	ScanDepth       int                   `json:"scanDepth"`       // 扫描深度
	SaveToDB        bool                  `json:"saveToDB"`        // 是否保存到数据库
	VerifyVuln      bool                  `json:"verifyVuln"`      // 是否验证漏洞
	MinimumSeverity models.VulnerabilitySeverity `json:"minimumSeverity"` // 最低严重性级别
}

// VulnScanResult 漏洞扫描结果
type VulnScanResult struct {
	TotalTargets    int                      `json:"totalTargets"`
	ScannedTargets  int                      `json:"scannedTargets"`
	TotalVulns      int                      `json:"totalVulns"`
	Vulnerabilities []models.Vulnerability   `json:"vulnerabilities"`
	POCResults      []models.POCResult       `json:"pocResults"`
	Summary         models.VulnScanSummary   `json:"summary"`
	ErrorCount      int                      `json:"errorCount"`
	Errors          []string                 `json:"errors"`
}

// NewVulnScanExecutor 创建漏洞扫描执行器
func NewVulnScanExecutor(db *mongo.Database, pocEngine *vulnscan.POCEngine, handler *vulnscan.VulnHandler) *VulnScanExecutor {
	return &VulnScanExecutor{
		db:        db,
		pocEngine: pocEngine,
		handler:   handler,
	}
}

// Execute 执行漏洞扫描任务
func (e *VulnScanExecutor) Execute(ctx context.Context, task *models.Task) (*models.TaskResult, error) {
	// 解析配置
	var config VulnScanConfig
	
	// 将map转换为JSON再解析为结构体
	configBytes, err := json.Marshal(task.Config)
	if err != nil {
		return nil, fmt.Errorf("序列化任务配置失败: %v", err)
	}
	
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("解析漏洞扫描配置失败: %v", err)
	}

	// 初始化结果
	result := &VulnScanResult{
		TotalTargets:    len(config.Targets),
		ScannedTargets:  0,
		TotalVulns:      0,
		Vulnerabilities: []models.Vulnerability{},
		POCResults:      []models.POCResult{},
		Summary: models.VulnScanSummary{
			TotalTargets:   len(config.Targets),
			ScannedTargets: 0,
			TotalVulns:     0,
			CriticalVulns:  0,
			HighVulns:      0,
			MediumVulns:    0,
			LowVulns:       0,
			InfoVulns:      0,
			VulnTypes:      make(map[string]int),
		},
		ErrorCount: 0,
		Errors:     []string{},
	}

	// 更新任务状态为运行中
	if err := e.updateTaskStatus(task, string(models.TaskStatusRunning), 0, result); err != nil {
		return nil, fmt.Errorf("更新任务状态失败: %v", err)
	}

	// 获取POC列表
	pocs, err := e.getPOCs(config.POCIDs, config.POCCategories)
	if err != nil {
		return nil, fmt.Errorf("获取POC列表失败: %v", err)
	}

	if len(pocs) == 0 {
		return nil, fmt.Errorf("没有找到可用的POC")
	}

	// 对每个目标执行扫描
	for i, target := range config.Targets {
		select {
		case <-ctx.Done():
			return e.createTaskResult(task, result, string(models.TaskStatusCancelled))
		default:
		}

		// 扫描单个目标
		targetResult, err := e.scanTarget(ctx, target, pocs, &config, task)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("扫描目标 %s 失败: %v", target, err))
		} else {
			// 合并结果
			result.Vulnerabilities = append(result.Vulnerabilities, targetResult.Vulnerabilities...)
			result.POCResults = append(result.POCResults, targetResult.POCResults...)
			result.TotalVulns += targetResult.TotalVulns
			
			// 更新统计
			e.updateSummary(&result.Summary, targetResult.Vulnerabilities)
		}

		result.ScannedTargets++
		result.Summary.ScannedTargets = result.ScannedTargets

		// 更新进度
		progress := float64(i+1) / float64(len(config.Targets)) * 100
		if err := e.updateTaskStatus(task, string(models.TaskStatusRunning), progress, result); err != nil {
			return nil, fmt.Errorf("更新任务进度失败: %v", err)
		}
	}

	// 任务完成
	status := models.TaskStatusCompleted
	if result.ErrorCount > 0 && result.TotalVulns == 0 {
		status = models.TaskStatusFailed
	}

	if err := e.updateTaskStatus(task, string(status), 100, result); err != nil {
		return nil, fmt.Errorf("更新任务完成状态失败: %v", err)
	}

	return e.createTaskResult(task, result, string(status))
}

// scanTarget 扫描单个目标
func (e *VulnScanExecutor) scanTarget(ctx context.Context, targetURL string, pocs []models.POC, config *VulnScanConfig, task *models.Task) (*VulnScanResult, error) {
	result := &VulnScanResult{
		TotalTargets:    1,
		ScannedTargets:  0,
		TotalVulns:      0,
		Vulnerabilities: []models.Vulnerability{},
		POCResults:      []models.POCResult{},
	}

	// 创建扫描目标
	target := vulnscan.POCTarget{
		URL:    targetURL,
		Host:   extractHost(targetURL),
		Port:   extractPort(targetURL),
		Scheme: extractScheme(targetURL),
		Path:   extractPath(targetURL),
		Query:  extractQuery(targetURL),
		Extra:  make(map[string]string),
	}

	// 添加自定义头和Cookie
	if config.CustomHeaders != nil {
		for k, v := range config.CustomHeaders {
			target.Extra[k] = v
		}
	}
	if config.Cookies != "" {
		target.Extra["cookies"] = config.Cookies
	}

	// 对每个POC执行扫描
	for _, poc := range pocs {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		// 检查严重性过滤
		if config.MinimumSeverity != "" && !e.severityFilter(poc.Severity, config.MinimumSeverity) {
			continue
		}

		// 执行POC
		pocResult, err := e.pocEngine.ExecutePOC(ctx, &poc, target)
		if err != nil {
			// POC执行错误，记录但继续
			continue
		}

		// 设置任务ID
		pocResult.TaskID = task.ID

		// 保存POC结果
		if config.SaveToDB {
			if err := e.handler.HandlePOCResult(pocResult); err != nil {
				// 记录错误但继续
			}
		}

		result.POCResults = append(result.POCResults, *pocResult)

		// 如果发现漏洞，创建漏洞记录
		if pocResult.Success {
			vuln := e.createVulnerabilityFromPOC(&poc, pocResult, task, targetURL)
			
			if config.SaveToDB {
				if err := e.handler.HandleVulnerability(&vuln); err != nil {
					// 记录错误但继续
				}
			}
			
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
			result.TotalVulns++
		}
	}

	result.ScannedTargets = 1
	return result, nil
}

// getPOCs 获取POC列表
func (e *VulnScanExecutor) getPOCs(pocIDs []string, categories []string) ([]models.POC, error) {
	query := make(map[string]interface{})
	
	// 构建查询条件
	if len(pocIDs) > 0 {
		objectIDs := make([]primitive.ObjectID, 0, len(pocIDs))
		for _, id := range pocIDs {
			if objID, err := primitive.ObjectIDFromHex(id); err == nil {
				objectIDs = append(objectIDs, objID)
			}
		}
		if len(objectIDs) > 0 {
			query["_id"] = map[string]interface{}{"$in": objectIDs}
		}
	}
	
	if len(categories) > 0 {
		query["category"] = map[string]interface{}{"$in": categories}
	}

	// 只获取已启用的POC
	query["enabled"] = true

	// 获取POC列表
	pocs, _, err := e.handler.GetPOCs(query, 1, 1000) // 限制最多1000个POC
	return pocs, err
}

// createVulnerabilityFromPOC 从POC结果创建漏洞记录
func (e *VulnScanExecutor) createVulnerabilityFromPOC(poc *models.POC, pocResult *models.POCResult, task *models.Task, targetURL string) models.Vulnerability {
	vuln := models.Vulnerability{
		ID:            primitive.NewObjectID(),
		ProjectID:     task.ProjectID,
		TaskID:        task.ID,
		Title:         poc.Name,
		Description:   poc.Description,
		CVEID:         poc.CVEID,
		CWEID:         poc.CWEID,
		Severity:      poc.Severity,
		Status:        models.StatusUnverified,
		Type:          poc.Type,
		AffectedURL:   targetURL,
		AffectedHost:  extractHost(targetURL),
		Payload:       pocResult.Payload,
		Request:       pocResult.Request,
		Response:      pocResult.Response,
		POCName:       poc.Name,
		DiscoveredAt:  time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Score:         e.calculateScore(poc.Severity),
	}

	return vuln
}

// updateTaskStatus 更新任务状态
func (e *VulnScanExecutor) updateTaskStatus(task *models.Task, status string, progress float64, result *VulnScanResult) error {
	// 这里应该调用TaskManager的更新方法
	// 暂时简化实现
	return nil
}

// updateSummary 更新扫描摘要
func (e *VulnScanExecutor) updateSummary(summary *models.VulnScanSummary, vulns []models.Vulnerability) {
	for _, vuln := range vulns {
		summary.TotalVulns++
		
		switch vuln.Severity {
		case models.SeverityCritical:
			summary.CriticalVulns++
		case models.SeverityHigh:
			summary.HighVulns++
		case models.SeverityMedium:
			summary.MediumVulns++
		case models.SeverityLow:
			summary.LowVulns++
		case models.SeverityInfo:
			summary.InfoVulns++
		}
		
		// 统计漏洞类型
		typeStr := string(vuln.Type)
		summary.VulnTypes[typeStr]++
	}
}

// severityFilter 严重性过滤器
func (e *VulnScanExecutor) severityFilter(pocSeverity, minSeverity models.VulnerabilitySeverity) bool {
	severityLevels := map[models.VulnerabilitySeverity]int{
		models.SeverityInfo:     1,
		models.SeverityLow:      2,
		models.SeverityMedium:   3,
		models.SeverityHigh:     4,
		models.SeverityCritical: 5,
	}
	
	return severityLevels[pocSeverity] >= severityLevels[minSeverity]
}

// calculateScore 计算漏洞评分
func (e *VulnScanExecutor) calculateScore(severity models.VulnerabilitySeverity) float64 {
	scoreMap := map[models.VulnerabilitySeverity]float64{
		models.SeverityInfo:     2.0,
		models.SeverityLow:      4.0,
		models.SeverityMedium:   6.0,
		models.SeverityHigh:     8.0,
		models.SeverityCritical: 10.0,
	}
	
	return scoreMap[severity]
}

// createTaskResult 创建任务结果
func (e *VulnScanExecutor) createTaskResult(task *models.Task, result *VulnScanResult, status string) (*models.TaskResult, error) {
	// 将结果数据转换为map[string]interface{}
	resultMap := map[string]interface{}{
		"totalTargets":    result.TotalTargets,
		"scannedTargets":  result.ScannedTargets,
		"totalVulns":      result.TotalVulns,
		"vulnerabilities": result.Vulnerabilities,
		"pocResults":      result.POCResults,
		"summary":         result.Summary,
		"errorCount":      result.ErrorCount,
		"errors":          result.Errors,
	}

	// 生成摘要字符串
	summary := fmt.Sprintf("扫描完成: %d/%d个目标，发现%d个漏洞",
		result.ScannedTargets, result.TotalTargets, result.TotalVulns)
	if result.ErrorCount > 0 {
		summary += fmt.Sprintf("，%d个错误", result.ErrorCount)
	}

	taskResult := &models.TaskResult{
		ID:        primitive.NewObjectID(),
		TaskID:    task.ID,
		Status:    status,
		Data:      resultMap,
		Summary:   summary,
		StartTime: time.Now(), // 这里简化处理，实际应该记录真实开始时间
		EndTime:   time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return taskResult, nil
}

// GetSupportedTypes 获取支持的任务类型
func (e *VulnScanExecutor) GetSupportedTypes() []string {
	return []string{"vuln_scan"}
}

// GetExecutorInfo 获取执行器信息
func (e *VulnScanExecutor) GetExecutorInfo() models.ExecutorInfo {
	return models.ExecutorInfo{
		Name:        "VulnScanExecutor",
		Version:     "1.0.0",
		Description: "漏洞扫描执行器，支持多种POC类型的漏洞检测",
		Author:      "Stellar Team",
	}
}

// URL解析辅助函数
func extractHost(url string) string {
	// 简化实现，实际应该使用url.Parse
	return url
}

func extractPort(url string) int {
	// 简化实现
	return 80
}

func extractScheme(url string) string {
	// 简化实现
	return "http"
}

func extractPath(url string) string {
	// 简化实现
	return "/"
}

func extractQuery(url string) string {
	// 简化实现
	return ""
}