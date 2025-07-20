#!/bin/bash

# 认证状态持久化测试脚本

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
FRONTEND_URL="http://localhost:5173"

# 测试用户信息
TEST_USERNAME="persistence_test_user"
TEST_EMAIL="persistence_test@example.com"
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
check_services() {
    log_info "检查服务状态..."
    
    # 检查后端服务
    if curl -s "$BASE_URL/health" > /dev/null; then
        log_success "后端服务运行正常"
    else
        log_error "后端服务未运行，请先启动服务"
        exit 1
    fi
    
    # 检查前端服务
    if curl -s "$FRONTEND_URL" > /dev/null; then
        log_success "前端服务运行正常"
    else
        log_warning "前端服务未运行，请先启动前端开发服务器"
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
        return 0
    else
        log_error "会话验证失败: $(echo "$response" | jq -r '.message')"
        return 1
    fi
}

# 测试前端页面访问
test_frontend_access() {
    log_info "测试前端页面访问..."
    
    # 测试访问主页面
    local response=$(curl -s -I "$FRONTEND_URL" | head -n 1)
    if echo "$response" | grep -q "200"; then
        log_success "前端主页面可访问"
    else
        log_warning "前端主页面访问失败: $response"
    fi
    
    # 测试访问测试页面
    local test_response=$(curl -s -I "$FRONTEND_URL/test-auth" | head -n 1)
    if echo "$test_response" | grep -q "200"; then
        log_success "认证测试页面可访问"
    else
        log_warning "认证测试页面访问失败: $test_response"
    fi
}

# 测试localStorage持久化（通过浏览器自动化）
test_localstorage_persistence() {
    log_info "测试localStorage持久化..."
    
    # 这里可以添加浏览器自动化测试
    # 例如使用Selenium或Playwright
    log_warning "localStorage持久化测试需要浏览器自动化，请手动测试"
    log_info "手动测试步骤："
    log_info "1. 访问 $FRONTEND_URL/login"
    log_info "2. 使用测试账户登录: $TEST_USERNAME / $TEST_PASSWORD"
    log_info "3. 登录成功后访问 $FRONTEND_URL/test-auth"
    log_info "4. 刷新页面，检查认证状态是否保持"
    log_info "5. 关闭浏览器，重新打开访问 $FRONTEND_URL"
    log_info "6. 检查是否自动跳转到dashboard"
}

# 测试会话状态API
test_session_apis() {
    log_info "测试会话状态API..."
    
    # 测试获取会话状态
    local status_response=$(curl -s -X GET "$API_BASE/auth/session/status" \
        -H "Authorization: Bearer $TOKEN")
    
    local status_code=$(echo "$status_response" | jq -r '.code')
    if [ "$status_code" = "200" ]; then
        log_success "获取会话状态成功"
        local username=$(echo "$status_response" | jq -r '.data.username')
        local time_until_expiry=$(echo "$status_response" | jq -r '.data.time_until_expiry')
        log_info "会话信息: 用户=$username, 剩余时间=${time_until_expiry}秒"
    else
        log_error "获取会话状态失败: $(echo "$status_response" | jq -r '.message')"
    fi
    
    # 测试刷新会话
    local refresh_response=$(curl -s -X POST "$API_BASE/auth/session/refresh" \
        -H "Authorization: Bearer $TOKEN")
    
    local refresh_code=$(echo "$refresh_response" | jq -r '.code')
    if [ "$refresh_code" = "200" ]; then
        log_success "会话刷新成功"
    else
        log_warning "会话刷新失败或Redis未配置: $(echo "$refresh_response" | jq -r '.message')"
    fi
}

# 清理测试用户
cleanup_test_user() {
    log_info "清理测试用户..."
    
    if [ -n "$USER_ID" ]; then
        log_info "测试用户ID: $USER_ID"
        # 这里可以添加删除用户的API调用
    fi
}

# 主测试流程
main() {
    log_info "开始认证状态持久化测试..."
    
    # 检查服务状态
    check_services
    
    # 注册/登录测试用户
    register_test_user
    
    # 测试会话验证
    test_session_verification
    
    # 测试会话状态API
    test_session_apis
    
    # 测试前端页面访问
    test_frontend_access
    
    # 测试localStorage持久化
    test_localstorage_persistence
    
    log_success "认证状态持久化测试完成！"
    log_info "请手动测试前端页面的认证状态持久化功能"
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