package docs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// DocumentationGenerator 文档生成器
type DocumentationGenerator struct {
	outputDir string
	templates map[string]*template.Template
}

// PluginDocumentation 插件文档
type PluginDocumentation struct {
	Metadata     *PluginMetadata     `json:"metadata"`
	Parameters   []*Parameter        `json:"parameters"`
	Functions    []*Function         `json:"functions"`
	Examples     []*Example          `json:"examples"`
	Installation *InstallationGuide  `json:"installation"`
	Changelog    []*ChangelogEntry   `json:"changelog"`
	GeneratedAt  time.Time           `json:"generated_at"`
}

// PluginMetadata 插件元数据
type PluginMetadata struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Language    string   `json:"language"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	License     string   `json:"license"`
	Homepage    string   `json:"homepage"`
	Repository  string   `json:"repository"`
}

// Parameter 参数文档
type Parameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description"`
	Example     interface{} `json:"example,omitempty"`
	Validation  string      `json:"validation,omitempty"`
}

// Function 函数文档
type Function struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Parameters  []*Parameter `json:"parameters"`
	Returns     *ReturnValue `json:"returns"`
	Examples    []*Example   `json:"examples"`
}

// ReturnValue 返回值文档
type ReturnValue struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Example     string `json:"example,omitempty"`
}

// Example 示例
type Example struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Input       map[string]interface{} `json:"input"`
	Output      interface{}            `json:"output"`
	Code        string                 `json:"code,omitempty"`
}

// InstallationGuide 安装指南
type InstallationGuide struct {
	Prerequisites []string `json:"prerequisites"`
	Steps         []string `json:"steps"`
	Verification  []string `json:"verification"`
	Troubleshooting []TroubleshootingItem `json:"troubleshooting"`
}

// TroubleshootingItem 故障排除项
type TroubleshootingItem struct {
	Problem  string `json:"problem"`
	Solution string `json:"solution"`
}

// ChangelogEntry 更新日志条目
type ChangelogEntry struct {
	Version     string    `json:"version"`
	Date        time.Time `json:"date"`
	Type        string    `json:"type"` // added, changed, deprecated, removed, fixed, security
	Description string    `json:"description"`
}

// NewDocumentationGenerator 创建文档生成器
func NewDocumentationGenerator(outputDir string) *DocumentationGenerator {
	return &DocumentationGenerator{
		outputDir: outputDir,
		templates: make(map[string]*template.Template),
	}
}

// GenerateDocumentation 生成插件文档
func (g *DocumentationGenerator) GenerateDocumentation(pluginPath string) (*PluginDocumentation, error) {
	// 解析插件文件
	metadata, err := g.parsePluginMetadata(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("解析插件元数据失败: %v", err)
	}

	// 分析代码
	functions, err := g.analyzeCode(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("分析代码失败: %v", err)
	}

	// 提取参数信息
	parameters, err := g.extractParameters(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("提取参数失败: %v", err)
	}

	// 生成示例
	examples := g.generateExamples(metadata, parameters)

	// 创建安装指南
	installation := g.createInstallationGuide(metadata)

	// 解析更新日志
	changelog, err := g.parseChangelog(filepath.Dir(pluginPath))
	if err != nil {
		// 如果没有更新日志，创建默认的
		changelog = []*ChangelogEntry{
			{
				Version:     metadata.Version,
				Date:        time.Now(),
				Type:        "added",
				Description: "初始版本",
			},
		}
	}

	doc := &PluginDocumentation{
		Metadata:     metadata,
		Parameters:   parameters,
		Functions:    functions,
		Examples:     examples,
		Installation: installation,
		Changelog:    changelog,
		GeneratedAt:  time.Now(),
	}

	return doc, nil
}

// parsePluginMetadata 解析插件元数据
func (g *DocumentationGenerator) parsePluginMetadata(pluginPath string) (*PluginMetadata, error) {
	// 尝试从metadata.json读取
	metadataPath := filepath.Join(filepath.Dir(pluginPath), "metadata.json")
	if data, err := os.ReadFile(metadataPath); err == nil {
		var metadata PluginMetadata
		if err := json.Unmarshal(data, &metadata); err == nil {
			return &metadata, nil
		}
	}

	// 从代码注释中提取
	return g.extractMetadataFromCode(pluginPath)
}

