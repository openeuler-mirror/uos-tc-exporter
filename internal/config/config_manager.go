// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// ConfigManager 配置管理器，负责配置的加载、验证和热重载
type ConfigManager struct {
	configPath  string
	config      *Config
	watcher     *fsnotify.Watcher
	reloadChan  chan struct{}
	stopChan    chan struct{}
	mu          sync.RWMutex
	reloadDelay time.Duration
	onReload    func(*Config) error
	lastReload  time.Time
	reloadCount int64

	// 影子构建和原子切换相关字段
	shadowConfig    *Config          // 影子配置
	configPtr       *atomic.Value    // 原子指针，用于原子切换
	rollbackConfig  *Config          // 回滚配置
	validationRules *ValidationRules // 验证规则
}

// ValidationRules 配置验证规则
type ValidationRules struct {
	RequiredFields []string
	RangeRules     map[string]RangeRule
	ConflictRules  []ConflictRule
	CustomRules    []CustomValidationRule
}

// RangeRule 范围验证规则
type RangeRule struct {
	Min interface{}
	Max interface{}
}

// ConflictRule 冲突验证规则
type ConflictRule struct {
	Fields []string
	Check  func(map[string]interface{}) error
}

// CustomValidationRule 自定义验证规则
type CustomValidationRule struct {
	Field string
	Check func(interface{}) error
}

// ConfigValidationError 配置验证错误
type ConfigValidationError struct {
	Field   string
	Message string
	Type    string // "required", "range", "conflict", "custom"
}

