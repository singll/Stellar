package sensitive

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"sort"
	"strings"
	"time"

	"github.com/StellarServer/internal/models"
)

// XMLReport XML报告结构
type XMLReport struct {
	XMLName  xml.Name     `xml:"SensitiveDetectionReport"`
	Metadata XMLMetadata  `xml:"Metadata"`
	Summary  XMLSummary   `xml:"Summary"`
	Findings []XMLFinding `xml:"Findings>Finding"`
}

// XMLMetadata XML元数据结构
type XMLMetadata struct {
	DetectionID   string `xml:"DetectionID"`
	Name          string `xml:"Name"`
	ProjectID     string `xml:"ProjectID"`
	GeneratedAt   string `xml:"GeneratedAt"`
	TotalFindings int    `xml:"TotalFindings"`
	FilteredCount int    `xml:"FilteredCount"`
	Status        string `xml:"Status"`
}

// XMLSummary XML摘要结构
type XMLSummary struct {
	RiskStatistics     map[string]int `xml:"RiskStatistics"`
	CategoryStatistics map[string]int `xml:"CategoryStatistics"`
}

// XMLFinding XML发现结构
type XMLFinding struct {
	Target       string `xml:"Target"`
	TargetType   string `xml:"TargetType"`
	RuleName     string `xml:"RuleName"`
	RiskLevel    string `xml:"RiskLevel"`
	Category     string `xml:"Category"`
	MatchedText  string `xml:"MatchedText"`
	LineNumber   int    `xml:"LineNumber"`
	Context      string `xml:"Context"`
	CreatedAt    string `xml:"CreatedAt"`
}

// ReportGenerator 报告生成器
type ReportGenerator struct {
	detector *Detector
}

// NewReportGenerator 创建报告生成器
func NewReportGenerator(detector *Detector) *ReportGenerator {
	return &ReportGenerator{
		detector: detector,
	}
}

// ReportFormat 报告格式类型
type ReportFormat string

const (
	ReportFormatHTML ReportFormat = "html"
	ReportFormatJSON ReportFormat = "json"
	ReportFormatCSV  ReportFormat = "csv"
	ReportFormatXML  ReportFormat = "xml"
	ReportFormatPDF  ReportFormat = "pdf"
	ReportFormatTXT  ReportFormat = "txt"
)

// ReportRequest 报告生成请求
type ReportRequest struct {
	DetectionID     string       `json:"detectionId" binding:"required"`
	Format          ReportFormat `json:"format" binding:"required"`
	IncludeSummary  bool         `json:"includeSummary"`
	IncludeDetails  bool         `json:"includeDetails"`
	FilterRiskLevel []string     `json:"filterRiskLevel"` // high, medium, low
	FilterCategory  []string     `json:"filterCategory"`
	SortBy          string       `json:"sortBy"`          // riskLevel, category, target, time
	SortOrder       string       `json:"sortOrder"`       // asc, desc
	Template        string       `json:"template"`        // 自定义模板名称
}

// ReportResult 报告生成结果
type ReportResult struct {
	Content     []byte    `json:"content"`
	ContentType string    `json:"contentType"`
	Filename    string    `json:"filename"`
	Size        int       `json:"size"`
	GeneratedAt time.Time `json:"generatedAt"`
}

// GenerateReport 生成报告
func (rg *ReportGenerator) GenerateReport(result *models.SensitiveDetectionResult, req ReportRequest) (*ReportResult, error) {
	// 过滤和排序数据
	filteredFindings := rg.filterFindings(result.Findings, req)
	sortedFindings := rg.sortFindings(filteredFindings, req)

	// 根据格式生成报告
	switch req.Format {
	case ReportFormatHTML:
		return rg.generateHTMLReport(result, sortedFindings, req)
	case ReportFormatJSON:
		return rg.generateJSONReport(result, sortedFindings, req)
	case ReportFormatCSV:
		return rg.generateCSVReport(result, sortedFindings, req)
	case ReportFormatXML:
		return rg.generateXMLReport(result, sortedFindings, req)
	case ReportFormatTXT:
		return rg.generateTXTReport(result, sortedFindings, req)
	default:
		return nil, fmt.Errorf("不支持的报告格式: %s", req.Format)
	}
}

