package api

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/services/taskmanager"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	TaskManager *taskmanager.TaskManager
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskManager *taskmanager.TaskManager) *TaskHandler {
	return &TaskHandler{
		TaskManager: taskManager,
	}
}

// RegisterRoutes 注册路由
func (h *TaskHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("", h.SubmitTask)
	router.GET("", h.ListTasks)
	router.GET("/:id", h.GetTask)
	router.DELETE("/:id", h.CancelTask)
	router.GET("/:id/result", h.GetTaskResult)
}

// SubmitTask 提交任务
func (h *TaskHandler) SubmitTask(c *gin.Context) {
	// TODO: 实现提交任务
	c.JSON(200, gin.H{"message": "功能待实现"})
}

// ListTasks 列出任务
func (h *TaskHandler) ListTasks(c *gin.Context) {
	// TODO: 实现列出任务
	c.JSON(200, gin.H{"message": "功能待实现"})
}

// GetTask 获取任务
func (h *TaskHandler) GetTask(c *gin.Context) {
	// TODO: 实现获取任务
	c.JSON(200, gin.H{"message": "功能待实现"})
}

// CancelTask 取消任务
func (h *TaskHandler) CancelTask(c *gin.Context) {
	// TODO: 实现取消任务
	c.JSON(200, gin.H{"message": "功能待实现"})
}

// GetTaskResult 获取任务结果
func (h *TaskHandler) GetTaskResult(c *gin.Context) {
	// TODO: 实现获取任务结果
	c.JSON(200, gin.H{"message": "功能待实现"})
}

// TaskAPI 任务API处理器
type TaskAPI struct {
	taskManager *taskmanager.TaskManager
	db          *mongo.Database
}

// NewTaskAPI 创建任务API处理器
func NewTaskAPI(taskManager *taskmanager.TaskManager, db *mongo.Database) *TaskAPI {
	return &TaskAPI{
		taskManager: taskManager,
		db:          db,
	}
}

// RegisterRoutes 注册路由
func (api *TaskAPI) RegisterRoutes(router *gin.Engine) {
	taskGroup := router.Group("/api/tasks")
	{
		taskGroup.POST("", api.CreateTask)
		taskGroup.GET("", api.ListTasks)
		taskGroup.GET("/:id", api.GetTask)
		taskGroup.PUT("/:id/status", api.UpdateTaskStatus)
		taskGroup.POST("/:id/cancel", api.CancelTask)
		taskGroup.GET("/:id/result", api.GetTaskResult)
		taskGroup.GET("/:id/results", api.GetTaskResults)
		taskGroup.GET("/:id/export", api.ExportTaskResults)
		taskGroup.POST("/:id/retry", api.RetryTask)
	}
}

// CreateTask 创建任务
// @Summary 创建任务
// @Description 创建新的任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param task body models.TaskCreateRequest true "任务信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tasks [post]
func (api *TaskAPI) CreateTask(c *gin.Context) {
	var req models.TaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	// 验证必填字段
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务名称不能为空"})
		return
	}
	if req.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务类型不能为空"})
		return
	}
	if req.ProjectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目ID不能为空"})
		return
	}

	// 创建任务
	task := models.Task{
		ID:          primitive.NewObjectID(),
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Status:      string(models.TaskStatusPending),
		Priority:    req.Priority,
		CreatedAt:   time.Now(),
		Timeout:     req.Timeout,
		MaxRetries:  req.MaxRetries,
		DependsOn:   req.DependsOn,
		Tags:        req.Tags,
		CallbackURL: req.CallbackURL,
		Params:      req.Params,
	}

	// 设置项目ID
	projectID, err := primitive.ObjectIDFromHex(req.ProjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}
	task.ProjectID = projectID

	// 设置创建者ID
	userID, exists := c.Get("userID")
	if exists {
		if userIDStr, ok := userID.(string); ok {
			creatorID, err := primitive.ObjectIDFromHex(userIDStr)
			if err == nil {
				task.CreatedBy = creatorID
			}
		}
	}

	// 提交任务
	err = api.taskManager.SubmitTask(&task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      task.ID.Hex(),
		"name":    task.Name,
		"status":  task.Status,
		"message": "任务创建成功",
	})
}

