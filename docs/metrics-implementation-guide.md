# Metrics 架构重构实现指南

## 1. 实现步骤详解

### 1.1 第一步：创建目录结构

```bash
mkdir -p internal/metrics/{interfaces,base,collectors/{qdisc,class,app,business},factories,config,registry,utils}
```

### 1.2 第二步：实现核心接口

#### 1.2.1 基础收集器接口
```go
// internal/metrics/interfaces/collector.go
package interfaces

import "github.com/prometheus/client_golang/prometheus"

// MetricCollector 定义指标收集器的统一接口
type MetricCollector interface {
    // Collect 收集指标数据
    Collect(ch chan<- prometheus.Metric)
    
    // ID 返回收集器的唯一标识符
    ID() string
    
    // Name 返回收集器的名称
    Name() string
    
    // Description 返回收集器的描述
    Description() string
    
    // Enabled 检查收集器是否启用
    Enabled() bool
    
    // SetEnabled 设置收集器启用状态
    SetEnabled(enabled bool)
    
    // GetConfig 获取收集器配置
    GetConfig() interface{}
    
    // SetConfig 设置收集器配置
    SetConfig(config interface{}) error
}

// QdiscCollector 定义 qdisc 指标收集器的接口
type QdiscCollector interface {
    MetricCollector
    
    // GetQdiscType 返回 qdisc 类型
    GetQdiscType() string
    
    // GetSupportedMetrics 返回支持的指标列表
    GetSupportedMetrics() []string
    
    // ValidateQdisc 验证 qdisc 是否支持
    ValidateQdisc(qdisc interface{}) bool
}

// ClassCollector 定义 class 指标收集器的接口
type ClassCollector interface {
    MetricCollector
    
    // GetClassType 返回 class 类型
    GetClassType() string
    
    // GetSupportedMetrics 返回支持的指标列表
    GetSupportedMetrics() []string
    
    // ValidateClass 验证 class 是否支持
    ValidateClass(class interface{}) bool
}
```

#### 1.2.2 配置接口
```go
// internal/metrics/interfaces/config.go
package interfaces

import "time"

// ConfigProvider 定义配置提供者接口
type ConfigProvider interface {
    // GetConfig 获取配置
    GetConfig() interface{}
    
    // ReloadConfig 重新加载配置
    ReloadConfig() error
    
    // ValidateConfig 验证配置
    ValidateConfig() error
}

// CollectorConfig 收集器配置接口
type CollectorConfig interface {
    // IsEnabled 检查是否启用
    IsEnabled() bool
    
    // GetTimeout 获取超时时间
    GetTimeout() time.Duration
    
    // GetRetryCount 获取重试次数
    GetRetryCount() int
    
    // GetMetrics 获取指标配置
    GetMetrics() map[string]MetricConfig
}

// MetricConfig 指标配置接口
type MetricConfig interface {
    // GetName 获取指标名称
    GetName() string
    
    // IsEnabled 检查是否启用
    IsEnabled() bool
    
    // GetHelp 获取帮助信息
    GetHelp() string
    
    // GetType 获取指标类型
    GetType() string
    
    // GetLabels 获取标签列表
    GetLabels() []string
}
```

### 1.3 第三步：实现基础类

