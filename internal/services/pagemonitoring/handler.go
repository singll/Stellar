package pagemonitoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ResultHandler 结果处理器
type ResultHandler struct {
	service *PageMonitoringService
}

// NewResultHandler 创建结果处理器
func NewResultHandler(service *PageMonitoringService) *ResultHandler {
	return &ResultHandler{
		service: service,
	}
}

// HandleResult 处理监控结果
func (h *ResultHandler) HandleResult(result map[string]interface{}) error {
	// 解析监控ID
	monitoringIDStr, ok := result["monitoringId"].(string)
	if !ok {
		return fmt.Errorf("无效的监控ID")
	}

	monitoringID, err := primitive.ObjectIDFromHex(monitoringIDStr)
	if err != nil {
		return fmt.Errorf("无效的监控ID格式: %v", err)
	}

	// 获取监控任务
	var monitoring models.PageMonitoring
	err = h.service.db.Collection("page_monitoring").FindOne(
		h.service.ctx,
		bson.M{"_id": monitoringID},
	).Decode(&monitoring)

	if err != nil {
		return fmt.Errorf("获取监控任务失败: %v", err)
	}

	// 解析快照数据
	snapshotData, ok := result["snapshot"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("无效的快照数据")
	}

	// 创建新的快照
	snapshot := &models.PageSnapshot{
		ID:           primitive.NewObjectID(),
		MonitoringID: monitoringID,
		URL:          monitoring.URL,
		CreatedAt:    time.Now(),
	}

	// 填充快照数据
	if statusCode, ok := snapshotData["statusCode"].(float64); ok {
		snapshot.StatusCode = int(statusCode)
	}

	if html, ok := snapshotData["html"].(string); ok {
		snapshot.HTML = html
	}

	if text, ok := snapshotData["text"].(string); ok {
		snapshot.Text = text
	}

	if contentHash, ok := snapshotData["contentHash"].(string); ok {
		snapshot.ContentHash = contentHash
	}

	if size, ok := snapshotData["size"].(float64); ok {
		snapshot.Size = int(size)
	}

	if loadTime, ok := snapshotData["loadTime"].(float64); ok {
		snapshot.LoadTime = int(loadTime)
	}

	// 解析响应头
	if headers, ok := snapshotData["headers"].(map[string]interface{}); ok {
		snapshot.Headers = make(map[string]string)
		for key, value := range headers {
			if strValue, ok := value.(string); ok {
				snapshot.Headers[key] = strValue
			}
		}
	}

	// 保存快照
	_, err = h.service.db.Collection("page_snapshots").InsertOne(h.service.ctx, snapshot)
	if err != nil {
		return fmt.Errorf("保存快照失败: %v", err)
	}

	// 判断是否为首次监控
	if monitoring.LatestSnapshot == nil {
		// 首次监控，直接更新监控任务
		_, err = h.service.db.Collection("page_monitoring").UpdateOne(
			h.service.ctx,
			bson.M{"_id": monitoringID},
			bson.M{
				"$set": bson.M{
					"latestSnapshot": snapshot,
					"updatedAt":      time.Now(),
				},
			},
		)
		if err != nil {
			return fmt.Errorf("更新监控任务失败: %v", err)
		}
		return nil
	}

	// 非首次监控，比较快照
	previousSnapshot := monitoring.LatestSnapshot
	change, similarity, diff := CompareSnapshots(previousSnapshot, snapshot, monitoring.Config)

	// 保存变更记录
	if change.Status == models.PageChangeStatusChanged {
		_, err = h.service.db.Collection("page_changes").InsertOne(h.service.ctx, change)
		if err != nil {
			return fmt.Errorf("保存变更记录失败: %v", err)
		}

		// 更新监控任务
		_, err = h.service.db.Collection("page_monitoring").UpdateOne(
			h.service.ctx,
			bson.M{"_id": monitoringID},
			bson.M{
				"$set": bson.M{
					"previousSnapshot": previousSnapshot,
					"latestSnapshot":   snapshot,
					"hasChanged":       true,
					"similarity":       similarity,
					"updatedAt":        time.Now(),
				},
				"$inc": bson.M{
					"changeCount": 1,
				},
			},
		)
		if err != nil {
			return fmt.Errorf("更新监控任务失败: %v", err)
		}

		// 发送通知
		if monitoring.Config.NotifyOnChange {
			h.sendNotification(&monitoring, snapshot, previousSnapshot, similarity, diff)
		}
	} else {
		// 更新监控任务
		_, err = h.service.db.Collection("page_monitoring").UpdateOne(
			h.service.ctx,
			bson.M{"_id": monitoringID},
			bson.M{
				"$set": bson.M{
					"previousSnapshot": previousSnapshot,
					"latestSnapshot":   snapshot,
					"hasChanged":       false,
					"similarity":       similarity,
					"updatedAt":        time.Now(),
				},
			},
		)
		if err != nil {
			return fmt.Errorf("更新监控任务失败: %v", err)
		}
	}

	return nil
}

// sendNotification 发送通知
func (h *ResultHandler) sendNotification(monitoring *models.PageMonitoring, newSnapshot, oldSnapshot *models.PageSnapshot, similarity float64, diff string) {
	// 构建通知内容
	notificationData := map[string]interface{}{
		"monitoringId":   monitoring.ID.Hex(),
		"monitoringName": monitoring.Name,
		"url":            monitoring.URL,
		"changeTime":     time.Now().Format("2006-01-02 15:04:05"),
		"similarity":     similarity,
		"statusCode":     newSnapshot.StatusCode,
		"diffUrl":        fmt.Sprintf("/api/monitoring/%s/diff", monitoring.ID.Hex()),
	}

	// 根据通知方式发送通知
	for _, method := range monitoring.Config.NotifyMethods {
		switch method {
		case "webhook":
			h.sendWebhookNotification(monitoring.Config.NotifyConfig["webhook"], notificationData)
		case "email":
			// 实现邮件通知
		case "sms":
			// 实现短信通知
		}
	}
}

// sendWebhookNotification 发送Webhook通知
func (h *ResultHandler) sendWebhookNotification(webhookURL string, data map[string]interface{}) {
	if webhookURL == "" {
		return
	}

	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("序列化通知数据失败: %v", err)
		return
	}

	// 发送请求
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("发送Webhook通知失败: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("Webhook通知返回错误状态码: %d", resp.StatusCode)
	}
}
