package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	// 初始化测试用JWT配置
	InitJWTConfig("test-secret", 1) // 1小时
}

func TestGenerateAndParseJWT_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	roles := []string{"admin", "user"}

	token, err := GenerateJWT(userID, roles)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ParseJWT(token)
	assert.NoError(t, err)
	assert.Equal(t, userID.Hex(), claims.UserID)
	assert.ElementsMatch(t, roles, claims.Roles)
	assert.WithinDuration(t, time.Now(), claims.IssuedAt.Time, time.Minute)
	assert.WithinDuration(t, time.Now().Add(1*time.Hour), claims.ExpiresAt.Time, time.Minute)
}

func TestParseJWT_InvalidToken(t *testing.T) {
	_, err := ParseJWT("invalid.token.here")
	assert.Error(t, err)
	appErr, ok := err.(*AppError)
	assert.True(t, ok)
	assert.Equal(t, ErrorTypeAuth, appErr.Type)
	assert.Equal(t, "JWT_PARSE_ERROR", appErr.Code)
}

func TestParseJWT_ExpiredToken(t *testing.T) {
	claims := CustomClaims{
		UserID: primitive.NewObjectID().Hex(),
		Roles:  []string{"user"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // 已过期
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("test-secret"))
	assert.NoError(t, err)

	_, err = ParseJWT(tokenStr)
	assert.Error(t, err)
	appErr, ok := err.(*AppError)
	assert.True(t, ok)
	// 过期token会被ParseWithClaims判为无效
	assert.Equal(t, ErrorTypeAuth, appErr.Type)
	assert.Equal(t, "JWT_INVALID", appErr.Code)
}

func TestGenerateJWT_SignError(t *testing.T) {
	// 临时破坏jwtConfig.Secret
	oldSecret := jwtConfig.Secret
	jwtConfig.Secret = ""
	defer func() { jwtConfig.Secret = oldSecret }()

	userID := primitive.NewObjectID()
	_, err := GenerateJWT(userID, []string{"user"})
	assert.Error(t, err)
	appErr, ok := err.(*AppError)
	assert.True(t, ok)
	assert.Equal(t, ErrorTypeAuth, appErr.Type)
	assert.Equal(t, "JWT_SIGN_ERROR", appErr.Code)
}
