# 测试修改密码功能
# 这个脚本测试以下功能：
# 1. 用户登录
# 2. 修改密码
# 3. 使用新密码登录
# 4. 验证旧密码失效

Write-Host "🔐 开始测试修改密码功能..." -ForegroundColor Cyan

# 设置测试环境
$BASE_URL = "http://localhost:8090"
$API_BASE = "$BASE_URL/api/v1"

# 测试用户信息
$TEST_USER = "testuser"
$TEST_EMAIL = "testuser@example.com"
$OLD_PASSWORD = "TestPassword123"
$NEW_PASSWORD = "NewPassword456!"

Write-Host "📋 测试用户信息:" -ForegroundColor Yellow
Write-Host "  用户名: $TEST_USER"
Write-Host "  邮箱: $TEST_EMAIL"
Write-Host "  原密码: $OLD_PASSWORD"
Write-Host "  新密码: $NEW_PASSWORD"
Write-Host "  API地址: $API_BASE"

# 颜色输出函数
function Write-Success {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor Red
}

function Write-Info {
    param([string]$Message)
    Write-Host "ℹ️  $Message" -ForegroundColor Blue
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠️  $Message" -ForegroundColor Yellow
}

# 全局变量
$script:TOKEN = $null

# 检查服务是否运行
function Test-ServiceStatus {
    Write-Info "检查服务状态..."
    try {
        $response = Invoke-RestMethod -Uri "$BASE_URL/health" -Method Get -TimeoutSec 5
        Write-Success "服务正在运行"
        return $true
    }
    catch {
        Write-Error "服务未运行，请先启动服务"
        return $false
    }
}

# 注册测试用户
function Register-TestUser {
    Write-Info "注册测试用户..."
    $body = @{
        username = $TEST_USER
        email = $TEST_EMAIL
        password = $OLD_PASSWORD
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$API_BASE/auth/register" -Method Post -Body $body -ContentType "application/json"
        if ($response.code -eq 200) {
            Write-Success "用户注册成功"
            return $true
        }
    }
    catch {
        Write-Warning "用户可能已存在或注册失败: $($_.Exception.Message)"
        return $true  # 继续测试，用户可能已存在
    }
}

# 测试登录
function Test-Login {
    param([string]$Password)
    Write-Info "测试用户登录 (密码: $Password)..."
    $body = @{
        username = $TEST_USER
        password = $Password
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$API_BASE/auth/login" -Method Post -Body $body -ContentType "application/json"
        if ($response.code -eq 200) {
            Write-Success "登录成功"
            $script:TOKEN = $response.data.token
            if ($TOKEN) {
                Write-Success "获取到JWT令牌"
                return $true
            } else {
                Write-Error "未获取到JWT令牌"
                return $false
            }
        } else {
            Write-Error "登录失败: $($response.message)"
            return $false
        }
    }
    catch {
        Write-Error "登录失败: $($_.Exception.Message)"
        return $false
    }
}

# 测试修改密码
function Test-ChangePassword {
    Write-Info "测试修改密码..."
    if (-not $TOKEN) {
        Write-Error "没有可用的令牌进行测试"
        return $false
    }

    $body = @{
        oldPassword = $OLD_PASSWORD
        newPassword = $NEW_PASSWORD
    } | ConvertTo-Json

    try {
        $headers = @{
            "Authorization" = "Bearer $TOKEN"
            "Content-Type" = "application/json"
        }
        $response = Invoke-RestMethod -Uri "$API_BASE/auth/password" -Method Put -Body $body -Headers $headers
        if ($response.code -eq 200) {
            Write-Success "密码修改成功"
            return $true
        } else {
            Write-Error "密码修改失败: $($response.message)"
            return $false
        }
    }
    catch {
        Write-Error "密码修改失败: $($_.Exception.Message)"
        return $false
    }
}

# 测试使用新密码登录
function Test-NewPasswordLogin {
    Write-Info "测试使用新密码登录..."
    $body = @{
        username = $TEST_USER
        password = $NEW_PASSWORD
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$API_BASE/auth/login" -Method Post -Body $body -ContentType "application/json"
        if ($response.code -eq 200) {
            Write-Success "新密码登录成功"
            $script:TOKEN = $response.data.token
            return $true
        } else {
            Write-Error "新密码登录失败: $($response.message)"
            return $false
        }
    }
    catch {
        Write-Error "新密码登录失败: $($_.Exception.Message)"
        return $false
    }
}

# 测试旧密码失效
function Test-OldPasswordInvalid {
    Write-Info "测试旧密码失效..."
    $body = @{
        username = $TEST_USER
        password = $OLD_PASSWORD
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$API_BASE/auth/login" -Method Post -Body $body -ContentType "application/json"
        if ($response.code -eq 401) {
            Write-Success "旧密码已失效（符合预期）"
            return $true
        } else {
            Write-Error "旧密码仍然有效（不符合预期）: $($response.message)"
            return $false
        }
    }
    catch {
        if ($_.Exception.Response.StatusCode -eq 401) {
            Write-Success "旧密码已失效（符合预期）"
            return $true
        } else {
            Write-Error "测试旧密码失效时发生错误: $($_.Exception.Message)"
            return $false
        }
    }
}

