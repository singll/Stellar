# Claude 开发助手记忆文件

## 项目概述
**项目名称**: Stellar - 安全资产管理平台  
**版本**: 1.0  
**类型**: 网络安全工具平台  
**架构**: 前后端分离 + 微服务架构  

## 技术栈信息

### 后端技术栈
- **核心语言**: Go 1.24.3
- **Web框架**: Gin v1.10.1
- **数据库**: 
  - 主数据库: MongoDB (go.mongodb.org/mongo-driver v1.17.4)
  - 开发环境: mongodb://192.168.7.216:27017
  - 数据库名: stellarserver_dev (开发) / stellarserver (生产)
- **缓存**: Redis v9.10.0 (192.168.7.128:6379)
- **认证**: JWT (github.com/golang-jwt/jwt/v4 v4.5.0)
- **日志**: github.com/rs/zerolog v1.32.0
- **定时任务**: github.com/robfig/cron/v3 v3.0.1

### 前端技术栈
- **核心框架**: Svelte v5.7.0 + SvelteKit v2.0.0
- **构建工具**: Vite v6.0.0
- **语言**: TypeScript v5.0.0
- **UI框架**: shadcn-svelte v0.9.0 + Tailwind CSS v3.4.1
- **状态管理**: @tanstack/store v0.7.1
- **HTTP客户端**: axios v1.6.0
- **包管理器**: pnpm

### 开发环境
- **后端端口**: 8090
- **前端端口**: 5173
- **代理配置**: Vite代理API请求到后端

## 项目结构

### 目录架构
```
/root/Stellar/
├── cmd/                    # 程序入口
├── internal/               # 后端核心代码
│   ├── api/               # API接口层
│   ├── models/            # 数据模型
│   ├── services/          # 业务服务
│   ├── utils/             # 工具包
│   └── pkg/               # 公共包
├── web/                   # 前端代码
│   ├── src/               # 源码目录
│   └── package.json       # 前端依赖
├── configs/               # 配置文件
├── scripts/               # 脚本文件
└── Makefile              # 构建脚本
```

### 微服务架构
- **资产管理服务** (Asset Management)
- **项目管理服务** (Project Management)  
- **任务管理服务** (Task Management)
- **节点管理服务** (Node Management)
- **漏洞扫描服务** (Vulnerability Scan)
- **端口扫描服务** (Port Scan)
- **子域名枚举服务** (Subdomain)
- **敏感信息检测服务** (Sensitive)
- **插件系统** (Plugin System)

## API接口规范

### 路由前缀
- **API版本**: `/api/v1`
- **健康检查**: `/health`
- **WebSocket**: `/ws`

### 主要接口模块
- `/api/v1/assets` - 资产管理
- `/api/v1/projects` - 项目管理
- `/api/v1/tasks` - 任务管理
- `/api/v1/auth` - 身份认证
- `/api/v1/nodes` - 节点管理
- `/api/v1/vulnerabilities` - 漏洞管理
- `/api/v1/portscan` - 端口扫描
- `/api/v1/subdomains` - 子域名
- `/api/v1/plugins` - 插件管理

### 认证机制
- **认证方式**: JWT Bearer Token
- **会话管理**: Redis存储
- **权限验证**: 中间件拦截

## 开发规范

### 🌏 语言使用规范
- **交流语言**: 全程使用中文与用户交流
- **代码注释**: 使用中文编写代码注释
- **变量命名**: 变量名和函数名使用英文，保持代码规范
- **文档编写**: 技术文档和说明均使用中文

### 构建命令
```bash
# 启动开发环境 (前端5173 + 后端8090)
make dev

# 停止开发环境
make dev-stop

# 检查环境状态
make status

# 查看日志
make logs

# 运行测试
make test

# 构建生产版本
make build
```

### 手动启动 (开发期间)
```bash
# 后端手动启动
go run ./cmd/main.go -config configs/config.dev.yaml -log-level debug

# 前端手动启动
cd web && pnpm dev
```

### 代码规范
- **错误处理**: 使用 `internal/pkg/errors` 统一错误处理
- **日志记录**: 使用 `internal/pkg/logger` 结构化日志
- **配置管理**: YAML配置文件 + 环境变量
- **数据验证**: Gin的ShouldBindJSON + 自定义验证器
- **中文开发**: 优先使用中文进行代码交流和文档编写

## 数据库设计

