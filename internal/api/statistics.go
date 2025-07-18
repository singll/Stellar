package api

import (
	"net/http"
	"time"

	pkgerrors "github.com/StellarServer/internal/pkg/errors"
	"github.com/StellarServer/internal/pkg/logger"
	"github.com/StellarServer/internal/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// StatisticsResponse 统计响应结构
type StatisticsResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

// StatisticsAPI 统计API结构
type StatisticsAPI struct {
	db *mongo.Database
}

// NewStatisticsAPI 创建统计API实例
func NewStatisticsAPI(db *mongo.Database) *StatisticsAPI {
	return &StatisticsAPI{
		db: db,
	}
}

// DashboardStats 获取仪表盘统计数据
func (api *StatisticsAPI) DashboardStats(c *gin.Context) {
	// 获取统计数据
	stats, err := api.getDashboardStats(c)
	if err != nil {
		logger.Error("Failed to get dashboard stats", map[string]interface{}{"error": err})
		utils.HandleError(c, pkgerrors.NewInternalError("获取统计数据失败"))
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": stats,
	})
}

// 获取仪表盘统计数据
func (api *StatisticsAPI) getDashboardStats(c *gin.Context) (map[string]interface{}, error) {
	// 统计资产数量
	assetCount, err := api.db.Collection("asset").CountDocuments(c, bson.M{})
	if err != nil {
		return nil, pkgerrors.NewInternalError("Failed to count assets")
	}

	// 统计漏洞数量
	vulnCount, err := api.db.Collection("vulnerability").CountDocuments(c, bson.M{})
	if err != nil {
		return nil, pkgerrors.NewInternalError("Failed to count vulnerabilities")
	}

	// 统计项目数量
	projectCount, err := api.db.Collection("project").CountDocuments(c, bson.M{})
	if err != nil {
		return nil, pkgerrors.NewInternalError("Failed to count projects")
	}

	// 统计今日任务数量
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	todayTaskCount, err := api.db.Collection("task").CountDocuments(c, bson.M{
		"createTime": bson.M{
			"$gte": today,
			"$lt":  tomorrow,
		},
	})
	if err != nil {
		return nil, pkgerrors.NewInternalError("Failed to count today's tasks")
	}

	// 统计资产类型分布
	assetTypes, err := api.getAssetTypeDistribution(c)
	if err != nil {
		return nil, pkgerrors.NewInternalError("Failed to get asset type distribution")
	}

	// 统计漏洞风险等级分布
	vulnLevels, err := api.getVulnerabilityLevelDistribution(c)
	if err != nil {
		return nil, pkgerrors.NewInternalError("Failed to get vulnerability level distribution")
	}

	// 统计任务执行趋势
	taskTrend, err := api.getTaskExecutionTrend(c)
	if err != nil {
		return nil, pkgerrors.NewInternalError("Failed to get task execution trend")
	}

	// 获取最近漏洞
	recentVulns, err := api.getRecentVulnerabilities(c)
	if err != nil {
		return nil, pkgerrors.NewInternalError("Failed to get recent vulnerabilities")
	}

	// 获取最近任务
	recentTasks, err := api.getRecentTasks(c)
	if err != nil {
		return nil, pkgerrors.NewInternalError("Failed to get recent tasks")
	}

	return map[string]interface{}{
		"summary": map[string]interface{}{
			"assetCount":     assetCount,
			"vulnCount":      vulnCount,
			"projectCount":   projectCount,
			"todayTaskCount": todayTaskCount,
		},
		"assetTypes":  assetTypes,
		"vulnLevels":  vulnLevels,
		"taskTrend":   taskTrend,
		"recentVulns": recentVulns,
		"recentTasks": recentTasks,
	}, nil
}

// 获取资产类型分布
func (api *StatisticsAPI) getAssetTypeDistribution(c *gin.Context) ([]map[string]interface{}, error) {
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": "$type",
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
		{
			"$project": bson.M{
				"_id":   0,
				"name":  "$_id",
				"value": "$count",
			},
		},
	}

	cursor, err := api.db.Collection("asset").Aggregate(c, pipeline)
	if err != nil {
		return nil, pkgerrors.NewInternalError("Failed to aggregate asset types")
	}

	var results []map[string]interface{}
	if err = cursor.All(c, &results); err != nil {
		return nil, pkgerrors.NewInternalError("Failed to decode asset type results")
	}

	return results, nil
}

