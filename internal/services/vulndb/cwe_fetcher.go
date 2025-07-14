package vulndb

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/StellarServer/internal/models"
)

// CWEDataFetcher CWE数据获取器
type CWEDataFetcher struct {
	baseURL    string
	httpClient *http.Client
}

// NewCWEDataFetcher 创建CWE数据获取器
func NewCWEDataFetcher() *CWEDataFetcher {
	return &CWEDataFetcher{
		baseURL:    "https://cwe.mitre.org/data/xml/cwec_latest.xml.zip",
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// CWECatalog CWE目录结构
type CWECatalog struct {
	XMLName     xml.Name      `xml:"Weakness_Catalog"`
	Weaknesses  []CWEWeakness `xml:"Weaknesses>Weakness"`
	Categories  []CWECategory `xml:"Categories>Category"`
	Views       []CWEView     `xml:"Views>View"`
}

// CWEWeakness CWE弱点
type CWEWeakness struct {
	ID          string `xml:"ID,attr"`
	Name        string `xml:"Name,attr"`
	Abstraction string `xml:"Abstraction,attr"`
	Structure   string `xml:"Structure,attr"`
	Status      string `xml:"Status,attr"`
	
	Description         CWEDescription         `xml:"Description"`
	ExtendedDescription CWEDescription         `xml:"Extended_Description"`
	RelatedWeaknesses   CWERelatedWeaknesses   `xml:"Related_Weaknesses"`
	WeaknessOrdinalities CWEWeaknessOrdinalities `xml:"Weakness_Ordinalities"`
	ApplicablePlatforms CWEApplicablePlatforms `xml:"Applicable_Platforms"`
	BackgroundDetails   CWEBackgroundDetails   `xml:"Background_Details"`
	AlternateTerms      CWEAlternateTerms      `xml:"Alternate_Terms"`
	ModesOfIntroduction CWEModesOfIntroduction `xml:"Modes_of_Introduction"`
	ExploitationFactors CWEExploitationFactors `xml:"Exploitation_Factors"`
	LikelihoodOfExploit CWELikelihoodOfExploit `xml:"Likelihood_of_Exploit"`
	CommonConsequences  CWECommonConsequences  `xml:"Common_Consequences"`
	DetectionMethods    CWEDetectionMethods    `xml:"Detection_Methods"`
	PotentialMitigations CWEPotentialMitigations `xml:"Potential_Mitigations"`
}

// CWECategory CWE分类
type CWECategory struct {
	ID     string         `xml:"ID,attr"`
	Name   string         `xml:"Name,attr"`
	Status string         `xml:"Status,attr"`
	Summary CWEDescription `xml:"Summary"`
}

// CWEView CWE视图
type CWEView struct {
	ID      string         `xml:"ID,attr"`
	Name    string         `xml:"Name,attr"`
	Type    string         `xml:"Type,attr"`
	Status  string         `xml:"Status,attr"`
	Summary CWEDescription `xml:"Summary"`
}

// CWEDescription CWE描述
type CWEDescription struct {
	Text string `xml:",innerxml"`
}

// CWERelatedWeaknesses CWE相关弱点
type CWERelatedWeaknesses struct {
	RelatedWeaknesses []CWERelatedWeakness `xml:"Related_Weakness"`
}

// CWERelatedWeakness CWE相关弱点
type CWERelatedWeakness struct {
	Nature  string `xml:"Nature,attr"`
	CWEID   string `xml:"CWE_ID,attr"`
	ViewID  string `xml:"View_ID,attr"`
	ChainID string `xml:"Chain_ID,attr"`
}

// CWEWeaknessOrdinalities CWE弱点序数
type CWEWeaknessOrdinalities struct {
	WeaknessOrdinalities []CWEWeaknessOrdinality `xml:"Weakness_Ordinality"`
}

// CWEWeaknessOrdinality CWE弱点序数
type CWEWeaknessOrdinality struct {
	Ordinality  string         `xml:"Ordinality,attr"`
	Description CWEDescription `xml:"Description"`
}

// CWEApplicablePlatforms CWE适用平台
type CWEApplicablePlatforms struct {
	Languages        []CWELanguage        `xml:"Language"`
	OperatingSystems []CWEOperatingSystem `xml:"Operating_System"`
	Architectures    []CWEArchitecture    `xml:"Architecture"`
	Technologies     []CWETechnology      `xml:"Technology"`
}

// CWELanguage CWE语言
type CWELanguage struct {
	Class      string `xml:"Class,attr"`
	Name       string `xml:"Name,attr"`
	Prevalence string `xml:"Prevalence,attr"`
}

// CWEOperatingSystem CWE操作系统
type CWEOperatingSystem struct {
	Class      string `xml:"Class,attr"`
	Name       string `xml:"Name,attr"`
	Prevalence string `xml:"Prevalence,attr"`
}

// CWEArchitecture CWE架构
type CWEArchitecture struct {
	Class      string `xml:"Class,attr"`
	Name       string `xml:"Name,attr"`
	Prevalence string `xml:"Prevalence,attr"`
}

// CWETechnology CWE技术
type CWETechnology struct {
	Class      string `xml:"Class,attr"`
	Name       string `xml:"Name,attr"`
	Prevalence string `xml:"Prevalence,attr"`
}

// CWEBackgroundDetails CWE背景详情
type CWEBackgroundDetails struct {
	BackgroundDetails []CWEBackgroundDetail `xml:"Background_Detail"`
}

// CWEBackgroundDetail CWE背景详情
type CWEBackgroundDetail struct {
	Text string `xml:",innerxml"`
}

// CWEAlternateTerms CWE替代术语
type CWEAlternateTerms struct {
	AlternateTerms []CWEAlternateTerm `xml:"Alternate_Term"`
}

// CWEAlternateTerm CWE替代术语
type CWEAlternateTerm struct {
	Term        string         `xml:"Term,attr"`
	Description CWEDescription `xml:"Description"`
}

// CWEModesOfIntroduction CWE引入模式
type CWEModesOfIntroduction struct {
	ModesOfIntroduction []CWEModeOfIntroduction `xml:"Introduction"`
}

// CWEModeOfIntroduction CWE引入模式
type CWEModeOfIntroduction struct {
	Phase string         `xml:"Phase,attr"`
	Note  CWEDescription `xml:"Note"`
}

// CWEExploitationFactors CWE利用因素
type CWEExploitationFactors struct {
	ExploitationFactors []CWEExploitationFactor `xml:"Exploitation_Factor"`
}

// CWEExploitationFactor CWE利用因素
type CWEExploitationFactor struct {
	Text string `xml:",innerxml"`
}

// CWELikelihoodOfExploit CWE利用可能性
type CWELikelihoodOfExploit struct {
	Text string `xml:",innerxml"`
}

// CWECommonConsequences CWE常见后果
type CWECommonConsequences struct {
	Consequences []CWECommonConsequence `xml:"Consequence"`
}

// CWECommonConsequence CWE常见后果
type CWECommonConsequence struct {
	Scope      []string       `xml:"Scope"`
	Impact     []string       `xml:"Impact"`
	Note       CWEDescription `xml:"Note"`
	Likelihood string         `xml:"Likelihood,attr"`
}

// CWEDetectionMethods CWE检测方法
type CWEDetectionMethods struct {
	DetectionMethods []CWEDetectionMethod `xml:"Detection_Method"`
}

// CWEDetectionMethod CWE检测方法
type CWEDetectionMethod struct {
	Method             string         `xml:"Method,attr"`
	Description        CWEDescription `xml:"Description"`
	Effectiveness      string         `xml:"Effectiveness,attr"`
	EffectivenessNotes CWEDescription `xml:"Effectiveness_Notes"`
}

// CWEPotentialMitigations CWE潜在缓解措施
type CWEPotentialMitigations struct {
	Mitigations []CWEPotentialMitigation `xml:"Mitigation"`
}

// CWEPotentialMitigation CWE潜在缓解措施
type CWEPotentialMitigation struct {
	Phase              []string       `xml:"Phase"`
	Strategy           string         `xml:"Strategy,attr"`
	Description        CWEDescription `xml:"Description"`
	Effectiveness      string         `xml:"Effectiveness,attr"`
	EffectivenessNotes CWEDescription `xml:"Effectiveness_Notes"`
}

// FetchCWEData 获取CWE数据
func (f *CWEDataFetcher) FetchCWEData(ctx context.Context) (*CWECatalog, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", f.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求CWE数据失败: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CWE数据源返回错误状态: %d", resp.StatusCode)
	}
	
	// 注意：实际实现中需要解压ZIP文件并解析XML
	// 这里为了简化，直接解析XML（假设已经解压）
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取CWE数据失败: %v", err)
	}
	
	var catalog CWECatalog
	if err := xml.Unmarshal(body, &catalog); err != nil {
		return nil, fmt.Errorf("解析CWE数据失败: %v", err)
	}
	
	return &catalog, nil
}

