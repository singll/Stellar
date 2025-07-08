package subdomain

import (
	"fmt"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/miekg/dns"
)

// DNSResolver DNS解析器
type DNSResolver struct {
	// DNS服务器列表
	Servers []string
	// 超时时间(秒)
	Timeout int
	// 重试次数
	RetryCount int
	// DNS客户端
	client *dns.Client
}

// NewResolver 创建默认DNS解析器
func NewResolver() *DNSResolver {
	return NewDNSResolver(nil, 5, 3)
}

// NewDNSResolver 创建DNS解析器
func NewDNSResolver(servers []string, timeout, retryCount int) *DNSResolver {
	// 如果没有指定DNS服务器，使用默认服务器
	if len(servers) == 0 {
		servers = []string{
			"8.8.8.8:53",         // Google DNS
			"8.8.4.4:53",         // Google DNS
			"1.1.1.1:53",         // Cloudflare DNS
			"1.0.0.1:53",         // Cloudflare DNS
			"114.114.114.114:53", // 114 DNS
		}
	}

	// 设置默认超时时间
	if timeout <= 0 {
		timeout = 5
	}

	// 设置默认重试次数
	if retryCount <= 0 {
		retryCount = 3
	}

	// 创建DNS客户端
	client := &dns.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	return &DNSResolver{
		Servers:    servers,
		Timeout:    timeout,
		RetryCount: retryCount,
		client:     client,
	}
}

// Resolve 解析域名
func (r *DNSResolver) Resolve(domain string) ([]models.DNSRecord, error) {
	var records []models.DNSRecord

	// 尝试解析A记录
	aRecords, err := r.lookupA(domain)
	if err == nil {
		records = append(records, aRecords...)
	}

	// 尝试解析AAAA记录
	aaaaRecords, err := r.lookupAAAA(domain)
	if err == nil {
		records = append(records, aaaaRecords...)
	}

	// 尝试解析CNAME记录
	cnameRecords, err := r.lookupCNAME(domain)
	if err == nil {
		records = append(records, cnameRecords...)
	}

	// 如果没有找到任何记录，返回错误
	if len(records) == 0 {
		return nil, fmt.Errorf("未找到任何DNS记录")
	}

	return records, nil
}

// lookupA 查询A记录
func (r *DNSResolver) lookupA(domain string) ([]models.DNSRecord, error) {
	var records []models.DNSRecord

	// 创建DNS消息
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	m.RecursionDesired = true

	// 发送查询
	resp, err := r.query(m)
	if err != nil {
		return nil, err
	}

	// 解析响应
	for _, ans := range resp.Answer {
		if a, ok := ans.(*dns.A); ok {
			records = append(records, models.DNSRecord{
				Type:  "A",
				Value: a.A.String(),
				TTL:   int(a.Hdr.Ttl),
			})
		}
	}

	return records, nil
}

// lookupAAAA 查询AAAA记录
func (r *DNSResolver) lookupAAAA(domain string) ([]models.DNSRecord, error) {
	var records []models.DNSRecord

	// 创建DNS消息
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeAAAA)
	m.RecursionDesired = true

	// 发送查询
	resp, err := r.query(m)
	if err != nil {
		return nil, err
	}

	// 解析响应
	for _, ans := range resp.Answer {
		if aaaa, ok := ans.(*dns.AAAA); ok {
			records = append(records, models.DNSRecord{
				Type:  "AAAA",
				Value: aaaa.AAAA.String(),
				TTL:   int(aaaa.Hdr.Ttl),
			})
		}
	}

	return records, nil
}

// lookupCNAME 查询CNAME记录
func (r *DNSResolver) lookupCNAME(domain string) ([]models.DNSRecord, error) {
	var records []models.DNSRecord

	// 创建DNS消息
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeCNAME)
	m.RecursionDesired = true

	// 发送查询
	resp, err := r.query(m)
	if err != nil {
		return nil, err
	}

	// 解析响应
	for _, ans := range resp.Answer {
		if cname, ok := ans.(*dns.CNAME); ok {
			records = append(records, models.DNSRecord{
				Type:  "CNAME",
				Value: cname.Target,
				TTL:   int(cname.Hdr.Ttl),
			})
		}
	}

	return records, nil
}

// LookupNS 查询NS记录
func (r *DNSResolver) LookupNS(domain string) ([]string, error) {
	var nameservers []string

	// 创建DNS消息
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	m.RecursionDesired = true

	// 发送查询
	resp, err := r.query(m)
	if err != nil {
		return nil, err
	}

	// 解析响应
	for _, ans := range resp.Answer {
		if ns, ok := ans.(*dns.NS); ok {
			nameservers = append(nameservers, ns.Ns)
		}
	}

	// 如果没有找到任何NS记录，返回错误
	if len(nameservers) == 0 {
		return nil, fmt.Errorf("未找到任何NS记录")
	}

	return nameservers, nil
}

// query 发送DNS查询
func (r *DNSResolver) query(m *dns.Msg) (*dns.Msg, error) {
	var resp *dns.Msg
	var err error

	// 尝试每个DNS服务器
	for _, server := range r.Servers {
		// 重试指定次数
		for i := 0; i < r.RetryCount; i++ {
			resp, _, err = r.client.Exchange(m, server)
			if err == nil && resp != nil && resp.Rcode == dns.RcodeSuccess {
				return resp, nil
			}
			// 如果查询失败，等待一段时间后重试
			time.Sleep(time.Duration(i*100) * time.Millisecond)
		}
	}

	// 如果所有DNS服务器都查询失败，返回最后一个错误
	if err != nil {
		return nil, err
	}

	// 如果没有错误但响应为空或响应码不是成功，返回一个通用错误
	return nil, fmt.Errorf("DNS查询失败")
}
