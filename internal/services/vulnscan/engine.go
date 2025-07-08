package vulnscan

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/time/rate"
)

// 添加一个适配器，将 VulnPlugin 转换为 Plugin
type vulnPluginAdapter struct {
	plugin VulnPlugin
}

// Info 实现 Plugin 接口的 Info 方法
func (a *vulnPluginAdapter) Info() PluginInfo {
	info := a.plugin.Info()
	return PluginInfo{
		ID:             info.ID.Hex(),
		Name:           info.Name,
		Description:    info.Description,
		Author:         info.Author,
		References:     info.References,
		CVEID:          info.CVEID,
		CWEID:          info.CWEID,
		Severity:       info.Severity,
		Type:           info.Type,
		Category:       info.Category,
		Tags:           info.Tags,
		RequiredParams: info.RequiredParams,
		DefaultParams:  info.DefaultParams,
	}
}

// Check 实现 Plugin 接口的 Check 方法
func (a *vulnPluginAdapter) Check(ctx context.Context, target Target, params map[string]string) (ScanResult, error) {
	vulnTarget := VulnTarget{
		URL:      target.URL,
		Host:     target.Host,
		Port:     target.Port,
		Protocol: target.Protocol,
		Path:     target.Path,
		Params:   params,
	}

	result, err := a.plugin.Scan(vulnTarget, map[string]interface{}{})
	if err != nil {
		return ScanResult{}, err
	}

	return ScanResult{
		Success:    result.Vulnerable,
		Output:     result.Details,
		Payload:    result.Payload,
		Request:    result.Request,
		Response:   result.Response,
		Screenshot: result.Screenshot,
	}, nil
}

// Init 实现 Plugin 接口的 Init 方法
func (a *vulnPluginAdapter) Init(config map[string]interface{}) error {
	// VulnPlugin 接口没有 Init 方法，这里简单返回 nil
	return nil
}

// Validate 实现 Plugin 接口的 Validate 方法
func (a *vulnPluginAdapter) Validate(params map[string]string) error {
	// VulnPlugin 接口没有 Validate 方法，这里简单返回 nil
	return nil
}

// 将 VulnPlugin 转换为 Plugin
func adaptVulnPlugin(plugin VulnPlugin) Plugin {
	return &vulnPluginAdapter{plugin: plugin}
}

// 将 VulnPlugin 切片转换为 Plugin 切片
func adaptVulnPlugins(plugins []VulnPlugin) []Plugin {
	result := make([]Plugin, len(plugins))
	for i, p := range plugins {
		result[i] = adaptVulnPlugin(p)
	}
	return result
}

// Engine 漏洞扫描引擎
type Engine struct {
	registry      PluginRegistry       // 插件注册表
	db            *mongo.Database      // 数据库
	limiter       *rate.Limiter        // 速率限制器
	resultHandler ResultHandler        // 结果处理器
	taskMap       map[string]*ScanTask // 任务映射
	taskMutex     sync.RWMutex         // 任务锁
}

// ScanTask 扫描任务
type ScanTask struct {
	ID              string                  // 任务ID
	Task            *models.VulnScanTask    // 任务信息
	Context         context.Context         // 上下文
	CancelFunc      context.CancelFunc      // 取消函数
	Progress        float64                 // 进度
	Status          string                  // 状态
	StartTime       time.Time               // 开始时间
	EndTime         time.Time               // 结束时间
	Plugins         []Plugin                // 使用的插件
	Results         []*models.POCResult     // 结果
	Vulnerabilities []*models.Vulnerability // 漏洞
	Error           error                   // 错误
	Mutex           sync.Mutex              // 锁
}

// ResultHandler 结果处理器接口
type ResultHandler interface {
	// HandlePOCResult 处理POC执行结果
	HandlePOCResult(result *models.POCResult) error

	// HandleVulnerability 处理漏洞
	HandleVulnerability(vuln *models.Vulnerability) error

	// UpdateTaskProgress 更新任务进度
	UpdateTaskProgress(taskID string, progress float64) error

	// UpdateTaskStatus 更新任务状态
	UpdateTaskStatus(taskID string, status string) error

	// FinishTask 完成任务
	FinishTask(task *models.VulnScanTask) error
}

