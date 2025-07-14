package portscan

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/time/rate"
)

// PortScanner 端口扫描器
type PortScanner struct {
	// 配置
	Config models.PortScanConfig
	// 结果通道
	ResultChan chan *models.PortScanResult
	// 任务ID
	TaskID string
	// 项目ID
	ProjectID string
	// 停止信号
	StopChan chan struct{}
	// 进度通道
	ProgressChan chan float64
	// 已扫描的主机端口集合
	scannedPorts sync.Map
	// 速率限制器
	limiter *rate.Limiter
	// 总扫描端口数
	totalPorts int
	// 已扫描端口数
	scannedCount int
	// 互斥锁
	mu sync.Mutex
}

// NewScanner 创建端口扫描器
func NewScanner(config models.PortScanConfig) *PortScanner {
	return NewPortScanner(config, "", "")
}

// NewPortScanner 创建端口扫描器
func NewPortScanner(config models.PortScanConfig, taskID, projectID string) *PortScanner {
	// 创建速率限制器
	var limiter *rate.Limiter
	if config.RateLimit > 0 {
		limiter = rate.NewLimiter(rate.Limit(config.RateLimit), config.RateLimit)
	}

	return &PortScanner{
		Config:       config,
		ResultChan:   make(chan *models.PortScanResult, 1000),
		TaskID:       taskID,
		ProjectID:    projectID,
		StopChan:     make(chan struct{}),
		ProgressChan: make(chan float64, 100),
		limiter:      limiter,
	}
}

// Start 开始端口扫描
func (s *PortScanner) Start(ctx context.Context, targets []string) error {
	// 解析端口范围
	ports, err := s.parsePorts(s.Config.Ports)
	if err != nil {
		return fmt.Errorf("解析端口范围失败: %v", err)
	}

	// 计算总扫描端口数
	s.totalPorts = len(targets) * len(ports)

	// 创建工作组
	var wg sync.WaitGroup

	// 启动结果处理协程
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.processResults(ctx)
	}()

	// 根据扫描类型执行不同的扫描策略
	switch s.Config.ScanType {
	case "tcp":
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.scanTCP(ctx, targets, ports)
		}()
	case "udp":
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.scanUDP(ctx, targets, ports)
		}()
	case "both":
		wg.Add(2)
		go func() {
			defer wg.Done()
			s.scanTCP(ctx, targets, ports)
		}()
		go func() {
			defer wg.Done()
			s.scanUDP(ctx, targets, ports)
		}()
	default:
		// 默认使用TCP扫描
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.scanTCP(ctx, targets, ports)
		}()
	}

	// 等待所有协程完成
	wg.Wait()
	close(s.ResultChan)
	close(s.ProgressChan)

	return nil
}

// Stop 停止端口扫描
func (s *PortScanner) Stop() {
	close(s.StopChan)
}

// scanTCP 执行TCP端口扫描
func (s *PortScanner) scanTCP(ctx context.Context, targets []string, ports []int) {
	// 创建工作池
	workerCount := s.Config.Concurrency
	if workerCount <= 0 {
		workerCount = 100 // 默认并发数
	}

	// 创建任务通道
	taskChan := make(chan scanTask, workerCount*2)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				select {
				case <-ctx.Done():
					return
				case <-s.StopChan:
					return
				default:
					// 如果有速率限制，则等待令牌
					if s.limiter != nil {
						s.limiter.Wait(ctx)
					}

					// 执行TCP扫描
					s.scanTCPPort(ctx, task.host, task.port)

					// 更新进度
					s.updateProgress()
				}
			}
		}()
	}

	// 发送扫描任务
	for _, target := range targets {
		// 检查目标是否在排除列表中
		if s.isExcluded(target) {
			continue
		}

		for _, port := range ports {
			select {
			case <-ctx.Done():
				close(taskChan)
				return
			case <-s.StopChan:
				close(taskChan)
				return
			default:
				taskChan <- scanTask{host: target, port: port}
			}
		}
	}

	// 关闭任务通道
	close(taskChan)

	// 等待所有工作协程完成
	wg.Wait()
}

