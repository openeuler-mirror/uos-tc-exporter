// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package config

type ManagerConfig struct {
	Collectors map[string]*CollectorConfig
}

// NewManagerConfig 创建管理器配置
func NewManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		Collectors: make(map[string]*CollectorConfig),
	}
}

// GetCollectorConfig 获取收集器配置
func (mc *ManagerConfig) GetCollectorConfig(name string) *CollectorConfig {
	if cfg, exists := mc.Collectors[name]; exists {
		return cfg
	}
	return nil
}

// SetCollectorConfig 设置收集器配置
func (mc *ManagerConfig) SetCollectorConfig(name string, cfg *CollectorConfig) {
	mc.Collectors[name] = cfg
}
