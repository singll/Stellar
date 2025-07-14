package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// VulnDbService 漏洞数据库服务
type VulnDbService struct {
	db         *mongo.Database
	collection *mongo.Collection
	updateCol  *mongo.Collection
}

// NewVulnDbService 创建漏洞数据库服务
func NewVulnDbService(db *mongo.Database) *VulnDbService {
	service := &VulnDbService{
		db:         db,
		collection: db.Collection("vuln_database"),
		updateCol:  db.Collection("vuln_updates"),
	}
	
	// 创建索引
	service.createIndexes()
	
	return service
}

// createIndexes 创建数据库索引
func (s *VulnDbService) createIndexes() {
	ctx := context.Background()
	
	// 漏洞数据库索引
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"cve_id", 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys: bson.D{{"cwe_id", 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys: bson.D{{"cnvd_id", 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys: bson.D{{"severity", 1}},
		},
		{
			Keys: bson.D{{"category", 1}},
		},
		{
			Keys: bson.D{{"source", 1}},
		},
		{
			Keys: bson.D{{"published_date", -1}},
		},
		{
			Keys: bson.D{{"cvss_score", -1}},
		},
		{
			Keys: bson.D{
				{"title", "text"},
				{"description", "text"},
			},
		},
	}
	
	_, err := s.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("创建漏洞数据库索引失败: %v", err)
	}
	
	// 更新记录索引
	updateIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"source", 1}, {"start_time", -1}},
		},
		{
			Keys: bson.D{{"status", 1}},
		},
	}
	
	_, err = s.updateCol.Indexes().CreateMany(ctx, updateIndexes)
	if err != nil {
		log.Printf("创建更新记录索引失败: %v", err)
	}
}

// SearchVulnerabilities 搜索漏洞
func (s *VulnDbService) SearchVulnerabilities(ctx context.Context, query models.VulnDbQuery) ([]*models.VulnDbInfo, int64, error) {
	filter := s.buildSearchFilter(query)
	
	// 构建排序
	sort := bson.D{}
	if query.SortBy != "" {
		order := 1
		if query.SortDesc {
			order = -1
		}
		sort = bson.D{{query.SortBy, order}}
	} else {
		sort = bson.D{{"published_date", -1}} // 默认按发布时间排序
	}
	
	// 计算总数
	total, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("统计漏洞数量失败: %v", err)
	}
	
	// 分页设置
	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	
	// 查询数据
	opts := options.Find().
		SetSort(sort).
		SetSkip(skip).
		SetLimit(limit)
	
	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("查询漏洞失败: %v", err)
	}
	defer cursor.Close(ctx)
	
	var vulns []*models.VulnDbInfo
	if err := cursor.All(ctx, &vulns); err != nil {
		return nil, 0, fmt.Errorf("解析漏洞数据失败: %v", err)
	}
	
	return vulns, total, nil
}

// buildSearchFilter 构建搜索过滤条件
func (s *VulnDbService) buildSearchFilter(query models.VulnDbQuery) bson.M {
	filter := bson.M{}
	
	// 基本查询
	if query.CVEId != "" {
		filter["cve_id"] = query.CVEId
	}
	if query.CWEId != "" {
		filter["cwe_id"] = query.CWEId
	}
	if query.CNVDId != "" {
		filter["cnvd_id"] = query.CNVDId
	}
	
	// 关键词搜索
	if query.Keyword != "" {
		filter["$text"] = bson.M{"$search": query.Keyword}
	}
	
	// 过滤条件
	if len(query.Severity) > 0 {
		filter["severity"] = bson.M{"$in": query.Severity}
	}
	if len(query.Category) > 0 {
		filter["category"] = bson.M{"$in": query.Category}
	}
	if len(query.VulnType) > 0 {
		filter["vuln_type"] = bson.M{"$in": query.VulnType}
	}
	if len(query.Source) > 0 {
		filter["source"] = bson.M{"$in": query.Source}
	}
	if len(query.Tags) > 0 {
		filter["tags"] = bson.M{"$in": query.Tags}
	}
	
	// 状态过滤
	if query.Status != "" {
		filter["status"] = query.Status
	}
	if query.Verified != nil {
		filter["verified"] = *query.Verified
	}
	if query.HasPOC != nil {
		filter["has_poc"] = *query.HasPOC
	}
	if query.HasExploit != nil {
		filter["has_exploit"] = *query.HasExploit
	}
	
	// CVSS分数范围
	if query.MinCVSSScore != nil || query.MaxCVSSScore != nil {
		scoreFilter := bson.M{}
		if query.MinCVSSScore != nil {
			scoreFilter["$gte"] = *query.MinCVSSScore
		}
		if query.MaxCVSSScore != nil {
			scoreFilter["$lte"] = *query.MaxCVSSScore
		}
		filter["cvss_score"] = scoreFilter
	}
	
	// 时间范围
	if query.PublishedAfter != nil || query.PublishedBefore != nil {
		dateFilter := bson.M{}
		if query.PublishedAfter != nil {
			dateFilter["$gte"] = *query.PublishedAfter
		}
		if query.PublishedBefore != nil {
			dateFilter["$lte"] = *query.PublishedBefore
		}
		filter["published_date"] = dateFilter
	}
	
	if query.ModifiedAfter != nil || query.ModifiedBefore != nil {
		dateFilter := bson.M{}
		if query.ModifiedAfter != nil {
			dateFilter["$gte"] = *query.ModifiedAfter
		}
		if query.ModifiedBefore != nil {
			dateFilter["$lte"] = *query.ModifiedBefore
		}
		filter["modified_date"] = dateFilter
	}
	
	return filter
}

