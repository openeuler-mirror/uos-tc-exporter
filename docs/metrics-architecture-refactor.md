# Metrics 架构重构设计文档

## 1. 概述

### 1.1 背景
当前 `internal/metrics` 目录下的代码存在以下问题：
- 代码风格不统一，不同 qdisc 类型的实现差异很大
- 重复代码严重，每个 qdisc 都有相似的 Collect 逻辑
- 缺乏抽象层，没有统一的接口和基类
- 可维护性差，添加新的 qdisc 类型需要大量重复代码
- 扩展性不足，难以添加新功能或修改现有逻辑
- 测试覆盖不完整，部分文件缺少测试

### 1.2 目标
- 统一代码风格和架构模式
- 减少重复代码，提高可维护性
- 增强扩展性，支持插件化架构
- 提高测试覆盖率和代码质量
- 保持向后兼容性

## 2. 当前架构分析

### 2.1 现有文件结构
```
internal/metrics/
├── app_metrics.go          # 应用指标
├── business_metrics.go     # 业务指标
├── manager.go             # 管理器
├── metrics.go             # 基础指标
├── q_*.go                 # 各种 qdisc 指标 (10+ 文件)
├── qclass.go              # class 指标
├── qdisc.go               # qdisc 基础
├── logger.go              # 日志
├── performance_wrapper.go # 性能包装器
└── *_test.go              # 测试文件
```

### 2.2 问题分析

#### 2.2.1 代码重复
- 每个 qdisc 文件都有相似的 `Collect` 方法实现
- 相同的网络命名空间遍历逻辑
- 重复的错误处理和日志记录

#### 2.2.2 缺乏抽象
- 没有统一的收集器接口
- 缺少基类来共享通用逻辑
- 指标定义分散在各个文件中

#### 2.2.3 配置管理混乱
- 硬编码的配置值
- 缺乏统一的配置管理
- 难以动态调整收集行为

#### 2.2.4 测试覆盖不足
- 部分文件缺少测试
- 测试风格不统一
- 缺乏集成测试

## 3. 新架构设计

### 3.1 整体架构

```
internal/metrics/
├── interfaces/           # 接口定义层
│   ├── collector.go     # 基础收集器接口
│   ├── qdisc.go         # qdisc 收集器接口
│   ├── class.go         # class 收集器接口
│   └── app.go           # 应用收集器接口
├── base/                # 基础实现层
│   ├── collector_base.go    # 基础收集器
│   ├── qdisc_base.go        # qdisc 基础实现
│   ├── class_base.go        # class 基础实现
│   └── metric_definitions.go # 指标定义
├── collectors/          # 具体实现层
│   ├── qdisc/           # qdisc 收集器
│   │   ├── codel.go     # Codel 收集器
│   │   ├── cbq.go       # CBQ 收集器
│   │   ├── htb.go       # HTB 收集器
│   │   └── ...
│   ├── class/           # class 收集器
│   ├── app/             # 应用收集器
│   └── business/        # 业务收集器
├── factories/           # 工厂层
│   ├── collector_factory.go  # 收集器工厂
│   ├── qdisc_factory.go      # qdisc 工厂
│   └── class_factory.go      # class 工厂
├── config/              # 配置层
│   ├── collector_config.go   # 收集器配置
│   ├── metrics_config.go     # 指标配置
│   └── manager_config.go     # 管理器配置
├── registry/            # 注册中心
│   ├── collector_registry.go # 收集器注册中心
│   └── metric_registry.go    # 指标注册中心
├── utils/               # 工具层
│   ├── tc_helper.go     # TC 操作辅助
│   ├── metric_helper.go # 指标辅助
│   └── validation.go    # 验证工具
└── manager_v2.go        # 新管理器
```

### 3.2 核心设计原则

#### 3.2.1 统一接口
- 所有收集器实现相同的 `MetricCollector` 接口
- 特定类型的收集器实现对应的扩展接口
- 统一的错误处理和日志记录

#### 3.2.2 代码复用
- 通过基类共享通用逻辑
- 工厂模式创建收集器实例
- 配置驱动的行为控制

#### 3.2.3 插件化架构
- 支持动态注册和发现收集器
- 配置驱动的收集器启用/禁用
- 支持自定义收集器扩展

