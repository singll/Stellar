package notification

import (
	"fmt"
	"time"
)

// Config 通知系统配置
type Config struct {
	Email   EmailConfig   `yaml:"email" json:"email"`
	Webhook WebhookConfig `yaml:"webhook" json:"webhook"`
	SMS     SMSConfig     `yaml:"sms" json:"sms"`
	Enabled bool          `yaml:"enabled" json:"enabled"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		Email: EmailConfig{
			SMTPHost:  "smtp.gmail.com",
			SMTPPort:  587,
			EnableTLS: true,
			FromName:  "Stellar Monitor",
			Username:  "", // 需要配置
			Password:  "", // 需要配置
			FromEmail: "", // 需要配置
			DefaultTo: "", // 需要配置
		},
		Webhook: WebhookConfig{
			Timeout:    30 * time.Second,
			RetryCount: 3,
		},
		SMS: SMSConfig{
			ServiceURL: "",
			APIKey:     "",
			APISecret:  "",
		},
		Enabled: true,
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证邮件配置
	if c.Email.SMTPHost == "" {
		return fmt.Errorf("SMTP主机地址不能为空")
	}
	if c.Email.SMTPPort <= 0 {
		return fmt.Errorf("SMTP端口必须大于0")
	}

	// 验证Webhook配置
	if c.Webhook.Timeout <= 0 {
		c.Webhook.Timeout = 30 * time.Second
	}
	if c.Webhook.RetryCount <= 0 {
		c.Webhook.RetryCount = 3
	}

	return nil
}

// IsEmailEnabled 检查邮件通知是否启用
func (c *Config) IsEmailEnabled() bool {
	return c.Enabled && c.Email.Username != "" && c.Email.Password != "" && c.Email.FromEmail != ""
}

// IsWebhookEnabled 检查Webhook通知是否启用
func (c *Config) IsWebhookEnabled() bool {
	return c.Enabled && c.Webhook.DefaultURL != ""
}

// IsSMSEnabled 检查短信通知是否启用
func (c *Config) IsSMSEnabled() bool {
	return c.Enabled && c.SMS.APIKey != "" && c.SMS.ServiceURL != ""
}