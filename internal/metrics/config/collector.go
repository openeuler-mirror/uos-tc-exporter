// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package config

import (
	"time"

	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
)

// CollectorConfig 收集器配置接口
// Updated fields to be exported

type CollectorConfig struct {
	Enabled    bool
	Timeout    time.Duration
	RetryCount int
	Metrics    map[string]MetricConfig
	Labels     []string
}

// NewCollectorConfig 创建收集器配置
func NewCollectorConfig() *CollectorConfig {
	return &CollectorConfig{
		Enabled:    true,
		Timeout:    30 * time.Second,
		RetryCount: 3,
		Metrics:    make(map[string]MetricConfig),
		Labels:     []string{"namespace", "device", "kind"},
	}
}

// IsEnabled 实现 CollectorConfig 接口
func (cc *CollectorConfig) IsEnabled() bool {
	return cc.Enabled
}

// GetTimeout 实现 CollectorConfig 接口
func (cc *CollectorConfig) GetTimeout() time.Duration {
	return cc.Timeout
}

// GetRetryCount 实现 CollectorConfig 接口
func (cc *CollectorConfig) GetRetryCount() int {
	return cc.RetryCount
}

// GetMetrics 实现 CollectorConfig 接口
func (cc *CollectorConfig) GetMetrics() map[string]interfaces.MetricConfig {
	convertedMetrics := make(map[string]interfaces.MetricConfig)
	for key, value := range cc.Metrics {
		convertedMetrics[key] = &value
	}
	return convertedMetrics
}

// SetEnabled 设置启用状态
func (cc *CollectorConfig) SetEnabled(enabled bool) {
	cc.Enabled = enabled
}

// SetTimeout 设置超时时间
func (cc *CollectorConfig) SetTimeout(timeout time.Duration) {
	cc.Timeout = timeout
}

// SetRetryCount 设置重试次数
func (cc *CollectorConfig) SetRetryCount(count int) {
	cc.RetryCount = count
}

// AddMetric 添加指标配置
func (cc *CollectorConfig) AddMetric(name string, config MetricConfig) {
	cc.Metrics[name] = config
}
