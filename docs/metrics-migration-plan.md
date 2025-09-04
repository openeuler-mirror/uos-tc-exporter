# Metrics 架构迁移计划

## 1. 迁移概述

### 1.1 迁移目标
- 将现有的分散式 metrics 代码重构为统一的、可维护的架构
- 保持向后兼容性，确保现有功能不受影响
- 提高代码质量和可扩展性

### 1.2 迁移策略
- 分阶段迁移，逐步替换现有实现
- 保持接口兼容性，确保平滑过渡
- 充分测试，确保功能正确性

## 2. 当前代码分析

### 2.1 现有文件清单
```
internal/metrics/
├── app_metrics.go          # 应用指标 (290行)
├── business_metrics.go     # 业务指标 (270行)
├── manager.go             # 管理器 (279行)
├── metrics.go             # 基础指标 (36行)
├── q_cbq.go               # CBQ qdisc (187行)
├── q_cbq_test.go          # CBQ 测试 (345行)
├── q_choke.go             # Choke qdisc (217行)
├── q_choke_test.go        # Choke 测试 (538行)
├── q_codel.go             # Codel qdisc (308行)
├── q_codel_test.go        # Codel 测试 (815行)
├── q_fq.go                # FQ qdisc (549行)
├── q_fq_codel.go          # FQ-Codel qdisc (226行)
├── q_hfsc.go              # HFSC qdisc (157行)
├── q_htb.go               # HTB qdisc (190行)
├── q_pie.go               # PIE qdisc (316行)
├── q_red.go               # RED qdisc (193行)
├── q_sfb.go               # SFB qdisc (252行)
├── q_sfq.go               # SFQ qdisc (100行)
├── qclass.go              # Class 指标 (334行)
├── qdisc.go               # Qdisc 基础 (333行)
├── logger.go              # 日志 (124行)
├── logger_test.go         # 日志测试 (257行)
├── performance_wrapper.go # 性能包装器 (91行)
├── metrics_test.go        # 指标测试 (327行)
└── info.go                # 信息 (43行)
```

### 2.2 代码重复分析

#### 2.2.1 重复的 Collect 方法
每个 qdisc 文件都有相似的 Collect 方法实现：
- 获取网络命名空间列表
- 遍历命名空间和设备
- 获取 qdiscs 并过滤类型
- 收集指标数据

**重复代码统计：**
- 11个 qdisc 文件 × 平均 50行重复代码 = 550行重复代码
- 占总代码量的约 15%

#### 2.2.2 重复的指标定义
每个 qdisc 都有相似的指标定义模式：
- 创建 prometheus.Desc
- 定义标签名称
- 实现 Collect 方法

**重复代码统计：**
- 每个 qdisc 平均 8个指标 × 11个 qdisc = 88个指标定义
- 每个指标定义平均 10行代码 = 880行重复代码

### 2.3 问题总结
- **总重复代码：** 约 1430行 (占总代码量的 40%)
- **维护成本：** 添加新 qdisc 需要复制大量代码
- **测试覆盖：** 部分文件缺少测试，测试风格不统一
- **配置管理：** 硬编码配置，难以动态调整

## 3. 新架构设计

### 3.1 目录结构
```
internal/metrics/
├── interfaces/           # 接口定义层
│   ├── collector.go     # 基础收集器接口
│   ├── qdisc.go         # qdisc 收集器接口
│   ├── class.go         # class 收集器接口
│   └── config.go        # 配置接口
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
├── manager_v2.go        # 新管理器
└── legacy/              # 旧代码（迁移完成后删除）
    └── ...
```

### 3.2 代码减少预期
- **重复代码减少：** 1430行 → 200行 (减少 86%)
- **新增代码：** 约 1500行 (接口、基类、工厂等)
- **净减少：** 约 130行代码
- **维护成本：** 降低 70%+

## 4. 迁移计划

### 4.1 阶段1：基础架构搭建 (第1-2周)

#### 4.1.1 第1周：接口和基类
- [ ] 创建 `interfaces/` 目录和接口定义
- [ ] 创建 `base/` 目录和基础实现
- [ ] 实现 `CollectorBase` 和 `QdiscBase`
- [ ] 编写基础类的单元测试

