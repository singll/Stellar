package vulndb

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/StellarServer/internal/services"
)

// VulnDbUpdater 漏洞数据库更新器
type VulnDbUpdater struct {
	vulnDbService *services.VulnDbService
	cveFetcher    *CVEDataFetcher
	cweFetcher    *CWEDataFetcher
	config        UpdaterConfig
	running       bool
	mutex         sync.RWMutex
}

// UpdaterConfig 更新器配置
type UpdaterConfig struct {
	CVEAPIKey           string        `json:"cve_api_key"`
	AutoUpdateEnabled   bool          `json:"auto_update_enabled"`
	UpdateInterval      time.Duration `json:"update_interval"`
	CVEUpdateEnabled    bool          `json:"cve_update_enabled"`
	CWEUpdateEnabled    bool          `json:"cwe_update_enabled"`
	CNVDUpdateEnabled   bool          `json:"cnvd_update_enabled"`
	BatchSize           int           `json:"batch_size"`
	MaxRetries          int           `json:"max_retries"`
	RetryDelay          time.Duration `json:"retry_delay"`
	ConcurrentWorkers   int           `json:"concurrent_workers"`
}

// DefaultUpdaterConfig 默认更新器配置
func DefaultUpdaterConfig() UpdaterConfig {
	return UpdaterConfig{
		AutoUpdateEnabled: true,
		UpdateInterval:    24 * time.Hour, // 每24小时更新一次
		CVEUpdateEnabled:  true,
		CWEUpdateEnabled:  true,
		CNVDUpdateEnabled: false, // 暂时禁用CNVD
		BatchSize:         100,
		MaxRetries:        3,
		RetryDelay:        5 * time.Second,
		ConcurrentWorkers: 3,
	}
}

// NewVulnDbUpdater 创建漏洞数据库更新器
func NewVulnDbUpdater(vulnDbService *services.VulnDbService, config UpdaterConfig) *VulnDbUpdater {
	updater := &VulnDbUpdater{
		vulnDbService: vulnDbService,
		cveFetcher:    NewCVEDataFetcher(config.CVEAPIKey),
		cweFetcher:    NewCWEDataFetcher(),
		config:        config,
	}
	
	return updater
}

// Start 启动自动更新
func (u *VulnDbUpdater) Start(ctx context.Context) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	
	if u.running {
		return fmt.Errorf("更新器已在运行")
	}
	
	if !u.config.AutoUpdateEnabled {
		return fmt.Errorf("自动更新未启用")
	}
	
	u.running = true
	
	// 启动定期更新
	go u.runPeriodicUpdate(ctx)
	
	log.Println("漏洞数据库更新器已启动")
	return nil
}

// Stop 停止自动更新
func (u *VulnDbUpdater) Stop() {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	
	u.running = false
	log.Println("漏洞数据库更新器已停止")
}

// IsRunning 检查是否在运行
func (u *VulnDbUpdater) IsRunning() bool {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	
	return u.running
}

// runPeriodicUpdate 运行定期更新
func (u *VulnDbUpdater) runPeriodicUpdate(ctx context.Context) {
	ticker := time.NewTicker(u.config.UpdateInterval)
	defer ticker.Stop()
	
	// 启动时立即执行一次更新
	u.performFullUpdate(ctx)
	
	for {
		select {
		case <-ctx.Done():
			log.Println("收到停止信号，退出定期更新")
			return
		case <-ticker.C:
			if !u.IsRunning() {
				return
			}
			
			log.Println("开始定期漏洞数据库更新")
			u.performFullUpdate(ctx)
		}
	}
}

// performFullUpdate 执行完整更新
func (u *VulnDbUpdater) performFullUpdate(ctx context.Context) {
	var wg sync.WaitGroup
	
	// 并发更新不同数据源
	if u.config.CVEUpdateEnabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := u.updateCVEData(ctx); err != nil {
				log.Printf("CVE数据更新失败: %v", err)
			}
		}()
	}
	
	if u.config.CWEUpdateEnabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := u.updateCWEData(ctx); err != nil {
				log.Printf("CWE数据更新失败: %v", err)
			}
		}()
	}
	
	if u.config.CNVDUpdateEnabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := u.updateCNVDData(ctx); err != nil {
				log.Printf("CNVD数据更新失败: %v", err)
			}
		}()
	}
	
	wg.Wait()
	log.Println("漏洞数据库更新完成")
}

// updateCVEData 更新CVE数据
func (u *VulnDbUpdater) updateCVEData(ctx context.Context) error {
	log.Println("开始更新CVE数据")
	
	updateRecord := &models.VulnUpdate{
		Source:        "CVE",
		UpdateType:    "incremental",
		StartTime:     time.Now(),
		Status:        "running",
		Progress:      0,
		CreatedAt:     time.Now(),
	}
	
	// 获取最近30天的CVE数据
	vulns, err := u.cveFetcher.FetchRecentCVEs(ctx, 30)
	if err != nil {
		updateRecord.Status = "failed"
		updateRecord.ErrorMessage = err.Error()
		updateRecord.EndTime = time.Now()
		return fmt.Errorf("获取CVE数据失败: %v", err)
	}
	
	updateRecord.RecordsTotal = int64(len(vulns))
	
	// 批量处理漏洞数据
	batchSize := u.config.BatchSize
	
	var newRecords, updatedRecords, errorRecords int64
	
	for i := 0; i < len(vulns); i += batchSize {
		end := i + batchSize
		if end > len(vulns) {
			end = len(vulns)
		}
		
		batch := vulns[i:end]
		batchNew, batchUpdated, batchErrors := u.processCVEBatch(ctx, batch)
		
		newRecords += batchNew
		updatedRecords += batchUpdated
		errorRecords += batchErrors
		
		// 更新进度
		progress := float64(end) / float64(len(vulns)) * 100
		updateRecord.Progress = progress
		updateRecord.RecordsNew = newRecords
		updateRecord.RecordsUpdated = updatedRecords
		updateRecord.RecordsError = errorRecords
		
		log.Printf("CVE更新进度: %.1f%% (%d/%d)", progress, end, len(vulns))
		
		// 避免过于频繁的请求
		if i+batchSize < len(vulns) {
			time.Sleep(u.config.RetryDelay)
		}
	}
	
	updateRecord.Status = "completed"
	updateRecord.EndTime = time.Now()
	updateRecord.Progress = 100
	
	log.Printf("CVE数据更新完成: 新增 %d, 更新 %d, 错误 %d", newRecords, updatedRecords, errorRecords)
	return nil
}

