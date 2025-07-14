package pagemonitoring

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AnalyticsService 监控分析服务
type AnalyticsService struct {
	db *mongo.Database
}

// MonitoringStats 监控统计
type MonitoringStats struct {
	// 基础统计
	TotalTasks      int64 `json:"total_tasks"`
	ActiveTasks     int64 `json:"active_tasks"`
	PausedTasks     int64 `json:"paused_tasks"`
	FailedTasks     int64 `json:"failed_tasks"`
	TotalSnapshots  int64 `json:"total_snapshots"`
	TotalChanges    int64 `json:"total_changes"`

	// 时间范围统计
	StatsForPeriod *PeriodStats `json:"stats_for_period"`

	// 任务执行统计
	TaskStats *TaskExecutionStats `json:"task_stats"`

	// 变更统计
	ChangeStats *ChangeAnalysisStats `json:"change_stats"`

	// 性能统计
	PerformanceStats *PerformanceStats `json:"performance_stats"`

	GeneratedAt time.Time `json:"generated_at"`
}

// PeriodStats 时间段统计
type PeriodStats struct {
	Period          string    `json:"period"` // "24h", "7d", "30d"
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	SnapshotCount   int64     `json:"snapshot_count"`
	ChangeCount     int64     `json:"change_count"`
	FailureCount    int64     `json:"failure_count"`
	AvgResponseTime float64   `json:"avg_response_time"`
	ChangeRate      float64   `json:"change_rate"` // 变更率
}

// TaskExecutionStats 任务执行统计
type TaskExecutionStats struct {
	TotalExecutions    int64                    `json:"total_executions"`
	SuccessfulRuns     int64                    `json:"successful_runs"`
	FailedRuns         int64                    `json:"failed_runs"`
	SuccessRate        float64                  `json:"success_rate"`
	AvgExecutionTime   float64                  `json:"avg_execution_time"`
	TopFailedTasks     []*TaskFailureInfo       `json:"top_failed_tasks"`
	ExecutionTrend     []*ExecutionTrendPoint   `json:"execution_trend"`
	HourlyDistribution []*HourlyExecutionStats  `json:"hourly_distribution"`
}

// TaskFailureInfo 任务失败信息
type TaskFailureInfo struct {
	TaskID       primitive.ObjectID `json:"task_id"`
	TaskName     string             `json:"task_name"`
	URL          string             `json:"url"`
	FailureCount int64              `json:"failure_count"`
	LastFailure  time.Time          `json:"last_failure"`
	LastError    string             `json:"last_error"`
}

// ExecutionTrendPoint 执行趋势点
type ExecutionTrendPoint struct {
	Time         time.Time `json:"time"`
	SuccessCount int64     `json:"success_count"`
	FailureCount int64     `json:"failure_count"`
	AvgTime      float64   `json:"avg_time"`
}

// HourlyExecutionStats 小时执行统计
type HourlyExecutionStats struct {
	Hour         int   `json:"hour"`
	ExecutionCount int64 `json:"execution_count"`
	SuccessCount int64 `json:"success_count"`
	FailureCount int64 `json:"failure_count"`
}

// ChangeAnalysisStats 变更分析统计
type ChangeAnalysisStats struct {
	TotalChanges        int64                 `json:"total_changes"`
	CriticalChanges     int64                 `json:"critical_changes"`   // 相似度 < 0.3
	MajorChanges        int64                 `json:"major_changes"`      // 相似度 < 0.7
	MinorChanges        int64                 `json:"minor_changes"`      // 相似度 >= 0.7
	AvgSimilarity       float64               `json:"avg_similarity"`
	MostChangedSites    []*SiteChangeInfo     `json:"most_changed_sites"`
	ChangeDistribution  []*ChangeDistribution `json:"change_distribution"`
	SimilarityTrend     []*SimilarityTrend    `json:"similarity_trend"`
}

// SiteChangeInfo 站点变更信息
type SiteChangeInfo struct {
	URL           string    `json:"url"`
	TaskName      string    `json:"task_name"`
	ChangeCount   int64     `json:"change_count"`
	LastChange    time.Time `json:"last_change"`
	AvgSimilarity float64   `json:"avg_similarity"`
}

