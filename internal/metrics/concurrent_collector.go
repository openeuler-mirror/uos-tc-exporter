// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics

import (
	"strings"
	"sync"
	"time"

	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
	"gitee.com/openeuler/uos-tc-exporter/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ConcurrentCollector 并发收集器
type ConcurrentCollector struct {
	collectors []interfaces.MetricCollector
	poolSize   int
	timeout    time.Duration
	logger     *logrus.Logger
	metrics    *InternalMetrics
}

// NewConcurrentCollector 创建新的并发收集器
func NewConcurrentCollector(poolSize int, timeout time.Duration, logger *logrus.Logger) *ConcurrentCollector {
	if logger == nil {
		logger = logrus.StandardLogger()
	}

	return &ConcurrentCollector{
		collectors: make([]interfaces.MetricCollector, 0),
		poolSize:   poolSize,
		timeout:    timeout,
		logger:     logger,
		metrics:    NewInternalMetrics(logger),
	}
}

// AddCollector 添加收集器
func (cc *ConcurrentCollector) AddCollector(collector interfaces.MetricCollector) {
	cc.collectors = append(cc.collectors, collector)
}

// collectionResult 收集结果
type collectionResult struct {
	collectorName string
	metrics       []prometheus.Metric
	duration      time.Duration
	err           error
}

// CollectAll 并发收集所有指标
func (cc *ConcurrentCollector) CollectAll(ch chan<- prometheus.Metric) error {
	startTime := time.Now()

	if len(cc.collectors) == 0 {
		cc.logger.Warn("No collectors registered")
		return nil
	}

	// 创建带缓冲的通道来收集结果
	resultCh := make(chan collectionResult, len(cc.collectors))
	errorCh := make(chan error, len(cc.collectors))

	// 创建信号量控制并发数
	semaphore := make(chan struct{}, cc.poolSize)

	var wg sync.WaitGroup

	// 启动所有收集器
	for i, collector := range cc.collectors {
		if !collector.Enabled() {
			cc.logger.Debugf("Collector %s is disabled, skipping", collector.ID())
			continue
		}

		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量

		go func(idx int, col interfaces.MetricCollector) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量

			collectorStart := time.Now()
			collectorName := col.ID()

			// 创建收集器的指标通道
			collectorCh := make(chan prometheus.Metric, 100)

			// 启动收集器
			go func() {
				defer close(collectorCh)
				col.Collect(collectorCh)
			}()

			// 处理收集结果
			var collectedMetrics []prometheus.Metric
			var collectionErr error

		collectLoop:
			for {
				select {
				case metric, ok := <-collectorCh:
					if !ok {
						break collectLoop
					}
					collectedMetrics = append(collectedMetrics, metric)
				case <-time.After(cc.timeout):
					collectionErr = errors.New(
						errors.ErrCodeMetricsCollect,
						"collector timeout",
					).WithContext("collector", collectorName).WithContext("timeout", cc.timeout.String())
					break collectLoop
				}
			}

			duration := time.Since(collectorStart)

			// 记录收集结果
			result := collectionResult{
				collectorName: collectorName,
				metrics:       collectedMetrics,
				duration:      duration,
				err:           collectionErr,
			}

			if collectionErr != nil {
				cc.metrics.RecordCollection(collectorName, duration, false, "timeout")
				errorCh <- collectionErr
			} else {
				cc.metrics.RecordCollection(collectorName, duration, true, "")
				resultCh <- result
			}
		}(i, collector)
	}

	// 等待所有收集器完成
	go func() {
		wg.Wait()
		close(resultCh)
		close(errorCh)
	}()

	// 处理结果
	var totalErrors int
	var totalMetrics int

	// 处理成功的结果
	resultDone := false
	errorDone := false

	for !resultDone || !errorDone {
		select {
		case result, ok := <-resultCh:
			if !ok {
				resultDone = true
				continue
			}

			// 将收集到的指标发送到输出通道
			for _, metric := range result.metrics {
				ch <- metric
				totalMetrics++
			}

			cc.logger.Debugf("Collector %s completed in %v, collected %d metrics",
				result.collectorName, result.duration, len(result.metrics))

		case err, ok := <-errorCh:
			if !ok {
				errorDone = true
				continue
			}

			totalErrors++
			cc.logger.Warnf("Collector failed: %v", err)
		}
	}

	totalDuration := time.Since(startTime)

	// 记录总体收集统计
	cc.logger.Infof("Collection completed in %v: %d metrics from %d collectors, %d errors",
		totalDuration, totalMetrics, len(cc.collectors), totalErrors)

	cc.metrics.RecordCollection("total", totalDuration, totalErrors == 0, "overall")

	if totalErrors > 0 {
		return errors.NewWithContext(
			errors.ErrCodeMetricsCollect,
			"collection completed with errors",
			map[string]interface{}{
				"total_collectors": len(cc.collectors),
				"successful":       len(cc.collectors) - totalErrors,
				"errors":           totalErrors,
				"total_metrics":    totalMetrics,
				"duration":         totalDuration.String(),
			},
		)
	}

	return nil
}

// GetStats 获取收集统计信息
func (cc *ConcurrentCollector) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_collectors":   len(cc.collectors),
		"pool_size":          cc.poolSize,
		"timeout":            cc.timeout.String(),
		"enabled_collectors": cc.getEnabledCollectors(),
	}
}

