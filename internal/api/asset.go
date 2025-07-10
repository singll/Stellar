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
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
			"details": err.Error(),
		})
		return
	}

	// 验证项目ID
	projectID, err := primitive.ObjectIDFromHex(req.ProjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目ID",
		})
		return
	}

	// 验证项目是否存在
	var project models.Project
	err = h.DB.Collection("projects").FindOne(c, bson.M{"_id": projectID}).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "项目不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询项目失败",
				"details": err.Error(),
			})
		}
		return
	}

	now := time.Now()

	// 创建基础资产
	baseAsset := models.BaseAsset{
		CreatedAt:    now,
		UpdatedAt:    now,
		LastScanTime: now,
		Type:         req.Type,
		ProjectID:    projectID,
		Tags:         req.Tags,
		TaskName:     req.TaskName,
		RootDomain:   req.RootDomain,
	}

	// 根据类型创建具体资产
	var asset interface{}
	var id primitive.ObjectID

	switch req.Type {
	case models.AssetTypeDomain:
		domainAsset := &models.DomainAsset{
			BaseAsset: baseAsset,
		}
		// 填充域名特定字段
		if domain, ok := req.Data["domain"].(string); ok {
			domainAsset.Domain = domain
		}
		if ips, ok := req.Data["ips"].([]interface{}); ok {
			for _, ip := range ips {
				if ipStr, ok := ip.(string); ok {
					domainAsset.IPs = append(domainAsset.IPs, ipStr)
				}
			}
		}
		if whois, ok := req.Data["whois"].(string); ok {
			domainAsset.Whois = whois
		}
		if icpInfo, ok := req.Data["icpInfo"].(map[string]interface{}); ok {
			domainAsset.ICPInfo = &models.ICPInfo{
				ICPNo:       utils.GetStringFromMap(icpInfo, "icpNo"),
				CompanyName: utils.GetStringFromMap(icpInfo, "companyName"),
				CompanyType: utils.GetStringFromMap(icpInfo, "companyType"),
				UpdateTime:  utils.GetStringFromMap(icpInfo, "updateTime"),
			}
		}
		asset = domainAsset

	case models.AssetTypeSubdomain:
		subdomainAsset := &models.SubdomainAsset{
			BaseAsset: baseAsset,
		}
		// 填充子域名特定字段
		if host, ok := req.Data["host"].(string); ok {
			subdomainAsset.Host = host
		}
		if ips, ok := req.Data["ips"].([]interface{}); ok {
			for _, ip := range ips {
				if ipStr, ok := ip.(string); ok {
					subdomainAsset.IPs = append(subdomainAsset.IPs, ipStr)
				}
			}
		}
		if cname, ok := req.Data["cname"].(string); ok {
			subdomainAsset.CNAME = cname
		}
		if dnsType, ok := req.Data["dnsType"].(string); ok {
			subdomainAsset.Type = dnsType
		}
		if value, ok := req.Data["value"].([]interface{}); ok {
			for _, v := range value {
				if vStr, ok := v.(string); ok {
					subdomainAsset.Value = append(subdomainAsset.Value, vStr)
				}
			}
		}
		if takeOver, ok := req.Data["takeOver"].(bool); ok {
			subdomainAsset.TakeOver = takeOver
		}
		asset = subdomainAsset

	// 其他资产类型处理...
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的资产类型",
		})
		return
	}

	// 创建资产
	id, err = h.AssetRepo.CreateAsset(c, req.Type, asset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建资产失败",
			"details": err.Error(),
		})
		return
	}

	// 设置ID并返回
	switch a := asset.(type) {
	case *models.DomainAsset:
		a.ID = id
	case *models.SubdomainAsset:
		a.ID = id
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": 201,
		"message": "资产创建成功",
		"data": asset,
	})
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
	var req struct {
		ProjectID  string           `form:"projectId"`
		Type       models.AssetType `form:"type"`
		RootDomain string           `form:"rootDomain"`
		Tags       []string         `form:"tags"`
		TaskName   string           `form:"taskName"`
		Search     string           `form:"search"`
		Page       int              `form:"page,default=1"`
		PageSize   int              `form:"pageSize,default=20"`
		SortBy     string           `form:"sortBy,default=createdAt"`
		SortDesc   bool             `form:"sortDesc,default=true"`
		StartTime  string           `form:"startTime"`
		EndTime    string           `form:"endTime"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的查询参数",
			"details": err.Error(),
		})
		return
	}

	// 验证资产类型
	if req.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "资产类型不能为空",
		})
		return
	}

	// 构建筛选条件
	filter := &models.AssetFilter{}

	// 项目ID筛选
	if req.ProjectID != "" {
		projectID, err := primitive.ObjectIDFromHex(req.ProjectID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的项目ID",
			})
			return
		}
		filter.ProjectID = &projectID
	}

	// 其他筛选条件
	if req.RootDomain != "" {
		filter.RootDomain = &req.RootDomain
	}
	if req.TaskName != "" {
		filter.TaskName = &req.TaskName
	}
	if len(req.Tags) > 0 {
		filter.Tags = req.Tags
	}
	if req.Search != "" {
		filter.Search = &req.Search
	}

	// 时间范围筛选
	if req.StartTime != "" || req.EndTime != "" {
		dateRange := &models.DateRange{}
		if req.StartTime != "" {
			startTime, err := time.Parse(time.RFC3339, req.StartTime)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    400,
					"message": "无效的开始时间格式",
				})
				return
			}
			dateRange.StartTime = &startTime
		}
		if req.EndTime != "" {
			endTime, err := time.Parse(time.RFC3339, req.EndTime)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    400,
					"message": "无效的结束时间格式",
				})
				return
			}
			dateRange.EndTime = &endTime
		}
		filter.DateRange = dateRange
	}

	// 构建查询选项
	opts := models.AssetListOptions{
		Filter:   filter,
		Page:     req.Page,
		PageSize: req.PageSize,
		SortBy:   req.SortBy,
		SortDesc: req.SortDesc,
	}

	// 使用Repository查询
	result, err := h.AssetRepo.ListAssets(c, req.Type, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询资产失败",
			"details": err.Error(),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": result,
	})
}

// GetAsset 获取单个资产
func (h *AssetHandler) GetAsset(c *gin.Context) {
	id := c.Param("id")
	assetType := c.Query("type")

	// 验证ID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的资产ID",
			"details": err.Error(),
		})
		return
	}

	// 验证资产类型
	if assetType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "资产类型不能为空",
		})
		return
	}

	// 使用Repository查询资产
	var asset interface{}
	err = h.AssetRepo.FindAssetByID(c, models.AssetType(assetType), objectID, &asset)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "资产不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询资产失败",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": asset,
	})
}

// UpdateAsset 更新资产
func (h *AssetHandler) UpdateAsset(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Type models.AssetType       `json:"type" binding:"required"`
		Data map[string]interface{} `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
			"details": err.Error(),
		})
		return
	}

	// 验证ID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的资产ID",
			"details": err.Error(),
		})
		return
	}

	// 验证资产是否存在
	var existingAsset interface{}
	err = h.AssetRepo.FindAssetByID(c, req.Type, objectID, &existingAsset)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "资产不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询资产失败",
				"details": err.Error(),
			})
		}
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
		if k != "_id" && k != "createdAt" && k != "type" && k != "projectId" { // 保护某些字段不被更新
			update["$set"].(bson.M)[k] = v
		}
	}

	// 使用Repository更新资产
	err = h.AssetRepo.UpdateAsset(c, req.Type, objectID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新资产失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "资产更新成功",
	})
}

