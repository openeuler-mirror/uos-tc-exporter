// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package config

// MetricConfig 指标配置实现
type MetricConfig struct {
	name    string
	enabled bool
	help    string
	mtype   string
	labels  []string
	buckets []float64
}

// NewMetricConfig 创建指标配置
func NewMetricConfig(name, help, mtype string) *MetricConfig {
	return &MetricConfig{
		name:    name,
		enabled: true,
		help:    help,
		mtype:   mtype,
		labels:  []string{"namespace", "device", "kind"},
	}
}

// GetName 实现 MetricConfig 接口
func (mc *MetricConfig) GetName() string {
	return mc.name
}

// IsEnabled 实现 MetricConfig 接口
func (mc *MetricConfig) IsEnabled() bool {
	return mc.enabled
}

// GetHelp 实现 MetricConfig 接口
func (mc *MetricConfig) GetHelp() string {
	return mc.help
}

// GetType 实现 MetricConfig 接口
func (mc *MetricConfig) GetType() string {
	return mc.mtype
}

// GetLabels 实现 MetricConfig 接口
func (mc *MetricConfig) GetLabels() []string {
	return mc.labels
}

// SetEnabled 设置启用状态
func (mc *MetricConfig) SetEnabled(enabled bool) {
	mc.enabled = enabled
}

// SetLabels 设置标签
func (mc *MetricConfig) SetLabels(labels []string) {
	mc.labels = labels
}

// SetBuckets 设置桶配置（用于直方图）
func (mc *MetricConfig) SetBuckets(buckets []float64) {
	mc.buckets = buckets
}
