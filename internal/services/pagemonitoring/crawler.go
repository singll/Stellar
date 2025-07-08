package pagemonitoring

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PageCrawler 页面抓取器
type PageCrawler struct {
	client *http.Client
}

// NewPageCrawler 创建页面抓取器
func NewPageCrawler(timeout int) *PageCrawler {
	if timeout <= 0 {
		timeout = 30
	}
	return &PageCrawler{
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// FetchPage 抓取页面
func (c *PageCrawler) FetchPage(url string, config models.MonitoringConfig) (*models.PageSnapshot, error) {
	startTime := time.Now()

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置默认请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	// 设置自定义请求头
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// 设置认证
	if config.Authentication.Type == "basic" {
		req.SetBasicAuth(config.Authentication.Username, config.Authentication.Password)
	} else if config.Authentication.Type == "cookie" && config.Authentication.Cookie != "" {
		req.Header.Set("Cookie", config.Authentication.Cookie)
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 计算加载时间
	loadTime := int(time.Since(startTime).Milliseconds())

	// 提取响应头
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// 解析HTML
	htmlContent := string(body)
	textContent, err := extractText(htmlContent, config)
	if err != nil {
		textContent = ""
	}

	// 计算哈希值
	contentHash := calculateHash(htmlContent)

	// 创建快照
	snapshot := &models.PageSnapshot{
		ID:          primitive.NewObjectID(),
		URL:         url,
		StatusCode:  resp.StatusCode,
		Headers:     headers,
		HTML:        htmlContent,
		Text:        textContent,
		ContentHash: contentHash,
		CreatedAt:   time.Now(),
		Size:        len(body),
		LoadTime:    loadTime,
	}

	return snapshot, nil
}

// extractText 提取文本内容
func extractText(html string, config models.MonitoringConfig) (string, error) {
	// 简单实现：移除HTML标签
	re := regexp.MustCompile("<[^>]*>")
	text := re.ReplaceAllString(html, "")

	// 如果需要忽略数字，则替换所有数字
	if config.IgnoreNumbers {
		re := regexp.MustCompile(`\d+`)
		text = re.ReplaceAllString(text, "X")
	}

	// 应用忽略模式
	for _, pattern := range config.IgnorePatterns {
		re, err := regexp.Compile(pattern)
		if err == nil {
			text = re.ReplaceAllString(text, "")
		}
	}

	return text, nil
}

// calculateHash 计算哈希值
func calculateHash(content string) string {
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}

// CompareSnapshots 比较快照
func CompareSnapshots(oldSnapshot, newSnapshot *models.PageSnapshot, config models.MonitoringConfig) (*models.PageChange, float64, string) {
	// 创建变更记录
	change := &models.PageChange{
		ID:            primitive.NewObjectID(),
		URL:           newSnapshot.URL,
		OldSnapshotID: oldSnapshot.ID,
		NewSnapshotID: newSnapshot.ID,
		ChangedAt:     time.Now(),
	}

	// 根据比较方法选择比较内容
	var oldContent, newContent string
	switch config.CompareMethod {
	case "text":
		oldContent = oldSnapshot.Text
		newContent = newSnapshot.Text
		change.DiffType = "text"
	case "hash":
		oldContent = oldSnapshot.ContentHash
		newContent = newSnapshot.ContentHash
		change.DiffType = "hash"
	default:
		oldContent = oldSnapshot.HTML
		newContent = newSnapshot.HTML
		change.DiffType = "html"
	}

	// 计算相似度
	similarity := calculateSimilarity(oldContent, newContent)
	change.Similarity = similarity

	// 判断是否有变更
	if similarity >= config.SimilarityThreshold {
		change.Status = models.PageChangeStatusUnchanged
		return change, similarity, ""
	}

	// 计算差异
	diff := calculateDiff(oldContent, newContent)
	change.Diff = diff
	change.Status = models.PageChangeStatusChanged

	return change, similarity, diff
}

// calculateSimilarity 计算相似度
func calculateSimilarity(oldContent, newContent string) float64 {
	// 如果内容完全相同，则相似度为1
	if oldContent == newContent {
		return 1.0
	}

	// 如果内容为空，则相似度为0
	if oldContent == "" || newContent == "" {
		return 0.0
	}

	// 简单实现：计算最长公共子序列长度
	lcs := longestCommonSubsequence(oldContent, newContent)
	maxLen := max(len(oldContent), len(newContent))

	if maxLen == 0 {
		return 1.0
	}

	return float64(lcs) / float64(maxLen)
}

// longestCommonSubsequence 计算最长公共子序列长度
func longestCommonSubsequence(s1, s2 string) int {
	m, n := len(s1), len(s2)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if s1[i-1] == s2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	return dp[m][n]
}

// calculateDiff 计算差异
func calculateDiff(oldContent, newContent string) string {
	// 简单实现：返回新内容
	return "变更内容过大，请查看完整内容"
}

// max 返回两个整数中的最大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
