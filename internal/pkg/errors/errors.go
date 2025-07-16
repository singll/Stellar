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

	// 业务错误代码
	CodeUserNotFound      ErrorCode = "USER_NOT_FOUND"
	CodeUserExists        ErrorCode = "USER_EXISTS"
	CodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	CodeProjectNotFound   ErrorCode = "PROJECT_NOT_FOUND"
	CodeAssetNotFound     ErrorCode = "ASSET_NOT_FOUND"
	CodeTaskNotFound      ErrorCode = "TASK_NOT_FOUND"
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

// 业务错误构造函数
func NewUserNotFoundError() *AppError {
	return NewAppError(CodeUserNotFound, "User not found", http.StatusNotFound)
}

func NewUserExistsError() *AppError {
	return NewAppError(CodeUserExists, "User already exists", http.StatusConflict)
}

func NewInvalidCredentialsError() *AppError {
	return NewAppError(CodeInvalidCredentials, "Invalid credentials", http.StatusUnauthorized)
}

func NewProjectNotFoundError() *AppError {
	return NewAppError(CodeProjectNotFound, "Project not found", http.StatusNotFound)
}

func NewAssetNotFoundError() *AppError {
	return NewAppError(CodeAssetNotFound, "Asset not found", http.StatusNotFound)
}

func NewTaskNotFoundError() *AppError {
	return NewAppError(CodeTaskNotFound, "Task not found", http.StatusNotFound)
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