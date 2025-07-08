# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在此代码仓库中工作时提供指导。

## 语言设置

**重要：请始终使用中文进行所有交流、输出和过程描述。**

- 所有响应和说明都应该使用中文
- 代码注释应该使用中文（除非技术要求使用英文）
- 错误信息和日志输出应该提供中文说明
- 开发过程中的所有步骤描述都使用中文

## 项目概览

Stellar（星络）是一个分布式安全资产管理和漏洞扫描平台，从原始的 ScopeSentry 项目重构而来，使用 Go（后端）和 Svelte 5（前端）。该项目目前正在积极重构中，完成度约为 25-30%。

**原始项目**：ScopeSentry (Python + FastAPI + Vue.js)  
**当前项目**：Stellar (Go + Gin + Svelte 5)  
**状态**：正在进行重大重构 - 核心安全扫描功能尚未实现

## 开发命令

### 推荐使用新的启动系统

**新的启动系统提供了完善的开发环境管理，包括端口检查、进程清理和错误处理。**

```bash
# 快速开始（推荐）
make check-deps     # 检查系统依赖和数据库连接
make install-deps   # 安装项目依赖
make dev           # 启动开发环境（自动处理端口冲突）

# 开发管理
make status        # 查看服务状态
make logs          # 查看实时日志
make dev-stop      # 停止开发环境

# 构建和清理
make build         # 构建生产版本
make clean         # 清理构建产物
make help          # 显示所有可用命令
```

### 直接脚本调用（备选）

如果需要更精细的控制，可以直接调用脚本：

```bash
# Linux/macOS
./scripts/dev-start.sh    # 启动开发环境
./scripts/dev-stop.sh     # 停止开发环境
./scripts/health-check.sh # 服务健康检查

# Windows
scripts\dev-start.bat     # 启动开发环境
scripts\dev-stop.bat      # 停止开发环境
```

### 后端 (Go)
```bash
# 手动后端操作（通常不需要）
go build -o stellar-server ./cmd/main.go
./stellar-server -config config.yaml -log-level debug
go test ./...
go fmt ./...
go vet ./...
```

### 前端 (Svelte)
```bash
cd web

# 手动前端操作（通常不需要）
pnpm install
pnpm run dev
pnpm run build
pnpm run test
pnpm run check
pnpm run format
pnpm run lint
```

### 启动系统特性

**自动端口管理**：
- 自动检测并清理占用的端口（8090, 5173）
- 智能进程终止（先TERM，后KILL）
- 端口释放验证

**依赖检查**：
- 验证系统工具（Go、pnpm、lsof）
- 测试数据库连接（MongoDB、Redis）
- 检查前端依赖完整性

**服务监控**：
- 实时服务状态检查
- 进程启动超时保护
- 详细的错误报告和日志

**日志管理**：
- 所有日志存储在 `logs/` 目录
- 支持实时日志查看
- PID文件管理，便于进程控制

## 架构概览

### 后端架构 (Go)
- **框架**：Gin web 框架
- **数据库**：MongoDB 用于数据存储，Redis 用于缓存/队列
- **认证**：基于 JWT 的认证，带有中间件保护
- **结构**：清洁架构，关注点分离

```
internal/
├── api/          # HTTP 处理器和路由
├── config/       # 配置管理
├── database/     # 数据库连接 (MongoDB/Redis)
├── models/       # 数据模型和仓库
├── services/     # 业务逻辑层
├── plugin/       # 插件系统架构
└── utils/        # 工具函数 (JWT, 日志等)
```

### 前端架构 (Svelte 5)
- **框架**：Svelte 5 + SvelteKit，启用 runes
- **UI**：shadcn-svelte + Tailwind CSS
- **状态管理**：Svelte 5 runes + TanStack Store
- **HTTP**：Axios 与代理配置
- **测试**：Vitest + Testing Library

```
web/src/
├── lib/
│   ├── api/          # API 客户端函数
│   ├── auth/         # 认证工具
│   ├── components/   # 可重用 UI 组件
│   ├── stores/       # 状态管理
│   ├── types/        # TypeScript 类型定义
│   └── utils/        # 工具函数
└── routes/           # SvelteKit 页面和布局
    ├── (app)/        # 认证应用路由
    └── (auth)/       # 认证路由
```