// NewEngine 创建漏洞扫描引擎
func NewEngine(registry PluginRegistry, db *mongo.Database, resultHandler ResultHandler) *Engine {
	return &Engine{
		registry:      registry,
		db:            db,
		resultHandler: resultHandler,
		taskMap:       make(map[string]*ScanTask),
	}
}

// CreateTask 创建扫描任务
func (e *Engine) CreateTask(task *models.VulnScanTask) (string, error) {
	// 设置默认值
	if task.ID.IsZero() {
		task.ID = primitive.NewObjectID()
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	task.Status = "pending"
	task.Progress = 0

	// 初始化结果摘要
	task.ResultSummary = models.VulnScanSummary{
		TotalTargets: len(task.Targets),
		VulnTypes:    make(map[string]int),
	}

	// 保存任务到数据库
	_, err := e.db.Collection("vuln_scan_tasks").InsertOne(context.Background(), task)
	if err != nil {
		return "", err
	}

	return task.ID.Hex(), nil
}

// StartTask 启动扫描任务
func (e *Engine) StartTask(taskID string) error {
	// 检查任务是否已经在运行
	e.taskMutex.RLock()
	_, exists := e.taskMap[taskID]
	e.taskMutex.RUnlock()
	if exists {
		return errors.New("任务已在运行中")
	}

	// 从数据库获取任务
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return fmt.Errorf("无效的任务ID: %v", err)
	}

	var task models.VulnScanTask
	err = e.db.Collection("vuln_scan_tasks").FindOne(context.Background(), map[string]interface{}{
		"_id": objID,
	}).Decode(&task)
	if err != nil {
		return fmt.Errorf("获取任务失败: %v", err)
	}

	// 检查任务状态
	if task.Status == "running" || task.Status == "completed" {
		return fmt.Errorf("任务状态不允许启动: %s", task.Status)
	}

	// 更新任务状态
	task.Status = "running"
	task.StartedAt = time.Now()
	err = e.resultHandler.UpdateTaskStatus(taskID, "running")
	if err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	// 创建上下文和取消函数
	ctx, cancel := context.WithCancel(context.Background())

	// 加载插件
	var vulnPlugins []VulnPlugin
	if len(task.Config.POCIDs) > 0 {
		// 按ID加载插件
		for _, pocID := range task.Config.POCIDs {
			plugin, err := e.registry.GetPlugin(pocID)
			if err != nil {
				continue
			}
			vulnPlugins = append(vulnPlugins, plugin)
		}
	} else if len(task.Config.POCCategories) > 0 {
		// 按分类加载插件
		for _, category := range task.Config.POCCategories {
			categoryPlugins := e.registry.ListPluginsByCategory(category)
			vulnPlugins = append(vulnPlugins, categoryPlugins...)
		}
	} else {
		// 加载所有插件
		vulnPlugins = e.registry.ListPlugins()
	}

	// 转换为 Plugin 接口
	plugins := adaptVulnPlugins(vulnPlugins)

	// 过滤插件
	if task.Config.MinimumSeverity != "" {
		var filteredPlugins []Plugin
		for _, plugin := range plugins {
			info := plugin.Info()
			if severityLevel(info.Severity) >= severityLevel(task.Config.MinimumSeverity) {
				filteredPlugins = append(filteredPlugins, plugin)
			}
		}
		plugins = filteredPlugins
	}

	// 创建扫描任务
	scanTask := &ScanTask{
		ID:              taskID,
		Task:            &task,
		Context:         ctx,
		CancelFunc:      cancel,
		Progress:        0,
		Status:          "running",
		StartTime:       time.Now(),
		Plugins:         plugins,
		Results:         make([]*models.POCResult, 0),
		Vulnerabilities: make([]*models.Vulnerability, 0),
	}

	// 添加到任务映射
	e.taskMutex.Lock()
	e.taskMap[taskID] = scanTask
	e.taskMutex.Unlock()

	// 启动任务
	go e.runTask(scanTask)

	return nil
}

// StopTask 停止扫描任务
func (e *Engine) StopTask(taskID string) error {
	// 检查任务是否在运行
	e.taskMutex.RLock()
	task, exists := e.taskMap[taskID]
	e.taskMutex.RUnlock()
	if !exists {
		return errors.New("任务未在运行中")
	}

	// 取消任务
	task.CancelFunc()

	// 更新任务状态
	task.Status = "stopped"
	task.EndTime = time.Now()
	err := e.resultHandler.UpdateTaskStatus(taskID, "stopped")
	if err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	// 从任务映射中移除
	e.taskMutex.Lock()
	delete(e.taskMap, taskID)
	e.taskMutex.Unlock()

	return nil
}

