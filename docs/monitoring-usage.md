# 应用监控指标使用指南

## 概述

本文档描述了如何在 `uos-tc-exporter` 中使用新添加的应用监控指标系统。该系统提供了全面的应用性能监控、业务指标监控和性能分析功能。

## 监控指标架构

### 1. 应用级别指标 (AppMetrics)
- **运行时指标**: Go运行时统计信息
- **性能指标**: 指标收集性能统计
- **系统指标**: 系统运行时间和进程信息
- **自定义指标**: 错误率和自定义统计

### 2. 业务指标 (BusinessMetrics)
- **TC相关指标**: 网络命名空间、接口、qdisc、class数量
- **网络性能指标**: 延迟、吞吐量、错误统计
- **服务状态指标**: 服务健康状态监控
- **配置指标**: 配置重载和版本信息

### 3. 性能监控包装器 (PerformanceWrapper)
- 自动为现有指标收集器添加性能监控
- 记录收集耗时和错误统计
- 非侵入式设计，不影响原有功能

### 4. 指标管理器 (Manager)
- 统一管理所有监控指标
- 提供统计信息收集和分析
- 支持配置化控制

## 使用方法

### 基本使用

```go
package main

import (
    "gitee.com/openeuler/uos-tc-exporter/internal/metrics"
    "time"
)

func main() {
    // 获取全局管理器
    manager := metrics.GetGlobalManager()
    
    // 记录指标收集
    startTime := time.Now()
    success := true
    var err error
    
    // 执行指标收集逻辑
    // ... 你的指标收集代码 ...
    
    // 记录收集结果
    duration := time.Since(startTime)
    manager.RecordCollection(duration, success, err)
}
```

### 业务指标更新

```go
// 更新TC统计信息
manager.UpdateTCStats(5, 20, 100, 50) // 5个命名空间，20个接口，100个qdisc，50个class

// 更新网络吞吐量
manager.UpdateNetworkThroughput(1024.5) // 1024.5 bytes/s

// 记录网络延迟
manager.RecordNetworkLatency(0.001) // 1ms

// 记录网络错误
manager.IncrementNetworkErrors()

// 设置服务健康状态
manager.SetServiceHealth("tc_exporter", "metrics_collector", true)
```

### 配置重载监控

```go
// 配置重载成功
manager.IncrementConfigReload()

// 配置重载失败
manager.IncrementConfigReloadErrors()

// 更新配置版本信息
manager.UpdateConfigVersion("tc-exporter.yaml", "1.1.0", "abc123")
```

## 配置选项

### 监控配置

```yaml
monitoring:
  enabled: true                           # 启用监控
  performance_monitoring: true            # 启用性能监控
  enable_business_metrics: true           # 启用业务指标
  collection_interval: "30s"              # 收集间隔
  stats_retention: "24h"                  # 统计保留时间
  app_info:
    version: "1.0.0"                      # 应用版本
    build_time: "2025-01-01"              # 构建时间
    go_version: "1.25.0"                  # Go版本
```

### 指标收集优化配置

```yaml
metrics:
  log_interval: 5                         # 日志输出间隔
  debug_logging: false                    # 调试日志
  performance_stats: true                 # 性能统计
```

## 监控指标说明

### 应用指标

| 指标名称 | 类型 | 说明 |
|---------|------|------|
| `app_info` | Gauge | 应用信息（版本、构建时间、Go版本） |
| `go_goroutines` | Gauge | 当前goroutine数量 |
| `go_threads` | Gauge | OS线程数量 |
| `go_heap_alloc_bytes` | Gauge | 堆内存分配字节数 |
| `go_heap_sys_bytes` | Gauge | 从系统获取的堆内存字节数 |
| `metrics_collection_duration_seconds` | Histogram | 指标收集耗时分布 |
| `metrics_collection_total` | Counter | 指标收集总次数 |
| `metrics_collection_errors_total` | Counter | 指标收集错误次数 |
| `system_uptime_seconds` | Gauge | 系统运行时间 |
| `process_start_time_seconds` | Gauge | 进程启动时间 |
| `custom_metrics_count` | Gauge | 自定义指标数量 |
| `error_rate` | Gauge | 错误率（0-1） |

