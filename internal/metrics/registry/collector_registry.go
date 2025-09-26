// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package registry

import (
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
