package executors

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SubdomainExecutor 子域名枚举执行器
type SubdomainExecutor struct {
	config      SubdomainConfig
	mutex       sync.RWMutex
	dnsCache    *DNSCache
	rateLimiter *rate.Limiter
	progress    *EnumProgress
	client      *http.Client
	resolvers   []*DNSResolver
}

// DNSResolver DNS解析器
type DNSResolver struct {
	server   string
	client   *net.Resolver
	timeout  time.Duration
	isDoH    bool
	httpClient *http.Client
}

// SubdomainConfig 子域名枚举配置
type SubdomainConfig struct {
	MaxWorkers         int           `json:"max_workers"`
	Timeout            time.Duration `json:"timeout"`
	WordlistPath       string        `json:"wordlist_path"`
	DNSServers         []string      `json:"dns_servers"`
	EnableWildcard     bool          `json:"enable_wildcard"`
	MaxRetries         int           `json:"max_retries"`
	EnumMethods        []string      `json:"enum_methods"`        // dns_brute, cert_transparency, search_engine
	RateLimit          float64       `json:"rate_limit"`          // 每秒请求数限制
	EnableDOH          bool          `json:"enable_doh"`          // DNS over HTTPS
	EnableRecursive    bool          `json:"enable_recursive"`    // 递归枚举
	MaxDepth           int           `json:"max_depth"`           // 最大递归深度
	VerifySubdomains   bool          `json:"verify_subdomains"`   // 验证子域名活跃性
	EnableCache        bool          `json:"enable_cache"`        // 启用DNS缓存
	CacheTimeout       time.Duration `json:"cache_timeout"`       // 缓存超时时间
	SearchEngineAPIs   map[string]string `json:"search_engine_apis"` // 搜索引擎API配置
}

