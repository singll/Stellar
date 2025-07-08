package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

// ErrorType 定义错误类型
type ErrorType string

const (
	// ErrorTypeValidation 验证错误
	ErrorTypeValidation ErrorType = "VALIDATION_ERROR"
	// ErrorTypeDatabase 数据库错误
	ErrorTypeDatabase ErrorType = "DATABASE_ERROR"
	// ErrorTypeAuth 认证错误
	ErrorTypeAuth ErrorType = "AUTH_ERROR"
	// ErrorTypeForbidden 权限错误
	ErrorTypeForbidden ErrorType = "FORBIDDEN_ERROR"
	// ErrorTypeNotFound 资源不存在错误
	ErrorTypeNotFound ErrorType = "NOT_FOUND_ERROR"
	// ErrorTypeInternal 内部服务器错误
	ErrorTypeInternal ErrorType = "INTERNAL_ERROR"
	// ErrorTypeExternal 外部服务错误
	ErrorTypeExternal ErrorType = "EXTERNAL_ERROR"
	// ErrorTypeConflict 资源冲突错误
	ErrorTypeConflict ErrorType = "CONFLICT_ERROR"
	// ErrorTypeBadRequest 请求错误
	ErrorTypeBadRequest ErrorType = "BAD_REQUEST_ERROR"
	// ErrorTypeRateLimit 速率限制错误
	ErrorTypeRateLimit ErrorType = "RATE_LIMIT_ERROR"
)

// AppError 定义应用错误
type AppError struct {
	Type      ErrorType `json:"type"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Details   any       `json:"details,omitempty"`
	HTTPCode  int       `json:"-"`
	StackInfo string    `json:"-"`
	Err       error     `json:"-"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Type, e.Code, e.Message)
}

// WithDetails 添加错误详情
func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}

// WithError 添加原始错误
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// MarshalJSON 自定义JSON序列化
func (e *AppError) MarshalJSON() ([]byte, error) {
	type Alias AppError
	return json.Marshal(&struct {
		*Alias
		HTTPCode  int    `json:"-"`
		StackInfo string `json:"-"`
		Err       error  `json:"-"`
	}{
		Alias: (*Alias)(e),
	})
}

// NewAppError 创建新的应用错误
func NewAppError(errType ErrorType, code string, message string) *AppError {
	httpCode := getHTTPCodeForErrorType(errType)
	stackInfo := getStackInfo(2)

	return &AppError{
		Type:      errType,
		Code:      code,
		Message:   message,
		HTTPCode:  httpCode,
		StackInfo: stackInfo,
	}
}

// ValidationError 创建验证错误
func ValidationError(code string, message string) *AppError {
	return NewAppError(ErrorTypeValidation, code, message)
}

// DatabaseError 创建数据库错误
func DatabaseError(code string, message string) *AppError {
	return NewAppError(ErrorTypeDatabase, code, message)
}

// AuthError 创建认证错误
func AuthError(code string, message string) *AppError {
	return NewAppError(ErrorTypeAuth, code, message)
}

// ForbiddenError 创建权限错误
func ForbiddenError(code string, message string) *AppError {
	return NewAppError(ErrorTypeForbidden, code, message)
}

// NotFoundError 创建资源不存在错误
func NotFoundError(code string, message string) *AppError {
	return NewAppError(ErrorTypeNotFound, code, message)
}

// InternalError 创建内部服务器错误
func InternalError(code string, message string) *AppError {
	return NewAppError(ErrorTypeInternal, code, message)
}

// ExternalError 创建外部服务错误
func ExternalError(code string, message string) *AppError {
	return NewAppError(ErrorTypeExternal, code, message)
}

// ConflictError 创建资源冲突错误
func ConflictError(code string, message string) *AppError {
	return NewAppError(ErrorTypeConflict, code, message)
}

// BadRequestError 创建请求错误
func BadRequestError(code string, message string) *AppError {
	return NewAppError(ErrorTypeBadRequest, code, message)
}

// RateLimitError 创建速率限制错误
func RateLimitError(code string, message string) *AppError {
	return NewAppError(ErrorTypeRateLimit, code, message)
}

// getHTTPCodeForErrorType 获取错误类型对应的HTTP状态码
func getHTTPCodeForErrorType(errType ErrorType) int {
	switch errType {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeDatabase:
		return http.StatusInternalServerError
	case ErrorTypeAuth:
		return http.StatusUnauthorized
	case ErrorTypeForbidden:
		return http.StatusForbidden
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeInternal:
		return http.StatusInternalServerError
	case ErrorTypeExternal:
		return http.StatusBadGateway
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeBadRequest:
		return http.StatusBadRequest
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// getStackInfo 获取堆栈信息
func getStackInfo(skip int) string {
	var stackBuilder strings.Builder
	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		name := fn.Name()
		if strings.Contains(name, "runtime.") {
			continue
		}
		stackBuilder.WriteString(fmt.Sprintf("%s:%d %s\n", file, line, name))
	}
	return stackBuilder.String()
}

// HandleError 处理错误并返回适当的HTTP响应
func HandleError(c *gin.Context, err error) {
	// 记录错误
	if appErr, ok := err.(*AppError); ok {
		if appErr.Err != nil {
			Error("应用错误", appErr.Err, "type", string(appErr.Type), "code", appErr.Code, "message", appErr.Message)
		} else {
			Error("应用错误", nil, "type", string(appErr.Type), "code", appErr.Code, "message", appErr.Message)
		}

		c.JSON(appErr.HTTPCode, gin.H{
			"error": appErr,
		})
		return
	}

	// 处理未知错误
	Error("未知错误", err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"type":    ErrorTypeInternal,
			"code":    "UNKNOWN_ERROR",
			"message": "发生未知错误",
		},
	})
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error struct {
		Type    ErrorType `json:"type"`
		Code    string    `json:"code"`
		Message string    `json:"message"`
		Details any       `json:"details,omitempty"`
	} `json:"error"`
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(errType ErrorType, code string, message string, details any) *ErrorResponse {
	resp := &ErrorResponse{}
	resp.Error.Type = errType
	resp.Error.Code = code
	resp.Error.Message = message
	resp.Error.Details = details
	return resp
}

// ErrorMiddleware 全局错误处理中间件
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			HandleError(c, err)
			c.Abort()
			return
		}
	}
}

// PanicRecoveryMiddleware 全局恐慌恢复中间件
func PanicRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch x := r.(type) {
				case string:
					err = fmt.Errorf(x)
				case error:
					err = x
				default:
					err = fmt.Errorf("未知恐慌: %v", r)
				}

				stack := FormatStackTrace(3)
				Error("恐慌恢复", err, "stack", stack)

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"type":    ErrorTypeInternal,
						"code":    "PANIC_RECOVERY",
						"message": "服务器内部错误",
					},
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
