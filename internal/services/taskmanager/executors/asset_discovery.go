package executors

import (
	"context"
	"fmt"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AssetDiscoveryExecutor 资产发现执行器
type AssetDiscoveryExecutor struct {
	db     *mongo.Database
	config AssetDiscoveryConfig
}

// AssetDiscoveryConfig 资产发现配置
type AssetDiscoveryConfig struct {
	EnableDomainAssets    bool `json:"enable_domain_assets"`
	EnableSubdomainAssets bool `json:"enable_subdomain_assets"`
	EnableIPAssets        bool `json:"enable_ip_assets"`
	EnablePortAssets      bool `json:"enable_port_assets"`
	EnableURLAssets       bool `json:"enable_url_assets"`
	AutoCreateAssets      bool `json:"auto_create_assets"`
}

// AssetDiscoveryResult 资产发现结果
type AssetDiscoveryResult struct {
	CreatedAssets []CreatedAsset `json:"created_assets"`
	UpdatedAssets []UpdatedAsset `json:"updated_assets"`
	SkippedAssets []SkippedAsset `json:"skipped_assets"`
	ErrorAssets   []ErrorAsset   `json:"error_assets"`
	Summary       DiscoverySummary `json:"summary"`
}

// CreatedAsset 创建的资产
type CreatedAsset struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Source string `json:"source"`
}

// UpdatedAsset 更新的资产
type UpdatedAsset struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Changes  []string `json:"changes"`
	Source   string `json:"source"`
}

// SkippedAsset 跳过的资产
type SkippedAsset struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// ErrorAsset 错误的资产
type ErrorAsset struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Error  string `json:"error"`
}

// DiscoverySummary 发现总结
type DiscoverySummary struct {
	TotalProcessed int `json:"total_processed"`
	CreatedCount   int `json:"created_count"`
	UpdatedCount   int `json:"updated_count"`
	SkippedCount   int `json:"skipped_count"`
	ErrorCount     int `json:"error_count"`
}

// NewAssetDiscoveryExecutor 创建资产发现执行器
func NewAssetDiscoveryExecutor(db *mongo.Database, config AssetDiscoveryConfig) *AssetDiscoveryExecutor {
	// 设置默认配置
	if !config.EnableDomainAssets && !config.EnableSubdomainAssets && 
		!config.EnableIPAssets && !config.EnablePortAssets && !config.EnableURLAssets {
		config.EnableDomainAssets = true
		config.EnableSubdomainAssets = true
		config.EnableIPAssets = true
		config.EnablePortAssets = true
		config.EnableURLAssets = true
	}

	return &AssetDiscoveryExecutor{
		db:     db,
		config: config,
	}
}

// Execute 执行资产发现任务
func (e *AssetDiscoveryExecutor) Execute(ctx context.Context, task *models.Task) (*models.TaskResult, error) {
	// 创建结果对象
	result := &models.TaskResult{
		ID:        primitive.NewObjectID(),
		TaskID:    task.ID,
		Status:    "running",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
		Data:      make(map[string]interface{}),
	}

	// 获取项目ID
	projectID, ok := task.Config["project_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid project_id")
	}

	// 获取源任务结果
	sourceResults, err := e.getSourceTaskResults(ctx, task)
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		return result, err
	}

	// 执行资产发现
	discoveryResult, err := e.discoverAssets(ctx, projectID, sourceResults, task)
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		return result, err
	}

	// 处理结果
	result.Status = "completed"
	result.Data["discovery_result"] = discoveryResult
	result.Data["project_id"] = projectID
	result.Summary = fmt.Sprintf("Created %d assets, updated %d assets, skipped %d assets",
		discoveryResult.Summary.CreatedCount,
		discoveryResult.Summary.UpdatedCount,
		discoveryResult.Summary.SkippedCount)

	return result, nil
}

// GetSupportedTypes 获取支持的任务类型
func (e *AssetDiscoveryExecutor) GetSupportedTypes() []string {
	return []string{"asset_discovery"}
}

// GetExecutorInfo 获取执行器信息
func (e *AssetDiscoveryExecutor) GetExecutorInfo() models.ExecutorInfo {
	return models.ExecutorInfo{
		Name:        "AssetDiscoveryExecutor",
		Version:     "1.0.0",
		Description: "Asset discovery executor that converts scan results into assets",
		Author:      "Stellar Team",
	}
}

// getSourceTaskResults 获取源任务结果
func (e *AssetDiscoveryExecutor) getSourceTaskResults(ctx context.Context, task *models.Task) (map[string]interface{}, error) {
	// 获取源任务ID
	sourceTaskID, ok := task.Config["source_task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid source_task_id")
	}

	// 查询源任务结果
	taskObjID, err := primitive.ObjectIDFromHex(sourceTaskID)
	if err != nil {
		return nil, fmt.Errorf("invalid source_task_id format: %v", err)
	}

	collection := e.db.Collection("task_results")
	var taskResult models.TaskResult
	err = collection.FindOne(ctx, map[string]interface{}{
		"task_id": taskObjID,
		"status":  "completed",
	}).Decode(&taskResult)
	if err != nil {
		return nil, fmt.Errorf("failed to find source task result: %v", err)
	}

	return taskResult.Data, nil
}

