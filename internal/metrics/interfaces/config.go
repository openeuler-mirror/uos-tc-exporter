// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package interfaces

import "time"

// ConfigProvider 定义配置提供者接口
type ConfigProvider interface {
	// GetConfig 获取配置
	GetConfig() interface{}

	// ReloadConfig 重新加载配置
	ReloadConfig() error

	// ValidateConfig 验证配置
	ValidateConfig() error
}

// CollectorConfig 收集器配置接口
type CollectorConfig interface {
	// IsEnabled 检查是否启用
	IsEnabled() bool

	// GetTimeout 获取超时时间
	GetTimeout() time.Duration

	// GetRetryCount 获取重试次数
	GetRetryCount() int

	// GetMetrics 获取指标配置
	GetMetrics() map[string]MetricConfig
}

// MetricConfig 指标配置接口
type MetricConfig interface {
	// GetName 获取指标名称
	GetName() string

	// IsEnabled 检查是否启用
	IsEnabled() bool

	// GetHelp 获取帮助信息
	GetHelp() string

	// GetType 获取指标类型
	GetType() string

	// GetLabels 获取标签列表
	GetLabels() []string
}
