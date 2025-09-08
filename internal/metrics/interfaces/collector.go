// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package interfaces

import "github.com/prometheus/client_golang/prometheus"

type MetricCollector interface {
	Collect(ch chan<- prometheus.Metric)
	ID() string
	Name() string
	Description() string
	Enabled() bool
	SetEnabled(enabled bool)
}
