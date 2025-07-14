package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ScanResultService 扫描结果服务
type ScanResultService struct {
	db                *mongo.Database
	vulnerabilityCol  *mongo.Collection
	pocResultCol      *mongo.Collection
	vulnDbService     *VulnDbService
}

// NewScanResultService 创建扫描结果服务
func NewScanResultService(db *mongo.Database, vulnDbService *VulnDbService) *ScanResultService {
	service := &ScanResultService{
		db:               db,
		vulnerabilityCol: db.Collection("vulnerabilities"),
		pocResultCol:     db.Collection("poc_results"),
		vulnDbService:    vulnDbService,
	}
	
	// 创建索引
	service.createIndexes()
	
	return service
}

// createIndexes 创建数据库索引
func (s *ScanResultService) createIndexes() {
	ctx := context.Background()
	
	// 漏洞索引
	vulnIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"projectId", 1}, {"createdAt", -1}},
		},
		{
			Keys: bson.D{{"taskId", 1}},
		},
		{
			Keys: bson.D{{"severity", 1}},
		},
		{
			Keys: bson.D{{"status", 1}},
		},
		{
			Keys: bson.D{{"type", 1}},
		},
		{
			Keys: bson.D{{"cveId", 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys: bson.D{{"affectedHost", 1}},
		},
	}
	
	s.vulnerabilityCol.Indexes().CreateMany(ctx, vulnIndexes)
	
	// POC结果索引
	pocIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"taskId", 1}, {"createdAt", -1}},
		},
		{
			Keys: bson.D{{"pocId", 1}},
		},
		{
			Keys: bson.D{{"success", 1}},
		},
		{
			Keys: bson.D{{"target", 1}},
		},
	}
	
	s.pocResultCol.Indexes().CreateMany(ctx, pocIndexes)
}

// ScanResultSummary 扫描结果摘要
type ScanResultSummary struct {
	TaskID           primitive.ObjectID            `json:"task_id"`
	TotalTargets     int                           `json:"total_targets"`
	ScannedTargets   int                           `json:"scanned_targets"`
	TotalVulns       int                           `json:"total_vulns"`
	VulnsBySeverity  map[string]int                `json:"vulns_by_severity"`
	VulnsByType      map[string]int                `json:"vulns_by_type"`
	VulnsByStatus    map[string]int                `json:"vulns_by_status"`
	HostsAffected    int                           `json:"hosts_affected"`
	TopVulns         []*VulnerabilityWithRisk      `json:"top_vulns"`
	RiskScore        float64                       `json:"risk_score"`
	RiskLevel        string                        `json:"risk_level"`
	CompletionRate   float64                       `json:"completion_rate"`
	ExecutionTime    int64                         `json:"execution_time"`
	LastUpdated      time.Time                     `json:"last_updated"`
}

// VulnerabilityWithRisk 带风险评估的漏洞
type VulnerabilityWithRisk struct {
	*models.Vulnerability
	RiskScore    float64 `json:"risk_score"`
	RiskLevel    string  `json:"risk_level"`
	RiskFactors  []string `json:"risk_factors"`
	CVSSInfo     *CVSSInfo `json:"cvss_info,omitempty"`
	POCResults   []*models.POCResult `json:"poc_results,omitempty"`
}

// CVSSInfo CVSS评分信息
type CVSSInfo struct {
	Score          float64 `json:"score"`
	Vector         string  `json:"vector"`
	Version        string  `json:"version"`
	Severity       string  `json:"severity"`
	AttackVector   string  `json:"attack_vector"`
	AttackComplexity string `json:"attack_complexity"`
	UserInteraction string `json:"user_interaction"`
	PrivilegesRequired string `json:"privileges_required"`
	Impact         CVSSImpact `json:"impact"`
}

// CVSSImpact CVSS影响评分
type CVSSImpact struct {
	Confidentiality string `json:"confidentiality"`
	Integrity       string `json:"integrity"`
	Availability    string `json:"availability"`
}

