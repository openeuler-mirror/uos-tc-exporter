// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package qdisc

import (
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/base"
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/config"
	"github.com/florianl/go-tc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// func init() {
// 	mc := map[string]config.MetricConfig{
// 		"ce_mark":        *config.NewMetricConfig("ce_mark", "Number of packets marked with CE (Congestion Experienced) by CoDel", "codel"),
// 		"count":          *config.NewMetricConfig("count", "Current number of packets in the CoDel queue", "codel"),
// 		"drop_next":      *config.NewMetricConfig("drop_next", "Time when the next packet will be dropped by CoDel", "codel"),
// 		"drop_overlimit": *config.NewMetricConfig("drop_overlimit", "Number of packets dropped because they exceeded the CoDel limit", "codel"),
// 		"dropping":       *config.NewMetricConfig("dropping", "Indicates whether CoDel is currently dropping packets", "codel"),
// 		"ecn_mark":       *config.NewMetricConfig("ecn_mark", "Number of packets marked with ECN (Explicit Congestion Notification) by CoDel", "codel"),
// 		"ldelay":         *config.NewMetricConfig("ldelay", "Last measured delay of packets in the CoDel queue (in microseconds)", "codel"),
// 		"max_packet":     *config.NewMetricConfig("max_packet", "Maximum packet size handled by CoDel (in bytes)", "codel"),
// 	}
// 	code := NewCodelCollector(config.CollectorConfig{
// 		Enabled:    true,
// 		Timeout:    5,
// 		RetryCount: 3,
// 		Metrics:    mc},
// 		logrus.StandardLogger(),
// 	)
// 	exporter.Register(code)
// }

type CodelCollector struct {
	*base.QdiscBase
}

func NewCodelCollector(cfg config.CollectorConfig, logger *logrus.Logger) *CodelCollector {
	base := base.NewQdiscBase("codel", "qdisc_codel", "Codel qdisc metrics", &cfg, logger)
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

// func (c *CodelCollector) initializeMetrics(cfg *config.CollectorConfig) {
// 	labelNames := c.LabelNames
// 	// CE Mark 指标
// 	c.AddMetric("ce_mark", prometheus.NewDesc(
// 		"qdisc_codel_ce_mark",
// 		"Number of packets marked with CE (Congestion Experienced) by CoDel",
// 		labelNames, nil,
// 	))

//		c.AddSupportedMetric("ce_mark")
//		// Count 指标
//		c.AddMetric("count", prometheus.NewDesc(
//			"qdisc_codel_count",
//			"Current number of packets in the CoDel queue",
//			labelNames, nil,
//		))
//		c.AddSupportedMetric("count")
//		// Drop Next 指标
//		c.AddMetric("drop_next", prometheus.NewDesc(
//			"qdisc_codel_drop_next",
//			"Time when the next packet will be dropped by CoDel",
//			labelNames, nil,
//		))
//		c.AddSupportedMetric("drop_next")
//		// Drop Overlimit 指标
//		c.AddMetric("drop_overlimit", prometheus.NewDesc(
//			"qdisc_codel_drop_overlimit",
//			"Number of packets dropped because they exceeded the CoDel limit",
//			labelNames, nil,
//		))
//		c.AddSupportedMetric("drop_overlimit")
//		// Dropping 指标
//		c.AddMetric("dropping", prometheus.NewDesc(
//			"qdisc_codel_dropping",
//			"Indicates whether CoDel is currently dropping packets",
//			labelNames, nil,
//		))
//		c.AddSupportedMetric("dropping")
//		// Ecn Mark 指标
//		c.AddMetric("ecn_mark", prometheus.NewDesc(
//			"qdisc_codel_ecn_mark",
//			"Number of packets marked with ECN (Explicit Congestion Notification) by CoDel",
//			labelNames, nil,
//		))
//		c.AddSupportedMetric("ecn_mark")
//		// LDelay 指标
//		c.AddMetric("ldelay", prometheus.NewDesc(
//			"qdisc_codel_ldelay",
//			"Last measured delay of packets in the CoDel queue (in microseconds)",
//			labelNames, nil,
//		))
//		c.AddSupportedMetric("ldelay")
//		// Max Packet 指标
//		c.AddMetric("max_packet", prometheus.NewDesc(
//			"qdisc_codel_max_packet",
//			"Maximum packet size handled by CoDel (in bytes)",
//			labelNames, nil,
//		))
//		c.AddSupportedMetric("max_packet")
//	}

func (c *CodelCollector) initializeMetrics(cfg *config.CollectorConfig) {
	labelNames := c.LabelNames
	for metricName, metricConfig := range cfg.GetMetrics() {
		desc := prometheus.NewDesc(
			"qdisc_codel_"+metricName,
			metricConfig.GetHelp(),
			labelNames, nil,
		)
		c.AddMetric(metricName, desc)
		c.AddSupportedMetric(metricName)
	}
}

// ValidateQdisc 验证 qdisc 是否支持
func (c *CodelCollector) ValidateQdisc(qdisc *tc.Object) bool {
	return qdisc.Kind == "codel"
}

// CollectQdiscMetrics 收集 qdisc 指标
func (c *CodelCollector) CollectQdiscMetrics(ch chan<- prometheus.Metric, ns, deviceName string, qdisc any) {
	tcQdisc, ok := qdisc.(*tc.Object)
	if !ok {
		c.Logger.Warnf("Invalid qdisc type for device %s in netns %s", deviceName, ns)
		return
	}
	if tcQdisc.XStats == nil {
		c.Logger.Warnf("No extended stats for codel qdisc on device %s in netns %s", deviceName, ns)
		return
	}
	if tcQdisc.XStats.Codel == nil {
		c.Logger.Debugf("No codel stats for codel qdisc on device %s in netns %s", deviceName, ns)
		return
	}
	attrs := tcQdisc.XStats.Codel

	// 根据配置收集指标
	for _, metricName := range c.GetSupportedMetrics() {
		var value float64
		switch metricName {
		case "ce_mark":
			value = float64(attrs.CeMark)
		case "count":
			value = float64(attrs.Count)
		case "drop_next":
			value = float64(attrs.DropNext)
		case "drop_overlimit":
			value = float64(attrs.DropOverlimit)
		case "dropping":
			value = float64(attrs.Dropping)
		case "ecn_mark":
			value = float64(attrs.EcnMark)
		case "ldelay":
			value = float64(attrs.LDelay)
		case "max_packet":
			value = float64(attrs.MaxPacket)
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
			ns, deviceName, "codel",
		)
	}
}
