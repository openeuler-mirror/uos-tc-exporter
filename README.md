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

## 快速开始

### 安装

```bash
# 克隆仓库
git clone https://gitee.com/openeuler/uos-tc-exporter.git
cd tc_exporter

# 编译
make build

# 安装到系统
sudo make install
```

### 配置

默认配置文件位置: `/etc/uos-exporter/tc-exporter.yaml`

```yaml
# 服务监听配置
address: "127.0.0.1"
port: 9062
metricsPath: "/metrics"

# 日志配置
log:
  level: "info"
  logPath: "/var/log/tc-exporter.log"
  maxSize: "10MB"
  maxAge: "168h"  # 7天

# 服务器配置
server:
  shutdownTimeout: "30s"
```

### 启动服务

```bash
# 直接运行
sudo ./tc-exporter

# 或使用 systemd 服务
sudo systemctl start uos-tc-exporter
sudo systemctl enable uos-tc-exporter
```

### 访问指标

```bash
# 获取指标数据
curl http://localhost:9062/metrics

# 查看首页
curl http://localhost:9062/
```

## 支持的队列规则

TC Exporter 支持多种队列规则，包括：

- **HTB** (Hierarchical Token Bucket)
- **CBQ** (Class Based Queueing) 
- **HFSC** (Hierarchical Fair Service Curve)
- **FQ** (Fair Queue)
- **FQ_CODEL** (Fair Queue with Controlled Delay)
- **CODEL** (Controlled Delay)
- **PIE** (Proportional Integral controller Enhanced)
- **RED** (Random Early Detection)
- **SFB** (Stochastic Fair Blue)
- **SFQ** (Stochastic Fairness Queueing)
- **CHOKE** (CHOose and Keep for responsive flows)

## 设计文档

详细的架构设计、开发指南和技术实现请参考：[设计文档](docs/design.md)

## 监控和告警

### Prometheus 配置

在 `prometheus.yml` 中添加：

```yaml
scrape_configs:
  - job_name: 'tc-exporter'
    static_configs:
      - targets: ['localhost:9062']
    scrape_interval: 15s
```

### 示例指标

```prometheus
# Qdisc 指标
tc_qdisc_bytes_total{device="eth0",qdisc="htb"} 6789012
tc_qdisc_packets_total{device="eth0",qdisc="htb"} 12345
tc_qdisc_drops_total{device="eth0",qdisc="htb"} 0

# Class 指标
tc_class_bytes_total{device="eth0",class="htb"} 3456789
tc_class_packets_total{device="eth0",class="htb"} 5432
```

## 故障排除

### 常见问题

1. **权限不足**: 使用 sudo 运行或设置 capabilities
2. **端口被占用**: 修改配置文件中的端口
3. **TC 配置不存在**: 检查网络接口的 TC 配置

### 调试模式

```bash
# 启用详细日志
export LOG_LEVEL=debug
sudo -E tc-exporter
```

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 支持

如果您遇到问题或有建议，请查看 [设计文档](docs/design.md) 或创建 Issue。

---

**注意**: 此导出器需要 root 权限或适当的 netlink 权限才能访问 TC 统计信息。
