package models

import "errors"

// 定义错误常量
var (
	ErrNameExists       = errors.New("名称已存在")
	ErrNotFound         = errors.New("记录不存在")
	ErrInvalidID        = errors.New("无效的ID")
	ErrInvalidParameter = errors.New("无效的参数")
	ErrDatabaseError    = errors.New("数据库错误")
	ErrUnauthorized     = errors.New("未授权的操作")
	ErrForbidden        = errors.New("禁止的操作")
)

// ErrorResponse 定义API错误响应结构
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
