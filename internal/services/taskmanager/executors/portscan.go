package executors

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"regexp"
	"crypto/tls"
	"golang.org/x/time/rate"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PortScanExecutor 端口扫描执行器
type PortScanExecutor struct {
	config      PortScanConfig
	mutex       sync.RWMutex
	serviceDB   *ServiceDatabase
	rateLimiter *rate.Limiter
	progress    *ScanProgress
}

// ServiceDatabase 服务指纹数据库
type ServiceDatabase struct {
	services map[int]ServiceInfo
	patterns map[string]ServicePattern
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Versions    []string `json:"versions"`
	Probes      []string `json:"probes"`
}

// ServicePattern 服务识别模式
type ServicePattern struct {
	Regex   *regexp.Regexp
	Service string
	Version string
}

// ScanProgress 扫描进度
type ScanProgress struct {
	Total     int64 `json:"total"`
	Current   int64 `json:"current"`
	Completed int64 `json:"completed"`
	Failed    int64 `json:"failed"`
	mutex     sync.RWMutex
}

// PortScanConfig 端口扫描配置
type PortScanConfig struct {
	MaxWorkers     int           `json:"max_workers"`
	Timeout        time.Duration `json:"timeout"`
	ConnectTimeout time.Duration `json:"connect_timeout"`
	EnableBanner   bool          `json:"enable_banner"`
	BannerTimeout  time.Duration `json:"banner_timeout"`
	MaxRetries     int           `json:"max_retries"`
	ScanMethod     string        `json:"scan_method"`     // tcp, udp, both
	EnableService  bool          `json:"enable_service"`  // 启用服务识别
	RateLimit      float64       `json:"rate_limit"`      // 每秒请求数限制
	EnableSSL      bool          `json:"enable_ssl"`      // 启用SSL探测
	UDPTimeout     time.Duration `json:"udp_timeout"`     // UDP超时时间
	MaxUDPRetries  int           `json:"max_udp_retries"` // UDP重试次数
}

// PortScanResult 端口扫描结果
type PortScanResult struct {
	Host          string            `json:"host"`
	Port          int               `json:"port"`
	Status        string            `json:"status"`
	Protocol      string            `json:"protocol"`
	Service       string            `json:"service,omitempty"`
	Banner        string            `json:"banner,omitempty"`
	Version       string            `json:"version,omitempty"`
	SSLInfo       *SSLInfo          `json:"ssl_info,omitempty"`
	ResponseTime  time.Duration     `json:"response_time"`
	Timestamp     time.Time         `json:"timestamp"`
	Error         string            `json:"error,omitempty"`
	Fingerprint   map[string]string `json:"fingerprint,omitempty"`
}

// SSLInfo SSL证书信息
type SSLInfo struct {
	Issuer      string    `json:"issuer"`
	Subject     string    `json:"subject"`
	NotBefore   time.Time `json:"not_before"`
	NotAfter    time.Time `json:"not_after"`
	Fingerprint string    `json:"fingerprint"`
	Version     string    `json:"version"`
	Cipher      string    `json:"cipher"`
}

// NewPortScanExecutor 创建端口扫描执行器
func NewPortScanExecutor(config PortScanConfig) *PortScanExecutor {
	// 设置默认值
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = 100
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}
	if config.ConnectTimeout <= 0 {
		config.ConnectTimeout = 3 * time.Second
	}
	if config.BannerTimeout <= 0 {
		config.BannerTimeout = 5 * time.Second
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 2
	}
	if config.ScanMethod == "" {
		config.ScanMethod = "tcp"
	}
	if config.RateLimit <= 0 {
		config.RateLimit = 100 // 每秒100个请求
	}
	if config.UDPTimeout <= 0 {
		config.UDPTimeout = 2 * time.Second
	}
	if config.MaxUDPRetries <= 0 {
		config.MaxUDPRetries = 3
	}

	// 创建速率限制器
	rateLimiter := rate.NewLimiter(rate.Limit(config.RateLimit), int(config.RateLimit))

	return &PortScanExecutor{
		config:      config,
		serviceDB:   NewServiceDatabase(),
		rateLimiter: rateLimiter,
		progress:    &ScanProgress{},
	}
}

