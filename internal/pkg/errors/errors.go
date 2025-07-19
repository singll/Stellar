package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode 错误代码类型
type ErrorCode string

const (
	// 通用错误代码
	CodeInternalError    ErrorCode = "INTERNAL_ERROR"
	CodeBadRequest       ErrorCode = "BAD_REQUEST"
	CodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	CodeForbidden        ErrorCode = "FORBIDDEN"
	CodeNotFound         ErrorCode = "NOT_FOUND"
	CodeConflict         ErrorCode = "CONFLICT"
	CodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	CodeTimeout          ErrorCode = "TIMEOUT"
	CodeRateLimit        ErrorCode = "RATE_LIMIT"

	// 业务错误代码
	CodeUserNotFound       ErrorCode = "USER_NOT_FOUND"
	CodeUserExists         ErrorCode = "USER_EXISTS"
	CodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	CodeProjectNotFound    ErrorCode = "PROJECT_NOT_FOUND"
	CodeAssetNotFound      ErrorCode = "ASSET_NOT_FOUND"
	CodeTaskNotFound       ErrorCode = "TASK_NOT_FOUND"
	CodeTaskRunning        ErrorCode = "TASK_RUNNING"
	CodeTaskCompleted      ErrorCode = "TASK_COMPLETED"
	CodeTaskFailed         ErrorCode = "TASK_FAILED"
	CodeTaskTimeout        ErrorCode = "TASK_TIMEOUT"
	CodeNodeNotFound       ErrorCode = "NODE_NOT_FOUND"
	CodeNodeOffline        ErrorCode = "NODE_OFFLINE"
	CodePluginNotFound     ErrorCode = "PLUGIN_NOT_FOUND"
	CodePluginError        ErrorCode = "PLUGIN_ERROR"
	CodeScanError          ErrorCode = "SCAN_ERROR"
	CodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	CodeRedisError         ErrorCode = "REDIS_ERROR"
	CodeNetworkError       ErrorCode = "NETWORK_ERROR"
	CodeFileError          ErrorCode = "FILE_ERROR"
	CodeConfigError        ErrorCode = "CONFIG_ERROR"
)

