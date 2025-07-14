package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReportService 报告生成服务
type ReportService struct {
	scanResultService *ScanResultService
	vulnDbService     *VulnDbService
}

// NewReportService 创建报告服务
func NewReportService(scanResultService *ScanResultService, vulnDbService *VulnDbService) *ReportService {
	return &ReportService{
		scanResultService: scanResultService,
		vulnDbService:     vulnDbService,
	}
}

// ReportFormat 报告格式
type ReportFormat string

const (
	FormatJSON ReportFormat = "json"
	FormatCSV  ReportFormat = "csv"
	FormatHTML ReportFormat = "html"
	FormatPDF  ReportFormat = "pdf"
)

// ReportOptions 报告选项
type ReportOptions struct {
	Format          ReportFormat `json:"format"`
	IncludeSummary  bool         `json:"include_summary"`
	IncludeDetails  bool         `json:"include_details"`
	IncludePOC      bool         `json:"include_poc"`
	IncludeCharts   bool         `json:"include_charts"`
	SeverityFilter  []string     `json:"severity_filter"`
	StatusFilter    []string     `json:"status_filter"`
	Language        string       `json:"language"` // zh-CN, en-US
	Template        string       `json:"template"` // 报告模板
}

// ScanReport 扫描报告
type ScanReport struct {
	Metadata        ReportMetadata          `json:"metadata"`
	Summary         *ScanResultSummary      `json:"summary"`
	Vulnerabilities []*VulnerabilityWithRisk `json:"vulnerabilities"`
	Statistics      *ReportStatistics       `json:"statistics"`
	Recommendations []string                `json:"recommendations"`
	GeneratedAt     time.Time               `json:"generated_at"`
}

// ReportMetadata 报告元数据
type ReportMetadata struct {
	TaskID          primitive.ObjectID `json:"task_id"`
	TaskName        string             `json:"task_name"`
	ProjectName     string             `json:"project_name"`
	ScanType        string             `json:"scan_type"`
	Targets         []string           `json:"targets"`
	ScanStartTime   time.Time          `json:"scan_start_time"`
	ScanEndTime     time.Time          `json:"scan_end_time"`
	GeneratedBy     string             `json:"generated_by"`
	ReportVersion   string             `json:"report_version"`
}

// ReportStatistics 报告统计信息
type ReportStatistics struct {
	TotalVulns        int                 `json:"total_vulns"`
	VulnsBySeverity   map[string]int      `json:"vulns_by_severity"`
	VulnsByType       map[string]int      `json:"vulns_by_type"`
	VulnsByStatus     map[string]int      `json:"vulns_by_status"`
	HostsAffected     int                 `json:"hosts_affected"`
	PortsAffected     []int               `json:"ports_affected"`
	TopCVEs           []string            `json:"top_cves"`
	TopCWEs           []string            `json:"top_cwes"`
	RiskDistribution  map[string]float64  `json:"risk_distribution"`
	ComplianceStatus  map[string]string   `json:"compliance_status"`
}

// GenerateReport 生成扫描报告
func (r *ReportService) GenerateReport(ctx context.Context, taskID primitive.ObjectID, options ReportOptions) ([]byte, string, error) {
	// 构建报告数据
	report, err := r.buildReportData(ctx, taskID, options)
	if err != nil {
		return nil, "", fmt.Errorf("构建报告数据失败: %v", err)
	}
	
	// 根据格式生成报告
	switch options.Format {
	case FormatJSON:
		return r.generateJSONReport(report)
	case FormatCSV:
		return r.generateCSVReport(report)
	case FormatHTML:
		return r.generateHTMLReport(report, options)
	case FormatPDF:
		return r.generatePDFReport(report, options)
	default:
		return nil, "", fmt.Errorf("不支持的报告格式: %s", options.Format)
	}
}

