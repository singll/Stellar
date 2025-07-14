package vulndb

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
)

// CWEClient CWE数据源客户端
type CWEClient struct {
	config CWEConfig
	client *http.Client
}

// NewCWEClient 创建CWE客户端
func NewCWEClient(config CWEConfig) *CWEClient {
	return &CWEClient{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

// FetchCWEData 获取CWE数据
func (c *CWEClient) FetchCWEData(ctx context.Context) ([]CWEData, error) {
	// TODO: 实现真实的CWE XML解析
	// CWE数据通常以XML格式提供
	
	req, err := http.NewRequestWithContext(ctx, "GET", c.config.XMLURL, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CWE数据请求失败，状态码: %d", resp.StatusCode)
	}
	
	var cweList CWEList
	if err := xml.NewDecoder(resp.Body).Decode(&cweList); err != nil {
		return nil, err
	}
	
	// 转换为内部格式
	var cweData []CWEData
	for _, weakness := range cweList.Weaknesses {
		cwe := CWEData{
			ID:           weakness.ID,
			Name:         weakness.Name,
			Description:  weakness.Description,
			Category:     weakness.Category,
			WeaknessType: weakness.WeaknessType,
		}
		cweData = append(cweData, cwe)
	}
	
	return cweData, nil
}

// CWE XML结构体
type CWEList struct {
	XMLName    xml.Name `xml:"Weakness_Catalog"`
	Weaknesses []struct {
		ID           string `xml:"ID,attr"`
		Name         string `xml:"Name,attr"`
		Description  string `xml:"Description"`
		Category     string `xml:"Category"`
		WeaknessType string `xml:"Type,attr"`
	} `xml:"Weaknesses>Weakness"`
}