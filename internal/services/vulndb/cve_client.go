package vulndb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CVEClient CVE数据源客户端
type CVEClient struct {
	config CVEConfig
	client *http.Client
}

// NewCVEClient 创建CVE客户端
func NewCVEClient(config CVEConfig) *CVEClient {
	return &CVEClient{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

// FetchRecentCVEs 获取最近的CVE数据
func (c *CVEClient) FetchRecentCVEs(ctx context.Context, since time.Time) ([]CVEData, error) {
	// TODO: 实现真实的CVE API调用
	// 这里提供一个模拟实现
	
	url := fmt.Sprintf("%s/cves/2.0?modStartDate=%s", c.config.APIURL, since.Format("2006-01-02T15:04:05"))
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CVE API请求失败，状态码: %d", resp.StatusCode)
	}
	
	var response struct {
		TotalResults int       `json:"totalResults"`
		CVEItems     []CVEItem `json:"vulnerabilities"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	
	// 转换为内部格式
	var cveData []CVEData
	for _, item := range response.CVEItems {
		cve := CVEData{
			ID:               item.CVE.ID,
			SourceIdentifier: "nvd@nist.gov",
			Published:        item.PublishedDate.Format(time.RFC3339),
			LastModified:     item.LastModifiedDate.Format(time.RFC3339),
			VulnStatus:       "Analyzed",
			Descriptions: []CVEDescription{
				{
					Lang:  "en",
					Value: item.CVE.Description,
				},
			},
		}
		cveData = append(cveData, cve)
	}
	
	return cveData, nil
}

// CVE API响应结构体
type CVEItem struct {
	CVE struct {
		ID          string `json:"id"`
		Description string `json:"description"`
	} `json:"cve"`
	Impact struct {
		BaseMetricV3 struct {
			CVSSV3 struct {
				BaseScore    float64 `json:"baseScore"`
				BaseSeverity string  `json:"baseSeverity"`
				VectorString string  `json:"vectorString"`
			} `json:"cvssV3"`
		} `json:"baseMetricV3"`
	} `json:"impact"`
	PublishedDate    time.Time `json:"publishedDate"`
	LastModifiedDate time.Time `json:"lastModifiedDate"`
}

func extractSeverity(impact interface{}) string {
	// TODO: 从impact中提取严重程度
	return "medium"
}

func extractCVSSScore(impact interface{}) float64 {
	// TODO: 从impact中提取CVSS分数
	return 5.0
}

func extractCVSSVector(impact interface{}) string {
	// TODO: 从impact中提取CVSS向量
	return ""
}