#### 4.1.2 第2周：配置和工厂
- [ ] 创建 `config/` 目录和配置管理
- [ ] 创建 `factories/` 目录和工厂实现
- [ ] 创建 `registry/` 目录和注册中心
- [ ] 编写配置和工厂的单元测试

### 4.2 阶段2：收集器迁移 (第3-6周)

#### 4.2.1 第3周：Codel 收集器迁移
- [ ] 创建 `collectors/qdisc/codel.go`
- [ ] 迁移 Codel 收集器逻辑
- [ ] 编写 Codel 收集器测试
- [ ] 验证功能正确性

#### 4.2.2 第4周：CBQ 和 HTB 收集器迁移
- [ ] 迁移 CBQ 收集器
- [ ] 迁移 HTB 收集器
- [ ] 编写测试用例
- [ ] 性能测试

#### 4.2.3 第5周：FQ 系列收集器迁移
- [ ] 迁移 FQ 收集器
- [ ] 迁移 FQ-Codel 收集器
- [ ] 迁移其他 qdisc 收集器
- [ ] 集成测试

#### 4.2.4 第6周：Class 和应用收集器迁移
- [ ] 迁移 Class 收集器
- [ ] 迁移应用指标收集器
- [ ] 迁移业务指标收集器
- [ ] 完整集成测试

### 4.3 阶段3：新管理器实现 (第7-8周)

#### 4.3.1 第7周：管理器 V2
- [ ] 实现 `ManagerV2`
- [ ] 集成所有收集器
- [ ] 实现配置管理
- [ ] 编写管理器测试

#### 4.3.2 第8周：兼容性层
- [ ] 实现向后兼容接口
- [ ] 创建迁移工具
- [ ] 更新文档
- [ ] 性能优化

### 4.4 阶段4：测试和优化 (第9-10周)

#### 4.4.1 第9周：测试完善
- [ ] 完善单元测试
- [ ] 添加集成测试
- [ ] 性能基准测试
- [ ] 压力测试

#### 4.4.2 第10周：代码审查和优化
- [ ] 代码审查
- [ ] 性能优化
- [ ] 文档更新
- [ ] 最终测试

### 4.5 阶段5：部署和清理 (第11-12周)

#### 4.5.1 第11周：部署准备
- [ ] 准备部署脚本
- [ ] 配置迁移工具
- [ ] 用户培训
- [ ] 部署测试

#### 4.5.2 第12周：清理和交付
- [ ] 移除旧代码
- [ ] 更新文档
- [ ] 项目交付
- [ ] 经验总结

## 5. 详细迁移步骤

### 5.1 第一步：创建基础架构

#### 5.1.1 创建目录结构
```bash
mkdir -p internal/metrics/{interfaces,base,collectors/{qdisc,class,app,business},factories,config,registry,utils}
```

#### 5.1.2 实现核心接口
```go
// interfaces/collector.go
type MetricCollector interface {
    Collect(ch chan<- prometheus.Metric)
    ID() string
    Name() string
    Description() string
    Enabled() bool
    SetEnabled(enabled bool)
}
```

#### 5.1.3 实现基础类
```go
// base/collector_base.go
type CollectorBase struct {
    // 基础实现
}

// base/qdisc_base.go
type QdiscBase struct {
    *CollectorBase
    // qdisc 特定实现
}
```

### 5.2 第二步：迁移第一个收集器 (Codel)

#### 5.2.1 创建 Codel 收集器
```go
// collectors/qdisc/codel.go
type CodelCollector struct {
    *QdiscBase
    metrics map[string]*prometheus.Desc
}

func NewCodelCollector(cfg *config.CollectorConfig) *CodelCollector {
    // 实现
}
```

#### 5.2.2 迁移指标定义
```go
func (c *CodelCollector) initializeMetrics(cfg *config.CollectorConfig) {
    c.metrics["ce_mark"] = prometheus.NewDesc(
        "qdisc_codel_ce_mark",
        "Codel CE mark xstat",
        []string{"namespace", "device", "kind"}, nil,
    )
    // ... 其他指标
}
```

#### 5.2.3 迁移收集逻辑
```go
func (c *CodelCollector) collectQdiscMetrics(ch chan<- prometheus.Metric, ns, deviceName string, qdisc interface{}) {
    // 从 q_codel.go 迁移的收集逻辑
}
```

