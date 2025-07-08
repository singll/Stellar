package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoMetadataStore MongoDB插件元数据存储
type MongoMetadataStore struct {
	db         *mongo.Database
	collection string
}

// NewMongoMetadataStore 创建MongoDB插件元数据存储
func NewMongoMetadataStore(db *mongo.Database, collection string) *MongoMetadataStore {
	return &MongoMetadataStore{
		db:         db,
		collection: collection,
	}
}

// GetPluginMetadata 获取插件元数据
func (s *MongoMetadataStore) GetPluginMetadata(id string) (*models.PluginMetadata, error) {
	var metadata models.PluginMetadata
	err := s.db.Collection(s.collection).FindOne(context.Background(), bson.M{"_id": id}).Decode(&metadata)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("插件元数据不存在: %s", id)
		}
		return nil, fmt.Errorf("获取插件元数据失败: %v", err)
	}
	return &metadata, nil
}

// SavePluginMetadata 保存插件元数据
func (s *MongoMetadataStore) SavePluginMetadata(metadata *models.PluginMetadata) error {
	// 设置更新时间
	metadata.UpdateTime = time.Now()

	// 更新或插入元数据
	_, err := s.db.Collection(s.collection).UpdateOne(
		context.Background(),
		bson.M{"_id": metadata.ID},
		bson.M{"$set": metadata},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("保存插件元数据失败: %v", err)
	}
	return nil
}

// ListPluginMetadata 列出插件元数据
func (s *MongoMetadataStore) ListPluginMetadata() ([]*models.PluginMetadata, error) {
	// 查询所有元数据
	cursor, err := s.db.Collection(s.collection).Find(context.Background(), bson.M{})
	if err != nil {
		return nil, fmt.Errorf("查询插件元数据失败: %v", err)
	}
	defer cursor.Close(context.Background())

	// 解析结果
	var metadataList []*models.PluginMetadata
	if err := cursor.All(context.Background(), &metadataList); err != nil {
		return nil, fmt.Errorf("解析插件元数据失败: %v", err)
	}
	return metadataList, nil
}

// DeletePluginMetadata 删除插件元数据
func (s *MongoMetadataStore) DeletePluginMetadata(id string) error {
	_, err := s.db.Collection(s.collection).DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("删除插件元数据失败: %v", err)
	}
	return nil
}

// ListPlugins 列出所有插件元数据
func (s *MongoMetadataStore) ListPlugins() ([]*models.PluginMetadata, error) {
	return s.ListPluginMetadata()
}

// ListPluginsByType 按类型列出插件元数据
func (s *MongoMetadataStore) ListPluginsByType(pluginType string) ([]*models.PluginMetadata, error) {
	// 查询指定类型的元数据
	cursor, err := s.db.Collection(s.collection).Find(
		context.Background(),
		bson.M{"type": pluginType},
	)
	if err != nil {
		return nil, fmt.Errorf("查询插件元数据失败: %v", err)
	}
	defer cursor.Close(context.Background())

	// 解析结果
	var metadataList []*models.PluginMetadata
	if err := cursor.All(context.Background(), &metadataList); err != nil {
		return nil, fmt.Errorf("解析插件元数据失败: %v", err)
	}
	return metadataList, nil
}

// ListPluginsByCategory 按分类列出插件元数据
func (s *MongoMetadataStore) ListPluginsByCategory(category string) ([]*models.PluginMetadata, error) {
	// 查询指定分类的元数据
	cursor, err := s.db.Collection(s.collection).Find(
		context.Background(),
		bson.M{"category": category},
	)
	if err != nil {
		return nil, fmt.Errorf("查询插件元数据失败: %v", err)
	}
	defer cursor.Close(context.Background())

	// 解析结果
	var metadataList []*models.PluginMetadata
	if err := cursor.All(context.Background(), &metadataList); err != nil {
		return nil, fmt.Errorf("解析插件元数据失败: %v", err)
	}
	return metadataList, nil
}

// ListPluginsByTag 按标签列出插件元数据
func (s *MongoMetadataStore) ListPluginsByTag(tag string) ([]*models.PluginMetadata, error) {
	// 查询包含指定标签的元数据
	cursor, err := s.db.Collection(s.collection).Find(
		context.Background(),
		bson.M{"tags": tag},
	)
	if err != nil {
		return nil, fmt.Errorf("查询插件元数据失败: %v", err)
	}
	defer cursor.Close(context.Background())

	// 解析结果
	var metadataList []*models.PluginMetadata
	if err := cursor.All(context.Background(), &metadataList); err != nil {
		return nil, fmt.Errorf("解析插件元数据失败: %v", err)
	}
	return metadataList, nil
}

