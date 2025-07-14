package vulndb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/StellarServer/internal/models"
)

// CVEDataFetcher CVE数据获取器
type CVEDataFetcher struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewCVEDataFetcher 创建CVE数据获取器
func NewCVEDataFetcher(apiKey string) *CVEDataFetcher {
	return &CVEDataFetcher{
		baseURL:    "https://services.nvd.nist.gov/rest/json/cves/2.0",
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// CVEResponse CVE API响应结构
type CVEResponse struct {
	ResultsPerPage  int               `json:"resultsPerPage"`
	StartIndex      int               `json:"startIndex"`
	TotalResults    int               `json:"totalResults"`
	Format          string            `json:"format"`
	Version         string            `json:"version"`
	Timestamp       string            `json:"timestamp"`
	Vulnerabilities []CVEVulnerability `json:"vulnerabilities"`
}

// CVEVulnerability CVE漏洞数据
type CVEVulnerability struct {
	CVE CVEData `json:"cve"`
}

// CVEData CVE核心数据
type CVEData struct {
	ID               string           `json:"id"`
	SourceIdentifier string           `json:"sourceIdentifier"`
	Published        string           `json:"published"`
	LastModified     string           `json:"lastModified"`
	VulnStatus       string           `json:"vulnStatus"`
	Descriptions     []CVEDescription `json:"descriptions"`
	Metrics          CVEMetrics       `json:"metrics,omitempty"`
	Weaknesses       []CVEWeakness    `json:"weaknesses,omitempty"`
	Configurations   []CVEConfigData  `json:"configurations,omitempty"`
	References       []CVEReference   `json:"references,omitempty"`
}

// CVEDescription CVE描述
type CVEDescription struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

// CVEMetrics CVE评分指标
type CVEMetrics struct {
	CvssMetricV31 []CVSSMetricV31 `json:"cvssMetricV31,omitempty"`
	CvssMetricV30 []CVSSMetricV30 `json:"cvssMetricV30,omitempty"`
	CvssMetricV2  []CVSSMetricV2  `json:"cvssMetricV2,omitempty"`
}

// CVSSMetricV31 CVSS v3.1评分
type CVSSMetricV31 struct {
	Source   string      `json:"source"`
	Type     string      `json:"type"`
	CvssData CVSSDataV31 `json:"cvssData"`
}

// CVSSDataV31 CVSS v3.1数据
type CVSSDataV31 struct {
	Version               string  `json:"version"`
	VectorString          string  `json:"vectorString"`
	AttackVector          string  `json:"attackVector"`
	AttackComplexity      string  `json:"attackComplexity"`
	PrivilegesRequired    string  `json:"privilegesRequired"`
	UserInteraction       string  `json:"userInteraction"`
	Scope                 string  `json:"scope"`
	ConfidentialityImpact string  `json:"confidentialityImpact"`
	IntegrityImpact       string  `json:"integrityImpact"`
	AvailabilityImpact    string  `json:"availabilityImpact"`
	BaseScore             float64 `json:"baseScore"`
	BaseSeverity          string  `json:"baseSeverity"`
}

// CVSSMetricV30 CVSS v3.0评分
type CVSSMetricV30 struct {
	Source   string      `json:"source"`
	Type     string      `json:"type"`
	CvssData CVSSDataV30 `json:"cvssData"`
}

// CVSSDataV30 CVSS v3.0数据
type CVSSDataV30 struct {
	Version               string  `json:"version"`
	VectorString          string  `json:"vectorString"`
	AttackVector          string  `json:"attackVector"`
	AttackComplexity      string  `json:"attackComplexity"`
	PrivilegesRequired    string  `json:"privilegesRequired"`
	UserInteraction       string  `json:"userInteraction"`
	Scope                 string  `json:"scope"`
	ConfidentialityImpact string  `json:"confidentialityImpact"`
	IntegrityImpact       string  `json:"integrityImpact"`
	AvailabilityImpact    string  `json:"availabilityImpact"`
	BaseScore             float64 `json:"baseScore"`
	BaseSeverity          string  `json:"baseSeverity"`
}

// CVSSMetricV2 CVSS v2评分
type CVSSMetricV2 struct {
	Source   string     `json:"source"`
	Type     string     `json:"type"`
	CvssData CVSSDataV2 `json:"cvssData"`
}

// CVSSDataV2 CVSS v2数据
type CVSSDataV2 struct {
	Version               string  `json:"version"`
	VectorString          string  `json:"vectorString"`
	AccessVector          string  `json:"accessVector"`
	AccessComplexity      string  `json:"accessComplexity"`
	Authentication        string  `json:"authentication"`
	ConfidentialityImpact string  `json:"confidentialityImpact"`
	IntegrityImpact       string  `json:"integrityImpact"`
	AvailabilityImpact    string  `json:"availabilityImpact"`
	BaseScore             float64 `json:"baseScore"`
}

// CVEWeakness CVE弱点信息
type CVEWeakness struct {
	Source      string               `json:"source"`
	Type        string               `json:"type"`
	Description []CVEWeaknessDesc    `json:"description"`
}

// CVEWeaknessDesc CVE弱点描述
type CVEWeaknessDesc struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

// CVEConfigData CVE配置信息
type CVEConfigData struct {
	Nodes []CVEConfigNode `json:"nodes"`
}

// CVEConfigNode CVE配置节点
type CVEConfigNode struct {
	Operator string        `json:"operator"`
	Negate   bool          `json:"negate"`
	CpeMatch []CVECpeMatch `json:"cpeMatch"`
}

// CVECpeMatch CVE CPE匹配
type CVECpeMatch struct {
	Vulnerable            bool   `json:"vulnerable"`
	Criteria              string `json:"criteria"`
	VersionStartIncluding string `json:"versionStartIncluding,omitempty"`
	VersionEndExcluding   string `json:"versionEndExcluding,omitempty"`
	MatchCriteriaId       string `json:"matchCriteriaId"`
}

// CVEReference CVE参考信息
type CVEReference struct {
	URL    string   `json:"url"`
	Source string   `json:"source"`
	Tags   []string `json:"tags,omitempty"`
}

// FetchCVEData 获取CVE数据
func (f *CVEDataFetcher) FetchCVEData(ctx context.Context, startIndex, resultsPerPage int) (*CVEResponse, error) {
	url := fmt.Sprintf("%s?startIndex=%d&resultsPerPage=%d", f.baseURL, startIndex, resultsPerPage)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	
	// 设置API密钥
	if f.apiKey != "" {
		req.Header.Set("apiKey", f.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求CVE数据失败: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CVE API返回错误状态: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	
	var cveResponse CVEResponse
	if err := json.Unmarshal(body, &cveResponse); err != nil {
		return nil, fmt.Errorf("解析CVE数据失败: %v", err)
	}
	
	return &cveResponse, nil
}

// FetchCVEByID 根据CVE ID获取特定CVE
func (f *CVEDataFetcher) FetchCVEByID(ctx context.Context, cveID string) (*CVEVulnerability, error) {
	url := fmt.Sprintf("%s?cveId=%s", f.baseURL, cveID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	
	if f.apiKey != "" {
		req.Header.Set("apiKey", f.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求CVE数据失败: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CVE API返回错误状态: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	
	var cveResponse CVEResponse
	if err := json.Unmarshal(body, &cveResponse); err != nil {
		return nil, fmt.Errorf("解析CVE数据失败: %v", err)
	}
	
	if len(cveResponse.Vulnerabilities) == 0 {
		return nil, fmt.Errorf("CVE %s 不存在", cveID)
	}
	
	return &cveResponse.Vulnerabilities[0], nil
}

// ConvertToVulnDbInfo 将CVE数据转换为内部漏洞数据格式
func (f *CVEDataFetcher) ConvertToVulnDbInfo(cveVuln *CVEVulnerability) (*models.VulnDbInfo, error) {
	cve := cveVuln.CVE
	
	vuln := &models.VulnDbInfo{
		CVEId:  cve.ID,
		Source: "CVE",
		Status: "active",
	}
	
	// 转换发布和修改时间
	if publishedTime, err := time.Parse(time.RFC3339, cve.Published); err == nil {
		vuln.PublishedDate = publishedTime
	}
	
	if modifiedTime, err := time.Parse(time.RFC3339, cve.LastModified); err == nil {
		vuln.ModifiedDate = modifiedTime
	}
	
	// 提取描述
	for _, desc := range cve.Descriptions {
		if desc.Lang == "en" {
			vuln.Description = desc.Value
			// 生成标题（取描述的前100个字符）
			if len(desc.Value) > 100 {
				vuln.Title = desc.Value[:100] + "..."
			} else {
				vuln.Title = desc.Value
			}
			break
		}
	}
	
	// 提取CVSS评分信息
	if len(cve.Metrics.CvssMetricV31) > 0 {
		metric := cve.Metrics.CvssMetricV31[0]
		vuln.CVSSScore = metric.CvssData.BaseScore
		vuln.CVSSVector = metric.CvssData.VectorString
		vuln.CVSSVersion = "3.1"
		vuln.Severity = strings.ToLower(metric.CvssData.BaseSeverity)
	} else if len(cve.Metrics.CvssMetricV30) > 0 {
		metric := cve.Metrics.CvssMetricV30[0]
		vuln.CVSSScore = metric.CvssData.BaseScore
		vuln.CVSSVector = metric.CvssData.VectorString
		vuln.CVSSVersion = "3.0"
		vuln.Severity = strings.ToLower(metric.CvssData.BaseSeverity)
	} else if len(cve.Metrics.CvssMetricV2) > 0 {
		metric := cve.Metrics.CvssMetricV2[0]
		vuln.CVSSScore = metric.CvssData.BaseScore
		vuln.CVSSVector = metric.CvssData.VectorString
		vuln.CVSSVersion = "2.0"
		// V2没有severity字段，根据分数计算
		vuln.Severity = f.calculateSeverityFromScore(metric.CvssData.BaseScore)
	}
	
	// 提取CWE信息
	for _, weakness := range cve.Weaknesses {
		for _, desc := range weakness.Description {
			if strings.HasPrefix(desc.Value, "CWE-") {
				vuln.CWEId = desc.Value
				break
			}
		}
		if vuln.CWEId != "" {
			break
		}
	}
	
	// 提取参考信息
	for _, ref := range cve.References {
		vulnRef := models.VulnReference{
			URL:    ref.URL,
			Source: ref.Source,
			RefTags: ref.Tags,
		}
		
		// 根据标签确定类型
		for _, tag := range ref.Tags {
			switch strings.ToLower(tag) {
			case "patch":
				vulnRef.Type = "patch"
			case "exploit":
				vulnRef.Type = "exploit"
			case "vendor":
				vulnRef.Type = "vendor"
			case "advisory":
				vulnRef.Type = "advisory"
			default:
				vulnRef.Type = "report"
			}
		}
		
		vuln.References = append(vuln.References, vulnRef)
	}
	
	// 提取受影响的产品
	for _, config := range cve.Configurations {
		for _, node := range config.Nodes {
			for _, cpe := range node.CpeMatch {
				if cpe.Vulnerable {
					vuln.Affected = append(vuln.Affected, cpe.Criteria)
				}
			}
		}
	}
	
	// 设置状态
	switch cve.VulnStatus {
	case "Analyzed":
		vuln.Verified = true
	case "Awaiting Analysis":
		vuln.Verified = false
	case "Undergoing Analysis":
		vuln.Verified = false
	default:
		vuln.Verified = false
	}
	
	return vuln, nil
}

// calculateSeverityFromScore 根据CVSS分数计算严重程度
func (f *CVEDataFetcher) calculateSeverityFromScore(score float64) string {
	if score >= 9.0 {
		return "critical"
	} else if score >= 7.0 {
		return "high"
	} else if score >= 4.0 {
		return "medium"
	} else if score > 0.0 {
		return "low"
	}
	return "info"
}

// FetchRecentCVEs 获取最近的CVE数据
func (f *CVEDataFetcher) FetchRecentCVEs(ctx context.Context, days int) ([]*models.VulnDbInfo, error) {
	var allVulns []*models.VulnDbInfo
	startIndex := 0
	resultsPerPage := 100
	
	for {
		response, err := f.FetchCVEData(ctx, startIndex, resultsPerPage)
		if err != nil {
			return nil, err
		}
		
		if len(response.Vulnerabilities) == 0 {
			break
		}
		
		cutoffDate := time.Now().AddDate(0, 0, -days)
		shouldStop := false
		
		for _, cveVuln := range response.Vulnerabilities {
			// 检查是否超过时间范围
			if publishedTime, err := time.Parse(time.RFC3339, cveVuln.CVE.Published); err == nil {
				if publishedTime.Before(cutoffDate) {
					shouldStop = true
					break
				}
			}
			
			vuln, err := f.ConvertToVulnDbInfo(&cveVuln)
			if err != nil {
				log.Printf("转换CVE %s 失败: %v", cveVuln.CVE.ID, err)
				continue
			}
			
			allVulns = append(allVulns, vuln)
		}
		
		if shouldStop || len(response.Vulnerabilities) < resultsPerPage {
			break
		}
		
		startIndex += resultsPerPage
		
		// 添加延迟以避免频率限制
		time.Sleep(1 * time.Second)
	}
	
	return allVulns, nil
}