package vulnscan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/StellarServer/internal/models"
	pkgerrors "github.com/StellarServer/internal/pkg/errors"
	"github.com/StellarServer/internal/pkg/logger"
	"gopkg.in/yaml.v3"
)

// YAMLPOCExecutor YAML POC执行器
type YAMLPOCExecutor struct {
	name           string
	supportedTypes []string
	httpClient     *http.Client
}

// YAMLPOCTemplate YAML POC模板结构
type YAMLPOCTemplate struct {
	Info      POCInfo           `yaml:"info"`
	Requests  []POCRequest      `yaml:"requests"`
	Variables map[string]string `yaml:"variables,omitempty"`
}

// POCInfo POC信息
type POCInfo struct {
	Name        string   `yaml:"name"`
	Author      string   `yaml:"author,omitempty"`
	Severity    string   `yaml:"severity"`
	Description string   `yaml:"description,omitempty"`
	Reference   []string `yaml:"reference,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
}

// POCRequest POC请求
type POCRequest struct {
	Method     string            `yaml:"method"`
	Path       string            `yaml:"path"`
	Headers    map[string]string `yaml:"headers,omitempty"`
	Body       string            `yaml:"body,omitempty"`
	Matchers   []POCMatcher      `yaml:"matchers"`
	Extractors []POCExtractor    `yaml:"extractors,omitempty"`
}

// POCMatcher POC匹配器
type POCMatcher struct {
	Type      string   `yaml:"type"` // status, word, regex, size
	Status    []int    `yaml:"status,omitempty"`
	Words     []string `yaml:"words,omitempty"`
	Regex     []string `yaml:"regex,omitempty"`
	Size      []int    `yaml:"size,omitempty"`
	Condition string   `yaml:"condition,omitempty"` // and, or
	Part      string   `yaml:"part,omitempty"`      // body, header, all
}

// POCExtractor POC提取器
type POCExtractor struct {
	Type  string   `yaml:"type"` // regex, xpath, json
	Name  string   `yaml:"name"`
	Part  string   `yaml:"part"` // body, header
	Group int      `yaml:"group,omitempty"`
	Regex []string `yaml:"regex,omitempty"`
	XPath []string `yaml:"xpath,omitempty"`
	JSON  []string `yaml:"json,omitempty"`
}

// NewYAMLPOCExecutor 创建YAML POC执行器
func NewYAMLPOCExecutor() *YAMLPOCExecutor {
	return &YAMLPOCExecutor{
		name:           "yaml",
		supportedTypes: []string{"yaml", "yml"},
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     30 * time.Second,
			},
		},
	}
}

// Execute 执行POC
func (e *YAMLPOCExecutor) Execute(ctx context.Context, poc *models.POC, target POCTarget) (*models.POCResult, error) {
	// 解析YAML模板
	template, err := e.parseTemplate(poc.Script)
	if err != nil {
		logger.Error("Execute parse template failed", map[string]interface{}{"pocID": poc.ID.Hex(), "error": err})
		return nil, pkgerrors.WrapError(err, pkgerrors.CodePluginError, "解析YAML模板失败", 500)
	}

	result := &models.POCResult{
		POCID:     poc.ID,
		Target:    target.URL,
		Success:   false,
		CreatedAt: time.Now(),
	}

	// 执行所有请求
	for i, request := range template.Requests {
		requestResult, err := e.executeRequest(ctx, request, target, template.Variables)
		if err != nil {
			logger.Error("Execute request failed", map[string]interface{}{"pocID": poc.ID.Hex(), "requestIndex": i, "error": err})
			result.Error = err.Error()
			continue
		}

		// 检查匹配器
		matched := e.checkMatchers(request.Matchers, requestResult)
		if matched {
			result.Success = true
			result.Output = template.Info.Description
			result.Response = requestResult.Response

			// 提取器提取数据
			if len(request.Extractors) > 0 {
				extracted := e.extractData(request.Extractors, requestResult)
				if len(extracted) > 0 {
					extractedData, _ := json.Marshal(extracted)
					result.Payload = string(extractedData)
				}
			}

			break // 找到漏洞就停止
		}

		// 记录第一个请求的详情
		if i == 0 {
			result.Request = requestResult.Request
			result.Response = requestResult.Response
		}
	}

	return result, nil
}

// RequestResult 请求结果
type RequestResult struct {
	Request    string
	Response   string
	StatusCode int
	Headers    map[string][]string
	Body       string
	Size       int
}

// executeRequest 执行单个请求
func (e *YAMLPOCExecutor) executeRequest(ctx context.Context, request POCRequest, target POCTarget, variables map[string]string) (*RequestResult, error) {
	// 构建URL
	url := target.URL
	if request.Path != "" {
		if strings.HasPrefix(request.Path, "/") {
			url = strings.TrimSuffix(target.URL, "/") + request.Path
		} else {
			url = strings.TrimSuffix(target.URL, "/") + "/" + request.Path
		}
	}

	// 替换变量
	url = e.replaceVariables(url, variables)
	body := e.replaceVariables(request.Body, variables)

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, request.Method, url, strings.NewReader(body))
	if err != nil {
		logger.Error("executeRequest create request failed", map[string]interface{}{"method": request.Method, "url": url, "error": err})
		return nil, pkgerrors.WrapError(err, pkgerrors.CodeNetworkError, "创建HTTP请求失败", 500)
	}

	// 设置请求头
	for key, value := range request.Headers {
		req.Header.Set(key, e.replaceVariables(value, variables))
	}

	// 执行请求
	resp, err := e.httpClient.Do(req)
	if err != nil {
		logger.Error("executeRequest do request failed", map[string]interface{}{"method": request.Method, "url": url, "error": err})
		return nil, pkgerrors.WrapError(err, pkgerrors.CodeNetworkError, "执行HTTP请求失败", 500)
	}
	defer resp.Body.Close()

	// 读取响应
	bodyBytes := make([]byte, 1024*1024) // 限制1MB
	n, _ := resp.Body.Read(bodyBytes)
	responseBody := string(bodyBytes[:n])

	return &RequestResult{
		Request:    fmt.Sprintf("%s %s", request.Method, url),
		Response:   responseBody,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       responseBody,
		Size:       len(responseBody),
	}, nil
}

// checkMatchers 检查匹配器
func (e *YAMLPOCExecutor) checkMatchers(matchers []POCMatcher, result *RequestResult) bool {
	if len(matchers) == 0 {
		return false
	}

	for _, matcher := range matchers {
		matched := false

		switch matcher.Type {
		case "status":
			for _, status := range matcher.Status {
				if result.StatusCode == status {
					matched = true
					break
				}
			}
		case "word":
			content := e.getMatchContent(result, matcher.Part)
			for _, word := range matcher.Words {
				if strings.Contains(content, word) {
					matched = true
					break
				}
			}
		case "regex":
			content := e.getMatchContent(result, matcher.Part)
			for _, pattern := range matcher.Regex {
				if matched, _ := regexp.MatchString(pattern, content); matched {
					matched = true
					break
				}
			}
		case "size":
			for _, size := range matcher.Size {
				if result.Size == size {
					matched = true
					break
				}
			}
		}

		// 根据条件决定是否继续
		if matcher.Condition == "or" && matched {
			return true
		}
		if matcher.Condition == "and" && !matched {
			return false
		}
		if matcher.Condition == "" && matched {
			return true
		}
	}

	return true // 默认所有匹配器都通过
}

// extractData 提取数据
func (e *YAMLPOCExecutor) extractData(extractors []POCExtractor, result *RequestResult) map[string]string {
	extracted := make(map[string]string)

	for _, extractor := range extractors {
		content := e.getMatchContent(result, extractor.Part)

		switch extractor.Type {
		case "regex":
			for _, pattern := range extractor.Regex {
				re, err := regexp.Compile(pattern)
				if err != nil {
					continue
				}
				matches := re.FindStringSubmatch(content)
				if len(matches) > extractor.Group {
					extracted[extractor.Name] = matches[extractor.Group]
				}
			}
			// TODO: 添加xpath和json提取器
		}
	}

	return extracted
}

// getMatchContent 获取匹配内容
func (e *YAMLPOCExecutor) getMatchContent(result *RequestResult, part string) string {
	switch part {
	case "header":
		headers := ""
		for key, values := range result.Headers {
			headers += key + ": " + strings.Join(values, ", ") + "\n"
		}
		return headers
	case "body":
		return result.Body
	default:
		return result.Response
	}
}

// replaceVariables 替换变量
func (e *YAMLPOCExecutor) replaceVariables(text string, variables map[string]string) string {
	for key, value := range variables {
		text = strings.ReplaceAll(text, "{{"+key+"}}", value)
	}
	return text
}

// parseTemplate 解析YAML模板
func (e *YAMLPOCExecutor) parseTemplate(script string) (*YAMLPOCTemplate, error) {
	var template YAMLPOCTemplate
	err := yaml.Unmarshal([]byte(script), &template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetSupportedTypes 获取支持的脚本类型
func (e *YAMLPOCExecutor) GetSupportedTypes() []string {
	return e.supportedTypes
}

// Validate 验证POC脚本
func (e *YAMLPOCExecutor) Validate(poc *models.POC) error {
	_, err := e.parseTemplate(poc.Script)
	return err
}

// GetName 获取执行器名称
func (e *YAMLPOCExecutor) GetName() string {
	return e.name
}
