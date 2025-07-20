# 星络（Stellar）数据库初始化修复总结

## 🎯 修复概述

本次修复解决了服务端在初次运行时对数据库和Redis初始化不完整的问题，确保系统能够正确创建用户并输出密码。

## 🔧 修复内容

### 1. MongoDB初始化修复

**问题**: `CreateDatabase()` 函数没有被调用，数据库初始化逻辑没有执行

**修复**: 
- 在 `internal/database/mongodb_manager.go` 中集成完整的初始化逻辑
- 自动检查数据库是否存在，不存在则进行初始化
- 创建管理员用户并生成随机密码
- 输出完整的用户信息供用户使用

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

**问题**: Redis没有初始化逻辑，只是连接测试

**修复**:
- 在 `internal/database/redis_manager.go` 中添加 `InitializeRedis()` 方法
- 检查Redis是否为空，为空则设置默认配置
- 创建系统配置、扫描配置、任务队列配置等

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

**问题**: 创建用户使用MD5加密，验证时使用bcrypt，导致认证失败

**修复**:
- 修改 `internal/models/user.go` 中的 `ValidateUser()` 函数
- 支持两种加密方式：MD5（向后兼容）和bcrypt（新标准）
- 自动密码升级：MD5用户登录时自动升级为bcrypt

### 4. 代码清理

**修复**:
- 删除旧的 `CreateDatabase()` 函数
- 清理未使用的导入
- 保留向后兼容的函数

## 🚀 使用方法

### 方法1: 直接运行（推荐）

```bash
# 1. 确保MongoDB和Redis运行
# Windows
mongod --dbpath ./data/db --port 27017
redis-server --port 6379

# Linux/Mac
mongod --dbpath ./data/db --port 27017 --fork --logpath ./data/mongod.log
redis-server --port 6379 --daemonize yes --logfile ./data/redis.log

# 2. 构建并运行应用
go build -o stellar cmd/main.go
./stellar -config configs/config.dev.yaml
```

### 方法2: 使用测试脚本

**Windows**:
```cmd
scripts\test-init.bat
```

**Linux/Mac**:
```bash
chmod +x scripts/test-init.sh
./scripts/test-init.sh
```

## 📋 验证步骤

### 1. 检查初始化输出

运行应用后，应该看到类似以下的输出：

```
🔧 正在初始化数据库: stellarserver_dev
📝 生成管理员密码: [随机12位密码]
✅ 创建管理员用户: admin
✅ 数据库初始化完成
🔑 管理员账户信息:
   用户名: admin
   密码: [随机12位密码]
   邮箱: admin@stellarserver.com

🔧 正在初始化Redis: 127.0.0.1:6379
✅ Redis初始化完成
📊 Redis连接信息:
   地址: 127.0.0.1:6379
   数据库: 0
   连接池大小: 10
```

### 2. 验证数据库

```bash
# 连接到MongoDB
mongosh stellarserver_dev

# 检查用户集合
db.user.find()

# 检查配置集合
db.config.find()
```

### 3. 验证Redis

```bash
# 连接到Redis
redis-cli

# 检查配置键
KEYS stellar:config:*
KEYS stellar:scan:*
```

### 4. 测试登录

使用输出的用户名和密码在前端界面登录，验证认证功能是否正常。

## ⚙️ 配置说明

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

## 🔒 安全注意事项

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

## 🐛 故障排除

### 常见问题

1. **MongoDB连接失败**
   - 检查MongoDB服务是否运行
   - 检查端口27017是否被占用
   - 检查数据目录权限

2. **Redis连接失败**
   - 检查Redis服务是否运行
   - 检查端口6379是否被占用
   - 检查Redis配置

3. **用户认证失败**
   - 确认数据库已正确初始化
   - 检查用户名和密码是否正确
   - 查看日志中的错误信息

### 日志查看

```bash
# MongoDB日志
tail -f ./data/mongod.log

# Redis日志
tail -f ./data/redis.log

# 应用日志
# 查看控制台输出或日志文件
```

## 📝 更新日志

### v1.0.1 (2024-01-XX)

- ✅ 修复MongoDB初始化不完整问题
- ✅ 修复密码输出缺失问题
- ✅ 修复Redis初始化缺失问题
- ✅ 修复用户认证兼容性问题
- ✅ 添加完整的初始化日志输出
- ✅ 添加测试脚本
- ✅ 添加故障排除文档

## 📞 技术支持

如果遇到问题，请：

1. 查看本文档的故障排除部分
2. 检查日志文件中的错误信息
3. 确认MongoDB和Redis服务正常运行
4. 验证配置文件是否正确

---

**注意**: 首次运行时会自动创建管理员账户，请妥善保存输出的密码信息！ 