## 关键技术细节

### 数据库配置
- **MongoDB**：默认连接到 `mongodb://192.168.7.216:27017`
- **Redis**：默认连接到 `192.168.7.128:6379`
- **数据库名**：`stellarserver`
- 配置详情请参考 `config.yaml`

### 认证系统
- 基于 JWT 的认证，可配置过期时间
- 对所有 `/api/v1/*` 路由的中间件保护（除认证端点外）
- 令牌存储在前端的 localStorage 中
- 实现了自动刷新机制

### API 设计
- 使用 `/api/v1/` 前缀的 RESTful API
- 一致的错误处理和响应格式
- 所有业务路由都需要认证
- WebSocket 支持，地址为 `/ws`

### 插件系统
- 模块化插件架构，带有沙盒执行
- 插件元数据存储在 MongoDB 中
- 基于注册表的插件管理
- 支持脚本和编译插件

## 实现状态

### ✅ 已完成功能
- **认证系统**：用户登录/登出，JWT 管理，中间件保护
- **项目管理**：CRUD 操作，项目列表，基础统计
- **资产管理**：基础资产模型，CRUD 操作，类型安全 API
- **数据库层**：MongoDB/Redis 集成，连接管理
- **前端基础**：Svelte 5 设置，组件库，路由

### ❓ 部分实现
- **任务管理**：框架存在，核心调度逻辑缺失
- **节点管理**：基础结构，健康检查未实现
- **插件系统**：架构完成，执行引擎部分完成

### ❌ 尚未实现（高优先级）
- **子域名枚举**：缺少核心安全功能
- **端口扫描**：缺少安全扫描功能
- **漏洞扫描**：POC 管理和执行缺失
- **目录扫描**：缺少 Web 应用安全测试
- **敏感信息检测**：缺少数据泄露检测
- **页面监控**：缺少网站变化监控
- **Web 爬虫**：缺少内容发现功能

## 开发指南

### 代码风格
- **Go**：遵循标准 Go 约定，使用 `gofmt` 和 `go vet`
- **TypeScript/Svelte**：使用提供的 Prettier 和 ESLint 配置
- **组件**：使用 Svelte 5 runes 语法（`$state`，`$props`，`$effect`）
- **API 调用**：使用 `lib/api/` 中的类型安全 API 客户端

### 数据库模式
- 使用 Repository 模式进行数据访问
- 在 `internal/models/` 中定义模型，使用 MongoDB 结构标签
- 实现正确的错误处理和验证
- 使用连接池和正确的资源清理

### 测试策略
- 后端：使用 Go 测试框架的单元测试
- 前端：使用 Vitest + Testing Library 的组件测试
- 集成：API 端点测试
- 手动：使用 `lib/dev/` 中的开发验证脚本

### 错误处理
- 后端：一致的错误响应格式，使用正确的 HTTP 状态码
- 前端：用户反馈的全局通知系统
- API：有意义的错误传播消息

## 配置管理

### 环境配置
- 主配置文件：`config.yaml`
- 支持不同环境（debug/release/test）
- 数据库连接、超时和服务参数可配置
- JWT 密钥和令牌过期设置

### 开发 vs 生产
- 开发：热重载，详细日志，启用 CORS
- 生产：优化构建，嵌入式前端，安全头
- 开发使用 `make dev`，生产使用 `make build`

## 安全考虑

- JWT 密钥应该是环境特定的
- 数据库凭据不应提交到仓库
- API 端点受认证中间件保护
- 对所有用户输入进行验证
- 配置速率限制和请求大小限制

## 性能优化

### 后端
- MongoDB 和 Redis 的连接池
- 基于中间件的请求日志和监控
- 优雅关闭处理
- 可配置的超时和并发限制

### 前端
- Svelte 5 编译时优化
- 路由和组件的懒加载
- 使用 runes 的高效状态管理
- 构建时资源优化

## 开发工作流

