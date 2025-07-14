package vulndb

import (
	"context"
	"fmt"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Service 漏洞数据库服务
type Service struct {
	db         *mongo.Database
	cveClient  *CVEClient
	cweClient  *CWEClient
	cnvdClient *CNVDClient
}

// Config 服务配置
type Config struct {
	UpdateInterval time.Duration `json:"update_interval"`
	CVEConfig      CVEConfig     `json:"cve_config"`
	CWEConfig      CWEConfig     `json:"cwe_config"`
	CNVDConfig     CNVDConfig    `json:"cnvd_config"`
}

// CVEConfig CVE数据源配置
type CVEConfig struct {
	APIURL      string        `json:"api_url"`
	APIKey      string        `json:"api_key"`
	Timeout     time.Duration `json:"timeout"`
	BatchSize   int           `json:"batch_size"`
	RateLimit   int           `json:"rate_limit"`
	LastUpdated time.Time     `json:"last_updated"`
}

// CWEConfig CWE数据源配置
type CWEConfig struct {
	XMLURL      string        `json:"xml_url"`
	Timeout     time.Duration `json:"timeout"`
	LastUpdated time.Time     `json:"last_updated"`
}

// CNVDConfig CNVD数据源配置
type CNVDConfig struct {
	APIURL      string        `json:"api_url"`
	APIKey      string        `json:"api_key"`
	Timeout     time.Duration `json:"timeout"`
	BatchSize   int           `json:"batch_size"`
	LastUpdated time.Time     `json:"last_updated"`
}

// NewService 创建漏洞数据库服务
func NewService(db *mongo.Database, config Config) *Service {
	return &Service{
		db:         db,
		cveClient:  NewCVEClient(config.CVEConfig),
		cweClient:  NewCWEClient(config.CWEConfig),
		cnvdClient: NewCNVDClient(config.CNVDConfig),
	}
}

// UpdateAllDatabases 更新所有漏洞数据库
func (s *Service) UpdateAllDatabases(ctx context.Context) error {
	// TODO: 实现并发更新所有数据源
	
	// 更新CVE数据库
	if err := s.UpdateCVEDatabase(ctx); err != nil {
		return fmt.Errorf("更新CVE数据库失败: %w", err)
	}

	// 更新CWE数据库
	if err := s.UpdateCWEDatabase(ctx); err != nil {
		return fmt.Errorf("更新CWE数据库失败: %w", err)
	}

	// 更新CNVD数据库
	if err := s.UpdateCNVDDatabase(ctx); err != nil {
		return fmt.Errorf("更新CNVD数据库失败: %w", err)
	}

	return nil
}

// UpdateCVEDatabase 更新CVE数据库
func (s *Service) UpdateCVEDatabase(ctx context.Context) error {
	// TODO: 实现CVE数据库更新逻辑
	updateRecord := &models.VulnUpdate{
		Source:     "CVE",
		UpdateType: "incremental",
		StartTime:  time.Now(),
		Status:     "running",
	}

	// 插入更新记录
	collection := s.db.Collection("vuln_updates")
	result, err := collection.InsertOne(ctx, updateRecord)
	if err != nil {
		return err
	}
	updateRecord.ID = result.InsertedID.(primitive.ObjectID)

	// 获取CVE数据
	cveData, err := s.cveClient.FetchRecentCVEs(ctx, time.Now().AddDate(0, 0, -30))
	if err != nil {
		s.updateFailure(ctx, updateRecord.ID, err)
		return err
	}

	// 处理CVE数据
	var recordsNew, recordsUpdated int64
	for _, cve := range cveData {
		vulnInfo := s.convertCVEToVulnInfo(cve)
		
		// 检查是否已存在
		filter := bson.M{"cve_id": vulnInfo.CVEId}
		var existing models.VulnDbInfo
		vulnCollection := s.db.Collection("vuln_database")
		err := vulnCollection.FindOne(ctx, filter).Decode(&existing)
		
		if err == mongo.ErrNoDocuments {
			// 新记录
			vulnInfo.CreatedAt = time.Now()
			vulnInfo.UpdatedAt = time.Now()
			_, err = vulnCollection.InsertOne(ctx, vulnInfo)
			if err == nil {
				recordsNew++
			}
		} else if err == nil {
			// 更新现有记录
			vulnInfo.ID = existing.ID
			vulnInfo.CreatedAt = existing.CreatedAt
			vulnInfo.UpdatedAt = time.Now()
			vulnInfo.Version = existing.Version + 1
			
			update := bson.M{"$set": vulnInfo}
			_, err = vulnCollection.UpdateOne(ctx, filter, update)
			if err == nil {
				recordsUpdated++
			}
		}
	}

	// 更新完成状态
	s.updateSuccess(ctx, updateRecord.ID, recordsNew, recordsUpdated, int64(len(cveData)))

	return nil
}

// UpdateCWEDatabase 更新CWE数据库
func (s *Service) UpdateCWEDatabase(ctx context.Context) error {
	// TODO: 实现CWE数据库更新逻辑
	updateRecord := &models.VulnUpdate{
		Source:     "CWE",
		UpdateType: "full",
		StartTime:  time.Now(),
		Status:     "running",
	}

	collection := s.db.Collection("vuln_updates")
	result, err := collection.InsertOne(ctx, updateRecord)
	if err != nil {
		return err
	}
	updateRecord.ID = result.InsertedID.(primitive.ObjectID)

	// 获取CWE数据
	cweData, err := s.cweClient.FetchCWEData(ctx)
	if err != nil {
		s.updateFailure(ctx, updateRecord.ID, err)
		return err
	}

	// 处理CWE数据并关联到漏洞信息
	var recordsNew, recordsUpdated int64
	for _, cwe := range cweData {
		// 将CWE数据转换为漏洞信息格式
		vulnInfo := s.convertCWEToVulnInfo(cwe)
		
		filter := bson.M{"cwe_id": vulnInfo.CWEId}
		var existing models.VulnDbInfo
		vulnCollection := s.db.Collection("vuln_database")
		err := vulnCollection.FindOne(ctx, filter).Decode(&existing)
		
		if err == mongo.ErrNoDocuments {
			vulnInfo.CreatedAt = time.Now()
			vulnInfo.UpdatedAt = time.Now()
			_, err = vulnCollection.InsertOne(ctx, vulnInfo)
			if err == nil {
				recordsNew++
			}
		} else if err == nil {
			// 合并CWE信息到现有记录
			existing.CWEId = vulnInfo.CWEId
			existing.Category = vulnInfo.Category
			existing.VulnType = vulnInfo.VulnType
			existing.UpdatedAt = time.Now()
			
			update := bson.M{"$set": existing}
			_, err = vulnCollection.UpdateOne(ctx, filter, update)
			if err == nil {
				recordsUpdated++
			}
		}
	}

	s.updateSuccess(ctx, updateRecord.ID, recordsNew, recordsUpdated, int64(len(cweData)))
	return nil
}

// UpdateCNVDDatabase 更新CNVD数据库
func (s *Service) UpdateCNVDDatabase(ctx context.Context) error {
	// TODO: 实现CNVD数据库更新逻辑
	updateRecord := &models.VulnUpdate{
		Source:     "CNVD",
		UpdateType: "incremental",
		StartTime:  time.Now(),
		Status:     "running",
	}

	collection := s.db.Collection("vuln_updates")
	result, err := collection.InsertOne(ctx, updateRecord)
	if err != nil {
		return err
	}
	updateRecord.ID = result.InsertedID.(primitive.ObjectID)

	// 获取CNVD数据
	cnvdData, err := s.cnvdClient.FetchRecentCNVDs(ctx, time.Now().AddDate(0, 0, -30))
	if err != nil {
		s.updateFailure(ctx, updateRecord.ID, err)
		return err
	}

	var recordsNew, recordsUpdated int64
	for _, cnvd := range cnvdData {
		vulnInfo := s.convertCNVDToVulnInfo(cnvd)
		
		filter := bson.M{"cnvd_id": vulnInfo.CNVDId}
		var existing models.VulnDbInfo
		vulnCollection := s.db.Collection("vuln_database")
		err := vulnCollection.FindOne(ctx, filter).Decode(&existing)
		
		if err == mongo.ErrNoDocuments {
			vulnInfo.CreatedAt = time.Now()
			vulnInfo.UpdatedAt = time.Now()
			_, err = vulnCollection.InsertOne(ctx, vulnInfo)
			if err == nil {
				recordsNew++
			}
		} else if err == nil {
			vulnInfo.ID = existing.ID
			vulnInfo.CreatedAt = existing.CreatedAt
			vulnInfo.UpdatedAt = time.Now()
			vulnInfo.Version = existing.Version + 1
			
			update := bson.M{"$set": vulnInfo}
			_, err = vulnCollection.UpdateOne(ctx, filter, update)
			if err == nil {
				recordsUpdated++
			}
		}
	}

	s.updateSuccess(ctx, updateRecord.ID, recordsNew, recordsUpdated, int64(len(cnvdData)))
	return nil
}

// GetVulnerabilityStats 获取漏洞数据库统计信息
func (s *Service) GetVulnerabilityStats(ctx context.Context) (*models.VulnDbStats, error) {
	collection := s.db.Collection("vuln_database")
	
	// 总数统计
	totalCount, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// 严重程度统计
	severityPipeline := []bson.M{
		{"$group": bson.M{
			"_id":   "$severity",
			"count": bson.M{"$sum": 1},
		}},
	}
	
	severityCursor, err := collection.Aggregate(ctx, severityPipeline)
	if err != nil {
		return nil, err
	}
	defer severityCursor.Close(ctx)

	severityCount := make(map[string]int64)
	for severityCursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := severityCursor.Decode(&result); err == nil {
			severityCount[result.ID] = result.Count
		}
	}

	// TODO: 实现其他统计项目

	return &models.VulnDbStats{
		TotalCount:     totalCount,
		SeverityCount:  severityCount,
		LastUpdateTime: time.Now(),
	}, nil
}