// GetVulnerabilityByID 根据ID获取漏洞信息
func (s *VulnDbService) GetVulnerabilityByID(ctx context.Context, id primitive.ObjectID) (*models.VulnDbInfo, error) {
	var vuln models.VulnDbInfo
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&vuln)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("漏洞不存在")
		}
		return nil, fmt.Errorf("获取漏洞信息失败: %v", err)
	}
	
	return &vuln, nil
}

// GetVulnerabilityByCVE 根据CVE编号获取漏洞信息
func (s *VulnDbService) GetVulnerabilityByCVE(ctx context.Context, cveId string) (*models.VulnDbInfo, error) {
	var vuln models.VulnDbInfo
	err := s.collection.FindOne(ctx, bson.M{"cve_id": cveId}).Decode(&vuln)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("CVE %s 不存在", cveId)
		}
		return nil, fmt.Errorf("获取CVE信息失败: %v", err)
	}
	
	return &vuln, nil
}

// GetStatistics 获取统计信息
func (s *VulnDbService) GetStatistics(ctx context.Context) (*models.VulnDbStats, error) {
	stats := &models.VulnDbStats{
		SeverityCount: make(map[string]int64),
		CategoryCount: make(map[string]int64),
		SourceCount:   make(map[string]int64),
	}
	
	// 总数统计
	total, err := s.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("统计总数失败: %v", err)
	}
	stats.TotalCount = total
	
	// 严重程度统计
	severityPipeline := mongo.Pipeline{
		{{"$group", bson.D{
			{"_id", "$severity"},
			{"count", bson.D{{"$sum", 1}}},
		}}},
	}
	cursor, err := s.collection.Aggregate(ctx, severityPipeline)
	if err != nil {
		return nil, fmt.Errorf("统计严重程度失败: %v", err)
	}
	defer cursor.Close(ctx)
	
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			stats.SeverityCount[result.ID] = result.Count
		}
	}
	
	// 类别统计
	categoryPipeline := mongo.Pipeline{
		{{"$group", bson.D{
			{"_id", "$category"},
			{"count", bson.D{{"$sum", 1}}},
		}}},
	}
	cursor, err = s.collection.Aggregate(ctx, categoryPipeline)
	if err != nil {
		return nil, fmt.Errorf("统计类别失败: %v", err)
	}
	defer cursor.Close(ctx)
	
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			stats.CategoryCount[result.ID] = result.Count
		}
	}
	
	// 数据源统计
	sourcePipeline := mongo.Pipeline{
		{{"$group", bson.D{
			{"_id", "$source"},
			{"count", bson.D{{"$sum", 1}}},
		}}},
	}
	cursor, err = s.collection.Aggregate(ctx, sourcePipeline)
	if err != nil {
		return nil, fmt.Errorf("统计数据源失败: %v", err)
	}
	defer cursor.Close(ctx)
	
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			stats.SourceCount[result.ID] = result.Count
		}
	}
	
	// 其他统计
	recentTime := time.Now().AddDate(0, 0, -30)
	recentCount, _ := s.collection.CountDocuments(ctx, bson.M{"created_at": bson.M{"$gte": recentTime}})
	stats.RecentCount = recentCount
	
	verifiedCount, _ := s.collection.CountDocuments(ctx, bson.M{"verified": true})
	stats.VerifiedCount = verifiedCount
	
	pocCount, _ := s.collection.CountDocuments(ctx, bson.M{"has_poc": true})
	stats.POCCount = pocCount
	
	exploitCount, _ := s.collection.CountDocuments(ctx, bson.M{"has_exploit": true})
	stats.ExploitCount = exploitCount
	
	// 最后更新时间
	var lastUpdate models.VulnUpdate
	opts := options.FindOne().SetSort(bson.D{{"end_time", -1}})
	err = s.updateCol.FindOne(ctx, bson.M{"status": "completed"}, opts).Decode(&lastUpdate)
	if err == nil {
		stats.LastUpdateTime = lastUpdate.EndTime
	}
	
	return stats, nil
}

