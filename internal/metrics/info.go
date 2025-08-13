// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics

import (
	"gitee.com/openeuler/uos-tc-exporter/internal/exporter"
	"gitee.com/openeuler/uos-tc-exporter/version"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	exporter.Register(
		NewBuildInfo("exporter_build_info",
			"exporter build info",
			[]string{"version",
				"revision",
				"branch",
				"goversion"}))
}

type BuildInfo struct {
	*baseMetrics
}

func NewBuildInfo(fqname, help string, labels []string) *BuildInfo {
	return &BuildInfo{NewMetrics(fqname, help, labels)}
}

func (c *BuildInfo) Collect(ch chan<- prometheus.Metric) {
	c.baseMetrics.collect(ch,
		1,
		[]string{version.Version,
			version.Revision,
			version.Branch,
			version.GoVersion})
}