// SubdomainResult 子域名枚举结果
type SubdomainResult struct {
	Subdomain    string            `json:"subdomain"`
	IPs          []string          `json:"ips"`
	CNAME        string            `json:"cname,omitempty"`
	Status       string            `json:"status"`
	Source       string            `json:"source"`
	ResponseTime time.Duration     `json:"response_time"`
	HTTPStatus   int               `json:"http_status,omitempty"`
	HTTPTitle    string            `json:"http_title,omitempty"`
	Technologies []string          `json:"technologies,omitempty"`
	Takeover     *TakeoverInfo     `json:"takeover,omitempty"`
	Timestamp    time.Time         `json:"timestamp"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// TakeoverInfo 子域名接管信息
type TakeoverInfo struct {
	Vulnerable bool   `json:"vulnerable"`
	Service    string `json:"service,omitempty"`
	Pattern    string `json:"pattern,omitempty"`
	CNAME      string `json:"cname,omitempty"`
}

// DNSCache DNS缓存
type DNSCache struct {
	cache  map[string]*CacheEntry
	mutex  sync.RWMutex
	ttl    time.Duration
}

// CacheEntry 缓存条目
type CacheEntry struct {
	result    SubdomainResult
	expiry    time.Time
	createdAt time.Time
}

// EnumProgress 枚举进度
type EnumProgress struct {
	Total     int64 `json:"total"`
	Current   int64 `json:"current"`
	Found     int64 `json:"found"`
	Failed    int64 `json:"failed"`
	mutex     sync.RWMutex
}

// NewSubdomainExecutor 创建子域名枚举执行器
func NewSubdomainExecutor(config SubdomainConfig) *SubdomainExecutor {
	// 设置默认值
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = 50
	}
	if config.Timeout <= 0 {
		config.Timeout = 5 * time.Second
	}
	if len(config.DNSServers) == 0 {
		config.DNSServers = []string{"8.8.8.8", "1.1.1.1", "114.114.114.114", "223.5.5.5"}
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	if len(config.EnumMethods) == 0 {
		config.EnumMethods = []string{"dns_brute"}
	}
	if config.RateLimit <= 0 {
		config.RateLimit = 10 // 每秒10个请求
	}
	if config.MaxDepth <= 0 {
		config.MaxDepth = 2
	}
	if config.CacheTimeout <= 0 {
		config.CacheTimeout = 5 * time.Minute
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext: (&net.Dialer{
				Timeout:   config.Timeout,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// 创建速率限制器
	rateLimiter := rate.NewLimiter(rate.Limit(config.RateLimit), int(config.RateLimit))

	// 初始化DNS缓存
	dnsCache := &DNSCache{
		cache: make(map[string]*CacheEntry),
		ttl:   config.CacheTimeout,
	}

	executor := &SubdomainExecutor{
		config:      config,
		dnsCache:    dnsCache,
		rateLimiter: rateLimiter,
		progress:    &EnumProgress{},
		client:      client,
	}

	// 初始化DNS解析器
	executor.initDNSResolvers()

	return executor
}

// Execute 执行子域名枚举任务
func (e *SubdomainExecutor) Execute(ctx context.Context, task *models.Task) (*models.TaskResult, error) {
	// 解析任务配置
	target, ok := task.Config["target"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid target domain")
	}

	// 验证域名格式
	if !e.isValidDomain(target) {
		return nil, fmt.Errorf("invalid domain format: %s", target)
	}

	// 初始化进度
	e.progress.mutex.Lock()
	e.progress.Total = 0
	e.progress.Current = 0
	e.progress.Found = 0
	e.progress.Failed = 0
	e.progress.mutex.Unlock()

	// 创建结果对象
	result := &models.TaskResult{
		ID:        primitive.NewObjectID(),
		TaskID:    task.ID,
		Status:    "running",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
		Data:      make(map[string]interface{}),
	}

	// 执行子域名枚举
	subdomains, err := e.enumerateSubdomains(ctx, target, task)
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		return result, err
	}

	// 去重和排序
	subdomains = e.deduplicateAndSort(subdomains)

	// 后处理：验证和接管检测
	if e.config.VerifySubdomains {
		subdomains = e.verifySubdomains(ctx, subdomains)
	}

	// 统计信息
	stats := e.generateStats(subdomains)

	// 完成时间和统计
	result.EndTime = time.Now()
	result.Status = "completed"
	result.Data["subdomains"] = subdomains
	result.Data["count"] = len(subdomains)
	result.Data["target"] = target
	result.Data["stats"] = stats
	result.Data["duration"] = result.EndTime.Sub(result.StartTime).Seconds()
	result.Summary = fmt.Sprintf("Found %d subdomains for %s using %v methods (%.2fs)", 
		len(subdomains), target, e.config.EnumMethods, result.EndTime.Sub(result.StartTime).Seconds())

	return result, nil
}

// GetSupportedTypes 获取支持的任务类型
func (e *SubdomainExecutor) GetSupportedTypes() []string {
	return []string{"subdomain_enum"}
}

// GetExecutorInfo 获取执行器信息
func (e *SubdomainExecutor) GetExecutorInfo() models.ExecutorInfo {
	return models.ExecutorInfo{
		Name:        "SubdomainExecutor",
		Version:     "2.0.0",
		Description: "Enhanced subdomain enumeration executor with multiple methods: DNS brute force, certificate transparency, search engines",
		Author:      "Stellar Team",
	}
}

// enumerateSubdomains 执行子域名枚举
func (e *SubdomainExecutor) enumerateSubdomains(ctx context.Context, target string, task *models.Task) ([]SubdomainResult, error) {
	var allSubdomains []SubdomainResult
	var mu sync.Mutex

	// 检查通配符解析
	wildcardIPs := make([]string, 0)
	if e.config.EnableWildcard {
		if hasWildcard, ips := e.checkWildcard(ctx, target); hasWildcard {
			wildcardIPs = ips
			// 不直接返回错误，而是继续执行但过滤通配符结果
		}
	}

	var wg sync.WaitGroup
	// 根据配置的枚举方法执行不同的枚举策略
	for _, method := range e.config.EnumMethods {
		wg.Add(1)
		go func(method string) {
			defer wg.Done()

			var results []SubdomainResult
			var err error

			switch method {
			case "dns_brute":
				results, err = e.dnsbruteForce(ctx, target, task)
			case "cert_transparency":
				results, err = e.certTransparency(ctx, target)
			case "search_engine":
				results, err = e.searchEngineQuery(ctx, target)
			case "dns_transfer":
				results, err = e.dnsZoneTransfer(ctx, target)
			default:
				fmt.Printf("Unknown enumeration method: %s\n", method)
				return
			}

			if err == nil && len(results) > 0 {
				// 过滤通配符结果
				if len(wildcardIPs) > 0 {
					results = e.filterWildcardResults(results, wildcardIPs)
				}

				mu.Lock()
				allSubdomains = append(allSubdomains, results...)
				mu.Unlock()
			}
		}(method)
	}

	wg.Wait()

	// 如果启用了递归枚举，对发现的子域名进行递归查找
	if e.config.EnableRecursive && len(allSubdomains) > 0 {
		recursiveResults := e.recursiveEnumeration(ctx, allSubdomains, target, 1)
		allSubdomains = append(allSubdomains, recursiveResults...)
	}

	return allSubdomains, nil
}

// getWordlist 获取字典
func (e *SubdomainExecutor) getWordlist(task *models.Task) ([]string, error) {
	// 从任务配置中获取自定义字典
	if customWordlist, ok := task.Config["wordlist"].([]interface{}); ok {
		wordlist := make([]string, len(customWordlist))
		for i, word := range customWordlist {
			if wordStr, ok := word.(string); ok {
				wordlist[i] = wordStr
			}
		}
		return wordlist, nil
	}

	// 使用默认字典
	return e.getDefaultWordlist(), nil
}

// getDefaultWordlist 获取默认字典
func (e *SubdomainExecutor) getDefaultWordlist() []string {
	// 返回一个更完整的子域名字典
	return []string{
		// 常见服务
		"www", "mail", "email", "ftp", "sftp", "ssh", "vpn", "proxy", "gateway",
		"admin", "administrator", "root", "api", "app", "application",
		
		// 开发和测试
		"dev", "devel", "development", "test", "testing", "staging", "stage",
		"beta", "alpha", "demo", "sandbox", "temp", "temporary", "lab",
		"experimental", "preview", "pre", "preprod", "production", "prod",
		
		// 内容管理
		"blog", "news", "forum", "wiki", "docs", "documentation", "help",
		"support", "faq", "kb", "knowledgebase", "tutorial", "guide",
		
		// 商务
		"shop", "store", "ecommerce", "cart", "checkout", "payment", "billing",
		"invoice", "order", "sales", "crm", "erp", "hr", "finance",
		
		// 技术基础设施
		"cdn", "cache", "redis", "memcache", "database", "db", "mysql", "postgres",
		"mongodb", "elasticsearch", "search", "solr", "ldap", "nfs", "smb",
		
		// 监控和管理
		"monitor", "monitoring", "metrics", "analytics", "stats", "statistics",
		"logs", "logging", "syslog", "kibana", "grafana", "prometheus",
		"nagios", "zabbix", "icinga", "status", "health", "ping",
		
		// 网络和安全
		"firewall", "ids", "ips", "proxy", "lb", "loadbalancer", "nginx",
		"apache", "iis", "tomcat", "jetty", "secure", "ssl", "tls",
		
		// 移动和媒体
		"m", "mobile", "wap", "mobi", "touch", "media", "images", "img",
		"static", "assets", "resources", "files", "download", "downloads",
		"upload", "uploads", "content", "video", "audio", "stream",
		
		// 版本控制和CI/CD
		"git", "svn", "cvs", "mercurial", "gitlab", "github", "bitbucket",
		"jenkins", "travis", "circleci", "bamboo", "ci", "cd", "build",
		"deploy", "deployment", "release", "artifact", "nexus", "artifactory",
		
		// CMS和框架
		"wp", "wordpress", "drupal", "joomla", "magento", "shopify", "prestashop",
		"opencart", "oscommerce", "zen", "cpanel", "plesk", "directadmin",
		"phpmyadmin", "adminer", "phpinfo", "info",
		
		// API版本
		"api-v1", "api-v2", "api-v3", "v1", "v2", "v3", "rest", "soap",
		"graphql", "webhook", "callbacks", "notifications",
		
		// 地理和语言
		"us", "eu", "asia", "china", "japan", "uk", "de", "fr", "es", "it",
		"en", "zh", "ja", "ko", "ru", "pt", "ar",
		
		// 环境和配置
		"local", "localhost", "internal", "private", "public", "external",
		"old", "new", "legacy", "archive", "backup", "bak", "mirror",
		"replica", "slave", "master", "primary", "secondary",
		
		// 特殊服务
		"webmail", "mail2", "smtp", "pop", "pop3", "imap", "exchange",
		"calendar", "contacts", "directory", "employees", "staff",
		"portal", "dashboard", "console", "panel", "control", "manage",
		
		// 开发工具
		"phpmyadmin", "mysql", "postgres", "oracle", "mssql", "redis",
		"rabbitmq", "kafka", "zookeeper", "consul", "vault", "nomad",
		
		// 通用前缀
		"sub", "subdomain", "host", "node", "server", "service", "app",
		"web", "site", "page", "home", "main", "index", "default",
		
		// 数字组合
		"1", "2", "3", "01", "02", "03", "001", "002", "003",
		"web1", "web2", "web3", "app1", "app2", "app3",
		"db1", "db2", "cache1", "cache2", "lb1", "lb2",
		
		// 云服务
		"aws", "azure", "gcp", "cloud", "k8s", "kubernetes", "docker",
		"swarm", "rancher", "openshift", "heroku", "digitalocean",
		
		// 其他常见
		"autoconfig", "autodiscover", "wpad", "broadcasthost", "isatap",
		"keyserver", "ocsp", "crl", "ca", "certificate", "cert",
	}
}

// isValidDomain 验证域名格式
func (e *SubdomainExecutor) isValidDomain(domain string) bool {
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}

	// 简单的域名格式验证
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return false
		}
	}

	return true
}

// initDNSResolvers 初始化DNS解析器
func (e *SubdomainExecutor) initDNSResolvers() {
	e.resolvers = make([]*DNSResolver, len(e.config.DNSServers))
	
	for i, server := range e.config.DNSServers {
		resolver := &DNSResolver{
			server:  server,
			timeout: e.config.Timeout,
		}
		
		// 检查是否为DoH服务器
		if e.config.EnableDOH && (strings.Contains(server, "dns.google") || strings.Contains(server, "cloudflare-dns.com")) {
			resolver.isDoH = true
			resolver.httpClient = e.client
		} else {
			// 传统DNS解析器
			resolver.client = &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{
						Timeout: e.config.Timeout,
					}
					return d.DialContext(ctx, network, server+":53")
				},
			}
		}
		
		e.resolvers[i] = resolver
	}
}

// dnsbruteForce DNS字典爆破
func (e *SubdomainExecutor) dnsbruteForce(ctx context.Context, target string, task *models.Task) ([]SubdomainResult, error) {
	// 获取字典
	wordlist, err := e.getWordlist(task)
	if err != nil {
		return nil, fmt.Errorf("failed to get wordlist: %v", err)
	}

	// 更新进度总数
	e.progress.mutex.Lock()
	e.progress.Total += int64(len(wordlist))
	e.progress.mutex.Unlock()

	// 创建工作池
	jobs := make(chan string, len(wordlist))
	results := make(chan SubdomainResult, len(wordlist))
	var wg sync.WaitGroup

	// 启动工作协程
	for i := 0; i < e.config.MaxWorkers; i++ {
		wg.Add(1)
		go e.dnsWorker(ctx, &wg, target, jobs, results)
	}

	// 发送任务
	go func() {
		defer close(jobs)
		for _, word := range wordlist {
			select {
			case jobs <- word:
			case <-ctx.Done():
				return
			}
		}
	}()

	// 等待完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	var subdomains []SubdomainResult
	for result := range results {
		if result.Status == "found" {
			subdomains = append(subdomains, result)
		}
	}

	return subdomains, nil
}

// dnsWorker DNS解析工作协程
func (e *SubdomainExecutor) dnsWorker(ctx context.Context, wg *sync.WaitGroup, target string, jobs <-chan string, results chan<- SubdomainResult) {
	defer wg.Done()

	for {
		select {
		case word, ok := <-jobs:
			if !ok {
				return
			}

			// 速率限制
			if err := e.rateLimiter.Wait(ctx); err != nil {
				return
			}

			// 更新进度
			e.progress.mutex.Lock()
			e.progress.Current++
			e.progress.mutex.Unlock()

			subdomain := fmt.Sprintf("%s.%s", word, target)
			result := e.resolveSubdomainEnhanced(ctx, subdomain)

			// 更新进度统计
			e.progress.mutex.Lock()
			if result.Status == "found" {
				e.progress.Found++
			} else {
				e.progress.Failed++
			}
			e.progress.mutex.Unlock()

			select {
			case results <- result:
			case <-ctx.Done():
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// resolveSubdomainEnhanced 增强的子域名解析
func (e *SubdomainExecutor) resolveSubdomainEnhanced(ctx context.Context, subdomain string) SubdomainResult {
	startTime := time.Now()
	result := SubdomainResult{
		Subdomain: subdomain,
		Status:    "not_found",
		Source:    "dns_brute",
		Timestamp: startTime,
		Metadata:  make(map[string]string),
	}

	// 检查缓存
	if e.config.EnableCache {
		if cached := e.getCachedResult(subdomain); cached != nil {
			return *cached
		}
	}

	// 尝试多个DNS服务器解析
	for _, resolver := range e.resolvers {
		if resolver.isDoH {
			// DoH解析
			if ips, cname := e.resolveWithDoH(ctx, subdomain, resolver); len(ips) > 0 || cname != "" {
				result.Status = "found"
				result.IPs = ips
				result.CNAME = cname
				break
			}
		} else {
			// 传统DNS解析
			if ips, cname := e.resolveWithDNS(ctx, subdomain, resolver); len(ips) > 0 || cname != "" {
				result.Status = "found"
				result.IPs = ips
				result.CNAME = cname
				break
			}
		}
	}

	result.ResponseTime = time.Since(startTime)

	// 缓存结果
	if e.config.EnableCache {
		e.setCachedResult(subdomain, result)
	}

	return result
}

// resolveWithDNS 使用传统DNS解析
func (e *SubdomainExecutor) resolveWithDNS(ctx context.Context, subdomain string, resolver *DNSResolver) ([]string, string) {
	var ips []string
	var cname string

	// 解析A记录
	if ipAddrs, err := resolver.client.LookupIPAddr(ctx, subdomain); err == nil && len(ipAddrs) > 0 {
		for _, ip := range ipAddrs {
			ips = append(ips, ip.IP.String())
		}
	}

	// 如果没有A记录，尝试解析CNAME记录
	if len(ips) == 0 {
		if cnameRecord, err := resolver.client.LookupCNAME(ctx, subdomain); err == nil && cnameRecord != subdomain+"." {
			cname = strings.TrimSuffix(cnameRecord, ".")
		}
	}

	return ips, cname
}

// resolveWithDoH 使用DNS over HTTPS解析
func (e *SubdomainExecutor) resolveWithDoH(ctx context.Context, subdomain string, resolver *DNSResolver) ([]string, string) {
	// DoH查询实现（Google DNS API格式）
	dohURL := fmt.Sprintf("https://%s/resolve?name=%s&type=A", resolver.server, subdomain)
	
	req, err := http.NewRequestWithContext(ctx, "GET", dohURL, nil)
	if err != nil {
		return nil, ""
	}
	
	req.Header.Set("Accept", "application/dns-json")
	req.Header.Set("User-Agent", "StellarServer/1.0")
	
	resp, err := resolver.httpClient.Do(req)
	if err != nil {
		return nil, ""
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, ""
	}
	
	// 解析DoH响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ""
	}
	
	var dnsResp struct {
		Status int `json:"Status"`
		Answer []struct {
			Name string `json:"name"`
			Type int    `json:"type"`
			Data string `json:"data"`
		} `json:"Answer"`
	}
	
	if err := json.Unmarshal(body, &dnsResp); err != nil {
		return nil, ""
	}
	
	if dnsResp.Status != 0 { // NOERROR
		return nil, ""
	}
	
	var ips []string
	var cname string
	
	for _, answer := range dnsResp.Answer {
		switch answer.Type {
		case 1: // A记录
			ips = append(ips, answer.Data)
		case 5: // CNAME记录
			cname = strings.TrimSuffix(answer.Data, ".")
		}
	}
	
	return ips, cname
}

// certTransparency 证书透明度日志查询
func (e *SubdomainExecutor) certTransparency(ctx context.Context, target string) ([]SubdomainResult, error) {
	var results []SubdomainResult
	
	// crt.sh查询
	crtResults, err := e.queryCrtSh(ctx, target)
	if err == nil {
		results = append(results, crtResults...)
	}
	
	// censys查询（如果有API密钥）
	if apiKey, exists := e.config.SearchEngineAPIs["censys"]; exists && apiKey != "" {
		censysResults, err := e.queryCensys(ctx, target)
		if err == nil {
			results = append(results, censysResults...)
		}
	}
	
	return results, nil
}

// queryCrtSh 查询crt.sh证书透明度日志
func (e *SubdomainExecutor) queryCrtSh(ctx context.Context, target string) ([]SubdomainResult, error) {
	url := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", target)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("crt.sh returned status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// 解析JSON响应
	var certEntries []struct {
		NameValue string `json:"name_value"`
	}
	
	if err := json.Unmarshal(body, &certEntries); err != nil {
		return nil, err
	}
	
	// 提取子域名
	subdomainSet := make(map[string]bool)
	for _, entry := range certEntries {
		names := strings.Split(entry.NameValue, "\n")
		for _, name := range names {
			name = strings.TrimSpace(name)
			if strings.HasSuffix(name, "."+target) && name != target {
				subdomainSet[name] = true
			}
		}
	}
	
	// 转换为结果
	var results []SubdomainResult
	for subdomain := range subdomainSet {
		result := SubdomainResult{
			Subdomain: subdomain,
			Status:    "found",
			Source:    "cert_transparency",
			Timestamp: time.Now(),
			Metadata:  map[string]string{"source": "crt.sh"},
		}
		results = append(results, result)
	}
	
	return results, nil
}

// queryCensys 查询Censys（需要API密钥）
func (e *SubdomainExecutor) queryCensys(ctx context.Context, target string) ([]SubdomainResult, error) {
	// Censys API查询实现
	// 需要API密钥和完整的API集成
	return nil, fmt.Errorf("censys API not implemented")
}

// searchEngineQuery 搜索引擎查询
func (e *SubdomainExecutor) searchEngineQuery(ctx context.Context, target string) ([]SubdomainResult, error) {
	var allResults []SubdomainResult
	
	// Google搜索
	googleResults, err := e.searchGoogle(ctx, target)
	if err == nil {
		allResults = append(allResults, googleResults...)
	}
	
	// Bing搜索
	bingResults, err := e.searchBing(ctx, target)
	if err == nil {
		allResults = append(allResults, bingResults...)
	}
	
	return allResults, nil
}

// searchGoogle Google搜索引擎查询
func (e *SubdomainExecutor) searchGoogle(ctx context.Context, target string) ([]SubdomainResult, error) {
	query := fmt.Sprintf("site:%s -www", target)
	url := fmt.Sprintf("https://www.google.com/search?q=%s&num=100", query)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("google search returned status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// 使用正则表达式提取子域名
	pattern := regexp.MustCompile(`https?://([a-zA-Z0-9.-]+\.` + regexp.QuoteMeta(target) + `)`)
	matches := pattern.FindAllStringSubmatch(string(body), -1)
	
	subdomainSet := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			subdomain := match[1]
			if subdomain != target {
				subdomainSet[subdomain] = true
			}
		}
	}
	
	var results []SubdomainResult
	for subdomain := range subdomainSet {
		result := SubdomainResult{
			Subdomain: subdomain,
			Status:    "found",
			Source:    "search_engine",
			Timestamp: time.Now(),
			Metadata:  map[string]string{"source": "google"},
		}
		results = append(results, result)
	}
	
	return results, nil
}

