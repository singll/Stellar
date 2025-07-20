# 会话管理功能测试脚本 (PowerShell版本)
# 测试Redis会话管理、状态检查、刷新和删除功能

param(
    [string]$BaseUrl = "http://localhost:8082"
)

# 配置
$ApiBase = "$BaseUrl/api/v1"

# 测试用户信息
$TestUsername = "session_test_user"
$TestEmail = "session_test@example.com"
$TestPassword = "test123456"

# 全局变量
$Script:Token = ""
$Script:UserId = ""

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

# 检查服务状态
function Test-ServiceStatus {
    Write-Info "检查服务状态..."
    
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/health" -Method Get -ErrorAction Stop
        Write-Success "后端服务运行正常"
    }
    catch {
        Write-Error "后端服务未运行，请先启动服务"
        exit 1
    }
}

# 清理测试用户
function Remove-TestUser {
    Write-Info "清理测试用户..."
    
    if ($Script:UserId) {
        Write-Info "测试用户ID: $($Script:UserId)"
        # 这里可以添加删除用户的API调用
    }
}

# 注册测试用户
function Register-TestUser {
    Write-Info "注册测试用户..."
    
    $body = @{
        username = $TestUsername
        email = $TestEmail
        password = $TestPassword
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$ApiBase/auth/register" -Method Post -Body $body -ContentType "application/json"
        
        if ($response.code -eq 200) {
            Write-Success "测试用户注册成功"
            $Script:Token = $response.data.token
            $Script:UserId = $response.data.user.id
        }
        else {
            throw "注册失败: $($response.message)"
        }
    }
    catch {
        Write-Warning "测试用户可能已存在，尝试登录..."
        Login-TestUser
    }
}

# 登录测试用户
function Login-TestUser {
    Write-Info "登录测试用户..."
    
    $body = @{
        username = $TestUsername
        password = $TestPassword
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$ApiBase/auth/login" -Method Post -Body $body -ContentType "application/json"
        
        if ($response.code -eq 200) {
            Write-Success "测试用户登录成功"
            $Script:Token = $response.data.token
            $Script:UserId = $response.data.user.id
        }
        else {
            throw "登录失败: $($response.message)"
        }
    }
    catch {
        Write-Error "登录失败: $_"
        exit 1
    }
}

# 测试会话验证
function Test-SessionVerification {
    Write-Info "测试会话验证..."
    
    $headers = @{
        "Authorization" = "Bearer $Script:Token"
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$ApiBase/auth/verify" -Method Get -Headers $headers
        
        if ($response.code -eq 200 -and $response.valid -eq $true) {
            Write-Success "会话验证成功"
            
            # 检查会话状态信息
            if ($response.session_status) {
                Write-Success "会话状态信息完整"
                $username = $response.session_status.username
                $isExpired = $response.session_status.is_expired
                $needsRefresh = $response.session_status.needs_refresh
                
                Write-Info "会话信息: 用户=$username, 过期=$isExpired, 需刷新=$needsRefresh"
            }
            else {
                Write-Warning "未返回会话状态信息（可能Redis未配置）"
            }
        }
        else {
            throw "会话验证失败: $($response.message)"
        }
    }
    catch {
        Write-Error "会话验证失败: $_"
        return $false
    }
    
    return $true
}

# 测试获取会话状态
function Test-GetSessionStatus {
    Write-Info "测试获取会话状态..."
    
    $headers = @{
        "Authorization" = "Bearer $Script:Token"
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$ApiBase/auth/session/status" -Method Get -Headers $headers
        
        if ($response.code -eq 200) {
            Write-Success "获取会话状态成功"
            
            if ($response.data) {
                $username = $response.data.username
                $timeUntilExpiry = $response.data.time_until_expiry
                $isExpired = $response.data.is_expired
                
                Write-Info "会话状态: 用户=$username, 剩余时间=${timeUntilExpiry}秒, 过期=$isExpired"
            }
            else {
                Write-Warning "未返回会话数据（可能Redis未配置）"
            }
        }
        else {
            throw "获取会话状态失败: $($response.message)"
        }
    }
    catch {
        Write-Error "获取会话状态失败: $_"
        return $false
    }
    
    return $true
}