// GetScanResultSummary 获取扫描结果摘要
func (s *ScanResultService) GetScanResultSummary(ctx context.Context, taskID primitive.ObjectID) (*ScanResultSummary, error) {
	summary := &ScanResultSummary{
		TaskID:          taskID,
		VulnsBySeverity: make(map[string]int),
		VulnsByType:     make(map[string]int),
		VulnsByStatus:   make(map[string]int),
		LastUpdated:     time.Now(),
	}
	
	// 获取任务信息
	var task models.VulnScanTask
	err := s.db.Collection("vuln_scan_tasks").FindOne(ctx, bson.M{"_id": taskID}).Decode(&task)
	if err != nil {
		return nil, fmt.Errorf("获取任务信息失败: %v", err)
	}
	
	summary.TotalTargets = len(task.Targets)
	summary.CompletionRate = task.Progress
	summary.ExecutionTime = int64(task.CompletedAt.Sub(task.StartedAt).Milliseconds())
	
	// 统计漏洞信息
	vulnFilter := bson.M{"taskId": taskID}
	
	// 总漏洞数
	total, err := s.vulnerabilityCol.CountDocuments(ctx, vulnFilter)
	if err != nil {
		return nil, fmt.Errorf("统计漏洞总数失败: %v", err)
	}
	summary.TotalVulns = int(total)
	
	// 按严重程度统计
	severityPipeline := mongo.Pipeline{
		{{"$match", vulnFilter}},
		{{"$group", bson.D{
			{"_id", "$severity"},
			{"count", bson.D{{"$sum", 1}}},
		}}},
	}
	
	cursor, err := s.vulnerabilityCol.Aggregate(ctx, severityPipeline)
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
	
	// 按类型统计
	typePipeline := mongo.Pipeline{
		{{"$match", vulnFilter}},
		{{"$group", bson.D{
			{"_id", "$type"},
			{"count", bson.D{{"$sum", 1}}},
		}}},
	}
	
	cursor, err = s.vulnerabilityCol.Aggregate(ctx, typePipeline)
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
	
	// 按状态统计
	statusPipeline := mongo.Pipeline{
		{{"$match", vulnFilter}},
		{{"$group", bson.D{
			{"_id", "$status"},
			{"count", bson.D{{"$sum", 1}}},
		}}},
	}
	
	cursor, err = s.vulnerabilityCol.Aggregate(ctx, statusPipeline)
	if err != nil {
		return nil, fmt.Errorf("统计漏洞状态失败: %v", err)
	}
	defer cursor.Close(ctx)
	
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			summary.VulnsByStatus[result.ID] = result.Count
		}
	}
	
	// 统计受影响主机数
	hostPipeline := mongo.Pipeline{
		{{"$match", vulnFilter}},
		{{"$group", bson.D{
			{"_id", "$affectedHost"},
		}}},
		{{"$count", "hosts"}},
	}
	
	cursor, err = s.vulnerabilityCol.Aggregate(ctx, hostPipeline)
	if err == nil {
		defer cursor.Close(ctx)
		if cursor.Next(ctx) {
			var result struct {
				Hosts int `bson:"hosts"`
			}
			if err := cursor.Decode(&result); err == nil {
				summary.HostsAffected = result.Hosts
			}
		}
	}
	
	// 获取前10个高风险漏洞
	topVulns, err := s.GetTopRiskVulnerabilities(ctx, taskID, 10)
	if err != nil {
		return nil, fmt.Errorf("获取高风险漏洞失败: %v", err)
	}
	summary.TopVulns = topVulns
	
	// 计算总体风险评分
	summary.RiskScore = s.calculateOverallRiskScore(summary.VulnsBySeverity)
	summary.RiskLevel = s.getRiskLevel(summary.RiskScore)
	
	return summary, nil
}

