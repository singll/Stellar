#!/bin/bash

# Stellar API 验收测试脚本
# 用于验证API接口的基本功能

# 配置
API_BASE_URL="http://localhost:8090"
TEST_USER_EMAIL="test@example.com"
TEST_USER_PASSWORD="testpass123"
JWT_TOKEN=""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查API响应
check_response() {
    local response=$1
    local expected_status=$2
    local test_name=$3
    
    local status=$(echo "$response" | jq -r '.status // empty')
    local success=$(echo "$response" | jq -r '.success // empty')
    
    if [[ "$status" == "$expected_status" ]] || [[ "$success" == "true" ]]; then
        log_success "$test_name: 通过"
        return 0
    else
        log_error "$test_name: 失败"
        echo "$response" | jq '.'
        return 1
    fi
}

# 1. 测试服务器连接
test_server_connection() {
    log_info "测试服务器连接..."
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$API_BASE_URL/health" 2>/dev/null)
    
    if [[ "$response" == "200" ]]; then
        log_success "服务器连接正常"
        return 0
    else
        log_error "服务器连接失败，HTTP状态码: $response"
        return 1
    fi
}

# 2. 测试用户认证
test_authentication() {
    log_info "测试用户认证..."
    
    # 尝试登录
    response=$(curl -s -X POST \
        "$API_BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_USER_EMAIL\",\"password\":\"$TEST_USER_PASSWORD\"}")
    
    JWT_TOKEN=$(echo "$response" | jq -r '.data.token // empty')
    
    if [[ -n "$JWT_TOKEN" && "$JWT_TOKEN" != "null" ]]; then
        log_success "用户认证成功，获取到JWT令牌"
        return 0
    else
        log_error "用户认证失败"
        echo "$response" | jq '.'
        return 1
    fi
}

# 3. 测试项目管理API
test_project_management() {
    log_info "测试项目管理API..."
    
    # 创建项目
    response=$(curl -s -X POST \
        "$API_BASE_URL/api/v1/projects" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "API测试项目",
            "description": "用于API测试的项目",
            "targets": ["example.com"]
        }')
    
    PROJECT_ID=$(echo "$response" | jq -r '.data.id // empty')
    
    if [[ -n "$PROJECT_ID" && "$PROJECT_ID" != "null" ]]; then
        log_success "项目创建成功，ID: $PROJECT_ID"
        
        # 获取项目列表
        response=$(curl -s -X GET \
            "$API_BASE_URL/api/v1/projects" \
            -H "Authorization: Bearer $JWT_TOKEN")
        
        check_response "$response" "" "获取项目列表"
        
        # 获取项目详情
        response=$(curl -s -X GET \
            "$API_BASE_URL/api/v1/projects/$PROJECT_ID" \
            -H "Authorization: Bearer $JWT_TOKEN")
        
        check_response "$response" "" "获取项目详情"
        
        return 0
    else
        log_error "项目创建失败"
        echo "$response" | jq '.'
        return 1
    fi
}

# 4. 测试任务管理API
test_task_management() {
    log_info "测试任务管理API..."
    
    if [[ -z "$PROJECT_ID" ]]; then
        log_error "需要先创建项目"
        return 1
    fi
    
    # 创建子域名枚举任务
    response=$(curl -s -X POST \
        "$API_BASE_URL/api/tasks" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"API测试-子域名枚举\",
            \"description\": \"用于API测试的子域名枚举任务\",
            \"type\": \"subdomain_enum\",
            \"projectId\": \"$PROJECT_ID\",
            \"priority\": 1,
            \"timeout\": 3600,
            \"params\": {
                \"target\": \"example.com\",
                \"max_workers\": 10,
                \"timeout\": 30,
                \"enum_methods\": [\"dns_brute\"]
            }
        }")
    
    TASK_ID=$(echo "$response" | jq -r '.id // empty')
    
    if [[ -n "$TASK_ID" && "$TASK_ID" != "null" ]]; then
        log_success "任务创建成功，ID: $TASK_ID"
        
        # 获取任务列表
        response=$(curl -s -X GET \
            "$API_BASE_URL/api/tasks" \
            -H "Authorization: Bearer $JWT_TOKEN")
        
        check_response "$response" "" "获取任务列表"
        
        # 获取任务详情
        response=$(curl -s -X GET \
            "$API_BASE_URL/api/tasks/$TASK_ID" \
            -H "Authorization: Bearer $JWT_TOKEN")
        
        check_response "$response" "" "获取任务详情"
        
        # 等待任务开始执行
        sleep 5
        
        # 获取任务结果
        response=$(curl -s -X GET \
            "$API_BASE_URL/api/tasks/$TASK_ID/results" \
            -H "Authorization: Bearer $JWT_TOKEN")
        
        check_response "$response" "" "获取任务结果"
        
        return 0
    else
        log_error "任务创建失败"
        echo "$response" | jq '.'
        return 1
    fi
}

