# 内存泄漏防护改进

## 问题描述

在原始实现中，`Registry.metrics` 切片只增不减，可能导致内存泄漏风险：

```go
// 原始实现 - 存在内存泄漏风险
func (r *Registry) Register(metrics Metric) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.metrics = append(r.metrics, metrics) // 只增不减
}
```

## 改进方案

### 1. 数据结构优化

将 `[]Metric` 切片改为 `map[string]Metric`，支持快速查找和删除：

```go
type Registry struct {
    metrics map[string]Metric // 使用map替代slice，支持快速查找和删除
    mu      sync.RWMutex
}
```

### 2. 指标生命周期管理

#### 2.1 指标标识符

为 `Metric` 接口添加 `ID()` 方法：

```go
type Metric interface {
    Collect(ch chan<- prometheus.Metric)
    // ID returns a unique identifier for this metric
    // This should be stable across program restarts and unique within the registry
    ID() string
}
```

#### 2.2 指标注销机制

提供多种注销方式：

```go
// 通过ID注销指标
func (r *Registry) Unregister(metricID string)

// 通过指标实例注销
func (r *Registry) UnregisterMetric(metric Metric)

// 清空所有指标
func (r *Registry) Clear()
```

### 3. 内存管理优化

#### 3.1 自动清理

- 支持指标的动态注册和注销
- 防止指标累积导致的内存泄漏
- 提供指标计数和状态监控

#### 3.2 并发安全

- 使用读写锁保护并发访问
- 支持高并发场景下的指标操作

## 使用示例

### 注册指标

```go
// 自动注册（在init()函数中）
func init() {
    exporter.Register(NewQdiscHtb())
}
```

### 注销指标

```go
// 通过ID注销
exporter.Unregister("qdisc_htb")

// 通过实例注销
htbMetric := NewQdiscHtb()
exporter.Register(htbMetric)
// ... 使用指标 ...
exporter.UnregisterMetric(htbMetric)
```

### 清空注册表

```go
// 清空所有指标（谨慎使用）
exporter.defaultReg.Clear()
```

## 性能影响

### 改进前
- 注册：O(1) 追加操作
- 查找：O(n) 线性搜索
- 删除：不支持

### 改进后
- 注册：O(1) map插入
- 查找：O(1) map查找
- 删除：O(1) map删除

## 向后兼容性

- 所有现有的 `exporter.Register()` 调用继续有效
- 新增的 `ID()` 方法有默认实现，不会破坏现有代码
- 保持原有的 `GetMetrics()` 接口不变

## 测试覆盖

创建了完整的测试套件验证：

- 内存泄漏防护功能
- 并发安全访问
- 指标生命周期管理
- 边界情况处理

## 最佳实践

1. **指标命名**：使用有意义的指标ID，便于管理和调试
2. **生命周期管理**：及时注销不再使用的指标
3. **监控指标数量**：定期检查注册表中的指标数量
4. **错误处理**：处理注销不存在指标的情况

## 未来改进方向

1. **指标过期机制**：自动清理长时间未使用的指标
2. **指标依赖管理**：支持指标间的依赖关系
3. **指标版本控制**：支持指标版本升级和兼容性
4. **性能指标**：添加注册表操作的性能监控
