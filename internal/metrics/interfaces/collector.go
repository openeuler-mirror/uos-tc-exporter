// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package interfaces

import "github.com/prometheus/client_golang/prometheus"

type MetricCollector interface {
	// Collect 收集指标数据
	Collect(ch chan<- prometheus.Metric)
	ID() string
	Name() string
	Description() string
	Enabled() bool
	SetEnabled(enabled bool)

	// GetConfig 获取收集器配置
	GetConfig() any

	// SetConfig 设置收集器配置
	SetConfig(config any) error
}
