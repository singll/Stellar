package api

import (
	"net/http"
	"strings"

	"github.com/StellarServer/internal/config"
	"github.com/StellarServer/internal/database"
	"github.com/StellarServer/internal/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	DB *mongo.Database
}

// NewProjectHandler 创建项目处理器
func NewProjectHandler(db *mongo.Database) *ProjectHandler {
	return &ProjectHandler{DB: db}
}

// RegisterRoutes 注册路由
func (h *ProjectHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("", h.GetProjects)
	router.GET("/all", h.GetAllProjects)
	router.GET("/search", h.SearchProjects)
	router.GET("/:id", h.GetProjectContent)
	router.POST("", h.AddProject)
	router.PUT("/:id", h.UpdateProject)
	router.DELETE("/:id", h.DeleteProject)
}

// GetProjects 获取项目列表
func (h *ProjectHandler) GetProjects(c *gin.Context) {
	var req struct {
		Search    string `form:"search"`
		PageIndex int    `form:"pageIndex"`
		PageSize  int    `form:"pageSize"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
		})
		return
	}

	if req.PageIndex <= 0 {
		req.PageIndex = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 直接使用 handler 的 h.DB
	db := h.DB

	resultList, tagNum, err := models.GetProjects(db, req.Search, req.PageIndex, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取项目列表失败",
		})
		return
	}

	go models.UpdateProjectAssetCount(db)

	// 获取"All"标签下的项目数据，这是实际的分页数据
	projects := resultList["All"]
	totalCount := tagNum["All"]

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"data":  projects,    // 当前页的项目数据
			"total": totalCount,  // 总项目数
			"page":  req.PageIndex,
			"limit": req.PageSize,
			"tags":  tagNum,      // 保留标签统计信息
		},
	})
}

// GetAllProjects 获取所有项目（用于下拉列表）
func (h *ProjectHandler) GetAllProjects(c *gin.Context) {
	db := h.DB
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": "$tag",
				"children": bson.M{
					"$push": bson.M{
						"value": bson.M{"$toString": "$_id"},
						"label": "$name",
					},
				},
			},
		},
		{
			"$project": bson.M{
				"_id":      0,
				"label":    "$_id",
				"value":    "",
				"children": 1,
			},
		},
	}

	cursor, err := db.Collection("project").Aggregate(c, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取项目列表失败",
		})
		return
	}

	var result []bson.M
	if err = cursor.All(c, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "处理项目列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list": result,
		},
	})
}

// SearchProjects 获取所有项目（用于前端搜索）
func (h *ProjectHandler) SearchProjects(c *gin.Context) {
	var req struct {
		Limit int `form:"limit"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
		})
		return
	}

	// 设置默认限制
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}

	db := h.DB

	// 查询所有项目，不进行服务端搜索过滤
	cursor, err := db.Collection("project").Find(c, bson.M{}, &options.FindOptions{
		Limit: int64Ptr(int64(req.Limit)),
		Sort:  bson.M{"updated": -1}, // 按更新时间倒序
		Projection: bson.M{
			"_id":         1,
			"name":        1,
			"description": 1,
			"tag":         1,
			"created":     1,
			"updated":     1,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取项目列表失败",
		})
		return
	}
	defer cursor.Close(c)

	var projects []bson.M
	if err = cursor.All(c, &projects); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "处理项目数据失败",
		})
		return
	}

	// 转换ID格式
	for _, project := range projects {
		if id, ok := project["_id"].(primitive.ObjectID); ok {
			project["id"] = id.Hex()
			delete(project, "_id")
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"projects": projects,
			"total":    len(projects),
		},
	})
}

// 辅助函数：转换int64指针
func int64Ptr(v int64) *int64 {
	return &v
}