// ChangeDistribution 变更分布
type ChangeDistribution struct {
	SimilarityRange string `json:"similarity_range"`
	Count           int64  `json:"count"`
	Percentage      float64 `json:"percentage"`
}

// SimilarityTrend 相似度趋势
type SimilarityTrend struct {
	Time          time.Time `json:"time"`
	AvgSimilarity float64   `json:"avg_similarity"`
	ChangeCount   int64     `json:"change_count"`
}

// PerformanceStats 性能统计
type PerformanceStats struct {
	AvgLoadTime    float64                `json:"avg_load_time"`
	MinLoadTime    float64                `json:"min_load_time"`
	MaxLoadTime    float64                `json:"max_load_time"`
	P95LoadTime    float64                `json:"p95_load_time"`
	P99LoadTime    float64                `json:"p99_load_time"`
	SlowestSites   []*SitePerformanceInfo `json:"slowest_sites"`
	FastestSites   []*SitePerformanceInfo `json:"fastest_sites"`
	LoadTimeTrend  []*LoadTimeTrend       `json:"load_time_trend"`
}

// SitePerformanceInfo 站点性能信息
type SitePerformanceInfo struct {
	URL         string  `json:"url"`
	TaskName    string  `json:"task_name"`
	AvgLoadTime float64 `json:"avg_load_time"`
	MinLoadTime float64 `json:"min_load_time"`
	MaxLoadTime float64 `json:"max_load_time"`
	SampleCount int64   `json:"sample_count"`
}

// LoadTimeTrend 加载时间趋势
type LoadTimeTrend struct {
	Time        time.Time `json:"time"`
	AvgLoadTime float64   `json:"avg_load_time"`
	SampleCount int64     `json:"sample_count"`
}

// TaskDetailStats 任务详细统计
type TaskDetailStats struct {
	TaskID           primitive.ObjectID `json:"task_id"`
	TaskName         string             `json:"task_name"`
	URL              string             `json:"url"`
	Status           string             `json:"status"`
	CreatedAt        time.Time          `json:"created_at"`
	LastRunAt        *time.Time         `json:"last_run_at"`
	NextRunAt        *time.Time         `json:"next_run_at"`
	
	// 执行统计
	TotalRuns        int64     `json:"total_runs"`
	SuccessfulRuns   int64     `json:"successful_runs"`
	FailedRuns       int64     `json:"failed_runs"`
	SuccessRate      float64   `json:"success_rate"`
	AvgExecutionTime float64   `json:"avg_execution_time"`
	
	// 变更统计
	TotalChanges     int64     `json:"total_changes"`
	LastChangeAt     *time.Time `json:"last_change_at"`
	AvgSimilarity    float64   `json:"avg_similarity"`
	
	// 性能统计
	AvgLoadTime      float64   `json:"avg_load_time"`
	MinLoadTime      float64   `json:"min_load_time"`
	MaxLoadTime      float64   `json:"max_load_time"`
	
	// 最近执行记录
	RecentExecutions []*ExecutionRecord `json:"recent_executions"`
	
	// 最近变更记录
	RecentChanges    []*ChangeRecord    `json:"recent_changes"`
}

// ExecutionRecord 执行记录
type ExecutionRecord struct {
	Timestamp    time.Time `json:"timestamp"`
	Success      bool      `json:"success"`
	ExecutionTime float64  `json:"execution_time"`
	LoadTime     float64   `json:"load_time"`
	Error        string    `json:"error,omitempty"`
}

// ChangeRecord 变更记录
type ChangeRecord struct {
	Timestamp  time.Time `json:"timestamp"`
	Similarity float64   `json:"similarity"`
	Status     string    `json:"status"`
	DiffPreview string   `json:"diff_preview"`
}

// NewAnalyticsService 创建分析服务
func NewAnalyticsService(db *mongo.Database) *AnalyticsService {
	return &AnalyticsService{
		db: db,
	}
}

