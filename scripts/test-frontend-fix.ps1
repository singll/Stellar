# 前端修复验证测试脚本 (PowerShell版本)

param(
    [string]$FrontendUrl = "http://localhost:5173"
)

# 日志函数
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# 检查前端服务状态
function Test-FrontendService {
    Write-Info "检查前端服务状态..."
    
    # 等待服务启动
    $maxAttempts = 30
    $attempt = 1
    
    while ($attempt -le $maxAttempts) {
        try {
            $response = Invoke-WebRequest -Uri $FrontendUrl -Method Get -ErrorAction Stop
            if ($response.StatusCode -eq 200) {
                Write-Success "前端服务运行正常"
                return $true
            }
        }
        catch {
            Write-Info "等待前端服务启动... (尝试 $attempt/$maxAttempts)"
            Start-Sleep -Seconds 2
            $attempt++
        }
    }
    
    Write-Error "前端服务启动超时"
    return $false
}

# 测试页面访问
function Test-PageAccess {
    Write-Info "测试页面访问..."
    
    # 测试主页面
    try {
        $mainResponse = Invoke-WebRequest -Uri $FrontendUrl -Method Get -ErrorAction Stop
        if ($mainResponse.StatusCode -eq 200) {
            Write-Success "主页面可访问"
        }
        else {
            Write-Error "主页面访问失败: $($mainResponse.StatusCode)"
            return $false
        }
    }
    catch {
        Write-Error "主页面访问失败: $_"
        return $false
    }
    
    # 测试登录页面
    try {
        $loginResponse = Invoke-WebRequest -Uri "$FrontendUrl/login" -Method Get -ErrorAction Stop
        if ($loginResponse.StatusCode -eq 200) {
            Write-Success "登录页面可访问"
        }
        else {
            Write-Error "登录页面访问失败: $($loginResponse.StatusCode)"
            return $false
        }
    }
    catch {
        Write-Error "登录页面访问失败: $_"
        return $false
    }
    
    # 测试简单测试页面
    try {
        $testResponse = Invoke-WebRequest -Uri "$FrontendUrl/test-simple" -Method Get -ErrorAction Stop
        if ($testResponse.StatusCode -eq 200) {
            Write-Success "简单测试页面可访问"
        }
        else {
            Write-Warning "简单测试页面访问失败: $($testResponse.StatusCode)"
        }
    }
    catch {
        Write-Warning "简单测试页面访问失败: $_"
    }
    
    # 测试认证测试页面
    try {
        $authTestResponse = Invoke-WebRequest -Uri "$FrontendUrl/test-auth" -Method Get -ErrorAction Stop
        if ($authTestResponse.StatusCode -eq 200) {
            Write-Success "认证测试页面可访问"
        }
        else {
            Write-Warning "认证测试页面访问失败: $($authTestResponse.StatusCode)"
        }
    }
    catch {
        Write-Warning "认证测试页面访问失败: $_"
    }
    
    return $true
}

# 检查页面内容
function Test-PageContent {
    Write-Info "检查页面内容..."
    
    # 检查主页面是否包含关键内容
    try {
        $mainContent = Invoke-WebRequest -Uri $FrontendUrl -Method Get -ErrorAction Stop
        if ($mainContent.Content -match "Stellar") {
            Write-Success "主页面内容正常"
        }
        else {
            Write-Error "主页面内容异常"
            return $false
        }
    }
    catch {
        Write-Error "主页面内容检查失败: $_"
        return $false
    }
    
    # 检查登录页面是否包含关键内容
    try {
        $loginContent = Invoke-WebRequest -Uri "$FrontendUrl/login" -Method Get -ErrorAction Stop
        if ($loginContent.Content -match "登录") {
            Write-Success "登录页面内容正常"
        }
        else {
            Write-Error "登录页面内容异常"
            return $false
        }
    }
    catch {
        Write-Error "登录页面内容检查失败: $_"
        return $false
    }
    
    return $true
}

# 主测试流程
function Start-FrontendFixTest {
    Write-Info "开始前端修复验证测试..."
    
    # 检查前端服务状态
    if (-not (Test-FrontendService)) {
        return
    }
    
    # 测试页面访问
    if (-not (Test-PageAccess)) {
        return
    }
    
    # 检查页面内容
    if (-not (Test-PageContent)) {
        return
    }
    
    Write-Success "前端修复验证测试完成！"
    Write-Info "请手动访问以下页面进行进一步测试："
    Write-Info "1. 主页面: $FrontendUrl"
    Write-Info "2. 登录页面: $FrontendUrl/login"
    Write-Info "3. 简单测试页面: $FrontendUrl/test-simple"
    Write-Info "4. 认证测试页面: $FrontendUrl/test-auth"
}

# 运行主测试
Start-FrontendFixTest 