// 获取漏洞风险等级分布
func (api *StatisticsAPI) getVulnerabilityLevelDistribution(c *gin.Context) ([]map[string]interface{}, error) {
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": "$level",
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
		{
			"$project": bson.M{
				"_id":   0,
				"name":  "$_id",
				"value": "$count",
			},
		},
	}

	cursor, err := api.db.Collection("vulnerability").Aggregate(c, pipeline)
	if err != nil {
		return nil, pkgerrors.NewInternalError("Failed to aggregate vulnerability levels")
	}

	var results []map[string]interface{}
	if err = cursor.All(c, &results); err != nil {
		return nil, pkgerrors.NewInternalError("Failed to decode vulnerability level results")
	}

	return results, nil
}

// 获取任务执行趋势
func (api *StatisticsAPI) getTaskExecutionTrend(c *gin.Context) (map[string]interface{}, error) {
	// 获取最近7天的日期
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -6)

	// 准备日期列表
	dates := []string{}
	for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format("01/02"))
	}

	// 查询每天的任务状态
	completed := []int64{}
	inProgress := []int64{}
	failed := []int64{}

	for i := 0; i < 7; i++ {
		day := startDate.AddDate(0, 0, i)
		nextDay := day.AddDate(0, 0, 1)

		// 查询已完成任务
		completedCount, err := api.db.Collection("task").CountDocuments(c, bson.M{
			"createTime": bson.M{
				"$gte": day,
				"$lt":  nextDay,
			},
			"status": "已完成",
		})
		if err != nil {
			return nil, pkgerrors.NewInternalError("Failed to count completed tasks for trend")
		}
		completed = append(completed, completedCount)

		// 查询进行中任务
		inProgressCount, err := api.db.Collection("task").CountDocuments(c, bson.M{
			"createTime": bson.M{
				"$gte": day,
				"$lt":  nextDay,
			},
			"status": bson.M{
				"$in": []string{"进行中", "排队中"},
			},
		})
		if err != nil {
			return nil, pkgerrors.NewInternalError("Failed to count in-progress tasks for trend")
		}
		inProgress = append(inProgress, inProgressCount)

		// 查询失败任务
		failedCount, err := api.db.Collection("task").CountDocuments(c, bson.M{
			"createTime": bson.M{
				"$gte": day,
				"$lt":  nextDay,
			},
			"status": "失败",
		})
		if err != nil {
			return nil, pkgerrors.NewInternalError("Failed to count failed tasks for trend")
		}
		failed = append(failed, failedCount)
	}

	return map[string]interface{}{
		"dates":      dates,
		"completed":  completed,
		"inProgress": inProgress,
		"failed":     failed,
	}, nil
}

// 获取最近漏洞
func (api *StatisticsAPI) getRecentVulnerabilities(c *gin.Context) ([]map[string]interface{}, error) {
	opts := options.Find().SetSort(bson.M{"createTime": -1}).SetLimit(5)
	cursor, err := api.db.Collection("vulnerability").Find(c, bson.M{}, opts)
	if err != nil {
		return nil, pkgerrors.NewInternalError("Failed to find recent vulnerabilities")
	}

	var vulns []map[string]interface{}
	if err = cursor.All(c, &vulns); err != nil {
		return nil, pkgerrors.NewInternalError("Failed to decode recent vulnerabilities")
	}

	// 处理结果
	result := []map[string]interface{}{}
	for _, vuln := range vulns {
		result = append(result, map[string]interface{}{
			"name":  vuln["name"],
			"level": vuln["level"],
			"asset": vuln["target"],
			"date":  formatTime(vuln["createTime"]),
		})
	}

	return result, nil
}

// 获取最近任务
func (api *StatisticsAPI) getRecentTasks(c *gin.Context) ([]map[string]interface{}, error) {
	opts := options.Find().SetSort(bson.M{"createTime": -1}).SetLimit(5)
	cursor, err := api.db.Collection("task").Find(c, bson.M{}, opts)
	if err != nil {
		return nil, pkgerrors.NewInternalError("Failed to find recent tasks")
	}

	var tasks []map[string]interface{}
	if err = cursor.All(c, &tasks); err != nil {
		return nil, pkgerrors.NewInternalError("Failed to decode recent tasks")
	}

	// 处理结果
	result := []map[string]interface{}{}
	for _, task := range tasks {
		result = append(result, map[string]interface{}{
			"name":   task["name"],
			"type":   task["type"],
			"status": task["status"],
			"date":   formatTime(task["createTime"]),
		})
	}

	return result, nil
}

