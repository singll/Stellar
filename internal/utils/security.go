package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// 常量定义
const (
	// 默认的加密密钥长度
	DefaultKeyLength = 32
	// 默认的盐长度
	DefaultSaltLength = 16
	// 默认的BCrypt成本
	DefaultBcryptCost = 12
	// 默认的JWT过期时间（小时）
	DefaultJWTExpiry = 24
)

// 错误定义
var (
	ErrInvalidKey       = errors.New("无效的加密密钥")
	ErrInvalidData      = errors.New("无效的数据")
	ErrInvalidSignature = errors.New("无效的签名")
	ErrTokenExpired     = errors.New("令牌已过期")
	ErrInvalidToken     = errors.New("无效的令牌")
)

// JWTClaims 定义JWT声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateRandomBytes 生成随机字节
func GenerateRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) (string, error) {
	b, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}

// HashPassword 使用BCrypt哈希密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultBcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// SHA256 计算SHA256哈希
func SHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SHA512 计算SHA512哈希
func SHA512(data string) string {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:])
}

// EncryptAES 使用AES加密数据
func EncryptAES(data, key []byte) ([]byte, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, ErrInvalidKey
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 创建随机数
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// DecryptAES 使用AES解密数据
func DecryptAES(data, key []byte) ([]byte, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, ErrInvalidKey
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 检查数据长度
	if len(data) < gcm.NonceSize() {
		return nil, ErrInvalidData
	}

	// 提取随机数
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// ValidateJWT 验证JWT令牌
func ValidateJWT(tokenString, secret string) (*JWTClaims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrTokenExpired
			}
		}
		return nil, ErrInvalidToken
	}

	// 获取声明
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// SanitizeInput 清理输入，防止XSS和SQL注入
func SanitizeInput(input string) string {
	// 移除HTML标签
	re := regexp.MustCompile("<[^>]*>")
	input = re.ReplaceAllString(input, "")

	// 移除SQL注入关键字
	input = strings.ReplaceAll(input, "'", "''")
	input = strings.ReplaceAll(input, "\"", "\"\"")
	input = strings.ReplaceAll(input, ";", "")
	input = strings.ReplaceAll(input, "--", "")
	input = strings.ReplaceAll(input, "/*", "")
	input = strings.ReplaceAll(input, "*/", "")

	return input
}

// IsStrongPassword 检查密码强度
func IsStrongPassword(password string) bool {
	// 密码长度至少8位
	if len(password) < 8 {
		return false
	}

	// 包含至少一个数字
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasNumber {
		return false
	}

	// 包含至少一个小写字母
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return false
	}

	// 包含至少一个大写字母
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return false
	}

	// 包含至少一个特殊字符
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
	if !hasSpecial {
		return false
	}

	return true
}

// MaskSensitiveData 掩盖敏感数据
func MaskSensitiveData(data, dataType string) string {
	switch dataType {
	case "email":
		parts := strings.Split(data, "@")
		if len(parts) != 2 {
			return data
		}
		username := parts[0]
		domain := parts[1]
		if len(username) <= 2 {
			return data
		}
		maskedUsername := username[:2] + strings.Repeat("*", len(username)-2)
		return maskedUsername + "@" + domain
	case "phone":
		if len(data) <= 4 {
			return data
		}
		return strings.Repeat("*", len(data)-4) + data[len(data)-4:]
	case "creditcard":
		if len(data) <= 4 {
			return data
		}
		return strings.Repeat("*", len(data)-4) + data[len(data)-4:]
	case "password":
		return strings.Repeat("*", len(data))
	default:
		return data
	}
}

// GenerateCSRFToken 生成CSRF令牌
func GenerateCSRFToken() (string, error) {
	token, err := GenerateRandomBytes(32)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(token), nil
}

// ValidateCSRFToken 验证CSRF令牌
func ValidateCSRFToken(token, storedToken string) bool {
	return token == storedToken
}
