#!/bin/bash

# Stellar 开发环境启动脚本
# 自动检查并清理端口占用，启动前后端服务

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

# 创建日志目录
create_log_dir() {
    if [ ! -d "$LOG_DIR" ]; then
        mkdir -p "$LOG_DIR"
        log_info "创建日志目录: $LOG_DIR"
    fi
}

# 检查命令是否存在
check_command() {
    if ! command -v "$1" &> /dev/null; then
        log_error "命令 '$1' 未找到，请安装后重试"
        exit 1
    fi
}

# 检查必要的工具
check_prerequisites() {
    log_info "检查必要工具..."
    check_command "go"
    check_command "pnpm"
    check_command "lsof"
    log_success "所有必要工具已安装"
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

# 获取占用端口的进程信息
get_port_process() {
    local port=$1
    lsof -Pi :$port -sTCP:LISTEN -n | tail -n +2
}

# 杀死占用端口的进程
kill_port_process() {
    local port=$1
    local pids=$(lsof -ti :$port)
    
    if [ -n "$pids" ]; then
        log_warning "正在终止占用端口 $port 的进程..."
        for pid in $pids; do
            local process_info=$(ps -p $pid -o pid,ppid,cmd --no-headers 2>/dev/null || echo "进程已退出")
            log_info "终止进程: $process_info"
            kill -TERM $pid 2>/dev/null || true
        done
        
        # 等待进程退出
        sleep 2
        
        # 强制杀死仍在运行的进程
        pids=$(lsof -ti :$port 2>/dev/null || true)
        if [ -n "$pids" ]; then
            log_warning "强制终止顽固进程..."
            for pid in $pids; do
                kill -KILL $pid 2>/dev/null || true
            done
            sleep 1
        fi
        
        log_success "端口 $port 已清理"
    fi
}

# 清理端口
cleanup_ports() {
    log_info "检查端口占用情况..."
    
    # 检查后端端口
    if check_port $BACKEND_PORT; then
        log_warning "端口 $BACKEND_PORT (后端) 被占用:"
        get_port_process $BACKEND_PORT
        kill_port_process $BACKEND_PORT
    else
        log_success "端口 $BACKEND_PORT (后端) 空闲"
    fi
    
    # 检查前端端口
    if check_port $FRONTEND_PORT; then
        log_warning "端口 $FRONTEND_PORT (前端) 被占用:"
        get_port_process $FRONTEND_PORT
        kill_port_process $FRONTEND_PORT
    else
        log_success "端口 $FRONTEND_PORT (前端) 空闲"
    fi
}

# 清理旧的PID文件
cleanup_pid_files() {
    log_info "清理旧的PID文件..."
    rm -f "$BACKEND_PID_FILE" "$FRONTEND_PID_FILE"
}

# 检查数据库连接
check_database_connection() {
    log_info "检查数据库连接..."
    
    # 检查MongoDB连接
    if timeout 5 bash -c "</dev/tcp/192.168.7.216/27017" 2>/dev/null; then
        log_success "MongoDB (192.168.7.216:27017) 连接正常"
    else
        log_warning "MongoDB (192.168.7.216:27017) 连接失败，请检查数据库服务"
    fi
    
    # 检查Redis连接
    if timeout 5 bash -c "</dev/tcp/192.168.7.128/6379" 2>/dev/null; then
        log_success "Redis (192.168.7.128:6379) 连接正常"
    else
        log_warning "Redis (192.168.7.128:6379) 连接失败，请检查缓存服务"
    fi
}

# 安装前端依赖
install_frontend_deps() {
    log_info "检查前端依赖..."
    cd "$PROJECT_ROOT/web"
    
    if [ ! -d "node_modules" ] || [ "package.json" -nt "node_modules" ]; then
        log_info "安装前端依赖..."
        pnpm install
        log_success "前端依赖安装完成"
    else
        log_success "前端依赖已是最新"
    fi
}

# 启动后端服务
start_backend() {
    log_info "启动后端服务..."
    cd "$PROJECT_ROOT"
    
    # 检查配置文件
    if [ ! -f "config.dev.yaml" ]; then
        log_error "配置文件 config.dev.yaml 不存在"
        exit 1
    fi
    
    # 启动后端服务
    nohup go run ./cmd/main.go -config config.dev.yaml -log-level debug \
        > "$LOG_DIR/backend.log" 2>&1 &
    
    local backend_pid=$!
    echo $backend_pid > "$BACKEND_PID_FILE"
    
    log_info "后端服务已启动 (PID: $backend_pid)"
    
    # 等待后端启动
    log_info "等待后端服务启动..."
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if check_port $BACKEND_PORT; then
            log_success "后端服务启动成功，监听端口 $BACKEND_PORT"
            return 0
        fi
        
        # 检查进程是否还在运行
        if ! kill -0 $backend_pid 2>/dev/null; then
            log_error "后端服务启动失败，请检查日志: $LOG_DIR/backend.log"
            exit 1
        fi
        
        sleep 1
        attempt=$((attempt + 1))
        echo -n "."
    done
    
    log_error "后端服务启动超时"
    exit 1
}

# 启动前端服务
start_frontend() {
    log_info "启动前端服务..."
    cd "$PROJECT_ROOT/web"
    
    # 启动前端开发服务器
    nohup pnpm run dev > "$LOG_DIR/frontend.log" 2>&1 &
    
    local frontend_pid=$!
    echo $frontend_pid > "$FRONTEND_PID_FILE"
    
    log_info "前端服务已启动 (PID: $frontend_pid)"
    
    # 等待前端启动
    log_info "等待前端服务启动..."
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if check_port $FRONTEND_PORT; then
            log_success "前端服务启动成功，监听端口 $FRONTEND_PORT"
            return 0
        fi
        
        # 检查进程是否还在运行
        if ! kill -0 $frontend_pid 2>/dev/null; then
            log_error "前端服务启动失败，请检查日志: $LOG_DIR/frontend.log"
            exit 1
        fi
        
        sleep 1
        attempt=$((attempt + 1))
        echo -n "."
    done
    
    log_error "前端服务启动超时"
    exit 1
}

# 显示服务状态
show_status() {
    echo
    log_success "开发环境启动完成！"
    echo
    echo "服务信息:"
    echo "  后端服务: http://localhost:$BACKEND_PORT"
    echo "  前端服务: http://localhost:$FRONTEND_PORT"
    echo "  API接口: http://localhost:$BACKEND_PORT/api/v1"
    echo "  WebSocket: ws://localhost:$BACKEND_PORT/ws"
    echo
    echo "日志文件:"
    echo "  后端日志: $LOG_DIR/backend.log"
    echo "  前端日志: $LOG_DIR/frontend.log"
    echo
    echo "停止服务: $PROJECT_ROOT/scripts/dev-stop.sh"
    echo "查看日志: tail -f $LOG_DIR/{backend,frontend}.log"
    echo
}

# 设置退出处理
cleanup_on_exit() {
    log_warning "收到退出信号，正在清理..."
    
    if [ -f "$BACKEND_PID_FILE" ]; then
        local backend_pid=$(cat "$BACKEND_PID_FILE")
        if kill -0 $backend_pid 2>/dev/null; then
            log_info "停止后端服务 (PID: $backend_pid)"
            kill -TERM $backend_pid 2>/dev/null || true
        fi
    fi
    
    if [ -f "$FRONTEND_PID_FILE" ]; then
        local frontend_pid=$(cat "$FRONTEND_PID_FILE")
        if kill -0 $frontend_pid 2>/dev/null; then
            log_info "停止前端服务 (PID: $frontend_pid)"
            kill -TERM $frontend_pid 2>/dev/null || true
        fi
    fi
    
    cleanup_pid_files
    log_info "清理完成"
}

# 主函数
main() {
    echo "======================================"
    echo "Stellar 开发环境启动脚本"
    echo "======================================"
    
    # 设置信号处理
    trap cleanup_on_exit EXIT INT TERM
    
    # 执行启动流程
    create_log_dir
    check_prerequisites
    cleanup_ports
    cleanup_pid_files
    check_database_connection
    install_frontend_deps
    start_backend
    start_frontend
    show_status
    
    # 如果是交互模式，等待用户输入
    if [ -t 0 ]; then
        echo "按 Ctrl+C 停止服务..."
        while true; do
            sleep 1
        done
    fi
}

# 执行主函数
main "$@"