package database

import (
	"context"
	"fmt"
	"time"

	"github.com/StellarServer/internal/config"
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
		return nil, err
	}

	manager.client = client
	return manager, nil
}

// connect establishes Redis connection
func (r *RedisManager) connect() (*redis.Client, error) {
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
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// GetClient returns the Redis client instance
func (r *RedisManager) GetClient() *redis.Client {
	return r.client
}

// Health checks Redis connection health
func (r *RedisManager) Health() error {
	if r.client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.client.Ping(ctx).Result()
	return err
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

// LPush pushes elements to list
func (r *RedisManager) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

// RPop pops element from list
func (r *RedisManager) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

// Publish publishes message to channel
func (r *RedisManager) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to channels
func (r *RedisManager) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// Pipeline creates a pipeline
func (r *RedisManager) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// TxPipeline creates a transaction pipeline
func (r *RedisManager) TxPipeline() redis.Pipeliner {
	return r.client.TxPipeline()
}