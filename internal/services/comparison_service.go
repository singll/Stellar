package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ComparisonService 结果对比和趋势分析服务
type ComparisonService struct {
	db                *mongo.Database
	vulnerabilityCol  *mongo.Collection
	taskCol           *mongo.Collection
	scanResultService *ScanResultService
}

// NewComparisonService 创建对比分析服务
func NewComparisonService(db *mongo.Database, scanResultService *ScanResultService) *ComparisonService {
	return &ComparisonService{
		db:                db,
		vulnerabilityCol:  db.Collection("vulnerabilities"),
		taskCol:           db.Collection("vuln_scan_tasks"),
		scanResultService: scanResultService,
	}
}

// ComparisonResult 对比结果
type ComparisonResult struct {
	BaselineTask    *TaskSummary          `json:"baseline_task"`
	ComparisonTask  *TaskSummary          `json:"comparison_task"`
	VulnChanges     *VulnerabilityChanges `json:"vuln_changes"`
	RiskChanges     *RiskChanges          `json:"risk_changes"`
	Trends          *TrendAnalysis        `json:"trends"`
	ComparedAt      time.Time             `json:"compared_at"`
}

// TaskSummary 任务摘要
type TaskSummary struct {
	TaskID      primitive.ObjectID `json:"task_id"`
	TaskName    string             `json:"task_name"`
	ScanTime    time.Time          `json:"scan_time"`
	TotalVulns  int                `json:"total_vulns"`
	VulnsByType map[string]int     `json:"vulns_by_type"`
	VulnsBySeverity map[string]int `json:"vulns_by_severity"`
	RiskScore   float64            `json:"risk_score"`
}

// VulnerabilityChanges 漏洞变化
type VulnerabilityChanges struct {
	NewVulns     []*models.Vulnerability `json:"new_vulns"`
	FixedVulns   []*models.Vulnerability `json:"fixed_vulns"`
	ChangedVulns []*VulnChange           `json:"changed_vulns"`
	Summary      ChangesSummary          `json:"summary"`
}

// VulnChange 漏洞变化详情
type VulnChange struct {
	VulnID      primitive.ObjectID `json:"vuln_id"`
	Title       string             `json:"title"`
	ChangeType  string             `json:"change_type"` // severity_changed, status_changed
	OldValue    string             `json:"old_value"`
	NewValue    string             `json:"new_value"`
	ChangedAt   time.Time          `json:"changed_at"`
}

// ChangesSummary 变化摘要
type ChangesSummary struct {
	NewVulnsCount     int `json:"new_vulns_count"`
	FixedVulnsCount   int `json:"fixed_vulns_count"`
	ChangedVulnsCount int `json:"changed_vulns_count"`
	NetChange         int `json:"net_change"` // 净变化（新增-修复）
}

// RiskChanges 风险变化
type RiskChanges struct {
	BaselineRisk   float64                `json:"baseline_risk"`
	CurrentRisk    float64                `json:"current_risk"`
	RiskDelta      float64                `json:"risk_delta"`
	RiskTrend      string                 `json:"risk_trend"` // increasing, decreasing, stable
	SeverityChanges map[string]int        `json:"severity_changes"`
	RiskFactors    []string               `json:"risk_factors"`
}

// TrendAnalysis 趋势分析
type TrendAnalysis struct {
	Period         string                 `json:"period"`          // daily, weekly, monthly
	DataPoints     []TrendDataPoint       `json:"data_points"`
	Metrics        TrendMetrics           `json:"metrics"`
	Predictions    *TrendPredictions      `json:"predictions,omitempty"`
}

// TrendDataPoint 趋势数据点
type TrendDataPoint struct {
	Timestamp   time.Time          `json:"timestamp"`
	TotalVulns  int                `json:"total_vulns"`
	RiskScore   float64            `json:"risk_score"`
	VulnsByType map[string]int     `json:"vulns_by_type"`
	VulnsBySeverity map[string]int `json:"vulns_by_severity"`
}

// TrendMetrics 趋势指标
type TrendMetrics struct {
	AvgVulnsPerScan      float64 `json:"avg_vulns_per_scan"`
	AvgRiskScore         float64 `json:"avg_risk_score"`
	VulnDiscoveryRate    float64 `json:"vuln_discovery_rate"`    // 漏洞发现率
	VulnFixRate          float64 `json:"vuln_fix_rate"`          // 漏洞修复率
	RiskReductionRate    float64 `json:"risk_reduction_rate"`    // 风险降低率
	SecurityImprovement  string  `json:"security_improvement"`   // improving, declining, stable
}

