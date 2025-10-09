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
		"choke_early":   *config.NewMetricConfig("choke_early", "Choke early xstat", "choke"),
		"choke_marked":  *config.NewMetricConfig("choke_marked", "Choke marked xstat", "choke"),
		"choke_matched": *config.NewMetricConfig("choke_matched", "Choke matched xstat", "choke"),
		"choke_other":   *config.NewMetricConfig("choke_other", "Choke other xstat", "choke"),
		"choke_pdrop":   *config.NewMetricConfig("choke_pdrop", "Choke pdrop xstat", "choke"),
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

type ChokeCollector struct {
	*base.QdiscBase
}

func NewChokeCollector(cfg config.CollectorConfig, logger *logrus.Logger) *CodelCollector {
	base := base.NewQdiscBase("choke", "qdisc_choke", "Choke qdisc metrics", &cfg, logger)
	collector := &CodelCollector{
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

func (c *ChokeCollector) initializeMetrics(cfg *config.CollectorConfig) {
	labelNames := c.LabelNames
	for metricName, metricConfig := range cfg.GetMetrics() {
		desc := prometheus.NewDesc(
			"qdisc_choke_"+metricName,
			metricConfig.GetHelp(),
			labelNames, nil,
		)
		c.AddMetric(metricName, desc)
		c.AddSupportedMetric(metricName)
	}
}

// ValidateQdisc 验证 qdisc 是否支持
func (c *ChokeCollector) ValidateQdisc(qdisc *tc.Object) bool {
	return true
}

// CollectQdiscMetrics 收集 qdisc 指标
func (c *ChokeCollector) CollectQdiscMetrics(ch chan<- prometheus.Metric, ns, deviceName string, qdisc any) {
	tcQdisc, ok := qdisc.(*tc.Object)
	if !ok {
		c.Logger.Warnf("Invalid qdisc type for device %s in netns %s", deviceName, ns)
		return
	}
	if tcQdisc.XStats == nil {
		c.Logger.Warnf("No extended stats for codel qdisc on device %s in netns %s", deviceName, ns)
		return
	}
	if tcQdisc.XStats.Choke == nil {
		c.Logger.Warnf("No codel stats for codel qdisc on device %s in netns %s", deviceName, ns)
		return
	}
	attrs := tcQdisc.XStats.Choke

	// 根据配置收集指标
	for _, metricName := range c.GetSupportedMetrics() {
		var value float64
		switch metricName {
		case "choke_early":
			value = float64(attrs.Early)
		case "choke_marked":
			value = float64(attrs.Marked)
		case "choke_matched":
			value = float64(attrs.Matched)
		case "choke_other":
			value = float64(attrs.Other)
		case "choke_pdrop":
			value = float64(attrs.PDrop)
		default:
			c.Logger.Warnf("Unsupported metric %s for choke qdisc on device %s in netns %s", metricName, deviceName, ns)
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
			ns, deviceName, "choke",
		)
	}
}
