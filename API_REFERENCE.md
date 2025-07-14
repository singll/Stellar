# Stellar API 接口文档

## 概述

Stellar 是一个分布式安全资产管理和漏洞扫描平台，提供 RESTful API 接口用于管理扫描任务、查看结果等功能。

## 认证

所有API请求（除了认证接口）都需要在请求头中包含JWT令牌：

```
Authorization: Bearer <jwt_token>
```

## 基础响应格式

### 成功响应
```json
{
  "success": true,
  "data": <response_data>,
  "message": "操作成功"
}
```

### 错误响应
```json
{
  "success": false,
  "error": "错误描述",
  "code": <error_code>
}
```

## 核心API接口

### 任务管理 API

#### 1. 创建任务
- **URL**: `POST /api/tasks`
- **描述**: 创建新的扫描任务
- **请求体**:
```json
{
  "name": "任务名称",
  "description": "任务描述",
  "type": "subdomain_enum|port_scan|vuln_scan",
  "projectId": "项目ID",
  "priority": 1,
  "timeout": 3600,
  "params": {
    "target": "example.com",
    "config": {}
  }
}
```

#### 2. 获取任务列表
- **URL**: `GET /api/tasks`
- **描述**: 获取任务列表
- **查询参数**:
  - `projectId`: 项目ID（可选）
  - `status`: 任务状态（可选）
  - `type`: 任务类型（可选）
  - `page`: 页码（默认1）
  - `limit`: 每页数量（默认20）

#### 3. 获取任务详情
- **URL**: `GET /api/tasks/{id}`
- **描述**: 获取指定任务的详细信息

#### 4. 获取任务结果
- **URL**: `GET /api/tasks/{id}/results`
- **描述**: 获取任务的详细结果
- **查询参数**:
  - `page`: 页码（默认1）
  - `limit`: 每页数量（默认50）

#### 5. 导出任务结果
- **URL**: `GET /api/tasks/{id}/export`
- **描述**: 导出任务结果
- **查询参数**:
  - `format`: 导出格式（csv|json，默认csv）

#### 6. 取消任务
- **URL**: `POST /api/tasks/{id}/cancel`
- **描述**: 取消正在执行的任务

#### 7. 重试任务
- **URL**: `POST /api/tasks/{id}/retry`
- **描述**: 重新执行失败的任务

### 项目管理 API

#### 1. 创建项目
- **URL**: `POST /api/v1/projects`
- **描述**: 创建新项目

#### 2. 获取项目列表
- **URL**: `GET /api/v1/projects`
- **描述**: 获取项目列表

#### 3. 获取项目详情
- **URL**: `GET /api/v1/projects/{id}`
- **描述**: 获取项目详情

### 资产管理 API

#### 1. 获取资产列表
- **URL**: `GET /api/v1/assets`
- **描述**: 获取资产列表

#### 2. 获取资产详情
- **URL**: `GET /api/v1/assets/{id}`
- **描述**: 获取资产详情

### 节点管理 API

#### 1. 获取节点列表
- **URL**: `GET /api/v1/nodes`
- **描述**: 获取节点列表

#### 2. 获取节点状态
- **URL**: `GET /api/v1/nodes/{id}/status`
- **描述**: 获取节点状态

## 子域名枚举 API

### 任务配置参数
```json
{
  "target": "example.com",
  "max_workers": 50,
  "timeout": 30,
  "wordlist_path": "/path/to/wordlist.txt",
  "dns_servers": ["8.8.8.8", "1.1.1.1"],
  "enable_wildcard": true,
  "max_retries": 3,
  "enum_methods": ["dns_brute", "certificate_transparency"],
  "rate_limit": 10,
  "enable_doh": false,
  "enable_recursive": false,
  "max_depth": 2,
  "verify_subdomains": true
}
```

### 结果格式
```json
{
  "subdomain": "www",
  "domain": "www.example.com",
  "ip": "93.184.216.34",
  "status": "valid",
  "source": "dns_brute",
  "response_time": 45,
  "created_at": "2024-01-01T12:00:00Z"
}
```

## 端口扫描 API

### 任务配置参数
```json
{
  "target": "example.com",
  "ports": "80,443,8080-8090",
  "scan_method": "tcp",
  "max_workers": 50,
  "timeout": 30,
  "enable_banner": true,
  "enable_ssl": true,
  "enable_service": true,
  "rate_limit": 100
}
```

### 结果格式
```json
{
  "host": "example.com",
  "port": 80,
  "protocol": "tcp",
  "status": "open",
  "service": "http",
  "version": "nginx/1.18.0",
  "response_time": 23,
  "banner": "HTTP/1.1 200 OK\r\nServer: nginx/1.18.0"
}
```

## 状态码说明

- **200 OK**: 请求成功
- **201 Created**: 资源创建成功
- **400 Bad Request**: 请求参数错误
- **401 Unauthorized**: 未授权访问
- **403 Forbidden**: 访问被拒绝
- **404 Not Found**: 资源不存在
- **500 Internal Server Error**: 服务器内部错误

## 错误处理

### 常见错误码
- `INVALID_REQUEST`: 请求参数无效
- `UNAUTHORIZED`: 未授权访问
- `RESOURCE_NOT_FOUND`: 资源不存在
- `TASK_EXECUTION_FAILED`: 任务执行失败
- `RATE_LIMIT_EXCEEDED`: 请求频率超限

### 错误响应示例
```json
{
  "success": false,
  "error": "任务ID不能为空",
  "code": "INVALID_REQUEST",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## 限流说明

- 每个用户每分钟最多发起 100 次API请求
- 每个项目最多同时运行 10 个任务
- 单个任务最大执行时间为 24 小时

## 示例用法

### 创建子域名枚举任务
```bash
curl -X POST \
  http://localhost:8090/api/tasks \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "example.com子域名枚举",
    "type": "subdomain_enum",
    "projectId": "507f1f77bcf86cd799439011",
    "params": {
      "target": "example.com",
      "max_workers": 50,
      "timeout": 30,
      "enum_methods": ["dns_brute", "certificate_transparency"]
    }
  }'
```

### 获取任务结果
```bash
curl -X GET \
  "http://localhost:8090/api/tasks/507f1f77bcf86cd799439012/results?page=1&limit=50" \
  -H "Authorization: Bearer <jwt_token>"
```

### 导出任务结果
```bash
curl -X GET \
  "http://localhost:8090/api/tasks/507f1f77bcf86cd799439012/export?format=csv" \
  -H "Authorization: Bearer <jwt_token>" \
  -o results.csv
```

## 开发和测试

### 本地开发环境
1. 克隆项目代码
2. 安装依赖：`go mod tidy`
3. 配置数据库连接
4. 启动服务：`make dev`

### API测试
推荐使用以下工具进行API测试：
- **Postman**: 图形化API测试工具
- **curl**: 命令行HTTP客户端
- **HTTPie**: 现代化的命令行HTTP客户端

### 测试用例
项目提供了完整的API测试用例，位于 `tests/api/` 目录下。

## 更新日志

### v1.0.0 (2024-01-01)
- 初始版本发布
- 支持子域名枚举、端口扫描、漏洞扫描
- 完整的任务管理系统
- 结果导出功能
- 分布式节点管理

## 联系方式

如有问题或建议，请联系开发团队或提交Issue。