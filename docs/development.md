# 开发文档

## 项目结构

基于实际代码的项目结构如下：

```
tc_exporter/                    # 项目根目录
├── config/                     # 配置相关
│   ├── config.go              # 配置结构定义和命令行解析
│   └── tc-exporter.yaml       # 默认配置文件
├── internal/                   # 内部包
│   ├── exporter/              # 导出器核心逻辑
│   │   ├── config.go          # 导出器配置管理
│   │   ├── metrics.go         # 指标协调器
│   │   └── registry.go        # 收集器注册表
│   ├── metrics/               # 指标定义和收集
│   │   ├── metrics.go         # 基础指标定义
│   │   ├── qdisc.go           # 通用 Qdisc 指标
│   │   ├── qclass.go          # Class 指标收集
│   │   ├── info.go            # 系统信息指标
│   │   ├── cpu.go             # CPU 相关指标
│   │   ├── q_htb.go           # HTB 队列规则
│   │   ├── q_cbq.go           # CBQ 队列规则
│   │   ├── q_fq.go            # FQ 队列规则
│   │   ├── q_fq_codel.go      # FQ_CODEL 队列规则
│   │   ├── q_codel.go         # CODEL 队列规则
│   │   ├── q_pie.go           # PIE 队列规则
│   │   ├── q_red.go           # RED 队列规则
│   │   ├── q_sfb.go           # SFB 队列规则
│   │   ├── q_sfq.go           # SFQ 队列规则
│   │   ├── q_choke.go         # CHOKE 队列规则
│   │   └── q_hfsc.go          # HFSC 队列规则
│   ├── server/                # HTTP 服务器
│   │   ├── server.go          # 主服务器实现
│   │   ├── landpage.go        # 首页处理
│   │   ├── request.go         # 请求处理逻辑
│   │   ├── head.go            # HTTP 头处理
│   │   └── ratelimit_middleware.go # 限流中间件
│   └── tc/                    # TC 操作封装
│       ├── tc.go              # TC 客户端实现
│       └── helper.go          # 辅助函数
├── pkg/                       # 公共包
│   ├── logger/                # 日志工具
│   │   ├── logger.go          # 日志接口定义
│   │   ├── output.go          # 输出管理
│   │   ├── filerotator.go     # 文件轮转
│   │   └── filerotator_test.go # 文件轮转测试
│   ├── ratelimit/             # 限流工具
│   │   ├── ratelimit.go       # 限流器实现
│   │   └── ratelimit_test.go  # 限流器测试
│   └── utils/                 # 通用工具
│       ├── file.go            # 文件操作工具
│       ├── http.go            # HTTP 相关工具
│       └── signal.go          # 信号处理
├── version/                   # 版本信息
│   └── version.go             # 版本定义
├── docs/                      # 文档目录
│   ├── architecture.md        # 架构设计文档
│   ├── supported-qdiscs.md    # 支持的队列规则
│   └── development.md         # 开发文档（本文件）
├── main.go                    # 程序入口
├── exporter.go                # 导出器主逻辑
├── go.mod                     # Go 模块定义
├── go.sum                     # 依赖校验和
├── Makefile                   # 构建脚本
├── uos-tc-exporter.service    # Systemd 服务文件
└── README.md                  # 项目说明文档
```

## 开发环境设置

### 依赖要求

- **Go 1.20+**: 编程语言版本
- **Linux 操作系统**: 必须，因为需要 netlink 支持
- **NET_ADMIN 权限**: 访问 TC 信息需要管理员权限
- **Git**: 版本控制和构建时版本信息注入

### 核心依赖库

根据 `go.mod` 文件，项目主要依赖：

