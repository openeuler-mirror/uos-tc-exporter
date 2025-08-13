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
		NewQdiscFq())
}

type QdiscFq struct {
	qdiscFqQcFlows
	qdiscFqHighPrioPackets
	qdiscFqTcpRetrans
	qdiscFqThrottled
	qdiscFqThrottledFlows
	qdiscFqFlowsPLimit
	qdiscFqPacketsTooLong
	qdiscFqAllocationErrors
	qdiscFqTimeNextDelayedFlow
	qdiscFqFlows
	qdiscFqInactiveFlows
	qdiscFqUnthrottledLatencyNs
	qdiscFqCeMark
	qdiscFqHorizonDrops
	qdiscFqHorizonCaps
	qdiscFqFastPathPackets
}

func NewQdiscFq() *QdiscFq {
	return &QdiscFq{
		qdiscFqQcFlows:              *newQdiscFqQcFlows(),
		qdiscFqHighPrioPackets:      *newQdiscFqHighPrioPackets(),
		qdiscFqTcpRetrans:           *newQdiscFqTcpRetrans(),
		qdiscFqThrottled:            *newQdiscFqThrottled(),
		qdiscFqThrottledFlows:       *newQdiscFqThrottledFlows(),
		qdiscFqFlowsPLimit:          *newQdiscFqFlowsPLimit(),
		qdiscFqPacketsTooLong:       *newQdiscFqPacketsTooLong(),
		qdiscFqAllocationErrors:     *newQdiscFqAllocationRrrors(),
		qdiscFqTimeNextDelayedFlow:  *newQdiscFqTimeNextDelayedFlow(),
		qdiscFqFlows:                *newQdiscFqFlows(),
		qdiscFqInactiveFlows:        *newQdiscFqInactiveFlows(),
		qdiscFqUnthrottledLatencyNs: *newQdiscFqUnthrottledLatencyNs(),
		qdiscFqCeMark:               *newQdiscFqCeMark(),
		qdiscFqHorizonDrops:         *newQdiscFqHorizonDrops(),
		qdiscFqHorizonCaps:          *newQdiscFqHorizonCaps(),
		qdiscFqFastPathPackets:      *newQdiscFqFastPathPackets(),
	}
}

func (qd *QdiscFq) Collect(ch chan<- prometheus.Metric) {
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
				if qdisc.Kind != "fq" {
					continue
				}
				if qdisc.XStats == nil {
					continue
				}
				if qdisc.XStats.Fq == nil {
					continue
				}
				qd.qdiscFqQcFlows.Collect(ch,
					float64(qdisc.XStats.Fq.GcFlows),
					[]string{ns,
						device.Attributes.Name,
						"fq"})
				qd.qdiscFqHighPrioPackets.Collect(ch,
					float64(qdisc.XStats.Fq.HighPrioPackets),
					[]string{ns,
						device.Attributes.Name,
						"fq"})
				qd.qdiscFqTcpRetrans.Collect(ch,
					float64(qdisc.XStats.Fq.TCPRetrans),
					[]string{ns,
						device.Attributes.Name,
						"fq"})
				qd.qdiscFqThrottled.Collect(ch,
					float64(qdisc.XStats.Fq.Throttled),
					[]string{ns,
						device.Attributes.Name,
						"fq"})
				qd.qdiscFqThrottledFlows.Collect(ch,
					float64(qdisc.XStats.Fq.ThrottledFlows),
					[]string{ns,
						device.Attributes.Name,
						"fq"})
				qd.qdiscFqFlowsPLimit.Collect(ch,
					float64(qdisc.XStats.Fq.FlowsPlimit),
					[]string{ns,
						device.Attributes.Name,
						"fq"})
				qd.qdiscFqPacketsTooLong.Collect(ch,
					float64(qdisc.XStats.Fq.PktsTooLong),
					[]string{ns,
						device.Attributes.Name,
						"fq"})
				qd.qdiscFqAllocationErrors.Collect(ch,
					float64(qdisc.XStats.Fq.AllocationErrors),
					[]string{ns,
						device.Attributes.Name,
						"fq"})
				qd.qdiscFqTimeNextDelayedFlow.Collect(ch,
					float64(qdisc.XStats.Fq.TimeNextDelayedFlow),
					[]string{ns,
						device.Attributes.Name,
						"fq"})
				qd.qdiscFqFlows.Collect(ch,
					float64(qdisc.XStats.Fq.Flows),
					[]string{ns,
						device.Attributes.Name,
						"fq"})
				qd.qdiscFqInactiveFlows.Collect(ch,
					float64(qdisc.XStats.Fq.InactiveFlows),
					[]string{ns,
						device.Attributes.Name,
						"fq"})
				qd.qdiscFqUnthrottledLatencyNs.Collect(ch,
					float64(qdisc.XStats.Fq.UnthrottleLatencyNs),
					[]string{ns,
						device.Attributes.Name,
						"fq"})
				qd.qdiscFqCeMark.Collect(ch,
					float64(qdisc.XStats.Fq.CEMark),
					[]string{ns,
						device.Attributes.Name,
						"fq"})

				qd.qdiscFqHorizonDrops.Collect(ch,
					float64(qdisc.XStats.Fq.HorizonDrops),
					[]string{ns,
						device.Attributes.Name,
						"fq"})
				qd.qdiscFqHorizonCaps.Collect(ch,
					float64(qdisc.XStats.Fq.HorizonCaps),
					[]string{ns,
						device.Attributes.Name,
						"fq"})
				qd.qdiscFqFastPathPackets.Collect(ch,
					float64(qdisc.XStats.Fq.FastpathPackets),
					[]string{ns,
						device.Attributes.Name,
						"fq"})

			}
		}
	}
}