// buildReportData 构建报告数据
func (r *ReportService) buildReportData(ctx context.Context, taskID primitive.ObjectID, options ReportOptions) (*ScanReport, error) {
	report := &ScanReport{
		GeneratedAt: time.Now(),
	}
	
	// 获取任务元数据
	metadata, err := r.getReportMetadata(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务元数据失败: %v", err)
	}
	report.Metadata = *metadata
	
	// 获取扫描结果摘要
	if options.IncludeSummary {
		summary, err := r.scanResultService.GetScanResultSummary(ctx, taskID)
		if err != nil {
			return nil, fmt.Errorf("获取扫描摘要失败: %v", err)
		}
		report.Summary = summary
	}
	
	// 获取漏洞详情
	if options.IncludeDetails {
		vulns, err := r.getFilteredVulnerabilities(ctx, taskID, options)
		if err != nil {
			return nil, fmt.Errorf("获取漏洞详情失败: %v", err)
		}
		report.Vulnerabilities = vulns
	}
	
	// 生成统计信息
	statistics, err := r.generateStatistics(ctx, taskID, report.Vulnerabilities)
	if err != nil {
		return nil, fmt.Errorf("生成统计信息失败: %v", err)
	}
	report.Statistics = statistics
	
	// 生成安全建议
	report.Recommendations = r.generateRecommendations(report.Statistics, report.Vulnerabilities)
	
	return report, nil
}

// getReportMetadata 获取报告元数据
func (r *ReportService) getReportMetadata(ctx context.Context, taskID primitive.ObjectID) (*ReportMetadata, error) {
	// 获取任务信息（这里需要从任务服务获取，暂时模拟）
	metadata := &ReportMetadata{
		TaskID:        taskID,
		TaskName:      "漏洞扫描任务",
		ProjectName:   "安全扫描项目",
		ScanType:      "漏洞扫描",
		Targets:       []string{"https://example.com", "192.168.1.100"},
		ScanStartTime: time.Now().Add(-2 * time.Hour),
		ScanEndTime:   time.Now().Add(-30 * time.Minute),
		GeneratedBy:   "Stellar Security Scanner",
		ReportVersion: "1.0",
	}
	
	return metadata, nil
}

// getFilteredVulnerabilities 获取过滤后的漏洞
func (r *ReportService) getFilteredVulnerabilities(ctx context.Context, taskID primitive.ObjectID, options ReportOptions) ([]*VulnerabilityWithRisk, error) {
	// 获取所有漏洞（分页处理）
	var allVulns []*VulnerabilityWithRisk
	page := 1
	pageSize := 100
	
	for {
		vulns, total, err := r.scanResultService.GetVulnerabilitiesByTask(ctx, taskID, page, pageSize)
		if err != nil {
			return nil, err
		}
		
		// 应用过滤器
		filteredVulns := r.applyFilters(vulns, options)
		allVulns = append(allVulns, filteredVulns...)
		
		if int64(page*pageSize) >= total {
			break
		}
		page++
	}
	
	return allVulns, nil
}

// applyFilters 应用过滤器
func (r *ReportService) applyFilters(vulns []*VulnerabilityWithRisk, options ReportOptions) []*VulnerabilityWithRisk {
	var filtered []*VulnerabilityWithRisk
	
	for _, vuln := range vulns {
		// 严重程度过滤
		if len(options.SeverityFilter) > 0 {
			severityMatch := false
			for _, severity := range options.SeverityFilter {
				if string(vuln.Severity) == severity {
					severityMatch = true
					break
				}
			}
			if !severityMatch {
				continue
			}
		}
		
		// 状态过滤
		if len(options.StatusFilter) > 0 {
			statusMatch := false
			for _, status := range options.StatusFilter {
				if string(vuln.Status) == status {
					statusMatch = true
					break
				}
			}
			if !statusMatch {
				continue
			}
		}
		
		filtered = append(filtered, vuln)
	}
	
	return filtered
}