```go
require (
    gitee.com/weidongkl/logrus-formatter v1.1.0  // 日志格式化
    github.com/alecthomas/kingpin v2.2.6+incompatible // 命令行解析
    github.com/dustin/go-humanize v1.0.1          // 人性化显示
    github.com/florianl/go-tc v0.4.5              // TC netlink 库
    github.com/jsimonetti/rtnetlink v1.4.2        // 路由 netlink
    github.com/mdlayher/netlink v1.7.2            // 底层 netlink
    github.com/prometheus/client_golang v1.20.5   // Prometheus 客户端
    github.com/sirupsen/logrus v1.9.3             // 结构化日志
    golang.org/x/sys v0.28.0                      // 系统调用
    gopkg.in/yaml.v2 v2.4.0                       // YAML 解析
)
```

### 本地开发

```bash
# 克隆仓库到本地
git clone <repository-url>
cd tc_exporter

# 安装依赖
go mod download

# 验证依赖
go mod tidy

# 运行测试
go test ./...

# 本地构建
make build

# 运行（需要 root 权限）
sudo ./build/bin/tc-exporter
```

### 开发配置

创建开发配置文件 `config/dev.yaml`:

```yaml
address: "127.0.0.1"
port: 9062
metricsPath: "/metrics"

log:
  level: "debug"
#  log_path: "./debug.log"

# 可选配置
scrape_uri: "http://localhost:9062/metrics"
insecure: false
```

## 代码规范

### 包组织原则

