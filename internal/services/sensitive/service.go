package sensitive

import (
	"context"
	"fmt"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO: 完善以下类型的实现
// TargetStats 目标统计信息
type TargetStats struct {
	// TODO: 添加字段实现
}

// TrendPoint 趋势点数据
type TrendPoint struct {
	// TODO: 添加字段实现
}

// ReportConfig 报告配置
type ReportConfig struct {
	// TODO: 添加字段实现
}

// Service 敏感信息检测服务
type Service struct {
	db                      *mongo.Database
	detector                *Detector
	engine                  *DetectionEngine
	ruleManager            *RuleManager
	reportGenerator        *ReportGenerator
	falsePositiveManager   *FalsePositiveManager
}

// NewService 创建敏感信息检测服务
func NewService(db *mongo.Database) *Service {
	// 创建检测引擎
	engine := NewDetectionEngine(db)
	
	// 创建检测器
	detector := NewDetector(db, engine.rules)
	
	// 创建其他组件
	ruleManager := NewRuleManager(db, engine)
	reportGenerator := NewReportGenerator(detector)
	falsePositiveManager := NewFalsePositiveManager(db, engine)
	
	service := &Service{
		db:                      db,
		detector:                detector,
		engine:                  engine,
		ruleManager:            ruleManager,
		reportGenerator:        reportGenerator,
		falsePositiveManager:   falsePositiveManager,
	}
	
	return service
}

// StartDetection 启动敏感信息检测
func (s *Service) StartDetection(ctx context.Context, req models.SensitiveDetectionRequest) (*models.SensitiveDetectionResult, error) {
	// 验证请求
	if err := s.validateRequest(req); err != nil {
		return nil, fmt.Errorf("请求验证失败: %v", err)
	}
	
	// 使用检测器启动检测
	return s.detector.Detect(ctx, req)
}

// ScanContent 扫描内容
func (s *Service) ScanContent(content, source string) (*DetectionResult, error) {
	return s.engine.ScanContent(content, source)
}

// ScanURL 扫描URL
func (s *Service) ScanURL(url string) (*DetectionResult, error) {
	return s.engine.ScanURL(url)
}

// ScanFile 扫描文件
func (s *Service) ScanFile(filePath string) (*DetectionResult, error) {
	return s.engine.ScanFile(filePath)
}

// ScanDirectory 扫描目录
func (s *Service) ScanDirectory(dirPath string, recursive bool) ([]*DetectionResult, error) {
	return s.engine.ScanDirectory(dirPath, recursive)
}

// GetResults 获取检测结果
func (s *Service) GetResults(projectID primitive.ObjectID, limit int) ([]*models.SensitiveDetectionResult, error) {
	return s.detector.GetResults(projectID, limit)
}

// GetResult 获取单个检测结果
func (s *Service) GetResult(resultID primitive.ObjectID) (*models.SensitiveDetectionResult, error) {
	return s.detector.GetResult(resultID)
}

// UpdateResult 更新检测结果
func (s *Service) UpdateResult(resultID primitive.ObjectID, updates map[string]interface{}) error {
	return s.detector.UpdateResult(resultID, updates)
}

// DeleteResult 删除检测结果
func (s *Service) DeleteResult(resultID primitive.ObjectID) error {
	return s.detector.DeleteResult(resultID)
}

// 规则管理相关方法
func (s *Service) CreateRuleSet(ruleset *RuleSet) error {
	return s.ruleManager.CreateRuleSet(ruleset)
}

func (s *Service) UpdateRuleSet(rulesetID string, updates map[string]interface{}) error {
	return s.ruleManager.UpdateRuleSet(rulesetID, updates)
}

func (s *Service) DeleteRuleSet(rulesetID string) error {
	return s.ruleManager.DeleteRuleSet(rulesetID)
}

func (s *Service) GetRuleSet(rulesetID string) (*RuleSet, error) {
	return s.ruleManager.GetRuleSet(rulesetID)
}

func (s *Service) ListRuleSets() []*RuleSet {
	return s.ruleManager.ListRuleSets()
}

func (s *Service) GetRuleStatistics() *RuleStatistics {
	return s.ruleManager.GetStatistics()
}

func (s *Service) ValidateRule(rule *DetectionRule) *RuleValidationResult {
	return s.ruleManager.ValidateRule(rule)
}

func (s *Service) ExportRuleSet(rulesetID string) ([]byte, error) {
	return s.ruleManager.ExportRuleSet(rulesetID)
}

func (s *Service) ImportRuleSet(data []byte) error {
	return s.ruleManager.ImportRuleSet(data)
}

func (s *Service) ImportRuleSetFromFile(filePath string) error {
	return s.ruleManager.ImportRuleSetFromFile(filePath)
}

func (s *Service) ExportRuleSetToFile(rulesetID, filePath string) error {
	return s.ruleManager.ExportRuleSetToFile(rulesetID, filePath)
}

func (s *Service) GetRuleTemplates() []*RuleTemplate {
	return s.ruleManager.GetRuleTemplates()
}

// 报告生成相关方法
func (s *Service) GenerateReport(result *models.SensitiveDetectionResult, req ReportRequest) (*ReportResult, error) {
	return s.reportGenerator.GenerateReport(result, req)
}

// 误报处理相关方法
func (s *Service) CreateWhitelistRule(rule *WhitelistRule) error {
	return s.falsePositiveManager.CreateWhitelistRule(rule)
}

func (s *Service) AddToReviewQueue(result *DetectionResult, matchIndex int, priority ReviewPriority) error {
	return s.falsePositiveManager.AddToReviewQueue(result, matchIndex, priority)
}

func (s *Service) ProcessReviewItem(itemID primitive.ObjectID, userID string, action ReviewAction, comments string) error {
	return s.falsePositiveManager.ProcessReviewItem(itemID, userID, action, comments)
}

func (s *Service) GetPendingReviews(limit int) ([]*ReviewItem, error) {
	return s.falsePositiveManager.GetPendingReviews(limit)
}

func (s *Service) GetSuggestions() []*AutoSuggestion {
	return s.falsePositiveManager.GetSuggestions()
}

func (s *Service) ApplySuggestion(suggestionID primitive.ObjectID, userID string) error {
	return s.falsePositiveManager.ApplySuggestion(suggestionID, userID)
}

// validateRequest 验证检测请求
func (s *Service) validateRequest(req models.SensitiveDetectionRequest) error {
	if req.Name == "" {
		return fmt.Errorf("检测名称不能为空")
	}
	
	if len(req.Targets) == 0 {
		return fmt.Errorf("检测目标不能为空")
	}
	
	if req.Config.Concurrency <= 0 {
		req.Config.Concurrency = 5 // 默认并发数
	}
	
	if req.Config.Timeout <= 0 {
		req.Config.Timeout = 30 // 默认超时时间30秒
	}
	
	return nil
}

// GetEngineStatus 获取引擎状态
func (s *Service) GetEngineStatus() map[string]interface{} {
	return map[string]interface{}{
		"rules_count":          len(s.engine.GetRules()),
		"rulesets_count":       len(s.ruleManager.ListRuleSets()),
		"whitelist_rules":      len(s.falsePositiveManager.whitelistRules),
		"pending_reviews":      len(s.falsePositiveManager.reviewQueue),
		"suggestions_count":    len(s.falsePositiveManager.GetSuggestions()),
		"last_updated":         time.Now(),
	}
}

// ReloadRules 重新加载规则
func (s *Service) ReloadRules() error {
	// 重新加载规则管理器中的规则
	s.ruleManager.reloadRules()
	
	// 重新加载误报管理器中的白名单规则
	s.falsePositiveManager.loadWhitelistRules()
	
	return nil
}

// GetDetectionHistory 获取检测历史
func (s *Service) GetDetectionHistory(projectID primitive.ObjectID, days int) ([]*models.SensitiveDetectionResult, error) {
	startTime := time.Now().AddDate(0, 0, -days)
	return s.detector.GetResultsByTimeRange(startTime, time.Now())
}

// GetDetectionStatistics 获取检测统计
func (s *Service) GetDetectionStatistics(projectID primitive.ObjectID) (*DetectionStatistics, error) {
	stats, err := s.detector.GetStatistics()
	if err != nil {
		return nil, err
	}
	
	// 转换为DetectionStatistics结构
	return &DetectionStatistics{
		TotalResults: stats["total_results"].(int64),
		ActiveRules:  stats["active_rules"].(int),
	}, nil
}

// DetectionStatistics 检测统计
type DetectionStatistics struct {
	TotalDetections    int64                     `json:"total_detections"`
	TotalResults       int64                     `json:"total_results"`
	ActiveRules        int                       `json:"active_rules"`
	TotalFindings      int64                     `json:"total_findings"`
	FindingsByCategory map[string]int64          `json:"findings_by_category"`
	FindingsBySeverity map[SeverityLevel]int64   `json:"findings_by_severity"`
	RecentActivity     []*ActivityPoint          `json:"recent_activity"`
	TopTargets         []*TargetStats            `json:"top_targets"`
	TrendData          []*TrendPoint             `json:"trend_data"`
}

// ActivityPoint 活动点
type ActivityPoint struct {
	Date     time.Time `json:"date"`
	Count    int64     `json:"count"`
	Findings int64     `json:"findings"`
}

// 批量操作方法
func (s *Service) BatchScanURLs(urls []string) ([]*DetectionResult, error) {
	var results []*DetectionResult
	
	for _, url := range urls {
		result, err := s.engine.ScanURL(url)
		if err != nil {
			// 记录错误但继续处理其他URL
			continue
		}
		results = append(results, result)
	}
	
	return results, nil
}

func (s *Service) BatchScanFiles(filePaths []string) ([]*DetectionResult, error) {
	var results []*DetectionResult
	
	for _, filePath := range filePaths {
		result, err := s.engine.ScanFile(filePath)
		if err != nil {
			// 记录错误但继续处理其他文件
			continue
		}
		results = append(results, result)
	}
	
	return results, nil
}

// 配置管理方法
func (s *Service) UpdateEngineConfig(config *DetectionConfig) {
	s.engine.LoadConfig(config)
}

func (s *Service) GetEngineConfig() *DetectionConfig {
	return s.engine.config
}

// 健康检查
func (s *Service) HealthCheck() error {
	// 检查数据库连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := s.db.Client().Ping(ctx, nil); err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}
	
	// 检查规则加载
	if len(s.engine.GetRules()) == 0 {
		return fmt.Errorf("没有加载任何检测规则")
	}
	
	return nil
}

// 清理资源
func (s *Service) Cleanup() error {
	// 这里可以添加清理逻辑，如关闭连接、清理缓存等
	return nil
}

// 兼容性方法，用于保持与现有API的兼容
func (s *Service) DetectSensitiveInfo(projectID string, req models.SensitiveDetectionRequest) (*models.SensitiveDetectionResult, error) {
	// 解析项目ID
	projID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return nil, fmt.Errorf("无效的项目ID: %v", err)
	}
	req.ProjectID = projID
	
	return s.StartDetection(context.Background(), req)
}

func (s *Service) GetDetectionResult(id string) (*models.SensitiveDetectionResult, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("无效的结果ID: %v", err)
	}
	return s.GetResult(objID)
}

func (s *Service) ListDetectionResults(projectID string, status string, limit int, skip int) ([]*models.SensitiveDetectionResult, error) {
	objID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return nil, fmt.Errorf("无效的项目ID: %v", err)
	}
	return s.GetResults(objID, limit)
}

func (s *Service) DeleteDetectionResult(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("无效的结果ID: %v", err)
	}
	return s.DeleteResult(objID)
}