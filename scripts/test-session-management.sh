#!/bin/bash

# 会话管理功能测试脚本
# 测试Redis会话管理、状态检查、刷新和删除功能

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
BASE_URL="http://localhost:8082"
API_BASE="$BASE_URL/api/v1"

# 测试用户信息
TEST_USERNAME="session_test_user"
TEST_EMAIL="session_test@example.com"
TEST_PASSWORD="test123456"

# 全局变量
TOKEN=""
USER_ID=""

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查服务状态
check_service() {
    log_info "检查服务状态..."
    
    if curl -s "$BASE_URL/health" > /dev/null; then
        log_success "后端服务运行正常"
    else
        log_error "后端服务未运行，请先启动服务"
        exit 1
    fi
}

# 清理测试用户
cleanup_test_user() {
    log_info "清理测试用户..."
    
    if [ -n "$USER_ID" ]; then
        # 这里可以添加删除用户的API调用
        log_info "测试用户ID: $USER_ID"
    fi
}

# 注册测试用户
register_test_user() {
    log_info "注册测试用户..."
    
    local response=$(curl -s -X POST "$API_BASE/auth/register" \
        -H "Content-Type: application/json" \
        -d "{
            \"username\": \"$TEST_USERNAME\",
            \"email\": \"$TEST_EMAIL\",
            \"password\": \"$TEST_PASSWORD\"
        }")
    
    local code=$(echo "$response" | jq -r '.code')
    
    if [ "$code" = "200" ]; then
        log_success "测试用户注册成功"
        TOKEN=$(echo "$response" | jq -r '.data.token')
        USER_ID=$(echo "$response" | jq -r '.data.user.id')
    else
        log_warning "测试用户可能已存在，尝试登录..."
        login_test_user
    fi
}

# 登录测试用户
login_test_user() {
    log_info "登录测试用户..."
    
    local response=$(curl -s -X POST "$API_BASE/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"username\": \"$TEST_USERNAME\",
            \"password\": \"$TEST_PASSWORD\"
        }")
    
    local code=$(echo "$response" | jq -r '.code')
    
    if [ "$code" = "200" ]; then
        log_success "测试用户登录成功"
        TOKEN=$(echo "$response" | jq -r '.data.token')
        USER_ID=$(echo "$response" | jq -r '.data.user.id')
    else
        log_error "登录失败: $(echo "$response" | jq -r '.message')"
        exit 1
    fi
}

# 测试会话验证
test_session_verification() {
    log_info "测试会话验证..."
    
    local response=$(curl -s -X GET "$API_BASE/auth/verify" \
        -H "Authorization: Bearer $TOKEN")
    
    local code=$(echo "$response" | jq -r '.code')
    local valid=$(echo "$response" | jq -r '.valid')
    
    if [ "$code" = "200" ] && [ "$valid" = "true" ]; then
        log_success "会话验证成功"
        
        # 检查会话状态信息
        local has_session_status=$(echo "$response" | jq -r 'has("session_status")')
        if [ "$has_session_status" = "true" ]; then
            log_success "会话状态信息完整"
            local username=$(echo "$response" | jq -r '.session_status.username')
            local is_expired=$(echo "$response" | jq -r '.session_status.is_expired')
            local needs_refresh=$(echo "$response" | jq -r '.session_status.needs_refresh')
            
            log_info "会话信息: 用户=$username, 过期=$is_expired, 需刷新=$needs_refresh"
        else
            log_warning "未返回会话状态信息（可能Redis未配置）"
        fi
    else
        log_error "会话验证失败: $(echo "$response" | jq -r '.message')"
        return 1
    fi
}

# 测试获取会话状态
test_get_session_status() {
    log_info "测试获取会话状态..."
    
    local response=$(curl -s -X GET "$API_BASE/auth/session/status" \
        -H "Authorization: Bearer $TOKEN")
    
    local code=$(echo "$response" | jq -r '.code')
    
    if [ "$code" = "200" ]; then
        log_success "获取会话状态成功"
        
        local data=$(echo "$response" | jq -r '.data')
        if [ "$data" != "null" ]; then
            local username=$(echo "$response" | jq -r '.data.username')
            local time_until_expiry=$(echo "$response" | jq -r '.data.time_until_expiry')
            local is_expired=$(echo "$response" | jq -r '.data.is_expired')
            
            log_info "会话状态: 用户=$username, 剩余时间=${time_until_expiry}秒, 过期=$is_expired"
        else
            log_warning "未返回会话数据（可能Redis未配置）"
        fi
    else
        log_error "获取会话状态失败: $(echo "$response" | jq -r '.message')"
        return 1
    fi
}

