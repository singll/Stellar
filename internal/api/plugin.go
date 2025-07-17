package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/plugin"
)

// PluginHandler 处理插件相关的API请求
type PluginHandler struct {
	manager *plugin.Manager
	store   plugin.MetadataStore
}

// NewPluginHandler 创建一个新的插件处理器
func NewPluginHandler(manager *plugin.Manager, store plugin.MetadataStore) *PluginHandler {
	return &PluginHandler{
		manager: manager,
		store:   store,
	}
}

// RegisterRoutes 注册插件相关的路由
func (h *PluginHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("", h.ListPlugins)
	router.GET("/:id", h.GetPlugin)
	router.POST("", h.InstallPlugin)
	router.DELETE("/:id", h.UninstallPlugin)
	router.PUT("/:id/enable", h.EnablePlugin)
	router.PUT("/:id/disable", h.DisablePlugin)
	router.PUT("/:id/config", h.UpdatePluginConfig)
	router.GET("/market", h.ListMarketPlugins)
	router.GET("/types", h.ListPluginTypes)
	router.GET("/categories", h.ListPluginCategories)
}

// ListPlugins 列出所有已安装的插件
func (h *PluginHandler) ListPlugins(c *gin.Context) {
	pluginType := c.Query("type")
	category := c.Query("category")
	tag := c.Query("tag")

	var plugins []*models.PluginMetadata
	var err error

	if pluginType != "" {
		plugins, err = h.store.ListPluginsByType(pluginType)
	} else if category != "" {
		plugins, err = h.store.ListPluginsByCategory(category)
	} else if tag != "" {
		plugins, err = h.store.ListPluginsByTag(tag)
	} else {
		plugins, err = h.store.ListPlugins()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plugins)
}