// extractMetadataFromCode 从代码中提取元数据
func (g *DocumentationGenerator) extractMetadataFromCode(pluginPath string) (*PluginMetadata, error) {
	file, err := os.Open(pluginPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	metadata := &PluginMetadata{
		Language: g.detectLanguage(pluginPath),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// 解析注释中的元数据
		if g.isComment(line, metadata.Language) {
			content := g.extractCommentContent(line, metadata.Language)
			g.parseMetadataComment(content, metadata)
		}
		
		// 如果已经扫描了足够的行，可以停止
		if scanner.Text() != "" && !g.isComment(line, metadata.Language) {
			break
		}
	}

	// 设置默认值
	if metadata.ID == "" {
		metadata.ID = strings.TrimSuffix(filepath.Base(pluginPath), filepath.Ext(pluginPath))
	}
	if metadata.Name == "" {
		metadata.Name = metadata.ID
	}
	if metadata.Version == "" {
		metadata.Version = "1.0.0"
	}

	return metadata, scanner.Err()
}

// detectLanguage 检测编程语言
func (g *DocumentationGenerator) detectLanguage(pluginPath string) string {
	ext := filepath.Ext(pluginPath)
	switch ext {
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".go":
		return "go"
	case ".yaml", ".yml":
		return "yaml"
	default:
		return "unknown"
	}
}

// isComment 检查是否为注释行
func (g *DocumentationGenerator) isComment(line, language string) bool {
	switch language {
	case "python":
		return strings.HasPrefix(line, "#") || strings.HasPrefix(line, "\"\"\"") || strings.HasPrefix(line, "'''")
	case "javascript":
		return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*")
	case "go":
		return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*")
	default:
		return strings.HasPrefix(line, "#")
	}
}

// extractCommentContent 提取注释内容
func (g *DocumentationGenerator) extractCommentContent(line, language string) string {
	switch language {
	case "python":
		if strings.HasPrefix(line, "#") {
			return strings.TrimSpace(line[1:])
		}
	case "javascript", "go":
		if strings.HasPrefix(line, "//") {
			return strings.TrimSpace(line[2:])
		}
		if strings.HasPrefix(line, "*") {
			return strings.TrimSpace(line[1:])
		}
	}
	return line
}

// parseMetadataComment 解析元数据注释
func (g *DocumentationGenerator) parseMetadataComment(content string, metadata *PluginMetadata) {
	// 查找特定的元数据标记
	patterns := map[string]*string{
		`@name\s+(.+)`:        &metadata.Name,
		`@version\s+(.+)`:     &metadata.Version,
		`@author\s+(.+)`:      &metadata.Author,
		`@description\s+(.+)`: &metadata.Description,
		`@category\s+(.+)`:    &metadata.Category,
		`@license\s+(.+)`:     &metadata.License,
		`@homepage\s+(.+)`:    &metadata.Homepage,
		`@repository\s+(.+)`:  &metadata.Repository,
	}

	for pattern, field := range patterns {
		if re := regexp.MustCompile(pattern); re.MatchString(content) {
			matches := re.FindStringSubmatch(content)
			if len(matches) > 1 {
				*field = strings.TrimSpace(matches[1])
			}
		}
	}

	// 解析标签
	if re := regexp.MustCompile(`@tags?\s+(.+)`); re.MatchString(content) {
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			tags := strings.Split(matches[1], ",")
			for _, tag := range tags {
				metadata.Tags = append(metadata.Tags, strings.TrimSpace(tag))
			}
		}
	}
}

// analyzeCode 分析代码结构
func (g *DocumentationGenerator) analyzeCode(pluginPath string) ([]*Function, error) {
	language := g.detectLanguage(pluginPath)
	
	switch language {
	case "python":
		return g.analyzePythonCode(pluginPath)
	case "javascript":
		return g.analyzeJavaScriptCode(pluginPath)
	case "go":
		return g.analyzeGoCode(pluginPath)
	default:
		return []*Function{}, nil
	}
}

// analyzePythonCode 分析Python代码
func (g *DocumentationGenerator) analyzePythonCode(pluginPath string) ([]*Function, error) {
	file, err := os.Open(pluginPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var functions []*Function
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// 查找函数定义
		if re := regexp.MustCompile(`def\s+(\w+)\s*\(([^)]*)\):`); re.MatchString(line) {
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				function := &Function{
					Name:       matches[1],
					Parameters: g.parseParameters(matches[2], "python"),
				}
				
				// 读取文档字符串
				function.Description = g.extractPythonDocstring(scanner)
				functions = append(functions, function)
			}
		}
	}

	return functions, scanner.Err()
}

