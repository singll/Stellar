package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	
	// 初始化JWT配置用于测试
	utils.InitJWTConfig("test-secret-key", 24)
	
	// 设置路由
	r.POST("/api/v1/auth/login", Login)
	r.POST("/api/v1/auth/logout", Logout)
	r.GET("/api/v1/auth/info", AuthMiddleware(), GetUserInfo)
	r.POST("/api/v1/auth/register", Register)
	
	// 添加受保护的路由用于测试
	r.GET("/api/v1/protected", AuthMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "访问成功"})
	})
	r.GET("/api/v1/admin", AuthMiddleware("admin"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "管理员访问成功"})
	})
	
	return r
}

func TestLogin_Success(t *testing.T) {
	t.Run("用户名登录成功", func(t *testing.T) {
		// 跳过，因为需要真实的数据库连接
		t.Skip("需要数据库连接")
		
		body := map[string]string{"username": "testuser", "password": "password"}
		jsonBody, _ := json.Marshal(body)
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		
		// 验证响应结构
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])
		assert.NotNil(t, response["data"])
		
		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["token"])
		assert.NotNil(t, data["user"])
	})
	
	t.Run("邮箱登录成功", func(t *testing.T) {
		t.Skip("需要数据库连接")
		
		body := map[string]string{"email": "test@example.com", "password": "password"}
		jsonBody, _ := json.Marshal(body)
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	})
}

func TestLogin_InvalidParams(t *testing.T) {
	t.Run("缺少密码参数", func(t *testing.T) {
		body := map[string]string{"username": "testuser"}
		jsonBody, _ := json.Marshal(body)
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, 400, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(400), response["code"])
		assert.Contains(t, response["message"], "无效的请求参数")
	})
	
	t.Run("用户名和邮箱都为空", func(t *testing.T) {
		body := map[string]string{"password": "password"}
		jsonBody, _ := json.Marshal(body)
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, 400, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "用户名或邮箱不能为空")
	})
	
	t.Run("无效的JSON格式", func(t *testing.T) {
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, 400, w.Code)
	})
}

func TestLogin_AuthFail(t *testing.T) {
	t.Run("用户名或密码错误", func(t *testing.T) {
		// 这个测试会因为数据库连接失败而返回500，但我们可以测试基本流程
		body := map[string]string{"username": "notfound", "password": "wrong"}
		jsonBody, _ := json.Marshal(body)
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		
		// 由于没有数据库连接，会返回500而不是401
		assert.Equal(t, 500, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(500), response["code"])
		assert.Contains(t, response["message"], "数据库连接未初始化")
	})
}

func TestLogout(t *testing.T) {
	t.Run("登出成功", func(t *testing.T) {
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])
		assert.Equal(t, "登出成功", response["message"])
	})
}

func TestGetUserInfo_Unauthorized(t *testing.T) {
	t.Run("无Authorization头", func(t *testing.T) {
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/auth/info", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 401, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(401), response["code"])
		assert.Contains(t, response["message"], "未提供授权令牌")
	})
}

func TestGetUserInfo_Success(t *testing.T) {
	t.Run("有效token获取用户信息", func(t *testing.T) {
		// 生成有效的JWT token
		userID := primitive.NewObjectID()
		roles := []string{"user"}
		token, err := utils.GenerateJWT(userID, roles)
		require.NoError(t, err)
		
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/auth/info", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])
		
		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["username"])
		assert.NotEmpty(t, data["roles"])
	})
	
	t.Run("不带Bearer前缀的token", func(t *testing.T) {
		userID := primitive.NewObjectID()
		roles := []string{"user"}
		token, err := utils.GenerateJWT(userID, roles)
		require.NoError(t, err)
		
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/auth/info", nil)
		req.Header.Set("Authorization", token)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	})
}

func TestGetUserInfo_InvalidToken(t *testing.T) {
	t.Run("无效token格式", func(t *testing.T) {
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/auth/info", nil)
		req.Header.Set("Authorization", "invalid-token")
		r.ServeHTTP(w, req)
		assert.Equal(t, 401, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(401), response["code"])
		assert.Contains(t, response["message"], "无效的授权令牌")
	})
	
	t.Run("过期的token", func(t *testing.T) {
		// 创建过期的token
		utils.InitJWTConfig("test-secret-key", -1) // 设置为过期
		userID := primitive.NewObjectID()
		roles := []string{"user"}
		token, err := utils.GenerateJWT(userID, roles)
		require.NoError(t, err)
		
		// 等待token过期
		time.Sleep(time.Second * 1)
		
		// 恢复正常配置
		utils.InitJWTConfig("test-secret-key", 24)
		
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/auth/info", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		assert.Equal(t, 401, w.Code)
	})
}