// DeleteAsset 删除资产
func (h *AssetHandler) DeleteAsset(c *gin.Context) {
	id := c.Param("id")
	assetType := c.Query("type")

	// 验证ID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的资产ID",
			"details": err.Error(),
		})
		return
	}

	// 验证资产类型
	if assetType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "资产类型不能为空",
		})
		return
	}

	// 验证资产是否存在
	var existingAsset interface{}
	err = h.AssetRepo.FindAssetByID(c, models.AssetType(assetType), objectID, &existingAsset)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "资产不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询资产失败",
				"details": err.Error(),
			})
		}
		return
	}

	// 使用Repository删除资产
	err = h.AssetRepo.DeleteAsset(c, models.AssetType(assetType), objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除资产失败",
			"details": err.Error(),
		})
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
		log.Printf("删除资产关系失败: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "资产删除成功",
	})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
			"details": err.Error(),
		})
		return
	}

	// 验证项目ID
	projectID, err := primitive.ObjectIDFromHex(req.ProjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目ID",
			"details": err.Error(),
		})
		return
	}

	// 检查项目是否存在
	var project models.Project
	err = h.DB.Collection("projects").FindOne(c, bson.M{"_id": projectID}).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "项目不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询项目失败",
				"details": err.Error(),
			})
		}
		return
	}

	// 准备批量插入的文档
	now := time.Now()
	var documents []interface{}

	for _, assetData := range req.Assets {
		// 创建基础资产数据
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

	// 使用Repository执行批量插入
	insertedIDs, err := h.AssetRepo.BatchCreateAssets(c, req.Type, documents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "批量创建资产失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": 201,
		"message": "批量创建资产成功",
		"data": gin.H{
			"insertedCount": len(insertedIDs),
			"insertedIds":   insertedIDs,
		},
	})
}