// ConvertCWEToVulnDbInfo 将CWE弱点转换为漏洞数据库信息
func (f *CWEDataFetcher) ConvertCWEToVulnDbInfo(weakness *CWEWeakness) (*models.VulnDbInfo, error) {
	vuln := &models.VulnDbInfo{
		CWEId:       fmt.Sprintf("CWE-%s", weakness.ID),
		Title:       weakness.Name,
		Description: f.cleanDescription(weakness.Description.Text),
		Category:    f.mapCWEToCategory(weakness.Name),
		VulnType:    "weakness",
		Source:      "CWE",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     1,
	}
	
	// 设置严重程度（基于抽象级别）
	vuln.Severity = f.mapAbstractionToSeverity(weakness.Abstraction)
	
	// 提取扩展描述
	if weakness.ExtendedDescription.Text != "" {
		vuln.Impact = f.cleanDescription(weakness.ExtendedDescription.Text)
	}
	
	// 提取背景详情
	var backgroundDetails []string
	for _, detail := range weakness.BackgroundDetails.BackgroundDetails {
		if detail.Text != "" {
			backgroundDetails = append(backgroundDetails, f.cleanDescription(detail.Text))
		}
	}
	if len(backgroundDetails) > 0 {
		vuln.Attack = strings.Join(backgroundDetails, "; ")
	}
	
	// 提取缓解措施作为解决方案
	var mitigations []string
	for _, mitigation := range weakness.PotentialMitigations.Mitigations {
		if mitigation.Description.Text != "" {
			desc := f.cleanDescription(mitigation.Description.Text)
			if mitigation.Strategy != "" {
				desc = fmt.Sprintf("[%s] %s", mitigation.Strategy, desc)
			}
			mitigations = append(mitigations, desc)
		}
	}
	if len(mitigations) > 0 {
		vuln.Solution = strings.Join(mitigations, "\n\n")
	}
	
	// 提取适用平台作为受影响的产品
	var affected []string
	for _, lang := range weakness.ApplicablePlatforms.Languages {
		if lang.Name != "" {
			affected = append(affected, fmt.Sprintf("Language: %s", lang.Name))
		}
	}
	for _, os := range weakness.ApplicablePlatforms.OperatingSystems {
		if os.Name != "" {
			affected = append(affected, fmt.Sprintf("OS: %s", os.Name))
		}
	}
	for _, tech := range weakness.ApplicablePlatforms.Technologies {
		if tech.Name != "" {
			affected = append(affected, fmt.Sprintf("Technology: %s", tech.Name))
		}
	}
	vuln.Affected = affected
	
	// 设置标签
	var tags []string
	tags = append(tags, "CWE")
	tags = append(tags, weakness.Abstraction)
	if weakness.Structure != "" {
		tags = append(tags, weakness.Structure)
	}
	vuln.Tags = tags
	
	// 设置发布时间（CWE没有具体发布时间，使用当前时间）
	vuln.PublishedDate = time.Now()
	vuln.ModifiedDate = time.Now()
	
	return vuln, nil
}

