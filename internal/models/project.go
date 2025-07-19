package models

import (
	"context"
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
	var objectIDs []primitive.ObjectID
	for _, id := range projectIDs {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		objectIDs = append(objectIDs, oid)
	}

	// 删除项目
	_, err := db.Collection("project").DeleteMany(
		context.Background(),
		bson.M{"_id": bson.M{"$in": objectIDs}},
	)
	if err != nil {
		return err
	}

	// 删除项目目标数据
	_, err = db.Collection("ProjectTargetData").DeleteMany(
		context.Background(),
		bson.M{"id": bson.M{"$in": projectIDs}},
	)
	return err
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

	// 获取标签统计信息
	pipeline := []bson.M{
		{"$group": bson.M{"_id": "$tag", "count": bson.M{"$sum": 1}}},
		{"$sort": bson.M{"count": -1}},
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
	for _, tag := range tagResults {
		tagNum[tag.ID] = tag.Count
		allNum += tag.Count
	}
	tagNum["All"] = allNum

	// 获取项目列表
	resultList := make(map[string][]map[string]interface{})

	for tag := range tagNum {
		var tagQuery bson.M
		if tag != "All" {
			if query != nil {
				tagQuery = bson.M{"$and": []bson.M{query, {"tag": tag}}}
			} else {
				tagQuery = bson.M{"tag": tag}
			}
		} else {
			tagQuery = query
		}

		opts := options.Find().
			SetSort(bson.M{"AssetCount": -1}).
			SetSkip(int64((pageIndex - 1) * pageSize)).
			SetLimit(int64(pageSize))

		cursor, err := db.Collection("project").Find(
			context.Background(),
			tagQuery,
			opts,
		)
		if err != nil {
			return nil, nil, err
		}

		var results []map[string]interface{}
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

			results = append(results, project)
		}

		if err := cursor.Err(); err != nil {
			return nil, nil, err
		}

		resultList[tag] = results
	}

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
