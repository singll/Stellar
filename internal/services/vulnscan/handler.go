package vulnscan

import (
	"context"
	"time"

	"github.com/StellarServer/internal/models"
	pkgerrors "github.com/StellarServer/internal/pkg/errors"
	"github.com/StellarServer/internal/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// VulnHandler 漏洞处理器
type VulnHandler struct {
	db *mongo.Database
}

// NewVulnHandler 创建漏洞处理器
func NewVulnHandler(db *mongo.Database) *VulnHandler {
	return &VulnHandler{
		db: db,
	}
}

// HandlePOCResult 处理POC执行结果
func (h *VulnHandler) HandlePOCResult(result *models.POCResult) error {
	// 保存POC结果到数据库
	_, err := h.db.Collection("poc_results").InsertOne(context.Background(), result)
	if err != nil {
		logger.Error("HandlePOCResult failed", map[string]interface{}{"error": err, "resultID": result.ID})
		return pkgerrors.WrapDatabaseError(err, "保存POC结果")
	}
	return nil
}

// HandleVulnerability 处理漏洞
func (h *VulnHandler) HandleVulnerability(vuln *models.Vulnerability) error {
	// 检查是否已存在相同漏洞
	filter := bson.M{
		"projectId":    vuln.ProjectID,
		"assetId":      vuln.AssetID,
		"title":        vuln.Title,
		"affectedUrl":  vuln.AffectedURL,
		"affectedHost": vuln.AffectedHost,
	}

	var existingVuln models.Vulnerability
	err := h.db.Collection("vulnerabilities").FindOne(context.Background(), filter).Decode(&existingVuln)
	if err == nil {
		// 漏洞已存在，更新信息
		update := bson.M{
			"$set": bson.M{
				"updatedAt":   time.Now(),
				"description": vuln.Description,
				"solution":    vuln.Solution,
				"references":  vuln.References,
				"payload":     vuln.Payload,
				"request":     vuln.Request,
				"response":    vuln.Response,
				"screenshot":  vuln.Screenshot,
			},
		}
		_, err = h.db.Collection("vulnerabilities").UpdateOne(
			context.Background(),
			bson.M{"_id": existingVuln.ID},
			update,
		)
		if err != nil {
			logger.Error("HandleVulnerability update existing vulnerability failed", map[string]interface{}{"vulnID": existingVuln.ID.Hex(), "error": err})
			return pkgerrors.WrapDatabaseError(err, "更新已存在的漏洞")
		}
		return nil
	}

	// 漏洞不存在，创建新漏洞
	_, err = h.db.Collection("vulnerabilities").InsertOne(context.Background(), vuln)
	if err != nil {
		logger.Error("HandleVulnerability create new vulnerability failed", map[string]interface{}{"vulnTitle": vuln.Title, "error": err})
		return pkgerrors.WrapDatabaseError(err, "创建新漏洞")
	}
	return nil
}

// UpdateTaskProgress 更新任务进度
func (h *VulnHandler) UpdateTaskProgress(taskID string, progress float64) error {
	// 解析任务ID
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		logger.Error("UpdateTaskProgress invalid taskID", map[string]interface{}{"taskID": taskID, "error": err})
		return pkgerrors.NewAppErrorWithCause(pkgerrors.CodeBadRequest, "无效的任务ID", 400, err)
	}

	// 更新任务进度
	update := bson.M{
		"$set": bson.M{
			"progress": progress,
		},
	}
	_, err = h.db.Collection("vuln_scan_tasks").UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		update,
	)
	if err != nil {
		logger.Error("UpdateTaskProgress failed", map[string]interface{}{"taskID": taskID, "error": err})
		return pkgerrors.WrapDatabaseError(err, "更新任务进度")
	}
	return nil
}

// UpdateTaskStatus 更新任务状态
func (h *VulnHandler) UpdateTaskStatus(taskID string, status string) error {
	// 解析任务ID
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		logger.Error("UpdateTaskStatus invalid taskID", map[string]interface{}{"taskID": taskID, "error": err})
		return pkgerrors.NewAppErrorWithCause(pkgerrors.CodeBadRequest, "无效的任务ID", 400, err)
	}

	// 更新任务状态
	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	// 如果任务完成或失败，设置完成时间
	if status == "completed" || status == "failed" || status == "stopped" {
		update["$set"].(bson.M)["completedAt"] = time.Now()
	}

	// 如果任务开始运行，设置开始时间
	if status == "running" {
		update["$set"].(bson.M)["startedAt"] = time.Now()
	}

	_, err = h.db.Collection("vuln_scan_tasks").UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		update,
	)
	if err != nil {
		logger.Error("UpdateTaskStatus failed", map[string]interface{}{"taskID": taskID, "status": status, "error": err})
		return pkgerrors.WrapDatabaseError(err, "更新任务状态")
	}
	return nil
}

