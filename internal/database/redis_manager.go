package database

import (
	"context"
	"fmt"
	"time"

	"github.com/StellarServer/internal/config"
	pkgerrors "github.com/StellarServer/internal/pkg/errors"
	"github.com/StellarServer/internal/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// RedisManager manages Redis connections
type RedisManager struct {
	client *redis.Client
	config config.RedisConfig
}

// NewRedisManager creates a new Redis manager
func NewRedisManager(cfg config.RedisConfig) (*RedisManager, error) {
	manager := &RedisManager{
		config: cfg,
	}

	client, err := manager.connect()
	if err != nil {
		logger.Error("NewRedisManager connect failed", map[string]interface{}{"config": cfg, "error": err})
		return nil, pkgerrors.WrapError(err, pkgerrors.CodeRedisError, "创建Redis管理器", 500)
	}

	manager.client = client

	// 初始化Redis
	if err := manager.InitializeRedis(); err != nil {
		logger.Error("NewRedisManager initialize redis failed", map[string]interface{}{"error": err})
		return nil, pkgerrors.WrapError(err, pkgerrors.CodeRedisError, "初始化Redis", 500)
	}

	return manager, nil
}

// connect establishes Redis connection
func (r *RedisManager) connect() (*redis.Client, error) {
	logger.Info("Connecting to Redis", map[string]interface{}{
		"addr":        r.config.Addr,
		"db":          r.config.DB,
		"poolSize":    r.config.PoolSize,
		"hasPassword": r.config.Password != "",
	})

	client := redis.NewClient(&redis.Options{
		Addr:     r.config.Addr,
		Password: r.config.Password,
		DB:       r.config.DB,
		PoolSize: r.config.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		logger.Error("Redis connection test failed", map[string]interface{}{
			"addr":  r.config.Addr,
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connection successful", map[string]interface{}{
		"addr": r.config.Addr,
	})

	return client, nil
}

// InitializeRedis 初始化Redis
func (r *RedisManager) InitializeRedis() error {
	ctx := context.Background()

	// 检查Redis是否为空（新安装）
	keys, err := r.client.Keys(ctx, "*").Result()
	if err != nil {
		return fmt.Errorf("failed to check Redis keys: %w", err)
	}

	if len(keys) == 0 {
		fmt.Printf("🔧 正在初始化Redis: %s\n", r.config.Addr)

		// 设置默认配置
		defaultConfigs := map[string]string{
			"stellar:config:version":     "1.0.0",
			"stellar:config:initialized": time.Now().Format(time.RFC3339),
			"stellar:config:timezone":    "Asia/Shanghai",
		}

		for key, value := range defaultConfigs {
			err := r.client.Set(ctx, key, value, 0).Err()
			if err != nil {
				return fmt.Errorf("failed to set Redis key %s: %w", key, err)
			}
		}

		// 设置默认的扫描配置
		scanConfigs := map[string]interface{}{
			"stellar:scan:subdomain:timeout":        10,
			"stellar:scan:subdomain:maxConcurrency": 100,
			"stellar:scan:portscan:timeout":         5,
			"stellar:scan:portscan:maxConcurrency":  50,
			"stellar:scan:vulnscan:timeout":         30,
			"stellar:scan:vulnscan:maxConcurrency":  10,
		}

		for key, value := range scanConfigs {
			err := r.client.Set(ctx, key, value, 0).Err()
			if err != nil {
				return fmt.Errorf("failed to set Redis key %s: %w", key, err)
			}
		}

		// 设置默认的任务队列配置
		queueConfigs := map[string]interface{}{
			"stellar:queue:maxSize":    1000,
			"stellar:queue:retryCount": 3,
			"stellar:queue:retryDelay": 300,
			"stellar:queue:timeout":    3600,
		}

		for key, value := range queueConfigs {
			err := r.client.Set(ctx, key, value, 0).Err()
			if err != nil {
				return fmt.Errorf("failed to set Redis key %s: %w", key, err)
			}
		}

		// 设置默认的节点管理配置
		nodeConfigs := map[string]interface{}{
			"stellar:node:heartbeatInterval": 30,
			"stellar:node:heartbeatTimeout":  90,
			"stellar:node:autoRemove":        false,
			"stellar:node:autoRemoveAfter":   86400,
		}

		for key, value := range nodeConfigs {
			err := r.client.Set(ctx, key, value, 0).Err()
			if err != nil {
				return fmt.Errorf("failed to set Redis key %s: %w", key, err)
			}
		}

		fmt.Printf("✅ Redis初始化完成\n")
		fmt.Printf("📊 Redis连接信息:\n")
		fmt.Printf("   地址: %s\n", r.config.Addr)
		fmt.Printf("   数据库: %d\n", r.config.DB)
		fmt.Printf("   连接池大小: %d\n", r.config.PoolSize)
		if r.config.Password != "" {
			fmt.Printf("   密码: 已配置\n")
		} else {
			fmt.Printf("   密码: 无\n")
		}

		logger.Info("Redis初始化成功", map[string]interface{}{
			"addr": r.config.Addr,
			"db":   r.config.DB,
		})
	} else {
		fmt.Printf("✅ Redis已初始化: %s (包含 %d 个键)\n", r.config.Addr, len(keys))
		logger.Info("Redis已存在，跳过初始化", map[string]interface{}{
			"addr": r.config.Addr,
			"keys": len(keys),
		})
	}

	return nil
}

// GetClient returns the Redis client instance
func (r *RedisManager) GetClient() *redis.Client {
	return r.client
}

// Health checks Redis connection health
func (r *RedisManager) Health() error {
	if r.client == nil {
		logger.Error("Health Redis client not initialized", nil)
		return pkgerrors.NewAppError(pkgerrors.CodeRedisError, "Redis客户端未初始化", 500)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.client.Ping(ctx).Result()
	if err != nil {
		logger.Error("Health Redis ping failed", map[string]interface{}{"error": err})
		return pkgerrors.WrapError(err, pkgerrors.CodeRedisError, "检查Redis连接健康状态", 500)
	}

	return nil
}

// Close closes Redis connection
func (r *RedisManager) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// Set sets a key-value pair with expiration
func (r *RedisManager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get gets a value by key
func (r *RedisManager) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Del deletes keys
func (r *RedisManager) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if keys exist
func (r *RedisManager) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Expire sets expiration for a key
func (r *RedisManager) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// HSet sets field in hash
func (r *RedisManager) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

// HGet gets field from hash
func (r *RedisManager) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll gets all fields from hash
func (r *RedisManager) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel deletes fields from hash
func (r *RedisManager) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// LPush pushes values to the left of a list
func (r *RedisManager) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

// RPop pops a value from the right of a list
func (r *RedisManager) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

// LLen gets the length of a list
func (r *RedisManager) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

// SAdd adds members to a set
func (r *RedisManager) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

// SMembers gets all members of a set
func (r *RedisManager) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// SRem removes members from a set
func (r *RedisManager) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

// ZAdd adds members to a sorted set
func (r *RedisManager) ZAdd(ctx context.Context, key string, score float64, member string) error {
	return r.client.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
}

// ZRange gets members from a sorted set by range
func (r *RedisManager) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(ctx, key, start, stop).Result()
}

// ZRem removes members from a sorted set
func (r *RedisManager) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.ZRem(ctx, key, members...).Err()
}

// Publish publishes a message to a channel
func (r *RedisManager) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to a channel
func (r *RedisManager) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// FlushDB flushes the current database
func (r *RedisManager) FlushDB(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// FlushAll flushes all databases
func (r *RedisManager) FlushAll(ctx context.Context) error {
	return r.client.FlushAll(ctx).Err()
}

// Info gets Redis server information
func (r *RedisManager) Info(ctx context.Context, section ...string) (string, error) {
	return r.client.Info(ctx, section...).Result()
}

// ConfigGet gets Redis configuration
func (r *RedisManager) ConfigGet(ctx context.Context, parameter string) (map[string]string, error) {
	return r.client.ConfigGet(ctx, parameter).Result()
}

// ConfigSet sets Redis configuration
func (r *RedisManager) ConfigSet(ctx context.Context, parameter, value string) error {
	return r.client.ConfigSet(ctx, parameter, value).Err()
}
