package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/services/nodemanager"
	"github.com/gin-gonic/gin"
)

// NodeHandler 节点处理器
type NodeHandler struct {
	NodeManager *nodemanager.NodeManager
	NodeRepo    models.NodeRepository
}

// NewNodeHandler 创建节点处理器
func NewNodeHandler(nodeManager *nodemanager.NodeManager, nodeRepo models.NodeRepository) *NodeHandler {
	return &NodeHandler{
		NodeManager: nodeManager,
		NodeRepo:    nodeRepo,
	}
}

// RegisterRoutes 注册路由
func (h *NodeHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("", h.GetNodes)
	router.POST("", h.CreateNode)
	router.GET("/:id", h.GetNode)
	router.PUT("/:id", h.UpdateNode)
	router.DELETE("/:id", h.DeleteNode)

	router.PUT("/:id/status", h.UpdateNodeStatus)
	router.POST("/:id/heartbeat", h.NodeHeartbeat)
	router.GET("/:id/health", h.GetNodeHealth)

	router.PUT("/:id/config", h.UpdateNodeConfig)
	router.GET("/:id/config", h.GetNodeConfig)

	router.GET("/:id/tasks", h.GetNodeTasks)
	router.GET("/:id/task-stats", h.GetNodeTaskStats)

	router.POST("/register", h.RegisterNode)
	router.POST("/unregister/:id", h.UnregisterNode)

	router.GET("/status/:status", h.GetNodesByStatus)
	router.GET("/role/:role", h.GetNodesByRole)
	router.GET("/tags/:tag", h.GetNodesByTag)

	router.GET("/stats", h.GetNodeStats)
	router.POST("/batch", h.BatchOperation)
	router.GET("/monitor", h.GetNodeMonitor)
	router.GET("/events", h.GetNodeEvents)
	router.POST("/maintenance/:id", h.SetMaintenanceMode)
	router.POST("/cleanup", h.CleanupOfflineNodes)
}

// GetNodes 获取节点列表
func (h *NodeHandler) GetNodes(c *gin.Context) {
	// 解析查询参数
	params := models.NodeQueryParams{
		Status:     c.Query("status"),
		Role:       c.Query("role"),
		Search:     c.Query("search"),
		OnlineOnly: c.Query("onlineOnly") == "true",
	}

	// 解析标签
	if tagsStr := c.Query("tags"); tagsStr != "" {
		params.Tags = strings.Split(tagsStr, ",")
	}

	// 解析分页参数
	if page, err := strconv.Atoi(c.Query("page")); err == nil {
		params.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("pageSize")); err == nil {
		params.PageSize = pageSize
	}

	// 解析排序参数
	params.SortBy = c.Query("sortBy")
	params.SortDesc = c.Query("sortDesc") == "true"

	// 查询节点
	nodes, total, err := h.NodeRepo.List(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询节点失败",
			"error":   err.Error(),
		})
		return
	}

	// 计算分页信息
	totalPages := (int(total) + params.PageSize - 1) / params.PageSize

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取节点列表成功",
		"data": gin.H{
			"items":      nodes,
			"total":      total,
			"page":       params.Page,
			"pageSize":   params.PageSize,
			"totalPages": totalPages,
		},
	})
}

// CreateNode 创建节点（手动添加）
func (h *NodeHandler) CreateNode(c *gin.Context) {
	var req models.NodeRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 注册节点
	resp, err := h.NodeManager.RegisterNode(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "创建节点失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建节点成功",
		"data":    resp,
	})
}

// GetNode 获取节点详情
func (h *NodeHandler) GetNode(c *gin.Context) {
	id := c.Param("id")

	node, err := h.NodeRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "节点不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取节点详情成功",
		"data":    node,
	})
}

// UpdateNode 更新节点信息
func (h *NodeHandler) UpdateNode(c *gin.Context) {
	id := c.Param("id")

	// 获取原节点信息
	node, err := h.NodeRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "节点不存在",
			"error":   err.Error(),
		})
		return
	}

	// 绑定更新数据
	var updateReq struct {
		Name   string             `json:"name,omitempty"`
		Role   string             `json:"role,omitempty"`
		Status string             `json:"status,omitempty"`
		Tags   []string           `json:"tags,omitempty"`
		Config *models.NodeConfig `json:"config,omitempty"`
	}

	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 更新字段
	if updateReq.Name != "" {
		node.Name = updateReq.Name
	}
	if updateReq.Role != "" {
		node.Type = models.NodeType(updateReq.Role)
	}
	if updateReq.Status != "" {
		node.Status = models.NodeStatus(updateReq.Status)
	}
	if updateReq.Tags != nil {
		node.Tags = updateReq.Tags
	}
	if updateReq.Config != nil {
		node.Config = *updateReq.Config
	}

	// 保存更新
	if err := h.NodeRepo.Update(c.Request.Context(), node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新节点失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新节点成功",
		"data":    node,
	})
}

