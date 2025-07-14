#!/bin/bash

# 测试插件管理前端页面的简单脚本

echo "=== 插件管理前端页面测试 ==="
echo

# 检查页面文件是否存在
echo "检查页面文件..."

pages=(
    "/root/Stellar/web/src/routes/(app)/plugins/+page.svelte"
    "/root/Stellar/web/src/routes/(app)/plugins/[id]/+page.svelte"
    "/root/Stellar/web/src/routes/(app)/plugins/install/+page.svelte"
    "/root/Stellar/web/src/lib/types/plugin.ts"
    "/root/Stellar/web/src/lib/api/plugin.ts"
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
if grep -q "\$state" /root/Stellar/web/src/routes/\(app\)/plugins/+page.svelte; then
    echo "✓ 正确使用 \$state"
else
    echo "⚠ 未发现 \$state 使用"
fi

# 检查$derived的使用
if grep -q "\$derived" /root/Stellar/web/src/routes/\(app\)/plugins/+page.svelte; then
    echo "✓ 正确使用 \$derived"
else
    echo "⚠ 未发现 \$derived 使用"
fi

# 检查路由跳转是否正确
echo
echo "检查路由跳转..."

if grep -q "goto('/plugins')" /root/Stellar/web/src/routes/\(app\)/plugins/\[id\]/+page.svelte; then
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

plugin_page="/root/Stellar/web/src/routes/(app)/plugins/+page.svelte"

for component in "${required_components[@]}"; do
    if grep -q "import.*$component" "$plugin_page"; then
        echo "✓ $component 组件已导入"
    else
        echo "⚠ $component 组件可能未导入"
    fi
done

echo
echo "检查类型定义..."

# 检查类型定义是否完整
types_file="/root/Stellar/web/src/lib/types/plugin.ts"

required_types=(
    "PluginMetadata"
    "PluginRunRecord"
    "PluginConfig"
    "YAMLPluginConfig"
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
api_file="/root/Stellar/web/src/lib/api/plugin.ts"

required_functions=(
    "getPlugins"
    "getPlugin"
    "installPlugin"
    "togglePlugin"
    "deletePlugin"
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

if grep -q "搜索插件" "$plugin_page"; then
    echo "✓ 插件搜索功能"
fi

if grep -q "过滤" "$plugin_page"; then
    echo "✓ 插件过滤功能"
fi

if grep -q "统计" "$plugin_page"; then
    echo "✓ 统计信息显示"
fi

if grep -q "togglePlugin" "$plugin_page"; then
    echo "✓ 插件启用/禁用功能"
fi

echo
echo "详情页面功能:"

detail_page="/root/Stellar/web/src/routes/(app)/plugins/[id]/+page.svelte"

if grep -q "运行记录" "$detail_page"; then
    echo "✓ 运行记录显示"
fi

if grep -q "TabsContent" "$detail_page"; then
    echo "✓ 标签页布局"
fi

if grep -q "配置" "$detail_page"; then
    echo "✓ 插件配置管理"
fi

echo
echo "安装页面功能:"

install_page="/root/Stellar/web/src/routes/(app)/plugins/install/+page.svelte"

if grep -q "文件上传" "$install_page"; then
    echo "✓ 文件上传安装"
fi

if grep -q "YAML" "$install_page"; then
    echo "✓ YAML配置安装"
fi

if grep -q "URL" "$install_page"; then
    echo "✓ URL下载安装"
fi

echo
echo "=== 检查完成 ==="
echo "建议手动测试页面功能和UI交互效果"