// GetTopRiskVulnerabilities 获取高风险漏洞列表
func (s *ScanResultService) GetTopRiskVulnerabilities(ctx context.Context, taskID primitive.ObjectID, limit int) ([]*VulnerabilityWithRisk, error) {
	filter := bson.M{"taskId": taskID}
	opts := options.Find().
		SetSort(bson.D{{"severity", -1}, {"score", -1}, {"createdAt", -1}}).
		SetLimit(int64(limit))
	
	cursor, err := s.vulnerabilityCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询漏洞失败: %v", err)
	}
	defer cursor.Close(ctx)
	
	var vulnerabilities []*models.Vulnerability
	if err := cursor.All(ctx, &vulnerabilities); err != nil {
		return nil, fmt.Errorf("解析漏洞数据失败: %v", err)
	}
	
	var result []*VulnerabilityWithRisk
	for _, vuln := range vulnerabilities {
		vulnWithRisk, err := s.assessVulnerabilityRisk(ctx, vuln)
		if err != nil {
			continue
		}
		result = append(result, vulnWithRisk)
	}
	
	// 按风险评分排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].RiskScore > result[j].RiskScore
	})
	
	return result, nil
}

// assessVulnerabilityRisk 评估漏洞风险
func (s *ScanResultService) assessVulnerabilityRisk(ctx context.Context, vuln *models.Vulnerability) (*VulnerabilityWithRisk, error) {
	vulnWithRisk := &VulnerabilityWithRisk{
		Vulnerability: vuln,
		RiskFactors:   []string{},
	}
	
	// 基础风险评分
	baseScore := s.getSeverityScore(vuln.Severity)
	
	// 获取CVSS信息
	if vuln.CVEID != "" {
		cvssInfo, err := s.getCVSSInfo(ctx, vuln.CVEID)
		if err == nil && cvssInfo != nil {
			vulnWithRisk.CVSSInfo = cvssInfo
			if cvssInfo.Score > 0 {
				baseScore = cvssInfo.Score
			}
		}
	}
	
	// 风险调整因子
	riskMultiplier := 1.0
	
	// 1. 是否有公开POC
	if vuln.POCName != "" {
		riskMultiplier += 0.3
		vulnWithRisk.RiskFactors = append(vulnWithRisk.RiskFactors, "存在POC验证")
	}
	
	// 2. 网络可达性
	if vuln.AffectedPort > 0 && vuln.AffectedPort < 1024 {
		riskMultiplier += 0.2
		vulnWithRisk.RiskFactors = append(vulnWithRisk.RiskFactors, "影响关键端口")
	}
	
	// 3. 认证需求
	if strings.Contains(strings.ToLower(vuln.Description), "authentication") ||
	   strings.Contains(strings.ToLower(vuln.Description), "login") {
		riskMultiplier += 0.1
		vulnWithRisk.RiskFactors = append(vulnWithRisk.RiskFactors, "可能绕过认证")
	}
	
	// 4. 远程执行
	if strings.Contains(strings.ToLower(vuln.Description), "remote") &&
	   strings.Contains(strings.ToLower(vuln.Description), "execution") {
		riskMultiplier += 0.5
		vulnWithRisk.RiskFactors = append(vulnWithRisk.RiskFactors, "远程代码执行")
	}
	
	// 5. 数据泄露
	if strings.Contains(strings.ToLower(vuln.Description), "disclosure") ||
	   strings.Contains(strings.ToLower(vuln.Description), "leak") {
		riskMultiplier += 0.3
		vulnWithRisk.RiskFactors = append(vulnWithRisk.RiskFactors, "信息泄露风险")
	}
	
	// 6. 业务关键性（根据端口判断）
	if s.isCriticalService(vuln.AffectedPort) {
		riskMultiplier += 0.4
		vulnWithRisk.RiskFactors = append(vulnWithRisk.RiskFactors, "影响关键业务服务")
	}
	
	// 7. 漏洞年龄（新漏洞风险更高）
	if vuln.DiscoveredAt.After(time.Now().AddDate(0, 0, -30)) {
		riskMultiplier += 0.2
		vulnWithRisk.RiskFactors = append(vulnWithRisk.RiskFactors, "近期发现漏洞")
	}
	
	// 计算最终风险评分
	vulnWithRisk.RiskScore = math.Min(baseScore*riskMultiplier, 10.0)
	vulnWithRisk.RiskLevel = s.getRiskLevel(vulnWithRisk.RiskScore)
	
	// 获取相关POC结果
	pocResults, err := s.getPOCResults(ctx, vuln.ID)
	if err == nil {
		vulnWithRisk.POCResults = pocResults
		if len(pocResults) > 0 {
			vulnWithRisk.RiskFactors = append(vulnWithRisk.RiskFactors, "POC验证成功")
		}
	}
	
	return vulnWithRisk, nil
}

