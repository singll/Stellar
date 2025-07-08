# Stellar 开发环境启动指南

本文档描述了如何使用新的启动系统来管理 Stellar 开发环境。

## 快速开始

### 1. 检查系统依赖
```bash
make check-deps
```

### 2. 安装项目依赖
```bash
make install-deps
```

### 3. 启动开发环境
```bash
make dev
```

### 4. 查看状态
```bash
make status
```

## 命令说明

### 开发命令

| 命令 | 功能 | 说明 |
|------|------|------|
| `make dev` | 启动开发环境 | 自动清理端口、启动前后端服务 |
| `make dev-stop` | 停止开发环境 | 安全停止所有相关进程 |
| `make status` | 查看状态 | 显示端口占用和日志信息 |
| `make logs` | 查看实时日志 | 同时显示前后端日志 |

### 依赖管理

| 命令 | 功能 | 说明 |
|------|------|------|
| `make check-deps` | 检查依赖 | 验证系统工具和数据库连接 |
| `make install-deps` | 安装依赖 | 安装Go模块和前端依赖 |

### 构建和清理

| 命令 | 功能 | 说明 |
|------|------|------|
| `make build` | 构建生产版本 | 编译前端并嵌入到Go二进制文件 |
| `make clean` | 清理构建产物 | 删除临时文件和日志 |
| `make clean-all` | 深度清理 | 额外删除依赖缓存 |

### 直接脚本调用

如果您不使用 Make，可以直接调用脚本：

#### Linux/macOS
```bash
# 启动开发环境
./scripts/dev-start.sh

# 停止开发环境
./scripts/dev-stop.sh

# 健康检查
./scripts/health-check.sh
```

#### Windows
```batch
REM 启动开发环境
scripts\dev-start.bat

REM 停止开发环境
scripts\dev-stop.bat
```

## 特性说明

### 自动端口清理

启动脚本会自动检查并清理占用的端口：
- 后端端口：8090
- 前端端口：5173

如果发现端口被占用，脚本会：
1. 显示占用进程信息
2. 尝试优雅终止进程（TERM信号）
3. 如果需要，强制终止进程（KILL信号）
4. 验证端口已释放

### 服务健康检查

启动脚本包含以下检查：
- 系统工具可用性（Go、pnpm、lsof）
- 数据库连接性（MongoDB、Redis）
- 前端依赖完整性
- 服务启动状态验证

### 日志管理

所有日志文件存储在 `logs/` 目录：
- `logs/backend.log` - 后端服务日志
- `logs/frontend.log` - 前端开发服务器日志
- `logs/backend.pid` - 后端进程ID文件
- `logs/frontend.pid` - 前端进程ID文件

### 错误处理

脚本包含完善的错误处理：
- 超时检测（服务启动最多等待30秒）
- 进程状态监控
- 优雅的错误恢复
- 详细的错误信息

## 故障排除

### 端口占用问题

如果遇到端口占用问题：

1. 查看哪个进程占用端口：
   ```bash
   lsof -Pi :8090 -sTCP:LISTEN  # 后端端口
   lsof -Pi :5173 -sTCP:LISTEN  # 前端端口
   ```

2. 手动停止开发环境：
   ```bash
   make dev-stop
   ```

3. 强制清理端口（如果必要）：
   ```bash
   sudo lsof -ti :8090 | xargs kill -9  # 强制清理后端端口
   sudo lsof -ti :5173 | xargs kill -9  # 强制清理前端端口
   ```

### 数据库连接问题

如果数据库连接失败：

1. 检查数据库服务状态：
   ```bash
   make check-deps
   ```

2. 验证网络连接：
   ```bash
   ping 192.168.7.216  # MongoDB主机
   ping 192.168.7.128  # Redis主机
   ```

3. 检查配置文件：
   ```bash
   cat config.dev.yaml
   ```

### 依赖问题

如果遇到依赖问题：

1. 重新安装依赖：
   ```bash
   make clean-all
   make install-deps
   ```

2. 检查Node.js版本：
   ```bash
   node --version  # 需要18+
   pnpm --version
   ```

3. 检查Go版本：
   ```bash
   go version  # 需要1.19+
   ```

### 前端构建问题

如果前端构建失败：

1. 清理前端缓存：
   ```bash
   cd web
   rm -rf node_modules .svelte-kit dist
   pnpm install
   ```

2. 检查TypeScript错误：
   ```bash
   cd web
   pnpm run check
   ```

## 性能优化

### 开发环境

- 前端使用Vite的热重载功能
- 后端使用`go run`直接运行源码
- 日志输出到文件，减少终端显示延迟

### 系统资源监控

使用健康检查脚本监控系统资源：
```bash
./scripts/health-check.sh
```

该脚本会报告：
- 内存使用率
- 磁盘使用率
- CPU负载
- 进程状态

## 配置说明

### 开发配置文件

主要配置文件：`config.dev.yaml`

关键配置项：
- 服务器端口：8090
- 数据库连接：MongoDB (192.168.7.216:27017)
- 缓存连接：Redis (192.168.7.128:6379)
- JWT密钥：开发专用密钥

### 前端配置

前端配置在 `web/` 目录：
- `vite.config.ts` - Vite配置
- `tsconfig.json` - TypeScript配置
- `tailwind.config.js` - Tailwind CSS配置

## 生产部署

生产环境构建：
```bash
make build
```

这会：
1. 构建前端静态文件
2. 将前端文件嵌入到Go二进制文件
3. 生成优化的`stellar-server`可执行文件

运行生产版本：
```bash
./stellar-server -config config.yaml
```

## 开发工作流建议

### 日常开发
1. `make dev` - 启动开发环境
2. 进行开发工作
3. `make logs` - 查看日志（如需要）
4. `make dev-stop` - 结束时停止服务

### 版本发布前
1. `make clean` - 清理旧文件
2. `make check-deps` - 验证依赖
3. `make build` - 构建生产版本
4. 测试生产版本

### 问题调试
1. `make status` - 查看服务状态
2. `./scripts/health-check.sh` - 运行健康检查
3. `make logs` - 查看详细日志
4. 根据错误信息进行相应处理