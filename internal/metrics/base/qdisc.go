// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package base

import (
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
	"gitee.com/openeuler/uos-tc-exporter/internal/tc"
	"github.com/jsimonetti/rtnetlink"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// QdiscBase qdisc 基础实现
type QdiscBase struct {
	*CollectorBase
	qdiscType        string
	supportedMetrics []string
	labelNames       []string
}

// NewQdiscBase 创建 qdisc 基础实例
func NewQdiscBase(qdiscType, name, description string, config interfaces.CollectorConfig, logger *logrus.Logger) *QdiscBase {
	base := NewCollectorBase("qdisc_"+qdiscType, name, description, config, logger)
	return &QdiscBase{
		CollectorBase:    base,
		qdiscType:        qdiscType,
		supportedMetrics: make([]string, 0),
		labelNames:       []string{"namespace", "device", "kind"},
	}
}

// collectMetrics 实现 qdisc 收集逻辑
func (qb *QdiscBase) collectMetrics(ch chan<- prometheus.Metric) {
	qb.logger.Info("Start collecting qdisc metrics")

	nsList, err := tc.GetNetNameSpaceList()
	if err != nil {
		qb.logger.Warnf("Get net namespace list failed: %v", err)
		qb.setLastError(err)
		return
	}

	if len(nsList) == 0 {
		qb.logger.Info("No net namespace found")
		return
	}

	for _, ns := range nsList {
		qb.collectForNamespace(ch, ns)
	}
}

// collectForNamespace 收集指定命名空间的指标
func (qb *QdiscBase) collectForNamespace(ch chan<- prometheus.Metric, ns string) {
	devices, err := tc.GetInterfaceInNetNS(ns)
	if err != nil {
		qb.logger.Warnf("Get interface in netns %s failed: %v", ns, err)
		return
	}

	for _, device := range devices {
		qb.collectForDevice(ch, ns, device)
	}
}

// collectForDevice 收集指定设备的指标
func (qb *QdiscBase) collectForDevice(ch chan<- prometheus.Metric, ns string, device rtnetlink.LinkMessage) {
	// 获取设备索引
	deviceIndex, deviceName := qb.extractDeviceInfo(device)

	qdiscs, err := tc.GetQdiscs(deviceIndex, ns)
	if err != nil {
		qb.logger.Warnf("Get qdiscs in netns %s failed: %v", ns, err)
		return
	}

	for _, qdisc := range qdiscs {
		if !qb.ValidateQdisc(&qdisc) {
			continue
		}

		qb.collectQdiscMetrics(ch, ns, deviceName, &qdisc)
	}
}

// extractDeviceInfo 提取设备信息
func (qb *QdiscBase) extractDeviceInfo(device rtnetlink.LinkMessage) (uint32, string) {
	// 这里需要根据实际的设备类型进行转换
	// 假设设备有 Index 和 Attributes.Name 字段
	return device.Index, device.Attributes.Name
}

// ValidateQdisc 验证 qdisc 是否支持
func (qb *QdiscBase) ValidateQdisc(qdisc any) bool {
	// 子类需要实现具体的验证逻辑
	return true
}

// collectQdiscMetrics 收集 qdisc 指标
func (qb *QdiscBase) collectQdiscMetrics(ch chan<- prometheus.Metric, ns, deviceName string, qdisc any) {
	// 子类需要实现具体的指标收集逻辑
}

// GetQdiscType 返回 qdisc 类型
func (qb *QdiscBase) GetQdiscType() string {
	return qb.qdiscType
}

// GetSupportedMetrics 返回支持的指标列表
func (qb *QdiscBase) GetSupportedMetrics() []string {
	return qb.supportedMetrics
}

// addSupportedMetric 添加支持的指标
func (qb *QdiscBase) addSupportedMetric(metricName string) {
	qb.supportedMetrics = append(qb.supportedMetrics, metricName)
}
