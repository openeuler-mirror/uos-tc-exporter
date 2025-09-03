// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Manager 监控指标管理器
// 负责统一管理所有监控指标，提供统一的接口
type Manager struct {
	mu sync.RWMutex

	// 指标收集器
	appMetrics      *AppMetrics
	businessMetrics *BusinessMetrics

	collectionStats *CollectionStats

	// 配置
	config *ManagerConfig
}

// ManagerConfig 管理器配置
type ManagerConfig struct {
	PerformanceMonitoring bool          `yaml:"performance_monitoring"`
	CollectionInterval    time.Duration `yaml:"collection_interval"`
	StatsRetention        time.Duration `yaml:"stats_retention"`
	EnableBusinessMetrics bool          `yaml:"enable_business_metrics"`
}

// CollectionStats 收集统计信息
type CollectionStats struct {
	mu sync.RWMutex

	TotalCollections      int64
	SuccessfulCollections int64
	FailedCollections     int64
	TotalDuration         time.Duration
	AverageDuration       time.Duration
	LastCollectionTime    time.Time
	LastErrorTime         time.Time
	LastError             error
}

// NewManager 创建新的监控指标管理器
func NewManager(config *ManagerConfig) *Manager {
	if config == nil {
		config = &ManagerConfig{
			PerformanceMonitoring: true,
			CollectionInterval:    30 * time.Second,
			StatsRetention:        24 * time.Hour,
			EnableBusinessMetrics: true,
		}
	}

	manager := &Manager{
		config:          config,
		collectionStats: &CollectionStats{},
	}

	// 初始化指标
	manager.initializeMetrics()

	// 启动后台任务
	if config.PerformanceMonitoring {
		go manager.startBackgroundTasks()
	}

	return manager
}

// initializeMetrics 初始化监控指标
func (manager *Manager) initializeMetrics() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// 应用指标
	manager.appMetrics = NewAppMetrics()

	// 业务指标（可选）
	if manager.config.EnableBusinessMetrics {
		manager.businessMetrics = NewBusinessMetrics()
	}

	logrus.Info("Metrics manager initialized")
}

// startBackgroundTasks 启动后台任务
func (manager *Manager) startBackgroundTasks() {
	ticker := time.NewTicker(manager.config.CollectionInterval)
	defer ticker.Stop()

	for range ticker.C {
		manager.updateStats()
		manager.cleanupOldStats()
	}
}

// updateStats 更新统计信息
func (manager *Manager) updateStats() {
	manager.collectionStats.mu.Lock()
	defer manager.collectionStats.mu.Unlock()

	// 更新平均持续时间
	if manager.collectionStats.TotalCollections > 0 {
		manager.collectionStats.AverageDuration = manager.collectionStats.TotalDuration / time.Duration(manager.collectionStats.TotalCollections)
	}
}

// cleanupOldStats 清理旧统计信息
func (manager *Manager) cleanupOldStats() {
	manager.collectionStats.mu.Lock()
	defer manager.collectionStats.mu.Unlock()

	cutoff := time.Now().Add(-manager.config.StatsRetention)

	// 如果最后收集时间太旧，重置统计
	if manager.collectionStats.LastCollectionTime.Before(cutoff) {
		manager.collectionStats.TotalCollections = 0
		manager.collectionStats.SuccessfulCollections = 0
		manager.collectionStats.FailedCollections = 0
		manager.collectionStats.TotalDuration = 0
		manager.collectionStats.AverageDuration = 0
		logrus.Debug("Collection stats reset due to retention policy")
	}
}

// RecordCollection 记录指标收集
func (manager *Manager) RecordCollection(duration time.Duration, success bool, err error) {
	manager.collectionStats.mu.Lock()
	defer manager.collectionStats.mu.Unlock()

	manager.collectionStats.TotalCollections++
	manager.collectionStats.TotalDuration += duration
	manager.collectionStats.LastCollectionTime = time.Now()

	if success {
		manager.collectionStats.SuccessfulCollections++
	} else {
		manager.collectionStats.FailedCollections++
		manager.collectionStats.LastErrorTime = time.Now()
		manager.collectionStats.LastError = err
	}

	// 更新应用指标
	if manager.appMetrics != nil {
		manager.appMetrics.RecordMetricsCollectionDuration(duration)
		manager.appMetrics.IncrementMetricsCollectionTotal()
		if !success {
			manager.appMetrics.IncrementMetricsCollectionErrors()
		}
	}
}

