package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// TemplateGenerator 插件模板生成器
type TemplateGenerator struct {
	templateDir string
	outputDir   string
}

// PluginTemplate 插件模板信息
type PluginTemplate struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Language    string            `json:"language"`
	Category    string            `json:"category"`
	Author      string            `json:"author"`
	Version     string            `json:"version"`
	Tags        []string          `json:"tags"`
	Config      map[string]string `json:"config"`
}

// NewTemplateGenerator 创建模板生成器
func NewTemplateGenerator(templateDir, outputDir string) *TemplateGenerator {
	return &TemplateGenerator{
		templateDir: templateDir,
		outputDir:   outputDir,
	}
}

// GeneratePlugin 生成插件
func (g *TemplateGenerator) GeneratePlugin(template *PluginTemplate) error {
	// 创建插件目录
	pluginDir := filepath.Join(g.outputDir, template.ID)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("创建插件目录失败: %v", err)
	}

	// 根据语言生成对应的文件
	switch template.Language {
	case "python":
		return g.generatePythonPlugin(pluginDir, template)
	case "javascript":
		return g.generateJavaScriptPlugin(pluginDir, template)
	case "yaml":
		return g.generateYAMLPlugin(pluginDir, template)
	case "go":
		return g.generateGoPlugin(pluginDir, template)
	default:
		return fmt.Errorf("不支持的插件语言: %s", template.Language)
	}
}

// generatePythonPlugin 生成Python插件
func (g *TemplateGenerator) generatePythonPlugin(pluginDir string, template *PluginTemplate) error {
	// 创建插件元数据文件
	if err := g.createMetadataFile(pluginDir, template); err != nil {
		return err
	}

	// 创建Python主文件
	pythonTemplate := `#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
{{.Name}} - {{.Description}}
Author: {{.Author}}
Version: {{.Version}}
"""

import json
import sys
import os
from datetime import datetime

# 插件参数（从环境变量获取）
PLUGIN_PARAMS = json.loads(os.environ.get('PLUGIN_PARAMS', '{}'))

# 全局结果对象
result = {
    "success": False,
    "data": None,
    "message": "",
    "timestamp": datetime.now().isoformat()
}

def set_result(success, data=None, message=""):
    """设置插件执行结果"""
    global result
    result["success"] = success
    result["data"] = data
    result["message"] = message

def log(message):
    """记录日志"""
    print(f"[LOG] {message}", file=sys.stderr)

def get_param(key, default=None):
    """获取插件参数"""
    return PLUGIN_PARAMS.get(key, default)

def main():
    """主函数 - 在这里实现你的插件逻辑"""
    try:
        log("{{.Name}} 插件开始执行")
        
        # TODO: 在这里实现你的插件逻辑
        # 示例：获取参数
        target = get_param("target", "")
        if not target:
            set_result(False, message="缺少必需的参数: target")
            return
        
        log(f"处理目标: {target}")
        
        # 示例：执行扫描逻辑
        scan_results = perform_scan(target)
        
        # 设置成功结果
        set_result(True, scan_results, "扫描完成")
        log("{{.Name}} 插件执行完成")
        
    except Exception as e:
        result["success"] = False
        result["error"] = str(e)
        result["message"] = f"执行异常: {e}"
        log(f"插件执行异常: {e}")

def perform_scan(target):
    """执行扫描逻辑"""
    # TODO: 在这里实现具体的扫描逻辑
    results = {
        "target": target,
        "status": "completed",
        "findings": [],
        "scan_time": datetime.now().isoformat()
    }
    
    # 示例：添加一些发现
    results["findings"].append({
        "type": "info",
        "title": "示例发现",
        "description": "这是一个示例发现",
        "severity": "low"
    })
    
    return results

if __name__ == "__main__":
    main()
    
    # 输出结果
    print("PLUGIN_RESULT_START")
    print(json.dumps(result, ensure_ascii=False, default=str))
    print("PLUGIN_RESULT_END")
`

	mainFile := filepath.Join(pluginDir, "main.py")
	if err := g.renderTemplate(pythonTemplate, template, mainFile); err != nil {
		return err
	}

	// 创建requirements.txt
	requirementsContent := `# Python依赖列表
# 请在这里添加你的插件依赖
requests>=2.25.0
`
	requirementsFile := filepath.Join(pluginDir, "requirements.txt")
	if err := os.WriteFile(requirementsFile, []byte(requirementsContent), 0644); err != nil {
		return fmt.Errorf("创建requirements.txt失败: %v", err)
	}

	// 创建README.md
	return g.createReadmeFile(pluginDir, template)
}

