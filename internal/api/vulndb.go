package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/services/vulndb"
)

// VulnDatabaseHandler 漏洞数据库API处理器
type VulnDatabaseHandler struct {
	service   *vulndb.Service
	scheduler *vulndb.Scheduler
}

// NewVulnDatabaseHandler 创建处理器
func NewVulnDatabaseHandler(service *vulndb.Service, scheduler *vulndb.Scheduler) *VulnDatabaseHandler {
	return &VulnDatabaseHandler{
		service:   service,
		scheduler: scheduler,
	}
}

// GetStats 获取漏洞数据库统计信息
// @Summary 获取漏洞数据库统计信息
// @Description 获取漏洞数据库的各项统计数据
// @Tags 漏洞数据库
// @Accept json
// @Produce json
// @Success 200 {object} models.VulnDbStats
// @Failure 500 {object} gin.H
// @Router /api/v1/vulndb/stats [get]
func (h *VulnDatabaseHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetVulnerabilityStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取统计信息失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SearchVulnerabilities 搜索漏洞信息
// @Summary 搜索漏洞信息
// @Description 根据条件搜索漏洞数据库
// @Tags 漏洞数据库
// @Accept json
// @Produce json
// @Param cve_id query string false "CVE编号"
// @Param cwe_id query string false "CWE编号"
// @Param cnvd_id query string false "CNVD编号"
// @Param keyword query string false "关键词"
// @Param severity query []string false "严重程度"
// @Param source query []string false "数据源"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "页面大小" default(20)
// @Success 200 {object} VulnSearchResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/vulndb/search [get]
func (h *VulnDatabaseHandler) SearchVulnerabilities(c *gin.Context) {
	var query models.VulnDbQuery

	// 解析查询参数
	query.CVEId = c.Query("cve_id")
	query.CWEId = c.Query("cwe_id")
	query.CNVDId = c.Query("cnvd_id")
	query.Keyword = c.Query("keyword")
	
	// 解析数组参数
	if severities, exists := c.GetQueryArray("severity"); exists {
		query.Severity = severities
	}
	if sources, exists := c.GetQueryArray("source"); exists {
		query.Source = sources
	}

	// 解析分页参数
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		query.Page = page
	}
	if pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20")); err == nil {
		query.PageSize = pageSize
	}

	// 解析排序参数
	query.SortBy = c.DefaultQuery("sort_by", "published_date")
	query.SortDesc = c.DefaultQuery("sort_desc", "true") == "true"

	// 执行搜索
	results, total, err := h.service.SearchVulnerabilities(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "搜索失败",
			"message": err.Error(),
		})
		return
	}

	response := VulnSearchResponse{
		Results:     results,
		Total:       total,
		Page:        query.Page,
		PageSize:    query.PageSize,
		TotalPages:  (total + int64(query.PageSize) - 1) / int64(query.PageSize),
	}

	c.JSON(http.StatusOK, response)
}

// UpdateDatabase 更新漏洞数据库
// @Summary 手动更新漏洞数据库
// @Description 手动触发漏洞数据库更新
// @Tags 漏洞数据库
// @Accept json
// @Produce json
// @Param source query string false "数据源" Enums(cve, cwe, cnvd, all)
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} gin.H
// @Router /api/v1/vulndb/update [post]
func (h *VulnDatabaseHandler) UpdateDatabase(c *gin.Context) {
	source := c.DefaultQuery("source", "all")
	
	var err error
	switch source {
	case "cve":
		err = h.service.UpdateCVEDatabase(c.Request.Context())
	case "cwe":
		err = h.service.UpdateCWEDatabase(c.Request.Context())
	case "cnvd":
		err = h.service.UpdateCNVDDatabase(c.Request.Context())
	case "all":
		err = h.scheduler.ForceUpdate()
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的数据源",
			"message": "支持的数据源: cve, cwe, cnvd, all",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新已开始",
	})
}

// GetVulnerability 获取单个漏洞详情
// @Summary 获取漏洞详情
// @Description 根据ID获取漏洞详细信息
// @Tags 漏洞数据库
// @Accept json
// @Produce json
// @Param id path string true "漏洞ID"
// @Success 200 {object} models.VulnDbInfo
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/vulndb/vulnerability/{id} [get]
func (h *VulnDatabaseHandler) GetVulnerability(c *gin.Context) {
	// TODO: 实现获取单个漏洞详情
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "功能未实现",
		"message": "TODO: 实现获取单个漏洞详情",
	})
}

// 响应结构体
type VulnSearchResponse struct {
	Results    []models.VulnDbInfo `json:"results"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int64               `json:"total_pages"`
}