// 格式化时间
func formatTime(t interface{}) string {
	if t == nil {
		return ""
	}

	switch v := t.(type) {
	case time.Time:
		return v.Format("2006-01-02")
	case primitive.DateTime:
		return time.Unix(int64(v)/1000, 0).Format("2006-01-02")
	default:
		return ""
	}
}

// AssetRelationship 获取资产关系图数据
func (api *StatisticsAPI) AssetRelationship(c *gin.Context) {
	projectId := c.Query("projectId")
	if projectId == "" {
		// 兼容 POST 方式
		var req struct {
			ProjectID string `json:"projectId"`
		}
		_ = c.ShouldBindJSON(&req)
		projectId = req.ProjectID
	}

	// 查询条件
	filter := bson.M{}
	if projectId != "" {
		filter["projectId"] = projectId
	}

	// 查询资产
	cursor, err := api.db.Collection("asset").Find(c, filter)
	if err != nil {
		logger.Error("Failed to find assets for relationship", map[string]interface{}{"error": err})
		utils.HandleError(c, pkgerrors.NewInternalError("获取资产数据失败"))
		return
	}

	var assets []map[string]interface{}
	if err = cursor.All(c, &assets); err != nil {
		logger.Error("Failed to decode assets for relationship", map[string]interface{}{"error": err})
		utils.HandleError(c, pkgerrors.NewInternalError("处理资产数据失败"))
		return
	}

	// 构建关系图数据
	nodes := []map[string]interface{}{}
	links := []map[string]interface{}{}
	nodeMap := make(map[string]bool)

	// 添加项目节点
	if projectId != "" {
		var project map[string]interface{}
		err = api.db.Collection("project").FindOne(c, bson.M{"_id": projectId}).Decode(&project)
		if err == nil && project != nil {
			projectNode := map[string]interface{}{
				"id":         projectId,
				"name":       project["name"],
				"category":   "project",
				"symbolSize": 50,
			}
			nodes = append(nodes, projectNode)
			nodeMap[projectId] = true
		}
	}

	// 添加资产节点和关系
	for _, asset := range assets {
		assetID := asset["_id"].(string)
		if !nodeMap[assetID] {
			assetNode := map[string]interface{}{
				"id":         assetID,
				"name":       asset["name"],
				"category":   asset["type"],
				"symbolSize": 30,
			}
			nodes = append(nodes, assetNode)
			nodeMap[assetID] = true
		}

		// 添加与项目的关系
		if projectId != "" {
			links = append(links, map[string]interface{}{
				"source": projectId,
				"target": assetID,
			})
		}

		// 添加资产间的关系
		if parent, ok := asset["parent"].(string); ok && parent != "" {
			if !nodeMap[parent] {
				// 查询父资产信息
				var parentAsset map[string]interface{}
				err = api.db.Collection("asset").FindOne(c, bson.M{"_id": parent}).Decode(&parentAsset)
				if err == nil && parentAsset != nil {
					parentNode := map[string]interface{}{
						"id":         parent,
						"name":       parentAsset["name"],
						"category":   parentAsset["type"],
						"symbolSize": 30,
					}
					nodes = append(nodes, parentNode)
					nodeMap[parent] = true
				}
			}

			links = append(links, map[string]interface{}{
				"source": parent,
				"target": assetID,
			})
		}
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"nodes": nodes,
			"links": links,
		},
	})
}