// searchBing Bing搜索引擎查询
func (e *SubdomainExecutor) searchBing(ctx context.Context, target string) ([]SubdomainResult, error) {
	query := fmt.Sprintf("site:%s -www", target)
	url := fmt.Sprintf("https://www.bing.com/search?q=%s&count=100", query)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bing search returned status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// 使用正则表达式提取子域名
	pattern := regexp.MustCompile(`https?://([a-zA-Z0-9.-]+\.` + regexp.QuoteMeta(target) + `)`)
	matches := pattern.FindAllStringSubmatch(string(body), -1)
	
	subdomainSet := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			subdomain := match[1]
			if subdomain != target {
				subdomainSet[subdomain] = true
			}
		}
	}
	
	var results []SubdomainResult
	for subdomain := range subdomainSet {
		result := SubdomainResult{
			Subdomain: subdomain,
			Status:    "found",
			Source:    "search_engine",
			Timestamp: time.Now(),
			Metadata:  map[string]string{"source": "bing"},
		}
		results = append(results, result)
	}
	
	return results, nil
}

// dnsZoneTransfer DNS区域传输
func (e *SubdomainExecutor) dnsZoneTransfer(ctx context.Context, target string) ([]SubdomainResult, error) {
	// DNS区域传输实现
	// 这是一个高级功能，需要目标服务器允许区域传输
	return nil, fmt.Errorf("DNS zone transfer not implemented")
}

