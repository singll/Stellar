#!/bin/bash

# Stellar 开发环境停止脚本
# 安全停止前后端服务

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
BACKEND_PORT=8090
FRONTEND_PORT=5173
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="${PROJECT_ROOT}/logs"
BACKEND_PID_FILE="${LOG_DIR}/backend.pid"
FRONTEND_PID_FILE="${LOG_DIR}/frontend.pid"

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

# 检查端口是否被占用
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 0  # 端口被占用
    else
        return 1  # 端口空闲
    fi
}

# 停止指定PID的进程
stop_process_by_pid() {
    local pid_file=$1
    local service_name=$2
    
    if [ ! -f "$pid_file" ]; then
        log_warning "$service_name PID文件不存在: $pid_file"
        return 1
    fi
    
    local pid=$(cat "$pid_file")
    
    if ! kill -0 $pid 2>/dev/null; then
        log_warning "$service_name 进程 (PID: $pid) 已经停止"
        rm -f "$pid_file"
        return 1
    fi
    
    log_info "停止 $service_name 进程 (PID: $pid)"
    
    # 先发送TERM信号
    kill -TERM $pid 2>/dev/null || true
    
    # 等待进程退出
    local max_attempts=10
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if ! kill -0 $pid 2>/dev/null; then
            log_success "$service_name 进程已停止"
            rm -f "$pid_file"
            return 0
        fi
        
        sleep 1
        attempt=$((attempt + 1))
        echo -n "."
    done
    
    # 如果进程仍在运行，强制杀死
    log_warning "强制终止 $service_name 进程 (PID: $pid)"
    kill -KILL $pid 2>/dev/null || true
    rm -f "$pid_file"
    
    if ! kill -0 $pid 2>/dev/null; then
        log_success "$service_name 进程已强制停止"
        return 0
    else
        log_error "无法停止 $service_name 进程 (PID: $pid)"
        return 1
    fi
}

# 通过端口停止进程
stop_process_by_port() {
    local port=$1
    local service_name=$2
    
    if ! check_port $port; then
        log_success "$service_name 端口 $port 已空闲"
        return 0
    fi
    
    log_info "发现端口 $port 上的 $service_name 进程"
    local pids=$(lsof -ti :$port)
    
    for pid in $pids; do
        local process_info=$(ps -p $pid -o pid,cmd --no-headers 2>/dev/null || echo "进程已退出")
        log_info "停止进程: $process_info"
        
        # 先发送TERM信号
        kill -TERM $pid 2>/dev/null || true
    done
    
    # 等待进程退出
    sleep 3
    
    # 检查是否还有进程占用端口
    if check_port $port; then
        log_warning "强制终止占用端口 $port 的进程"
        pids=$(lsof -ti :$port)
        for pid in $pids; do
            kill -KILL $pid 2>/dev/null || true
        done
        sleep 1
    fi
    
    if check_port $port; then
        log_error "无法释放端口 $port"
        return 1
    else
        log_success "$service_name 端口 $port 已释放"
        return 0
    fi
}

# 停止后端服务
stop_backend() {
    log_info "停止后端服务..."
    
    local stopped_by_pid=false
    local stopped_by_port=false
    
    # 先尝试通过PID文件停止
    if stop_process_by_pid "$BACKEND_PID_FILE" "后端服务"; then
        stopped_by_pid=true
    fi
    
    # 再检查端口，确保没有遗漏的进程
    if ! $stopped_by_pid || check_port $BACKEND_PORT; then
        if stop_process_by_port $BACKEND_PORT "后端服务"; then
            stopped_by_port=true
        fi
    fi
    
    if $stopped_by_pid || $stopped_by_port; then
        log_success "后端服务已停止"
    else
        log_warning "后端服务可能未在运行"
    fi
}

# 停止前端服务
stop_frontend() {
    log_info "停止前端服务..."
    
    local stopped_by_pid=false
    local stopped_by_port=false
    
    # 先尝试通过PID文件停止
    if stop_process_by_pid "$FRONTEND_PID_FILE" "前端服务"; then
        stopped_by_pid=true
    fi
    
    # 再检查端口，确保没有遗漏的进程
    if ! $stopped_by_pid || check_port $FRONTEND_PORT; then
        if stop_process_by_port $FRONTEND_PORT "前端服务"; then
            stopped_by_port=true
        fi
    fi
    
    if $stopped_by_pid || $stopped_by_port; then
        log_success "前端服务已停止"
    else
        log_warning "前端服务可能未在运行"
    fi
}

# 清理相关进程
cleanup_related_processes() {
    log_info "清理相关进程..."
    
    # 查找并清理可能的Go进程
    local go_processes=$(pgrep -f "go run.*cmd/main.go" || true)
    if [ -n "$go_processes" ]; then
        log_info "清理Go开发进程..."
        for pid in $go_processes; do
            local process_info=$(ps -p $pid -o pid,cmd --no-headers 2>/dev/null || echo "进程已退出")
            log_info "停止进程: $process_info"
            kill -TERM $pid 2>/dev/null || true
        done
        sleep 2
        
        # 强制清理
        go_processes=$(pgrep -f "go run.*cmd/main.go" || true)
        if [ -n "$go_processes" ]; then
            for pid in $go_processes; do
                kill -KILL $pid 2>/dev/null || true
            done
        fi
    fi
    
    # 查找并清理可能的Vite进程
    local vite_processes=$(pgrep -f "vite.*dev" || true)
    if [ -n "$vite_processes" ]; then
        log_info "清理Vite开发进程..."
        for pid in $vite_processes; do
            local process_info=$(ps -p $pid -o pid,cmd --no-headers 2>/dev/null || echo "进程已退出")
            log_info "停止进程: $process_info"
            kill -TERM $pid 2>/dev/null || true
        done
        sleep 2
        
        # 强制清理
        vite_processes=$(pgrep -f "vite.*dev" || true)
        if [ -n "$vite_processes" ]; then
            for pid in $vite_processes; do
                kill -KILL $pid 2>/dev/null || true
            done
        fi
    fi
}

# 显示最终状态
show_final_status() {
    echo
    log_info "检查最终状态..."
    
    if check_port $BACKEND_PORT; then
        log_warning "后端端口 $BACKEND_PORT 仍被占用"
        lsof -Pi :$BACKEND_PORT -sTCP:LISTEN
    else
        log_success "后端端口 $BACKEND_PORT 已释放"
    fi
    
    if check_port $FRONTEND_PORT; then
        log_warning "前端端口 $FRONTEND_PORT 仍被占用"
        lsof -Pi :$FRONTEND_PORT -sTCP:LISTEN
    else
        log_success "前端端口 $FRONTEND_PORT 已释放"
    fi
    
    echo
    log_success "Stellar 开发环境已停止"
}

# 主函数
main() {
    echo "======================================"
    echo "Stellar 开发环境停止脚本"
    echo "======================================"
    
    stop_backend
    stop_frontend
    cleanup_related_processes
    show_final_status
}

# 执行主函数
main "$@"