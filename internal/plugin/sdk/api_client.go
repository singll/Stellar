package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIClient Stellar API客户端
type APIClient struct {
	baseURL    string
	apiKey     string
	headers    map[string]string
	timeout    time.Duration
	httpClient *http.Client
}

// NewAPIClient 创建API客户端
func NewAPIClient(baseURL, apiKey string, timeout time.Duration) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		timeout: timeout,
		headers: make(map[string]string),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// SetHeader 设置请求头
func (c *APIClient) SetHeader(key, value string) {
	c.headers[key] = value
}

// Get 发送GET请求
func (c *APIClient) Get(endpoint string, headers map[string]string) ([]byte, error) {
	return c.request("GET", endpoint, nil, headers)
}

// Post 发送POST请求
func (c *APIClient) Post(endpoint string, data []byte, headers map[string]string) ([]byte, error) {
	return c.request("POST", endpoint, data, headers)
}

// Put 发送PUT请求
func (c *APIClient) Put(endpoint string, data []byte, headers map[string]string) ([]byte, error) {
	return c.request("PUT", endpoint, data, headers)
}

// Delete 发送DELETE请求
func (c *APIClient) Delete(endpoint string, headers map[string]string) ([]byte, error) {
	return c.request("DELETE", endpoint, nil, headers)
}

// request 通用请求方法
func (c *APIClient) request(method, endpoint string, data []byte, headers map[string]string) ([]byte, error) {
	url := c.baseURL + endpoint
	
	var body io.Reader
	if data != nil {
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置默认头部
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Stellar-Plugin-SDK/1.0")
	
	// 设置API密钥
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// 设置全局头部
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// 设置请求特定头部
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API请求失败: %d %s", resp.StatusCode, string(respData))
	}

	return respData, nil
}

// Stellar业务API封装

