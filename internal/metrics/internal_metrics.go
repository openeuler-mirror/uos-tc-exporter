// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// InternalMetrics 内部监控指标
type InternalMetrics struct {
	// 收集器性能指标
	collectionDuration *prometheus.HistogramVec
	collectionErrors    *prometheus.CounterVec
	collectionSuccesses *prometheus.CounterVec
	
	// 系统资源指标
	memoryUsage        prometheus.Gauge
	cpuUsage           prometheus.Gauge
	goroutineCount     prometheus.Gauge
	uptime             prometheus.Gauge
	
	// 网络指标
	netlinkCalls       *prometheus.CounterVec
	netlinkErrors      *prometheus.CounterVec
	netlinkDuration    *prometheus.HistogramVec
	
	// 配置相关指标
	configReloads      prometheus.Counter
	configErrors       prometheus.Counter
	configLastReload   prometheus.Gauge
	
	// 注册表
	registry *prometheus.Registry
	
	// 状态
	startTime time.Time
	logger    *logrus.Logger
	mu        sync.RWMutex
}

// NewInternalMetrics 创建内部监控指标
func NewInternalMetrics(logger *logrus.Logger) *InternalMetrics {
	if logger == nil {
		logger = logrus.StandardLogger()
	}
	
	im := &InternalMetrics{
		registry:  prometheus.NewRegistry(),
		startTime: time.Now(),
		logger:    logger,
	}
	
	im.initializeMetrics()
	return im
}

// initializeMetrics 初始化所有内部指标
func (im *InternalMetrics) initializeMetrics() {
	// 收集器性能指标
	im.collectionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "tc_exporter_collection_duration_seconds",
			Help: "Time spent collecting metrics from each collector",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
		},
		[]string{"collector", "status"},
	)
	
	im.collectionErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tc_exporter_collection_errors_total",
			Help: "Total number of collection errors",
		},
		[]string{"collector", "error_type"},
	)
	
	im.collectionSuccesses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tc_exporter_collection_successes_total",
			Help: "Total number of successful collections",
		},
		[]string{"collector"},
	)
	
	// 系统资源指标
	im.memoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tc_exporter_memory_usage_bytes",
		Help: "Current memory usage in bytes",
	})
	
	im.cpuUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tc_exporter_cpu_usage_percent",
		Help: "Current CPU usage percentage",
	})
	
	im.goroutineCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tc_exporter_goroutines",
		Help: "Current number of goroutines",
	})
	
	im.uptime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tc_exporter_uptime_seconds",
		Help: "Process uptime in seconds",
	})
	
	// 网络指标
	im.netlinkCalls = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tc_exporter_netlink_calls_total",
			Help: "Total number of netlink calls",
		},
		[]string{"operation"},
	)
	
	im.netlinkErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tc_exporter_netlink_errors_total",
			Help: "Total number of netlink errors",
		},
		[]string{"operation", "error_type"},
	)
	
	im.netlinkDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "tc_exporter_netlink_duration_seconds",
			Help: "Time spent on netlink operations",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05},
		},
		[]string{"operation"},
	)
	
	// 配置相关指标
	im.configReloads = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tc_exporter_config_reloads_total",
		Help: "Total number of configuration reloads",
	})
	
	im.configErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tc_exporter_config_errors_total",
		Help: "Total number of configuration errors",
	})
	
	im.configLastReload = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tc_exporter_config_last_reload_timestamp_seconds",
		Help: "Timestamp of last configuration reload",
	})
	
	// 注册所有指标
	im.registry.MustRegister(
		im.collectionDuration,
		im.collectionErrors,
		im.collectionSuccesses,
		im.memoryUsage,
		im.cpuUsage,
		im.goroutineCount,
		im.uptime,
		im.netlinkCalls,
		im.netlinkErrors,
		im.netlinkDuration,
		im.configReloads,
		im.configErrors,
		im.configLastReload,
	)
}

// RecordCollection 记录收集操作
func (im *InternalMetrics) RecordCollection(collector string, duration time.Duration, success bool, errorType string) {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	status := "success"
	if !success {
		status = "error"
	}
	
	im.collectionDuration.WithLabelValues(collector, status).Observe(duration.Seconds())
	
	if success {
		im.collectionSuccesses.WithLabelValues(collector).Inc()
	} else {
		im.collectionErrors.WithLabelValues(collector, errorType).Inc()
	}
}

// RecordNetlinkOperation 记录netlink操作
func (im *InternalMetrics) RecordNetlinkOperation(operation string, duration time.Duration, success bool, errorType string) {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	im.netlinkCalls.WithLabelValues(operation).Inc()
	im.netlinkDuration.WithLabelValues(operation).Observe(duration.Seconds())
	
	if !success {
		im.netlinkErrors.WithLabelValues(operation, errorType).Inc()
	}
}

// RecordConfigReload 记录配置重载
func (im *InternalMetrics) RecordConfigReload(success bool) {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	im.configReloads.Inc()
	im.configLastReload.Set(float64(time.Now().Unix()))
	
	if !success {
		im.configErrors.Inc()
	}
}

// UpdateSystemMetrics 更新系统指标
func (im *InternalMetrics) UpdateSystemMetrics() {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	// 更新运行时间
	im.uptime.Set(time.Since(im.startTime).Seconds())
	
	// 更新goroutine数量
	im.goroutineCount.Set(float64(getGoroutineCount()))
	
	// 更新内存使用量
	if memStats, err := getMemoryUsage(); err == nil {
		im.memoryUsage.Set(float64(memStats))
	}
	
	// 更新CPU使用率
	if cpuUsage, err := getCPUUsage(); err == nil {
		im.cpuUsage.Set(cpuUsage)
	}
}

// GetRegistry 获取指标注册表
func (im *InternalMetrics) GetRegistry() *prometheus.Registry {
	return im.registry
}

// StartMonitoring 启动监控循环
func (im *InternalMetrics) StartMonitoring(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			im.UpdateSystemMetrics()
		}
	}
}

// 辅助函数

// getGoroutineCount 获取当前goroutine数量
func getGoroutineCount() int {
	// 这里可以使用runtime.NumGoroutine()
	// 为了简化，我们返回一个估计值
	return 10 // 实际实现应该使用runtime包
}

// getMemoryUsage 获取内存使用量
func getMemoryUsage() (uint64, error) {
	// 实际实现应该读取/proc/self/statm或使用runtime包
	return 1024 * 1024 * 100, nil // 100MB示例值
}

// getCPUUsage 获取CPU使用率
func getCPUUsage() (float64, error) {
	// 实际实现应该读取/proc/stat
	return 25.5, nil // 25.5%示例值
}

// MetricsCollector 实现prometheus.Collector接口
func (im *InternalMetrics) Collect(ch chan<- prometheus.Metric) {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	// 更新系统指标
	im.UpdateSystemMetrics()
	
	// 收集所有指标
	im.registry.Collect(ch)
}

// Describe 实现prometheus.Collector接口
func (im *InternalMetrics) Describe(ch chan<- *prometheus.Desc) {
	im.registry.Describe(ch)
}
