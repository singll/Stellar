package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/StellarServer/internal/plugin"
)

// PluginMarketplaceHandler 插件市场处理器
type PluginMarketplaceHandler struct {
	marketplace *plugin.Marketplace
}

// NewPluginMarketplaceHandler 创建插件市场处理器
func NewPluginMarketplaceHandler(marketplace *plugin.Marketplace) *PluginMarketplaceHandler {
	return &PluginMarketplaceHandler{
		marketplace: marketplace,
	}
}

// SearchPluginsRequest 搜索插件请求
type SearchPluginsRequest struct {
	Query    string   `form:"query"`
	Category string   `form:"category"`
	Tags     []string `form:"tags"`
	Page     int      `form:"page,default=1"`
	PageSize int      `form:"page_size,default=20"`
}

// SearchPluginsResponse 搜索插件响应
type SearchPluginsResponse struct {
	Plugins    []*plugin.PluginPackage `json:"plugins"`
	Total      int                     `json:"total"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
	TotalPages int                     `json:"total_pages"`
}

// GetPluginResponse 获取插件响应
type GetPluginResponse struct {
	Plugin *plugin.PluginPackage `json:"plugin"`
}

// MarketplaceInstallRequest 从市场安装插件请求
type MarketplaceInstallRequest struct {
	PluginID string `json:"plugin_id" binding:"required"`
}

// InstallPluginResponse 安装插件响应
type InstallPluginResponse struct {
	Status *plugin.InstallStatus `json:"status"`
}

// MarketplaceStatsResponse 市场统计响应
type MarketplaceStatsResponse struct {
	Stats map[string]interface{} `json:"stats"`
}

// CategoriesResponse 分类响应
type CategoriesResponse struct {
	Categories []string `json:"categories"`
}

// TagsResponse 标签响应
type TagsResponse struct {
	Tags []string `json:"tags"`
}

// SearchPlugins 搜索插件
// @Summary 搜索插件
// @Description 根据查询条件搜索插件
// @Tags 插件市场
// @Accept json
// @Produce json
// @Param query query string false "搜索查询"
// @Param category query string false "插件分类"
// @Param tags query []string false "插件标签"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} SearchPluginsResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/marketplace/plugins/search [get]
func (h *PluginMarketplaceHandler) SearchPlugins(c *gin.Context) {
	var req SearchPluginsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "VALIDATION_ERROR",
				"code":    "INVALID_PARAMS", 
				"message": err.Error(),
			},
		})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 解析标签
	if tagsParam := c.Query("tags"); tagsParam != "" {
		req.Tags = strings.Split(tagsParam, ",")
	}

	// 搜索插件
	plugins, err := h.marketplace.Search(req.Query, req.Category, req.Tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "INTERNAL_ERROR",
				"code":    "SEARCH_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	// 分页处理
	total := len(plugins)
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var pagedPlugins []*plugin.PluginPackage
	if start < end {
		pagedPlugins = plugins[start:end]
	} else {
		pagedPlugins = []*plugin.PluginPackage{}
	}

	totalPages := (total + req.PageSize - 1) / req.PageSize

	c.JSON(http.StatusOK, SearchPluginsResponse{
		Plugins:    pagedPlugins,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	})
}

// GetPlugin 获取插件详情
// @Summary 获取插件详情
// @Description 根据ID获取插件详细信息
// @Tags 插件市场
// @Accept json
// @Produce json
// @Param id path string true "插件ID"
// @Success 200 {object} GetPluginResponse
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/marketplace/plugins/{id} [get]
func (h *PluginMarketplaceHandler) GetPlugin(c *gin.Context) {
	pluginID := c.Param("id")
	if pluginID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "VALIDATION_ERROR",
				"code":    "INVALID_PARAMS",
				"message": "插件ID不能为空",
			},
		})
		return
	}

	plugin, err := h.marketplace.GetPlugin(pluginID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"type":    "NOT_FOUND_ERROR",
				"code":    "PLUGIN_NOT_FOUND",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, GetPluginResponse{
		Plugin: plugin,
	})
}

// InstallPlugin 安装插件
// @Summary 安装插件
// @Description 安装指定的插件
// @Tags 插件市场
// @Accept json
// @Produce json
// @Param request body MarketplaceInstallRequest true "安装插件请求"
// @Success 200 {object} InstallPluginResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/marketplace/plugins/install [post]
func (h *PluginMarketplaceHandler) InstallPlugin(c *gin.Context) {
	var req MarketplaceInstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "VALIDATION_ERROR",
				"code":    "INVALID_PARAMS", 
				"message": err.Error(),
			},
		})
		return
	}

	status, err := h.marketplace.Install(req.PluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "INTERNAL_ERROR",
				"code":    "INSTALL_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, InstallPluginResponse{
		Status: status,
	})
}

// UpdateIndex 更新插件索引
// @Summary 更新插件索引
// @Description 从仓库更新插件索引
// @Tags 插件市场
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/marketplace/index/update [post]
func (h *PluginMarketplaceHandler) UpdateIndex(c *gin.Context) {
	err := h.marketplace.UpdateIndex()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "INTERNAL_ERROR",
				"code":    "UPDATE_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "索引更新成功",
	})
}

// GetStats 获取市场统计
// @Summary 获取市场统计
// @Description 获取插件市场的统计信息
// @Tags 插件市场
// @Accept json
// @Produce json
// @Success 200 {object} MarketplaceStatsResponse
// @Failure 500 {object} gin.H
// @Router /api/v1/marketplace/stats [get]
func (h *PluginMarketplaceHandler) GetStats(c *gin.Context) {
	stats := h.marketplace.GetStats()
	
	c.JSON(http.StatusOK, MarketplaceStatsResponse{
		Stats: stats,
	})
}

// GetCategories 获取插件分类
// @Summary 获取插件分类
// @Description 获取所有可用的插件分类
// @Tags 插件市场
// @Accept json
// @Produce json
// @Success 200 {object} CategoriesResponse
// @Failure 500 {object} gin.H
// @Router /api/v1/marketplace/categories [get]
func (h *PluginMarketplaceHandler) GetCategories(c *gin.Context) {
	categories := h.marketplace.GetCategories()
	
	c.JSON(http.StatusOK, CategoriesResponse{
		Categories: categories,
	})
}

// GetTags 获取插件标签
// @Summary 获取插件标签
// @Description 获取所有可用的插件标签
// @Tags 插件市场
// @Accept json
// @Produce json
// @Success 200 {object} TagsResponse
// @Failure 500 {object} gin.H
// @Router /api/v1/marketplace/tags [get]
func (h *PluginMarketplaceHandler) GetTags(c *gin.Context) {
	tags := h.marketplace.GetTags()
	
	c.JSON(http.StatusOK, TagsResponse{
		Tags: tags,
	})
}

// DownloadPlugin 下载插件
// @Summary 下载插件
// @Description 下载指定插件的安装包
// @Tags 插件市场
// @Accept json
// @Produce application/octet-stream
// @Param id path string true "插件ID"
// @Success 200 {file} binary
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/marketplace/plugins/{id}/download [get]
func (h *PluginMarketplaceHandler) DownloadPlugin(c *gin.Context) {
	pluginID := c.Param("id")
	if pluginID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "VALIDATION_ERROR",
				"code":    "INVALID_PARAMS",
				"message": "插件ID不能为空",
			},
		})
		return
	}

	// 获取插件信息
	plugin, err := h.marketplace.GetPlugin(pluginID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"type":    "NOT_FOUND_ERROR",
				"code":    "PLUGIN_NOT_FOUND",
				"message": err.Error(),
			},
		})
		return
	}

	// 下载插件
	downloadPath, err := h.marketplace.Download(plugin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "INTERNAL_ERROR",
				"code":    "DOWNLOAD_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	// 设置响应头
	c.Header("Content-Disposition", "attachment; filename="+plugin.ID+"_"+plugin.Version+".zip")
	c.Header("Content-Type", "application/octet-stream")
	
	// 返回文件
	c.File(downloadPath)
}