// DeleteNode 删除节点
func (h *NodeHandler) DeleteNode(c *gin.Context) {
	id := c.Param("id")

	err := h.NodeManager.RemoveNode(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除节点失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除节点成功",
	})
}

// UpdateNodeStatus 更新节点状态
func (h *NodeHandler) UpdateNodeStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	if err := h.NodeRepo.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新节点状态失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新节点状态成功",
	})
}

// NodeHeartbeat 节点心跳
func (h *NodeHandler) NodeHeartbeat(c *gin.Context) {
	id := c.Param("id")

	var heartbeat models.NodeHeartbeat
	if err := c.ShouldBindJSON(&heartbeat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "心跳数据格式错误",
			"error":   err.Error(),
		})
		return
	}

	if err := h.NodeManager.UpdateNodeStatus(id, heartbeat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新心跳失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "心跳更新成功",
	})
}

// GetNodeHealth 获取节点健康状态
func (h *NodeHandler) GetNodeHealth(c *gin.Context) {
	id := c.Param("id")

	node, err := h.NodeRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "节点不存在",
			"error":   err.Error(),
		})
		return
	}

	// 计算健康状态
	health := calculateNodeHealth(node)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取节点健康状态成功",
		"data":    health,
	})
}

// UpdateNodeConfig 更新节点配置
func (h *NodeHandler) UpdateNodeConfig(c *gin.Context) {
	id := c.Param("id")

	var config models.NodeConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "配置参数错误",
			"error":   err.Error(),
		})
		return
	}

	if err := h.NodeManager.UpdateNodeConfig(id, config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新节点配置失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新节点配置成功",
	})
}

// GetNodeConfig 获取节点配置
func (h *NodeHandler) GetNodeConfig(c *gin.Context) {
	id := c.Param("id")

	node, err := h.NodeRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "节点不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取节点配置成功",
		"data":    node.Config,
	})
}

// GetNodeTasks 获取节点任务
func (h *NodeHandler) GetNodeTasks(c *gin.Context) {
	// TODO: 实现获取节点任务逻辑
	// 这里需要与任务管理器集成

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取节点任务成功",
		"data":    []interface{}{},
	})
}

// GetNodeTaskStats 获取节点任务统计
func (h *NodeHandler) GetNodeTaskStats(c *gin.Context) {
	id := c.Param("id")

	stats, err := h.NodeRepo.GetNodeTaskStats(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取节点任务统计失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取节点任务统计成功",
		"data":    stats,
	})
}

// RegisterNode 节点注册
func (h *NodeHandler) RegisterNode(c *gin.Context) {
	var req models.NodeRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "注册请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	resp, err := h.NodeManager.RegisterNode(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "节点注册失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "节点注册成功",
		"data":    resp,
	})
}

// UnregisterNode 节点注销
func (h *NodeHandler) UnregisterNode(c *gin.Context) {
	id := c.Param("id")

	if err := h.NodeManager.RemoveNode(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "节点注销失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "节点注销成功",
	})
}

// GetNodesByStatus 按状态获取节点
func (h *NodeHandler) GetNodesByStatus(c *gin.Context) {
	status := c.Param("status")

	nodes, err := h.NodeRepo.GetByStatus(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询节点失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取节点列表成功",
		"data": gin.H{
			"nodes": nodes,
			"total": len(nodes),
		},
	})
}

// GetNodesByRole 按角色获取节点
func (h *NodeHandler) GetNodesByRole(c *gin.Context) {
	role := c.Param("role")

	nodes, err := h.NodeRepo.GetByRole(c.Request.Context(), role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询节点失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取节点列表成功",
		"data": gin.H{
			"nodes": nodes,
			"total": len(nodes),
		},
	})
}

// GetNodesByTag 按标签获取节点
func (h *NodeHandler) GetNodesByTag(c *gin.Context) {
	tag := c.Param("tag")

	nodes, err := h.NodeRepo.GetByTags(c.Request.Context(), []string{tag})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询节点失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取节点列表成功",
		"data": gin.H{
			"nodes": nodes,
			"total": len(nodes),
		},
	})
}

// GetNodeStats 获取节点统计
func (h *NodeHandler) GetNodeStats(c *gin.Context) {
	stats, err := h.NodeRepo.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取节点统计失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取节点统计成功",
		"data":    stats,
	})
}

