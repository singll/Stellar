package sensitive

import (
	"context"
	"fmt"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Service 敏感信息检测服务
type Service struct {
	db             *mongo.Database
	rulesColl      *mongo.Collection
	ruleGroupsColl *mongo.Collection
	whitelistColl  *mongo.Collection
	resultsColl    *mongo.Collection
	ctx            context.Context
}

// NewService 创建敏感信息检测服务
func NewService(db *mongo.Database) *Service {
	return &Service{
		db:             db,
		rulesColl:      db.Collection("sensitiveRules"),
		ruleGroupsColl: db.Collection("sensitiveRuleGroups"),
		whitelistColl:  db.Collection("sensitiveWhitelists"),
		resultsColl:    db.Collection("sensitiveResults"),
		ctx:            context.Background(),
	}
}

// CreateRule 创建敏感规则
func (s *Service) CreateRule(req models.SensitiveRuleCreateRequest) (*models.SensitiveRule, error) {
	rule := &models.SensitiveRule{
		ID:                    primitive.NewObjectID(),
		Name:                  req.Name,
		Description:           req.Description,
		Type:                  req.Type,
		Pattern:               req.Pattern,
		Category:              req.Category,
		RiskLevel:             req.RiskLevel,
		Tags:                  req.Tags,
		Enabled:               req.Enabled,
		Context:               req.Context,
		Examples:              req.Examples,
		FalsePositivePatterns: req.FalsePositivePatterns,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	_, err := s.rulesColl.InsertOne(s.ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("创建敏感规则失败: %v", err)
	}

	return rule, nil
}

// UpdateRule 更新敏感规则
func (s *Service) UpdateRule(id string, req models.SensitiveRuleUpdateRequest) (*models.SensitiveRule, error) {
	// 解析ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("无效的规则ID: %v", err)
	}

	// 构建更新内容
	update := bson.M{
		"$set": bson.M{
			"name":                  req.Name,
			"description":           req.Description,
			"pattern":               req.Pattern,
			"category":              req.Category,
			"riskLevel":             req.RiskLevel,
			"tags":                  req.Tags,
			"enabled":               req.Enabled,
			"context":               req.Context,
			"examples":              req.Examples,
			"falsePositivePatterns": req.FalsePositivePatterns,
			"updatedAt":             time.Now(),
		},
	}

	// 执行更新
	result := s.rulesColl.FindOneAndUpdate(
		s.ctx,
		bson.M{"_id": objID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	// 检查更新结果
	if result.Err() != nil {
		return nil, fmt.Errorf("更新敏感规则失败: %v", result.Err())
	}

	// 解析更新后的规则
	var rule models.SensitiveRule
	if err := result.Decode(&rule); err != nil {
		return nil, fmt.Errorf("解析更新后的规则失败: %v", err)
	}

	return &rule, nil
}

// DeleteRule 删除敏感规则
func (s *Service) DeleteRule(id string) error {
	// 解析ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("无效的规则ID: %v", err)
	}

	// 执行删除
	result, err := s.rulesColl.DeleteOne(s.ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("删除敏感规则失败: %v", err)
	}

	// 检查删除结果
	if result.DeletedCount == 0 {
		return fmt.Errorf("未找到要删除的规则: %s", id)
	}

	return nil
}

// GetRule 获取敏感规则
func (s *Service) GetRule(id string) (*models.SensitiveRule, error) {
	// 解析ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("无效的规则ID: %v", err)
	}

	// 查询规则
	var rule models.SensitiveRule
	err = s.rulesColl.FindOne(s.ctx, bson.M{"_id": objID}).Decode(&rule)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("未找到规则: %s", id)
		}
		return nil, fmt.Errorf("获取敏感规则失败: %v", err)
	}

	return &rule, nil
}

