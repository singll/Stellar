package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/StellarServer/internal/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// SessionData 会话数据结构
type SessionData struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Roles     []string  `json:"roles"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
}

// SessionManager 会话管理器
type SessionManager struct {
	redisClient *redis.Client
	// 会话过期时间：8小时
	sessionExpiry time.Duration
	// 刷新阈值：如果距离过期时间少于1小时，则自动刷新
	refreshThreshold time.Duration
}

// NewSessionManager 创建会话管理器
func NewSessionManager(redisClient *redis.Client) *SessionManager {
	return &SessionManager{
		redisClient:      redisClient,
		sessionExpiry:    8 * time.Hour,
		refreshThreshold: 1 * time.Hour,
	}
}

// CreateSession 创建新会话
func (sm *SessionManager) CreateSession(ctx context.Context, token string, userID, username string, roles []string) error {
	// 如果用户已有会话，先删除旧会话
	sm.deleteUserSessions(ctx, userID)

	sessionData := &SessionData{
		UserID:    userID,
		Username:  username,
		Roles:     roles,
		Token:     token,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}

	// 序列化会话数据
	sessionJSON, err := json.Marshal(sessionData)
	if err != nil {
		logger.Error("序列化会话数据失败", map[string]interface{}{
			"error": err,
			"user":  username,
		})
		return fmt.Errorf("序列化会话数据失败: %w", err)
	}

	// 存储到Redis，设置8小时过期时间
	sessionKey := sm.getSessionKey(token)
	userSessionKey := sm.getUserSessionKey(userID)

	// 使用事务确保原子性
	pipe := sm.redisClient.Pipeline()
	pipe.Set(ctx, sessionKey, sessionJSON, sm.sessionExpiry)
	pipe.Set(ctx, userSessionKey, token, sm.sessionExpiry) // 用户ID到token的映射

	_, err = pipe.Exec(ctx)
	if err != nil {
		logger.Error("存储会话到Redis失败", map[string]interface{}{
			"error": err,
			"user":  username,
			"key":   sessionKey,
		})
		return fmt.Errorf("存储会话失败: %w", err)
	}

	logger.Info("会话创建成功", map[string]interface{}{
		"user":        username,
		"session_key": sessionKey,
		"expiry":      sm.sessionExpiry,
	})

	return nil
}

// GetSession 获取会话数据
func (sm *SessionManager) GetSession(ctx context.Context, token string) (*SessionData, error) {
	sessionKey := sm.getSessionKey(token)

	// 从Redis获取会话数据
	sessionJSON, err := sm.redisClient.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			logger.Warn("会话不存在", map[string]interface{}{
				"session_key": sessionKey,
			})
			return nil, fmt.Errorf("会话不存在或已过期")
		}
		logger.Error("从Redis获取会话失败", map[string]interface{}{
			"error": err,
			"key":   sessionKey,
		})
		return nil, fmt.Errorf("获取会话失败: %w", err)
	}

	// 反序列化会话数据
	var sessionData SessionData
	err = json.Unmarshal([]byte(sessionJSON), &sessionData)
	if err != nil {
		logger.Error("反序列化会话数据失败", map[string]interface{}{
			"error": err,
			"key":   sessionKey,
		})
		return nil, fmt.Errorf("解析会话数据失败: %w", err)
	}

	// 检查是否需要刷新会话
	if sm.shouldRefreshSession(sessionData) {
		if err := sm.refreshSession(ctx, token, &sessionData); err != nil {
			logger.Warn("刷新会话失败，但继续使用当前会话", map[string]interface{}{
				"error": err,
				"user":  sessionData.Username,
			})
		}
	}

	return &sessionData, nil
}

// RefreshSession 刷新会话
func (sm *SessionManager) RefreshSession(ctx context.Context, token string) error {
	sessionData, err := sm.GetSession(ctx, token)
	if err != nil {
		return err
	}

	return sm.refreshSession(ctx, token, sessionData)
}

// DeleteSession 删除会话
func (sm *SessionManager) DeleteSession(ctx context.Context, token string) error {
	sessionKey := sm.getSessionKey(token)

	// 先获取会话数据以获取用户ID
	sessionData, err := sm.GetSession(ctx, token)
	if err == nil && sessionData != nil {
		// 删除用户会话映射
		userSessionKey := sm.getUserSessionKey(sessionData.UserID)
		sm.redisClient.Del(ctx, userSessionKey)
	}

	err = sm.redisClient.Del(ctx, sessionKey).Err()
	if err != nil {
		logger.Error("删除会话失败", map[string]interface{}{
			"error": err,
			"key":   sessionKey,
		})
		return fmt.Errorf("删除会话失败: %w", err)
	}

	logger.Info("会话删除成功", map[string]interface{}{
		"session_key": sessionKey,
	})

	return nil
}

// IsSessionValid 检查会话是否有效
func (sm *SessionManager) IsSessionValid(ctx context.Context, token string) bool {
	_, err := sm.GetSession(ctx, token)
	return err == nil
}

// GetSessionExpiry 获取会话过期时间
func (sm *SessionManager) GetSessionExpiry() time.Duration {
	return sm.sessionExpiry
}

// GetUserSession 根据用户ID获取会话
func (sm *SessionManager) GetUserSession(ctx context.Context, userID string) (*SessionData, error) {
	userSessionKey := sm.getUserSessionKey(userID)

	// 获取用户的token
	token, err := sm.redisClient.Get(ctx, userSessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("用户会话不存在")
		}
		return nil, fmt.Errorf("获取用户会话失败: %w", err)
	}

	// 使用token获取完整会话数据
	return sm.GetSession(ctx, token)
}

// DeleteUserSessions 删除用户的所有会话
func (sm *SessionManager) DeleteUserSessions(ctx context.Context, userID string) error {
	return sm.deleteUserSessions(ctx, userID)
}

// GetSessionStatus 获取会话状态信息
func (sm *SessionManager) GetSessionStatus(ctx context.Context, token string) (map[string]interface{}, error) {
	sessionData, err := sm.GetSession(ctx, token)
	if err != nil {
		return nil, err
	}

	expiresAt := sessionData.CreatedAt.Add(sm.sessionExpiry)
	timeUntilExpiry := expiresAt.Sub(time.Now())

	return map[string]interface{}{
		"user_id":           sessionData.UserID,
		"username":          sessionData.Username,
		"roles":             sessionData.Roles,
		"created_at":        sessionData.CreatedAt,
		"last_used":         sessionData.LastUsed,
		"expires_at":        expiresAt,
		"time_until_expiry": timeUntilExpiry,
		"is_expired":        timeUntilExpiry <= 0,
		"needs_refresh":     sm.shouldRefreshSession(*sessionData),
	}, nil
}

// shouldRefreshSession 检查是否需要刷新会话
func (sm *SessionManager) shouldRefreshSession(sessionData SessionData) bool {
	// 如果距离过期时间少于刷新阈值，则刷新
	timeUntilExpiry := sessionData.CreatedAt.Add(sm.sessionExpiry).Sub(time.Now())
	return timeUntilExpiry < sm.refreshThreshold
}

// refreshSession 刷新会话
func (sm *SessionManager) refreshSession(ctx context.Context, token string, sessionData *SessionData) error {
	// 更新最后使用时间
	sessionData.LastUsed = time.Now()

	// 重新序列化
	sessionJSON, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("序列化刷新后的会话数据失败: %w", err)
	}

	// 重新存储到Redis，重置过期时间
	sessionKey := sm.getSessionKey(token)
	userSessionKey := sm.getUserSessionKey(sessionData.UserID)

	pipe := sm.redisClient.Pipeline()
	pipe.Set(ctx, sessionKey, sessionJSON, sm.sessionExpiry)
	pipe.Set(ctx, userSessionKey, token, sm.sessionExpiry)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("刷新会话存储失败: %w", err)
	}

	logger.Info("会话刷新成功", map[string]interface{}{
		"user":        sessionData.Username,
		"session_key": sessionKey,
		"last_used":   sessionData.LastUsed,
	})

	return nil
}

// deleteUserSessions 删除用户的所有会话（内部方法）
func (sm *SessionManager) deleteUserSessions(ctx context.Context, userID string) error {
	userSessionKey := sm.getUserSessionKey(userID)

	// 获取用户的token
	token, err := sm.redisClient.Get(ctx, userSessionKey).Result()
	if err == nil {
		// 删除会话数据
		sessionKey := sm.getSessionKey(token)
		sm.redisClient.Del(ctx, sessionKey)
	}

	// 删除用户会话映射
	sm.redisClient.Del(ctx, userSessionKey)

	logger.Info("用户会话删除成功", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

// getSessionKey 生成会话键
func (sm *SessionManager) getSessionKey(token string) string {
	return fmt.Sprintf("session:%s", token)
}

// getUserSessionKey 生成用户会话键
func (sm *SessionManager) getUserSessionKey(userID string) string {
	return fmt.Sprintf("user_session:%s", userID)
}

// CleanupExpiredSessions 清理过期会话（可选的后台任务）
func (sm *SessionManager) CleanupExpiredSessions(ctx context.Context) error {
	// Redis会自动清理过期的键，这里可以添加额外的清理逻辑
	// 例如：清理用户相关的其他数据
	logger.Info("开始清理过期会话", nil)

	// 这里可以添加自定义的清理逻辑
	// 例如：清理用户活动日志、临时文件等

	logger.Info("过期会话清理完成", nil)
	return nil
}