// recursiveEnumeration 递归枚举
func (e *SubdomainExecutor) recursiveEnumeration(ctx context.Context, foundSubdomains []SubdomainResult, target string, currentDepth int) []SubdomainResult {
	if currentDepth >= e.config.MaxDepth {
		return nil
	}
	
	var newSubdomains []SubdomainResult
	
	// 对每个发现的子域名进行递归查找
	for _, subdomain := range foundSubdomains {
		// 使用常见的递归前缀
		recursivePrefixes := []string{"www", "m", "mobile", "api", "app", "dev", "test", "stage", "admin"}
		
		for _, prefix := range recursivePrefixes {
			newSubdomain := fmt.Sprintf("%s.%s", prefix, subdomain.Subdomain)
			result := e.resolveSubdomainEnhanced(ctx, newSubdomain)
			
			if result.Status == "found" {
				result.Source = "recursive"
				result.Metadata["depth"] = fmt.Sprintf("%d", currentDepth)
				newSubdomains = append(newSubdomains, result)
			}
		}
	}
	
	// 继续递归查找
	if len(newSubdomains) > 0 && currentDepth < e.config.MaxDepth {
		deeperResults := e.recursiveEnumeration(ctx, newSubdomains, target, currentDepth+1)
		newSubdomains = append(newSubdomains, deeperResults...)
	}
	
	return newSubdomains
}

