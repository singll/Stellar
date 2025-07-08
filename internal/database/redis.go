package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/StellarServer/internal/config"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
)

// ConnectRedis 连接Redis数据库
func ConnectRedis(cfg config.RedisConfig) (*redis.Client, error) {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:            cfg.Addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		ConnMaxLifetime: time.Duration(cfg.MaxConnAgeMS) * time.Millisecond,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// CloseRedis 关闭Redis连接
func CloseRedis() {
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}
}

// PublishLog 发布日志到Redis通道
func PublishLog(nodeName, logMessage string) error {
	client, err := ConnectRedis(config.RedisConfig{})
	if err != nil {
		return err
	}

	ctx := context.Background()
	return client.Publish(ctx, "log_channel", fmt.Sprintf("%s:%s", nodeName, logMessage)).Err()
}

// SubscribeLogChannel 订阅日志通道
func SubscribeLogChannel() {
	client, err := ConnectRedis(config.RedisConfig{})
	if err != nil {
		log.Printf("Error connecting to Redis for log subscription: %v", err)
		return
	}

	ctx := context.Background()
	pubsub := client.Subscribe(ctx, "log_channel")
	defer pubsub.Close()

	// 在后台处理消息
	go func() {
		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				log.Printf("Error receiving message from Redis: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// 处理日志消息
			processLogMessage(msg.Payload)
		}
	}()
}

// 处理日志消息
func processLogMessage(message string) {
	// TODO: 实现日志处理逻辑
}

// RefreshConfig 刷新配置
func RefreshConfig(target, configType, id string) error {
	client, err := ConnectRedis(config.RedisConfig{})
	if err != nil {
		return err
	}

	ctx := context.Background()
	message := fmt.Sprintf("%s:%s:%s", target, configType, id)
	return client.Publish(ctx, "config_channel", message).Err()
}
