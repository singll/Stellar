package api

import (
	"github.com/StellarServer/internal/services/nodemanager"
	"github.com/gin-gonic/gin"
)

// NodeHandler 节点处理器
type NodeHandler struct {
	NodeManager *nodemanager.NodeManager
}

// NewNodeHandler 创建节点处理器
func NewNodeHandler(nodeManager *nodemanager.NodeManager) *NodeHandler {
	return &NodeHandler{
		NodeManager: nodeManager,
	}
}

// RegisterRoutes 注册路由
func (h *NodeHandler) RegisterRoutes(router *gin.RouterGroup) {
	nodeGroup := router.Group("/nodes")
	{
		nodeGroup.GET("", h.GetAllNodes)
		nodeGroup.GET("/:id", h.GetNode)
		nodeGroup.DELETE("/:id", h.RemoveNode)
		nodeGroup.PUT("/:id/config", h.UpdateNodeConfig)
		nodeGroup.GET("/status/:status", h.GetNodesByStatus)
		nodeGroup.GET("/role/:role", h.GetNodesByRole)
	}
}

// GetAllNodes 获取所有节点
func (h *NodeHandler) GetAllNodes(c *gin.Context) {
	nodes := h.NodeManager.GetAllNodes()
	c.JSON(200, gin.H{
		"nodes": nodes,
		"total": len(nodes),
	})
}

// GetNode 获取节点
func (h *NodeHandler) GetNode(c *gin.Context) {
	id := c.Param("id")
	node, err := h.NodeManager.GetNode(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "节点不存在"})
		return
	}
	c.JSON(200, node)
}

// RemoveNode 移除节点
func (h *NodeHandler) RemoveNode(c *gin.Context) {
	id := c.Param("id")
	err := h.NodeManager.RemoveNode(id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "节点已移除"})
}

// UpdateNodeConfig 更新节点配置
func (h *NodeHandler) UpdateNodeConfig(c *gin.Context) {
	// TODO: 实现更新节点配置
	c.JSON(200, gin.H{"message": "功能待实现"})
}

// GetNodesByStatus 按状态获取节点
func (h *NodeHandler) GetNodesByStatus(c *gin.Context) {
	status := c.Param("status")
	nodes := h.NodeManager.GetNodesByStatus(status)
	c.JSON(200, gin.H{
		"nodes": nodes,
		"total": len(nodes),
	})
}

// GetNodesByRole 按角色获取节点
func (h *NodeHandler) GetNodesByRole(c *gin.Context) {
	role := c.Param("role")
	nodes := h.NodeManager.GetNodesByRole(role)
	c.JSON(200, gin.H{
		"nodes": nodes,
		"total": len(nodes),
	})
}