type qdiscFqQcFlows struct {
	*baseMetrics
}

func newQdiscFqQcFlows() *qdiscFqQcFlows {
	logrus.Debug("create qdisc_fq_gc_flows")
	return &qdiscFqQcFlows{
		NewMetrics(
			"qdisc_fq_gc_flows",
			"FQ gc flow counter",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscFqQcFlows) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqHighPrioPackets struct {
	*baseMetrics
}

func newQdiscFqHighPrioPackets() *qdiscFqHighPrioPackets {
	logrus.Debug("create qdisc_fq_high_prio_packets")
	return &qdiscFqHighPrioPackets{
		NewMetrics(
			"qdisc_fq_high_prio_packets",
			"FQ high prio packets counter",
			[]string{"namespace",
				"device",
				"kind"})}
}
func (qd *qdiscFqHighPrioPackets) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqTcpRetrans struct {
	*baseMetrics
}

func newQdiscFqTcpRetrans() *qdiscFqTcpRetrans {
	logrus.Debug("create qdisc_fq_tcp_retrans")
	return &qdiscFqTcpRetrans{
		NewMetrics(
			"qdisc_fq_tcp_retrans",
			"FQ TCP retransmits",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscFqTcpRetrans) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqThrottled struct {
	*baseMetrics
}

func newQdiscFqThrottled() *qdiscFqThrottled {
	logrus.Debug("create qdisc_fq_throttled")
	return &qdiscFqThrottled{
		NewMetrics(
			"qdisc_fq_throttled",
			"FQ throttled counter",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscFqThrottled) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqThrottledFlows struct {
	*baseMetrics
}

func newQdiscFqThrottledFlows() *qdiscFqThrottledFlows {
	logrus.Debug("create qdisc_fq_throttled_flows")
	return &qdiscFqThrottledFlows{
		NewMetrics(
			"qdisc_fq_throttled_flows",
			"FQ throttled flows counter",
			[]string{"namespace",
				"device",
				"kind"})}
}
func (qd *qdiscFqThrottledFlows) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqFlowsPLimit struct {
	*baseMetrics
}

func newQdiscFqFlowsPLimit() *qdiscFqFlowsPLimit {
	logrus.Debug("create qdisc_fq_flows_p_limit")
	return &qdiscFqFlowsPLimit{
		NewMetrics(
			"qdisc_fq_flows_p_limit",
			"FQ flows p limit",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscFqFlowsPLimit) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqPacketsTooLong struct {
	*baseMetrics
}

func newQdiscFqPacketsTooLong() *qdiscFqPacketsTooLong {
	logrus.Debug("create qdisc_fq_packets_too_long")
	return &qdiscFqPacketsTooLong{
		NewMetrics(
			"qdisc_fq_packets_too_long",
			"FQ packets too long",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscFqPacketsTooLong) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqAllocationErrors struct {
	*baseMetrics
}

func newQdiscFqAllocationRrrors() *qdiscFqAllocationErrors {
	logrus.Debug("create qdisc_fq_allocation_rrrors")
	return &qdiscFqAllocationErrors{
		NewMetrics(
			"qdisc_fq_allocation_errors",
			"FQ allocation errors",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscFqAllocationErrors) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqTimeNextDelayedFlow struct {
	*baseMetrics
}

func newQdiscFqTimeNextDelayedFlow() *qdiscFqTimeNextDelayedFlow {
	logrus.Debug("create qdisc_fq_time_next_delayed_flow")
	return &qdiscFqTimeNextDelayedFlow{
		NewMetrics(
			"qdisc_fq_time_next_delayed_flow",
			"FQ time next delayed flow",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscFqTimeNextDelayedFlow) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqFlows struct {
	*baseMetrics
}

func newQdiscFqFlows() *qdiscFqFlows {
	logrus.Debug("create qdisc_fq_flows")
	return &qdiscFqFlows{
		NewMetrics(
			"qdisc_fq_flows",
			"FQ flows counter",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscFqFlows) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqInactiveFlows struct {
	*baseMetrics
}

func newQdiscFqInactiveFlows() *qdiscFqInactiveFlows {
	logrus.Debug("create qdisc_fq_inactive_flows")
	return &qdiscFqInactiveFlows{
		NewMetrics(
			"qdisc_fq_inactive_flows",
			"FQ inactive flows counter",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscFqInactiveFlows) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqUnthrottledLatencyNs struct {
	*baseMetrics
}

func newQdiscFqUnthrottledLatencyNs() *qdiscFqUnthrottledLatencyNs {
	logrus.Debug("create qdisc_fq_unthrottled_latency_ns")
	return &qdiscFqUnthrottledLatencyNs{
		NewMetrics(
			"qdisc_fq_unthrottled_latency_ns",
			"FQ unthrottled latency ns",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscFqUnthrottledLatencyNs) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqCeMark struct {
	*baseMetrics
}

func newQdiscFqCeMark() *qdiscFqCeMark {
	logrus.Debug("create qdisc_fq_ce_mark")
	return &qdiscFqCeMark{
		NewMetrics(
			"qdisc_fq_ce_mark",
			"FQ CE mark",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscFqCeMark) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqHorizonDrops struct {
	*baseMetrics
}

func newQdiscFqHorizonDrops() *qdiscFqHorizonDrops {
	logrus.Debug("create qdisc_fq_horizon_drops")
	return &qdiscFqHorizonDrops{
		NewMetrics(
			"qdisc_fq_horizon_drops",
			"FQ horizon drops",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscFqHorizonDrops) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqHorizonCaps struct {
	*baseMetrics
}

func newQdiscFqHorizonCaps() *qdiscFqHorizonCaps {
	logrus.Debug("create qdisc_fq_horizon_caps")
	return &qdiscFqHorizonCaps{
		NewMetrics(
			"qdisc_fq_horizon_caps",
			"FQ horizon caps",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscFqHorizonCaps) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}

type qdiscFqFastPathPackets struct {
	*baseMetrics
}

func newQdiscFqFastPathPackets() *qdiscFqFastPathPackets {
	logrus.Debug("create qdisc_fq_fast_path_packets")
	return &qdiscFqFastPathPackets{
		NewMetrics(
			"qdisc_fq_fast_path_packets",
			"FQ fast path packets",
			[]string{"namespace",
				"device",
				"kind"})}
}

func (qd *qdiscFqFastPathPackets) Collect(ch chan<- prometheus.Metric,
	value float64,
	labels []string) {
	qd.collect(ch,
		value,
		labels)
}
