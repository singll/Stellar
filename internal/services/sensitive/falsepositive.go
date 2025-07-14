package sensitive

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FalsePositiveManager 误报管理器
type FalsePositiveManager struct {
	db               *mongo.Database
	engine           *DetectionEngine
	whitelistRules   []*WhitelistRule
	learningEngine   *LearningEngine
	reviewQueue      []*ReviewItem
}

// WhitelistRule 白名单规则
type WhitelistRule struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Type        WhitelistType      `json:"type"`
	Pattern     string             `json:"pattern"`
	Regex       *regexp.Regexp     `json:"-"`
	Target      string             `json:"target"`     // 目标类型：content, file_path, rule_id
	Category    string             `json:"category"`   // 适用的类别
	RuleID      string             `json:"rule_id"`    // 适用的规则ID
	Enabled     bool               `json:"enabled"`
	CreatedBy   string             `json:"created_by"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	Statistics  *WhitelistStats    `json:"statistics"`
}

// WhitelistType 白名单类型
type WhitelistType string

const (
	WhitelistTypeContent   WhitelistType = "content"   // 内容白名单
	WhitelistTypeFilePath  WhitelistType = "file_path" // 文件路径白名单
	WhitelistTypeRule      WhitelistType = "rule"      // 规则白名单
	WhitelistTypeGlobal    WhitelistType = "global"    // 全局白名单
)

// WhitelistStats 白名单统计
type WhitelistStats struct {
	TotalMatches     int64     `json:"total_matches"`
	RecentMatches    int64     `json:"recent_matches"`    // 最近7天匹配次数
	LastMatchedAt    time.Time `json:"last_matched_at"`
	EffectivenessRate float64  `json:"effectiveness_rate"` // 有效率
}

// ReviewItem 审查项目
type ReviewItem struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DetectionID  primitive.ObjectID `json:"detection_id"`
	ResultID     primitive.ObjectID `json:"result_id"`
	MatchIndex   int                `json:"match_index"`
	RuleID       string             `json:"rule_id"`
	RuleName     string             `json:"rule_name"`
	Category     string             `json:"category"`
	Severity     SeverityLevel      `json:"severity"`
	Target       string             `json:"target"`
	MatchedText  string             `json:"matched_text"`
	Context      string             `json:"context"`
	Status       ReviewStatus       `json:"status"`
	ReviewedBy   string             `json:"reviewed_by"`
	ReviewedAt   *time.Time         `json:"reviewed_at"`
	Comments     string             `json:"comments"`
	Action       ReviewAction       `json:"action"`
	Priority     ReviewPriority     `json:"priority"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// ReviewStatus 审查状态
type ReviewStatus string

const (
	ReviewStatusPending    ReviewStatus = "pending"     // 待审查
	ReviewStatusReviewing  ReviewStatus = "reviewing"   // 审查中
	ReviewStatusCompleted  ReviewStatus = "completed"   // 已完成
	ReviewStatusSkipped    ReviewStatus = "skipped"     // 已跳过
)

// ReviewAction 审查动作
type ReviewAction string

const (
	ReviewActionConfirm      ReviewAction = "confirm"        // 确认为真阳性
	ReviewActionFalsePositive ReviewAction = "false_positive" // 标记为误报
	ReviewActionCreateRule   ReviewAction = "create_rule"    // 创建白名单规则
	ReviewActionUpdateRule   ReviewAction = "update_rule"    // 更新现有规则
	ReviewActionIgnore       ReviewAction = "ignore"         // 忽略
)

// ReviewPriority 审查优先级
type ReviewPriority string

const (
	ReviewPriorityHigh   ReviewPriority = "high"
	ReviewPriorityMedium ReviewPriority = "medium"
	ReviewPriorityLow    ReviewPriority = "low"
)

// LearningEngine 学习引擎
type LearningEngine struct {
	patternAnalyzer *PatternAnalyzer
	behaviorTracker *BehaviorTracker
	suggestions     []*AutoSuggestion
}

