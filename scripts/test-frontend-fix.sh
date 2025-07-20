#!/bin/bash

# 前端修复验证测试脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
FRONTEND_URL="http://localhost:5173"

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

# 检查前端服务状态
check_frontend_service() {
    log_info "检查前端服务状态..."
    
    # 等待服务启动
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$FRONTEND_URL" > /dev/null 2>&1; then
            log_success "前端服务运行正常"
            return 0
        else
            log_info "等待前端服务启动... (尝试 $attempt/$max_attempts)"
            sleep 2
            attempt=$((attempt + 1))
        fi
    done
    
    log_error "前端服务启动超时"
    return 1
}

# 测试页面访问
test_page_access() {
    log_info "测试页面访问..."
    
    # 测试主页面
    local main_response=$(curl -s -I "$FRONTEND_URL" | head -n 1)
    if echo "$main_response" | grep -q "200"; then
        log_success "主页面可访问"
    else
        log_error "主页面访问失败: $main_response"
        return 1
    fi
    
    # 测试登录页面
    local login_response=$(curl -s -I "$FRONTEND_URL/login" | head -n 1)
    if echo "$login_response" | grep -q "200"; then
        log_success "登录页面可访问"
    else
        log_error "登录页面访问失败: $login_response"
        return 1
    fi
    
    # 测试简单测试页面
    local test_response=$(curl -s -I "$FRONTEND_URL/test-simple" | head -n 1)
    if echo "$test_response" | grep -q "200"; then
        log_success "简单测试页面可访问"
    else
        log_warning "简单测试页面访问失败: $test_response"
    fi
    
    # 测试认证测试页面
    local auth_test_response=$(curl -s -I "$FRONTEND_URL/test-auth" | head -n 1)
    if echo "$auth_test_response" | grep -q "200"; then
        log_success "认证测试页面可访问"
    else
        log_warning "认证测试页面访问失败: $auth_test_response"
    fi
}

# 检查页面内容
check_page_content() {
    log_info "检查页面内容..."
    
    # 检查主页面是否包含关键内容
    local main_content=$(curl -s "$FRONTEND_URL")
    if echo "$main_content" | grep -q "Stellar"; then
        log_success "主页面内容正常"
    else
        log_error "主页面内容异常"
        return 1
    fi
    
    # 检查登录页面是否包含关键内容
    local login_content=$(curl -s "$FRONTEND_URL/login")
    if echo "$login_content" | grep -q "登录"; then
        log_success "登录页面内容正常"
    else
        log_error "登录页面内容异常"
        return 1
    fi
}

# 主测试流程
main() {
    log_info "开始前端修复验证测试..."
    
    # 检查前端服务状态
    check_frontend_service
    
    # 测试页面访问
    test_page_access
    
    # 检查页面内容
    check_page_content
    
    log_success "前端修复验证测试完成！"
    log_info "请手动访问以下页面进行进一步测试："
    log_info "1. 主页面: $FRONTEND_URL"
    log_info "2. 登录页面: $FRONTEND_URL/login"
    log_info "3. 简单测试页面: $FRONTEND_URL/test-simple"
    log_info "4. 认证测试页面: $FRONTEND_URL/test-auth"
}

# 运行主测试
main "$@" 