### 业务指标

| 指标名称 | 类型 | 说明 |
|---------|------|------|
| `tc_namespaces_total` | Gauge | TC命名空间总数 |
| `tc_interfaces_total` | Gauge | TC接口总数 |
| `tc_qdiscs_total` | Gauge | TC qdisc总数 |
| `tc_classes_total` | Gauge | TC class总数 |
| `network_latency_seconds` | Histogram | 网络操作延迟分布 |
| `network_throughput_bytes_per_second` | Gauge | 网络吞吐量 |
| `network_errors_total` | Counter | 网络错误总数 |
| `service_health` | Gauge | 服务健康状态 |
| `last_update_timestamp` | Gauge | 最后更新时间戳 |
| `config_reload_total` | Counter | 配置重载总次数 |
| `config_reload_errors_total` | Counter | 配置重载错误次数 |
| `config_version` | Gauge | 配置版本信息 |

## 性能监控

### 自动性能监控

系统会自动为所有指标收集器添加性能监控：

```go
// 原始指标收集器
originalMetric := &MyMetrics{}

// 使用性能监控包装器
wrappedMetric := metrics.WrapWithPerformanceMonitoring(originalMetric)

// 注册包装后的指标
exporter.Register(wrappedMetric)
```

### 性能统计

- **收集耗时**: 每次指标收集的耗时统计
- **成功率**: 成功收集的指标比例
- **错误统计**: 收集过程中的错误记录
- **性能趋势**: 长期性能变化趋势

## 最佳实践

### 1. 指标命名规范
- 使用下划线分隔的小写字母
- 添加适当的单位后缀（如 `_seconds`, `_bytes`）
- 使用描述性的名称

### 2. 标签使用
- 避免高基数标签
- 使用有意义的标签值
- 保持标签数量合理

### 3. 指标收集频率
- 根据业务需求设置合适的收集频率
- 避免过于频繁的收集影响性能
- 监控指标收集本身的性能

### 4. 错误处理
- 记录所有错误和异常情况
- 提供错误分类和统计
- 设置合理的错误阈值告警

## 故障排除

### 常见问题

1. **指标不显示**
   - 检查监控是否启用
   - 验证指标注册是否成功
   - 查看日志中的错误信息

2. **性能影响**
   - 调整收集频率
   - 检查指标数量是否过多
   - 优化指标收集逻辑

3. **内存泄漏**
   - 检查指标标签是否过多
   - 验证指标清理逻辑
   - 监控内存使用情况

### 调试方法

1. **启用调试日志**
   ```yaml
   log:
     level: "debug"
   ```

2. **查看指标统计**
   ```go
   stats := manager.GetStats()
   logrus.Infof("Collection stats: %+v", stats)
   ```

3. **监控指标数量**
   ```go
   count := registry.GetMetricCount()
   logrus.Infof("Total metrics: %d", count)
   ```

## 扩展开发

### 添加新指标

1. 创建新的指标结构体
2. 实现 `Metric` 接口
3. 在 `init()` 函数中注册
4. 添加相应的配置选项

### 自定义监控逻辑

1. 继承 `Manager` 结构体
2. 重写相关方法
3. 添加自定义统计逻辑
4. 集成到现有系统

## 总结

新的监控指标系统为 `uos-tc-exporter` 提供了全面的监控能力，包括：

- **应用性能监控**: 运行时统计、性能分析
- **业务指标监控**: TC相关统计、网络性能
- **性能监控包装**: 自动性能分析、错误统计
- **统一管理**: 集中配置、统计管理

该系统采用非侵入式设计，对现有代码影响最小，同时提供了丰富的监控功能和灵活的配置选项。
