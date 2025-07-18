package api

import (
	"net/http"
	"time"

	"github.com/StellarServer/internal/models"
	pkgerrors "github.com/StellarServer/internal/pkg/errors"
	"github.com/StellarServer/internal/pkg/logger"
	"github.com/StellarServer/internal/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	DB *mongo.Database
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(db *mongo.Database) *AuthHandler {
	return &AuthHandler{DB: db}
}

// RegisterRoutes 注册路由
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/login", h.Login)
	router.GET("/info", AuthMiddleware(), GetUserInfo) // 用户信息需要认证
	router.POST("/logout", Logout)                     // logout不需要认证中间件
	router.POST("/register", h.Register)
}

// JWTSecret JWT密钥
var JWTSecret = []byte("StellarServer-JWT-Secret")

// Claims JWT声明
type Claims struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.StandardClaims
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Code int `json:"code"`
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	Code int `json:"code"`
	Data struct {
		Username string   `json:"username"`
		Roles    []string `json:"roles"`
	} `json:"data"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// Login 处理登录请求
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Login参数绑定失败", map[string]interface{}{"error": err})
		utils.HandleError(c, pkgerrors.NewAppErrorWithCause(pkgerrors.CodeBadRequest, "无效的请求参数", 400, err))
		return
	}

	db := h.DB
	if db == nil {
		logger.Error("Login数据库未初始化", nil)
		utils.HandleError(c, pkgerrors.NewAppError(pkgerrors.CodeInternalError, "数据库连接未初始化", 500))
		return
	}

	// 支持用户名或邮箱登录
	identifier := req.Username
	if identifier == "" {
		identifier = req.Email
	}
	if identifier == "" {
		logger.Error("Login用户名或邮箱为空", nil)
		utils.HandleError(c, pkgerrors.NewAppError(pkgerrors.CodeBadRequest, "用户名或邮箱不能为空", 400))
		return
	}

	user, err := models.ValidateUser(db, identifier, req.Password)
	if err != nil {
		logger.Error("Login用户校验失败", map[string]interface{}{"identifier": identifier, "error": err})
		utils.HandleError(c, err)
		return
	}

	// 生成JWT令牌
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: user.Username,
		Roles:    user.Roles,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "StellarServer",
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		logger.Error("Login生成JWT失败", map[string]interface{}{"error": err})
		utils.HandleError(c, pkgerrors.NewAppErrorWithCause(pkgerrors.CodeInternalError, "生成令牌失败", 500, err))
		return
	}

	// 返回令牌和用户信息
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": gin.H{
			"token": tokenString,
			"user": gin.H{
				"id":        user.ID.Hex(),
				"username":  user.Username,
				"email":     user.Email,
				"roles":     user.Roles,
				"created":   user.Created,
				"lastLogin": user.LastLogin,
			},
		},
	})
}

// GetUserInfo 获取用户信息
func GetUserInfo(c *gin.Context) {
	// 从中间件获取用户信息
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户信息未找到",
		})
		return
	}

	roles, exists := c.Get("roles")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户角色信息未找到",
		})
		return
	}

	// 返回用户信息
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"username": username,
			"roles":    roles,
		},
	})
}

// Logout 处理登出请求
func Logout(c *gin.Context) {
	// 实际上，服务端不需要做任何特殊处理
	// 客户端只需要删除本地存储的令牌即可
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登出成功",
	})
}

// AuthMiddleware JWT认证中间件，支持可选角色权限校验
func AuthMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未提供授权令牌",
			})
			c.Abort()
			return
		}

		// 提取Bearer token
		tokenString := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		// 解析令牌
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return JWTSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的授权令牌",
			})
			c.Abort()
			return
		}

		// 角色权限校验（如有指定）
		if len(allowedRoles) > 0 {
			userRoles := make(map[string]struct{})
			for _, r := range claims.Roles {
				userRoles[r] = struct{}{}
			}
			allowed := false
			for _, ar := range allowedRoles {
				if _, ok := userRoles[ar]; ok {
					allowed = true
					break
				}
			}
			if !allowed {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "权限不足，禁止访问",
				})
				c.Abort()
				return
			}
		}

		// 将用户信息存储到上下文中
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}

// Register 注册新用户
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Register参数绑定失败", map[string]interface{}{"error": err})
		utils.HandleError(c, pkgerrors.NewAppErrorWithCause(pkgerrors.CodeBadRequest, "无效的请求参数", 400, err))
		return
	}

	// 建议：复用handler中的DB连接
	db := h.DB
	if db == nil {
		logger.Error("Register数据库未初始化", nil)
		utils.HandleError(c, pkgerrors.NewAppError(pkgerrors.CodeInternalError, "数据库连接未初始化", 500))
		return
	}

	// 创建用户，默认角色为user
	user, err := models.CreateUser(db, req.Username, req.Email, req.Password, []string{"user"})
	if err != nil {
		logger.Error("Register用户创建失败", map[string]interface{}{"username": req.Username, "error": err})
		utils.HandleError(c, err)
		return
	}

	// 注册成功后自动生成JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: user.Username,
		Roles:    user.Roles,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "StellarServer",
			Subject:   user.Username,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		logger.Error("Register生成JWT失败", map[string]interface{}{"error": err})
		utils.HandleError(c, pkgerrors.NewAppErrorWithCause(pkgerrors.CodeInternalError, "生成令牌失败", 500, err))
		return
	}

	// 返回结构与前端register/login一致
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "注册成功",
		"data": gin.H{
			"token": tokenString,
			"user": gin.H{
				"id":        user.ID.Hex(),
				"username":  user.Username,
				"email":     user.Email,
				"roles":     user.Roles,
				"created":   user.Created,
				"lastLogin": user.LastLogin,
			},
		},
	})
}
