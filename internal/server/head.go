// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package server

import (
	_ "embed"
	"net/http"

	"github.com/sirupsen/logrus"
)

//go:embed favicon.ico
var faviconData []byte

type favicon struct{}

func NewFavicon() *favicon {
	return &favicon{}
}

func (f *favicon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	_, err := w.Write(faviconData)
	if err != nil {
		logrus.Errorf("failed to write favicon response: %v", err)
	}
}
