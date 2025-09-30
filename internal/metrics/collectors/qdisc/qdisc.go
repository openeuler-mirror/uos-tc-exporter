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
		"bytes_total":   *config.NewMetricConfig("bytes_total", "QdiscPie byte counter", "qdisc"),
		"packets_total": *config.NewMetricConfig("packets_total", "QdiscPie packet counter", "qdisc"),
	}
	qc := NewQdiscCollector(config.CollectorConfig{
		Enabled:    true,
		Timeout:    5,
		RetryCount: 3,
		Metrics:    mc},
		logrus.New(),
	)
	exporter.Register(qc)
}

type QdiscCollector struct {
	*base.QdiscBase
}

func NewQdiscCollector(cfg config.CollectorConfig, logger *logrus.Logger) *QdiscCollector {
	base := base.NewQdiscBase("codel", "qdisc_codel", "Codel qdisc metrics", &cfg, logger)
	collector := &QdiscCollector{
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

func (c *QdiscCollector) initializeMetrics(cfg *config.CollectorConfig) {
	labelNames := c.LabelNames
	for metricName, metricConfig := range cfg.GetMetrics() {
		desc := prometheus.NewDesc(
			"qdisc_qdisc_"+metricName,
			metricConfig.GetHelp(),
			labelNames, nil,
		)
		c.AddMetric(metricName, desc)
		c.AddSupportedMetric(metricName)
	}
}

// ValidateQdisc 验证 qdisc 是否支持
func (c *QdiscCollector) ValidateQdisc(qdisc *tc.Object) bool {
	return true
}

// CollectQdiscMetrics 收集 qdisc 指标
func (c *QdiscCollector) CollectQdiscMetrics(ch chan<- prometheus.Metric, ns, deviceName string, qdisc any) {
	tcQdisc, ok := qdisc.(*tc.Object)
	if !ok {
		c.Logger.Warnf("Invalid qdisc type for device %s in netns %s", deviceName, ns)
		return
	}
	if tcQdisc.Stats == nil {
		c.Logger.Warnf("No extended stats for  qdisc on device %s in netns %s", deviceName, ns)
		return
	}

	attrs := tcQdisc.Stats
	// 根据配置收集指标
	for _, metricName := range c.GetSupportedMetrics() {
		var value float64
		switch metricName {
		case "bytes_total":
			value = float64(attrs.Bytes)
		case "packets_total":
			value = float64(attrs.Packets)
		default:
			c.Logger.Warnf("Unsupported metric %s for codel qdisc on device %s in netns %s", metricName, deviceName, ns)
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
			ns, deviceName, "qdisc",
		)
	}
}