// GetMonitoringStats 获取监控统计
func (s *AnalyticsService) GetMonitoringStats(period string) (*MonitoringStats, error) {
	stats := &MonitoringStats{
		GeneratedAt: time.Now(),
	}

	// 获取基础统计
	if err := s.getBasicStats(stats); err != nil {
		return nil, fmt.Errorf("获取基础统计失败: %v", err)
	}

	// 获取时间段统计
	periodStats, err := s.getPeriodStats(period)
	if err != nil {
		return nil, fmt.Errorf("获取时间段统计失败: %v", err)
	}
	stats.StatsForPeriod = periodStats

	// 获取任务执行统计
	taskStats, err := s.getTaskExecutionStats(period)
	if err != nil {
		return nil, fmt.Errorf("获取任务执行统计失败: %v", err)
	}
	stats.TaskStats = taskStats

	// 获取变更分析统计
	changeStats, err := s.getChangeAnalysisStats(period)
	if err != nil {
		return nil, fmt.Errorf("获取变更分析统计失败: %v", err)
	}
	stats.ChangeStats = changeStats

	// 获取性能统计
	performanceStats, err := s.getPerformanceStats(period)
	if err != nil {
		return nil, fmt.Errorf("获取性能统计失败: %v", err)
	}
	stats.PerformanceStats = performanceStats

	return stats, nil
}

// getBasicStats 获取基础统计
func (s *AnalyticsService) getBasicStats(stats *MonitoringStats) error {
	ctx := context.Background()

	// 统计任务数量
	tasksCollection := s.db.Collection("monitoring_tasks")
	
	totalTasks, err := tasksCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return err
	}
	stats.TotalTasks = totalTasks

	activeTasks, err := tasksCollection.CountDocuments(ctx, bson.M{"status": models.MonitoringTaskStatusActive})
	if err != nil {
		return err
	}
	stats.ActiveTasks = activeTasks

	pausedTasks, err := tasksCollection.CountDocuments(ctx, bson.M{"status": models.MonitoringTaskStatusPaused})
	if err != nil {
		return err
	}
	stats.PausedTasks = pausedTasks

	failedTasks, err := tasksCollection.CountDocuments(ctx, bson.M{"status": models.MonitoringTaskStatusFailed})
	if err != nil {
		return err
	}
	stats.FailedTasks = failedTasks

	// 统计快照数量
	snapshotsCollection := s.db.Collection("page_snapshots")
	totalSnapshots, err := snapshotsCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return err
	}
	stats.TotalSnapshots = totalSnapshots

	// 统计变更数量
	changesCollection := s.db.Collection("page_changes")
	totalChanges, err := changesCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return err
	}
	stats.TotalChanges = totalChanges

	return nil
}

// getPeriodStats 获取时间段统计
func (s *AnalyticsService) getPeriodStats(period string) (*PeriodStats, error) {
	endTime := time.Now()
	var startTime time.Time

	switch period {
	case "24h":
		startTime = endTime.Add(-24 * time.Hour)
	case "7d":
		startTime = endTime.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = endTime.Add(-30 * 24 * time.Hour)
	default:
		startTime = endTime.Add(-24 * time.Hour)
		period = "24h"
	}

	stats := &PeriodStats{
		Period:    period,
		StartTime: startTime,
		EndTime:   endTime,
	}

	ctx := context.Background()
	timeFilter := bson.M{
		"created_at": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	// 统计快照数量
	snapshotsCollection := s.db.Collection("page_snapshots")
	snapshotCount, err := snapshotsCollection.CountDocuments(ctx, timeFilter)
	if err != nil {
		return nil, err
	}
	stats.SnapshotCount = snapshotCount

	// 统计变更数量
	changesCollection := s.db.Collection("page_changes")
	changeCount, err := changesCollection.CountDocuments(ctx, bson.M{
		"changed_at": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	})
	if err != nil {
		return nil, err
	}
	stats.ChangeCount = changeCount

	// 统计失败数量（从任务更新时间判断）
	tasksCollection := s.db.Collection("monitoring_tasks")
	failureCount, err := tasksCollection.CountDocuments(ctx, bson.M{
		"status": models.MonitoringTaskStatusFailed,
		"updated_at": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	})
	if err != nil {
		return nil, err
	}
	stats.FailureCount = failureCount

	// 计算平均响应时间
	pipeline := []bson.M{
		{"$match": timeFilter},
		{"$group": bson.M{
			"_id": nil,
			"avg_load_time": bson.M{"$avg": "$load_time"},
		}},
	}

	cursor, err := snapshotsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			AvgLoadTime float64 `bson:"avg_load_time"`
		}
		if err := cursor.Decode(&result); err == nil {
			stats.AvgResponseTime = result.AvgLoadTime
		}
	}

	// 计算变更率
	if stats.SnapshotCount > 0 {
		stats.ChangeRate = float64(stats.ChangeCount) / float64(stats.SnapshotCount) * 100
	}

	return stats, nil
}

