package vulndb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CNVDClient CNVD数据源客户端
type CNVDClient struct {
	config CNVDConfig
	client *http.Client
}

// NewCNVDClient 创建CNVD客户端
func NewCNVDClient(config CNVDConfig) *CNVDClient {
	return &CNVDClient{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

// FetchRecentCNVDs 获取最近的CNVD数据
func (c *CNVDClient) FetchRecentCNVDs(ctx context.Context, since time.Time) ([]CNVDData, error) {
	// TODO: 实现真实的CNVD API调用
	// 注意：CNVD的API可能需要特殊的认证和请求格式
	
	url := fmt.Sprintf("%s/api/vulns?startDate=%s", c.config.APIURL, since.Format("2006-01-02"))
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	if c.config.APIKey != "" {
		req.Header.Set("X-API-Key", c.config.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Stellar-VulnDB-Client/1.0")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CNVD API请求失败，状态码: %d", resp.StatusCode)
	}
	
	var response struct {
		Success bool         `json:"success"`
		Data    []CNVDItem   `json:"data"`
		Total   int          `json:"total"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	
	if !response.Success {
		return nil, fmt.Errorf("CNVD API返回错误")
	}
	
	// 转换为内部格式
	var cnvdData []CNVDData
	for _, item := range response.Data {
		cnvd := CNVDData{
			ID:           item.CNVDID,
			Title:        item.Title,
			Description:  item.Description,
			Severity:     item.Severity,
			CVSSScore:    item.CVSSScore,
			PublishedDate: item.PublishDate,
			ModifiedDate:  item.UpdateDate,
		}
		cnvdData = append(cnvdData, cnvd)
	}
	
	return cnvdData, nil
}

// CNVD API响应结构体
type CNVDItem struct {
	CNVDID      string    `json:"cnvd_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	CVSSScore   float64   `json:"cvss_score"`
	PublishDate time.Time `json:"publish_date"`
	UpdateDate  time.Time `json:"update_date"`
	Category    string    `json:"category"`
	Vendor      string    `json:"vendor"`
	Product     string    `json:"product"`
}