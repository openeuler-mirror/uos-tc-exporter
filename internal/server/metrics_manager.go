// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package server

import (
	"gitee.com/openeuler/uos-tc-exporter/internal/exporter"
	_ "gitee.com/openeuler/uos-tc-exporter/internal/metrics_bak"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/sirupsen/logrus"
)

// MetricsManager 负责指标管理
type MetricsManager struct {
	promReg *prometheus.Registry
}

// NewMetricsManager 创建新的指标管理器
func NewMetricsManager() *MetricsManager {
	return &MetricsManager{
		promReg: prometheus.NewRegistry(),
	}
}

// Setup 设置指标注册表
func (mm *MetricsManager) Setup() {
	// 注册默认收集器
	if *enableDefaultPromReg {
		mm.promReg.MustRegister(
			collectors.NewGoCollector())
		mm.promReg.MustRegister(
			collectors.NewProcessCollector(
				collectors.ProcessCollectorOpts(
					prometheus.ProcessCollectorOpts{})))
	}

	// 注册自定义指标
	exporter.RegisterPrometheus(mm.promReg)
}

// GetRegistry 获取Prometheus注册表
func (mm *MetricsManager) GetRegistry() *prometheus.Registry {
	return mm.promReg
}

// Reload 重新加载指标配置
func (mm *MetricsManager) Reload() error {
	logrus.Info("Reloading metrics configuration")

	// 清理现有注册表
	mm.promReg = prometheus.NewRegistry()

	// 重新设置
	mm.Setup()

	logrus.Info("Metrics configuration reloaded successfully")
	return nil
}
