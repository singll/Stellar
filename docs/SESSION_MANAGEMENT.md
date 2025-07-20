# 会话管理功能文档

## 概述

Stellar项目的会话管理功能基于Redis实现，提供了完整的登录状态管理、会话验证、状态检查和自动刷新功能。该功能解决了前端刷新后跳转到登录页的问题，提升了用户体验。

## 功能特性

### 1. Redis会话管理
- **8小时有效期**：会话默认有效期为8小时
- **自动刷新**：距离过期时间少于1小时时自动刷新
- **会话替换**：重新登录时自动删除旧会话，创建新会话
- **分布式支持**：支持多实例部署和会话共享

### 2. 会话状态检查
- **实时验证**：每次访问时验证会话有效性
- **状态信息**：提供详细的会话状态信息
- **过期检测**：自动检测会话是否过期
- **刷新提醒**：提示是否需要刷新会话

### 3. 前端状态管理
- **自动恢复**：刷新页面后自动恢复认证状态
- **智能跳转**：已登录用户不会看到登录页面
- **状态同步**：前后端状态保持同步
- **错误处理**：完善的错误处理和用户提示

## 技术实现

### 后端架构

#### 会话管理器 (`internal/services/session/session.go`)

```go
type SessionManager struct {
    redisClient      *redis.Client
    sessionExpiry    time.Duration    // 8小时
    refreshThreshold time.Duration    // 1小时
}
```

**核心方法：**
- `CreateSession()` - 创建新会话，支持会话替换
- `GetSession()` - 获取会话数据，自动刷新
- `DeleteSession()` - 删除会话
- `IsSessionValid()` - 验证会话有效性
- `GetSessionStatus()` - 获取详细会话状态
- `RefreshSession()` - 手动刷新会话

#### 认证API (`internal/api/auth.go`)

**新增API端点：**
- `GET /api/v1/auth/verify` - 验证会话状态
- `GET /api/v1/auth/session/status` - 获取会话状态信息
- `POST /api/v1/auth/session/refresh` - 刷新会话

#### 会话中间件 (`internal/api/middleware/session.go`)

```go
func SessionMiddleware(sessionManager *session.SessionManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        if sessionManager != nil {
            c.Set("session_manager", sessionManager)
        }
        c.Next()
    }
}
```

### 前端架构

#### 认证Store (`web/src/lib/stores/auth.ts`)

```typescript
export const auth = {
    // 初始化认证状态
    async initialize() {
        const { token, user } = auth.state;
        if (token && user) {
            const isValid = await this.verifySession();
            if (!isValid) {
                this.clearState();
            }
        }
    },
    
    // 验证会话
    async verifySession() {
        const response = await authApi.verifySession();
        if (response.valid) {
            // 检查是否需要刷新
            if (response.session_status?.needs_refresh) {
                await this.refreshSession();
            }
            return true;
        }
        return false;
    },
    
    // 刷新会话
    async refreshSession() {
        const response = await authApi.refreshSession();
        return response.code === 200;
    }
};
```

#### 路由守卫 (`web/src/lib/guards/auth.guard.ts`)

```typescript
export const authGuard: Handle = async ({ event, resolve }) => {
    const { isAuthenticated, token } = auth.state;
    
    // 如果未认证，尝试从localStorage恢复状态
    if (!isAuthenticated || !token) {
        const storedState = localStorage.getItem('auth_state');
        if (storedState) {
            const parsedState = JSON.parse(storedState);
            if (parsedState.token && parsedState.user) {
                const isValid = await auth.verifySession();
                if (isValid) {
                    return resolve(event);
                }
            }
        }
        throw redirect(303, `/login?redirect=${encodeURIComponent(path)}`);
    }
    
    return resolve(event);
};
```

## 配置要求

### Redis配置

确保Redis服务正常运行，配置示例：

```yaml
redis:
  addr: "127.0.0.1:6379"
  password: "redis"
  db: 0
  poolSize: 10
  minIdleConns: 5
  maxConnAgeMS: 30000
```

### 环境变量

```bash
# Redis配置
export REDIS_ADDR=127.0.0.1:6379
export REDIS_PASSWORD=your_redis_password

# JWT配置
export JWT_SECRET=your_jwt_secret
```

## 使用流程

### 1. 登录流程

1. 用户提交登录信息
2. 后端验证用户凭据
3. 生成JWT令牌
4. 创建Redis会话（如果Redis可用）
5. 返回令牌和用户信息
6. 前端保存认证状态

### 2. 会话验证流程

1. 前端发送请求时携带JWT令牌
2. 后端验证JWT令牌有效性
3. 如果Redis可用，验证Redis会话
4. 检查会话是否需要刷新
5. 返回验证结果和会话状态

### 3. 会话刷新流程

1. 检测到会话接近过期（少于1小时）
2. 自动更新最后使用时间
3. 重置会话过期时间
4. 返回刷新后的状态信息

### 4. 登出流程

