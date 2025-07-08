package assetdiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/redis/go-redis/v9"
)

// RedisResultHandler Redis结果处理器
type RedisResultHandler struct {
	client *redis.Client
}

// NewRedisResultHandler 创建Redis结果处理器
func NewRedisResultHandler(client *redis.Client) *RedisResultHandler {
	return &RedisResultHandler{
		client: client,
	}
}

// HandleDiscoveryResult 处理资产发现结果
func (h *RedisResultHandler) HandleDiscoveryResult(task *DiscoveryTask, result *models.DiscoveryResult) error {
	// 序列化结果
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	// 保存到Redis
	ctx := context.Background()
	key := fmt.Sprintf("discovery:result:%s:%s", task.ID, result.ID.Hex())
	err = h.client.Set(ctx, key, data, 24*time.Hour).Err()
	if err != nil {
		return err
	}

	// 添加到任务结果集合
	listKey := fmt.Sprintf("discovery:task:%s:results", task.ID)
	return h.client.RPush(ctx, listKey, key).Err()
}

// UpdateTaskStatus 更新任务状态
func (h *RedisResultHandler) UpdateTaskStatus(taskID string, status string, progress float64) error {
	// 创建状态数据
	data := map[string]interface{}{
		"status":    status,
		"progress":  progress,
		"updatedAt": time.Now(),
	}

	// 如果任务完成或失败，设置完成时间
	if status == "completed" || status == "failed" || status == "stopped" {
		data["completedAt"] = time.Now()
	}

	// 如果任务开始运行，设置开始时间
	if status == "running" {
		data["startedAt"] = time.Now()
	}

	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 保存到Redis
	ctx := context.Background()
	key := fmt.Sprintf("discovery:task:%s:status", taskID)
	return h.client.Set(ctx, key, jsonData, 24*time.Hour).Err()
}

// SaveDiscoveryResult 保存资产发现结果
func (h *RedisResultHandler) SaveDiscoveryResult(result *models.DiscoveryResult) error {
	// 序列化结果
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	// 保存到Redis
	ctx := context.Background()
	key := fmt.Sprintf("discovery:result:%s", result.ID.Hex())
	return h.client.Set(ctx, key, data, 24*time.Hour).Err()
}

// GetDiscoveryResults 获取资产发现结果
func (h *RedisResultHandler) GetDiscoveryResults(taskID string) ([]*models.DiscoveryResult, error) {
	ctx := context.Background()

	// 获取任务结果列表
	listKey := fmt.Sprintf("discovery:task:%s:results", taskID)
	keys, err := h.client.LRange(ctx, listKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	// 如果没有结果，返回空列表
	if len(keys) == 0 {
		return []*models.DiscoveryResult{}, nil
	}

	// 获取每个结果
	var results []*models.DiscoveryResult
	for _, key := range keys {
		// 获取结果数据
		data, err := h.client.Get(ctx, key).Bytes()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, err
		}

		// 反序列化结果
		var result models.DiscoveryResult
		if err := json.Unmarshal(data, &result); err != nil {
			continue
		}

		results = append(results, &result)
	}

	return results, nil
}
