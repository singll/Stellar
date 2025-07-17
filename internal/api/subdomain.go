package api

import (
	"bufio"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/pkg/logger"
	"github.com/StellarServer/internal/services/subdomain"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SubdomainHandler 处理子域名枚举相关的API请求
type SubdomainHandler struct {
	DB *mongo.Database
}

// NewSubdomainHandler 创建子域名处理器
func NewSubdomainHandler(db *mongo.Database) *SubdomainHandler {
	return &SubdomainHandler{DB: db}
}

// RegisterRoutes 注册子域名枚举相关的路由
func (h *SubdomainHandler) RegisterRoutes(router *gin.RouterGroup) {
	// 无需 router.Use(AuthMiddleware())，统一由路由分组控制
	router.POST("/tasks", h.CreateEnumTask)
	router.GET("/tasks", h.ListEnumTasks)
	router.GET("/tasks/:id", h.GetEnumTask)
	router.DELETE("/tasks/:id", h.DeleteEnumTask)
	router.GET("/results", h.ListEnumResults)
	router.GET("/results/:id", h.GetEnumResult)
	router.POST("/dictionaries", h.UploadDictionary)
	router.GET("/dictionaries", h.ListDictionaries)
	router.DELETE("/dictionaries/:id", h.DeleteDictionary)
}

// CreateEnumTaskRequest 创建子域名枚举任务的请求
type CreateEnumTaskRequest struct {
	ProjectID  string                     `json:"projectId" binding:"required"`
	RootDomain string                     `json:"rootDomain" binding:"required"`
	TaskName   string                     `json:"taskName"`
	Config     models.SubdomainEnumConfig `json:"config"`
	Tags       []string                   `json:"tags"`
}

// CreateEnumTask 创建子域名枚举任务
func (h *SubdomainHandler) CreateEnumTask(c *gin.Context) {
	var req CreateEnumTaskRequest
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

	// 设置默认配置
	if req.Config.DictionaryPath == "" {
		req.Config.DictionaryPath = "dicts/subdomain_dict.txt" // 默认字典路径
	}
	if len(req.Config.Methods) == 0 {
		req.Config.Methods = []string{"dns_brute"} // 默认使用DNS爆破
	}
	if req.Config.Concurrency <= 0 {
		req.Config.Concurrency = 50 // 默认并发数
	}
	if req.Config.Timeout <= 0 {
		req.Config.Timeout = 5 // 默认超时时间
	}
	if req.Config.RetryCount <= 0 {
		req.Config.RetryCount = 3 // 默认重试次数
	}
	if req.Config.RateLimit <= 0 {
		req.Config.RateLimit = 100 // 默认速率限制
	}

	// 创建任务
	now := time.Now()
	task := models.SubdomainEnumTask{
		ID:         primitive.NewObjectID(),
		ProjectID:  projectID,
		RootDomain: req.RootDomain,
		TaskName:   req.TaskName,
		Status:     "pending",
		CreatedAt:  now,
		Config:     req.Config,
		Tags:       req.Tags,
		ResultSummary: models.SubdomainEnumSummary{
			MethodStats: make(map[string]int),
		},
	}

	// 保存任务到数据库
	_, err = h.DB.Collection("subdomain_enum_tasks").InsertOne(c, task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	// 启动任务（在实际应用中，这应该由任务调度系统处理）
	go h.runEnumTask(task)

	c.JSON(http.StatusCreated, gin.H{
		"id":     task.ID.Hex(),
		"status": task.Status,
	})
}

// ListEnumTasks 列出子域名枚举任务
func (h *SubdomainHandler) ListEnumTasks(c *gin.Context) {
	// 获取查询参数
	projectID := c.Query("projectId")
	status := c.Query("status")
	page, _ := c.GetQuery("page")
	pageSize, _ := c.GetQuery("pageSize")

	// 构建查询条件
	filter := bson.M{}
	if projectID != "" {
		objID, err := primitive.ObjectIDFromHex(projectID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}
		filter["projectId"] = objID
	}
	if status != "" {
		filter["status"] = status
	}

	// 设置分页
	options := options.Find()
	options.SetSort(bson.M{"createdAt": -1})
	if page != "" && pageSize != "" {
		var pageNum, pageSizeNum int64
		pageNum, _ = strconv.ParseInt(page, 10, 64)
		pageSizeNum, _ = strconv.ParseInt(pageSize, 10, 64)
		if pageNum > 0 && pageSizeNum > 0 {
			options.SetSkip((pageNum - 1) * pageSizeNum)
			options.SetLimit(pageSizeNum)
		}
	}

	// 查询数据库
	cursor, err := h.DB.Collection("subdomain_enum_tasks").Find(c, filter, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query tasks"})
		return
	}
	defer cursor.Close(c)

	// 解析结果
	var tasks []models.SubdomainEnumTask
	if err := cursor.All(c, &tasks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse tasks"})
		return
	}

	// 获取总数
	count, err := h.DB.Collection("subdomain_enum_tasks").CountDocuments(c, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"total": count,
	})
}

// GetEnumTask 获取子域名枚举任务详情
func (h *SubdomainHandler) GetEnumTask(c *gin.Context) {
	id := c.Param("id")
	taskID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// 查询任务
	var task models.SubdomainEnumTask
	err = h.DB.Collection("subdomain_enum_tasks").FindOne(c, bson.M{"_id": taskID}).Decode(&task)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteEnumTask 删除子域名枚举任务
func (h *SubdomainHandler) DeleteEnumTask(c *gin.Context) {
	id := c.Param("id")
	taskID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// 删除任务
	result, err := h.DB.Collection("subdomain_enum_tasks").DeleteOne(c, bson.M{"_id": taskID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// 删除相关的结果
	_, _ = h.DB.Collection("subdomain_results").DeleteMany(c, bson.M{"taskId": taskID})

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// ListEnumResults 列出子域名枚举结果
func (h *SubdomainHandler) ListEnumResults(c *gin.Context) {
	// 获取查询参数
	taskID := c.Query("taskId")
	projectID := c.Query("projectId")
	rootDomain := c.Query("rootDomain")
	subdomain := c.Query("subdomain")
	isResolved := c.Query("isResolved")
	isWildcard := c.Query("isWildcard")
	isTakeOver := c.Query("isTakeOver")
	page, _ := c.GetQuery("page")
	pageSize, _ := c.GetQuery("pageSize")

	// 构建查询条件
	filter := bson.M{}
	if taskID != "" {
		objID, err := primitive.ObjectIDFromHex(taskID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
			return
		}
		filter["taskId"] = objID
	}
	if projectID != "" {
		objID, err := primitive.ObjectIDFromHex(projectID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}
		filter["projectId"] = objID
	}
	if rootDomain != "" {
		filter["rootDomain"] = rootDomain
	}
	if subdomain != "" {
		filter["subdomain"] = bson.M{"$regex": subdomain, "$options": "i"}
	}
	if isResolved != "" {
		filter["isResolved"] = isResolved == "true"
	}
	if isWildcard != "" {
		filter["isWildcard"] = isWildcard == "true"
	}
	if isTakeOver != "" {
		filter["isTakeOver"] = isTakeOver == "true"
	}

	// 设置分页
	options := options.Find()
	options.SetSort(bson.M{"createdAt": -1})
	if page != "" && pageSize != "" {
		var pageNum, pageSizeNum int64
		pageNum, _ = strconv.ParseInt(page, 10, 64)
		pageSizeNum, _ = strconv.ParseInt(pageSize, 10, 64)
		if pageNum > 0 && pageSizeNum > 0 {
			options.SetSkip((pageNum - 1) * pageSizeNum)
			options.SetLimit(pageSizeNum)
		}
	}

	// 查询数据库
	cursor, err := h.DB.Collection("subdomain_results").Find(c, filter, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query results"})
		return
	}
	defer cursor.Close(c)

	// 解析结果
	var results []models.SubdomainResult
	if err := cursor.All(c, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse results"})
		return
	}

	// 获取总数
	count, err := h.DB.Collection("subdomain_results").CountDocuments(c, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count results"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"total":   count,
	})
}

// GetEnumResult 获取子域名枚举结果详情
func (h *SubdomainHandler) GetEnumResult(c *gin.Context) {
	id := c.Param("id")
	resultID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid result ID"})
		return
	}

	// 查询结果
	var result models.SubdomainResult
	err = h.DB.Collection("subdomain_results").FindOne(c, bson.M{"_id": resultID}).Decode(&result)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Result not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UploadDictionary 上传子域名字典
func (h *SubdomainHandler) UploadDictionary(c *gin.Context) {
	// 获取文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 获取参数
	name := c.PostForm("name")
	description := c.PostForm("description")
	isDefault := c.PostForm("isDefault") == "true"
	tags := c.PostFormArray("tags")

	if name == "" {
		name = file.Filename
	}

	// 保存文件
	dictionaryDir := "dicts"
	if _, err := os.Stat(dictionaryDir); os.IsNotExist(err) {
		os.MkdirAll(dictionaryDir, 0755)
	}

	filename := filepath.Join(dictionaryDir, name)
	if err := c.SaveUploadedFile(file, filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// 计算字典中的行数
	f, err := os.Open(filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			count++
		}
	}

	// 创建字典记录
	dictionary := models.SubdomainDictionary{
		ID:          primitive.NewObjectID(),
		Name:        name,
		Description: description,
		Count:       count,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		FilePath:    filename,
		IsDefault:   isDefault,
		Tags:        tags,
	}

	// 保存到数据库
	_, err = h.DB.Collection("subdomain_dictionaries").InsertOne(c, dictionary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save dictionary"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    dictionary.ID.Hex(),
		"name":  dictionary.Name,
		"count": dictionary.Count,
	})
}

// ListDictionaries 列出子域名字典
func (h *SubdomainHandler) ListDictionaries(c *gin.Context) {
	// 查询数据库
	cursor, err := h.DB.Collection("subdomain_dictionaries").Find(c, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query dictionaries"})
		return
	}
	defer cursor.Close(c)

	// 解析结果
	var dictionaries []models.SubdomainDictionary
	if err := cursor.All(c, &dictionaries); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse dictionaries"})
		return
	}

	c.JSON(http.StatusOK, dictionaries)
}

// DeleteDictionary 删除子域名字典
func (h *SubdomainHandler) DeleteDictionary(c *gin.Context) {
	id := c.Param("id")
	dictionaryID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dictionary ID"})
		return
	}

	// 查询字典
	var dictionary models.SubdomainDictionary
	err = h.DB.Collection("subdomain_dictionaries").FindOne(c, bson.M{"_id": dictionaryID}).Decode(&dictionary)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dictionary not found"})
		return
	}

	// 删除文件
	if dictionary.FilePath != "" {
		os.Remove(dictionary.FilePath)
	}

	// 删除记录
	result, err := h.DB.Collection("subdomain_dictionaries").DeleteOne(c, bson.M{"_id": dictionaryID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete dictionary"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dictionary not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dictionary deleted successfully"})
}