// ListRules 列出敏感规则
func (s *Service) ListRules(category string, riskLevel string, enabled *bool, limit int, skip int) ([]*models.SensitiveRule, error) {
	// 构建查询条件
	filter := bson.M{}
	if category != "" {
		filter["category"] = category
	}
	if riskLevel != "" {
		filter["riskLevel"] = riskLevel
	}
	if enabled != nil {
		filter["enabled"] = *enabled
	}

	// 设置查询选项
	opts := options.Find().
		SetSort(bson.M{"updatedAt": -1}).
		SetLimit(int64(limit)).
		SetSkip(int64(skip))

	// 执行查询
	cursor, err := s.rulesColl.Find(s.ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询敏感规则失败: %v", err)
	}
	defer cursor.Close(s.ctx)

	// 解析结果
	var rules []*models.SensitiveRule
	if err := cursor.All(s.ctx, &rules); err != nil {
		return nil, fmt.Errorf("解析敏感规则失败: %v", err)
	}

	return rules, nil
}

// CreateRuleGroup 创建敏感规则组
func (s *Service) CreateRuleGroup(req models.SensitiveRuleGroupCreateRequest) (*models.SensitiveRuleGroup, error) {
	// 解析规则ID
	var ruleIDs []primitive.ObjectID
	for _, idStr := range req.Rules {
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			return nil, fmt.Errorf("无效的规则ID: %s - %v", idStr, err)
		}
		ruleIDs = append(ruleIDs, id)
	}

	// 创建规则组
	group := &models.SensitiveRuleGroup{
		ID:          primitive.NewObjectID(),
		Name:        req.Name,
		Description: req.Description,
		Rules:       ruleIDs,
		Enabled:     req.Enabled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存规则组
	_, err := s.ruleGroupsColl.InsertOne(s.ctx, group)
	if err != nil {
		return nil, fmt.Errorf("创建敏感规则组失败: %v", err)
	}

	return group, nil
}

// UpdateRuleGroup 更新敏感规则组
func (s *Service) UpdateRuleGroup(id string, req models.SensitiveRuleGroupUpdateRequest) (*models.SensitiveRuleGroup, error) {
	// 解析ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("无效的规则组ID: %v", err)
	}

	// 解析规则ID
	var ruleIDs []primitive.ObjectID
	for _, idStr := range req.Rules {
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			return nil, fmt.Errorf("无效的规则ID: %s - %v", idStr, err)
		}
		ruleIDs = append(ruleIDs, id)
	}

	// 构建更新内容
	update := bson.M{
		"$set": bson.M{
			"name":        req.Name,
			"description": req.Description,
			"rules":       ruleIDs,
			"enabled":     req.Enabled,
			"updatedAt":   time.Now(),
		},
	}

	// 执行更新
	result := s.ruleGroupsColl.FindOneAndUpdate(
		s.ctx,
		bson.M{"_id": objID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	// 检查更新结果
	if result.Err() != nil {
		return nil, fmt.Errorf("更新敏感规则组失败: %v", result.Err())
	}

	// 解析更新后的规则组
	var group models.SensitiveRuleGroup
	if err := result.Decode(&group); err != nil {
		return nil, fmt.Errorf("解析更新后的规则组失败: %v", err)
	}

	return &group, nil
}

// DeleteRuleGroup 删除敏感规则组
func (s *Service) DeleteRuleGroup(id string) error {
	// 解析ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("无效的规则组ID: %v", err)
	}

	// 执行删除
	result, err := s.ruleGroupsColl.DeleteOne(s.ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("删除敏感规则组失败: %v", err)
	}

	// 检查删除结果
	if result.DeletedCount == 0 {
		return fmt.Errorf("未找到要删除的规则组: %s", id)
	}

	return nil
}

// GetRuleGroup 获取敏感规则组
func (s *Service) GetRuleGroup(id string) (*models.SensitiveRuleGroup, error) {
	// 解析ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("无效的规则组ID: %v", err)
	}

	// 查询规则组
	var group models.SensitiveRuleGroup
	err = s.ruleGroupsColl.FindOne(s.ctx, bson.M{"_id": objID}).Decode(&group)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("未找到规则组: %s", id)
		}
		return nil, fmt.Errorf("获取敏感规则组失败: %v", err)
	}

	return &group, nil
}

