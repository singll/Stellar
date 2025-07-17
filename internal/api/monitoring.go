package api

import (
	"net/http"
	"strconv"

	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/services/pagemonitoring"
	"github.com/gin-gonic/gin"
)

// MonitoringHandler 页面监控处理器
type MonitoringHandler struct {
	service *pagemonitoring.PageMonitoringService
}

// NewMonitoringHandler 创建页面监控处理器
func NewMonitoringHandler(service *pagemonitoring.PageMonitoringService) *MonitoringHandler {
	return &MonitoringHandler{
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *MonitoringHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("", h.CreateMonitoring)
	router.GET("", h.ListMonitorings)
	router.GET("/:id", h.GetMonitoring)
	router.PUT("/:id", h.UpdateMonitoring)
	router.DELETE("/:id", h.DeleteMonitoring)
	router.GET("/:id/snapshots", h.GetSnapshots)
	router.GET("/:id/changes", h.GetChanges)
	router.GET("/:id/diff", h.GetDiff)
}

// CreateMonitoring 创建监控任务
// @Summary 创建页面监控任务
// @Description 创建新的页面监控任务
// @Tags 页面监控
// @Accept json
// @Produce json
// @Param request body models.PageMonitoringCreateRequest true "监控任务信息"
// @Success 200 {object} Response{data=models.PageMonitoring}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/monitoring [post]
func (h *MonitoringHandler) CreateMonitoring(c *gin.Context) {
	var req models.PageMonitoringCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
			"error":   err.Error(),
		})
		return
	}

	monitoring, err := h.service.CreateMonitoring(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建监控任务失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建监控任务成功",
		"data":    monitoring,
	})
}

// ListMonitorings 列出监控任务
// @Summary 列出页面监控任务
// @Description 获取页面监控任务列表
// @Tags 页面监控
// @Accept json
// @Produce json
// @Param projectId query string false "项目ID"
// @Param status query string false "状态"
// @Param tags query []string false "标签"
// @Param url query string false "URL关键字"
// @Param name query string false "名称关键字"
// @Param limit query int false "限制数量"
// @Param offset query int false "偏移量"
// @Param sortBy query string false "排序字段"
// @Param sortOrder query string false "排序方向"
// @Success 200 {object} Response{data=[]models.PageMonitoring}
// @Failure 500 {object} Response
// @Router /api/monitoring [get]
func (h *MonitoringHandler) ListMonitorings(c *gin.Context) {
	// 解析查询参数
	req := models.PageMonitoringQueryRequest{
		ProjectID: c.Query("projectId"),
		Status:    c.Query("status"),
		URL:       c.Query("url"),
		Name:      c.Query("name"),
		SortBy:    c.Query("sortBy"),
		SortOrder: c.Query("sortOrder"),
	}

	// 解析标签
	if tags := c.QueryArray("tags"); len(tags) > 0 {
		req.Tags = tags
	}

	// 解析分页参数
	if limit, err := strconv.Atoi(c.Query("limit")); err == nil {
		req.Limit = limit
	}
	if offset, err := strconv.Atoi(c.Query("offset")); err == nil {
		req.Offset = offset
	}

	// 查询数据
	monitorings, total, err := h.service.ListMonitorings(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询监控任务失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  monitorings,
			"total": total,
		},
	})
}

// GetMonitoring 获取监控任务
// @Summary 获取页面监控任务
// @Description 获取指定ID的页面监控任务详情
// @Tags 页面监控
// @Accept json
// @Produce json
// @Param id path string true "监控任务ID"
// @Success 200 {object} Response{data=models.PageMonitoring}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /api/monitoring/{id} [get]
func (h *MonitoringHandler) GetMonitoring(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少监控任务ID",
		})
		return
	}

	monitoring, err := h.service.GetMonitoring(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取监控任务失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": monitoring,
	})
}

// UpdateMonitoring 更新监控任务
// @Summary 更新页面监控任务
// @Description 更新指定ID的页面监控任务
// @Tags 页面监控
// @Accept json
// @Produce json
// @Param id path string true "监控任务ID"
// @Param request body models.PageMonitoringUpdateRequest true "更新信息"
// @Success 200 {object} Response{data=models.PageMonitoring}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /api/monitoring/{id} [put]
func (h *MonitoringHandler) UpdateMonitoring(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少监控任务ID",
		})
		return
	}

	var req models.PageMonitoringUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
			"error":   err.Error(),
		})
		return
	}

	monitoring, err := h.service.UpdateMonitoring(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新监控任务失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新监控任务成功",
		"data":    monitoring,
	})
}

// DeleteMonitoring 删除监控任务
// @Summary 删除页面监控任务
// @Description 删除指定ID的页面监控任务
// @Tags 页面监控
// @Accept json
// @Produce json
// @Param id path string true "监控任务ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/monitoring/{id} [delete]
func (h *MonitoringHandler) DeleteMonitoring(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少监控任务ID",
		})
		return
	}

	err := h.service.DeleteMonitoring(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除监控任务失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除监控任务成功",
	})
}

// GetSnapshots 获取快照列表
// @Summary 获取页面快照列表
// @Description 获取指定监控任务的页面快照列表
// @Tags 页面监控
// @Accept json
// @Produce json
// @Param id path string true "监控任务ID"
// @Param limit query int false "限制数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} Response{data=[]models.PageSnapshot}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/monitoring/{id}/snapshots [get]
func (h *MonitoringHandler) GetSnapshots(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少监控任务ID",
		})
		return
	}

	// 解析分页参数
	limit := 10
	offset := 0
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	snapshots, total, err := h.service.GetSnapshots(id, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取快照列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  snapshots,
			"total": total,
		},
	})
}

// GetChanges 获取变更记录
// @Summary 获取页面变更记录
// @Description 获取指定监控任务的页面变更记录
// @Tags 页面监控
// @Accept json
// @Produce json
// @Param id path string true "监控任务ID"
// @Param limit query int false "限制数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} Response{data=[]models.PageChange}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/monitoring/{id}/changes [get]
func (h *MonitoringHandler) GetChanges(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少监控任务ID",
		})
		return
	}

	// 解析分页参数
	limit := 10
	offset := 0
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	changes, total, err := h.service.GetChanges(id, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取变更记录失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  changes,
			"total": total,
		},
	})
}

// GetDiff 获取差异内容
// @Summary 获取页面差异内容
// @Description 获取指定监控任务的最新页面差异内容
// @Tags 页面监控
// @Accept json
// @Produce json
// @Param id path string true "监控任务ID"
// @Success 200 {object} Response{data=string}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /api/monitoring/{id}/diff [get]
func (h *MonitoringHandler) GetDiff(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少监控任务ID",
		})
		return
	}

	// 获取监控任务
	monitoring, err := h.service.GetMonitoring(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取监控任务失败",
			"error":   err.Error(),
		})
		return
	}

	// 检查是否有快照
	if monitoring.LatestSnapshot == nil || monitoring.PreviousSnapshot == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "没有足够的快照用于比较",
		})
		return
	}

	// 比较快照
	_, _, diff := pagemonitoring.CompareSnapshots(monitoring.PreviousSnapshot, monitoring.LatestSnapshot, monitoring.Config)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"diff": diff,
		},
	})
}