// generateStatistics 生成统计信息
func (r *ReportService) generateStatistics(ctx context.Context, taskID primitive.ObjectID, vulns []*VulnerabilityWithRisk) (*ReportStatistics, error) {
	stats := &ReportStatistics{
		VulnsBySeverity:  make(map[string]int),
		VulnsByType:      make(map[string]int),
		VulnsByStatus:    make(map[string]int),
		RiskDistribution: make(map[string]float64),
		ComplianceStatus: make(map[string]string),
	}
	
	stats.TotalVulns = len(vulns)
	
	hostSet := make(map[string]bool)
	portSet := make(map[int]bool)
	cveSet := make(map[string]bool)
	cweSet := make(map[string]bool)
	
	riskLevels := make(map[string]int)
	
	for _, vuln := range vulns {
		// 按严重程度统计
		stats.VulnsBySeverity[string(vuln.Severity)]++
		
		// 按类型统计
		stats.VulnsByType[string(vuln.Type)]++
		
		// 按状态统计
		stats.VulnsByStatus[string(vuln.Status)]++
		
		// 收集主机
		if vuln.AffectedHost != "" {
			hostSet[vuln.AffectedHost] = true
		}
		
		// 收集端口
		if vuln.AffectedPort > 0 {
			portSet[vuln.AffectedPort] = true
		}
		
		// 收集CVE
		if vuln.CVEID != "" {
			cveSet[vuln.CVEID] = true
		}
		
		// 收集CWE
		if vuln.CWEID != "" {
			cweSet[vuln.CWEID] = true
		}
		
		// 风险等级分布
		riskLevels[vuln.RiskLevel]++
	}
	
	stats.HostsAffected = len(hostSet)
	
	// 转换端口集合为切片
	for port := range portSet {
		stats.PortsAffected = append(stats.PortsAffected, port)
	}
	
	// 提取前10个CVE和CWE
	stats.TopCVEs = r.getTopItems(cveSet, 10)
	stats.TopCWEs = r.getTopItems(cweSet, 10)
	
	// 计算风险分布
	for riskLevel, count := range riskLevels {
		if stats.TotalVulns > 0 {
			stats.RiskDistribution[riskLevel] = float64(count) / float64(stats.TotalVulns) * 100
		}
	}
	
	// 生成合规状态
	stats.ComplianceStatus = r.generateComplianceStatus(stats)
	
	return stats, nil
}

// getTopItems 获取前N个项目
func (r *ReportService) getTopItems(itemSet map[string]bool, limit int) []string {
	var items []string
	count := 0
	for item := range itemSet {
		if count >= limit {
			break
		}
		items = append(items, item)
		count++
	}
	return items
}

// generateComplianceStatus 生成合规状态
func (r *ReportService) generateComplianceStatus(stats *ReportStatistics) map[string]string {
	compliance := make(map[string]string)
	
	// OWASP Top 10 合规检查
	criticalCount := stats.VulnsBySeverity["critical"]
	highCount := stats.VulnsBySeverity["high"]
	
	if criticalCount == 0 && highCount == 0 {
		compliance["OWASP_Top10"] = "合规"
	} else if criticalCount == 0 && highCount <= 5 {
		compliance["OWASP_Top10"] = "基本合规"
	} else {
		compliance["OWASP_Top10"] = "不合规"
	}
	
	// PCI DSS 合规检查
	if stats.TotalVulns <= 10 && criticalCount == 0 {
		compliance["PCI_DSS"] = "合规"
	} else {
		compliance["PCI_DSS"] = "不合规"
	}
	
	// ISO 27001 合规检查
	if stats.TotalVulns <= 20 && criticalCount <= 1 {
		compliance["ISO_27001"] = "合规"
	} else {
		compliance["ISO_27001"] = "不合规"
	}
	
	return compliance
}

