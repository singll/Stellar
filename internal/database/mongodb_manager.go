package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/StellarServer/internal/config"
	pkgerrors "github.com/StellarServer/internal/pkg/errors"
	"github.com/StellarServer/internal/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
)

// MongoDBManager manages MongoDB connections
type MongoDBManager struct {
	client   *mongo.Client
	database *mongo.Database
	config   config.MongoDBConfig
}

// NewMongoDBManager creates a new MongoDB manager
func NewMongoDBManager(cfg config.MongoDBConfig) (*MongoDBManager, error) {
	manager := &MongoDBManager{
		config: cfg,
	}

	client, err := manager.connect()
	if err != nil {
		logger.Error("NewMongoDBManager connect failed", map[string]interface{}{"config": cfg, "error": err})
		return nil, pkgerrors.WrapDatabaseError(err, "创建MongoDB管理器")
	}

	manager.client = client
	manager.database = client.Database(cfg.Database)

	// 初始化数据库
	if err := manager.InitializeDatabase(); err != nil {
		logger.Error("NewMongoDBManager initialize database failed", map[string]interface{}{"error": err})
		return nil, pkgerrors.WrapDatabaseError(err, "初始化数据库")
	}

	return manager, nil
}

// connect establishes MongoDB connection
func (m *MongoDBManager) connect() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := m.config.URI
	if m.config.Username != "" && m.config.Password != "" {
		uri = m.addAuthToURI(m.config.URI, m.config.Username, m.config.Password)
	}

	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(m.config.MaxPoolSize).
		SetMinPoolSize(m.config.MinPoolSize).
		SetMaxConnIdleTime(time.Duration(m.config.MaxIdleTimeMS) * time.Millisecond)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, nil
}

// addAuthToURI adds authentication to MongoDB URI
func (m *MongoDBManager) addAuthToURI(uri, user, pass string) string {
	if !strings.HasPrefix(uri, "mongodb://") {
		return uri
	}
	uri = strings.TrimPrefix(uri, "mongodb://")
	return "mongodb://" + user + ":" + pass + "@" + uri
}

