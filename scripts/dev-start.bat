@echo off
setlocal enabledelayedexpansion

REM Stellar 开发环境启动脚本 (Windows)
REM 自动检查并清理端口占用，启动前后端服务

title Stellar Development Environment

REM 配置变量
set BACKEND_PORT=8090
set FRONTEND_PORT=5173
set PROJECT_ROOT=%~dp0..
set LOG_DIR=%PROJECT_ROOT%\logs
set BACKEND_PID_FILE=%LOG_DIR%\backend.pid
set FRONTEND_PID_FILE=%LOG_DIR%\frontend.pid

REM 颜色定义 (Windows 10+)
set ESC=[
set RED=%ESC%[31m
set GREEN=%ESC%[32m
set YELLOW=%ESC%[33m
set BLUE=%ESC%[34m
set NC=%ESC%[0m

echo ======================================
echo Stellar 开发环境启动脚本 (Windows)
echo ======================================

REM 创建日志目录
if not exist "%LOG_DIR%" (
    mkdir "%LOG_DIR%"
    echo [INFO] 创建日志目录: %LOG_DIR%
)

REM 检查必要工具
echo [INFO] 检查必要工具...
where go >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go 未安装或未添加到 PATH
    pause
    exit /b 1
)

where pnpm >nul 2>&1
if errorlevel 1 (
    echo [ERROR] pnpm 未安装或未添加到 PATH
    pause
    exit /b 1
)

echo [SUCCESS] 所有必要工具已安装

REM 清理端口 (Windows 使用netstat)
echo [INFO] 检查端口占用情况...

REM 检查后端端口
netstat -ano | findstr ":%BACKEND_PORT%" | findstr "LISTENING" >nul
if not errorlevel 1 (
    echo [WARNING] 端口 %BACKEND_PORT% 被占用，正在清理...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%BACKEND_PORT%" ^| findstr "LISTENING"') do (
        echo [INFO] 终止进程 PID: %%a
        taskkill /PID %%a /F >nul 2>&1
    )
    timeout /t 2 >nul
    echo [SUCCESS] 端口 %BACKEND_PORT% 已清理
) else (
    echo [SUCCESS] 端口 %BACKEND_PORT% 空闲
)

REM 检查前端端口
netstat -ano | findstr ":%FRONTEND_PORT%" | findstr "LISTENING" >nul
if not errorlevel 1 (
    echo [WARNING] 端口 %FRONTEND_PORT% 被占用，正在清理...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%FRONTEND_PORT%" ^| findstr "LISTENING"') do (
        echo [INFO] 终止进程 PID: %%a
        taskkill /PID %%a /F >nul 2>&1
    )
    timeout /t 2 >nul
    echo [SUCCESS] 端口 %FRONTEND_PORT% 已清理
) else (
    echo [SUCCESS] 端口 %FRONTEND_PORT% 空闲
)

REM 清理旧的PID文件
if exist "%BACKEND_PID_FILE%" del "%BACKEND_PID_FILE%"
if exist "%FRONTEND_PID_FILE%" del "%FRONTEND_PID_FILE%"

REM 检查数据库连接 (简化版)
echo [INFO] 检查数据库连接...
ping -n 1 192.168.7.216 >nul 2>&1
if errorlevel 1 (
    echo [WARNING] MongoDB 主机 192.168.7.216 不可达
) else (
    echo [SUCCESS] MongoDB 主机连接正常
)

ping -n 1 192.168.7.128 >nul 2>&1
if errorlevel 1 (
    echo [WARNING] Redis 主机 192.168.7.128 不可达
) else (
    echo [SUCCESS] Redis 主机连接正常
)

REM 安装前端依赖
echo [INFO] 检查前端依赖...
cd /d "%PROJECT_ROOT%\web"

if not exist "node_modules" (
    echo [INFO] 安装前端依赖...
    pnpm install
    if errorlevel 1 (
        echo [ERROR] 前端依赖安装失败
        pause
        exit /b 1
    )
    echo [SUCCESS] 前端依赖安装完成
) else (
    echo [SUCCESS] 前端依赖已存在
)

REM 启动后端服务
echo [INFO] 启动后端服务...
cd /d "%PROJECT_ROOT%"

if not exist "config.dev.yaml" (
    echo [ERROR] 配置文件 config.dev.yaml 不存在
    pause
    exit /b 1
)

start /b cmd /c "go run ./cmd/main.go -config config.dev.yaml -log-level debug > \"%LOG_DIR%\backend.log\" 2>&1"

REM 等待后端启动
echo [INFO] 等待后端服务启动...
set /a attempt=0
:wait_backend
set /a attempt+=1
netstat -ano | findstr ":%BACKEND_PORT%" | findstr "LISTENING" >nul
if not errorlevel 1 (
    echo [SUCCESS] 后端服务启动成功，监听端口 %BACKEND_PORT%
    goto backend_ready
)

if %attempt% geq 30 (
    echo [ERROR] 后端服务启动超时
    pause
    exit /b 1
)

timeout /t 1 >nul
goto wait_backend

:backend_ready

REM 启动前端服务
echo [INFO] 启动前端服务...
cd /d "%PROJECT_ROOT%\web"

start /b cmd /c "pnpm run dev > \"%LOG_DIR%\frontend.log\" 2>&1"

REM 等待前端启动
echo [INFO] 等待前端服务启动...
set /a attempt=0
:wait_frontend
set /a attempt+=1
netstat -ano | findstr ":%FRONTEND_PORT%" | findstr "LISTENING" >nul
if not errorlevel 1 (
    echo [SUCCESS] 前端服务启动成功，监听端口 %FRONTEND_PORT%
    goto frontend_ready
)

if %attempt% geq 30 (
    echo [ERROR] 前端服务启动超时
    pause
    exit /b 1
)

timeout /t 1 >nul
goto wait_frontend

:frontend_ready

REM 显示服务状态
echo.
echo [SUCCESS] 开发环境启动完成！
echo.
echo 服务信息:
echo   后端服务: http://localhost:%BACKEND_PORT%
echo   前端服务: http://localhost:%FRONTEND_PORT%
echo   API接口: http://localhost:%BACKEND_PORT%/api/v1
echo   WebSocket: ws://localhost:%BACKEND_PORT%/ws
echo.
echo 日志文件:
echo   后端日志: %LOG_DIR%\backend.log
echo   前端日志: %LOG_DIR%\frontend.log
echo.
echo 停止服务: %PROJECT_ROOT%\scripts\dev-stop.bat
echo.
echo 按任意键退出启动脚本（服务将继续运行）...
pause >nul