// extractPythonDocstring 提取Python文档字符串
func (g *DocumentationGenerator) extractPythonDocstring(scanner *bufio.Scanner) string {
	var docstring []string
	inDocstring := false
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if strings.HasPrefix(line, `"""`) || strings.HasPrefix(line, `'''`) {
			if !inDocstring {
				inDocstring = true
				// 如果文档字符串在同一行结束
				if strings.Count(line, `"""`) == 2 || strings.Count(line, `'''`) == 2 {
					return strings.Trim(line, `"'`)
				}
				continue
			} else {
				break
			}
		}
		
		if inDocstring {
			docstring = append(docstring, line)
		} else if line != "" {
			break
		}
	}
	
	return strings.Join(docstring, "\n")
}

// analyzeJavaScriptCode 分析JavaScript代码
func (g *DocumentationGenerator) analyzeJavaScriptCode(pluginPath string) ([]*Function, error) {
	// 简化实现，实际应该使用AST解析
	file, err := os.Open(pluginPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var functions []*Function
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// 查找函数定义
		patterns := []string{
			`function\s+(\w+)\s*\(([^)]*)\)`,
			`const\s+(\w+)\s*=\s*function\s*\(([^)]*)\)`,
			`(\w+)\s*:\s*function\s*\(([^)]*)\)`,
		}
		
		for _, pattern := range patterns {
			if re := regexp.MustCompile(pattern); re.MatchString(line) {
				matches := re.FindStringSubmatch(line)
				if len(matches) > 1 {
					function := &Function{
						Name:       matches[1],
						Parameters: g.parseParameters(matches[2], "javascript"),
					}
					functions = append(functions, function)
					break
				}
			}
		}
	}

	return functions, scanner.Err()
}

// analyzeGoCode 分析Go代码
func (g *DocumentationGenerator) analyzeGoCode(pluginPath string) ([]*Function, error) {
	// 简化实现，实际应该使用go/ast包
	file, err := os.Open(pluginPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var functions []*Function
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// 查找函数定义
		if re := regexp.MustCompile(`func\s+(\w+)\s*\(([^)]*)\)`); re.MatchString(line) {
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				function := &Function{
					Name:       matches[1],
					Parameters: g.parseParameters(matches[2], "go"),
				}
				functions = append(functions, function)
			}
		}
	}

	return functions, scanner.Err()
}

// parseParameters 解析参数
func (g *DocumentationGenerator) parseParameters(paramStr, language string) []*Parameter {
	if paramStr == "" {
		return []*Parameter{}
	}

	var parameters []*Parameter
	params := strings.Split(paramStr, ",")
	
	for _, param := range params {
		param = strings.TrimSpace(param)
		if param == "" {
			continue
		}

		parameter := &Parameter{
			Name: param,
			Type: "any",
		}

		// 根据语言解析参数类型
		switch language {
		case "python":
			if strings.Contains(param, ":") {
				parts := strings.Split(param, ":")
				parameter.Name = strings.TrimSpace(parts[0])
				parameter.Type = strings.TrimSpace(parts[1])
			}
		case "go":
			parts := strings.Fields(param)
			if len(parts) >= 2 {
				parameter.Name = parts[0]
				parameter.Type = parts[1]
			}
		}

		parameters = append(parameters, parameter)
	}

	return parameters
}

