// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics

import (
	"sync"
	"time"

	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
	"gitee.com/openeuler/uos-tc-exporter/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ConcurrentCollector 并发收集器
type ConcurrentCollector struct {
	collectors []interfaces.MetricCollector
	poolSize   int
	timeout    time.Duration
	logger     *logrus.Logger
	metrics    *InternalMetrics
}

// NewConcurrentCollector 创建新的并发收集器
func NewConcurrentCollector(poolSize int, timeout time.Duration, logger *logrus.Logger) *ConcurrentCollector {
	if logger == nil {
		logger = logrus.StandardLogger()
	}

	return &ConcurrentCollector{
		collectors: make([]interfaces.MetricCollector, 0),
		poolSize:   poolSize,
		timeout:    timeout,
		logger:     logger,
		metrics:    NewInternalMetrics(logger),
	}
}

// AddCollector 添加收集器
func (cc *ConcurrentCollector) AddCollector(collector interfaces.MetricCollector) {
	cc.collectors = append(cc.collectors, collector)
}

// collectionResult 收集结果
type collectionResult struct {
	collectorName string
	metrics       []prometheus.Metric
	duration      time.Duration
	err           error
}
