#!/bin/bash

# YAML插件测试脚本
# 用于测试YAML插件的基本功能

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGINS_DIR="$SCRIPT_DIR/examples"

echo "=== Stellar YAML插件测试工具 ==="
echo

# 检查依赖
check_dependencies() {
    echo "检查依赖..."
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        echo "错误: 需要安装Go"
        exit 1
    fi
    
    # 检查Python3
    if ! command -v python3 &> /dev/null; then
        echo "警告: Python3未安装，Python插件将无法测试"
    fi
    
    # 检查Node.js
    if ! command -v node &> /dev/null; then
        echo "警告: Node.js未安装，JavaScript插件将无法测试"
    fi
    
    # 检查yaml解析工具
    if ! command -v yq &> /dev/null; then
        echo "警告: yq未安装，建议安装以便更好地解析YAML"
    fi
    
    echo "依赖检查完成"
    echo
}

# 解析YAML插件配置
parse_yaml_plugin() {
    local plugin_file="$1"
    
    if [[ ! -f "$plugin_file" ]]; then
        echo "错误: 插件文件不存在: $plugin_file"
        return 1
    fi
    
    echo "解析插件: $(basename "$plugin_file")"
    
    # 提取基本信息
    local plugin_id=$(grep "^id:" "$plugin_file" | sed 's/id: *//')
    local plugin_name=$(grep "^name:" "$plugin_file" | sed 's/name: *//')
    local plugin_version=$(grep "^version:" "$plugin_file" | sed 's/version: *//' | tr -d '"')
    local plugin_language=$(grep -A 10 "^script:" "$plugin_file" | grep "language:" | sed 's/.*language: *//')
    
    echo "  ID: $plugin_id"
    echo "  名称: $plugin_name"
    echo "  版本: $plugin_version"
    echo "  语言: $plugin_language"
    echo
}

# 提取脚本内容
extract_script() {
    local plugin_file="$1"
    local output_file="$2"
    
    # 使用awk提取content部分
    awk '
    /content: \|/ { 
        in_content = 1
        next
    }
    in_content && /^[[:space:]]*[^[:space:]]/ && !/^[[:space:]]*#/ && !/^[[:space:]]*$/ {
        if ($0 !~ /^[[:space:]]*content:/ && $0 !~ /^[[:space:]]*language:/ && $0 !~ /^[[:space:]]*entry:/) {
            in_content = 0
        }
    }
    in_content && /^[[:space:]]+/ {
        # 移除4个空格的缩进
        sub(/^[[:space:]]{4}/, "")
        print
    }
    ' "$plugin_file" > "$output_file"
}

# 测试Python插件
test_python_plugin() {
    local plugin_file="$1"
    local temp_dir=$(mktemp -d)
    local script_file="$temp_dir/plugin.py"
    
    echo "测试Python插件..."
    
    extract_script "$plugin_file" "$script_file"
    
    # 创建测试参数
    local test_params='{"domain": "example.com", "config": {"timeout": 10}}'
    
    echo "执行插件..."
    echo "$test_params" | python3 "$script_file" 2>/dev/null || {
        echo "插件执行失败，查看脚本内容:"
        echo "--- 脚本开始 ---"
        cat "$script_file"
        echo "--- 脚本结束 ---"
    }
    
    rm -rf "$temp_dir"
    echo
}

# 测试JavaScript插件
test_javascript_plugin() {
    local plugin_file="$1"
    local temp_dir=$(mktemp -d)
    local script_file="$temp_dir/plugin.js"
    
    echo "测试JavaScript插件..."
    
    extract_script "$plugin_file" "$script_file"
    
    # 创建测试参数
    local test_params='{"url": "https://example.com", "config": {"timeout": 10000}}'
    
    echo "执行插件..."
    echo "$test_params" | node "$script_file" 2>/dev/null || {
        echo "插件执行失败，查看脚本内容:"
        echo "--- 脚本开始 ---"
        cat "$script_file"
        echo "--- 脚本结束 ---"
    }
    
    rm -rf "$temp_dir"
    echo
}

# 测试Shell插件
test_shell_plugin() {
    local plugin_file="$1"
    local temp_dir=$(mktemp -d)
    local script_file="$temp_dir/plugin.sh"
    
    echo "测试Shell插件..."
    
    extract_script "$plugin_file" "$script_file"
    chmod +x "$script_file"
    
    # 创建测试参数
    local test_params='{"target": "127.0.0.1", "config": {"top_ports": 10}}'
    
    echo "执行插件..."
    echo "$test_params" | bash "$script_file" 2>/dev/null || {
        echo "插件执行失败，查看脚本内容:"
        echo "--- 脚本开始 ---"
        cat "$script_file"
        echo "--- 脚本结束 ---"
    }
    
    rm -rf "$temp_dir"
    echo
}

# 测试单个插件
test_plugin() {
    local plugin_file="$1"
    
    echo "=== 测试插件: $(basename "$plugin_file") ==="
    
    parse_yaml_plugin "$plugin_file"
    
    # 获取插件语言
    local language=$(grep -A 10 "^script:" "$plugin_file" | grep "language:" | sed 's/.*language: *//')
    
    case "$language" in
        "python")
            if command -v python3 &> /dev/null; then
                test_python_plugin "$plugin_file"
            else
                echo "跳过Python插件测试(Python3未安装)"
                echo
            fi
            ;;
        "javascript")
            if command -v node &> /dev/null; then
                test_javascript_plugin "$plugin_file"
            else
                echo "跳过JavaScript插件测试(Node.js未安装)"
                echo
            fi
            ;;
        "shell")
            test_shell_plugin "$plugin_file"
            ;;
        *)
            echo "不支持的插件语言: $language"
            echo
            ;;
    esac
}

# 主函数
main() {
    check_dependencies
    
    if [[ $# -eq 1 ]]; then
        # 测试指定插件
        test_plugin "$1"
    else
        # 测试所有示例插件
        echo "测试所有示例插件..."
        echo
        
        for plugin_file in "$PLUGINS_DIR"/*.yaml; do
            if [[ -f "$plugin_file" ]]; then
                test_plugin "$plugin_file"
            fi
        done
    fi
    
    echo "=== 测试完成 ==="
}

# 显示帮助信息
show_help() {
    echo "用法: $0 [插件文件路径]"
    echo
    echo "参数:"
    echo "  插件文件路径    要测试的YAML插件文件路径"
    echo "                 如果不指定，将测试所有示例插件"
    echo
    echo "示例:"
    echo "  $0                                    # 测试所有示例插件"
    echo "  $0 examples/subdomain_hunter.yaml    # 测试指定插件"
    echo
}

# 检查参数
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    show_help
    exit 0
fi

# 运行主函数
main "$@"