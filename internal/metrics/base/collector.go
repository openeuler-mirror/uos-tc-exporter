// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package base

import (
	"fmt"
	"sync"
	"time"

	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type CollectorBase struct {
	mu          sync.RWMutex
	id          string
	name        string
	description string
	enabled     bool
	config      interfaces.CollectorConfig
	logger      *logrus.Logger
	metrics     map[string]*prometheus.Desc
	lastError   error
	lastCollect time.Time
}

func NewCollectorBase(id, name, description string, config interfaces.CollectorConfig, logger *logrus.Logger) *CollectorBase {
	return &CollectorBase{
		id:          id,
		name:        name,
		description: description,
		enabled:     true,
		config:      config,
		logger:      logger,
		metrics:     make(map[string]*prometheus.Desc),
	}
}
func (cb *CollectorBase) Collect(ch chan<- prometheus.Metric) {
	cb.mu.RLock()
	if !cb.enabled {
		cb.mu.RUnlock()
		return
	}
	cb.mu.RUnlock()
	defer func() {
		cb.mu.Lock()
		cb.lastCollect = time.Now()
		cb.mu.Unlock()
	}()
	cb.collectMetrics(ch)
}
func (cb *CollectorBase) ID() string {
	return cb.id
}
func (cb *CollectorBase) Name() string {
	return cb.name
}
func (cb *CollectorBase) Description() string {
	return cb.description
}
func (cb *CollectorBase) Enabled() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.enabled
}
func (cb *CollectorBase) SetEnabled(enabled bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.enabled = enabled
}

func (cb *CollectorBase) GetConfig() interfaces.CollectorConfig {
	return cb.config
}
func (cb *CollectorBase) SetConfig(config any) error {
	collectorConfig, ok := config.(interfaces.CollectorConfig)
	if !ok {
		return fmt.Errorf("invalid config type")
	}
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.config = collectorConfig
	return nil
}

// collectMetrics 子类需要实现的收集逻辑
func (cb *CollectorBase) collectMetrics(ch chan<- prometheus.Metric) {
	// 默认实现为空，子类需要重写
}

// addMetric 添加指标描述符
func (cb *CollectorBase) addMetric(name string, desc *prometheus.Desc) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.metrics[name] = desc
}

// getMetric 获取指标描述符
func (cb *CollectorBase) getMetric(name string) (*prometheus.Desc, bool) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	desc, exists := cb.metrics[name]
	return desc, exists
}

// setLastError 设置最后错误
func (cb *CollectorBase) setLastError(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.lastError = err
}

// getLastError 获取最后错误
func (cb *CollectorBase) getLastError() error {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.lastError
}