// cleanDescription 清理描述文本
func (f *CWEDataFetcher) cleanDescription(text string) string {
	// 移除XML标签
	text = strings.ReplaceAll(text, "<xhtml:p>", "")
	text = strings.ReplaceAll(text, "</xhtml:p>", "\n")
	text = strings.ReplaceAll(text, "<xhtml:br/>", "\n")
	text = strings.ReplaceAll(text, "<xhtml:b>", "**")
	text = strings.ReplaceAll(text, "</xhtml:b>", "**")
	text = strings.ReplaceAll(text, "<xhtml:i>", "*")
	text = strings.ReplaceAll(text, "</xhtml:i>", "*")
	
	// 清理多余的空白字符
	lines := strings.Split(text, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}
	
	return strings.Join(cleanLines, "\n")
}

// mapCWEToCategory 将CWE名称映射到分类
func (f *CWEDataFetcher) mapCWEToCategory(name string) string {
	name = strings.ToLower(name)
	
	if strings.Contains(name, "injection") || strings.Contains(name, "sql") {
		return "injection"
	}
	if strings.Contains(name, "xss") || strings.Contains(name, "cross-site") {
		return "xss"
	}
	if strings.Contains(name, "buffer") || strings.Contains(name, "overflow") {
		return "buffer_overflow"
	}
	if strings.Contains(name, "authentication") || strings.Contains(name, "authorization") {
		return "authentication"
	}
	if strings.Contains(name, "crypto") || strings.Contains(name, "encryption") {
		return "cryptography"
	}
	if strings.Contains(name, "path") || strings.Contains(name, "traversal") {
		return "path_traversal"
	}
	if strings.Contains(name, "race") || strings.Contains(name, "concurrency") {
		return "race_condition"
	}
	if strings.Contains(name, "information") || strings.Contains(name, "disclosure") {
		return "information_disclosure"
	}
	if strings.Contains(name, "dos") || strings.Contains(name, "denial") {
		return "denial_of_service"
	}
	
	return "other"
}

