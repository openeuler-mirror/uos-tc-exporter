// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type RegistryV2 struct {
	mng *ManagerV2
}

func NewRegistryV2() *RegistryV2 {
	return &RegistryV2{
		mng: NewManagerV2(nil, logrus.StandardLogger()),
	}
}
func (r *RegistryV2) Describe(descs chan<- *prometheus.Desc) {
}

func (r *RegistryV2) Collect(ch chan<- prometheus.Metric) {
	r.mng.CollectAll(ch)
}
