// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package exporter

import "github.com/prometheus/client_golang/prometheus"

type Metric interface {
	Collect(ch chan<- prometheus.Metric)
}
