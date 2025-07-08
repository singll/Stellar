package utils

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	// CPU使用率
	CPUUsage float64 `json:"cpu_usage"`
	// 内存使用量（字节）
	MemoryUsage uint64 `json:"memory_usage"`
	// 内存分配量（字节）
	MemoryAllocated uint64 `json:"memory_allocated"`
	// 系统内存总量（字节）
	MemoryTotal uint64 `json:"memory_total"`
	// 协程数量
	GoroutineCount int `json:"goroutine_count"`
	// GC暂停时间（纳秒）
	GCPauseNs uint64 `json:"gc_pause_ns"`
	// GC次数
	GCCount uint32 `json:"gc_count"`
	// 采集时间
	CollectedAt time.Time `json:"collected_at"`
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	// 是否运行中
	running bool
	// 监控间隔
	interval time.Duration
	// 指标通道
	metricsChan chan PerformanceMetrics
	// 历史指标
	history []PerformanceMetrics
	// 历史指标最大数量
	maxHistory int
	// 停止通道
	stopChan chan struct{}
	// 互斥锁
	mu sync.RWMutex
	// 上次CPU样本
	lastCPUSample time.Time
	// 上次CPU使用时间
	lastCPUUsage time.Duration
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor(interval time.Duration, maxHistory int) *PerformanceMonitor {
	if interval < time.Second {
		interval = time.Second
	}
	if maxHistory < 1 {
		maxHistory = 100
	}
	return &PerformanceMonitor{
		interval:    interval,
		metricsChan: make(chan PerformanceMetrics, 10),
		history:     make([]PerformanceMetrics, 0, maxHistory),
		maxHistory:  maxHistory,
		stopChan:    make(chan struct{}),
	}
}

// Start 启动监控
func (m *PerformanceMonitor) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return
	}

	m.running = true
	m.lastCPUSample = time.Now()
	m.lastCPUUsage = 0

	// 启动采集协程
	go m.collect()

	// 启动处理协程
	go m.process()

	Info("性能监控已启动", "interval", m.interval.String())
}

// Stop 停止监控
func (m *PerformanceMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false
	close(m.stopChan)
	Info("性能监控已停止")
}

// GetLatestMetrics 获取最新指标
func (m *PerformanceMonitor) GetLatestMetrics() *PerformanceMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.history) == 0 {
		return nil
	}
	metrics := m.history[len(m.history)-1]
	return &metrics
}

// GetMetricsHistory 获取历史指标
func (m *PerformanceMonitor) GetMetricsHistory() []PerformanceMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 复制历史数据
	history := make([]PerformanceMetrics, len(m.history))
	copy(history, m.history)
	return history
}

// IsRunning 是否正在运行
func (m *PerformanceMonitor) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// collect 采集指标
func (m *PerformanceMonitor) collect() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := m.collectMetrics()
			m.metricsChan <- metrics
		case <-m.stopChan:
			return
		}
	}
}

// process 处理指标
func (m *PerformanceMonitor) process() {
	for {
		select {
		case metrics := <-m.metricsChan:
			m.mu.Lock()
			// 添加到历史记录
			m.history = append(m.history, metrics)
			// 如果超过最大历史记录数，删除最旧的记录
			if len(m.history) > m.maxHistory {
				m.history = m.history[1:]
			}
			m.mu.Unlock()
		case <-m.stopChan:
			return
		}
	}
}

// collectMetrics 采集当前指标
func (m *PerformanceMonitor) collectMetrics() PerformanceMetrics {
	var metrics PerformanceMetrics
	metrics.CollectedAt = time.Now()

	// 采集内存指标
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics.MemoryUsage = memStats.Sys
	metrics.MemoryAllocated = memStats.Alloc
	metrics.GCPauseNs = memStats.PauseNs[(memStats.NumGC+255)%256]
	metrics.GCCount = memStats.NumGC

	// 采集协程数量
	metrics.GoroutineCount = runtime.NumGoroutine()

	// 简单估算CPU使用率
	now := time.Now()
	duration := now.Sub(m.lastCPUSample)
	if duration.Seconds() > 0 {
		// 由于无法直接获取CPU使用率，这里使用一个简单的估算
		metrics.CPUUsage = float64(runtime.NumGoroutine()) / float64(runtime.NumCPU()) * 25
		// 限制最大值为100%
		if metrics.CPUUsage > 100 {
			metrics.CPUUsage = 100
		}
	}
	m.lastCPUSample = now

	return metrics
}

// FormatMetrics 格式化指标
func FormatMetrics(metrics PerformanceMetrics) string {
	return fmt.Sprintf(
		"CPU: %.2f%%, 内存: %s/%s, 协程: %d, GC暂停: %s, GC次数: %d",
		metrics.CPUUsage,
		FormatBytes(metrics.MemoryAllocated),
		FormatBytes(metrics.MemoryUsage),
		metrics.GoroutineCount,
		FormatDuration(time.Duration(metrics.GCPauseNs)),
		metrics.GCCount,
	)
}

// FormatBytes 格式化字节数
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration 格式化时间
func FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%d ns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.2f µs", float64(d.Nanoseconds())/1000)
	} else if d < time.Second {
		return fmt.Sprintf("%.2f ms", float64(d.Nanoseconds())/1000000)
	} else if d < time.Minute {
		return fmt.Sprintf("%.2f s", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.2f m", d.Minutes())
	}
	return fmt.Sprintf("%.2f h", d.Hours())
}

// GlobalMonitor 全局性能监控器
var GlobalMonitor *PerformanceMonitor

// InitGlobalMonitor 初始化全局性能监控器
func InitGlobalMonitor(interval time.Duration, maxHistory int) {
	GlobalMonitor = NewPerformanceMonitor(interval, maxHistory)
	GlobalMonitor.Start()
}

// StopGlobalMonitor 停止全局性能监控器
func StopGlobalMonitor() {
	if GlobalMonitor != nil {
		GlobalMonitor.Stop()
	}
}

// GetGlobalMetrics 获取全局指标
func GetGlobalMetrics() *PerformanceMetrics {
	if GlobalMonitor != nil {
		return GlobalMonitor.GetLatestMetrics()
	}
	return nil
}

// LogPerformance 记录性能指标
func LogPerformance() {
	if GlobalMonitor != nil {
		metrics := GlobalMonitor.GetLatestMetrics()
		if metrics != nil {
			Info("性能指标",
				"cpu", fmt.Sprintf("%.2f%%", metrics.CPUUsage),
				"memory", FormatBytes(metrics.MemoryAllocated),
				"goroutines", metrics.GoroutineCount,
				"gc_pause", FormatDuration(time.Duration(metrics.GCPauseNs)),
				"gc_count", metrics.GCCount,
			)
		}
	}
}
