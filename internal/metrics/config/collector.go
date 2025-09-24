// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package config

import (
	"time"

	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
)

// CollectorConfig 收集器配置接口
type CollectorConfig struct {
	enabled    bool
	timeout    time.Duration
	retryCount int
	metrics    map[string]MetricConfig
	// labels     []string
}

// IsEnabled 实现 CollectorConfig 接口
func (cc *CollectorConfig) IsEnabled() bool {
	return cc.enabled
}

// GetTimeout 实现 CollectorConfig 接口
func (cc *CollectorConfig) GetTimeout() time.Duration {
	return cc.timeout
}

// GetRetryCount 实现 CollectorConfig 接口
func (cc *CollectorConfig) GetRetryCount() int {
	return cc.retryCount
}

// GetMetrics 实现 CollectorConfig 接口
func (cc *CollectorConfig) GetMetrics() map[string]interfaces.MetricConfig {
	convertedMetrics := make(map[string]interfaces.MetricConfig)
	for key, value := range cc.metrics {
		convertedMetrics[key] = &value
	}
	return convertedMetrics
}

// SetEnabled 设置启用状态
func (cc *CollectorConfig) SetEnabled(enabled bool) {
	cc.enabled = enabled
}

// SetTimeout 设置超时时间
func (cc *CollectorConfig) SetTimeout(timeout time.Duration) {
	cc.timeout = timeout
}

// SetRetryCount 设置重试次数
func (cc *CollectorConfig) SetRetryCount(count int) {
	cc.retryCount = count
}

// AddMetric 添加指标配置
func (cc *CollectorConfig) AddMetric(name string, config MetricConfig) {
	cc.metrics[name] = config
}
