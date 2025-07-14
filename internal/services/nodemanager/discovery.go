package nodemanager

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
)

// DiscoveryService 节点发现服务
type DiscoveryService struct {
	registry      *RegistryService
	discoverers   map[string]Discoverer
	config        *DiscoveryConfig
	stopChan      chan struct{}
	wg            sync.WaitGroup
	mu            sync.RWMutex
	discoveredNodes map[string]*DiscoveredNode
}

// DiscoveryConfig 发现配置
type DiscoveryConfig struct {
	Methods          []string      `json:"methods"`           // 发现方法：dns, broadcast, static, consul, etcd
	Interval         time.Duration `json:"interval"`          // 发现间隔
	BroadcastPort    int           `json:"broadcast_port"`    // 广播端口
	BroadcastAddress string        `json:"broadcast_address"` // 广播地址
	StaticNodes      []string      `json:"static_nodes"`      // 静态节点列表
	DNSName          string        `json:"dns_name"`          // DNS 名称
	ConsulConfig     *ConsulConfig `json:"consul_config"`     // Consul 配置
	ETCDConfig       *ETCDConfig   `json:"etcd_config"`       // ETCD 配置
}

// ConsulConfig Consul 配置
type ConsulConfig struct {
	Address    string `json:"address"`
	ServiceName string `json:"service_name"`
	Token      string `json:"token"`
}

// ETCDConfig ETCD 配置
type ETCDConfig struct {
	Endpoints []string `json:"endpoints"`
	Prefix    string   `json:"prefix"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
}

// DiscoveredNode 发现的节点
type DiscoveredNode struct {
	Address     string                 `json:"address"`
	Port        int                    `json:"port"`
	Type        models.NodeType        `json:"type"`
	Metadata    map[string]string      `json:"metadata"`
	LastSeen    time.Time              `json:"last_seen"`
	Source      string                 `json:"source"` // 发现来源
	HealthCheck *HealthCheckResult     `json:"health_check"`
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Healthy      bool          `json:"healthy"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error"`
	CheckTime    time.Time     `json:"check_time"`
}

// Discoverer 发现器接口
type Discoverer interface {
	Start(ctx context.Context) error
	Stop() error
	GetNodes() ([]*DiscoveredNode, error)
	GetName() string
}

// NewDiscoveryService 创建发现服务
func NewDiscoveryService(registry *RegistryService, config *DiscoveryConfig) *DiscoveryService {
	service := &DiscoveryService{
		registry:        registry,
		discoverers:     make(map[string]Discoverer),
		config:          config,
		stopChan:        make(chan struct{}),
		discoveredNodes: make(map[string]*DiscoveredNode),
	}

	// 初始化发现器
	service.initDiscoverers()

	return service
}

// Start 启动发现服务
func (s *DiscoveryService) Start(ctx context.Context) error {
	// 启动所有发现器
	for name, discoverer := range s.discoverers {
		s.wg.Add(1)
		go func(name string, d Discoverer) {
			defer s.wg.Done()
			if err := d.Start(ctx); err != nil {
				fmt.Printf("发现器 %s 启动失败: %v\n", name, err)
			}
		}(name, discoverer)
	}

	// 启动主发现循环
	s.wg.Add(1)
	go s.discoveryLoop(ctx)

	return nil
}

// Stop 停止发现服务
func (s *DiscoveryService) Stop() error {
	close(s.stopChan)

	// 停止所有发现器
	for _, discoverer := range s.discoverers {
		discoverer.Stop()
	}

	s.wg.Wait()
	return nil
}

// GetDiscoveredNodes 获取发现的节点
func (s *DiscoveryService) GetDiscoveredNodes() []*DiscoveredNode {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var nodes []*DiscoveredNode
	for _, node := range s.discoveredNodes {
		nodes = append(nodes, node)
	}

	// 按最后见到时间排序
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].LastSeen.After(nodes[j].LastSeen)
	})

	return nodes
}

// initDiscoverers 初始化发现器
func (s *DiscoveryService) initDiscoverers() {
	for _, method := range s.config.Methods {
		switch method {
		case "dns":
			s.discoverers["dns"] = NewDNSDiscoverer(s.config.DNSName)
		case "broadcast":
			s.discoverers["broadcast"] = NewBroadcastDiscoverer(s.config.BroadcastAddress, s.config.BroadcastPort)
		case "static":
			s.discoverers["static"] = NewStaticDiscoverer(s.config.StaticNodes)
		case "consul":
			if s.config.ConsulConfig != nil {
				s.discoverers["consul"] = NewConsulDiscoverer(s.config.ConsulConfig)
			}
		case "etcd":
			if s.config.ETCDConfig != nil {
				s.discoverers["etcd"] = NewETCDDiscoverer(s.config.ETCDConfig)
			}
		}
	}
}