func (e *ConfigValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s (%s)", e.Field, e.Message, e.Type)
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager(configPath string) (*ConfigManager, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	// 初始化原子指针
	configPtr := &atomic.Value{}
	configPtr.Store(&DefaultConfig)

	cm := &ConfigManager{
		configPath:      configPath,
		config:          &DefaultConfig,
		watcher:         watcher,
		reloadChan:      make(chan struct{}, 1),
		stopChan:        make(chan struct{}),
		reloadDelay:     2 * time.Second, // 防抖延迟
		lastReload:      time.Now(),
		configPtr:       configPtr,
		validationRules: createDefaultValidationRules(),
	}

	// 设置默认配置
	*cm.config = DefaultConfig
	cm.configPtr.Store(cm.config)

	return cm, nil
}

// createDefaultValidationRules 创建默认验证规则
func createDefaultValidationRules() *ValidationRules {
	return &ValidationRules{
		RequiredFields: []string{"address", "port", "metricsPath"},
		RangeRules: map[string]RangeRule{
			"port":                   {Min: 1, Max: 65535},
			"server.shutdownTimeout": {Min: time.Second, Max: 5 * time.Minute},
		},
		ConflictRules: []ConflictRule{
			{
				Fields: []string{"address", "port"},
				Check: func(fields map[string]interface{}) error {
					address, ok1 := fields["address"].(string)
					port, ok2 := fields["port"].(int)
					if !ok1 || !ok2 {
						return nil
					}

					// 检查是否为特权端口但地址不是localhost
					if port < 1024 && address != "127.0.0.1" && address != "localhost" {
						return fmt.Errorf("privileged port %d requires localhost address for security", port)
					}
					return nil
				},
			},
		},
		CustomRules: []CustomValidationRule{
			{
				Field: "address",
				Check: func(value interface{}) error {
					addr, ok := value.(string)
					if !ok {
						return fmt.Errorf("address must be a string")
					}
					if addr == "" {
						return fmt.Errorf("address cannot be empty")
					}
					return nil
				},
			},
			{
				Field: "metricsPath",
				Check: func(value interface{}) error {
					path, ok := value.(string)
					if !ok {
						return fmt.Errorf("metricsPath must be a string")
					}
					if path == "" {
						return fmt.Errorf("metricsPath cannot be empty")
					}
					if path[0] != '/' {
						return fmt.Errorf("metricsPath must start with '/'")
					}
					return nil
				},
			},
		},
	}
}

// loadConfigFromFile 从文件加载配置（内部辅助方法）
func (cm *ConfigManager) loadConfigFromFile() (*Config, error) {
	// 检查文件是否存在
	if !cm.fileExists() {
		return nil, errors.New("config file not found")
	}

	// 读取配置文件
	content, err := os.ReadFile(cm.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析YAML
	var newConfig Config
	if err := yaml.Unmarshal(content, &newConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 使用增强的验证
	if err := cm.validateConfig(&newConfig); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &newConfig, nil
}

// validateConfig 使用结构化验证规则验证配置
func (cm *ConfigManager) validateConfig(config *Config) error {
	var validationErrors []*ConfigValidationError

	// 1. 必填字段验证
	if err := cm.validateRequiredFields(config); err != nil {
		validationErrors = append(validationErrors, err...)
	}

	// 2. 范围验证
	if err := cm.validateRanges(config); err != nil {
		validationErrors = append(validationErrors, err...)
	}

	// 3. 冲突验证
	if err := cm.validateConflicts(config); err != nil {
		validationErrors = append(validationErrors, err...)
	}

	// 4. 自定义验证
	if err := cm.validateCustomRules(config); err != nil {
		validationErrors = append(validationErrors, err...)
	}

	// 5. 调用原有的验证方法
	if err := config.Validate(); err != nil {
		validationErrors = append(validationErrors, &ConfigValidationError{
			Field:   "config",
			Message: err.Error(),
			Type:    "custom",
		})
	}

	// 返回所有验证错误
	if len(validationErrors) > 0 {
		var errorMessages []string
		for _, err := range validationErrors {
			errorMessages = append(errorMessages, err.Error())
		}
		return fmt.Errorf("configuration validation failed:\n%s", fmt.Sprintf("%v", errorMessages))
	}

	return nil
}

// validateRequiredFields 验证必填字段
func (cm *ConfigManager) validateRequiredFields(config *Config) []*ConfigValidationError {
	var errors []*ConfigValidationError

	for _, field := range cm.validationRules.RequiredFields {
		switch field {
		case "address":
			if config.Address == "" {
				errors = append(errors, &ConfigValidationError{
					Field:   field,
					Message: "address is required",
					Type:    "required",
				})
			}
		case "port":
			if config.Port == 0 {
				errors = append(errors, &ConfigValidationError{
					Field:   field,
					Message: "port is required",
					Type:    "required",
				})
			}
		case "metricsPath":
			if config.MetricsPath == "" {
				errors = append(errors, &ConfigValidationError{
					Field:   field,
					Message: "metricsPath is required",
					Type:    "required",
				})
			}
		}
	}

	return errors
}

// validateRanges 验证范围
func (cm *ConfigManager) validateRanges(config *Config) []*ConfigValidationError {
	var errors []*ConfigValidationError

	for field, rule := range cm.validationRules.RangeRules {
		switch field {
		case "port":
			if port, ok := rule.Min.(int); ok && config.Port < port {
				errors = append(errors, &ConfigValidationError{
					Field:   field,
					Message: fmt.Sprintf("port must be >= %d, got %d", port, config.Port),
					Type:    "range",
				})
			}
			if port, ok := rule.Max.(int); ok && config.Port > port {
				errors = append(errors, &ConfigValidationError{
					Field:   field,
					Message: fmt.Sprintf("port must be <= %d, got %d", port, config.Port),
					Type:    "range",
				})
			}
		case "server.shutdownTimeout":
			if min, ok := rule.Min.(time.Duration); ok && config.Server.ShutdownTimeout < min {
				errors = append(errors, &ConfigValidationError{
					Field:   field,
					Message: fmt.Sprintf("shutdownTimeout must be >= %v, got %v", min, config.Server.ShutdownTimeout),
					Type:    "range",
				})
			}
			if max, ok := rule.Max.(time.Duration); ok && config.Server.ShutdownTimeout > max {
				errors = append(errors, &ConfigValidationError{
					Field:   field,
					Message: fmt.Sprintf("shutdownTimeout must be <= %v, got %v", max, config.Server.ShutdownTimeout),
					Type:    "range",
				})
			}
		}
	}

	return errors
}

// validateConflicts 验证冲突
func (cm *ConfigManager) validateConflicts(config *Config) []*ConfigValidationError {
	var errors []*ConfigValidationError

	for _, rule := range cm.validationRules.ConflictRules {
		fields := make(map[string]interface{})

		// 提取相关字段的值
		for _, field := range rule.Fields {
			switch field {
			case "address":
				fields[field] = config.Address
			case "port":
				fields[field] = config.Port
			case "metricsPath":
				fields[field] = config.MetricsPath
			}
		}

		// 执行冲突检查
		if err := rule.Check(fields); err != nil {
			errors = append(errors, &ConfigValidationError{
				Field:   fmt.Sprintf("conflict(%s)", fmt.Sprintf("%v", rule.Fields)),
				Message: err.Error(),
				Type:    "conflict",
			})
		}
	}

	return errors
}

// validateCustomRules 验证自定义规则
func (cm *ConfigManager) validateCustomRules(config *Config) []*ConfigValidationError {
	var errors []*ConfigValidationError

	for _, rule := range cm.validationRules.CustomRules {
		var value interface{}

		switch rule.Field {
		case "address":
			value = config.Address
		case "port":
			value = config.Port
		case "metricsPath":
			value = config.MetricsPath
		default:
			continue
		}

		if err := rule.Check(value); err != nil {
			errors = append(errors, &ConfigValidationError{
				Field:   rule.Field,
				Message: err.Error(),
				Type:    "custom",
			})
		}
	}

	return errors
}

// LoadConfig 加载配置文件
func (cm *ConfigManager) LoadConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	newConfig, err := cm.loadConfigFromFile()
	if err != nil {
		logrus.Warnf("Config file %s not found", cm.configPath)
		return err
	}

	// 应用配置
	cm.config = newConfig
	logrus.Debugf("Config loaded: address=%s, port=%d, metricsPath=%s", cm.config.Address, cm.config.Port, cm.config.MetricsPath)
	return nil
}

// Reload 重新加载配置文件（使用影子构建+原子切换）
func (cm *ConfigManager) Reload() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 1. 影子构建：先构建新配置但不应用
	shadowConfig, err := cm.buildShadowConfig()
	if err != nil {
		logrus.Errorf("Shadow config build failed: %v", err)
		return fmt.Errorf("shadow config build failed: %w", err)
	}

	// 2. 保存当前配置用于回滚
	rollbackConfig := cm.config
	cm.rollbackConfig = rollbackConfig

	// 3. 原子切换：使用原子操作切换配置
	cm.configPtr.Store(shadowConfig)
	cm.config = shadowConfig
	cm.shadowConfig = nil // 清空影子配置

	// 4. 更新统计信息
	cm.lastReload = time.Now()
	atomic.AddInt64(&cm.reloadCount, 1)

	logrus.Infof("Config reloaded successfully from %s (reload count: %d)", cm.configPath, atomic.LoadInt64(&cm.reloadCount))
	logrus.Debugf("Config reloaded: address=%s, port=%d, metricsPath=%s", cm.config.Address, cm.config.Port, cm.config.MetricsPath)

	// 5. 调用重载回调（如果设置）
	if cm.onReload != nil {
		if err := cm.onReload(cm.config); err != nil {
			logrus.Errorf("Config reload callback failed: %v", err)
			// 回滚到旧配置
			if err := cm.rollbackToConfig(rollbackConfig); err != nil {
				logrus.Errorf("Failed to rollback config: %v", err)
			}
			return err
		}
	}

	// 6. 清空回滚配置（成功应用后不再需要）
	cm.rollbackConfig = nil

	return nil
}

// buildShadowConfig 构建影子配置
func (cm *ConfigManager) buildShadowConfig() (*Config, error) {
	// 从文件加载新配置
	newConfig, err := cm.loadConfigFromFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load config from file: %w", err)
	}

	// 使用增强的验证
	if err := cm.validateConfig(newConfig); err != nil {
		return nil, fmt.Errorf("shadow config validation failed: %w", err)
	}

	// 检查配置冲突（与当前配置比较）
	if err := cm.checkConfigConflicts(newConfig); err != nil {
		return nil, fmt.Errorf("config conflicts detected: %w", err)
	}

	// 存储为影子配置
	cm.shadowConfig = newConfig

	logrus.Debugf("Shadow config built successfully: address=%s, port=%d, metricsPath=%s",
		newConfig.Address, newConfig.Port, newConfig.MetricsPath)

	return newConfig, nil
}