// FinishTask 完成任务
func (h *VulnHandler) FinishTask(task *models.VulnScanTask) error {
	// 更新任务状态
	update := bson.M{
		"$set": bson.M{
			"status":        "completed",
			"completedAt":   time.Now(),
			"progress":      100,
			"resultSummary": task.ResultSummary,
		},
	}
	_, err := h.db.Collection("vuln_scan_tasks").UpdateOne(
		context.Background(),
		bson.M{"_id": task.ID},
		update,
	)
	if err != nil {
		logger.Error("FinishTask failed", map[string]interface{}{"taskID": task.ID.Hex(), "error": err})
		return pkgerrors.WrapDatabaseError(err, "完成任务")
	}
	return nil
}

// GetVulnerabilities 获取漏洞列表
func (h *VulnHandler) GetVulnerabilities(projectID string, query map[string]interface{}, page, pageSize int) ([]models.Vulnerability, int, error) {
	// 解析项目ID
	objID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		logger.Error("GetVulnerabilities invalid projectID", map[string]interface{}{"projectID": projectID, "error": err})
		return nil, 0, pkgerrors.NewAppErrorWithCause(pkgerrors.CodeBadRequest, "无效的项目ID", 400, err)
	}

	// 构建查询条件
	filter := bson.M{"projectId": objID}
	for key, value := range query {
		filter[key] = value
	}

	// 计算总数
	total, err := h.db.Collection("vulnerabilities").CountDocuments(context.Background(), filter)
	if err != nil {
		logger.Error("GetVulnerabilities count documents failed", map[string]interface{}{"projectID": projectID, "error": err})
		return nil, 0, pkgerrors.WrapDatabaseError(err, "查询漏洞总数")
	}

	// 分页查询
	skip := (page - 1) * pageSize
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(pageSize)).SetSort(bson.D{{"createdAt", -1}})
	cursor, err := h.db.Collection("vulnerabilities").Find(context.Background(), filter, opts)
	if err != nil {
		logger.Error("GetVulnerabilities find documents failed", map[string]interface{}{"projectID": projectID, "error": err})
		return nil, 0, pkgerrors.WrapDatabaseError(err, "查询漏洞列表")
	}
	defer cursor.Close(context.Background())

	// 解析结果
	var vulns []models.Vulnerability
	if err = cursor.All(context.Background(), &vulns); err != nil {
		logger.Error("GetVulnerabilities decode documents failed", map[string]interface{}{"projectID": projectID, "error": err})
		return nil, 0, pkgerrors.WrapDatabaseError(err, "解析漏洞数据")
	}

	return vulns, int(total), nil
}

// GetVulnerabilityByID 根据ID获取漏洞
func (h *VulnHandler) GetVulnerabilityByID(vulnID string) (*models.Vulnerability, error) {
	// 解析漏洞ID
	objID, err := primitive.ObjectIDFromHex(vulnID)
	if err != nil {
		logger.Error("GetVulnerabilityByID invalid vulnID", map[string]interface{}{"vulnID": vulnID, "error": err})
		return nil, pkgerrors.NewAppErrorWithCause(pkgerrors.CodeBadRequest, "无效的漏洞ID", 400, err)
	}

	// 查询漏洞
	var vuln models.Vulnerability
	err = h.db.Collection("vulnerabilities").FindOne(
		context.Background(),
		bson.M{"_id": objID},
	).Decode(&vuln)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Warn("GetVulnerabilityByID vulnerability not found", map[string]interface{}{"vulnID": vulnID})
			return nil, pkgerrors.NewNotFoundError("漏洞不存在")
		}
		logger.Error("GetVulnerabilityByID failed", map[string]interface{}{"vulnID": vulnID, "error": err})
		return nil, pkgerrors.WrapDatabaseError(err, "查询漏洞详情")
	}

	return &vuln, nil
}

