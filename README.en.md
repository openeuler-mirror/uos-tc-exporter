# TC Exporter for Prometheus

TC Exporter is a Prometheus exporter that exports Linux Traffic Control (TC) statistics through the [go-tc](https://github.com/florianl/go-tc) netlink library. It supports monitoring various queue disciplines (qdisc) and classes, providing detailed metrics for network traffic control.

## Features

- 🔍 **Comprehensive TC Monitoring**: Supports monitoring of various queue disciplines and classes
- 📊 **Prometheus Metrics**: Provides standard Prometheus metrics format
- 🌐 **Network Namespace Support**: Supports monitoring TC configurations in different network namespaces
- ⚡ **High Performance**: Efficient data collection based on netlink
- 🔧 **Flexible Configuration**: Supports YAML configuration files
- 📝 **Detailed Logging**: Configurable log levels and output
- 🚀 **System Service**: Provides systemd service files
- 🛡️ **Rate Limiting**: Built-in request rate limiting mechanism
- 📈 **System Monitoring**: Provides CPU and memory usage monitoring

### Supported Queue Disciplines

TC Exporter supports multiple queue disciplines through a modular collector architecture, with dedicated implementations for each qdisc type:

**Hierarchical Queue Disciplines**:
- **HTB** (Hierarchical Token Bucket) - Hierarchical token bucket
- **CBQ** (Class Based Queueing) - Class-based queuing
- **HFSC** (Hierarchical Fair Service Curve) - Hierarchical fair service curve

**Queue Management Algorithms**:
- **FQ** (Fair Queue) - Fair queue
- **FQ_CODEL** - Fair queue with CoDel algorithm
- **CODEL** (Controlled Delay) - Controlled delay algorithm
- **PIE** (Proportional Integral controller Enhanced) - Proportional integral controller enhanced

**Randomized Queue Disciplines**:
- **RED** (Random Early Detection) - Random early detection
- **SFB** (Stochastic Fair Blue) - Stochastic fair blue algorithm
- **SFQ** (Stochastic Fairness Queueing) - Stochastic fairness queueing
- **CHOKE** (CHOose and Keep for responsive flows) - Choose and keep for responsive flows

For detailed queue discipline descriptions, please refer to: [Supported Queue Disciplines Documentation](docs/supported-qdiscs.md)

## Design Documentation

- [Architecture Design](docs/architecture.md) - System architecture and core components
- [Supported Queue Disciplines](docs/supported-qdiscs.md) - Detailed introduction and implementation of various qdiscs
- [Development Guide](docs/development.md) - Project structure and development guidelines

## Installation

### System Requirements

- **Operating System**: Linux (netlink support required)
- **Permissions**: root privileges or NET_ADMIN capabilities
- **Go Version**: 1.20+ (only required for building from source)

### Build from Source

```bash
# Clone repository
git clone https://gitee.com/openeuler/uos-tc-exporter.git
cd tc_exporter

# Build
make build

# Install to system (default /usr/bin)
sudo make install
```

### Using Makefile

```bash
# View all available commands
make help

# Build project (includes version information injection)
make build

# Install binary files, configuration files and system services
sudo make install

# Clean build files
make clean
```

## Configuration

### Configuration File

Default configuration file location: `/etc/uos-exporter/tc-exporter.yaml`

```yaml
# Service listening configuration
address: "127.0.0.1"
port: 9062
metricsPath: "/metrics"

# Log configuration
log:
  level: "debug"
  # log_path: "/var/log/exporter.log"
```

### Command Line Arguments

TC Exporter uses [kingpin](https://github.com/alecthomas/kingpin) for command line parsing:

```bash
# View help
tc-exporter --help

# Specify configuration file
tc-exporter --config.file=/path/to/config.yaml

```

## Usage

### Starting the Service

#### Direct Execution

```bash
# Run in foreground (requires root privileges)
sudo ./tc-exporter

# Run with specified configuration file
sudo ./tc-exporter --config.file=/etc/uos-exporter/tc-exporter.yaml
```

#### System Service

```bash
# Start service
sudo systemctl start uos-tc-exporter

# Enable auto-start on boot
sudo systemctl enable uos-tc-exporter

# Check service status
sudo systemctl status uos-tc-exporter

# View service logs
sudo journalctl -u uos-tc-exporter -f
```

### Accessing Metrics

#### Direct Access

```bash
# Get all metrics data
curl http://localhost:9062/metrics

# View homepage (includes basic information)
curl http://localhost:9062/

# Check specific metrics
curl -s http://localhost:9062/metrics | grep tc_qdisc_bytes
```

#### Prometheus Configuration

Add the following configuration to `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'tc-exporter'
    static_configs:
      - targets: ['localhost:9062']
    scrape_interval: 15s
    scrape_timeout: 10s
    metrics_path: '/metrics'
```

### Example Metrics

TC Exporter provides rich metrics data:

```prometheus
# Generic Qdisc metrics
tc_qdisc_bytes_total{device="eth0",qdisc="htb",handle="1:0"} 6789012
tc_qdisc_packets_total{device="eth0",qdisc="htb",handle="1:0"} 12345
tc_qdisc_drops_total{device="eth0",qdisc="htb",handle="1:0"} 0
tc_qdisc_overlimits_total{device="eth0",qdisc="htb",handle="1:0"} 5
tc_qdisc_bps{device="eth0",qdisc="htb",handle="1:0"} 1048576
tc_qdisc_pps{device="eth0",qdisc="htb",handle="1:0"} 150

# HTB-specific metrics
tc_qdisc_htb_borrows{device="eth0",handle="1:1"} 10
tc_qdisc_htb_lends{device="eth0",handle="1:1"} 5

# Class-level metrics
tc_class_bytes_total{device="eth0",class="htb",handle="1:1"} 3456789
tc_class_packets_total{device="eth0",class="htb",handle="1:1"} 5432
tc_class_drops_total{device="eth0",class="htb",handle="1:1"} 0

# System information metrics
tc_exporter_build_info{version="1.0.0",revision="abc123",branch="main",go_version="go1.20.5"} 1
tc_exporter_cpu_usage 25.6
tc_exporter_memory_usage 134217728
```

## Development

### Quick Start

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Local build
make build

# Run (requires root privileges)
sudo ./build/bin/tc-exporter
```

### Project Structure

The project adopts a clear modular architecture:

```
tc_exporter/
├── config/          # Configuration management
├── internal/        # Internal implementation
│   ├── exporter/    # Collector registration and management
│   ├── metrics/     # Various metric implementations
│   ├── server/      # HTTP server
│   └── tc/          # TC operation wrapper
├── pkg/             # Public utility packages
│   ├── logger/      # Logging system
│   ├── ratelimit/   # Rate limiter
│   └── utils/       # Common utilities
├── version/         # Version information
└── docs/            # Documentation
```

### Detailed Development Guide

Please refer to the [Development Documentation](docs/development.md) for:
- Complete project structure description
- Development environment setup
- Code standards and commit conventions
- How to add new queue discipline support
- Testing guides and debugging techniques

## Monitoring and Alerting

### Grafana Dashboard

You can create Grafana dashboards to visualize TC metrics:

**Recommended Charts**:
- TC Qdisc throughput trends
- Class-level bandwidth usage
- Packet drop and overlimit statistics
- System resource usage monitoring

**Example Queries**:
```promql
# Network interface throughput
rate(tc_qdisc_bytes_total[5m]) * 8

# Packet drop rate
rate(tc_qdisc_drops_total[5m]) / rate(tc_qdisc_packets_total[5m]) * 100

# HTB bandwidth borrowing frequency
rate(tc_qdisc_htb_borrows[5m])
```

### Prometheus Alert Rules

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

## Troubleshooting

### Common Issues

#### 1. Insufficient Permissions

```bash
# Error message: permission denied or operation not permitted

# Solution 1: Run with sudo
sudo tc-exporter

# Solution 2: Set capabilities
sudo setcap cap_net_admin+ep ./tc-exporter

# Solution 3: Add user to netdev group (some distributions)
sudo usermod -a -G netdev $USER
```

#### 2. Port Already in Use

```bash
# Error message: bind: address already in use

# Check port usage
sudo netstat -tlnp | grep 9062
sudo ss -tlnp | grep 9062

# Solution: Modify port in configuration file
port: 9063
```

#### 3. TC Configuration Not Found

```bash
# Check TC configuration
tc qdisc show
tc class show

# If no qdisc exists, create test configuration
sudo tc qdisc add dev eth0 root handle 1: htb default 12
sudo tc class add dev eth0 parent 1: classid 1:1 htb rate 1mbit
```

#### 4. Service Startup Failed

```bash
# View detailed error information
sudo journalctl -u uos-tc-exporter -n 50

# Check configuration file syntax
tc-exporter --config.file=/etc/uos-exporter/tc-exporter.yaml --help

# Check binary file permissions
ls -la /usr/bin/tc-exporter
```

### Debug Mode

#### Enable Verbose Logging

```bash
# Method 1: Environment variable
export LOG_LEVEL=debug
sudo -E tc-exporter

# Method 2: Configuration file
log:
  level: "debug"
  log_path: "/var/log/tc-exporter-debug.log"

# Method 3: Temporary file logging
sudo tc-exporter 2>&1 | tee debug.log
```

#### Network Debugging

```bash
# Check netlink communication
sudo strace -e trace=socket,bind,sendto,recvfrom tc-exporter

# Monitor network interface status
watch -n 1 'tc qdisc show; echo "---"; tc class show'

# Check network interfaces
ip link show
ip addr show
```

#### HTTP Debugging

```bash
# Test metrics endpoint
curl -v http://localhost:9062/metrics

# Check response headers
curl -I http://localhost:9062/

# Performance test
time curl -s http://localhost:9062/metrics > /dev/null
```

### Performance Optimization

#### Reduce Metrics Collection Frequency

If system load is high, you can:
1. Adjust Prometheus scrape interval
2. Use rate limiting features
3. Monitor specific network interfaces

#### Memory Usage Optimization

```bash
# Monitor memory usage
ps aux | grep tc-exporter
cat /proc/$(pgrep tc-exporter)/status | grep VmRSS

# If memory usage is too high, check:
# 1. Number of network interfaces
# 2. qdisc/class configuration complexity
# 3. Metrics collection frequency
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Related Links

- **Dependencies**:
  - [go-tc](https://github.com/florianl/go-tc) - TC netlink library
  - [Prometheus Go Client](https://github.com/prometheus/client_golang) - Prometheus client
  - [Logrus](https://github.com/sirupsen/logrus) - Structured logging
  - [Kingpin](https://github.com/alecthomas/kingpin) - Command line parsing

- **Documentation References**:
  - [Linux Traffic Control](https://tldp.org/HOWTO/Traffic-Control-HOWTO/)
  - [Prometheus](https://prometheus.io/)
  - [Netlink](https://man7.org/linux/man-pages/man7/netlink.7.html)

## Support

If you encounter issues or have suggestions:

1. Check the [Troubleshooting](#troubleshooting) section
2. Review [Issues](https://github.com/your-org/tc_exporter/issues) for known problems
3. Create a new Issue with detailed problem description
4. Contact the maintenance team

---

**Note**: This exporter requires root privileges or appropriate netlink permissions to access TC statistics. It is recommended to run as a systemd service in production environments.