// getTaskExecutionStats 获取任务执行统计
func (s *AnalyticsService) getTaskExecutionStats(period string) (*TaskExecutionStats, error) {
	endTime := time.Now()
	var startTime time.Time

	switch period {
	case "24h":
		startTime = endTime.Add(-24 * time.Hour)
	case "7d":
		startTime = endTime.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = endTime.Add(-30 * 24 * time.Hour)
	default:
		startTime = endTime.Add(-24 * time.Hour)
	}

	ctx := context.Background()
	stats := &TaskExecutionStats{}

	// 从快照集合统计执行情况（每个快照代表一次执行）
	snapshotsCollection := s.db.Collection("page_snapshots")
	timeFilter := bson.M{
		"created_at": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	totalExecutions, err := snapshotsCollection.CountDocuments(ctx, timeFilter)
	if err != nil {
		return nil, err
	}
	stats.TotalExecutions = totalExecutions

	// 成功执行数 = 快照数量（能创建快照说明执行成功）
	stats.SuccessfulRuns = totalExecutions

	// 失败执行数（从任务失败记录统计）
	tasksCollection := s.db.Collection("monitoring_tasks")
	failedRuns, err := tasksCollection.CountDocuments(ctx, bson.M{
		"status": models.MonitoringTaskStatusFailed,
		"updated_at": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	})
	if err != nil {
		return nil, err
	}
	stats.FailedRuns = failedRuns

	// 计算成功率
	totalRuns := stats.SuccessfulRuns + stats.FailedRuns
	if totalRuns > 0 {
		stats.SuccessRate = float64(stats.SuccessfulRuns) / float64(totalRuns) * 100
	}

	// 计算平均执行时间（使用加载时间作为近似）
	pipeline := []bson.M{
		{"$match": timeFilter},
		{"$group": bson.M{
			"_id": nil,
			"avg_execution_time": bson.M{"$avg": "$load_time"},
		}},
	}

	cursor, err := snapshotsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			AvgExecutionTime float64 `bson:"avg_execution_time"`
		}
		if err := cursor.Decode(&result); err == nil {
			stats.AvgExecutionTime = result.AvgExecutionTime
		}
	}
	cursor.Close(ctx)

	// 获取失败最多的任务
	topFailedTasks, err := s.getTopFailedTasks(startTime, endTime)
	if err != nil {
		return nil, err
	}
	stats.TopFailedTasks = topFailedTasks

	// 获取执行趋势
	executionTrend, err := s.getExecutionTrend(startTime, endTime)
	if err != nil {
		return nil, err
	}
	stats.ExecutionTrend = executionTrend

	// 获取小时分布
	hourlyDistribution, err := s.getHourlyDistribution(startTime, endTime)
	if err != nil {
		return nil, err
	}
	stats.HourlyDistribution = hourlyDistribution

	return stats, nil
}