1. **config/**: 全局配置相关代码
2. **internal/**: 项目内部实现，不对外暴露
3. **pkg/**: 可复用的公共包
4. **version/**: 版本信息管理

### 命名规范

- **文件命名**: 使用下划线分隔（如 `q_htb.go`）
- **包命名**: 简短有意义的单词
- **接口命名**: 以 -er 结尾（如 Collector）
- **常量命名**: 使用驼峰式或全大写

### 代码风格

```bash
# 格式化代码
go fmt ./...

# 静态检查
go vet ./...

# 代码规范检查（需要安装 golint）
golint ./...
```

### 提交规范

```bash
# 提交前检查
make lint
make test

# 提交信息格式
<type>(<scope>): <description>

# 类型说明：
# feat: 新功能
# fix: 修复bug
# docs: 文档更新
# style: 代码格式
# refactor: 重构
# test: 测试相关
# chore: 构建过程或辅助工具的变动

# 示例：
feat(metrics): add support for cake qdisc
fix(server): handle graceful shutdown
docs(readme): update installation guide
```

## 添加新的队列规则支持

### 1. 创建新的指标收集器

在 `internal/metrics/` 目录下创建新文件（例如 `q_cake.go`）：

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/sirupsen/logrus"
    "tc_exporter/internal/exporter"
    "tc_exporter/internal/tc"
)

// 在 init 函数中注册收集器
func init() {
    exporter.Register(NewQdiscCake())
}

// 定义 CAKE qdisc 特有的指标结构
type QdiscCake struct {
    // 嵌入通用的 qdisc 指标
    qdiscBytesTotal
    qdiscPacketsTotal
    qdiscDropsTotal
    qdiscOverlimitsTotal
    
    // CAKE 特有的指标
    cakeMemoryUsage  *baseMetrics
    cakeFlows        *baseMetrics
    // ... 其他 CAKE 特有指标
}

// 构造函数
func NewQdiscCake() *QdiscCake {
    return &QdiscCake{
        cakeMemoryUsage: NewMetrics(
            "tc_qdisc_cake_memory_usage",
            "CAKE memory usage in bytes",
            []string{"device", "handle", "parent"},
        ),
        cakeFlows: NewMetrics(
            "tc_qdisc_cake_flows",
            "Number of active flows in CAKE",
            []string{"device", "handle", "parent"},
        ),
        // ... 初始化其他指标
    }
}

// 实现 Prometheus Collector 接口
func (q *QdiscCake) Describe(ch chan<- *prometheus.Desc) {
    q.cakeMemoryUsage.desc.Describe(ch)
    q.cakeFlows.desc.Describe(ch)
    // ... 描述其他指标
}

func (q *QdiscCake) Collect(ch chan<- prometheus.Metric) {
    // 获取所有网络接口
    devices, err := tc.GetDevices()
    if err != nil {
        logrus.Errorf("Failed to get devices: %v", err)
        return
    }
    
    for _, device := range devices {
        // 获取该设备上的 CAKE qdisc
        qdiscs, err := tc.GetQdiscs(device, "cake")
        if err != nil {
            logrus.Debugf("No CAKE qdisc found on %s: %v", device, err)
            continue
        }
        
        for _, qdisc := range qdiscs {
            // 收集 CAKE 特有的统计数据
            q.collectCakeStats(ch, device, qdisc)
        }
    }
}

func (q *QdiscCake) collectCakeStats(ch chan<- prometheus.Metric, device string, qdisc *tc.QdiscInfo) {
    labels := []string{device, qdisc.Handle, qdisc.Parent}
    
    // 收集内存使用情况
    if memUsage := qdisc.GetMemoryUsage(); memUsage > 0 {
        q.cakeMemoryUsage.collect(ch, float64(memUsage), labels)
    }
    
    // 收集流数量
    if flows := qdisc.GetFlowCount(); flows > 0 {
        q.cakeFlows.collect(ch, float64(flows), labels)
    }
    
    // ... 收集其他 CAKE 特有数据
}
```

### 2. 扩展 TC 客户端（如果需要）

如果新的队列规则需要特殊的数据获取逻辑，在 `internal/tc/` 中扩展：

```go
// 在 tc.go 中添加新的方法
func GetCakeStats(device string) (*CakeStats, error) {
    // 实现 CAKE 特有的数据获取逻辑
    // 使用 go-tc 库的相关接口
}
```

### 3. 添加测试

创建对应的测试文件 `internal/metrics/q_cake_test.go`：

```go
package metrics

import (
    "testing"
    "github.com/prometheus/client_golang/prometheus/testutil"
)

func TestQdiscCakeCollector(t *testing.T) {
    // 创建收集器实例
    collector := NewQdiscCake()
    
    // 测试 Describe 方法
    descCh := make(chan *prometheus.Desc, 10)
    collector.Describe(descCh)
    close(descCh)
    
    // 验证描述数量
    count := 0
    for range descCh {
        count++
    }
    
    if count == 0 {
        t.Error("Expected at least one metric description")
    }
    
    // 测试 Collect 方法
    // 注意：这需要在有 CAKE qdisc 的环境中运行
    metricCh := make(chan prometheus.Metric, 10)
    collector.Collect(metricCh)
    close(metricCh)
    
    // 验证指标收集
    // ...
}
```

### 4. 更新文档

在 `docs/supported-qdiscs.md` 中添加新队列规则的说明：

```markdown
### CAKE (Common Applications Kept Enhanced)
**实现文件**: `internal/metrics/q_cake.go`

- **用途**: 现代智能队列管理
- **特点**:
  - 自动带宽检测
  - 流级别公平性
  - 低延迟优化
  - 智能 ECN 支持

**专有指标**:
- `tc_qdisc_cake_memory_usage`: 内存使用量
- `tc_qdisc_cake_flows`: 活跃流数量

**常用场景**:
- 家庭网络
- 小型企业
- 移动网络
```

## 测试指南

### 单元测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/metrics

# 带覆盖率的测试
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 集成测试

```bash
# 需要 root 权限的集成测试
sudo go test -tags=integration ./internal/tc/

# 或者在测试环境中运行
make test-integration
```

### 模拟测试环境

```bash
# 创建测试 TC 配置
sudo tc qdisc add dev lo root handle 1: htb default 12
sudo tc class add dev lo parent 1: classid 1:1 htb rate 1mbit
sudo tc class add dev lo parent 1:1 classid 1:12 htb rate 1mbit ceil 1mbit

# 运行测试
go test ./internal/tc/

# 清理测试配置
sudo tc qdisc del dev lo root
```

### 性能测试

```bash
# CPU 性能分析
go test -cpuprofile cpu.prof ./internal/metrics
go tool pprof cpu.prof

# 内存性能分析
go test -memprofile mem.prof ./internal/metrics
go tool pprof mem.prof

# 基准测试
go test -bench=. ./internal/metrics
```

## 构建和发布

### 使用 Makefile

项目提供了完整的 Makefile 支持：

```bash
# 查看所有可用命令
make help

# 构建项目（包含版本信息注入）
make build

# 清理构建文件
make clean

# 安装到系统（需要 root 权限）
sudo make install

# 安装配置
sudo make install-config
```

### 版本管理

版本信息通过构建时注入：

```bash
# Makefile 中的版本注入
LDFLAGS := "-s -w \
    -X tc_exporter/version.Version=${VERSION} \
    -X tc_exporter/version.Revision=$(shell git rev-list -1 HEAD) \
    -X tc_exporter/version.Branch=$(shell git rev-parse --abbrev-ref HEAD)"
```

### 手动构建

```bash
# 基本构建
go build -o tc-exporter

# 带版本信息的构建
go build -ldflags "-X tc_exporter/version.Version=1.0.0" -o tc-exporter

# 交叉编译（例如 ARM64）
GOOS=linux GOARCH=arm64 go build -o tc-exporter-arm64
```

## 调试技巧

### 启用详细日志

```bash
# 环境变量方式
export LOG_LEVEL=debug
./tc-exporter

# 配置文件方式
log:
  level: "debug"
  log_path: "./debug.log"
```

### 使用 delve 调试器

```bash
# 安装 delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试程序（需要 root 权限）
sudo dlv exec ./tc-exporter -- --config.file=./config/tc-exporter.yaml

# 在调试器中设置断点
(dlv) break main.main
(dlv) continue
```

### 网络调试

```bash
# 检查 netlink 通信
strace -e trace=socket,bind,sendto,recvfrom ./tc-exporter

# 监控系统调用
sudo perf trace -p $(pgrep tc-exporter)

# 检查文件描述符
sudo lsof -p $(pgrep tc-exporter)
```

### HTTP 接口调试

```bash
# 检查指标端点
curl -v http://localhost:9062/metrics

# 检查服务状态
curl http://localhost:9062/

# 性能测试
ab -n 100 -c 10 http://localhost:9062/metrics
```

## 常见问题解决

### 权限问题

```bash
# 方案1：使用 sudo 运行
sudo ./tc-exporter

# 方案2：设置 capabilities
sudo setcap cap_net_admin+ep ./tc-exporter

# 方案3：添加用户到 netdev 组（某些发行版）
sudo usermod -a -G netdev $USER
```

### 编译问题

```bash
# 清理模块缓存
go clean -modcache

# 重新下载依赖
go mod download

# 验证模块
go mod verify

# 更新依赖到最新版本
go get -u ./...
```

### 运行时问题

```bash
# 检查 TC 配置是否存在
tc qdisc show
tc class show

# 检查网络接口
ip link show

# 检查进程状态
ps aux | grep tc-exporter
systemctl status uos-tc-exporter
```

## 贡献指南

### Pull Request 流程

1. **Fork** 项目到自己的 GitHub 账户
2. **Clone** fork 的仓库到本地
3. **创建** 功能分支：`git checkout -b feature/amazing-feature`
4. **开发** 并测试新功能
5. **提交** 更改：`git commit -m 'feat: add amazing feature'`
6. **推送** 分支：`git push origin feature/amazing-feature`
7. **创建** Pull Request

### 代码审查要点

- 代码风格符合 Go 规范
- 添加适当的测试覆盖
- 更新相关文档
- 性能影响评估
- 安全考虑

### 发布流程

1. 更新 `version/version.go` 中的版本号
2. 更新 `CHANGELOG.md`
3. 创建 Git tag
4. 自动构建和发布（通过 CI/CD）

## 性能优化建议

### 内存优化

- 使用对象池减少 GC 压力
- 及时释放不再使用的资源
- 避免在热路径中分配大量临时对象

### CPU 优化

- 使用合适的数据结构
- 避免不必要的系统调用
- 缓存频繁访问的数据

### 网络优化

- 批量处理 netlink 请求
- 使用连接复用
- 合理设置超时时间 