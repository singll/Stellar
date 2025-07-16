package middleware

import (
	"net/http"
	"time"

	"github.com/StellarServer/internal/pkg/logger"
	"github.com/gin-gonic/gin"
)

// CORS 跨域资源共享中间件
func CORS() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)

		// 构建日志信息
		logData := map[string]interface{}{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"query":      raw,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"latency":    latency,
			"size":       c.Writer.Size(),
		}

		// 添加用户信息（如果已认证）
		if userID, exists := c.Get("user_id"); exists {
			logData["user_id"] = userID
		}
		if username, exists := c.Get("username"); exists {
			logData["username"] = username
		}

		// 记录日志
		if c.Writer.Status() >= 400 {
			logger.Error("HTTP Request Error", logData)
		} else {
			logger.Info("HTTP Request", logData)
		}
	})
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered", map[string]interface{}{
					"error":  err,
					"method": c.Request.Method,
					"path":   c.Request.URL.Path,
					"ip":     c.ClientIP(),
				})

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	})
}

// RateLimit 限流中间件
func RateLimit() gin.HandlerFunc {
	// TODO: 实现基于Redis的限流逻辑
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()
	})
}

// Security 安全头中间件
func Security() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	})
}