# 测试会话刷新
test_session_refresh() {
    log_info "测试会话刷新..."
    
    local response=$(curl -s -X POST "$API_BASE/auth/session/refresh" \
        -H "Authorization: Bearer $TOKEN")
    
    local code=$(echo "$response" | jq -r '.code')
    
    if [ "$code" = "200" ]; then
        log_success "会话刷新成功"
        
        local data=$(echo "$response" | jq -r '.data')
        if [ "$data" != "null" ]; then
            local last_used=$(echo "$response" | jq -r '.data.last_used')
            log_info "会话刷新时间: $last_used"
        fi
    else
        log_warning "会话刷新失败或Redis未配置: $(echo "$response" | jq -r '.message')"
    fi
}

# 测试重新登录（会话替换）
test_relogin() {
    log_info "测试重新登录（会话替换）..."
    
    local response=$(curl -s -X POST "$API_BASE/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"username\": \"$TEST_USERNAME\",
            \"password\": \"$TEST_PASSWORD\"
        }")
    
    local code=$(echo "$response" | jq -r '.code')
    
    if [ "$code" = "200" ]; then
        log_success "重新登录成功"
        local new_token=$(echo "$response" | jq -r '.data.token')
        
        # 验证新token
        local verify_response=$(curl -s -X GET "$API_BASE/auth/verify" \
            -H "Authorization: Bearer $new_token")
        
        local verify_code=$(echo "$verify_response" | jq -r '.code')
        local valid=$(echo "$verify_response" | jq -r '.valid')
        
        if [ "$verify_code" = "200" ] && [ "$valid" = "true" ]; then
            log_success "新会话验证成功"
            TOKEN="$new_token"
        else
            log_error "新会话验证失败"
            return 1
        fi
    else
        log_error "重新登录失败: $(echo "$response" | jq -r '.message')"
        return 1
    fi
}

# 测试登出（会话删除）
test_logout() {
    log_info "测试登出（会话删除）..."
    
    local response=$(curl -s -X POST "$API_BASE/auth/logout" \
        -H "Authorization: Bearer $TOKEN")
    
    local code=$(echo "$response" | jq -r '.code')
    
    if [ "$code" = "200" ]; then
        log_success "登出成功"
        
        # 验证会话是否已删除
        local verify_response=$(curl -s -X GET "$API_BASE/auth/verify" \
            -H "Authorization: Bearer $TOKEN")
        
        local verify_code=$(echo "$verify_response" | jq -r '.code')
        
        if [ "$verify_code" = "401" ]; then
            log_success "会话已成功删除"
        else
            log_warning "会话可能未完全删除（JWT仍有效）"
        fi
    else
        log_error "登出失败: $(echo "$response" | jq -r '.message')"
        return 1
    fi
}

# 测试无效token
test_invalid_token() {
    log_info "测试无效token..."
    
    local response=$(curl -s -X GET "$API_BASE/auth/verify" \
        -H "Authorization: Bearer invalid_token_123")
    
    local code=$(echo "$response" | jq -r '.code')
    
    if [ "$code" = "401" ]; then
        log_success "无效token正确被拒绝"
    else
        log_error "无效token未被正确拒绝"
        return 1
    fi
}

# 主测试流程
main() {
    log_info "开始会话管理功能测试..."
    
    # 检查服务状态
    check_service
    
    # 注册/登录测试用户
    register_test_user
    
    # 测试会话验证
    test_session_verification
    
    # 测试获取会话状态
    test_get_session_status
    
    # 测试会话刷新
    test_session_refresh
    
    # 测试重新登录（会话替换）
    test_relogin
    
    # 再次验证会话
    test_session_verification
    
    # 测试登出（会话删除）
    test_logout
    
    # 测试无效token
    test_invalid_token
    
    log_success "所有会话管理功能测试完成！"
}

# 清理函数
cleanup() {
    log_info "清理测试环境..."
    cleanup_test_user
}

# 设置退出时清理
trap cleanup EXIT

# 运行主测试
main "$@" 