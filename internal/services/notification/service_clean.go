package notification

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/StellarServer/internal/models"
)

// NotificationService 通知服务
type NotificationService struct {
	emailConfig     EmailConfig
	webhookConfig   WebhookConfig
	smsConfig       SMSConfig
	templateManager *TemplateManager
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	Username     string
	Password     string
	FromEmail    string
	FromName     string
	DefaultTo    string
	EnableTLS    bool
}

// WebhookConfig Webhook配置
type WebhookConfig struct {
	DefaultURL string
	Timeout    time.Duration
	RetryCount int
}

// SMSConfig 短信配置
type SMSConfig struct {
	APIKey      string
	APISecret   string
	ServiceURL  string
	DefaultFrom string
}

// NewNotificationService 创建通知服务
func NewNotificationService(emailConfig EmailConfig, webhookConfig WebhookConfig, smsConfig SMSConfig) *NotificationService {
	return &NotificationService{
		emailConfig:     emailConfig,
		webhookConfig:   webhookConfig,
		smsConfig:       smsConfig,
		templateManager: NewTemplateManager(),
	}
}

// SendPageChangeNotification 发送页面变化通知
func (ns *NotificationService) SendPageChangeNotification(ctx context.Context, monitoring *models.PageMonitoring, change *models.PageChange) error {
	if !monitoring.Config.NotifyOnChange {
		log.Printf("页面监控 %s 未启用变化通知", monitoring.Name)
		return nil
	}

	// 构建模板数据
	templateData := NewTemplateData(monitoring, change)
	
	// 发送通知到各个配置的渠道
	var errors []error
	for _, method := range monitoring.Config.NotifyMethods {
		// 获取接收者信息
		target := ns.getNotificationTarget(monitoring.Config.NotifyConfig, method)

		// 发送通知
		if err := ns.sendNotificationByMethod(ctx, method, target, templateData); err != nil {
			log.Printf("发送 %s 通知失败: %v", method, err)
			errors = append(errors, err)
		} else {
			log.Printf("成功发送 %s 通知到 %s", method, target)
		}
	}

	// 如果所有通知都失败，返回错误
	if len(errors) == len(monitoring.Config.NotifyMethods) && len(errors) > 0 {
		return fmt.Errorf("所有通知渠道都失败: %v", errors[0])
	}

	return nil
}

// sendNotificationByMethod 根据方法发送通知
func (ns *NotificationService) sendNotificationByMethod(ctx context.Context, method, target string, data *TemplateData) error {
	switch method {
	case "email":
		return ns.sendEmailNotificationWithTemplate(ctx, target, data)
	case "webhook":
		return ns.sendWebhookNotificationWithTemplate(ctx, target, data)
	case "sms":
		return ns.sendSMSNotificationWithTemplate(ctx, target, data)
	default:
		return fmt.Errorf("不支持的通知类型: %s", method)
	}
}

