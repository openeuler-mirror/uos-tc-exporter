// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package collectors

import (
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type TcCollector struct {
	mng *metrics.ManagerV2
}

func NewTcCollector() *TcCollector {
	return &TcCollector{
		mng: metrics.NewManagerV2(nil, logrus.StandardLogger()),
	}
}
func (r *TcCollector) Describe(descs chan<- *prometheus.Desc) {
}

func (r *TcCollector) Collect(ch chan<- prometheus.Metric) {
	r.mng.CollectAll(ch)
}

type CollectorFunc func(ch chan<- prometheus.Metric)

func (f CollectorFunc) Describe(descs chan<- *prometheus.Desc) {
}

func (f CollectorFunc) Collect(ch chan<- prometheus.Metric) {
	f(ch)
}