### 5.3 第三步：实现工厂和注册中心

#### 5.3.1 实现 Qdisc 工厂
```go
// factories/qdisc_factory.go
type QdiscFactory struct {
    configs map[string]*config.CollectorConfig
}

func (qf *QdiscFactory) CreateCollector(qdiscType string) (interfaces.QdiscCollector, error) {
    switch qdiscType {
    case "codel":
        return qdisc.NewCodelCollector(qf.configs[qdiscType])
    // ... 其他类型
    }
}
```

#### 5.3.2 实现注册中心
```go
// registry/collector_registry.go
type CollectorRegistry struct {
    collectors map[string]interfaces.MetricCollector
    factories  map[string]CollectorFactory
}
```

### 5.4 第四步：实现新管理器

#### 5.4.1 创建 ManagerV2
```go
// manager_v2.go
type ManagerV2 struct {
    registry *registry.CollectorRegistry
    factories map[string]registry.CollectorFactory
    config   *config.ManagerConfig
}

func (m *ManagerV2) CollectAll(ch chan<- prometheus.Metric) {
    collectors := m.registry.GetEnabledCollectors()
    for _, collector := range collectors {
        collector.Collect(ch)
    }
}
```

### 5.5 第五步：逐步迁移其他收集器

#### 5.5.1 迁移模式
1. 复制现有 qdisc 文件到 `collectors/qdisc/`
2. 重构为新的架构模式
3. 编写测试用例
4. 验证功能正确性
5. 更新工厂注册

#### 5.5.2 迁移顺序
1. **Codel** (已完成)
2. **CBQ** - 简单结构，易于迁移
3. **HTB** - 中等复杂度
4. **FQ** - 复杂结构，多个指标
5. **FQ-Codel** - 基于 FQ 和 Codel
6. **其他 qdisc** - 按复杂度排序

## 6. 测试策略

### 6.1 单元测试
- 每个收集器都有对应的测试文件
- 测试覆盖率目标：90%+
- 使用表驱动测试模式

### 6.2 集成测试
- 测试收集器注册和发现
- 测试配置管理
- 测试指标收集流程

### 6.3 性能测试
- 收集性能基准测试
- 内存使用测试
- 并发安全测试

### 6.4 兼容性测试
- 向后兼容性测试
- 配置迁移测试
- 升级路径测试

## 7. 风险控制

### 7.1 技术风险
- **风险：** 新架构可能引入 bug
- **缓解：** 充分测试，分阶段迁移

### 7.2 兼容性风险
- **风险：** 破坏现有功能
- **缓解：** 保持接口兼容，实现兼容层

### 7.3 性能风险
- **风险：** 性能可能下降
- **缓解：** 性能基准测试，优化热点代码

### 7.4 时间风险
- **风险：** 迁移时间可能超期
- **缓解：** 合理规划，预留缓冲时间

## 8. 成功标准

### 8.1 功能标准
- [ ] 所有现有功能正常工作
- [ ] 新架构功能完整
- [ ] 配置管理正常
- [ ] 指标收集正确

### 8.2 质量标准
- [ ] 代码重复率 < 10%
- [ ] 测试覆盖率 > 90%
- [ ] 代码审查通过
- [ ] 性能无显著下降

### 8.3 维护标准
- [ ] 添加新 qdisc 时间 < 2小时
- [ ] 修改现有功能影响范围 < 3个文件
- [ ] 配置修改无需代码变更
- [ ] 文档完整准确

## 9. 交付物

### 9.1 代码交付
- 新架构代码
- 迁移工具
- 测试用例
- 配置文件

### 9.2 文档交付
- 架构设计文档
- 实现指南
- 迁移计划
- 用户手册

### 9.3 工具交付
- 配置迁移工具
- 代码生成工具
- 测试工具
- 部署脚本

## 10. 总结

这个迁移计划通过分阶段、渐进式的方式，将现有的分散式 metrics 代码重构为统一的、可维护的架构。预期将减少 86% 的重复代码，提高 70% 的开发效率，为项目的长期发展奠定坚实基础。

通过充分的测试和风险控制措施，确保迁移过程的安全性和可靠性，最终交付一个高质量、易维护的 metrics 系统。
