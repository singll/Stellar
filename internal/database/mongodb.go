package database

import (
	"context"
	"crypto/md5"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/StellarServer/internal/config"
	"github.com/StellarServer/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	client *mongo.Client
	db     *mongo.Database
)

// ConnectMongoDB 连接MongoDB数据库
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

// CreateDatabase 初始化数据库
func CreateDatabase() error {
	var checkFlag = 0
	var err error

	// 尝试连接数据库
	for {
		// 创建默认配置
		cfg := config.MongoDBConfig{
			URI:           config.MONGODB_IP,
			Database:      config.MONGODB_DATABASE,
			MaxPoolSize:   100,
			MinPoolSize:   10,
			MaxIdleTimeMS: 30000,
		}

		_, err = ConnectMongoDB(cfg)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Second)
		checkFlag++
		if checkFlag == 10 {
			log.Printf("Error creating database: %v", err)
			return err
		}
	}

	// 获取数据库列表
	databaseNames, err := client.ListDatabaseNames(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	// 检查数据库是否存在
	dbExists := false
	for _, name := range databaseNames {
		if name == config.MONGODB_DATABASE {
			dbExists = true
			break
		}
	}

	// 如果数据库不存在，创建数据库和初始集合
	if !dbExists {
		// 生成随机密码
		password := config.GenerateRandomString(8)
		fmt.Println("\n" + strings.Repeat("=", 50))
		fmt.Println("✨✨✨ 重要提示: 请查看以下用户名/密码 ✨✨✨")
		fmt.Println(strings.Repeat("=", 50))
		fmt.Printf("🔑 用户名/密码: StellarServer/%s\n", password)
		fmt.Println(strings.Repeat("=", 50))
		fmt.Println("✅ 请确保正确复制用户名/密码!\n")
		fmt.Println("✅ 初始密码已存储在PASSWORD文件中\n")

		// 保存密码到文件
		err = utils.WriteToFile("PASSWORD", password)
		if err != nil {
			return err
		}

		totalSteps := 16
		// 创建用户集合并插入管理员用户
		collection := db.Collection("user")
		passwordHash := fmt.Sprintf("%x", md5.Sum([]byte(password)))
		_, err = collection.InsertOne(context.Background(), bson.M{
			"username": "StellarServer",
			"password": passwordHash,
		})
		if err != nil {
			return err
		}

		log.Println("项目初始化中")
		utils.PrintProgressBar(1, totalSteps, "install")

		// 创建配置集合
		collection = db.Collection("config")
		_, err = collection.InsertOne(context.Background(), bson.M{
			"name":  "ModulesConfig",
			"value": GetModulesConfig(),
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

		utils.PrintProgressBar(2, totalSteps, "install")

		// 创建subfinder配置
		_, err = collection.InsertOne(context.Background(), bson.M{
			"name":  "SubfinderApiConfig",
			"value": GetSubfinderApiConfig(),
			"type":  "subfinder",
		})
		if err != nil {
			return err
		}

		utils.PrintProgressBar(3, totalSteps, "install")

		// 创建rad配置
		_, err = collection.InsertOne(context.Background(), bson.M{
			"name":  "RadConfig",
			"value": GetRadConfig(),
			"type":  "rad",
		})
		if err != nil {
			return err
		}

		utils.PrintProgressBar(4, totalSteps, "install")

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

		utils.PrintProgressBar(5, totalSteps, "install")

		// 更新目录扫描默认字典
		content := GetDirDict()
		size := float64(len(content)) / (1024 * 1024)
		result, err := db.Collection("dictionary").InsertOne(context.Background(), bson.M{
			"name":     "default",
			"category": "dir",
			"size":     fmt.Sprintf("%.2f", size),
		})
		if err != nil {
			return err
		}

		// 创建GridFS桶并上传文件
		bucket, err := CreateGridFSBucket(db)
		if err != nil {
			return err
		}

		err = UploadToGridFS(bucket, result.InsertedID.(primitive.ObjectID).Hex(), []byte(content))
		if err != nil {
			return err
		}

		utils.PrintProgressBar(6, totalSteps, "install")

		// 更新子域名默认字典
		content = GetDomainDict()
		size = float64(len(content)) / (1024 * 1024)
		result, err = db.Collection("dictionary").InsertOne(context.Background(), bson.M{
			"name":     "default",
			"category": "subdomain",
			"size":     fmt.Sprintf("%.2f", size),
		})
		if err != nil {
			return err
		}

		err = UploadToGridFS(bucket, result.InsertedID.(primitive.ObjectID).Hex(), []byte(content))
		if err != nil {
			return err
		}

		utils.PrintProgressBar(7, totalSteps, "install")

		// 插入敏感信息规则
		sensitiveData := GetSensitiveRules()
		if len(sensitiveData) > 0 {
			_, err = db.Collection("SensitiveRule").InsertMany(context.Background(), sensitiveData)
			if err != nil {
				return err
			}
		}

		utils.PrintProgressBar(8, totalSteps, "install")

		// 创建定时任务
		_, err = db.Collection("ScheduledTasks").InsertOne(context.Background(), bson.M{
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

		utils.PrintProgressBar(9, totalSteps, "install")

		// 创建通知API集合
		err = db.CreateCollection(context.Background(), "notification")
		if err != nil {
			return err
		}

		utils.PrintProgressBar(10, totalSteps, "install")

		// 默认端口
		_, err = db.Collection("PortDict").InsertMany(context.Background(), GetPortDict())
		if err != nil {
			return err
		}

		utils.PrintProgressBar(11, totalSteps, "install")

		// POC导入
		pocData := GetPocList()
		_, err = db.Collection("PocList").InsertMany(context.Background(), pocData)
		if err != nil {
			return err
		}

		utils.PrintProgressBar(12, totalSteps, "install")
		utils.PrintProgressBar(13, totalSteps, "install")

		// 指纹导入
		fingerprint := GetFingerprint()
		_, err = db.Collection("FingerprintRules").InsertMany(context.Background(), fingerprint)
		if err != nil {
			return err
		}

		utils.PrintProgressBar(14, totalSteps, "install")

		// 创建默认插件
		_, err = db.Collection("plugins").InsertMany(context.Background(), GetDefaultPlugins())
		if err != nil {
			return err
		}

		utils.PrintProgressBar(15, totalSteps, "install")

		// 创建默认扫描模板
		_, err = db.Collection("ScanTemplates").InsertOne(context.Background(), GetDefaultScanTemplate())
		if err != nil {
			return err
		}

		utils.PrintProgressBar(16, totalSteps, "install")

		// 创建索引
		// 页面监控URL不重复
		_, err = db.Collection("PageMonitoring").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys:    bson.D{{Key: "url", Value: 1}},
			Options: options.Index().SetUnique(true),
		})
		if err != nil {
			return err
		}

		// 页面监控Body的MD5不重复
		_, err = db.Collection("PageMonitoringBody").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys:    bson.D{{Key: "md5", Value: 1}},
			Options: options.Index().SetUnique(true),
		})
		if err != nil {
			return err
		}

		// 创建RootDomain索引
		_, err = db.Collection("PageMonitoringBody").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys:    bson.D{{Key: "domain", Value: 1}},
			Options: options.Index().SetUnique(true),
		})
		if err != nil {
			return err
		}

		// 创建asset集合索引
		_, err = db.Collection("asset").Indexes().CreateMany(context.Background(), []mongo.IndexModel{
			{Keys: bson.D{{Key: "time", Value: 1}}},
			{Keys: bson.D{{Key: "url", Value: 1}}},
			{Keys: bson.D{{Key: "host", Value: 1}}},
			{Keys: bson.D{{Key: "ip", Value: 1}}},
			{Keys: bson.D{{Key: "port", Value: 1}}},
			{Keys: bson.D{{Key: "host", Value: 1}, {Key: "port", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "project", Value: 1}}},
			{Keys: bson.D{{Key: "taskName", Value: 1}}},
			{Keys: bson.D{{Key: "rootDomain", Value: 1}}},
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
			_, err = db.Collection(collName).Indexes().CreateMany(context.Background(), []mongo.IndexModel{
				{Keys: bson.D{{Key: "project", Value: 1}}},
				{Keys: bson.D{{Key: "taskName", Value: 1}}},
				{Keys: bson.D{{Key: "rootDomain", Value: 1}}},
			})
			if err != nil {
				return err
			}
		}

		// 为RootDomain集合创建额外的索引
		_, err = db.Collection("RootDomain").Indexes().CreateMany(context.Background(), []mongo.IndexModel{
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
			_, err = db.Collection(collName).Indexes().CreateMany(context.Background(), []mongo.IndexModel{
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
		userCollection := db.Collection("user")
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

		log.Println("项目初始化成功")
	} else {
		// 如果数据库已存在，读取时区配置
		var result struct {
			Value string `bson:"value"`
		}
		err = db.Collection("config").FindOne(context.Background(), bson.M{"name": "timezone"}).Decode(&result)
		if err == nil {
			config.SetTimezone(result.Value)
		}

		// 检查并创建页面监控任务
		var pageMonTask struct {
			ID string `bson:"id"`
		}
		err = db.Collection("ScheduledTasks").FindOne(context.Background(), bson.M{"id": "page_monitoring"}).Decode(&pageMonTask)
		if err == mongo.ErrNoDocuments {
			_, err = db.Collection("ScheduledTasks").InsertOne(context.Background(), bson.M{
				"id":    "page_monitoring",
				"name":  "Page Monitoring",
				"hour":  24,
				"type":  "Page Monitoring",
				"state": true,
			})
			if err != nil {
				return err
			}
		}
	}

	// 加载指纹和项目数据
	err = LoadFingerprint()
	if err != nil {
		return err
	}

	err = LoadProjects()
	if err != nil {
		return err
	}

	return nil
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

// 以下是获取默认配置和数据的函数，需要根据原Python代码实现
func GetModulesConfig() interface{} {
	// 实现获取模块配置的逻辑
	return bson.M{}
}

func GetSubfinderApiConfig() interface{} {
	// 实现获取Subfinder API配置的逻辑
	return bson.M{}
}

func GetRadConfig() interface{} {
	// 实现获取Rad配置的逻辑
	return bson.M{}
}

func GetDirDict() string {
	// 实现获取目录字典的逻辑
	return ""
}

func GetDomainDict() string {
	// 实现获取域名字典的逻辑
	return ""
}

func GetSensitiveRules() []interface{} {
	// 实现获取敏感信息规则的逻辑
	return []interface{}{}
}

func GetPortDict() []interface{} {
	// 实现获取端口字典的逻辑
	return []interface{}{}
}

func GetPocList() []interface{} {
	// 实现获取POC列表的逻辑
	return []interface{}{}
}

func GetFingerprint() []interface{} {
	// 实现获取指纹的逻辑
	return []interface{}{}
}

func GetDefaultPlugins() []interface{} {
	// 实现获取默认插件的逻辑
	return []interface{}{}
}

func GetDefaultScanTemplate() interface{} {
	// 实现获取默认扫描模板的逻辑
	return bson.M{}
}

// GetMongoDB 获取MongoDB数据库实例
func GetMongoDB() *mongo.Database {
	return db
}
