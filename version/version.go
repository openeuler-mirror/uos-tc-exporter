// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package version

import "runtime"

var (
	Version   = "1.0.0"
	Revision  string
	Branch    string
	GoVersion = runtime.Version()
)
