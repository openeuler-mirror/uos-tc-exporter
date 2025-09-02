// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics

import (
	"runtime"
	"time"

	"gitee.com/openeuler/uos-tc-exporter/internal/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func init() {
	exporter.Register(NewAppMetrics())
}

// AppMetrics 应用级别的监控指标
type AppMetrics struct {
	// 应用信息指标
	appInfo *prometheus.GaugeVec

	// 运行时指标
	goGoroutines   prometheus.GaugeFunc
	goThreads      prometheus.GaugeFunc
	goHeapAlloc    prometheus.GaugeFunc
	goHeapSys      prometheus.GaugeFunc
	goHeapIdle     prometheus.GaugeFunc
	goHeapInuse    prometheus.GaugeFunc
	goHeapReleased prometheus.GaugeFunc
	goHeapObjects  prometheus.GaugeFunc

	// 性能指标
	metricsCollectionDuration prometheus.Histogram
	metricsCollectionTotal    prometheus.Counter
	metricsCollectionErrors   prometheus.Counter

	// 系统指标
	systemUptime     prometheus.GaugeFunc
	processStartTime prometheus.GaugeFunc

	// 自定义指标
	customMetricsCount prometheus.GaugeFunc
	errorRate          prometheus.GaugeFunc
}

// NewAppMetrics 创建新的应用监控指标实例
func NewAppMetrics() *AppMetrics {
	am := &AppMetrics{
		// 应用信息指标
		appInfo: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "app_info",
				Help: "Application information",
			},
			[]string{"version", "build_time", "go_version"},
		),

		// 运行时指标
		goGoroutines: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "app_go_goroutines",
				Help: "Number of goroutines that currently exist (app-specific)",
			},
			func() float64 { return float64(runtime.NumGoroutine()) },
		),

		goThreads: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "app_go_threads",
				Help: "Number of OS threads created (app-specific)",
			},
			func() float64 { return float64(runtime.GOMAXPROCS(0)) },
		),

		goHeapAlloc: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "app_go_heap_alloc_bytes",
				Help: "Heap memory usage: bytes allocated (app-specific)",
			},
			func() float64 {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				return float64(m.HeapAlloc)
			},
		),

		goHeapSys: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "app_go_heap_sys_bytes",
				Help: "Heap memory usage: bytes obtained from system (app-specific)",
			},
			func() float64 {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				return float64(m.HeapSys)
			},
		),

		goHeapIdle: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "app_go_heap_idle_bytes",
				Help: "Heap memory usage: bytes in idle spans (app-specific)",
			},
			func() float64 {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				return float64(m.HeapIdle)
			},
		),

		goHeapInuse: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "app_go_heap_inuse_bytes",
				Help: "Heap memory usage: bytes in in-use spans (app-specific)",
			},
			func() float64 {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				return float64(m.HeapInuse)
			},
		),

		goHeapReleased: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "app_go_heap_released_bytes",
				Help: "Heap memory usage: bytes released to OS (app-specific)",
			},
			func() float64 {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				return float64(m.HeapReleased)
			},
		),

		goHeapObjects: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "app_go_heap_objects",
				Help: "Heap memory usage: total number of allocated objects (app-specific)",
			},
			func() float64 {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				return float64(m.HeapObjects)
			},
		),

		// 性能指标
		metricsCollectionDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "metrics_collection_duration_seconds",
				Help:    "Time spent collecting metrics",
				Buckets: prometheus.DefBuckets,
			},
		),

		metricsCollectionTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "metrics_collection_total",
				Help: "Total number of metrics collections",
			},
		),

		metricsCollectionErrors: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "metrics_collection_errors_total",
				Help: "Total number of metrics collection errors",
			},
		),

		// 系统指标
		systemUptime: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "app_system_uptime_seconds",
				Help: "System uptime in seconds (app-specific)",
			},
			func() float64 { return float64(time.Now().Unix()) },
		),

		processStartTime: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "app_process_start_time_seconds",
				Help: "Start time of the process since unix epoch in seconds (app-specific)",
			},
			func() float64 { return float64(processStartTime) },
		),

		// 自定义指标
		customMetricsCount: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "custom_metrics_count",
				Help: "Number of custom metrics registered",
			},
			func() float64 { return float64(getCustomMetricsCount()) },
		),

		errorRate: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "error_rate",
				Help: "Current error rate (0-1)",
			},
			func() float64 { return getErrorRate() },
		),
	}

	// 设置应用信息
	am.setAppInfo()

	return am
}

// Collect 实现Metric接口
func (am *AppMetrics) Collect(ch chan<- prometheus.Metric) {
	// 收集所有指标
	am.appInfo.Collect(ch)
	am.goGoroutines.Collect(ch)
	am.goThreads.Collect(ch)
	am.goHeapAlloc.Collect(ch)
	am.goHeapSys.Collect(ch)
	am.goHeapIdle.Collect(ch)
	am.goHeapInuse.Collect(ch)
	am.goHeapReleased.Collect(ch)
	am.goHeapObjects.Collect(ch)
	am.metricsCollectionDuration.Collect(ch)
	am.metricsCollectionTotal.Collect(ch)
	am.metricsCollectionErrors.Collect(ch)
	am.systemUptime.Collect(ch)
	am.processStartTime.Collect(ch)
	am.customMetricsCount.Collect(ch)
	am.errorRate.Collect(ch)
}

// ID 实现Metric接口
func (am *AppMetrics) ID() string {
	return "app_metrics"
}

// setAppInfo 设置应用信息
func (am *AppMetrics) setAppInfo() {
	am.appInfo.WithLabelValues("1.0.0", "2025-01-01", runtime.Version()).Set(1)
}

// RecordMetricsCollectionDuration 记录指标收集耗时
func (am *AppMetrics) RecordMetricsCollectionDuration(duration time.Duration) {
	am.metricsCollectionDuration.Observe(duration.Seconds())
}

// IncrementMetricsCollectionTotal 增加指标收集总数
func (am *AppMetrics) IncrementMetricsCollectionTotal() {
	am.metricsCollectionTotal.Inc()
}

// IncrementMetricsCollectionErrors 增加指标收集错误数
func (am *AppMetrics) IncrementMetricsCollectionErrors() {
	am.metricsCollectionErrors.Inc()
}

// 全局变量和函数
var (
	processStartTime = time.Now().Unix()
	errorCount       = 0
	totalOperations  = 0
)

// getCustomMetricsCount 获取自定义指标数量
func getCustomMetricsCount() int {
	// 这里可以从registry获取实际的指标数量
	return 0 // 暂时返回0，后续可以集成registry
}

// getErrorRate 获取错误率
func getErrorRate() float64 {
	if totalOperations == 0 {
		return 0
	}
	return float64(errorCount) / float64(totalOperations)
}

// RecordError 记录错误
func RecordError() {
	errorCount++
	totalOperations++
}

// RecordOperation 记录操作
func RecordOperation() {
	totalOperations++
}
