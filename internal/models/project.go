package models

import (
	"context"
	"fmt"
	"time"

	pkgerrors "github.com/StellarServer/internal/pkg/errors"
	"github.com/StellarServer/internal/pkg/logger"
	"github.com/StellarServer/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Project 项目模型
type Project struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Tag         string             `bson:"tag" json:"tag"`
	Logo        string             `bson:"logo" json:"logo"`
	AssetCount  int                `bson:"AssetCount" json:"AssetCount"`
	Node        []string           `bson:"node" json:"node"`
	RootDomains []string           `bson:"root_domains" json:"root_domains"`
	AllNode     bool               `bson:"allNode" json:"allNode"`
	Duplicates  bool               `bson:"duplicates" json:"duplicates"`
	Template    string             `bson:"template" json:"template"`
	Ignore      string             `bson:"ignore" json:"ignore"`
	Type        string             `bson:"tp" json:"tp"`
	Hour        int                `bson:"hour" json:"hour"`
	Created     time.Time          `bson:"created" json:"created"`
	Updated     time.Time          `bson:"updated" json:"updated"`
}

// ProjectTargetData 项目目标数据
type ProjectTargetData struct {
	ID     string `bson:"id" json:"id"`
	Target string `bson:"target" json:"target"`
}

// CreateProject 创建项目
func CreateProject(db *mongo.Database, project *Project, target string) (string, error) {
	// 检查项目名是否已存在
	count, err := db.Collection("project").CountDocuments(context.Background(), bson.M{"name": project.Name})
	if err != nil {
		logger.Error("CreateProject count documents failed", map[string]interface{}{"projectName": project.Name, "error": err})
		return "", pkgerrors.WrapDatabaseError(err, "检查项目名是否存在")
	}
	if count > 0 {
		logger.Error("CreateProject project name exists", map[string]interface{}{"projectName": project.Name})
		return "", pkgerrors.NewAppError(pkgerrors.CodeConflict, "项目名已存在", 409)
	}

	// 处理目标数据
	targetList, err := utils.GetTargetList(target, project.Ignore)
	if err != nil {
		logger.Error("CreateProject get target list failed", map[string]interface{}{"target": target, "error": err})
		return "", pkgerrors.WrapError(err, pkgerrors.CodeValidationFailed, "处理目标数据失败", 400)
	}

	// 提取根域名
	rootDomains := []string{}
	for _, tg := range targetList {
		var rootDomain string
		if utils.HasPrefix(tg, "CMP:", "ICP:", "APP:", "APP-ID:") {
			rootDomain = utils.RemovePrefix(tg, "CMP:", "ICP:", "APP:", "APP-ID:")
			if utils.HasPrefix(tg, "ICP:") {
				rootDomain = utils.GetBeforeLastDash(rootDomain)
			}
		} else {
			rootDomain = utils.GetRootDomain(tg)
		}

		if rootDomain != "" && !utils.StringInSlice(rootDomain, rootDomains) {
			rootDomains = append(rootDomains, rootDomain)
		}
	}

	// 设置项目属性
	project.RootDomains = rootDomains
	project.Type = "project"
	project.Created = time.Now()
	project.Updated = time.Now()

	// 插入项目
	result, err := db.Collection("project").InsertOne(context.Background(), project)
	if err != nil {
		logger.Error("CreateProject insert project failed", map[string]interface{}{"projectName": project.Name, "error": err})
		return "", pkgerrors.WrapDatabaseError(err, "创建项目")
	}

	projectID := result.InsertedID.(primitive.ObjectID).Hex()

	// 插入项目目标数据
	_, err = db.Collection("ProjectTargetData").InsertOne(context.Background(), bson.M{
		"id":     projectID,
		"target": target,
	})
	if err != nil {
		logger.Error("CreateProject insert target data failed", map[string]interface{}{"projectID": projectID, "error": err})
		return "", pkgerrors.WrapDatabaseError(err, "创建项目目标数据")
	}

	return projectID, nil
}

