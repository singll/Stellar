package taskmanager

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TaskScheduler 任务调度器
type TaskScheduler struct {
	db           *mongo.Database
	taskManager  *TaskManager
	cron         *cron.Cron
	scheduleRules map[string]*models.TaskScheduleRule
	cronJobs     map[string]cron.EntryID
	mutex        sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewTaskScheduler 创建任务调度器
func NewTaskScheduler(db *mongo.Database, taskManager *TaskManager) *TaskScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	
	// 创建cron调度器，使用UTC时间
	c := cron.New(cron.WithLocation(time.UTC))
	
	scheduler := &TaskScheduler{
		db:           db,
		taskManager:  taskManager,
		cron:         c,
		scheduleRules: make(map[string]*models.TaskScheduleRule),
		cronJobs:     make(map[string]cron.EntryID),
		ctx:          ctx,
		cancel:       cancel,
	}
	
	return scheduler
}

// Start 启动任务调度器
func (s *TaskScheduler) Start() error {
	// 加载调度规则
	if err := s.loadScheduleRules(); err != nil {
		return fmt.Errorf("加载调度规则失败: %v", err)
	}
	
	// 启动cron调度器
	s.cron.Start()
	
	// 启动定期检查
	go s.runPeriodicCheck()
	
	log.Printf("任务调度器启动成功，加载了 %d 个调度规则", len(s.scheduleRules))
	return nil
}

// Stop 停止任务调度器
func (s *TaskScheduler) Stop() {
	s.cancel()
	s.cron.Stop()
	log.Println("任务调度器已停止")
}

// CreateScheduleRule 创建调度规则
func (s *TaskScheduler) CreateScheduleRule(rule *models.TaskScheduleRule) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// 设置默认值
	rule.ID = primitive.NewObjectID()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	rule.RunCount = 0
	
	// 计算下次执行时间
	if err := s.calculateNextRunTime(rule); err != nil {
		return fmt.Errorf("计算下次执行时间失败: %v", err)
	}
	
	// 保存到数据库
	_, err := s.db.Collection("task_schedule_rules").InsertOne(s.ctx, rule)
	if err != nil {
		return fmt.Errorf("保存调度规则失败: %v", err)
	}
	
	// 添加到内存
	s.scheduleRules[rule.ID.Hex()] = rule
	
	// 如果规则启用，添加到cron
	if rule.Enabled {
		if err := s.addCronJob(rule); err != nil {
			log.Printf("添加cron任务失败: %v", err)
		}
	}
	
	return nil
}

// UpdateScheduleRule 更新调度规则
func (s *TaskScheduler) UpdateScheduleRule(ruleID string, updates *models.TaskScheduleRule) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	objID, err := primitive.ObjectIDFromHex(ruleID)
	if err != nil {
		return fmt.Errorf("无效的规则ID: %v", err)
	}
	
	// 获取原规则
	rule, exists := s.scheduleRules[ruleID]
	if !exists {
		return fmt.Errorf("调度规则不存在")
	}
	
	// 如果cron表达式或时区改变，需要重新计算
	if updates.CronExpression != "" && updates.CronExpression != rule.CronExpression {
		rule.CronExpression = updates.CronExpression
		if err := s.calculateNextRunTime(rule); err != nil {
			return fmt.Errorf("计算下次执行时间失败: %v", err)
		}
	}
	
	// 更新其他字段
	if updates.Name != "" {
		rule.Name = updates.Name
	}
	if updates.Description != "" {
		rule.Description = updates.Description
	}
	if updates.Enabled != rule.Enabled {
		rule.Enabled = updates.Enabled
	}
	if updates.MaxRuns != nil {
		rule.MaxRuns = updates.MaxRuns
	}
	
	rule.UpdatedAt = time.Now()
	
	// 更新数据库
	_, err = s.db.Collection("task_schedule_rules").UpdateOne(
		s.ctx,
		bson.M{"_id": objID},
		bson.M{"$set": rule},
	)
	if err != nil {
		return fmt.Errorf("更新调度规则失败: %v", err)
	}
	
	// 重新调度cron任务
	s.removeCronJob(ruleID)
	if rule.Enabled {
		if err := s.addCronJob(rule); err != nil {
			log.Printf("添加cron任务失败: %v", err)
		}
	}
	
	return nil
}