// scanUDP 执行UDP端口扫描
func (s *PortScanner) scanUDP(ctx context.Context, targets []string, ports []int) {
	// 创建工作池
	workerCount := s.Config.Concurrency
	if workerCount <= 0 {
		workerCount = 50 // UDP扫描默认并发数较低
	}

	// 创建任务通道
	taskChan := make(chan scanTask, workerCount*2)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				select {
				case <-ctx.Done():
					return
				case <-s.StopChan:
					return
				default:
					// 如果有速率限制，则等待令牌
					if s.limiter != nil {
						s.limiter.Wait(ctx)
					}

					// 执行UDP扫描
					s.scanUDPPort(ctx, task.host, task.port)

					// 更新进度
					s.updateProgress()
				}
			}
		}()
	}

	// 发送扫描任务
	for _, target := range targets {
		// 检查目标是否在排除列表中
		if s.isExcluded(target) {
			continue
		}

		for _, port := range ports {
			select {
			case <-ctx.Done():
				close(taskChan)
				return
			case <-s.StopChan:
				close(taskChan)
				return
			default:
				taskChan <- scanTask{host: target, port: port}
			}
		}
	}

	// 关闭任务通道
	close(taskChan)

	// 等待所有工作协程完成
	wg.Wait()
}

// scanTCPPort 扫描单个TCP端口
func (s *PortScanner) scanTCPPort(ctx context.Context, host string, port int) {
	// 构建扫描结果的唯一键
	key := fmt.Sprintf("%s:%d:tcp", host, port)

	// 检查是否已扫描
	if _, exists := s.scannedPorts.Load(key); exists {
		return
	}
	s.scannedPorts.Store(key, true)

	// 设置超时
	timeout := time.Duration(s.Config.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	// 根据扫描方法选择不同的扫描策略
	var status string
	var err error

	switch s.Config.ScanMethod {
	case "connect":
		status, err = s.connectScan(host, port, timeout)
	case "syn":
		status, err = s.synScan(host, port, timeout)
	default:
		// 默认使用全连接扫描
		status, err = s.connectScan(host, port, timeout)
	}

	if err != nil {
		// 端口关闭或扫描出错
		return
	}

	// 创建扫描结果
	result := &models.PortScanResult{
		TaskID:    s.getObjectID(s.TaskID),
		ProjectID: s.getObjectID(s.ProjectID),
		Host:      host,
		Port:      port,
		Protocol:  "tcp",
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 如果需要服务识别
	if s.Config.ServiceDetection && status == "open" {
		s.detectService(ctx, result)
	}

	// 发送结果
	select {
	case s.ResultChan <- result:
	default:
		// 通道已满，丢弃结果
	}
}

// scanUDPPort 扫描单个UDP端口
func (s *PortScanner) scanUDPPort(ctx context.Context, host string, port int) {
	// 构建扫描结果的唯一键
	key := fmt.Sprintf("%s:%d:udp", host, port)

	// 检查是否已扫描
	if _, exists := s.scannedPorts.Load(key); exists {
		return
	}
	s.scannedPorts.Store(key, true)

	// 设置超时
	timeout := time.Duration(s.Config.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second // UDP扫描通常需要更长的超时时间
	}

	// UDP扫描实现
	status, err := s.udpScan(host, port, timeout)
	if err != nil {
		// 端口关闭或扫描出错
		return
	}

	// 创建扫描结果
	result := &models.PortScanResult{
		TaskID:    s.getObjectID(s.TaskID),
		ProjectID: s.getObjectID(s.ProjectID),
		Host:      host,
		Port:      port,
		Protocol:  "udp",
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 如果需要服务识别
	if s.Config.ServiceDetection && status == "open" {
		s.detectService(ctx, result)
	}

	// 发送结果
	select {
	case s.ResultChan <- result:
	default:
		// 通道已满，丢弃结果
	}
}

// connectScan 执行全连接扫描
func (s *PortScanner) connectScan(host string, port int, timeout time.Duration) (string, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return "closed", err
	}
	defer conn.Close()
	return "open", nil
}

// synScan 执行SYN扫描
// 注意：SYN扫描需要原始套接字权限，通常需要root权限
func (s *PortScanner) synScan(host string, port int, timeout time.Duration) (string, error) {
	// 这里是SYN扫描的简化实现
	// 实际上，SYN扫描需要使用原始套接字，这里简化为全连接扫描
	// 在生产环境中，可以考虑使用第三方库或调用系统命令实现
	return s.connectScan(host, port, timeout)
}

// udpScan 执行UDP扫描
func (s *PortScanner) udpScan(host string, port int, timeout time.Duration) (string, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("udp", address, timeout)
	if err != nil {
		return "closed", err
	}
	defer conn.Close()

	// 发送一个空的UDP数据包
	_, err = conn.Write([]byte{})
	if err != nil {
		return "closed", err
	}

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(timeout))

	// 尝试读取响应
	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	if err != nil {
		// 如果是超时错误，可能是端口开放但没有响应
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return "open|filtered", nil
		}
		// 如果是连接拒绝错误，端口可能是关闭的
		return "closed", err
	}

	// 收到响应，端口是开放的
	return "open", nil
}