// getTopFailedTasks 获取失败最多的任务
func (s *AnalyticsService) getTopFailedTasks(startTime, endTime time.Time) ([]*TaskFailureInfo, error) {
	ctx := context.Background()
	tasksCollection := s.db.Collection("monitoring_tasks")

	// 查询失败的任务
	filter := bson.M{
		"status": models.MonitoringTaskStatusFailed,
		"updated_at": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	cursor, err := tasksCollection.Find(ctx, filter, &options.FindOptions{
		Limit: ptr(int64(10)),
		Sort:  bson.M{"updated_at": -1},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var failedTasks []*TaskFailureInfo
	for cursor.Next(ctx) {
		var task models.MonitoringTask
		if err := cursor.Decode(&task); err != nil {
			continue
		}

		failureInfo := &TaskFailureInfo{
			TaskID:       task.ID,
			TaskName:     task.Name,
			URL:          task.URL,
			FailureCount: 1, // 简化实现，实际应该统计具体失败次数
			LastFailure:  task.UpdatedAt,
			LastError:    task.Message,
		}
		failedTasks = append(failedTasks, failureInfo)
	}

	return failedTasks, nil
}

// getExecutionTrend 获取执行趋势
func (s *AnalyticsService) getExecutionTrend(startTime, endTime time.Time) ([]*ExecutionTrendPoint, error) {
	ctx := context.Background()
	snapshotsCollection := s.db.Collection("page_snapshots")

	// 按小时分组统计
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{
					"$gte": startTime,
					"$lte": endTime,
				},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"year":  bson.M{"$year": "$created_at"},
					"month": bson.M{"$month": "$created_at"},
					"day":   bson.M{"$dayOfMonth": "$created_at"},
					"hour":  bson.M{"$hour": "$created_at"},
				},
				"success_count": bson.M{"$sum": 1},
				"avg_time":     bson.M{"$avg": "$load_time"},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	cursor, err := snapshotsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var trends []*ExecutionTrendPoint
	for cursor.Next(ctx) {
		var result struct {
			ID struct {
				Year  int `bson:"year"`
				Month int `bson:"month"`
				Day   int `bson:"day"`
				Hour  int `bson:"hour"`
			} `bson:"_id"`
			SuccessCount int64   `bson:"success_count"`
			AvgTime      float64 `bson:"avg_time"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		point := &ExecutionTrendPoint{
			Time:         time.Date(result.ID.Year, time.Month(result.ID.Month), result.ID.Day, result.ID.Hour, 0, 0, 0, time.UTC),
			SuccessCount: result.SuccessCount,
			FailureCount: 0, // 简化实现
			AvgTime:      result.AvgTime,
		}
		trends = append(trends, point)
	}

	return trends, nil
}

// getHourlyDistribution 获取小时分布
func (s *AnalyticsService) getHourlyDistribution(startTime, endTime time.Time) ([]*HourlyExecutionStats, error) {
	ctx := context.Background()
	snapshotsCollection := s.db.Collection("page_snapshots")

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{
					"$gte": startTime,
					"$lte": endTime,
				},
			},
		},
		{
			"$group": bson.M{
				"_id":            bson.M{"$hour": "$created_at"},
				"execution_count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	cursor, err := snapshotsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	hourlyStats := make([]*HourlyExecutionStats, 24)
	for i := 0; i < 24; i++ {
		hourlyStats[i] = &HourlyExecutionStats{
			Hour: i,
		}
	}

	for cursor.Next(ctx) {
		var result struct {
			Hour           int   `bson:"_id"`
			ExecutionCount int64 `bson:"execution_count"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		if result.Hour >= 0 && result.Hour < 24 {
			hourlyStats[result.Hour].ExecutionCount = result.ExecutionCount
			hourlyStats[result.Hour].SuccessCount = result.ExecutionCount // 简化实现
		}
	}

	return hourlyStats, nil
}

// getChangeAnalysisStats 获取变更分析统计
func (s *AnalyticsService) getChangeAnalysisStats(period string) (*ChangeAnalysisStats, error) {
	endTime := time.Now()
	var startTime time.Time

	switch period {
	case "24h":
		startTime = endTime.Add(-24 * time.Hour)
	case "7d":
		startTime = endTime.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = endTime.Add(-30 * 24 * time.Hour)
	default:
		startTime = endTime.Add(-24 * time.Hour)
	}

	ctx := context.Background()
	changesCollection := s.db.Collection("page_changes")
	
	timeFilter := bson.M{
		"changed_at": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	stats := &ChangeAnalysisStats{}

	// 统计总变更数
	totalChanges, err := changesCollection.CountDocuments(ctx, timeFilter)
	if err != nil {
		return nil, err
	}
	stats.TotalChanges = totalChanges

	// 按相似度分类统计
	criticalFilter := bson.M{
		"changed_at": bson.M{"$gte": startTime, "$lte": endTime},
		"similarity": bson.M{"$lt": 0.3},
	}
	criticalChanges, err := changesCollection.CountDocuments(ctx, criticalFilter)
	if err != nil {
		return nil, err
	}
	stats.CriticalChanges = criticalChanges

	majorFilter := bson.M{
		"changed_at": bson.M{"$gte": startTime, "$lte": endTime},
		"similarity": bson.M{"$gte": 0.3, "$lt": 0.7},
	}
	majorChanges, err := changesCollection.CountDocuments(ctx, majorFilter)
	if err != nil {
		return nil, err
	}
	stats.MajorChanges = majorChanges

	minorFilter := bson.M{
		"changed_at": bson.M{"$gte": startTime, "$lte": endTime},
		"similarity": bson.M{"$gte": 0.7},
	}
	minorChanges, err := changesCollection.CountDocuments(ctx, minorFilter)
	if err != nil {
		return nil, err
	}
	stats.MinorChanges = minorChanges

	// 计算平均相似度
	pipeline := []bson.M{
		{"$match": timeFilter},
		{"$group": bson.M{
			"_id": nil,
			"avg_similarity": bson.M{"$avg": "$similarity"},
		}},
	}

	cursor, err := changesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			AvgSimilarity float64 `bson:"avg_similarity"`
		}
		if err := cursor.Decode(&result); err == nil {
			stats.AvgSimilarity = result.AvgSimilarity
		}
	}
	cursor.Close(ctx)

	// 获取变更最多的站点
	mostChangedSites, err := s.getMostChangedSites(startTime, endTime)
	if err != nil {
		return nil, err
	}
	stats.MostChangedSites = mostChangedSites

	// 生成变更分布
	stats.ChangeDistribution = []*ChangeDistribution{
		{
			SimilarityRange: "0.0-0.3 (关键变更)",
			Count:           stats.CriticalChanges,
			Percentage:      float64(stats.CriticalChanges) / float64(stats.TotalChanges) * 100,
		},
		{
			SimilarityRange: "0.3-0.7 (重要变更)",
			Count:           stats.MajorChanges,
			Percentage:      float64(stats.MajorChanges) / float64(stats.TotalChanges) * 100,
		},
		{
			SimilarityRange: "0.7-1.0 (轻微变更)",
			Count:           stats.MinorChanges,
			Percentage:      float64(stats.MinorChanges) / float64(stats.TotalChanges) * 100,
		},
	}

	return stats, nil
}

// getMostChangedSites 获取变更最多的站点
func (s *AnalyticsService) getMostChangedSites(startTime, endTime time.Time) ([]*SiteChangeInfo, error) {
	ctx := context.Background()
	changesCollection := s.db.Collection("page_changes")

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"changed_at": bson.M{
					"$gte": startTime,
					"$lte": endTime,
				},
			},
		},
		{
			"$group": bson.M{
				"_id":            "$url",
				"change_count":   bson.M{"$sum": 1},
				"avg_similarity": bson.M{"$avg": "$similarity"},
				"last_change":    bson.M{"$max": "$changed_at"},
			},
		},
		{
			"$sort": bson.M{"change_count": -1},
		},
		{
			"$limit": 10,
		},
	}

	cursor, err := changesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sites []*SiteChangeInfo
	for cursor.Next(ctx) {
		var result struct {
			URL           string    `bson:"_id"`
			ChangeCount   int64     `bson:"change_count"`
			AvgSimilarity float64   `bson:"avg_similarity"`
			LastChange    time.Time `bson:"last_change"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		site := &SiteChangeInfo{
			URL:           result.URL,
			TaskName:      result.URL, // 简化实现，实际应该关联任务名称
			ChangeCount:   result.ChangeCount,
			LastChange:    result.LastChange,
			AvgSimilarity: result.AvgSimilarity,
		}
		sites = append(sites, site)
	}

	return sites, nil
}

// getPerformanceStats 获取性能统计
func (s *AnalyticsService) getPerformanceStats(period string) (*PerformanceStats, error) {
	endTime := time.Now()
	var startTime time.Time

	switch period {
	case "24h":
		startTime = endTime.Add(-24 * time.Hour)
	case "7d":
		startTime = endTime.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = endTime.Add(-30 * 24 * time.Hour)
	default:
		startTime = endTime.Add(-24 * time.Hour)
	}

	ctx := context.Background()
	snapshotsCollection := s.db.Collection("page_snapshots")
	
	timeFilter := bson.M{
		"created_at": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	stats := &PerformanceStats{}

	// 获取基础性能统计
	pipeline := []bson.M{
		{"$match": timeFilter},
		{"$group": bson.M{
			"_id":          nil,
			"avg_load_time": bson.M{"$avg": "$load_time"},
			"min_load_time": bson.M{"$min": "$load_time"},
			"max_load_time": bson.M{"$max": "$load_time"},
			"load_times":   bson.M{"$push": "$load_time"},
		}},
	}

	cursor, err := snapshotsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			AvgLoadTime float64   `bson:"avg_load_time"`
			MinLoadTime float64   `bson:"min_load_time"`
			MaxLoadTime float64   `bson:"max_load_time"`
			LoadTimes   []float64 `bson:"load_times"`
		}

		if err := cursor.Decode(&result); err == nil {
			stats.AvgLoadTime = result.AvgLoadTime
			stats.MinLoadTime = result.MinLoadTime
			stats.MaxLoadTime = result.MaxLoadTime

			// 计算百分位数
			if len(result.LoadTimes) > 0 {
				sort.Float64s(result.LoadTimes)
				stats.P95LoadTime = percentile(result.LoadTimes, 95)
				stats.P99LoadTime = percentile(result.LoadTimes, 99)
			}
		}
	}
	cursor.Close(ctx)

	// 获取最慢和最快的站点
	slowestSites, err := s.getSitesByPerformance(startTime, endTime, false)
	if err != nil {
		return nil, err
	}
	stats.SlowestSites = slowestSites

	fastestSites, err := s.getSitesByPerformance(startTime, endTime, true)
	if err != nil {
		return nil, err
	}
	stats.FastestSites = fastestSites

	return stats, nil
}

// getSitesByPerformance 按性能获取站点
func (s *AnalyticsService) getSitesByPerformance(startTime, endTime time.Time, fastest bool) ([]*SitePerformanceInfo, error) {
	ctx := context.Background()
	snapshotsCollection := s.db.Collection("page_snapshots")

	sortOrder := -1 // 默认降序（最慢）
	if fastest {
		sortOrder = 1 // 升序（最快）
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{
					"$gte": startTime,
					"$lte": endTime,
				},
			},
		},
		{
			"$group": bson.M{
				"_id":            "$url",
				"avg_load_time":  bson.M{"$avg": "$load_time"},
				"min_load_time":  bson.M{"$min": "$load_time"},
				"max_load_time":  bson.M{"$max": "$load_time"},
				"sample_count":   bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"avg_load_time": sortOrder},
		},
		{
			"$limit": 5,
		},
	}

	cursor, err := snapshotsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sites []*SitePerformanceInfo
	for cursor.Next(ctx) {
		var result struct {
			URL         string  `bson:"_id"`
			AvgLoadTime float64 `bson:"avg_load_time"`
			MinLoadTime float64 `bson:"min_load_time"`
			MaxLoadTime float64 `bson:"max_load_time"`
			SampleCount int64   `bson:"sample_count"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		site := &SitePerformanceInfo{
			URL:         result.URL,
			TaskName:    result.URL, // 简化实现
			AvgLoadTime: result.AvgLoadTime,
			MinLoadTime: result.MinLoadTime,
			MaxLoadTime: result.MaxLoadTime,
			SampleCount: result.SampleCount,
		}
		sites = append(sites, site)
	}

	return sites, nil
}

