# 优雅关闭功能

## 概述

uos-tc-exporter 提供了可配置的优雅关闭功能，确保在服务停止时能够安全地清理资源并完成正在进行的请求。

## 配置选项

### 服务器配置

在配置文件中添加 `server` 部分来配置优雅关闭行为：

```yaml
server:
  # 优雅关闭超时时间
  # 支持的时间单位：ns, us (or µs), ms, s, m, h
  # 示例：30s, 1m, 2m30s, 1h30m
  shutdownTimeout: "30s"
```

### 默认值

如果不配置 `shutdownTimeout`，系统将使用默认值：**30秒**

## 工作原理

### 1. HTTP服务器关闭

- 停止接受新的HTTP连接
- 等待现有连接完成（最长等待时间：`shutdownTimeout`）
- 如果超时，记录警告日志

### 2. 配置监控关闭

- 停止配置文件监控
- 清理文件系统监听器

### 3. 资源清理

- 使用 `sync.WaitGroup` 协调各个组件的关闭
- 并发关闭各个组件以提高效率
- 收集关闭过程中的错误信息

## 配置示例

### 快速关闭（适合开发环境）

```yaml
server:
  shutdownTimeout: "5s"
```

### 标准关闭（适合生产环境）

```yaml
server:
  shutdownTimeout: "30s"
```

### 长时间关闭（适合高负载环境）

```yaml
server:
  shutdownTimeout: "2m"
```

## 日志信息

系统会在关闭过程中记录详细的日志信息：

```
INFO: Server shutdown timeout set to: 30s
INFO: HTTP server shutdown timeout set to: 30s
INFO: Config watching stopped
INFO: HTTP server gracefully stopped
INFO: All server components stopped successfully
INFO: Server stopped gracefully
```

## 超时处理

如果某个组件在指定时间内无法完成关闭：

1. 记录超时警告日志
2. 继续等待其他组件关闭
3. 在日志中标记关闭状态

## 最佳实践

1. **生产环境**：建议设置 30-60 秒的超时时间
2. **高负载环境**：可能需要更长的超时时间（1-2分钟）
3. **开发环境**：可以使用较短的超时时间（5-10秒）
4. **监控**：关注关闭超时的日志，调整超时时间

## 故障排除

### 常见问题

1. **关闭超时**：增加 `shutdownTimeout` 值
2. **资源泄漏**：检查组件是否正确实现了关闭逻辑
3. **性能问题**：优化关闭逻辑，减少不必要的等待时间

### 调试建议

1. 启用 debug 级别日志
2. 监控关闭过程的耗时
3. 检查是否有长时间运行的goroutine