// generateJavaScriptPlugin 生成JavaScript插件
func (g *TemplateGenerator) generateJavaScriptPlugin(pluginDir string, template *PluginTemplate) error {
	// 创建插件元数据文件
	if err := g.createMetadataFile(pluginDir, template); err != nil {
		return err
	}

	// 创建JavaScript主文件
	jsTemplate := `#!/usr/bin/env node
/**
 * {{.Name}} - {{.Description}}
 * Author: {{.Author}}
 * Version: {{.Version}}
 */

// 插件参数（从环境变量获取）
const PLUGIN_PARAMS = JSON.parse(process.env.PLUGIN_PARAMS || '{}');

// 全局结果对象
const result = {
    success: false,
    data: null,
    message: "",
    timestamp: new Date().toISOString()
};

/**
 * 设置插件执行结果
 */
function setResult(success, data = null, message = "") {
    result.success = success;
    result.data = data;
    result.message = message;
}

/**
 * 记录日志
 */
function log(message) {
    console.error('[LOG]', message);
}

/**
 * 获取插件参数
 */
function getParam(key, defaultValue = null) {
    return PLUGIN_PARAMS[key] || defaultValue;
}

/**
 * 主函数 - 在这里实现你的插件逻辑
 */
async function main() {
    try {
        log('{{.Name}} 插件开始执行');
        
        // TODO: 在这里实现你的插件逻辑
        // 示例：获取参数
        const target = getParam('target', '');
        if (!target) {
            setResult(false, null, '缺少必需的参数: target');
            return;
        }
        
        log(` + "处理目标: ${target}" + `);
        
        // 示例：执行扫描逻辑
        const scanResults = await performScan(target);
        
        // 设置成功结果
        setResult(true, scanResults, '扫描完成');
        log('{{.Name}} 插件执行完成');
        
    } catch (error) {
        result.success = false;
        result.error = error.message;
        result.message = ` + "执行异常: ${error.message}" + `;
        log(` + "插件执行异常: ${error.message}" + `);
    }
}

/**
 * 执行扫描逻辑
 */
async function performScan(target) {
    // TODO: 在这里实现具体的扫描逻辑
    const results = {
        target: target,
        status: 'completed',
        findings: [],
        scan_time: new Date().toISOString()
    };
    
    // 示例：添加一些发现
    results.findings.push({
        type: 'info',
        title: '示例发现',
        description: '这是一个示例发现',
        severity: 'low'
    });
    
    return results;
}

// 执行主函数
(async function() {
    await main();
    
    // 输出结果
    console.log('PLUGIN_RESULT_START');
    console.log(JSON.stringify(result, null, 2));
    console.log('PLUGIN_RESULT_END');
})();
`

	mainFile := filepath.Join(pluginDir, "main.js")
	if err := g.renderTemplate(jsTemplate, template, mainFile); err != nil {
		return err
	}

	// 创建package.json
	packageTemplate := `{
  "name": "{{.ID}}",
  "version": "{{.Version}}",
  "description": "{{.Description}}",
  "main": "main.js",
  "author": "{{.Author}}",
  "license": "MIT",
  "dependencies": {
    "axios": "^0.24.0"
  },
  "scripts": {
    "start": "node main.js",
    "test": "echo \"Error: no test specified\" && exit 1"
  }
}
`

	packageFile := filepath.Join(pluginDir, "package.json")
	if err := g.renderTemplate(packageTemplate, template, packageFile); err != nil {
		return err
	}

	// 创建README.md
	return g.createReadmeFile(pluginDir, template)
}

