// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package config

import "time"

type ManagerConfig struct {
	PerformanceMonitoring bool          `yaml:"performance_monitoring"`
	CollectionInterval    time.Duration `yaml:"collection_interval"`
	StatsRetention        time.Duration `yaml:"stats_retention"`
	EnableBusinessMetrics bool          `yaml:"enable_business_metrics"`
}