// ListRuleGroups 列出敏感规则组
func (s *Service) ListRuleGroups(enabled *bool, limit int, skip int) ([]*models.SensitiveRuleGroup, error) {
	// 构建查询条件
	filter := bson.M{}
	if enabled != nil {
		filter["enabled"] = *enabled
	}

	// 设置查询选项
	opts := options.Find().
		SetSort(bson.M{"updatedAt": -1}).
		SetLimit(int64(limit)).
		SetSkip(int64(skip))

	// 执行查询
	cursor, err := s.ruleGroupsColl.Find(s.ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询敏感规则组失败: %v", err)
	}
	defer cursor.Close(s.ctx)

	// 解析结果
	var groups []*models.SensitiveRuleGroup
	if err := cursor.All(s.ctx, &groups); err != nil {
		return nil, fmt.Errorf("解析敏感规则组失败: %v", err)
	}

	return groups, nil
}

// CreateWhitelist 创建敏感信息白名单
func (s *Service) CreateWhitelist(req models.SensitiveWhitelistCreateRequest) (*models.SensitiveWhitelist, error) {
	whitelist := &models.SensitiveWhitelist{
		ID:          primitive.NewObjectID(),
		Type:        req.Type,
		Value:       req.Value,
		Description: req.Description,
		ExpiresAt:   req.ExpiresAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err := s.whitelistColl.InsertOne(s.ctx, whitelist)
	if err != nil {
		return nil, fmt.Errorf("创建敏感信息白名单失败: %v", err)
	}

	return whitelist, nil
}

// UpdateWhitelist 更新敏感信息白名单
func (s *Service) UpdateWhitelist(id string, req models.SensitiveWhitelistUpdateRequest) (*models.SensitiveWhitelist, error) {
	// 解析ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("无效的白名单ID: %v", err)
	}

	// 构建更新内容
	update := bson.M{
		"$set": bson.M{
			"description": req.Description,
			"expiresAt":   req.ExpiresAt,
			"updatedAt":   time.Now(),
		},
	}

	// 执行更新
	result := s.whitelistColl.FindOneAndUpdate(
		s.ctx,
		bson.M{"_id": objID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	// 检查更新结果
	if result.Err() != nil {
		return nil, fmt.Errorf("更新敏感信息白名单失败: %v", result.Err())
	}

	// 解析更新后的白名单
	var whitelist models.SensitiveWhitelist
	if err := result.Decode(&whitelist); err != nil {
		return nil, fmt.Errorf("解析更新后的白名单失败: %v", err)
	}

	return &whitelist, nil
}

// DeleteWhitelist 删除敏感信息白名单
func (s *Service) DeleteWhitelist(id string) error {
	// 解析ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("无效的白名单ID: %v", err)
	}

	// 执行删除
	result, err := s.whitelistColl.DeleteOne(s.ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("删除敏感信息白名单失败: %v", err)
	}

	// 检查删除结果
	if result.DeletedCount == 0 {
		return fmt.Errorf("未找到要删除的白名单: %s", id)
	}

	return nil
}

// GetWhitelist 获取敏感信息白名单
func (s *Service) GetWhitelist(id string) (*models.SensitiveWhitelist, error) {
	// 解析ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("无效的白名单ID: %v", err)
	}

	// 查询白名单
	var whitelist models.SensitiveWhitelist
	err = s.whitelistColl.FindOne(s.ctx, bson.M{"_id": objID}).Decode(&whitelist)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("未找到白名单: %s", id)
		}
		return nil, fmt.Errorf("获取敏感信息白名单失败: %v", err)
	}

	return &whitelist, nil
}