# 5. 测试端口扫描API
test_port_scan() {
    log_info "测试端口扫描API..."
    
    if [[ -z "$PROJECT_ID" ]]; then
        log_error "需要先创建项目"
        return 1
    fi
    
    # 创建端口扫描任务
    response=$(curl -s -X POST \
        "$API_BASE_URL/api/tasks" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"API测试-端口扫描\",
            \"description\": \"用于API测试的端口扫描任务\",
            \"type\": \"port_scan\",
            \"projectId\": \"$PROJECT_ID\",
            \"priority\": 1,
            \"timeout\": 3600,
            \"params\": {
                \"target\": \"example.com\",
                \"ports\": \"80,443\",
                \"scan_method\": \"tcp\",
                \"max_workers\": 10,
                \"timeout\": 30
            }
        }")
    
    PORTSCAN_TASK_ID=$(echo "$response" | jq -r '.id // empty')
    
    if [[ -n "$PORTSCAN_TASK_ID" && "$PORTSCAN_TASK_ID" != "null" ]]; then
        log_success "端口扫描任务创建成功，ID: $PORTSCAN_TASK_ID"
        return 0
    else
        log_error "端口扫描任务创建失败"
        echo "$response" | jq '.'
        return 1
    fi
}

# 6. 测试结果导出API
test_export_results() {
    log_info "测试结果导出API..."
    
    if [[ -z "$TASK_ID" ]]; then
        log_error "需要先创建任务"
        return 1
    fi
    
    # 等待任务完成或有部分结果
    sleep 10
    
    # 测试CSV导出
    response=$(curl -s -X GET \
        "$API_BASE_URL/api/tasks/$TASK_ID/export?format=csv" \
        -H "Authorization: Bearer $JWT_TOKEN")
    
    if [[ -n "$response" ]]; then
        log_success "CSV导出测试通过"
    else
        log_error "CSV导出测试失败"
        return 1
    fi
    
    # 测试JSON导出
    response=$(curl -s -X GET \
        "$API_BASE_URL/api/tasks/$TASK_ID/export?format=json" \
        -H "Authorization: Bearer $JWT_TOKEN")
    
    if [[ -n "$response" ]]; then
        log_success "JSON导出测试通过"
    else
        log_error "JSON导出测试失败"
        return 1
    fi
    
    return 0
}

# 7. 测试节点管理API
test_node_management() {
    log_info "测试节点管理API..."
    
    # 获取节点列表
    response=$(curl -s -X GET \
        "$API_BASE_URL/api/v1/nodes" \
        -H "Authorization: Bearer $JWT_TOKEN")
    
    check_response "$response" "" "获取节点列表"
    
    return 0
}

# 8. 测试资产管理API
test_asset_management() {
    log_info "测试资产管理API..."
    
    # 获取资产列表
    response=$(curl -s -X GET \
        "$API_BASE_URL/api/v1/assets" \
        -H "Authorization: Bearer $JWT_TOKEN")
    
    check_response "$response" "" "获取资产列表"
    
    return 0
}

# 9. 清理测试数据
cleanup_test_data() {
    log_info "清理测试数据..."
    
    # 取消任务
    if [[ -n "$TASK_ID" ]]; then
        curl -s -X POST \
            "$API_BASE_URL/api/tasks/$TASK_ID/cancel" \
            -H "Authorization: Bearer $JWT_TOKEN" > /dev/null
    fi
    
    if [[ -n "$PORTSCAN_TASK_ID" ]]; then
        curl -s -X POST \
            "$API_BASE_URL/api/tasks/$PORTSCAN_TASK_ID/cancel" \
            -H "Authorization: Bearer $JWT_TOKEN" > /dev/null
    fi
    
    # 删除项目
    if [[ -n "$PROJECT_ID" ]]; then
        curl -s -X DELETE \
            "$API_BASE_URL/api/v1/projects/$PROJECT_ID" \
            -H "Authorization: Bearer $JWT_TOKEN" > /dev/null
    fi
    
    log_success "测试数据清理完成"
}

# 主测试函数
main() {
    log_info "开始Stellar API验收测试"
    log_info "测试目标: $API_BASE_URL"
    echo ""
    
    # 检查依赖
    if ! command -v curl &> /dev/null; then
        log_error "curl命令未找到，请安装curl"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_error "jq命令未找到，请安装jq"
        exit 1
    fi
    
    # 运行测试
    local failed_tests=0
    
    test_server_connection || ((failed_tests++))
    echo ""
    
    test_authentication || ((failed_tests++))
    echo ""
    
    if [[ -n "$JWT_TOKEN" ]]; then
        test_project_management || ((failed_tests++))
        echo ""
        
        test_task_management || ((failed_tests++))
        echo ""
        
        test_port_scan || ((failed_tests++))
        echo ""
        
        test_export_results || ((failed_tests++))
        echo ""
        
        test_node_management || ((failed_tests++))
        echo ""
        
        test_asset_management || ((failed_tests++))
        echo ""
        
        cleanup_test_data
        echo ""
    else
        log_error "跳过需要认证的测试"
        ((failed_tests++))
    fi
    
    # 输出测试结果
    if [[ $failed_tests -eq 0 ]]; then
        log_success "所有测试通过！"
        exit 0
    else
        log_error "有 $failed_tests 个测试失败"
        exit 1
    fi
}

# 运行主函数
main "$@"