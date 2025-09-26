// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package registry

import (
	"fmt"
	"sync"

	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
)

// CollectorFactory 收集器工厂接口
type CollectorFactory interface {
	CreateCollector(collectorType string) (interfaces.MetricCollector, error)
	GetSupportedTypes() []string
}

// CollectorRegistry 收集器注册中心
type CollectorRegistry struct {
	mu sync.RWMutex
	// 使用 map 存储收集器，键为收集器的唯一标识符
	collectors map[string]interfaces.MetricCollector
	factories  map[string]CollectorFactory
}

// NewCollectorRegistry 创建收集器注册中心
func NewCollectorRegistry() *CollectorRegistry {
	return &CollectorRegistry{
		collectors: make(map[string]interfaces.MetricCollector),
		factories:  make(map[string]CollectorFactory),
	}
}

func (cr *CollectorRegistry) Register(collector interfaces.MetricCollector) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	if _, exists := cr.collectors[collector.ID()]; exists {
		return fmt.Errorf("collector with ID %s already registered", collector.ID())
	}
	cr.collectors[collector.ID()] = collector
	return nil
}

func (cr *CollectorRegistry) Unregister(collectorID string) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	if _, exists := cr.collectors[collectorID]; !exists {
		return
	}
	delete(cr.collectors, collectorID)
}

func (cr *CollectorRegistry) GetCollector(collectorID string) (interfaces.MetricCollector, bool) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	collector, exists := cr.collectors[collectorID]
	return collector, exists
}
func (cr *CollectorRegistry) GetAllCollectors() []interfaces.MetricCollector {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	collectors := make([]interfaces.MetricCollector, 0, len(cr.collectors))
	for _, collector := range cr.collectors {
		collectors = append(collectors, collector)
	}
	return collectors
}

func (cr *CollectorRegistry) GetEnableCollectors() []interfaces.MetricCollector {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	collectors := make([]interfaces.MetricCollector, 0, len(cr.collectors))
	for _, collector := range cr.collectors {
		if collector.Enabled() {
			collectors = append(collectors, collector)
		}
	}
	return collectors
}

func (cr *CollectorRegistry) RegisterFactory(factoryName string, factory CollectorFactory) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	if _, exists := cr.factories[factoryName]; exists {
		return fmt.Errorf("factory with name %s already registered", factoryName)
	}
	cr.factories[factoryName] = factory
	return nil
}

func (cr *CollectorRegistry) UnregisterFactory(factoryName string) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	if _, exists := cr.factories[factoryName]; !exists {
		return
	}
	delete(cr.factories, factoryName)
}

func (cr *CollectorRegistry) GetFactory(factoryName string) (CollectorFactory, bool) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	factory, exists := cr.factories[factoryName]
	return factory, exists
}

func (cr *CollectorRegistry) CreateCollector(factoryName, collectorType string) (interfaces.MetricCollector, error) {
	cr.mu.RLock()
	factory, exists := cr.factories[factoryName]
	cr.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("factory with name %s not found", factoryName)
	}
	return factory.CreateCollector(collectorType)
}