// filterFindings 过滤发现结果
func (rg *ReportGenerator) filterFindings(findings []*models.SensitiveFinding, req ReportRequest) []*models.SensitiveFinding {
	var filtered []*models.SensitiveFinding

	for _, finding := range findings {
		// 风险等级过滤
		if len(req.FilterRiskLevel) > 0 {
			found := false
			for _, level := range req.FilterRiskLevel {
				if finding.RiskLevel == level {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// 分类过滤
		if len(req.FilterCategory) > 0 {
			found := false
			for _, category := range req.FilterCategory {
				if finding.Category == category {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		filtered = append(filtered, finding)
	}

	return filtered
}

// sortFindings 排序发现结果
func (rg *ReportGenerator) sortFindings(findings []*models.SensitiveFinding, req ReportRequest) []*models.SensitiveFinding {
	if req.SortBy == "" {
		req.SortBy = "riskLevel"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	sort.Slice(findings, func(i, j int) bool {
		var compare int
		switch req.SortBy {
		case "riskLevel":
			compare = compareRiskLevel(findings[i].RiskLevel, findings[j].RiskLevel)
		case "category":
			compare = strings.Compare(findings[i].Category, findings[j].Category)
		case "target":
			compare = strings.Compare(findings[i].Target, findings[j].Target)
		case "time":
			compare = int(findings[i].CreatedAt.Unix() - findings[j].CreatedAt.Unix())
		default:
			compare = compareRiskLevel(findings[i].RiskLevel, findings[j].RiskLevel)
		}

		if req.SortOrder == "desc" {
			return compare > 0
		}
		return compare < 0
	})

	return findings
}

// compareRiskLevel 比较风险等级
func compareRiskLevel(level1, level2 string) int {
	levels := map[string]int{
		"critical": 4,
		"high":     3,
		"medium":   2,
		"low":      1,
		"info":     0,
	}

	return levels[level1] - levels[level2]
}

// generateHTMLReport 生成HTML报告
func (rg *ReportGenerator) generateHTMLReport(result *models.SensitiveDetectionResult, findings []*models.SensitiveFinding, req ReportRequest) (*ReportResult, error) {
	tmpl := rg.getHTMLTemplate(req.Template)

	data := struct {
		Result         *models.SensitiveDetectionResult
		Findings       []*models.SensitiveFinding
		GeneratedAt    time.Time
		TotalFindings  int
		FilteredCount  int
		RiskStatistics map[string]int
		CategoryStats  map[string]int
	}{
		Result:         result,
		Findings:       findings,
		GeneratedAt:    time.Now(),
		TotalFindings:  len(result.Findings),
		FilteredCount:  len(findings),
		RiskStatistics: rg.calculateRiskStatistics(findings),
		CategoryStats:  rg.calculateCategoryStatistics(findings),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("生成HTML报告失败: %v", err)
	}

	filename := fmt.Sprintf("sensitive_report_%s_%s.html", 
		result.Name, 
		time.Now().Format("20060102_150405"))

	return &ReportResult{
		Content:     buf.Bytes(),
		ContentType: "text/html; charset=utf-8",
		Filename:    filename,
		Size:        buf.Len(),
		GeneratedAt: time.Now(),
	}, nil
}

// generateJSONReport 生成JSON报告
func (rg *ReportGenerator) generateJSONReport(result *models.SensitiveDetectionResult, findings []*models.SensitiveFinding, req ReportRequest) (*ReportResult, error) {
	report := map[string]interface{}{
		"metadata": map[string]interface{}{
			"detectionId":   result.ID.Hex(),
			"name":          result.Name,
			"projectId":     result.ProjectID.Hex(),
			"generatedAt":   time.Now().Format(time.RFC3339),
			"totalFindings": len(result.Findings),
			"filteredCount": len(findings),
			"status":        result.Status,
			"startTime":     result.StartTime.Format(time.RFC3339),
			"endTime":       result.EndTime.Format(time.RFC3339),
		},
		"summary": map[string]interface{}{
			"riskStatistics":    rg.calculateRiskStatistics(findings),
			"categoryStatistics": rg.calculateCategoryStatistics(findings),
			"targetStatistics":  rg.calculateTargetStatistics(findings),
		},
	}

	if req.IncludeDetails {
		report["findings"] = findings
	}

	content, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("生成JSON报告失败: %v", err)
	}

	filename := fmt.Sprintf("sensitive_report_%s_%s.json", 
		result.Name, 
		time.Now().Format("20060102_150405"))

	return &ReportResult{
		Content:     content,
		ContentType: "application/json",
		Filename:    filename,
		Size:        len(content),
		GeneratedAt: time.Now(),
	}, nil
}

// generateCSVReport 生成CSV报告
func (rg *ReportGenerator) generateCSVReport(result *models.SensitiveDetectionResult, findings []*models.SensitiveFinding, req ReportRequest) (*ReportResult, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入头部
	headers := []string{
		"序号", "目标", "目标类型", "规则名称", "风险等级", "分类",
		"匹配文本", "行号", "上下文", "发现时间", "状态",
	}
	writer.Write(headers)

	// 写入数据
	for i, finding := range findings {
		record := []string{
			fmt.Sprintf("%d", i+1),
			finding.Target,
			finding.TargetType,
			finding.RuleName,
			finding.RiskLevel,
			finding.Category,
			truncateString(finding.MatchedText, 100),
			fmt.Sprintf("%d", finding.LineNumber),
			truncateString(finding.Context, 150),
			finding.CreatedAt.Format("2006-01-02 15:04:05"),
			"", // Status字段暂时为空
		}
		writer.Write(record)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("生成CSV报告失败: %v", err)
	}

	filename := fmt.Sprintf("sensitive_report_%s_%s.csv", 
		result.Name, 
		time.Now().Format("20060102_150405"))

	return &ReportResult{
		Content:     buf.Bytes(),
		ContentType: "text/csv",
		Filename:    filename,
		Size:        buf.Len(),
		GeneratedAt: time.Now(),
	}, nil
}

// generateXMLReport 生成XML报告
func (rg *ReportGenerator) generateXMLReport(result *models.SensitiveDetectionResult, findings []*models.SensitiveFinding, req ReportRequest) (*ReportResult, error) {
	xmlFindings := make([]XMLFinding, len(findings))
	for i, finding := range findings {
		xmlFindings[i] = XMLFinding{
			Target:      finding.Target,
			TargetType:  finding.TargetType,
			RuleName:    finding.RuleName,
			RiskLevel:   finding.RiskLevel,
			Category:    finding.Category,
			MatchedText: finding.MatchedText,
			LineNumber:  finding.LineNumber,
			Context:     finding.Context,
			CreatedAt:   finding.CreatedAt.Format(time.RFC3339),
		}
	}

	report := XMLReport{
		Metadata: XMLMetadata{
			DetectionID:   result.ID.Hex(),
			Name:          result.Name,
			ProjectID:     result.ProjectID.Hex(),
			GeneratedAt:   time.Now().Format(time.RFC3339),
			TotalFindings: len(result.Findings),
			FilteredCount: len(findings),
			Status:        string(result.Status),
		},
		Summary: XMLSummary{
			RiskStatistics:     rg.calculateRiskStatistics(findings),
			CategoryStatistics: rg.calculateCategoryStatistics(findings),
		},
		Findings: xmlFindings,
	}

	content, err := xml.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("生成XML报告失败: %v", err)
	}

	// 添加XML声明
	xmlContent := []byte(xml.Header + string(content))

	filename := fmt.Sprintf("sensitive_report_%s_%s.xml", 
		result.Name, 
		time.Now().Format("20060102_150405"))

	return &ReportResult{
		Content:     xmlContent,
		ContentType: "application/xml",
		Filename:    filename,
		Size:        len(xmlContent),
		GeneratedAt: time.Now(),
	}, nil
}

// generateTXTReport 生成文本报告
func (rg *ReportGenerator) generateTXTReport(result *models.SensitiveDetectionResult, findings []*models.SensitiveFinding, req ReportRequest) (*ReportResult, error) {
	var buf bytes.Buffer

	// 报告头部
	buf.WriteString("=================================================================\n")
	buf.WriteString("                   敏感信息检测报告\n")
	buf.WriteString("=================================================================\n\n")

	// 基本信息
	buf.WriteString("检测信息:\n")
	buf.WriteString(fmt.Sprintf("  检测名称: %s\n", result.Name))
	buf.WriteString(fmt.Sprintf("  项目ID: %s\n", result.ProjectID.Hex()))
	buf.WriteString(fmt.Sprintf("  开始时间: %s\n", result.StartTime.Format("2006-01-02 15:04:05")))
	buf.WriteString(fmt.Sprintf("  结束时间: %s\n", result.EndTime.Format("2006-01-02 15:04:05")))
	buf.WriteString(fmt.Sprintf("  检测状态: %s\n", result.Status))
	buf.WriteString(fmt.Sprintf("  生成时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// 统计信息
	buf.WriteString("统计摘要:\n")
	buf.WriteString(fmt.Sprintf("  总发现数: %d\n", len(result.Findings)))
	buf.WriteString(fmt.Sprintf("  筛选后: %d\n", len(findings)))

	riskStats := rg.calculateRiskStatistics(findings)
	buf.WriteString("  风险等级分布:\n")
	for level, count := range riskStats {
		buf.WriteString(fmt.Sprintf("    %s: %d\n", level, count))
	}

	categoryStats := rg.calculateCategoryStatistics(findings)
	buf.WriteString("  分类分布:\n")
	for category, count := range categoryStats {
		buf.WriteString(fmt.Sprintf("    %s: %d\n", category, count))
	}
	buf.WriteString("\n")

	// 详细发现
	if req.IncludeDetails && len(findings) > 0 {
		buf.WriteString("详细发现:\n")
		buf.WriteString("-----------------------------------------------------------------\n")

		for i, finding := range findings {
			buf.WriteString(fmt.Sprintf("发现 #%d:\n", i+1))
			buf.WriteString(fmt.Sprintf("  目标: %s\n", finding.Target))
			buf.WriteString(fmt.Sprintf("  类型: %s\n", finding.TargetType))
			buf.WriteString(fmt.Sprintf("  规则: %s\n", finding.RuleName))
			buf.WriteString(fmt.Sprintf("  风险等级: %s\n", finding.RiskLevel))
			buf.WriteString(fmt.Sprintf("  分类: %s\n", finding.Category))
			buf.WriteString(fmt.Sprintf("  匹配文本: %s\n", truncateString(finding.MatchedText, 200)))
			if finding.LineNumber > 0 {
				buf.WriteString(fmt.Sprintf("  行号: %d\n", finding.LineNumber))
			}
			if finding.Context != "" {
				buf.WriteString(fmt.Sprintf("  上下文: %s\n", truncateString(finding.Context, 300)))
			}
			buf.WriteString(fmt.Sprintf("  发现时间: %s\n", finding.CreatedAt.Format("2006-01-02 15:04:05")))
			buf.WriteString("-----------------------------------------------------------------\n")
		}
	}

	filename := fmt.Sprintf("sensitive_report_%s_%s.txt", 
		result.Name, 
		time.Now().Format("20060102_150405"))

	return &ReportResult{
		Content:     buf.Bytes(),
		ContentType: "text/plain; charset=utf-8",
		Filename:    filename,
		Size:        buf.Len(),
		GeneratedAt: time.Now(),
	}, nil
}

// 统计计算函数
func (rg *ReportGenerator) calculateRiskStatistics(findings []*models.SensitiveFinding) map[string]int {
	stats := make(map[string]int)
	for _, finding := range findings {
		stats[finding.RiskLevel]++
	}
	return stats
}

func (rg *ReportGenerator) calculateCategoryStatistics(findings []*models.SensitiveFinding) map[string]int {
	stats := make(map[string]int)
	for _, finding := range findings {
		stats[finding.Category]++
	}
	return stats
}

func (rg *ReportGenerator) calculateTargetStatistics(findings []*models.SensitiveFinding) map[string]int {
	stats := make(map[string]int)
	for _, finding := range findings {
		stats[finding.Target]++
	}
	return stats
}

// getHTMLTemplate 获取HTML模板
func (rg *ReportGenerator) getHTMLTemplate(templateName string) *template.Template {
	if templateName != "" {
		// 这里可以加载自定义模板
		// 暂时使用默认模板
	}

	// 创建模板并注册函数
	tmpl := template.New("report").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"truncate": func(s string, length int) string {
			return truncateString(s, length)
		},
	})

	return template.Must(tmpl.Parse(defaultHTMLTemplate))
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// defaultHTMLTemplate 默认HTML模板
const defaultHTMLTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>敏感信息检测报告 - {{.Result.Name}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { border-bottom: 2px solid #3498db; padding-bottom: 20px; margin-bottom: 30px; }
        .title { color: #2c3e50; margin: 0; }
        .subtitle { color: #7f8c8d; margin: 10px 0 0 0; }
        .section { margin-bottom: 30px; }
        .section-title { color: #2c3e50; border-left: 4px solid #3498db; padding-left: 10px; margin-bottom: 15px; }
        .info-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 15px; margin-bottom: 20px; }
        .info-item { background: #f8f9fa; padding: 15px; border-radius: 4px; border-left: 3px solid #3498db; }
        .info-label { font-weight: bold; color: #34495e; }
        .info-value { color: #2c3e50; margin-top: 5px; }
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; }
        .stat-card { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 20px; border-radius: 8px; text-align: center; }
        .stat-number { font-size: 2em; font-weight: bold; margin-bottom: 5px; }
        .stat-label { font-size: 0.9em; opacity: 0.9; }
        .risk-high { background: linear-gradient(135deg, #FF6B6B 0%, #EE5A24 100%); }
        .risk-medium { background: linear-gradient(135deg, #FFA726 0%, #FB8C00 100%); }
        .risk-low { background: linear-gradient(135deg, #66BB6A 0%, #43A047 100%); }
        .findings-table { width: 100%; border-collapse: collapse; margin-top: 15px; }
        .findings-table th, .findings-table td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        .findings-table th { background-color: #3498db; color: white; font-weight: bold; }
        .findings-table tr:nth-child(even) { background-color: #f2f2f2; }
        .risk-badge { padding: 4px 8px; border-radius: 12px; font-size: 0.8em; font-weight: bold; color: white; }
        .risk-badge.high { background-color: #e74c3c; }
        .risk-badge.medium { background-color: #f39c12; }
        .risk-badge.low { background-color: #27ae60; }
        .risk-badge.info { background-color: #3498db; }
        .matched-text { font-family: monospace; background: #f8f9fa; padding: 2px 4px; border-radius: 3px; }
        .footer { text-align: center; margin-top: 40px; padding-top: 20px; border-top: 1px solid #ddd; color: #7f8c8d; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="title">敏感信息检测报告</h1>
            <p class="subtitle">检测名称: {{.Result.Name}} | 生成时间: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</p>
        </div>

        <div class="section">
            <h2 class="section-title">检测信息</h2>
            <div class="info-grid">
                <div class="info-item">
                    <div class="info-label">项目ID</div>
                    <div class="info-value">{{.Result.ProjectID.Hex}}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">开始时间</div>
                    <div class="info-value">{{.Result.StartTime.Format "2006-01-02 15:04:05"}}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">结束时间</div>
                    <div class="info-value">{{.Result.EndTime.Format "2006-01-02 15:04:05"}}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">检测状态</div>
                    <div class="info-value">{{.Result.Status}}</div>
                </div>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">统计摘要</h2>
            <div class="stats-grid">
                <div class="stat-card">
                    <div class="stat-number">{{.TotalFindings}}</div>
                    <div class="stat-label">总发现数</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">{{.FilteredCount}}</div>
                    <div class="stat-label">筛选结果</div>
                </div>
                {{range $level, $count := .RiskStatistics}}
                <div class="stat-card risk-{{$level}}">
                    <div class="stat-number">{{$count}}</div>
                    <div class="stat-label">{{$level}} 风险</div>
                </div>
                {{end}}
            </div>
        </div>

        {{if .Findings}}
        <div class="section">
            <h2 class="section-title">详细发现 ({{len .Findings}} 项)</h2>
            <table class="findings-table">
                <thead>
                    <tr>
                        <th>序号</th>
                        <th>目标</th>
                        <th>规则</th>
                        <th>风险等级</th>
                        <th>分类</th>
                        <th>匹配文本</th>
                        <th>发现时间</th>
                    </tr>
                </thead>
                <tbody>
                    {{range $index, $finding := .Findings}}
                    <tr>
                        <td>{{add $index 1}}</td>
                        <td>{{$finding.Target}}</td>
                        <td>{{$finding.RuleName}}</td>
                        <td><span class="risk-badge {{$finding.RiskLevel}}">{{$finding.RiskLevel}}</span></td>
                        <td>{{$finding.Category}}</td>
                        <td><code class="matched-text">{{truncate $finding.MatchedText 50}}</code></td>
                        <td>{{$finding.FoundAt.Format "01-02 15:04"}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
        {{end}}

        <div class="footer">
            <p>此报告由 Stellar 安全平台自动生成</p>
        </div>
    </div>
</body>
</html>`