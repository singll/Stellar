#!/bin/bash

# Stellar 项目快速启动脚本
# 用于验证基本功能和数据库连接

echo "🌟 Stellar 项目启动检查..."

# 检查Go环境
echo "📋 检查Go环境..."
if ! command -v go &> /dev/null; then
    echo "❌ Go未安装，请先安装Go"
    exit 1
fi
echo "✅ Go版本: $(go version)"

# 检查依赖
echo "📋 检查Go依赖..."
if [ ! -f "go.mod" ]; then
    echo "❌ go.mod文件不存在"
    exit 1
fi

echo "📦 下载依赖..."
go mod download
if [ $? -ne 0 ]; then
    echo "❌ 依赖下载失败"
    exit 1
fi

# 检查配置文件
echo "📋 检查配置文件..."
CONFIG_FILE="config.dev.yaml"
if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ 配置文件 $CONFIG_FILE 不存在"
    exit 1
fi
echo "✅ 使用配置文件: $CONFIG_FILE"

# 显示数据库连接信息
echo "🔗 数据库连接信息:"
echo "   MongoDB: 192.168.7.216:27017 (数据库: stellarserver_dev)"
echo "   Redis: 192.168.7.128:6379 (db: 1)"
echo "   ⚠️  请确保数据库服务可访问"
echo ""

echo "🚀 启动Stellar服务器..."
echo "📍 后端服务: http://0.0.0.0:8090"
echo "📍 前端服务: http://0.0.0.0:5173"
echo "📍 API地址: http://0.0.0.0:8090/api/v1"
echo "📍 WebSocket: ws://0.0.0.0:8090/ws"
echo ""

# 启动服务器
go run ./cmd/main.go -config "$CONFIG_FILE" -log-level debug -show-routes