// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package exporter

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var defaultReg = NewRegistry()

// func init() {
// 	defaultReg = NewRegistry()
// }

type Registry struct {
	metrics map[string]Metric // 使用map替代slice，支持快速查找和删除
	mu      sync.RWMutex
}

func Register(metric Metric) {
	defaultReg.Register(metric)
}

// Unregister removes a metric from the registry by its identifier
func Unregister(metricID string) {
	defaultReg.Unregister(metricID)
}

// UnregisterMetric removes a metric by its instance
func UnregisterMetric(metric Metric) {
	defaultReg.UnregisterMetric(metric)
}

func RegisterPrometheus(reg *prometheus.Registry) {
	logrus.Info("Registering default registry to prometheus registry")
	reg.MustRegister(defaultReg)
}

func NewRegistry() *Registry {
	return &Registry{
		metrics: make(map[string]Metric),
	}
}

func (r *Registry) Register(metric Metric) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 使用Metric的ID()方法获取唯一标识符
	metricID := metric.ID()
	r.metrics[metricID] = metric
	logrus.Debugf("Registered metric: %s", metricID)
}

func (r *Registry) Unregister(metricID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.metrics[metricID]; exists {
		delete(r.metrics, metricID)
		logrus.Debugf("Unregistered metric: %s", metricID)
	} else {
		logrus.Warnf("Attempted to unregister non-existent metric: %s", metricID)
	}
}

func (r *Registry) UnregisterMetric(metric Metric) {
	r.mu.Lock()
	defer r.mu.Unlock()

	metricID := generateMetricID(metric)
	if _, exists := r.metrics[metricID]; exists {
		delete(r.metrics, metricID)
		logrus.Debugf("Unregistered metric: %s", metricID)
	} else {
		logrus.Warnf("Attempted to unregister non-existent metric: %s", metricID)
	}
}

func (r *Registry) GetMetrics() []Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := make([]Metric, 0, len(r.metrics))
	for _, metric := range r.metrics {
		metrics = append(metrics, metric)
	}
	return metrics
}

// GetMetricCount returns the current number of registered metrics
func (r *Registry) GetMetricCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.metrics)
}

// Clear removes all metrics from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	oldCount := len(r.metrics)
	r.metrics = make(map[string]Metric)
	logrus.Infof("Cleared registry, removed %d metrics", oldCount)
}

func (r *Registry) Describe(descs chan<- *prometheus.Desc) {
}

func (r *Registry) Collect(ch chan<- prometheus.Metric) {
	for _, m := range r.GetMetrics() {
		func() {
			defer func() {
				if err := recover(); err != nil {
					logrus.Warnf("collector panic recovered: %v", err)
				}
			}()
			m.Collect(ch)
		}()
	}
}

// generateMetricID generates a unique identifier for a metric
// This is a fallback implementation - the primary method should be Metric.ID()
func generateMetricID(metric Metric) string {
	// 优先使用Metric的ID()方法
	if id := metric.ID(); id != "" {
		return id
	}

	// 如果没有ID()方法，使用指针地址作为后备方案
	return fmt.Sprintf("%p", metric)
}
