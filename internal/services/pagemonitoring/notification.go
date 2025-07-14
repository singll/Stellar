package pagemonitoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationService 通知服务
type NotificationService struct {
	config   *NotificationConfig
	channels map[string]NotificationChannel
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	// 邮件配置
	Email EmailConfig `json:"email"`
	// Webhook配置
	Webhook WebhookConfig `json:"webhook"`
	// 钉钉配置
	DingTalk DingTalkConfig `json:"dingtalk"`
	// 微信配置
	WeChat WeChatConfig `json:"wechat"`
	// Slack配置
	Slack SlackConfig `json:"slack"`
	// 默认通知级别
	DefaultLevel NotificationLevel `json:"default_level"`
	// 启用的通知渠道
	EnabledChannels []string `json:"enabled_channels"`
	// 通知频率限制
	RateLimit RateLimitConfig `json:"rate_limit"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	FromAddress  string `json:"from_address"`
	ToAddresses  []string `json:"to_addresses"`
	UseTLS       bool   `json:"use_tls"`
	Subject      string `json:"subject"`
}

// WebhookConfig Webhook配置
type WebhookConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Timeout int               `json:"timeout"`
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	WebhookURL string `json:"webhook_url"`
	Secret     string `json:"secret"`
	AtMobiles  []string `json:"at_mobiles"`
	AtAll      bool   `json:"at_all"`
}

// WeChatConfig 微信配置
type WeChatConfig struct {
	WebhookURL string `json:"webhook_url"`
	CorpID     string `json:"corp_id"`
	CorpSecret string `json:"corp_secret"`
	AgentID    int    `json:"agent_id"`
}

// SlackConfig Slack配置
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
	IconEmoji  string `json:"icon_emoji"`
}

// RateLimitConfig 频率限制配置
type RateLimitConfig struct {
	MaxPerHour   int           `json:"max_per_hour"`
	MaxPerDay    int           `json:"max_per_day"`
	CooldownTime time.Duration `json:"cooldown_time"`
}

// NotificationMessage 通知消息
type NotificationMessage struct {
	ID        primitive.ObjectID     `json:"id"`
	Type      NotificationType       `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Level     NotificationLevel      `json:"level"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	SentAt    *time.Time             `json:"sent_at,omitempty"`
	Status    NotificationStatus     `json:"status"`
	Channels  []string               `json:"channels"`
	RetryCount int                   `json:"retry_count"`
}

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypePageChange NotificationType = "page_change"
	NotificationTypeTaskError  NotificationType = "task_error"
	NotificationTypeSystemAlert NotificationType = "system_alert"
	NotificationTypeTaskStart  NotificationType = "task_start"
	NotificationTypeTaskComplete NotificationType = "task_complete"
)

// NotificationLevel 通知级别
type NotificationLevel string

const (
	NotificationLevelInfo     NotificationLevel = "info"
	NotificationLevelWarning  NotificationLevel = "warning"
	NotificationLevelError    NotificationLevel = "error"
	NotificationLevelCritical NotificationLevel = "critical"
)

// NotificationStatus 通知状态
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
	NotificationStatusRetrying NotificationStatus = "retrying"
)

// NotificationChannel 通知渠道接口
type NotificationChannel interface {
	Send(message *NotificationMessage) error
	GetName() string
	IsEnabled() bool
}

// NewNotificationService 创建通知服务
func NewNotificationService() *NotificationService {
	service := &NotificationService{
		channels: make(map[string]NotificationChannel),
		config: &NotificationConfig{
			DefaultLevel: NotificationLevelInfo,
			RateLimit: RateLimitConfig{
				MaxPerHour:   100,
				MaxPerDay:    1000,
				CooldownTime: 5 * time.Minute,
			},
		},
	}

	// 初始化默认通知渠道
	service.initDefaultChannels()
	
	return service
}

// LoadConfig 加载配置
func (s *NotificationService) LoadConfig(config *NotificationConfig) {
	s.config = config
	s.initChannelsFromConfig()
}

// initDefaultChannels 初始化默认通知渠道
func (s *NotificationService) initDefaultChannels() {
	// 注册邮件通知渠道
	s.channels["email"] = &EmailChannel{service: s}
	
	// 注册Webhook通知渠道
	s.channels["webhook"] = &WebhookChannel{service: s}
	
	// 注册钉钉通知渠道
	s.channels["dingtalk"] = &DingTalkChannel{service: s}
	
	// 注册微信通知渠道
	s.channels["wechat"] = &WeChatChannel{service: s}
	
	// 注册Slack通知渠道
	s.channels["slack"] = &SlackChannel{service: s}
}

// initChannelsFromConfig 根据配置初始化通知渠道
func (s *NotificationService) initChannelsFromConfig() {
	// 这里可以根据配置动态启用/禁用通知渠道
	for _, channelName := range s.config.EnabledChannels {
		if channel, exists := s.channels[channelName]; exists {
			log.Printf("启用通知渠道: %s", channel.GetName())
		}
	}
}

// Send 发送通知
func (s *NotificationService) Send(message *NotificationMessage) error {
	if message.ID.IsZero() {
		message.ID = primitive.NewObjectID()
	}
	
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}
	
	// 应用频率限制
	if !s.checkRateLimit(message) {
		log.Printf("通知被频率限制跳过: %s", message.Title)
		return fmt.Errorf("通知频率超过限制")
	}
	
	// 选择通知渠道
	channels := s.selectChannels(message)
	if len(channels) == 0 {
		return fmt.Errorf("没有可用的通知渠道")
	}
	
	var lastError error
	sentCount := 0
	
	// 向每个渠道发送通知
	for _, channelName := range channels {
		channel, exists := s.channels[channelName]
		if !exists || !channel.IsEnabled() {
			continue
		}
		
		if err := channel.Send(message); err != nil {
			log.Printf("通过渠道 %s 发送通知失败: %v", channelName, err)
			lastError = err
		} else {
			log.Printf("通过渠道 %s 发送通知成功: %s", channelName, message.Title)
			sentCount++
		}
	}
	
	// 更新消息状态
	if sentCount > 0 {
		now := time.Now()
		message.SentAt = &now
		message.Status = NotificationStatusSent
		return nil
	} else {
		message.Status = NotificationStatusFailed
		return fmt.Errorf("所有通知渠道发送失败: %v", lastError)
	}
}

// checkRateLimit 检查频率限制
func (s *NotificationService) checkRateLimit(message *NotificationMessage) bool {
	// 这里简化实现，实际应该使用Redis或内存缓存来跟踪发送频率
	// 可以根据消息类型、级别等进行不同的频率限制
	return true
}

// selectChannels 选择通知渠道
func (s *NotificationService) selectChannels(message *NotificationMessage) []string {
	// 如果消息指定了通知渠道，使用指定的渠道
	if len(message.Channels) > 0 {
		return message.Channels
	}
	
	// 根据通知级别选择合适的渠道
	var channels []string
	switch message.Level {
	case NotificationLevelCritical:
		// 关键级别：使用所有可用渠道
		channels = append(channels, "email", "dingtalk", "wechat", "slack", "webhook")
	case NotificationLevelError:
		// 错误级别：使用邮件和即时通信工具
		channels = append(channels, "email", "dingtalk", "wechat")
	case NotificationLevelWarning:
		// 警告级别：使用即时通信工具
		channels = append(channels, "dingtalk", "wechat")
	case NotificationLevelInfo:
		// 信息级别：使用Webhook
		channels = append(channels, "webhook")
	}
	
	// 过滤启用的渠道
	var enabledChannels []string
	for _, channel := range channels {
		if s.isChannelEnabled(channel) {
			enabledChannels = append(enabledChannels, channel)
		}
	}
	
	return enabledChannels
}

// isChannelEnabled 检查渠道是否启用
func (s *NotificationService) isChannelEnabled(channelName string) bool {
	if len(s.config.EnabledChannels) == 0 {
		// 如果没有配置启用渠道，默认启用Webhook
		return channelName == "webhook"
	}
	
	for _, enabled := range s.config.EnabledChannels {
		if enabled == channelName {
			return true
		}
	}
	return false
}

// SendPageChangeNotification 发送页面变更通知
func (s *NotificationService) SendPageChangeNotification(taskName, url string, similarity float64, diff string) error {
	level := NotificationLevelInfo
	title := fmt.Sprintf("页面变更检测: %s", taskName)
	
	// 根据相似度调整通知级别
	if similarity < 0.3 {
		level = NotificationLevelCritical
		title = fmt.Sprintf("重大页面变更: %s", taskName)
	} else if similarity < 0.7 {
		level = NotificationLevelWarning
		title = fmt.Sprintf("页面变更警告: %s", taskName)
	}
	
	message := &NotificationMessage{
		Type:    NotificationTypePageChange,
		Title:   title,
		Message: fmt.Sprintf("检测到页面 %s 发生变更，相似度: %.2f", url, similarity),
		Level:   level,
		Data: map[string]interface{}{
			"task_name":  taskName,
			"url":        url,
			"similarity": similarity,
			"diff":       diff,
			"timestamp":  time.Now(),
		},
	}
	
	return s.Send(message)
}

// SendTaskErrorNotification 发送任务错误通知
func (s *NotificationService) SendTaskErrorNotification(taskName, url, errorMsg string) error {
	message := &NotificationMessage{
		Type:    NotificationTypeTaskError,
		Title:   fmt.Sprintf("监控任务失败: %s", taskName),
		Message: fmt.Sprintf("任务 %s 执行失败: %s", taskName, errorMsg),
		Level:   NotificationLevelError,
		Data: map[string]interface{}{
			"task_name": taskName,
			"url":       url,
			"error":     errorMsg,
			"timestamp": time.Now(),
		},
	}
	
	return s.Send(message)
}

// EmailChannel 邮件通知渠道
type EmailChannel struct {
	service *NotificationService
}

func (c *EmailChannel) GetName() string {
	return "email"
}

func (c *EmailChannel) IsEnabled() bool {
	config := c.service.config.Email
	return config.SMTPHost != "" && config.Username != "" && len(config.ToAddresses) > 0
}

func (c *EmailChannel) Send(message *NotificationMessage) error {
	config := c.service.config.Email
	
	// 构建邮件内容
	subject := config.Subject
	if subject == "" {
		subject = "Stellar 监控通知"
	}
	subject = fmt.Sprintf("%s - %s", subject, message.Title)
	
	body := c.buildEmailBody(message)
	
	// 发送邮件
	auth := smtp.PlainAuth("", config.Username, config.Password, config.SMTPHost)
	
	for _, to := range config.ToAddresses {
		msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", to, subject, body)
		
		addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)
		if err := smtp.SendMail(addr, auth, config.FromAddress, []string{to}, []byte(msg)); err != nil {
			return fmt.Errorf("发送邮件失败: %v", err)
		}
	}
	
	return nil
}

func (c *EmailChannel) buildEmailBody(message *NotificationMessage) string {
	var body strings.Builder
	
	body.WriteString("<html><body>")
	body.WriteString(fmt.Sprintf("<h2>%s</h2>", message.Title))
	body.WriteString(fmt.Sprintf("<p><strong>级别:</strong> %s</p>", message.Level))
	body.WriteString(fmt.Sprintf("<p><strong>时间:</strong> %s</p>", message.CreatedAt.Format("2006-01-02 15:04:05")))
	body.WriteString(fmt.Sprintf("<p><strong>消息:</strong> %s</p>", message.Message))
	
	if len(message.Data) > 0 {
		body.WriteString("<h3>详细信息:</h3><ul>")
		for key, value := range message.Data {
			body.WriteString(fmt.Sprintf("<li><strong>%s:</strong> %v</li>", key, value))
		}
		body.WriteString("</ul>")
	}
	
	body.WriteString("</body></html>")
	return body.String()
}

// WebhookChannel Webhook通知渠道
type WebhookChannel struct {
	service *NotificationService
}

func (c *WebhookChannel) GetName() string {
	return "webhook"
}

func (c *WebhookChannel) IsEnabled() bool {
	return c.service.config.Webhook.URL != ""
}

func (c *WebhookChannel) Send(message *NotificationMessage) error {
	config := c.service.config.Webhook
	
	// 构建请求数据
	data := map[string]interface{}{
		"id":         message.ID.Hex(),
		"type":       message.Type,
		"title":      message.Title,
		"message":    message.Message,
		"level":      message.Level,
		"data":       message.Data,
		"created_at": message.CreatedAt,
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %v", err)
	}
	
	// 创建HTTP请求
	method := config.Method
	if method == "" {
		method = "POST"
	}
	
	req, err := http.NewRequest(method, config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}
	
	// 发送请求
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送Webhook请求失败: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Webhook请求失败，状态码: %d，响应: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// DingTalkChannel 钉钉通知渠道
type DingTalkChannel struct {
	service *NotificationService
}

func (c *DingTalkChannel) GetName() string {
	return "dingtalk"
}

func (c *DingTalkChannel) IsEnabled() bool {
	return c.service.config.DingTalk.WebhookURL != ""
}

func (c *DingTalkChannel) Send(message *NotificationMessage) error {
	config := c.service.config.DingTalk
	
	// 构建钉钉消息格式
	content := c.buildDingTalkMessage(message)
	
	data := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"title": message.Title,
			"text":  content,
		},
	}
	
	// 添加@信息
	if len(config.AtMobiles) > 0 || config.AtAll {
		data["at"] = map[string]interface{}{
			"atMobiles": config.AtMobiles,
			"isAtAll":   config.AtAll,
		}
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化钉钉消息失败: %v", err)
	}
	
	// 发送请求
	resp, err := http.Post(config.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送钉钉消息失败: %v", err)
	}
	defer resp.Body.Close()
	
	return nil
}

func (c *DingTalkChannel) buildDingTalkMessage(message *NotificationMessage) string {
	var content strings.Builder
	
	content.WriteString(fmt.Sprintf("## %s\n\n", message.Title))
	content.WriteString(fmt.Sprintf("**级别:** %s\n\n", message.Level))
	content.WriteString(fmt.Sprintf("**时间:** %s\n\n", message.CreatedAt.Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("**消息:** %s\n\n", message.Message))
	
	if url, exists := message.Data["url"]; exists {
		content.WriteString(fmt.Sprintf("**URL:** %s\n\n", url))
	}
	
	if similarity, exists := message.Data["similarity"]; exists {
		content.WriteString(fmt.Sprintf("**相似度:** %.2f\n\n", similarity))
	}
	
	return content.String()
}

// WeChatChannel 微信通知渠道
type WeChatChannel struct {
	service *NotificationService
}

func (c *WeChatChannel) GetName() string {
	return "wechat"
}

func (c *WeChatChannel) IsEnabled() bool {
	return c.service.config.WeChat.WebhookURL != ""
}

func (c *WeChatChannel) Send(message *NotificationMessage) error {
	config := c.service.config.WeChat
	
	// 构建微信消息格式
	content := c.buildWeChatMessage(message)
	
	data := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"content": content,
		},
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化微信消息失败: %v", err)
	}
	
	// 发送请求
	resp, err := http.Post(config.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送微信消息失败: %v", err)
	}
	defer resp.Body.Close()
	
	return nil
}

func (c *WeChatChannel) buildWeChatMessage(message *NotificationMessage) string {
	var content strings.Builder
	
	content.WriteString(fmt.Sprintf("## %s\n", message.Title))
	content.WriteString(fmt.Sprintf("> **级别:** %s\n", message.Level))
	content.WriteString(fmt.Sprintf("> **时间:** %s\n", message.CreatedAt.Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("> **消息:** %s\n", message.Message))
	
	return content.String()
}

// SlackChannel Slack通知渠道
type SlackChannel struct {
	service *NotificationService
}

func (c *SlackChannel) GetName() string {
	return "slack"
}

func (c *SlackChannel) IsEnabled() bool {
	return c.service.config.Slack.WebhookURL != ""
}

func (c *SlackChannel) Send(message *NotificationMessage) error {
	config := c.service.config.Slack
	
	// 构建Slack消息格式
	data := map[string]interface{}{
		"text":     message.Title,
		"channel":  config.Channel,
		"username": config.Username,
	}
	
	if config.IconEmoji != "" {
		data["icon_emoji"] = config.IconEmoji
	}
	
	// 添加富文本附件
	attachment := map[string]interface{}{
		"color":     c.getLevelColor(message.Level),
		"title":     message.Title,
		"text":      message.Message,
		"timestamp": message.CreatedAt.Unix(),
		"fields": []map[string]interface{}{
			{
				"title": "级别",
				"value": string(message.Level),
				"short": true,
			},
			{
				"title": "类型",
				"value": string(message.Type),
				"short": true,
			},
		},
	}
	
	data["attachments"] = []interface{}{attachment}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化Slack消息失败: %v", err)
	}
	
	// 发送请求
	resp, err := http.Post(config.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送Slack消息失败: %v", err)
	}
	defer resp.Body.Close()
	
	return nil
}

func (c *SlackChannel) getLevelColor(level NotificationLevel) string {
	switch level {
	case NotificationLevelCritical:
		return "danger"
	case NotificationLevelError:
		return "danger"
	case NotificationLevelWarning:
		return "warning"
	case NotificationLevelInfo:
		return "good"
	default:
		return "good"
	}
}