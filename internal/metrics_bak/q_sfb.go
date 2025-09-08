// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics_bak

import (
	"gitee.com/openeuler/uos-tc-exporter/internal/exporter"
	"gitee.com/openeuler/uos-tc-exporter/internal/tc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func init() {
	exporter.Register(
		NewQdiscSfb())
}

type QdiscSfb struct {
	qdiscSfbAvgProbe
	qdiscSfbBucketDrop
	qdiscSfbChildDrop
	qdiscSfbEarlyDrop
	qdiscSfbMarked
	qdiscSfbMaxProb
	qdiscSfbMaxQlen
	qdiscSfbPenaltyDrop
	qdiscSfbQueueDrop
}

func NewQdiscSfb() *QdiscSfb {
	return &QdiscSfb{
		qdiscSfbAvgProbe: *newQdiscSfbAvgProbe(),
	}
}

func (qd *QdiscSfb) Collect(ch chan<- prometheus.Metric) {
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
				if qdisc.Kind != "sfb" {
					continue
				}
				if qdisc.XStats == nil {
					continue
				}
				if qdisc.XStats.Sfb == nil {
					continue
				}
				qd.qdiscSfbAvgProbe.Collect(ch,
					float64(qdisc.XStats.Sfb.AvgProb),
					[]string{ns,
						device.Attributes.Name,
						"sfb"})

				qd.qdiscSfbBucketDrop.Collect(ch,
					float64(qdisc.XStats.Sfb.BucketDrop),
					[]string{ns,
						device.Attributes.Name,
						"sfb"})

				qd.qdiscSfbChildDrop.Collect(ch,
					float64(qdisc.XStats.Sfb.ChildDrop),
					[]string{ns,
						device.Attributes.Name,
						"sfb"})

				qd.qdiscSfbEarlyDrop.Collect(ch,
					float64(qdisc.XStats.Sfb.EarlyDrop),
					[]string{ns,
						device.Attributes.Name,
						"sfb"})

				qd.qdiscSfbMarked.Collect(ch,
					float64(qdisc.XStats.Sfb.Marked),
					[]string{ns,
						device.Attributes.Name,
						"sfb"})

				qd.qdiscSfbMaxProb.Collect(ch,
					float64(qdisc.XStats.Sfb.MaxProb),
					[]string{ns,
						device.Attributes.Name,
						"sfb"})

				qd.qdiscSfbMaxQlen.Collect(ch,
					float64(qdisc.XStats.Sfb.MaxQlen),
					[]string{ns,
						device.Attributes.Name,
						"sfb"})

				qd.qdiscSfbPenaltyDrop.Collect(ch,
					float64(qdisc.XStats.Sfb.PenaltyDrop),
					[]string{ns,
						device.Attributes.Name,
						"sfb"})

				qd.qdiscSfbQueueDrop.Collect(ch,
					float64(qdisc.XStats.Sfb.QueueDrop),
					[]string{ns,
						device.Attributes.Name,
						"sfb"})

			}
		}
	}
}

// ID returns a unique identifier for this metric
func (qd *QdiscSfb) ID() string {
	return "qdisc_sfb"
}

type qdiscSfbAvgProbe struct {
	*baseMetrics
}

func newQdiscSfbAvgProbe() *qdiscSfbAvgProbe {
	logrus.Debug("create qdiscPieAvgDqRate")
	return &qdiscSfbAvgProbe{
		NewMetrics(
			"qdisc_sfb_avg_probe",
			"SFB avg probe xstat",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscSfbAvgProbe) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscSfbBucketDrop struct {
	*baseMetrics
}

func (qd *qdiscSfbBucketDrop) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscSfbChildDrop struct {
	*baseMetrics
}

func (qd *qdiscSfbChildDrop) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscSfbEarlyDrop struct {
	*baseMetrics
}

func (qd *qdiscSfbEarlyDrop) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscSfbMarked struct {
	*baseMetrics
}

func (qd *qdiscSfbMarked) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscSfbMaxProb struct {
	*baseMetrics
}

func (qd *qdiscSfbMaxProb) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscSfbMaxQlen struct {
	*baseMetrics
}

func (qd *qdiscSfbMaxQlen) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscSfbPenaltyDrop struct {
	*baseMetrics
}

func (qd *qdiscSfbPenaltyDrop) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscSfbQueueDrop struct {
	*baseMetrics
}

func (qd *qdiscSfbQueueDrop) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}
