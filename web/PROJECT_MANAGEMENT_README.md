# 项目管理功能 - 使用指南

本文档说明新实现的项目管理功能的使用方法和技术细节。

## 🎯 功能概述

项目管理模块是 Stellar 平台的核心功能之一，提供了完整的项目生命周期管理：

- ✅ 项目创建、编辑、删除
- ✅ 项目列表展示和筛选
- ✅ 项目详情查看
- ✅ 项目统计数据
- ✅ 项目成员管理
- ✅ 项目活动日志
- ✅ 数据导出功能

## 📁 文件结构

```
web/src/
├── lib/
│   ├── types/project.ts              # 项目类型定义
│   ├── api/projects.ts               # 项目API客户端
│   ├── components/project/           # 项目相关组件
│   │   ├── ProjectCard.svelte       # 项目卡片组件
│   │   └── index.ts                 # 组件导出
│   └── dev/verify-project-features.ts # 功能验证脚本
└── routes/(app)/projects/            # 项目页面路由
    ├── +page.svelte                 # 项目列表页
    ├── +page.ts                     # 列表页数据加载
    ├── create/+page.svelte          # 创建项目页
    └── [id]/
        ├── +page.svelte             # 项目详情页
        └── +page.ts                 # 详情页数据加载
```

## 🚀 快速开始

### 1. 访问项目管理

启动开发服务器后，访问以下路径：

- 项目列表：`http://localhost:5173/projects`
- 创建项目：`http://localhost:5173/projects/create`
- 项目详情：`http://localhost:5173/projects/{id}`

### 2. 创建第一个项目

1. 点击导航栏中的"项目管理"
2. 点击"创建项目"按钮
3. 填写项目信息：
   - **项目名称**（必填）：项目的唯一标识名称
   - **项目描述**（可选）：详细描述项目目的和范围
   - **目标地址**（可选）：要扫描的域名或IP地址
   - **项目颜色**：用于界面区分的颜色标识
   - **私有项目**：是否设为私有项目
4. 点击"创建项目"完成

### 3. 管理项目

在项目列表页面，可以进行以下操作：

- **查看详情**：点击项目卡片或"查看详情"按钮
- **编辑项目**：通过右上角的"更多"菜单
- **复制项目**：快速创建相似项目
- **删除项目**：永久删除项目（不可逆）
- **导出数据**：下载项目数据（JSON/CSV/XLSX格式）

## 🛠 技术实现

### API 客户端

项目API客户端(`ProjectAPI`)提供了完整的CRUD操作：

```typescript
import { ProjectAPI } from '$lib/api/projects';

// 获取项目列表
const projects = await ProjectAPI.getProjects({
  page: 1,
  limit: 20,
  search: '搜索关键词'
});

// 创建项目
const newProject = await ProjectAPI.createProject({
  name: '项目名称',
  description: '项目描述',
  target: 'example.com'
});

// 获取项目详情
const project = await ProjectAPI.getProject(projectId);
```

### 类型安全

所有项目数据都有完整的TypeScript类型定义：

```typescript
interface Project {
  id: string;
  name: string;
  description?: string;
  target?: string;
  scan_status?: string;
  color?: string;
  is_private?: boolean;
  assets_count?: number;
  vulnerabilities_count?: number;
  tasks_count?: number;
  created_by?: string;
  created_at: string;
  updated_at: string;
}
```

### 状态管理

遵循 Svelte 5 的最佳实践：

- 使用 `$state()` 管理组件内部状态
- 使用 `$props()` 接收父组件传递的数据
- 使用 `$effect()` 处理副作用操作

### 错误处理

统一的错误处理机制：

- API错误通过通知系统显示给用户
- 网络错误和超时有友好的提示
- 表单验证提供实时反馈

## 🧪 功能验证

项目包含完整的功能验证脚本，可以在开发环境中使用：

```typescript
import { quickVerifyProjectFeatures } from '$lib/dev/verify-project-features';

// 在浏览器控制台中运行
await quickVerifyProjectFeatures();
```

验证内容包括：

- ✅ TypeScript 类型定义
- ✅ API 客户端配置
- ✅ CRUD 操作完整性
- ✅ 数据持久化
- ✅ 错误处理机制

## 🎨 UI/UX 设计

### 设计原则

- **直观性**：清晰的视觉层次和操作流程
- **一致性**：统一的设计语言和交互模式
- **响应式**：适配不同屏幕尺寸
- **可访问性**：支持键盘导航和屏幕阅读器

### 组件库

使用 shadcn-svelte 组件库确保：

- 现代化的视觉设计
- 良好的可访问性支持
- 一致的交互体验
- 易于维护和扩展

## 🔗 API 集成

### 后端 API 规范

项目管理功能与后端 API 的集成遵循 RESTful 设计：

```
GET    /api/v1/projects          # 获取项目列表
POST   /api/v1/projects          # 创建项目
GET    /api/v1/projects/{id}     # 获取项目详情
PUT    /api/v1/projects/{id}     # 更新项目
DELETE /api/v1/projects/{id}     # 删除项目
GET    /api/v1/projects/stats    # 获取项目统计
```

### 认证授权

- 使用 JWT 令牌进行身份验证
- 请求自动添加 Authorization 头
- 支持令牌刷新机制
- 权限检查基于用户角色

## 📱 移动端适配

项目管理界面完全支持移动设备：

- 响应式布局自动适配
- 触摸友好的交互设计
- 优化的移动端导航
- 压缩的信息展示

## 🔧 开发指南

### 添加新功能

1. 在 `types/project.ts` 中定义新的类型
2. 在 `api/projects.ts` 中添加API方法
3. 创建或更新相关的Svelte组件
4. 添加路由页面（如需要）
5. 更新验证脚本

### 代码规范

- 严格遵循 TypeScript 类型检查
- 使用 ESLint 和 Prettier 格式化代码
- 组件使用 Svelte 5 语法
- API 调用使用统一的错误处理

### 测试策略

- 类型检查确保编译时安全
- 功能验证脚本测试运行时行为
- 手动测试覆盖用户交互场景
- 集成测试验证前后端协作

## 🚧 待开发功能

以下功能已在架构中预留，待后续开发：

- [ ] 高级筛选和排序
- [ ] 批量操作（批量删除、移动等）
- [ ] 项目模板系统
- [ ] 项目克隆和分支
- [ ] 详细的权限管理
- [ ] 项目标签系统
- [ ] 项目归档功能
- [ ] 更丰富的导出格式

## 📞 支持与反馈

如果在使用过程中遇到问题或有功能建议：

1. 检查浏览器控制台的错误信息
2. 运行功能验证脚本定位问题
3. 查看网络请求确认API连接状态
4. 参考本文档的故障排除部分

---

**注意**：这是 Stellar 项目管理功能的初始版本，我们会根据用户反馈持续改进和扩展功能。 