// BatchDeleteAssets 批量删除资产
func (h *AssetHandler) BatchDeleteAssets(c *gin.Context) {
	var req struct {
		Type models.AssetType `json:"type" binding:"required"`
		IDs  []string         `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
			"details": err.Error(),
		})
		return
	}

	// 转换ID字符串为ObjectID
	var objectIDs []primitive.ObjectID
	for _, id := range req.IDs {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的资产ID",
				"details": fmt.Sprintf("ID %s 格式错误: %v", id, err),
			})
			return
		}
		objectIDs = append(objectIDs, objectID)
	}

	// 使用Repository批量删除资产
	deletedCount, err := h.AssetRepo.BatchDeleteAssets(c, req.Type, objectIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "批量删除资产失败",
			"details": err.Error(),
		})
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
		log.Printf("删除资产关系失败: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "批量删除资产成功",
		"data": gin.H{
			"deletedCount": deletedCount,
		},
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
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
			"details": err.Error(),
		})
		return
	}

	// 验证项目ID
	objProjectID, err := primitive.ObjectIDFromHex(req.ProjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的项目ID",
			"details": err.Error(),
		})
		return
	}

	// 检查项目是否存在
	var project models.Project
	err = h.DB.Collection("projects").FindOne(c, bson.M{"_id": objProjectID}).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "项目不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询项目失败",
				"details": err.Error(),
			})
		}
		return
	}

	// 验证资产ID
	sourceAssetID, err := primitive.ObjectIDFromHex(req.SourceAssetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的源资产ID",
			"details": err.Error(),
		})
		return
	}

	targetAssetID, err := primitive.ObjectIDFromHex(req.TargetAssetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的目标资产ID",
			"details": err.Error(),
		})
		return
	}

	// 创建资产关系
	relation := models.AssetRelation{
		SourceAssetID: sourceAssetID,
		TargetAssetID: targetAssetID,
		RelationType:  req.RelationType,
		ProjectID:     objProjectID,
	}

	// 使用Repository插入关系
	relationID, err := h.AssetRepo.CreateAssetRelation(c, relation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建资产关系失败",
			"details": err.Error(),
		})
		return
	}

	// 设置ID并返回
	relation.ID = relationID
	c.JSON(http.StatusCreated, gin.H{
		"code": 201,
		"message": "资产关系创建成功",
		"data": relation,
	})
}

// GetAssetRelations 获取资产关系
func (h *AssetHandler) GetAssetRelations(c *gin.Context) {
	// 获取查询参数
	projectID := c.Query("projectId")
	assetID := c.Query("assetId")

	// 验证项目ID
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "项目ID不能为空",
		})
		return
	}

	// 验证资产ID
	if assetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "资产ID不能为空",
		})
		return
	}

	// 验证资产ID格式
	objectID, err := primitive.ObjectIDFromHex(assetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的资产ID",
			"details": err.Error(),
		})
		return
	}

	// 使用Repository查询关系
	relations, err := h.AssetRepo.FindAssetRelations(c, objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询资产关系失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": relations,
	})
}