// UpdateFromCVE 从CVE数据源更新
func (s *VulnDbService) UpdateFromCVE(ctx context.Context) error {
	updateRecord := &models.VulnUpdate{
		ID:         primitive.NewObjectID(),
		Source:     "CVE",
		UpdateType: "incremental",
		StartTime:  time.Now(),
		Status:     "running",
		Progress:   0,
		CreatedAt:  time.Now(),
	}
	
	// 插入更新记录
	_, err := s.updateCol.InsertOne(ctx, updateRecord)
	if err != nil {
		return fmt.Errorf("创建更新记录失败: %v", err)
	}
	
	// 启动后台更新
	go s.performCVEUpdate(updateRecord.ID)
	
	return nil
}

// performCVEUpdate 执行CVE数据更新
func (s *VulnDbService) performCVEUpdate(updateID primitive.ObjectID) {
	ctx := context.Background()
	
	defer func() {
		// 更新状态为完成
		filter := bson.M{"_id": updateID}
		update := bson.M{
			"$set": bson.M{
				"status":    "completed",
				"end_time":  time.Now(),
				"progress":  100,
			},
		}
		s.updateCol.UpdateOne(ctx, filter, update)
	}()
	
	// 获取CVE数据 (这里使用模拟数据，实际应该从CVE API获取)
	vulns := s.getMockCVEData()
	
	total := int64(len(vulns))
	processed := int64(0)
	newRecords := int64(0)
	updatedRecords := int64(0)
	errorRecords := int64(0)
	
	for _, vuln := range vulns {
		// 检查是否已存在
		existing, err := s.GetVulnerabilityByCVE(ctx, vuln.CVEId)
		if err == nil && existing != nil {
			// 更新现有记录
			filter := bson.M{"cve_id": vuln.CVEId}
			update := bson.M{
				"$set": bson.M{
					"title":          vuln.Title,
					"description":    vuln.Description,
					"severity":       vuln.Severity,
					"cvss_score":     vuln.CVSSScore,
					"cvss_vector":    vuln.CVSSVector,
					"modified_date":  vuln.ModifiedDate,
					"updated_at":     time.Now(),
					"version":        existing.Version + 1,
				},
			}
			_, err = s.collection.UpdateOne(ctx, filter, update)
			if err != nil {
				log.Printf("更新CVE %s 失败: %v", vuln.CVEId, err)
				errorRecords++
			} else {
				updatedRecords++
			}
		} else {
			// 插入新记录
			vuln.ID = primitive.NewObjectID()
			vuln.CreatedAt = time.Now()
			vuln.UpdatedAt = time.Now()
			vuln.Version = 1
			
			_, err = s.collection.InsertOne(ctx, vuln)
			if err != nil {
				log.Printf("插入CVE %s 失败: %v", vuln.CVEId, err)
				errorRecords++
			} else {
				newRecords++
			}
		}
		
		processed++
		progress := float64(processed) / float64(total) * 100
		
		// 更新进度
		filter := bson.M{"_id": updateID}
		update := bson.M{
			"$set": bson.M{
				"progress":         progress,
				"records_total":    total,
				"records_new":      newRecords,
				"records_updated":  updatedRecords,
				"records_error":    errorRecords,
			},
		}
		s.updateCol.UpdateOne(ctx, filter, update)
	}
}

// getMockCVEData 获取模拟CVE数据
func (s *VulnDbService) getMockCVEData() []*models.VulnDbInfo {
	return []*models.VulnDbInfo{
		{
			CVEId:         "CVE-2024-0001",
			Title:         "示例SQL注入漏洞",
			Description:   "Web应用程序中的SQL注入漏洞，允许攻击者执行任意SQL查询",
			Severity:      "high",
			CVSSScore:     8.5,
			CVSSVector:    "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:N",
			CVSSVersion:   "3.1",
			Category:      "injection",
			VulnType:      "sql_injection",
			Source:        "CVE",
			Status:        "active",
			PublishedDate: time.Now().AddDate(0, 0, -7),
			ModifiedDate:  time.Now().AddDate(0, 0, -1),
			References: []models.VulnReference{
				{
					Type:   "advisory",
					URL:    "https://example.com/advisory/CVE-2024-0001",
					Title:  "安全公告",
					Source: "vendor",
				},
			},
		},
		{
			CVEId:         "CVE-2024-0002", 
			Title:         "跨站脚本攻击漏洞",
			Description:   "Web应用程序存在存储型XSS漏洞",
			Severity:      "medium",
			CVSSScore:     6.1,
			CVSSVector:    "CVSS:3.1/AV:N/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N",
			CVSSVersion:   "3.1",
			Category:      "xss",
			VulnType:      "stored_xss",
			Source:        "CVE",
			Status:        "active",
			PublishedDate: time.Now().AddDate(0, 0, -5),
			ModifiedDate:  time.Now().AddDate(0, 0, -2),
		},
	}
}

