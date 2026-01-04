// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT
package qdisc

import (
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/config"
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/core/base"
	"github.com/florianl/go-tc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type FqCollector struct {
	*base.QdiscBase
}

func NewFqCollector(cfg config.CollectorConfig, logger *logrus.Logger) *FqCollector {
	base := base.NewQdiscBase("fq", "qdisc_fq", "Fq qdisc metrics", &cfg, logger)
	collector := &FqCollector{
		QdiscBase: base,
	}
	collector.initializeMetrics(&cfg)
	// Wire hooks so that base dispatch calls concrete implementations
	collector.SetQdiscHooks(
		func(qdisc any) bool {
			tcObj, ok := qdisc.(*tc.Object)
			if !ok {
				return false
			}
			return collector.ValidateQdisc(tcObj)
		},
		func(ch chan<- prometheus.Metric, ns, deviceName string, qdisc any) {
			collector.CollectQdiscMetrics(ch, ns, deviceName, qdisc)
		},
	)
	return collector
}

func (c *FqCollector) initializeMetrics(cfg *config.CollectorConfig) {
	labelNames := c.LabelNames
	for metricName, metricConfig := range cfg.GetMetrics() {
		desc := prometheus.NewDesc(
			"qdisc_fq_"+metricName,
			metricConfig.GetHelp(),
			labelNames, nil,
		)
		c.AddMetric(metricName, desc)
		c.AddSupportedMetric(metricName)
	}
}

// ValidateQdisc 验证 qdisc 是否支持
func (c *FqCollector) ValidateQdisc(qdisc *tc.Object) bool {
	return qdisc.Kind == "fq"
}

// CollectQdiscMetrics 收集 qdisc 指标
func (c *FqCollector) CollectQdiscMetrics(ch chan<- prometheus.Metric, ns, deviceName string, qdisc any) {
	tcQdisc, ok := qdisc.(*tc.Object)
	if !ok {
		c.Logger.Warnf("Invalid qdisc type for device %s in netns %s", deviceName, ns)
		return
	}
	if tcQdisc.XStats == nil {
		c.Logger.Debugf("No extended stats for fq qdisc on device %s in netns %s", deviceName, ns)
		return
	}
	if tcQdisc.XStats.Fq == nil {
		c.Logger.Debugf("No fq stats for cbq qdisc on device %s in netns %s", deviceName, ns)
		return
	}
	attrs := tcQdisc.XStats.Fq

	// 根据配置收集指标
	for _, metricName := range c.GetSupportedMetrics() {
		var value float64
		switch metricName {
		case "fq_gc_flows":
			value = float64(attrs.GcFlows)
		case "fq_high_prio_packets":
			value = float64(attrs.HighPrioPackets)
		default:
			c.Logger.Warnf("Unsupported metric %s for cbq qdisc on device %s in netns %s", metricName, deviceName, ns)
			continue
		}
		desc, ok := c.GetMetric(metricName)
		if !ok {
			c.Logger.Warnf("Metric descriptor for %s not found on device %s in netns %s", metricName, deviceName, ns)
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			value,
			ns, deviceName, "cbq",
		)
	}
}

func NewFqConfig(name, help string) config.MetricConfig {
	return *config.NewMetricConfig(name, help, "fq")
}