### 添加新功能
1. **定义类型**：在 `lib/types/` 中添加 TypeScript 接口
2. **后端 API**：在 `internal/api/` 中实现处理器
3. **前端 API**：在 `lib/api/` 中创建客户端函数
4. **组件**：在 `lib/components/` 中构建 UI 组件
5. **页面**：在 `routes/` 中添加路由
6. **测试**：在 `lib/dev/` 中添加验证

### 数据库变更
1. 更新 `internal/models/` 中的模型
2. 如需要，添加迁移脚本
3. 更新 API 处理器和验证
4. 使用不同数据场景测试

### 安全功能实现
1. 研究原始 ScopeSentry 实现作为参考
2. 设计具有适当并发性的 Go 原生实现
3. 与任务管理系统集成
4. 添加适当的配置和错误处理
5. 为配置和结果创建前端界面

## 常见问题和解决方案

### 开发设置
- 确保在启动服务前运行 MongoDB 和 Redis
- 前端代理配置将 `/api` 路由到 `localhost:8090` 的后端
- 使用 `make dev` 同时启动两个服务

### 数据库连接
- 检查 `config.yaml` 中的正确数据库 URL
- 验证 MongoDB 和 Redis 服务可访问
- 开发时连接池设置可能需要调整

### 认证问题
- JWT 密钥在请求间必须一致
- 检查配置中的令牌过期设置
- 前端自动处理令牌刷新

## 参考和资源

- **原始项目**：ScopeSentry-main（Python 后端），ScopeSentry-UI-main（Vue.js 前端）
- **开发计划**：参见 `DEV_PLAN.md` 和 `DEV_PLAN/` 目录
- **项目文档**：`README.md` 概览
- **前端指南**：`web/PROJECT_MANAGEMENT_README.md` 功能使用

## 开发优先任务

基于当前重构状态，专注于这些领域：

1. **任务管理系统**：完成任务调度和分发逻辑
2. **核心安全功能**：实现子域名枚举、端口扫描、漏洞检测
3. **节点管理**：完成分布式节点协调
4. **前端完成**：为安全功能和系统管理添加缺失页面
5. **集成测试**：确保所有组件正常协作

## 技术规范和版本要求

### 前端技术规范 (Svelte 5)

**关键版本信息**：
- **Svelte**: 5.x (使用 runes 模式)
- **SvelteKit**: 最新版本，支持 Svelte 5
- **TypeScript**: 5.x
- **Node.js**: 18+ 

**Svelte 5 核心语法规则**：

#### 1. 响应式状态管理
```javascript
// ✅ 正确：使用 Svelte 5 runes
let count = $state(0);
let name = $state('');
let isLoading = $state(false);

// ❌ 错误：Svelte 4 语法
let count = 0;
```

#### 2. 派生状态
```javascript
// ✅ 正确：使用 $derived
let doubled = $derived(count * 2);
let fullName = $derived(`${firstName} ${lastName}`);

// ❌ 错误：Svelte 4 语法
$: doubled = count * 2;
```

#### 3. 副作用处理
```javascript
// ✅ 正确：使用 $effect
$effect(() => {
  console.log('Count changed:', count);
});

// ❌ 错误：Svelte 4 语法
$: console.log('Count changed:', count);
```

#### 4. Props 定义
```javascript
// ✅ 正确：使用 $props
let { title, items = [], onClick } = $props();

// ❌ 错误：Svelte 4 语法
export let title;
export let items = [];
export let onClick;
```

#### 5. Store 使用
```javascript
// ✅ 正确：兼容 Svelte 5 的 store 使用
import { writable } from 'svelte/store';
import { get } from 'svelte/store';

const store = writable(initialValue);

// 在组件中使用
let storeValue = $state();
store.subscribe(value => storeValue = value);

// 或者直接使用
const currentValue = get(store);
```

#### 6. 路由跳转语法
```javascript
// ✅ 正确：使用标准路由路径
await goto('/dashboard');
await goto('/login');
await goto('/projects/123');

// ❌ 错误：SvelteKit 路由组语法（仅用于文件组织）
await goto('/(app)/dashboard');  // 这是文件组织，不是路由路径
await goto('/(auth)/login');     // 这是文件组织，不是路由路径
```

