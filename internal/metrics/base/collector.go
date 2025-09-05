// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package base

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type CollectorBase struct {
	id          string
	name        string
	description string
	enabled     bool
	logger      *logrus.Logger
	metrics     map[string]*prometheus.Desc
	lastError   error
	lastCollect time.Time
}

func NewCollectorBase(id, name, description string, logger *logrus.Logger) *CollectorBase {
	return &CollectorBase{
		id:          id,
		name:        name,
		description: description,
		logger:      logger,
		metrics:     make(map[string]*prometheus.Desc),
		lastError:   nil,
	}
}
func (cb *CollectorBase) Collect(ch chan<- prometheus.Metric) {
	if !cb.enabled {
		return
	}
}
func (cb *CollectorBase) ID() string {
	return cb.id
}
func (cb *CollectorBase) Name() string {
	return cb.name
}
func (cb *CollectorBase) Description() string {
	return cb.description
}
func (cb *CollectorBase) Enabled() bool {
	return cb.enabled
}
func (cb *CollectorBase) SetEnabled(enabled bool) {
	cb.enabled = enabled
}
func (cb *CollectorBase) GetLastError() error {
	return cb.lastError
}
func (cb *CollectorBase) SetLastError(err error) {
	cb.lastError = err
}
func (cb *CollectorBase) GetLastCollect() time.Time {
	return cb.lastCollect
}
func (cb *CollectorBase) SetLastCollect(lastCollect time.Time) {
	cb.lastCollect = lastCollect
}
