package models

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AssetRepository struct {
	db *mongo.Database
}

func NewAssetRepository(db *mongo.Database) *AssetRepository {
	return &AssetRepository{db: db}
}

// AssetFilter 定义资产筛选条件
type AssetFilter struct {
	ProjectID  *primitive.ObjectID `json:"projectId,omitempty"`
	Tags       []string            `json:"tags,omitempty"`
	Type       *AssetType          `json:"type,omitempty"`
	RootDomain *string             `json:"rootDomain,omitempty"`
	TaskName   *string             `json:"taskName,omitempty"`
	Search     *string             `json:"search,omitempty"` // 模糊搜索
	DateRange  *DateRange          `json:"dateRange,omitempty"`
}

// DateRange 定义日期范围
type DateRange struct {
	StartTime *time.Time `json:"startTime,omitempty"`
	EndTime   *time.Time `json:"endTime,omitempty"`
}

// AssetListOptions 定义列表查询选项
type AssetListOptions struct {
	Filter   *AssetFilter `json:"filter,omitempty"`
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
	SortBy   string       `json:"sortBy"`   // 排序字段
	SortDesc bool         `json:"sortDesc"` // 是否降序
}

// AssetListResult 定义列表查询结果
type AssetListResult struct {
	Items      []interface{} `json:"items"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"pageSize"`
	TotalPages int           `json:"totalPages"`
}

// CreateAsset 创建资产（支持多类型）
func (r *AssetRepository) CreateAsset(ctx context.Context, assetType AssetType, asset interface{}) (primitive.ObjectID, error) {
	collName := AssetCollection(assetType)
	result, err := r.db.Collection(collName).InsertOne(ctx, asset)
	if err != nil {
		return primitive.NilObjectID, err
	}
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, errors.New("插入资产返回ID类型错误")
	}
	return id, nil
}

// FindAssetByID 根据ID查询资产（需指定类型）
func (r *AssetRepository) FindAssetByID(ctx context.Context, assetType AssetType, id primitive.ObjectID, result interface{}) error {
	collName := AssetCollection(assetType)
	return r.db.Collection(collName).FindOne(ctx, bson.M{"_id": id}).Decode(result)
}

// UpdateAsset 更新资产
func (r *AssetRepository) UpdateAsset(ctx context.Context, assetType AssetType, id primitive.ObjectID, update bson.M) error {
	collName := AssetCollection(assetType)
	// 添加更新时间
	if update["$set"] == nil {
		update["$set"] = bson.M{}
	}
	update["$set"].(bson.M)["updatedAt"] = time.Now()
	
	_, err := r.db.Collection(collName).UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// DeleteAsset 删除资产
func (r *AssetRepository) DeleteAsset(ctx context.Context, assetType AssetType, id primitive.ObjectID) error {
	collName := AssetCollection(assetType)
	_, err := r.db.Collection(collName).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// ListAssets 分页查询资产列表
func (r *AssetRepository) ListAssets(ctx context.Context, assetType AssetType, opts AssetListOptions) (*AssetListResult, error) {
	collName := AssetCollection(assetType)
	collection := r.db.Collection(collName)
	
	// 构建查询条件
	filter := r.buildFilter(opts.Filter)
	
	// 计算总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	// 设置默认分页参数
	if opts.PageSize <= 0 {
		opts.PageSize = 20
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}
	
	// 计算跳过的文档数
	skip := (opts.Page - 1) * opts.PageSize
	
	// 构建查询选项
	findOpts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(opts.PageSize))
	
	// 设置排序
	if opts.SortBy != "" {
		sortOrder := 1
		if opts.SortDesc {
			sortOrder = -1
		}
		findOpts.SetSort(bson.D{{Key: opts.SortBy, Value: sortOrder}})
	} else {
		// 默认按创建时间降序
		findOpts.SetSort(bson.D{{Key: "createdAt", Value: -1}})
	}
	
	// 查询数据
	cursor, err := collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	// 解析结果
	var items []interface{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	
	// 计算总页数
	totalPages := int((total + int64(opts.PageSize) - 1) / int64(opts.PageSize))
	
	return &AssetListResult{
		Items:      items,
		Total:      total,
		Page:       opts.Page,
		PageSize:   opts.PageSize,
		TotalPages: totalPages,
	}, nil
}

// BatchCreateAssets 批量创建资产
func (r *AssetRepository) BatchCreateAssets(ctx context.Context, assetType AssetType, assets []interface{}) ([]primitive.ObjectID, error) {
	if len(assets) == 0 {
		return nil, errors.New("批量创建资产列表不能为空")
	}
	
	collName := AssetCollection(assetType)
	result, err := r.db.Collection(collName).InsertMany(ctx, assets)
	if err != nil {
		return nil, err
	}
	
	ids := make([]primitive.ObjectID, len(result.InsertedIDs))
	for i, id := range result.InsertedIDs {
		if objID, ok := id.(primitive.ObjectID); ok {
			ids[i] = objID
		}
	}
	
	return ids, nil
}

// BatchUpdateAssets 批量更新资产
func (r *AssetRepository) BatchUpdateAssets(ctx context.Context, assetType AssetType, filter bson.M, update bson.M) (int64, error) {
	collName := AssetCollection(assetType)
	
	// 添加更新时间
	if update["$set"] == nil {
		update["$set"] = bson.M{}
	}
	update["$set"].(bson.M)["updatedAt"] = time.Now()
	
	result, err := r.db.Collection(collName).UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}
	
	return result.ModifiedCount, nil
}

// BatchDeleteAssets 批量删除资产
func (r *AssetRepository) BatchDeleteAssets(ctx context.Context, assetType AssetType, ids []primitive.ObjectID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	
	collName := AssetCollection(assetType)
	filter := bson.M{"_id": bson.M{"$in": ids}}
	
	result, err := r.db.Collection(collName).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	
	return result.DeletedCount, nil
}

// FindAssetsByProject 根据项目ID查询资产
func (r *AssetRepository) FindAssetsByProject(ctx context.Context, assetType AssetType, projectID primitive.ObjectID) ([]interface{}, error) {
	collName := AssetCollection(assetType)
	
	cursor, err := r.db.Collection(collName).Find(ctx, bson.M{"projectId": projectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var assets []interface{}
	if err = cursor.All(ctx, &assets); err != nil {
		return nil, err
	}
	
	return assets, nil
}

// FindAssetsByTags 根据标签查询资产
func (r *AssetRepository) FindAssetsByTags(ctx context.Context, assetType AssetType, tags []string) ([]interface{}, error) {
	collName := AssetCollection(assetType)
	
	filter := bson.M{"tags": bson.M{"$in": tags}}
	cursor, err := r.db.Collection(collName).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var assets []interface{}
	if err = cursor.All(ctx, &assets); err != nil {
		return nil, err
	}
	
	return assets, nil
}

// AddAssetChange 添加资产变更记录
func (r *AssetRepository) AddAssetChange(ctx context.Context, assetType AssetType, assetID primitive.ObjectID, change AssetChange) error {
	collName := AssetCollection(assetType)
	
	change.Time = time.Now()
	update := bson.M{
		"$push": bson.M{"changeHistory": change},
		"$set":  bson.M{"updatedAt": time.Now()},
	}
	
	_, err := r.db.Collection(collName).UpdateOne(ctx, bson.M{"_id": assetID}, update)
	return err
}

// CreateAssetRelation 创建资产关系
func (r *AssetRepository) CreateAssetRelation(ctx context.Context, relation AssetRelation) (primitive.ObjectID, error) {
	relation.CreatedAt = time.Now()
	relation.UpdatedAt = time.Now()
	
	result, err := r.db.Collection("asset_relations").InsertOne(ctx, relation)
	if err != nil {
		return primitive.NilObjectID, err
	}
	
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, errors.New("插入关系返回ID类型错误")
	}
	return id, nil
}

// FindAssetRelations 查询资产关系
func (r *AssetRepository) FindAssetRelations(ctx context.Context, assetID primitive.ObjectID) ([]AssetRelation, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"sourceAssetId": assetID},
			{"targetAssetId": assetID},
		},
	}
	
	cursor, err := r.db.Collection("asset_relations").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var relations []AssetRelation
	if err = cursor.All(ctx, &relations); err != nil {
		return nil, err
	}
	
	return relations, nil
}