// Execute 执行端口扫描任务
func (e *PortScanExecutor) Execute(ctx context.Context, task *models.Task) (*models.TaskResult, error) {
	// 解析任务配置
	target, ok := task.Config["target"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid target")
	}

	// 解析端口范围
	ports, err := e.parsePorts(task.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ports: %v", err)
	}

	// 解析扫描方法
	scanMethod := "tcp"
	if method, ok := task.Config["scan_method"].(string); ok {
		scanMethod = method
	}

	// 初始化进度
	e.progress.mutex.Lock()
	e.progress.Total = int64(len(ports))
	e.progress.Current = 0
	e.progress.Completed = 0
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

	// 执行端口扫描
	scanResults, err := e.scanPorts(ctx, target, ports, scanMethod, task)
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		return result, err
	}

	// 处理结果
	openPorts := make([]PortScanResult, 0)
	serviceStats := make(map[string]int)
	
	for _, scanResult := range scanResults {
		if scanResult.Status == "open" {
			openPorts = append(openPorts, scanResult)
			if scanResult.Service != "" {
				serviceStats[scanResult.Service]++
			}
		}
	}

	// 完成时间和统计
	result.EndTime = time.Now()
	result.Status = "completed"
	result.Data["open_ports"] = openPorts
	result.Data["scanned_ports"] = scanResults
	result.Data["open_count"] = len(openPorts)
	result.Data["total_scanned"] = len(scanResults)
	result.Data["target"] = target
	result.Data["scan_method"] = scanMethod
	result.Data["service_stats"] = serviceStats
	result.Data["scan_duration"] = result.EndTime.Sub(result.StartTime).Seconds()
	result.Summary = fmt.Sprintf("Found %d open ports on %s (scanned %d ports in %.2fs)", 
		len(openPorts), target, len(scanResults), result.EndTime.Sub(result.StartTime).Seconds())

	return result, nil
}

// GetSupportedTypes 获取支持的任务类型
func (e *PortScanExecutor) GetSupportedTypes() []string {
	return []string{"port_scan"}
}

// GetExecutorInfo 获取执行器信息
func (e *PortScanExecutor) GetExecutorInfo() models.ExecutorInfo {
	return models.ExecutorInfo{
		Name:        "PortScanExecutor",
		Version:     "2.0.0",
		Description: "Enhanced port scanning executor with TCP/UDP support, service identification, and SSL detection",
		Author:      "Stellar Team",
	}
}

// parsePorts 解析端口配置
func (e *PortScanExecutor) parsePorts(config map[string]interface{}) ([]int, error) {
	var ports []int

	// 从配置中获取端口信息
	if portsConfig, ok := config["ports"]; ok {
		switch v := portsConfig.(type) {
		case []interface{}:
			// 端口列表
			for _, p := range v {
				if port, ok := p.(float64); ok {
					ports = append(ports, int(port))
				}
			}
		case string:
			// 端口范围字符串，如 "1-1000" 或 "80,443,8080"
			if strings.Contains(v, "-") {
				// 解析范围
				parts := strings.Split(v, "-")
				if len(parts) == 2 {
					start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
					end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
					if err1 == nil && err2 == nil && start <= end {
						for i := start; i <= end; i++ {
							ports = append(ports, i)
						}
					}
				}
			} else if strings.Contains(v, ",") {
				// 解析列表
				parts := strings.Split(v, ",")
				for _, part := range parts {
					if port, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
						ports = append(ports, port)
					}
				}
			} else {
				// 单个端口
				if port, err := strconv.Atoi(v); err == nil {
					ports = append(ports, port)
				}
			}
		}
	}

	// 如果没有指定端口，使用默认端口
	if len(ports) == 0 {
		ports = e.getDefaultPorts()
	}

	return ports, nil
}