// AppError 应用程序错误
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
	Cause      error     `json:"-"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap 支持错误链
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError 创建新的应用程序错误
func NewAppError(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// NewAppErrorWithCause 创建带原因的应用程序错误
func NewAppErrorWithCause(code ErrorCode, message string, httpStatus int, cause error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Cause:      cause,
	}
}

// 预定义错误构造函数
func NewInternalError(message string) *AppError {
	return NewAppError(CodeInternalError, message, http.StatusInternalServerError)
}

func NewBadRequestError(message string) *AppError {
	return NewAppError(CodeBadRequest, message, http.StatusBadRequest)
}

func NewUnauthorizedError(message string) *AppError {
	return NewAppError(CodeUnauthorized, message, http.StatusUnauthorized)
}

func NewForbiddenError(message string) *AppError {
	return NewAppError(CodeForbidden, message, http.StatusForbidden)
}

func NewNotFoundError(message string) *AppError {
	return NewAppError(CodeNotFound, message, http.StatusNotFound)
}

func NewConflictError(message string) *AppError {
	return NewAppError(CodeConflict, message, http.StatusConflict)
}

func NewValidationError(message string) *AppError {
	return NewAppError(CodeValidationFailed, message, http.StatusBadRequest)
}

func NewTimeoutError(message string) *AppError {
	return NewAppError(CodeTimeout, message, http.StatusRequestTimeout)
}

func NewRateLimitError(message string) *AppError {
	return NewAppError(CodeRateLimit, message, http.StatusTooManyRequests)
}

// 业务错误构造函数
func NewUserNotFoundError() *AppError {
	return NewAppError(CodeUserNotFound, "用户不存在", http.StatusNotFound)
}

func NewUserExistsError() *AppError {
	return NewAppError(CodeUserExists, "用户已存在", http.StatusConflict)
}

func NewInvalidCredentialsError() *AppError {
	return NewAppError(CodeInvalidCredentials, "用户名或密码错误", http.StatusUnauthorized)
}

func NewProjectNotFoundError() *AppError {
	return NewAppError(CodeProjectNotFound, "项目不存在", http.StatusNotFound)
}

func NewAssetNotFoundError() *AppError {
	return NewAppError(CodeAssetNotFound, "资产不存在", http.StatusNotFound)
}

func NewTaskNotFoundError() *AppError {
	return NewAppError(CodeTaskNotFound, "任务不存在", http.StatusNotFound)
}

func NewTaskRunningError() *AppError {
	return NewAppError(CodeTaskRunning, "任务正在运行中", http.StatusConflict)
}

func NewTaskCompletedError() *AppError {
	return NewAppError(CodeTaskCompleted, "任务已完成", http.StatusConflict)
}

func NewTaskFailedError() *AppError {
	return NewAppError(CodeTaskFailed, "任务执行失败", http.StatusInternalServerError)
}

func NewTaskTimeoutError() *AppError {
	return NewAppError(CodeTaskTimeout, "任务执行超时", http.StatusRequestTimeout)
}

func NewNodeNotFoundError() *AppError {
	return NewAppError(CodeNodeNotFound, "节点不存在", http.StatusNotFound)
}

func NewNodeOfflineError() *AppError {
	return NewAppError(CodeNodeOffline, "节点离线", http.StatusServiceUnavailable)
}

func NewPluginNotFoundError(pluginID string) *AppError {
	return NewAppError(CodePluginNotFound, fmt.Sprintf("插件不存在: %s", pluginID), http.StatusNotFound)
}

func NewPluginError(message string) *AppError {
	return NewAppError(CodePluginError, message, http.StatusInternalServerError)
}

func NewScanError(message string) *AppError {
	return NewAppError(CodeScanError, message, http.StatusInternalServerError)
}

func NewDatabaseError(message string) *AppError {
	return NewAppError(CodeDatabaseError, message, http.StatusInternalServerError)
}

func NewRedisError(message string) *AppError {
	return NewAppError(CodeRedisError, message, http.StatusInternalServerError)
}

func NewNetworkError(message string) *AppError {
	return NewAppError(CodeNetworkError, message, http.StatusBadGateway)
}

func NewFileError(message string) *AppError {
	return NewAppError(CodeFileError, message, http.StatusInternalServerError)
}

func NewConfigError(message string) *AppError {
	return NewAppError(CodeConfigError, message, http.StatusInternalServerError)
}

// IsAppError 检查是否为应用程序错误
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

// WrapError 包装标准错误为应用程序错误
func WrapError(err error, code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Cause:      err,
	}
}

// WrapDatabaseError 包装数据库错误
func WrapDatabaseError(err error, operation string) *AppError {
	return WrapError(err, CodeDatabaseError, fmt.Sprintf("数据库操作失败: %s", operation), http.StatusInternalServerError)
}

// WrapValidationError 包装验证错误
func WrapValidationError(err error, field string) *AppError {
	return WrapError(err, CodeValidationFailed, fmt.Sprintf("参数验证失败: %s", field), http.StatusBadRequest)
}

// WrapNetworkError 包装网络错误
func WrapNetworkError(err error, operation string) *AppError {
	return WrapError(err, CodeNetworkError, fmt.Sprintf("网络操作失败: %s", operation), http.StatusBadGateway)
}

// WrapFileError 包装文件操作错误
func WrapFileError(err error, operation string) *AppError {
	return WrapError(err, CodeFileError, fmt.Sprintf("文件操作失败: %s", operation), http.StatusInternalServerError)
}

// WrapTaskError 包装任务相关错误
func WrapTaskError(err error, taskID string, operation string) *AppError {
	return WrapError(err, CodeTaskNotFound, fmt.Sprintf("任务操作失败 [%s]: %s", taskID, operation), http.StatusInternalServerError)
}

// WrapPluginError 包装插件错误
func WrapPluginError(err error, pluginID string, operation string) *AppError {
	return WrapError(err, CodePluginError, fmt.Sprintf("插件操作失败 [%s]: %s", pluginID, operation), http.StatusInternalServerError)
}
