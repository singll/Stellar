@echo off
echo ========================================
echo 星络（Stellar）数据库初始化测试
echo ========================================
echo.

echo 正在启动测试环境...
echo.

REM 检查MongoDB是否运行
echo [1/4] 检查MongoDB服务状态...
mongod --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ MongoDB未安装或未在PATH中
    echo 请先安装MongoDB: https://docs.mongodb.com/manual/installation/
    pause
    exit /b 1
)

REM 检查Redis是否运行
echo [2/4] 检查Redis服务状态...
redis-cli ping >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Redis未安装或未在PATH中
    echo 请先安装Redis: https://redis.io/download
    pause
    exit /b 1
)

echo [3/4] 启动MongoDB服务...
start /B mongod --dbpath ./data/db --port 27017
timeout /t 3 /nobreak >nul

echo [4/4] 启动Redis服务...
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
echo 🚀 启动应用进行初始化测试...
echo.

REM 运行应用（使用开发配置）
stellar.exe -config configs/config.dev.yaml

echo.
echo 测试完成！
echo.
echo 如果看到以下信息，说明初始化成功：
echo - ✅ 数据库初始化完成
echo - ✅ Redis初始化完成  
echo - 🔑 管理员账户信息（包含用户名和密码）
echo.
pause 