1. 用户点击登出
2. 前端调用登出API
3. 后端删除Redis会话
4. 前端清理本地状态
5. 跳转到登录页面

## API接口

### 会话验证

```http
GET /api/v1/auth/verify
Authorization: Bearer <token>

Response:
{
  "code": 200,
  "message": "会话验证成功",
  "valid": true,
  "user": {
    "username": "admin",
    "roles": ["admin"]
  },
  "session_status": {
    "user_id": "507f1f77bcf86cd799439011",
    "username": "admin",
    "roles": ["admin"],
    "created_at": "2024-01-01T00:00:00Z",
    "last_used": "2024-01-01T07:00:00Z",
    "expires_at": "2024-01-01T08:00:00Z",
    "time_until_expiry": 3600,
    "is_expired": false,
    "needs_refresh": false
  }
}
```

### 获取会话状态

```http
GET /api/v1/auth/session/status
Authorization: Bearer <token>

Response:
{
  "code": 200,
  "message": "获取会话状态成功",
  "data": {
    "user_id": "507f1f77bcf86cd799439011",
    "username": "admin",
    "roles": ["admin"],
    "created_at": "2024-01-01T00:00:00Z",
    "last_used": "2024-01-01T07:00:00Z",
    "expires_at": "2024-01-01T08:00:00Z",
    "time_until_expiry": 3600,
    "is_expired": false,
    "needs_refresh": false
  }
}
```

### 刷新会话

```http
POST /api/v1/auth/session/refresh
Authorization: Bearer <token>

Response:
{
  "code": 200,
  "message": "会话刷新成功",
  "data": {
    "user_id": "507f1f77bcf86cd799439011",
    "username": "admin",
    "roles": ["admin"],
    "created_at": "2024-01-01T00:00:00Z",
    "last_used": "2024-01-01T07:30:00Z",
    "expires_at": "2024-01-01T08:30:00Z",
    "time_until_expiry": 3600,
    "is_expired": false,
    "needs_refresh": false
  }
}
```

## 测试

### 自动化测试

使用提供的测试脚本验证功能：

```bash
# Linux/macOS
./scripts/test-session-management.sh

# Windows PowerShell
.\scripts\test-session-management.ps1
```

### 手动测试步骤

1. **启动服务**
   ```bash
   # 启动Redis
   redis-server
   
   # 启动Stellar服务
   go run cmd/main.go -config configs/config.dev.yaml
   ```

2. **登录测试**
   - 访问登录页面
   - 输入用户名和密码
   - 验证自动跳转到dashboard

3. **会话持久性测试**
   - 登录后关闭浏览器
   - 重新打开浏览器访问应用
   - 验证无需重新登录

4. **会话刷新测试**
   - 等待会话接近过期
   - 验证自动刷新功能

5. **重新登录测试**
   - 在已登录状态下重新登录
   - 验证旧会话被替换

6. **登出测试**
   - 点击登出按钮
   - 验证会话被删除
   - 验证跳转到登录页面

## 故障排除

### 常见问题

1. **Redis连接失败**
   - 检查Redis服务状态
   - 验证连接配置
   - 查看日志错误信息

2. **会话验证失败**
   - 检查Redis中的会话数据
   - 验证JWT令牌有效性
   - 确认中间件配置正确

3. **前端刷新跳转问题**
   - 检查认证状态初始化
   - 验证路由守卫配置
   - 查看浏览器控制台错误

### 调试方法

1. **后端调试**
   ```bash
   # 查看Redis会话数据
   redis-cli keys "session:*"
   redis-cli get "session:your_token"
   redis-cli keys "user_session:*"
   ```

2. **前端调试**
   ```javascript
   // 检查认证状态
   console.log(auth.state);
   
   // 验证会话
   auth.verifySession().then(console.log);
   
   // 获取会话状态
   auth.getSessionStatus().then(console.log);
   ```

## 性能优化

### Redis优化建议

```bash
# 配置Redis最大内存
redis-cli config set maxmemory 4gb
redis-cli config set maxmemory-policy allkeys-lru

# 启用持久化
redis-cli config set save "900 1 300 10 60 10000"
```

### 会话清理

- Redis自动清理过期会话
- 可配置定期清理任务
- 支持手动清理用户会话

## 安全考虑

### 会话安全
- 会话数据在Redis中加密存储
- 支持会话撤销（登出时删除）
- 自动刷新防止会话劫持
- 重新登录时替换旧会话

### JWT安全
- 使用强密钥
- 合理的过期时间
- 支持令牌撤销
- 双重验证机制

## 总结

会话管理功能显著提升了用户体验：

1. **无缝体验**：刷新页面后认证状态自动恢复
2. **智能跳转**：已登录用户不会看到登录页面
3. **状态同步**：前后端状态保持同步
4. **安全可靠**：双重验证和自动刷新机制
5. **可扩展性**：支持分布式部署和会话共享

该功能完全解决了前端刷新跳转问题，为用户提供了流畅的认证体验。 