// checkWildcard 增强的通配符检测
func (e *SubdomainExecutor) checkWildcard(ctx context.Context, domain string) (bool, []string) {
	var wildcardIPs []string
	
	// 生成多个随机子域名进行测试
	testSubdomains := []string{
		fmt.Sprintf("nonexistent-%d", rand.Intn(1000000)),
		fmt.Sprintf("random-%d", rand.Intn(1000000)),
		fmt.Sprintf("test-%d", rand.Intn(1000000)),
	}
	
	resolver := e.resolvers[0] // 使用第一个解析器
	
	for _, testSubdomain := range testSubdomains {
		fullDomain := fmt.Sprintf("%s.%s", testSubdomain, domain)
		
		if resolver.isDoH {
			continue // 简化版跳过DoH测试
		}
		
		ips, err := resolver.client.LookupIPAddr(ctx, fullDomain)
		if err == nil && len(ips) > 0 {
			// 发现通配符解析
			for _, ip := range ips {
				wildcardIPs = append(wildcardIPs, ip.IP.String())
			}
			return true, wildcardIPs
		}
	}
	
	return false, wildcardIPs
}

// filterWildcardResults 过滤通配符结果
func (e *SubdomainExecutor) filterWildcardResults(results []SubdomainResult, wildcardIPs []string) []SubdomainResult {
	var filtered []SubdomainResult
	
	for _, result := range results {
		isWildcard := false
		
		// 检查IP是否与通配符IP相同
		for _, resultIP := range result.IPs {
			for _, wildcardIP := range wildcardIPs {
				if resultIP == wildcardIP {
					isWildcard = true
					break
				}
			}
			if isWildcard {
				break
			}
		}
		
		if !isWildcard {
			filtered = append(filtered, result)
		}
	}
	
	return filtered
}

