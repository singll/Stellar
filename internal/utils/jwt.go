package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JWTConfig struct {
	Secret      string
	ExpireHours int
}

type CustomClaims struct {
	UserID string   `json:"userId"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

var jwtConfig JWTConfig

func InitJWTConfig(secret string, expireHours int) {
	jwtConfig = JWTConfig{
		Secret:      secret,
		ExpireHours: expireHours,
	}
}

func GenerateJWT(userID primitive.ObjectID, roles []string) (string, error) {
	claims := CustomClaims{
		UserID: userID.Hex(),
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(jwtConfig.ExpireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(jwtConfig.Secret))
	if err != nil {
		return "", AuthError("JWT_SIGN_ERROR", "生成JWT失败").WithError(err)
	}
	return signed, nil
}

// ParseJWT 解析并校验JWT，返回自定义claims，错误时返回统一AppError
func ParseJWT(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtConfig.Secret), nil
	})
	if err != nil {
		return nil, AuthError("JWT_PARSE_ERROR", "JWT解析失败").WithError(err)
	}
	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, AuthError("JWT_INVALID", "无效的JWT Token")
	}
	return claims, nil
}
