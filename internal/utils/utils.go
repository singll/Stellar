package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

// WriteToFile 写入文件
func WriteToFile(filename string, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

// PrintProgressBar 打印进度条
func PrintProgressBar(current, total int, prefix string) {
	percent := float64(current) / float64(total) * 100
	filled := int(percent / 10)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", 10-filled)
	fmt.Printf("\r%s: [%s] %.1f%%", prefix, bar, percent)
	if current == total {
		fmt.Println()
	}
}

// GetNowTime 获取当前时间的格式化字符串
func GetNowTime() string {
	return time.Now().Format("2006-01-02-15-04-05")
}

// GetRootDomain 获取根域名
func GetRootDomain(domain string) string {
	// 检查是否是URL
	if strings.Contains(domain, "://") {
		parsedURL, err := url.Parse(domain)
		if err == nil {
			domain = parsedURL.Hostname()
		}
	}

	// 移除可能的端口号
	if strings.Contains(domain, ":") {
		parts := strings.Split(domain, ":")
		domain = parts[0]
	}

	// 使用publicsuffix库获取有效的顶级域名
	suffix, icann := publicsuffix.PublicSuffix(domain)
	if icann {
		parts := strings.Split(domain, ".")
		suffixParts := strings.Split(suffix, ".")
		if len(parts) > len(suffixParts) {
			rootDomain := parts[len(parts)-len(suffixParts)-1] + "." + suffix
			return rootDomain
		}
	}

	// 如果无法通过publicsuffix库处理，尝试简单的方法
	parts := strings.Split(domain, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "." + parts[len(parts)-1]
	}

	return domain
}

// IsDomainValid 检查域名是否有效
func IsDomainValid(domain string) bool {
	// 简单的域名验证
	return strings.Contains(domain, ".")
}

// IsIPAddress 检查是否是IP地址
func IsIPAddress(ip string) bool {
	// 简单的IP地址验证
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		for _, c := range part {
			if c < '0' || c > '9' {
				return false
			}
		}
		num := 0
		for _, c := range part {
			num = num*10 + int(c-'0')
		}
		if num > 255 {
			return false
		}
	}
	return true
}

// GenerateRandomPassword 生成随机密码
func GenerateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(1 * time.Nanosecond) // 确保随机性
	}
	return string(b)
}

// FormatFileSize 格式化文件大小
func FormatFileSize(size float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	unitIndex := 0
	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}
	return fmt.Sprintf("%.2f %s", size, units[unitIndex])
}

// StringInSlice 检查字符串是否在切片中
func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// RemoveDuplicates 移除切片中的重复元素
func RemoveDuplicates(strSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// SanitizeFilename 清理文件名
func SanitizeFilename(filename string) string {
	// 替换不安全的字符
	unsafe := []string{"/", "\\", "?", "%", "*", ":", "|", "\"", "<", ">", ".."}
	result := filename
	for _, char := range unsafe {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}

// MD5 计算字符串的MD5哈希值
func MD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// TrimSpace 去除字符串前后的空白字符
func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

// GetStringFromMap 从map中获取字符串值
func GetStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}
