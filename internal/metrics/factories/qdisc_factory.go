// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package factories

import (
	"errors"

	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/collectors/qdisc"
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/config"
	"gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
	"github.com/sirupsen/logrus"
)

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

func (qf *QdiscFactory) CreateCollector(qdiscType string) (interfaces.MetricCollector, error) {
	var cfg *config.CollectorConfig
	cfg, exists := qf.GetConfig(qdiscType)
	if !exists {
		cfg = config.NewCollectorConfig()
		qf.AddConfig(qdiscType, cfg)
	}
	logger := logrus.New()
	switch qdiscType {
	case "codel":
		return qdisc.NewCodelCollector(*cfg, logger), nil
	case "qdisc":
		return qdisc.NewQdiscCollector(*cfg, logger), nil
	default:
		return nil, errors.New("unsupported qdisc type: " + qdiscType)
	}
}
