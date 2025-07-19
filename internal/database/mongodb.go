package database

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/StellarServer/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	client *mongo.Client
	db     *mongo.Database
)

// ConnectMongoDB 连接MongoDB数据库（向后兼容）
func ConnectMongoDB(cfg config.MongoDBConfig) (*mongo.Client, error) {
	// 设置连接超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := cfg.URI
	// 如有账号密码，拼接到uri
	if cfg.Username != "" && cfg.Password != "" {
		// 处理mongodb://host:port/database 变为 mongodb://user:pass@host:port/database
		uri = addMongoAuthToURI(cfg.URI, cfg.Username, cfg.Password)
	}

	// 创建连接选项
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetMaxConnIdleTime(time.Duration(cfg.MaxIdleTimeMS) * time.Millisecond)

	// 连接MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// 测试连接
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return client, nil
}

// addMongoAuthToURI 拼接账号密码到MongoDB URI
func addMongoAuthToURI(uri, user, pass string) string {
	// 只处理mongodb://开头
	if !strings.HasPrefix(uri, "mongodb://") {
		return uri
	}
	uri = strings.TrimPrefix(uri, "mongodb://")
	return "mongodb://" + user + ":" + pass + "@" + uri
}

// CloseMongoDB 关闭MongoDB连接
func CloseMongoDB() {
	if client != nil {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}
}

// CreateGridFSBucket 创建GridFS桶
func CreateGridFSBucket(db *mongo.Database) (*gridfs.Bucket, error) {
	bucket, err := gridfs.NewBucket(
		db,
		options.GridFSBucket(),
	)
	return bucket, err
}

// UploadToGridFS 上传文件到GridFS
func UploadToGridFS(bucket *gridfs.Bucket, filename string, data []byte) error {
	uploadStream, err := bucket.OpenUploadStream(filename)
	if err != nil {
		return err
	}
	defer uploadStream.Close()

	_, err = uploadStream.Write(data)
	return err
}

// LoadFingerprint 加载指纹数据
func LoadFingerprint() error {
	// 实现加载指纹的逻辑
	return nil
}

// LoadProjects 加载项目数据
func LoadProjects() error {
	// 实现加载项目的逻辑
	return nil
}

// GetMongoDB 获取MongoDB数据库实例（向后兼容）
func GetMongoDB() *mongo.Database {
	return db
}