// ListTasks 列出任务
// @Summary 列出任务
// @Description 获取任务列表
// @Tags 任务管理
// @Produce json
// @Param projectId query string false "项目ID"
// @Param status query string false "任务状态"
// @Param type query string false "任务类型"
// @Param limit query int false "限制数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/tasks [get]
func (api *TaskAPI) ListTasks(c *gin.Context) {
	// 获取查询参数
	projectID := c.Query("projectId")
	status := c.Query("status")
	taskType := c.Query("type")
	limit := 20
	offset := 0

	// 获取任务列表
	tasks, total, err := api.taskManager.ListTasks(projectID, status, taskType, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务列表失败: " + err.Error()})
		return
	}

	// 转换为响应格式
	var resp []gin.H
	for _, task := range tasks {
		resp = append(resp, gin.H{
			"id":          task.ID.Hex(),
			"name":        task.Name,
			"description": task.Description,
			"type":        task.Type,
			"status":      task.Status,
			"priority":    task.Priority,
			"projectId":   task.ProjectID.Hex(),
			"createdAt":   task.CreatedAt,
			"startedAt":   task.StartedAt,
			"completedAt": task.CompletedAt,
			"progress":    task.Progress,
			"nodeId":      task.NodeID,
			"tags":        task.Tags,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":  resp,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetTask 获取任务详情
// @Summary 获取任务详情
// @Description 获取指定任务的详细信息
// @Tags 任务管理
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tasks/{id} [get]
func (api *TaskAPI) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务ID不能为空"})
		return
	}

	// 获取任务详情
	task, err := api.taskManager.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务详情失败: " + err.Error()})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 转换为响应格式
	resp := gin.H{
		"id":          task.ID.Hex(),
		"name":        task.Name,
		"description": task.Description,
		"type":        task.Type,
		"status":      task.Status,
		"priority":    task.Priority,
		"projectId":   task.ProjectID.Hex(),
		"createdAt":   task.CreatedAt,
		"startedAt":   task.StartedAt,
		"completedAt": task.CompletedAt,
		"timeout":     task.Timeout,
		"retryCount":  task.RetryCount,
		"maxRetries":  task.MaxRetries,
		"progress":    task.Progress,
		"nodeId":      task.NodeID,
		"dependsOn":   task.DependsOn,
		"tags":        task.Tags,
		"error":       task.Error,
		"callbackURL": task.CallbackURL,
		"params":      task.Params,
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateTaskStatus 更新任务状态
// @Summary 更新任务状态
// @Description 更新指定任务的状态
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Param status body models.TaskUpdateRequest true "任务状态更新"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tasks/{id}/status [put]
func (api *TaskAPI) UpdateTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务ID不能为空"})
		return
	}

	var req models.TaskUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	// 更新任务状态
	err := api.taskManager.UpdateTaskStatus(taskID, req.Status, req.Progress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新任务状态失败: " + err.Error()})
		return
	}

	// 如果有结果数据，保存结果
	if req.ResultData != nil {
		resultData, ok := req.ResultData.(map[string]interface{})
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "结果数据格式不正确",
			})
			return
		}
		result := &models.TaskResult{
			Status:  req.Status,
			Data:    resultData,
			Summary: "",
			Error:   req.Error,
		}
		err = api.taskManager.SaveTaskResult(taskID, result)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存任务结果失败: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       taskID,
		"status":   req.Status,
		"progress": req.Progress,
		"message":  "任务状态更新成功",
	})
}

// CancelTask 取消任务
// @Summary 取消任务
// @Description 取消指定的任务
// @Tags 任务管理
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tasks/{id}/cancel [post]
func (api *TaskAPI) CancelTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务ID不能为空"})
		return
	}

	// 取消任务
	err := api.taskManager.CancelTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消任务失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      taskID,
		"message": "任务取消成功",
	})
}

// GetTaskResult 获取任务结果
// @Summary 获取任务结果
// @Description 获取指定任务的结果
// @Tags 任务管理
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tasks/{id}/result [get]
func (api *TaskAPI) GetTaskResult(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务ID不能为空"})
		return
	}

	// 获取任务结果
	result, err := api.taskManager.GetTaskResult(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务结果失败: " + err.Error()})
		return
	}

	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务结果不存在"})
		return
	}

	// 转换为响应格式
	resp := gin.H{
		"id":        result.ID.Hex(),
		"taskId":    result.TaskID.Hex(),
		"status":    result.Status,
		"data":      result.Data,
		"summary":   result.Summary,
		"createdAt": result.CreatedAt,
		"endTime":   result.EndTime,
		"updatedAt": result.UpdatedAt,
	}

	c.JSON(http.StatusOK, resp)
}

