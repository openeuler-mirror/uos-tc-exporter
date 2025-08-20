// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"gitee.com/openeuler/uos-tc-exporter/internal/exporter"
	_ "gitee.com/openeuler/uos-tc-exporter/internal/metrics"
	"gitee.com/openeuler/uos-tc-exporter/pkg/errors"
	"gitee.com/openeuler/uos-tc-exporter/pkg/logger"
	"gitee.com/openeuler/uos-tc-exporter/pkg/ratelimit"
	"gitee.com/openeuler/uos-tc-exporter/pkg/utils"
	"github.com/alecthomas/kingpin"
	"github.com/dustin/go-humanize"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
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
	Name         string
	Version      string
	CommonConfig exporter.Config
	configMgr    *exporter.ConfigManager
	promReg      *prometheus.Registry
	handlers     []HandlerFunc
	ExitSignal   chan struct{}
	Error        error
	callback     sync.Once
	server       *http.Server
}

func NewServer(name, version string) *Server {
	if version == "" {
		version = defaultSeverVersion
	}

	// 创建配置管理器
	configMgr, err := exporter.NewConfigManager(*exporter.Configfile)
	if err != nil {
		logrus.Warnf("Failed to create config manager: %v, will use static config", err)
		configMgr = nil
	}

	s := &Server{
		Name:         name,
		Version:      version,
		CommonConfig: exporter.DefaultConfig,
		configMgr:    configMgr,
		promReg:      prometheus.NewRegistry(),
		ExitSignal:   make(chan struct{}),
	}
	return s
}

func (s *Server) SetUp() error {
	defer func() {
		if s.Error != nil {
			logrus.Errorf("SetUp error: %v", s.Error)
		}
	}()
	err := s.parse()
	if err != nil {
		logrus.Errorf("Parsing command line arguments failed: %v", err)
		return err
	}
	err = s.loadConfig()
	if err != nil {
		logrus.Errorf("Loading config file failed: %v", err)
		return err
	}
	err = s.setupLog()
	if err != nil {
		logrus.Errorf("SetUp error: %v", err)
		return err
	}
	logrus.Info("setup prom")
	s.setupPromReg()

	err = s.setupHttpServer()
	if err != nil {
		logrus.Errorf("SetUp error: %v", err)
		return err
	}

	// 启动配置监控（如果配置管理器可用）
	if s.configMgr != nil {
		if err := s.startConfigWatching(); err != nil {
			logrus.Warnf("Failed to start config watching: %v, config hot reload will be disabled", err)
		} else {
			logrus.Info("Config hot reload enabled")
		}
	}

	return nil
}

func (s *Server) setupLog() error {
	size, err := humanize.ParseBytes(s.CommonConfig.Logging.MaxSize)
	if err != nil {
		logrus.Errorf("Parsing log size failed: %v", err)
		return err
	}
	logConfig := logger.NewConfig(s.CommonConfig.Logging.Level, s.CommonConfig.Logging.LogPath, int64(size), s.CommonConfig.Logging.MaxAge)
	logger.Init(logConfig)
	return nil
}

func (s *Server) setupPromReg() {
	if *enableDefaultPromReg {
		s.promReg.MustRegister(
			collectors.NewGoCollector())
		s.promReg.MustRegister(
			collectors.NewProcessCollector(
				collectors.ProcessCollectorOpts(
					prometheus.ProcessCollectorOpts{})))
	}
}

func (s *Server) setupHttpServer() error {
	exporter.RegisterPrometheus(s.promReg)
	mux := http.NewServeMux()
	mux.Handle(s.CommonConfig.MetricsPath, s)
	if *UseRatelimit {
		rateLimiter, err := ratelimit.NewRateLimiter(*rateLimitInterval, *rateLimitSize)
		if err != nil {
			customErr := errors.Wrap(err, errors.ErrCodeRateLimit, "failed to initialize rate limiter middleware")
			customErr.WithContext("interval", *rateLimitInterval).WithContext("size", *rateLimitSize)
			logrus.WithFields(logrus.Fields{
				"error_code": customErr.Code,
				"error":      customErr.Error(),
				"interval":   *rateLimitInterval,
				"size":       *rateLimitSize,
			}).Error("Rate limiter middleware initialization failed")
			return customErr
		}
		s.Use(Ratelimit(rateLimiter))
	}
	addr := fmt.Sprintf("%s:%d", s.CommonConfig.Address, s.CommonConfig.Port)
	schema := "http"
	fmt.Fprintf(os.Stdout, "Listening and serving %s on [%s://%s]\n", s.Name, schema, addr)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	landConfig := LandingPageConfig{
		Name:    s.Name,
		Version: s.Version,
		Links: []LandingPageLinks{
			{
				Text:    "Metrics",
				Address: s.CommonConfig.MetricsPath,
			},
		},
	}
	landPage, err := NewLandingPage(landConfig)
	if err != nil {
		customErr := errors.Wrap(err, errors.ErrCodeLandingPage, "failed to create landing page")
		customErr.WithContext("config", landConfig)
		logrus.WithFields(logrus.Fields{
			"error_code": customErr.Code,
			"error":      customErr.Error(),
			"config":     landConfig,
		}).Error("Landing page creation failed")
		return customErr
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		landPage.ServeHTTP(w, r)
	})
	favicon := NewFavicon()
	mux.Handle("/favicon.ico", favicon)
	s.server = server
	logrus.Infof("Server is running on %s", addr)
	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := s.createRequest(w, r)
	for _, handler := range s.handlers {
		handler(req)
		if req.Error != nil {
			return
		}
	}
	promhttp.HandlerFor(s.promReg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}

