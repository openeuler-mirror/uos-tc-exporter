// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"os"

	"gitee.com/openeuler/uos-tc-exporter/internal/exporter"
	"gitee.com/openeuler/uos-tc-exporter/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// ConfigManager 负责配置管理
type ConfigManager struct {
	configMgr *exporter.ConfigManager
	config    exporter.Config
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager() *ConfigManager {
	// 创建配置管理器
	configMgr, err := exporter.NewConfigManager(*exporter.Configfile)
	if err != nil {
		logrus.Warnf("Failed to create config manager: %v, will use static config", err)
		configMgr = nil
	}

	return &ConfigManager{
		configMgr: configMgr,
		config:    exporter.DefaultConfig,
	}
}

// LoadConfig 加载配置
func (cm *ConfigManager) LoadConfig() error {
	if cm.configMgr != nil {
		// 使用配置管理器加载配置
		if err := cm.configMgr.LoadConfig(); err != nil {
			logrus.Errorf("Failed to load config via config manager: %v", err)
			logrus.Info("Use default config")
			return nil
		}

		// 获取配置并同步
		cm.config = cm.configMgr.GetConfig()

		// 设置配置重载回调
		cm.configMgr.SetReloadCallback(cm.onConfigReload)

		logrus.Infof("Loaded config via config manager from: %s", *exporter.Configfile)
		logrus.Info("Config file loaded and validated successfully")
		return nil
	}

	// 回退到原有的静态配置加载方式
	content, err := os.ReadFile(*exporter.Configfile)
	if err != nil {
		logrus.Errorf("Failed to read config file: %v", err)
		logrus.Info("Use default config")
		return nil
	}
	err = yaml.Unmarshal(content, &cm.config)
	if err != nil {
		logrus.Errorf("Failed to parse config file: %v", err)
		logrus.Info("Use default config")
		return nil
	}

	// 验证配置的有效性
	if err := cm.config.Validate(); err != nil {
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
	logrus.Info("Config file loaded and validated successfully")
	return nil
}

// GetConfig 获取当前配置
func (cm *ConfigManager) GetConfig() exporter.Config {
	return cm.config
}

// StartWatching 启动配置监控
func (cm *ConfigManager) StartWatching() error {
	if cm.configMgr == nil {
		return nil
	}

	// 启动配置监控
	if err := cm.configMgr.StartWatching(context.TODO()); err != nil {
		return err
	}

	logrus.Info("Config hot reload enabled")
	return nil
}

// StopWatching 停止配置监控
func (cm *ConfigManager) StopWatching() error {
	if cm.configMgr == nil {
		return nil
	}

	if err := cm.configMgr.StopWatching(); err != nil {
		return err
	}

	logrus.Info("Config watching stopped")
	return nil
}

// onConfigReload 配置重载回调函数
func (cm *ConfigManager) onConfigReload(newConfig *exporter.Config) error {
	logrus.Info("Configuration reload triggered")

	// 同步新配置
	cm.config = *newConfig

	logrus.Info("Configuration reload completed successfully")
	return nil
}