#### 7. 事件处理
```javascript
// ✅ 正确：现代事件处理
<button onclick={() => count++}>增加</button>
<input oninput={(e) => name = e.target.value} />

// ✅ 也支持：传统语法
<button on:click={() => count++}>增加</button>
```

#### 8. 条件渲染和循环
```javascript
// ✅ 正确：保持原有语法
{#if condition}
  <div>内容</div>
{/if}

{#each items as item, index (item.id)}
  <div>{item.name}</div>
{/each}
```

### 路由系统规范

**SvelteKit 路由组 vs 实际路径**：

#### 文件组织（路由组）
```
routes/
├── (app)/           # 路由组：需要认证的页面
│   ├── dashboard/
│   ├── projects/
│   └── +layout.svelte
├── (auth)/          # 路由组：认证相关页面
│   ├── login/
│   ├── register/
│   └── +layout.svelte
└── +page.svelte
```

#### 实际路由路径
```javascript
// ✅ 正确：使用实际路径进行跳转
goto('/dashboard')     // → routes/(app)/dashboard/+page.svelte
goto('/login')         // → routes/(auth)/login/+page.svelte
goto('/projects/123')  // → routes/(app)/projects/[id]/+page.svelte

// ❌ 错误：不要在 goto() 中使用路由组语法
goto('/(app)/dashboard')  // 这会导致 404 错误
goto('/(auth)/login')     // 这会导致 404 错误
```

### 常见错误模式和修复

#### 错误 1：路由跳转失败
```javascript
// ❌ 问题代码
await goto('/(app)/dashboard');

// ✅ 修复
await goto('/dashboard');
```

#### 错误 2：响应式状态不更新
```javascript
// ❌ 问题代码
let count = 0;  // Svelte 4 语法

// ✅ 修复
let count = $state(0);  // Svelte 5 runes
```

#### 错误 3：Store 订阅内存泄漏
```javascript
// ❌ 问题代码
get state() {
  let currentState;
  store.subscribe(state => { currentState = state; })();
  return currentState!;
}

// ✅ 修复
get state() {
  let currentState = initialState;
  const unsubscribe = store.subscribe(state => { currentState = state; });
  unsubscribe();
  return currentState;
}
```

### 开发检查清单

在开发或修复代码时，请按以下清单检查：

1. **[ ] 使用 Svelte 5 runes 语法**
   - `$state()` 用于响应式变量
   - `$derived()` 用于计算属性
   - `$effect()` 用于副作用
   - `$props()` 用于组件属性

2. **[ ] 路由跳转使用正确格式**
   - 使用 `/path` 而不是 `/(group)/path`
   - 确保路径与文件结构匹配

3. **[ ] Store 使用正确模式**
   - 正确订阅和取消订阅
   - 避免内存泄漏

4. **[ ] TypeScript 类型安全**
   - 定义明确的接口
   - 使用类型断言谨慎

5. **[ ] 运行验证命令**
   ```bash
   pnpm run check     # TypeScript 检查
   pnpm run format    # 代码格式化
   pnpm run lint      # 代码检查
   ```

### 版本兼容性说明

**当前项目配置**：
- 项目使用 Svelte 5 + runes 模式
- SvelteKit 配置支持路由组（文件组织）
- TypeScript 严格模式启用
- ESLint + Prettier 配置完整

**升级注意事项**：
- 从 Svelte 4 迁移时必须重写所有响应式语法
- 路由跳转逻辑需要检查和修正
- Store 使用模式需要适配 Svelte 5

## AI 助手注意事项

- 这个项目是一个重大重构工作 - 许多功能以框架形式存在但缺乏实现
- 参考原始 ScopeSentry 代码库了解功能需求和实现模式
- 优先考虑核心安全功能而非 UI 优化
- 在整个技术栈中保持类型安全
- 遵循已建立的架构模式以保持一致性
- 始终使用开发验证脚本测试新功能
- **重要：所有交流和输出都应使用中文**
- **严格遵循上述技术规范，避免版本兼容性问题**