// InitializeDatabase 初始化数据库
func (m *MongoDBManager) InitializeDatabase() error {
	// 获取数据库列表
	databaseNames, err := m.client.ListDatabaseNames(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	// 检查数据库是否存在
	dbExists := false
	for _, name := range databaseNames {
		if name == m.config.Database {
			dbExists = true
			break
		}
	}

	// 如果数据库不存在，创建数据库和初始集合
	if !dbExists {
		logger.Info("数据库不存在，开始初始化", map[string]interface{}{
			"database": m.config.Database,
		})
		fmt.Printf("🔧 正在初始化数据库: %s\n", m.config.Database)

		// 生成随机密码
		password := config.GenerateRandomString(12)
		fmt.Printf("📝 生成管理员密码: %s\n", password)
		fmt.Printf("⚠️  请妥善保存此密码，首次登录后建议修改\n")

		// 创建用户集合并插入管理员用户
		collection := m.database.Collection("user")
		// 使用bcrypt加密密码，更安全
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		_, err = collection.InsertOne(context.Background(), bson.M{
			"username":       "admin",
			"hashedPassword": string(hashedPassword),
			"email":          "admin@stellarserver.com",
			"roles":          []string{"admin"},
			"created":        time.Now(),
			"lastLogin":      time.Now(),
		})
		if err != nil {
			return err
		}

		fmt.Printf("✅ 创建管理员用户: admin\n")

		// 创建配置集合
		collection = m.database.Collection("config")
		_, err = collection.InsertOne(context.Background(), bson.M{
			"name":  "ModulesConfig",
			"value": getModulesConfig(),
			"type":  "system",
		})
		if err != nil {
			return err
		}

		_, err = collection.InsertOne(context.Background(), bson.M{
			"name":  "timezone",
			"value": "Asia/Shanghai",
			"type":  "system",
		})
		if err != nil {
			return err
		}

		// 创建subfinder配置
		_, err = collection.InsertOne(context.Background(), bson.M{
			"name":  "SubfinderApiConfig",
			"value": getSubfinderApiConfig(),
			"type":  "subfinder",
		})
		if err != nil {
			return err
		}

		// 创建rad配置
		_, err = collection.InsertOne(context.Background(), bson.M{
			"name":  "RadConfig",
			"value": getRadConfig(),
			"type":  "rad",
		})
		if err != nil {
			return err
		}

		// 创建通知配置
		_, err = collection.InsertOne(context.Background(), bson.M{
			"name":                          "notification",
			"dirScanNotification":           true,
			"portScanNotification":          true,
			"sensitiveNotification":         true,
			"subdomainTakeoverNotification": true,
			"pageMonNotification":           true,
			"subdomainNotification":         true,
			"vulNotification":               true,
			"type":                          "notification",
		})
		if err != nil {
			return err
		}

		// 创建定时任务
		_, err = m.database.Collection("ScheduledTasks").InsertOne(context.Background(), bson.M{
			"id":      "page_monitoring",
			"name":    "Page Monitoring",
			"hour":    24,
			"node":    []string{},
			"allNode": true,
			"type":    "Page Monitoring",
			"state":   true,
		})
		if err != nil {
			return err
		}

		// 创建通知API集合
		err = m.database.CreateCollection(context.Background(), "notification")
		if err != nil {
			return err
		}

		// 创建索引
		// 页面监控URL不重复
		_, err = m.database.Collection("PageMonitoring").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys:    bson.D{{Key: "url", Value: 1}},
			Options: options.Index().SetUnique(true),
		})
		if err != nil {
			return err
		}

		// 页面监控Body的MD5不重复
		_, err = m.database.Collection("PageMonitoring").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys:    bson.D{{Key: "bodyMD5", Value: 1}},
			Options: options.Index().SetUnique(true),
		})
		if err != nil {
			return err
		}

		// 创建其他集合的索引
		collections := []string{
			"subdomain", "UrlScan", "crawler", "SensitiveResult",
			"DirScanResult", "vulnerability",
		}

		for _, collName := range collections {
			_, err = m.database.Collection(collName).Indexes().CreateMany(context.Background(), []mongo.IndexModel{
				{Keys: bson.D{{Key: "project", Value: 1}}},
				{Keys: bson.D{{Key: "taskName", Value: 1}}},
				{Keys: bson.D{{Key: "rootDomain", Value: 1}}},
			})
			if err != nil {
				return err
			}
		}

		// 为RootDomain集合创建额外的索引
		_, err = m.database.Collection("RootDomain").Indexes().CreateMany(context.Background(), []mongo.IndexModel{
			{Keys: bson.D{{Key: "project", Value: 1}}},
			{Keys: bson.D{{Key: "taskName", Value: 1}}},
			{Keys: bson.D{{Key: "domain", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "time", Value: 1}}},
		})
		if err != nil {
			return err
		}

		// 为app和mp集合创建索引
		for _, collName := range []string{"app", "mp"} {
			_, err = m.database.Collection(collName).Indexes().CreateMany(context.Background(), []mongo.IndexModel{
				{Keys: bson.D{{Key: "project", Value: 1}}},
				{Keys: bson.D{{Key: "taskName", Value: 1}}},
				{Keys: bson.D{{Key: "time", Value: 1}}},
				{Keys: bson.D{{Key: "name", Value: 1}}},
			})
			if err != nil {
				return err
			}
		}

		// 创建用户集合唯一索引
		userCollection := m.database.Collection("user")
		_, err = userCollection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "username", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		})
		if err != nil {
			return err
		}

		fmt.Printf("✅ 数据库初始化完成\n")
		fmt.Printf("🔑 管理员账户信息:\n")
		fmt.Printf("   用户名: admin\n")
		fmt.Printf("   密码: %s\n", password)
		fmt.Printf("   邮箱: admin@stellarserver.com\n")
		fmt.Printf("⚠️  请使用以上信息登录系统，登录后请及时修改密码\n")

		logger.Info("数据库初始化成功", map[string]interface{}{
			"database": m.config.Database,
		})
	} else {
		logger.Info("数据库已存在，跳过初始化", map[string]interface{}{
			"database": m.config.Database,
		})
		fmt.Printf("✅ 数据库已存在: %s\n", m.config.Database)
	}

	return nil
}

// GetDatabase returns the MongoDB database instance
func (m *MongoDBManager) GetDatabase() *mongo.Database {
	return m.database
}

// Health checks MongoDB connection health
func (m *MongoDBManager) Health() error {
	if m.client == nil {
		return pkgerrors.NewAppError(pkgerrors.CodeDatabaseError, "MongoDB客户端未初始化", 500)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.client.Ping(ctx, readpref.Primary())
	if err != nil {
		return pkgerrors.WrapError(err, pkgerrors.CodeDatabaseError, "检查MongoDB连接健康状态", 500)
	}

	return nil
}

// Close closes MongoDB connection
func (m *MongoDBManager) Close() error {
	if m.client != nil {
		return m.client.Disconnect(context.Background())
	}
	return nil
}

// Transaction executes a function within a MongoDB transaction
func (m *MongoDBManager) Transaction(fn func(sessCtx mongo.SessionContext) error) error {
	session, err := m.client.StartSession()
	if err != nil {
		logger.Error("Transaction start session failed", map[string]interface{}{"error": err})
		return pkgerrors.WrapDatabaseError(err, "启动MongoDB会话")
	}
	defer session.EndSession(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})

	if err != nil {
		logger.Error("Transaction execution failed", map[string]interface{}{"error": err})
		return pkgerrors.WrapDatabaseError(err, "执行MongoDB事务")
	}

	return nil
}

// 以下是获取默认配置和数据的函数
func getModulesConfig() interface{} {
	return bson.M{
		"subdomain": true,
		"portscan":  true,
		"vulnscan":  true,
		"dirscan":   true,
		"sensitive": true,
		"pagemoni":  true,
	}
}

func getSubfinderApiConfig() interface{} {
	return bson.M{
		"enabled": false,
		"api_key": "",
		"url":     "https://api.subfinder.io",
	}
}

func getRadConfig() interface{} {
	return bson.M{
		"enabled": false,
		"path":    "/usr/local/bin/rad",
		"timeout": 30,
	}
}
