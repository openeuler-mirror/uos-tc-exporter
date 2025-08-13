// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics

import (
	"gitee.com/openeuler/uos-tc-exporter/internal/exporter"
	"gitee.com/openeuler/uos-tc-exporter/internal/tc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func init() {
	exporter.Register(
		NewQdiscPie())
}

type QdiscPie struct {
	qdiscPieAvgDqRate
	qdiscPieDelay
	qdiscPieDropped
	qdiscPieEcnMark
	qdiscPieMaxq
	qdiscPieOverLimit
	qdiscPiePacketsIn
	qdiscPieProb
}

func NewQdiscPie() *QdiscPie {
	return &QdiscPie{
		qdiscPieAvgDqRate: *newQdiscPieAvgDqRate(),
		qdiscPieDelay:     *newQdiscPieDelay(),
		qdiscPieDropped:   *newQdiscPieDropped(),
		qdiscPieEcnMark:   *newQdiscPieEcnMark(),
		qdiscPieMaxq:      *newQdiscPieMaxq(),
		qdiscPieOverLimit: *newQdiscPieOverLimit(),
		qdiscPiePacketsIn: *newQdiscPiePacketsIn(),
		qdiscPieProb:      *newQdiscPieProb(),
	}
}

func (qd *QdiscPie) Collect(ch chan<- prometheus.Metric) {
	logrus.Info("Start collecting qdisc metrics")
	logrus.Info("get net namespace list")
	nsList, err := tc.GetNetNameSpaceList()
	if err != nil {
		logrus.Warnf("Get net namespace list failed: %v", err)
		return
	}
	if len(nsList) == 0 {
		logrus.Info("No net namespace found")
		return
	}
	for _, ns := range nsList {
		devices, err := tc.GetInterfaceInNetNS(ns)
		if err != nil {
			logrus.Warnf("Get interface in netns %s failed: %v", ns, err)
			continue
		}
		for _, device := range devices {
			qdiscs, err := tc.GetQdiscs(device.Index, ns)
			if err != nil {
				logrus.Warnf("Get qdiscs in netns %s failed: %v", ns, err)
				continue
			}
			for _, qdisc := range qdiscs {
				if qdisc.Kind != "pie" {
					continue
				}
				if qdisc.XStats == nil {
					continue
				}
				if qdisc.XStats.Pie == nil {
					continue
				}
				qd.qdiscPieAvgDqRate.Collect(ch,
					float64(qdisc.XStats.Pie.AvgDqRate),
					[]string{ns,
						device.Attributes.Name,
						"pie"})

				qd.qdiscPieDelay.Collect(ch,
					float64(qdisc.XStats.Pie.Delay),
					[]string{ns,
						device.Attributes.Name,
						"pie"})

				qd.qdiscPieDropped.Collect(ch,
					float64(qdisc.XStats.Pie.Dropped),
					[]string{ns,
						device.Attributes.Name,
						"pie"})

				qd.qdiscPieEcnMark.Collect(ch,
					float64(qdisc.XStats.Pie.EcnMark),
					[]string{ns,
						device.Attributes.Name,
						"pie"})

				qd.qdiscPieMaxq.Collect(ch,
					float64(qdisc.XStats.Pie.Maxq),
					[]string{ns,
						device.Attributes.Name,
						"pie"})

				qd.qdiscPieOverLimit.Collect(ch,
					float64(qdisc.XStats.Pie.Overlimit),
					[]string{ns,
						device.Attributes.Name,
						"pie"})

				qd.qdiscPiePacketsIn.Collect(ch,
					float64(qdisc.XStats.Pie.PacketsIn),
					[]string{ns,
						device.Attributes.Name,
						"pie"})

				qd.qdiscPieProb.Collect(ch,
					float64(qdisc.XStats.Pie.Prob),
					[]string{ns,
						device.Attributes.Name,
						"pie"})
			}
		}
	}
}

type qdiscPieAvgDqRate struct {
	*baseMetrics
}

func newQdiscPieAvgDqRate() *qdiscPieAvgDqRate {
	logrus.Debug("create qdiscPieAvgDqRate")
	return &qdiscPieAvgDqRate{
		NewMetrics(
			"qdisc_pie_avg_dq_rate",
			"PIE avgdqrate xstat",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscPieAvgDqRate) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscPieDelay struct {
	*baseMetrics
}

func newQdiscPieDelay() *qdiscPieDelay {
	logrus.Debug("create qdiscPieDelay")
	return &qdiscPieDelay{
		NewMetrics(
			"qdisc_pie_delay",
			"PIE delay xstat",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscPieDelay) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscPieDropped struct {
	*baseMetrics
}

func newQdiscPieDropped() *qdiscPieDropped {
	logrus.Debug("create qdiscPieDropped")
	return &qdiscPieDropped{
		NewMetrics(
			"qdisc_pie_dropped",
			"PIE dropped xstat",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscPieDropped) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscPieEcnMark struct {
	*baseMetrics
}

func newQdiscPieEcnMark() *qdiscPieEcnMark {
	logrus.Debug("create qdiscPieEcnMark")
	return &qdiscPieEcnMark{
		NewMetrics(
			"qdisc_pie_ecn_mark",
			"PIE ecnmark xstat",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscPieEcnMark) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscPieMaxq struct {
	*baseMetrics
}

func newQdiscPieMaxq() *qdiscPieMaxq {
	logrus.Debug("create qdiscPieMaxq")
	return &qdiscPieMaxq{
		NewMetrics(
			"qdisc_pie_maxq",
			"PIE maxq xstat",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscPieMaxq) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscPieOverLimit struct {
	*baseMetrics
}

func newQdiscPieOverLimit() *qdiscPieOverLimit {
	logrus.Debug("create qdiscPieOverLimit")
	return &qdiscPieOverLimit{
		NewMetrics(
			"qdisc_pie_overlimit",
			"PIE overlimit xstat",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscPieOverLimit) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscPiePacketsIn struct {
	*baseMetrics
}

func newQdiscPiePacketsIn() *qdiscPiePacketsIn {
	logrus.Debug("create qdiscPiePacketsIn")
	return &qdiscPiePacketsIn{
		NewMetrics(
			"qdisc_pie_packets_in",
			"PIE packets_in xstat",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscPiePacketsIn) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscPieProb struct {
	*baseMetrics
}

func newQdiscPieProb() *qdiscPieProb {
	logrus.Debug("create qdiscPieProb")
	return &qdiscPieProb{
		NewMetrics(
			"qdisc_pie_prob",
			"PIE prob xstat",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscPieProb) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}