// checkConfigConflicts 检查配置冲突
func (cm *ConfigManager) checkConfigConflicts(newConfig *Config) error {
	currentConfig := cm.configPtr.Load().(*Config)

	// 检查端口冲突
	if newConfig.Port == currentConfig.Port && newConfig.Address == currentConfig.Address {
		// 如果地址和端口都相同，需要检查是否有其他服务占用
		logrus.Warnf("New config uses same address:port as current config: %s:%d",
			newConfig.Address, newConfig.Port)
	}

	// 检查指标路径冲突
	if newConfig.MetricsPath == currentConfig.MetricsPath {
		logrus.Debugf("Metrics path unchanged: %s", newConfig.MetricsPath)
	}

	// 可以添加更多冲突检查逻辑
	return nil
}

// rollbackToConfig 回滚到指定配置
func (cm *ConfigManager) rollbackToConfig(rollbackConfig *Config) error {
	if rollbackConfig == nil {
		return fmt.Errorf("rollback config is nil")
	}

	// 原子切换回旧配置
	cm.configPtr.Store(rollbackConfig)
	cm.config = rollbackConfig

	logrus.Warnf("Config rolled back to previous version: address=%s, port=%d, metricsPath=%s",
		rollbackConfig.Address, rollbackConfig.Port, rollbackConfig.MetricsPath)

	return nil
}

