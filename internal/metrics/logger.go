// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MetricsLogger 提供优化的指标收集日志输出
type MetricsLogger struct {
	mu           sync.RWMutex
	lastLogTime  map[string]time.Time
	logInterval  time.Duration
	metricsCount map[string]int
	errorCount   map[string]int
}

var (
	globalLogger *MetricsLogger
	once         sync.Once
)

// GetMetricsLogger 获取全局指标日志记录器实例
func GetMetricsLogger() *MetricsLogger {
	once.Do(func() {
		globalLogger = &MetricsLogger{
			lastLogTime:  make(map[string]time.Time),
			logInterval:  5 * time.Second, // 5秒内不重复输出相同类型的日志
			metricsCount: make(map[string]int),
			errorCount:   make(map[string]int),
		}
	})
	return globalLogger
}

// LogCollectionStart 记录指标收集开始（减少重复日志）
func (ml *MetricsLogger) LogCollectionStart(collectorType string) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	now := time.Now()
	if lastTime, exists := ml.lastLogTime[collectorType]; !exists || now.Sub(lastTime) > ml.logInterval {
		logrus.Debugf("Starting %s metrics collection", collectorType)
		ml.lastLogTime[collectorType] = now
		ml.metricsCount[collectorType] = 0
		ml.errorCount[collectorType] = 0
	}
}

// LogCollectionComplete 记录指标收集完成（包含统计信息）
func (ml *MetricsLogger) LogCollectionComplete(collectorType string) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	if count, exists := ml.metricsCount[collectorType]; exists {
		logrus.Debugf("%s metrics collection completed, collected %d metrics", collectorType, count)
	}
}

// LogError 记录错误（避免重复错误日志）
func (ml *MetricsLogger) LogError(collectorType, operation string, err error) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	key := collectorType + ":" + operation
	now := time.Now()

	if lastTime, exists := ml.lastLogTime[key]; !exists || now.Sub(lastTime) > ml.logInterval {
		logrus.Warnf("%s %s failed: %v", collectorType, operation, err)
		ml.lastLogTime[key] = now
		ml.errorCount[collectorType]++
	}
}

// LogNoData 记录无数据情况（减少重复日志）
func (ml *MetricsLogger) LogNoData(collectorType, reason string) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	key := collectorType + ":no_data:" + reason
	now := time.Now()

	if lastTime, exists := ml.lastLogTime[key]; !exists || now.Sub(lastTime) > ml.logInterval {
		logrus.Debugf("%s: %s", collectorType, reason)
		ml.lastLogTime[key] = now
	}
}

// IncrementMetricsCount 增加指标计数
func (ml *MetricsLogger) IncrementMetricsCount(collectorType string, count int) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	if _, exists := ml.metricsCount[collectorType]; exists {
		ml.metricsCount[collectorType] += count
	}
}

// GetStats 获取收集统计信息
func (ml *MetricsLogger) GetStats() map[string]interface{} {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	stats := make(map[string]interface{})
	for collectorType, count := range ml.metricsCount {
		stats[collectorType+"_metrics"] = count
	}
	for collectorType, count := range ml.errorCount {
		stats[collectorType+"_errors"] = count
	}
	return stats
}

// SetLogInterval 设置日志输出间隔
func (ml *MetricsLogger) SetLogInterval(interval time.Duration) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.logInterval = interval
}