// discoveryLoop 发现循环
func (s *DiscoveryService) discoveryLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.performDiscovery(ctx)
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// performDiscovery 执行发现
func (s *DiscoveryService) performDiscovery(ctx context.Context) {
	allNodes := make(map[string]*DiscoveredNode)

	// 从所有发现器收集节点
	for name, discoverer := range s.discoverers {
		nodes, err := discoverer.GetNodes()
		if err != nil {
			fmt.Printf("发现器 %s 获取节点失败: %v\n", name, err)
			continue
		}

		for _, node := range nodes {
			key := fmt.Sprintf("%s:%d", node.Address, node.Port)
			node.Source = name
			node.LastSeen = time.Now()
			
			// 执行健康检查
			node.HealthCheck = s.performHealthCheck(node.Address, node.Port)
			
			allNodes[key] = node
		}
	}

	// 更新发现的节点
	s.mu.Lock()
	s.discoveredNodes = allNodes
	s.mu.Unlock()

	// 自动注册健康的节点
	s.autoRegisterNodes(ctx, allNodes)
}

// performHealthCheck 执行健康检查
func (s *DiscoveryService) performHealthCheck(address string, port int) *HealthCheckResult {
	result := &HealthCheckResult{
		CheckTime: time.Now(),
	}

	start := time.Now()
	
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", address, port), 5*time.Second)
	if err != nil {
		result.Healthy = false
		result.Error = err.Error()
		result.ResponseTime = time.Since(start)
		return result
	}
	
	conn.Close()
	result.Healthy = true
	result.ResponseTime = time.Since(start)
	
	return result
}

// autoRegisterNodes 自动注册节点
func (s *DiscoveryService) autoRegisterNodes(ctx context.Context, nodes map[string]*DiscoveredNode) {
	for _, node := range nodes {
		// 只注册健康的节点
		if node.HealthCheck == nil || !node.HealthCheck.Healthy {
			continue
		}

		// 检查节点是否已注册
		existing, err := s.registry.findNodeByAddress(ctx, node.Address, node.Port)
		if err == nil && existing != nil {
			continue // 节点已存在
		}

		// 自动注册节点
		registration := &models.NodeRegistration{
			Name:         fmt.Sprintf("discovered-%s-%d", node.Address, node.Port),
			IP:           node.Address,
			Port:         node.Port,
			Type:         node.Type,
			Version:      "auto-discovered",
			Capabilities: []string{"auto-discovered"},
			Metadata:     node.Metadata,
			Secret:       "auto-discovery-secret", // 应该使用更安全的方式
		}

		if registration.Metadata == nil {
			registration.Metadata = make(map[string]string)
		}
		registration.Metadata["discovery_source"] = node.Source
		registration.Metadata["auto_registered"] = "true"

		// 注册节点
		_, err = s.registry.RegisterNode(ctx, registration)
		if err != nil {
			fmt.Printf("自动注册节点失败 %s:%d: %v\n", node.Address, node.Port, err)
		} else {
			fmt.Printf("自动注册节点成功: %s:%d\n", node.Address, node.Port)
		}
	}
}

// DNS 发现器
type DNSDiscoverer struct {
	dnsName string
}

func NewDNSDiscoverer(dnsName string) *DNSDiscoverer {
	return &DNSDiscoverer{dnsName: dnsName}
}

func (d *DNSDiscoverer) Start(ctx context.Context) error {
	return nil // DNS 发现器不需要持续运行
}

func (d *DNSDiscoverer) Stop() error {
	return nil
}