#### 1.3.1 基础收集器
```go
// internal/metrics/base/collector_base.go
package base

import (
    "fmt"
    "sync"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/sirupsen/logrus"
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
)

// CollectorBase 基础收集器实现
type CollectorBase struct {
    mu          sync.RWMutex
    id          string
    name        string
    description string
    enabled     bool
    config      interfaces.CollectorConfig
    logger      *logrus.Logger
    metrics     map[string]*prometheus.Desc
    lastError   error
    lastCollect time.Time
}

// NewCollectorBase 创建基础收集器
func NewCollectorBase(id, name, description string, config interfaces.CollectorConfig) *CollectorBase {
    return &CollectorBase{
        id:          id,
        name:        name,
        description: description,
        enabled:     true,
        config:      config,
        logger:      logrus.New(),
        metrics:     make(map[string]*prometheus.Desc),
    }
}

// Collect 实现 MetricCollector 接口
func (cb *CollectorBase) Collect(ch chan<- prometheus.Metric) {
    cb.mu.RLock()
    if !cb.enabled {
        cb.mu.RUnlock()
        return
    }
    cb.mu.RUnlock()
    
    start := time.Now()
    defer func() {
        cb.mu.Lock()
        cb.lastCollect = time.Now()
        cb.mu.Unlock()
        
        duration := time.Since(start)
        cb.logger.Debugf("Collector %s collected metrics in %v", cb.id, duration)
    }()
    
    // 子类实现具体的收集逻辑
    cb.collectMetrics(ch)
}

// ID 实现 MetricCollector 接口
func (cb *CollectorBase) ID() string {
    return cb.id
}

// Name 实现 MetricCollector 接口
func (cb *CollectorBase) Name() string {
    return cb.name
}

// Description 实现 MetricCollector 接口
func (cb *CollectorBase) Description() string {
    return cb.description
}

// Enabled 实现 MetricCollector 接口
func (cb *CollectorBase) Enabled() bool {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    return cb.enabled
}

// SetEnabled 实现 MetricCollector 接口
func (cb *CollectorBase) SetEnabled(enabled bool) {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    cb.enabled = enabled
}

// GetConfig 实现 MetricCollector 接口
func (cb *CollectorBase) GetConfig() interface{} {
    return cb.config
}

// SetConfig 实现 MetricCollector 接口
func (cb *CollectorBase) SetConfig(config interface{}) error {
    if collectorConfig, ok := config.(interfaces.CollectorConfig); ok {
        cb.mu.Lock()
        defer cb.mu.Unlock()
        cb.config = collectorConfig
        return nil
    }
    return fmt.Errorf("invalid config type")
}

// collectMetrics 子类需要实现的收集逻辑
func (cb *CollectorBase) collectMetrics(ch chan<- prometheus.Metric) {
    // 默认实现为空，子类需要重写
}

// addMetric 添加指标描述符
func (cb *CollectorBase) addMetric(name string, desc *prometheus.Desc) {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    cb.metrics[name] = desc
}

// getMetric 获取指标描述符
func (cb *CollectorBase) getMetric(name string) (*prometheus.Desc, bool) {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    desc, exists := cb.metrics[name]
    return desc, exists
}

// setLastError 设置最后错误
func (cb *CollectorBase) setLastError(err error) {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    cb.lastError = err
}

// getLastError 获取最后错误
func (cb *CollectorBase) getLastError() error {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    return cb.lastError
}
```

#### 1.3.2 Qdisc 基础实现
```go
// internal/metrics/base/qdisc_base.go
package base

import (
    "gitee.com/openeuler/uos-tc-exporter/internal/tc"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/sirupsen/logrus"
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
)

// QdiscBase qdisc 基础实现
type QdiscBase struct {
    *CollectorBase
    qdiscType        string
    supportedMetrics []string
    labelNames       []string
}

// NewQdiscBase 创建 qdisc 基础实例
func NewQdiscBase(qdiscType, name, description string, config interfaces.CollectorConfig) *QdiscBase {
    base := NewCollectorBase(
        "qdisc_"+qdiscType,
        name,
        description,
        config,
    )
    
    return &QdiscBase{
        CollectorBase:   base,
        qdiscType:       qdiscType,
        supportedMetrics: make([]string, 0),
        labelNames:      []string{"namespace", "device", "kind"},
    }
}

// collectMetrics 实现 qdisc 收集逻辑
func (qb *QdiscBase) collectMetrics(ch chan<- prometheus.Metric) {
    qb.logger.Info("Start collecting qdisc metrics")
    
    nsList, err := tc.GetNetNameSpaceList()
    if err != nil {
        qb.logger.Warnf("Get net namespace list failed: %v", err)
        qb.setLastError(err)
        return
    }
    
    if len(nsList) == 0 {
        qb.logger.Info("No net namespace found")
        return
    }
    
    for _, ns := range nsList {
        qb.collectForNamespace(ch, ns)
    }
}

// collectForNamespace 收集指定命名空间的指标
func (qb *QdiscBase) collectForNamespace(ch chan<- prometheus.Metric, ns string) {
    devices, err := tc.GetInterfaceInNetNS(ns)
    if err != nil {
        qb.logger.Warnf("Get interface in netns %s failed: %v", ns, err)
        return
    }
    
    for _, device := range devices {
        qb.collectForDevice(ch, ns, device)
    }
}

// collectForDevice 收集指定设备的指标
func (qb *QdiscBase) collectForDevice(ch chan<- prometheus.Metric, ns string, device interface{}) {
    // 获取设备索引
    deviceIndex, deviceName, err := qb.extractDeviceInfo(device)
    if err != nil {
        qb.logger.Warnf("Extract device info failed: %v", err)
        return
    }
    
    qdiscs, err := tc.GetQdiscs(deviceIndex, ns)
    if err != nil {
        qb.logger.Warnf("Get qdiscs in netns %s failed: %v", ns, err)
        return
    }
    
    for _, qdisc := range qdiscs {
        if !qb.ValidateQdisc(&qdisc) {
            continue
        }
        
        qb.collectQdiscMetrics(ch, ns, deviceName, &qdisc)
    }
}

// extractDeviceInfo 提取设备信息
func (qb *QdiscBase) extractDeviceInfo(device interface{}) (uint32, string, error) {
    // 这里需要根据实际的设备类型进行转换
    // 假设设备有 Index 和 Attributes.Name 字段
    if dev, ok := device.(interface {
        GetIndex() uint32
        GetName() string
    }); ok {
        return dev.GetIndex(), dev.GetName(), nil
    }
    return 0, "", fmt.Errorf("unsupported device type")
}

// ValidateQdisc 验证 qdisc 是否支持
func (qb *QdiscBase) ValidateQdisc(qdisc interface{}) bool {
    // 子类需要实现具体的验证逻辑
    return true
}

// collectQdiscMetrics 收集 qdisc 指标
func (qb *QdiscBase) collectQdiscMetrics(ch chan<- prometheus.Metric, ns, deviceName string, qdisc interface{}) {
    // 子类需要实现具体的指标收集逻辑
}

// GetQdiscType 返回 qdisc 类型
func (qb *QdiscBase) GetQdiscType() string {
    return qb.qdiscType
}

// GetSupportedMetrics 返回支持的指标列表
func (qb *QdiscBase) GetSupportedMetrics() []string {
    return qb.supportedMetrics
}

// addSupportedMetric 添加支持的指标
func (qb *QdiscBase) addSupportedMetric(metricName string) {
    qb.supportedMetrics = append(qb.supportedMetrics, metricName)
}
```