// generateYAMLPlugin 生成YAML插件
func (g *TemplateGenerator) generateYAMLPlugin(pluginDir string, template *PluginTemplate) error {
	// 创建插件元数据文件
	if err := g.createMetadataFile(pluginDir, template); err != nil {
		return err
	}

	// 创建YAML插件文件
	yamlTemplate := `# {{.Name}} - {{.Description}}
# Author: {{.Author}}
# Version: {{.Version}}

id: {{.ID}}
name: {{.Name}}
version: {{.Version}}
author: {{.Author}}
description: {{.Description}}
type: scanner
category: {{.Category}}
tags:{{range .Tags}}
  - {{.}}{{end}}

# 插件配置参数
config:
  target:
    type: string
    required: true
    description: "扫描目标"
  timeout:
    type: integer
    default: 30
    description: "超时时间（秒）"
  user_agent:
    type: string
    default: "Stellar-Scanner/1.0"
    description: "User-Agent字符串"

# 脚本配置
script:
  language: python
  entry: main
  content: |
    #!/usr/bin/env python3
    # -*- coding: utf-8 -*-
    
    import requests
    import json
    from datetime import datetime
    
    def main():
        """主扫描函数"""
        try:
            # 获取配置参数
            target = PLUGIN_PARAMS.get('target', '')
            timeout = PLUGIN_PARAMS.get('timeout', 30)
            user_agent = PLUGIN_PARAMS.get('user_agent', 'Stellar-Scanner/1.0')
            
            if not target:
                set_result(False, message="缺少必需的参数: target")
                return
            
            log(f"开始扫描目标: {target}")
            
            # 执行HTTP请求
            headers = {'User-Agent': user_agent}
            response = requests.get(target, headers=headers, timeout=timeout)
            
            # 分析响应
            results = analyze_response(target, response)
            
            # 设置结果
            set_result(True, results, "扫描完成")
            log("扫描成功完成")
            
        except requests.RequestException as e:
            set_result(False, message=f"请求失败: {e}")
        except Exception as e:
            set_result(False, message=f"扫描异常: {e}")
    
    def analyze_response(target, response):
        """分析HTTP响应"""
        results = {
            "target": target,
            "status_code": response.status_code,
            "headers": dict(response.headers),
            "content_length": len(response.content),
            "findings": [],
            "scan_time": datetime.now().isoformat()
        }
        
        # 基础安全检查
        findings = []
        
        # 检查安全头
        security_headers = [
            'X-Frame-Options',
            'X-XSS-Protection', 
            'X-Content-Type-Options',
            'Strict-Transport-Security',
            'Content-Security-Policy'
        ]
        
        for header in security_headers:
            if header not in response.headers:
                findings.append({
                    "type": "security",
                    "title": f"缺少安全头: {header}",
                    "description": f"响应中缺少推荐的安全头 {header}",
                    "severity": "medium"
                })
        
        # 检查服务器信息泄露
        if 'Server' in response.headers:
            findings.append({
                "type": "info_disclosure",
                "title": "服务器信息泄露",
                "description": f"服务器头泄露了版本信息: {response.headers['Server']}",
                "severity": "low"
            })
        
        results["findings"] = findings
        return results

# 依赖列表
dependencies:
  - requests>=2.25.0

# 兼容性信息
compatibility:
  min_version: "0.6.0"
  platforms:
    - linux
    - windows
    - macos
`

	yamlFile := filepath.Join(pluginDir, "plugin.yaml")
	if err := g.renderTemplate(yamlTemplate, template, yamlFile); err != nil {
		return err
	}

	// 创建README.md
	return g.createReadmeFile(pluginDir, template)
}

