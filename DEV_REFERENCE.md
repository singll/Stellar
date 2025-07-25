# Stellar 开发快速参考

## 🚀 快速启动

### 环境启动 (手动)
```bash
# 后端启动 (端口8090)
go run ./cmd/main.go -config configs/config.dev.yaml -log-level debug

# 前端启动 (端口5173) 
cd web && pnpm dev

# 或使用Makefile (会自动管理两个服务)
make dev
```

### 环境检查
```bash
# 检查服务状态
make status

# 查看实时日志  
make logs

# 检查端口占用
lsof -i :8090  # 后端
lsof -i :5173  # 前端
```

## 🏗️ 项目架构速览

```
Stellar/
├── cmd/main.go              # 程序入口
├── internal/
│   ├── api/                 # API接口层
│   │   ├── asset.go        # 资产管理API
│   │   ├── project.go      # 项目管理API
│   │   ├── auth.go         # 认证API
│   │   └── router/         # 路由配置
│   ├── models/             # 数据模型
│   ├── services/           # 业务服务
│   └── utils/              # 工具包
├── web/                    # 前端(Svelte5)
│   ├── src/routes/         # 页面路由
│   └── src/lib/api/        # API客户端
├── configs/                # 配置文件
└── Makefile               # 构建脚本
```

## 🔧 常用开发命令

### 后端开发
```bash
# 运行后端
go run ./cmd/main.go -config configs/config.dev.yaml

# 格式化代码
go fmt ./...

# 安装依赖
go mod tidy

# 运行测试
go test ./...
```

### 前端开发
```bash
cd web

# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev

# 构建生产版本
pnpm build

# 运行测试
pnpm test
```

## 📡 API 测试示例

### 认证流程
```bash
# 1. 注册用户
curl -X POST "http://localhost:8090/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpassword", 
    "email": "test@example.com"
  }'

# 2. 登录获取Token
curl -X POST "http://localhost:8090/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpassword"
  }'
```

### 资产管理
```bash
# 创建资产 (需要Bearer Token)
curl -X POST "http://localhost:8090/api/v1/assets" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "type": "domain",
    "projectId": "",
    "tags": ["test"],
    "data": {
      "domain": "example.com"
    }
  }'

# 获取资产列表
curl -X GET "http://localhost:8090/api/v1/assets?type=domain" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 🗄️ 数据库连接

### MongoDB (主数据库)
- **开发环境**: mongodb://192.168.7.216:27017
- **数据库名**: stellarserver_dev
- **集合**: projects, domain_assets, users, etc.

### Redis (缓存)
- **地址**: 192.168.7.128:6379 
- **用途**: 会话存储、任务队列、数据缓存

## 🚨 常见问题排查

### 1. 资产创建500错误
- **症状**: POST /api/v1/assets 返回 UNKNOWN_ERROR
- **原因**: 项目ID验证失败或数据库连接问题
- **解决**: 检查项目是否存在，验证数据库连接

### 2. 认证失败
- **症状**: 401 Unauthorized 
- **原因**: JWT Token过期或无效
- **解决**: 重新登录获取新Token

### 3. 前端代理失败
- **症状**: API请求404或超时
- **原因**: 后端服务未启动或端口冲突
- **解决**: 检查后端8090端口是否正常监听

### 4. 数据库连接失败
- **症状**: 启动时数据库连接错误
- **原因**: MongoDB/Redis服务未启动或网络问题
- **解决**: 验证数据库服务状态和网络连通性

## 🔍 日志查看

### 后端日志
```bash
# 实时日志
tail -f /root/Stellar/logs/backend.log

# 或查看控制台输出 (调试模式)
go run ./cmd/main.go -config configs/config.dev.yaml -log-level debug
```

### 前端日志
```bash
# 开发服务器日志
tail -f /root/Stellar/logs/frontend.log

# 浏览器控制台 (F12)
```

## 🧪 测试策略

### 单元测试
```bash
# 后端测试
make test-backend

# 前端测试  
make test-frontend
```

### 集成测试
```bash
# E2E测试 (需要环境运行)
make test-e2e
```

### API测试
- 使用 Postman 集合
- curl 命令脚本
- 自动化测试套件

---

## 💡 开发提醒

### 🚨 重要规则
1. **不要自动启动环境** - 开发期间需要手动启动验证
2. **资产创建宽松验证** - 项目ID不存在时继续创建资产
3. **详细日志记录** - 关键操作要有日志跟踪

### 📋 提交前检查
- [ ] 代码格式化完成
- [ ] 单元测试通过
- [ ] API功能验证
- [ ] 日志输出正常
- [ ] 没有硬编码配置

---
*快速参考文档 v1.0*  
*最后更新: 2025-07-24*