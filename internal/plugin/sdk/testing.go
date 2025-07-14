package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// PluginTester 插件测试器
type PluginTester struct {
	workDir     string
	timeout     time.Duration
	enableDebug bool
	logFile     *os.File
}

// TestCase 测试用例
type TestCase struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Params      map[string]interface{} `json:"params"`
	Expected    TestExpectation        `json:"expected"`
	Timeout     time.Duration          `json:"timeout"`
}

// TestExpectation 测试期望
type TestExpectation struct {
	Success      bool                   `json:"success"`
	ErrorPattern string                 `json:"error_pattern,omitempty"`
	DataPattern  string                 `json:"data_pattern,omitempty"`
	MinTime      time.Duration          `json:"min_time,omitempty"`
	MaxTime      time.Duration          `json:"max_time,omitempty"`
	CustomCheck  func(interface{}) bool `json:"-"`
}

// TestResult 测试结果
type TestResult struct {
	TestCase    *TestCase     `json:"test_case"`
	Success     bool          `json:"success"`
	ActualTime  time.Duration `json:"actual_time"`
	Output      string        `json:"output"`
	Error       string        `json:"error,omitempty"`
	PluginData  interface{}   `json:"plugin_data,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
}

// TestSuite 测试套件
type TestSuite struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	TestCases   []*TestCase `json:"test_cases"`
}

// TestReport 测试报告
type TestReport struct {
	PluginPath     string        `json:"plugin_path"`
	TestSuite      *TestSuite    `json:"test_suite"`
	Results        []*TestResult `json:"results"`
	Summary        TestSummary   `json:"summary"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	TotalDuration  time.Duration `json:"total_duration"`
}

// TestSummary 测试摘要
type TestSummary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// NewPluginTester 创建插件测试器
func NewPluginTester(workDir string) *PluginTester {
	return &PluginTester{
		workDir:     workDir,
		timeout:     60 * time.Second,
		enableDebug: false,
	}
}

// SetTimeout 设置超时时间
func (t *PluginTester) SetTimeout(timeout time.Duration) {
	t.timeout = timeout
}

// EnableDebug 启用调试模式
func (t *PluginTester) EnableDebug(enable bool) {
	t.enableDebug = enable
}

// SetLogFile 设置日志文件
func (t *PluginTester) SetLogFile(logPath string) error {
	if t.logFile != nil {
		t.logFile.Close()
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	t.logFile = file
	return nil
}

// log 记录日志
func (t *PluginTester) log(format string, args ...interface{}) {
	message := fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
	
	if t.enableDebug {
		fmt.Print(message)
	}
	
	if t.logFile != nil {
		t.logFile.WriteString(message)
	}
}

// RunTests 运行测试套件
func (t *PluginTester) RunTests(pluginPath string, testSuite *TestSuite) (*TestReport, error) {
	t.log("开始测试插件: %s", pluginPath)
	
	report := &TestReport{
		PluginPath: pluginPath,
		TestSuite:  testSuite,
		Results:    make([]*TestResult, 0),
		StartTime:  time.Now(),
	}

	// 运行每个测试用例
	for _, testCase := range testSuite.TestCases {
		t.log("运行测试用例: %s", testCase.Name)
		
		result := t.runSingleTest(pluginPath, testCase)
		report.Results = append(report.Results, result)
		
		if result.Success {
			report.Summary.Passed++
			t.log("测试用例 %s: 通过", testCase.Name)
		} else {
			report.Summary.Failed++
			t.log("测试用例 %s: 失败 - %s", testCase.Name, result.Error)
		}
		
		report.Summary.Total++
	}

	report.EndTime = time.Now()
	report.TotalDuration = report.EndTime.Sub(report.StartTime)

	t.log("测试完成: %d/%d 通过", report.Summary.Passed, report.Summary.Total)
	return report, nil
}

// runSingleTest 运行单个测试用例
func (t *PluginTester) runSingleTest(pluginPath string, testCase *TestCase) *TestResult {
	result := &TestResult{
		TestCase:  testCase,
		Timestamp: time.Now(),
	}

	startTime := time.Now()
	
	// 执行插件
	output, pluginData, err := t.executePlugin(pluginPath, testCase.Params, testCase.Timeout)
	
	result.ActualTime = time.Since(startTime)
	result.Output = output
	result.PluginData = pluginData

	if err != nil {
		result.Error = err.Error()
		result.Success = false
		return result
	}

	// 验证测试期望
	result.Success = t.validateExpectation(testCase.Expected, result)
	
	return result
}

// executePlugin 执行插件
func (t *PluginTester) executePlugin(pluginPath string, params map[string]interface{}, timeout time.Duration) (string, interface{}, error) {
	// 检测插件类型
	pluginType := t.detectPluginType(pluginPath)
	
	// 准备参数
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return "", nil, fmt.Errorf("序列化参数失败: %v", err)
	}

	// 创建上下文
	if timeout <= 0 {
		timeout = t.timeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 根据插件类型执行
	var cmd *exec.Cmd
	switch pluginType {
	case "python":
		cmd = exec.CommandContext(ctx, "python3", pluginPath)
	case "javascript":
		cmd = exec.CommandContext(ctx, "node", pluginPath)
	case "go":
		// 先编译Go插件
		if err := t.compileGoPlugin(pluginPath); err != nil {
			return "", nil, fmt.Errorf("编译Go插件失败: %v", err)
		}
		execPath := strings.TrimSuffix(pluginPath, ".go")
		cmd = exec.CommandContext(ctx, execPath)
	case "yaml":
		return "", nil, fmt.Errorf("YAML插件需要通过插件引擎执行")
	default:
		return "", nil, fmt.Errorf("不支持的插件类型: %s", pluginType)
	}

	// 设置环境变量
	cmd.Env = append(os.Environ(),
		"PLUGIN_PARAMS="+string(paramsJSON),
		"PLUGIN_ENV=testing",
		"PLUGIN_LOG_LEVEL=DEBUG",
	)

	// 设置工作目录
	cmd.Dir = filepath.Dir(pluginPath)

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), nil, fmt.Errorf("插件执行失败: %v", err)
	}

	// 解析插件结果
	pluginData := t.parsePluginOutput(string(output))
	
	return string(output), pluginData, nil
}