// Asset 资产信息
type Asset struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	ProjectID   string    `json:"project_id"`
	Tags        []string  `json:"tags"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Task 任务信息
type Task struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status"`
	ProjectID string                 `json:"project_id"`
	Config    map[string]interface{} `json:"config"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// ScanResult 扫描结果
type ScanResult struct {
	ID       string                 `json:"id"`
	TaskID   string                 `json:"task_id"`
	Target   string                 `json:"target"`
	Type     string                 `json:"type"`
	Status   string                 `json:"status"`
	Data     map[string]interface{} `json:"data"`
	CreateAt time.Time              `json:"created_at"`
}

// GetAssets 获取资产列表
func (c *APIClient) GetAssets(projectID string) ([]*Asset, error) {
	endpoint := "/api/v1/assets"
	if projectID != "" {
		endpoint += "?project_id=" + projectID
	}

	data, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Assets []*Asset `json:"assets"`
	}
	
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("解析资产列表失败: %v", err)
	}

	return response.Assets, nil
}

// CreateAsset 创建资产
func (c *APIClient) CreateAsset(asset *Asset) (*Asset, error) {
	data, err := json.Marshal(asset)
	if err != nil {
		return nil, fmt.Errorf("序列化资产数据失败: %v", err)
	}

	respData, err := c.Post("/api/v1/assets", data, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Asset *Asset `json:"asset"`
	}
	
	if err := json.Unmarshal(respData, &response); err != nil {
		return nil, fmt.Errorf("解析创建资产响应失败: %v", err)
	}

	return response.Asset, nil
}

// GetTasks 获取任务列表
func (c *APIClient) GetTasks(projectID string) ([]*Task, error) {
	endpoint := "/api/v1/tasks"
	if projectID != "" {
		endpoint += "?project_id=" + projectID
	}

	data, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Tasks []*Task `json:"tasks"`
	}
	
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("解析任务列表失败: %v", err)
	}

	return response.Tasks, nil
}

// CreateTask 创建任务
func (c *APIClient) CreateTask(task *Task) (*Task, error) {
	data, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("序列化任务数据失败: %v", err)
	}

	respData, err := c.Post("/api/v1/tasks", data, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Task *Task `json:"task"`
	}
	
	if err := json.Unmarshal(respData, &response); err != nil {
		return nil, fmt.Errorf("解析创建任务响应失败: %v", err)
	}

	return response.Task, nil
}

// GetTaskStatus 获取任务状态
func (c *APIClient) GetTaskStatus(taskID string) (string, error) {
	data, err := c.Get("/api/v1/tasks/"+taskID+"/status", nil)
	if err != nil {
		return "", err
	}

	var response struct {
		Status string `json:"status"`
	}
	
	if err := json.Unmarshal(data, &response); err != nil {
		return "", fmt.Errorf("解析任务状态失败: %v", err)
	}

	return response.Status, nil
}

// SubmitScanResult 提交扫描结果
func (c *APIClient) SubmitScanResult(result *ScanResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("序列化扫描结果失败: %v", err)
	}

	_, err = c.Post("/api/v1/scan-results", data, nil)
	if err != nil {
		return fmt.Errorf("提交扫描结果失败: %v", err)
	}

	return nil
}

// GetScanResults 获取扫描结果
func (c *APIClient) GetScanResults(taskID string) ([]*ScanResult, error) {
	endpoint := "/api/v1/scan-results"
	if taskID != "" {
		endpoint += "?task_id=" + taskID
	}

	data, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Results []*ScanResult `json:"results"`
	}
	
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("解析扫描结果失败: %v", err)
	}

	return response.Results, nil
}

// SendNotification 发送通知
func (c *APIClient) SendNotification(title, message, level string) error {
	notification := map[string]interface{}{
		"title":   title,
		"message": message,
		"level":   level,
		"source":  "plugin",
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("序列化通知数据失败: %v", err)
	}

	_, err = c.Post("/api/v1/notifications", data, nil)
	if err != nil {
		return fmt.Errorf("发送通知失败: %v", err)
	}

	return nil
}

// UpdateProgress 更新任务进度
func (c *APIClient) UpdateProgress(taskID string, progress float64, message string) error {
	progressData := map[string]interface{}{
		"progress": progress,
		"message":  message,
	}

	data, err := json.Marshal(progressData)
	if err != nil {
		return fmt.Errorf("序列化进度数据失败: %v", err)
	}

	_, err = c.Put("/api/v1/tasks/"+taskID+"/progress", data, nil)
	if err != nil {
		return fmt.Errorf("更新任务进度失败: %v", err)
	}

	return nil
}

// UploadFile 上传文件
func (c *APIClient) UploadFile(filename string, content []byte) (string, error) {
	// 创建多部分表单数据
	var buffer bytes.Buffer
	// 这里简化实现，实际应该使用multipart/form-data
	
	// 对于简化，我们将文件内容作为JSON上传
	fileData := map[string]interface{}{
		"filename": filename,
		"content":  content,
	}

	data, err := json.Marshal(fileData)
	if err != nil {
		return "", fmt.Errorf("序列化文件数据失败: %v", err)
	}

	respData, err := c.Post("/api/v1/files/upload", data, nil)
	if err != nil {
		return "", err
	}

	var response struct {
		FileID string `json:"file_id"`
		URL    string `json:"url"`
	}
	
	if err := json.Unmarshal(respData, &response); err != nil {
		return "", fmt.Errorf("解析上传响应失败: %v", err)
	}

	return response.URL, nil
}

// DownloadFile 下载文件
func (c *APIClient) DownloadFile(fileID string) ([]byte, error) {
	return c.Get("/api/v1/files/"+fileID, nil)
}

// GetSystemInfo 获取系统信息
func (c *APIClient) GetSystemInfo() (map[string]interface{}, error) {
	data, err := c.Get("/api/v1/system/info", nil)
	if err != nil {
		return nil, err
	}

	var systemInfo map[string]interface{}
	if err := json.Unmarshal(data, &systemInfo); err != nil {
		return nil, fmt.Errorf("解析系统信息失败: %v", err)
	}

	return systemInfo, nil
}

// Ping 检查API连接
func (c *APIClient) Ping() error {
	_, err := c.Get("/api/v1/ping", nil)
	return err
}