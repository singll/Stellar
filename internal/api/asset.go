package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/utils"
)

// AssetHandler 处理资产相关的API请求
type AssetHandler struct {
	DB        *mongo.Database
	AssetRepo *models.AssetRepository
}

// NewAssetHandler 创建资产处理器
func NewAssetHandler(db *mongo.Database) *AssetHandler {
	return &AssetHandler{
		DB:        db,
		AssetRepo: models.NewAssetRepository(db),
	}
}

// RegisterRoutes 注册资产相关的路由
func (h *AssetHandler) RegisterRoutes(router *gin.RouterGroup) {
	assetGroup := router.Group("/assets")
	// 统一应用JWT认证中间件
	assetGroup.Use(AuthMiddleware())
	{
		assetGroup.POST("", h.CreateAsset)
		assetGroup.GET("", h.ListAssets)
		assetGroup.GET("/:id", h.GetAsset)
		assetGroup.PUT("/:id", h.UpdateAsset)
		assetGroup.DELETE("/:id", h.DeleteAsset)
		assetGroup.POST("/batch", h.BatchCreateAssets)
		assetGroup.DELETE("/batch", h.BatchDeleteAssets)
		assetGroup.POST("/import", h.ImportAssets)
		assetGroup.GET("/export", h.ExportAssets)
		assetGroup.GET("/relations", h.GetAssetRelations)
		assetGroup.POST("/relations", h.CreateAssetRelation)
	}
}

// CreateAssetRequest 创建资产的请求
type CreateAssetRequest struct {
	Type       models.AssetType       `json:"type" binding:"required"`
	ProjectID  string                 `json:"projectId" binding:"required"`
	RootDomain string                 `json:"rootDomain"`
	TaskName   string                 `json:"taskName"`
	Tags       []string               `json:"tags"`
	Data       map[string]interface{} `json:"data" binding:"required"`
}

