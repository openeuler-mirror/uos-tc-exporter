# TC Exporter 设计文档

## 项目概述

TC Exporter 是一个用于监控 Linux 流量控制（Traffic Control，TC）系统的 Prometheus 导出器。它通过 netlink 接口与内核 TC 子系统通信，收集各种队列规则（qdisc）和类（class）的统计信息，并以 Prometheus 格式暴露指标。

## 架构设计

### 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Prometheus    │    │   TC Exporter   │    │   Linux Kernel  │
│                 │◄──►│                 │◄──►│                 │
│   - 拉取指标     │    │   - HTTP Server │    │   - TC 子系统    │
│   - 存储数据     │    │   - 指标收集器   │    │   - Netlink     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   配置文件       │
                       │   - YAML 格式    │
                       │   - 日志配置     │
                       │   - 服务配置     │
                       └─────────────────┘
```

### 核心组件

#### 1. HTTP 服务器 (`internal/server/`)
- **职责**: 提供 Prometheus 指标端点和管理界面
- **功能**:
  - 监听配置的端口（默认 9062）
  - 处理 `/metrics` 请求
  - 提供健康检查端点
  - 支持限流保护

#### 2. 指标收集器 (`internal/metrics/`)
- **职责**: 收集和格式化 TC 指标
- **支持的队列规则**:
  - CBQ (Class Based Queueing)
  - CHOKE (CHOose and Keep for responsive flows, CHOose and Kill for unresponsive flows)
  - CODEL (Controlled Delay)
  - 其他常见 qdisc 类型

#### 3. TC 客户端 (`internal/tc/`)
- **职责**: 通过 netlink 与内核 TC 子系统通信
- **功能**:
  - 获取网络接口列表
  - 查询 qdisc 配置和统计
  - 查询 class 配置和统计
  - 支持网络命名空间

#### 4. 配置管理 (`internal/exporter/`)
- **职责**: 管理应用配置
- **功能**:
  - 解析命令行参数
  - 加载 YAML 配置文件
  - 配置验证和默认值设置

#### 5. 日志系统 (`pkg/logging/`)
- **职责**: 提供结构化日志输出
- **功能**:
  - 多级别日志支持
  - 文件轮转
  - 结构化日志格式

## 数据流程

### 启动流程
```
main.go → exporter.go → server.NewServer()
│
├── 初始化日志系统 (pkg/logging)
├── 解析配置文件 (config/)
├── 创建指标收集器 (internal/metrics)
├── 注册收集器 (internal/exporter)
└── 启动HTTP服务器 (internal/server)
```

### 指标收集流程
```
HTTP Request (/metrics)
│
├── 限流检查 (pkg/ratelimit)
├── 指标收集
│   ├── TC 数据收集 (internal/tc)
│   │   └── Netlink 通信
│   ├── 系统信息收集
│   └── Qdisc/Class 指标收集
│
└── Prometheus 格式响应
```

## 技术实现

### 核心依赖
- **Go 1.20+**: 编程语言
- **github.com/florianl/go-tc**: TC netlink 库
- **github.com/prometheus/client_golang**: Prometheus 客户端
- **github.com/sirupsen/logrus**: 结构化日志
- **github.com/alecthomas/kingpin**: 命令行参数解析

### 配置示例

```yaml
# tc-exporter.yaml
log:
  level: info
  logPath: /var/log/tc-exporter.log
  maxSize: 10MB
  maxAge: 168h  # 7天

address: 127.0.0.1
port: 9062
metricsPath: /metrics

server:
  shutdownTimeout: 30s
```

### 指标示例

```prometheus
# HELP tc_qdisc_packets_total Total packets processed by qdisc
# TYPE tc_qdisc_packets_total counter
tc_qdisc_packets_total{device="eth0",qdisc="cbq"} 123456

# HELP tc_qdisc_bytes_total Total bytes processed by qdisc
# TYPE tc_qdisc_bytes_total counter
tc_qdisc_bytes_total{device="eth0",qdisc="cbq"} 789012345

# HELP tc_qdisc_drops_total Total packets dropped by qdisc
# TYPE tc_qdisc_drops_total counter
tc_qdisc_drops_total{device="eth0",qdisc="cbq"} 123
```

## 部署方案

### Systemd 服务部署
```ini
# /usr/lib/systemd/system/uos-tc-exporter.service
[Unit]
Description=UOS TC Exporter
After=network.target

[Service]
Type=simple
User=tc-exporter
Group=tc-exporter
ExecStart=/usr/bin/uos-tc-exporter
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

### Docker 容器部署
```dockerfile
FROM golang:1.20 AS builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
COPY --from=builder /app/bin/uos-tc-exporter /usr/local/bin/
EXPOSE 9062
ENTRYPOINT ["/usr/local/bin/uos-tc-exporter"]
```

## 监控和运维

### 关键指标
- 指标收集耗时
- 错误计数
- 请求频率
- 系统资源使用情况

### 日志配置
- 日志级别可配置（debug, info, warn, error）
- 支持文件轮转
- 结构化日志格式便于解析

## 扩展性设计

### 插件化架构
- 模块化收集器设计
- 统一指标接口
- 动态注册机制

### 配置灵活性
- YAML 配置文件支持
- 命令行参数覆盖
- 环境变量支持

## 安全考虑

### 权限要求
- 需要 NET_ADMIN 权限访问 TC 信息
- 建议使用非特权用户运行

### 网络安全
- 默认监听本地地址
- 支持配置监听地址和端口
- 限流保护防止过度请求

## 故障排除

### 常见问题
1. **权限不足**: 确保运行用户有足够权限
2. **网络接口不存在**: 检查配置的网络接口名称
3. **TC 子系统未启用**: 确认系统支持 TC 功能

### 日志分析
- 查看日志文件获取详细错误信息
- 启用 debug 级别日志进行调试

## 未来发展

### 计划功能
- 支持更多 qdisc 类型
- 增强配置验证
- 改进性能监控
- 容器化优化

### 社区贡献
- 欢迎提交 issue 和 pull request
- 文档改进和翻译
- 测试用例补充

---

*最后更新: 2025年10月30日*
*版本: 1.0.0*
