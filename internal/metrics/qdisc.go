// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics

import (
	"gitee.com/openeuler/uos-tc-exporter/internal/exporter"
	"gitee.com/openeuler/uos-tc-exporter/internal/tc"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	exporter.Register(
		NewQdisc())
}

type Qdisc struct {
	qdiscBytesTotal
	qdiscpacketsTotal
	qdiscDropsTotal
	qdiscOverlimitsTotal
	qdiscBps
	qdiscPps
	qdiscQlenTotal
	qdiscBacklogTotal
	qdiscRequeuesTotal
}

func NewQdisc() *Qdisc {
	return &Qdisc{
		qdiscBytesTotal:      *newQdiscBytesTotal(),
		qdiscpacketsTotal:    *newQdiscpacketsTotal(),
		qdiscDropsTotal:      *NewQdiscDropTotal(),
		qdiscOverlimitsTotal: *NewQdiscOverlimitsTotal(),
		qdiscBps:             *NewQdiscBps(),
		qdiscPps:             *NewQdiscPps(),
		qdiscQlenTotal:       *NewQdiscQlenTotal(),
		qdiscBacklogTotal:    *NewQdiscBacklogTotal(),
		qdiscRequeuesTotal:   *NewQdiscRequeuesTotal(),
	}
}

func (qd *Qdisc) Collect(ch chan<- prometheus.Metric) {
	// 使用优化的日志记录器
	logger := GetMetricsLogger()
	logger.LogCollectionStart("qdisc")

	nsList, err := tc.GetNetNameSpaceList()
	if err != nil {
		logger.LogError("qdisc", "get net namespace list", err)
		return
	}
	if len(nsList) == 0 {
		logger.LogNoData("qdisc", "No net namespace found")
		return
	}

	for _, ns := range nsList {
		devices, err := tc.GetInterfaceInNetNS(ns)
		if err != nil {
			logger.LogError("qdisc", "get interface in netns", err)
			continue
		}
		for _, device := range devices {
			qdiscs, err := tc.GetQdiscs(device.Index, ns)
			if err != nil {
				logger.LogError("qdisc", "get qdiscs in netns", err)
				continue
			}
			for _, qdisc := range qdiscs {
				//handleMaj, handleMin := tc.HandleStr(qdisc.Handle)
				//parentMaj, parentMin := tc.HandleStr(qdisc.Parent)
				var bytes, packets, drops, overlimits, qlen, backlog float64
				if qdisc.Stats2 != nil {
					bytes = float64(qdisc.Stats2.Bytes)
					packets = float64(qdisc.Stats2.Packets)
					drops = float64(qdisc.Stats2.Drops)
					overlimits = float64(qdisc.Stats2.Overlimits)
					qlen = float64(qdisc.Stats2.Qlen)
					backlog = float64(qdisc.Stats2.Backlog)
					qd.qdiscRequeuesTotal.Collect(ch,
						float64(qdisc.Stats2.Requeues),
						[]string{ns,
							device.Attributes.Name})
					logger.IncrementMetricsCount("qdisc", 1)
				}
				if qdisc.Stats != nil {
					bytes = float64(qdisc.Stats.Bytes)
					packets = float64(qdisc.Stats.Packets)
					drops = float64(qdisc.Stats.Drops)
					overlimits = float64(qdisc.Stats.Overlimits)
					qlen = float64(qdisc.Stats.Qlen)
					backlog = float64(qdisc.Stats.Backlog)
				}
				if qdisc.Stats2 == nil && qdisc.Stats == nil {
					continue
				}
				qd.qdiscBytesTotal.Collect(ch,
					bytes,
					[]string{ns,
						device.Attributes.Name})
				qd.qdiscpacketsTotal.Collect(ch,
					packets,
					[]string{ns,
						device.Attributes.Name})
				qd.qdiscDropsTotal.Collect(ch,
					drops,
					[]string{ns,
						device.Attributes.Name})
				qd.qdiscOverlimitsTotal.Collect(ch,
					overlimits,
					[]string{ns,
						device.Attributes.Name})
				qd.qdiscQlenTotal.Collect(ch,
					qlen,
					[]string{ns,
						device.Attributes.Name})
				qd.qdiscBacklogTotal.Collect(ch,
					backlog,
					[]string{ns,
						device.Attributes.Name})
				if qdisc.Stats == nil {
					continue
				}
				qd.qdiscBps.Collect(ch,
					float64(qdisc.Stats.Bps),
					[]string{ns,
						device.Attributes.Name})
				qd.qdiscPps.Collect(ch,
					float64(qdisc.Stats.Pps),
					[]string{ns,
						device.Attributes.Name})
				logger.IncrementMetricsCount("qdisc", 8) // 统计收集的指标数量
			}
		}
	}

	// 记录收集完成
	logger.LogCollectionComplete("qdisc")
}

