package middleware

import (
	"net/http"
	"strings"

	"github.com/StellarServer/internal/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 获取Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing authorization header",
			})
			c.Abort()
			return
		}

		// 检查Bearer token格式
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// 验证JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 验证签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.NewUnauthorizedError("Invalid signing method")
			}
			// 返回密钥
			return []byte("your-secret-key"), nil // TODO: 从配置获取
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// 提取claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
			c.Set("username", claims["username"])
			c.Set("roles", claims["roles"])
		}

		c.Next()
	})
}

// RoleMiddleware 角色权限中间件
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 先确保用户已认证
		userRoles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "No roles found in token",
			})
			c.Abort()
			return
		}

		// 检查用户角色
		roles, ok := userRoles.([]interface{})
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid roles format",
			})
			c.Abort()
			return
		}

		// 检查是否有允许的角色
		hasPermission := false
		for _, userRole := range roles {
			userRoleStr, ok := userRole.(string)
			if !ok {
				continue
			}
			for _, allowedRole := range allowedRoles {
				if userRoleStr == allowedRole {
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}