@echo off
echo ========================================
echo 星络（Stellar）密码修复验证测试
echo ========================================
echo.

echo 正在启动测试环境...
echo.

REM 检查MongoDB是否运行
echo [1/5] 检查MongoDB服务状态...
mongod --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ MongoDB未安装或未在PATH中
    echo 请先安装MongoDB: https://docs.mongodb.com/manual/installation/
    pause
    exit /b 1
)

REM 检查Redis是否运行
echo [2/5] 检查Redis服务状态...
redis-cli ping >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Redis未安装或未在PATH中
    echo 请先安装Redis: https://redis.io/download
    pause
    exit /b 1
)

REM 清理旧数据
echo [3/5] 清理旧数据...
if exist data\db rmdir /s /q data\db
mkdir data\db

echo [4/5] 启动MongoDB服务...
start /B mongod --dbpath ./data/db --port 27017
timeout /t 3 /nobreak >nul

echo [5/5] 启动Redis服务...
start /B redis-server --port 6379
timeout /t 2 /nobreak >nul

echo.
echo ✅ 测试环境准备完成
echo.

echo 正在构建并运行应用...
echo.

REM 构建应用
go build -o stellar.exe cmd/main.go
if %errorlevel% neq 0 (
    echo ❌ 构建失败
    pause
    exit /b 1
)

echo.
echo 🚀 启动应用进行密码修复验证...
echo.

REM 运行应用（使用开发配置）
start /B stellar.exe -config configs/config.dev.yaml

REM 等待应用启动
timeout /t 5 /nobreak >nul

echo.
echo 📋 验证步骤：
echo 1. 检查控制台输出是否包含管理员密码
echo 2. 使用输出的用户名和密码尝试登录
echo 3. 验证登录是否成功
echo.

echo 应用正在运行
echo 请查看上面的输出，找到管理员密码并尝试登录
echo 登录地址: http://localhost:8082
echo.

echo 按任意键停止应用...
pause >nul

REM 停止应用
taskkill /f /im stellar.exe >nul 2>&1

REM 清理进程
echo 正在清理测试环境...
taskkill /f /im mongod.exe >nul 2>&1
taskkill /f /im redis-server.exe >nul 2>&1
echo ✅ 清理完成

echo.
echo 测试完成！
echo 如果登录成功，说明密码修复有效。
pause 