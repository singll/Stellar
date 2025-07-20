# 认证功能增强文档

## 概述

本次更新对Stellar项目的登录和验证功能进行了全面优化，主要实现了以下功能：

1. **自动跳转优化**：登录后如果没有指定重定向路由，自动跳转到dashboard
2. **Redis会话管理**：使用Redis保持登录状态，默认8小时有效时间，使用时自动刷新
3. **智能认证检查**：已登录用户不需要再次登录，避免重复跳转到登录界面

## 功能特性

### 1. 自动跳转优化

#### 后端实现
- 登录成功后返回JWT令牌和用户信息
- 支持Redis会话存储（如果Redis可用）

#### 前端实现
- 登录成功后检查URL参数中的`redirect`参数
- 如果没有重定向参数，默认跳转到`/dashboard`
- 已登录用户访问登录页会自动跳转到dashboard

```typescript
// 登录成功后的跳转逻辑
const urlParams = new URLSearchParams(window.location.search);
const redirectUrl = urlParams.get('redirect');
const targetUrl = redirectUrl || '/dashboard';
await goto(targetUrl);
```

### 2. Redis会话管理

#### 会话管理器 (`internal/services/session/session.go`)

**核心特性：**
- **8小时有效期**：会话默认有效期为8小时
- **自动刷新**：距离过期时间少于1小时时自动刷新
- **Redis存储**：使用Redis存储会话数据，支持分布式部署
- **会话验证**：每次访问时验证会话有效性

**主要方法：**
```go
// 创建新会话
func (sm *SessionManager) CreateSession(ctx context.Context, token string, userID, username string, roles []string) error

// 获取会话数据（自动刷新）
func (sm *SessionManager) GetSession(ctx context.Context, token string) (*SessionData, error)

// 删除会话
func (sm *SessionManager) DeleteSession(ctx context.Context, token string) error

// 验证会话有效性
func (sm *SessionManager) IsSessionValid(ctx context.Context, token string) bool
```

#### 会话数据结构
```go
type SessionData struct {
    UserID    string    `json:"user_id"`
    Username  string    `json:"username"`
    Roles     []string  `json:"roles"`
    Token     string    `json:"token"`
    CreatedAt time.Time `json:"created_at"`
    LastUsed  time.Time `json:"last_used"`
}
```

### 3. 智能认证检查

#### 认证守卫优化 (`web/src/lib/guards/auth.guard.ts`)

**改进逻辑：**
1. 首先检查用户是否已认证
2. 如果已认证且访问登录页，自动跳转到dashboard
3. 如果未认证且访问受保护页面，重定向到登录页并记录原路径

```typescript
// 认证路由守卫
export const authGuard: Handle = async ({ event, resolve }) => {
    const path = event.url.pathname;
    const { isAuthenticated, token } = auth.state;

    // 如果已认证且访问登录页，重定向到dashboard
    if (isAuthenticated && token && path === '/login') {
        throw redirect(303, '/dashboard');
    }

    // 如果是公开路由，直接放行
    if (isPublicRoute(path)) {
        return resolve(event);
    }

    // 如果未认证且不是公开路由，重定向到登录页
    if (!isAuthenticated || !token) {
        throw redirect(303, `/login?redirect=${encodeURIComponent(path)}`);
    }

    return resolve(event);
};
```

## 技术实现

### 后端架构

#### 1. 会话中间件 (`internal/api/middleware/session.go`)
```go
// 会话中间件，注入会话管理器到请求上下文
func SessionMiddleware(sessionManager *session.SessionManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        if sessionManager != nil {
            c.Set("session_manager", sessionManager)
        }
        c.Next()
    }
}
```

#### 2. 认证中间件增强 (`internal/api/auth.go`)
- 支持JWT令牌验证
- 集成Redis会话验证
- 自动刷新会话（如果Redis可用）

#### 3. 路由管理器更新 (`internal/api/router/manager.go`)
- 为认证路由组添加会话中间件
- 为公开路由组也添加会话中间件（支持登出时删除会话）

### 前端架构

#### 1. 认证Store增强 (`web/src/lib/stores/auth.ts`)
```typescript
// 会话验证功能
async verifySession() {
    try {
        const response = await authApi.verifySession();
        if (response.code === 200 && response.valid) {
            // 会话有效，更新用户信息
            if (response.user && auth.state.user) {
                store.update((state) => ({
                    ...state,
                    user: {
                        ...state.user!,
                        username: response.user!.username,
                        roles: response.user!.roles,
                    },
                }));
            }
            return true;
        }
        return false;
    } catch (error) {
        console.error('会话验证失败:', error);
        return false;
    }
}
```

#### 2. API客户端增强 (`web/src/lib/api/auth.ts`)
- 添加会话验证API
- 支持会话状态检查

#### 3. 路由守卫优化
- 智能重定向逻辑
- 避免重复登录

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

## 测试

### 自动化测试脚本
使用提供的测试脚本验证功能：
```bash
chmod +x scripts/test-auth-enhancement.sh
./scripts/test-auth-enhancement.sh
```

### 手动测试步骤

1. **登录测试**
   - 访问登录页面
   - 输入正确的用户名和密码
   - 验证自动跳转到dashboard

2. **会话持久性测试**
   - 登录后关闭浏览器
   - 重新打开浏览器访问应用
   - 验证无需重新登录

3. **已登录用户测试**
   - 在已登录状态下访问登录页
   - 验证自动跳转到dashboard

4. **会话过期测试**
   - 等待会话过期（或手动删除Redis中的会话）
   - 验证需要重新登录

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

## 安全考虑

### 会话安全
- 会话数据在Redis中加密存储
- 支持会话撤销（登出时删除）
- 自动刷新防止会话劫持

### JWT安全
- 使用强密钥
- 合理的过期时间
- 支持令牌撤销

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

3. **前端跳转问题**
   - 检查认证状态
   - 验证路由守卫配置
   - 查看浏览器控制台错误

### 调试方法

1. **后端调试**
   ```bash
   # 查看Redis会话数据
   redis-cli keys "session:*"
   redis-cli get "session:your_token"
   ```

2. **前端调试**
   ```javascript
   // 检查认证状态
   console.log(auth.state);
   
   // 验证会话
   auth.verifySession().then(console.log);
   ```

## 总结

本次认证功能增强显著提升了用户体验：

1. **用户体验优化**：自动跳转和智能认证检查
2. **安全性提升**：Redis会话管理和自动刷新
3. **可扩展性**：支持分布式部署和会话共享
4. **维护性**：清晰的代码结构和完善的文档

这些改进使Stellar项目的认证系统更加健壮和用户友好。 