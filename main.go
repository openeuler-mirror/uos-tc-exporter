package main

import (
	"time"

	"gitee.com/openeuler/uos-tc-exporter/pkg/logger"
	"gitee.com/openeuler/uos-tc-exporter/version"

	"github.com/sirupsen/logrus"
)

var (
	Name    = "uos_tc_exporter"
	Version = version.Version
)

func init() {
	logConfig := logger.NewConfig("debug", "/var/log/uos_tc_exporter.log", 1024, time.Hour*24)
	logger.Init(logConfig)
}

func main() {

	logrus.Info("uos_tc_exporter start")
}
