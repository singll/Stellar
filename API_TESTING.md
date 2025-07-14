# Stellar API 测试配置

## 测试环境配置

### 本地开发环境
```bash
# 服务器配置
API_BASE_URL=http://localhost:8090
DB_HOST=192.168.7.216
DB_PORT=27017
REDIS_HOST=192.168.7.128
REDIS_PORT=6379

# 测试用户
TEST_USER_EMAIL=test@example.com
TEST_USER_PASSWORD=testpass123
```

### 测试用例清单

#### 1. 基础功能测试
- [x] 服务器连接测试
- [x] 用户认证测试
- [x] JWT令牌验证
- [x] 权限验证

#### 2. 项目管理测试
- [x] 创建项目
- [x] 获取项目列表
- [x] 获取项目详情
- [x] 更新项目信息
- [x] 删除项目

#### 3. 任务管理测试
- [x] 创建扫描任务
- [x] 获取任务列表
- [x] 获取任务详情
- [x] 更新任务状态
- [x] 取消任务
- [x] 重试任务

#### 4. 子域名枚举测试
- [x] 创建子域名枚举任务
- [x] 验证配置参数
- [x] 检查执行状态
- [x] 获取枚举结果
- [x] 验证结果格式

#### 5. 端口扫描测试
- [x] 创建端口扫描任务
- [x] 验证扫描配置
- [x] 检查扫描进度
- [x] 获取扫描结果
- [x] 验证结果完整性

#### 6. 结果导出测试
- [x] CSV格式导出
- [x] JSON格式导出
- [x] 大数据量导出
- [x] 导出文件完整性

#### 7. 节点管理测试
- [x] 获取节点列表
- [x] 节点状态查询
- [x] 节点健康检查
- [x] 节点负载信息

#### 8. 资产管理测试
- [x] 获取资产列表
- [x] 资产详情查询
- [x] 资产分类统计
- [x] 资产搜索过滤

## 性能测试

### 并发测试
```bash
# 使用Apache Bench进行并发测试
ab -n 1000 -c 10 -H "Authorization: Bearer <jwt_token>" http://localhost:8090/api/v1/projects

# 使用curl进行压力测试
for i in {1..100}; do
  curl -s -X GET "http://localhost:8090/api/tasks" \
    -H "Authorization: Bearer <jwt_token>" &
done
wait
```

### 负载测试
```bash
# 创建多个并发任务
for i in {1..10}; do
  curl -s -X POST "http://localhost:8090/api/tasks" \
    -H "Authorization: Bearer <jwt_token>" \
    -H "Content-Type: application/json" \
    -d '{
      "name": "负载测试任务'$i'",
      "type": "subdomain_enum",
      "projectId": "'$PROJECT_ID'",
      "params": {"target": "example'$i'.com"}
    }' &
done
wait
```

## 自动化测试

### 使用测试脚本
```bash
# 运行完整的API测试
./test_api.sh

# 运行特定测试
./test_api.sh --test authentication
./test_api.sh --test task_management
./test_api.sh --test export_results
```

### 集成到CI/CD
```yaml
# GitHub Actions 示例
name: API Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.19
      - name: Start services
        run: |
          make dev &
          sleep 30
      - name: Run API tests
        run: ./test_api.sh
```

## 测试数据

### 测试用户数据
```json
{
  "email": "test@example.com",
  "password": "testpass123",
  "role": "admin"
}
```

### 测试项目数据
```json
{
  "name": "API测试项目",
  "description": "用于API测试的项目",
  "targets": ["example.com", "test.com"],
  "created_by": "test_user"
}
```

### 测试任务数据
```json
{
  "subdomain_enum": {
    "name": "子域名枚举测试",
    "type": "subdomain_enum",
    "params": {
      "target": "example.com",
      "max_workers": 10,
      "timeout": 30,
      "enum_methods": ["dns_brute"]
    }
  },
  "port_scan": {
    "name": "端口扫描测试",
    "type": "port_scan",
    "params": {
      "target": "example.com",
      "ports": "80,443,8080",
      "scan_method": "tcp",
      "max_workers": 20
    }
  }
}
```

## 预期结果

### 子域名枚举结果
```json
{
  "subdomain": "www",
  "domain": "www.example.com",
  "ip": "93.184.216.34",
  "status": "valid",
  "source": "dns_brute",
  "response_time": 45
}
```

### 端口扫描结果
```json
{
  "host": "example.com",
  "port": 80,
  "protocol": "tcp",
  "status": "open",
  "service": "http",
  "version": "nginx/1.18.0",
  "banner": "HTTP/1.1 200 OK"
}
```

## 错误处理测试

### 常见错误场景
1. **无效令牌**: 返回401 Unauthorized
2. **缺少参数**: 返回400 Bad Request
3. **资源不存在**: 返回404 Not Found
4. **权限不足**: 返回403 Forbidden
5. **服务器错误**: 返回500 Internal Server Error

### 错误响应格式验证
```json
{
  "success": false,
  "error": "错误描述",
  "code": "ERROR_CODE",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## 安全测试

### 认证测试
- JWT令牌有效性验证
- 令牌过期处理
- 刷新令牌机制

### 授权测试
- 角色权限验证
- 资源访问控制
- 跨项目访问限制

### 输入验证测试
- SQL注入防护
- XSS防护
- 参数验证

## 测试报告

测试完成后，应生成详细的测试报告，包括：
- 测试用例执行情况
- 性能指标统计
- 错误日志分析
- 改进建议

执行 `./test_api.sh` 后会自动生成测试报告。