// generateGoPlugin 生成Go插件
func (g *TemplateGenerator) generateGoPlugin(pluginDir string, template *PluginTemplate) error {
	// 创建插件元数据文件
	if err := g.createMetadataFile(pluginDir, template); err != nil {
		return err
	}

	// 创建Go插件文件
	goTemplate := `package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// PluginInfo 插件信息
type PluginInfo struct {
	ID          string   ` + "`json:\"id\"`" + `
	Name        string   ` + "`json:\"name\"`" + `
	Version     string   ` + "`json:\"version\"`" + `
	Author      string   ` + "`json:\"author\"`" + `
	Description string   ` + "`json:\"description\"`" + `
	Category    string   ` + "`json:\"category\"`" + `
	Tags        []string ` + "`json:\"tags\"`" + `
}

// ScanResult 扫描结果
type ScanResult struct {
	Target    string      ` + "`json:\"target\"`" + `
	Status    string      ` + "`json:\"status\"`" + `
	Findings  []Finding   ` + "`json:\"findings\"`" + `
	ScanTime  time.Time   ` + "`json:\"scan_time\"`" + `
}

// Finding 发现项
type Finding struct {
	Type        string ` + "`json:\"type\"`" + `
	Title       string ` + "`json:\"title\"`" + `
	Description string ` + "`json:\"description\"`" + `
	Severity    string ` + "`json:\"severity\"`" + `
}

// Result 插件执行结果
type Result struct {
	Success   bool        ` + "`json:\"success\"`" + `
	Data      interface{} ` + "`json:\"data\"`" + `
	Message   string      ` + "`json:\"message\"`" + `
	Error     string      ` + "`json:\"error,omitempty\"`" + `
	Timestamp time.Time   ` + "`json:\"timestamp\"`" + `
}

// Plugin 插件接口
type Plugin interface {
	Info() PluginInfo
	Execute(ctx context.Context, params map[string]interface{}) (*Result, error)
}

// {{.Name}} 插件实现
type {{.Name}}Plugin struct{}

// Info 返回插件信息
func (p *{{.Name}}Plugin) Info() PluginInfo {
	return PluginInfo{
		ID:          "{{.ID}}",
		Name:        "{{.Name}}",
		Version:     "{{.Version}}",
		Author:      "{{.Author}}",
		Description: "{{.Description}}",
		Category:    "{{.Category}}",
		Tags:        []string{{"{{range .Tags}}"{{.}}", {{end}}},
	}
}

// Execute 执行插件扫描
func (p *{{.Name}}Plugin) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	// 获取参数
	target, ok := params["target"].(string)
	if !ok || target == "" {
		return &Result{
			Success:   false,
			Message:   "缺少必需的参数: target",
			Timestamp: time.Now(),
		}, nil
	}

	log.Printf("开始扫描目标: %s", target)

	// 执行扫描
	scanResult, err := p.performScan(ctx, target, params)
	if err != nil {
		return &Result{
			Success:   false,
			Error:     err.Error(),
			Message:   "扫描失败",
			Timestamp: time.Now(),
		}, nil
	}

	return &Result{
		Success:   true,
		Data:      scanResult,
		Message:   "扫描完成",
		Timestamp: time.Now(),
	}, nil
}

// performScan 执行具体的扫描逻辑
func (p *{{.Name}}Plugin) performScan(ctx context.Context, target string, params map[string]interface{}) (*ScanResult, error) {
	// TODO: 在这里实现具体的扫描逻辑
	
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 发送HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 分析响应
	findings := p.analyzeResponse(resp)

	result := &ScanResult{
		Target:   target,
		Status:   "completed",
		Findings: findings,
		ScanTime: time.Now(),
	}

	return result, nil
}

// analyzeResponse 分析HTTP响应
func (p *{{.Name}}Plugin) analyzeResponse(resp *http.Response) []Finding {
	var findings []Finding

	// 检查安全头
	securityHeaders := []string{
		"X-Frame-Options",
		"X-XSS-Protection",
		"X-Content-Type-Options",
		"Strict-Transport-Security",
		"Content-Security-Policy",
	}

	for _, header := range securityHeaders {
		if resp.Header.Get(header) == "" {
			findings = append(findings, Finding{
				Type:        "security",
				Title:       fmt.Sprintf("缺少安全头: %s", header),
				Description: fmt.Sprintf("响应中缺少推荐的安全头 %s", header),
				Severity:    "medium",
			})
		}
	}

	// 检查服务器信息泄露
	if server := resp.Header.Get("Server"); server != "" {
		findings = append(findings, Finding{
			Type:        "info_disclosure",
			Title:       "服务器信息泄露",
			Description: fmt.Sprintf("服务器头泄露了版本信息: %s", server),
			Severity:    "low",
		})
	}

	return findings
}

func main() {
	// 从环境变量获取参数
	paramsStr := os.Getenv("PLUGIN_PARAMS")
	var params map[string]interface{}
	if paramsStr != "" {
		if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
			log.Fatalf("解析参数失败: %v", err)
		}
	}

	// 创建插件实例
	plugin := &{{.Name}}Plugin{}

	// 执行插件
	ctx := context.Background()
	result, err := plugin.Execute(ctx, params)
	if err != nil {
		log.Fatalf("插件执行失败: %v", err)
	}

	// 输出结果
	fmt.Println("PLUGIN_RESULT_START")
	if jsonData, err := json.MarshalIndent(result, "", "  "); err == nil {
		fmt.Println(string(jsonData))
	}
	fmt.Println("PLUGIN_RESULT_END")
}

// 导出插件实例（用于动态加载）
var Plugin {{.Name}}Plugin
`

	mainFile := filepath.Join(pluginDir, "main.go")
	if err := g.renderTemplate(goTemplate, template, mainFile); err != nil {
		return err
	}

	// 创建go.mod
	goModTemplate := `module {{.ID}}

go 1.19

require (
	// 在这里添加你的依赖
)
`

	goModFile := filepath.Join(pluginDir, "go.mod")
	if err := g.renderTemplate(goModTemplate, template, goModFile); err != nil {
		return err
	}

	// 创建README.md
	return g.createReadmeFile(pluginDir, template)
}

