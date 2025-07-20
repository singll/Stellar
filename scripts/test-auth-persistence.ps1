# 认证状态持久化测试脚本 (PowerShell版本)

param(
    [string]$BaseUrl = "http://localhost:8082",
    [string]$FrontendUrl = "http://localhost:5173"
)

# 配置
$ApiBase = "$BaseUrl/api/v1"

# 测试用户信息
$TestUsername = "persistence_test_user"
$TestEmail = "persistence_test@example.com"
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
function Test-Services {
    Write-Info "检查服务状态..."
    
    # 检查后端服务
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/health" -Method Get -ErrorAction Stop
        Write-Success "后端服务运行正常"
    }
    catch {
        Write-Error "后端服务未运行，请先启动服务"
        exit 1
    }
    
    # 检查前端服务
    try {
        $response = Invoke-RestMethod -Uri "$FrontendUrl" -Method Get -ErrorAction Stop
        Write-Success "前端服务运行正常"
    }
    catch {
        Write-Warning "前端服务未运行，请先启动前端开发服务器"
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
            return $true
        }
        else {
            throw "会话验证失败: $($response.message)"
        }
    }
    catch {
        Write-Error "会话验证失败: $_"
        return $false
    }
}

# 测试前端页面访问
function Test-FrontendAccess {
    Write-Info "测试前端页面访问..."
    
    # 测试访问主页面
    try {
        $response = Invoke-WebRequest -Uri "$FrontendUrl" -Method Get -ErrorAction Stop
        if ($response.StatusCode -eq 200) {
            Write-Success "前端主页面可访问"
        }
        else {
            Write-Warning "前端主页面访问失败: $($response.StatusCode)"
        }
    }
    catch {
        Write-Warning "前端主页面访问失败: $_"
    }
    
    # 测试访问测试页面
    try {
        $testResponse = Invoke-WebRequest -Uri "$FrontendUrl/test-auth" -Method Get -ErrorAction Stop
        if ($testResponse.StatusCode -eq 200) {
            Write-Success "认证测试页面可访问"
        }
        else {
            Write-Warning "认证测试页面访问失败: $($testResponse.StatusCode)"
        }
    }
    catch {
        Write-Warning "认证测试页面访问失败: $_"
    }
}

# 测试localStorage持久化
function Test-LocalStoragePersistence {
    Write-Info "测试localStorage持久化..."
    
    Write-Warning "localStorage持久化测试需要浏览器自动化，请手动测试"
    Write-Info "手动测试步骤："
    Write-Info "1. 访问 $FrontendUrl/login"
    Write-Info "2. 使用测试账户登录: $TestUsername / $TestPassword"
    Write-Info "3. 登录成功后访问 $FrontendUrl/test-auth"
    Write-Info "4. 刷新页面，检查认证状态是否保持"
    Write-Info "5. 关闭浏览器，重新打开访问 $FrontendUrl"
    Write-Info "6. 检查是否自动跳转到dashboard"
}

# 测试会话状态API
function Test-SessionApis {
    Write-Info "测试会话状态API..."
    
    $headers = @{
        "Authorization" = "Bearer $Script:Token"
    }
    
    # 测试获取会话状态
    try {
        $statusResponse = Invoke-RestMethod -Uri "$ApiBase/auth/session/status" -Method Get -Headers $headers
        
        if ($statusResponse.code -eq 200) {
            Write-Success "获取会话状态成功"
            $username = $statusResponse.data.username
            $timeUntilExpiry = $statusResponse.data.time_until_expiry
            Write-Info "会话信息: 用户=$username, 剩余时间=${timeUntilExpiry}秒"
        }
        else {
            throw "获取会话状态失败: $($statusResponse.message)"
        }
    }
    catch {
        Write-Error "获取会话状态失败: $_"
    }
    
    # 测试刷新会话
    try {
        $refreshResponse = Invoke-RestMethod -Uri "$ApiBase/auth/session/refresh" -Method Post -Headers $headers
        
        if ($refreshResponse.code -eq 200) {
            Write-Success "会话刷新成功"
        }
        else {
            Write-Warning "会话刷新失败或Redis未配置: $($refreshResponse.message)"
        }
    }
    catch {
        Write-Warning "会话刷新失败或Redis未配置: $_"
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

# 主测试流程
function Start-AuthPersistenceTest {
    Write-Info "开始认证状态持久化测试..."
    
    # 检查服务状态
    Test-Services
    
    # 注册/登录测试用户
    Register-TestUser
    
    # 测试会话验证
    Test-SessionVerification
    
    # 测试会话状态API
    Test-SessionApis
    
    # 测试前端页面访问
    Test-FrontendAccess
    
    # 测试localStorage持久化
    Test-LocalStoragePersistence
    
    Write-Success "认证状态持久化测试完成！"
    Write-Info "请手动测试前端页面的认证状态持久化功能"
}

# 清理函数
function Invoke-Cleanup {
    Write-Info "清理测试环境..."
    Remove-TestUser
}

# 设置退出时清理
try {
    # 运行主测试
    Start-AuthPersistenceTest
}
finally {
    # 清理测试环境
    Invoke-Cleanup
} 