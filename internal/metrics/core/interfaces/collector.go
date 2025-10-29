// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package interfaces

import "github.com/prometheus/client_golang/prometheus"

type MetricCollector interface {
	Identifiable
	// Collect 收集指标数据
	Collectible
	// 配置
	Configurable
}

type Identifiable interface {
	ID() string
	Name() string
	Description() string
}

type Configurable interface {
	GetConfig() any
	SetConfig(any) error
}

type Collectible interface {
	Collect(chan<- prometheus.Metric)
	Enabled() bool
	SetEnabled(enabled bool)
}