func (s *Server) Use(handlerFuncs ...HandlerFunc) {
	s.handlers = append(s.handlers, handlerFuncs...)
}

func (s *Server) createRequest(w http.ResponseWriter, r *http.Request) *Request {
	req := NewRequest(w, r)
	req.handlers = s.handlers
	return req
}

func (s *Server) Run() error {
	go utils.HandleSignals(s.Exit)
	logrus.Infof("%s successfully setup. SetUp running.", s.Name)

	logrus.Infof("Running %s", s.Name)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		customErr := errors.Wrap(err, errors.ErrCodeServerRun, "server listen and serve failed")
		customErr.WithContext("address", s.server.Addr)
		logrus.WithFields(logrus.Fields{
			"error_code": customErr.Code,
			"error":      customErr.Error(),
			"address":    s.server.Addr,
		}).Error("Server listen and serve failed")
		return customErr
	}
	return nil
}

func (s *Server) PrintVersion() {
	logrus.Printf("%s version: %s\n", s.Name, s.Version)
}

func (s *Server) Stop() {
	logrus.Info("Stopping Server")
	logger.LogOutput("Shutting down server...")

	// 停止配置监控
	if s.configMgr != nil {
		if err := s.configMgr.StopWatching(); err != nil {
			logrus.Warnf("Failed to stop config watching: %v", err)
		} else {
			logrus.Info("Config watching stopped")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logrus.Warn("Server shutdown timed out")
		} else {
			customErr := errors.Wrap(err, errors.ErrCodeServerShutdown, "server shutdown failed")
			customErr.WithContext("timeout", "1s")
			logrus.WithFields(logrus.Fields{
				"error_code": customErr.Code,
				"error":      customErr.Error(),
				"timeout":    "1s",
			}).Error("Server shutdown failed")
		}
	} else {
		logrus.Info("Server gracefully stopped")
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

func (s *Server) loadConfig() error {
	if s.configMgr != nil {
		// 使用配置管理器加载配置
		if err := s.configMgr.LoadConfig(); err != nil {
			logrus.Errorf("Failed to load config via config manager: %v", err)
			logrus.Info("Use default config")
			return nil
		}

		// 获取配置并同步到服务器
		s.CommonConfig = s.configMgr.GetConfig()

		// 设置配置重载回调
		s.configMgr.SetReloadCallback(s.onConfigReload)

		logrus.Infof("Loaded config via config manager from: %s", *exporter.Configfile)
		logrus.Info("CommonConfig file loaded and validated successfully")
		return nil
	}

	// 回退到原有的静态配置加载方式
	content, err := os.ReadFile(*exporter.Configfile)
	if err != nil {
		logrus.Errorf("Failed to read config file: %v", err)
		logrus.Info("Use default config")
		return nil
	}
	err = yaml.Unmarshal(content, &s.CommonConfig)
	if err != nil {
		logrus.Errorf("Failed to parse config file: %v", err)
		logrus.Info("Use default config")
		return nil
	}

	// 验证配置的有效性
	if err := s.CommonConfig.Validate(); err != nil {
		customErr := errors.Wrap(err, errors.ErrCodeConfig, "configuration validation failed")
		customErr.WithContext("config_file", *exporter.Configfile)
		logrus.WithFields(logrus.Fields{
			"error_code":  customErr.Code,
			"error":       customErr.Error(),
			"config_file": *exporter.Configfile,
		}).Error("Configuration validation failed")
		return customErr
	}

	logrus.Infof("Loaded config file from: %s", *exporter.Configfile)
	logrus.Info("CommonConfig file loaded and validated successfully")
	return nil
}

// onConfigReload 配置重载回调函数
func (s *Server) onConfigReload(newConfig *exporter.Config) error {
	logrus.Info("Configuration reload triggered")

	// 同步新配置到服务器
	s.CommonConfig = *newConfig

	// 重新设置日志配置
	if err := s.setupLog(); err != nil {
		logrus.Errorf("Failed to update logging config: %v", err)
		return err
	}

	// 重新设置HTTP服务器（如果需要）
	if err := s.setupHttpServer(); err != nil {
		logrus.Errorf("Failed to update HTTP server config: %v", err)
		return err
	}

	logrus.Info("Configuration reload completed successfully")
	return nil
}

// startConfigWatching 启动配置监控
func (s *Server) startConfigWatching() error {
	if s.configMgr == nil {
		return fmt.Errorf("config manager not available")
	}

	// 创建上下文
	ctx := context.Background()

	// 启动配置监控
	if err := s.configMgr.StartWatching(ctx); err != nil {
		return fmt.Errorf("failed to start config watching: %w", err)
	}

	return nil
}
