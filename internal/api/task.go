package api

import (
	"net/http"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/services/taskmanager"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	taskGroup := router.Group("/tasks")
	{
		taskGroup.POST("", h.SubmitTask)
		taskGroup.GET("", h.ListTasks)
		taskGroup.GET("/:id", h.GetTask)
		taskGroup.DELETE("/:id", h.CancelTask)
		taskGroup.GET("/:id/result", h.GetTaskResult)
	}
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
}

// NewTaskAPI 创建任务API处理器
func NewTaskAPI(taskManager *taskmanager.TaskManager) *TaskAPI {
	return &TaskAPI{
		taskManager: taskManager,
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
		Status:      models.TaskStatusPending,
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
		result := &models.TaskResult{
			Status:  req.Status,
			Data:    req.ResultData,
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
		"id":          result.ID.Hex(),
		"taskId":      result.TaskID.Hex(),
		"status":      result.Status,
		"data":        result.Data,
		"summary":     result.Summary,
		"createdAt":   result.CreatedAt,
		"completedAt": result.CompletedAt,
		"error":       result.Error,
	}

	c.JSON(http.StatusOK, resp)
}
