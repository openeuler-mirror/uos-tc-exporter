// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics_bak

import (
	"time"

	"gitee.com/openeuler/uos-tc-exporter/internal/exporter"
	"github.com/prometheus/client_golang/prometheus"
)

// PerformanceWrapper 性能监控包装器
// 用于包装现有的指标收集器，添加性能监控功能
type PerformanceWrapper struct {
	originalMetric exporter.Metric
	appMetrics     *AppMetrics
}

// NewPerformanceWrapper 创建新的性能监控包装器
func NewPerformanceWrapper(originalMetric exporter.Metric) *PerformanceWrapper {
	return &PerformanceWrapper{
		originalMetric: originalMetric,
		appMetrics:     GetAppMetricsInstance(),
	}
}

// Collect 实现Metric接口，添加性能监控
func (pw *PerformanceWrapper) Collect(ch chan<- prometheus.Metric) {
	startTime := time.Now()

	// 记录操作开始
	RecordOperation()

	// 执行原始指标收集
	func() {
		defer func() {
			if r := recover(); r != nil {
				// 记录panic错误
				RecordError()
				pw.appMetrics.IncrementMetricsCollectionErrors()
			}
		}()

		pw.originalMetric.Collect(ch)
	}()

	// 记录性能指标
	duration := time.Since(startTime)
	pw.appMetrics.RecordMetricsCollectionDuration(duration)
	pw.appMetrics.IncrementMetricsCollectionTotal()
}

// ID 实现Metric接口
func (pw *PerformanceWrapper) ID() string {
	return "perf_wrapper_" + pw.originalMetric.ID()
}

// GetOriginalMetric 获取原始指标收集器
func (pw *PerformanceWrapper) GetOriginalMetric() exporter.Metric {
	return pw.originalMetric
}

// 全局变量
var globalAppMetrics *AppMetrics

// GetAppMetricsInstance 获取全局AppMetrics实例
func GetAppMetricsInstance() *AppMetrics {
	if globalAppMetrics == nil {
		globalAppMetrics = NewAppMetrics()
	}
	return globalAppMetrics
}

// WrapWithPerformanceMonitoring 使用性能监控包装指标收集器
// 这是一个工厂函数，用于创建包装后的指标收集器
func WrapWithPerformanceMonitoring(originalMetric exporter.Metric) exporter.Metric {
	return NewPerformanceWrapper(originalMetric)
}

// WrapAllMetrics 包装所有已注册的指标收集器
// 这个函数可以在应用启动时调用，自动为所有指标添加性能监控
func WrapAllMetrics() {
	// 注意：这个功能需要registry提供相应的接口
	// 目前registry没有提供获取和替换指标的方法
	// 可以通过以下方式实现：
	// 1. 扩展registry接口
	// 2. 在指标注册时直接使用包装器
	// 3. 使用反射（不推荐）
}