// TrendPredictions 趋势预测
type TrendPredictions struct {
	NextScanRisk        float64   `json:"next_scan_risk"`
	PredictedVulnCount  int       `json:"predicted_vuln_count"`
	RiskTrend           string    `json:"risk_trend"`
	RecommendedActions  []string  `json:"recommended_actions"`
}

// CompareTasks 对比两次扫描任务
func (c *ComparisonService) CompareTasks(ctx context.Context, baselineTaskID, comparisonTaskID primitive.ObjectID) (*ComparisonResult, error) {
	// 获取基线任务摘要
	baselineSummary, err := c.getTaskSummary(ctx, baselineTaskID)
	if err != nil {
		return nil, fmt.Errorf("获取基线任务摘要失败: %v", err)
	}
	
	// 获取对比任务摘要
	comparisonSummary, err := c.getTaskSummary(ctx, comparisonTaskID)
	if err != nil {
		return nil, fmt.Errorf("获取对比任务摘要失败: %v", err)
	}
	
	// 分析漏洞变化
	vulnChanges, err := c.analyzeVulnerabilityChanges(ctx, baselineTaskID, comparisonTaskID)
	if err != nil {
		return nil, fmt.Errorf("分析漏洞变化失败: %v", err)
	}
	
	// 分析风险变化
	riskChanges := c.analyzeRiskChanges(baselineSummary, comparisonSummary, vulnChanges)
	
	// 趋势分析
	trends, err := c.analyzeTrends(ctx, []primitive.ObjectID{baselineTaskID, comparisonTaskID})
	if err != nil {
		return nil, fmt.Errorf("趋势分析失败: %v", err)
	}
	
	result := &ComparisonResult{
		BaselineTask:   baselineSummary,
		ComparisonTask: comparisonSummary,
		VulnChanges:    vulnChanges,
		RiskChanges:    riskChanges,
		Trends:         trends,
		ComparedAt:     time.Now(),
	}
	
	return result, nil
}

// getTaskSummary 获取任务摘要
func (c *ComparisonService) getTaskSummary(ctx context.Context, taskID primitive.ObjectID) (*TaskSummary, error) {
	// 获取任务基本信息
	var task models.VulnScanTask
	err := c.taskCol.FindOne(ctx, bson.M{"_id": taskID}).Decode(&task)
	if err != nil {
		return nil, fmt.Errorf("获取任务信息失败: %v", err)
	}
	
	summary := &TaskSummary{
		TaskID:          taskID,
		TaskName:        task.TaskName,
		ScanTime:        task.StartedAt,
		VulnsByType:     make(map[string]int),
		VulnsBySeverity: make(map[string]int),
	}
	
	// 统计漏洞信息
	vulnFilter := bson.M{"taskId": taskID}
	
	// 总漏洞数
	total, err := c.vulnerabilityCol.CountDocuments(ctx, vulnFilter)
	if err != nil {
		return nil, fmt.Errorf("统计漏洞总数失败: %v", err)
	}
	summary.TotalVulns = int(total)
	
	// 按类型统计
	typePipeline := mongo.Pipeline{
		{{"$match", vulnFilter}},
		{{"$group", bson.D{
			{"_id", "$type"},
			{"count", bson.D{{"$sum", 1}}},
		}}},
	}
	
	cursor, err := c.vulnerabilityCol.Aggregate(ctx, typePipeline)
	if err != nil {
		return nil, fmt.Errorf("统计漏洞类型失败: %v", err)
	}
	defer cursor.Close(ctx)
	
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			summary.VulnsByType[result.ID] = result.Count
		}
	}
	
	// 按严重程度统计
	severityPipeline := mongo.Pipeline{
		{{"$match", vulnFilter}},
		{{"$group", bson.D{
			{"_id", "$severity"},
			{"count", bson.D{{"$sum", 1}}},
		}}},
	}
	
	cursor, err = c.vulnerabilityCol.Aggregate(ctx, severityPipeline)
	if err != nil {
		return nil, fmt.Errorf("统计严重程度失败: %v", err)
	}
	defer cursor.Close(ctx)
	
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			summary.VulnsBySeverity[result.ID] = result.Count
		}
	}
	
	// 计算风险评分
	summary.RiskScore = c.calculateTaskRiskScore(summary.VulnsBySeverity)
	
	return summary, nil
}