// processCVEBatch 处理CVE批次数据
func (u *VulnDbUpdater) processCVEBatch(ctx context.Context, batch []*models.VulnDbInfo) (int64, int64, int64) {
	var newRecords, updatedRecords, errorRecords int64
	
	for _, vuln := range batch {
		if vuln.CVEId == "" {
			errorRecords++
			continue
		}
		
		// 尝试获取现有记录
		existing, err := u.vulnDbService.GetVulnerabilityByCVE(ctx, vuln.CVEId)
		if err != nil && !strings.Contains(err.Error(), "不存在") {
			log.Printf("检查CVE %s 时出错: %v", vuln.CVEId, err)
			errorRecords++
			continue
		}
		
		if existing != nil {
			// 更新现有记录
			if err := u.vulnDbService.UpdateVulnerability(ctx, existing.ID, vuln); err != nil {
				log.Printf("更新CVE %s 失败: %v", vuln.CVEId, err)
				errorRecords++
			} else {
				updatedRecords++
			}
		} else {
			// 创建新记录
			if err := u.vulnDbService.CreateCustomVulnerability(ctx, vuln); err != nil {
				log.Printf("创建CVE %s 失败: %v", vuln.CVEId, err)
				errorRecords++
			} else {
				newRecords++
			}
		}
	}
	
	return newRecords, updatedRecords, errorRecords
}

// updateCWEData 更新CWE数据
func (u *VulnDbUpdater) updateCWEData(ctx context.Context) error {
	log.Println("开始更新CWE数据")
	
	// 使用模拟数据进行更新
	vulns := u.cweFetcher.GetMockCWEData()
	
	var newRecords, updatedRecords, errorRecords int64
	
	for _, vuln := range vulns {
		// 检查是否已存在
		query := models.VulnDbQuery{
			CWEId:    vuln.CWEId,
			PageSize: 1,
		}
		
		existingVulns, total, err := u.vulnDbService.SearchVulnerabilities(ctx, query)
		if err != nil {
			log.Printf("检查CWE %s 时出错: %v", vuln.CWEId, err)
			errorRecords++
			continue
		}
		
		if total > 0 && len(existingVulns) > 0 {
			// 更新现有记录
			if err := u.vulnDbService.UpdateVulnerability(ctx, existingVulns[0].ID, vuln); err != nil {
				log.Printf("更新CWE %s 失败: %v", vuln.CWEId, err)
				errorRecords++
			} else {
				updatedRecords++
			}
		} else {
			// 创建新记录
			if err := u.vulnDbService.CreateCustomVulnerability(ctx, vuln); err != nil {
				log.Printf("创建CWE %s 失败: %v", vuln.CWEId, err)
				errorRecords++
			} else {
				newRecords++
			}
		}
	}
	
	log.Printf("CWE数据更新完成: 新增 %d, 更新 %d, 错误 %d", newRecords, updatedRecords, errorRecords)
	return nil
}

// updateCNVDData 更新CNVD数据
func (u *VulnDbUpdater) updateCNVDData(ctx context.Context) error {
	log.Println("开始更新CNVD数据")
	
	// TODO: 实现CNVD数据获取和更新
	// 由于CNVD API限制，这里暂时使用模拟实现
	
	log.Println("CNVD数据更新完成（模拟）")
	return nil
}

// UpdateNow 立即执行更新
func (u *VulnDbUpdater) UpdateNow(ctx context.Context) error {
	log.Println("开始手动漏洞数据库更新")
	u.performFullUpdate(ctx)
	return nil
}

// UpdateCVEOnly 仅更新CVE数据
func (u *VulnDbUpdater) UpdateCVEOnly(ctx context.Context) error {
	return u.updateCVEData(ctx)
}

// UpdateCWEOnly 仅更新CWE数据
func (u *VulnDbUpdater) UpdateCWEOnly(ctx context.Context) error {
	return u.updateCWEData(ctx)
}

// GetConfig 获取更新器配置
func (u *VulnDbUpdater) GetConfig() UpdaterConfig {
	return u.config
}

// UpdateConfig 更新配置
func (u *VulnDbUpdater) UpdateConfig(config UpdaterConfig) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	
	u.config = config
	
	// 更新CVE获取器的API密钥
	if config.CVEAPIKey != "" {
		u.cveFetcher = NewCVEDataFetcher(config.CVEAPIKey)
	}
	
	log.Println("更新器配置已更新")
}

// GetStats 获取更新统计信息
func (u *VulnDbUpdater) GetStats(ctx context.Context) (*models.VulnDbStats, error) {
	return u.vulnDbService.GetStatistics(ctx)
}

// GetUpdateHistory 获取更新历史
func (u *VulnDbUpdater) GetUpdateHistory(ctx context.Context, page, pageSize int) ([]*models.VulnUpdate, int64, error) {
	return u.vulnDbService.GetUpdateHistory(ctx, page, pageSize)
}