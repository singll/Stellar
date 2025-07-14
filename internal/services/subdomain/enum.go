package subdomain

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/miekg/dns"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/time/rate"
)

// Enumerator 子域名枚举器接口
type Enumerator interface {
	// EnumerateSubdomains 枚举子域名
	EnumerateSubdomains(ctx context.Context, rootDomain string, taskID string, projectID string, config models.SubdomainEnumConfig) error
	// StopEnumeration 停止枚举
	StopEnumeration(taskID string) error
	// GetEnumerationStatus 获取枚举状态
	GetEnumerationStatus(taskID string) (string, error)
	// GetEnumerationProgress 获取枚举进度
	GetEnumerationProgress(taskID string) (float64, error)
}

// SubdomainEnumerator 子域名枚举器实现
type SubdomainEnumerator struct {
	db       *mongo.Database
	resolver *DNSResolver
	config   models.SubdomainEnumConfig
	tasks    map[string]*EnumService
	mutex    sync.RWMutex
}

// NewEnumerator 创建子域名枚举器
func NewEnumerator(db *mongo.Database, resolver *DNSResolver, config models.SubdomainEnumConfig) *SubdomainEnumerator {
	return &SubdomainEnumerator{
		db:       db,
		resolver: resolver,
		config:   config,
		tasks:    make(map[string]*EnumService),
	}
}

// EnumerateSubdomains 枚举子域名
func (e *SubdomainEnumerator) EnumerateSubdomains(ctx context.Context, rootDomain string, taskID string, projectID string, config models.SubdomainEnumConfig) error {
	// 创建枚举服务
	service := NewEnumService(config, rootDomain, taskID, projectID)

	// 保存任务
	e.mutex.Lock()
	e.tasks[taskID] = service
	e.mutex.Unlock()

	// 启动枚举
	go func() {
		err := service.Start(ctx)
		if err != nil {
			log.Printf("子域名枚举失败: %v", err)
		}

		// 任务完成后从映射中删除
		e.mutex.Lock()
		delete(e.tasks, taskID)
		e.mutex.Unlock()
	}()

	return nil
}

// StopEnumeration 停止枚举
func (e *SubdomainEnumerator) StopEnumeration(taskID string) error {
	e.mutex.RLock()
	service, exists := e.tasks[taskID]
	e.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("任务不存在")
	}

	service.Stop()
	return nil
}

// GetEnumerationStatus 获取枚举状态
func (e *SubdomainEnumerator) GetEnumerationStatus(taskID string) (string, error) {
	e.mutex.RLock()
	_, exists := e.tasks[taskID]
	e.mutex.RUnlock()

	if exists {
		return "running", nil
	}

	// 从数据库查询任务状态
	// TODO: 实现从数据库查询任务状态的逻辑

	return "completed", nil
}

// GetEnumerationProgress 获取枚举进度
func (e *SubdomainEnumerator) GetEnumerationProgress(taskID string) (float64, error) {
	e.mutex.RLock()
	service, exists := e.tasks[taskID]
	e.mutex.RUnlock()

	if !exists {
		// 从数据库查询进度
		// TODO: 实现从数据库查询进度的逻辑
		return 100.0, nil
	}

	// 从通道读取最新进度
	select {
	case progress := <-service.ProgressChan:
		return progress, nil
	default:
		// 如果通道中没有数据，返回0
		return 0.0, nil
	}
}

// EnumService 子域名枚举服务
type EnumService struct {
	// 配置
	Config models.SubdomainEnumConfig
	// 结果通道
	ResultChan chan *models.SubdomainResult
	// 任务ID
	TaskID string
	// 项目ID
	ProjectID string
	// 根域名
	RootDomain string
	// 停止信号
	StopChan chan struct{}
	// 进度通道
	ProgressChan chan float64
	// 已发现的子域名集合
	foundSubdomains sync.Map
	// DNS解析器
	resolver *DNSResolver
	// 速率限制器
	limiter *rate.Limiter
}

// NewEnumService 创建子域名枚举服务
func NewEnumService(config models.SubdomainEnumConfig, rootDomain, taskID, projectID string) *EnumService {
	// 创建速率限制器
	var limiter *rate.Limiter
	if config.RateLimit > 0 {
		limiter = rate.NewLimiter(rate.Limit(config.RateLimit), config.RateLimit)
	}

	// 创建DNS解析器
	resolver := NewDNSResolver(config.ResolverServers, config.Timeout, config.RetryCount)

	return &EnumService{
		Config:       config,
		ResultChan:   make(chan *models.SubdomainResult, 1000),
		TaskID:       taskID,
		ProjectID:    projectID,
		RootDomain:   rootDomain,
		StopChan:     make(chan struct{}),
		ProgressChan: make(chan float64, 100),
		resolver:     resolver,
		limiter:      limiter,
	}
}

