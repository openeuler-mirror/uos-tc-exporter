// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package qdisc

import (
	"gitee.com/openeuler/uos-tc-exporter/internal/exporter"
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/base"
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/config"
	"github.com/florianl/go-tc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func init() {
	mc := map[string]config.MetricConfig{
		"cbq_avg_idle":    NewCbqConfig("cbq_avg_idle", "CBQ avg idle xstat"),
		"cbq_borrows":     NewCbqConfig("cbq_borrows", "CBQ borrows xstat"),
		"cbq_overactions": NewCbqConfig("cbq_overactions", "CBQ overactions xstat"),
		"cbq_undertime":   NewCbqConfig("cbq_undertime", "CBQ undetime xstat"),
	}
	code := NewChokeCollector(config.CollectorConfig{
		Enabled:    true,
		Timeout:    5,
		RetryCount: 3,
		Metrics:    mc},
		logrus.New(),
	)
	exporter.Register(code)
}

type CbqCollector struct {
	*base.QdiscBase
}

func NewCbqCollector(cfg config.CollectorConfig, logger *logrus.Logger) *CbqCollector {
	base := base.NewQdiscBase("cbq", "qdisc_cbq", "Cbq qdisc metrics", &cfg, logger)
	collector := &CbqCollector{
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

func (c *CbqCollector) initializeMetrics(cfg *config.CollectorConfig) {
	labelNames := c.LabelNames
	for metricName, metricConfig := range cfg.GetMetrics() {
		desc := prometheus.NewDesc(
			"qdisc_cbq_"+metricName,
			metricConfig.GetHelp(),
			labelNames, nil,
		)
		c.AddMetric(metricName, desc)
		c.AddSupportedMetric(metricName)
	}
}

// ValidateQdisc 验证 qdisc 是否支持
func (c *CbqCollector) ValidateQdisc(qdisc *tc.Object) bool {
	return qdisc.Kind == "cbq"
}

// CollectQdiscMetrics 收集 qdisc 指标
func (c *CbqCollector) CollectQdiscMetrics(ch chan<- prometheus.Metric, ns, deviceName string, qdisc any) {
	tcQdisc, ok := qdisc.(*tc.Object)
	if !ok {
		c.Logger.Warnf("Invalid qdisc type for device %s in netns %s", deviceName, ns)
		return
	}
	if tcQdisc.XStats == nil {
		c.Logger.Debugf("No extended stats for cbq qdisc on device %s in netns %s", deviceName, ns)
		return
	}
	if tcQdisc.XStats.Choke == nil {
		c.Logger.Debugf("No cbq stats for cbq qdisc on device %s in netns %s", deviceName, ns)
		return
	}
	attrs := tcQdisc.XStats.Cbq

	// 根据配置收集指标
	for _, metricName := range c.GetSupportedMetrics() {
		var value float64
		switch metricName {
		case "cbq_avg_idle":
			value = float64(attrs.AvgIdle)
		case "cbq_borrows":
			value = float64(attrs.Borrows)
		case "cbq_overactions":
			value = float64(attrs.Overactions)
		case "cbq_undertime":
			value = float64(attrs.Undertime)
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

func NewCbqConfig(name, help string) config.MetricConfig {
	return *config.NewMetricConfig(name, help, "cbq")
}
