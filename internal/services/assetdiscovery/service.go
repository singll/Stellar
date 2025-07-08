package assetdiscovery

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ServiceScanner 服务扫描器
type ServiceScanner struct {
	concurrency   int
	timeout       time.Duration
	retryCount    int
	resultHandler ResultHandler
}

// NewServiceScanner 创建服务扫描器
func NewServiceScanner(concurrency int, timeout time.Duration, retryCount int, resultHandler ResultHandler) *ServiceScanner {
	if concurrency <= 0 {
		concurrency = 50
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	if retryCount <= 0 {
		retryCount = 2
	}

	return &ServiceScanner{
		concurrency:   concurrency,
		timeout:       timeout,
		retryCount:    retryCount,
		resultHandler: resultHandler,
	}
}

// 服务发现功能实现
func (s *ServiceScanner) ScanServices(task *DiscoveryTask) error {
	// 创建服务扫描器
	scanner := NewServiceScanner(
		task.Task.Config.Concurrency,
		time.Duration(task.Task.Config.Timeout)*time.Second,
		task.Task.Config.RetryCount,
		s.resultHandler,
	)

	// 解析目标
	var targets []string
	for _, target := range task.Task.Targets {
		// 检查是否已经是IP:端口格式
		if strings.Contains(target, ":") {
			targets = append(targets, target)
			continue
		}

		// 尝试解析为IP或域名
		ips, err := parseTarget(target)
		if err != nil {
			continue
		}

		// 为每个IP添加常用端口
		for _, ip := range ips {
			// 如果配置了端口范围，使用配置的端口
			if len(task.Task.Config.PortRanges) > 0 {
				for _, portRange := range task.Task.Config.PortRanges {
					ports, err := expandPortRange(portRange)
					if err != nil {
						continue
					}
					for _, port := range ports {
						targets = append(targets, fmt.Sprintf("%s:%d", ip, port))
					}
				}
			} else {
				// 否则使用默认常用端口
				commonPorts := []int{21, 22, 23, 25, 53, 80, 110, 111, 135, 139, 143, 443, 445, 993, 995, 1723, 3306, 3389, 5900, 8080, 8443}
				for _, port := range commonPorts {
					targets = append(targets, fmt.Sprintf("%s:%d", ip, port))
				}
			}
		}
	}

	// 更新任务总目标数
	totalTargets := len(targets)
	if totalTargets == 0 {
		return fmt.Errorf("未找到有效的服务目标")
	}

	// 更新任务信息
	task.Task.ResultSummary.TotalTargets = totalTargets

	// 创建工作通道
	targetChan := make(chan string, totalTargets)
	resultChan := make(chan *models.DiscoveryResult, totalTargets)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < scanner.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range targetChan {
				// 检查任务是否被取消
				select {
				case <-task.Context.Done():
					return
				default:
				}

				// 扫描服务
				result := scanService(target, scanner.timeout, scanner.retryCount)
				if result != nil {
					resultChan <- result
				}
			}
		}()
	}

	// 发送任务到通道
	go func() {
		for _, target := range targets {
			targetChan <- target
		}
		close(targetChan)
	}()

	// 处理结果
	go func() {
		scannedCount := 0
		discoveredCount := 0
		serviceTypes := make(map[string]int)

		for result := range resultChan {
			scannedCount++

			// 如果发现服务，则计数
			if len(result.Services) > 0 {
				discoveredCount++

				// 统计服务类型
				for _, service := range result.Services {
					serviceTypes[service.Name]++
				}

				// 保存结果
				result.TaskID = task.Task.ID
				result.ID = primitive.NewObjectID()
				result.FirstSeen = time.Now()
				result.LastSeen = time.Now()

				// 处理结果
				s.resultHandler.HandleDiscoveryResult(task, result)

				// 添加到任务结果
				task.Mutex.Lock()
				task.Results = append(task.Results, result)
				task.Mutex.Unlock()
			}

			// 更新进度
			progress := float64(scannedCount) / float64(totalTargets) * 100
			task.Mutex.Lock()
			task.Progress = progress
			task.Mutex.Unlock()

			// 更新数据库中的任务进度
			s.resultHandler.UpdateTaskStatus(task.ID, "running", progress)

			// 更新任务摘要
			task.Task.ResultSummary.ScannedTargets = scannedCount
			task.Task.ResultSummary.ServiceCount = discoveredCount
			task.Task.ResultSummary.ServiceTypes = serviceTypes
		}
	}()

	// 等待所有工作协程完成
	wg.Wait()
	close(resultChan)

	// 给结果处理协程一些时间完成
	time.Sleep(1 * time.Second)

	return nil
}

// expandPortRange 展开端口范围
func expandPortRange(portRange string) ([]int, error) {
	var ports []int

	// 处理单个端口
	if !strings.Contains(portRange, "-") {
		port, err := strconv.Atoi(portRange)
		if err != nil {
			return nil, err
		}
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("无效的端口: %d", port)
		}
		return []int{port}, nil
	}

	// 处理端口范围
	parts := strings.Split(portRange, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("无效的端口范围: %s", portRange)
	}

	startPort, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}

	endPort, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}

	if startPort < 1 || startPort > 65535 || endPort < 1 || endPort > 65535 || startPort > endPort {
		return nil, fmt.Errorf("无效的端口范围: %s", portRange)
	}

	// 如果范围太大，可能导致内存问题
	if endPort-startPort > 1000 {
		return nil, fmt.Errorf("端口范围太大: %s", portRange)
	}

	for port := startPort; port <= endPort; port++ {
		ports = append(ports, port)
	}

	return ports, nil
}