// runEnumTask 运行子域名枚举任务
func (h *SubdomainHandler) runEnumTask(task models.SubdomainEnumTask) {
	// 更新任务状态为运行中
	update := bson.M{
		"$set": bson.M{
			"status":    "running",
			"startedAt": time.Now(),
		},
	}
	_, err := h.DB.Collection("subdomain_enum_tasks").UpdateOne(
		context.Background(),
		bson.M{"_id": task.ID},
		update,
	)
	if err != nil {
		logger.Error("Failed to update task status", map[string]interface{}{"error": err})
		return
	}

	// 创建子域名枚举服务
	enumService := subdomain.NewEnumService(
		task.Config,
		task.RootDomain,
		task.ID.Hex(),
		task.ProjectID.Hex(),
	)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动结果处理协程
	resultChan := enumService.ResultChan
	progressChan := enumService.ProgressChan
	var resultCount, resolvedCount, unresolvedCount, wildcardCount, takeOverCount int
	methodStats := make(map[string]int)

	// 处理进度更新
	go func() {
		for progress := range progressChan {
			// 更新任务进度
			update := bson.M{
				"$set": bson.M{
					"progress": progress,
				},
			}
			_, _ = h.DB.Collection("subdomain_enum_tasks").UpdateOne(
				context.Background(),
				bson.M{"_id": task.ID},
				update,
			)
		}
	}()

	// 处理结果
	go func() {
		for result := range resultChan {
			// 保存结果到数据库
			result.ID = primitive.NewObjectID()
			_, err := h.DB.Collection("subdomain_results").InsertOne(context.Background(), result)
			if err != nil {
				logger.Error("Failed to save result", map[string]interface{}{"error": err})
				continue
			}

			// 更新统计信息
			resultCount++
			if result.IsResolved {
				resolvedCount++
			} else {
				unresolvedCount++
			}
			if result.IsWildcard {
				wildcardCount++
			}
			if result.IsTakeOver {
				takeOverCount++
			}
			methodStats[result.Source]++

			// 如果需要，创建资产
			if task.Config.SaveToDB {
				// 创建子域名资产
				asset := &models.SubdomainAsset{
					BaseAsset: models.BaseAsset{
						ID:           primitive.NewObjectID(),
						CreatedAt:    time.Now(),
						UpdatedAt:    time.Now(),
						LastScanTime: time.Now(),
						Type:         models.AssetTypeSubdomain,
						ProjectID:    task.ProjectID,
						Tags:         task.Tags,
						TaskName:     task.TaskName,
						RootDomain:   task.RootDomain,
					},
					Host:     result.Subdomain,
					IPs:      result.IPs,
					CNAME:    result.CNAME,
					TakeOver: result.IsTakeOver,
				}

				// 保存资产
				_, err := h.DB.Collection(models.AssetCollection(models.AssetTypeSubdomain)).InsertOne(context.Background(), asset)
				if err != nil {
					logger.Error("Failed to save asset", map[string]interface{}{"error": err})
				} else {
					// 更新结果中的资产ID
					update := bson.M{
						"$set": bson.M{
							"assetId": asset.ID,
						},
					}
					_, _ = h.DB.Collection("subdomain_results").UpdateOne(
						context.Background(),
						bson.M{"_id": result.ID},
						update,
					)
				}
			}
		}
	}()

	// 启动枚举
	err = enumService.Start(ctx)
	if err != nil {
		// 更新任务状态为失败
		update := bson.M{
			"$set": bson.M{
				"status":    "failed",
				"error":     err.Error(),
				"updatedAt": time.Now(),
			},
		}
		_, _ = h.DB.Collection("subdomain_enum_tasks").UpdateOne(
			context.Background(),
			bson.M{"_id": task.ID},
			update,
		)
		return
	}

	// 更新任务状态为已完成
	update = bson.M{
		"$set": bson.M{
			"status":      "completed",
			"completedAt": time.Now(),
			"updatedAt":   time.Now(),
			"resultSummary": bson.M{
				"totalFound":       resultCount,
				"newFound":         resultCount, // 简化处理，实际应该计算新发现的子域名
				"resolvedCount":    resolvedCount,
				"unresolvedCount":  unresolvedCount,
				"wildcardCount":    wildcardCount,
				"takeOverCount":    takeOverCount,
				"methodStats":      methodStats,
				"processedRecords": resultCount,
			},
		},
	}
	_, _ = h.DB.Collection("subdomain_enum_tasks").UpdateOne(
		context.Background(),
		bson.M{"_id": task.ID},
		update,
	)
}