// GetProject 获取项目
func GetProject(db *mongo.Database, projectID string) (*Project, error) {
	id, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		logger.Error("GetProject invalid projectID", map[string]interface{}{"projectID": projectID, "error": err})
		return nil, pkgerrors.NewAppErrorWithCause(pkgerrors.CodeBadRequest, "无效的项目ID", 400, err)
	}

	var project Project
	err = db.Collection("project").FindOne(context.Background(), bson.M{"_id": id}).Decode(&project)
	if err != nil {
		logger.Error("GetProject find project failed", map[string]interface{}{"projectID": projectID, "error": err})
		if err == mongo.ErrNoDocuments {
			return nil, pkgerrors.NewNotFoundError("项目不存在")
		}
		return nil, pkgerrors.WrapDatabaseError(err, "查询项目")
	}

	return &project, nil
}

// GetProjectTargetData 获取项目目标数据
func GetProjectTargetData(db *mongo.Database, projectID string) (*ProjectTargetData, error) {
	var targetData ProjectTargetData
	err := db.Collection("ProjectTargetData").FindOne(context.Background(), bson.M{"id": projectID}).Decode(&targetData)
	if err != nil {
		logger.Error("GetProjectTargetData find target data failed", map[string]interface{}{"projectID": projectID, "error": err})
		if err == mongo.ErrNoDocuments {
			return nil, pkgerrors.NewNotFoundError("项目目标数据不存在")
		}
		return nil, pkgerrors.WrapDatabaseError(err, "查询项目目标数据")
	}

	return &targetData, nil
}

// UpdateProject 更新项目
func UpdateProject(db *mongo.Database, projectID string, updateData bson.M) error {
	id, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return err
	}

	updateData["updated"] = time.Now()
	_, err = db.Collection("project").UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		bson.M{"$set": updateData},
	)
	return err
}

