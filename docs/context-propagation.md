# 上下文贯穿与优雅关闭实现

本文档详细说明了 `uos-tc-exporter` 中上下文（Context）贯穿实现和优雅关闭机制，确保所有 goroutine 能够正确响应取消信号并优雅退出。

## 架构概览

### 上下文传播路径

```
main() -> Run() -> Server.SetUp() -> Server.Run() -> HttpServer.RunWithContext()
                -> ConfigManager.StartWatching()
                -> MetricsManager.CollectAllWithContext()
                -> TcObjectCollector.collectObjectsWithContext()
```

### 关键组件

1. **主入口层** (`main.go`, `exporter.go`)
2. **服务器层** (`internal/server/server.go`)
3. **HTTP 层** (`internal/server/http_server.go`)
4. **指标收集层** (`internal/metrics/manager_v2.go`)
5. **TC 操作层** (`internal/tc/operations.go`)

## 实现细节

### 1. 主入口层 - 信号处理与 errgroup 管理

**文件**: `exporter.go`

```go
func Run(name string, version string) error {
    // 创建根上下文，支持信号取消
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 监听系统信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        sig := <-sigChan
        logrus.Infof("Received signal %v, initiating graceful shutdown", sig)
        cancel()
    }()

    // 使用 errgroup 管理 goroutine 生命周期
    g, gCtx := errgroup.WithContext(ctx)
    
    // 启动服务器运行
    g.Go(func() error {
        return s.Run(gCtx)
    })

    // 等待上下文取消或服务器错误
    <-gCtx.Done()
    
    // 优雅关闭
    s.Stop()
    
    // 等待所有 goroutine 完成
    if err := g.Wait(); err != nil {
        logrus.Errorf("Server exited with error: %v", err)
        return err
    }
    
    return nil
}
```

**关键特性**:
- 使用 `context.WithCancel` 创建可取消的根上下文
- 监听 `SIGINT` 和 `SIGTERM` 信号，自动触发优雅关闭
- 使用 `errgroup.WithContext` 管理所有 goroutine 生命周期
- 确保所有 goroutine 在退出前完成清理工作

### 2. 服务器层 - 上下文传递与配置监控

**文件**: `internal/server/server.go`

```go
func (s *Server) SetUp(ctx context.Context) error {
    // ... 初始化逻辑 ...
    
    // 启动配置监控，传递上下文
    if err := s.configMgr.StartWatching(ctx); err != nil {
        logrus.Warnf("Failed to start config watching: %v, config hot reload will be disabled", err)
    } else {
        logrus.Info("Config hot reload enabled")
    }
    
    return nil
}

func (s *Server) Run(ctx context.Context) error {
    logrus.Infof("Running %s", s.Name)
    
    // 使用上下文运行 HTTP 服务器
    return s.httpServer.RunWithContext(ctx)
}

func (s *Server) StopWithContext(ctx context.Context) {
    // 创建带超时的关闭上下文
    shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
    defer cancel()
    
    // 协调关闭各个组件
    // ...
}
```

**关键特性**:
- 所有方法都接受 `context.Context` 参数
- 配置监控使用传入的上下文，支持取消
- 优雅关闭支持超时控制
- 使用 `WaitGroup` 协调多个组件的关闭

### 3. HTTP 层 - 上下文感知的服务器运行

**文件**: `internal/server/http_server.go`

```go
func (hs *HttpServer) RunWithContext(ctx context.Context) error {
    if hs.server == nil {
        return fmt.Errorf("HTTP server not initialized")
    }

    logrus.Infof("Running HTTP server on %s", hs.server.Addr)
    
    // 在单独的 goroutine 中启动服务器
    errChan := make(chan error, 1)
    go func() {
        if err := hs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            // 错误处理...
            errChan <- customErr
        }
    }()

    // 等待上下文取消或服务器错误
    select {
    case <-ctx.Done():
        logrus.Info("Context cancelled, shutting down HTTP server")
        return ctx.Err()
    case err := <-errChan:
        return err
    }
}
```

**关键特性**:
- HTTP 服务器在独立 goroutine 中运行
- 使用 `select` 监听上下文取消和服务器错误
- 上下文取消时立即返回，触发优雅关闭

### 4. 指标收集层 - 上下文感知的指标收集

**文件**: `internal/metrics/manager_v2.go`

```go
func (m *ManagerV2) CollectAllWithContext(ctx context.Context, ch chan<- prometheus.Metric) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        fmt.Printf("Collection took %v\n", duration)
    }()
    
    // 检查上下文是否已取消
    select {
    case <-ctx.Done():
        m.logger.Warnf("Collection cancelled due to context: %v", ctx.Err())
        return
    default:
    }
    
    collectors := m.registry.GetEnableCollectors()
    for _, collector := range collectors {
        // 检查上下文是否已取消
        select {
        case <-ctx.Done():
            m.logger.Warnf("Collection cancelled during collector %s: %v", collector.ID(), ctx.Err())
            return
        default:
        }
        
        m.logger.Debugf("Collecting from collector: %s", collector.ID())
        collector.Collect(ch)
    }
}
```

