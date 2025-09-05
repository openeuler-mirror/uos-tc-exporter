// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics_bak

import (
	"github.com/prometheus/client_golang/prometheus"
)

type baseMetrics struct {
	labels []string
	desc   *prometheus.Desc
	fqname string // 添加fqname字段用于ID生成
}

func NewMetrics(fqname, help string, labels []string) *baseMetrics {
	return &baseMetrics{
		labels: labels,
		fqname: fqname,
		desc: prometheus.NewDesc(
			fqname,
			help,
			labels,
			nil),
	}
}

func (c *baseMetrics) collect(ch chan<- prometheus.Metric, value float64, labels []string) {
	ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, value, labels...)
}

// ID returns a unique identifier for this metric
func (c *baseMetrics) ID() string {
	return c.fqname
}