// DeleteProject 删除项目
func DeleteProject(db *mongo.Database, projectIDs []string) error {
	if len(projectIDs) == 0 {
		return pkgerrors.NewAppError(pkgerrors.CodeBadRequest, "项目ID列表不能为空", 400)
	}

	var objectIDs []primitive.ObjectID
	for _, id := range projectIDs {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			logger.Error("DeleteProject invalid project ID", map[string]interface{}{"projectID": id, "error": err})
			return pkgerrors.NewAppErrorWithCause(pkgerrors.CodeBadRequest, "无效的项目ID", 400, err)
		}
		objectIDs = append(objectIDs, oid)
	}

	ctx := context.Background()

	// 验证项目是否存在并获取项目信息
	count, err := db.Collection("project").CountDocuments(ctx, bson.M{"_id": bson.M{"$in": objectIDs}})
	if err != nil {
		logger.Error("DeleteProject verify projects exist failed", map[string]interface{}{"projectIDs": projectIDs, "error": err})
		return pkgerrors.WrapDatabaseError(err, "验证项目存在性")
	}
	if count != int64(len(projectIDs)) {
		logger.Warn("DeleteProject some projects not found", map[string]interface{}{"requestedCount": len(projectIDs), "foundCount": count})
		return pkgerrors.NewAppError(pkgerrors.CodeNotFound, "部分项目不存在", 404)
	}

	// 1. 删除项目主表
	result, err := db.Collection("project").DeleteMany(ctx, bson.M{"_id": bson.M{"$in": objectIDs}})
	if err != nil {
		logger.Error("DeleteProject delete projects failed", map[string]interface{}{"projectIDs": projectIDs, "error": err})
		return pkgerrors.WrapDatabaseError(err, "删除项目")
	}
	
	if result.DeletedCount == 0 {
		logger.Warn("DeleteProject no projects deleted", map[string]interface{}{"projectIDs": projectIDs})
		return pkgerrors.NewAppError(pkgerrors.CodeNotFound, "未找到要删除的项目", 404)
	}
	
	logger.Info("DeleteProject projects deleted", map[string]interface{}{"projectIDs": projectIDs, "deletedCount": result.DeletedCount})

	// 2. 删除项目目标数据
	targetResult, err := db.Collection("ProjectTargetData").DeleteMany(ctx, bson.M{"id": bson.M{"$in": projectIDs}})
	if err != nil {
		logger.Error("DeleteProject delete target data failed", map[string]interface{}{"projectIDs": projectIDs, "error": err})
		// 继续执行，但记录错误
	} else {
		logger.Info("DeleteProject target data deleted", map[string]interface{}{"projectIDs": projectIDs, "deletedCount": targetResult.DeletedCount})
	}

	// 3. 删除关联的资产数据
	assetCollections := []string{
		"domain", "subdomain", "ip", "port", "url", "http", "app", "miniapp", "assets",
	}
	
	for _, collection := range assetCollections {
		assetResult, err := db.Collection(collection).DeleteMany(ctx, bson.M{"projectId": bson.M{"$in": objectIDs}})
		if err != nil {
			logger.Error("DeleteProject delete assets failed", map[string]interface{}{
				"collection": collection,
				"projectIDs": projectIDs,
				"error": err,
			})
			// 继续执行，不直接返回错误
		} else if assetResult.DeletedCount > 0 {
			logger.Info("DeleteProject assets deleted", map[string]interface{}{
				"collection": collection,
				"projectIDs": projectIDs,
				"deletedCount": assetResult.DeletedCount,
			})
		}
	}

	// 4. 删除关联的漏洞数据
	vulnResult, err := db.Collection("vulnerabilities").DeleteMany(ctx, bson.M{"projectId": bson.M{"$in": objectIDs}})
	if err != nil {
		logger.Error("DeleteProject delete vulnerabilities failed", map[string]interface{}{"projectIDs": projectIDs, "error": err})
		// 继续执行，不直接返回错误
	} else if vulnResult.DeletedCount > 0 {
		logger.Info("DeleteProject vulnerabilities deleted", map[string]interface{}{"projectIDs": projectIDs, "deletedCount": vulnResult.DeletedCount})
	}

	// 5. 删除任务调度规则
	ruleResult, err := db.Collection("TaskScheduleRule").DeleteMany(ctx, bson.M{"projectId": bson.M{"$in": projectIDs}})
	if err != nil {
		logger.Error("DeleteProject delete task rules failed", map[string]interface{}{"projectIDs": projectIDs, "error": err})
		// 继续执行，不直接返回错误
	} else if ruleResult.DeletedCount > 0 {
		logger.Info("DeleteProject task rules deleted", map[string]interface{}{"projectIDs": projectIDs, "deletedCount": ruleResult.DeletedCount})
	}

	// 6. 删除任务历史记录
	historyResult, err := db.Collection("TaskHistory").DeleteMany(ctx, bson.M{"projectId": bson.M{"$in": projectIDs}})
	if err != nil {
		logger.Error("DeleteProject delete task history failed", map[string]interface{}{"projectIDs": projectIDs, "error": err})
		// 继续执行，不直接返回错误
	} else if historyResult.DeletedCount > 0 {
		logger.Info("DeleteProject task history deleted", map[string]interface{}{"projectIDs": projectIDs, "deletedCount": historyResult.DeletedCount})
	}

	logger.Info("DeleteProject cascade delete completed", map[string]interface{}{"projectIDs": projectIDs})
	return nil
}