// FileMetadataStore 文件插件元数据存储
type FileMetadataStore struct {
	filePath string
	metadata map[string]*models.PluginMetadata
}

// NewFileMetadataStore 创建文件插件元数据存储
func NewFileMetadataStore(filePath string) (*FileMetadataStore, error) {
	store := &FileMetadataStore{
		filePath: filePath,
		metadata: make(map[string]*models.PluginMetadata),
	}

	// 加载元数据
	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

// load 加载元数据
func (s *FileMetadataStore) load() error {
	// 检查文件是否存在
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		// 文件不存在，创建空文件
		return s.save()
	}

	// 读取文件
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return fmt.Errorf("读取元数据文件失败: %v", err)
	}

	// 解析JSON
	var metadataList []*models.PluginMetadata
	if err := json.Unmarshal(data, &metadataList); err != nil {
		return fmt.Errorf("解析元数据文件失败: %v", err)
	}

	// 构建映射
	for _, metadata := range metadataList {
		s.metadata[metadata.ID] = metadata
	}

	return nil
}

// save 保存元数据
func (s *FileMetadataStore) save() error {
	// 构建元数据列表
	metadataList := make([]*models.PluginMetadata, 0, len(s.metadata))
	for _, metadata := range s.metadata {
		metadataList = append(metadataList, metadata)
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(metadataList, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化元数据失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("写入元数据文件失败: %v", err)
	}

	return nil
}

// GetPluginMetadata 获取插件元数据
func (s *FileMetadataStore) GetPluginMetadata(id string) (*models.PluginMetadata, error) {
	metadata, exists := s.metadata[id]
	if !exists {
		return nil, fmt.Errorf("插件元数据不存在: %s", id)
	}
	return metadata, nil
}

// SavePluginMetadata 保存插件元数据
func (s *FileMetadataStore) SavePluginMetadata(metadata *models.PluginMetadata) error {
	// 设置更新时间
	metadata.UpdateTime = time.Now()

	// 保存元数据
	s.metadata[metadata.ID] = metadata

	// 保存到文件
	return s.save()
}

// ListPluginMetadata 列出插件元数据
func (s *FileMetadataStore) ListPluginMetadata() ([]*models.PluginMetadata, error) {
	metadataList := make([]*models.PluginMetadata, 0, len(s.metadata))
	for _, metadata := range s.metadata {
		metadataList = append(metadataList, metadata)
	}
	return metadataList, nil
}

// DeletePluginMetadata 删除插件元数据
func (s *FileMetadataStore) DeletePluginMetadata(id string) error {
	if _, exists := s.metadata[id]; !exists {
		return fmt.Errorf("插件元数据不存在: %s", id)
	}

	// 删除元数据
	delete(s.metadata, id)

	// 保存到文件
	return s.save()
}

// ListPlugins 列出所有插件元数据
func (s *FileMetadataStore) ListPlugins() ([]*models.PluginMetadata, error) {
	return s.ListPluginMetadata()
}

// ListPluginsByType 按类型列出插件元数据
func (s *FileMetadataStore) ListPluginsByType(pluginType string) ([]*models.PluginMetadata, error) {
	var result []*models.PluginMetadata
	for _, metadata := range s.metadata {
		if metadata.Type == pluginType {
			result = append(result, metadata)
		}
	}
	return result, nil
}

// ListPluginsByCategory 按分类列出插件元数据
func (s *FileMetadataStore) ListPluginsByCategory(category string) ([]*models.PluginMetadata, error) {
	var result []*models.PluginMetadata
	for _, metadata := range s.metadata {
		if metadata.Category == category {
			result = append(result, metadata)
		}
	}
	return result, nil
}

// ListPluginsByTag 按标签列出插件元数据
func (s *FileMetadataStore) ListPluginsByTag(tag string) ([]*models.PluginMetadata, error) {
	var result []*models.PluginMetadata
	for _, metadata := range s.metadata {
		for _, t := range metadata.Tags {
			if t == tag {
				result = append(result, metadata)
				break
			}
		}
	}
	return result, nil
}