// RollbackToPrevious 回滚到上一个配置
func (cm *ConfigManager) RollbackToPrevious() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.rollbackConfig == nil {
		return fmt.Errorf("no rollback config available")
	}

	return cm.rollbackToConfig(cm.rollbackConfig)
}

// GetConfig 获取当前配置的副本（使用原子操作）
func (cm *ConfigManager) GetConfig() Config {
	config := cm.configPtr.Load().(*Config)
	if config == nil {
		return DefaultConfig
	}
	return *config
}

// GetConfigPtr 获取当前配置的指针（用于内部使用）
func (cm *ConfigManager) GetConfigPtr() *Config {
	return cm.configPtr.Load().(*Config)
}

// GetShadowConfig 获取影子配置（用于调试）
func (cm *ConfigManager) GetShadowConfig() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.shadowConfig
}

// GetRollbackConfig 获取回滚配置（用于调试）
func (cm *ConfigManager) GetRollbackConfig() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.rollbackConfig
}

// ValidateConfigFile 验证配置文件（不加载）
func (cm *ConfigManager) ValidateConfigFile() error {
	// 检查文件是否存在
	if !cm.fileExists() {
		return errors.New("config file not found")
	}

	// 读取配置文件
	content, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析YAML
	var testConfig Config
	if err := yaml.Unmarshal(content, &testConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// 使用增强的验证
	return cm.validateConfig(&testConfig)
}

// SetValidationRules 设置自定义验证规则
func (cm *ConfigManager) SetValidationRules(rules *ValidationRules) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.validationRules = rules
}

// GetValidationRules 获取当前验证规则
func (cm *ConfigManager) GetValidationRules() *ValidationRules {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.validationRules
}

// SetReloadCallback 设置配置重载回调函数
func (cm *ConfigManager) SetReloadCallback(callback func(*Config) error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.onReload = callback
}

// StartWatching 开始监控配置文件变化
func (cm *ConfigManager) StartWatching(ctx context.Context) error {
	// 确保配置目录存在
	configDir := filepath.Dir(cm.configPath)
	if err := cm.watcher.Add(configDir); err != nil {
		return fmt.Errorf("failed to watch config directory %s: %w", configDir, err)
	}

	// 启动监控goroutine
	go cm.watchLoop(ctx)

	logrus.Infof("Started watching config file: %s", cm.configPath)
	return nil
}

// StopWatching 停止监控配置文件变化
func (cm *ConfigManager) StopWatching() error {
	close(cm.stopChan)
	if cm.watcher != nil {
		return cm.watcher.Close()
	}
	return nil
}

// watchLoop 文件监控循环
func (cm *ConfigManager) watchLoop(ctx context.Context) {
	var reloadTimer *time.Timer
	configPathClean := filepath.Clean(cm.configPath)

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopChan:
			return
		case event, ok := <-cm.watcher.Events:
			if !ok {
				return
			}

			// 只处理配置文件的变化
			if filepath.Clean(event.Name) == configPathClean {
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					// 防抖处理：延迟重载
					if reloadTimer != nil {
						reloadTimer.Stop()
					}
					reloadTimer = time.AfterFunc(cm.reloadDelay, func() {
						select {
						case cm.reloadChan <- struct{}{}:
							// 成功发送重载信号
						default:
							// 通道已满，跳过本次重载
						}
					})
				}
			}
		case err, ok := <-cm.watcher.Errors:
			if !ok {
				return
			}
			logrus.Errorf("Config file watcher error: %v", err)
		case <-cm.reloadChan:
			// 执行配置重载
			if err := cm.Reload(); err != nil {
				logrus.Errorf("Failed to reload config: %v", err)
			}
		}
	}
}

// fileExists 检查配置文件是否存在
func (cm *ConfigManager) fileExists() bool {
	_, err := os.Stat(cm.configPath)
	return err == nil
}

// GetStats 获取配置管理器统计信息
func (cm *ConfigManager) GetStats() map[string]any {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return map[string]any{
		"config_path":      cm.configPath,
		"last_reload":      cm.lastReload,
		"reload_count":     atomic.LoadInt64(&cm.reloadCount),
		"is_watching":      cm.watcher != nil,
		"reload_delay":     cm.reloadDelay,
		"has_shadow":       cm.shadowConfig != nil,
		"has_rollback":     cm.rollbackConfig != nil,
		"validation_rules": cm.validationRules != nil,
	}
}