#### 3.2.4 测试友好
- 每个组件都有对应接口
- 支持依赖注入和模拟
- 完整的单元测试和集成测试

## 4. 详细设计

### 4.1 接口层设计

#### 4.1.1 基础收集器接口
```go
type MetricCollector interface {
    Collect(ch chan<- prometheus.Metric)
    ID() string
    Name() string
    Description() string
    Enabled() bool
    SetEnabled(enabled bool)
}
```

#### 4.1.2 Qdisc 收集器接口
```go
type QdiscCollector interface {
    MetricCollector
    GetQdiscType() string
    GetSupportedMetrics() []string
    ValidateQdisc(qdisc *tc.Object) bool
}
```

#### 4.1.3 Class 收集器接口
```go
type ClassCollector interface {
    MetricCollector
    GetClassType() string
    GetSupportedMetrics() []string
    ValidateClass(class *tc.Object) bool
}
```

### 4.2 基础实现层设计

#### 4.2.1 基础收集器
```go
type CollectorBase struct {
    id          string
    name        string
    description string
    enabled     bool
    config      *config.CollectorConfig
    logger      *logrus.Logger
}

func (cb *CollectorBase) Collect(ch chan<- prometheus.Metric) {
    if !cb.enabled {
        return
    }
    // 通用收集逻辑
}
```

#### 4.2.2 Qdisc 基础实现
```go
type QdiscBase struct {
    *CollectorBase
    qdiscType        string
    supportedMetrics []string
    metrics          map[string]*MetricDefinition
}

func (qb *QdiscBase) Collect(ch chan<- prometheus.Metric) {
    // 通用 qdisc 收集逻辑
    nsList, err := tc.GetNetNameSpaceList()
    if err != nil {
        qb.logger.Warnf("Get net namespace list failed: %v", err)
        return
    }
    
    for _, ns := range nsList {
        qb.collectForNamespace(ch, ns)
    }
}
```

### 4.3 配置管理设计

#### 4.3.1 收集器配置
```go
type CollectorConfig struct {
    Enabled           bool                    `yaml:"enabled"`
    CollectionTimeout time.Duration          `yaml:"collection_timeout"`
    RetryCount        int                     `yaml:"retry_count"`
    Metrics           map[string]MetricConfig `yaml:"metrics"`
    Labels            []string                `yaml:"labels"`
}

type MetricConfig struct {
    Name       string   `yaml:"name"`
    Enabled    bool     `yaml:"enabled"`
    Help       string   `yaml:"help"`
    Type       string   `yaml:"type"`
    Labels     []string `yaml:"labels"`
    Buckets    []float64 `yaml:"buckets,omitempty"`
}
```

#### 4.3.2 管理器配置
```go
type ManagerConfig struct {
    PerformanceMonitoring bool                    `yaml:"performance_monitoring"`
    CollectionInterval    time.Duration          `yaml:"collection_interval"`
    StatsRetention        time.Duration          `yaml:"stats_retention"`
    Collectors            map[string]CollectorConfig `yaml:"collectors"`
    Logging               LoggingConfig          `yaml:"logging"`
}
```

### 4.4 工厂模式设计

#### 4.4.1 收集器工厂
```go
type CollectorFactory interface {
    CreateCollector(collectorType string, config interface{}) (MetricCollector, error)
    GetSupportedTypes() []string
}

type QdiscFactory struct {
    configs map[string]*config.QdiscConfig
}

func (qf *QdiscFactory) CreateCollector(qdiscType string) (QdiscCollector, error) {
    switch qdiscType {
    case "codel":
        return qdisc.NewCodelCollector(qf.configs[qdiscType])
    case "cbq":
        return qdisc.NewCbqCollector(qf.configs[qdiscType])
    // ... 其他类型
    default:
        return nil, fmt.Errorf("unsupported qdisc type: %s", qdiscType)
    }
}
```

### 4.5 注册中心设计

#### 4.5.1 收集器注册中心
```go
type CollectorRegistry struct {
    mu         sync.RWMutex
    collectors map[string]MetricCollector
    factories  map[string]CollectorFactory
}

func (cr *CollectorRegistry) Register(collector MetricCollector) error {
    cr.mu.Lock()
    defer cr.mu.Unlock()
    
    if _, exists := cr.collectors[collector.ID()]; exists {
        return fmt.Errorf("collector %s already registered", collector.ID())
    }
    
    cr.collectors[collector.ID()] = collector
    return nil
}

func (cr *CollectorRegistry) GetCollector(id string) (MetricCollector, bool) {
    cr.mu.RLock()
    defer cr.mu.RUnlock()
    collector, exists := cr.collectors[id]
    return collector, exists
}
```

