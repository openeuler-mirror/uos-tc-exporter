// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"gitee.com/openeuler/uos-tc-exporter/internal/server"
	"gitee.com/openeuler/uos-tc-exporter/pkg/errors"
	"gitee.com/openeuler/uos-tc-exporter/pkg/logger"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func Run(name string, version string) error {
	logger.InitDefaultLog()

	// 创建根上下文，支持信号取消
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logrus.Infof("Received signal %v, initiating graceful shutdown", sig)
		cancel()
	}()

	s := server.NewServer(name, version)
	s.PrintVersion()

	err := s.SetUp(ctx)
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

	// 使用 errgroup 管理 goroutine 生命周期
	g, gCtx := errgroup.WithContext(ctx)

	// 启动服务器运行
	g.Go(func() error {
		return s.Run(gCtx)
	})

	// 等待上下文取消或服务器错误
	<-gCtx.Done()

	// 优雅关闭
	s.Stop()

	// 等待所有 goroutine 完成
	if err := g.Wait(); err != nil {
		logrus.Errorf("Server exited with error: %v", err)
		return err
	}

	logrus.Info("Exit exporter server completed")
	return nil
}
