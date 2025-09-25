// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package factories

import "gitee.com/openeuler/uos-tc-exporter/internal/metrics/config"

type QdiscFactory struct {
	configs map[string]*config.CollectorConfig
}

func NewQdiscFactory() *QdiscFactory {
	return &QdiscFactory{
		configs: make(map[string]*config.CollectorConfig),
	}
}

func (qf *QdiscFactory) GetConfig(qdiscType string) (*config.CollectorConfig, bool) {
	cfg, exists := qf.configs[qdiscType]
	return cfg, exists
}

func (qf *QdiscFactory) AddConfig(qdiscType string, cfg *config.CollectorConfig) {
	qf.configs[qdiscType] = cfg
}
func (qf *QdiscFactory) RemoveConfig(qdiscType string) {
	delete(qf.configs, qdiscType)
}

func (qf *QdiscFactory) GetSupportedTypes() []string {
	return []string{
		"codel", "cbq", "htb", "fq", "fq_codel",
		"choke", "pie", "red", "sfb", "sfq", "hfsc",
	}
}