// GetPlugin 获取指定插件的详细信息
func (h *PluginHandler) GetPlugin(c *gin.Context) {
	id := c.Param("id")
	metadata, err := h.store.GetPluginMetadata(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	c.JSON(http.StatusOK, metadata)
}

// InstallPluginRequest 安装插件的请求
type InstallPluginRequest struct {
	Source      string                 `json:"source" binding:"required"`
	SourceType  string                 `json:"source_type" binding:"required"`
	Config      map[string]interface{} `json:"config"`
	AutoEnable  bool                   `json:"auto_enable"`
	ForceUpdate bool                   `json:"force_update"`
}

// InstallPlugin 安装一个新插件
func (h *PluginHandler) InstallPlugin(c *gin.Context) {
	var req InstallPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证源类型
	if req.SourceType != "file" && req.SourceType != "url" && req.SourceType != "market" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source type"})
		return
	}

	// 根据源类型处理插件安装
	var pluginID string
	var err error

	switch req.SourceType {
	case "file":
		// 从本地文件安装
		pluginID, err = h.manager.InstallPluginFromFile(req.Source, req.Config, req.ForceUpdate)
	case "url":
		// 从URL安装
		pluginID, err = h.manager.InstallPluginFromURL(req.Source, req.Config, req.ForceUpdate)
	case "market":
		// 从插件市场安装
		pluginID, err = h.manager.InstallPluginFromMarket(req.Source, req.Config, req.ForceUpdate)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 如果需要自动启用
	if req.AutoEnable {
		if err := h.manager.EnablePlugin(pluginID); err != nil {
			// 仅记录错误，不影响安装结果
			c.JSON(http.StatusOK, gin.H{
				"id":      pluginID,
				"status":  "installed",
				"enabled": false,
				"message": "Plugin installed but failed to enable: " + err.Error(),
			})
			return
		}
	}

	metadata, _ := h.store.GetPluginMetadata(pluginID)
	c.JSON(http.StatusOK, gin.H{
		"id":       pluginID,
		"status":   "installed",
		"enabled":  req.AutoEnable,
		"metadata": metadata,
	})
}

// UninstallPlugin 卸载指定的插件
func (h *PluginHandler) UninstallPlugin(c *gin.Context) {
	id := c.Param("id")

	// 检查插件是否存在
	_, err := h.store.GetPluginMetadata(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	// 卸载插件
	if err := h.manager.UninstallPlugin(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "uninstalled"})
}

// EnablePlugin 启用指定的插件
func (h *PluginHandler) EnablePlugin(c *gin.Context) {
	id := c.Param("id")

	// 检查插件是否存在
	metadata, err := h.store.GetPluginMetadata(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	// 如果插件已经启用，直接返回成功
	if metadata.Enabled {
		c.JSON(http.StatusOK, gin.H{"status": "already enabled"})
		return
	}

	// 启用插件
	if err := h.manager.EnablePlugin(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "enabled"})
}

// DisablePlugin 禁用指定的插件
func (h *PluginHandler) DisablePlugin(c *gin.Context) {
	id := c.Param("id")

	// 检查插件是否存在
	metadata, err := h.store.GetPluginMetadata(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	// 如果插件已经禁用，直接返回成功
	if !metadata.Enabled {
		c.JSON(http.StatusOK, gin.H{"status": "already disabled"})
		return
	}

	// 禁用插件
	if err := h.manager.DisablePlugin(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "disabled"})
}

// UpdatePluginConfigRequest 更新插件配置的请求
type UpdatePluginConfigRequest struct {
	Config map[string]interface{} `json:"config" binding:"required"`
}

// UpdatePluginConfig 更新插件配置
func (h *PluginHandler) UpdatePluginConfig(c *gin.Context) {
	id := c.Param("id")

	var req UpdatePluginConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查插件是否存在
	metadata, err := h.store.GetPluginMetadata(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	// 更新配置
	metadata.Config = req.Config
	metadata.UpdateTime = time.Now()

	// 保存更新后的元数据
	if err := h.store.SavePluginMetadata(metadata); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 如果插件已启用，需要重新加载
	if metadata.Enabled {
		if err := h.manager.ReloadPlugin(id); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status":  "config updated",
				"message": "Config updated but failed to reload plugin: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "config updated",
		"config": metadata.Config,
	})
}

// MarketPlugin 插件市场中的插件信息
type MarketPlugin struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Author      string                 `json:"author"`
	Type        string                 `json:"type"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Source      string                 `json:"source"`
	SourceType  string                 `json:"source_type"`
	Config      map[string]interface{} `json:"config"`
	Installed   bool                   `json:"installed"`
}

// ListMarketPlugins 列出插件市场中的插件
func (h *PluginHandler) ListMarketPlugins(c *gin.Context) {
	// 这里应该实现从插件市场获取插件列表的逻辑
	// 目前先返回一些示例数据
	marketPlugins := []MarketPlugin{
		{
			ID:          "com.stellar.plugin.subdomain-finder",
			Name:        "Subdomain Finder",
			Description: "Advanced subdomain enumeration plugin",
			Version:     "1.0.0",
			Author:      "Stellar Team",
			Type:        "subdomain",
			Category:    "discovery",
			Tags:        []string{"subdomain", "discovery", "recon"},
			Source:      "https://plugins.stellarserver.io/subdomain-finder-1.0.0.zip",
			SourceType:  "url",
			Config:      map[string]interface{}{"threads": 10},
			Installed:   false,
		},
		{
			ID:          "com.stellar.plugin.port-scanner",
			Name:        "Advanced Port Scanner",
			Description: "Fast and accurate port scanner",
			Version:     "1.0.0",
			Author:      "Stellar Team",
			Type:        "portscan",
			Category:    "discovery",
			Tags:        []string{"port", "scanner", "discovery"},
			Source:      "https://plugins.stellarserver.io/port-scanner-1.0.0.zip",
			SourceType:  "url",
			Config:      map[string]interface{}{"threads": 100, "timeout": 5},
			Installed:   false,
		},
	}

	// 检查插件是否已安装
	installedPlugins, _ := h.store.ListPlugins()
	installedMap := make(map[string]bool)
	for _, p := range installedPlugins {
		installedMap[p.ID] = true
	}

	for i := range marketPlugins {
		marketPlugins[i].Installed = installedMap[marketPlugins[i].ID]
	}

	c.JSON(http.StatusOK, marketPlugins)
}

// ListPluginTypes 列出所有支持的插件类型
func (h *PluginHandler) ListPluginTypes(c *gin.Context) {
	types := []string{
		"vulnerability",
		"discovery",
		"subdomain",
		"portscan",
		"sensitive",
		"monitoring",
		"utility",
	}

	c.JSON(http.StatusOK, types)
}

// ListPluginCategories 列出所有支持的插件分类
func (h *PluginHandler) ListPluginCategories(c *gin.Context) {
	categories := []string{
		"scanner",
		"discovery",
		"monitoring",
		"utility",
		"integration",
	}

	c.JSON(http.StatusOK, categories)
}
