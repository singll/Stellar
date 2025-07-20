# 密码加密修复说明

## 问题描述

在服务端初始化时，发现以下问题：

1. **字段名不匹配**：数据库初始化时使用 `"password"` 字段，但用户模型中定义的是 `"hashedPassword"` 字段
2. **加密方式不一致**：初始化时使用MD5加密，但验证时期望bcrypt加密
3. **安全性问题**：MD5加密不够安全，容易被破解

## 修复内容

### 1. 字段名统一

**文件**: `internal/database/mongodb_manager.go`

**修复前**:
```go
_, err = collection.InsertOne(context.Background(), bson.M{
    "username":  "admin",
    "password":  passwordHash,  // ❌ 错误的字段名
    "email":     "admin@stellarserver.com",
    "roles":     []string{"admin"},
    "created":   time.Now(),
    "lastLogin": time.Now(),
})
```

**修复后**:
```go
_, err = collection.InsertOne(context.Background(), bson.M{
    "username":       "admin",
    "hashedPassword": string(hashedPassword),  // ✅ 正确的字段名
    "email":          "admin@stellarserver.com",
    "roles":          []string{"admin"},
    "created":        time.Now(),
    "lastLogin":      time.Now(),
})
```

### 2. 加密方式升级

**修复前**:
```go
// 使用MD5加密（不安全）
passwordHash := fmt.Sprintf("%x", md5.Sum([]byte(password)))
```

**修复后**:
```go
// 使用bcrypt加密（更安全）
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
if err != nil {
    return fmt.Errorf("failed to hash password: %w", err)
}
```

### 3. 向后兼容性

**文件**: `internal/models/user.go`

为了保持向后兼容性，`ValidateUser` 函数仍然支持两种加密方式：

1. **bcrypt验证**（新用户，优先）
2. **MD5验证**（旧用户，向后兼容）
3. **自动升级**：MD5用户登录时自动升级为bcrypt

```go
// 首先尝试bcrypt验证（新用户）
if len(user.HashedPassword) > 0 {
    if bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)) == nil {
        passwordValid = true
    }
}

// 如果bcrypt验证失败，尝试MD5验证（向后兼容）
if !passwordValid {
    if len(user.HashedPassword) == 32 {
        hashedInput := fmt.Sprintf("%x", md5.Sum([]byte(password)))
        if hashedInput == user.HashedPassword {
            passwordValid = true
            // 自动升级为bcrypt
            go func() {
                // ... 升级逻辑
            }()
        }
    }
}
```

## 安全改进

### 1. 加密强度提升

- **MD5** → **bcrypt**：从32位哈希升级到自适应哈希
- **成本因子**：使用 `bcrypt.DefaultCost`（当前为12）
- **盐值**：bcrypt自动生成随机盐值

### 2. 字段名规范化

- 统一使用 `hashedPassword` 字段名
- 避免与明文密码字段混淆
- 提高代码可读性和维护性

### 3. 错误处理改进

- 添加密码哈希失败的错误处理
- 提供更详细的错误信息
- 确保系统稳定性

## 测试验证

### 1. 使用测试脚本

**Linux/Mac**:
```bash
chmod +x scripts/test-password-fix.sh
./scripts/test-password-fix.sh
```

**Windows**:
```cmd
scripts\test-password-fix.bat
```

### 2. 手动测试

```bash
# 1. 清理旧数据
rm -rf ./data/db
mkdir -p ./data/db

# 2. 启动服务
mongod --dbpath ./data/db --port 27017 --fork --logpath ./data/mongod.log
redis-server --port 6379 --daemonize yes --logfile ./data/redis.log

# 3. 运行应用
go build -o stellar cmd/main.go
./stellar -config configs/config.dev.yaml
```

### 3. 验证步骤

1. **检查初始化输出**：确认包含管理员密码
2. **尝试登录**：使用输出的用户名和密码
3. **验证成功**：确认能够正常登录系统

## 影响范围

### 1. 新安装

- ✅ 完全兼容，使用bcrypt加密
- ✅ 字段名正确，验证正常
- ✅ 安全性更高

### 2. 现有用户

- ✅ 向后兼容，MD5用户仍可登录
- ✅ 自动升级，首次登录后升级为bcrypt
- ✅ 无数据丢失

### 3. 开发环境

- ✅ 测试脚本更新
- ✅ 文档完善
- ✅ 错误处理改进

## 配置说明

### 开发环境

**文件**: `configs/config.dev.yaml`

```yaml
mongodb:
  uri: "mongodb://127.0.0.1:27017"
  database: "stellarserver_dev"
  user: "admin"
  password: "admin"

redis:
  addr: "127.0.0.1:6379"
  password: "redis"
  db: 0
  poolSize: 10
```

### 生产环境

**文件**: `configs/config.prod.yaml`

```yaml
mongodb:
  uri: "mongodb://192.168.7.216:27017"
  database: "stellarserver"
  user: "admin"
  password: "admin"

redis:
  addr: "192.168.7.128:6379"
  password: "redis"
  db: 0
  poolSize: 100
```

## 故障排除

### 1. 登录失败

**问题**: 使用初始密码登录失败

**解决方案**:
1. 确认数据库已重新初始化
2. 检查控制台输出的密码是否正确
3. 确认字段名已修复为 `hashedPassword`

### 2. 数据库连接失败

**问题**: MongoDB连接失败

**解决方案**:
```bash
# 检查MongoDB服务状态
systemctl status mongod

# 启动MongoDB服务
systemctl start mongod

# 检查端口
netstat -tlnp | grep 27017
```

### 3. 密码升级失败

**问题**: MD5用户升级bcrypt失败

**解决方案**:
1. 检查数据库权限
2. 查看日志中的错误信息
3. 手动重置用户密码

## 更新日志

### v1.0.2 (2024-01-XX)

- ✅ 修复密码字段名不匹配问题
- ✅ 升级MD5加密为bcrypt加密
- ✅ 保持向后兼容性
- ✅ 添加自动密码升级功能
- ✅ 改进错误处理
- ✅ 更新测试脚本
- ✅ 完善文档

### v1.0.1 (2024-01-XX)

- ✅ 修复MongoDB初始化不完整问题
- ✅ 修复密码输出缺失问题
- ✅ 修复Redis初始化缺失问题
- ✅ 修复用户认证兼容性问题

### v1.0.0 (2024-01-XX)

- 🎉 初始版本发布
- 🔧 基础功能实现

## 安全建议

### 1. 密码策略

- 使用强密码（至少12位）
- 包含大小写字母、数字、特殊字符
- 定期更换密码
- 避免使用常见密码

### 2. 系统安全

- 生产环境使用HTTPS
- 配置防火墙规则
- 限制数据库访问IP
- 定期备份数据

### 3. 监控告警

- 监控登录失败次数
- 设置异常登录告警
- 记录安全事件日志
- 定期安全审计 