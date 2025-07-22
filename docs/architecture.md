# TC Exporter 架构设计

## 整体架构

TC Exporter 采用模块化设计，通过 [go-tc](https://github.com/florianl/go-tc) netlink 库与 Linux 内核 TC 子系统通信，收集流量控制统计数据并以 Prometheus 格式提供指标。

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

## 核心组件

### 1. HTTP Server (`internal/server/`)
- **职责**: 提供 Prometheus 指标端点和管理界面
- **主要文件**:
  - `server.go`: 主服务器实现
  - `landpage.go`: 首页处理
  - `request.go`: 请求处理逻辑
  - `ratelimit_middleware.go`: 限流中间件
- **功能**:
  - 监听端口 9062（可配置）
  - 处理 `/metrics` 请求
  - 提供管理界面
  - 支持限流保护

### 2. Metrics Collector (`internal/metrics/`)
- **职责**: 收集和格式化 TC 指标
- **主要文件**:
  - `metrics.go`: 基础指标定义
  - `qdisc.go`: Qdisc 通用指标收集
  - `qclass.go`: Class 指标收集
  - `q_*.go`: 各种具体的 qdisc 实现
  - `info.go`: 系统信息指标
  - `cpu.go`: CPU 相关指标
- **支持的队列规则**:
  - HTB (`q_htb.go`)
  - CBQ (`q_cbq.go`)
  - FQ (`q_fq.go`)
  - FQ_CODEL (`q_fq_codel.go`)
  - CODEL (`q_codel.go`)
  - PIE (`q_pie.go`)
  - RED (`q_red.go`)
  - SFB (`q_sfb.go`)
  - SFQ (`q_sfq.go`)
  - CHOKE (`q_choke.go`)
  - HFSC (`q_hfsc.go`)

### 3. TC Client (`internal/tc/`)
- **职责**: 通过 netlink 与内核 TC 子系统通信
- **主要文件**:
  - `tc.go`: TC 操作封装
  - `helper.go`: 辅助函数
- **功能**:
  - 获取网络接口列表
  - 查询 qdisc 配置和统计
  - 查询 class 配置和统计
  - 支持网络命名空间
- **实现**: 基于 `github.com/florianl/go-tc` 库

### 4. Exporter Registry (`internal/exporter/`)
- **职责**: 管理指标收集器注册和协调
- **主要文件**:
  - `registry.go`: 收集器注册表
  - `metrics.go`: 指标协调器
  - `config.go`: 导出器配置
- **功能**:
  - 注册各种指标收集器
  - 协调指标收集过程
  - 管理收集器生命周期

### 5. Configuration Manager (`config/`)
- **职责**: 管理应用配置
- **主要文件**:
  - `config.go`: 配置结构定义和命令行解析
  - `tc-exporter.yaml`: 默认配置文件
- **配置项**:
  - 服务监听地址和端口
  - 指标路径配置
  - 日志配置
  - 抓取URI配置
- **特性**:
  - 支持命令行参数覆盖
  - YAML 配置文件支持
  - 配置验证

### 6. 公共工具包 (`pkg/`)

#### Logger (`pkg/logger/`)
- **职责**: 提供结构化日志输出
- **主要文件**:
  - `logger.go`: 日志接口定义
  - `output.go`: 输出管理
  - `filerotator.go`: 文件轮转
- **功能**:
  - 多级别日志（debug, info, warn, error）
  - 支持文件输出和轮转
  - 结构化日志格式

#### Rate Limiter (`pkg/ratelimit/`)
- **职责**: 提供限流功能
- **文件**: `ratelimit.go`
- **功能**: 保护服务免受过度请求

#### Utils (`pkg/utils/`)
- **职责**: 通用工具函数
- **主要文件**:
  - `http.go`: HTTP 相关工具
  - `file.go`: 文件操作工具
  - `signal.go`: 信号处理

### 7. Version Management (`version/`)
- **职责**: 版本信息管理
- **文件**: `version.go`
- **功能**:
  - 版本号定义
  - Git 版本信息（通过构建时注入）
  - Go 版本信息

## 数据流程

### 1. 启动阶段
```
main.go → exporter.go → server.NewServer()
│
├── 初始化日志系统 (pkg/logger)
├── 解析配置文件 (config/)
├── 创建指标收集器 (internal/metrics)
├── 注册收集器 (internal/exporter)
└── 启动HTTP服务器 (internal/server)
```

### 2. 指标收集阶段
```
HTTP Request (/metrics)
│
├── Rate Limiting Check (pkg/ratelimit)
├── Metrics Collection
│   ├── TC Data Collection (internal/tc)
│   │   └── Netlink Communication
│   ├── System Info Collection (internal/metrics/info.go)
│   ├── CPU Metrics Collection (internal/metrics/cpu.go)
│   └── Qdisc/Class Metrics (internal/metrics/q_*.go)
│
└── Prometheus Format Response
```

### 3. 错误处理
- 网络接口不存在时跳过
- Netlink 通信失败时记录错误并返回默认值
- 配置错误时提供详细错误信息
- 支持优雅关闭

## 技术栈

### 核心依赖
- **Go 1.20+**: 编程语言
- **github.com/florianl/go-tc**: TC netlink 库
- **github.com/prometheus/client_golang**: Prometheus 客户端
- **github.com/sirupsen/logrus**: 结构化日志
- **github.com/alecthomas/kingpin**: 命令行参数解析
- **gopkg.in/yaml.v2**: YAML 配置解析

### 系统要求
- Linux 操作系统（netlink 支持）
- NET_ADMIN 权限（访问 TC 信息）

## 性能考虑

### 优化策略
- **高效的 netlink 通信**: 使用专门的 go-tc 库
- **并发安全**: 支持多个 Prometheus 请求同时处理
- **限流保护**: 防止过度请求影响系统性能
- **错误恢复**: 网络异常时自动重试
- **资源管理**: 及时释放 netlink 连接

### 监控指标
- 指标收集耗时
- 错误计数
- 请求频率
- 系统资源使用情况

## 扩展性设计

### 插件化架构
- **模块化收集器**: 每个 qdisc 类型独立实现
- **统一接口**: 通过 baseMetrics 提供统一的指标接口
- **注册机制**: 通过 registry 动态注册收集器

### 配置化
- **YAML 配置**: 灵活的配置管理
- **命令行覆盖**: 支持运行时参数调整
- **环境变量**: 支持容器化部署

### 可维护性
- **清晰的包结构**: 职责分离，便于维护
- **完整的测试**: 单元测试和集成测试
- **文档化**: 详细的代码注释和文档

## 部署架构

### Systemd 服务
- **服务文件**: `uos-tc-exporter.service`
- **安装位置**: `/usr/lib/systemd/system/`
- **配置位置**: `/etc/uos-exporter/`
- **日志**: systemd journal 或自定义日志文件

### 容器化支持
- 支持 Docker 部署
- 需要特权模式访问宿主机网络栈
- 配置文件可通过卷挂载 