// detectService 检测端口服务
func (s *PortScanner) detectService(ctx context.Context, result *models.PortScanResult) {
	// 这里实现服务识别逻辑
	// 可以使用常见端口映射、发送探测包、分析响应等方法
	// 简化实现，仅根据常见端口映射
	service, product := s.getCommonService(result.Port)
	result.Service = service
	result.Product = product

	// 如果是常见服务，尝试获取横幅信息
	if service != "unknown" {
		banner := s.getBanner(result.Host, result.Port, result.Protocol)
		result.Banner = banner

		// 根据横幅进一步识别服务和版本
		if banner != "" {
			s.identifyServiceFromBanner(result, banner)
		}
	}
}

// getCommonService 根据端口号获取常见服务
func (s *PortScanner) getCommonService(port int) (string, string) {
	commonPorts := map[int][]string{
		21:    {"ftp", "FTP Server"},
		22:    {"ssh", "SSH Server"},
		23:    {"telnet", "Telnet Server"},
		25:    {"smtp", "SMTP Server"},
		53:    {"dns", "DNS Server"},
		80:    {"http", "Web Server"},
		110:   {"pop3", "POP3 Server"},
		143:   {"imap", "IMAP Server"},
		443:   {"https", "Web Server"},
		445:   {"smb", "SMB Server"},
		3306:  {"mysql", "MySQL Database"},
		3389:  {"rdp", "Remote Desktop"},
		5432:  {"postgresql", "PostgreSQL Database"},
		6379:  {"redis", "Redis Database"},
		8080:  {"http-alt", "Web Server"},
		9200:  {"elasticsearch", "Elasticsearch"},
		27017: {"mongodb", "MongoDB Database"},
	}

	if service, ok := commonPorts[port]; ok {
		return service[0], service[1]
	}
	return "unknown", "Unknown Service"
}

// getBanner 获取服务横幅
func (s *PortScanner) getBanner(host string, port int, protocol string) string {
	address := fmt.Sprintf("%s:%d", host, port)
	timeout := time.Duration(s.Config.Timeout) * time.Second

	if protocol == "tcp" {
		conn, err := net.DialTimeout("tcp", address, timeout)
		if err != nil {
			return ""
		}
		defer conn.Close()

		// 设置读取超时
		conn.SetReadDeadline(time.Now().Add(timeout))

		// 对于某些服务，需要发送特定的探测包
		switch port {
		case 80, 8080, 443:
			// HTTP/HTTPS
			_, err = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
		case 25, 587:
			// SMTP
			_, err = conn.Write([]byte("EHLO example.com\r\n"))
		case 21:
			// FTP
			// FTP服务器通常会自动发送横幅
		case 22:
			// SSH
			// SSH服务器通常会自动发送横幅
		default:
			// 对于其他服务，不发送任何数据，只尝试读取
		}

		// 读取响应
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			return ""
		}

		return string(buf[:n])
	}

	return ""
}