// generateRecommendations 生成安全建议
func (r *ReportService) generateRecommendations(stats *ReportStatistics, vulns []*VulnerabilityWithRisk) []string {
	var recommendations []string
	
	// 基于统计信息的建议
	if stats.VulnsBySeverity["critical"] > 0 {
		recommendations = append(recommendations, "立即修复所有严重漏洞，这些漏洞可能导致系统完全沦陷")
	}
	
	if stats.VulnsBySeverity["high"] > 5 {
		recommendations = append(recommendations, "尽快修复高危漏洞，建议在7天内完成修复")
	}
	
	if stats.HostsAffected > 10 {
		recommendations = append(recommendations, "影响主机数量较多，建议优先修复网络边界设备漏洞")
	}
	
	// 基于漏洞类型的建议
	if stats.VulnsByType["web"] > 0 {
		recommendations = append(recommendations, "加强Web应用安全防护，部署WAF并进行代码审计")
	}
	
	if stats.VulnsByType["network"] > 0 {
		recommendations = append(recommendations, "加强网络安全配置，关闭不必要的服务和端口")
	}
	
	// 基于风险因素的建议
	hasRCE := false
	hasInfoDisclosure := false
	
	for _, vuln := range vulns {
		for _, factor := range vuln.RiskFactors {
			if strings.Contains(factor, "远程代码执行") {
				hasRCE = true
			}
			if strings.Contains(factor, "信息泄露") {
				hasInfoDisclosure = true
			}
		}
	}
	
	if hasRCE {
		recommendations = append(recommendations, "发现远程代码执行漏洞，需要立即修复并加强访问控制")
	}
	
	if hasInfoDisclosure {
		recommendations = append(recommendations, "发现信息泄露漏洞，需要加强数据保护措施")
	}
	
	// 通用建议
	recommendations = append(recommendations, "建立定期漏洞扫描机制，及时发现和修复安全问题")
	recommendations = append(recommendations, "加强安全意识培训，提高开发和运维人员的安全意识")
	recommendations = append(recommendations, "建立完善的漏洞管理流程，确保漏洞得到及时处理")
	
	return recommendations
}

// generateJSONReport 生成JSON报告
func (r *ReportService) generateJSONReport(report *ScanReport) ([]byte, string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("序列化JSON失败: %v", err)
	}
	
	filename := fmt.Sprintf("scan_report_%s.json", time.Now().Format("20060102_150405"))
	return data, filename, nil
}

// generateCSVReport 生成CSV报告
func (r *ReportService) generateCSVReport(report *ScanReport) ([]byte, string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	
	// 写入标题行
	headers := []string{
		"漏洞标题", "严重程度", "类型", "状态", "风险评分", "风险等级",
		"受影响主机", "受影响端口", "CVE编号", "CWE编号", "发现时间",
		"风险因素", "描述", "解决方案",
	}
	writer.Write(headers)
	
	// 写入漏洞数据
	for _, vuln := range report.Vulnerabilities {
		record := []string{
			vuln.Title,
			string(vuln.Severity),
			string(vuln.Type),
			string(vuln.Status),
			fmt.Sprintf("%.1f", vuln.RiskScore),
			vuln.RiskLevel,
			vuln.AffectedHost,
			strconv.Itoa(vuln.AffectedPort),
			vuln.CVEID,
			vuln.CWEID,
			vuln.DiscoveredAt.Format("2006-01-02 15:04:05"),
			strings.Join(vuln.RiskFactors, "; "),
			vuln.Description,
			vuln.Solution,
		}
		writer.Write(record)
	}
	
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", fmt.Errorf("写入CSV失败: %v", err)
	}
	
	filename := fmt.Sprintf("scan_report_%s.csv", time.Now().Format("20060102_150405"))
	return buf.Bytes(), filename, nil
}

// generateHTMLReport 生成HTML报告
func (r *ReportService) generateHTMLReport(report *ScanReport, options ReportOptions) ([]byte, string, error) {
	// HTML模板
	htmlTemplate := r.getHTMLTemplate(options.Language)
	
	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return nil, "", fmt.Errorf("解析HTML模板失败: %v", err)
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, report); err != nil {
		return nil, "", fmt.Errorf("生成HTML报告失败: %v", err)
	}
	
	filename := fmt.Sprintf("scan_report_%s.html", time.Now().Format("20060102_150405"))
	return buf.Bytes(), filename, nil
}

