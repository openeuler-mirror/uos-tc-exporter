// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics

import (
	"fmt"
	"sync"
	"time"

	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/config"
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/factories"
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/registry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type ManagerV2 struct {
	mu        sync.RWMutex
	registry  *registry.CollectorRegistry
	factories map[string]registry.CollectorFactory
	config    *config.ManagerConfig
	stats     *CollectionStats
	logger    *logrus.Logger
	// Add fields as necessary
}

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

func NewManagerV2(cfg *config.ManagerConfig, logger *logrus.Logger) *ManagerV2 {
	defaultCfg := config.ManagerConfig{
		PerformanceMonitoring: true,
		CollectionInterval:    30 * time.Second,
		StatsRetention:        24 * time.Hour,
		EnableBusinessMetrics: true,
	}
	if cfg == nil {
		cfg = &defaultCfg
	}
	if logger == nil {
		logger = logrus.New()
	}
	m := &ManagerV2{
		registry:  registry.NewCollectorRegistry(),
		factories: make(map[string]registry.CollectorFactory),
		config:    cfg,
		stats:     &CollectionStats{},
		logger:    logger,
	}
	// Additional initialization logic can be added here
	m.initializeFactories()
	m.registerCollectors()
	return m
}

func (m *ManagerV2) initializeFactories() {
	// Initialize and register different factories
	m.logger.Info("Initializing Qdisc Factory")
	qdiscFactory := factories.NewQdiscFactory()
	m.factories["qdisc"] = qdiscFactory
	m.registry.RegisterFactory("qdisc", qdiscFactory)
	// Add other factories as needed
}

func (m *ManagerV2) registerCollectors() {
	// 注册 qdisc 收集器
	qdiscTypes := []string{"codel", "cbq", "htb", "fq", "fq_codel", "choke", "pie", "red", "sfb", "sfq", "hfsc"}
	for _, qdiscType := range qdiscTypes {
		collector, err := m.registry.CreateCollector("qdisc", qdiscType)
		if err == nil {
			m.registry.Register(collector)
		} else {
			m.logger.Warnf("Failed to create qdisc collector %s: %v", qdiscType, err)
		}
	}
}

func (m *ManagerV2) GetStats() CollectionStats {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()
	return *m.stats
}
func (m *ManagerV2) Shutdown() {
	m.logger.Info("Shutting down ManagerV2")
}

// CollectAll 收集所有指标
func (m *ManagerV2) CollectAll(ch chan<- prometheus.Metric) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		// m.stats.RecordCollection(duration, true, nil)
		fmt.Printf("Collection took %v\n", duration)
	}()
	collectors := m.registry.GetEnableCollectors()
	for _, collector := range collectors {
		collector.Collect(ch)
	}

}

// GetCollector 获取收集器
func (m *ManagerV2) GetCollector(id string) (interfaces.MetricCollector, bool) {
	return m.registry.GetCollector(id)
}

// EnableCollector 启用收集器
func (m *ManagerV2) EnableCollector(id string) error {
	collector, exists := m.registry.GetCollector(id)
	if !exists {
		return fmt.Errorf("collector %s not found", id)
	}

	collector.SetEnabled(true)
	return nil
}

// DisableCollector 禁用收集器
func (m *ManagerV2) DisableCollector(id string) error {
	collector, exists := m.registry.GetCollector(id)
	if !exists {
		return fmt.Errorf("collector %s not found", id)
	}

	collector.SetEnabled(false)
	return nil
}