### 1.4 第四步：实现配置管理

#### 1.4.1 收集器配置
```go
// internal/metrics/config/collector_config.go
package config

import (
    "time"
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
)

// CollectorConfig 收集器配置实现
type CollectorConfig struct {
    enabled           bool
    collectionTimeout time.Duration
    retryCount        int
    metrics           map[string]MetricConfig
    labels            []string
}

// NewCollectorConfig 创建收集器配置
func NewCollectorConfig() *CollectorConfig {
    return &CollectorConfig{
        enabled:           true,
        collectionTimeout: 30 * time.Second,
        retryCount:        3,
        metrics:           make(map[string]MetricConfig),
        labels:            []string{"namespace", "device", "kind"},
    }
}

// IsEnabled 实现 CollectorConfig 接口
func (cc *CollectorConfig) IsEnabled() bool {
    return cc.enabled
}

// GetTimeout 实现 CollectorConfig 接口
func (cc *CollectorConfig) GetTimeout() time.Duration {
    return cc.collectionTimeout
}

// GetRetryCount 实现 CollectorConfig 接口
func (cc *CollectorConfig) GetRetryCount() int {
    return cc.retryCount
}

// GetMetrics 实现 CollectorConfig 接口
func (cc *CollectorConfig) GetMetrics() map[string]MetricConfig {
    return cc.metrics
}

// SetEnabled 设置启用状态
func (cc *CollectorConfig) SetEnabled(enabled bool) {
    cc.enabled = enabled
}

// SetTimeout 设置超时时间
func (cc *CollectorConfig) SetTimeout(timeout time.Duration) {
    cc.collectionTimeout = timeout
}

// SetRetryCount 设置重试次数
func (cc *CollectorConfig) SetRetryCount(count int) {
    cc.retryCount = count
}

// AddMetric 添加指标配置
func (cc *CollectorConfig) AddMetric(name string, config MetricConfig) {
    cc.metrics[name] = config
}

// MetricConfig 指标配置实现
type MetricConfig struct {
    name    string
    enabled bool
    help    string
    mtype   string
    labels  []string
    buckets []float64
}

// NewMetricConfig 创建指标配置
func NewMetricConfig(name, help, mtype string) *MetricConfig {
    return &MetricConfig{
        name:    name,
        enabled: true,
        help:    help,
        mtype:   mtype,
        labels:  []string{"namespace", "device", "kind"},
    }
}

// GetName 实现 MetricConfig 接口
func (mc *MetricConfig) GetName() string {
    return mc.name
}

// IsEnabled 实现 MetricConfig 接口
func (mc *MetricConfig) IsEnabled() bool {
    return mc.enabled
}

// GetHelp 实现 MetricConfig 接口
func (mc *MetricConfig) GetHelp() string {
    return mc.help
}

// GetType 实现 MetricConfig 接口
func (mc *MetricConfig) GetType() string {
    return mc.mtype
}

// GetLabels 实现 MetricConfig 接口
func (mc *MetricConfig) GetLabels() []string {
    return mc.labels
}

// SetEnabled 设置启用状态
func (mc *MetricConfig) SetEnabled(enabled bool) {
    mc.enabled = enabled
}

// SetLabels 设置标签
func (mc *MetricConfig) SetLabels(labels []string) {
    mc.labels = labels
}

// SetBuckets 设置桶配置（用于直方图）
func (mc *MetricConfig) SetBuckets(buckets []float64) {
    mc.buckets = buckets
}
```