// identifyServiceFromBanner 从横幅中识别服务和版本
func (s *PortScanner) identifyServiceFromBanner(result *models.PortScanResult, banner string) {
	// 这里可以实现更复杂的服务和版本识别逻辑
	// 简化实现，仅使用一些基本规则

	// HTTP服务器识别
	if result.Service == "http" || result.Service == "https" {
		if strings.Contains(banner, "Server: Apache") {
			result.Product = "Apache"
			// 尝试提取版本
			if idx := strings.Index(banner, "Server: Apache/"); idx != -1 {
				version := banner[idx+15:]
				if endIdx := strings.Index(version, "\r\n"); endIdx != -1 {
					result.Version = version[:endIdx]
				}
			}
		} else if strings.Contains(banner, "Server: nginx") {
			result.Product = "Nginx"
			// 尝试提取版本
			if idx := strings.Index(banner, "Server: nginx/"); idx != -1 {
				version := banner[idx+14:]
				if endIdx := strings.Index(version, "\r\n"); endIdx != -1 {
					result.Version = version[:endIdx]
				}
			}
		} else if strings.Contains(banner, "Server: Microsoft-IIS") {
			result.Product = "Microsoft IIS"
			// 尝试提取版本
			if idx := strings.Index(banner, "Server: Microsoft-IIS/"); idx != -1 {
				version := banner[idx+22:]
				if endIdx := strings.Index(version, "\r\n"); endIdx != -1 {
					result.Version = version[:endIdx]
				}
			}
		}
	}

	// SSH服务器识别
	if result.Service == "ssh" {
		if strings.Contains(banner, "OpenSSH") {
			result.Product = "OpenSSH"
			// 尝试提取版本
			if idx := strings.Index(banner, "OpenSSH_"); idx != -1 {
				version := banner[idx+8:]
				if endIdx := strings.Index(version, " "); endIdx != -1 {
					result.Version = version[:endIdx]
				}
			}
		}
	}

	// FTP服务器识别
	if result.Service == "ftp" {
		if strings.Contains(banner, "FileZilla") {
			result.Product = "FileZilla FTP Server"
			// 尝试提取版本
			if idx := strings.Index(banner, "FileZilla Server "); idx != -1 {
				version := banner[idx+17:]
				if endIdx := strings.Index(version, " "); endIdx != -1 {
					result.Version = version[:endIdx]
				}
			}
		} else if strings.Contains(banner, "vsFTPd") {
			result.Product = "vsFTPd"
			// 尝试提取版本
			if idx := strings.Index(banner, "vsFTPd "); idx != -1 {
				version := banner[idx+7:]
				if endIdx := strings.Index(version, " "); endIdx != -1 {
					result.Version = version[:endIdx]
				}
			}
		}
	}

	// 保存原始横幅
	result.Banner = banner
}

// processResults 处理结果
func (s *PortScanner) processResults(ctx context.Context) {
	// 在这里处理结果，例如去重、验证、保存到数据库等
	// 这个函数会在单独的协程中运行
	for result := range s.ResultChan {
		// 检查是否已经存在
		key := fmt.Sprintf("%s:%d:%s", result.Host, result.Port, result.Protocol)
		if _, exists := s.scannedPorts.Load(key); exists {
			continue
		}
		s.scannedPorts.Store(key, true)

		// 所有结果都会通过ResultChan发送到TaskManager的handleResults处理
		// TaskManager会负责调用ResultHandler.HandleResult来保存到数据库
		// 这里不需要直接操作数据库，保持职责分离
	}
}

// parsePorts 解析端口范围
func (s *PortScanner) parsePorts(portsStr string) ([]int, error) {
	var ports []int

	// 如果端口字符串为空，使用默认端口
	if portsStr == "" {
		portsStr = "1-1000"
	}

	// 分割端口范围
	ranges := strings.Split(portsStr, ",")
	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}

		// 检查是否是范围
		if strings.Contains(r, "-") {
			parts := strings.Split(r, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("无效的端口范围: %s", r)
			}

			start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, fmt.Errorf("无效的起始端口: %s", parts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, fmt.Errorf("无效的结束端口: %s", parts[1])
			}

			if start > end {
				return nil, fmt.Errorf("起始端口大于结束端口: %d > %d", start, end)
			}

			for i := start; i <= end; i++ {
				if i > 0 && i < 65536 {
					ports = append(ports, i)
				}
			}
		} else {
			// 单个端口
			port, err := strconv.Atoi(r)
			if err != nil {
				return nil, fmt.Errorf("无效的端口: %s", r)
			}

			if port > 0 && port < 65536 {
				ports = append(ports, port)
			}
		}
	}

	return ports, nil
}

// updateProgress 更新进度
func (s *PortScanner) updateProgress() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.scannedCount++
	if s.totalPorts > 0 {
		progress := float64(s.scannedCount) / float64(s.totalPorts) * 100
		select {
		case s.ProgressChan <- progress:
		default:
			// 通道已满，丢弃进度更新
		}
	}
}

// isExcluded 检查目标是否在排除列表中
func (s *PortScanner) isExcluded(target string) bool {
	for _, excluded := range s.Config.ExcludeHosts {
		if target == excluded {
			return true
		}
	}
	return false
}

// getObjectID 从字符串获取ObjectID
func (s *PortScanner) getObjectID(id string) primitive.ObjectID {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NewObjectID()
	}
	return objID
}

// scanTask 扫描任务
type scanTask struct {
	host string
	port int
}