// GetTaskResults 获取任务详细结果
// @Summary 获取任务详细结果
// @Description 获取指定任务的详细结果列表
// @Tags 任务管理
// @Produce json
// @Param id path string true "任务ID"
// @Param page query int false "页码"
// @Param limit query int false "每页数量"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tasks/{id}/results [get]
func (api *TaskAPI) GetTaskResults(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务ID不能为空"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 1000 {
		limit = 50
	}

	// 首先获取任务信息
	task, err := api.taskManager.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务失败: " + err.Error()})
		return
	}
	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 根据任务类型获取相应的结果
	var results []interface{}
	var total int64

	switch task.Type {
	case models.TaskTypeSubdomainEnum:
		results, total, err = api.getSubdomainResults(task.ID, page, limit)
	case models.TaskTypePortScan:
		results, total, err = api.getPortScanResults(task.ID, page, limit)
	case models.TaskTypeVulnScan:
		results, total, err = api.getVulnScanResults(task.ID, page, limit)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的任务类型"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取结果失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"total":   total,
		"page":    page,
		"limit":   limit,
		"pages":   (total + int64(limit) - 1) / int64(limit),
	})
}

// ExportTaskResults 导出任务结果
// @Summary 导出任务结果
// @Description 导出指定任务的结果为指定格式
// @Tags 任务管理
// @Produce json
// @Param id path string true "任务ID"
// @Param format query string false "导出格式(csv|json)" default(csv)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tasks/{id}/export [get]
func (api *TaskAPI) ExportTaskResults(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务ID不能为空"})
		return
	}

	format := c.DefaultQuery("format", "csv")
	if format != "csv" && format != "json" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的导出格式"})
		return
	}

	// 获取任务信息
	task, err := api.taskManager.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务失败: " + err.Error()})
		return
	}
	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 获取所有结果
	var results []interface{}
	var filename string

	switch task.Type {
	case models.TaskTypeSubdomainEnum:
		results, _, err = api.getSubdomainResults(task.ID, 1, 10000)
		filename = fmt.Sprintf("subdomain_%s_%s.%s", task.Name, time.Now().Format("20060102_150405"), format)
	case models.TaskTypePortScan:
		results, _, err = api.getPortScanResults(task.ID, 1, 10000)
		filename = fmt.Sprintf("portscan_%s_%s.%s", task.Name, time.Now().Format("20060102_150405"), format)
	case models.TaskTypeVulnScan:
		results, _, err = api.getVulnScanResults(task.ID, 1, 10000)
		filename = fmt.Sprintf("vulnscan_%s_%s.%s", task.Name, time.Now().Format("20060102_150405"), format)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的任务类型"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取结果失败: " + err.Error()})
		return
	}

	// 根据格式导出
	if format == "csv" {
		api.exportCSV(c, results, filename, task.Type)
	} else {
		api.exportJSON(c, results, filename)
	}
}

// RetryTask 重新运行任务
// @Summary 重新运行任务
// @Description 重新运行指定的任务
// @Tags 任务管理
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tasks/{id}/retry [post]
func (api *TaskAPI) RetryTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务ID不能为空"})
		return
	}

	// 获取原始任务
	originalTask, err := api.taskManager.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务失败: " + err.Error()})
		return
	}
	if originalTask == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 创建新任务
	newTask := models.Task{
		ID:          primitive.NewObjectID(),
		Name:        originalTask.Name + " (重试)",
		Description: originalTask.Description,
		Type:        originalTask.Type,
		Status:      string(models.TaskStatusPending),
		Priority:    originalTask.Priority,
		ProjectID:   originalTask.ProjectID,
		CreatedBy:   originalTask.CreatedBy,
		CreatedAt:   time.Now(),
		Timeout:     originalTask.Timeout,
		MaxRetries:  originalTask.MaxRetries,
		Tags:        originalTask.Tags,
		CallbackURL: originalTask.CallbackURL,
		Params:      originalTask.Params,
	}

	// 提交新任务
	err = api.taskManager.SubmitTask(&newTask)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重新启动任务失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         newTask.ID.Hex(),
		"name":       newTask.Name,
		"status":     newTask.Status,
		"originalId": taskID,
		"message":    "任务重新启动成功",
	})
}