// discoverAssets 发现资产
func (e *AssetDiscoveryExecutor) discoverAssets(ctx context.Context, projectID string, sourceResults map[string]interface{}, task *models.Task) (*AssetDiscoveryResult, error) {
	result := &AssetDiscoveryResult{
		CreatedAssets: make([]CreatedAsset, 0),
		UpdatedAssets: make([]UpdatedAsset, 0),
		SkippedAssets: make([]SkippedAsset, 0),
		ErrorAssets:   make([]ErrorAsset, 0),
	}

	// 获取任务类型以确定如何处理结果
	sourceTaskType, ok := task.Config["source_task_type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid source_task_type")
	}

	// 根据源任务类型处理不同的结果
	switch sourceTaskType {
	case "subdomain_enum":
		if err := e.processSubdomainResults(ctx, projectID, sourceResults, result); err != nil {
			return nil, err
		}
	case "port_scan":
		if err := e.processPortScanResults(ctx, projectID, sourceResults, result); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported source task type: %s", sourceTaskType)
	}

	// 更新总结
	result.Summary = DiscoverySummary{
		TotalProcessed: len(result.CreatedAssets) + len(result.UpdatedAssets) + len(result.SkippedAssets) + len(result.ErrorAssets),
		CreatedCount:   len(result.CreatedAssets),
		UpdatedCount:   len(result.UpdatedAssets),
		SkippedCount:   len(result.SkippedAssets),
		ErrorCount:     len(result.ErrorAssets),
	}

	return result, nil
}

// processSubdomainResults 处理子域名枚举结果
func (e *AssetDiscoveryExecutor) processSubdomainResults(ctx context.Context, projectID string, sourceResults map[string]interface{}, result *AssetDiscoveryResult) error {
	// 获取子域名结果
	subdomainsData, ok := sourceResults["subdomains"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid subdomains data format")
	}

	// 获取根域名
	rootDomain, _ := sourceResults["target"].(string)

	// 处理每个子域名
	for _, subdomainData := range subdomainsData {
		subdomainMap, ok := subdomainData.(map[string]interface{})
		if !ok {
			continue
		}

		subdomain, _ := subdomainMap["subdomain"].(string)
		if subdomain == "" {
			continue
		}

		// 创建子域名资产
		if e.config.EnableSubdomainAssets {
			asset, err := e.createSubdomainAsset(ctx, projectID, rootDomain, subdomainMap)
			if err != nil {
				result.ErrorAssets = append(result.ErrorAssets, ErrorAsset{
					Type:  "subdomain",
					Name:  subdomain,
					Error: err.Error(),
				})
				continue
			}

			result.CreatedAssets = append(result.CreatedAssets, CreatedAsset{
				ID:     asset.ID.Hex(),
				Type:   "subdomain",
				Name:   subdomain,
				Source: "subdomain_enum",
			})
		}

		// 创建IP资产
		if e.config.EnableIPAssets {
			if ips, ok := subdomainMap["ips"].([]interface{}); ok {
				for _, ipData := range ips {
					if ip, ok := ipData.(string); ok && ip != "" {
						asset, err := e.createIPAsset(ctx, projectID, ip, rootDomain)
						if err != nil {
							result.ErrorAssets = append(result.ErrorAssets, ErrorAsset{
								Type:  "ip",
								Name:  ip,
								Error: err.Error(),
							})
							continue
						}

						result.CreatedAssets = append(result.CreatedAssets, CreatedAsset{
							ID:     asset.ID.Hex(),
							Type:   "ip",
							Name:   ip,
							Source: "subdomain_enum",
						})
					}
				}
			}
		}
	}

	return nil
}

// processPortScanResults 处理端口扫描结果
func (e *AssetDiscoveryExecutor) processPortScanResults(ctx context.Context, projectID string, sourceResults map[string]interface{}, result *AssetDiscoveryResult) error {
	// 获取开放端口结果
	openPortsData, ok := sourceResults["open_ports"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid open_ports data format")
	}

	// 处理每个开放端口
	for _, portData := range openPortsData {
		portMap, ok := portData.(map[string]interface{})
		if !ok {
			continue
		}

		host, _ := portMap["host"].(string)
		port, _ := portMap["port"].(float64)
		service, _ := portMap["service"].(string)

		if host == "" || port == 0 {
			continue
		}

		// 创建端口资产
		if e.config.EnablePortAssets {
			asset, err := e.createPortAsset(ctx, projectID, host, int(port), service, portMap)
			if err != nil {
				result.ErrorAssets = append(result.ErrorAssets, ErrorAsset{
					Type:  "port",
					Name:  fmt.Sprintf("%s:%d", host, int(port)),
					Error: err.Error(),
				})
				continue
			}

			result.CreatedAssets = append(result.CreatedAssets, CreatedAsset{
				ID:     asset.ID.Hex(),
				Type:   "port",
				Name:   fmt.Sprintf("%s:%d", host, int(port)),
				Source: "port_scan",
			})
		}

		// 如果端口是web服务，创建URL资产
		if e.config.EnableURLAssets && (service == "http" || service == "https" || int(port) == 80 || int(port) == 443 || int(port) == 8080 || int(port) == 8443) {
			protocol := "http"
			if service == "https" || int(port) == 443 || int(port) == 8443 {
				protocol = "https"
			}

			url := fmt.Sprintf("%s://%s:%d", protocol, host, int(port))
			asset, err := e.createURLAsset(ctx, projectID, url, host, int(port))
			if err != nil {
				result.ErrorAssets = append(result.ErrorAssets, ErrorAsset{
					Type:  "url",
					Name:  url,
					Error: err.Error(),
				})
				continue
			}

			result.CreatedAssets = append(result.CreatedAssets, CreatedAsset{
				ID:     asset.ID.Hex(),
				Type:   "url",
				Name:   url,
				Source: "port_scan",
			})
		}
	}

	return nil
}

