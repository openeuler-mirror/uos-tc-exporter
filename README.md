# TC Exporter for Prometheus

TC Exporter 是一个用于 Prometheus 的导出器，能够通过 [go-tc](https://github.com/florianl/go-tc) netlink 库导出 Linux Traffic Control (TC) 的统计信息。它支持多种队列规则（qdisc）和类（class）的监控，为网络流量控制提供详细的指标数据。

## 功能特性

- 🔍 **全面的 TC 监控**: 支持多种队列规则和类的监控
- 📊 **Prometheus 指标**: 提供标准的 Prometheus 指标格式
- 🌐 **网络命名空间支持**: 支持监控不同网络命名空间中的 TC 配置
- ⚡ **高性能**: 基于 netlink 的高效数据收集
- 🔧 **灵活配置**: 支持 YAML 配置文件
- 📝 **详细日志**: 可配置的日志级别和输出
- 🚀 **系统服务**: 提供 systemd 服务文件
- 🛡️ **限流保护**: 内置请求限流机制
- 📈 **系统监控**: 提供 CPU 和内存使用率监控

### 支持的队列规则

TC Exporter 通过模块化的收集器架构支持多种队列规则，每种 qdisc 都有专门的实现：

**分层队列规则**:
- **HTB** (Hierarchical Token Bucket) - 分层令牌桶
- **CBQ** (Class Based Queueing) - 基于类的队列
- **HFSC** (Hierarchical Fair Service Curve) - 分层公平服务曲线

**队列管理算法**:
- **FQ** (Fair Queue) - 公平队列
- **FQ_CODEL** - 公平队列与 CoDel 算法
- **CODEL** (Controlled Delay) - 控制延迟算法
- **PIE** (Proportional Integral controller Enhanced) - 比例积分控制器增强

**随机化队列规则**:
- **RED** (Random Early Detection) - 随机早期检测
- **SFB** (Stochastic Fair Blue) - 随机公平蓝色算法
- **SFQ** (Stochastic Fairness Queueing) - 随机公平队列
- **CHOKE** (CHOose and Keep for responsive flows) - 响应流选择保持

详细的队列规则说明请参考：[支持的队列规则文档](docs/supported-qdiscs.md)

## 设计文档

- [架构设计](docs/architecture.md) - 系统架构和核心组件说明
- [支持的队列规则](docs/supported-qdiscs.md) - 各种 qdisc 的详细介绍和实现
- [开发文档](docs/development.md) - 项目结构和开发指南
- [优雅关闭功能](docs/graceful-shutdown.md) - 可配置的优雅关闭机制

## 安装

### 系统要求

- **操作系统**: Linux（需要 netlink 支持）
- **权限**: root 权限或 NET_ADMIN capabilities
- **Go 版本**: 1.20+（仅从源码编译时需要）

### 从源码编译

```bash
# 克隆仓库
git clone https://gitee.com/openeuler/uos-tc-exporter.git
cd tc_exporter

# 编译
make build

# 安装到系统（默认 /usr/bin）
sudo make install
```

### 使用 Makefile

```bash
# 查看所有可用命令
make help

# 编译项目（包含版本信息注入）
make build

# 安装二进制文件、配置文件和系统服务
sudo make install

# 清理构建文件
make clean
```

## 配置

### 配置文件

默认配置文件位置: `/etc/uos-exporter/tc-exporter.yaml`

```yaml
# 服务监听配置
address: "127.0.0.1"
port: 9062
metricsPath: "/metrics"

# 日志配置
log:
  level: "debug"
  # log_path: "/var/log/exporter.log"

# 服务器配置
server:
  # 优雅关闭超时时间，支持时间单位：30s, 1m, 2m30s 等
  shutdownTimeout: "30s"

```

### 命令行参数

TC Exporter 使用 [kingpin](https://github.com/alecthomas/kingpin) 进行命令行解析：

```bash
# 查看帮助
tc-exporter --help

# 指定配置文件
tc-exporter --config.file=/path/to/config.yaml


```

## 使用方法

### 启动服务

#### 直接运行

```bash
# 前台运行（需要 root 权限）
sudo ./tc-exporter

# 指定配置文件运行
sudo ./tc-exporter --config.file=/etc/uos-exporter/tc-exporter.yaml
```

#### 系统服务

```bash
# 启动服务
sudo systemctl start uos-tc-exporter

# 设置开机自启
sudo systemctl enable uos-tc-exporter

# 查看服务状态
sudo systemctl status uos-tc-exporter

# 查看服务日志
sudo journalctl -u uos-tc-exporter -f
```

### 访问指标

#### 直接访问

```bash
# 获取所有指标数据
curl http://localhost:9062/metrics

# 查看首页（包含基本信息）
curl http://localhost:9062/

# 检查特定指标
curl -s http://localhost:9062/metrics | grep tc_qdisc_bytes
```

#### Prometheus 配置

在 `prometheus.yml` 中添加以下配置：

```yaml
scrape_configs:
  - job_name: 'tc-exporter'
    static_configs:
      - targets: ['localhost:9062']
    scrape_interval: 15s
    scrape_timeout: 10s
    metrics_path: '/metrics'
```

### 示例指标

TC Exporter 提供了丰富的指标数据：

```prometheus
# 通用 Qdisc 指标
tc_qdisc_bytes_total{device="eth0",qdisc="htb",handle="1:0"} 6789012
tc_qdisc_packets_total{device="eth0",qdisc="htb",handle="1:0"} 12345
tc_qdisc_drops_total{device="eth0",qdisc="htb",handle="1:0"} 0
tc_qdisc_overlimits_total{device="eth0",qdisc="htb",handle="1:0"} 5
tc_qdisc_bps{device="eth0",qdisc="htb",handle="1:0"} 1048576
tc_qdisc_pps{device="eth0",qdisc="htb",handle="1:0"} 150

# HTB 特有指标
tc_qdisc_htb_borrows{device="eth0",handle="1:1"} 10
tc_qdisc_htb_lends{device="eth0",handle="1:1"} 5

# Class 级别指标
tc_class_bytes_total{device="eth0",class="htb",handle="1:1"} 3456789
tc_class_packets_total{device="eth0",class="htb",handle="1:1"} 5432
tc_class_drops_total{device="eth0",class="htb",handle="1:1"} 0

# 系统信息指标
tc_exporter_build_info{version="1.0.0",revision="abc123",branch="main",go_version="go1.20.5"} 1
tc_exporter_cpu_usage 25.6
tc_exporter_memory_usage 134217728
```

## 开发

### 快速开始

```bash
# 安装依赖
go mod download

# 运行测试
go test ./...

# 本地构建
make build

# 运行（需要 root 权限）
sudo ./build/bin/tc-exporter
```

### 项目结构

项目采用清晰的模块化架构：

```
tc_exporter/
├── config/          # 配置管理
├── internal/        # 内部实现
│   ├── exporter/    # 收集器注册和管理
│   ├── metrics/     # 各种指标实现
│   ├── server/      # HTTP 服务器
│   └── tc/          # TC 操作封装
├── pkg/             # 公共工具包
│   ├── logger/      # 日志系统
│   ├── ratelimit/   # 限流器
│   └── utils/       # 通用工具
├── version/         # 版本信息
└── docs/            # 文档
```

### 详细开发指南

请参考 [开发文档](docs/development.md) 了解：
- 完整的项目结构说明
- 开发环境设置
- 代码规范和提交规范
- 如何添加新的队列规则支持
- 测试指南和调试技巧

## 监控和告警

### Grafana 仪表板

可以创建 Grafana 仪表板来可视化 TC 指标：

**推荐图表**:
- TC Qdisc 吞吐量趋势
- Class 级别带宽使用情况
- 丢包率和超限统计
- 系统资源使用监控

**示例查询**:
```promql
# 网络接口吞吐量
rate(tc_qdisc_bytes_total[5m]) * 8

# 丢包率
rate(tc_qdisc_drops_total[5m]) / rate(tc_qdisc_packets_total[5m]) * 100

# HTB 借用带宽频率
rate(tc_qdisc_htb_borrows[5m])
```

### Prometheus 告警规则

```yaml
groups:
  - name: tc-exporter
    rules:
      - alert: TCExporterDown
        expr: up{job="tc-exporter"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "TC Exporter is down"
          description: "TC Exporter has been down for more than 1 minute"
          
      - alert: HighPacketDrops
        expr: rate(tc_qdisc_drops_total[5m]) > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High packet drops detected on {{ $labels.device }}"
          description: "Device {{ $labels.device }} is dropping {{ $value }} packets/sec"
          
      - alert: HighBandwidthUtilization
        expr: rate(tc_qdisc_bytes_total[5m]) * 8 > 800000000  # 800 Mbps
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High bandwidth utilization on {{ $labels.device }}"
          description: "Device {{ $labels.device }} bandwidth usage is {{ $value | humanize }}bps"
```

## 故障排除

### 常见问题

#### 1. 权限不足

```bash
# 错误信息：permission denied 或 operation not permitted

# 解决方案1：使用 sudo 运行
sudo tc-exporter

# 解决方案2：设置 capabilities
sudo setcap cap_net_admin+ep ./tc-exporter

# 解决方案3：添加用户到 netdev 组（某些发行版）
sudo usermod -a -G netdev $USER
```

#### 2. 端口被占用

```bash
# 错误信息：bind: address already in use

# 检查端口使用情况
sudo netstat -tlnp | grep 9062
sudo ss -tlnp | grep 9062

# 解决方案：修改配置文件中的端口
port: 9063
```

#### 3. TC 配置不存在

```bash
# 检查 TC 配置
tc qdisc show
tc class show

# 如果没有 qdisc，创建测试配置
sudo tc qdisc add dev eth0 root handle 1: htb default 12
sudo tc class add dev eth0 parent 1: classid 1:1 htb rate 1mbit
```

#### 4. 服务启动失败

```bash
# 查看详细错误信息
sudo journalctl -u uos-tc-exporter -n 50

# 检查配置文件语法
tc-exporter --config.file=/etc/uos-exporter/tc-exporter.yaml --help

# 检查二进制文件权限
ls -la /usr/bin/tc-exporter
```

### 调试模式

#### 启用详细日志

```bash
# 方法1：环境变量
export LOG_LEVEL=debug
sudo -E tc-exporter

# 方法2：配置文件
log:
  level: "debug"
  log_path: "/var/log/tc-exporter-debug.log"

# 方法3：临时文件日志
sudo tc-exporter 2>&1 | tee debug.log
```

#### 网络调试

```bash
# 检查 netlink 通信
sudo strace -e trace=socket,bind,sendto,recvfrom tc-exporter

# 监控网络接口状态
watch -n 1 'tc qdisc show; echo "---"; tc class show'

# 检查网络接口
ip link show
ip addr show
```

#### HTTP 调试

```bash
# 测试指标端点
curl -v http://localhost:9062/metrics

# 检查响应头
curl -I http://localhost:9062/

# 性能测试
time curl -s http://localhost:9062/metrics > /dev/null
```

### 性能优化

#### 减少指标收集频率

如果系统负载较高，可以：
1. 调整 Prometheus 抓取间隔
2. 使用限流功能
3. 监控特定网络接口

#### 内存使用优化

```bash
# 监控内存使用
ps aux | grep tc-exporter
cat /proc/$(pgrep tc-exporter)/status | grep VmRSS

# 如果内存使用过高，检查：
# 1. 网络接口数量
# 2. qdisc/class 配置复杂度
# 3. 指标收集频率
```


## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 相关链接

- **依赖库**:
  - [go-tc](https://github.com/florianl/go-tc) - TC netlink 库
  - [Prometheus Go Client](https://github.com/prometheus/client_golang) - Prometheus 客户端
  - [Logrus](https://github.com/sirupsen/logrus) - 结构化日志
  - [Kingpin](https://github.com/alecthomas/kingpin) - 命令行解析

- **文档参考**:
  - [Linux Traffic Control](https://tldp.org/HOWTO/Traffic-Control-HOWTO/)
  - [Prometheus](https://prometheus.io/)
  - [Netlink](https://man7.org/linux/man-pages/man7/netlink.7.html)

## 支持

如果您遇到问题或有建议，请：

1. 查看 [故障排除](#故障排除) 部分
2. 检查 [Issues](https://gitee.com/openeuler/uos-tc-exporter) 中的已知问题
3. 创建新的 Issue 详细描述问题
4. 联系维护团队

---

**注意**: 此导出器需要 root 权限或适当的 netlink 权限才能访问 TC 统计信息。建议在生产环境中使用 systemd 服务方式运行。