// Start 开始子域名枚举
func (s *EnumService) Start(ctx context.Context) error {
	// 创建工作组
	var wg sync.WaitGroup

	// 启动结果处理协程
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.processResults(ctx)
	}()

	// 根据配置的方法执行不同的枚举策略
	for _, method := range s.Config.Methods {
		switch method {
		case "dns_brute":
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := s.doBruteForceDNS(ctx)
				if err != nil {
					log.Printf("DNS爆破失败: %v", err)
				}
			}()
		case "dns_zone_transfer":
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := s.doZoneTransfer(ctx)
				if err != nil {
					log.Printf("DNS区域传输失败: %v", err)
				}
			}()
		case "search_engines":
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := s.doSearchEngines(ctx)
				if err != nil {
					log.Printf("搜索引擎查询失败: %v", err)
				}
			}()
		case "certificate_transparency":
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := s.doCertificateTransparency(ctx)
				if err != nil {
					log.Printf("证书透明度日志查询失败: %v", err)
				}
			}()
		}
	}

	// 等待所有协程完成
	wg.Wait()
	close(s.ResultChan)
	close(s.ProgressChan)

	return nil
}

// Stop 停止子域名枚举
func (s *EnumService) Stop() {
	close(s.StopChan)
}

// processResults 处理结果
func (s *EnumService) processResults(ctx context.Context) {
	// 在这里处理结果，例如去重、验证、保存到数据库等
	// 这个函数会在单独的协程中运行
	for result := range s.ResultChan {
		// 检查是否已经存在
		if _, exists := s.foundSubdomains.Load(result.Subdomain); exists {
			continue
		}

		// 标记为已发现
		s.foundSubdomains.Store(result.Subdomain, true)

		// 如果需要验证子域名
		if s.Config.VerifySubdomains {
			s.verifySubdomain(ctx, result)
		}

		// 如果需要检查是否可能被接管
		s.checkTakeOver(result)

		// 如果需要保存到数据库
		if s.Config.SaveToDB {
			// 这里应该有保存到数据库的代码
			// saveToDatabase(result)
		}

		// 如果需要递归搜索
		if s.Config.RecursiveSearch && result.IsResolved {
			// 这里应该有递归搜索的代码
			// s.doRecursiveSearch(ctx, result.Subdomain)
		}
	}
}

// doBruteForceDNS 执行DNS爆破
func (s *EnumService) doBruteForceDNS(ctx context.Context) error {
	// 打开字典文件
	file, err := os.Open(s.Config.DictionaryPath)
	if err != nil {
		return fmt.Errorf("打开字典文件失败: %v", err)
	}
	defer file.Close()

	// 创建工作池
	workerCount := s.Config.Concurrency
	if workerCount <= 0 {
		workerCount = 10 // 默认并发数
	}

	// 创建任务通道
	taskChan := make(chan string, workerCount*2)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for subdomain := range taskChan {
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

					// 构造完整域名
					fullDomain := subdomain + "." + s.RootDomain

					// 解析域名
					records, err := s.resolver.Resolve(fullDomain)
					if err != nil {
						continue
					}

					// 如果解析成功，创建结果
					if len(records) > 0 {
						taskID, _ := primitive.ObjectIDFromHex(s.TaskID)
						projectID, _ := primitive.ObjectIDFromHex(s.ProjectID)

						result := &models.SubdomainResult{
							TaskID:     taskID,
							ProjectID:  projectID,
							RootDomain: s.RootDomain,
							Subdomain:  fullDomain,
							Records:    records,
							IsResolved: true,
							Source:     "dns_brute",
							CreatedAt:  time.Now(),
							UpdatedAt:  time.Now(),
						}

						// 提取IP和CNAME
						for _, record := range records {
							if record.Type == "A" || record.Type == "AAAA" {
								result.IPs = append(result.IPs, record.Value)
							} else if record.Type == "CNAME" {
								result.CNAME = record.Value
							}
						}

						// 发送结果
						select {
						case s.ResultChan <- result:
						default:
							// 通道已满，丢弃结果
						}
					}
				}
			}
		}()
	}

	// 读取字典文件并发送任务
	scanner := bufio.NewScanner(file)
	lineCount := 0
	totalLines := 0

	// 首先计算总行数
	for scanner.Scan() {
		totalLines++
	}

	// 重置文件指针
	file.Seek(0, 0)
	scanner = bufio.NewScanner(file)

	// 发送任务
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			close(taskChan)
			return ctx.Err()
		case <-s.StopChan:
			close(taskChan)
			return nil
		default:
			subdomain := strings.TrimSpace(scanner.Text())
			if subdomain != "" && !strings.HasPrefix(subdomain, "#") {
				taskChan <- subdomain
			}
			lineCount++

			// 更新进度
			if lineCount%100 == 0 {
				progress := float64(lineCount) / float64(totalLines) * 100
				select {
				case s.ProgressChan <- progress:
				default:
					// 通道已满，丢弃进度更新
				}
			}
		}
	}

	// 关闭任务通道
	close(taskChan)

	// 等待所有工作协程完成
	wg.Wait()

	return nil
}