### 4.6 具体收集器实现设计

#### 4.6.1 Codel 收集器示例
```go
type CodelCollector struct {
    *QdiscBase
    metrics map[string]*prometheus.Desc
}

func NewCodelCollector(cfg *config.QdiscConfig) *CodelCollector {
    base := NewQdiscBase("codel", "qdisc_codel", "Codel qdisc metrics", cfg)
    
    c := &CodelCollector{
        QdiscBase: base,
        metrics:   make(map[string]*prometheus.Desc),
    }
    
    c.initializeMetrics(cfg)
    return c
}

func (c *CodelCollector) initializeMetrics(cfg *config.QdiscConfig) {
    labelNames := []string{"namespace", "device", "kind"}
    
    c.metrics["ce_mark"] = prometheus.NewDesc(
        "qdisc_codel_ce_mark",
        "Codel CE mark xstat",
        labelNames, nil,
    )
    
    c.metrics["count"] = prometheus.NewDesc(
        "qdisc_codel_count",
        "Codel count xstat",
        labelNames, nil,
    )
    
    // ... 其他指标定义
}

func (c *CodelCollector) collectForDevice(ch chan<- prometheus.Metric, ns string, device interface{}) {
    qdiscs, err := tc.GetQdiscs(device.Index, ns)
    if err != nil {
        c.logger.Warnf("Get qdiscs in netns %s failed: %v", ns, err)
        return
    }
    
    for _, qdisc := range qdiscs {
        if !c.ValidateQdisc(&qdisc) {
            continue
        }
        
        c.collectQdiscMetrics(ch, ns, device, &qdisc)
    }
}
```

## 5. 迁移策略

### 5.1 阶段1：基础架构搭建
1. 创建新的目录结构
2. 实现核心接口和基类
3. 实现配置管理系统
4. 实现工厂和注册中心

### 5.2 阶段2：逐步迁移
1. 迁移一个 qdisc 类型作为示例（如 Codel）
2. 验证新架构的正确性
3. 逐步迁移其他 qdisc 类型
4. 迁移应用和业务指标

### 5.3 阶段3：测试和优化
1. 完善单元测试
2. 添加集成测试
3. 性能测试和优化
4. 代码审查

### 5.4 阶段4：清理和文档
1. 移除旧代码
2. 更新文档
3. 培训团队成员
4. 建立代码规范

## 6. 预期收益

### 6.1 代码质量提升
- 减少重复代码 80%+
- 统一代码风格和架构
- 提高代码可读性和可维护性

### 6.2 开发效率提升
- 新 qdisc 类型开发时间减少 70%+
- 统一的错误处理和日志记录
- 更好的调试和测试支持

### 6.3 系统稳定性提升
- 更好的错误处理
- 统一的配置管理
- 完整的测试覆盖

### 6.4 扩展性提升
- 插件化架构支持
- 配置驱动的行为控制
- 支持自定义收集器

## 7. 风险评估和缓解

### 7.1 风险识别
- 迁移过程中可能引入 bug
- 性能可能受到影响
- 团队成员需要学习新架构

### 7.2 缓解措施
- 分阶段迁移，逐步验证
- 保持向后兼容性
- 充分的测试覆盖
- 详细的文档和培训

## 8. 实施计划

阶段1：创建新架构
创建接口和基础类
实现工厂和注册中心
创建配置管理
阶段2：迁移现有收集器
逐个迁移 qdisc 收集器
迁移应用和业务收集器
保持向后兼容
阶段3：优化和测试
性能优化
完善测试覆盖
文档更新
阶段4：清理旧代码
移除重复代码
统一命名规范
代码审查

## 9. 总结

本重构方案通过引入统一的接口、基类和工厂模式，将显著提高代码的可维护性、可扩展性和一致性。通过分阶段迁移策略，可以最小化风险并确保系统稳定性。预期将减少 80%+ 的重复代码，提高 70%+ 的开发效率，为项目的长期发展奠定坚实基础。