// GetProjectContent 获取项目详情
func (h *ProjectHandler) GetProjectContent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "项目ID不能为空",
		})
		return
	}
	db := h.DB
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID格式",
		})
		return
	}
	var project bson.M
	err = db.Collection("project").FindOne(c, bson.M{"_id": objID}).Decode(&project)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "项目不存在",
		})
		return
	}
	var targetData bson.M
	err = db.Collection("ProjectTargetData").FindOne(c, bson.M{"id": id}).Decode(&targetData)
	if err != nil {
		targetData = bson.M{"target": ""} // 允许目标数据缺失
	}
	result := gin.H{
		"name":           project["name"],
		"tag":            project["tag"],
		"target":         targetData["target"],
		"node":           project["node"],
		"logo":           project["logo"],
		"scheduledTasks": project["scheduledTasks"],
		"hour":           project["hour"],
		"allNode":        project["allNode"],
		"duplicates":     project["duplicates"],
		"template":       project["template"],
		"ignore":         project["ignore"],
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": result,
	})
}

// AddProject 添加项目
func (h *ProjectHandler) AddProject(c *gin.Context) {
	var req ProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
		})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "项目名称不能为空",
		})
		return
	}

	db := h.DB
	project := &models.Project{
		Name:       req.Name,
		Tag:        req.Tag,
		Logo:       req.Logo,
		Node:       req.Node,
		AllNode:    req.AllNode,
		Duplicates: req.Duplicates,
		Template:   req.Template,
		Ignore:     req.Ignore,
		Hour:       req.Hour,
	}

	projectID, err := models.CreateProject(db, project, req.Target)
	if err != nil {
		if err == models.ErrNameExists {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "项目名称已存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "创建项目失败: " + err.Error(),
			})
		}
		return
	}

	// TODO: 定时任务和立即运行逻辑
	go database.RefreshConfig("all", "project", projectID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "项目添加成功",
		"data": gin.H{
			"id": projectID,
		},
	})
}

// UpdateProject 更新项目
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	var req ProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
		})
		return
	}

	if req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "项目ID不能为空",
		})
		return
	}

	db := h.DB
	updateData := bson.M{
		"name":       req.Name,
		"tag":        req.Tag,
		"logo":       req.Logo,
		"node":       req.Node,
		"allNode":    req.AllNode,
		"duplicates": req.Duplicates,
		"template":   req.Template,
		"ignore":     req.Ignore,
		"hour":       req.Hour,
	}

	err := models.UpdateProject(db, req.ID, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新项目失败: " + err.Error(),
		})
		return
	}

	_, err = db.Collection("ProjectTargetData").UpdateOne(
		c,
		bson.M{"id": req.ID},
		bson.M{"$set": bson.M{"target": req.Target}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新项目目标数据失败",
		})
		return
	}

	go database.RefreshConfig("all", "project", req.ID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "项目更新成功",
	})
}

// DeleteProject 删除项目
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	var req struct {
		IDs  []string `json:"ids"`
		DelA bool     `json:"delA"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数: " + err.Error(),
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "项目ID不能为空",
		})
		return
	}

	// 验证每个 ID 格式
	for _, id := range req.IDs {
		if _, err := primitive.ObjectIDFromHex(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的项目ID格式: " + id,
			})
			return
		}
	}

	db := h.DB
	err := models.DeleteProject(db, req.IDs)
	if err != nil {
		// 根据错误类型返回不同的 HTTP 状态码
		if err.Error() == "部分项目不存在" || err.Error() == "未找到要删除的项目" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": err.Error(),
			})
		} else if err.Error() == "项目ID列表不能为空" || strings.Contains(err.Error(), "无效的项目ID") {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "删除项目失败: " + err.Error(),
			})
		}
		return
	}

	// TODO: 删除关联资产和定时任务

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "项目删除成功",
	})
}

// ProjectRequest 项目请求结构
type ProjectRequest struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Tag            string   `json:"tag"`
	Target         string   `json:"target"`
	Node           []string `json:"node"`
	Logo           string   `json:"logo"`
	AllNode        bool     `json:"allNode"`
	ScheduledTasks bool     `json:"scheduledTasks"`
	Hour           int      `json:"hour"`
	RunNow         bool     `json:"runNow"`
	Duplicates     bool     `json:"duplicates"`
	Template       string   `json:"template"`
	Ignore         string   `json:"ignore"`
}

// ProjectResponse 项目响应结构
type ProjectResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

// GetProjects 获取项目列表
func GetProjects(c *gin.Context) {
	var req struct {
		Search    string `json:"search"`
		PageIndex int    `json:"pageIndex"`
		PageSize  int    `json:"pageSize"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
		})
		return
	}

	// 设置默认值
	if req.PageIndex <= 0 {
		req.PageIndex = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 创建MongoDB配置
	mongoConfig := config.MongoDBConfig{
		URI:           config.MONGODB_IP,
		Database:      config.MONGODB_DATABASE,
		MaxPoolSize:   100,
		MinPoolSize:   10,
		MaxIdleTimeMS: 30000,
	}

	// 连接数据库
	client, err := database.ConnectMongoDB(mongoConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库连接失败",
		})
		return
	}

	// 获取数据库实例
	db := client.Database(config.MONGODB_DATABASE)

	// 获取项目列表
	resultList, tagNum, err := models.GetProjects(db, req.Search, req.PageIndex, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取项目列表失败",
		})
		return
	}

	// 更新项目资产计数
	go models.UpdateProjectAssetCount(db)

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"result": resultList,
			"tag":    tagNum,
		},
	})
}