// UpdateFromCNVD 从CNVD数据源更新
func (s *VulnDbService) UpdateFromCNVD(ctx context.Context) error {
	updateRecord := &models.VulnUpdate{
		ID:         primitive.NewObjectID(),
		Source:     "CNVD",
		UpdateType: "incremental",
		StartTime:  time.Now(),
		Status:     "running",
		Progress:   0,
		CreatedAt:  time.Now(),
	}
	
	_, err := s.updateCol.InsertOne(ctx, updateRecord)
	if err != nil {
		return fmt.Errorf("创建CNVD更新记录失败: %v", err)
	}
	
	go s.performCNVDUpdate(updateRecord.ID)
	
	return nil
}

// performCNVDUpdate 执行CNVD数据更新
func (s *VulnDbService) performCNVDUpdate(updateID primitive.ObjectID) {
	ctx := context.Background()
	
	defer func() {
		filter := bson.M{"_id": updateID}
		update := bson.M{
			"$set": bson.M{
				"status":    "completed",
				"end_time":  time.Now(),
				"progress":  100,
			},
		}
		s.updateCol.UpdateOne(ctx, filter, update)
	}()
	
	// 模拟CNVD数据更新
	log.Println("CNVD数据更新完成")
}

// GetUpdateHistory 获取更新历史
func (s *VulnDbService) GetUpdateHistory(ctx context.Context, page, pageSize int) ([]*models.VulnUpdate, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	total, err := s.updateCol.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, fmt.Errorf("统计更新记录失败: %v", err)
	}
	
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	
	opts := options.Find().
		SetSort(bson.D{{"start_time", -1}}).
		SetSkip(skip).
		SetLimit(limit)
	
	cursor, err := s.updateCol.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("查询更新记录失败: %v", err)
	}
	defer cursor.Close(ctx)
	
	var updates []*models.VulnUpdate
	if err := cursor.All(ctx, &updates); err != nil {
		return nil, 0, fmt.Errorf("解析更新记录失败: %v", err)
	}
	
	return updates, total, nil
}

// CreateCustomVulnerability 创建自定义漏洞
func (s *VulnDbService) CreateCustomVulnerability(ctx context.Context, vuln *models.VulnDbInfo) error {
	vuln.ID = primitive.NewObjectID()
	vuln.Source = "Custom"
	vuln.Status = "active"
	vuln.CreatedAt = time.Now()
	vuln.UpdatedAt = time.Now()
	vuln.Version = 1
	
	_, err := s.collection.InsertOne(ctx, vuln)
	if err != nil {
		return fmt.Errorf("创建自定义漏洞失败: %v", err)
	}
	
	return nil
}

// UpdateVulnerability 更新漏洞信息
func (s *VulnDbService) UpdateVulnerability(ctx context.Context, id primitive.ObjectID, vuln *models.VulnDbInfo) error {
	existing, err := s.GetVulnerabilityByID(ctx, id)
	if err != nil {
		return err
	}
	
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"title":        vuln.Title,
			"description":  vuln.Description,
			"severity":     vuln.Severity,
			"cvss_score":   vuln.CVSSScore,
			"category":     vuln.Category,
			"vuln_type":    vuln.VulnType,
			"solution":     vuln.Solution,
			"verified":     vuln.Verified,
			"has_poc":      vuln.HasPOC,
			"has_exploit":  vuln.HasExploit,
			"updated_at":   time.Now(),
			"version":      existing.Version + 1,
		},
	}
	
	_, err = s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("更新漏洞信息失败: %v", err)
	}
	
	return nil
}

// DeleteVulnerability 删除漏洞
func (s *VulnDbService) DeleteVulnerability(ctx context.Context, id primitive.ObjectID) error {
	result, err := s.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("删除漏洞失败: %v", err)
	}
	
	if result.DeletedCount == 0 {
		return fmt.Errorf("漏洞不存在")
	}
	
	return nil
}