// CreateAsset 创建资产
func (h *AssetHandler) CreateAsset(c *gin.Context) {
	var req CreateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err error

	projectID, err := primitive.ObjectIDFromHex(req.ProjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var project models.Project
	err = h.DB.Collection("projects").FindOne(c, bson.M{"_id": projectID}).Decode(&project)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	now := time.Now()

	switch req.Type {
	case models.AssetTypeDomain:
		asset := &models.DomainAsset{
			BaseAsset: models.BaseAsset{
				CreatedAt:    now,
				UpdatedAt:    now,
				LastScanTime: now,
				Type:         req.Type,
				ProjectID:    projectID,
				Tags:         req.Tags,
				TaskName:     req.TaskName,
				RootDomain:   req.RootDomain,
			},
		}
		if domain, ok := req.Data["domain"].(string); ok {
			asset.Domain = domain
		}
		if ips, ok := req.Data["ips"].([]interface{}); ok {
			for _, ip := range ips {
				if ipStr, ok := ip.(string); ok {
					asset.IPs = append(asset.IPs, ipStr)
				}
			}
		}
		if whois, ok := req.Data["whois"].(string); ok {
			asset.Whois = whois
		}
		if icpInfo, ok := req.Data["icpInfo"].(map[string]interface{}); ok {
			asset.ICPInfo = &models.ICPInfo{
				ICPNo:       utils.GetStringFromMap(icpInfo, "icpNo"),
				CompanyName: utils.GetStringFromMap(icpInfo, "companyName"),
				CompanyType: utils.GetStringFromMap(icpInfo, "companyType"),
				UpdateTime:  utils.GetStringFromMap(icpInfo, "updateTime"),
			}
		}
		id, err := h.AssetRepo.CreateAsset(c, req.Type, asset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		asset.ID = id
		c.JSON(http.StatusOK, gin.H{"data": asset})
		return

	case models.AssetTypeSubdomain:
		asset := &models.SubdomainAsset{
			BaseAsset: models.BaseAsset{
				CreatedAt:    now,
				UpdatedAt:    now,
				LastScanTime: now,
				Type:         req.Type,
				ProjectID:    projectID,
				Tags:         req.Tags,
				TaskName:     req.TaskName,
				RootDomain:   req.RootDomain,
			},
		}
		if host, ok := req.Data["host"].(string); ok {
			asset.Host = host
		}
		if ips, ok := req.Data["ips"].([]interface{}); ok {
			for _, ip := range ips {
				if ipStr, ok := ip.(string); ok {
					asset.IPs = append(asset.IPs, ipStr)
				}
			}
		}
		if cname, ok := req.Data["cname"].(string); ok {
			asset.CNAME = cname
		}
		if dnsType, ok := req.Data["dnsType"].(string); ok {
			asset.Type = dnsType
		}
		if value, ok := req.Data["value"].([]interface{}); ok {
			for _, v := range value {
				if vStr, ok := v.(string); ok {
					asset.Value = append(asset.Value, vStr)
				}
			}
		}
		if takeOver, ok := req.Data["takeOver"].(bool); ok {
			asset.TakeOver = takeOver
		}
		id, err := h.AssetRepo.CreateAsset(c, req.Type, asset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		asset.ID = id
		c.JSON(http.StatusOK, gin.H{"data": asset})
		return

	case models.AssetTypeIP:
		asset := &models.IPAsset{
			BaseAsset: models.BaseAsset{
				CreatedAt:    now,
				UpdatedAt:    now,
				LastScanTime: now,
				Type:         req.Type,
				ProjectID:    projectID,
				Tags:         req.Tags,
				TaskName:     req.TaskName,
				RootDomain:   req.RootDomain,
			},
		}
		if ip, ok := req.Data["ip"].(string); ok {
			asset.IP = ip
		}
		if asn, ok := req.Data["asn"].(string); ok {
			asset.ASN = asn
		}
		if isp, ok := req.Data["isp"].(string); ok {
			asset.ISP = isp
		}
		if location, ok := req.Data["location"].(map[string]interface{}); ok {
			asset.Location = &models.IPLocation{
				Country:     utils.GetStringFromMap(location, "country"),
				CountryCode: utils.GetStringFromMap(location, "countryCode"),
				Region:      utils.GetStringFromMap(location, "region"),
				City:        utils.GetStringFromMap(location, "city"),
			}
			if lat, ok := location["latitude"].(float64); ok {
				asset.Location.Latitude = lat
			}
			if lng, ok := location["longitude"].(float64); ok {
				asset.Location.Longitude = lng
			}
		}
		if fingerprint, ok := req.Data["fingerprint"].(map[string]interface{}); ok {
			asset.Fingerprint = fingerprint
		}
		id, err := h.AssetRepo.CreateAsset(c, req.Type, asset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		asset.ID = id
		c.JSON(http.StatusOK, gin.H{"data": asset})
		return

	case models.AssetTypePort:
		asset := &models.PortAsset{
			BaseAsset: models.BaseAsset{
				CreatedAt:    now,
				UpdatedAt:    now,
				LastScanTime: now,
				Type:         req.Type,
				ProjectID:    projectID,
				Tags:         req.Tags,
				TaskName:     req.TaskName,
				RootDomain:   req.RootDomain,
			},
		}
		if ip, ok := req.Data["ip"].(string); ok {
			asset.IP = ip
		}
		if port, ok := req.Data["port"].(float64); ok {
			asset.Port = int(port)
		}
		if service, ok := req.Data["service"].(string); ok {
			asset.Service = service
		}
		if protocol, ok := req.Data["protocol"].(string); ok {
			asset.Protocol = protocol
		}
		if version, ok := req.Data["version"].(string); ok {
			asset.Version = version
		}
		if banner, ok := req.Data["banner"].(string); ok {
			asset.Banner = banner
		}
		if tls, ok := req.Data["tls"].(bool); ok {
			asset.TLS = tls
		}
		if transport, ok := req.Data["transport"].(string); ok {
			asset.Transport = transport
		}
		if status, ok := req.Data["status"].(string); ok {
			asset.Status = status
		}
		id, err := h.AssetRepo.CreateAsset(c, req.Type, asset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		asset.ID = id
		c.JSON(http.StatusOK, gin.H{"data": asset})
		return

	case models.AssetTypeURL:
		asset := &models.URLAsset{
			BaseAsset: models.BaseAsset{
				CreatedAt:    now,
				UpdatedAt:    now,
				LastScanTime: now,
				Type:         req.Type,
				ProjectID:    projectID,
				Tags:         req.Tags,
				TaskName:     req.TaskName,
				RootDomain:   req.RootDomain,
			},
		}
		if url, ok := req.Data["url"].(string); ok {
			asset.URL = url
		}
		if host, ok := req.Data["host"].(string); ok {
			asset.Host = host
		}
		if path, ok := req.Data["path"].(string); ok {
			asset.Path = path
		}
		if query, ok := req.Data["query"].(string); ok {
			asset.Query = query
		}
		if fragment, ok := req.Data["fragment"].(string); ok {
			asset.Fragment = fragment
		}
		if statusCode, ok := req.Data["statusCode"].(float64); ok {
			asset.StatusCode = int(statusCode)
		}
		if title, ok := req.Data["title"].(string); ok {
			asset.Title = title
		}
		if contentType, ok := req.Data["contentType"].(string); ok {
			asset.ContentType = contentType
		}
		if contentLength, ok := req.Data["contentLength"].(float64); ok {
			asset.ContentLength = int(contentLength)
		}
		if hash, ok := req.Data["hash"].(string); ok {
			asset.Hash = hash
		}
		if screenshot, ok := req.Data["screenshot"].(string); ok {
			asset.Screenshot = screenshot
		}
		if technologies, ok := req.Data["technologies"].([]interface{}); ok {
			for _, tech := range technologies {
				if techStr, ok := tech.(string); ok {
					asset.Technologies = append(asset.Technologies, techStr)
				}
			}
		}
		if headers, ok := req.Data["headers"].(map[string]interface{}); ok {
			asset.Headers = make(map[string]string)
			for k, v := range headers {
				if vStr, ok := v.(string); ok {
					asset.Headers[k] = vStr
				}
			}
		}
		if favicon, ok := req.Data["favicon"].(map[string]interface{}); ok {
			asset.Favicon = &models.FaviconInfo{
				Path:    utils.GetStringFromMap(favicon, "path"),
				MMH3:    utils.GetStringFromMap(favicon, "mmh3"),
				Content: utils.GetStringFromMap(favicon, "content"),
			}
		}
		if metadata, ok := req.Data["metadata"].(map[string]interface{}); ok {
			asset.Metadata = metadata
		}
		id, err := h.AssetRepo.CreateAsset(c, req.Type, asset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		asset.ID = id
		c.JSON(http.StatusOK, gin.H{"data": asset})
		return

	case models.AssetTypeHTTP:
		asset := &models.HTTPAsset{
			BaseAsset: models.BaseAsset{
				CreatedAt:    now,
				UpdatedAt:    now,
				LastScanTime: now,
				Type:         req.Type,
				ProjectID:    projectID,
				Tags:         req.Tags,
				TaskName:     req.TaskName,
				RootDomain:   req.RootDomain,
			},
		}
		if host, ok := req.Data["host"].(string); ok {
			asset.Host = host
		}
		if ip, ok := req.Data["ip"].(string); ok {
			asset.IP = ip
		}
		if port, ok := req.Data["port"].(float64); ok {
			asset.Port = int(port)
		}
		if url, ok := req.Data["url"].(string); ok {
			asset.URL = url
		}
		if title, ok := req.Data["title"].(string); ok {
			asset.Title = title
		}
		if statusCode, ok := req.Data["statusCode"].(float64); ok {
			asset.StatusCode = int(statusCode)
		}
		if contentType, ok := req.Data["contentType"].(string); ok {
			asset.ContentType = contentType
		}
		if contentLength, ok := req.Data["contentLength"].(float64); ok {
			asset.ContentLength = int(contentLength)
		}
		if webServer, ok := req.Data["webServer"].(string); ok {
			asset.WebServer = webServer
		}
		if tls, ok := req.Data["tls"].(bool); ok {
			asset.TLS = tls
		}
		if hash, ok := req.Data["hash"].(string); ok {
			asset.Hash = hash
		}
		if cdnName, ok := req.Data["cdnName"].(string); ok {
			asset.CDNName = cdnName
		}
		if cdn, ok := req.Data["cdn"].(bool); ok {
			asset.CDN = cdn
		}
		if screenshot, ok := req.Data["screenshot"].(string); ok {
			asset.Screenshot = screenshot
		}
		if technologies, ok := req.Data["technologies"].([]interface{}); ok {
			for _, tech := range technologies {
				if techStr, ok := tech.(string); ok {
					asset.Technologies = append(asset.Technologies, techStr)
				}
			}
		}
		if headers, ok := req.Data["headers"].(map[string]interface{}); ok {
			asset.Headers = make(map[string]string)
			for k, v := range headers {
				if vStr, ok := v.(string); ok {
					asset.Headers[k] = vStr
				}
			}
		}
		if favicon, ok := req.Data["favicon"].(map[string]interface{}); ok {
			asset.Favicon = &models.FaviconInfo{
				Path:    utils.GetStringFromMap(favicon, "path"),
				MMH3:    utils.GetStringFromMap(favicon, "mmh3"),
				Content: utils.GetStringFromMap(favicon, "content"),
			}
		}
		if jarm, ok := req.Data["jarm"].(string); ok {
			asset.JARM = jarm
		}
		if metadata, ok := req.Data["metadata"].(map[string]interface{}); ok {
			asset.Metadata = metadata
		}
		id, err := h.AssetRepo.CreateAsset(c, req.Type, asset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		asset.ID = id
		c.JSON(http.StatusOK, gin.H{"data": asset})
		return

	case models.AssetTypeApp:
		asset := &models.AppAsset{
			BaseAsset: models.BaseAsset{
				CreatedAt:    now,
				UpdatedAt:    now,
				LastScanTime: now,
				Type:         req.Type,
				ProjectID:    projectID,
				Tags:         req.Tags,
				TaskName:     req.TaskName,
				RootDomain:   req.RootDomain,
			},
		}
		if appName, ok := req.Data["appName"].(string); ok {
			asset.AppName = appName
		}
		if packageName, ok := req.Data["packageName"].(string); ok {
			asset.PackageName = packageName
		}
		if platform, ok := req.Data["platform"].(string); ok {
			asset.Platform = platform
		}
		if version, ok := req.Data["version"].(string); ok {
			asset.Version = version
		}
		if developer, ok := req.Data["developer"].(string); ok {
			asset.Developer = developer
		}
		if downloadURL, ok := req.Data["downloadUrl"].(string); ok {
			asset.DownloadURL = downloadURL
		}
		if description, ok := req.Data["description"].(string); ok {
			asset.Description = description
		}
		if permissions, ok := req.Data["permissions"].([]interface{}); ok {
			for _, perm := range permissions {
				if permStr, ok := perm.(string); ok {
					asset.Permissions = append(asset.Permissions, permStr)
				}
			}
		}
		if sha256, ok := req.Data["sha256"].(string); ok {
			asset.SHA256 = sha256
		}
		if iconURL, ok := req.Data["iconUrl"].(string); ok {
			asset.IconURL = iconURL
		}
		if metadata, ok := req.Data["metadata"].(map[string]interface{}); ok {
			asset.Metadata = metadata
		}
		id, err := h.AssetRepo.CreateAsset(c, req.Type, asset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		asset.ID = id
		c.JSON(http.StatusOK, gin.H{"data": asset})
		return

	case models.AssetTypeMiniApp:
		asset := &models.MiniAppAsset{
			BaseAsset: models.BaseAsset{
				CreatedAt:    now,
				UpdatedAt:    now,
				LastScanTime: now,
				Type:         req.Type,
				ProjectID:    projectID,
				Tags:         req.Tags,
				TaskName:     req.TaskName,
				RootDomain:   req.RootDomain,
			},
		}

		// 填充特定字段
		if appName, ok := req.Data["appName"].(string); ok {
			asset.AppName = appName
		}
		if appID, ok := req.Data["appId"].(string); ok {
			asset.AppID = appID
		}
		if platform, ok := req.Data["platform"].(string); ok {
			asset.Platform = platform
		}
		if developer, ok := req.Data["developer"].(string); ok {
			asset.Developer = developer
		}
		if description, ok := req.Data["description"].(string); ok {
			asset.Description = description
		}
		if iconURL, ok := req.Data["iconUrl"].(string); ok {
			asset.IconURL = iconURL
		}
		if qrCodeURL, ok := req.Data["qrCodeUrl"].(string); ok {
			asset.QRCodeURL = qrCodeURL
		}
		if metadata, ok := req.Data["metadata"].(map[string]interface{}); ok {
			asset.Metadata = metadata
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported asset type"})
		return
	}
}

// ListAssetsRequest 列出资产的请求
type ListAssetsRequest struct {
	ProjectID  string           `form:"projectId"`
	Type       models.AssetType `form:"type"`
	RootDomain string           `form:"rootDomain"`
	Tags       []string         `form:"tags"`
	Page       int              `form:"page,default=1"`
	PageSize   int              `form:"pageSize,default=20"`
	SortBy     string           `form:"sortBy,default=createdAt"`
	SortOrder  int              `form:"sortOrder,default=-1"` // 1: 升序, -1: 降序
	Query      string           `form:"query"`                // 全文搜索
}

// ListAssets 列出资产
func (h *AssetHandler) ListAssets(c *gin.Context) {
	var req ListAssetsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 构建查询条件
	filter := bson.M{}

	// 项目ID过滤
	if req.ProjectID != "" {
		projectID, err := primitive.ObjectIDFromHex(req.ProjectID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}
		filter["projectId"] = projectID
	}

	// 资产类型过滤
	if req.Type != "" {
		filter["type"] = req.Type
	}

	// 根域名过滤
	if req.RootDomain != "" {
		filter["rootDomain"] = req.RootDomain
	}

	// 标签过滤
	if len(req.Tags) > 0 {
		filter["tags"] = bson.M{"$in": req.Tags}
	}

	// 全文搜索
	if req.Query != "" {
		filter["$text"] = bson.M{"$search": req.Query}
	}

	// 分页
	skip := (req.Page - 1) * req.PageSize
	limit := req.PageSize

	// 排序
	sort := bson.D{{req.SortBy, req.SortOrder}}

	// 查询选项
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(sort)

	// 确定集合名称
	collectionName := models.AssetCollection(req.Type)
	if req.Type == "" {
		// 如果未指定类型，默认查询所有类型的资产
		// 这里需要实现跨集合查询，或者选择一个默认集合
		collectionName = "domain_assets" // 默认查询域名资产
	}

	// 执行查询
	cursor, err := h.DB.Collection(collectionName).Find(c, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(c)

	// 获取总数
	total, err := h.DB.Collection(collectionName).CountDocuments(c, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 解析结果
	var assets []map[string]interface{}
	if err := cursor.All(c, &assets); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  req.Page,
		"size":  req.PageSize,
		"data":  assets,
	})
}

// GetAsset 获取单个资产
func (h *AssetHandler) GetAsset(c *gin.Context) {
	id := c.Param("id")
	assetType := c.Query("type")

	// 验证ID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset ID"})
		return
	}

	// 验证资产类型
	if assetType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Asset type is required"})
		return
	}

	// 确定集合名称
	collectionName := models.AssetCollection(models.AssetType(assetType))

	// 查询资产
	var asset map[string]interface{}
	err = h.DB.Collection(collectionName).FindOne(c, bson.M{"_id": objectID}).Decode(&asset)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, asset)
}

// UpdateAsset 更新资产
func (h *AssetHandler) UpdateAsset(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Type models.AssetType       `json:"type" binding:"required"`
		Data map[string]interface{} `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证ID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset ID"})
		return
	}

	// 确定集合名称
	collectionName := models.AssetCollection(req.Type)

	// 获取当前资产
	var currentAsset map[string]interface{}
	err = h.DB.Collection(collectionName).FindOne(c, bson.M{"_id": objectID}).Decode(&currentAsset)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 准备更新数据
	update := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	// 添加其他需要更新的字段
	for k, v := range req.Data {
		if k != "_id" && k != "createdAt" && k != "type" { // 保护某些字段不被更新
			update["$set"].(bson.M)[k] = v
		}
	}

	// 执行更新
	_, err = h.DB.Collection(collectionName).UpdateOne(
		c,
		bson.M{"_id": objectID},
		update,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Asset updated successfully"})
}

// DeleteAsset 删除资产
func (h *AssetHandler) DeleteAsset(c *gin.Context) {
	id := c.Param("id")
	assetType := c.Query("type")

	// 验证ID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset ID"})
		return
	}

	// 验证资产类型
	if assetType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Asset type is required"})
		return
	}

	// 确定集合名称
	collectionName := models.AssetCollection(models.AssetType(assetType))

	// 删除资产
	result, err := h.DB.Collection(collectionName).DeleteOne(c, bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
		return
	}

	// 删除相关的资产关系
	_, err = h.DB.Collection("asset_relations").DeleteMany(c, bson.M{
		"$or": []bson.M{
			{"sourceAssetId": objectID},
			{"targetAssetId": objectID},
		},
	})

	if err != nil {
		// 仅记录错误，不影响主要操作的结果
		// TODO: 记录日志
	}

	c.JSON(http.StatusOK, gin.H{"message": "Asset deleted successfully"})
}

// BatchCreateAssets 批量创建资产
func (h *AssetHandler) BatchCreateAssets(c *gin.Context) {
	var req struct {
		Type       models.AssetType         `json:"type" binding:"required"`
		ProjectID  string                   `json:"projectId" binding:"required"`
		RootDomain string                   `json:"rootDomain"`
		TaskName   string                   `json:"taskName"`
		Tags       []string                 `json:"tags"`
		Assets     []map[string]interface{} `json:"assets" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证项目ID
	projectID, err := primitive.ObjectIDFromHex(req.ProjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// 检查项目是否存在
	var project models.Project
	err = h.DB.Collection("projects").FindOne(c, bson.M{"_id": projectID}).Decode(&project)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// 确定集合名称
	collectionName := models.AssetCollection(req.Type)

	// 准备批量插入的文档
	now := time.Now()
	var documents []interface{}

	for _, assetData := range req.Assets {
		// 根据资产类型创建资产
		// 这里简化处理，实际应该根据不同类型创建不同的资产对象
		asset := bson.M{
			"createdAt":    now,
			"updatedAt":    now,
			"lastScanTime": now,
			"type":         req.Type,
			"projectId":    projectID,
			"tags":         req.Tags,
			"taskName":     req.TaskName,
			"rootDomain":   req.RootDomain,
		}

		// 合并资产特定数据
		for k, v := range assetData {
			if k != "_id" && k != "createdAt" && k != "updatedAt" && k != "type" && k != "projectId" {
				asset[k] = v
			}
		}

		documents = append(documents, asset)
	}

	// 执行批量插入
	result, err := h.DB.Collection(collectionName).InsertMany(c, documents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"insertedCount": len(result.InsertedIDs),
		"insertedIds":   result.InsertedIDs,
	})
}

// BatchDeleteAssets 批量删除资产
func (h *AssetHandler) BatchDeleteAssets(c *gin.Context) {
	var req struct {
		Type models.AssetType `json:"type" binding:"required"`
		IDs  []string         `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转换ID字符串为ObjectID
	var objectIDs []primitive.ObjectID
	for _, id := range req.IDs {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset ID: " + id})
			return
		}
		objectIDs = append(objectIDs, objectID)
	}

	// 确定集合名称
	collectionName := models.AssetCollection(req.Type)

	// 批量删除资产
	result, err := h.DB.Collection(collectionName).DeleteMany(c, bson.M{
		"_id": bson.M{"$in": objectIDs},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 删除相关的资产关系
	_, err = h.DB.Collection("asset_relations").DeleteMany(c, bson.M{
		"$or": []bson.M{
			{"sourceAssetId": bson.M{"$in": objectIDs}},
			{"targetAssetId": bson.M{"$in": objectIDs}},
		},
	})

	if err != nil {
		// 仅记录错误，不影响主要操作的结果
		// TODO: 记录日志
	}

	c.JSON(http.StatusOK, gin.H{
		"deletedCount": result.DeletedCount,
	})
}

// ImportAssets 导入资产
func (h *AssetHandler) ImportAssets(c *gin.Context) {
	// 获取文件和参数
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	projectID := c.PostForm("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	assetType := c.PostForm("type")
	if assetType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Asset type is required"})
		return
	}

	// 验证项目ID
	objProjectID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// 检查项目是否存在
	var project models.Project
	err = h.DB.Collection("projects").FindOne(c, bson.M{"_id": objProjectID}).Decode(&project)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// 打开文件
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer f.Close()

	// 根据文件类型处理导入
	fileExt := strings.ToLower(filepath.Ext(file.Filename))
	var assets []map[string]interface{}

	switch fileExt {
	case ".csv":
		// 处理CSV文件
		reader := csv.NewReader(f)
		records, err := reader.ReadAll()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV file"})
			return
		}

		// 确保CSV有数据
		if len(records) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "CSV file has no data or headers"})
			return
		}

		// 获取CSV头部
		headers := records[0]

		// 将CSV数据转换为资产数据
		for i := 1; i < len(records); i++ {
			asset := make(map[string]interface{})
			for j, value := range records[i] {
				if j < len(headers) {
					asset[headers[j]] = value
				}
			}
			assets = append(assets, asset)
		}

	case ".json":
		// 处理JSON文件
		var jsonData []map[string]interface{}
		decoder := json.NewDecoder(f)
		if err := decoder.Decode(&jsonData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON file"})
			return
		}
		assets = jsonData

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported file format. Please use CSV or JSON."})
		return
	}

	// 准备批量插入的文档
	now := time.Now()
	var documents []interface{}
	collectionName := models.AssetCollection(models.AssetType(assetType))

	for _, assetData := range assets {
		// 创建基础资产数据
		asset := bson.M{
			"createdAt":    now,
			"updatedAt":    now,
			"lastScanTime": now,
			"type":         assetType,
			"projectId":    objProjectID,
			"tags":         []string{},
			"rootDomain":   project.Name, // 默认使用项目名称作为根域名
		}

		// 合并资产特定数据
		for k, v := range assetData {
			if k != "_id" && k != "createdAt" && k != "updatedAt" && k != "type" && k != "projectId" {
				asset[k] = v
			}
		}

		documents = append(documents, asset)
	}

	// 执行批量插入
	if len(documents) > 0 {
		result, err := h.DB.Collection(collectionName).InsertMany(c, documents)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":       true,
			"importedCount": len(result.InsertedIDs),
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid assets to import"})
	}
}

// ExportAssets 导出资产
func (h *AssetHandler) ExportAssets(c *gin.Context) {
	// 获取查询参数
	projectID := c.Query("projectId")
	assetType := c.Query("type")
	format := c.DefaultQuery("format", "json") // 默认导出为JSON格式
	filename := c.DefaultQuery("filename", "assets-export-"+time.Now().Format("20060102150405"))

	// 验证必要参数
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	if assetType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Asset type is required"})
		return
	}

	// 验证项目ID
	objProjectID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// 构建查询条件
	filter := bson.M{"projectId": objProjectID}

	// 确定集合名称
	collectionName := models.AssetCollection(models.AssetType(assetType))

	// 执行查询
	cursor, err := h.DB.Collection(collectionName).Find(c, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(c)

	// 解析结果
	var assets []map[string]interface{}
	if err := cursor.All(c, &assets); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 根据格式导出
	switch strings.ToLower(format) {
	case "csv":
		// 导出为CSV
		if len(assets) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No assets found"})
			return
		}

		// 创建CSV写入器
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", filename))

		writer := csv.NewWriter(c.Writer)
		defer writer.Flush()

		// 提取所有可能的列名
		columnSet := make(map[string]bool)
		for _, asset := range assets {
			for k := range asset {
				columnSet[k] = true
			}
		}

		// 排序列名以确保一致性
		columns := make([]string, 0, len(columnSet))
		for col := range columnSet {
			columns = append(columns, col)
		}
		sort.Strings(columns)

		// 写入CSV头部
		if err := writer.Write(columns); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV header"})
			return
		}

		// 写入CSV数据
		for _, asset := range assets {
			row := make([]string, len(columns))
			for i, col := range columns {
				if val, ok := asset[col]; ok {
					row[i] = fmt.Sprintf("%v", val)
				}
			}
			if err := writer.Write(row); err != nil {
				// 记录错误但继续处理
				log.Printf("Error writing CSV row: %v", err)
			}
		}

	case "json":
		// 导出为JSON
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.json", filename))
		c.JSON(http.StatusOK, assets)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported export format. Please use CSV or JSON."})
	}
}

// CreateAssetRelation 创建资产关系
func (h *AssetHandler) CreateAssetRelation(c *gin.Context) {
	var req struct {
		SourceAssetID string `json:"sourceAssetId" binding:"required"`
		TargetAssetID string `json:"targetAssetId" binding:"required"`
		RelationType  string `json:"relationType" binding:"required"`
		ProjectID     string `json:"projectId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证项目ID
	objProjectID, err := primitive.ObjectIDFromHex(req.ProjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// 检查项目是否存在
	var project models.Project
	err = h.DB.Collection("projects").FindOne(c, bson.M{"_id": objProjectID}).Decode(&project)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// 验证资产ID
	sourceAssetID, err := primitive.ObjectIDFromHex(req.SourceAssetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source asset ID"})
		return
	}

	targetAssetID, err := primitive.ObjectIDFromHex(req.TargetAssetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid target asset ID"})
		return
	}

	// 创建资产关系
	relation := bson.M{
		"sourceAssetId": sourceAssetID,
		"targetAssetId": targetAssetID,
		"relationType":  req.RelationType,
		"projectId":     objProjectID,
	}

	// 插入关系
	_, err = h.DB.Collection("asset_relations").InsertOne(c, relation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Asset relation created successfully"})
}

// GetAssetRelations 获取资产关系
func (h *AssetHandler) GetAssetRelations(c *gin.Context) {
	// 获取查询参数
	projectID := c.Query("projectId")
	assetType := c.Query("type")

	// 验证项目ID
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	// 验证资产类型
	if assetType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Asset type is required"})
		return
	}

	// 验证资产ID
	objectID, err := primitive.ObjectIDFromHex(assetType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset ID"})
		return
	}

	// 构建查询条件
	filter := bson.M{
		"projectId": projectID,
		"$or": []bson.M{
			{"sourceAssetId": objectID},
			{"targetAssetId": objectID},
		},
	}

	// 执行查询
	cursor, err := h.DB.Collection("asset_relations").Find(c, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(c)

	// 解析结果
	var relations []map[string]interface{}
	if err := cursor.All(c, &relations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, relations)
}