// UpdateVulnerabilityStatus 更新漏洞状态
func (h *VulnHandler) UpdateVulnerabilityStatus(vulnID string, status models.VulnerabilityStatus) error {
	// 解析漏洞ID
	objID, err := primitive.ObjectIDFromHex(vulnID)
	if err != nil {
		logger.Error("UpdateVulnerabilityStatus invalid vulnID", map[string]interface{}{"vulnID": vulnID, "error": err})
		return pkgerrors.NewAppErrorWithCause(pkgerrors.CodeBadRequest, "无效的漏洞ID", 400, err)
	}

	// 更新状态
	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	// 如果状态为已修复，设置修复时间
	if status == models.StatusFixed {
		update["$set"].(bson.M)["fixedAt"] = time.Now()
	}

	// 如果状态为已验证，设置验证时间
	if status == models.StatusVerified {
		update["$set"].(bson.M)["verifiedAt"] = time.Now()
	}

	_, err = h.db.Collection("vulnerabilities").UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		update,
	)
	if err != nil {
		logger.Error("UpdateVulnerabilityStatus failed", map[string]interface{}{"vulnID": vulnID, "status": status, "error": err})
		return pkgerrors.WrapDatabaseError(err, "更新漏洞状态")
	}
	return nil
}

// GetTaskSummary 获取任务摘要
func (h *VulnHandler) GetTaskSummary(taskID string) (*models.VulnScanSummary, error) {
	// 解析任务ID
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		logger.Error("GetTaskSummary invalid taskID", map[string]interface{}{"taskID": taskID, "error": err})
		return nil, pkgerrors.NewAppErrorWithCause(pkgerrors.CodeBadRequest, "无效的任务ID", 400, err)
	}

	// 查询任务
	var task models.VulnScanTask
	err = h.db.Collection("vuln_scan_tasks").FindOne(
		context.Background(),
		bson.M{"_id": objID},
	).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Warn("GetTaskSummary task not found", map[string]interface{}{"taskID": taskID})
			return nil, pkgerrors.NewTaskNotFoundError()
		}
		logger.Error("GetTaskSummary failed", map[string]interface{}{"taskID": taskID, "error": err})
		return nil, pkgerrors.WrapDatabaseError(err, "查询任务摘要")
	}

	return &task.ResultSummary, nil
}

// GetTaskResults 获取任务结果
func (h *VulnHandler) GetTaskResults(taskID string) ([]models.POCResult, error) {
	// 解析任务ID
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		logger.Error("GetTaskResults invalid taskID", map[string]interface{}{"taskID": taskID, "error": err})
		return nil, pkgerrors.NewAppErrorWithCause(pkgerrors.CodeBadRequest, "无效的任务ID", 400, err)
	}

	// 查询结果
	cursor, err := h.db.Collection("poc_results").Find(
		context.Background(),
		bson.M{"taskId": objID},
	)
	if err != nil {
		logger.Error("GetTaskResults find results failed", map[string]interface{}{"taskID": taskID, "error": err})
		return nil, pkgerrors.WrapDatabaseError(err, "查询任务结果")
	}
	defer cursor.Close(context.Background())

	// 解析结果
	var results []models.POCResult
	if err = cursor.All(context.Background(), &results); err != nil {
		logger.Error("GetTaskResults decode results failed", map[string]interface{}{"taskID": taskID, "error": err})
		return nil, pkgerrors.WrapDatabaseError(err, "解析任务结果")
	}

	return results, nil
}

// GetTaskVulnerabilities 获取任务漏洞
func (h *VulnHandler) GetTaskVulnerabilities(taskID string) ([]models.Vulnerability, error) {
	// 解析任务ID
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		logger.Error("GetTaskVulnerabilities invalid taskID", map[string]interface{}{"taskID": taskID, "error": err})
		return nil, pkgerrors.NewAppErrorWithCause(pkgerrors.CodeBadRequest, "无效的任务ID", 400, err)
	}

	// 查询漏洞
	cursor, err := h.db.Collection("vulnerabilities").Find(
		context.Background(),
		bson.M{"taskId": objID},
	)
	if err != nil {
		logger.Error("GetTaskVulnerabilities find vulnerabilities failed", map[string]interface{}{"taskID": taskID, "error": err})
		return nil, pkgerrors.WrapDatabaseError(err, "查询任务漏洞")
	}
	defer cursor.Close(context.Background())

	// 解析结果
	var vulns []models.Vulnerability
	if err = cursor.All(context.Background(), &vulns); err != nil {
		logger.Error("GetTaskVulnerabilities decode vulnerabilities failed", map[string]interface{}{"taskID": taskID, "error": err})
		return nil, pkgerrors.WrapDatabaseError(err, "解析任务漏洞")
	}

	return vulns, nil
}