// scanService 扫描单个服务
func scanService(target string, timeout time.Duration, retryCount int) *models.DiscoveryResult {
	// 解析目标
	host, portStr, err := net.SplitHostPort(target)
	if err != nil {
		return nil
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil
	}

	// 创建结果
	result := &models.DiscoveryResult{
		Target:    target,
		IP:        host,
		AssetType: "service",
		IsAlive:   true,
		Services:  []models.ServiceInfo{},
		Ports:     []models.PortInfo{},
	}

	// 检查端口是否开放
	portInfo := models.PortInfo{
		Port:     port,
		Protocol: "tcp",
		State:    "closed",
	}

	// 尝试连接
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		return nil // 端口未开放，不返回结果
	}
	defer conn.Close()

	// 端口开放
	portInfo.State = "open"
	result.Ports = append(result.Ports, portInfo)

	// 尝试获取服务banner
	banner := getBanner(conn, timeout)
	if banner != "" {
		portInfo.Banner = banner

		// 根据banner识别服务
		service := identifyService(port, banner)
		if service.Name != "" {
			result.Services = append(result.Services, service)
		}
	} else {
		// 根据端口识别常见服务
		service := identifyServiceByPort(port)
		if service.Name != "" {
			result.Services = append(result.Services, service)
		}
	}

	return result
}

// getBanner 获取服务banner
func getBanner(conn net.Conn, timeout time.Duration) string {
	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(timeout))

	// 发送一些数据触发响应
	// 不同的服务可能需要不同的触发数据
	_, err := conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	if err != nil {
		return ""
	}

	// 读取响应
	reader := bufio.NewReader(conn)
	var banner strings.Builder

	// 最多读取10行或1024字节
	for i := 0; i < 10; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		banner.WriteString(line)
		if banner.Len() > 1024 {
			break
		}
	}

	return strings.TrimSpace(banner.String())
}

// identifyService 根据banner识别服务
func identifyService(port int, banner string) models.ServiceInfo {
	service := models.ServiceInfo{
		Port:     port,
		Protocol: "tcp",
	}

	// 根据banner识别服务
	lowerBanner := strings.ToLower(banner)

	// HTTP服务
	if strings.HasPrefix(lowerBanner, "http/") || strings.Contains(lowerBanner, "server:") {
		service.Name = "http"
		// 提取服务器类型
		if serverIndex := strings.Index(lowerBanner, "server:"); serverIndex != -1 {
			serverEnd := strings.Index(lowerBanner[serverIndex:], "\r\n")
			if serverEnd != -1 {
				serverInfo := lowerBanner[serverIndex+7 : serverIndex+serverEnd]
				service.Product = strings.TrimSpace(serverInfo)
			}
		}
		return service
	}

	// SSH服务
	if strings.Contains(lowerBanner, "ssh") {
		service.Name = "ssh"
		// 提取版本信息
		if strings.Contains(lowerBanner, "openssh") {
			service.Product = "OpenSSH"
			// 提取版本号
			if versionIndex := strings.Index(lowerBanner, "openssh_"); versionIndex != -1 {
				versionEnd := strings.Index(lowerBanner[versionIndex:], " ")
				if versionEnd != -1 {
					service.Version = lowerBanner[versionIndex+8 : versionIndex+versionEnd]
				}
			}
		}
		return service
	}

	// FTP服务
	if strings.Contains(lowerBanner, "ftp") {
		service.Name = "ftp"
		// 提取版本信息
		if strings.Contains(lowerBanner, "filezilla") {
			service.Product = "FileZilla"
		} else if strings.Contains(lowerBanner, "vsftpd") {
			service.Product = "vsftpd"
		}
		return service
	}

	// SMTP服务
	if strings.Contains(lowerBanner, "smtp") {
		service.Name = "smtp"
		if strings.Contains(lowerBanner, "postfix") {
			service.Product = "Postfix"
		} else if strings.Contains(lowerBanner, "exim") {
			service.Product = "Exim"
		}
		return service
	}

	// 数据库服务
	if strings.Contains(lowerBanner, "mysql") {
		service.Name = "mysql"
		service.Product = "MySQL"
		return service
	}

	if strings.Contains(lowerBanner, "postgresql") {
		service.Name = "postgresql"
		service.Product = "PostgreSQL"
		return service
	}

	// 未识别，保存原始banner
	service.Name = "unknown"
	service.ExtraInfo = banner
	return service
}

// identifyServiceByPort 根据端口识别常见服务
func identifyServiceByPort(port int) models.ServiceInfo {
	service := models.ServiceInfo{
		Port:     port,
		Protocol: "tcp",
	}

	// 根据常见端口识别服务
	switch port {
	case 21:
		service.Name = "ftp"
	case 22:
		service.Name = "ssh"
	case 23:
		service.Name = "telnet"
	case 25:
		service.Name = "smtp"
	case 53:
		service.Name = "dns"
	case 80:
		service.Name = "http"
	case 110:
		service.Name = "pop3"
	case 143:
		service.Name = "imap"
	case 443:
		service.Name = "https"
	case 445:
		service.Name = "smb"
	case 993:
		service.Name = "imaps"
	case 995:
		service.Name = "pop3s"
	case 3306:
		service.Name = "mysql"
	case 3389:
		service.Name = "rdp"
	case 5432:
		service.Name = "postgresql"
	case 5900:
		service.Name = "vnc"
	case 8080:
		service.Name = "http-proxy"
	case 8443:
		service.Name = "https-alt"
	default:
		service.Name = "unknown"
	}

	return service
}