// calculateTaskRiskScore 计算任务风险评分
func (c *ComparisonService) calculateTaskRiskScore(vulnsBySeverity map[string]int) float64 {
	totalScore := 0.0
	totalVulns := 0
	
	for severity, count := range vulnsBySeverity {
		var weight float64
		switch severity {
		case "critical":
			weight = 10.0
		case "high":
			weight = 7.0
		case "medium":
			weight = 4.0
		case "low":
			weight = 2.0
		case "info":
			weight = 0.5
		default:
			weight = 1.0
		}
		
		totalScore += weight * float64(count)
		totalVulns += count
	}
	
	if totalVulns == 0 {
		return 0.0
	}
	
	return totalScore / float64(totalVulns)
}

// analyzeVulnerabilityChanges 分析漏洞变化
func (c *ComparisonService) analyzeVulnerabilityChanges(ctx context.Context, baselineTaskID, comparisonTaskID primitive.ObjectID) (*VulnerabilityChanges, error) {
	changes := &VulnerabilityChanges{
		NewVulns:     []*models.Vulnerability{},
		FixedVulns:   []*models.Vulnerability{},
		ChangedVulns: []*VulnChange{},
		Summary:      ChangesSummary{},
	}
	
	// 获取基线漏洞
	baselineVulns, err := c.getTaskVulnerabilities(ctx, baselineTaskID)
	if err != nil {
		return nil, fmt.Errorf("获取基线漏洞失败: %v", err)
	}
	
	// 获取对比漏洞
	comparisonVulns, err := c.getTaskVulnerabilities(ctx, comparisonTaskID)
	if err != nil {
		return nil, fmt.Errorf("获取对比漏洞失败: %v", err)
	}
	
	// 创建漏洞映射（基于URL+端口+POC名称）
	baselineMap := make(map[string]*models.Vulnerability)
	for _, vuln := range baselineVulns {
		key := c.getVulnKey(vuln)
		baselineMap[key] = vuln
	}
	
	comparisonMap := make(map[string]*models.Vulnerability)
	for _, vuln := range comparisonVulns {
		key := c.getVulnKey(vuln)
		comparisonMap[key] = vuln
	}
	
	// 分析新增漏洞
	for key, vuln := range comparisonMap {
		if _, exists := baselineMap[key]; !exists {
			changes.NewVulns = append(changes.NewVulns, vuln)
		}
	}
	
	// 分析修复漏洞
	for key, vuln := range baselineMap {
		if _, exists := comparisonMap[key]; !exists {
			changes.FixedVulns = append(changes.FixedVulns, vuln)
		}
	}
	
	// 分析变化漏洞
	for key, baselineVuln := range baselineMap {
		if comparisonVuln, exists := comparisonMap[key]; exists {
			vulnChanges := c.compareVulnerabilities(baselineVuln, comparisonVuln)
			changes.ChangedVulns = append(changes.ChangedVulns, vulnChanges...)
		}
	}
	
	// 统计摘要
	changes.Summary.NewVulnsCount = len(changes.NewVulns)
	changes.Summary.FixedVulnsCount = len(changes.FixedVulns)
	changes.Summary.ChangedVulnsCount = len(changes.ChangedVulns)
	changes.Summary.NetChange = changes.Summary.NewVulnsCount - changes.Summary.FixedVulnsCount
	
	return changes, nil
}

// getVulnKey 生成漏洞唯一标识
func (c *ComparisonService) getVulnKey(vuln *models.Vulnerability) string {
	return fmt.Sprintf("%s:%d:%s", vuln.AffectedURL, vuln.AffectedPort, vuln.POCName)
}

// getTaskVulnerabilities 获取任务的所有漏洞
func (c *ComparisonService) getTaskVulnerabilities(ctx context.Context, taskID primitive.ObjectID) ([]*models.Vulnerability, error) {
	filter := bson.M{"taskId": taskID}
	cursor, err := c.vulnerabilityCol.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var vulns []*models.Vulnerability
	if err := cursor.All(ctx, &vulns); err != nil {
		return nil, err
	}
	
	return vulns, nil
}