// detectPluginType 检测插件类型
func (t *PluginTester) detectPluginType(pluginPath string) string {
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
		// 尝试通过shebang检测
		if content, err := os.ReadFile(pluginPath); err == nil {
			firstLine := strings.Split(string(content), "\n")[0]
			if strings.Contains(firstLine, "python") {
				return "python"
			}
			if strings.Contains(firstLine, "node") {
				return "javascript"
			}
		}
		return "unknown"
	}
}

// compileGoPlugin 编译Go插件
func (t *PluginTester) compileGoPlugin(pluginPath string) error {
	execPath := strings.TrimSuffix(pluginPath, ".go")
	cmd := exec.Command("go", "build", "-o", execPath, pluginPath)
	cmd.Dir = filepath.Dir(pluginPath)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("编译失败: %v\n%s", err, string(output))
	}
	
	return nil
}

// parsePluginOutput 解析插件输出
func (t *PluginTester) parsePluginOutput(output string) interface{} {
	// 查找插件结果标记
	startMarker := "PLUGIN_RESULT_START"
	endMarker := "PLUGIN_RESULT_END"
	
	startIdx := strings.Index(output, startMarker)
	endIdx := strings.Index(output, endMarker)
	
	if startIdx == -1 || endIdx == -1 {
		return nil
	}
	
	jsonStr := output[startIdx+len(startMarker):endIdx]
	jsonStr = strings.TrimSpace(jsonStr)
	
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.log("解析插件结果失败: %v", err)
		return nil
	}
	
	return result
}

// validateExpectation 验证测试期望
func (t *PluginTester) validateExpectation(expected TestExpectation, result *TestResult) bool {
	pluginResult, ok := result.PluginData.(map[string]interface{})
	if !ok {
		return !expected.Success // 如果无法解析结果，期望应该是失败的
	}

	// 检查成功状态
	success, _ := pluginResult["success"].(bool)
	if success != expected.Success {
		return false
	}

	// 检查错误模式
	if expected.ErrorPattern != "" && expected.Success == false {
		errorStr, _ := pluginResult["error"].(string)
		if !strings.Contains(errorStr, expected.ErrorPattern) {
			return false
		}
	}

	// 检查数据模式
	if expected.DataPattern != "" && expected.Success == true {
		dataStr := fmt.Sprintf("%v", pluginResult["data"])
		if !strings.Contains(dataStr, expected.DataPattern) {
			return false
		}
	}

	// 检查执行时间
	if expected.MinTime > 0 && result.ActualTime < expected.MinTime {
		return false
	}
	if expected.MaxTime > 0 && result.ActualTime > expected.MaxTime {
		return false
	}

	// 自定义检查
	if expected.CustomCheck != nil {
		return expected.CustomCheck(pluginResult["data"])
	}

	return true
}

// LoadTestSuite 加载测试套件
func (t *PluginTester) LoadTestSuite(testSuitePath string) (*TestSuite, error) {
	data, err := os.ReadFile(testSuitePath)
	if err != nil {
		return nil, fmt.Errorf("读取测试套件文件失败: %v", err)
	}

	var testSuite TestSuite
	if err := json.Unmarshal(data, &testSuite); err != nil {
		return nil, fmt.Errorf("解析测试套件失败: %v", err)
	}

	return &testSuite, nil
}