### 1.5 第五步：实现具体收集器

#### 1.5.1 Codel 收集器实现
```go
// internal/metrics/collectors/qdisc/codel.go
package qdisc

import (
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics/base"
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics/config"
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
    "github.com/prometheus/client_golang/prometheus"
)

// CodelCollector Codel qdisc 收集器
type CodelCollector struct {
    *base.QdiscBase
    metrics map[string]*prometheus.Desc
}

// NewCodelCollector 创建 Codel 收集器
func NewCodelCollector(cfg *config.CollectorConfig) *CodelCollector {
    base := base.NewQdiscBase("codel", "qdisc_codel", "Codel qdisc metrics", cfg)
    
    c := &CodelCollector{
        QdiscBase: base,
        metrics:   make(map[string]*prometheus.Desc),
    }
    
    c.initializeMetrics(cfg)
    return c
}

// initializeMetrics 初始化指标
func (c *CodelCollector) initializeMetrics(cfg *config.CollectorConfig) {
    labelNames := []string{"namespace", "device", "kind"}
    
    // CE Mark 指标
    c.metrics["ce_mark"] = prometheus.NewDesc(
        "qdisc_codel_ce_mark",
        "Codel CE mark xstat",
        labelNames, nil,
    )
    c.addSupportedMetric("ce_mark")
    
    // Count 指标
    c.metrics["count"] = prometheus.NewDesc(
        "qdisc_codel_count",
        "Codel count xstat",
        labelNames, nil,
    )
    c.addSupportedMetric("count")
    
    // Drop Next 指标
    c.metrics["drop_next"] = prometheus.NewDesc(
        "qdisc_codel_drop_next",
        "Codel drop next xstat",
        labelNames, nil,
    )
    c.addSupportedMetric("drop_next")
    
    // Drop Overlimit 指标
    c.metrics["drop_overlimit"] = prometheus.NewDesc(
        "qdisc_codel_drop_overlimit",
        "Codel drop overlimit xstat",
        labelNames, nil,
    )
    c.addSupportedMetric("drop_overlimit")
    
    // Dropping 指标
    c.metrics["dropping"] = prometheus.NewDesc(
        "qdisc_codel_dropping",
        "Codel dropping xstat",
        labelNames, nil,
    )
    c.addSupportedMetric("dropping")
    
    // ECN Mark 指标
    c.metrics["ecn_mark"] = prometheus.NewDesc(
        "qdisc_codel_ecn_mark",
        "Codel ECN mark xstat",
        labelNames, nil,
    )
    c.addSupportedMetric("ecn_mark")
    
    // LDelay 指标
    c.metrics["ldelay"] = prometheus.NewDesc(
        "qdisc_codel_ldelay",
        "Codel ldelay xstat",
        labelNames, nil,
    )
    c.addSupportedMetric("ldelay")
    
    // Max Packet 指标
    c.metrics["max_packet"] = prometheus.NewDesc(
        "qdisc_codel_max_packet",
        "Codel max packet xstat",
        labelNames, nil,
    )
    c.addSupportedMetric("max_packet")
}

// ValidateQdisc 验证 qdisc 是否支持
func (c *CodelCollector) ValidateQdisc(qdisc interface{}) bool {
    // 这里需要根据实际的 qdisc 结构进行验证
    // 假设 qdisc 有 Kind 字段
    if q, ok := qdisc.(interface{ GetKind() string }); ok {
        return q.GetKind() == "codel"
    }
    return false
}

// collectQdiscMetrics 收集 qdisc 指标
func (c *CodelCollector) collectQdiscMetrics(ch chan<- prometheus.Metric, ns, deviceName string, qdisc interface{}) {
    // 这里需要根据实际的 qdisc 结构进行数据提取
    // 假设 qdisc 有 XStats.Codel 字段
    if q, ok := qdisc.(interface{ GetXStats() interface{} }); ok {
        xstats := q.GetXStats()
        if codelStats, ok := xstats.(interface{ GetCodel() interface{} }); ok {
            codel := codelStats.GetCodel()
            c.collectCodelMetrics(ch, ns, deviceName, codel)
        }
    }
}

// collectCodelMetrics 收集 Codel 特定指标
func (c *CodelCollector) collectCodelMetrics(ch chan<- prometheus.Metric, ns, deviceName string, codel interface{}) {
    labels := []string{ns, deviceName, "codel"}
    
    // 这里需要根据实际的 Codel 结构进行数据提取
    // 假设 codel 有各种字段
    if c, ok := codel.(interface {
        GetCeMark() uint64
        GetCount() uint64
        GetDropNext() uint64
        GetDropOverlimit() uint64
        GetDropping() uint64
        GetEcnMark() uint64
        GetLDelay() uint64
        GetMaxPacket() uint64
    }); ok {
        // 收集 CE Mark
        if desc, exists := c.metrics["ce_mark"]; exists {
            ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 
                float64(c.GetCeMark()), labels...)
        }
        
        // 收集 Count
        if desc, exists := c.metrics["count"]; exists {
            ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 
                float64(c.GetCount()), labels...)
        }
        
        // 收集 Drop Next
        if desc, exists := c.metrics["drop_next"]; exists {
            ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 
                float64(c.GetDropNext()), labels...)
        }
        
        // 收集 Drop Overlimit
        if desc, exists := c.metrics["drop_overlimit"]; exists {
            ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 
                float64(c.GetDropOverlimit()), labels...)
        }
        
        // 收集 Dropping
        if desc, exists := c.metrics["dropping"]; exists {
            ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 
                float64(c.GetDropping()), labels...)
        }
        
        // 收集 ECN Mark
        if desc, exists := c.metrics["ecn_mark"]; exists {
            ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 
                float64(c.GetEcnMark()), labels...)
        }
        
        // 收集 LDelay
        if desc, exists := c.metrics["ldelay"]; exists {
            ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 
                float64(c.GetLDelay()), labels...)
        }
        
        // 收集 Max Packet
        if desc, exists := c.metrics["max_packet"]; exists {
            ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 
                float64(c.GetMaxPacket()), labels...)
        }
    }
}
```