// GetStats 获取统计信息
func (manager *Manager) GetStats() *CollectionStats {
	manager.collectionStats.mu.RLock()
	defer manager.collectionStats.mu.RUnlock()

	// 返回副本以避免并发访问问题
	stats := &CollectionStats{
		TotalCollections:      manager.collectionStats.TotalCollections,
		SuccessfulCollections: manager.collectionStats.SuccessfulCollections,
		FailedCollections:     manager.collectionStats.FailedCollections,
		TotalDuration:         manager.collectionStats.TotalDuration,
		AverageDuration:       manager.collectionStats.AverageDuration,
		LastCollectionTime:    manager.collectionStats.LastCollectionTime,
		LastErrorTime:         manager.collectionStats.LastErrorTime,
		LastError:             manager.collectionStats.LastError,
	}

	return stats
}

// UpdateTCStats 更新TC统计信息
func (manager *Manager) UpdateTCStats(namespaces, interfaces, qdiscs, classes int) {
	if manager.businessMetrics != nil {
		UpdateTCStats(namespaces, interfaces, qdiscs, classes)
	}
}

// UpdateNetworkThroughput 更新网络吞吐量
func (manager *Manager) UpdateNetworkThroughput(throughput float64) {
	if manager.businessMetrics != nil {
		UpdateNetworkThroughput(throughput)
	}
}

// RecordNetworkLatency 记录网络延迟
func (manager *Manager) RecordNetworkLatency(duration float64) {
	if manager.businessMetrics != nil {
		manager.businessMetrics.RecordNetworkLatency(duration)
	}
}

// IncrementNetworkErrors 增加网络错误计数
func (manager *Manager) IncrementNetworkErrors() {
	if manager.businessMetrics != nil {
		manager.businessMetrics.IncrementNetworkErrors()
	}
}

// SetServiceHealth 设置服务健康状态
func (manager *Manager) SetServiceHealth(serviceName, component string, healthy bool) {
	if manager.businessMetrics != nil {
		manager.businessMetrics.SetServiceHealth(serviceName, component, healthy)
	}
}

// IncrementConfigReload 增加配置重载计数
func (manager *Manager) IncrementConfigReload() {
	if manager.businessMetrics != nil {
		manager.businessMetrics.IncrementConfigReload()
	}
}

// IncrementConfigReloadErrors 增加配置重载错误计数
func (manager *Manager) IncrementConfigReloadErrors() {
	if manager.businessMetrics != nil {
		manager.businessMetrics.IncrementConfigReloadErrors()
	}
}

// UpdateConfigVersion 更新配置版本信息
func (manager *Manager) UpdateConfigVersion(configFile, version, hash string) {
	if manager.businessMetrics != nil {
		manager.businessMetrics.UpdateConfigVersion(configFile, version, hash)
	}
}

// GetAppMetrics 获取应用指标实例
func (manager *Manager) GetAppMetrics() *AppMetrics {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return manager.appMetrics
}

// GetBusinessMetrics 获取业务指标实例
func (manager *Manager) GetBusinessMetrics() *BusinessMetrics {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return manager.businessMetrics
}

// IsPerformanceMonitoringEnabled 检查是否启用了性能监控
func (manager *Manager) IsPerformanceMonitoringEnabled() bool {
	return manager.config.PerformanceMonitoring
}

// Shutdown 关闭管理器
func (manager *Manager) Shutdown() {
	logrus.Info("Shutting down metrics manager")
	// 这里可以添加清理逻辑
}

// 全局管理器实例
var globalManager *Manager
var globalManagerOnce sync.Once

// GetGlobalManager 获取全局管理器实例
func GetGlobalManager() *Manager {
	globalManagerOnce.Do(func() {
		globalManager = NewManager(nil)
	})
	return globalManager
}

// InitializeGlobalManager 初始化全局管理器
func InitializeGlobalManager(config *ManagerConfig) {
	globalManagerOnce.Do(func() {
		globalManager = NewManager(config)
	})
}