# 测试会话刷新
function Test-SessionRefresh {
    Write-Info "测试会话刷新..."
    
    $headers = @{
        "Authorization" = "Bearer $Script:Token"
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$ApiBase/auth/session/refresh" -Method Post -Headers $headers
        
        if ($response.code -eq 200) {
            Write-Success "会话刷新成功"
            
            if ($response.data) {
                $lastUsed = $response.data.last_used
                Write-Info "会话刷新时间: $lastUsed"
            }
        }
        else {
            Write-Warning "会话刷新失败或Redis未配置: $($response.message)"
        }
    }
    catch {
        Write-Warning "会话刷新失败或Redis未配置: $_"
    }
}

# 测试重新登录（会话替换）
function Test-Relogin {
    Write-Info "测试重新登录（会话替换）..."
    
    $body = @{
        username = $TestUsername
        password = $TestPassword
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$ApiBase/auth/login" -Method Post -Body $body -ContentType "application/json"
        
        if ($response.code -eq 200) {
            Write-Success "重新登录成功"
            $newToken = $response.data.token
            
            # 验证新token
            $verifyHeaders = @{
                "Authorization" = "Bearer $newToken"
            }
            
            $verifyResponse = Invoke-RestMethod -Uri "$ApiBase/auth/verify" -Method Get -Headers $verifyHeaders
            
            if ($verifyResponse.code -eq 200 -and $verifyResponse.valid -eq $true) {
                Write-Success "新会话验证成功"
                $Script:Token = $newToken
            }
            else {
                throw "新会话验证失败"
            }
        }
        else {
            throw "重新登录失败: $($response.message)"
        }
    }
    catch {
        Write-Error "重新登录失败: $_"
        return $false
    }
    
    return $true
}

# 测试登出（会话删除）
function Test-Logout {
    Write-Info "测试登出（会话删除）..."
    
    $headers = @{
        "Authorization" = "Bearer $Script:Token"
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$ApiBase/auth/logout" -Method Post -Headers $headers
        
        if ($response.code -eq 200) {
            Write-Success "登出成功"
            
            # 验证会话是否已删除
            try {
                $verifyResponse = Invoke-RestMethod -Uri "$ApiBase/auth/verify" -Method Get -Headers $headers
                
                if ($verifyResponse.code -eq 401) {
                    Write-Success "会话已成功删除"
                }
                else {
                    Write-Warning "会话可能未完全删除（JWT仍有效）"
                }
            }
            catch {
                Write-Success "会话已成功删除"
            }
        }
        else {
            throw "登出失败: $($response.message)"
        }
    }
    catch {
        Write-Error "登出失败: $_"
        return $false
    }
    
    return $true
}

# 测试无效token
function Test-InvalidToken {
    Write-Info "测试无效token..."
    
    $headers = @{
        "Authorization" = "Bearer invalid_token_123"
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$ApiBase/auth/verify" -Method Get -Headers $headers
        Write-Error "无效token未被正确拒绝"
        return $false
    }
    catch {
        Write-Success "无效token正确被拒绝"
        return $true
    }
}

# 主测试流程
function Start-SessionManagementTest {
    Write-Info "开始会话管理功能测试..."
    
    # 检查服务状态
    Test-ServiceStatus
    
    # 注册/登录测试用户
    Register-TestUser
    
    # 测试会话验证
    Test-SessionVerification
    
    # 测试获取会话状态
    Test-GetSessionStatus
    
    # 测试会话刷新
    Test-SessionRefresh
    
    # 测试重新登录（会话替换）
    Test-Relogin
    
    # 再次验证会话
    Test-SessionVerification
    
    # 测试登出（会话删除）
    Test-Logout
    
    # 测试无效token
    Test-InvalidToken
    
    Write-Success "所有会话管理功能测试完成！"
}

# 清理函数
function Invoke-Cleanup {
    Write-Info "清理测试环境..."
    Remove-TestUser
}

# 设置退出时清理
try {
    # 运行主测试
    Start-SessionManagementTest
}
finally {
    # 清理测试环境
    Invoke-Cleanup
} 