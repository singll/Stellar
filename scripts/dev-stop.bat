@echo off
setlocal enabledelayedexpansion

REM Stellar 开发环境停止脚本 (Windows)

title Stellar Development Environment - Stop

REM 配置变量
set BACKEND_PORT=8090
set FRONTEND_PORT=5173
set PROJECT_ROOT=%~dp0..
set LOG_DIR=%PROJECT_ROOT%\logs

echo ======================================
echo Stellar 开发环境停止脚本 (Windows)
echo ======================================

REM 停止后端服务
echo [INFO] 停止后端服务...
netstat -ano | findstr ":%BACKEND_PORT%" | findstr "LISTENING" >nul
if not errorlevel 1 (
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%BACKEND_PORT%" ^| findstr "LISTENING"') do (
        echo [INFO] 停止后端进程 PID: %%a
        taskkill /PID %%a /F >nul 2>&1
    )
    echo [SUCCESS] 后端服务已停止
) else (
    echo [WARNING] 后端服务可能未在运行
)

REM 停止前端服务
echo [INFO] 停止前端服务...
netstat -ano | findstr ":%FRONTEND_PORT%" | findstr "LISTENING" >nul
if not errorlevel 1 (
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%FRONTEND_PORT%" ^| findstr "LISTENING"') do (
        echo [INFO] 停止前端进程 PID: %%a
        taskkill /PID %%a /F >nul 2>&1
    )
    echo [SUCCESS] 前端服务已停止
) else (
    echo [WARNING] 前端服务可能未在运行
)

REM 清理相关进程
echo [INFO] 清理相关进程...

REM 清理Go进程
tasklist | findstr "go.exe" >nul
if not errorlevel 1 (
    echo [INFO] 清理Go开发进程...
    taskkill /IM go.exe /F >nul 2>&1
)

REM 清理Node进程 (谨慎处理，只清理特定的)
for /f "tokens=2" %%a in ('tasklist ^| findstr "node.exe"') do (
    REM 这里可以添加更精确的进程识别逻辑
    REM 暂时跳过，避免误杀其他Node应用
)

REM 清理PID文件
if exist "%LOG_DIR%\backend.pid" del "%LOG_DIR%\backend.pid"
if exist "%LOG_DIR%\frontend.pid" del "%LOG_DIR%\frontend.pid"

timeout /t 2 >nul

REM 检查最终状态
echo.
echo [INFO] 检查最终状态...

netstat -ano | findstr ":%BACKEND_PORT%" | findstr "LISTENING" >nul
if not errorlevel 1 (
    echo [WARNING] 后端端口 %BACKEND_PORT% 仍被占用
) else (
    echo [SUCCESS] 后端端口 %BACKEND_PORT% 已释放
)

netstat -ano | findstr ":%FRONTEND_PORT%" | findstr "LISTENING" >nul
if not errorlevel 1 (
    echo [WARNING] 前端端口 %FRONTEND_PORT% 仍被占用
) else (
    echo [SUCCESS] 前端端口 %FRONTEND_PORT% 已释放
)

echo.
echo [SUCCESS] Stellar 开发环境已停止
echo.
pause