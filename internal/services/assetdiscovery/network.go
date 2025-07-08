package assetdiscovery

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NetworkScanner 网络扫描器
type NetworkScanner struct {
	concurrency   int
	timeout       time.Duration
	retryCount    int
	resultHandler ResultHandler
}

// NewNetworkScanner 创建网络扫描器
func NewNetworkScanner(concurrency int, timeout time.Duration, retryCount int, resultHandler ResultHandler) *NetworkScanner {
	if concurrency <= 0 {
		concurrency = 100
	}
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	if retryCount <= 0 {
		retryCount = 2
	}

	return &NetworkScanner{
		concurrency:   concurrency,
		timeout:       timeout,
		retryCount:    retryCount,
		resultHandler: resultHandler,
	}
}

// ScanNetwork 扫描网络
func (s *NetworkScanner) ScanNetwork(task *DiscoveryTask) error {
	// 创建网络扫描器
	scanner := NewNetworkScanner(
		task.Task.Config.Concurrency,
		time.Duration(task.Task.Config.Timeout)*time.Second,
		task.Task.Config.RetryCount,
		s.resultHandler,
	)

	// 解析目标
	var ipList []string
	for _, target := range task.Task.Targets {
		ips, err := parseTarget(target)
		if err != nil {
			continue
		}
		ipList = append(ipList, ips...)
	}

	// 排除IP
	if len(task.Task.Config.ExcludeIPs) > 0 {
		excludeMap := make(map[string]bool)
		for _, ip := range task.Task.Config.ExcludeIPs {
			excludeMap[ip] = true
		}

		var filteredIPs []string
		for _, ip := range ipList {
			if !excludeMap[ip] {
				filteredIPs = append(filteredIPs, ip)
			}
		}
		ipList = filteredIPs
	}

	// 更新任务总目标数
	totalTargets := len(ipList)
	if totalTargets == 0 {
		return fmt.Errorf("未找到有效的目标IP")
	}

	// 更新任务信息
	task.Task.ResultSummary.TotalTargets = totalTargets

	// 创建工作通道
	ipChan := make(chan string, totalTargets)
	resultChan := make(chan *models.DiscoveryResult, totalTargets)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < scanner.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range ipChan {
				// 检查任务是否被取消
				select {
				case <-task.Context.Done():
					return
				default:
				}

				// 扫描主机
				result := scanHost(ip, scanner.timeout, scanner.retryCount, task.Task.Config)
				if result != nil {
					resultChan <- result
				}
			}
		}()
	}

	// 发送任务到通道
	go func() {
		for _, ip := range ipList {
			ipChan <- ip
		}
		close(ipChan)
	}()

	// 处理结果
	go func() {
		scannedCount := 0
		discoveredCount := 0

		for result := range resultChan {
			scannedCount++

			// 如果主机存活，则计数
			if result.IsAlive {
				discoveredCount++

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
			task.Task.ResultSummary.DiscoveredAssets = discoveredCount
			task.Task.ResultSummary.HostCount = discoveredCount
		}
	}()

	// 等待所有工作协程完成
	wg.Wait()
	close(resultChan)

	// 给结果处理协程一些时间完成
	time.Sleep(1 * time.Second)

	return nil
}

// parseTarget 解析目标，支持IP、IP段、CIDR
func parseTarget(target string) ([]string, error) {
	// 检查是否是CIDR
	if _, ipnet, err := net.ParseCIDR(target); err == nil {
		return expandCIDR(ipnet)
	}

	// 检查是否是IP范围 (192.168.1.1-192.168.1.254)
	if isIPRange(target) {
		return expandIPRange(target)
	}

	// 检查是否是单个IP
	if net.ParseIP(target) != nil {
		return []string{target}, nil
	}

	// 尝试解析域名
	if ips, err := net.LookupIP(target); err == nil && len(ips) > 0 {
		var result []string
		for _, ip := range ips {
			if ipv4 := ip.To4(); ipv4 != nil {
				result = append(result, ipv4.String())
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("无法解析目标: %s", target)
}

// expandCIDR 展开CIDR为IP列表
func expandCIDR(ipnet *net.IPNet) ([]string, error) {
	var ips []string

	// 获取CIDR的起始IP和掩码长度
	ip := ipnet.IP.To4()
	if ip == nil {
		return nil, fmt.Errorf("仅支持IPv4 CIDR")
	}

	// 计算IP数量
	mask := ipnet.Mask
	ones, bits := mask.Size()
	if bits != 32 {
		return nil, fmt.Errorf("无效的IPv4掩码")
	}

	// 计算主机数量
	hostCount := 1 << (bits - ones)

	// 如果主机数量太大，可能导致内存问题
	if hostCount > 65536 {
		return nil, fmt.Errorf("CIDR范围太大，包含超过65536个IP")
	}

	// 获取起始IP的整数表示
	start := uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])

	// 遍历所有IP
	for i := uint32(0); i < uint32(hostCount); i++ {
		// 跳过网络地址和广播地址
		if i == 0 || i == uint32(hostCount-1) {
			continue
		}

		// 计算当前IP
		current := start + i

		// 转换回IP字符串
		ipStr := fmt.Sprintf("%d.%d.%d.%d",
			(current>>24)&0xFF,
			(current>>16)&0xFF,
			(current>>8)&0xFF,
			current&0xFF,
		)

		ips = append(ips, ipStr)
	}

	return ips, nil
}

// isIPRange 检查是否是IP范围
func isIPRange(target string) bool {
	// TODO: 实现IP范围检查
	return false
}

// expandIPRange 展开IP范围为IP列表
func expandIPRange(ipRange string) ([]string, error) {
	// TODO: 实现IP范围展开
	return nil, fmt.Errorf("IP范围展开功能尚未实现")
}

// scanHost 扫描单个主机
func scanHost(ip string, timeout time.Duration, retryCount int, config models.DiscoveryConfig) *models.DiscoveryResult {
	// 创建结果
	result := &models.DiscoveryResult{
		Target:    ip,
		IP:        ip,
		AssetType: "host",
		IsAlive:   false,
	}

	// 检查主机是否存活
	alive := checkHostAlive(ip, timeout, retryCount)
	if !alive && config.OnlyAliveHosts {
		return nil // 只返回活跃主机
	}

	result.IsAlive = alive

	// 如果主机存活，进行进一步检测
	if alive {
		// 如果配置了端口扫描，则扫描端口
		if len(config.PortRanges) > 0 {
			// TODO: 实现端口扫描
		}

		// 如果配置了操作系统检测，则检测操作系统
		if config.OSDetect {
			// TODO: 实现操作系统检测
		}

		// 如果配置了服务检测，则检测服务
		if config.ServiceDetect {
			// TODO: 实现服务检测
		}
	}

	return result
}

// checkHostAlive 检查主机是否存活
func checkHostAlive(ip string, timeout time.Duration, retryCount int) bool {
	// 尝试ICMP ping
	if pingHost(ip, timeout) {
		return true
	}

	// 尝试TCP连接常见端口
	commonPorts := []int{80, 443, 22, 21, 25, 3389, 8080, 8443}
	for _, port := range commonPorts {
		if checkTCPPort(ip, port, timeout) {
			return true
		}
	}

	return false
}

// pingHost 使用ICMP ping主机
func pingHost(ip string, timeout time.Duration) bool {
	// TODO: 实现ICMP ping
	// 注意：在某些系统上可能需要root权限
	return false
}

// checkTCPPort 检查TCP端口是否开放
func checkTCPPort(ip string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
