package vulndb

import (
	"context"
	"time"

	"github.com/StellarServer/internal/utils"
)

// Scheduler 漏洞数据库自动更新调度器
type Scheduler struct {
	service *Service
	config  Config
	ticker  *time.Ticker
	stopCh  chan bool
}

// NewScheduler 创建调度器
func NewScheduler(service *Service, config Config) *Scheduler {
	return &Scheduler{
		service: service,
		config:  config,
		stopCh:  make(chan bool),
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	if s.config.UpdateInterval == 0 {
		s.config.UpdateInterval = 24 * time.Hour // 默认每天更新一次
	}
	
	s.ticker = time.NewTicker(s.config.UpdateInterval)
	
	utils.Info("漏洞数据库调度器已启动", "间隔", s.config.UpdateInterval)
	
	// 立即执行一次更新
	go s.updateAllDatabases()
	
	// 定期更新
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.updateAllDatabases()
			case <-s.stopCh:
				return
			}
		}
	}()
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopCh)
	utils.Info("漏洞数据库调度器已停止")
}

// updateAllDatabases 更新所有数据库
func (s *Scheduler) updateAllDatabases() {
	ctx := context.Background()
	
	utils.Info("开始更新漏洞数据库")
	
	if err := s.service.UpdateAllDatabases(ctx); err != nil {
		utils.Error("更新漏洞数据库失败", err)
	} else {
		utils.Info("漏洞数据库更新完成")
	}
}

// ForceUpdate 强制立即更新
func (s *Scheduler) ForceUpdate() error {
	ctx := context.Background()
	return s.service.UpdateAllDatabases(ctx)
}