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
		NewQdiscFqCodel())
}

type QdiscFqCodel struct {
	qdiscFqCodelCeMark
	qdiscFqCodelDropOverlimit
	qdiscFqCodelDropOverMemory
	qdiscFqCodelEcnMark
	qdiscFqCodelMaxPacket
	qdiscFqCodelMemoryUsage
	qdiscFqCodelNewFlowsCount
	qdiscFqCodelNewFlowsLen
}

func NewQdiscFqCodel() *QdiscFqCodel {
	return &QdiscFqCodel{
		qdiscFqCodelCeMark: *newQdiscFqCodelCeMark(),
	}
}

func (qd *QdiscFqCodel) Collect(ch chan<- prometheus.Metric) {
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
				if qdisc.Kind != "fq_codel" {
					continue
				}
				if qdisc.XStats == nil {
					continue
				}
				if qdisc.XStats.FqCodel == nil {
					continue
				}
				qd.qdiscFqCodelCeMark.Collect(ch,
					float64(qdisc.XStats.FqCodel.Qd.CeMark),
					[]string{ns,
						device.Attributes.Name,
						"fq_codel"})
				qd.qdiscFqCodelDropOverlimit.Collect(ch,
					float64(qdisc.XStats.FqCodel.Qd.DropOverlimit),
					[]string{ns,
						device.Attributes.Name,
						"fq_codel"})
				qd.qdiscFqCodelDropOverMemory.Collect(ch,
					float64(qdisc.XStats.FqCodel.Qd.DropOvermemory),
					[]string{ns,
						device.Attributes.Name,
						"fq_codel"})

				qd.qdiscFqCodelEcnMark.Collect(ch,
					float64(qdisc.XStats.FqCodel.Qd.EcnMark),
					[]string{ns,
						device.Attributes.Name,
						"fq_codel"})
				qd.qdiscFqCodelMaxPacket.Collect(ch,
					float64(qdisc.XStats.FqCodel.Qd.MaxPacket),
					[]string{ns,
						device.Attributes.Name,
						"fq_codel"})
				qd.qdiscFqCodelMemoryUsage.Collect(ch,
					float64(qdisc.XStats.FqCodel.Qd.MemoryUsage),
					[]string{ns,
						device.Attributes.Name,
						"fq_codel"})
				qd.qdiscFqCodelNewFlowsCount.Collect(ch,
					float64(qdisc.XStats.FqCodel.Qd.NewFlowCount),
					[]string{ns,
						device.Attributes.Name,
						"fq_codel"})
				qd.qdiscFqCodelNewFlowsLen.Collect(ch,
					float64(qdisc.XStats.FqCodel.Qd.NewFlowsLen),
					[]string{ns,
						device.Attributes.Name,
						"fq_codel"})
			}
		}
	}
}

type qdiscFqCodelCeMark struct {
	*baseMetrics
}

func newQdiscFqCodelCeMark() *qdiscFqCodelCeMark {
	logrus.Debug("create qdiscFqCodelCeMark")
	return &qdiscFqCodelCeMark{
		NewMetrics(
			"qdisc_fq_codel_ce_mark",
			"fq_codel ce mark xstat",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscFqCodelCeMark) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqCodelDropOverlimit struct {
	*baseMetrics
}

func (qd *qdiscFqCodelDropOverlimit) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqCodelDropOverMemory struct {
	*baseMetrics
}

func (qd *qdiscFqCodelDropOverMemory) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqCodelEcnMark struct {
	*baseMetrics
}

func (qd *qdiscFqCodelEcnMark) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqCodelMaxPacket struct {
	*baseMetrics
}

func (qd *qdiscFqCodelMaxPacket) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqCodelMemoryUsage struct {
	*baseMetrics
}

func (qd *qdiscFqCodelMemoryUsage) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqCodelNewFlowsCount struct {
	*baseMetrics
}

func (qd *qdiscFqCodelNewFlowsCount) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqCodelNewFlowsLen struct {
	*baseMetrics
}

func (qd *qdiscFqCodelNewFlowsLen) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}