// getEnabledCollectors 获取启用的收集器列表
func (cc *ConcurrentCollector) getEnabledCollectors() []string {
	var enabled []string
	for _, collector := range cc.collectors {
		if collector.Enabled() {
			enabled = append(enabled, collector.ID())
		}
	}
	return enabled
}

// SetPoolSize 设置并发池大小
func (cc *ConcurrentCollector) SetPoolSize(size int) {
	if size > 0 {
		cc.poolSize = size
	}
}

// SetTimeout 设置超时时间
func (cc *ConcurrentCollector) SetTimeout(timeout time.Duration) {
	if timeout > 0 {
		cc.timeout = timeout
	}
}

// EnableCollector 启用特定收集器
func (cc *ConcurrentCollector) EnableCollector(id string) error {
	for _, collector := range cc.collectors {
		if collector.ID() == id {
			collector.SetEnabled(true)
			return nil
		}
	}
	return errors.New(
		errors.ErrCodeMetricsCollect,
		"collector not found",
	).WithContext("collector_id", id)
}

// DisableCollector 禁用特定收集器
func (cc *ConcurrentCollector) DisableCollector(id string) error {
	for _, collector := range cc.collectors {
		if collector.ID() == id {
			collector.SetEnabled(false)
			return nil
		}
	}
	return errors.New(
		errors.ErrCodeMetricsCollect,
		"collector not found",
	).WithContext("collector_id", id)
}

// GetCollector 获取特定收集器
func (cc *ConcurrentCollector) GetCollector(id string) (interfaces.MetricCollector, bool) {
	for _, collector := range cc.collectors {
		if collector.ID() == id {
			return collector, true
		}
	}
	return nil, false
}

// OptimizePoolSize 根据收集器数量优化池大小
func (cc *ConcurrentCollector) OptimizePoolSize() {
	total := len(cc.collectors)

	// 根据收集器数量动态调整池大小
	switch {
	case total <= 2:
		cc.poolSize = 1
	case total <= 5:
		cc.poolSize = 2
	case total <= 10:
		cc.poolSize = 4
	case total <= 20:
		cc.poolSize = 8
	default:
		cc.poolSize = 16
	}

	cc.logger.Infof("Optimized pool size to %d for %d collectors", cc.poolSize, total)
}

// BatchCollect 批量收集（按类型分组）
func (cc *ConcurrentCollector) BatchCollect(ch chan<- prometheus.Metric, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 5 // 默认批次大小
	}

	// 按类型分组收集器
	collectorGroups := cc.groupCollectorsByType()

	var wg sync.WaitGroup
	errorCh := make(chan error, len(collectorGroups))

	// 分批处理
	for groupName, collectors := range collectorGroups {
		if len(collectors) == 0 {
			continue
		}

		wg.Add(1)

		go func(group string, cols []interfaces.MetricCollector) {
			defer wg.Done()

			// 为每个组创建子收集器
			subCollector := NewConcurrentCollector(cc.poolSize, cc.timeout, cc.logger)
			for _, col := range cols {
				subCollector.AddCollector(col)
			}

			// 收集该组的指标
			if err := subCollector.CollectAll(ch); err != nil {
				errorCh <- errors.Wrap(
					err,
					errors.ErrCodeMetricsCollect,
					"batch collection failed",
				).WithContext("group", group)
			}
		}(groupName, collectors)
	}

	// 等待所有批次完成
	go func() {
		wg.Wait()
		close(errorCh)
	}()

	// 收集错误
	var errorsList []error
	for err := range errorCh {
		errorsList = append(errorsList, err)
	}

	if len(errorsList) > 0 {
		return errors.NewWithContext(
			errors.ErrCodeMetricsCollect,
			"batch collection completed with errors",
			map[string]interface{}{
				"total_batches": len(collectorGroups),
				"errors":        len(errorsList),
			},
		)
	}

	return nil
}

// groupCollectorsByType 按类型分组收集器
func (cc *ConcurrentCollector) groupCollectorsByType() map[string][]interfaces.MetricCollector {
	groups := make(map[string][]interfaces.MetricCollector)

	for _, collector := range cc.collectors {
		// 根据收集器ID推断类型（例如："qdisc_htb" -> "qdisc"）
		collectorType := inferCollectorType(collector.ID())
		groups[collectorType] = append(groups[collectorType], collector)
	}

	return groups
}

// inferCollectorType 从收集器ID推断类型
func inferCollectorType(id string) string {
	// 简单的类型推断逻辑
	if strings.HasPrefix(id, "qdisc_") {
		return "qdisc"
	} else if strings.HasPrefix(id, "class_") {
		return "class"
	} else if strings.HasPrefix(id, "app_") {
		return "app"
	} else if strings.HasPrefix(id, "business_") {
		return "business"
	}
	return "other"
}

// HealthCheck 健康检查
func (cc *ConcurrentCollector) HealthCheck() map[string]interface{} {
	health := make(map[string]interface{})

	var enabledCount int
	// var totalMetrics int

	for _, collector := range cc.collectors {
		if collector.Enabled() {
			enabledCount++
			// 这里可以添加更详细的健康检查逻辑
		}
	}

	health["total_collectors"] = len(cc.collectors)
	health["enabled_collectors"] = enabledCount
	health["pool_size"] = cc.poolSize
	health["timeout"] = cc.timeout.String()
	health["status"] = "healthy"

	if enabledCount == 0 {
		health["status"] = "warning"
		health["message"] = "No collectors enabled"
	}

	return health
}
