package models

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AssetRepository struct {
	db *mongo.Database
}

func NewAssetRepository(db *mongo.Database) *AssetRepository {
	return &AssetRepository{db: db}
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