// getSeverityScore 获取严重程度基础评分
func (s *ScanResultService) getSeverityScore(severity models.VulnerabilitySeverity) float64 {
	switch severity {
	case models.SeverityCritical:
		return 9.0
	case models.SeverityHigh:
		return 7.0
	case models.SeverityMedium:
		return 5.0
	case models.SeverityLow:
		return 3.0
	case models.SeverityInfo:
		return 1.0
	default:
		return 5.0
	}
}

// getCVSSInfo 获取CVSS信息
func (s *ScanResultService) getCVSSInfo(ctx context.Context, cveID string) (*CVSSInfo, error) {
	if s.vulnDbService == nil {
		return nil, fmt.Errorf("漏洞数据库服务未初始化")
	}
	
	vulnDb, err := s.vulnDbService.GetVulnerabilityByCVE(ctx, cveID)
	if err != nil {
		return nil, err
	}
	
	if vulnDb.CVSSScore == 0 {
		return nil, fmt.Errorf("CVSS评分不存在")
	}
	
	cvssInfo := &CVSSInfo{
		Score:    vulnDb.CVSSScore,
		Vector:   vulnDb.CVSSVector,
		Version:  vulnDb.CVSSVersion,
		Severity: vulnDb.Severity,
	}
	
	// 解析CVSS向量
	if vulnDb.CVSSVector != "" {
		cvssInfo.parseVector(vulnDb.CVSSVector, s)
	}
	
	return cvssInfo, nil
}

// parseVector 解析CVSS向量
func (c *CVSSInfo) parseVector(vector string, srs *ScanResultService) {
	parts := strings.Split(vector, "/")
	for _, part := range parts {
		if strings.Contains(part, ":") {
			kv := strings.Split(part, ":")
			if len(kv) != 2 {
				continue
			}
			
			key, value := kv[0], kv[1]
			switch key {
			case "AV":
				c.AttackVector = srs.mapAttackVector(value)
			case "AC":
				c.AttackComplexity = srs.mapAttackComplexity(value)
			case "PR":
				c.PrivilegesRequired = srs.mapPrivilegesRequired(value)
			case "UI":
				c.UserInteraction = srs.mapUserInteraction(value)
			case "C":
				c.Impact.Confidentiality = srs.mapImpact(value)
			case "I":
				c.Impact.Integrity = srs.mapImpact(value)
			case "A":
				c.Impact.Availability = srs.mapImpact(value)
			}
		}
	}
}

// mapAttackVector 映射攻击向量
func (s *ScanResultService) mapAttackVector(value string) string {
	switch value {
	case "N":
		return "Network"
	case "A":
		return "Adjacent"
	case "L":
		return "Local"
	case "P":
		return "Physical"
	default:
		return value
	}
}

// mapAttackComplexity 映射攻击复杂度
func (s *ScanResultService) mapAttackComplexity(value string) string {
	switch value {
	case "L":
		return "Low"
	case "H":
		return "High"
	default:
		return value
	}
}

// mapPrivilegesRequired 映射所需权限
func (s *ScanResultService) mapPrivilegesRequired(value string) string {
	switch value {
	case "N":
		return "None"
	case "L":
		return "Low"
	case "H":
		return "High"
	default:
		return value
	}
}

// mapUserInteraction 映射用户交互
func (s *ScanResultService) mapUserInteraction(value string) string {
	switch value {
	case "N":
		return "None"
	case "R":
		return "Required"
	default:
		return value
	}
}

