// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package utils

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
)

func HandleSignals(function func()) {
	var callback sync.Once
	sigc := make(chan os.Signal, 1)
	defer close(sigc)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM)

	sig := <-sigc
	logrus.Infof("service received signal: %v", sig)
	callback.Do(function)
}
