package api

import (
	"net/http"

	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/services/sensitive"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SensitiveHandler 敏感信息处理器
type SensitiveHandler struct {
	service *sensitive.Service
}

// NewSensitiveHandler 创建敏感信息处理器
func NewSensitiveHandler(service *sensitive.Service) *SensitiveHandler {
	return &SensitiveHandler{
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *SensitiveHandler) RegisterRoutes(router *gin.RouterGroup) {
	sensitiveRoutes := router.Group("/sensitive")
	{
		sensitiveRoutes.POST("/scan", h.ScanForSensitiveInfo)
		sensitiveRoutes.GET("/list", h.ListSensitiveInfo)
		sensitiveRoutes.GET("/:id", h.GetSensitiveInfo)
		sensitiveRoutes.PUT("/:id/status", h.UpdateSensitiveInfoStatus)
		sensitiveRoutes.DELETE("/:id", h.DeleteSensitiveInfo)
		sensitiveRoutes.GET("/rules", h.ListSensitiveRules)
		sensitiveRoutes.POST("/rules", h.CreateSensitiveRule)
		sensitiveRoutes.PUT("/rules/:id", h.UpdateSensitiveRule)
		sensitiveRoutes.DELETE("/rules/:id", h.DeleteSensitiveRule)
	}
}

// ScanRequest 扫描请求
type ScanRequest struct {
	ProjectID string   `json:"projectId" binding:"required"`
	URLs      []string `json:"urls" binding:"required"`
	RuleIDs   []string `json:"ruleIds"`
}

// ScanForSensitiveInfo 扫描敏感信息
func (h *SensitiveHandler) ScanForSensitiveInfo(c *gin.Context) {
	var req ScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	projectID, err := primitive.ObjectIDFromHex(req.ProjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID",
		})
		return
	}

	// 创建扫描任务
	taskID, err := h.service.CreateScanTask(projectID, req.URLs, req.RuleIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create scan task: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"taskId": taskID.Hex(),
		"status": "created",
	})
}

// ListSensitiveInfo 列出敏感信息
func (h *SensitiveHandler) ListSensitiveInfo(c *gin.Context) {
	projectID := c.Query("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Project ID is required",
		})
		return
	}

	objID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID",
		})
		return
	}

	// 获取分页参数
	page := 1
	limit := 20
	// TODO: 从请求中获取分页参数

	// 获取敏感信息列表
	results, total, err := h.service.ListSensitiveInfo(objID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list sensitive info: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"data":  results,
	})
}

// GetSensitiveInfo 获取敏感信息详情
func (h *SensitiveHandler) GetSensitiveInfo(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID",
		})
		return
	}

	// 获取敏感信息详情
	info, err := h.service.GetSensitiveInfo(objID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get sensitive info: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, info)
}

// UpdateStatusRequest 更新状态请求
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateSensitiveInfoStatus 更新敏感信息状态
func (h *SensitiveHandler) UpdateSensitiveInfoStatus(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID",
		})
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// 更新状态
	err = h.service.UpdateSensitiveInfoStatus(objID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "updated",
	})
}

// DeleteSensitiveInfo 删除敏感信息
func (h *SensitiveHandler) DeleteSensitiveInfo(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID",
		})
		return
	}

	// 删除敏感信息
	err = h.service.DeleteSensitiveInfo(objID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete sensitive info: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "deleted",
	})
}

// ListSensitiveRules 列出敏感规则
func (h *SensitiveHandler) ListSensitiveRules(c *gin.Context) {
	// 获取敏感规则列表
	rules, err := h.service.ListSensitiveRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list sensitive rules: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, rules)
}

// CreateSensitiveRule 创建敏感规则
func (h *SensitiveHandler) CreateSensitiveRule(c *gin.Context) {
	var rule models.SensitiveRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// 创建敏感规则
	id, err := h.service.CreateSensitiveRule(rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create sensitive rule: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     id.Hex(),
		"status": "created",
	})
}

// UpdateSensitiveRule 更新敏感规则
func (h *SensitiveHandler) UpdateSensitiveRule(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID",
		})
		return
	}

	var rule models.SensitiveRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	rule.ID = objID

	// 更新敏感规则
	err = h.service.UpdateSensitiveRule(rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update sensitive rule: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "updated",
	})
}

// DeleteSensitiveRule 删除敏感规则
func (h *SensitiveHandler) DeleteSensitiveRule(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID",
		})
		return
	}

	// 删除敏感规则
	err = h.service.DeleteSensitiveRule(objID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete sensitive rule: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "deleted",
	})
}