// ListWhitelists 列出敏感信息白名单
func (s *Service) ListWhitelists(whitelistType string, limit int, skip int) ([]*models.SensitiveWhitelist, error) {
	// 构建查询条件
	filter := bson.M{}
	if whitelistType != "" {
		filter["type"] = whitelistType
	}

	// 设置查询选项
	opts := options.Find().
		SetSort(bson.M{"updatedAt": -1}).
		SetLimit(int64(limit)).
		SetSkip(int64(skip))

	// 执行查询
	cursor, err := s.whitelistColl.Find(s.ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询敏感信息白名单失败: %v", err)
	}
	defer cursor.Close(s.ctx)

	// 解析结果
	var whitelists []*models.SensitiveWhitelist
	if err := cursor.All(s.ctx, &whitelists); err != nil {
		return nil, fmt.Errorf("解析敏感信息白名单失败: %v", err)
	}

	return whitelists, nil
}

// DetectSensitiveInfo 敏感信息检测
func (s *Service) DetectSensitiveInfo(projectID string, req models.SensitiveDetectionRequest) (*models.SensitiveDetectionResult, error) {
	// 解析项目ID
	projID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return nil, fmt.Errorf("无效的项目ID: %v", err)
	}
	req.ProjectID = projID

	// 验证请求
	if len(req.Targets) == 0 {
		return nil, fmt.Errorf("目标不能为空")
	}
	if len(req.RuleGroups) == 0 && len(req.Rules) == 0 {
		return nil, fmt.Errorf("规则组或规则不能同时为空")
	}

	// 设置默认值
	if req.Name == "" {
		req.Name = fmt.Sprintf("敏感信息检测-%s", time.Now().Format("2006-01-02 15:04:05"))
	}
	if req.Config.Concurrency <= 0 {
		req.Config.Concurrency = 10
	}
	if req.Config.Timeout <= 0 {
		req.Config.Timeout = 30
	}
	if req.Config.ContextLines <= 0 {
		req.Config.ContextLines = 3
	}

	// 获取规则
	var rules []*models.SensitiveRule

	// 如果指定了规则组，获取规则组中的规则
	if len(req.RuleGroups) > 0 {
		for _, groupID := range req.RuleGroups {
			group, err := s.GetRuleGroup(groupID.Hex())
			if err != nil {
				return nil, fmt.Errorf("获取规则组失败: %v", err)
			}

			for _, ruleID := range group.Rules {
				rule, err := s.GetRule(ruleID.Hex())
				if err != nil {
					return nil, fmt.Errorf("获取规则失败: %v", err)
				}
				rules = append(rules, rule)
			}
		}
	}

	// 如果指定了规则，获取规则
	if len(req.Rules) > 0 {
		for _, ruleID := range req.Rules {
			rule, err := s.GetRule(ruleID.Hex())
			if err != nil {
				return nil, fmt.Errorf("获取规则失败: %v", err)
			}
			rules = append(rules, rule)
		}
	}

	// 去重
	ruleMap := make(map[string]*models.SensitiveRule)
	for _, rule := range rules {
		ruleMap[rule.ID.Hex()] = rule
	}

	rules = make([]*models.SensitiveRule, 0, len(ruleMap))
	for _, rule := range ruleMap {
		rules = append(rules, rule)
	}

	// 获取白名单
	var whitelist []*models.SensitiveWhitelist
	cursor, err := s.whitelistColl.Find(s.ctx, bson.M{"enabled": true})
	if err != nil {
		return nil, fmt.Errorf("获取白名单失败: %v", err)
	}
	defer cursor.Close(s.ctx)

	if err := cursor.All(s.ctx, &whitelist); err != nil {
		return nil, fmt.Errorf("解析白名单失败: %v", err)
	}

	// 创建检测器
	logConfig := utils.DefaultLogConfig()
	logger, err := utils.NewLogger(logConfig)
	if err != nil {
		return nil, fmt.Errorf("创建日志记录器失败: %v", err)
	}
	detector := NewDetector(rules, whitelist, req.Config, logger)

	// 执行检测
	result, err := detector.Detect(s.ctx, req)
	if err != nil {
		return nil, fmt.Errorf("执行敏感信息检测失败: %v", err)
	}

	// 保存结果到数据库
	_, err = s.resultsColl.InsertOne(s.ctx, result)
	if err != nil {
		return nil, fmt.Errorf("保存检测结果失败: %v", err)
	}

	return result, nil
}

