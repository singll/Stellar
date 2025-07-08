.PHONY: dev build clean help

# ====================================================================================
# 开发环境
# 使用concurrently同时启动Go后端和Vite前端开发服务器
# ====================================================================================
dev:
	@echo "Starting services for development..."
	@echo "Please ensure you have run 'npm install' in the 'web' directory."
	@cd web && npm run dev:fullstack

# ====================================================================================
# 构建
# 构建前端并将其嵌入到Go二进制文件中
# ====================================================================================
build:
	@echo "Building frontend..."
	@cd web && npm install && npm run build
	@echo "Building Go backend with embedded frontend..."
	@go build -o stellar-server ./cmd/main.go
	@echo "Build complete! Run with ./stellar-server"

# ====================================================================================
# 清理
# 清理构建产物
# ====================================================================================
clean:
	@echo "Cleaning up build artifacts..."
	@rm -f stellar-server stellar-server.exe
	@rm -rf ./web/dist
	@echo "Cleanup complete."

help:
	@echo "Available commands:"
	@echo "  make dev          - Start backend and frontend for development"
	@echo "  make build        - Build frontend and embed into the Go binary"
	@echo "  make clean        - Clean up build artifacts" 