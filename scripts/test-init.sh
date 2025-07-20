#!/bin/bash

echo "========================================"
echo "星络（Stellar）数据库初始化测试"
echo "========================================"
echo

echo "正在启动测试环境..."
echo

# 检查MongoDB是否运行
echo "[1/4] 检查MongoDB服务状态..."
if ! command -v mongod &> /dev/null; then
    echo "❌ MongoDB未安装或未在PATH中"
    echo "请先安装MongoDB: https://docs.mongodb.com/manual/installation/"
    exit 1
fi

# 检查Redis是否运行
echo "[2/4] 检查Redis服务状态..."
if ! command -v redis-server &> /dev/null; then
    echo "❌ Redis未安装或未在PATH中"
    echo "请先安装Redis: https://redis.io/download"
    exit 1
fi

# 创建数据目录
mkdir -p ./data/db

echo "[3/4] 启动MongoDB服务..."
mongod --dbpath ./data/db --port 27017 --fork --logpath ./data/mongod.log
sleep 3

echo "[4/4] 启动Redis服务..."
redis-server --port 6379 --daemonize yes --logfile ./data/redis.log
sleep 2

echo
echo "✅ 测试环境准备完成"
echo

echo "正在构建并运行应用..."
echo

# 构建应用
go build -o stellar cmd/main.go
if [ $? -ne 0 ]; then
    echo "❌ 构建失败"
    exit 1
fi

echo
echo "🚀 启动应用进行初始化测试..."
echo

# 运行应用（使用开发配置）
./stellar -config configs/config.dev.yaml

echo
echo "测试完成！"
echo
echo "如果看到以下信息，说明初始化成功："
echo "- ✅ 数据库初始化完成"
echo "- ✅ Redis初始化完成"
echo "- 🔑 管理员账户信息（包含用户名和密码）"
echo

# 清理进程
echo "正在清理测试环境..."
pkill -f mongod
pkill -f redis-server
echo "✅ 清理完成" 