### 1.6 第六步：实现工厂模式

#### 1.6.1 Qdisc 工厂
```go
// internal/metrics/factories/qdisc_factory.go
package factories

import (
    "fmt"
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics/collectors/qdisc"
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics/config"
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
)

// QdiscFactory qdisc 收集器工厂
type QdiscFactory struct {
    configs map[string]*config.CollectorConfig
}

// NewQdiscFactory 创建 qdisc 工厂
func NewQdiscFactory() *QdiscFactory {
    return &QdiscFactory{
        configs: make(map[string]*config.CollectorConfig),
    }
}

// CreateCollector 创建收集器
func (qf *QdiscFactory) CreateCollector(qdiscType string) (interfaces.QdiscCollector, error) {
    cfg, exists := qf.configs[qdiscType]
    if !exists {
        cfg = config.NewCollectorConfig()
        qf.configs[qdiscType] = cfg
    }
    
    switch qdiscType {
    case "codel":
        return qdisc.NewCodelCollector(cfg), nil
    case "cbq":
        return qdisc.NewCbqCollector(cfg), nil
    case "htb":
        return qdisc.NewHtbCollector(cfg), nil
    case "fq":
        return qdisc.NewFqCollector(cfg), nil
    case "fq_codel":
        return qdisc.NewFqCodelCollector(cfg), nil
    case "choke":
        return qdisc.NewChokeCollector(cfg), nil
    case "pie":
        return qdisc.NewPieCollector(cfg), nil
    case "red":
        return qdisc.NewRedCollector(cfg), nil
    case "sfb":
        return qdisc.NewSfbCollector(cfg), nil
    case "sfq":
        return qdisc.NewSfqCollector(cfg), nil
    case "hfsc":
        return qdisc.NewHfscCollector(cfg), nil
    default:
        return nil, fmt.Errorf("unsupported qdisc type: %s", qdiscType)
    }
}

// GetSupportedTypes 获取支持的 qdisc 类型
func (qf *QdiscFactory) GetSupportedTypes() []string {
    return []string{
        "codel", "cbq", "htb", "fq", "fq_codel",
        "choke", "pie", "red", "sfb", "sfq", "hfsc",
    }
}

// SetConfig 设置配置
func (qf *QdiscFactory) SetConfig(qdiscType string, cfg *config.CollectorConfig) {
    qf.configs[qdiscType] = cfg
}

// GetConfig 获取配置
func (qf *QdiscFactory) GetConfig(qdiscType string) (*config.CollectorConfig, bool) {
    cfg, exists := qf.configs[qdiscType]
    return cfg, exists
}
```

