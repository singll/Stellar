package assetdiscovery

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/StellarServer/internal/models"
)

// WebScanner Web扫描器
type WebScanner struct {
	concurrency    int
	timeout        time.Duration
	retryCount     int
	followRedirect bool
	userAgent      string
}

// NewWebScanner 创建Web扫描器
func NewWebScanner(concurrency int, timeout time.Duration, retryCount int, followRedirect bool) *WebScanner {
	if concurrency <= 0 {
		concurrency = 20
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if retryCount <= 0 {
		retryCount = 2
	}

	return &WebScanner{
		concurrency:    concurrency,
		timeout:        timeout,
		retryCount:     retryCount,
		followRedirect: followRedirect,
		userAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	}
}

// 实现Web资产发现功能
// 注意：此方法在discovery.go中已声明，这里只提供实现细节
// 不要在此文件中重复声明runWebDiscovery方法

// scanWebTarget 扫描Web目标
func scanWebTarget(url string, timeout time.Duration, retryCount int, followRedirect bool, customHeaders map[string]string) *models.DiscoveryResult {
	// 确保URL以http://或https://开头
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	// 创建结果
	result := &models.DiscoveryResult{
		Target:    url,
		AssetType: "webapp",
		IsAlive:   false,
		WebApps:   []models.WebAppInfo{},
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !followRedirect {
				return http.ErrUseLastResponse
			}
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// 设置自定义请求头
	for key, value := range customHeaders {
		req.Header.Set(key, value)
	}

	// 发送请求
	var resp *http.Response
	var lastErr error

	// 重试机制
	for i := 0; i <= retryCount; i++ {
		resp, lastErr = client.Do(req)
		if lastErr == nil {
			break
		}
		time.Sleep(time.Duration(i*500) * time.Millisecond)
	}

	if lastErr != nil || resp == nil {
		return nil
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // 限制读取1MB
	if err != nil {
		return nil
	}

	// 解析网页内容
	webApp := parseWebResponse(url, resp, body)
	result.WebApps = append(result.WebApps, webApp)
	result.IsAlive = true

	// 提取IP地址
	host := extractHost(url)
	ips, err := net.LookupIP(host)
	if err == nil && len(ips) > 0 {
		for _, ip := range ips {
			if ipv4 := ip.To4(); ipv4 != nil {
				result.IP = ipv4.String()
				break
			}
		}
	}

	return result
}

// parseWebResponse 解析Web响应
func parseWebResponse(url string, resp *http.Response, body []byte) models.WebAppInfo {
	webApp := models.WebAppInfo{
		URL:          url,
		StatusCode:   resp.StatusCode,
		Headers:      make(map[string]string),
		Technologies: []string{},
	}

	// 提取标题
	titleRegex := regexp.MustCompile(`<title[^>]*>(.*?)</title>`)
	titleMatches := titleRegex.FindSubmatch(body)
	if len(titleMatches) > 1 {
		webApp.Title = string(titleMatches[1])
	}

	// 提取服务器信息
	if server := resp.Header.Get("Server"); server != "" {
		webApp.Server = server
		webApp.Technologies = append(webApp.Technologies, "Server: "+server)
	}

	// 提取响应头
	for key, values := range resp.Header {
		if len(values) > 0 {
			webApp.Headers[key] = values[0]
		}
	}

	// 提取技术栈
	detectTechnologies(&webApp, body, resp.Header)

	// 提取Cookies
	for _, cookie := range resp.Cookies() {
		webApp.Cookies = append(webApp.Cookies, cookie.Name+"="+cookie.Value)
	}

	return webApp
}

// detectTechnologies 检测Web技术栈
func detectTechnologies(webApp *models.WebAppInfo, body []byte, headers http.Header) {
	bodyStr := string(body)

	// 检测Web框架
	frameworks := map[string]string{
		"jQuery":    `jquery[.-]([0-9.]+)`,
		"Bootstrap": `bootstrap[.-]([0-9.]+)`,
		"React":     `react[.-]([0-9.]+)`,
		"Angular":   `angular[.-]([0-9.]+)`,
		"Vue.js":    `vue[.-]([0-9.]+)`,
		"Laravel":   `laravel`,
		"Django":    `django`,
		"Flask":     `flask`,
		"Express":   `express`,
		"Spring":    `spring`,
		"WordPress": `wp-content|wordpress`,
		"Joomla":    `joomla`,
		"Drupal":    `drupal`,
		"Magento":   `magento`,
		"Shopify":   `shopify`,
	}

	// 检测Web服务器
	servers := map[string]string{
		"Apache":     `apache`,
		"Nginx":      `nginx`,
		"IIS":        `iis|microsoft-iis`,
		"Tomcat":     `tomcat`,
		"Cloudflare": `cloudflare`,
		"Cloudfront": `cloudfront`,
		"Fastly":     `fastly`,
		"Akamai":     `akamai`,
	}

	// 检测编程语言
	languages := map[string]string{
		"PHP":     `php`,
		"ASP.NET": `asp\.net`,
		"Ruby":    `ruby|rails`,
		"Python":  `python|django|flask`,
		"Java":    `java|jsp|jsessionid`,
		"Node.js": `node|express`,
	}

	// 检测数据库
	databases := map[string]string{
		"MySQL":      `mysql`,
		"PostgreSQL": `postgresql|postgres`,
		"MongoDB":    `mongodb`,
		"SQL Server": `sqlserver|mssql`,
		"Oracle":     `oracle`,
		"Redis":      `redis`,
	}

	// 检测框架
	for name, pattern := range frameworks {
		if regexp.MustCompile(`(?i)` + pattern).MatchString(bodyStr) {
			webApp.Technologies = append(webApp.Technologies, name)
		}
	}

	// 检测服务器
	for name, pattern := range servers {
		if regexp.MustCompile(`(?i)`+pattern).MatchString(bodyStr) ||
			regexp.MustCompile(`(?i)`+pattern).MatchString(webApp.Server) {
			webApp.Technologies = append(webApp.Technologies, name)
		}
	}

	// 检测编程语言
	for name, pattern := range languages {
		if regexp.MustCompile(`(?i)` + pattern).MatchString(bodyStr) {
			webApp.Technologies = append(webApp.Technologies, name)
		}
	}

	// 检测数据库
	for name, pattern := range databases {
		if regexp.MustCompile(`(?i)` + pattern).MatchString(bodyStr) {
			webApp.Technologies = append(webApp.Technologies, name)
		}
	}

	// 检测特定HTTP头
	if _, ok := headers["X-Powered-By"]; ok {
		webApp.Technologies = append(webApp.Technologies, "X-Powered-By: "+headers.Get("X-Powered-By"))
	}
}

// extractHost 从URL中提取主机名
func extractHost(urlStr string) string {
	// 移除协议部分
	hostPart := urlStr
	if strings.Contains(urlStr, "://") {
		parts := strings.Split(urlStr, "://")
		if len(parts) > 1 {
			hostPart = parts[1]
		}
	}

	// 移除路径部分
	if strings.Contains(hostPart, "/") {
		hostPart = strings.Split(hostPart, "/")[0]
	}

	// 移除端口部分
	if strings.Contains(hostPart, ":") {
		hostPart = strings.Split(hostPart, ":")[0]
	}

	return hostPart
}
