package pagemonitoring

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strings"
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
		Size:        int64(len(body)),
		LoadTime:    int64(loadTime),
	}

	return snapshot, nil
}

// preprocessHTML 预处理HTML内容以提高比较准确性
func preprocessHTML(html string, config models.MonitoringConfig) string {
	// 1. 移除空白字符和换行符的影响
	content := regexp.MustCompile(`\s+`).ReplaceAllString(html, " ")
	content = strings.TrimSpace(content)
	
	// 2. 如果配置忽略数字，替换所有数字
	if config.IgnoreNumbers {
		content = regexp.MustCompile(`\d+`).ReplaceAllString(content, "NUM")
	}
	
	// 3. 应用忽略模式
	for _, pattern := range config.IgnorePatterns {
		if re, err := regexp.Compile(pattern); err == nil {
			content = re.ReplaceAllString(content, "")
		}
	}
	
	// 4. 移除动态内容（时间戳、随机ID等）
	content = removeDynamicContent(content)
	
	// 5. 标准化HTML属性顺序
	content = normalizeHTMLAttributes(content)
	
	return content
}

// removeDynamicContent 移除动态内容
func removeDynamicContent(html string) string {
	// 移除时间戳（各种格式）
	patterns := []string{
		`\d{4}-\d{2}-\d{2}[\sT]\d{2}:\d{2}:\d{2}`, // 2024-07-11 10:30:00 或 2024-07-11T10:30:00
		`\d{2}/\d{2}/\d{4}\s+\d{2}:\d{2}:\d{2}`,   // 07/11/2024 10:30:00
		`timestamp="[^"]*"`,                        // timestamp="xxx"
		`time="[^"]*"`,                            // time="xxx"
		`data-time="[^"]*"`,                       // data-time="xxx"
		`_time\d+`,                                // _time1234567890
		`cache-bust=\d+`,                          // cache-bust=123456
		`v=\d+`,                                   // v=123456 (version parameters)
		`nonce="[^"]*"`,                           // nonce="random_string"
		`csrf-token="[^"]*"`,                      // csrf-token="random_string"
	}
	
	content := html
	for _, pattern := range patterns {
		if re, err := regexp.Compile(pattern); err == nil {
			content = re.ReplaceAllString(content, "DYNAMIC_CONTENT")
		}
	}
	
	return content
}

// normalizeHTMLAttributes 标准化HTML属性顺序
func normalizeHTMLAttributes(html string) string {
	// 简化实现：移除属性值中的额外空格
	re := regexp.MustCompile(`(\w+)="\s*([^"]*?)\s*"`)
	return re.ReplaceAllString(html, `$1="$2"`)
}

// extractText 提取文本内容 - 优化版本
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

// CompareSnapshots 比较快照 - 优化版本
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
		// 对于哈希比较，直接比较哈希值
		if oldSnapshot.ContentHash == newSnapshot.ContentHash {
			change.Status = models.PageChangeStatusUnchanged
			return change, 1.0, ""
		} else {
			change.Status = models.PageChangeStatusChanged
			change.Similarity = 0.0
			change.Diff = "内容哈希值不同，页面已发生变化"
			return change, 0.0, change.Diff
		}
	default:
		// HTML比较 - 使用智能预处理
		oldContent = preprocessHTML(oldSnapshot.HTML, config)
		newContent = preprocessHTML(newSnapshot.HTML, config)
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

// calculateSimilarity 计算相似度 - 优化版本
func calculateSimilarity(oldContent, newContent string) float64 {
	// 如果内容完全相同，则相似度为1
	if oldContent == newContent {
		return 1.0
	}

	// 如果内容为空，则相似度为0
	if oldContent == "" || newContent == "" {
		return 0.0
	}

	// 使用混合算法计算相似度
	// 1. 计算编辑距离相似度（权重45%）
	editDistSim := calculateEditDistanceSimilarity(oldContent, newContent)
	
	// 2. 计算余弦相似度（权重25%）
	cosineSim := calculateCosineSimilarity(oldContent, newContent)
	
	// 3. 计算Jaccard相似度（权重20%）
	jaccardSim := calculateJaccardSimilarity(oldContent, newContent)
	
	// 4. 计算最长公共子序列相似度（权重10%）
	lcsSim := calculateLCSSimilarity(oldContent, newContent)
	
	// 加权平均，增加编辑距离权重以提高敏感度
	finalSimilarity := 0.45*editDistSim + 0.25*cosineSim + 0.20*jaccardSim + 0.10*lcsSim
	
	return math.Min(1.0, math.Max(0.0, finalSimilarity))
}