// ID returns a unique identifier for this metric
func (qd *Qdisc) ID() string {
	return "qdisc"
}

type qdiscBytesTotal struct {
	*baseMetrics
}

func newQdiscBytesTotal() *qdiscBytesTotal {
	return &qdiscBytesTotal{
		NewMetrics(
			"qdisc_bytes_total",
			"QdiscPie byte counter",
			[]string{"namespace",
				"device"})}
}

func (qd *qdiscBytesTotal) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscpacketsTotal struct {
	*baseMetrics
}

func newQdiscpacketsTotal() *qdiscpacketsTotal {
	return &qdiscpacketsTotal{
		NewMetrics(
			"qdisc_packets_total",
			"QdiscPie packet counter",
			[]string{"namespace",
				"device"})}
}

func (qd *qdiscpacketsTotal) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscDropsTotal struct {
	*baseMetrics
}

func NewQdiscDropTotal() *qdiscDropsTotal {
	return &qdiscDropsTotal{
		NewMetrics(
			"qdisc_drops_total",
			"QdiscPie queue drops",
			[]string{"namespace",
				"device"})}
}

func (qd *qdiscDropsTotal) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscOverlimitsTotal struct {
	*baseMetrics
}

func NewQdiscOverlimitsTotal() *qdiscOverlimitsTotal {
	return &qdiscOverlimitsTotal{
		NewMetrics(
			"qdisc_overlimits_total",
			"QdiscPie queue overlimits",
			[]string{"namespace",
				"device"})}
}

func (qd *qdiscOverlimitsTotal) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscBps struct {
	*baseMetrics
}

func NewQdiscBps() *qdiscBps {
	return &qdiscBps{
		NewMetrics(
			"qdisc_bps",
			"QdiscPie byte rate",
			[]string{"namespace",
				"device"})}
}

func (qd *qdiscBps) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscPps struct {
	*baseMetrics
}

func NewQdiscPps() *qdiscPps {
	return &qdiscPps{
		NewMetrics(
			"qdisc_pps",
			"QdiscPie packet rate",
			[]string{"namespace",
				"device"})}
}

func (qd *qdiscPps) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscQlenTotal struct {
	*baseMetrics
}

func NewQdiscQlenTotal() *qdiscQlenTotal {
	return &qdiscQlenTotal{
		NewMetrics(
			"qdisc_qlen_total",
			"QdiscPie queue length",
			[]string{"namespace",
				"device"})}
}

func (qd *qdiscQlenTotal) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscBacklogTotal struct {
	*baseMetrics
}

func NewQdiscBacklogTotal() *qdiscBacklogTotal {
	return &qdiscBacklogTotal{
		NewMetrics(
			"qdisc_backlog_total",
			"QdiscPie backlog",
			[]string{"namespace",
				"device"})}
}
func (qd *qdiscBacklogTotal) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscRequeuesTotal struct {
	*baseMetrics
}

func NewQdiscRequeuesTotal() *qdiscRequeuesTotal {
	return &qdiscRequeuesTotal{
		NewMetrics(
			"qdisc_requeues_total",
			"QdiscPie requeues",
			[]string{"namespace",
				"device"})}
}
func (qd *qdiscRequeuesTotal) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}