### 1.7 第七步：实现注册中心

#### 1.7.1 收集器注册中心
```go
// internal/metrics/registry/collector_registry.go
package registry

import (
    "fmt"
    "sync"
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
)

// CollectorRegistry 收集器注册中心
type CollectorRegistry struct {
    mu         sync.RWMutex
    collectors map[string]interfaces.MetricCollector
    factories  map[string]CollectorFactory
}

// CollectorFactory 收集器工厂接口
type CollectorFactory interface {
    CreateCollector(collectorType string) (interfaces.MetricCollector, error)
    GetSupportedTypes() []string
}

// NewCollectorRegistry 创建收集器注册中心
func NewCollectorRegistry() *CollectorRegistry {
    return &CollectorRegistry{
        collectors: make(map[string]interfaces.MetricCollector),
        factories:  make(map[string]CollectorFactory),
    }
}

// Register 注册收集器
func (cr *CollectorRegistry) Register(collector interfaces.MetricCollector) error {
    cr.mu.Lock()
    defer cr.mu.Unlock()
    
    if _, exists := cr.collectors[collector.ID()]; exists {
        return fmt.Errorf("collector %s already registered", collector.ID())
    }
    
    cr.collectors[collector.ID()] = collector
    return nil
}

// Unregister 注销收集器
func (cr *CollectorRegistry) Unregister(id string) error {
    cr.mu.Lock()
    defer cr.mu.Unlock()
    
    if _, exists := cr.collectors[id]; !exists {
        return fmt.Errorf("collector %s not found", id)
    }
    
    delete(cr.collectors, id)
    return nil
}

// GetCollector 获取收集器
func (cr *CollectorRegistry) GetCollector(id string) (interfaces.MetricCollector, bool) {
    cr.mu.RLock()
    defer cr.mu.RUnlock()
    collector, exists := cr.collectors[id]
    return collector, exists
}

// GetAllCollectors 获取所有收集器
func (cr *CollectorRegistry) GetAllCollectors() map[string]interfaces.MetricCollector {
    cr.mu.RLock()
    defer cr.mu.RUnlock()
    
    result := make(map[string]interfaces.MetricCollector)
    for id, collector := range cr.collectors {
        result[id] = collector
    }
    return result
}

// GetEnabledCollectors 获取启用的收集器
func (cr *CollectorRegistry) GetEnabledCollectors() map[string]interfaces.MetricCollector {
    cr.mu.RLock()
    defer cr.mu.RUnlock()
    
    result := make(map[string]interfaces.MetricCollector)
    for id, collector := range cr.collectors {
        if collector.Enabled() {
            result[id] = collector
        }
    }
    return result
}

// RegisterFactory 注册工厂
func (cr *CollectorRegistry) RegisterFactory(factoryType string, factory CollectorFactory) {
    cr.mu.Lock()
    defer cr.mu.Unlock()
    cr.factories[factoryType] = factory
}

// CreateCollector 创建收集器
func (cr *CollectorRegistry) CreateCollector(factoryType, collectorType string) (interfaces.MetricCollector, error) {
    cr.mu.RLock()
    factory, exists := cr.factories[factoryType]
    cr.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("factory %s not found", factoryType)
    }
    
    return factory.CreateCollector(collectorType)
}

// GetFactory 获取工厂
func (cr *CollectorRegistry) GetFactory(factoryType string) (CollectorFactory, bool) {
    cr.mu.RLock()
    defer cr.mu.RUnlock()
    factory, exists := cr.factories[factoryType]
    return factory, exists
}
```

