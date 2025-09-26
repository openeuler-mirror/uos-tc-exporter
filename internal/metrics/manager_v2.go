// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics

import (
	"sync"
	"time"

	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/config"
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/registry"
	"github.com/sirupsen/logrus"
)

type ManagerV2 struct {
	mu        sync.RWMutex
	registry  *registry.CollectorRegistry
	factories map[string]*registry.CollectorFactory
	config    *config.ManagerConfig
	stats     *CollectionStats
	logger    *logrus.Logger
	// Add fields as necessary
}

type CollectionStats struct {
	mu sync.RWMutex

	TotalCollections      int64
	SuccessfulCollections int64
	FailedCollections     int64
	TotalDuration         time.Duration
	AverageDuration       time.Duration
	LastCollectionTime    time.Time
	LastErrorTime         time.Time
	LastError             error
}
