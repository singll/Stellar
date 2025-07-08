.PHONY: dev dev-stop build clean help install-deps check-deps status logs

# ====================================================================================
# 开发环境
# 使用高级启动脚本，包含端口检查、进程清理和错误处理
# ====================================================================================
dev:
	@echo "启动 Stellar 开发环境..."
	@./scripts/dev-start.sh

# 停止开发环境
dev-stop:
	@echo "停止 Stellar 开发环境..."
	@./scripts/dev-stop.sh

# 检查开发环境状态
status:
	@echo "检查 Stellar 开发环境状态..."
	@echo "======================================"
	@echo "端口占用情况:"
	@if lsof -Pi :8090 -sTCP:LISTEN -t >/dev/null 2>&1; then \
		echo "  后端 (8090): ✓ 运行中"; \
		lsof -Pi :8090 -sTCP:LISTEN; \
	else \
		echo "  后端 (8090): ✗ 未运行"; \
	fi
	@if lsof -Pi :5173 -sTCP:LISTEN -t >/dev/null 2>&1; then \
		echo "  前端 (5173): ✓ 运行中"; \
		lsof -Pi :5173 -sTCP:LISTEN; \
	else \
		echo "  前端 (5173): ✗ 未运行"; \
	fi
	@echo "======================================"
	@echo "日志文件:"
	@if [ -f "logs/backend.log" ]; then \
		echo "  后端日志: logs/backend.log ($(wc -l < logs/backend.log) 行)"; \
	else \
		echo "  后端日志: 不存在"; \
	fi
	@if [ -f "logs/frontend.log" ]; then \
		echo "  前端日志: logs/frontend.log ($(wc -l < logs/frontend.log) 行)"; \
	else \
		echo "  前端日志: 不存在"; \
	fi

# 查看实时日志
logs:
	@echo "查看 Stellar 开发环境日志..."
	@if [ -f "logs/backend.log" ] && [ -f "logs/frontend.log" ]; then \
		echo "按 Ctrl+C 退出日志查看"; \
		tail -f logs/backend.log logs/frontend.log; \
	else \
		echo "日志文件不存在，请先启动开发环境"; \
	fi

# 检查依赖
check-deps:
	@echo "检查系统依赖..."
	@echo "======================================"
	@if command -v go >/dev/null 2>&1; then \
		echo "Go: ✓ $(go version)"; \
	else \
		echo "Go: ✗ 未安装"; \
	fi
	@if command -v pnpm >/dev/null 2>&1; then \
		echo "pnpm: ✓ $(pnpm --version)"; \
	else \
		echo "pnpm: ✗ 未安装"; \
	fi
	@if command -v lsof >/dev/null 2>&1; then \
		echo "lsof: ✓ 已安装"; \
	else \
		echo "lsof: ✗ 未安装"; \
	fi
	@echo "======================================"
	@echo "数据库连接:"
	@if timeout 5 bash -c '</dev/tcp/192.168.7.216/27017' 2>/dev/null; then \
		echo "MongoDB (192.168.7.216:27017): ✓ 连接正常"; \
	else \
		echo "MongoDB (192.168.7.216:27017): ✗ 连接失败"; \
	fi
	@if timeout 5 bash -c '</dev/tcp/192.168.7.128/6379' 2>/dev/null; then \
		echo "Redis (192.168.7.128:6379): ✓ 连接正常"; \
	else \
		echo "Redis (192.168.7.128:6379): ✗ 连接失败"; \
	fi

# 安装依赖
install-deps:
	@echo "安装项目依赖..."
	@echo "安装 Go 模块..."
	@go mod tidy
	@go mod download
	@echo "安装前端依赖..."
	@cd web && pnpm install
	@echo "创建必要目录..."
	@mkdir -p logs scripts
	@echo "依赖安装完成"

# ====================================================================================
# 构建
# 构建前端并将其嵌入到Go二进制文件中
# ====================================================================================
build:
	@echo "构建 Stellar 生产版本..."
	@echo "构建前端..."
	@cd web && pnpm install && pnpm run build
	@echo "构建Go后端并嵌入前端..."
	@go build -ldflags "-s -w" -o stellar-server ./cmd/main.go
	@echo "构建完成! 运行: ./stellar-server"

# ====================================================================================
# 清理
# 清理构建产物和临时文件
# ====================================================================================
clean:
	@echo "清理构建产物和临时文件..."
	@rm -f stellar-server stellar-server.exe
	@rm -rf ./web/dist
	@rm -rf ./web/.svelte-kit
	@rm -rf ./web/node_modules/.vite
	@rm -f ./logs/*.log
	@rm -f ./logs/*.pid
	@echo "清理完成"

# 深度清理（包括依赖）
clean-all: clean
	@echo "执行深度清理..."
	@rm -rf ./web/node_modules
	@go clean -cache -modcache -i -r
	@echo "深度清理完成"

help:
	@echo "Stellar 开发环境命令:"
	@echo "======================================"
	@echo "开发命令:"
	@echo "  make dev          - 启动开发环境（前端+后端）"
	@echo "  make dev-stop     - 停止开发环境"
	@echo "  make status       - 查看开发环境状态"
	@echo "  make logs         - 查看实时日志"
	@echo ""
	@echo "依赖管理:"
	@echo "  make check-deps   - 检查系统依赖和数据库连接"
	@echo "  make install-deps - 安装项目依赖"
	@echo ""
	@echo "构建和清理:"
	@echo "  make build        - 构建生产版本"
	@echo "  make clean        - 清理构建产物"
	@echo "  make clean-all    - 深度清理（包括依赖）"
	@echo ""
	@echo "其他:"
	@echo "  make help         - 显示此帮助信息"
	@echo "======================================"
	@echo "快速开始:"
	@echo "  1. make check-deps  # 检查依赖"
	@echo "  2. make install-deps # 安装依赖"
	@echo "  3. make dev         # 启动开发环境" 