# 测试密码验证错误
function Test-PasswordValidation {
    Write-Info "测试密码验证错误..."
    if (-not $TOKEN) {
        Write-Error "没有可用的令牌进行测试"
        return $false
    }

    # 测试错误原密码
    $body = @{
        oldPassword = "WrongPassword123"
        newPassword = "AnotherPassword456!"
    } | ConvertTo-Json

    try {
        $headers = @{
            "Authorization" = "Bearer $TOKEN"
            "Content-Type" = "application/json"
        }
        $response = Invoke-RestMethod -Uri "$API_BASE/auth/password" -Method Put -Body $body -Headers $headers
        if ($response.code -eq 400) {
            Write-Success "错误原密码验证正确"
        } else {
            Write-Error "错误原密码验证失败: $($response.message)"
            return $false
        }
    }
    catch {
        if ($_.Exception.Response.StatusCode -eq 400) {
            Write-Success "错误原密码验证正确"
        } else {
            Write-Error "测试错误原密码时发生错误: $($_.Exception.Message)"
            return $false
        }
    }

    # 测试短密码
    $body = @{
        oldPassword = $NEW_PASSWORD
        newPassword = "123"
    } | ConvertTo-Json

    try {
        $headers = @{
            "Authorization" = "Bearer $TOKEN"
            "Content-Type" = "application/json"
        }
        $response = Invoke-RestMethod -Uri "$API_BASE/auth/password" -Method Put -Body $body -Headers $headers
        if ($response.code -eq 400) {
            Write-Success "短密码验证正确"
            return $true
        } else {
            Write-Error "短密码验证失败: $($response.message)"
            return $false
        }
    }
    catch {
        if ($_.Exception.Response.StatusCode -eq 400) {
            Write-Success "短密码验证正确"
            return $true
        } else {
            Write-Error "测试短密码时发生错误: $($_.Exception.Message)"
            return $false
        }
    }
}

# 测试未认证访问
function Test-UnauthenticatedAccess {
    Write-Info "测试未认证访问..."
    $body = @{
        oldPassword = $NEW_PASSWORD
        newPassword = "AnotherPassword789!"
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$API_BASE/auth/password" -Method Put -Body $body -ContentType "application/json"
        Write-Error "未认证访问应该被拒绝"
        return $false
    }
    catch {
        if ($_.Exception.Response.StatusCode -eq 401) {
            Write-Success "未认证访问被正确拒绝"
            return $true
        } else {
            Write-Error "未认证访问测试失败: $($_.Exception.Message)"
            return $false
        }
    }
}

# 恢复原密码（清理测试）
function Restore-OriginalPassword {
    Write-Info "恢复原密码..."
    if (-not $TOKEN) {
        Write-Warning "没有可用的令牌，跳过恢复原密码"
        return $true
    }

    $body = @{
        oldPassword = $NEW_PASSWORD
        newPassword = $OLD_PASSWORD
    } | ConvertTo-Json

    try {
        $headers = @{
            "Authorization" = "Bearer $TOKEN"
            "Content-Type" = "application/json"
        }
        $response = Invoke-RestMethod -Uri "$API_BASE/auth/password" -Method Put -Body $body -Headers $headers
        if ($response.code -eq 200) {
            Write-Success "原密码恢复成功"
            return $true
        } else {
            Write-Warning "原密码恢复失败: $($response.message)"
            return $false
        }
    }
    catch {
        Write-Warning "原密码恢复失败: $($_.Exception.Message)"
        return $false
    }
}

# 主测试流程
function Start-Test {
    Write-Host "`n🚀 开始执行修改密码功能测试..." -ForegroundColor Green

    # 检查服务状态
    if (-not (Test-ServiceStatus)) {
        return
    }

    # 注册用户
    if (-not (Register-TestUser)) {
        return
    }

    # 使用原密码登录
    if (-not (Test-Login -Password $OLD_PASSWORD)) {
        return
    }

    # 测试修改密码
    if (-not (Test-ChangePassword)) {
        return
    }

    # 测试使用新密码登录
    if (-not (Test-NewPasswordLogin)) {
        return
    }

    # 测试旧密码失效
    if (-not (Test-OldPasswordInvalid)) {
        return
    }

    # 测试密码验证错误
    if (-not (Test-PasswordValidation)) {
        return
    }

    # 测试未认证访问
    if (-not (Test-UnauthenticatedAccess)) {
        return
    }

    # 恢复原密码
    Restore-OriginalPassword

    Write-Host "`n🎉 所有测试完成！" -ForegroundColor Green
    Write-Host "✅ 修改密码功能测试通过" -ForegroundColor Green
}

# 执行测试
Start-Test 