**关键特性**:
- 在收集开始前检查上下文状态
- 在每次收集器处理前检查上下文状态
- 上下文取消时立即停止收集，避免资源浪费

### 5. TC 操作层 - 上下文感知的网络操作

**文件**: `internal/tc/operations.go`

```go
func (tcoc *TcObjectCollector) collectObjectsWithContext(
    ctx context.Context,
    devID uint32,
    collectFunc func(*tc.Tc) ([]tc.Object, error),
) ([]tc.Object, error) {
    // 检查上下文是否已取消
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // 获取 TC 连接
    sock, err := tcoc.connManager.GetTcConn()
    if err != nil {
        return nil, err
    }
    defer sock.Close()

    // 使用带超时的上下文进行收集
    done := make(chan struct{})
    var objects []tc.Object
    var collectErr error

    go func() {
        defer close(done)
        objects, collectErr = collectFunc(sock)
    }()

    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    case <-done:
        if collectErr != nil {
            return nil, collectErr
        }
    }

    // 过滤和返回结果...
}
```

**关键特性**:
- 在操作开始前检查上下文状态
- 使用 goroutine 执行可能阻塞的网络操作
- 使用 `select` 监听上下文取消和操作完成
- 确保网络连接正确关闭

## 使用方式

### 1. 启动服务

```bash
# 正常启动
./uos-tc-exporter

# 使用自定义配置
./uos-tc-exporter -c /path/to/config.yaml
```

### 2. 优雅关闭

```bash
# 发送 SIGTERM 信号
kill -TERM <pid>

# 发送 SIGINT 信号 (Ctrl+C)
kill -INT <pid>
```

### 3. 配置超时

在配置文件中设置关闭超时时间：

```yaml
server:
  shutdownTimeout: 30s  # 默认30秒
```

## 错误处理

### 1. 上下文取消错误

当上下文被取消时，各个组件会：
- 记录相应的警告日志
- 停止当前操作
- 返回 `context.Canceled` 或 `context.DeadlineExceeded` 错误

### 2. 超时处理

- HTTP 服务器有内置的超时设置
- 优雅关闭有配置的超时时间
- TC 操作支持上下文超时

### 3. 资源清理

- 所有网络连接都会在操作完成后关闭
- 配置监控会在关闭时停止
- HTTP 服务器会等待现有请求完成

## 监控与调试

### 1. 日志级别

```bash
# 启用调试日志查看上下文传播
export LOG_LEVEL=debug
./uos-tc-exporter
```

### 2. 关键日志

- `"Received signal %v, initiating graceful shutdown"` - 信号接收
- `"Context cancelled, shutting down HTTP server"` - HTTP 服务器关闭
- `"Collection cancelled due to context"` - 指标收集取消
- `"Server stopped gracefully"` - 服务器成功关闭

### 3. 健康检查

服务提供以下健康检查端点：
- `/health` - 整体健康状态
- `/ready` - 服务就绪状态
- `/live` - 服务存活状态

## 最佳实践

### 1. 上下文传递

- 所有长时间运行的操作都应该接受 `context.Context` 参数
- 在操作开始前检查上下文状态
- 在循环中定期检查上下文状态

### 2. 错误处理

- 区分上下文取消错误和其他错误
- 上下文取消通常不是真正的错误，应该优雅处理
- 记录有意义的错误信息

### 3. 资源管理

- 使用 `defer` 确保资源清理
- 在 goroutine 中也要处理资源清理
- 避免资源泄漏

### 4. 测试

- 测试上下文取消场景
- 测试超时场景
- 测试信号处理

## 故障排除

### 1. 服务无法正常关闭

- 检查是否有 goroutine 没有正确响应上下文取消
- 检查是否有阻塞的网络操作
- 增加关闭超时时间

### 2. 资源泄漏

- 检查所有网络连接是否正确关闭
- 检查是否有未清理的 goroutine
- 使用 `go tool pprof` 分析内存使用

### 3. 性能问题

- 检查上下文检查的频率
- 避免在热路径中进行昂贵的上下文检查
- 考虑使用 `context.WithTimeout` 限制操作时间

---

本文档描述了 `uos-tc-exporter` 中上下文贯穿和优雅关闭的完整实现。通过这种设计，服务能够正确响应取消信号，确保资源得到正确清理，提供更好的可靠性和可维护性。