// createMetadataFile 创建插件元数据文件
func (g *TemplateGenerator) createMetadataFile(pluginDir string, template *PluginTemplate) error {
	metadata := map[string]interface{}{
		"id":          template.ID,
		"name":        template.Name,
		"version":     template.Version,
		"author":      template.Author,
		"description": template.Description,
		"language":    template.Language,
		"category":    template.Category,
		"tags":        template.Tags,
		"created_at":  time.Now().Format(time.RFC3339),
		"config":      template.Config,
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化元数据失败: %v", err)
	}

	metadataFile := filepath.Join(pluginDir, "metadata.json")
	if err := os.WriteFile(metadataFile, data, 0644); err != nil {
		return fmt.Errorf("创建元数据文件失败: %v", err)
	}

	return nil
}

// createReadmeFile 创建README文件
func (g *TemplateGenerator) createReadmeFile(pluginDir string, template *PluginTemplate) error {
	readmeTemplate := `# {{.Name}}

{{.Description}}

## 基本信息

- **作者**: {{.Author}}
- **版本**: {{.Version}}
- **语言**: {{.Language}}
- **分类**: {{.Category}}
- **标签**: {{range .Tags}}{{.}} {{end}}

## 安装

请按照Stellar插件系统的标准流程安装此插件。

## 使用方法

### 参数配置

| 参数名 | 类型 | 必需 | 默认值 | 描述 |
|--------|------|------|--------|------|
| target | string | 是 | - | 扫描目标 |

### 使用示例

` + "```json" + `
{
  "target": "https://example.com"
}
` + "```" + `

## 开发

### 开发环境要求

{{if eq .Language "python"}}
- Python 3.7+
- pip

### 安装依赖

` + "```bash" + `
pip install -r requirements.txt
` + "```" + `
{{else if eq .Language "javascript"}}
- Node.js 14+
- npm

### 安装依赖

` + "```bash" + `
npm install
` + "```" + `
{{else if eq .Language "go"}}
- Go 1.19+

### 安装依赖

` + "```bash" + `
go mod tidy
` + "```" + `
{{end}}

### 测试

请使用Stellar提供的插件测试工具进行测试。

## 许可证

MIT License

## 更新日志

### v{{.Version}}
- 初始版本
`

	readmeFile := filepath.Join(pluginDir, "README.md")
	return g.renderTemplate(readmeTemplate, template, readmeFile)
}

// renderTemplate 渲染模板
func (g *TemplateGenerator) renderTemplate(templateStr string, data interface{}, outputFile string) error {
	tmpl, err := template.New("plugin").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("渲染模板失败: %v", err)
	}

	return nil
}

// GetAvailableTemplates 获取可用的模板
func (g *TemplateGenerator) GetAvailableTemplates() []string {
	return []string{
		"python-scanner",
		"javascript-scanner", 
		"yaml-scanner",
		"go-scanner",
		"python-crawler",
		"javascript-crawler",
		"python-exploit",
		"javascript-exploit",
	}
}

// CreateFromTemplate 从预设模板创建插件
func (g *TemplateGenerator) CreateFromTemplate(templateName, pluginID, pluginName, author string) error {
	var template *PluginTemplate

	switch templateName {
	case "python-scanner":
		template = &PluginTemplate{
			ID:          pluginID,
			Name:        pluginName,
			Description: "基于Python的Web安全扫描器",
			Language:    "python",
			Category:    "scanner",
			Author:      author,
			Version:     "1.0.0",
			Tags:        []string{"scanner", "web", "security"},
			Config:      make(map[string]string),
		}
	case "javascript-scanner":
		template = &PluginTemplate{
			ID:          pluginID,
			Name:        pluginName,
			Description: "基于JavaScript的Web安全扫描器",
			Language:    "javascript", 
			Category:    "scanner",
			Author:      author,
			Version:     "1.0.0",
			Tags:        []string{"scanner", "web", "security"},
			Config:      make(map[string]string),
		}
	case "yaml-scanner":
		template = &PluginTemplate{
			ID:          pluginID,
			Name:        pluginName,
			Description: "基于YAML配置的安全扫描器",
			Language:    "yaml",
			Category:    "scanner",
			Author:      author,
			Version:     "1.0.0",
			Tags:        []string{"scanner", "config", "yaml"},
			Config:      make(map[string]string),
		}
	case "go-scanner":
		template = &PluginTemplate{
			ID:          pluginID,
			Name:        pluginName,
			Description: "基于Go的高性能安全扫描器",
			Language:    "go",
			Category:    "scanner",
			Author:      author,
			Version:     "1.0.0",
			Tags:        []string{"scanner", "go", "performance"},
			Config:      make(map[string]string),
		}
	default:
		return fmt.Errorf("未知的模板: %s", templateName)
	}

	return g.GeneratePlugin(template)
}