// SaveTestReport 保存测试报告
func (t *PluginTester) SaveTestReport(report *TestReport, outputPath string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化测试报告失败: %v", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("保存测试报告失败: %v", err)
	}

	return nil
}

// GenerateHTMLReport 生成HTML测试报告
func (t *PluginTester) GenerateHTMLReport(report *TestReport, outputPath string) error {
	htmlTemplate := `<!DOCTYPE html>
<html>
<head>
    <title>插件测试报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .summary { margin: 20px 0; }
        .test-case { border: 1px solid #ddd; margin: 10px 0; padding: 15px; border-radius: 5px; }
        .passed { border-left: 5px solid #4CAF50; }
        .failed { border-left: 5px solid #f44336; }
        .details { background: #f9f9f9; padding: 10px; margin-top: 10px; border-radius: 3px; }
        pre { background: #f5f5f5; padding: 10px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="header">
        <h1>插件测试报告</h1>
        <p><strong>插件路径:</strong> {{.PluginPath}}</p>
        <p><strong>测试套件:</strong> {{.TestSuite.Name}}</p>
        <p><strong>开始时间:</strong> {{.StartTime.Format "2006-01-02 15:04:05"}}</p>
        <p><strong>结束时间:</strong> {{.EndTime.Format "2006-01-02 15:04:05"}}</p>
        <p><strong>总耗时:</strong> {{.TotalDuration}}</p>
    </div>
    
    <div class="summary">
        <h2>测试摘要</h2>
        <p><strong>总计:</strong> {{.Summary.Total}}</p>
        <p><strong>通过:</strong> {{.Summary.Passed}}</p>
        <p><strong>失败:</strong> {{.Summary.Failed}}</p>
        <p><strong>跳过:</strong> {{.Summary.Skipped}}</p>
        <p><strong>成功率:</strong> {{printf "%.1f%%" (div (mul (float64 .Summary.Passed) 100.0) (float64 .Summary.Total))}}</p>
    </div>
    
    <div class="test-results">
        <h2>测试结果</h2>
        {{range .Results}}
        <div class="test-case {{if .Success}}passed{{else}}failed{{end}}">
            <h3>{{.TestCase.Name}}</h3>
            <p>{{.TestCase.Description}}</p>
            <p><strong>状态:</strong> {{if .Success}}通过{{else}}失败{{end}}</p>
            <p><strong>执行时间:</strong> {{.ActualTime}}</p>
            {{if .Error}}
            <p><strong>错误:</strong> {{.Error}}</p>
            {{end}}
            <div class="details">
                <h4>输出:</h4>
                <pre>{{.Output}}</pre>
            </div>
        </div>
        {{end}}
    </div>
</body>
</html>`

	// 这里简化处理，实际应该使用html/template
	htmlContent := strings.ReplaceAll(htmlTemplate, "{{.PluginPath}}", report.PluginPath)
	htmlContent = strings.ReplaceAll(htmlContent, "{{.TestSuite.Name}}", report.TestSuite.Name)
	
	if err := os.WriteFile(outputPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("保存HTML报告失败: %v", err)
	}

	return nil
}

// CreateDefaultTestSuite 创建默认测试套件
func (t *PluginTester) CreateDefaultTestSuite(pluginPath string) *TestSuite {
	return &TestSuite{
		Name:        fmt.Sprintf("默认测试套件 - %s", filepath.Base(pluginPath)),
		Description: "自动生成的默认测试用例",
		TestCases: []*TestCase{
			{
				Name:        "基础功能测试",
				Description: "测试插件基础功能",
				Params: map[string]interface{}{
					"target": "https://httpbin.org/get",
				},
				Expected: TestExpectation{
					Success: true,
				},
				Timeout: 30 * time.Second,
			},
			{
				Name:        "参数验证测试",
				Description: "测试插件参数验证",
				Params:      map[string]interface{}{},
				Expected: TestExpectation{
					Success:      false,
					ErrorPattern: "target",
				},
				Timeout: 10 * time.Second,
			},
			{
				Name:        "超时测试",
				Description: "测试插件超时处理",
				Params: map[string]interface{}{
					"target": "https://httpbin.org/delay/10",
				},
				Expected: TestExpectation{
					Success: false,
				},
				Timeout: 5 * time.Second,
			},
		},
	}
}

// Close 关闭测试器
func (t *PluginTester) Close() error {
	if t.logFile != nil {
		return t.logFile.Close()
	}
	return nil
}