// PatternAnalyzer 模式分析器
type PatternAnalyzer struct {
	commonPatterns    map[string]*PatternInfo
	falsePositivePatterns map[string]*PatternInfo
}

// PatternInfo 模式信息
type PatternInfo struct {
	Pattern     string    `json:"pattern"`
	Frequency   int64     `json:"frequency"`
	Category    string    `json:"category"`
	Confidence  float64   `json:"confidence"`
	LastSeen    time.Time `json:"last_seen"`
	Examples    []string  `json:"examples"`
}

// BehaviorTracker 行为跟踪器
type BehaviorTracker struct {
	userActions     map[string]*UserBehavior
	ruleFeedback    map[string]*RuleFeedback
}

// UserBehavior 用户行为
type UserBehavior struct {
	UserID              string    `json:"user_id"`
	TotalReviews        int64     `json:"total_reviews"`
	FalsePositiveRate   float64   `json:"false_positive_rate"`
	AverageReviewTime   float64   `json:"average_review_time"`
	PreferredActions    map[ReviewAction]int64 `json:"preferred_actions"`
	LastActivity        time.Time `json:"last_activity"`
}

// RuleFeedback 规则反馈
type RuleFeedback struct {
	RuleID            string    `json:"rule_id"`
	TotalDetections   int64     `json:"total_detections"`
	FalsePositives    int64     `json:"false_positives"`
	TruePositives     int64     `json:"true_positives"`
	FalsePositiveRate float64   `json:"false_positive_rate"`
	Accuracy          float64   `json:"accuracy"`
	LastFeedback      time.Time `json:"last_feedback"`
}