// GetAllProjects 获取所有项目（用于下拉列表）
func GetAllProjects(c *gin.Context) {
	// 创建MongoDB配置
	mongoConfig := config.MongoDBConfig{
		URI:           config.MONGODB_IP,
		Database:      config.MONGODB_DATABASE,
		MaxPoolSize:   100,
		MinPoolSize:   10,
		MaxIdleTimeMS: 30000,
	}

	// 连接数据库
	client, err := database.ConnectMongoDB(mongoConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库连接失败",
		})
		return
	}

	// 获取数据库实例
	db := client.Database(config.MONGODB_DATABASE)

	// 聚合查询，按标签分组
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": "$tag",
				"children": bson.M{
					"$push": bson.M{
						"value": bson.M{"$toString": "$_id"},
						"label": "$name",
					},
				},
			},
		},
		{
			"$project": bson.M{
				"_id":      0,
				"label":    "$_id",
				"value":    "",
				"children": 1,
			},
		},
	}

	cursor, err := db.Collection("project").Aggregate(c, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取项目列表失败",
		})
		return
	}

	var result []bson.M
	if err = cursor.All(c, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "处理项目列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list": result,
		},
	})
}

// GetProjectContent 获取项目详情
func GetProjectContent(c *gin.Context) {
	var req struct {
		ID string `json:"id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
		})
		return
	}

	if req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "ID不能为空",
		})
		return
	}

	// 创建MongoDB配置
	mongoConfig := config.MongoDBConfig{
		URI:           config.MONGODB_IP,
		Database:      config.MONGODB_DATABASE,
		MaxPoolSize:   100,
		MinPoolSize:   10,
		MaxIdleTimeMS: 30000,
	}

	// 连接数据库
	client, err := database.ConnectMongoDB(mongoConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库连接失败",
		})
		return
	}

	// 获取数据库实例
	db := client.Database(config.MONGODB_DATABASE)

	// 获取项目
	id, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID格式",
		})
		return
	}

	var project bson.M
	err = db.Collection("project").FindOne(c, bson.M{"_id": id}).Decode(&project)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "项目不存在",
		})
		return
	}

	// 获取项目目标数据
	var targetData bson.M
	err = db.Collection("ProjectTargetData").FindOne(c, bson.M{"id": req.ID}).Decode(&targetData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取项目目标数据失败",
		})
		return
	}

	// 构建响应
	result := gin.H{
		"name":           project["name"],
		"tag":            project["tag"],
		"target":         targetData["target"],
		"node":           project["node"],
		"logo":           project["logo"],
		"scheduledTasks": project["scheduledTasks"],
		"hour":           project["hour"],
		"allNode":        project["allNode"],
		"duplicates":     project["duplicates"],
		"template":       project["template"],
		"ignore":         project["ignore"],
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": result,
	})
}

// AddProject 添加项目
func AddProject(c *gin.Context) {
	var req ProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
		})
		return
	}

	// 验证必填字段
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "项目名称不能为空",
		})
		return
	}

	// 创建MongoDB配置
	mongoConfig := config.MongoDBConfig{
		URI:           config.MONGODB_IP,
		Database:      config.MONGODB_DATABASE,
		MaxPoolSize:   100,
		MinPoolSize:   10,
		MaxIdleTimeMS: 30000,
	}

	// 连接数据库
	client, err := database.ConnectMongoDB(mongoConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库连接失败",
		})
		return
	}

	// 获取数据库实例
	db := client.Database(config.MONGODB_DATABASE)

	// 创建项目对象
	project := &models.Project{
		Name:       req.Name,
		Tag:        req.Tag,
		Logo:       req.Logo,
		Node:       req.Node,
		AllNode:    req.AllNode,
		Duplicates: req.Duplicates,
		Template:   req.Template,
		Ignore:     req.Ignore,
		Hour:       req.Hour,
	}

	// 创建项目
	projectID, err := models.CreateProject(db, project, req.Target)
	if err != nil {
		if err == models.ErrNameExists {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "项目名称已存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "创建项目失败: " + err.Error(),
			})
		}
		return
	}

	// 如果需要创建定时任务
	if req.ScheduledTasks {
		// TODO: 添加定时任务
	}

	// 如果需要立即运行
	if req.RunNow {
		// TODO: 立即运行任务
	}

	// 刷新配置
	go database.RefreshConfig("all", "project", projectID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "项目添加成功",
	})
}

// UpdateProject 更新项目
func UpdateProject(c *gin.Context) {
	var req ProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
		})
		return
	}

	// 验证必填字段
	if req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "项目ID不能为空",
		})
		return
	}

	// 创建MongoDB配置
	mongoConfig := config.MongoDBConfig{
		URI:           config.MONGODB_IP,
		Database:      config.MONGODB_DATABASE,
		MaxPoolSize:   100,
		MinPoolSize:   10,
		MaxIdleTimeMS: 30000,
	}

	// 连接数据库
	client, err := database.ConnectMongoDB(mongoConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库连接失败",
		})
		return
	}

	// 获取数据库实例
	db := client.Database(config.MONGODB_DATABASE)

	// 更新项目
	updateData := bson.M{
		"name":       req.Name,
		"tag":        req.Tag,
		"logo":       req.Logo,
		"node":       req.Node,
		"allNode":    req.AllNode,
		"duplicates": req.Duplicates,
		"template":   req.Template,
		"ignore":     req.Ignore,
		"hour":       req.Hour,
	}

	err = models.UpdateProject(db, req.ID, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新项目失败: " + err.Error(),
		})
		return
	}

	// 更新项目目标数据
	_, err = db.Collection("ProjectTargetData").UpdateOne(
		c,
		bson.M{"id": req.ID},
		bson.M{"$set": bson.M{"target": req.Target}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新项目目标数据失败",
		})
		return
	}

	// TODO: 处理定时任务的更新

	// 刷新配置
	go database.RefreshConfig("all", "project", req.ID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "项目更新成功",
	})
}

// DeleteProject 删除项目
func DeleteProject(c *gin.Context) {
	var req struct {
		IDs  []string `json:"ids"`
		DelA bool     `json:"delA"` // 是否删除关联资产
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "项目ID不能为空",
		})
		return
	}

	// 创建MongoDB配置
	mongoConfig := config.MongoDBConfig{
		URI:           config.MONGODB_IP,
		Database:      config.MONGODB_DATABASE,
		MaxPoolSize:   100,
		MinPoolSize:   10,
		MaxIdleTimeMS: 30000,
	}

	// 连接数据库
	client, err := database.ConnectMongoDB(mongoConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库连接失败",
		})
		return
	}

	// 获取数据库实例
	db := client.Database(config.MONGODB_DATABASE)

	// 删除项目
	err = models.DeleteProject(db, req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除项目失败: " + err.Error(),
		})
		return
	}

	// 如果需要删除关联资产
	if req.DelA {
		// TODO: 删除关联资产
	}

	// TODO: 删除关联的定时任务

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "项目删除成功",
	})
}