// DeleteScheduleRule 删除调度规则
func (s *TaskScheduler) DeleteScheduleRule(ruleID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	objID, err := primitive.ObjectIDFromHex(ruleID)
	if err != nil {
		return fmt.Errorf("无效的规则ID: %v", err)
	}
	
	// 移除cron任务
	s.removeCronJob(ruleID)
	
	// 从数据库删除
	_, err = s.db.Collection("task_schedule_rules").DeleteOne(
		s.ctx,
		bson.M{"_id": objID},
	)
	if err != nil {
		return fmt.Errorf("删除调度规则失败: %v", err)
	}
	
	// 从内存删除
	delete(s.scheduleRules, ruleID)
	
	return nil
}

// ToggleScheduleRule 切换调度规则状态
func (s *TaskScheduler) ToggleScheduleRule(ruleID string, enabled bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	objID, err := primitive.ObjectIDFromHex(ruleID)
	if err != nil {
		return fmt.Errorf("无效的规则ID: %v", err)
	}
	
	rule, exists := s.scheduleRules[ruleID]
	if !exists {
		return fmt.Errorf("调度规则不存在")
	}
	
	rule.Enabled = enabled
	rule.UpdatedAt = time.Now()
	
	// 更新数据库
	_, err = s.db.Collection("task_schedule_rules").UpdateOne(
		s.ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{
			"enabled": enabled,
			"updatedAt": time.Now(),
		}},
	)
	if err != nil {
		return fmt.Errorf("更新调度规则失败: %v", err)
	}
	
	// 重新调度cron任务
	s.removeCronJob(ruleID)
	if enabled {
		if err := s.addCronJob(rule); err != nil {
			log.Printf("添加cron任务失败: %v", err)
		}
	}
	
	return nil
}

// TriggerScheduleRule 手动触发调度规则
func (s *TaskScheduler) TriggerScheduleRule(ruleID string) (*models.Task, error) {
	s.mutex.RLock()
	rule, exists := s.scheduleRules[ruleID]
	s.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("调度规则不存在")
	}
	
	// 执行任务
	task, err := s.executeScheduledTask(rule)
	if err != nil {
		return nil, fmt.Errorf("执行调度任务失败: %v", err)
	}
	
	return task, nil
}

// GetScheduleRules 获取调度规则列表
func (s *TaskScheduler) GetScheduleRules(projectID string) ([]*models.TaskScheduleRule, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	var rules []*models.TaskScheduleRule
	for _, rule := range s.scheduleRules {
		if projectID == "" || rule.ProjectID == projectID {
			rules = append(rules, rule)
		}
	}
	
	return rules, nil
}

