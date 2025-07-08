package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/api/v1/auth/login", Login)
	r.POST("/api/v1/auth/logout", Logout)
	r.GET("/api/v1/auth/info", GetUserInfo)
	return r
}

func TestLogin_Success(t *testing.T) {
	t.Run("login success", func(t *testing.T) {
		// TODO: mock models.ValidateUser 返回 user
		// 构造请求
		body := map[string]string{"username": "testuser", "password": "password"}
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
	// ... existing code ...
	// 缺少参数
	body := map[string]string{"username": ""}
	jsonBody, _ := json.Marshal(body)
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestLogin_AuthFail(t *testing.T) {
	// ... existing code ...
	// 用户名或密码错误
	body := map[string]string{"username": "notfound", "password": "wrong"}
	jsonBody, _ := json.Marshal(body)
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestLogout(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestGetUserInfo_Unauthorized(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/auth/info", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestGetUserInfo_Success(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	// 构造有效token
	token := "valid-token" // TODO: mock有效token生成
	req, _ := http.NewRequest("GET", "/api/v1/auth/info", nil)
	req.Header.Set("Authorization", token)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestGetUserInfo_InvalidToken(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/auth/info", nil)
	req.Header.Set("Authorization", "invalid-token")
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestProtectedRoute_Forbidden(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	// 构造仅user角色的token，访问需要admin的接口
	token := "user-role-token" // TODO: mock user角色token
	req, _ := http.NewRequest("POST", "/api/v1/assets", nil)
	req.Header.Set("Authorization", token)
	r.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)
}

// TODO: 补充GetUserInfo成功场景、token无效等分支
