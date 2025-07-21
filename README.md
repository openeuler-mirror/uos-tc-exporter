# TC Exporter for Prometheus

TC Exporter 是一个用于 Prometheus 的Exporter，能够通过 netlink 库导出 Linux Traffic Control (TC) 的统计信息。它支持多种队列规则（qdisc）和类（class）的监控，为网络流量控制提供详细的指标数据。

## 功能特性

- 🔍 **全面的 TC 监控**: 支持多种队列规则和类的监控
- 📊 **Prometheus 指标**: 提供标准的 Prometheus 指标格式
- 🌐 **网络命名空间支持**: 支持监控不同网络命名空间中的 TC 配置
- ⚡ **高性能**: 基于 netlink 的高效数据收集
- 🔧 **灵活配置**: 支持 YAML 配置文件
- 📝 **详细日志**: 可配置的日志级别和输出
- 🚀 **系统服务**: 提供 systemd 服务文件

### 支持的队列规则 (Qdisc)

- **HTB (Hierarchical Token Bucket)**: 分层令牌桶
- **CBQ (Class Based Queueing)**: 基于类的队列
- **FQ (Fair Queue)**: 公平队列
- **FQ_CODEL**: 公平队列与 CoDel 算法
- **CODEL**: Controlled Delay 算法
- **PIE**: Proportional Integral controller Enhanced
- **RED**: Random Early Detection
- **SFB**: Stochastic Fair Blue
- **SFQ**: Stochastic Fairness Queueing
- **CHOKE**: CHOose and Keep for responsive flows
- **HFSC**: Hierarchical Fair Service Curve

## 架构设计

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

1. **HTTP Server**: 提供 Prometheus 指标端点
2. **Metrics Collector**: 收集和格式化 TC 指标
3. **TC Client**: 通过 netlink 与内核 TC 子系统通信
4. **Configuration Manager**: 管理应用配置
5. **Logger**: 提供结构化日志输出

## 安装

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/your-org/tc_exporter.git
cd tc_exporter

# 编译
make build

# 安装到系统
sudo make install
```

### 使用 Makefile

```bash
# 查看所有可用命令
make help

# 编译项目
make build

# 安装到系统 (默认 /usr/bin)
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

# 抓取配置
scrape_uri: "http://localhost:9062/metrics"
```

### 命令行参数

```bash
# 查看帮助
tc-exporter --help

# 指定配置文件
tc-exporter --config.file=/path/to/config.yaml

# 指定监听地址
tc-exporter --web.listen-address=:9062

# 启用默认 Prometheus 注册表
tc-exporter --enable-default-prom-reg
```

## 使用方法

### 启动服务

```bash
# 直接运行
./tc-exporter

# 作为系统服务启动
sudo systemctl start uos-tc-exporter
sudo systemctl enable uos-tc-exporter

# 查看服务状态
sudo systemctl status uos-tc-exporter
```

### 访问指标

```bash
# 获取指标数据
curl http://localhost:9062/metrics

# 使用 Prometheus 抓取
# 在 prometheus.yml 中添加:
scrape_configs:
  - job_name: 'tc-exporter'
    static_configs:
      - targets: ['localhost:9062']
```

### 示例指标

```prometheus
# TC Qdisc 指标
tc_qdisc_packets{device="eth0",qdisc="htb",handle="1:0"} 12345
tc_qdisc_bytes{device="eth0",qdisc="htb",handle="1:0"} 6789012
tc_qdisc_drops{device="eth0",qdisc="htb",handle="1:0"} 0

# TC Class 指标
tc_class_packets{device="eth0",class="htb",handle="1:1"} 5432
tc_class_bytes{device="eth0",class="htb",handle="1:1"} 3456789
tc_class_drops{device="eth0",class="htb",handle="1:1"} 0

# 系统信息
tc_exporter_build_info{version="1.0.0",revision="abc123"} 1
```

## 开发

### 项目结构

```
tc_exporter/
├── config/                 # 配置文件
│   ├── config.go          # 配置结构定义
│   └── tc-exporter.yaml   # 默认配置文件
├── internal/              # 内部包
│   ├── exporter/          # 导出器核心逻辑
│   ├── metrics/           # 指标定义和收集
│   ├── server/            # HTTP 服务器
│   └── tc/                # TC 操作封装
├── pkg/                   # 公共包
│   ├── logger/            # 日志工具
│   ├── ratelimit/         # 限流工具
│   └── utils/             # 通用工具
├── version/               # 版本信息
├── main.go               # 程序入口
├── exporter.go           # 导出器主逻辑
├── Makefile              # 构建脚本
└── README.md             # 项目文档
```

### 添加新的指标

1. 在 `internal/metrics/` 目录下创建新的指标文件
2. 实现指标收集逻辑
3. 在 `internal/exporter/` 中注册新的收集器
4. 更新文档

### 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/metrics

# 带覆盖率的测试
go test -cover ./...
```

## 监控和告警

### Grafana 仪表板

可以创建 Grafana 仪表板来可视化 TC 指标：

- TC Qdisc 统计
- TC Class 统计
- 网络接口流量控制效果
- 丢包率监控

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
          
      - alert: HighPacketDrops
        expr: tc_qdisc_drops > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High packet drops detected"
```

## 故障排除

### 常见问题

1. **权限不足**
   ```bash
   # 确保有足够的权限访问 netlink
   sudo tc-exporter
   ```

2. **端口被占用**
   ```bash
   # 检查端口使用情况
   netstat -tlnp | grep 9062
   
   # 修改配置文件中的端口
   ```

3. **TC 配置不存在**
   ```bash
   # 检查 TC 配置
   tc qdisc show
   tc class show
   ```

### 日志级别

```bash
# 设置调试日志
export LOG_LEVEL=debug
tc-exporter

# 或在配置文件中设置
log:
  level: "debug"
```

## 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 相关链接

- [Linux Traffic Control](https://tldp.org/HOWTO/Traffic-Control-HOWTO/)
- [Prometheus](https://prometheus.io/)
- [Netlink](https://man7.org/linux/man-pages/man7/netlink.7.html)

---

**注意**: 此导出器需要 root 权限或适当的 netlink 权限才能访问 TC 统计信息。
