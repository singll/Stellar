package notification

import (
	"bytes"
	"context"
	"encoding/json"
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

// NotificationRequest 通知请求
type NotificationRequest struct {
	Type      string                 `json:"type"`      // email, webhook, sms
	To        string                 `json:"to"`        // 接收者
	Subject   string                 `json:"subject"`   // 主题
	Content   string                 `json:"content"`   // 内容
	Data      map[string]interface{} `json:"data"`      // 额外数据
	Template  string                 `json:"template"`  // 模板名称
}

// NotificationResponse 通知响应
type NotificationResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"messageId"`
	Error     string `json:"error,omitempty"`
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

// SendNotification 发送通知
func (ns *NotificationService) SendNotification(ctx context.Context, request *NotificationRequest) error {
	switch request.Type {
	case "email":
		return ns.sendEmailNotification(ctx, request)
	case "webhook":
		return ns.sendWebhookNotification(ctx, request)
	case "sms":
		return ns.sendSMSNotification(ctx, request)
	default:
		return fmt.Errorf("不支持的通知类型: %s", request.Type)
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
func (ns *NotificationService) sendEmailNotification(ctx context.Context, request *NotificationRequest) error {
	// 使用默认接收者（如果没有指定）
	to := request.To
	if to == "" {
		to = ns.emailConfig.DefaultTo
	}
	if to == "" {
		return fmt.Errorf("邮件接收者不能为空")
	}

	// 构建邮件内容
	from := fmt.Sprintf("%s <%s>", ns.emailConfig.FromName, ns.emailConfig.FromEmail)
	
	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = request.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"utf-8\""
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	// 构建邮件消息
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + ns.renderEmailTemplate(request)

	// 发送邮件
	auth := smtp.PlainAuth("", ns.emailConfig.Username, ns.emailConfig.Password, ns.emailConfig.SMTPHost)
	addr := fmt.Sprintf("%s:%d", ns.emailConfig.SMTPHost, ns.emailConfig.SMTPPort)
	
	err := smtp.SendMail(addr, auth, ns.emailConfig.FromEmail, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	log.Printf("邮件通知发送成功: %s -> %s", request.Subject, to)
	return nil
}

// sendWebhookNotification 发送Webhook通知
func (ns *NotificationService) sendWebhookNotification(ctx context.Context, request *NotificationRequest) error {
	// 使用默认URL（如果没有指定）
	url := request.To
	if url == "" {
		url = ns.webhookConfig.DefaultURL
	}
	if url == "" {
		return fmt.Errorf("Webhook URL不能为空")
	}

	// 构建Webhook载荷
	payload := map[string]interface{}{
		"type":      "page_change",
		"timestamp": time.Now().Unix(),
		"subject":   request.Subject,
		"content":   request.Content,
		"data":      request.Data,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("构建Webhook载荷失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
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

// sendSMSNotification 发送短信通知
func (ns *NotificationService) sendSMSNotification(ctx context.Context, request *NotificationRequest) error {
	// 简化的短信发送实现
	// 实际项目中需要集成具体的短信服务商API
	log.Printf("SMS通知发送: %s -> %s", request.Subject, request.To)
	
	// 这里只是示例，实际需要调用短信API
	if ns.smsConfig.APIKey == "" {
		return fmt.Errorf("短信服务未配置")
	}

	// TODO: 实现具体的短信发送逻辑
	return nil
}

// buildChangeNotificationContent 构建变化通知内容
func (ns *NotificationService) buildChangeNotificationContent(monitoring *models.PageMonitoring, change *models.PageChange) string {
	// 计算相似度百分比
	similarityPercent := fmt.Sprintf("%.1f%%", change.Similarity*100)
	
	// 获取变化状态描述
	var statusDesc string
	switch change.Status {
	case "new":
		statusDesc = "新增内容"
	case "changed":
		statusDesc = "内容变化"
	case "removed":
		statusDesc = "内容移除"
	default:
		statusDesc = "未知变化"
	}

	content := fmt.Sprintf(`
页面监控告警通知

监控名称: %s
监控URL: %s
变化时间: %s
变化类型: %s
相似度: %s

变化详情:
%s

请及时查看并处理相关变化。

--
此邮件由 Stellar 页面监控系统自动发送
`, 
		monitoring.Name,
		monitoring.URL,
		change.ChangedAt.Format("2006-01-02 15:04:05"),
		statusDesc,
		similarityPercent,
		ns.truncateString(change.Diff, 500),
	)

	return content
}

// renderEmailTemplate 渲染邮件模板
func (ns *NotificationService) renderEmailTemplate(request *NotificationRequest) string {
	// 简单的HTML邮件模板
	if request.Template == "page_change" {
		return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>%s</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .content { background-color: #ffffff; padding: 20px; border: 1px solid #e9ecef; border-radius: 5px; }
        .footer { margin-top: 20px; padding: 10px; text-align: center; color: #6c757d; font-size: 12px; }
        .alert { background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
        pre { background-color: #f8f9fa; padding: 10px; border-radius: 5px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>🔔 页面监控告警</h2>
        </div>
        <div class="content">
            <div class="alert">
                <strong>检测到页面变化！</strong>
            </div>
            <pre>%s</pre>
        </div>
        <div class="footer">
            此邮件由 Stellar 页面监控系统自动发送<br>
            请勿直接回复此邮件
        </div>
    </div>
</body>
</html>
`, request.Subject, request.Content)
	}

	// 默认纯文本格式
	return fmt.Sprintf(`
<html>
<body style="font-family: Arial, sans-serif;">
    <h3>%s</h3>
    <pre>%s</pre>
</body>
</html>
`, request.Subject, request.Content)
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

// truncateString 截断字符串
func (ns *NotificationService) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// TestNotification 测试通知发送
func (ns *NotificationService) TestNotification(ctx context.Context, method, target string) error {
	request := &NotificationRequest{
		Type:    method,
		To:      target,
		Subject: "Stellar 通知测试",
		Content: "这是一条测试通知，如果您收到此消息，说明通知系统配置正确。",
		Data: map[string]interface{}{
			"test": true,
			"timestamp": time.Now().Unix(),
		},
	}

	return ns.SendNotification(ctx, request)
}