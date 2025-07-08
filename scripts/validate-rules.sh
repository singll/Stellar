#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 规则文件目录
RULES_DIR=".cursor/rules"
MAIN_RULES_FILE=".cursorrules"

# 检查规则文件格式
check_format() {
    echo -e "${YELLOW}检查规则文件格式...${NC}"
    local has_error=0

    # 检查主规则文件
    if [ ! -f "$MAIN_RULES_FILE" ]; then
        echo -e "${RED}错误: 主规则文件 $MAIN_RULES_FILE 不存在${NC}"
        has_error=1
    fi

    # 检查规则目录
    if [ ! -d "$RULES_DIR" ]; then
        echo -e "${RED}错误: 规则目录 $RULES_DIR 不存在${NC}"
        has_error=1
    fi

    # 检查每个规则文件
    for file in "$RULES_DIR"/*.mdc; do
        if [ -f "$file" ]; then
            # 检查文件头部格式
            if ! grep -q "^---$" "$file" || ! grep -q "^规则名称:" "$file" || ! grep -q "^版本:" "$file"; then
                echo -e "${RED}错误: $file 缺少必要的头部信息${NC}"
                has_error=1
            fi

            # 检查文件大小
            if [ $(wc -c < "$file") -lt 100 ]; then
                echo -e "${YELLOW}警告: $file 内容可能不完整${NC}"
            fi
        fi
    done

    if [ $has_error -eq 0 ]; then
        echo -e "${GREEN}规则文件格式检查通过${NC}"
    else
        echo -e "${RED}规则文件格式检查失败${NC}"
        return 1
    fi
}

# 检查规则依赖关系
check_deps() {
    echo -e "${YELLOW}检查规则依赖关系...${NC}"
    local has_error=0

    for file in "$RULES_DIR"/*.mdc; do
        if [ -f "$file" ]; then
            # 获取依赖规则列表
            deps=$(grep "^  - " "$file" | cut -d' ' -f4)
            
            # 检查每个依赖是否存在
            for dep in $deps; do
                if [ ! -f "$RULES_DIR/$dep" ]; then
                    echo -e "${RED}错误: $file 依赖的规则文件 $dep 不存在${NC}"
                    has_error=1
                fi
            done
        fi
    done

    if [ $has_error -eq 0 ]; then
        echo -e "${GREEN}规则依赖关系检查通过${NC}"
    else
        echo -e "${RED}规则依赖关系检查失败${NC}"
        return 1
    fi
}

# 检查规则版本一致性
check_version() {
    echo -e "${YELLOW}检查规则版本一致性...${NC}"
    local has_error=0
    local version_pattern="版本: ([0-9]+\.[0-9]+\.[0-9]+)"

    for file in "$RULES_DIR"/*.mdc; do
        if [ -f "$file" ]; then
            if ! grep -q "^版本: [0-9]\+\.[0-9]\+\.[0-9]\+$" "$file"; then
                echo -e "${RED}错误: $file 版本号格式不正确${NC}"
                has_error=1
            fi
        fi
    done

    if [ $has_error -eq 0 ]; then
        echo -e "${GREEN}规则版本一致性检查通过${NC}"
    else
        echo -e "${RED}规则版本一致性检查失败${NC}"
        return 1
    fi
}

# 生成规则应用报告
generate_report() {
    echo -e "${YELLOW}生成规则应用报告...${NC}"
    local report_file="rules-report.md"

    # 创建报告文件
    cat > "$report_file" << EOF
# Stellar 项目规则应用报告

生成时间: $(date)

## 规则文件统计

- 主规则文件: $(if [ -f "$MAIN_RULES_FILE" ]; then echo "✅"; else echo "❌"; fi)
- 规则文件数量: $(ls "$RULES_DIR"/*.mdc 2>/dev/null | wc -l)

## 规则优先级分布

$(for file in "$RULES_DIR"/*.mdc; do
    if [ -f "$file" ]; then
        priority=$(grep "优先级:" "$file" | cut -d' ' -f2)
        echo "- $(basename "$file"): $priority"
    fi
done)

## 规则依赖关系

$(for file in "$RULES_DIR"/*.mdc; do
    if [ -f "$file" ]; then
        echo "### $(basename "$file")"
        grep "^  - " "$file" | sed 's/^  - /- /'
        echo
    fi
done)

## 最后更新时间

$(for file in "$RULES_DIR"/*.mdc; do
    if [ -f "$file" ]; then
        update_time=$(grep "最后更新:" "$file" | cut -d' ' -f2)
        echo "- $(basename "$file"): $update_time"
    fi
done)
EOF

    echo -e "${GREEN}规则应用报告已生成: $report_file${NC}"
}

# 主函数
main() {
    case "$1" in
        "format")
            check_format
            ;;
        "deps")
            check_deps
            ;;
        "version")
            check_version
            ;;
        "report")
            generate_report
            ;;
        *)
            echo "用法: $0 {format|deps|version|report}"
            echo "  format  - 检查规则文件格式"
            echo "  deps    - 检查规则依赖关系"
            echo "  version - 检查规则版本一致性"
            echo "  report  - 生成规则应用报告"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@" 