// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics_bak

import (
	"time"

	"gitee.com/openeuler/uos-tc-exporter/internal/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func init() {
	exporter.Register(NewBusinessMetrics())
}

// BusinessMetrics 业务相关的监控指标
type BusinessMetrics struct {
	// TC相关指标
	tcNamespacesTotal prometheus.GaugeFunc
	tcInterfacesTotal prometheus.GaugeFunc
	tcQdiscsTotal     prometheus.GaugeFunc
	tcClassesTotal    prometheus.GaugeFunc

	// 网络性能指标
	networkLatency    prometheus.Histogram
	networkThroughput prometheus.GaugeFunc
	networkErrors     prometheus.Counter

	// 业务状态指标
	serviceHealth  *prometheus.GaugeVec
	lastUpdateTime prometheus.GaugeFunc

	// 配置相关指标
	configReloadTotal  prometheus.Counter
	configReloadErrors prometheus.Counter
	configVersion      *prometheus.GaugeVec
}

// NewBusinessMetrics 创建新的业务监控指标实例
func NewBusinessMetrics() *BusinessMetrics {
	bm := &BusinessMetrics{
		// TC相关指标
		tcNamespacesTotal: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "tc_namespaces_total",
				Help: "Total number of network namespaces",
			},
			func() float64 { return getTCNamespacesCount() },
		),

		tcInterfacesTotal: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "tc_interfaces_total",
				Help: "Total number of network interfaces",
			},
			func() float64 { return getTCInterfacesCount() },
		),

		tcQdiscsTotal: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "tc_qdiscs_total",
				Help: "Total number of qdiscs",
			},
			func() float64 { return getTCQdiscsCount() },
		),

		tcClassesTotal: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "tc_classes_total",
				Help: "Total number of classes",
			},
			func() float64 { return getTCClassesCount() },
		),

		// 网络性能指标
		networkLatency: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "network_latency_seconds",
				Help:    "Network operation latency",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to 1s
			},
		),

		networkThroughput: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "network_throughput_bytes_per_second",
				Help: "Network throughput in bytes per second",
			},
			func() float64 { return getNetworkThroughput() },
		),

		networkErrors: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "network_errors_total",
				Help: "Total number of network errors",
			},
		),

		// 业务状态指标
		serviceHealth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "service_health",
				Help: "Service health status (1=healthy, 0=unhealthy)",
			},
			[]string{"service_name", "component"},
		),

		lastUpdateTime: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "last_update_timestamp",
				Help: "Timestamp of last successful update",
			},
			func() float64 { return getLastUpdateTimestamp() },
		),

		// 配置相关指标
		configReloadTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "config_reload_total",
				Help: "Total number of configuration reloads",
			},
		),

		configReloadErrors: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "config_reload_errors_total",
				Help: "Total number of configuration reload errors",
			},
		),

		configVersion: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "config_version",
				Help: "Configuration version information",
			},
			[]string{"config_file", "version", "hash"},
		),
	}

	// 初始化默认值
	bm.initializeDefaults()

	return bm
}

// Collect 实现Metric接口
func (bm *BusinessMetrics) Collect(ch chan<- prometheus.Metric) {
	// 收集所有指标
	bm.tcNamespacesTotal.Collect(ch)
	bm.tcInterfacesTotal.Collect(ch)
	bm.tcQdiscsTotal.Collect(ch)
	bm.tcClassesTotal.Collect(ch)
	bm.networkLatency.Collect(ch)
	bm.networkThroughput.Collect(ch)
	bm.networkErrors.Collect(ch)
	bm.serviceHealth.Collect(ch)
	bm.lastUpdateTime.Collect(ch)
	bm.configReloadTotal.Collect(ch)
	bm.configReloadErrors.Collect(ch)
	bm.configVersion.Collect(ch)
}

// ID 实现Metric接口
func (bm *BusinessMetrics) ID() string {
	return "business_metrics"
}

// initializeDefaults 初始化默认值
func (bm *BusinessMetrics) initializeDefaults() {
	// 设置服务健康状态
	bm.serviceHealth.WithLabelValues("tc_exporter", "main").Set(1)
	bm.serviceHealth.WithLabelValues("tc_exporter", "metrics_collector").Set(1)

	// 设置配置版本
	bm.configVersion.WithLabelValues("tc-exporter.yaml", "1.0", "default").Set(1)
}

// RecordNetworkLatency 记录网络延迟
func (bm *BusinessMetrics) RecordNetworkLatency(duration float64) {
	bm.networkLatency.Observe(duration)
}

// IncrementNetworkErrors 增加网络错误计数
func (bm *BusinessMetrics) IncrementNetworkErrors() {
	bm.networkErrors.Inc()
}

// SetServiceHealth 设置服务健康状态
func (bm *BusinessMetrics) SetServiceHealth(serviceName, component string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	bm.serviceHealth.WithLabelValues(serviceName, component).Set(value)
}

// IncrementConfigReload 增加配置重载计数
func (bm *BusinessMetrics) IncrementConfigReload() {
	bm.configReloadTotal.Inc()
}

// IncrementConfigReloadErrors 增加配置重载错误计数
func (bm *BusinessMetrics) IncrementConfigReloadErrors() {
	bm.configReloadErrors.Inc()
}

// UpdateConfigVersion 更新配置版本信息
func (bm *BusinessMetrics) UpdateConfigVersion(configFile, version, hash string) {
	// 重置所有标签组合
	bm.configVersion.Reset()
	// 设置新的版本信息
	bm.configVersion.WithLabelValues(configFile, version, hash).Set(1)
}

// 全局变量和函数
var (
	lastUpdateTimestamp = float64(0)
	tcNamespacesCount   = 0
	tcInterfacesCount   = 0
	tcQdiscsCount       = 0
	tcClassesCount      = 0
	networkThroughput   = float64(0)
)

// getTCNamespacesCount 获取TC命名空间数量
func getTCNamespacesCount() float64 {
	return float64(tcNamespacesCount)
}

// getTCInterfacesCount 获取TC接口数量
func getTCInterfacesCount() float64 {
	return float64(tcInterfacesCount)
}

// getTCQdiscsCount 获取TC qdisc数量
func getTCQdiscsCount() float64 {
	return float64(tcQdiscsCount)
}

// getTCClassesCount 获取TC class数量
func getTCClassesCount() float64 {
	return float64(tcClassesCount)
}

// getNetworkThroughput 获取网络吞吐量
func getNetworkThroughput() float64 {
	return networkThroughput
}

// getLastUpdateTimestamp 获取最后更新时间戳
func getLastUpdateTimestamp() float64 {
	return lastUpdateTimestamp
}

// UpdateTCStats 更新TC统计信息
func UpdateTCStats(namespaces, interfaces, qdiscs, classes int) {
	tcNamespacesCount = namespaces
	tcInterfacesCount = interfaces
	tcQdiscsCount = qdiscs
	tcClassesCount = classes
	lastUpdateTimestamp = float64(time.Now().Unix())
}

// UpdateNetworkThroughput 更新网络吞吐量
func UpdateNetworkThroughput(throughput float64) {
	networkThroughput = throughput
}
