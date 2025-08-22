// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package exporter

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var defaultReg *Registry

func init() {
	defaultReg = NewRegistry()
}

type Registry struct {
	metrics []Metric
	mu      sync.RWMutex
}

func Register(metric Metric) {
	defaultReg.Register(metric)
}

func RegisterPrometheus(reg *prometheus.Registry) {
	logrus.Info("Registering default registry to prometheus registry")
	reg.MustRegister(defaultReg)
}

func NewRegistry() *Registry {
	return &Registry{
		metrics: []Metric{},
	}
}

func (r *Registry) Register(metrics Metric) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics = append(r.metrics, metrics)
}

func (r *Registry) GetMetrics() []Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.metrics
}

func (r *Registry) Describe(descs chan<- *prometheus.Desc) {
}

func (r *Registry) Collect(ch chan<- prometheus.Metric) {
	for _, m := range r.GetMetrics() {
		func() {
			defer func() {
				if err := recover(); err != nil {
					logrus.Warnf("collector panic recovered: %v", err)
				}
			}()
			m.Collect(ch)
		}()
	}
}