// GetAssetStats 获取资产统计信息
func (r *AssetRepository) GetAssetStats(ctx context.Context, projectID *primitive.ObjectID) (map[AssetType]int64, error) {
	stats := make(map[AssetType]int64)
	
	assetTypes := []AssetType{
		AssetTypeDomain, AssetTypeSubdomain, AssetTypeIP, AssetTypePort,
		AssetTypeURL, AssetTypeHTTP, AssetTypeApp, AssetTypeMiniApp,
	}
	
	for _, assetType := range assetTypes {
		collName := AssetCollection(assetType)
		
		var filter bson.M
		if projectID != nil {
			filter = bson.M{"projectId": *projectID}
		} else {
			filter = bson.M{}
		}
		
		count, err := r.db.Collection(collName).CountDocuments(ctx, filter)
		if err != nil {
			return nil, err
		}
		
		stats[assetType] = count
	}
	
	return stats, nil
}

// buildFilter 构建查询过滤条件
func (r *AssetRepository) buildFilter(filter *AssetFilter) bson.M {
	if filter == nil {
		return bson.M{}
	}
	
	query := bson.M{}
	
	if filter.ProjectID != nil {
		query["projectId"] = *filter.ProjectID
	}
	
	if len(filter.Tags) > 0 {
		query["tags"] = bson.M{"$in": filter.Tags}
	}
	
	if filter.Type != nil {
		query["type"] = *filter.Type
	}
	
	if filter.RootDomain != nil {
		query["rootDomain"] = *filter.RootDomain
	}
	
	if filter.TaskName != nil {
		query["taskName"] = *filter.TaskName
	}
	
	if filter.Search != nil && *filter.Search != "" {
		// 构建模糊搜索条件，根据不同资产类型搜索不同字段
		searchRegex := bson.M{"$regex": *filter.Search, "$options": "i"}
		query["$or"] = []bson.M{
			{"domain": searchRegex},
			{"host": searchRegex},
			{"ip": searchRegex},
			{"url": searchRegex},
			{"title": searchRegex},
			{"appName": searchRegex},
			{"packageName": searchRegex},
		}
	}
	
	if filter.DateRange != nil {
		dateFilter := bson.M{}
		if filter.DateRange.StartTime != nil {
			dateFilter["$gte"] = *filter.DateRange.StartTime
		}
		if filter.DateRange.EndTime != nil {
			dateFilter["$lte"] = *filter.DateRange.EndTime
		}
		if len(dateFilter) > 0 {
			query["createdAt"] = dateFilter
		}
	}
	
	return query
}