// GetTaskDetailStats 获取任务详细统计
func (s *AnalyticsService) GetTaskDetailStats(taskID primitive.ObjectID) (*TaskDetailStats, error) {
	ctx := context.Background()

	// 获取任务基本信息
	tasksCollection := s.db.Collection("monitoring_tasks")
	var task models.MonitoringTask
	err := tasksCollection.FindOne(ctx, bson.M{"_id": taskID}).Decode(&task)
	if err != nil {
		return nil, fmt.Errorf("获取任务信息失败: %v", err)
	}

	stats := &TaskDetailStats{
		TaskID:    task.ID,
		TaskName:  task.Name,
		URL:       task.URL,
		Status:    string(task.Status),
		CreatedAt: task.CreatedAt,
		LastRunAt: &task.LastRunAt, // 使用指针
	}

	// 统计执行记录
	snapshotsCollection := s.db.Collection("page_snapshots")
	successfulRuns, err := snapshotsCollection.CountDocuments(ctx, bson.M{"task_id": taskID})
	if err != nil {
		return nil, err
	}
	stats.SuccessfulRuns = successfulRuns
	stats.TotalRuns = successfulRuns // 简化实现

	if stats.TotalRuns > 0 {
		stats.SuccessRate = float64(stats.SuccessfulRuns) / float64(stats.TotalRuns) * 100
	}

	// 计算平均执行时间
	pipeline := []bson.M{
		{"$match": bson.M{"task_id": taskID}},
		{"$group": bson.M{
			"_id":              nil,
			"avg_execution_time": bson.M{"$avg": "$load_time"},
			"avg_load_time":    bson.M{"$avg": "$load_time"},
			"min_load_time":    bson.M{"$min": "$load_time"},
			"max_load_time":    bson.M{"$max": "$load_time"},
		}},
	}

	cursor, err := snapshotsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			AvgExecutionTime float64 `bson:"avg_execution_time"`
			AvgLoadTime      float64 `bson:"avg_load_time"`
			MinLoadTime      float64 `bson:"min_load_time"`
			MaxLoadTime      float64 `bson:"max_load_time"`
		}
		if err := cursor.Decode(&result); err == nil {
			stats.AvgExecutionTime = result.AvgExecutionTime
			stats.AvgLoadTime = result.AvgLoadTime
			stats.MinLoadTime = result.MinLoadTime
			stats.MaxLoadTime = result.MaxLoadTime
		}
	}
	cursor.Close(ctx)

	// 统计变更记录
	changesCollection := s.db.Collection("page_changes")
	totalChanges, err := changesCollection.CountDocuments(ctx, bson.M{"task_id": taskID})
	if err != nil {
		return nil, err
	}
	stats.TotalChanges = totalChanges

	// 获取最近的执行记录
	recentExecutions, err := s.getRecentExecutions(taskID, 10)
	if err != nil {
		return nil, err
	}
	stats.RecentExecutions = recentExecutions

	// 获取最近的变更记录
	recentChanges, err := s.getRecentChanges(taskID, 10)
	if err != nil {
		return nil, err
	}
	stats.RecentChanges = recentChanges

	return stats, nil
}