### 1.8 第八步：实现新管理器

#### 1.8.1 管理器 V2
```go
// internal/metrics/manager_v2.go
package metrics

import (
    "sync"
    "time"
    "github.com/prometheus/client_golang/prometheus"
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics/registry"
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics/factories"
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics/config"
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics/interfaces"
)

// ManagerV2 新版本管理器
type ManagerV2 struct {
    mu         sync.RWMutex
    registry   *registry.CollectorRegistry
    factories  map[string]registry.CollectorFactory
    config     *config.ManagerConfig
    stats      *CollectionStats
    logger     *logrus.Logger
}

// NewManagerV2 创建新版本管理器
func NewManagerV2(cfg *config.ManagerConfig) *ManagerV2 {
    m := &ManagerV2{
        registry:  registry.NewCollectorRegistry(),
        factories: make(map[string]registry.CollectorFactory),
        config:    cfg,
        stats:     &CollectionStats{},
        logger:    logrus.New(),
    }
    
    m.initializeFactories()
    m.registerCollectors()
    
    return m
}

// initializeFactories 初始化工厂
func (m *ManagerV2) initializeFactories() {
    // 注册 qdisc 工厂
    qdiscFactory := factories.NewQdiscFactory()
    m.registry.RegisterFactory("qdisc", qdiscFactory)
    m.factories["qdisc"] = qdiscFactory
    
    // 注册 class 工厂
    classFactory := factories.NewClassFactory()
    m.registry.RegisterFactory("class", classFactory)
    m.factories["class"] = classFactory
    
    // 注册 app 工厂
    appFactory := factories.NewAppFactory()
    m.registry.RegisterFactory("app", appFactory)
    m.factories["app"] = appFactory
    
    // 注册 business 工厂
    businessFactory := factories.NewBusinessFactory()
    m.registry.RegisterFactory("business", businessFactory)
    m.factories["business"] = businessFactory
}

// registerCollectors 注册收集器
func (m *ManagerV2) registerCollectors() {
    // 注册 qdisc 收集器
    qdiscTypes := []string{"codel", "cbq", "htb", "fq", "fq_codel", "choke", "pie", "red", "sfb", "sfq", "hfsc"}
    for _, qdiscType := range qdiscTypes {
        if collector, err := m.registry.CreateCollector("qdisc", qdiscType); err == nil {
            m.registry.Register(collector)
        } else {
            m.logger.Warnf("Failed to create qdisc collector %s: %v", qdiscType, err)
        }
    }
    
    // 注册应用收集器
    if appCollector, err := m.registry.CreateCollector("app", "app_metrics"); err == nil {
        m.registry.Register(appCollector)
    } else {
        m.logger.Warnf("Failed to create app collector: %v", err)
    }
    
    // 注册业务收集器
    if businessCollector, err := m.registry.CreateCollector("business", "business_metrics"); err == nil {
        m.registry.Register(businessCollector)
    } else {
        m.logger.Warnf("Failed to create business collector: %v", err)
    }
}

// CollectAll 收集所有指标
func (m *ManagerV2) CollectAll(ch chan<- prometheus.Metric) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        m.stats.RecordCollection(duration, true, nil)
    }()
    
    collectors := m.registry.GetEnabledCollectors()
    for _, collector := range collectors {
        if collector.Enabled() {
            collector.Collect(ch)
        }
    }
}

// GetCollector 获取收集器
func (m *ManagerV2) GetCollector(id string) (interfaces.MetricCollector, bool) {
    return m.registry.GetCollector(id)
}

// EnableCollector 启用收集器
func (m *ManagerV2) EnableCollector(id string) error {
    collector, exists := m.registry.GetCollector(id)
    if !exists {
        return fmt.Errorf("collector %s not found", id)
    }
    
    collector.SetEnabled(true)
    return nil
}

// DisableCollector 禁用收集器
func (m *ManagerV2) DisableCollector(id string) error {
    collector, exists := m.registry.GetCollector(id)
    if !exists {
        return fmt.Errorf("collector %s not found", id)
    }
    
    collector.SetEnabled(false)
    return nil
}

// GetStats 获取统计信息
func (m *ManagerV2) GetStats() *CollectionStats {
    return m.stats
}

// Shutdown 关闭管理器
func (m *ManagerV2) Shutdown() {
    m.logger.Info("Shutting down metrics manager v2")
    // 这里可以添加清理逻辑
}
```

