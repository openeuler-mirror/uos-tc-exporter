// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/config"
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/core/interfaces"
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/registry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type ManagerV2 struct {
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
		logger = logrus.StandardLogger()
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
	qdiscFactory := registry.NewQdiscFactory()

	qdiscFactory.AddConfig("qdisc", m.getQdiscConfig())
	m.factories["qdisc"] = qdiscFactory
	err := m.registry.RegisterFactory("qdisc", qdiscFactory)
	if err != nil {
		m.logger.Errorf("Failed to register Qdisc factory: %v", err)
	}
	// Add other factories as needed
}

func (m *ManagerV2) getQdiscConfig() *config.CollectorConfig {
	mc := map[string]config.MetricConfig{
		"bytes_total":      *config.NewMetricConfig("bytes_total", "QdiscPie byte counter", "qdisc"),
		"packets_total":    *config.NewMetricConfig("packets_total", "QdiscPie packet counter", "qdisc"),
		"drops_total":      *config.NewMetricConfig("drops_total", "QdiscPie queue drops", "qdisc"),
		"overlimits_total": *config.NewMetricConfig("overlimits", "QdiscPie queue overlimits", "qdisc"),
		"bps":              *config.NewMetricConfig("bps", "QdiscPie bytes per second", "qdisc"),
		"pps":              *config.NewMetricConfig("pps", "QdiscPie packets per second", "qdisc"),
		"qlen":             *config.NewMetricConfig("qlen", "QdiscPie current queue length", "qdisc"),
		"backlog":          *config.NewMetricConfig("backlog", "QdiscPie current backlog in bytes", "qdisc"),
		"requeues_total":   *config.NewMetricConfig("requeues_total", "QdiscPie number of requeues", "qdisc"),
	}
	cfg := config.NewCollectorConfig()
	cfg.Metrics = mc
	return cfg
}

func (m *ManagerV2) registerCollectors() {
	// 注册 qdisc 收集器
	qdiscTypes := []string{"codel", "cbq", "htb", "fq", "fq_codel", "choke", "pie", "red", "sfb", "sfq", "hfsc", "qdisc"}
	for _, qdiscType := range qdiscTypes {
		collector, err := m.registry.CreateCollector("qdisc", qdiscType)
		if err == nil {
			_ = m.registry.Register(collector)
		} else {
			m.logger.Warnf("Failed to create qdisc collector %s: %v", qdiscType, err)
		}
	}
}

func (m *ManagerV2) GetStats() *CollectionStats {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()
	// fix: returning a copy of stats can lead to issues with the embedded RWMutex
	return m.stats
}
func (m *ManagerV2) Shutdown() {
	m.logger.Info("Shutting down ManagerV2")
}

// CollectAll 收集所有指标
func (m *ManagerV2) CollectAll(ch chan<- prometheus.Metric) {
	m.CollectAllWithContext(context.Background(), ch)
}

// CollectAllWithContext 使用上下文收集所有指标
func (m *ManagerV2) CollectAllWithContext(ctx context.Context, ch chan<- prometheus.Metric) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		m.updateStats(duration, true, nil)
		m.logger.Debugf("Collection completed in %v", duration)
	}()

	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		m.logger.Warnf("Collection cancelled due to context: %v", err)
		return
	}

	collectors := m.registry.GetEnableCollectors()
	if len(collectors) == 0 {
		m.logger.Warn("No enabled collectors found")
		return
	}

	// 使用并发收集提高性能
	m.collectConcurrently(ctx, ch, collectors)
}

// collectConcurrently 并发收集指标
func (m *ManagerV2) collectConcurrently(ctx context.Context, ch chan<- prometheus.Metric, collectors []interfaces.MetricCollector) {
	start := time.Now()
	var wg sync.WaitGroup
	collectorCh := make(chan interfaces.MetricCollector, len(collectors))
	errorCh := make(chan error, len(collectors))

	// 启动固定数量的工作goroutine
	numWorkers := min(len(collectors), 10) // 限制最大并发数
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go m.collectWorker(ctx, &wg, collectorCh, ch, errorCh)
	}

	// 分发收集器到工作队列
	for _, collector := range collectors {
		collectorCh <- collector
	}
	close(collectorCh)

	// 等待所有工作完成
	wg.Wait()
	close(errorCh)

	// 处理收集错误
	var errors []error
	for err := range errorCh {
		if err != nil {
			errors = append(errors, err)
			m.logger.Errorf("Collector error: %v", err)
		}
	}

	if len(errors) > 0 {
		m.updateStats(time.Since(start), false, fmt.Errorf("%d collectors failed", len(errors)))
	}
}

// collectWorker 单个收集工作goroutine
func (m *ManagerV2) collectWorker(ctx context.Context, wg *sync.WaitGroup,
	collectorCh <-chan interfaces.MetricCollector,
	ch chan<- prometheus.Metric, errorCh chan<- error) {
	defer wg.Done()

	for collector := range collectorCh {
		// 检查上下文是否已取消
		if err := ctx.Err(); err != nil {
			m.logger.Debugf("Collection cancelled for worker: %v", err)
			return
		}

		collectorStart := time.Now()
		m.logger.Debugf("Collecting from collector: %s", collector.ID())

		// 使用安全执行避免panic
		err := m.safeCollect(collector, ch)
		collectorDuration := time.Since(collectorStart)

		if err != nil {
			errorCh <- fmt.Errorf("collector %s failed after %v: %w", collector.ID(), collectorDuration, err)
		} else {
			m.logger.Debugf("Collector %s completed in %v", collector.ID(), collectorDuration)
		}
	}
}

// safeCollect 安全执行收集操作
func (m *ManagerV2) safeCollect(collector interfaces.MetricCollector, ch chan<- prometheus.Metric) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during collection: %v", r)
		}
	}()

	collector.Collect(ch)
	return nil
}

// updateStats 更新收集统计信息
func (m *ManagerV2) updateStats(duration time.Duration, success bool, err error) {
	m.stats.mu.Lock()
	defer m.stats.mu.Unlock()

	m.stats.TotalCollections++
	m.stats.TotalDuration += duration

	if success {
		m.stats.SuccessfulCollections++
	} else {
		m.stats.FailedCollections++
		m.stats.LastError = err
		m.stats.LastErrorTime = time.Now()
	}

	// 计算平均持续时间
	if m.stats.TotalCollections > 0 {
		m.stats.AverageDuration = m.stats.TotalDuration / time.Duration(m.stats.TotalCollections)
	}

	m.stats.LastCollectionTime = time.Now()
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