// deduplicateAndSort 去重和排序
func (e *SubdomainExecutor) deduplicateAndSort(subdomains []SubdomainResult) []SubdomainResult {
	seen := make(map[string]bool)
	var unique []SubdomainResult
	
	for _, subdomain := range subdomains {
		if !seen[subdomain.Subdomain] {
			seen[subdomain.Subdomain] = true
			unique = append(unique, subdomain)
		}
	}
	
	// 按子域名排序
	sort.Slice(unique, func(i, j int) bool {
		return unique[i].Subdomain < unique[j].Subdomain
	})
	
	return unique
}

// verifySubdomains 验证子域名活跃性
func (e *SubdomainExecutor) verifySubdomains(ctx context.Context, subdomains []SubdomainResult) []SubdomainResult {
	var verified []SubdomainResult
	
	for i, subdomain := range subdomains {
		// HTTP验证
		if httpStatus, title, technologies := e.verifyHTTP(ctx, subdomain.Subdomain); httpStatus > 0 {
			subdomains[i].HTTPStatus = httpStatus
			subdomains[i].HTTPTitle = title
			subdomains[i].Technologies = technologies
		}
		
		// 子域名接管检测
		if takeover := e.checkSubdomainTakeover(ctx, subdomain); takeover != nil {
			subdomains[i].Takeover = takeover
		}
		
		verified = append(verified, subdomains[i])
	}
	
	return verified
}