// compareVulnerabilities 对比单个漏洞
func (c *ComparisonService) compareVulnerabilities(baseline, comparison *models.Vulnerability) []*VulnChange {
	var changes []*VulnChange
	
	// 严重程度变化
	if baseline.Severity != comparison.Severity {
		changes = append(changes, &VulnChange{
			VulnID:     comparison.ID,
			Title:      comparison.Title,
			ChangeType: "severity_changed",
			OldValue:   string(baseline.Severity),
			NewValue:   string(comparison.Severity),
			ChangedAt:  comparison.UpdatedAt,
		})
	}
	
	// 状态变化
	if baseline.Status != comparison.Status {
		changes = append(changes, &VulnChange{
			VulnID:     comparison.ID,
			Title:      comparison.Title,
			ChangeType: "status_changed",
			OldValue:   string(baseline.Status),
			NewValue:   string(comparison.Status),
			ChangedAt:  comparison.UpdatedAt,
		})
	}
	
	return changes
}

// analyzeRiskChanges 分析风险变化
func (c *ComparisonService) analyzeRiskChanges(baseline, comparison *TaskSummary, vulnChanges *VulnerabilityChanges) *RiskChanges {
	changes := &RiskChanges{
		BaselineRisk:    baseline.RiskScore,
		CurrentRisk:     comparison.RiskScore,
		SeverityChanges: make(map[string]int),
		RiskFactors:     []string{},
	}
	
	changes.RiskDelta = changes.CurrentRisk - changes.BaselineRisk
	
	// 确定风险趋势
	if changes.RiskDelta > 0.5 {
		changes.RiskTrend = "increasing"
		changes.RiskFactors = append(changes.RiskFactors, "风险评分显著上升")
	} else if changes.RiskDelta < -0.5 {
		changes.RiskTrend = "decreasing"
		changes.RiskFactors = append(changes.RiskFactors, "风险评分显著下降")
	} else {
		changes.RiskTrend = "stable"
		changes.RiskFactors = append(changes.RiskFactors, "风险评分基本稳定")
	}
	
	// 分析严重程度变化
	for severity := range baseline.VulnsBySeverity {
		baselineCount := baseline.VulnsBySeverity[severity]
		comparisonCount := comparison.VulnsBySeverity[severity]
		changes.SeverityChanges[severity] = comparisonCount - baselineCount
	}
	
	// 添加风险因素
	if vulnChanges.Summary.NewVulnsCount > 0 {
		changes.RiskFactors = append(changes.RiskFactors, 
			fmt.Sprintf("新增 %d 个漏洞", vulnChanges.Summary.NewVulnsCount))
	}
	
	if vulnChanges.Summary.FixedVulnsCount > 0 {
		changes.RiskFactors = append(changes.RiskFactors, 
			fmt.Sprintf("修复 %d 个漏洞", vulnChanges.Summary.FixedVulnsCount))
	}
	
	return changes
}

// analyzeTrends 分析趋势
func (c *ComparisonService) analyzeTrends(ctx context.Context, taskIDs []primitive.ObjectID) (*TrendAnalysis, error) {
	trends := &TrendAnalysis{
		Period:     "scan_based",
		DataPoints: []TrendDataPoint{},
	}
	
	// 为每个任务创建数据点
	for _, taskID := range taskIDs {
		summary, err := c.getTaskSummary(ctx, taskID)
		if err != nil {
			continue
		}
		
		dataPoint := TrendDataPoint{
			Timestamp:       summary.ScanTime,
			TotalVulns:      summary.TotalVulns,
			RiskScore:       summary.RiskScore,
			VulnsByType:     summary.VulnsByType,
			VulnsBySeverity: summary.VulnsBySeverity,
		}
		
		trends.DataPoints = append(trends.DataPoints, dataPoint)
	}
	
	// 按时间排序
	sort.Slice(trends.DataPoints, func(i, j int) bool {
		return trends.DataPoints[i].Timestamp.Before(trends.DataPoints[j].Timestamp)
	})
	
	// 计算趋势指标
	trends.Metrics = c.calculateTrendMetrics(trends.DataPoints)
	
	// 生成预测
	trends.Predictions = c.generatePredictions(trends.DataPoints, trends.Metrics)
	
	return trends, nil
}