// VulnerabilityAnalysis 获取漏洞分析数据
func (api *StatisticsAPI) VulnerabilityAnalysis(c *gin.Context) {
	var req struct {
		ProjectID string `json:"projectId"`
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request parameters for vulnerability analysis", map[string]interface{}{"error": err})
		utils.HandleError(c, pkgerrors.NewBadRequestError("无效的请求参数"))
		return
	}

	// 查询条件
	filter := bson.M{}
	if req.ProjectID != "" {
		filter["projectId"] = req.ProjectID
	}

	// 处理时间范围
	if req.StartTime != "" && req.EndTime != "" {
		startTime, err := time.Parse("2006-01-02", req.StartTime)
		if err != nil {
			logger.Error("Invalid start time format for vulnerability analysis", map[string]interface{}{"error": err})
			utils.HandleError(c, pkgerrors.NewBadRequestError("无效的开始时间格式"))
			return
		}

		endTime, err := time.Parse("2006-01-02", req.EndTime)
		if err != nil {
			logger.Error("Invalid end time format for vulnerability analysis", map[string]interface{}{"error": err})
			utils.HandleError(c, pkgerrors.NewBadRequestError("无效的结束时间格式"))
			return
		}
		endTime = endTime.Add(24 * time.Hour)

		filter["createTime"] = bson.M{
			"$gte": startTime,
			"$lt":  endTime,
		}
	}

	// 按风险等级分组统计
	levelPipeline := []bson.M{
		{
			"$match": filter,
		},
		{
			"$group": bson.M{
				"_id": "$level",
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
		{
			"$project": bson.M{
				"_id":   0,
				"name":  "$_id",
				"value": "$count",
			},
		},
	}

	levelCursor, err := api.db.Collection("vulnerability").Aggregate(c, levelPipeline)
	if err != nil {
		logger.Error("Failed to aggregate vulnerability level stats", map[string]interface{}{"error": err})
		utils.HandleError(c, pkgerrors.NewInternalError("获取漏洞等级统计失败"))
		return
	}

	var levelStats []map[string]interface{}
	if err = levelCursor.All(c, &levelStats); err != nil {
		logger.Error("Failed to decode vulnerability level stats", map[string]interface{}{"error": err})
		utils.HandleError(c, pkgerrors.NewInternalError("处理漏洞等级统计失败"))
		return
	}

	// 按漏洞类型分组统计
	typePipeline := []bson.M{
		{
			"$match": filter,
		},
		{
			"$group": bson.M{
				"_id": "$type",
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
		{
			"$project": bson.M{
				"_id":   0,
				"name":  "$_id",
				"value": "$count",
			},
		},
	}

	typeCursor, err := api.db.Collection("vulnerability").Aggregate(c, typePipeline)
	if err != nil {
		logger.Error("Failed to aggregate vulnerability type stats", map[string]interface{}{"error": err})
		utils.HandleError(c, pkgerrors.NewInternalError("获取漏洞类型统计失败"))
		return
	}

	var typeStats []map[string]interface{}
	if err = typeCursor.All(c, &typeStats); err != nil {
		logger.Error("Failed to decode vulnerability type stats", map[string]interface{}{"error": err})
		utils.HandleError(c, pkgerrors.NewInternalError("处理漏洞类型统计失败"))
		return
	}

	// 按时间趋势统计
	var trendStats map[string]interface{}
	if req.StartTime != "" && req.EndTime != "" {
		trendStats, err = api.getVulnerabilityTrend(c, filter)
		if err != nil {
			logger.Error("Failed to aggregate vulnerability trend stats", map[string]interface{}{"error": err})
			utils.HandleError(c, pkgerrors.NewInternalError("获取漏洞趋势统计失败"))
			return
		}
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"levelStats": levelStats,
			"typeStats":  typeStats,
			"trendStats": trendStats,
		},
	})
}

// 获取漏洞趋势统计
func (api *StatisticsAPI) getVulnerabilityTrend(c *gin.Context, filter bson.M) (map[string]interface{}, error) {
	// 获取时间范围
	var startTime, endTime time.Time
	if timeRange, ok := filter["createTime"].(bson.M); ok {
		if start, ok := timeRange["$gte"].(time.Time); ok {
			startTime = start
		}
		if end, ok := timeRange["$lt"].(time.Time); ok {
			endTime = end
		}
	}

	if startTime.IsZero() || endTime.IsZero() {
		return nil, nil
	}

	// 准备日期列表
	dates := []string{}
	for d := startTime; d.Before(endTime); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format("01/02"))
	}

	// 查询每天的漏洞数量
	counts := []int64{}
	for i := 0; i < len(dates); i++ {
		day := startTime.AddDate(0, 0, i)
		nextDay := day.AddDate(0, 0, 1)

		dayFilter := bson.M{}
		for k, v := range filter {
			if k != "createTime" {
				dayFilter[k] = v
			}
		}
		dayFilter["createTime"] = bson.M{
			"$gte": day,
			"$lt":  nextDay,
		}

		count, err := api.db.Collection("vulnerability").CountDocuments(c, dayFilter)
		if err != nil {
			return nil, pkgerrors.NewInternalError("Failed to count vulnerabilities for trend")
		}
		counts = append(counts, count)
	}

	return map[string]interface{}{
		"dates":  dates,
		"counts": counts,
	}, nil
}

// RegisterRoutes 注册统计API路由
func (api *StatisticsAPI) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/dashboard/stats", api.DashboardStats)
	router.GET("/asset/relationship", api.AssetRelationship)
	router.GET("/vulnerability/analysis", api.VulnerabilityAnalysis)
}