// 获取子域名枚举结果
func (api *TaskAPI) getSubdomainResults(taskID primitive.ObjectID, page, limit int) ([]interface{}, int64, error) {
	collection := api.db.Collection("subdomain_results")

	// 构建查询条件
	filter := bson.M{"task_id": taskID}

	// 获取总数
	total, err := collection.CountDocuments(nil, filter)
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	skip := (page - 1) * limit
	cursor, err := collection.Find(nil, filter, options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{"created_at", -1}}),
	)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(nil)

	var results []interface{}
	for cursor.Next(nil) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, total, nil
}

// 获取端口扫描结果
func (api *TaskAPI) getPortScanResults(taskID primitive.ObjectID, page, limit int) ([]interface{}, int64, error) {
	collection := api.db.Collection("port_scan_results")

	// 构建查询条件
	filter := bson.M{"task_id": taskID}

	// 获取总数
	total, err := collection.CountDocuments(nil, filter)
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	skip := (page - 1) * limit
	cursor, err := collection.Find(nil, filter, options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{"created_at", -1}}),
	)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(nil)

	var results []interface{}
	for cursor.Next(nil) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, total, nil
}

// 获取漏洞扫描结果
func (api *TaskAPI) getVulnScanResults(taskID primitive.ObjectID, page, limit int) ([]interface{}, int64, error) {
	collection := api.db.Collection("vulnerability_results")

	// 构建查询条件
	filter := bson.M{"task_id": taskID}

	// 获取总数
	total, err := collection.CountDocuments(nil, filter)
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	skip := (page - 1) * limit
	cursor, err := collection.Find(nil, filter, options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{"created_at", -1}}),
	)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(nil)

	var results []interface{}
	for cursor.Next(nil) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, total, nil
}

// 导出CSV格式
func (api *TaskAPI) exportCSV(c *gin.Context, results []interface{}, filename string, taskType string) {
	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "没有可导出的数据",
		})
		return
	}

	// 设置CSV响应头
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// 根据任务类型写入不同的CSV头
	switch taskType {
	case models.TaskTypeSubdomainEnum:
		writer.Write([]string{"子域名", "完整域名", "IP地址", "状态", "来源", "响应时间", "发现时间"})
		for _, result := range results {
			if data, ok := result.(bson.M); ok {
				writer.Write([]string{
					getStringValue(data, "subdomain"),
					getStringValue(data, "domain"),
					getStringValue(data, "ip"),
					getStringValue(data, "status"),
					getStringValue(data, "source"),
					getStringValue(data, "response_time"),
					getStringValue(data, "created_at"),
				})
			}
		}
	case models.TaskTypePortScan:
		writer.Write([]string{"主机", "端口", "协议", "状态", "服务", "版本", "响应时间", "Banner"})
		for _, result := range results {
			if data, ok := result.(bson.M); ok {
				writer.Write([]string{
					getStringValue(data, "host"),
					getStringValue(data, "port"),
					getStringValue(data, "protocol"),
					getStringValue(data, "status"),
					getStringValue(data, "service"),
					getStringValue(data, "version"),
					getStringValue(data, "response_time"),
					getStringValue(data, "banner"),
				})
			}
		}
	case models.TaskTypeVulnScan:
		writer.Write([]string{"目标", "漏洞名称", "严重级别", "CVSS评分", "描述", "状态", "发现时间"})
		for _, result := range results {
			if data, ok := result.(bson.M); ok {
				writer.Write([]string{
					getStringValue(data, "target"),
					getStringValue(data, "name"),
					getStringValue(data, "severity"),
					getStringValue(data, "cvss_score"),
					getStringValue(data, "description"),
					getStringValue(data, "status"),
					getStringValue(data, "created_at"),
				})
			}
		}
	}
}

// 导出JSON格式
func (api *TaskAPI) exportJSON(c *gin.Context, results []interface{}, filename string) {
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	c.JSON(http.StatusOK, gin.H{
		"results":     results,
		"total":       len(results),
		"exported_at": time.Now().Format("2006-01-02 15:04:05"),
	})
}

// 获取字符串值的辅助函数
func getStringValue(data bson.M, key string) string {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", value)
	}
	return ""
}