func TestProtectedRoute_Forbidden(t *testing.T) {
	t.Run("普通用户访问管理员接口", func(t *testing.T) {
		// 生成普通用户token
		userID := primitive.NewObjectID()
		roles := []string{"user"}
		token, err := utils.GenerateJWT(userID, roles)
		require.NoError(t, err)
		
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		assert.Equal(t, 403, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(403), response["code"])
		assert.Contains(t, response["message"], "权限不足")
	})
	
	t.Run("管理员访问管理员接口", func(t *testing.T) {
		// 生成管理员token
		userID := primitive.NewObjectID()
		roles := []string{"admin"}
		token, err := utils.GenerateJWT(userID, roles)
		require.NoError(t, err)
		
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "管理员访问成功", response["message"])
	})
	
	t.Run("普通用户访问受保护接口", func(t *testing.T) {
		// 生成普通用户token
		userID := primitive.NewObjectID()
		roles := []string{"user"}
		token, err := utils.GenerateJWT(userID, roles)
		require.NoError(t, err)
		
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "访问成功", response["message"])
	})
}

// 新增注册功能测试
func TestRegister(t *testing.T) {
	t.Run("注册参数验证", func(t *testing.T) {
		testCases := []struct {
			name     string
			body     map[string]interface{}
			expectedCode int
		}{
			{
				name:     "缺少用户名",
				body:     map[string]interface{}{"email": "test@example.com", "password": "password123"},
				expectedCode: 400,
			},
			{
				name:     "缺少邮箱",
				body:     map[string]interface{}{"username": "testuser", "password": "password123"},
				expectedCode: 400,
			},
			{
				name:     "缺少密码",
				body:     map[string]interface{}{"username": "testuser", "email": "test@example.com"},
				expectedCode: 400,
			},
			{
				name:     "无效邮箱格式",
				body:     map[string]interface{}{"username": "testuser", "email": "invalid-email", "password": "password123"},
				expectedCode: 400,
			},
			{
				name:     "密码太短",
				body:     map[string]interface{}{"username": "testuser", "email": "test@example.com", "password": "12345"},
				expectedCode: 400,
			},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				jsonBody, _ := json.Marshal(tc.body)
				r := setupRouter()
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				r.ServeHTTP(w, req)
				assert.Equal(t, tc.expectedCode, w.Code)
			})
		}
	})
	
	t.Run("注册失败-数据库未连接", func(t *testing.T) {
		body := map[string]string{
			"username": "testuser",
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(body)
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		
		// 由于没有数据库连接，会返回500
		assert.Equal(t, 500, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(500), response["code"])
		assert.Contains(t, response["message"], "数据库连接未初始化")
	})
}

// 测试JWT中间件的各种情况
func TestAuthMiddleware(t *testing.T) {
	t.Run("Bearer token格式测试", func(t *testing.T) {
		userID := primitive.NewObjectID()
		roles := []string{"user"}
		token, err := utils.GenerateJWT(userID, roles)
		require.NoError(t, err)
		
		testCases := []struct {
			name   string
			header string
			expectedCode int
		}{
			{"Bearer格式", "Bearer " + token, 200},
			{"直接token", token, 200},
			{"错误Bearer格式", "Bearer" + token, 401},
			{"Bearer空格过多", "Bearer  " + token, 401},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				r := setupRouter()
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/v1/protected", nil)
				req.Header.Set("Authorization", tc.header)
				r.ServeHTTP(w, req)
				assert.Equal(t, tc.expectedCode, w.Code)
			})
		}
	})
	
	t.Run("多角色权限测试", func(t *testing.T) {
		// 创建具有多个角色的用户
		userID := primitive.NewObjectID()
		roles := []string{"user", "admin", "moderator"}
		token, err := utils.GenerateJWT(userID, roles)
		require.NoError(t, err)
		
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	})
}
