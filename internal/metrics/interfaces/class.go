// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package interfaces

import "github.com/florianl/go-tc"

type ClassCollector interface {
	MetricCollector
	GetClassType() string
	GetSupportedMetrics() []string
	ValidateClass(class *tc.Object) bool
}