// GetScanTasks 获取扫描任务列表
func (h *VulnHandler) GetScanTasks(filter map[string]interface{}, page, pageSize int) ([]models.VulnScanTask, int, error) {
	// 计算总数
	total, err := h.db.Collection("vuln_scan_tasks").CountDocuments(context.Background(), filter)
	if err != nil {
		logger.Error("GetScanTasks count documents failed", map[string]interface{}{"filter": filter, "error": err})
		return nil, 0, pkgerrors.WrapDatabaseError(err, "查询扫描任务总数")
	}

	// 分页查询
	skip := (page - 1) * pageSize
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(pageSize)).SetSort(bson.D{{"createdAt", -1}})
	cursor, err := h.db.Collection("vuln_scan_tasks").Find(context.Background(), filter, opts)
	if err != nil {
		logger.Error("GetScanTasks find documents failed", map[string]interface{}{"filter": filter, "error": err})
		return nil, 0, pkgerrors.WrapDatabaseError(err, "查询扫描任务列表")
	}
	defer cursor.Close(context.Background())

	// 解析结果
	var tasks []models.VulnScanTask
	if err = cursor.All(context.Background(), &tasks); err != nil {
		logger.Error("GetScanTasks decode documents failed", map[string]interface{}{"filter": filter, "error": err})
		return nil, 0, pkgerrors.WrapDatabaseError(err, "解析扫描任务数据")
	}

	return tasks, int(total), nil
}

// GetScanTask 获取扫描任务详情
func (h *VulnHandler) GetScanTask(taskID string) (*models.VulnScanTask, error) {
	// 解析任务ID
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		logger.Error("GetScanTask invalid taskID", map[string]interface{}{"taskID": taskID, "error": err})
		return nil, pkgerrors.NewAppErrorWithCause(pkgerrors.CodeBadRequest, "无效的任务ID", 400, err)
	}

	// 查询任务
	var task models.VulnScanTask
	err = h.db.Collection("vuln_scan_tasks").FindOne(
		context.Background(),
		bson.M{"_id": objID},
	).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Warn("GetScanTask task not found", map[string]interface{}{"taskID": taskID})
			return nil, pkgerrors.NewTaskNotFoundError()
		}
		logger.Error("GetScanTask failed", map[string]interface{}{"taskID": taskID, "error": err})
		return nil, pkgerrors.WrapDatabaseError(err, "查询扫描任务详情")
	}

	return &task, nil
}

// DeleteScanTask 删除扫描任务
func (h *VulnHandler) DeleteScanTask(taskID string) error {
	// 解析任务ID
	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		logger.Error("DeleteScanTask invalid taskID", map[string]interface{}{"taskID": taskID, "error": err})
		return pkgerrors.NewAppErrorWithCause(pkgerrors.CodeBadRequest, "无效的任务ID", 400, err)
	}

	// 删除任务
	_, err = h.db.Collection("vuln_scan_tasks").DeleteOne(
		context.Background(),
		bson.M{"_id": objID},
	)
	if err != nil {
		logger.Error("DeleteScanTask delete task failed", map[string]interface{}{"taskID": taskID, "error": err})
		return pkgerrors.WrapDatabaseError(err, "删除扫描任务")
	}

	// 删除相关的POC结果
	_, err = h.db.Collection("poc_results").DeleteMany(
		context.Background(),
		bson.M{"taskId": objID},
	)
	if err != nil {
		logger.Error("DeleteScanTask delete poc results failed", map[string]interface{}{"taskID": taskID, "error": err})
		return pkgerrors.WrapDatabaseError(err, "删除任务相关的POC结果")
	}

	return nil
}