// GetTaskStatus 获取任务状态
func (e *Engine) GetTaskStatus(taskID string) (string, error) {
	// 检查任务是否在运行
	e.taskMutex.RLock()
	task, exists := e.taskMap[taskID]
	e.taskMutex.RUnlock()
	if exists {
		return task.Status, nil
	}

	// 从数据库获取任务
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return "", fmt.Errorf("无效的任务ID: %v", err)
	}

	var dbTask models.VulnScanTask
	err = e.db.Collection("vuln_scan_tasks").FindOne(context.Background(), map[string]interface{}{
		"_id": objID,
	}).Decode(&dbTask)
	if err != nil {
		return "", fmt.Errorf("获取任务失败: %v", err)
	}

	return dbTask.Status, nil
}

// GetTaskProgress 获取任务进度
func (e *Engine) GetTaskProgress(taskID string) (float64, error) {
	// 检查任务是否在运行
	e.taskMutex.RLock()
	task, exists := e.taskMap[taskID]
	e.taskMutex.RUnlock()
	if exists {
		return task.Progress, nil
	}

	// 从数据库获取任务
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return 0, fmt.Errorf("无效的任务ID: %v", err)
	}

	var dbTask models.VulnScanTask
	err = e.db.Collection("vuln_scan_tasks").FindOne(context.Background(), map[string]interface{}{
		"_id": objID,
	}).Decode(&dbTask)
	if err != nil {
		return 0, fmt.Errorf("获取任务失败: %v", err)
	}

	return dbTask.Progress, nil
}

// runTask 执行扫描任务
func (e *Engine) runTask(task *ScanTask) {
	defer task.CancelFunc()

	// 创建速率限制器
	var limiter *rate.Limiter
	if task.Task.Config.RateLimit > 0 {
		limiter = rate.NewLimiter(rate.Limit(task.Task.Config.RateLimit), task.Task.Config.RateLimit)
	}

	// 创建工作池
	workerCount := task.Task.Config.Concurrency
	if workerCount <= 0 {
		workerCount = 10
	}

	// 创建工作通道
	type workItem struct {
		target Target
		plugin Plugin
	}
	workChan := make(chan workItem, workerCount*2)

	// 创建结果通道
	resultChan := make(chan ScanResult, workerCount*2)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range workChan {
				select {
				case <-task.Context.Done():
					return
				default:
					// 如果有速率限制，则等待令牌
					if limiter != nil {
						limiter.Wait(task.Context)
					}

					// 执行POC检测
					result, err := e.executePOC(task.Context, work.plugin, work.target, task.Task.Config)
					if err != nil {
						continue
					}

					// 发送结果
					select {
					case resultChan <- result:
					default:
						// 通道已满，丢弃结果
					}
				}
			}
		}()
	}

	// 启动结果处理协程
	resultsDone := make(chan struct{})
	go func() {
		defer close(resultsDone)
		for result := range resultChan {
			// 处理结果
			e.handleScanResult(task, result)
		}
	}()

	// 发送扫描任务
	for _, targetStr := range task.Task.Targets {
		// 解析目标
		target := parseTarget(targetStr, task.Task.TargetType)

		// 对每个目标应用所有插件
		for _, plugin := range task.Plugins {
			select {
			case <-task.Context.Done():
				// 任务被取消
				close(workChan)
				return
			default:
				// 发送工作
				workChan <- workItem{
					target: target,
					plugin: plugin,
				}
			}
		}
	}

	// 关闭工作通道
	close(workChan)

	// 等待所有工作完成
	wg.Wait()

	// 关闭结果通道
	close(resultChan)

	// 等待结果处理完成
	<-resultsDone

	// 检查是否被取消
	select {
	case <-task.Context.Done():
		// 任务被取消
		task.Status = "stopped"
		e.resultHandler.UpdateTaskStatus(task.ID, "stopped")
	default:
		// 任务完成
		task.Status = "completed"
		task.EndTime = time.Now()
		task.Progress = 100
		task.Task.Status = "completed"
		task.Task.CompletedAt = time.Now()
		task.Task.Progress = 100
		e.resultHandler.FinishTask(task.Task)
	}

	// 从任务映射中移除
	e.taskMutex.Lock()
	delete(e.taskMap, task.ID)
	e.taskMutex.Unlock()
}