// calculateEditDistanceSimilarity 计算编辑距离相似度
func calculateEditDistanceSimilarity(s1, s2 string) float64 {
	// 使用优化的编辑距离算法（Levenshtein距离）
	dist := levenshteinDistance(s1, s2)
	maxLen := max(len(s1), len(s2))
	
	if maxLen == 0 {
		return 1.0
	}
	
	return 1.0 - float64(dist)/float64(maxLen)
}

// levenshteinDistance 计算Levenshtein编辑距离
func levenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	m, n := len(r1), len(r2)
	
	// 优化：如果一个字符串为空，距离就是另一个的长度
	if m == 0 {
		return n
	}
	if n == 0 {
		return m
	}
	
	// 使用滚动数组优化空间复杂度
	prev := make([]int, n+1)
	curr := make([]int, n+1)
	
	// 初始化第一行
	for j := 0; j <= n; j++ {
		prev[j] = j
	}
	
	for i := 1; i <= m; i++ {
		curr[0] = i
		for j := 1; j <= n; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}
			
			curr[j] = min(min(curr[j-1]+1, prev[j]+1), prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	
	return prev[n]
}

// calculateCosineSimilarity 计算余弦相似度
func calculateCosineSimilarity(s1, s2 string) float64 {
	// 将文本转换为词向量
	vec1 := stringToVector(s1)
	vec2 := stringToVector(s2)
	
	// 计算余弦相似度
	return cosineSimilarity(vec1, vec2)
}

// stringToVector 将字符串转换为词频向量
func stringToVector(s string) map[string]int {
	vector := make(map[string]int)
	
	// 分词：按空格和标点符号分割
	words := regexp.MustCompile(`[\s\p{P}]+`).Split(s, -1)
	
	for _, word := range words {
		word = strings.ToLower(strings.TrimSpace(word))
		if word != "" && len(word) > 1 { // 忽略单字符词
			vector[word]++
		}
	}
	
	return vector
}

// cosineSimilarity 计算两个向量的余弦相似度
func cosineSimilarity(vec1, vec2 map[string]int) float64 {
	if len(vec1) == 0 || len(vec2) == 0 {
		return 0.0
	}
	
	// 计算点积
	dotProduct := 0.0
	for word, count1 := range vec1 {
		if count2, exists := vec2[word]; exists {
			dotProduct += float64(count1 * count2)
		}
	}
	
	// 计算向量模长
	norm1 := 0.0
	for _, count := range vec1 {
		norm1 += float64(count * count)
	}
	norm1 = math.Sqrt(norm1)
	
	norm2 := 0.0
	for _, count := range vec2 {
		norm2 += float64(count * count)
	}
	norm2 = math.Sqrt(norm2)
	
	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}
	
	return dotProduct / (norm1 * norm2)
}

// calculateJaccardSimilarity 计算Jaccard相似度
func calculateJaccardSimilarity(s1, s2 string) float64 {
	// 将字符串转换为字符n-gram集合
	set1 := stringToNGramSet(s1, 3) // 使用3-gram
	set2 := stringToNGramSet(s2, 3)
	
	// 计算交集大小
	intersection := 0
	for ngram := range set1 {
		if set2[ngram] {
			intersection++
		}
	}
	
	// 计算并集大小
	union := len(set1) + len(set2) - intersection
	
	if union == 0 {
		return 1.0
	}
	
	return float64(intersection) / float64(union)
}

// stringToNGramSet 将字符串转换为n-gram集合
func stringToNGramSet(s string, n int) map[string]bool {
	set := make(map[string]bool)
	runes := []rune(s)
	
	for i := 0; i <= len(runes)-n; i++ {
		ngram := string(runes[i : i+n])
		set[ngram] = true
	}
	
	return set
}

// calculateLCSSimilarity 计算最长公共子序列相似度
func calculateLCSSimilarity(s1, s2 string) float64 {
	lcs := longestCommonSubsequence(s1, s2)
	maxLen := max(len(s1), len(s2))
	
	if maxLen == 0 {
		return 1.0
	}
	
	return float64(lcs) / float64(maxLen)
}

// longestCommonSubsequence 计算最长公共子序列长度（优化版本）
func longestCommonSubsequence(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	m, n := len(r1), len(r2)
	
	// 优化：确保第一个字符串较短，减少空间复杂度
	if m > n {
		r1, r2 = r2, r1
		m, n = n, m
	}
	
	// 使用滚动数组优化空间复杂度
	prev := make([]int, m+1)
	curr := make([]int, m+1)
	
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if r2[i-1] == r1[j-1] {
				curr[j] = prev[j-1] + 1
			} else {
				curr[j] = max(prev[j], curr[j-1])
			}
		}
		prev, curr = curr, prev
	}
	
	return prev[m]
}

