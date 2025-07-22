# 支持的队列规则 (Qdisc)

TC Exporter 通过 [go-tc](https://github.com/florianl/go-tc) 库支持监控多种 Linux Traffic Control 队列规则的统计信息。每种队列规则都有对应的收集器实现，通过注册机制动态加载。

## 通用指标 (所有 Qdisc)

所有队列规则都支持以下基础指标（由 `qdisc.go` 实现）：

- `tc_qdisc_bytes_total`: 处理的总字节数
- `tc_qdisc_packets_total`: 处理的总数据包数
- `tc_qdisc_drops_total`: 丢弃的数据包总数
- `tc_qdisc_overlimits_total`: 超出限制的总次数
- `tc_qdisc_bps`: 当前字节传输速率
- `tc_qdisc_pps`: 当前数据包传输速率

## 分层队列规则

### HTB (Hierarchical Token Bucket) - 分层令牌桶
**实现文件**: `internal/metrics/q_htb.go`

- **用途**: 提供分层带宽控制和流量整形
- **特点**:
  - 支持带宽保证和限制
  - 可以借用父类未使用的带宽
  - 适用于复杂的带宽分配场景
  - 支持优先级调度

**专有指标**:
- `tc_qdisc_htb_borrows`: 借用带宽的次数
- `tc_qdisc_htb_ctokens`: C 令牌数
- `tc_qdisc_htb_giants`: 大包处理次数
- `tc_qdisc_htb_lends`: 出借带宽的次数

**常用场景**: 
- ISP 带宽管理
- 企业网络 QoS
- 多租户带宽隔离

### CBQ (Class Based Queueing) - 基于类的队列
**实现文件**: `internal/metrics/q_cbq.go`

- **用途**: 基于类的流量分类和调度
- **特点**:
  - 支持流量分类
  - 提供优先级调度
  - 可配置借用机制
  - 传统的带宽管理方案

**常用场景**:
- 传统的流量分类
- 简单的优先级调度
- 兼容性要求较高的环境

### HFSC (Hierarchical Fair Service Curve) - 分层公平服务曲线
**实现文件**: `internal/metrics/q_hfsc.go`

- **用途**: 提供实时和链路共享保证
- **特点**:
  - 支持实时服务保证
  - 链路共享算法
  - 延迟敏感应用友好
  - 复杂的服务曲线算法

**常用场景**:
- VoIP 应用
- 视频流媒体
- 实时游戏
- 金融交易系统

## 队列管理算法

### FQ (Fair Queue) - 公平队列
**实现文件**: `internal/metrics/q_fq.go`

- **用途**: 为每个流提供公平的带宽分配
- **特点**:
  - 流级别的公平性
  - 防止单一流占用过多带宽
  - 低延迟特性
  - 现代高性能网络栈

**常用场景**:
- 数据中心网络
- 高性能计算环境
- 云计算平台

### FQ_CODEL - 公平队列与 CoDel 算法
**实现文件**: `internal/metrics/q_fq_codel.go`

- **用途**: 结合公平队列和 CoDel 主动队列管理
- **特点**:
  - 流级别隔离
  - 主动丢包管理
  - 低延迟保证
  - 自适应算法

**常用场景**:
- 现代 Linux 系统默认
- 互联网连接
- 混合流量环境
- 边缘计算

### CODEL (Controlled Delay) - 控制延迟算法
**实现文件**: `internal/metrics/q_codel.go`

- **用途**: 主动队列管理，控制排队延迟
- **特点**:
  - 自适应丢包
  - 无需配置参数
  - 有效控制缓冲膨胀
  - 延迟目标导向

**常用场景**:
- 家庭路由器
- 边缘设备
- 延迟敏感应用
- IoT 网关

### PIE (Proportional Integral controller Enhanced) - 比例积分控制器增强
**实现文件**: `internal/metrics/q_pie.go`

- **用途**: 基于延迟的主动队列管理
- **特点**:
  - 延迟目标导向
  - 快速响应
  - 稳定的队列长度
  - 自适应控制

**常用场景**:
- 高速网络环境
- 数据中心互连
- 运营商网络
- 5G 网络

## 随机化队列规则

### RED (Random Early Detection) - 随机早期检测
**实现文件**: `internal/metrics/q_red.go`

- **用途**: 通过随机丢包避免拥塞
- **特点**:
  - 队列长度感知
  - 随机丢包策略
  - 避免全局同步
  - 经典的 AQM 算法

**常用场景**:
- 传统网络设备
- TCP 流量优化
- 拥塞避免
- 学术研究

### SFB (Stochastic Fair Blue) - 随机公平蓝色算法
**实现文件**: `internal/metrics/q_sfb.go`

- **用途**: 基于流的公平性和蓝色算法
- **特点**:
  - 流级别保护
  - 恶意流检测
  - 动态阈值调整
  - 抗攻击能力

**常用场景**:
- 抗 DDoS 攻击
- 多流环境
- 服务质量保护
- 网络安全

### SFQ (Stochastic Fairness Queueing) - 随机公平队列
**实现文件**: `internal/metrics/q_sfq.go`

- **用途**: 简单的流级别公平性
- **特点**:
  - 哈希分类
  - 轮询调度
  - 低开销实现
  - 简单有效

**常用场景**:
- 简单流量管理
- 资源受限环境
- 基本公平性需求
- 嵌入式系统

### CHOKE (CHOose and Keep for responsive flows) - 响应流选择保持
**实现文件**: `internal/metrics/q_choke.go`

- **用途**: 优先保护响应式流量
- **特点**:
  - 流响应性检测
  - 选择性丢包
  - TCP 友好
  - 非响应流惩罚

**常用场景**:
- 混合协议环境
- TCP/UDP 流量共存
- 响应性优化
- 实时通信保护

## Class 级别指标

对于支持类（class）的分层队列规则（如 HTB、CBQ、HFSC），TC Exporter 还收集类级别的指标（由 `qclass.go` 实现）：

- `tc_class_bytes_total`: 类处理的总字节数
- `tc_class_packets_total`: 类处理的总数据包数
- `tc_class_drops_total`: 类丢弃的总数据包数
- `tc_class_overlimits_total`: 类超出限制的总次数

## 系统信息指标

除了 TC 相关指标，exporter 还提供系统信息指标（由 `info.go` 和 `cpu.go` 实现）：

- `tc_exporter_build_info`: 构建信息（版本、修订版本等）
- `tc_exporter_cpu_usage`: CPU 使用率
- `tc_exporter_memory_usage`: 内存使用情况

## 标签信息

所有指标都包含以下标签用于区分：

- `device`: 网络接口名称（如 eth0、wlan0）
- `qdisc`: 队列规则类型（如 htb、fq_codel）
- `handle`: qdisc/class 句柄标识符
- `parent`: 父级句柄（仅适用于 class）
- `kind`: 队列规则种类

## 实现架构

### 注册机制
每个队列规则的实现都通过 `init()` 函数自动注册到 exporter：

```go
func init() {
    exporter.Register(NewQdiscHtb())
}
```

### 统一接口
所有收集器都实现统一的 Prometheus Collector 接口：
- `Describe(chan<- *prometheus.Desc)`
- `Collect(chan<- prometheus.Metric)`

### 数据获取
通过 `internal/tc` 包封装的 netlink 接口获取内核数据：
- 自动发现网络接口
- 查询 qdisc 和 class 配置
- 获取统计数据

## 配置建议

### 高吞吐量场景
- **推荐**: FQ, FQ_CODEL
- **特点**: 优化并发处理，减少锁竞争

### 低延迟场景  
- **推荐**: CODEL, PIE, FQ_CODEL
- **特点**: 主动队列管理，控制排队延迟

### 带宽管理场景
- **推荐**: HTB, HFSC
- **特点**: 精确的带宽控制和分配

### 简单公平性场景
- **推荐**: SFQ, FQ
- **特点**: 实现简单，开销较低

### DDoS 防护场景
- **推荐**: SFB, CHOKE
- **特点**: 恶意流检测和保护

## 性能考虑

- **轻量级实现**: 每个收集器独立运行，避免相互影响
- **缓存机制**: 避免频繁的 netlink 调用
- **错误处理**: 单个队列规则出错不影响其他收集器
- **可扩展性**: 新的队列规则可以通过添加对应的收集器文件轻松支持 