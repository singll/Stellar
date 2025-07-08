#!/bin/bash

# Stellar 服务健康检查脚本
# 检查后端API、前端服务和数据库连接状态

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
BACKEND_URL="http://localhost:8090"
FRONTEND_URL="http://localhost:5173"
MONGODB_HOST="192.168.7.216"
MONGODB_PORT="27017"
REDIS_HOST="192.168.7.128"
REDIS_PORT="6379"

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

# 检查HTTP服务
check_http_service() {
    local url=$1
    local service_name=$2
    local timeout=${3:-5}
    
    if curl -s --max-time $timeout "$url" >/dev/null 2>&1; then
        log_success "$service_name 服务健康 ($url)"
        return 0
    else
        log_error "$service_name 服务异常 ($url)"
        return 1
    fi
}

# 检查API端点
check_api_endpoint() {
    local endpoint=$1
    local description=$2
    local expected_status=${3:-200}
    
    local full_url="${BACKEND_URL}${endpoint}"
    
    if command -v curl >/dev/null 2>&1; then
        local response=$(curl -s -w "%{http_code}" -o /dev/null --max-time 5 "$full_url" 2>/dev/null || echo "000")
        
        if [ "$response" = "$expected_status" ]; then
            log_success "$description API 正常 ($endpoint)"
            return 0
        else
            log_warning "$description API 异常 ($endpoint) - HTTP状态码: $response"
            return 1
        fi
    else
        log_warning "curl 未安装，跳过 $description API 检查"
        return 1
    fi
}

# 检查端口连通性
check_port_connectivity() {
    local host=$1
    local port=$2
    local service_name=$3
    local timeout=${4:-5}
    
    if timeout $timeout bash -c "</dev/tcp/$host/$port" 2>/dev/null; then
        log_success "$service_name 端口连通 ($host:$port)"
        return 0
    else
        log_error "$service_name 端口不通 ($host:$port)"
        return 1
    fi
}

# 检查系统资源
check_system_resources() {
    log_info "检查系统资源..."
    
    # 内存使用率
    if command -v free >/dev/null 2>&1; then
        local mem_usage=$(free | grep Mem | awk '{printf "%.1f", ($3/$2) * 100.0}')
        if (( $(echo "$mem_usage > 90" | bc -l) )); then
            log_warning "内存使用率较高: ${mem_usage}%"
        else
            log_success "内存使用正常: ${mem_usage}%"
        fi
    fi
    
    # 磁盘使用率
    if command -v df >/dev/null 2>&1; then
        local disk_usage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
        if [ "$disk_usage" -gt 90 ]; then
            log_warning "磁盘使用率较高: ${disk_usage}%"
        else
            log_success "磁盘使用正常: ${disk_usage}%"
        fi
    fi
    
    # CPU负载
    if command -v uptime >/dev/null 2>&1; then
        local load_avg=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
        log_info "系统负载: $load_avg"
    fi
}

# 检查进程状态
check_processes() {
    log_info "检查关键进程..."
    
    # 检查Go进程
    local go_processes=$(pgrep -f "go run.*cmd/main.go" | wc -l)
    if [ "$go_processes" -gt 0 ]; then
        log_success "Go后端进程运行中 ($go_processes 个)"
    else
        log_warning "Go后端进程未找到"
    fi
    
    # 检查Node进程（前端）
    local node_processes=$(pgrep -f "vite.*dev" | wc -l)
    if [ "$node_processes" -gt 0 ]; then
        log_success "前端开发进程运行中 ($node_processes 个)"
    else
        log_warning "前端开发进程未找到"
    fi
}

# 主健康检查函数
main_health_check() {
    echo "======================================"
    echo "Stellar 服务健康检查"
    echo "======================================"
    
    local error_count=0
    
    # 检查前端服务
    log_info "检查前端服务..."
    if ! check_http_service "$FRONTEND_URL" "前端"; then
        error_count=$((error_count + 1))
    fi
    
    # 检查后端服务
    log_info "检查后端服务..."
    if ! check_http_service "$BACKEND_URL" "后端"; then
        error_count=$((error_count + 1))
    fi
    
    # 检查API端点
    if curl -s --max-time 2 "$BACKEND_URL" >/dev/null 2>&1; then
        log_info "检查关键API端点..."
        
        # 检查认证相关API（这些通常返回401是正常的）
        check_api_endpoint "/api/v1/auth/info" "用户信息" "401"
        
        # 检查项目API
        check_api_endpoint "/api/v1/projects/projects" "项目列表" "401"
        
        # 检查WebSocket（简单的连通性检查）
        if command -v nc >/dev/null 2>&1; then
            if echo "" | nc -w 1 localhost 8090 >/dev/null 2>&1; then
                log_success "WebSocket端口可访问"
            else
                log_warning "WebSocket端口不可访问"
            fi
        fi
    fi
    
    # 检查数据库连接
    log_info "检查数据库连接..."
    if ! check_port_connectivity "$MONGODB_HOST" "$MONGODB_PORT" "MongoDB"; then
        error_count=$((error_count + 1))
    fi
    
    if ! check_port_connectivity "$REDIS_HOST" "$REDIS_PORT" "Redis"; then
        error_count=$((error_count + 1))
    fi
    
    # 检查系统资源
    check_system_resources
    
    # 检查进程状态
    check_processes
    
    # 输出总结
    echo "======================================"
    if [ $error_count -eq 0 ]; then
        log_success "所有服务运行正常！"
        exit 0
    else
        log_error "发现 $error_count 个问题，请检查服务状态"
        exit 1
    fi
}

# 执行健康检查
main_health_check "$@"