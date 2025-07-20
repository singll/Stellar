# 数据库初始化修复说明

## 问题描述

在初次运行服务端时，发现以下问题：

1. **MongoDB初始化不完整**：`CreateDatabase()` 函数没有被调用，导致数据库初始化逻辑没有执行
2. **密码输出缺失**：虽然生成了随机密码，但没有输出给用户
3. **Redis初始化缺失**：Redis没有初始化逻辑，只是连接测试
4. **用户认证问题**：创建的用户使用的是MD5加密，但验证时使用的是bcrypt，导致认证失败

## 修复内容

### 1. MongoDB初始化修复

**文件**: `internal/database/mongodb_manager.go`

**修复内容**:
- 在 `NewMongoDBManager()` 中自动调用 `InitializeDatabase()`
- 完整的数据库初始化逻辑，包括：
  - 检查数据库是否存在
  - 创建管理员用户（用户名：admin，邮箱：admin@stellarserver.com）
  - 生成12位随机密码并输出
  - 创建所有必要的集合和索引
  - 插入默认配置数据

**输出示例**:
```
🔧 正在初始化数据库: stellarserver_dev
📝 生成管理员密码: Ab3x9Kp2mN7q
⚠️  请妥善保存此密码，首次登录后建议修改
✅ 创建管理员用户: admin
✅ 数据库初始化完成
🔑 管理员账户信息:
   用户名: admin
   密码: Ab3x9Kp2mN7q
   邮箱: admin@stellarserver.com
⚠️  请使用以上信息登录系统，登录后请及时修改密码
```

### 2. Redis初始化修复

**文件**: `internal/database/redis_manager.go`

**修复内容**:
- 在 `NewRedisManager()` 中自动调用 `InitializeRedis()`
- 完整的Redis初始化逻辑，包括：
  - 检查Redis是否为空（新安装）
  - 设置默认配置（版本、时区等）
  - 设置扫描配置（超时、并发数等）
  - 设置任务队列配置
  - 设置节点管理配置

**输出示例**:
```
🔧 正在初始化Redis: 127.0.0.1:6379
✅ Redis初始化完成
📊 Redis连接信息:
   地址: 127.0.0.1:6379
   数据库: 0
   连接池大小: 10
   密码: 无
```

### 3. 用户认证修复

**文件**: `internal/models/user.go`

**修复内容**:
- 修改 `ValidateUser()` 函数，支持两种加密方式：
  - **MD5加密**：向后兼容，用于旧用户
  - **bcrypt加密**：新标准，用于新用户
- 自动密码升级：当MD5用户登录时，自动将密码升级为bcrypt
- 保持向后兼容性

**认证流程**:
1. 首先尝试bcrypt验证（新用户）
2. 如果失败，尝试MD5验证（向后兼容）
3. 如果是MD5密码，自动升级为bcrypt（后台异步）

### 4. 代码清理

**文件**: `internal/database/mongodb.go`

**修复内容**:
- 删除旧的 `CreateDatabase()` 函数
- 删除未使用的导入
- 保留向后兼容的函数

## 测试方法

### 1. 使用测试脚本

```bash
# Windows
scripts/test-init.bat

# Linux/Mac
chmod +x scripts/test-init.sh
./scripts/test-init.sh
```

### 2. 手动测试

```bash
# 1. 确保MongoDB和Redis运行
mongod --dbpath ./data/db --port 27017
redis-server --port 6379

# 2. 构建并运行应用
go build -o stellar cmd/main.go
./stellar -config configs/config.dev.yaml
```

### 3. 验证初始化结果

**MongoDB验证**:
```bash
# 连接到MongoDB
mongosh stellarserver_dev

# 检查用户集合
db.user.find()

# 检查配置集合
db.config.find()
```

**Redis验证**:
```bash
# 连接到Redis
redis-cli

# 检查配置键
KEYS stellar:config:*
KEYS stellar:scan:*
```

## 配置说明

### 开发环境配置

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

### 生产环境配置

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

## 安全注意事项

1. **密码安全**：
   - 生成的随机密码长度为12位
   - 首次登录后建议立即修改密码
   - 支持密码复杂度验证

2. **数据库安全**：
   - 生产环境必须设置MongoDB认证
   - 生产环境必须设置Redis密码
   - 定期备份数据库

3. **网络安全**：
   - 生产环境使用HTTPS
   - 配置防火墙规则
   - 限制数据库访问IP

## 故障排除

### 1. MongoDB连接失败

**错误**: `failed to connect to MongoDB`

**解决方案**:
```bash
# 检查MongoDB服务状态
systemctl status mongod

# 启动MongoDB服务
systemctl start mongod

# 检查端口是否被占用
netstat -tlnp | grep 27017
```

### 2. Redis连接失败

**错误**: `failed to connect to Redis`

**解决方案**:
```bash
# 检查Redis服务状态
systemctl status redis

# 启动Redis服务
systemctl start redis

# 检查端口是否被占用
netstat -tlnp | grep 6379
```

### 3. 用户认证失败

**错误**: `Invalid credentials`

**解决方案**:
1. 检查用户名和密码是否正确
2. 确认数据库已正确初始化
3. 查看日志中的错误信息

### 4. 权限问题

**错误**: `Permission denied`

**解决方案**:
```bash
# 检查数据目录权限
ls -la data/

# 修改权限
chmod 755 data/
chown -R $USER:$USER data/
```

## 更新日志

### v1.0.1 (2024-01-XX)

- ✅ 修复MongoDB初始化不完整问题
- ✅ 修复密码输出缺失问题
- ✅ 修复Redis初始化缺失问题
- ✅ 修复用户认证兼容性问题
- ✅ 添加完整的初始化日志输出
- ✅ 添加测试脚本
- ✅ 添加故障排除文档

### v1.0.0 (2024-01-XX)

- 🎉 初始版本发布
- �� 基础功能实现
- 🔧 基础配置支持 