// createSubdomainAsset 创建子域名资产
func (e *AssetDiscoveryExecutor) createSubdomainAsset(ctx context.Context, projectID, rootDomain string, subdomainData map[string]interface{}) (*models.SubdomainAsset, error) {
	subdomain, _ := subdomainData["subdomain"].(string)
	cname, _ := subdomainData["cname"].(string)
	
	// 解析IP地址
	var ips []string
	if ipsData, ok := subdomainData["ips"].([]interface{}); ok {
		for _, ipData := range ipsData {
			if ip, ok := ipData.(string); ok {
				ips = append(ips, ip)
			}
		}
	}

	// 转换projectID字符串为ObjectID
	projectObjID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %v", err)
	}

	asset := &models.SubdomainAsset{
		BaseAsset: models.BaseAsset{
			ID:           primitive.NewObjectID(),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			LastScanTime: time.Now(),
			Type:         "subdomain",
			ProjectID:    projectObjID,
			Tags:         []string{"auto-discovered"},
			RootDomain:   rootDomain,
		},
		Host:  subdomain,
		IPs:   ips,
		CNAME: cname,
	}

	// 保存到数据库
	collection := e.db.Collection("assets")
	_, err = collection.InsertOne(ctx, asset)
	if err != nil {
		return nil, err
	}

	return asset, nil
}

// createIPAsset 创建IP资产
func (e *AssetDiscoveryExecutor) createIPAsset(ctx context.Context, projectID, ip, rootDomain string) (*models.IPAsset, error) {
	// 转换projectID字符串为ObjectID
	projectObjID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %v", err)
	}

	asset := &models.IPAsset{
		BaseAsset: models.BaseAsset{
			ID:           primitive.NewObjectID(),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			LastScanTime: time.Now(),
			Type:         "ip",
			ProjectID:    projectObjID,
			Tags:         []string{"auto-discovered"},
			RootDomain:   rootDomain,
		},
		IP: ip,
	}

	// 保存到数据库
	collection := e.db.Collection("assets")
	_, err = collection.InsertOne(ctx, asset)
	if err != nil {
		return nil, err
	}

	return asset, nil
}

// createPortAsset 创建端口资产
func (e *AssetDiscoveryExecutor) createPortAsset(ctx context.Context, projectID, host string, port int, service string, portData map[string]interface{}) (*models.PortAsset, error) {
	banner, _ := portData["banner"].(string)
	protocol, _ := portData["protocol"].(string)
	status, _ := portData["status"].(string)

	// 转换projectID字符串为ObjectID
	projectObjID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %v", err)
	}

	asset := &models.PortAsset{
		BaseAsset: models.BaseAsset{
			ID:           primitive.NewObjectID(),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			LastScanTime: time.Now(),
			Type:         "port",
			ProjectID:    projectObjID,
			Tags:         []string{"auto-discovered"},
		},
		IP:       host,
		Port:     port,
		Service:  service,
		Protocol: protocol,
		Banner:   banner,
		Status:   status,
	}

	// 保存到数据库
	collection := e.db.Collection("assets")
	_, err = collection.InsertOne(ctx, asset)
	if err != nil {
		return nil, err
	}

	return asset, nil
}

// createURLAsset 创建URL资产
func (e *AssetDiscoveryExecutor) createURLAsset(ctx context.Context, projectID, url, host string, port int) (*models.URLAsset, error) {
	// 转换projectID字符串为ObjectID
	projectObjID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %v", err)
	}

	asset := &models.URLAsset{
		BaseAsset: models.BaseAsset{
			ID:           primitive.NewObjectID(),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			LastScanTime: time.Now(),
			Type:         "url",
			ProjectID:    projectObjID,
			Tags:         []string{"auto-discovered"},
		},
		URL:  url,
		Host: host,
	}

	// 保存到数据库
	collection := e.db.Collection("assets")
	_, err = collection.InsertOne(ctx, asset)
	if err != nil {
		return nil, err
	}

	return asset, nil
}