// getRecentExecutions 获取最近的执行记录
func (s *AnalyticsService) getRecentExecutions(taskID primitive.ObjectID, limit int) ([]*ExecutionRecord, error) {
	ctx := context.Background()
	snapshotsCollection := s.db.Collection("page_snapshots")

	cursor, err := snapshotsCollection.Find(ctx, bson.M{"task_id": taskID}, &options.FindOptions{
		Limit: ptr(int64(limit)),
		Sort:  bson.M{"created_at": -1},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var executions []*ExecutionRecord
	for cursor.Next(ctx) {
		var snapshot models.PageSnapshot
		if err := cursor.Decode(&snapshot); err != nil {
			continue
		}

		execution := &ExecutionRecord{
			Timestamp:     snapshot.CreatedAt,
			Success:       true,
			ExecutionTime: float64(snapshot.LoadTime),
			LoadTime:      float64(snapshot.LoadTime),
		}
		executions = append(executions, execution)
	}

	return executions, nil
}

// getRecentChanges 获取最近的变更记录
func (s *AnalyticsService) getRecentChanges(taskID primitive.ObjectID, limit int) ([]*ChangeRecord, error) {
	ctx := context.Background()
	changesCollection := s.db.Collection("page_changes")

	cursor, err := changesCollection.Find(ctx, bson.M{"task_id": taskID}, &options.FindOptions{
		Limit: ptr(int64(limit)),
		Sort:  bson.M{"changed_at": -1},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var changes []*ChangeRecord
	for cursor.Next(ctx) {
		var change models.PageChange
		if err := cursor.Decode(&change); err != nil {
			continue
		}

		record := &ChangeRecord{
			Timestamp:   change.ChangedAt,
			Similarity:  change.Similarity,
			Status:      string(change.Status),
			DiffPreview: s.truncateString(change.Diff, 200),
		}
		changes = append(changes, record)
	}

	return changes, nil
}

// truncateString 截断字符串
func (s *AnalyticsService) truncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen] + "..."
}

// percentile 计算百分位数
func percentile(data []float64, p float64) float64 {
	if len(data) == 0 {
		return 0
	}
	
	index := p / 100.0 * float64(len(data)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	
	if lower == upper {
		return data[lower]
	}
	
	weight := index - float64(lower)
	return data[lower]*(1-weight) + data[upper]*weight
}

// ptr 创建指针
func ptr(i int64) *int64 {
	return &i
}