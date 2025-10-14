#!/bin/bash

# TC Exporter 部署脚本
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi

    if ! docker info &> /dev/null; then
        log_error "Docker 守护进程未运行，请启动 Docker"
        exit 1
    fi

    log_info "Docker 检查通过"
}

# 构建Docker镜像
build_image() {
    local image_name="tc-exporter:latest"

    log_info "开始构建 Docker 镜像: $image_name"

    if docker build -t $image_name .; then
        log_info "Docker 镜像构建成功"
    else
        log_error "Docker 镜像构建失败"
        exit 1
    fi
}

# 运行容器
run_container() {
    local image_name="tc-exporter:latest"
    local container_name="tc-exporter"
    local port=${1:-9062}

    # 检查容器是否已存在
    if docker ps -a --format "table {{.Names}}" | grep -q "^${container_name}$"; then
        log_warn "容器 $container_name 已存在，正在停止并删除..."
        docker stop $container_name > /dev/null 2>&1 || true
        docker rm $container_name > /dev/null 2>&1 || true
    fi

    log_info "启动容器 $container_name，端口映射: $port:9062"

    # 运行容器
    if docker run -d \
        --name $container_name \
        --restart unless-stopped \
        --privileged \
        --cap-add=NET_ADMIN \
        -p $port:9062 \
        -v /var/log:/var/log \
        $image_name; then
        log_info "容器启动成功"
    else
        log_error "容器启动失败"
        exit 1
    fi
}

# 健康检查
health_check() {
    local port=${1:-9062}
    local max_attempts=30
    local attempt=1

    log_info "等待服务启动..."

    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:$port/health > /dev/null 2>&1; then
            log_info "服务健康检查通过"
            return 0
        fi

        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done

    log_error "服务健康检查失败，请检查容器日志"
    docker logs tc-exporter
    return 1
}

# 显示使用信息
show_usage() {
    echo "TC Exporter 部署脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -p, --port PORT     指定映射端口 (默认: 9062)"
    echo "  -b, --build-only    仅构建镜像，不运行容器"
    echo "  -h, --help         显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0                    # 使用默认端口 9062 部署"
    echo "  $0 -p 8080            # 使用端口 8080 部署"
    echo "  $0 --build-only       # 仅构建镜像"
}

# 主函数
main() {
    local port=9062
    local build_only=false

    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -p|--port)
                port="$2"
                shift 2
                ;;
            -b|--build-only)
                build_only=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                show_usage
                exit 1
                ;;
        esac
    done

    # 验证端口号
    if ! [[ $port =~ ^[0-9]+$ ]] || [ $port -lt 1 ] || [ $port -gt 65535 ]; then
        log_error "无效的端口号: $port"
        exit 1
    fi

    log_info "开始部署 TC Exporter"

    # 检查Docker
    check_docker

    # 构建镜像
    build_image

    if [ "$build_only" = true ]; then
        log_info "构建完成，跳过容器运行"
        exit 0
    fi

    # 运行容器
    run_container $port

    # 健康检查
    if health_check $port; then
        log_info "TC Exporter 部署成功!"
        log_info "访问地址: http://localhost:$port"
        log_info "健康检查: http://localhost:$port/health"
        log_info "指标端点: http://localhost:$port/metrics"
    else
        log_error "部署失败"
        exit 1
    fi
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