// GetDetectionResult 获取检测结果
func (s *Service) GetDetectionResult(id string) (*models.SensitiveDetectionResult, error) {
	// 解析ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("无效的结果ID: %v", err)
	}

	// 查询结果
	var result models.SensitiveDetectionResult
	err = s.resultsColl.FindOne(s.ctx, bson.M{"_id": objID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("未找到结果: %s", id)
		}
		return nil, fmt.Errorf("获取检测结果失败: %v", err)
	}

	return &result, nil
}

// ListDetectionResults 列出检测结果
func (s *Service) ListDetectionResults(projectID string, status string, limit int, skip int) ([]*models.SensitiveDetectionResult, error) {
	// 构建查询条件
	filter := bson.M{}
	if projectID != "" {
		objID, err := primitive.ObjectIDFromHex(projectID)
		if err != nil {
			return nil, fmt.Errorf("无效的项目ID: %v", err)
		}
		filter["projectId"] = objID
	}
	if status != "" {
		filter["status"] = status
	}

	// 设置查询选项
	opts := options.Find().
		SetSort(bson.M{"startTime": -1}).
		SetLimit(int64(limit)).
		SetSkip(int64(skip))

	// 执行查询
	cursor, err := s.resultsColl.Find(s.ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询检测结果失败: %v", err)
	}
	defer cursor.Close(s.ctx)

	// 解析结果
	var results []*models.SensitiveDetectionResult
	if err := cursor.All(s.ctx, &results); err != nil {
		return nil, fmt.Errorf("解析检测结果失败: %v", err)
	}

	return results, nil
}

// DeleteDetectionResult 删除检测结果
func (s *Service) DeleteDetectionResult(id string) error {
	// 解析ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("无效的结果ID: %v", err)
	}

	// 执行删除
	result, err := s.resultsColl.DeleteOne(s.ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("删除检测结果失败: %v", err)
	}

	// 检查删除结果
	if result.DeletedCount == 0 {
		return fmt.Errorf("未找到要删除的结果: %s", id)
	}

	return nil
}

// CreateScanTask 创建敏感信息扫描任务
func (s *Service) CreateScanTask(projectID primitive.ObjectID, urls []string, ruleIDs []string) (primitive.ObjectID, error) {
	// 将字符串ID转换为ObjectID
	var rules []primitive.ObjectID
	for _, id := range ruleIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return primitive.NilObjectID, fmt.Errorf("无效的规则ID: %s - %v", id, err)
		}
		rules = append(rules, objID)
	}

	// 创建敏感信息检测请求
	req := models.SensitiveDetectionRequest{
		ProjectID: projectID,
		Name:      fmt.Sprintf("敏感信息扫描-%s", time.Now().Format("2006-01-02 15:04:05")),
		Targets:   urls,
		Rules:     rules,
		Config: models.SensitiveDetectionConfig{
			Concurrency:  10,
			Timeout:      30,
			ContextLines: 3,
			FollowLinks:  true,
			UserAgent:    "StellarServer/1.0",
		},
	}

	// 执行敏感信息检测
	result, err := s.DetectSensitiveInfo(projectID.Hex(), req)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return result.ID, nil
}

