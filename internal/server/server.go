// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"math"
	"sync"
	"time"

	"gitee.com/openeuler/uos-tc-exporter/internal/config"
	logger "gitee.com/openeuler/uos-tc-exporter/pkg/logging"
	"github.com/alecthomas/kingpin"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
)

var (
	defaultSeverVersion  = "1.0.0"
	enableDefaultPromReg *bool
)

func init() {
	enableDefaultPromReg = kingpin.Flag(
		"enable-default-prom-reg",
		"enable default prom reg").
		Bool()
}

type Server struct {
	Name       string
	Version    string
	configMgr  *config.ConfigManager
	metricsMgr *MetricsManager
	httpServer *HttpServer
	ExitSignal chan struct{}
	Error      error
	callback   sync.Once
}

func NewServer(name, version string) *Server {
	if version == "" {
		version = defaultSeverVersion
	}

	s := &Server{
		Name:       name,
		Version:    version,
		ExitSignal: make(chan struct{}),
	}
	return s
}

func (s *Server) SetUp(ctx context.Context) error {
	defer func() {
		if s.Error != nil {
			logrus.Errorf("SetUp error: %v", s.Error)
		}
	}()

	// 解析命令行参数
	err := s.parse()
	if err != nil {
		logrus.Errorf("Parsing command line arguments failed: %v", err)
		return err
	}

	// 初始化配置管理器
	// s.configMgr = NewConfigManager()
	s.configMgr, err = config.NewConfigManager(*config.Configfile)
	if err != nil {
		return err
	}
	err = s.configMgr.LoadConfig()
	if err != nil {
		logrus.Errorf("Loading config file failed: %v", err)
		logrus.Info("Use default config")
		// return err
	}

	// 设置日志
	err = s.setupLog()
	if err != nil {
		logrus.Errorf("SetUp error: %v", err)
		return err
	}

	// 初始化指标管理器
	logrus.Info("setup prom")
	s.metricsMgr = NewMetricsManager()
	s.metricsMgr.Setup()

	// 初始化HTTP服务器
	s.httpServer = NewHttpServer(s.configMgr.GetConfig(), s.configMgr.GetConfig().MetricsPath, s.metricsMgr.GetRegistry())
	err = s.httpServer.Setup(s.metricsMgr)
	if err != nil {
		logrus.Errorf("SetUp error: %v", err)
		return err
	}

	// 启动配置监控
	if err := s.configMgr.StartWatching(ctx); err != nil {
		logrus.Warnf("Failed to start config watching: %v, config hot reload will be disabled", err)
	} else {
		logrus.Info("Config hot reload enabled")
	}

	return nil
}

func (s *Server) setupLog() error {
	config := s.configMgr.GetConfig()
	size, err := humanize.ParseBytes(config.Logging.MaxSize)
	if err != nil {
		logrus.Errorf("Parsing log size failed: %v", err)
		return err
	}
	if size > math.MaxInt64 {
		// 预防整数溢出
		size = math.MaxInt64
	}
	logConfig := logger.NewConfig(config.Logging.Level, config.Logging.LogPath, int64(size), config.Logging.MaxAge)
	logger.Init(logConfig)
	return nil
}

// setupPromReg 已移至 MetricsManager

// 这些方法已移至 HttpServer 结构体

func (s *Server) Run(ctx context.Context) error {
	logrus.Infof("%s successfully setup. SetUp running.", s.Name)
	logrus.Infof("Running %s", s.Name)

	// 使用上下文运行 HTTP 服务器
	return s.httpServer.RunWithContext(ctx)
}

func (s *Server) PrintVersion() {
	logrus.Printf("%s version: %s\n", s.Name, s.Version)
}

func (s *Server) Stop() {
	s.StopWithContext(context.Background())
}

// StopWithContext 使用上下文停止服务器
func (s *Server) StopWithContext(ctx context.Context) {
	logrus.Info("Stopping Server")
	logger.LogOutput("Shutting down server...")

	// 获取配置中的关闭超时时间
	shutdownTimeout := s.configMgr.GetConfig().Server.ShutdownTimeout
	if shutdownTimeout == 0 {
		shutdownTimeout = 30 * time.Second // 默认30秒
	}

	logrus.Infof("Server shutdown timeout set to: %v", shutdownTimeout)

	// 创建带超时的关闭上下文
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// 使用WaitGroup来协调各个组件的关闭
	var wg sync.WaitGroup
	errors := make(chan error, 2) // 最多2个错误（配置监控和HTTP服务器）

	// 停止配置监控
	if s.configMgr != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.configMgr.StopWatching(); err != nil {
				logrus.Warnf("Failed to stop config watching: %v", err)
				errors <- err
			} else {
				logrus.Info("Config watching stopped")
			}
		}()
	}

	// 停止HTTP服务器
	if s.httpServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.httpServer.Stop(); err != nil {
				logrus.Warnf("Failed to stop HTTP server: %v", err)
				errors <- err
			}
		}()
	}

	// 等待所有组件关闭完成或超时
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logrus.Info("All server components stopped successfully")
	case <-shutdownCtx.Done():
		logrus.Warnf("Server shutdown timed out after %v", shutdownTimeout)
		// 强制关闭
		if s.httpServer != nil {
			logrus.Warn("Force closing HTTP server")
			// 这里可以添加强制关闭逻辑
		}
	}

	// 检查是否有错误发生
	close(errors)
	var errorCount int
	for err := range errors {
		if err != nil {
			errorCount++
		}
	}

	if errorCount > 0 {
		logrus.Warnf("Server stopped with %d errors", errorCount)
	} else {
		logrus.Info("Server stopped gracefully")
	}
}

func (s *Server) Exit() {
	s.callback.Do(func() {
		close(s.ExitSignal)
	})
}

func (s *Server) parse() error {
	kingpin.Parse()
	return nil
}

// 这些方法已移至 ConfigManager 结构体