// mapImpact 映射影响级别
func (s *ScanResultService) mapImpact(value string) string {
	switch value {
	case "N":
		return "None"
	case "L":
		return "Low"
	case "H":
		return "High"
	default:
		return value
	}
}

// isCriticalService 判断是否为关键服务端口
func (s *ScanResultService) isCriticalService(port int) bool {
	criticalPorts := []int{
		22, 23, 25, 53, 80, 110, 143, 443, 993, 995, // 基础服务
		1433, 1521, 3306, 5432, 6379, 27017,        // 数据库
		8080, 8443, 9000, 9090,                      // Web服务
		21, 22, 3389,                                // 远程访问
	}
	
	for _, criticalPort := range criticalPorts {
		if port == criticalPort {
			return true
		}
	}
	
	return false
}

// calculateOverallRiskScore 计算总体风险评分
func (s *ScanResultService) calculateOverallRiskScore(vulnsBySeverity map[string]int) float64 {
	if len(vulnsBySeverity) == 0 {
		return 0.0
	}
	
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
	
	avgScore := totalScore / float64(totalVulns)
	
	// 根据漏洞数量调整评分（漏洞数量越多，整体风险越高）
	volumeMultiplier := 1.0 + math.Log10(float64(totalVulns+1))*0.1
	
	return math.Min(avgScore*volumeMultiplier, 10.0)
}

// getRiskLevel 获取风险等级
func (s *ScanResultService) getRiskLevel(score float64) string {
	if score >= 9.0 {
		return "极高"
	} else if score >= 7.0 {
		return "高"
	} else if score >= 5.0 {
		return "中"
	} else if score >= 3.0 {
		return "低"
	} else {
		return "极低"
	}
}

// getPOCResults 获取POC执行结果
func (s *ScanResultService) getPOCResults(ctx context.Context, vulnID primitive.ObjectID) ([]*models.POCResult, error) {
	filter := bson.M{"vulnId": vulnID}
	opts := options.Find().SetSort(bson.D{{"createdAt", -1}}).SetLimit(5)
	
	cursor, err := s.pocResultCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var results []*models.POCResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	
	return results, nil
}

// GetVulnerabilityDetails 获取漏洞详细信息
func (s *ScanResultService) GetVulnerabilityDetails(ctx context.Context, vulnID primitive.ObjectID) (*VulnerabilityWithRisk, error) {
	var vuln models.Vulnerability
	err := s.vulnerabilityCol.FindOne(ctx, bson.M{"_id": vulnID}).Decode(&vuln)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("漏洞不存在")
		}
		return nil, fmt.Errorf("获取漏洞详情失败: %v", err)
	}
	
	return s.assessVulnerabilityRisk(ctx, &vuln)
}

// GetVulnerabilitiesByTask 获取任务的所有漏洞
func (s *ScanResultService) GetVulnerabilitiesByTask(ctx context.Context, taskID primitive.ObjectID, page, pageSize int) ([]*VulnerabilityWithRisk, int64, error) {
	filter := bson.M{"taskId": taskID}
	
	// 计算总数
	total, err := s.vulnerabilityCol.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("统计漏洞数量失败: %v", err)
	}
	
	// 分页查询
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	
	opts := options.Find().
		SetSort(bson.D{{"severity", -1}, {"createdAt", -1}}).
		SetSkip(skip).
		SetLimit(limit)
	
	cursor, err := s.vulnerabilityCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("查询漏洞失败: %v", err)
	}
	defer cursor.Close(ctx)
	
	var vulnerabilities []*models.Vulnerability
	if err := cursor.All(ctx, &vulnerabilities); err != nil {
		return nil, 0, fmt.Errorf("解析漏洞数据失败: %v", err)
	}
	
	var result []*VulnerabilityWithRisk
	for _, vuln := range vulnerabilities {
		vulnWithRisk, err := s.assessVulnerabilityRisk(ctx, vuln)
		if err != nil {
			continue
		}
		result = append(result, vulnWithRisk)
	}
	
	return result, total, nil
}