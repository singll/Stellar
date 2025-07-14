#!/bin/bash

# 测试敏感信息检测前端页面的简单脚本

echo "=== 敏感信息检测前端页面测试 ==="
echo

# 检查页面文件是否存在
echo "检查页面文件..."

pages=(
    "/root/Stellar/web/src/routes/(app)/sensitive/+page.svelte"
    "/root/Stellar/web/src/routes/(app)/sensitive/[id]/+page.svelte"
    "/root/Stellar/web/src/routes/(app)/sensitive/create/+page.svelte"
    "/root/Stellar/web/src/routes/(app)/sensitive/rules/+page.svelte"
    "/root/Stellar/web/src/lib/types/sensitive.ts"
    "/root/Stellar/web/src/lib/api/sensitive.ts"
)

for page in "${pages[@]}"; do
    if [[ -f "$page" ]]; then
        echo "✓ $page"
    else
        echo "✗ $page (不存在)"
    fi
done

echo
echo "检查Svelte语法..."

# 检查是否使用了正确的Svelte 5语法
echo "验证Svelte 5 runes语法..."

# 检查$state的使用
if grep -q "\$state" /root/Stellar/web/src/routes/\(app\)/sensitive/+page.svelte; then
    echo "✓ 正确使用 \$state"
else
    echo "⚠ 未发现 \$state 使用"
fi

# 检查$derived的使用
if grep -q "\$derived" /root/Stellar/web/src/routes/\(app\)/sensitive/+page.svelte; then
    echo "✓ 正确使用 \$derived"
else
    echo "⚠ 未发现 \$derived 使用"
fi

# 检查路由跳转是否正确
echo
echo "检查路由跳转..."

if grep -q "goto('/sensitive')" /root/Stellar/web/src/routes/\(app\)/sensitive/\[id\]/+page.svelte; then
    echo "✓ 路由跳转格式正确"
else
    echo "⚠ 可能存在路由跳转问题"
fi

# 检查组件导入
echo
echo "检查UI组件导入..."

required_components=(
    "Badge"
    "Button" 
    "Card"
    "Input"
    "Tabs"
)

sensitive_page="/root/Stellar/web/src/routes/(app)/sensitive/+page.svelte"

for component in "${required_components[@]}"; do
    if grep -q "import.*$component" "$sensitive_page"; then
        echo "✓ $component 组件已导入"
    else
        echo "⚠ $component 组件可能未导入"
    fi
done

echo
echo "检查类型定义..."

# 检查类型定义是否完整
types_file="/root/Stellar/web/src/lib/types/sensitive.ts"

required_types=(
    "SensitiveRule"
    "SensitiveRuleGroup"
    "SensitiveDetectionResult"
    "SensitiveFinding"
    "SensitiveDetectionRequest"
)

for type in "${required_types[@]}"; do
    if grep -q "interface $type" "$types_file"; then
        echo "✓ $type 类型已定义"
    else
        echo "⚠ $type 类型可能未定义"
    fi
done

echo
echo "检查API客户端..."

# 检查API函数是否完整
api_file="/root/Stellar/web/src/lib/api/sensitive.ts"

required_functions=(
    "getSensitiveDetectionResults"
    "getSensitiveDetectionResult" 
    "createSensitiveDetection"
    "getSensitiveRules"
    "getSensitiveRuleGroups"
)

for func in "${required_functions[@]}"; do
    if grep -q "export.*function $func" "$api_file"; then
        echo "✓ $func 函数已定义"
    else
        echo "⚠ $func 函数可能未定义"
    fi
done

echo
echo "=== 页面功能特性检查 ==="

# 检查主要功能特性
echo "主页面功能:"

if grep -q "搜索检测" "$sensitive_page"; then
    echo "✓ 检测搜索功能"
fi

if grep -q "过滤" "$sensitive_page"; then
    echo "✓ 检测过滤功能"
fi

if grep -q "统计" "$sensitive_page"; then
    echo "✓ 统计信息显示"
fi

if grep -q "deleteResult" "$sensitive_page"; then
    echo "✓ 检测结果删除功能"
fi

echo
echo "详情页面功能:"

detail_page="/root/Stellar/web/src/routes/(app)/sensitive/[id]/+page.svelte"

if grep -q "发现详情" "$detail_page"; then
    echo "✓ 敏感信息发现显示"
fi

if grep -q "风险统计" "$detail_page"; then
    echo "✓ 风险统计显示"
fi

if grep -q "TabsContent" "$detail_page"; then
    echo "✓ 标签页布局"
fi

if grep -q "exportResults" "$detail_page"; then
    echo "✓ 结果导出功能"
fi

echo
echo "创建页面功能:"

create_page="/root/Stellar/web/src/routes/(app)/sensitive/create/+page.svelte"

if grep -q "检测目标" "$create_page"; then
    echo "✓ 目标配置功能"
fi

if grep -q "检测规则" "$create_page"; then
    echo "✓ 规则选择功能"
fi

if grep -q "testRules" "$create_page"; then
    echo "✓ 规则测试功能"
fi

if grep -q "validateForm" "$create_page"; then
    echo "✓ 表单验证功能"
fi

echo
echo "规则管理页面功能:"

rules_page="/root/Stellar/web/src/routes/(app)/sensitive/rules/+page.svelte"

if grep -q "规则组" "$rules_page"; then
    echo "✓ 规则组管理功能"
fi

if grep -q "toggleRule" "$rules_page"; then
    echo "✓ 规则启用/禁用功能"
fi

if grep -q "风险等级" "$rules_page"; then
    echo "✓ 风险等级显示"
fi

# 检查导航菜单是否更新
echo
echo "检查导航菜单:"

layout_file="/root/Stellar/web/src/routes/(app)/+layout.svelte"

if grep -q "敏感信息检测" "$layout_file"; then
    echo "✓ 导航菜单已添加敏感信息检测"
else
    echo "⚠ 导航菜单可能未添加敏感信息检测"
fi

echo
echo "=== 检查完成 ==="
echo "建议手动测试页面功能和UI交互效果"