// executePOC 执行POC检测
func (e *Engine) executePOC(ctx context.Context, plugin Plugin, target Target, config models.VulnScanConfig) (ScanResult, error) {
	// 设置超时上下文
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 执行检测
	startTime := time.Now()
	result, err := plugin.Check(timeoutCtx, target, plugin.Info().DefaultParams)
	executionTime := time.Since(startTime)
	result.ExecutionTime = executionTime

	return result, err
}

// handleScanResult 处理扫描结果
func (e *Engine) handleScanResult(task *ScanTask, result ScanResult) {
	// 更新任务进度
	task.Mutex.Lock()
	task.Progress += 1.0 / float64(len(task.Task.Targets)*len(task.Plugins)) * 100
	progress := task.Progress
	task.Mutex.Unlock()

	// 更新数据库中的进度
	e.resultHandler.UpdateTaskProgress(task.ID, progress)

	// 如果没有发现漏洞，直接返回
	if !result.Success || result.Vulnerability == nil {
		return
	}

	// 创建POC结果
	pocResult := &models.POCResult{
		ID:            primitive.NewObjectID(),
		TaskID:        task.Task.ID,
		Target:        result.Vulnerability.AffectedURL,
		Success:       result.Success,
		Output:        result.Output,
		Error:         result.Error,
		ExecutionTime: result.ExecutionTime.Milliseconds(),
		CreatedAt:     time.Now(),
		Request:       result.Request,
		Response:      result.Response,
		Payload:       result.Payload,
		Screenshot:    result.Screenshot,
	}

	// 保存POC结果
	err := e.resultHandler.HandlePOCResult(pocResult)
	if err != nil {
		return
	}

	// 设置漏洞关联
	result.Vulnerability.TaskID = task.Task.ID
	result.Vulnerability.ProjectID = task.Task.ProjectID
	result.Vulnerability.DiscoveredAt = time.Now()
	result.Vulnerability.CreatedAt = time.Now()
	result.Vulnerability.UpdatedAt = time.Now()
	result.Vulnerability.Status = models.StatusUnverified

	// 保存漏洞
	err = e.resultHandler.HandleVulnerability(result.Vulnerability)
	if err != nil {
		return
	}

	// 添加到任务结果
	task.Mutex.Lock()
	task.Results = append(task.Results, pocResult)
	task.Vulnerabilities = append(task.Vulnerabilities, result.Vulnerability)
	task.Mutex.Unlock()

	// 更新任务摘要
	updateTaskSummary(task, result.Vulnerability)
}

// updateTaskSummary 更新任务摘要
func updateTaskSummary(task *ScanTask, vuln *models.Vulnerability) {
	task.Mutex.Lock()
	defer task.Mutex.Unlock()

	// 更新漏洞总数
	task.Task.ResultSummary.TotalVulns++

	// 更新不同严重性级别的漏洞数量
	switch vuln.Severity {
	case models.SeverityCritical:
		task.Task.ResultSummary.CriticalVulns++
	case models.SeverityHigh:
		task.Task.ResultSummary.HighVulns++
	case models.SeverityMedium:
		task.Task.ResultSummary.MediumVulns++
	case models.SeverityLow:
		task.Task.ResultSummary.LowVulns++
	case models.SeverityInfo:
		task.Task.ResultSummary.InfoVulns++
	}

	// 更新漏洞类型统计
	typeStr := string(vuln.Type)
	if count, ok := task.Task.ResultSummary.VulnTypes[typeStr]; ok {
		task.Task.ResultSummary.VulnTypes[typeStr] = count + 1
	} else {
		task.Task.ResultSummary.VulnTypes[typeStr] = 1
	}
}

// parseTarget 解析目标
func parseTarget(targetStr string, targetType string) Target {
	// 简单实现，实际应用中需要更复杂的解析逻辑
	target := Target{
		URL:  targetStr,
		Host: targetStr,
	}

	return target
}

// severityLevel 获取严重性级别的数值
func severityLevel(severity models.VulnerabilitySeverity) int {
	switch severity {
	case models.SeverityCritical:
		return 5
	case models.SeverityHigh:
		return 4
	case models.SeverityMedium:
		return 3
	case models.SeverityLow:
		return 2
	case models.SeverityInfo:
		return 1
	default:
		return 0
	}
}
