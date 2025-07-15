# start-dev.ps1
# 星络（Stellar）本地开发环境一键启动脚本（适用于 Win11 + PowerShell 7）

Write-Host "==== 星络（Stellar）开发环境启动 ====" -ForegroundColor Cyan

# 检查Go环境
$goVersion = go version 2>$null
if (-not $goVersion) {
    Write-Error "未检测到Go环境，请先安装Go 1.21+"
    exit 1
}
Write-Host "Go环境: $goVersion"

# 检查Node.js和pnpm
$nodeVersion = node -v 2>$null
$pnpmVersion = pnpm -v 2>$null
if (-not $nodeVersion) {
    Write-Error "未检测到Node.js，请先安装Node.js 20+"
    exit 1
}
if (-not $pnpmVersion) {
    Write-Error "未检测到pnpm，请先全局安装pnpm 8+（npm install -g pnpm）"
    exit 1
}
Write-Host "Node.js: $nodeVersion"
Write-Host "pnpm: $pnpmVersion"

# 启动后端服务
Write-Host "`n[1/2] 启动后端服务..." -ForegroundColor Yellow
Start-Process -NoNewWindow -FilePath "pwsh" -ArgumentList "-NoExit", "-Command", "cd $PSScriptRoot; go run cmd/main.go -config config.yaml" 

# 启动前端服务
Write-Host "`n[2/2] 启动前端开发服务器..." -ForegroundColor Yellow
Start-Process -NoNewWindow -FilePath "pwsh" -ArgumentList "-NoExit", "-Command", "cd $PSScriptRoot\web; pnpm install; pnpm dev"

Write-Host "`n==== 所有服务已启动，请在浏览器访问 http://localhost:5173 ====" -ForegroundColor Green
Write-Host "如需停止服务，请手动关闭对应终端窗口。"