// GetPOCs 获取POC列表
func (h *VulnHandler) GetPOCs(query map[string]interface{}, page, pageSize int) ([]models.POC, int, error) {
	// 计算总数
	total, err := h.db.Collection("pocs").CountDocuments(context.Background(), query)
	if err != nil {
		logger.Error("GetPOCs count documents failed", map[string]interface{}{"query": query, "error": err})
		return nil, 0, pkgerrors.WrapDatabaseError(err, "查询POC总数")
	}

	// 分页查询
	skip := (page - 1) * pageSize
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(pageSize)).SetSort(bson.D{{"createdAt", -1}})
	cursor, err := h.db.Collection("pocs").Find(context.Background(), query, opts)
	if err != nil {
		logger.Error("GetPOCs find documents failed", map[string]interface{}{"query": query, "error": err})
		return nil, 0, pkgerrors.WrapDatabaseError(err, "查询POC列表")
	}
	defer cursor.Close(context.Background())

	// 解析结果
	var pocs []models.POC
	if err = cursor.All(context.Background(), &pocs); err != nil {
		logger.Error("GetPOCs decode documents failed", map[string]interface{}{"query": query, "error": err})
		return nil, 0, pkgerrors.WrapDatabaseError(err, "解析POC数据")
	}

	return pocs, int(total), nil
}

// GetPOCByID 根据ID获取POC
func (h *VulnHandler) GetPOCByID(pocID string) (*models.POC, error) {
	// 解析POC ID
	objID, err := primitive.ObjectIDFromHex(pocID)
	if err != nil {
		logger.Error("GetPOCByID invalid pocID", map[string]interface{}{"pocID": pocID, "error": err})
		return nil, pkgerrors.NewAppErrorWithCause(pkgerrors.CodeBadRequest, "无效的POC ID", 400, err)
	}

	// 查询POC
	var poc models.POC
	err = h.db.Collection("pocs").FindOne(
		context.Background(),
		bson.M{"_id": objID},
	).Decode(&poc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Warn("GetPOCByID poc not found", map[string]interface{}{"pocID": pocID})
			return nil, pkgerrors.NewPluginNotFoundError(pocID)
		}
		logger.Error("GetPOCByID failed", map[string]interface{}{"pocID": pocID, "error": err})
		return nil, pkgerrors.WrapDatabaseError(err, "查询POC详情")
	}

	return &poc, nil
}

// CreatePOC 创建POC
func (h *VulnHandler) CreatePOC(poc *models.POC) error {
	// 保存POC到数据库
	_, err := h.db.Collection("pocs").InsertOne(context.Background(), poc)
	if err != nil {
		logger.Error("CreatePOC failed", map[string]interface{}{"pocName": poc.Name, "error": err})
		return pkgerrors.WrapDatabaseError(err, "创建POC")
	}
	return nil
}

// UpdatePOC 更新POC
func (h *VulnHandler) UpdatePOC(poc *models.POC) error {
	// 更新POC
	update := bson.M{
		"$set": bson.M{
			"name":           poc.Name,
			"description":    poc.Description,
			"author":         poc.Author,
			"references":     poc.References,
			"cveId":          poc.CVEID,
			"cweId":          poc.CWEID,
			"severity":       poc.Severity,
			"type":           poc.Type,
			"category":       poc.Category,
			"script":         poc.Script,
			"scriptType":     poc.ScriptType,
			"updatedAt":      poc.UpdatedAt,
			"tags":           poc.Tags,
			"enabled":        poc.Enabled,
			"requiredParams": poc.RequiredParams,
			"defaultParams":  poc.DefaultParams,
		},
	}
	_, err := h.db.Collection("pocs").UpdateOne(
		context.Background(),
		bson.M{"_id": poc.ID},
		update,
	)
	if err != nil {
		logger.Error("UpdatePOC failed", map[string]interface{}{"pocID": poc.ID.Hex(), "pocName": poc.Name, "error": err})
		return pkgerrors.WrapDatabaseError(err, "更新POC")
	}
	return nil
}

// DeletePOC 删除POC
func (h *VulnHandler) DeletePOC(pocID string) error {
	// 解析POC ID
	objID, err := primitive.ObjectIDFromHex(pocID)
	if err != nil {
		logger.Error("DeletePOC invalid pocID", map[string]interface{}{"pocID": pocID, "error": err})
		return pkgerrors.NewAppErrorWithCause(pkgerrors.CodeBadRequest, "无效的POC ID", 400, err)
	}

	// 删除POC
	_, err = h.db.Collection("pocs").DeleteOne(
		context.Background(),
		bson.M{"_id": objID},
	)
	if err != nil {
		logger.Error("DeletePOC failed", map[string]interface{}{"pocID": pocID, "error": err})
		return pkgerrors.WrapDatabaseError(err, "删除POC")
	}
	return nil
}