// GetProjects 获取项目列表
func GetProjects(db *mongo.Database, search string, pageIndex, pageSize int) (map[string][]map[string]interface{}, map[string]int, error) {
	// 构建查询条件
	var query bson.M
	if search != "" {
		query = bson.M{
			"$or": []bson.M{
				{"name": bson.M{"$regex": search, "$options": "i"}},
				{"root_domain": bson.M{"$regex": search, "$options": "i"}},
			},
		}
	} else {
		query = bson.M{}
	}

	// 先验证数据库中项目的实际总数
	totalProjectsInDB, err := db.Collection("project").CountDocuments(context.Background(), query)
	if err != nil {
		fmt.Printf("GetProjects - 查询项目总数失败: %v\n", err)
	} else {
		fmt.Printf("GetProjects - 数据库中项目总数: %d (搜索条件: '%s')\n", totalProjectsInDB, search)
	}

	// 获取标签统计信息 - 修复：应用搜索条件到统计查询
	var pipeline []bson.M
	if search != "" {
		// 如果有搜索条件，先应用搜索条件
		pipeline = []bson.M{
			{"$match": query},
			{"$group": bson.M{"_id": "$tag", "count": bson.M{"$sum": 1}}},
			{"$sort": bson.M{"count": -1}},
		}
	} else {
		// 没有搜索条件时，直接分组统计
		pipeline = []bson.M{
			{"$group": bson.M{"_id": "$tag", "count": bson.M{"$sum": 1}}},
			{"$sort": bson.M{"count": -1}},
		}
	}

	cursor, err := db.Collection("project").Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, nil, err
	}

	var tagResults []struct {
		ID    string `bson:"_id"`
		Count int    `bson:"count"`
	}
	if err = cursor.All(context.Background(), &tagResults); err != nil {
		return nil, nil, err
	}

	// 构建标签统计数据
	tagNum := make(map[string]int)
	allNum := 0
	fmt.Printf("GetProjects - 聚合查询返回的标签结果: %+v\n", tagResults)
	for _, tag := range tagResults {
		tagNum[tag.ID] = tag.Count
		allNum += tag.Count
	}
	tagNum["All"] = allNum
	
	// 调试日志
	fmt.Printf("GetProjects - 搜索条件: %s, 分页: %d/%d, 统计结果: %+v, 数据库总数应该是: %d\n", search, pageIndex, pageSize, tagNum, totalProjectsInDB)

	// 获取项目列表 - 修复：只返回一次查询结果，避免重复数据
	resultList := make(map[string][]map[string]interface{})
	
	// 初始化所有标签为空数组
	for tag := range tagNum {
		resultList[tag] = []map[string]interface{}{}
	}
	
	// 只查询一次，获取分页数据
	opts := options.Find().
		SetSort(bson.M{"AssetCount": -1}).
		SetSkip(int64((pageIndex - 1) * pageSize)).
		SetLimit(int64(pageSize))

	cursor, err = db.Collection("project").Find(
		context.Background(),
		query, // 使用原始搜索条件，不按标签过滤
		opts,
	)
	if err != nil {
		return nil, nil, err
	}

	var allResults []map[string]interface{}
	for cursor.Next(context.Background()) {
		var project bson.M
		if err := cursor.Decode(&project); err != nil {
			return nil, nil, err
		}

		// 转换ID为字符串
		if id, ok := project["_id"].(primitive.ObjectID); ok {
			project["id"] = id.Hex()
			delete(project, "_id")
		}

		// 确保AssetCount字段存在
		if _, ok := project["AssetCount"]; !ok {
			project["AssetCount"] = 0
		}

		allResults = append(allResults, project)
	}

	if err := cursor.Err(); err != nil {
		return nil, nil, err
	}

	// 将所有结果分配给"All"标签，其他标签保持空数组
	resultList["All"] = allResults
	
	// 调试日志 - 验证返回的数据
	fmt.Printf("GetProjects - 返回项目数量: %d, 总数统计: %d\n", len(allResults), tagNum["All"])

	return resultList, tagNum, nil
}

// UpdateProjectAssetCount 更新项目资产计数
func UpdateProjectAssetCount(db *mongo.Database) error {
	// 获取所有项目ID
	cursor, err := db.Collection("project").Find(
		context.Background(),
		bson.M{},
		options.Find().SetProjection(bson.M{"_id": 1}),
	)
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	var projects []struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	if err = cursor.All(context.Background(), &projects); err != nil {
		return err
	}

	// 更新每个项目的资产计数
	for _, project := range projects {
		projectID := project.ID.Hex()
		count, err := db.Collection("asset").CountDocuments(
			context.Background(),
			bson.M{"project": projectID},
		)
		if err != nil {
			return err
		}

		if count > 0 {
			_, err = db.Collection("project").UpdateOne(
				context.Background(),
				bson.M{"_id": project.ID},
				bson.M{"$set": bson.M{"AssetCount": count}},
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