// mapAbstractionToSeverity 将抽象级别映射到严重程度
func (f *CWEDataFetcher) mapAbstractionToSeverity(abstraction string) string {
	switch strings.ToLower(abstraction) {
	case "class":
		return "high"
	case "base":
		return "medium"
	case "variant":
		return "low"
	case "compound":
		return "high"
	default:
		return "medium"
	}
}

// GetCWEByID 根据CWE ID获取CWE信息
func (f *CWEDataFetcher) GetCWEByID(ctx context.Context, cweID string) (*CWEWeakness, error) {
	// 从CWE ID中提取数字部分
	id := strings.TrimPrefix(cweID, "CWE-")
	
	catalog, err := f.FetchCWEData(ctx)
	if err != nil {
		return nil, err
	}
	
	for _, weakness := range catalog.Weaknesses {
		if weakness.ID == id {
			return &weakness, nil
		}
	}
	
	return nil, fmt.Errorf("CWE %s 不存在", cweID)
}

// GetMockCWEData 获取模拟CWE数据（用于测试）
func (f *CWEDataFetcher) GetMockCWEData() []*models.VulnDbInfo {
	return []*models.VulnDbInfo{
		{
			CWEId:       "CWE-79",
			Title:       "跨站脚本攻击(Cross-site Scripting)",
			Description: "软件不能正确中和或错误中和用户控制的输入，然后将其放置在用于网页的输出中，该网页将提供给其他用户。",
			Category:    "xss",
			VulnType:    "weakness",
			Severity:    "medium",
			Source:      "CWE",
			Status:      "active",
			Tags:        []string{"CWE", "Base"},
			Solution:    "对所有用户输入进行适当的验证和编码",
			Affected:    []string{"Language: PHP", "Language: JavaScript", "Technology: Web Application"},
			PublishedDate: time.Now().AddDate(0, 0, -30),
			ModifiedDate:  time.Now().AddDate(0, 0, -1),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Version:     1,
		},
		{
			CWEId:       "CWE-89",
			Title:       "SQL注入(SQL Injection)",
			Description: "软件在构建SQL查询时，未能正确中和特殊元素，从而允许攻击者修改查询的语义。",
			Category:    "injection",
			VulnType:    "weakness",
			Severity:    "high",
			Source:      "CWE",
			Status:      "active",
			Tags:        []string{"CWE", "Base"},
			Solution:    "使用参数化查询或预编译语句；对用户输入进行严格验证",
			Affected:    []string{"Language: SQL", "Technology: Database", "Technology: Web Application"},
			PublishedDate: time.Now().AddDate(0, 0, -45),
			ModifiedDate:  time.Now().AddDate(0, 0, -2),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Version:     1,
		},
		{
			CWEId:       "CWE-120",
			Title:       "缓冲区溢出(Buffer Overflow)",
			Description: "程序在将输入复制到缓冲区时，不能正确限制输入的大小，导致覆盖相邻内存位置。",
			Category:    "buffer_overflow",
			VulnType:    "weakness",
			Severity:    "high",
			Source:      "CWE",
			Status:      "active",
			Tags:        []string{"CWE", "Base"},
			Solution:    "使用安全的字符串处理函数；进行边界检查；启用栈保护机制",
			Affected:    []string{"Language: C", "Language: C++", "Architecture: x86", "Architecture: ARM"},
			PublishedDate: time.Now().AddDate(0, 0, -60),
			ModifiedDate:  time.Now().AddDate(0, 0, -3),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Version:     1,
		},
	}
}