### MongoDB集合
- **projects** - 项目信息
- **domain_assets** - 域名资产
- **subdomain_assets** - 子域名资产
- **ip_assets** - IP资产
- **port_assets** - 端口资产
- **url_assets** - URL资产
- **http_assets** - HTTP服务资产
- **app_assets** - 应用资产
- **miniapp_assets** - 小程序资产
- **users** - 用户信息
- **asset_relations** - 资产关系

### Redis键命名
- **session:*** - 用户会话
- **task:*** - 任务状态
- **cache:*** - 数据缓存

## 开发注意事项

### 🚨 重要提醒
1. **全程中文交流** - 必须使用中文与用户进行所有交互和说明
2. **不要自动启动环境** - 开发期间告知用户手动启动验证
3. **项目ID验证** - 创建资产时项目ID是可选的,不存在时创建独立资产
4. **错误处理** - 使用宽松模式,避免因验证失败阻止核心功能
5. **日志记录** - 详细记录请求参数和处理过程便于调试

### 📋 日志文件管理

#### 日志文件位置
- **后端日志**: `/root/Stellar/logs/backend.log`
- **前端日志**: `/root/Stellar/logs/frontend.log`

#### 调试时日志查看指引
调试问题时，优先查看相关日志文件：

```bash
# 实时查看后端日志
tail -f /root/Stellar/logs/backend.log

# 实时查看前端日志  
tail -f /root/Stellar/logs/frontend.log

# 搜索特定关键词
grep -i "error\|错误" /root/Stellar/logs/backend.log
grep -i "GetProjects\|分页" /root/Stellar/logs/backend.log
grep -i "API\|接口" /root/Stellar/logs/frontend.log

# 查看最近的日志（最后100行）
tail -100 /root/Stellar/logs/backend.log
tail -100 /root/Stellar/logs/frontend.log
```

#### 常见问题日志查找
- **分页数据错误** → 后端日志搜索"GetProjects"
- **API接口报错** → 后端日志搜索"error"或HTTP状态码
- **前端页面异常** → 前端日志搜索"error"或组件名
- **数据库连接问题** → 后端日志搜索"mongo\|database"
- **认证授权问题** → 后端日志搜索"auth\|token\|jwt"

#### 🔧 调试流程快速参考
当遇到问题时，按以下顺序进行调试：

1. **确定问题类型**
   - 前端页面问题 → 查看前端日志
   - API接口问题 → 查看后端日志
   - 数据流问题 → 同时查看前后端日志

2. **使用正确的日志查看命令**
   ```bash
   # 当前分页问题调试
   tail -f /root/Stellar/logs/backend.log | grep "GetProjects"
   
   # API错误调试  
   tail -f /root/Stellar/logs/backend.log | grep -i "error\|panic\|fatal"
   
   # 前端错误调试
   tail -f /root/Stellar/logs/frontend.log | grep -i "error\|failed"
   ```

3. **关键日志标识符**
   - `GetProjects -` → 项目列表相关日志
   - `CreateAsset` → 资产创建相关日志
   - `[API]` → 前端API调用日志
   - `[项目管理]` → 项目管理页面日志

### 已知问题修复记录
- ✅ **资产创建500错误** - 修复项目ID验证过于严格的问题 (2025-07-24)
  - 问题: 项目不存在时直接返回错误
  - 解决: 改为宽松模式,项目不存在时创建独立资产

- 🔄 **分页数据显示错误** - 正在调试中 (2025-07-24)
  - 问题: 切换每页显示数量时总数显示错误
  - 调试: 已添加详细日志到`/root/Stellar/logs/backend.log`
  - 状态: 需要查看后端日志确认数据库查询结果

### 测试验证
- **API测试**: 使用curl或Postman
- **认证测试**: 需要先注册/登录获取JWT Token
- **数据库连接**: MongoDB (192.168.7.216:27017) + Redis (192.168.7.128:6379)

## 文件路径速查

### 关键文件位置
- **主程序入口**: `/root/Stellar/cmd/main.go`
- **资产API**: `/root/Stellar/internal/api/asset.go`
- **项目API**: `/root/Stellar/internal/api/project.go`  
- **认证API**: `/root/Stellar/internal/api/auth.go`
- **路由配置**: `/root/Stellar/internal/api/router/`
- **前端主页**: `/root/Stellar/web/src/routes/(app)/assets/new/+page.svelte`
- **API客户端**: `/root/Stellar/web/src/lib/api/asset.ts`
- **开发配置**: `/root/Stellar/configs/config.dev.yaml`
- **构建脚本**: `/root/Stellar/Makefile`

---
*最后更新: 2025-07-24*  
*Claude助手专用记忆文件 - 请勿删除*