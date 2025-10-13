# 构建阶段
FROM golang:1.20-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache git make

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN make build

# 运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates

# 创建非root用户
RUN addgroup -S tc-exporter && adduser -S tc-exporter -G tc-exporter

# 创建工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/build/bin/tc-exporter .

# 复制配置文件
COPY config/tc-exporter.yaml /etc/uos-exporter/tc-exporter.yaml

# 创建必要的目录
RUN mkdir -p /var/log && \
    mkdir -p /etc/uos-exporter && \
    chown -R tc-exporter:tc-exporter /root/ /var/log /etc/uos-exporter

# 切换到非root用户
USER tc-exporter

# 暴露端口
EXPOSE 9062

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9062/health || exit 1

# 设置容器启动命令
CMD ["./tc-exporter", "--config.file=/etc/uos-exporter/tc-exporter.yaml"]