// extractParameters 提取插件参数
func (g *DocumentationGenerator) extractParameters(pluginPath string) ([]*Parameter, error) {
	// 从代码中查找参数使用
	file, err := os.Open(pluginPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	paramMap := make(map[string]*Parameter)
	scanner := bufio.NewScanner(file)
	
	// 查找参数访问模式
	patterns := []string{
		`get_param\(\s*["'](\w+)["']`,
		`getParam\(\s*["'](\w+)["']`,
		`PLUGIN_PARAMS\[["'](\w+)["']\]`,
		`params\[["'](\w+)["']\]`,
	}

	for scanner.Scan() {
		line := scanner.Text()
		
		for _, pattern := range patterns {
			if re := regexp.MustCompile(pattern); re.MatchString(line) {
				matches := re.FindAllStringSubmatch(line, -1)
				for _, match := range matches {
					if len(match) > 1 {
						paramName := match[1]
						if _, exists := paramMap[paramName]; !exists {
							paramMap[paramName] = &Parameter{
								Name:        paramName,
								Type:        "string",
								Required:    false,
								Description: fmt.Sprintf("插件参数: %s", paramName),
							}
						}
					}
				}
			}
		}
	}

	var parameters []*Parameter
	for _, param := range paramMap {
		parameters = append(parameters, param)
	}

	return parameters, scanner.Err()
}

// generateExamples 生成示例
func (g *DocumentationGenerator) generateExamples(metadata *PluginMetadata, parameters []*Parameter) []*Example {
	examples := []*Example{
		{
			Title:       "基础使用示例",
			Description: "展示插件的基本使用方法",
			Input:       make(map[string]interface{}),
			Output: map[string]interface{}{
				"success": true,
				"data":    "示例输出",
				"message": "执行成功",
			},
		},
	}

	// 为每个参数添加示例值
	for _, param := range parameters {
		switch param.Name {
		case "target":
			examples[0].Input["target"] = "https://example.com"
		case "timeout":
			examples[0].Input["timeout"] = 30
		case "user_agent":
			examples[0].Input["user_agent"] = "Stellar-Scanner/1.0"
		default:
			examples[0].Input[param.Name] = fmt.Sprintf("示例_%s", param.Name)
		}
	}

	return examples
}

// createInstallationGuide 创建安装指南
func (g *DocumentationGenerator) createInstallationGuide(metadata *PluginMetadata) *InstallationGuide {
	guide := &InstallationGuide{
		Prerequisites: []string{},
		Steps: []string{
			"下载插件文件",
			"将插件放置到Stellar插件目录",
			"在Stellar中安装插件",
			"配置插件参数",
		},
		Verification: []string{
			"检查插件是否在插件列表中显示",
			"运行测试任务验证功能",
		},
		Troubleshooting: []TroubleshootingItem{
			{
				Problem:  "插件无法加载",
				Solution: "检查插件文件是否完整，依赖是否安装",
			},
			{
				Problem:  "执行失败",
				Solution: "检查参数配置是否正确，查看日志获取详细错误信息",
			},
		},
	}

	// 根据语言添加特定的前置条件
	switch metadata.Language {
	case "python":
		guide.Prerequisites = append(guide.Prerequisites, "Python 3.7+")
	case "javascript":
		guide.Prerequisites = append(guide.Prerequisites, "Node.js 14+")
	case "go":
		guide.Prerequisites = append(guide.Prerequisites, "Go 1.19+")
	}

	return guide
}

// parseChangelog 解析更新日志
func (g *DocumentationGenerator) parseChangelog(pluginDir string) ([]*ChangelogEntry, error) {
	changelogPath := filepath.Join(pluginDir, "CHANGELOG.md")
	
	data, err := os.ReadFile(changelogPath)
	if err != nil {
		return nil, err
	}

	var entries []*ChangelogEntry
	lines := strings.Split(string(data), "\n")
	
	for _, line := range lines {
		// 简化的解析，实际应该支持更复杂的格式
		if re := regexp.MustCompile(`##\s+\[?(\d+\.\d+\.\d+)\]?`); re.MatchString(line) {
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				entry := &ChangelogEntry{
					Version: matches[1],
					Date:    time.Now(), // 实际应该从文档中解析
					Type:    "changed",
				}
				entries = append(entries, entry)
			}
		}
	}

	return entries, nil
}

// SaveAsMarkdown 保存为Markdown文档
func (g *DocumentationGenerator) SaveAsMarkdown(doc *PluginDocumentation, outputPath string) error {
	markdownTemplate := `# {{.Metadata.Name}}

{{.Metadata.Description}}

## 基本信息

- **ID**: {{.Metadata.ID}}
- **版本**: {{.Metadata.Version}}
- **作者**: {{.Metadata.Author}}
- **语言**: {{.Metadata.Language}}
- **分类**: {{.Metadata.Category}}
- **标签**: {{range .Metadata.Tags}}{{.}} {{end}}
{{if .Metadata.License}}- **许可证**: {{.Metadata.License}}{{end}}
{{if .Metadata.Homepage}}- **主页**: {{.Metadata.Homepage}}{{end}}
{{if .Metadata.Repository}}- **仓库**: {{.Metadata.Repository}}{{end}}

## 参数配置

{{range .Parameters}}### {{.Name}}

- **类型**: {{.Type}}
- **必需**: {{if .Required}}是{{else}}否{{end}}
{{if .Default}}- **默认值**: {{.Default}}{{end}}
- **描述**: {{.Description}}
{{if .Example}}- **示例**: {{.Example}}{{end}}

{{end}}

## 使用示例

{{range .Examples}}### {{.Title}}

{{.Description}}

**输入参数**:
` + "```json" + `
{{marshal .Input}}
` + "```" + `

**输出结果**:
` + "```json" + `
{{marshal .Output}}
` + "```" + `

{{end}}

## 安装指南

### 前置要求

{{range .Installation.Prerequisites}}- {{.}}
{{end}}

### 安装步骤

{{range $i, $step := .Installation.Steps}}{{add $i 1}}. {{$step}}
{{end}}

### 验证安装

{{range .Installation.Verification}}- {{.}}
{{end}}

### 故障排除

{{range .Installation.Troubleshooting}}**问题**: {{.Problem}}
**解决方案**: {{.Solution}}

{{end}}

## 更新日志

{{range .Changelog}}### v{{.Version}} - {{.Date.Format "2006-01-02"}}

- **{{.Type}}**: {{.Description}}

{{end}}

---

*文档生成时间: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}*
`

	// 这里简化处理，实际应该使用text/template
	content := strings.ReplaceAll(markdownTemplate, "{{.Metadata.Name}}", doc.Metadata.Name)
	content = strings.ReplaceAll(content, "{{.Metadata.Description}}", doc.Metadata.Description)
	
	return os.WriteFile(outputPath, []byte(content), 0644)
}