func (d *DNSDiscoverer) GetNodes() ([]*DiscoveredNode, error) {
	if d.dnsName == "" {
		return nil, nil
	}

	ips, err := net.LookupIP(d.dnsName)
	if err != nil {
		return nil, err
	}

	var nodes []*DiscoveredNode
	for _, ip := range ips {
		if ip.To4() != nil { // 只处理 IPv4
			node := &DiscoveredNode{
				Address: ip.String(),
				Port:    8090, // 默认端口
				Type:    models.NodeTypeWorker,
				Metadata: map[string]string{
					"dns_name": d.dnsName,
				},
			}
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

func (d *DNSDiscoverer) GetName() string {
	return "dns"
}

// 广播发现器
type BroadcastDiscoverer struct {
	address string
	port    int
	conn    *net.UDPConn
}

func NewBroadcastDiscoverer(address string, port int) *BroadcastDiscoverer {
	return &BroadcastDiscoverer{
		address: address,
		port:    port,
	}
}

func (d *BroadcastDiscoverer) Start(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", d.address, d.port))
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	d.conn = conn

	// 监听广播消息
	go d.listenForBroadcasts(ctx)

	return nil
}

func (d *BroadcastDiscoverer) Stop() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

func (d *BroadcastDiscoverer) GetNodes() ([]*DiscoveredNode, error) {
	// 广播发现器通过监听获取节点，这里返回空
	return nil, nil
}

func (d *BroadcastDiscoverer) GetName() string {
	return "broadcast"
}

func (d *BroadcastDiscoverer) listenForBroadcasts(ctx context.Context) {
	buffer := make([]byte, 1024)
	
	for {
		select {
		case <-ctx.Done():
			return
		default:
			d.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, addr, err := d.conn.ReadFromUDP(buffer)
			if err != nil {
				continue
			}

			// 解析广播消息
			var broadcastData struct {
				Type     string            `json:"type"`
				Port     int               `json:"port"`
				NodeType string            `json:"node_type"`
				Metadata map[string]string `json:"metadata"`
			}

			if err := json.Unmarshal(buffer[:n], &broadcastData); err != nil {
				continue
			}

			if broadcastData.Type == "node_announcement" {
				// 处理节点通告
				fmt.Printf("发现广播节点: %s:%d\n", addr.IP.String(), broadcastData.Port)
			}
		}
	}
}

// 静态发现器
type StaticDiscoverer struct {
	nodes []string
}

func NewStaticDiscoverer(nodes []string) *StaticDiscoverer {
	return &StaticDiscoverer{nodes: nodes}
}

func (d *StaticDiscoverer) Start(ctx context.Context) error {
	return nil
}

func (d *StaticDiscoverer) Stop() error {
	return nil
}

func (d *StaticDiscoverer) GetNodes() ([]*DiscoveredNode, error) {
	var discoveredNodes []*DiscoveredNode

	for _, nodeAddr := range d.nodes {
		host, port, err := net.SplitHostPort(nodeAddr)
		if err != nil {
			continue
		}

		portNum := 8090 // 默认端口
		if port != "" {
			if p, err := net.LookupPort("tcp", port); err == nil {
				portNum = p
			}
		}

		node := &DiscoveredNode{
			Address: host,
			Port:    portNum,
			Type:    models.NodeTypeWorker,
			Metadata: map[string]string{
				"static": "true",
			},
		}
		discoveredNodes = append(discoveredNodes, node)
	}

	return discoveredNodes, nil
}

func (d *StaticDiscoverer) GetName() string {
	return "static"
}

// Consul 发现器（占位实现）
type ConsulDiscoverer struct {
	config *ConsulConfig
}

func NewConsulDiscoverer(config *ConsulConfig) *ConsulDiscoverer {
	return &ConsulDiscoverer{config: config}
}

func (d *ConsulDiscoverer) Start(ctx context.Context) error {
	// TODO: 实现 Consul 客户端
	return nil
}

func (d *ConsulDiscoverer) Stop() error {
	return nil
}

func (d *ConsulDiscoverer) GetNodes() ([]*DiscoveredNode, error) {
	// TODO: 从 Consul 获取节点
	return nil, nil
}

func (d *ConsulDiscoverer) GetName() string {
	return "consul"
}

// ETCD 发现器（占位实现）
type ETCDDiscoverer struct {
	config *ETCDConfig
}

func NewETCDDiscoverer(config *ETCDConfig) *ETCDDiscoverer {
	return &ETCDDiscoverer{config: config}
}

func (d *ETCDDiscoverer) Start(ctx context.Context) error {
	// TODO: 实现 ETCD 客户端
	return nil
}

func (d *ETCDDiscoverer) Stop() error {
	return nil
}

func (d *ETCDDiscoverer) GetNodes() ([]*DiscoveredNode, error) {
	// TODO: 从 ETCD 获取节点
	return nil, nil
}

func (d *ETCDDiscoverer) GetName() string {
	return "etcd"
}