// AutoSuggestion 自动建议
type AutoSuggestion struct {
	ID          primitive.ObjectID `json:"id"`
	Type        SuggestionType     `json:"type"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Pattern     string             `json:"pattern"`
	RuleID      string             `json:"rule_id"`
	Category    string             `json:"category"`
	Confidence  float64            `json:"confidence"`
	Examples    []string           `json:"examples"`
	Impact      string             `json:"impact"`
	CreatedAt   time.Time          `json:"created_at"`
	Status      SuggestionStatus   `json:"status"`
}

// SuggestionType 建议类型
type SuggestionType string

const (
	SuggestionTypeWhitelist     SuggestionType = "whitelist"      // 白名单建议
	SuggestionTypeRuleUpdate    SuggestionType = "rule_update"    // 规则更新建议
	SuggestionTypeNewRule       SuggestionType = "new_rule"       // 新规则建议
	SuggestionTypeThreshold     SuggestionType = "threshold"      // 阈值调整建议
)

// SuggestionStatus 建议状态
type SuggestionStatus string

const (
	SuggestionStatusPending  SuggestionStatus = "pending"
	SuggestionStatusAccepted SuggestionStatus = "accepted"
	SuggestionStatusRejected SuggestionStatus = "rejected"
)

// NewFalsePositiveManager 创建误报管理器
func NewFalsePositiveManager(db *mongo.Database, engine *DetectionEngine) *FalsePositiveManager {
	manager := &FalsePositiveManager{
		db:             db,
		engine:         engine,
		whitelistRules: []*WhitelistRule{},
		reviewQueue:    []*ReviewItem{},
		learningEngine: NewLearningEngine(),
	}
	
	// 加载白名单规则
	manager.loadWhitelistRules()
	
	return manager
}

// NewLearningEngine 创建学习引擎
func NewLearningEngine() *LearningEngine {
	return &LearningEngine{
		patternAnalyzer: &PatternAnalyzer{
			commonPatterns:         make(map[string]*PatternInfo),
			falsePositivePatterns: make(map[string]*PatternInfo),
		},
		behaviorTracker: &BehaviorTracker{
			userActions:  make(map[string]*UserBehavior),
			ruleFeedback: make(map[string]*RuleFeedback),
		},
		suggestions: []*AutoSuggestion{},
	}
}

// loadWhitelistRules 加载白名单规则
func (fpm *FalsePositiveManager) loadWhitelistRules() {
	ctx := context.Background()
	collection := fpm.db.Collection("whitelist_rules")
	
	cursor, err := collection.Find(ctx, bson.M{"enabled": true})
	if err != nil {
		return
	}
	defer cursor.Close(ctx)
	
	var rules []*WhitelistRule
	if err := cursor.All(ctx, &rules); err != nil {
		return
	}
	
	// 编译正则表达式
	for _, rule := range rules {
		if rule.Pattern != "" {
			if regex, err := regexp.Compile(rule.Pattern); err == nil {
				rule.Regex = regex
			}
		}
	}
	
	fpm.whitelistRules = rules
}

// CreateWhitelistRule 创建白名单规则
func (fpm *FalsePositiveManager) CreateWhitelistRule(rule *WhitelistRule) error {
	// 验证规则
	if err := fpm.validateWhitelistRule(rule); err != nil {
		return fmt.Errorf("白名单规则验证失败: %v", err)
	}
	
	// 编译正则表达式
	if rule.Pattern != "" {
		regex, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return fmt.Errorf("正则表达式编译失败: %v", err)
		}
		rule.Regex = regex
	}
	
	// 设置时间戳
	now := time.Now()
	rule.ID = primitive.NewObjectID()
	rule.CreatedAt = now
	rule.UpdatedAt = now
	rule.Enabled = true
	rule.Statistics = &WhitelistStats{}
	
	// 保存到数据库
	ctx := context.Background()
	collection := fpm.db.Collection("whitelist_rules")
	
	_, err := collection.InsertOne(ctx, rule)
	if err != nil {
		return fmt.Errorf("保存白名单规则失败: %v", err)
	}
	
	// 添加到内存
	fpm.whitelistRules = append(fpm.whitelistRules, rule)
	
	return nil
}

// validateWhitelistRule 验证白名单规则
func (fpm *FalsePositiveManager) validateWhitelistRule(rule *WhitelistRule) error {
	if rule.Name == "" {
		return fmt.Errorf("规则名称不能为空")
	}
	
	if rule.Type == "" {
		return fmt.Errorf("规则类型不能为空")
	}
	
	validTypes := map[WhitelistType]bool{
		WhitelistTypeContent:  true,
		WhitelistTypeFilePath: true,
		WhitelistTypeRule:     true,
		WhitelistTypeGlobal:   true,
	}
	
	if !validTypes[rule.Type] {
		return fmt.Errorf("无效的规则类型: %s", rule.Type)
	}
	
	if rule.Pattern == "" {
		return fmt.Errorf("规则模式不能为空")
	}
	
	// 验证正则表达式
	if _, err := regexp.Compile(rule.Pattern); err != nil {
		return fmt.Errorf("无效的正则表达式: %v", err)
	}
	
	return nil
}

// IsWhitelisted 检查是否在白名单中
func (fpm *FalsePositiveManager) IsWhitelisted(match *SensitiveMatch, target, content string) bool {
	for _, rule := range fpm.whitelistRules {
		if !rule.Enabled {
			continue
		}
		
		// 检查规则适用性
		if !fpm.isRuleApplicable(rule, match, target) {
			continue
		}
		
		// 检查模式匹配
		if fpm.matchesWhitelistRule(rule, match, target, content) {
			// 更新统计信息
			fpm.updateWhitelistStats(rule)
			return true
		}
	}
	
	return false
}

// isRuleApplicable 检查规则是否适用
func (fpm *FalsePositiveManager) isRuleApplicable(rule *WhitelistRule, match *SensitiveMatch, target string) bool {
	// 检查类别
	if rule.Category != "" && rule.Category != match.Category {
		return false
	}
	
	// 检查规则ID
	if rule.RuleID != "" && rule.RuleID != match.RuleID {
		return false
	}
	
	return true
}

// matchesWhitelistRule 检查是否匹配白名单规则
func (fpm *FalsePositiveManager) matchesWhitelistRule(rule *WhitelistRule, match *SensitiveMatch, target, content string) bool {
	switch rule.Type {
	case WhitelistTypeContent:
		return rule.Regex != nil && rule.Regex.MatchString(match.Match)
		
	case WhitelistTypeFilePath:
		return rule.Regex != nil && rule.Regex.MatchString(target)
		
	case WhitelistTypeRule:
		return rule.RuleID == match.RuleID
		
	case WhitelistTypeGlobal:
		// 全局规则可以匹配内容或路径
		if rule.Regex != nil {
			return rule.Regex.MatchString(match.Match) || rule.Regex.MatchString(target)
		}
		
	default:
		return false
	}
	
	return false
}

// updateWhitelistStats 更新白名单统计
func (fpm *FalsePositiveManager) updateWhitelistStats(rule *WhitelistRule) {
	rule.Statistics.TotalMatches++
	rule.Statistics.LastMatchedAt = time.Now()
	
	// 异步更新数据库
	go func() {
		ctx := context.Background()
		collection := fpm.db.Collection("whitelist_rules")
		
		filter := bson.M{"_id": rule.ID}
		update := bson.M{"$set": bson.M{
			"statistics.total_matches":   rule.Statistics.TotalMatches,
			"statistics.last_matched_at": rule.Statistics.LastMatchedAt,
			"updated_at":                 time.Now(),
		}}
		
		collection.UpdateOne(ctx, filter, update)
	}()
}

// AddToReviewQueue 添加到审查队列
func (fpm *FalsePositiveManager) AddToReviewQueue(result *DetectionResult, matchIndex int, priority ReviewPriority) error {
	if matchIndex >= len(result.Matches) {
		return fmt.Errorf("无效的匹配索引")
	}
	
	match := result.Matches[matchIndex]
	
	reviewItem := &ReviewItem{
		ID:          primitive.NewObjectID(),
		DetectionID: primitive.NewObjectID(), // 应该从检测请求中获取
		ResultID:    result.ID,
		MatchIndex:  matchIndex,
		RuleID:      match.RuleID,
		RuleName:    match.RuleName,
		Category:    match.Category,
		Severity:    match.Severity,
		Target:      result.URL,
		MatchedText: match.Match,
		Context:     match.Context,
		Status:      ReviewStatusPending,
		Priority:    priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// 保存到数据库
	ctx := context.Background()
	collection := fpm.db.Collection("review_queue")
	
	_, err := collection.InsertOne(ctx, reviewItem)
	if err != nil {
		return fmt.Errorf("添加审查项目失败: %v", err)
	}
	
	// 添加到内存队列
	fpm.reviewQueue = append(fpm.reviewQueue, reviewItem)
	
	return nil
}

// ProcessReviewItem 处理审查项目
func (fpm *FalsePositiveManager) ProcessReviewItem(itemID primitive.ObjectID, userID string, action ReviewAction, comments string) error {
	// 查找审查项目
	item, err := fpm.getReviewItem(itemID)
	if err != nil {
		return err
	}
	
	// 更新审查项目
	now := time.Now()
	item.Status = ReviewStatusCompleted
	item.Action = action
	item.ReviewedBy = userID
	item.ReviewedAt = &now
	item.Comments = comments
	item.UpdatedAt = now
	
	// 根据动作执行相应操作
	switch action {
	case ReviewActionFalsePositive:
		err = fpm.handleFalsePositive(item)
	case ReviewActionCreateRule:
		err = fpm.handleCreateRule(item)
	case ReviewActionUpdateRule:
		err = fpm.handleUpdateRule(item)
	case ReviewActionConfirm:
		err = fpm.handleConfirm(item)
	}
	
	if err != nil {
		return err
	}
	
	// 更新数据库
	ctx := context.Background()
	collection := fpm.db.Collection("review_queue")
	
	filter := bson.M{"_id": itemID}
	update := bson.M{"$set": item}
	
	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("更新审查项目失败: %v", err)
	}
	
	// 学习用户行为
	fpm.learningEngine.LearnFromUserAction(userID, action, item)
	
	return nil
}

// handleFalsePositive 处理误报
func (fpm *FalsePositiveManager) handleFalsePositive(item *ReviewItem) error {
	// 自动生成白名单规则建议
	suggestion := fpm.generateWhitelistSuggestion(item)
	
	// 添加到建议列表
	fpm.learningEngine.suggestions = append(fpm.learningEngine.suggestions, suggestion)
	
	// 记录误报模式
	fpm.learningEngine.patternAnalyzer.RecordFalsePositive(item.MatchedText, item.Category)
	
	return nil
}

// handleCreateRule 处理创建规则
func (fpm *FalsePositiveManager) handleCreateRule(item *ReviewItem) error {
	// 自动创建白名单规则
	rule := &WhitelistRule{
		Name:        fmt.Sprintf("Auto-generated whitelist for %s", item.RuleName),
		Description: fmt.Sprintf("误报规则，源于审查项目 %s", item.ID.Hex()),
		Type:        WhitelistTypeContent,
		Pattern:     regexp.QuoteMeta(item.MatchedText), // 精确匹配
		Category:    item.Category,
		RuleID:      item.RuleID,
		CreatedBy:   item.ReviewedBy,
	}
	
	return fpm.CreateWhitelistRule(rule)
}

// handleUpdateRule 处理更新规则
func (fpm *FalsePositiveManager) handleUpdateRule(item *ReviewItem) error {
	// 这里应该更新检测规则以减少误报
	// 实现逻辑根据具体需求而定
	return nil
}

// handleConfirm 处理确认
func (fpm *FalsePositiveManager) handleConfirm(item *ReviewItem) error {
	// 记录真阳性
	fpm.learningEngine.behaviorTracker.RecordTruePositive(item.RuleID)
	return nil
}

// getReviewItem 获取审查项目
func (fpm *FalsePositiveManager) getReviewItem(itemID primitive.ObjectID) (*ReviewItem, error) {
	ctx := context.Background()
	collection := fpm.db.Collection("review_queue")
	
	var item ReviewItem
	err := collection.FindOne(ctx, bson.M{"_id": itemID}).Decode(&item)
	if err != nil {
		return nil, fmt.Errorf("获取审查项目失败: %v", err)
	}
	
	return &item, nil
}

// generateWhitelistSuggestion 生成白名单建议
func (fpm *FalsePositiveManager) generateWhitelistSuggestion(item *ReviewItem) *AutoSuggestion {
	return &AutoSuggestion{
		ID:          primitive.NewObjectID(),
		Type:        SuggestionTypeWhitelist,
		Title:       fmt.Sprintf("为 '%s' 创建白名单规则", item.MatchedText),
		Description: fmt.Sprintf("基于审查项目 %s 的误报反馈", item.ID.Hex()),
		Pattern:     regexp.QuoteMeta(item.MatchedText),
		RuleID:      item.RuleID,
		Category:    item.Category,
		Confidence:  0.8, // 基于用户反馈的高置信度
		Examples:    []string{item.MatchedText},
		Impact:      "将减少此模式的误报",
		CreatedAt:   time.Now(),
		Status:      SuggestionStatusPending,
	}
}

// GetPendingReviews 获取待审查项目
func (fpm *FalsePositiveManager) GetPendingReviews(limit int) ([]*ReviewItem, error) {
	ctx := context.Background()
	collection := fpm.db.Collection("review_queue")
	
	filter := bson.M{"status": ReviewStatusPending}
	opts := &options.FindOptions{
		Limit: &[]int64{int64(limit)}[0],
		Sort:  bson.M{"priority": 1, "created_at": 1}, // 按优先级和时间排序
	}
	
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var items []*ReviewItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	
	return items, nil
}

// GetSuggestions 获取自动建议
func (fpm *FalsePositiveManager) GetSuggestions() []*AutoSuggestion {
	return fpm.learningEngine.suggestions
}

// ApplySuggestion 应用建议
func (fpm *FalsePositiveManager) ApplySuggestion(suggestionID primitive.ObjectID, userID string) error {
	// 查找建议
	var suggestion *AutoSuggestion
	for _, s := range fpm.learningEngine.suggestions {
		if s.ID == suggestionID {
			suggestion = s
			break
		}
	}
	
	if suggestion == nil {
		return fmt.Errorf("建议不存在: %s", suggestionID.Hex())
	}
	
	switch suggestion.Type {
	case SuggestionTypeWhitelist:
		// 创建白名单规则
		rule := &WhitelistRule{
			Name:        suggestion.Title,
			Description: suggestion.Description,
			Type:        WhitelistTypeContent,
			Pattern:     suggestion.Pattern,
			Category:    suggestion.Category,
			RuleID:      suggestion.RuleID,
			CreatedBy:   userID,
		}
		
		err := fpm.CreateWhitelistRule(rule)
		if err != nil {
			return err
		}
		
	case SuggestionTypeRuleUpdate:
		// 更新检测规则
		// 实现逻辑根据具体需求
		
	default:
		return fmt.Errorf("不支持的建议类型: %s", suggestion.Type)
	}
	
	// 标记建议为已接受
	suggestion.Status = SuggestionStatusAccepted
	
	return nil
}

// 学习引擎方法
func (le *LearningEngine) LearnFromUserAction(userID string, action ReviewAction, item *ReviewItem) {
	// 更新用户行为
	if behavior, exists := le.behaviorTracker.userActions[userID]; exists {
		behavior.TotalReviews++
		behavior.PreferredActions[action]++
		behavior.LastActivity = time.Now()
		
		// 计算误报率
		if action == ReviewActionFalsePositive {
			behavior.FalsePositiveRate = float64(behavior.PreferredActions[ReviewActionFalsePositive]) / float64(behavior.TotalReviews)
		}
	} else {
		// 新用户
		le.behaviorTracker.userActions[userID] = &UserBehavior{
			UserID:            userID,
			TotalReviews:      1,
			PreferredActions:  map[ReviewAction]int64{action: 1},
			LastActivity:      time.Now(),
		}
	}
	
	// 更新规则反馈
	if feedback, exists := le.behaviorTracker.ruleFeedback[item.RuleID]; exists {
		feedback.TotalDetections++
		
		if action == ReviewActionFalsePositive {
			feedback.FalsePositives++
		} else if action == ReviewActionConfirm {
			feedback.TruePositives++
		}
		
		// 重新计算准确率
		total := feedback.FalsePositives + feedback.TruePositives
		if total > 0 {
			feedback.FalsePositiveRate = float64(feedback.FalsePositives) / float64(total)
			feedback.Accuracy = float64(feedback.TruePositives) / float64(total)
		}
		
		feedback.LastFeedback = time.Now()
	} else {
		// 新规则
		feedback := &RuleFeedback{
			RuleID:          item.RuleID,
			TotalDetections: 1,
			LastFeedback:    time.Now(),
		}
		
		if action == ReviewActionFalsePositive {
			feedback.FalsePositives = 1
		} else if action == ReviewActionConfirm {
			feedback.TruePositives = 1
		}
		
		le.behaviorTracker.ruleFeedback[item.RuleID] = feedback
	}
}

func (pa *PatternAnalyzer) RecordFalsePositive(pattern, category string) {
	if info, exists := pa.falsePositivePatterns[pattern]; exists {
		info.Frequency++
		info.LastSeen = time.Now()
	} else {
		pa.falsePositivePatterns[pattern] = &PatternInfo{
			Pattern:   pattern,
			Frequency: 1,
			Category:  category,
			LastSeen:  time.Now(),
		}
	}
}

func (bt *BehaviorTracker) RecordTruePositive(ruleID string) {
	if feedback, exists := bt.ruleFeedback[ruleID]; exists {
		feedback.TruePositives++
		
		// 重新计算准确率
		total := feedback.FalsePositives + feedback.TruePositives
		if total > 0 {
			feedback.Accuracy = float64(feedback.TruePositives) / float64(total)
			feedback.FalsePositiveRate = float64(feedback.FalsePositives) / float64(total)
		}
	}
}