// SaveAsHTML 保存为HTML文档
func (g *DocumentationGenerator) SaveAsHTML(doc *PluginDocumentation, outputPath string) error {
	htmlTemplate := `<!DOCTYPE html>
<html>
<head>
    <title>{{.Metadata.Name}} - 插件文档</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif; margin: 40px; line-height: 1.6; }
        .header { border-bottom: 2px solid #eee; padding-bottom: 20px; margin-bottom: 30px; }
        .metadata { background: #f8f9fa; padding: 20px; border-radius: 8px; margin: 20px 0; }
        .section { margin: 30px 0; }
        .parameter { border-left: 4px solid #007bff; padding-left: 20px; margin: 15px 0; }
        .example { background: #f8f9fa; padding: 20px; border-radius: 8px; margin: 15px 0; }
        pre { background: #f1f3f4; padding: 15px; border-radius: 4px; overflow-x: auto; }
        .badge { background: #007bff; color: white; padding: 2px 8px; border-radius: 12px; font-size: 0.8em; }
        .required { background: #dc3545; }
        .optional { background: #6c757d; }
    </style>
</head>
<body>
    <div class="header">
        <h1>{{.Metadata.Name}}</h1>
        <p>{{.Metadata.Description}}</p>
    </div>
    
    <div class="metadata">
        <h2>基本信息</h2>
        <table>
            <tr><td><strong>ID</strong></td><td>{{.Metadata.ID}}</td></tr>
            <tr><td><strong>版本</strong></td><td>{{.Metadata.Version}}</td></tr>
            <tr><td><strong>作者</strong></td><td>{{.Metadata.Author}}</td></tr>
            <tr><td><strong>语言</strong></td><td>{{.Metadata.Language}}</td></tr>
            <tr><td><strong>分类</strong></td><td>{{.Metadata.Category}}</td></tr>
        </table>
    </div>
    
    <div class="section">
        <h2>参数配置</h2>
        {{range .Parameters}}
        <div class="parameter">
            <h3>{{.Name}} <span class="badge {{if .Required}}required{{else}}optional{{end}}">{{if .Required}}必需{{else}}可选{{end}}</span></h3>
            <p><strong>类型</strong>: {{.Type}}</p>
            <p><strong>描述</strong>: {{.Description}}</p>
        </div>
        {{end}}
    </div>
    
    <div class="section">
        <h2>使用示例</h2>
        {{range .Examples}}
        <div class="example">
            <h3>{{.Title}}</h3>
            <p>{{.Description}}</p>
            <h4>输入参数:</h4>
            <pre>{{marshal .Input}}</pre>
            <h4>输出结果:</h4>
            <pre>{{marshal .Output}}</pre>
        </div>
        {{end}}
    </div>
    
    <footer>
        <p><em>文档生成时间: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</em></p>
    </footer>
</body>
</html>`

	// 简化处理
	content := strings.ReplaceAll(htmlTemplate, "{{.Metadata.Name}}", doc.Metadata.Name)
	return os.WriteFile(outputPath, []byte(content), 0644)
}

// SaveAsJSON 保存为JSON文档
func (g *DocumentationGenerator) SaveAsJSON(doc *PluginDocumentation, outputPath string) error {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化文档失败: %v", err)
	}

	return os.WriteFile(outputPath, data, 0644)
}