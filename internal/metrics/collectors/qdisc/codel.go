// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package qdisc

import (
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/base"
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type CodelCollector struct {
	*base.QdiscBase
	metrics map[string]*prometheus.Desc
}

func NewCodelCollector(config interfaces.CollectorConfig, logger *logrus.Logger) *CodelCollector {
	collector := &CodelCollector{
		QdiscBase: base.NewQdiscBase("codel", "CoDel Queue Discipline", "CoDel (Controlled Delay) is an active queue management algorithm designed to combat bufferbloat.", config, logger),
		metrics:   make(map[string]*prometheus.Desc),
	}

	collector.supportedMetrics = []string{
		"backlog_bytes",
		"backlog_packets",
		"maxpacket",
		"drops",
		"overlimits",
		"requeues",
		"newflows",
		"oldflows",
		"ecn_mark",
		"drop_overlimit",
	}

	for _, metric := range collector.supportedMetrics {
		desc := prometheus.NewDesc(
			prometheus.BuildFQName("tc", "qdisc_codel", metric),
			metric+" of CoDel qdisc",
			collector.LabelNames,
			nil,
		)
		collector.metrics[metric] = desc
	}

	return collector
}

func (cc *CodelCollector) Collect(ch chan<- prometheus.Metric) {
	cc.collectMetrics(ch)
}

func (cc *CodelCollector) getQdiscType() string {
	return cc.QdiscType
}

func (cc *CodelCollector) getMetrics() map[string]*prometheus.Desc {
	return cc.metrics
}
