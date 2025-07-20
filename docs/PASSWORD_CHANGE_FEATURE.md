# 修改密码功能实现文档

## 功能概述

本文档描述了星络（Stellar）项目中修改密码功能的完整实现，包括后端API、前端界面和用户体验设计。

## 功能特性

### 核心功能
- ✅ 修改当前登录用户的密码
- ✅ 密码强度实时检测
- ✅ 密码确认验证
- ✅ 原密码验证
- ✅ 安全建议提示

### 用户体验
- ✅ 实时密码强度指示器
- ✅ 密码可见性切换
- ✅ 表单验证和错误提示
- ✅ 加载状态反馈
- ✅ 成功/失败通知

## 技术实现

### 后端API

#### 1. 路由注册
```go
// internal/api/auth.go
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
    // ... 其他路由
    router.PUT("/password", AuthMiddleware(), h.ChangePassword) // 修改密码需要认证
}
```

#### 2. 请求结构
```go
type ChangePasswordRequest struct {
    OldPassword string `json:"oldPassword" binding:"required"`
    NewPassword string `json:"newPassword" binding:"required,min=6"`
}
```

#### 3. 处理逻辑
```go
func (h *AuthHandler) ChangePassword(c *gin.Context) {
    // 1. 参数验证
    // 2. 获取当前用户
    // 3. 验证原密码
    // 4. 更新新密码
    // 5. 返回结果
}
```

#### 4. 数据库操作
```go
// internal/models/user.go
func UpdatePassword(db *mongo.Database, username, oldPassword, newPassword string) error {
    // 1. 获取用户信息
    // 2. 验证原密码
    // 3. 加密新密码
    // 4. 更新数据库
}
```

### 前端实现

#### 1. API客户端
```typescript
// web/src/lib/api/auth.ts
export const authApi = {
    async changePassword(data: { oldPassword: string; newPassword: string }): Promise<void> {
        await api.put('/auth/password', data);
    }
};
```

#### 2. 状态管理
```typescript
// web/src/lib/stores/auth.ts
export const auth = {
    async updatePassword(data: any) {
        await authApi.updatePassword(data);
        notifications.success('密码更新成功');
    }
};
```

#### 3. 页面组件
```svelte
<!-- web/src/routes/(app)/settings/security/+page.svelte -->
<script lang="ts">
    // 表单数据
    let oldPassword = '';
    let newPassword = '';
    let confirmPassword = '';
    
    // 密码强度检查
    function checkPasswordStrength(password: string) {
        // 实现密码强度算法
    }
    
    // 修改密码
    async function handleChangePassword() {
        await authApi.changePassword({
            oldPassword,
            newPassword
        });
    }
</script>
```

## 安全特性

### 1. 密码验证
- 原密码必须正确
- 新密码不能与原密码相同
- 新密码长度至少6位

### 2. 密码强度检测
- 长度检查（8位以上）
- 包含小写字母
- 包含大写字母
- 包含数字
- 包含特殊字符

### 3. 认证保护
- 需要有效的JWT令牌
- 只能修改当前登录用户的密码
- 会话验证

### 4. 数据安全
- 密码使用bcrypt加密存储
- 支持MD5密码自动升级到bcrypt
- 传输过程中使用HTTPS

## 用户界面

### 页面布局
```
┌─────────────────────────────────────────────────────────────┐
│ 安全设置                                                     │
├─────────────────────────────────────────────────────────────┤
│ ┌─────────────────┐  ┌─────────────────┐                    │
│ │   修改密码卡片   │  │   账户状态       │                    │
│ │                 │  │                 │                    │
│ │ • 当前密码      │  │ • 登录状态       │                    │
│ │ • 新密码        │  │ • 用户名         │                    │
│ │ • 确认密码      │  │ • 邮箱           │                    │
│ │ • 密码强度      │  │                 │                    │
│ │ • 提交按钮      │  └─────────────────┘                    │
│ └─────────────────┘  ┌─────────────────┐                    │
│ ┌─────────────────┐  │   快速操作       │                    │
│ │   安全建议       │  │                 │                    │
│ │                 │  │ • 返回设置       │                    │
│ │ • 使用强密码    │  │ • 返回仪表盘     │                    │
│ │ • 定期更换      │  └─────────────────┘                    │
│ │ • 不要重复使用  │                                        │
│ └─────────────────┘                                        │
└─────────────────────────────────────────────────────────────┘
```

### 交互特性
- 实时密码强度指示器
- 密码可见性切换按钮
- 表单验证提示
- 加载状态显示
- 成功/失败通知

## API接口规范

### 请求
```http
PUT /api/v1/auth/password
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
    "oldPassword": "当前密码",
    "newPassword": "新密码"
}
```

### 响应
#### 成功响应
```json
{
    "code": 200,
    "message": "密码修改成功"
}
```

#### 错误响应
```json
{
    "code": 400,
    "message": "原密码错误"
}
```

```json
{
    "code": 400,
    "message": "新密码长度不能少于6位"
}
```

```json
{
    "code": 401,
    "message": "用户信息未找到"
}
```

## 测试覆盖

### 后端测试
- ✅ 参数验证测试
- ✅ 原密码验证测试
- ✅ 新密码长度验证测试
- ✅ 数据库操作测试
- ✅ 错误处理测试

### 前端测试
- ✅ API调用测试
- ✅ 表单验证测试
- ✅ 密码强度检测测试
- ✅ 用户交互测试
- ✅ 错误处理测试

## 部署说明

### 环境要求
- Go 1.24.3+
- MongoDB 6.0+
- Node.js 20+
- Svelte 5.7.0+

### 配置项
```yaml
# config.yaml
auth:
  jwt_secret: "your-jwt-secret"
  password_min_length: 6
  bcrypt_cost: 12
```

### 启动步骤
1. 启动后端服务
```bash
go run cmd/main.go -config configs/config.dev.yaml
```

2. 启动前端开发服务器
```bash
cd web
pnpm install
pnpm dev
```

3. 访问安全设置页面
```
http://localhost:5173/settings/security
```

## 使用指南

### 用户操作流程
1. 登录系统
2. 进入设置页面 (`/settings`)
3. 点击"安全设置"卡片
4. 输入当前密码
5. 输入新密码（查看强度指示器）
6. 确认新密码
7. 点击"修改密码"按钮
8. 查看操作结果

### 密码要求
- 长度：至少6位（建议8位以上）
- 复杂度：建议包含大小写字母、数字和特殊字符
- 唯一性：不能与当前密码相同

## 故障排除

### 常见问题
1. **原密码错误**
   - 检查密码是否正确
   - 确认大小写是否匹配

2. **新密码不符合要求**
   - 确保长度至少6位
   - 检查是否包含必要字符

3. **网络错误**
   - 检查网络连接
   - 确认服务器状态

4. **认证失败**
   - 重新登录
   - 检查JWT令牌是否有效

### 日志查看
```bash
# 后端日志
tail -f logs/stellar.log

# 前端控制台
# 打开浏览器开发者工具查看Console
```

## 更新历史

### v1.0.0 (2024-01-XX)
- ✅ 初始版本实现
- ✅ 基础密码修改功能
- ✅ 密码强度检测
- ✅ 用户界面设计
- ✅ 安全特性实现

## 贡献指南

### 开发规范
- 遵循项目代码规范
- 添加必要的测试用例
- 更新相关文档
- 提交前进行代码审查

### 测试要求
- 单元测试覆盖率 > 80%
- 集成测试覆盖主要流程
- E2E测试覆盖用户操作

---

**注意**：此功能涉及用户安全，请确保在生产环境中充分测试并遵循安全最佳实践。 