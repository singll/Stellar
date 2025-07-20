package middleware

import (
	"github.com/StellarServer/internal/services/session"
	"github.com/gin-gonic/gin"
)

// SessionMiddleware 会话中间件，注入会话管理器到请求上下文
func SessionMiddleware(sessionManager *session.SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if sessionManager != nil {
			c.Set("session_manager", sessionManager)
		}
		c.Next()
	}
}
