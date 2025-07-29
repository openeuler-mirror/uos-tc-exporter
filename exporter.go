package main

import (
	"gitee.com/openeuler/uos-tc-exporter/internal/server"
	"gitee.com/openeuler/uos-tc-exporter/pkg/logger"
	"github.com/sirupsen/logrus"
)

func Run(name string, version string) error {
	logger.InitDefaultLog()
	s := server.NewServer(name, version)

	s.PrintVersion()
	err := s.SetUp()
	if err != nil {
		logrus.Errorf("SetUp error: %v", err)
		return err
	}
	go func() {
		err := s.Run()
		if err != nil {
			logrus.Errorf("Run error: %v", err)
			s.Error = err
		}

		s.Exit()
	}()
	<-s.ExitSignal
	s.Stop()
	logrus.Info("Exit exporter server completed")
	return s.Error
}