// doZoneTransfer 执行DNS区域传输
func (s *EnumService) doZoneTransfer(ctx context.Context) error {
	// 首先获取域名的NS记录
	nsRecords, err := s.resolver.LookupNS(s.RootDomain)
	if err != nil {
		return fmt.Errorf("获取NS记录失败: %v", err)
	}

	// 对每个NS服务器尝试区域传输
	for _, ns := range nsRecords {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.StopChan:
			return nil
		default:
			// 尝试区域传输
			subdomains, err := s.zoneTransfer(ns, s.RootDomain)
			if err != nil {
				continue
			}

			// 处理结果
			for _, subdomain := range subdomains {
				// 解析域名
				records, err := s.resolver.Resolve(subdomain)
				if err != nil {
					continue
				}

				// 如果解析成功，创建结果
				if len(records) > 0 {
					taskID, _ := primitive.ObjectIDFromHex(s.TaskID)
					projectID, _ := primitive.ObjectIDFromHex(s.ProjectID)

					result := &models.SubdomainResult{
						TaskID:     taskID,
						ProjectID:  projectID,
						RootDomain: s.RootDomain,
						Subdomain:  subdomain,
						Records:    records,
						IsResolved: true,
						Source:     "dns_zone_transfer",
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					}

					// 提取IP和CNAME
					for _, record := range records {
						if record.Type == "A" || record.Type == "AAAA" {
							result.IPs = append(result.IPs, record.Value)
						} else if record.Type == "CNAME" {
							result.CNAME = record.Value
						}
					}

					// 发送结果
					select {
					case s.ResultChan <- result:
					default:
						// 通道已满，丢弃结果
					}
				}
			}
		}
	}

	return nil
}

// zoneTransfer 执行区域传输
func (s *EnumService) zoneTransfer(nameserver, domain string) ([]string, error) {
	// 创建DNS消息
	m := new(dns.Msg)
	m.SetAxfr(dns.Fqdn(domain))

	// 创建传输
	t := new(dns.Transfer)
	env, err := t.In(m, nameserver+":53")
	if err != nil {
		return nil, err
	}

	// 收集子域名
	var subdomains []string
	for e := range env {
		if e.Error != nil {
			return nil, e.Error
		}
		for _, rr := range e.RR {
			if rr.Header().Name != "" {
				subdomains = append(subdomains, strings.TrimSuffix(rr.Header().Name, "."))
			}
		}
	}

	return subdomains, nil
}

// doSearchEngines 从搜索引擎获取子域名
func (s *EnumService) doSearchEngines(ctx context.Context) error {
	// 通过特定的搜索引擎查询语法查找子域名
	// 这里使用简单的搜索引擎查询，实际应用中可能需要API密钥
	
	// 构造搜索查询
	queries := []string{
		fmt.Sprintf("site:*.%s", s.RootDomain),
		fmt.Sprintf("site:%s", s.RootDomain),
	}
	
	// 这里应该实现HTTP请求到搜索引擎并解析结果
	// 由于搜索引擎反爬虫机制，这里只是示例框架
	// 在实际应用中需要使用API或代理服务
	
	log.Printf("搜索引擎查询: %v", queries)
	
	// 模拟找到的子域名（实际应用中需要解析搜索结果）
	// 这里应该有HTTP请求和HTML解析逻辑
	
	return nil
}

// doCertificateTransparency 从证书透明度日志获取子域名
func (s *EnumService) doCertificateTransparency(ctx context.Context) error {
	// 使用证书透明度日志服务查询子域名
	// 这里使用crt.sh作为示例，实际应用中可以使用多个CT日志源
	
	// 构造查询URL
	url := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", s.RootDomain)
	
	// 这里应该实现HTTP请求到CT日志服务并解析JSON结果
	// 由于需要HTTP客户端和JSON解析，这里只是示例框架
	
	log.Printf("证书透明度日志查询: %s", url)
	
	// 模拟处理CT日志结果
	// 实际应用中需要：
	// 1. 发送HTTP请求到CT日志服务
	// 2. 解析JSON响应
	// 3. 提取子域名
	// 4. 去重和验证
	// 5. 发送到结果通道
	
	return nil
}

// verifySubdomain 验证子域名
func (s *EnumService) verifySubdomain(ctx context.Context, result *models.SubdomainResult) {
	// 检查是否为泛解析域名
	isWildcard, err := s.checkWildcard(result.RootDomain)
	if err == nil && isWildcard {
		result.IsWildcard = true
	}

	// 其他验证逻辑...
}

// checkWildcard 检查是否为泛解析域名
func (s *EnumService) checkWildcard(domain string) (bool, error) {
	// 生成随机子域名
	randomSubdomain := fmt.Sprintf("%d.%s", time.Now().UnixNano(), domain)

	// 尝试解析
	_, err := s.resolver.Resolve(randomSubdomain)
	if err == nil {
		// 如果能解析成功，说明是泛解析
		return true, nil
	}

	return false, nil
}

// checkTakeOver 检查子域名是否可能被接管
func (s *EnumService) checkTakeOver(result *models.SubdomainResult) {
	// 如果有CNAME记录但无法解析，可能存在接管风险
	if result.CNAME != "" && !result.IsResolved {
		// 检查是否匹配已知的可接管服务
		// 这里应该有更复杂的逻辑，例如检查CNAME是否指向已知的可接管服务
		result.IsTakeOver = true
		result.TakeOverType = "unresolved_cname"
	}
}