// calculateDiff 计算差异 - 优化版本
func calculateDiff(oldContent, newContent string) string {
	// 如果内容较短，直接生成diff
	if len(oldContent) < 10000 && len(newContent) < 10000 {
		return generateUnifiedDiff(oldContent, newContent)
	}
	
	// 对于大内容，计算摘要性差异
	return generateSummaryDiff(oldContent, newContent)
}

// generateUnifiedDiff 生成统一格式的diff
func generateUnifiedDiff(oldContent, newContent string) string {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")
	
	// 简化的diff算法：找出不同的行
	var diff strings.Builder
	oldLen, newLen := len(oldLines), len(newLines)
	maxLen := max(oldLen, newLen)
	
	changeCount := 0
	for i := 0; i < maxLen && changeCount < 20; i++ { // 限制显示的变更数量
		oldLine := ""
		newLine := ""
		
		if i < oldLen {
			oldLine = oldLines[i]
		}
		if i < newLen {
			newLine = newLines[i]
		}
		
		if oldLine != newLine {
			changeCount++
			if oldLine != "" {
				diff.WriteString("- ")
				diff.WriteString(truncateString(oldLine, 200))
				diff.WriteString("\n")
			}
			if newLine != "" {
				diff.WriteString("+ ")
				diff.WriteString(truncateString(newLine, 200))
				diff.WriteString("\n")
			}
		}
	}
	
	if changeCount >= 20 {
		diff.WriteString("... (显示了前20个变更)\n")
	}
	
	result := diff.String()
	if result == "" {
		return "内容结构发生微小变化"
	}
	
	return result
}

// generateSummaryDiff 生成摘要性差异
func generateSummaryDiff(oldContent, newContent string) string {
	var summary strings.Builder
	
	// 计算长度变化
	oldLen, newLen := len(oldContent), len(newContent)
	lengthChange := newLen - oldLen
	
	if lengthChange > 0 {
		summary.WriteString(fmt.Sprintf("内容增加了 %d 个字符\n", lengthChange))
	} else if lengthChange < 0 {
		summary.WriteString(fmt.Sprintf("内容减少了 %d 个字符\n", -lengthChange))
	}
	
	// 分析关键词变化
	oldWords := extractKeywords(oldContent)
	newWords := extractKeywords(newContent)
	
	addedWords := make(map[string]bool)
	removedWords := make(map[string]bool)
	
	// 找出新增的关键词
	for word := range newWords {
		if !oldWords[word] {
			addedWords[word] = true
		}
	}
	
	// 找出删除的关键词
	for word := range oldWords {
		if !newWords[word] {
			removedWords[word] = true
		}
	}
	
	if len(addedWords) > 0 {
		summary.WriteString("新增关键词: ")
		count := 0
		for word := range addedWords {
			if count < 10 { // 限制显示数量
				summary.WriteString(word + " ")
				count++
			}
		}
		summary.WriteString("\n")
	}
	
	if len(removedWords) > 0 {
		summary.WriteString("删除关键词: ")
		count := 0
		for word := range removedWords {
			if count < 10 { // 限制显示数量
				summary.WriteString(word + " ")
				count++
			}
		}
		summary.WriteString("\n")
	}
	
	result := summary.String()
	if result == "" {
		return "检测到内容变化，但无法生成详细差异信息"
	}
	
	return result
}

// extractKeywords 提取关键词
func extractKeywords(content string) map[string]bool {
	keywords := make(map[string]bool)
	
	// 提取长度大于3的单词作为关键词
	words := regexp.MustCompile(`[\p{L}\p{N}]{4,}`).FindAllString(content, -1)
	
	for _, word := range words {
		word = strings.ToLower(word)
		// 过滤常见词汇
		if !isCommonWord(word) {
			keywords[word] = true
		}
	}
	
	return keywords
}

// isCommonWord 判断是否为常见词汇
func isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "that": true, "have": true,
		"for": true, "not": true, "with": true, "you": true,
		"this": true, "but": true, "his": true, "from": true,
		"they": true, "she": true, "her": true, "been": true,
		"than": true, "its": true, "were": true, "said": true,
		"each": true, "which": true, "their": true, "time": true,
	}
	
	return commonWords[word]
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// max 返回两个整数中的最大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min 返回两个整数中的最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
