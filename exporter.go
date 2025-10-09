// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package main

import (
	"gitee.com/openeuler/uos-tc-exporter/internal/server"
	"gitee.com/openeuler/uos-tc-exporter/pkg/errors"
	"gitee.com/openeuler/uos-tc-exporter/pkg/logger"
	"github.com/sirupsen/logrus"
)

func Run(name string, version string) error {
	logger.InitDefaultLog()
	s := server.NewServer(name, version)

	s.PrintVersion()
	err := s.SetUp()
	if err != nil {
		customErr := errors.Wrap(err, errors.ErrCodeServerSetup, "server setup failed")
		customErr.WithContext("server_name", name).WithContext("server_version", version)
		logrus.WithFields(logrus.Fields{
			"error_code":     customErr.Code,
			"error":          customErr.Error(),
			"server_name":    name,
			"server_version": version,
		}).Errorln("Server setup failed")
		return customErr
	}
	go func() {
		err := s.Run()
		if err != nil {
			customErr := errors.Wrap(err, errors.ErrCodeServerRun, "server run failed")
			customErr.WithContext("server_name", name).WithContext("server_version", version)
			logrus.WithFields(logrus.Fields{
				"error_code":     customErr.Code,
				"error":          customErr.Error(),
				"server_name":    name,
				"server_version": version,
			}).Error("Server run failed")
			s.Error = customErr
		}

		s.Exit()
	}()
	<-s.ExitSignal
	s.Stop()
	logrus.Info("Exit exporter server completed")
	return s.Error
}