// BatchOperation 批量操作
func (h *NodeHandler) BatchOperation(c *gin.Context) {
	var req struct {
		Action  string      `json:"action" binding:"required"`
		NodeIds []string    `json:"nodeIds" binding:"required"`
		Data    interface{} `json:"data,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	switch req.Action {
	case "delete":
		err := h.NodeRepo.BatchDelete(c.Request.Context(), req.NodeIds)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "批量删除失败",
				"error":   err.Error(),
			})
			return
		}
	case "updateStatus":
		statusData, ok := req.Data.(map[string]interface{})
		if !ok || statusData["status"] == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "缺少状态参数",
			})
			return
		}
		status := statusData["status"].(string)
		err := h.NodeRepo.BatchUpdateStatus(c.Request.Context(), req.NodeIds, status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "批量更新状态失败",
				"error":   err.Error(),
			})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的操作类型",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "批量操作成功",
	})
}

// GetNodeMonitor 获取节点监控信息
func (h *NodeHandler) GetNodeMonitor(c *gin.Context) {
	// TODO: 实现节点监控逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取节点监控信息成功",
		"data":    gin.H{},
	})
}

// GetNodeEvents 获取节点事件
func (h *NodeHandler) GetNodeEvents(c *gin.Context) {
	// TODO: 实现节点事件查询逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取节点事件成功",
		"data":    []interface{}{},
	})
}

// SetMaintenanceMode 设置维护模式
func (h *NodeHandler) SetMaintenanceMode(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Maintenance bool   `json:"maintenance"`
		Reason      string `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	status := models.NodeStatusMaintenance
	if !req.Maintenance {
		status = models.NodeStatusOnline
	}

	if err := h.NodeRepo.UpdateStatus(c.Request.Context(), id, string(status)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "设置维护模式失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "设置维护模式成功",
	})
}

// CleanupOfflineNodes 清理离线节点
func (h *NodeHandler) CleanupOfflineNodes(c *gin.Context) {
	var req struct {
		TimeoutHours int `json:"timeoutHours"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.TimeoutHours = 24 // 默认24小时
	}

	if req.TimeoutHours <= 0 {
		req.TimeoutHours = 24
	}

	// 清理离线节点
	timeout := time.Hour * time.Duration(req.TimeoutHours)
	count, err := h.NodeRepo.CleanupOfflineNodes(c.Request.Context(), timeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "清理离线节点失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "清理离线节点成功",
		"data": gin.H{
			"cleanedCount": count,
		},
	})
}

// calculateNodeHealth 计算节点健康状态
func calculateNodeHealth(node *models.Node) map[string]interface{} {
	health := map[string]interface{}{
		"nodeId":   node.ID.Hex(),
		"name":     node.Name,
		"status":   node.Status,
		"healthy":  node.Status == models.NodeStatusOnline,
		"lastSeen": node.LastHeartbeatTime,
		"uptime":   node.StatusInfo.UptimeSeconds,
	}

	// 计算健康评分
	score := 100
	if node.Status != models.NodeStatusOnline {
		score = 0
	} else {
		// CPU使用率检查
		if node.StatusInfo.CpuUsage > 90 {
			score -= 30
		} else if node.StatusInfo.CpuUsage > 70 {
			score -= 15
		}

		// 内存使用率检查
		if node.Config.MaxMemoryUsage > 0 {
			memUsagePercent := float64(node.StatusInfo.MemoryUsage) / float64(node.Config.MaxMemoryUsage) * 100
			if memUsagePercent > 90 {
				score -= 30
			} else if memUsagePercent > 70 {
				score -= 15
			}
		}

		// 任务负载检查
		if node.StatusInfo.RunningTasks > node.Config.MaxConcurrentTasks {
			score -= 20
		}
	}

	if score < 0 {
		score = 0
	}

	health["score"] = score
	health["issues"] = generateHealthIssues(node)

	return health
}

// generateHealthIssues 生成健康问题列表
func generateHealthIssues(node *models.Node) []string {
	var issues []string

	if node.Status != models.NodeStatusOnline {
		issues = append(issues, "节点离线")
	}

	if node.StatusInfo.CpuUsage > 90 {
		issues = append(issues, "CPU使用率过高")
	}

	if node.Config.MaxMemoryUsage > 0 {
		memUsagePercent := float64(node.StatusInfo.MemoryUsage) / float64(node.Config.MaxMemoryUsage) * 100
		if memUsagePercent > 90 {
			issues = append(issues, "内存使用率过高")
		}
	}

	if node.StatusInfo.RunningTasks > node.Config.MaxConcurrentTasks {
		issues = append(issues, "任务负载过高")
	}

	return issues
}