// generatePDFReport 生成PDF报告
func (r *ReportService) generatePDFReport(report *ScanReport, options ReportOptions) ([]byte, string, error) {
	// 先生成HTML，然后转换为PDF
	htmlData, _, err := r.generateHTMLReport(report, options)
	if err != nil {
		return nil, "", err
	}
	
	// TODO: 实现HTML到PDF的转换
	// 这里可以使用wkhtmltopdf或其他PDF生成库
	// 暂时返回HTML内容
	
	filename := fmt.Sprintf("scan_report_%s.pdf", time.Now().Format("20060102_150405"))
	return htmlData, filename, nil
}

// getHTMLTemplate 获取HTML模板
func (r *ReportService) getHTMLTemplate(language string) string {
	// 简化的HTML模板
	return `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Stellar 安全扫描报告</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; margin: 40px; }
        .header { text-align: center; border-bottom: 2px solid #333; padding-bottom: 20px; }
        .section { margin: 30px 0; }
        .vuln-item { border: 1px solid #ddd; margin: 10px 0; padding: 15px; border-radius: 5px; }
        .critical { border-left: 5px solid #d32f2f; }
        .high { border-left: 5px solid #f57c00; }
        .medium { border-left: 5px solid #fbc02d; }
        .low { border-left: 5px solid #388e3c; }
        .info { border-left: 5px solid #0288d1; }
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; }
        .stat-card { background: #f5f5f5; padding: 20px; border-radius: 8px; text-align: center; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #f5f5f5; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Stellar 安全扫描报告</h1>
        <p>任务名称: {{.Metadata.TaskName}}</p>
        <p>生成时间: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</p>
    </div>
    
    {{if .Summary}}
    <div class="section">
        <h2>扫描摘要</h2>
        <div class="stats-grid">
            <div class="stat-card">
                <h3>漏洞总数</h3>
                <p style="font-size: 2em; color: #d32f2f;">{{.Summary.TotalVulns}}</p>
            </div>
            <div class="stat-card">
                <h3>受影响主机</h3>
                <p style="font-size: 2em; color: #f57c00;">{{.Summary.HostsAffected}}</p>
            </div>
            <div class="stat-card">
                <h3>风险评分</h3>
                <p style="font-size: 2em; color: #fbc02d;">{{printf "%.1f" .Summary.RiskScore}}</p>
            </div>
            <div class="stat-card">
                <h3>风险等级</h3>
                <p style="font-size: 2em; color: #388e3c;">{{.Summary.RiskLevel}}</p>
            </div>
        </div>
    </div>
    {{end}}
    
    <div class="section">
        <h2>漏洞详情</h2>
        {{range .Vulnerabilities}}
        <div class="vuln-item {{.Severity}}">
            <h3>{{.Title}}</h3>
            <p><strong>严重程度:</strong> {{.Severity}} | <strong>类型:</strong> {{.Type}} | <strong>风险评分:</strong> {{printf "%.1f" .RiskScore}}</p>
            <p><strong>受影响资产:</strong> {{.AffectedHost}}{{if .AffectedPort}}:{{.AffectedPort}}{{end}}</p>
            {{if .CVEID}}<p><strong>CVE编号:</strong> {{.CVEID}}</p>{{end}}
            <p><strong>描述:</strong> {{.Description}}</p>
            {{if .Solution}}<p><strong>解决方案:</strong> {{.Solution}}</p>{{end}}
            {{if .RiskFactors}}<p><strong>风险因素:</strong> {{range .RiskFactors}}{{.}} {{end}}</p>{{end}}
        </div>
        {{end}}
    </div>
    
    <div class="section">
        <h2>安全建议</h2>
        <ul>
        {{range .Recommendations}}
            <li>{{.}}</li>
        {{end}}
        </ul>
    </div>
</body>
</html>
`
}