// getDefaultPorts 获取默认端口列表
func (e *PortScanExecutor) getDefaultPorts() []int {
	// 常用端口列表
	return []int{
		21, 22, 23, 25, 53, 80, 110, 111, 135, 139, 143, 443, 993, 995,
		1723, 3306, 3389, 5432, 5900, 6379, 8080, 8443, 9200, 27017,
		// Web服务端口
		80, 443, 8080, 8443, 8000, 8001, 8008, 8888, 9000, 9001, 9999,
		// 数据库端口
		1433, 1521, 3306, 5432, 6379, 9042, 27017, 28017,
		// 远程访问端口
		22, 23, 3389, 5900, 5901, 5902, 5903, 5904, 5905,
		// 邮件服务端口
		25, 110, 143, 465, 587, 993, 995,
		// FTP端口
		21, 22,
		// DNS端口
		53,
		// 其他常用端口
		135, 139, 445, 1723, 8009, 8010, 8011, 8012, 8013, 8014, 8015,
	}
}

// scanPorts 扫描端口
func (e *PortScanExecutor) scanPorts(ctx context.Context, target string, ports []int, scanMethod string, task *models.Task) ([]PortScanResult, error) {
	// 创建工作池
	jobs := make(chan ScanJob, len(ports))
	results := make(chan PortScanResult, len(ports))
	var wg sync.WaitGroup

	// 启动工作协程
	for i := 0; i < e.config.MaxWorkers; i++ {
		wg.Add(1)
		go e.worker(ctx, &wg, jobs, results)
	}

	// 发送扫描任务
	go func() {
		defer close(jobs)
		for _, port := range ports {
			job := ScanJob{
				Host:       target,
				Port:       port,
				ScanMethod: scanMethod,
			}
			select {
			case jobs <- job:
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
	var scanResults []PortScanResult
	for result := range results {
		scanResults = append(scanResults, result)
	}

	return scanResults, nil
}

// ScanJob 扫描任务
type ScanJob struct {
	Host       string
	Port       int
	ScanMethod string
}

// worker 工作协程
func (e *PortScanExecutor) worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan ScanJob, results chan<- PortScanResult) {
	defer wg.Done()

	for {
		select {
		case job, ok := <-jobs:
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

			// 执行扫描
			var result PortScanResult
			switch job.ScanMethod {
			case "tcp":
				result = e.scanTCPPort(ctx, job.Host, job.Port)
			case "udp":
				result = e.scanUDPPort(ctx, job.Host, job.Port)
			case "both":
				// 先扫描TCP再扫描UDP
				tcpResult := e.scanTCPPort(ctx, job.Host, job.Port)
				udpResult := e.scanUDPPort(ctx, job.Host, job.Port)
				// 优先返回开放的端口结果
				if tcpResult.Status == "open" {
					result = tcpResult
				} else if udpResult.Status == "open" {
					result = udpResult
				} else {
					result = tcpResult // 默认返回TCP结果
				}
			default:
				result = e.scanTCPPort(ctx, job.Host, job.Port)
			}

			// 更新进度
			e.progress.mutex.Lock()
			if result.Status == "open" {
				e.progress.Completed++
			} else if result.Error != "" {
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

// scanTCPPort 扫描TCP端口
func (e *PortScanExecutor) scanTCPPort(ctx context.Context, host string, port int) PortScanResult {
	startTime := time.Now()
	result := PortScanResult{
		Host:      host,
		Port:      port,
		Protocol:  "tcp",
		Status:    "closed",
		Timestamp: startTime,
	}

	// 尝试连接多次
	for retry := 0; retry < e.config.MaxRetries; retry++ {
		address := fmt.Sprintf("%s:%d", host, port)
		
		// 创建带超时的连接
		connectCtx, cancel := context.WithTimeout(ctx, e.config.ConnectTimeout)
		defer cancel()

		dialer := &net.Dialer{}
		conn, err := dialer.DialContext(connectCtx, "tcp", address)
		if err != nil {
			if retry == e.config.MaxRetries-1 {
				result.Error = err.Error()
			}
			continue
		}
		defer conn.Close()

		// 连接成功，端口开放
		result.Status = "open"
		result.ResponseTime = time.Since(startTime)
		result.Service = e.identifyService(port)

		// 如果启用了banner抓取
		if e.config.EnableBanner {
			banner, version := e.grabBanner(conn, port)
			if banner != "" {
				result.Banner = banner
			}
			if version != "" {
				result.Version = version
			}
		}

		// 如果启用了SSL检测且是常见SSL端口
		if e.config.EnableSSL && e.isSSLPort(port) {
			if sslInfo := e.grabSSLInfo(host, port); sslInfo != nil {
				result.SSLInfo = sslInfo
			}
		}

		// 进一步的服务识别
		if e.config.EnableService {
			fingerprint := e.generateFingerprint(result.Banner, result.Service, port)
			result.Fingerprint = fingerprint
		}

		break
	}

	result.ResponseTime = time.Since(startTime)
	return result
}

// scanUDPPort 扫描UDP端口
func (e *PortScanExecutor) scanUDPPort(ctx context.Context, host string, port int) PortScanResult {
	startTime := time.Now()
	result := PortScanResult{
		Host:      host,
		Port:      port,
		Protocol:  "udp",
		Status:    "filtered", // UDP默认为过滤状态
		Timestamp: startTime,
	}

	// 尝试连接多次
	for retry := 0; retry < e.config.MaxUDPRetries; retry++ {
		address := fmt.Sprintf("%s:%d", host, port)

		conn, err := net.DialTimeout("udp", address, e.config.UDPTimeout)
		if err != nil {
			if retry == e.config.MaxUDPRetries-1 {
				result.Error = err.Error()
			}
			continue
		}
		defer conn.Close()

		// 发送UDP探测数据
		probe := e.getUDPProbe(port)
		if _, err := conn.Write(probe); err != nil {
			continue
		}

		// 设置读取超时
		conn.SetReadDeadline(time.Now().Add(e.config.UDPTimeout))
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		
		if err == nil && n > 0 {
			// 有响应，端口开放
			result.Status = "open"
			result.Service = e.identifyService(port)
			response := string(buffer[:n])
			if len(response) > 0 {
				result.Banner = e.cleanBanner(response)
			}
			break
		} else if err != nil {
			// 查看是否是超时错误
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// 超时，端口可能开放但不响应
				result.Status = "open|filtered"
			} else {
				// 其他错误
				result.Status = "filtered"
			}
		}
	}

	result.ResponseTime = time.Since(startTime)
	return result
}

// identifyService 识别服务类型
func (e *PortScanExecutor) identifyService(port int) string {
	if service, exists := e.serviceDB.services[port]; exists {
		return service.Name
	}
	return "unknown"
}

// grabBanner 抓取banner信息
func (e *PortScanExecutor) grabBanner(conn net.Conn, port int) (string, string) {
	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(e.config.BannerTimeout))

	// 根据端口类型发送不同的探测数据
	var probe []byte
	var expectResponse bool
	
	switch port {
	case 80, 8080, 8000, 8001, 8008, 8888:
		probe = []byte("GET / HTTP/1.1\r\nHost: " + conn.RemoteAddr().String() + "\r\nUser-Agent: StellarScanner/2.0\r\n\r\n")
		expectResponse = true
	case 443, 8443:
		// HTTPS需要特殊处理，这里先跳过
		return "", ""
	case 21:
		// FTP通常会主动发送banner
		probe = []byte("")
		expectResponse = true
	case 22:
		// SSH通常会主动发送banner
		probe = []byte("")
		expectResponse = true
	case 25:
		// SMTP
		probe = []byte("HELO stellarscanner.local\r\n")
		expectResponse = true
	case 110:
		// POP3
		probe = []byte("USER test\r\n")
		expectResponse = true
	case 143:
		// IMAP
		probe = []byte("A001 CAPABILITY\r\n")
		expectResponse = true
	case 3306:
		// MySQL
		probe = []byte("")
		expectResponse = true
	case 5432:
		// PostgreSQL
		probe = []byte("")
		expectResponse = true
	case 6379:
		// Redis
		probe = []byte("PING\r\n")
		expectResponse = true
	case 9200:
		// Elasticsearch
		probe = []byte("GET / HTTP/1.1\r\nHost: " + conn.RemoteAddr().String() + "\r\n\r\n")
		expectResponse = true
	case 27017:
		// MongoDB
		probe = []byte("")
		expectResponse = true
	default:
		// 通用探测
		probe = []byte("\r\n")
		expectResponse = false
	}

	// 发送探测数据
	if len(probe) > 0 {
		if _, err := conn.Write(probe); err != nil {
			return "", ""
		}
	}

	// 读取响应
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		if !expectResponse {
			return "", ""
		}
		// 对于期望响应的协议，继续尝试读取
		time.Sleep(100 * time.Millisecond)
		n, err = conn.Read(buffer)
		if err != nil {
			return "", ""
		}
	}

	banner := string(buffer[:n])
	banner = e.cleanBanner(banner)

	// 提取版本信息
	version := e.extractVersion(banner, port)

	return banner, version
}

// cleanBanner 清理banner
func (e *PortScanExecutor) cleanBanner(banner string) string {
	// 移除控制字符
	banner = strings.ReplaceAll(banner, "\r", "")
	banner = strings.ReplaceAll(banner, "\n", " ")
	banner = strings.ReplaceAll(banner, "\t", " ")
	banner = strings.TrimSpace(banner)

	// 限制banner长度
	if len(banner) > 500 {
		banner = banner[:500] + "..."
	}

	return banner
}

// extractVersion 提取版本信息
func (e *PortScanExecutor) extractVersion(banner string, port int) string {
	// 基础版本提取规则
	versionPatterns := map[int][]*regexp.Regexp{
		22: {
			regexp.MustCompile(`SSH-([\d\.]+)`),
			regexp.MustCompile(`OpenSSH_([\d\.]+)`),
		},
		25: {
			regexp.MustCompile(`ESMTP ([\w\d\.\-]+)`),
			regexp.MustCompile(`Postfix ([\d\.]+)`),
		},
		80: {
			regexp.MustCompile(`Server: ([\w\d\.\-/\s]+)`),
			regexp.MustCompile(`Apache/([\d\.]+)`),
			regexp.MustCompile(`nginx/([\d\.]+)`),
		},
		3306: {
			regexp.MustCompile(`([\d\.]+)-MariaDB`),
			regexp.MustCompile(`([\d\.]+)-MySQL`),
		},
	}

	if patterns, exists := versionPatterns[port]; exists {
		for _, pattern := range patterns {
			if matches := pattern.FindStringSubmatch(banner); len(matches) > 1 {
				return matches[1]
			}
		}
	}

	// 通用版本提取
	generalPatterns := []*regexp.Regexp{
		regexp.MustCompile(`version ([\d\.]+)`),
		regexp.MustCompile(`v([\d\.]+)`),
		regexp.MustCompile(`([\d\.]+)`),
	}

	for _, pattern := range generalPatterns {
		if matches := pattern.FindStringSubmatch(banner); len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// isSSLPort 检查是否为SSL端口
func (e *PortScanExecutor) isSSLPort(port int) bool {
	sslPorts := []int{443, 8443, 993, 995, 465, 587, 636, 989, 990, 992, 993, 994, 995}
	for _, p := range sslPorts {
		if port == p {
			return true
		}
	}
	return false
}

// grabSSLInfo 获取SSL证书信息
func (e *PortScanExecutor) grabSSLInfo(host string, port int) *SSLInfo {
	address := fmt.Sprintf("%s:%d", host, port)
	
	// 配置 TLS
	config := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// 连接并获取证书
	conn, err := tls.DialWithDialer(&net.Dialer{
		Timeout: e.config.ConnectTimeout,
	}, "tcp", address, config)
	if err != nil {
		return nil
	}
	defer conn.Close()

	// 获取证书信息
	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return nil
	}

	cert := state.PeerCertificates[0]
	sslInfo := &SSLInfo{
		Issuer:      cert.Issuer.CommonName,
		Subject:     cert.Subject.CommonName,
		NotBefore:   cert.NotBefore,
		NotAfter:    cert.NotAfter,
		Fingerprint: fmt.Sprintf("%x", cert.Raw),
		Version:     getTLSVersion(state.Version),
		Cipher:      getCipherSuite(state.CipherSuite),
	}

	return sslInfo
}

// getTLSVersion 获取TLS版本
func getTLSVersion(version uint16) string {
	switch version {
	case tls.VersionSSL30:
		return "SSLv3"
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return "Unknown"
	}
}

// getCipherSuite 获取加密套件
func getCipherSuite(suite uint16) string {
	cipherSuites := map[uint16]string{
		tls.TLS_RSA_WITH_RC4_128_SHA:                "TLS_RSA_WITH_RC4_128_SHA",
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA:           "TLS_RSA_WITH_3DES_EDE_CBC_SHA",
		tls.TLS_RSA_WITH_AES_128_CBC_SHA:            "TLS_RSA_WITH_AES_128_CBC_SHA",
		tls.TLS_RSA_WITH_AES_256_CBC_SHA:            "TLS_RSA_WITH_AES_256_CBC_SHA",
		tls.TLS_RSA_WITH_AES_128_CBC_SHA256:         "TLS_RSA_WITH_AES_128_CBC_SHA256",
		tls.TLS_RSA_WITH_AES_128_GCM_SHA256:         "TLS_RSA_WITH_AES_128_GCM_SHA256",
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384:         "TLS_RSA_WITH_AES_256_GCM_SHA384",
		tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA:        "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:    "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:    "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
		tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA:          "TLS_ECDHE_RSA_WITH_RC4_128_SHA",
		tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:     "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA",
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:      "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:      "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256: "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256",
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256:   "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256",
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:   "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256: "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:   "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384: "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305:    "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305:  "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
	}

	if name, exists := cipherSuites[suite]; exists {
		return name
	}
	return fmt.Sprintf("Unknown(0x%04x)", suite)
}

// getUDPProbe 获取UDP探测数据
func (e *PortScanExecutor) getUDPProbe(port int) []byte {
	switch port {
	case 53:
		// DNS查询
		return []byte{0x12, 0x34, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x00, 0x01, 0x00, 0x01}
	case 123:
		// NTP查询
		return []byte{0x1b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	case 161:
		// SNMP查询
		return []byte{0x30, 0x26, 0x02, 0x01, 0x01, 0x04, 0x06, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0xa0, 0x19, 0x02, 0x04, 0x00, 0x00, 0x00, 0x00, 0x02, 0x01, 0x00, 0x02, 0x01, 0x00, 0x30, 0x0b, 0x30, 0x09, 0x06, 0x05, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x05, 0x00}
	case 500:
		// IKE查询
		return []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x10, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	case 1900:
		// SSDP查询
		return []byte("M-SEARCH * HTTP/1.1\r\nHost: 239.255.255.250:1900\r\nMan: \"ssdp:discover\"\r\nST: upnp:rootdevice\r\nMX: 3\r\n\r\n")
	default:
		// 通用UDP探测
		return []byte("\x00\x00\x00\x00")
	}
}

// generateFingerprint 生成服务指纹
func (e *PortScanExecutor) generateFingerprint(banner, service string, port int) map[string]string {
	fingerprint := make(map[string]string)
	
	fingerprint["port"] = fmt.Sprintf("%d", port)
	fingerprint["service"] = service
	
	if banner != "" {
		fingerprint["banner_length"] = fmt.Sprintf("%d", len(banner))
		fingerprint["banner_hash"] = fmt.Sprintf("%x", banner)[:16]
		
		// 提取关键字
		bannerLower := strings.ToLower(banner)
		if strings.Contains(bannerLower, "apache") {
			fingerprint["webserver"] = "apache"
		} else if strings.Contains(bannerLower, "nginx") {
			fingerprint["webserver"] = "nginx"
		} else if strings.Contains(bannerLower, "iis") {
			fingerprint["webserver"] = "iis"
		}
		
		if strings.Contains(bannerLower, "mysql") {
			fingerprint["database"] = "mysql"
		} else if strings.Contains(bannerLower, "postgresql") {
			fingerprint["database"] = "postgresql"
		} else if strings.Contains(bannerLower, "redis") {
			fingerprint["database"] = "redis"
		}
		
		if strings.Contains(bannerLower, "ssh") {
			fingerprint["protocol"] = "ssh"
		} else if strings.Contains(bannerLower, "ftp") {
			fingerprint["protocol"] = "ftp"
		} else if strings.Contains(bannerLower, "smtp") {
			fingerprint["protocol"] = "smtp"
		}
	}
	
	return fingerprint
}

// NewServiceDatabase 创建服务数据库
func NewServiceDatabase() *ServiceDatabase {
	db := &ServiceDatabase{
		services: make(map[int]ServiceInfo),
		patterns: make(map[string]ServicePattern),
	}
	
	// 初始化服务数据
	db.initializeServices()
	
	return db
}

// initializeServices 初始化服务数据
func (db *ServiceDatabase) initializeServices() {
	// 常见服务端口映射
	services := map[int]ServiceInfo{
		21:    {Name: "ftp", Description: "File Transfer Protocol", Probes: []string{"USER anonymous\r\n"}},
		22:    {Name: "ssh", Description: "Secure Shell", Probes: []string{""}},
		23:    {Name: "telnet", Description: "Telnet", Probes: []string{"\r\n"}},
		25:    {Name: "smtp", Description: "Simple Mail Transfer Protocol", Probes: []string{"HELO test\r\n"}},
		53:    {Name: "dns", Description: "Domain Name System", Probes: []string{""}},
		80:    {Name: "http", Description: "Hypertext Transfer Protocol", Probes: []string{"GET / HTTP/1.1\r\n\r\n"}},
		110:   {Name: "pop3", Description: "Post Office Protocol v3", Probes: []string{"USER test\r\n"}},
		143:   {Name: "imap", Description: "Internet Message Access Protocol", Probes: []string{"A001 CAPABILITY\r\n"}},
		443:   {Name: "https", Description: "HTTP over TLS/SSL", Probes: []string{""}},
		993:   {Name: "imaps", Description: "IMAP over TLS/SSL", Probes: []string{""}},
		995:   {Name: "pop3s", Description: "POP3 over TLS/SSL", Probes: []string{""}},
		1433:  {Name: "mssql", Description: "Microsoft SQL Server", Probes: []string{""}},
		1521:  {Name: "oracle", Description: "Oracle Database", Probes: []string{""}},
		3306:  {Name: "mysql", Description: "MySQL Database", Probes: []string{""}},
		3389:  {Name: "rdp", Description: "Remote Desktop Protocol", Probes: []string{""}},
		5432:  {Name: "postgresql", Description: "PostgreSQL Database", Probes: []string{""}},
		5900:  {Name: "vnc", Description: "Virtual Network Computing", Probes: []string{""}},
		6379:  {Name: "redis", Description: "Redis Database", Probes: []string{"PING\r\n"}},
		8080:  {Name: "http-proxy", Description: "HTTP Proxy", Probes: []string{"GET / HTTP/1.1\r\n\r\n"}},
		8443:  {Name: "https-alt", Description: "HTTPS Alternative", Probes: []string{""}},
		9200:  {Name: "elasticsearch", Description: "Elasticsearch", Probes: []string{"GET / HTTP/1.1\r\n\r\n"}},
		27017: {Name: "mongodb", Description: "MongoDB Database", Probes: []string{""}},
	}
	
	for port, service := range services {
		db.services[port] = service
	}
}