// sendEmailNotificationWithTemplate 使用模板发送邮件通知
func (ns *NotificationService) sendEmailNotificationWithTemplate(ctx context.Context, to string, data *TemplateData) error {
	// 使用默认接收者（如果没有指定）
	if to == "" {
		to = ns.emailConfig.DefaultTo
	}
	if to == "" {
		return fmt.Errorf("邮件接收者不能为空")
	}

	// 渲染邮件模板
	content, err := ns.templateManager.RenderTemplate("page_change_email", data)
	if err != nil {
		return fmt.Errorf("渲染邮件模板失败: %v", err)
	}

	// 构建邮件内容
	from := fmt.Sprintf("%s <%s>", ns.emailConfig.FromName, ns.emailConfig.FromEmail)
	subject := fmt.Sprintf("页面监控告警: %s 检测到变化", data.Monitoring.Name)
	
	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"utf-8\""
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	// 构建邮件消息
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + content

	// 发送邮件
	auth := smtp.PlainAuth("", ns.emailConfig.Username, ns.emailConfig.Password, ns.emailConfig.SMTPHost)
	addr := fmt.Sprintf("%s:%d", ns.emailConfig.SMTPHost, ns.emailConfig.SMTPPort)
	
	err = smtp.SendMail(addr, auth, ns.emailConfig.FromEmail, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	log.Printf("邮件通知发送成功: %s -> %s", subject, to)
	return nil
}

// sendWebhookNotificationWithTemplate 使用模板发送Webhook通知
func (ns *NotificationService) sendWebhookNotificationWithTemplate(ctx context.Context, url string, data *TemplateData) error {
	// 使用默认URL（如果没有指定）
	if url == "" {
		url = ns.webhookConfig.DefaultURL
	}
	if url == "" {
		return fmt.Errorf("Webhook URL不能为空")
	}

	// 渲染Webhook模板
	content, err := ns.templateManager.RenderTemplate("page_change_webhook", data)
	if err != nil {
		return fmt.Errorf("渲染Webhook模板失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(content))
	if err != nil {
		return fmt.Errorf("创建Webhook请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Stellar-Notification/1.0")

	// 发送请求
	client := &http.Client{
		Timeout: ns.webhookConfig.Timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送Webhook请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Webhook返回错误状态码: %d", resp.StatusCode)
	}

	log.Printf("Webhook通知发送成功: %s", url)
	return nil
}

// sendSMSNotificationWithTemplate 使用模板发送短信通知
func (ns *NotificationService) sendSMSNotificationWithTemplate(ctx context.Context, phone string, data *TemplateData) error {
	// 简化的短信发送实现
	message := fmt.Sprintf("页面监控告警: %s 检测到变化，相似度 %.1f%%。详情请查看控制台。", 
		data.Monitoring.Name, data.Change.Similarity*100)
	
	log.Printf("SMS通知发送: %s -> %s", message, phone)
	
	// TODO: 实现具体的短信发送逻辑
	if ns.smsConfig.APIKey == "" {
		return fmt.Errorf("短信服务未配置")
	}

	return nil
}

// getNotificationTarget 获取通知目标
func (ns *NotificationService) getNotificationTarget(config map[string]string, method string) string {
	switch method {
	case "email":
		if email, ok := config["email"]; ok {
			return email
		}
		return ns.emailConfig.DefaultTo
	case "webhook":
		if url, ok := config["webhook_url"]; ok {
			return url
		}
		return ns.webhookConfig.DefaultURL
	case "sms":
		if phone, ok := config["phone"]; ok {
			return phone
		}
		return ""
	}
	return ""
}

// TestNotification 测试通知发送
func (ns *NotificationService) TestNotification(ctx context.Context, method, target string) error {
	templateData := NewTestTemplateData(method, target)

	switch method {
	case "email":
		return ns.sendTestEmailNotification(ctx, target, templateData)
	case "webhook":
		return ns.sendTestWebhookNotification(ctx, target, templateData)
	case "sms":
		return ns.sendTestSMSNotification(ctx, target, templateData)
	default:
		return fmt.Errorf("不支持的通知类型: %s", method)
	}
}

// sendTestEmailNotification 发送测试邮件通知
func (ns *NotificationService) sendTestEmailNotification(ctx context.Context, to string, data *TemplateData) error {
	if to == "" {
		to = ns.emailConfig.DefaultTo
	}
	if to == "" {
		return fmt.Errorf("邮件接收者不能为空")
	}

	// 渲染测试邮件模板
	content, err := ns.templateManager.RenderTemplate("test_notification", data)
	if err != nil {
		return fmt.Errorf("渲染测试邮件模板失败: %v", err)
	}

	// 构建邮件内容
	from := fmt.Sprintf("%s <%s>", ns.emailConfig.FromName, ns.emailConfig.FromEmail)
	subject := "Stellar 通知测试"
	
	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"utf-8\""
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	// 构建邮件消息
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + content

	// 发送邮件
	auth := smtp.PlainAuth("", ns.emailConfig.Username, ns.emailConfig.Password, ns.emailConfig.SMTPHost)
	addr := fmt.Sprintf("%s:%d", ns.emailConfig.SMTPHost, ns.emailConfig.SMTPPort)
	
	err = smtp.SendMail(addr, auth, ns.emailConfig.FromEmail, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("发送测试邮件失败: %v", err)
	}

	log.Printf("测试邮件发送成功: %s", to)
	return nil
}

// sendTestWebhookNotification 发送测试Webhook通知
func (ns *NotificationService) sendTestWebhookNotification(ctx context.Context, url string, data *TemplateData) error {
	if url == "" {
		url = ns.webhookConfig.DefaultURL
	}
	if url == "" {
		return fmt.Errorf("Webhook URL不能为空")
	}

	// 构建测试载荷
	payload := map[string]interface{}{
		"type":      "test_notification",
		"timestamp": data.Timestamp.Unix(),
		"message":   "这是一条来自 Stellar 的测试通知",
		"data":      data,
	}

	// 创建HTTP请求
	content := fmt.Sprintf(`{"type": "test_notification", "timestamp": %d, "message": "这是一条来自 Stellar 的测试通知"}`, data.Timestamp.Unix())
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(content))
	if err != nil {
		return fmt.Errorf("创建测试Webhook请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Stellar-Notification/1.0")

	// 发送请求
	client := &http.Client{
		Timeout: ns.webhookConfig.Timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送测试Webhook请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("测试Webhook返回错误状态码: %d", resp.StatusCode)
	}

	log.Printf("测试Webhook发送成功: %s", url)
	return nil
}

// sendTestSMSNotification 发送测试短信通知
func (ns *NotificationService) sendTestSMSNotification(ctx context.Context, phone string, data *TemplateData) error {
	message := "这是一条来自 Stellar 页面监控系统的测试短信，如果您收到此消息，说明短信通知配置正确。"
	
	log.Printf("测试SMS通知发送: %s -> %s", message, phone)
	
	// TODO: 实现具体的短信发送逻辑
	if ns.smsConfig.APIKey == "" {
		return fmt.Errorf("短信服务未配置")
	}

	return nil
}