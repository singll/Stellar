package api

import (
	"net/http"
	"strconv"

	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/services/portscan"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PortScanAPI 端口扫描API
type PortScanAPI struct {
	taskManager *portscan.TaskManager
}

// NewPortScanAPI 创建端口扫描API
func NewPortScanAPI(taskManager *portscan.TaskManager) *PortScanAPI {
	return &PortScanAPI{
		taskManager: taskManager,
	}
}

// RegisterRoutes 注册路由
func (api *PortScanAPI) RegisterRoutes(router *gin.RouterGroup) {
	portScanGroup := router.Group("/portscan")
	{
		portScanGroup.POST("/tasks", api.CreateTask)
		portScanGroup.GET("/tasks", api.ListTasks)
		portScanGroup.GET("/tasks/:id", api.GetTask)
		portScanGroup.POST("/tasks/:id/start", api.StartTask)
		portScanGroup.POST("/tasks/:id/stop", api.StopTask)
		portScanGroup.GET("/tasks/:id/status", api.GetTaskStatus)
		portScanGroup.GET("/tasks/:id/progress", api.GetTaskProgress)
		portScanGroup.GET("/tasks/:id/results", api.GetTaskResults)
	}
}

// CreateTask 创建端口扫描任务
// @Summary 创建端口扫描任务
// @Description 创建一个新的端口扫描任务
// @Tags 端口扫描
// @Accept json
// @Produce json
// @Param task body models.PortScanTask true "端口扫描任务"
// @Success 201 {object} map[string]string "任务ID"
// @Failure 400 {object} models.ErrorResponse "请求错误"
// @Failure 500 {object} models.ErrorResponse "服务器错误"
// @Router /api/v1/portscan/tasks [post]
func (api *PortScanAPI) CreateTask(c *gin.Context) {
	var task models.PortScanTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 验证必要参数
	if len(task.Targets) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "目标列表不能为空",
		})
		return
	}

	// 获取项目ID
	projectIDStr := c.Query("projectId")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "项目ID不能为空",
		})
		return
	}

	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "无效的项目ID: " + err.Error(),
		})
		return
	}
	task.ProjectID = projectID

	// 设置默认值
	if task.Name == "" {
		task.Name = "端口扫描任务"
	}
	if task.Config.Ports == "" {
		task.Config.Ports = "1-1000,3389,8080-8090"
	}
	if task.Config.ScanType == "" {
		task.Config.ScanType = "tcp"
	}
	if task.Config.ScanMethod == "" {
		task.Config.ScanMethod = "connect"
	}
	if task.Config.Concurrency <= 0 {
		task.Config.Concurrency = 100
	}
	if task.Config.Timeout <= 0 {
		task.Config.Timeout = 3
	}
	if task.Config.RetryCount <= 0 {
		task.Config.RetryCount = 2
	}

	// 创建任务
	taskID, err := api.taskManager.CreateTask(&task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "创建任务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"taskId": taskID,
	})
}

// ListTasks 列出端口扫描任务
// @Summary 列出端口扫描任务
// @Description 列出所有端口扫描任务
// @Tags 端口扫描
// @Produce json
// @Param projectId query string false "项目ID"
// @Param status query string false "任务状态"
// @Param limit query int false "限制数量"
// @Param skip query int false "跳过数量"
// @Success 200 {array} models.PortScanTask "任务列表"
// @Failure 500 {object} models.ErrorResponse "服务器错误"
// @Router /api/v1/portscan/tasks [get]
func (api *PortScanAPI) ListTasks(c *gin.Context) {
	// 获取查询参数
	projectID := c.Query("projectId")
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "10")
	skipStr := c.DefaultQuery("skip", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}
	skip, err := strconv.Atoi(skipStr)
	if err != nil {
		skip = 0
	}

	// 构建查询条件
	query := make(map[string]interface{})
	if projectID != "" {
		projectObjID, err := primitive.ObjectIDFromHex(projectID)
		if err == nil {
			query["projectId"] = projectObjID
		}
	}
	if status != "" {
		query["status"] = status
	}

	// 查询任务列表
	tasks, err := api.taskManager.ListTasks(query, limit, skip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "获取任务列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetTask 获取端口扫描任务详情
// @Summary 获取端口扫描任务详情
// @Description 获取指定任务的详细信息
// @Tags 端口扫描
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} models.PortScanTask "任务详情"
// @Failure 400 {object} models.ErrorResponse "请求错误"
// @Failure 404 {object} models.ErrorResponse "任务不存在"
// @Failure 500 {object} models.ErrorResponse "服务器错误"
// @Router /api/v1/portscan/tasks/{id} [get]
func (api *PortScanAPI) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "任务ID不能为空",
		})
		return
	}

	task, err := api.taskManager.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "获取任务失败: " + err.Error(),
		})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Message: "任务不存在",
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// StartTask 启动端口扫描任务
// @Summary 启动端口扫描任务
// @Description 启动指定的端口扫描任务
// @Tags 端口扫描
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]string "成功信息"
// @Failure 400 {object} models.ErrorResponse "请求错误"
// @Failure 500 {object} models.ErrorResponse "服务器错误"
// @Router /api/v1/portscan/tasks/{id}/start [post]
func (api *PortScanAPI) StartTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "任务ID不能为空",
		})
		return
	}

	err := api.taskManager.StartTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "启动任务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "任务已启动",
	})
}

