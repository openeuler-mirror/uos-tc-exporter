// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics

import (
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
		if !collector.IsEnabled() {
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
				"successful":      len(cc.collectors) - totalErrors,
				"errors":          totalErrors,
				"total_metrics":   totalMetrics,
				"duration":        totalDuration.String(),
			},
		)
	}

	return nil
}