// verifyHTTP HTTP验证
func (e *SubdomainExecutor) verifyHTTP(ctx context.Context, subdomain string) (int, string, []string) {
	// 尝试HTTP和HTTPS
	schemes := []string{"https", "http"}
	
	for _, scheme := range schemes {
		url := fmt.Sprintf("%s://%s", scheme, subdomain)
		
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue
		}
		
		req.Header.Set("User-Agent", "StellarScanner/2.0")
		
		resp, err := e.client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		
		// 读取响应内容以提取标题
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, "", nil
		}
		
		// 提取标题
		titleRegex := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
		titleMatches := titleRegex.FindStringSubmatch(string(body))
		var title string
		if len(titleMatches) > 1 {
			title = strings.TrimSpace(titleMatches[1])
		}
		
		// 检测技术栈
		technologies := e.detectTechnologies(resp.Header, string(body))
		
		return resp.StatusCode, title, technologies
	}
	
	return 0, "", nil
}

// detectTechnologies 检测技术栈
func (e *SubdomainExecutor) detectTechnologies(headers http.Header, body string) []string {
	var technologies []string
	
	// 检测服务器软件
	if server := headers.Get("Server"); server != "" {
		technologies = append(technologies, "Server: "+server)
	}
	
	// 检测CMS和框架
	if strings.Contains(body, "wp-content") {
		technologies = append(technologies, "WordPress")
	}
	if strings.Contains(body, "Powered by Drupal") {
		technologies = append(technologies, "Drupal")
	}
	if strings.Contains(body, "Joomla") {
		technologies = append(technologies, "Joomla")
	}
	
	return technologies
}

