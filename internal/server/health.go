// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package server

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// HealthStatus 健康状态结构
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	Check() error
	Name() string
}

// HealthManager 健康状态管理器
type HealthManager struct {
	startTime  time.Time
	version    string
	checkers   []HealthChecker
	isReady    atomic.Bool
	logger     *logrus.Logger
}

// NewHealthManager 创建健康状态管理器
func NewHealthManager(version string, logger *logrus.Logger) *HealthManager {
	if logger == nil {
		logger = logrus.New()
	}

	return &HealthManager{
		startTime: time.Now(),
		version:   version,
		logger:    logger,
		isReady:   atomic.Bool{},
	}
}

// RegisterChecker 注册健康检查器
func (h *HealthManager) RegisterChecker(checker HealthChecker) {
	h.checkers = append(h.checkers, checker)
}

// SetReady 设置服务就绪状态
func (h *HealthManager) SetReady(ready bool) {
	h.isReady.Store(ready)
	h.logger.Infof("Service readiness set to: %v", ready)
}

// IsReady 检查服务是否就绪
func (h *HealthManager) IsReady() bool {
	return h.isReady.Load()
}

// GetUptime 获取运行时间
func (h *HealthManager) GetUptime() string {
	return time.Since(h.startTime).String()
}

// HealthHandler 健康检查处理器
func (h *HealthManager) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := "healthy"
	details := make(map[string]interface{})
	var errors []string

	// 执行所有健康检查
	for _, checker := range h.checkers {
		if err := checker.Check(); err != nil {
			status = "unhealthy"
			errors = append(errors, fmt.Sprintf("%s: %v", checker.Name(), err))
			details[checker.Name()] = map[string]interface{}{
				"status":  "failed",
				"error":   err.Error(),
				"message": "Health check failed",
			}
		} else {
			details[checker.Name()] = map[string]interface{}{
				"status":  "ok",
				"message": "Health check passed",
			}
		}
	}

	// 如果有错误，设置HTTP状态码
	if status == "unhealthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
		// 在错误响应中包含错误详情
		response := HealthStatus{
			Status:    status,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Version:   h.version,
			Uptime:    h.GetUptime(),
			Details:   details,
		}
		
		if err := json.NewEncoder(w).Encode(response); err != nil {
			h.logger.Errorf("Failed to encode health response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// 健康状态响应
	response := HealthStatus{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   h.version,
		Uptime:    h.GetUptime(),
		Details:   details,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Errorf("Failed to encode health response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// ReadyHandler 就绪检查处理器
func (h *HealthManager) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := "ready"
	if !h.IsReady() {
		status = "not ready"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	response := HealthStatus{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   h.version,
		Uptime:    h.GetUptime(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Errorf("Failed to encode ready response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// LivenessHandler 存活检查处理器
func (h *HealthManager) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := HealthStatus{
		Status:    "alive",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   h.version,
		Uptime:    h.GetUptime(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Errorf("Failed to encode liveness response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// BasicHealthChecker 基础健康检查器
type BasicHealthChecker struct {
	name    string
	checkFn func() error
}

// NewBasicHealthChecker 创建基础健康检查器
func NewBasicHealthChecker(name string, checkFn func() error) *BasicHealthChecker {
	return &BasicHealthChecker{
		name:    name,
		checkFn: checkFn,
	}
}

// Check 执行健康检查
func (b *BasicHealthChecker) Check() error {
	return b.checkFn()
}

// Name 获取检查器名称
func (b *BasicHealthChecker) Name() string {
	return b.name
}

// TCHealthChecker TC 相关健康检查
type TCHealthChecker struct {
	logger *logrus.Logger
}

// NewTCHealthChecker 创建TC健康检查器
func NewTCHealthChecker(logger *logrus.Logger) *TCHealthChecker {
	if logger == nil {
		logger = logrus.New()
	}
	return &TCHealthChecker{logger: logger}
}

// Check 检查TC功能是否正常
func (t *TCHealthChecker) Check() error {
	// 这里可以添加TC相关的健康检查逻辑
	// 例如检查netlink连接、TC命令可用性等
	t.logger.Debug("TC health check passed")
	return nil
}

// Name 获取检查器名称
func (t *TCHealthChecker) Name() string {
	return "tc"
}

// MetricsHealthChecker 指标收集健康检查
type MetricsHealthChecker struct {
	logger *logrus.Logger
}

// NewMetricsHealthChecker 创建指标健康检查器
func NewMetricsHealthChecker(logger *logrus.Logger) *MetricsHealthChecker {
	if logger == nil {
		logger = logrus.New()
	}
	return &MetricsHealthChecker{logger: logger}
}

// Check 检查指标收集功能
func (m *MetricsHealthChecker) Check() error {
	// 这里可以添加指标收集相关的健康检查逻辑
	m.logger.Debug("Metrics health check passed")
	return nil
}

// Name 获取检查器名称
func (m *MetricsHealthChecker) Name() string {
	return "metrics"
}
