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
	"gitee.com/openeuler/uos-tc-exporter/pkg/errors"
	"gitee.com/openeuler/uos-tc-exporter/pkg/ratelimit"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// HttpServer 负责HTTP服务器管理
type HttpServer struct {
	server      *http.Server
	handlers    []HandlerFunc
	handlersMu  sync.RWMutex // 保护handlers切片的读写锁
	config      exporter.Config
	metricsPath string
	promReg     *prometheus.Registry
}

// NewHttpServer 创建新的HTTP服务器
func NewHttpServer(config exporter.Config, metricsPath string, promReg *prometheus.Registry) *HttpServer {
	return &HttpServer{
		config:      config,
		metricsPath: metricsPath,
		promReg:     promReg,
	}
}

// Setup 设置HTTP服务器
func (hs *HttpServer) Setup(metricsManager *MetricsManager) error {
	// 设置HTTP多路复用器
	mux := http.NewServeMux()

	// 注册指标端点
	mux.Handle(hs.metricsPath, hs)

	// 设置限流中间件
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
		hs.Use(Ratelimit(rateLimiter))
	}

	// 设置地址
	addr := fmt.Sprintf("%s:%d", hs.config.Address, hs.config.Port)
	schema := "http"
	fmt.Fprintf(os.Stdout, "Listening and serving on [%s://%s]\n", schema, addr)

	// 创建HTTP服务器
	hs.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// 设置着陆页
	if err := hs.setupLandingPage(mux); err != nil {
		return err
	}

	// 设置favicon
	favicon := NewFavicon()
	mux.Handle("/favicon.ico", favicon)

	logrus.Infof("HTTP server is running on %s", addr)
	return nil
}

// setupLandingPage 设置着陆页
func (hs *HttpServer) setupLandingPage(mux *http.ServeMux) error {
	landConfig := LandingPageConfig{
		Name:    "TC Exporter",
		Version: "1.0.0",
		Links: []LandingPageLinks{
			{
				Text:    "Metrics",
				Address: hs.metricsPath,
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

	return nil
}

// ServeHTTP 实现http.Handler接口
func (hs *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := hs.createRequest(w, r)
	hs.handlersMu.RLock()
	defer hs.handlersMu.RUnlock()
	for _, handler := range hs.handlers {
		handler(req)
		if req.Error != nil {
			return
		}
	}

	// 处理指标请求 - 这里需要从外部传入metricsManager
	// 暂时使用默认的prometheus注册表
	promhttp.HandlerFor(hs.promReg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}

// Use 添加中间件处理器
func (hs *HttpServer) Use(handlerFuncs ...HandlerFunc) {
	hs.handlersMu.Lock()
	defer hs.handlersMu.Unlock()
	hs.handlers = append(hs.handlers, handlerFuncs...)
}

// createRequest 创建请求对象
func (hs *HttpServer) createRequest(w http.ResponseWriter, r *http.Request) *Request {
	req := NewRequest(w, r)
	hs.handlersMu.RLock()
	defer hs.handlersMu.RUnlock()
	req.handlers = hs.handlers
	return req
}

// Run 启动HTTP服务器
func (hs *HttpServer) Run() error {
	if hs.server == nil {
		return fmt.Errorf("HTTP server not initialized")
	}

	logrus.Infof("Running HTTP server on %s", hs.server.Addr)
	if err := hs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		customErr := errors.Wrap(err, errors.ErrCodeServerRun, "HTTP server listen and serve failed")
		customErr.WithContext("address", hs.server.Addr)
		logrus.WithFields(logrus.Fields{
			"error_code": customErr.Code,
			"error":      customErr.Error(),
			"address":    hs.server.Addr,
		}).Error("HTTP server listen and serve failed")
		return customErr
	}
	return nil
}

// Stop 停止HTTP服务器
func (hs *HttpServer) Stop() error {
	if hs.server == nil {
		return nil
	}

	logrus.Info("Stopping HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := hs.server.Shutdown(ctx); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logrus.Warn("HTTP server shutdown timed out")
		} else {
			customErr := errors.Wrap(err, errors.ErrCodeServerShutdown, "HTTP server shutdown failed")
			customErr.WithContext("timeout", "1s")
			logrus.WithFields(logrus.Fields{
				"error_code": customErr.Code,
				"error":      customErr.Error(),
				"timeout":    "1s",
			}).Error("HTTP server shutdown failed")
		}
		return err
	}

	logrus.Info("HTTP server gracefully stopped")
	return nil
}

// GetServer 获取HTTP服务器实例
func (hs *HttpServer) GetServer() *http.Server {
	return hs.server
}
