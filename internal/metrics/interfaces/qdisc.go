// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package interfaces

import "github.com/florianl/go-tc"

type QdiscCollector interface {
	MetricCollector
	GetQdiscType() string
	GetSupportedMetrics() []string
	ValidateQdisc(qdisc *tc.Object) bool
}
