package notification

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/StellarServer/internal/models"
)

// TemplateManager 模板管理器
type TemplateManager struct {
	templates map[string]*template.Template
}

// NewTemplateManager 创建模板管理器
func NewTemplateManager() *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[string]*template.Template),
	}
	tm.loadDefaultTemplates()
	return tm
}

// loadDefaultTemplates 加载默认模板
func (tm *TemplateManager) loadDefaultTemplates() {
	// 页面变化通知邮件模板
	pageChangeEmailTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>页面监控告警 - {{.Monitoring.Name}}</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
            line-height: 1.6; 
            color: #333; 
            margin: 0; 
            padding: 0; 
            background-color: #f5f5f5;
        }
        .container { 
            max-width: 600px; 
            margin: 20px auto; 
            background-color: #ffffff; 
            border-radius: 8px; 
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header { 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white; 
            padding: 30px 20px; 
            text-align: center; 
        }
        .header h1 { 
            margin: 0; 
            font-size: 24px; 
            font-weight: 600; 
        }
        .content { 
            padding: 30px 20px; 
        }
        .alert { 
            background-color: #fff3cd; 
            border-left: 4px solid #ffc107; 
            padding: 15px 20px; 
            margin-bottom: 25px; 
            border-radius: 4px;
        }
        .alert.danger { 
            background-color: #f8d7da; 
            border-left-color: #dc3545; 
        }
        .alert.success { 
            background-color: #d4edda; 
            border-left-color: #28a745; 
        }
        .info-card {
            background-color: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 6px;
            padding: 20px;
            margin-bottom: 20px;
        }
        .info-row {
            display: flex;
            justify-content: space-between;
            margin-bottom: 10px;
            border-bottom: 1px solid #e9ecef;
            padding-bottom: 8px;
        }
        .info-row:last-child {
            border-bottom: none;
            margin-bottom: 0;
            padding-bottom: 0;
        }
        .label { 
            font-weight: 600; 
            color: #495057;
            flex-shrink: 0;
            width: 120px;
        }
        .value { 
            color: #212529;
            word-break: break-word;
        }
        .diff-content { 
            background-color: #f8f9fa; 
            border: 1px solid #e9ecef;
            padding: 15px; 
            border-radius: 6px; 
            font-family: 'Monaco', 'Courier New', monospace; 
            font-size: 13px;
            white-space: pre-wrap;
            max-height: 300px;
            overflow-y: auto;
        }
        .footer { 
            background-color: #f8f9fa;
            border-top: 1px solid #e9ecef;
            padding: 20px; 
            text-align: center; 
            color: #6c757d; 
            font-size: 14px; 
        }
        .status-badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 600;
            text-transform: uppercase;
        }
        .status-changed { background-color: #ffc107; color: #212529; }
        .status-new { background-color: #17a2b8; color: white; }
        .status-removed { background-color: #dc3545; color: white; }
        .btn {
            display: inline-block;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 12px 24px;
            text-decoration: none;
            border-radius: 6px;
            font-weight: 500;
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔔 页面监控告警</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">检测到页面内容发生变化</p>
        </div>
        
        <div class="content">
            <div class="alert {{if lt .Change.Similarity 0.5}}danger{{else if lt .Change.Similarity 0.8}}{{else}}success{{end}}">
                <strong>{{.Monitoring.Name}}</strong> 的页面内容发生了变化！
                相似度从上次检查的 <strong>{{printf "%.1f%%" (mul .Change.Similarity 100)}}</strong>
            </div>

            <div class="info-card">
                <div class="info-row">
                    <span class="label">监控名称:</span>
                    <span class="value">{{.Monitoring.Name}}</span>
                </div>
                <div class="info-row">
                    <span class="label">监控URL:</span>
                    <span class="value"><a href="{{.Monitoring.URL}}" target="_blank">{{.Monitoring.URL}}</a></span>
                </div>
                <div class="info-row">
                    <span class="label">变化时间:</span>
                    <span class="value">{{.Change.ChangedAt.Format "2006-01-02 15:04:05"}}</span>
                </div>
                <div class="info-row">
                    <span class="label">变化类型:</span>
                    <span class="value">
                        <span class="status-badge status-{{.Change.Status}}">
                            {{if eq .Change.Status "new"}}新增内容
                            {{else if eq .Change.Status "changed"}}内容变化
                            {{else if eq .Change.Status "removed"}}内容移除
                            {{else}}未知变化{{end}}
                        </span>
                    </span>
                </div>
                <div class="info-row">
                    <span class="label">相似度:</span>
                    <span class="value">{{printf "%.1f%%" (mul .Change.Similarity 100)}}</span>
                </div>
                <div class="info-row">
                    <span class="label">差异类型:</span>
                    <span class="value">{{.Change.DiffType}}</span>
                </div>
            </div>

            {{if .Change.Diff}}
            <h3>变化详情:</h3>
            <div class="diff-content">{{.Change.Diff}}</div>
            {{end}}

            <div style="text-align: center;">
                <a href="{{.DashboardURL}}/monitoring/{{.Monitoring.ID}}" class="btn" target="_blank">
                    查看详细信息
                </a>
            </div>
        </div>

        <div class="footer">
            此邮件由 <strong>Stellar 页面监控系统</strong> 自动发送<br>
            发送时间: {{.Timestamp.Format "2006-01-02 15:04:05"}} | 请勿直接回复此邮件
        </div>
    </div>
</body>
</html>`

	// 页面变化Webhook模板
	pageChangeWebhookTemplate := `{
  "type": "page_change_alert",
  "timestamp": {{.Timestamp.Unix}},
  "monitoring": {
    "id": "{{.Monitoring.ID}}",
    "name": "{{.Monitoring.Name}}",
    "url": "{{.Monitoring.URL}}"
  },
  "change": {
    "id": "{{.Change.ID}}",
    "status": "{{.Change.Status}}",
    "similarity": {{.Change.Similarity}},
    "changed_at": "{{.Change.ChangedAt.Format "2006-01-02T15:04:05Z07:00"}}",
    "diff_type": "{{.Change.DiffType}}"
  },
  "alert_level": "{{if lt .Change.Similarity 0.5}}high{{else if lt .Change.Similarity 0.8}}medium{{else}}low{{end}}",
  "message": "页面 {{.Monitoring.Name}} 检测到变化，相似度为 {{printf \"%.1f%%\" (mul .Change.Similarity 100)}}"
}`

	// 通知测试模板
	testNotificationTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Stellar 通知测试</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 20px; }
        .container { max-width: 500px; margin: 0 auto; padding: 20px; border: 1px solid #ddd; border-radius: 8px; }
        .header { text-align: center; background-color: #e8f5e8; padding: 20px; border-radius: 6px; margin-bottom: 20px; }
        .content { padding: 20px 0; }
        .footer { text-align: center; color: #666; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>✅ 通知测试成功</h2>
        </div>
        <div class="content">
            <p>恭喜！您的通知配置工作正常。</p>
            <p><strong>测试时间:</strong> {{.Timestamp.Format "2006-01-02 15:04:05"}}</p>
            <p><strong>通知类型:</strong> {{.Type}}</p>
            {{if .Target}}
            <p><strong>接收目标:</strong> {{.Target}}</p>
            {{end}}
        </div>
        <div class="footer">
            此邮件由 Stellar 页面监控系统发送
        </div>
    </div>
</body>
</html>`

	// 注册模板
	tm.RegisterTemplate("page_change_email", pageChangeEmailTemplate)
	tm.RegisterTemplate("page_change_webhook", pageChangeWebhookTemplate)
	tm.RegisterTemplate("test_notification", testNotificationTemplate)
}

// RegisterTemplate 注册模板
func (tm *TemplateManager) RegisterTemplate(name, content string) error {
	// 创建模板函数
	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}

	tmpl, err := template.New(name).Funcs(funcMap).Parse(content)
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	tm.templates[name] = tmpl
	return nil
}

// RenderTemplate 渲染模板
func (tm *TemplateManager) RenderTemplate(name string, data interface{}) (string, error) {
	tmpl, exists := tm.templates[name]
	if !exists {
		return "", fmt.Errorf("模板 %s 不存在", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("渲染模板失败: %v", err)
	}

	return buf.String(), nil
}

// GetAvailableTemplates 获取可用模板列表
func (tm *TemplateManager) GetAvailableTemplates() []string {
	var templates []string
	for name := range tm.templates {
		templates = append(templates, name)
	}
	return templates
}

// TemplateData 模板数据结构
type TemplateData struct {
	Monitoring   *models.PageMonitoring `json:"monitoring"`
	Change       *models.PageChange     `json:"change"`
	Timestamp    time.Time              `json:"timestamp"`
	DashboardURL string                 `json:"dashboard_url"`
	Type         string                 `json:"type"`
	Target       string                 `json:"target"`
}

// NewTemplateData 创建模板数据
func NewTemplateData(monitoring *models.PageMonitoring, change *models.PageChange) *TemplateData {
	return &TemplateData{
		Monitoring:   monitoring,
		Change:       change,
		Timestamp:    time.Now(),
		DashboardURL: "http://localhost:5173", // 应该从配置中获取
		Type:         "page_change",
	}
}

// NewTestTemplateData 创建测试模板数据
func NewTestTemplateData(notificationType, target string) *TemplateData {
	return &TemplateData{
		Timestamp: time.Now(),
		Type:      notificationType,
		Target:    target,
	}
}