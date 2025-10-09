// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package exporter

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
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
	reloadCount int
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager(configPath string) (*ConfigManager, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	cm := &ConfigManager{
		configPath:  configPath,
		config:      &DefaultConfig,
		watcher:     watcher,
		reloadChan:  make(chan struct{}, 1),
		stopChan:    make(chan struct{}),
		reloadDelay: 2 * time.Second, // 防抖延迟
		lastReload:  time.Now(),
	}

	// 设置默认配置
	*cm.config = DefaultConfig

	return cm, nil
}

// LoadConfig 加载配置文件
func (cm *ConfigManager) LoadConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 检查文件是否存在
	if !cm.fileExists() {
		logrus.Warnf("Config file %s not found, using default config", cm.configPath)
		return errors.New("config file not found")
	}
	// 读取配置文件
	content, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", cm.configPath, err)
	}

	// 解析YAML
	var newConfig Config
	if err := yaml.Unmarshal(content, &newConfig); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", cm.configPath, err)
	}

	// 验证配置
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 应用配置
	cm.config = &newConfig

	logrus.Debugf("Config : address=%s, port=%d, metricsPath=%s",
		cm.config.Address, cm.config.Port, cm.config.MetricsPath)
	return nil
}

func (cm *ConfigManager) Reload() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 检查文件是否存在
	if !cm.fileExists() {
		logrus.Warnf("Config file %s not found, using default config", cm.configPath)
		return errors.New("config file not found")
	}
	// 读取配置文件
	content, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", cm.configPath, err)
	}

	// 解析YAML
	var newConfig Config
	if err := yaml.Unmarshal(content, &newConfig); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", cm.configPath, err)
	}

	// 验证配置
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 应用新配置
	oldConfig := cm.config
	cm.config = &newConfig
	cm.lastReload = time.Now()
	cm.reloadCount++

	logrus.Infof("Config reloaded successfully from %s (reload count: %d)", cm.configPath, cm.reloadCount)
	logrus.Debugf("Config changed: address=%s, port=%d, metricsPath=%s",
		cm.config.Address, cm.config.Port, cm.config.MetricsPath)

	// 调用重载回调
	if cm.onReload != nil {
		if err := cm.onReload(cm.config); err != nil {
			logrus.Errorf("Config reload callback failed: %v", err)
			// 回滚到旧配置
			cm.config = oldConfig
			return err
		}
	}

	return nil
}

// GetConfig 获取当前配置的副本
func (cm *ConfigManager) GetConfig() Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.config == nil {
		return DefaultConfig
	}
	return *cm.config
}

// GetConfigPtr 获取当前配置的指针（用于内部使用）
func (cm *ConfigManager) GetConfigPtr() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
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
			if filepath.Clean(event.Name) == filepath.Clean(cm.configPath) {
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					// 防抖处理：延迟重载
					// 保证在配置文件频繁变化时，只进行一次重载
					if reloadTimer != nil {
						reloadTimer.Stop()
					}
					reloadTimer = time.AfterFunc(cm.reloadDelay, func() {
						select {
						// 非阻塞发送重载信号
						case cm.reloadChan <- struct{}{}:
						default:
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
		"config_path":  cm.configPath,
		"last_reload":  cm.lastReload,
		"reload_count": cm.reloadCount,
		"is_watching":  cm.watcher != nil,
		"reload_delay": cm.reloadDelay,
	}
}