## 2. 测试实现

### 2.1 单元测试示例
```go
// internal/metrics/collectors/qdisc/codel_test.go
package qdisc

import (
    "testing"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/stretchr/testify/assert"
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics/config"
)

func TestNewCodelCollector(t *testing.T) {
    cfg := config.NewCollectorConfig()
    collector := NewCodelCollector(cfg)
    
    assert.NotNil(t, collector)
    assert.Equal(t, "qdisc_codel", collector.ID())
    assert.Equal(t, "qdisc_codel", collector.Name())
    assert.Equal(t, "Codel qdisc metrics", collector.Description())
    assert.True(t, collector.Enabled())
}

func TestCodelCollector_GetQdiscType(t *testing.T) {
    cfg := config.NewCollectorConfig()
    collector := NewCodelCollector(cfg)
    
    assert.Equal(t, "codel", collector.GetQdiscType())
}

func TestCodelCollector_GetSupportedMetrics(t *testing.T) {
    cfg := config.NewCollectorConfig()
    collector := NewCodelCollector(cfg)
    
    metrics := collector.GetSupportedMetrics()
    assert.Contains(t, metrics, "ce_mark")
    assert.Contains(t, metrics, "count")
    assert.Contains(t, metrics, "drop_next")
    assert.Contains(t, metrics, "drop_overlimit")
    assert.Contains(t, metrics, "dropping")
    assert.Contains(t, metrics, "ecn_mark")
    assert.Contains(t, metrics, "ldelay")
    assert.Contains(t, metrics, "max_packet")
}

func TestCodelCollector_Collect(t *testing.T) {
    cfg := config.NewCollectorConfig()
    collector := NewCodelCollector(cfg)
    
    ch := make(chan prometheus.Metric, 10)
    collector.Collect(ch)
    
    // 验证指标被收集
    close(ch)
    metrics := make([]prometheus.Metric, 0)
    for metric := range ch {
        metrics = append(metrics, metric)
    }
    
    // 这里需要根据实际的收集逻辑进行验证
    assert.GreaterOrEqual(t, len(metrics), 0)
}
```

## 3. 配置示例

### 3.1 YAML 配置示例
```yaml
# config/metrics.yaml
manager:
  performance_monitoring: true
  collection_interval: 30s
  stats_retention: 24h
  logging:
    level: info
    format: json

collectors:
  qdisc:
    enabled: true
    collection_timeout: 30s
    retry_count: 3
    types:
      codel:
        enabled: true
        metrics:
          ce_mark:
            enabled: true
            help: "Codel CE mark xstat"
          count:
            enabled: true
            help: "Codel count xstat"
      cbq:
        enabled: true
        metrics:
          avg_idle:
            enabled: true
            help: "CBQ avg idle xstat"
      htb:
        enabled: true
        metrics:
          borrows:
            enabled: true
            help: "HTB borrows xstat"
  
  app:
    enabled: true
    collection_timeout: 10s
    retry_count: 1
  
  business:
    enabled: true
    collection_timeout: 15s
    retry_count: 2
```

## 4. 使用示例

### 4.1 基本使用
```go
package main

import (
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics/config"
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics"
)

func main() {
    // 创建配置
    cfg := config.NewManagerConfig()
    
    // 创建管理器
    manager := metrics.NewManagerV2(cfg)
    
    // 收集指标
    ch := make(chan prometheus.Metric, 100)
    manager.CollectAll(ch)
    
    // 处理指标
    for metric := range ch {
        // 处理指标
        fmt.Println(metric.Desc().String())
    }
}
```

### 4.2 动态管理
```go
// 启用特定收集器
err := manager.EnableCollector("qdisc_codel")
if err != nil {
    log.Fatal(err)
}

// 禁用特定收集器
err = manager.DisableCollector("qdisc_cbq")
if err != nil {
    log.Fatal(err)
}

// 获取统计信息
stats := manager.GetStats()
fmt.Printf("Total collections: %d\n", stats.TotalCollections)
fmt.Printf("Successful collections: %d\n", stats.SuccessfulCollections)
fmt.Printf("Failed collections: %d\n", stats.FailedCollections)
```

这个实现指南提供了完整的代码示例和实现步骤，可以帮助开发团队按照统一的架构进行代码重构。