// ListSensitiveInfo 列出敏感信息
func (s *Service) ListSensitiveInfo(projectID primitive.ObjectID, page, limit int) ([]*models.SensitiveDetectionResult, int64, error) {
	// 计算跳过的记录数
	skip := (page - 1) * limit

	// 构建查询条件
	filter := bson.M{"projectId": projectID}

	// 获取总记录数
	total, err := s.resultsColl.CountDocuments(s.ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("计算敏感信息总数失败: %v", err)
	}

	// 查询敏感信息列表
	results, err := s.ListDetectionResults(projectID.Hex(), "", limit, skip)
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// GetSensitiveInfo 获取敏感信息详情
func (s *Service) GetSensitiveInfo(id primitive.ObjectID) (*models.SensitiveDetectionResult, error) {
	return s.GetDetectionResult(id.Hex())
}

// UpdateSensitiveInfoStatus 更新敏感信息状态
func (s *Service) UpdateSensitiveInfoStatus(id primitive.ObjectID, status string) error {
	// 验证状态值
	validStatus := map[string]bool{
		string(models.SensitiveDetectionStatusPending):   true,
		string(models.SensitiveDetectionStatusRunning):   true,
		string(models.SensitiveDetectionStatusCompleted): true,
		string(models.SensitiveDetectionStatusFailed):    true,
		string(models.SensitiveDetectionStatusCancelled): true,
	}

	if !validStatus[status] {
		return fmt.Errorf("无效的状态值: %s", status)
	}

	// 更新状态
	update := bson.M{
		"$set": bson.M{
			"status":    models.SensitiveDetectionStatus(status),
			"updatedAt": time.Now(),
		},
	}

	result, err := s.resultsColl.UpdateOne(s.ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("更新敏感信息状态失败: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("未找到要更新的敏感信息: %s", id.Hex())
	}

	return nil
}

// DeleteSensitiveInfo 删除敏感信息
func (s *Service) DeleteSensitiveInfo(id primitive.ObjectID) error {
	return s.DeleteDetectionResult(id.Hex())
}

// ListSensitiveRules 列出敏感规则
func (s *Service) ListSensitiveRules() ([]*models.SensitiveRule, error) {
	// 使用现有的ListRules方法，不过滤任何条件，获取所有规则
	return s.ListRules("", "", nil, 100, 0)
}

// CreateSensitiveRule 创建敏感规则
func (s *Service) CreateSensitiveRule(rule models.SensitiveRule) (primitive.ObjectID, error) {
	// 将SensitiveRule转换为SensitiveRuleCreateRequest
	req := models.SensitiveRuleCreateRequest{
		Name:                  rule.Name,
		Description:           rule.Description,
		Type:                  rule.Type,
		Pattern:               rule.Pattern,
		Category:              rule.Category,
		RiskLevel:             rule.RiskLevel,
		Tags:                  rule.Tags,
		Enabled:               rule.Enabled,
		Context:               rule.Context,
		Examples:              rule.Examples,
		FalsePositivePatterns: rule.FalsePositivePatterns,
	}

	// 使用现有的CreateRule方法
	createdRule, err := s.CreateRule(req)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return createdRule.ID, nil
}

// UpdateSensitiveRule 更新敏感规则
func (s *Service) UpdateSensitiveRule(rule models.SensitiveRule) error {
	// 将SensitiveRule转换为SensitiveRuleUpdateRequest
	req := models.SensitiveRuleUpdateRequest{
		Name:                  rule.Name,
		Description:           rule.Description,
		Pattern:               rule.Pattern,
		Category:              rule.Category,
		RiskLevel:             rule.RiskLevel,
		Tags:                  rule.Tags,
		Enabled:               rule.Enabled,
		Context:               rule.Context,
		Examples:              rule.Examples,
		FalsePositivePatterns: rule.FalsePositivePatterns,
	}

	// 使用现有的UpdateRule方法
	_, err := s.UpdateRule(rule.ID.Hex(), req)
	return err
}

// DeleteSensitiveRule 删除敏感规则
func (s *Service) DeleteSensitiveRule(id primitive.ObjectID) error {
	return s.DeleteRule(id.Hex())
}