// calculateTrendMetrics 计算趋势指标
func (c *ComparisonService) calculateTrendMetrics(dataPoints []TrendDataPoint) TrendMetrics {
	if len(dataPoints) == 0 {
		return TrendMetrics{}
	}
	
	totalVulns := 0
	totalRisk := 0.0
	
	for _, point := range dataPoints {
		totalVulns += point.TotalVulns
		totalRisk += point.RiskScore
	}
	
	metrics := TrendMetrics{
		AvgVulnsPerScan: float64(totalVulns) / float64(len(dataPoints)),
		AvgRiskScore:    totalRisk / float64(len(dataPoints)),
	}
	
	// 计算趋势
	if len(dataPoints) >= 2 {
		first := dataPoints[0]
		last := dataPoints[len(dataPoints)-1]
		
		vulnChange := last.TotalVulns - first.TotalVulns
		riskChange := last.RiskScore - first.RiskScore
		
		if vulnChange < 0 && riskChange < 0 {
			metrics.SecurityImprovement = "improving"
		} else if vulnChange > 0 || riskChange > 0 {
			metrics.SecurityImprovement = "declining"
		} else {
			metrics.SecurityImprovement = "stable"
		}
		
		// 计算发现率和修复率（简化计算）
		metrics.VulnDiscoveryRate = float64(vulnChange) / float64(len(dataPoints)-1)
		if vulnChange < 0 {
			metrics.VulnFixRate = float64(-vulnChange) / float64(first.TotalVulns) * 100
		}
		
		metrics.RiskReductionRate = (first.RiskScore - last.RiskScore) / first.RiskScore * 100
	}
	
	return metrics
}

// generatePredictions 生成预测
func (c *ComparisonService) generatePredictions(dataPoints []TrendDataPoint, metrics TrendMetrics) *TrendPredictions {
	if len(dataPoints) < 2 {
		return nil
	}
	
	last := dataPoints[len(dataPoints)-1]
	
	predictions := &TrendPredictions{
		RecommendedActions: []string{},
	}
	
	// 简单的线性预测
	riskTrend := metrics.AvgRiskScore
	vulnTrend := int(metrics.AvgVulnsPerScan)
	
	if metrics.SecurityImprovement == "improving" {
		riskTrend = last.RiskScore * 0.9
		vulnTrend = int(float64(last.TotalVulns) * 0.9)
		predictions.RiskTrend = "improving"
		predictions.RecommendedActions = append(predictions.RecommendedActions, "继续当前的安全改进措施")
	} else if metrics.SecurityImprovement == "declining" {
		riskTrend = last.RiskScore * 1.1
		vulnTrend = int(float64(last.TotalVulns) * 1.1)
		predictions.RiskTrend = "declining"
		predictions.RecommendedActions = append(predictions.RecommendedActions, "需要加强安全防护措施")
		predictions.RecommendedActions = append(predictions.RecommendedActions, "增加扫描频率")
	} else {
		predictions.RiskTrend = "stable"
		predictions.RecommendedActions = append(predictions.RecommendedActions, "保持当前的安全措施")
	}
	
	predictions.NextScanRisk = riskTrend
	predictions.PredictedVulnCount = vulnTrend
	
	return predictions
}

// GetProjectTrends 获取项目趋势分析
func (c *ComparisonService) GetProjectTrends(ctx context.Context, projectID primitive.ObjectID, days int) (*TrendAnalysis, error) {
	// 获取项目在指定天数内的所有扫描任务
	startTime := time.Now().AddDate(0, 0, -days)
	filter := bson.M{
		"projectId": projectID,
		"startedAt": bson.M{"$gte": startTime},
		"status":    "completed",
	}
	
	opts := options.Find().SetSort(bson.D{{"startedAt", 1}})
	cursor, err := c.taskCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询扫描任务失败: %v", err)
	}
	defer cursor.Close(ctx)
	
	var tasks []models.VulnScanTask
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("解析任务数据失败: %v", err)
	}
	
	// 提取任务ID
	var taskIDs []primitive.ObjectID
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID)
	}
	
	return c.analyzeTrends(ctx, taskIDs)
}