// loadScheduleRules 加载调度规则
func (s *TaskScheduler) loadScheduleRules() error {
	cursor, err := s.db.Collection("task_schedule_rules").Find(s.ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(s.ctx)
	
	for cursor.Next(s.ctx) {
		var rule models.TaskScheduleRule
		if err := cursor.Decode(&rule); err != nil {
			log.Printf("解析调度规则失败: %v", err)
			continue
		}
		
		// 重新计算下次执行时间
		if err := s.calculateNextRunTime(&rule); err != nil {
			log.Printf("计算下次执行时间失败: %v", err)
			continue
		}
		
		s.scheduleRules[rule.ID.Hex()] = &rule
		
		// 如果规则启用，添加到cron
		if rule.Enabled {
			if err := s.addCronJob(&rule); err != nil {
				log.Printf("添加cron任务失败: %v", err)
			}
		}
	}
	
	return nil
}

// addCronJob 添加cron任务
func (s *TaskScheduler) addCronJob(rule *models.TaskScheduleRule) error {
	jobFunc := func() {
		// 检查是否达到最大执行次数
		if rule.MaxRuns != nil && rule.RunCount >= *rule.MaxRuns {
			log.Printf("调度规则 %s 已达到最大执行次数，停止调度", rule.Name)
			s.removeCronJob(rule.ID.Hex())
			return
		}
		
		// 执行任务
		task, err := s.executeScheduledTask(rule)
		if err != nil {
			log.Printf("执行调度任务失败: %v", err)
			return
		}
		
		log.Printf("调度任务执行成功: %s (任务ID: %s)", rule.Name, task.ID.Hex())
		
		// 更新执行统计
		s.updateRuleStats(rule)
	}
	
	// 添加到cron
	entryID, err := s.cron.AddFunc(rule.CronExpression, jobFunc)
	if err != nil {
		return fmt.Errorf("添加cron任务失败: %v", err)
	}
	
	s.cronJobs[rule.ID.Hex()] = entryID
	return nil
}

// removeCronJob 移除cron任务
func (s *TaskScheduler) removeCronJob(ruleID string) {
	if entryID, exists := s.cronJobs[ruleID]; exists {
		s.cron.Remove(entryID)
		delete(s.cronJobs, ruleID)
	}
}

// executeScheduledTask 执行调度任务
func (s *TaskScheduler) executeScheduledTask(rule *models.TaskScheduleRule) (*models.Task, error) {
	// 获取任务模板
	var template models.TaskTemplate
	templateObjID, err := primitive.ObjectIDFromHex(rule.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("无效的模板ID: %v", err)
	}
	
	err = s.db.Collection("task_templates").FindOne(s.ctx, bson.M{"_id": templateObjID}).Decode(&template)
	if err != nil {
		return nil, fmt.Errorf("查找任务模板失败: %v", err)
	}
	
	// 转换项目ID
	projectObjID, err := primitive.ObjectIDFromHex(rule.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("无效的项目ID: %v", err)
	}
	
	// 创建任务
	task := &models.Task{
		ID:        primitive.NewObjectID(),
		Name:      fmt.Sprintf("[调度] %s", template.Name),
		Type:      template.Type,
		Status:    string(models.TaskStatusPending),
		Priority:  template.Priority,
		ProjectID: projectObjID,
		Config:    template.Config,
		Timeout:   template.Timeout,
		MaxRetries: template.MaxRetries,
		Tags:      template.Tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		// 将创建者设置为空的ObjectID，因为这是系统调度的任务
		CreatedBy: primitive.NilObjectID,
		// 移除Metadata字段，因为Task结构体中没有这个字段
	}
	
	// 提交任务到任务管理器
	if err := s.taskManager.SubmitTask(task); err != nil {
		return nil, fmt.Errorf("提交任务失败: %v", err)
	}
	
	return task, nil
}

// updateRuleStats 更新规则统计信息
func (s *TaskScheduler) updateRuleStats(rule *models.TaskScheduleRule) {
	rule.RunCount++
	rule.LastRunTime = time.Now()
	
	// 计算下次执行时间
	if err := s.calculateNextRunTime(rule); err != nil {
		log.Printf("计算下次执行时间失败: %v", err)
	}
	
	// 更新数据库
	_, err := s.db.Collection("task_schedule_rules").UpdateOne(
		s.ctx,
		bson.M{"_id": rule.ID},
		bson.M{"$set": bson.M{
			"runCount": rule.RunCount,
			"lastRunTime": rule.LastRunTime,
			"nextRunTime": rule.NextRunTime,
			"updatedAt": time.Now(),
		}},
	)
	if err != nil {
		log.Printf("更新调度规则统计失败: %v", err)
	}
}

// calculateNextRunTime 计算下次执行时间
func (s *TaskScheduler) calculateNextRunTime(rule *models.TaskScheduleRule) error {
	// 解析cron表达式
	schedule, err := cron.ParseStandard(rule.CronExpression)
	if err != nil {
		return fmt.Errorf("解析cron表达式失败: %v", err)
	}
	
	// 计算下次执行时间
	now := time.Now()
	if rule.Timezone != "" {
		// 处理时区
		location, err := time.LoadLocation(rule.Timezone)
		if err != nil {
			log.Printf("加载时区失败，使用UTC: %v", err)
		} else {
			now = now.In(location)
		}
	}
	
	nextTime := schedule.Next(now)
	rule.NextRunTime = nextTime
	
	return nil
}

// runPeriodicCheck 运行定期检查
func (s *TaskScheduler) runPeriodicCheck() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.performHealthCheck()
		}
	}
}

// performHealthCheck 执行健康检查
func (s *TaskScheduler) performHealthCheck() {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// 检查是否有遗漏的任务
	for ruleID, rule := range s.scheduleRules {
		if !rule.Enabled {
			continue
		}
		
		// 检查是否有对应的cron任务
		if _, exists := s.cronJobs[ruleID]; !exists {
			log.Printf("发现遗漏的cron任务，重新添加: %s", rule.Name)
			if err := s.addCronJob(rule); err != nil {
				log.Printf("重新添加cron任务失败: %v", err)
			}
		}
	}
}

// GetStats 获取调度器统计信息
func (s *TaskScheduler) GetStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	stats := map[string]interface{}{
		"total_rules":   len(s.scheduleRules),
		"active_rules":  0,
		"cron_jobs":     len(s.cronJobs),
		"next_runs":     make([]map[string]interface{}, 0),
	}
	
	for _, rule := range s.scheduleRules {
		if rule.Enabled {
			stats["active_rules"] = stats["active_rules"].(int) + 1
		}
		
		if !rule.NextRunTime.IsZero() {
			stats["next_runs"] = append(stats["next_runs"].([]map[string]interface{}), map[string]interface{}{
				"rule_name": rule.Name,
				"next_run":  rule.NextRunTime.Format("2006-01-02 15:04:05"),
			})
		}
	}
	
	return stats
}