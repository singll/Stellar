package utils

import (
	"bufio"
	"strings"
)

// GetTargetList 从目标字符串中获取目标列表
func GetTargetList(target, ignore string) ([]string, error) {
	if target == "" {
		return []string{}, nil
	}

	// 解析目标
	var targetList []string
	scanner := bufio.NewScanner(strings.NewReader(target))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			targetList = append(targetList, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// 如果有忽略列表，过滤目标
	if ignore != "" {
		ignoreList := make(map[string]bool)
		scanner := bufio.NewScanner(strings.NewReader(ignore))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				ignoreList[line] = true
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		// 过滤目标
		var filteredList []string
		for _, target := range targetList {
			if !ignoreList[target] {
				filteredList = append(filteredList, target)
			}
		}
		targetList = filteredList
	}

	return targetList, nil
}

// HasPrefix 检查字符串是否以任何前缀开头
func HasPrefix(s string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

// RemovePrefix 移除字符串中的前缀
func RemovePrefix(s string, prefixes ...string) string {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return strings.TrimPrefix(s, prefix)
		}
	}
	return s
}

// GetBeforeLastDash 获取最后一个破折号前的内容
func GetBeforeLastDash(s string) string {
	index := strings.LastIndex(s, "-")
	if index != -1 {
		return s[:index]
	}
	return s
}