// checkSubdomainTakeover 检查子域名接管
func (e *SubdomainExecutor) checkSubdomainTakeover(ctx context.Context, result SubdomainResult) *TakeoverInfo {
	if result.CNAME == "" {
		return nil
	}
	
	// 常见的可接管服务模式
	patterns := map[string]string{
		"github.io":           "There isn't a GitHub Pages site here",
		"herokuapp.com":       "no such app",
		"s3.amazonaws.com":    "NoSuchBucket",
		"s3-website":          "NoSuchBucket",
		"cloudapp.net":        "page you're looking for can't be found",
		"azurewebsites.net":   "Error 404",
	}
	
	for service, pattern := range patterns {
		if strings.Contains(result.CNAME, service) {
			// 验证是否真的可以接管
			if e.verifyTakeover(ctx, result.Subdomain, pattern) {
				return &TakeoverInfo{
					Vulnerable: true,
					Service:    service,
					Pattern:    pattern,
					CNAME:      result.CNAME,
				}
			}
		}
	}
	
	return nil
}

// verifyTakeover 验证接管漏洞
func (e *SubdomainExecutor) verifyTakeover(ctx context.Context, subdomain, pattern string) bool {
	url := fmt.Sprintf("http://%s", subdomain)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}
	
	resp, err := e.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	
	return strings.Contains(string(body), pattern)
}

// generateStats 生成统计信息
func (e *SubdomainExecutor) generateStats(subdomains []SubdomainResult) map[string]interface{} {
	stats := make(map[string]interface{})
	
	// 按来源统计
	sourceStats := make(map[string]int)
	httpStats := make(map[int]int)
	takeoverCount := 0
	
	for _, subdomain := range subdomains {
		sourceStats[subdomain.Source]++
		
		if subdomain.HTTPStatus > 0 {
			httpStats[subdomain.HTTPStatus]++
		}
		
		if subdomain.Takeover != nil && subdomain.Takeover.Vulnerable {
			takeoverCount++
		}
	}
	
	stats["by_source"] = sourceStats
	stats["http_status"] = httpStats
	stats["takeover_vulnerable"] = takeoverCount
	stats["total_found"] = len(subdomains)
	
	return stats
}

// DNS缓存相关方法
func (e *SubdomainExecutor) getCachedResult(subdomain string) *SubdomainResult {
	e.dnsCache.mutex.RLock()
	defer e.dnsCache.mutex.RUnlock()
	
	if entry, exists := e.dnsCache.cache[subdomain]; exists {
		if time.Now().Before(entry.expiry) {
			return &entry.result
		}
		// 清理过期缓存
		delete(e.dnsCache.cache, subdomain)
	}
	
	return nil
}

func (e *SubdomainExecutor) setCachedResult(subdomain string, result SubdomainResult) {
	e.dnsCache.mutex.Lock()
	defer e.dnsCache.mutex.Unlock()
	
	e.dnsCache.cache[subdomain] = &CacheEntry{
		result:    result,
		expiry:    time.Now().Add(e.dnsCache.ttl),
		createdAt: time.Now(),
	}
}