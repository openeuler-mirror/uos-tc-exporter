# TC Exporter Docker 部署指南

## 概述

本文档介绍如何使用 Docker 部署 TC Exporter 服务。我们提供了多种部署方式，包括简单的 Docker 运行、使用部署脚本以及 Docker Compose。

## 快速开始

### 方式一：使用部署脚本（推荐）

```bash
# 给予脚本执行权限
chmod +x deploy.sh

# 使用默认端口部署
./deploy.sh

# 使用自定义端口部署
./deploy.sh -p 8080

# 仅构建镜像，不运行容器
./deploy.sh --build-only
```

### 方式二：手动 Docker 部署

```bash
# 构建镜像
docker build -t tc-exporter:latest .

# 运行容器
docker run -d \
  --name tc-exporter \
  --restart unless-stopped \
  --privileged \
  --cap-add=NET_ADMIN \
  -p 9062:9062 \
  -v /var/log:/var/log \
  tc-exporter:latest
```

### 方式三：Docker Compose 部署

```bash
# 创建 docker-compose.yml 文件后运行
docker-compose up -d
```

## 部署脚本功能

部署脚本 `deploy.sh` 提供以下功能：

- ✅ **自动检查**：检查 Docker 环境
- ✅ **镜像构建**：自动构建 Docker 镜像
- ✅ **容器管理**：自动处理现有容器
- ✅ **健康检查**：等待服务启动并验证
- ✅ **端口配置**：支持自定义端口映射
- ✅ **错误处理**：详细的错误信息和日志

## 端口配置

TC Exporter 默认使用端口 `9062`。如果需要修改端口，可以使用以下方式：

```bash
# 使用部署脚本
./deploy.sh -p 8080

# 手动运行
docker run -d -p 8080:9062 tc-exporter:latest
```

## 健康检查

部署完成后，可以通过以下端点检查服务状态：

- **健康检查**：`http://localhost:9062/health`
- **就绪检查**：`http://localhost:9062/ready`
- **存活检查**：`http://localhost:9062/live`
- **指标端点**：`http://localhost:9062/metrics`

## 容器配置

### 必需的特权

TC Exporter 需要以下特权来访问系统网络信息：

```yaml
privileged: true
cap_add:
  - NET_ADMIN
```

### 卷挂载

建议挂载以下卷：

- `/var/log`：日志目录
- `/etc/uos-exporter`：配置文件目录

## 环境变量

可以通过环境变量配置服务：

```bash
docker run -d \
  -e LOG_LEVEL=info \
  -e TC_EXPORTER_ADDRESS=0.0.0.0 \
  tc-exporter:latest
```

## 故障排除

### 常见问题

1. **权限不足**
   ```bash
   # 确保使用特权模式运行
   docker run --privileged --cap-add=NET_ADMIN ...
   ```

2. **端口冲突**
   ```bash
   # 检查端口占用
   netstat -tlnp | grep 9062
   # 或使用不同端口
   ./deploy.sh -p 9080
   ```

3. **健康检查失败**
   ```bash
   # 查看容器日志
   docker logs tc-exporter
   # 检查服务状态
   curl http://localhost:9062/health
   ```

### 日志查看

```bash
# 实时查看日志
docker logs -f tc-exporter

# 查看最近100行日志
docker logs --tail 100 tc-exporter
```

### 容器管理

```bash
# 停止容器
docker stop tc-exporter

# 启动容器
docker start tc-exporter

# 重启容器
docker restart tc-exporter

# 删除容器
docker rm tc-exporter

# 删除镜像
docker rmi tc-exporter:latest
```

## 生产环境建议

### 安全配置

1. **使用非root用户**：镜像已配置非root用户运行
2. **网络隔离**：使用自定义Docker网络
3. **资源限制**：设置内存和CPU限制
4. **日志管理**：配置日志轮转和存储

### 监控配置

1. **Prometheus**：配置抓取TC Exporter指标
2. **Grafana**：创建监控仪表板
3. **告警规则**：设置关键指标告警

### 高可用性

对于生产环境，建议：

- 使用容器编排平台（Kubernetes）
- 配置多个实例负载均衡
- 设置健康检查自动重启

## 更新部署

### 更新镜像

```bash
# 停止现有容器
docker stop tc-exporter
docker rm tc-exporter

# 重新构建和部署
./deploy.sh
```

### 滚动更新

如果使用编排平台，可以配置滚动更新策略确保服务不中断。

## 支持与反馈

如果在部署过程中遇到问题，请：

1. 查看容器日志：`docker logs tc-exporter`
2. 检查健康状态：`curl http://localhost:9062/health`
3. 查看项目文档和Issue

## 相关链接

- [项目主页](https://gitee.com/openeuler/uos-tc-exporter)
- [使用文档](README.md)
- [配置说明](config/tc-exporter.yaml)