// StopTask 停止端口扫描任务
// @Summary 停止端口扫描任务
// @Description 停止指定的端口扫描任务
// @Tags 端口扫描
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]string "成功信息"
// @Failure 400 {object} models.ErrorResponse "请求错误"
// @Failure 500 {object} models.ErrorResponse "服务器错误"
// @Router /api/v1/portscan/tasks/{id}/stop [post]
func (api *PortScanAPI) StopTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "任务ID不能为空",
		})
		return
	}

	err := api.taskManager.StopTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "停止任务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "任务已停止",
	})
}

// GetTaskStatus 获取任务状态
// @Summary 获取任务状态
// @Description 获取指定任务的状态
// @Tags 端口扫描
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]string "任务状态"
// @Failure 400 {object} models.ErrorResponse "请求错误"
// @Failure 500 {object} models.ErrorResponse "服务器错误"
// @Router /api/v1/portscan/tasks/{id}/status [get]
func (api *PortScanAPI) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "任务ID不能为空",
		})
		return
	}

	status, err := api.taskManager.GetTaskStatus(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "获取任务状态失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

// GetTaskProgress 获取任务进度
// @Summary 获取任务进度
// @Description 获取指定任务的进度
// @Tags 端口扫描
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]float64 "任务进度"
// @Failure 400 {object} models.ErrorResponse "请求错误"
// @Failure 500 {object} models.ErrorResponse "服务器错误"
// @Router /api/v1/portscan/tasks/{id}/progress [get]
func (api *PortScanAPI) GetTaskProgress(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "任务ID不能为空",
		})
		return
	}

	progress, err := api.taskManager.GetTaskProgress(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "获取任务进度失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"progress": progress,
	})
}

// GetTaskResults 获取任务结果
// @Summary 获取任务结果
// @Description 获取指定任务的结果
// @Tags 端口扫描
// @Produce json
// @Param id path string true "任务ID"
// @Param limit query int false "限制数量"
// @Param skip query int false "跳过数量"
// @Success 200 {array} models.PortScanResult "任务结果"
// @Failure 400 {object} models.ErrorResponse "请求错误"
// @Failure 500 {object} models.ErrorResponse "服务器错误"
// @Router /api/v1/portscan/tasks/{id}/results [get]
func (api *PortScanAPI) GetTaskResults(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "任务ID不能为空",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "100")
	skipStr := c.DefaultQuery("skip", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}
	skip, err := strconv.Atoi(skipStr)
	if err != nil {
		skip = 0
	}

	results, err := api.taskManager.GetTaskResults(taskID, limit, skip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "获取任务结果失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, results)
}

// PortScanHandler 端口扫描处理器
type PortScanHandler struct {
	DB      *mongo.Database
	Handler *portscan.Handler
}

// NewPortScanHandler 创建端口扫描处理器
func NewPortScanHandler(db *mongo.Database, handler *portscan.Handler) *PortScanHandler {
	return &PortScanHandler{
		DB:      db,
		Handler: handler,
	}
}

// RegisterRoutes 注册路由
func (h *PortScanHandler) RegisterRoutes(router *gin.RouterGroup) {
	scanGroup := router.Group("/portscan")
	{
		scanGroup.POST("/tasks", h.CreateScanTask)
		scanGroup.GET("/tasks", h.ListScanTasks)
		scanGroup.GET("/tasks/:id", h.GetScanTask)
		scanGroup.DELETE("/tasks/:id", h.DeleteScanTask)
		scanGroup.GET("/results", h.ListScanResults)
		scanGroup.GET("/results/:id", h.GetScanResult)
	}
}

// CreateScanTask 创建扫描任务
func (h *PortScanHandler) CreateScanTask(c *gin.Context) {
	// TODO: 实现创建扫描任务
	c.JSON(200, gin.H{"message": "功能待实现"})
}

// ListScanTasks 列出扫描任务
func (h *PortScanHandler) ListScanTasks(c *gin.Context) {
	// TODO: 实现列出扫描任务
	c.JSON(200, gin.H{"message": "功能待实现"})
}

// GetScanTask 获取扫描任务
func (h *PortScanHandler) GetScanTask(c *gin.Context) {
	// TODO: 实现获取扫描任务
	c.JSON(200, gin.H{"message": "功能待实现"})
}

// DeleteScanTask 删除扫描任务
func (h *PortScanHandler) DeleteScanTask(c *gin.Context) {
	// TODO: 实现删除扫描任务
	c.JSON(200, gin.H{"message": "功能待实现"})
}

// ListScanResults 列出扫描结果
func (h *PortScanHandler) ListScanResults(c *gin.Context) {
	// TODO: 实现列出扫描结果
	c.JSON(200, gin.H{"message": "功能待实现"})
}

// GetScanResult 获取扫描结果
func (h *PortScanHandler) GetScanResult(c *gin.Context) {
	// TODO: 实现获取扫描结果
	c.JSON(200, gin.H{"message": "功能待实现"})
}