// SearchVulnerabilities 搜索漏洞信息
func (s *Service) SearchVulnerabilities(ctx context.Context, query models.VulnDbQuery) ([]models.VulnDbInfo, int64, error) {
	collection := s.db.Collection("vuln_database")
	
	// 构建查询条件
	filter := s.buildSearchFilter(query)
	
	// 设置分页和排序
	opts := options.Find()
	if query.Page > 0 && query.PageSize > 0 {
		skip := int64((query.Page - 1) * query.PageSize)
		opts.SetSkip(skip).SetLimit(int64(query.PageSize))
	}
	
	if query.SortBy != "" {
		sortOrder := 1
		if query.SortDesc {
			sortOrder = -1
		}
		opts.SetSort(bson.M{query.SortBy: sortOrder})
	}

	// 执行查询
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []models.VulnDbInfo
	if err = cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// 辅助方法

func (s *Service) convertCVEToVulnInfo(cve CVEData) models.VulnDbInfo {
	vuln := models.VulnDbInfo{
		CVEId:  cve.ID,
		Source: "CVE",
		Status: "active",
	}
	
	// 从Descriptions提取描述和标题
	for _, desc := range cve.Descriptions {
		if desc.Lang == "en" {
			vuln.Description = desc.Value
			// 生成标题（取描述的前100个字符）
			if len(desc.Value) > 100 {
				vuln.Title = desc.Value[:100] + "..."
			} else {
				vuln.Title = desc.Value
			}
			break
		}
	}
	
	// 解析发布时间
	if publishedTime, err := time.Parse(time.RFC3339, cve.Published); err == nil {
		vuln.PublishedDate = publishedTime
	}
	
	// 解析修改时间
	if modifiedTime, err := time.Parse(time.RFC3339, cve.LastModified); err == nil {
		vuln.ModifiedDate = modifiedTime
	}
	
	// 从Metrics提取CVSS信息
	if len(cve.Metrics.CvssMetricV31) > 0 {
		metric := cve.Metrics.CvssMetricV31[0]
		vuln.CVSSScore = metric.CvssData.BaseScore
		vuln.CVSSVector = metric.CvssData.VectorString
		vuln.CVSSVersion = "3.1"
		vuln.Severity = metric.CvssData.BaseSeverity
	} else if len(cve.Metrics.CvssMetricV30) > 0 {
		metric := cve.Metrics.CvssMetricV30[0]
		vuln.CVSSScore = metric.CvssData.BaseScore
		vuln.CVSSVector = metric.CvssData.VectorString
		vuln.CVSSVersion = "3.0"
		vuln.Severity = metric.CvssData.BaseSeverity
	}
	
	return vuln
}

func (s *Service) convertCWEToVulnInfo(cwe CWEData) models.VulnDbInfo {
	// TODO: 实现CWE数据转换逻辑
	return models.VulnDbInfo{
		CWEId:       cwe.ID,
		Title:       cwe.Name,
		Description: cwe.Description,
		Category:    cwe.Category,
		VulnType:    cwe.WeaknessType,
		Source:      "CWE",
		Status:      "active",
	}
}

func (s *Service) convertCNVDToVulnInfo(cnvd CNVDData) models.VulnDbInfo {
	// TODO: 实现CNVD数据转换逻辑
	return models.VulnDbInfo{
		CNVDId:        cnvd.ID,
		Title:         cnvd.Title,
		Description:   cnvd.Description,
		Severity:      cnvd.Severity,
		CVSSScore:     cnvd.CVSSScore,
		PublishedDate: cnvd.PublishedDate,
		ModifiedDate:  cnvd.ModifiedDate,
		Source:        "CNVD",
		Status:        "active",
	}
}

func (s *Service) buildSearchFilter(query models.VulnDbQuery) bson.M {
	filter := bson.M{}

	// 基本查询
	if query.CVEId != "" {
		filter["cve_id"] = bson.M{"$regex": query.CVEId, "$options": "i"}
	}
	if query.CWEId != "" {
		filter["cwe_id"] = bson.M{"$regex": query.CWEId, "$options": "i"}
	}
	if query.CNVDId != "" {
		filter["cnvd_id"] = bson.M{"$regex": query.CNVDId, "$options": "i"}
	}
	if query.Keyword != "" {
		filter["$or"] = []bson.M{
			{"title": bson.M{"$regex": query.Keyword, "$options": "i"}},
			{"description": bson.M{"$regex": query.Keyword, "$options": "i"}},
		}
	}

	// 过滤条件
	if len(query.Severity) > 0 {
		filter["severity"] = bson.M{"$in": query.Severity}
	}
	if len(query.Source) > 0 {
		filter["source"] = bson.M{"$in": query.Source}
	}

	// TODO: 实现更多过滤条件

	return filter
}

func (s *Service) updateSuccess(ctx context.Context, updateID primitive.ObjectID, recordsNew, recordsUpdated, recordsTotal int64) {
	collection := s.db.Collection("vuln_updates")
	update := bson.M{
		"$set": bson.M{
			"end_time":        time.Now(),
			"status":          "completed",
			"records_new":     recordsNew,
			"records_updated": recordsUpdated,
			"records_total":   recordsTotal,
			"progress":        100.0,
		},
	}
	collection.UpdateOne(ctx, bson.M{"_id": updateID}, update)
}

func (s *Service) updateFailure(ctx context.Context, updateID primitive.ObjectID, err error) {
	collection := s.db.Collection("vuln_updates")
	update := bson.M{
		"$set": bson.M{
			"end_time":      time.Now(),
			"status":        "failed",
			"error_message": err.Error(),
		},
	}
	collection.UpdateOne(ctx, bson.M{"_id": updateID}, update)
}

// 数据结构定义 (TODO: 应该从实际API响应中获取)
type CWEData struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	WeaknessType string `json:"weakness_type"`
}

type CNVDData struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Severity     string    `json:"severity"`
	CVSSScore    float64   `json:"cvss_score"`
	PublishedDate time.Time `json:"published_date"`
	ModifiedDate time.Time `json:"modified_date"`
}