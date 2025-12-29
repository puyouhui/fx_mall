#!/bin/bash
# Docker 镜像构建和推送脚本
# 使用方法: ./build-and-push.sh [tag]
# 例如: ./build-and-push.sh latest 或 ./build-and-push.sh v0.0.6

set -e  # 遇到错误立即退出

# 配置
IMAGE_NAME="youhuipu/node-hiprint-transit"
PROXY_HOST="127.0.0.1"
PROXY_PORT="7897"
PROXY_URL="http://${PROXY_HOST}:${PROXY_PORT}"

# 生成时间戳标签（格式：YYYYMMDDHHmm，例如：202512291337）
generate_timestamp_tag() {
    date +"%Y%m%d%H%M"
}

# 获取标签
# 如果提供了参数，使用参数作为标签
# 如果没有提供参数，自动生成时间戳标签
if [ -n "$1" ] && [ "$1" != "-h" ] && [ "$1" != "--help" ] && [ "$1" != "help" ]; then
    TAG=$1
else
    TAG=$(generate_timestamp_tag)
fi

FULL_IMAGE_NAME="${IMAGE_NAME}:${TAG}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# 检查 Docker 是否运行
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        error "Docker 未运行，请先启动 Docker Desktop"
    fi
    success "Docker 运行正常"
}

# 拉取基础镜像（如果需要）
pull_base_image() {
    info "检查基础镜像 node:16-alpine..."
    
    if docker images node:16-alpine --format "{{.Repository}}:{{.Tag}}" | grep -q "node:16-alpine"; then
        success "基础镜像已存在"
    else
        warning "基础镜像不存在，正在拉取..."
        info "使用代理: ${PROXY_URL}"
        
        export HTTP_PROXY=${PROXY_URL}
        export HTTPS_PROXY=${PROXY_URL}
        export http_proxy=${PROXY_URL}
        export https_proxy=${PROXY_URL}
        
        if docker pull node:16-alpine; then
            success "基础镜像拉取成功"
        else
            error "基础镜像拉取失败，请检查网络或代理设置"
        fi
    fi
}

# 构建镜像
build_image() {
    info "开始构建镜像: ${FULL_IMAGE_NAME}"
    
    # 设置代理环境变量（用于构建过程中的网络请求）
    export HTTP_PROXY=${PROXY_URL}
    export HTTPS_PROXY=${PROXY_URL}
    export http_proxy=${PROXY_URL}
    export https_proxy=${PROXY_URL}
    
    # 构建镜像，传递代理参数
    # 只构建指定标签，不自动构建 latest
    if docker build \
        --build-arg HTTP_PROXY=${PROXY_URL} \
        --build-arg HTTPS_PROXY=${PROXY_URL} \
        --build-arg http_proxy=${PROXY_URL} \
        --build-arg https_proxy=${PROXY_URL} \
        -t ${FULL_IMAGE_NAME} \
        .; then
        success "镜像构建成功"
        
        # 显示镜像大小
        IMAGE_SIZE=$(docker images ${FULL_IMAGE_NAME} --format "{{.Size}}")
        info "镜像大小: ${IMAGE_SIZE}"
    else
        error "镜像构建失败"
    fi
}

# 推送镜像
push_image() {
    info "开始推送镜像到 Docker Hub..."
    info "镜像名称: ${FULL_IMAGE_NAME}"
    
    # 检查是否已登录 Docker Hub
    if ! docker info | grep -q "Username"; then
        warning "未检测到 Docker Hub 登录信息"
        info "请先登录 Docker Hub: docker login"
        read -p "是否现在登录? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker login
        else
            warning "跳过推送，镜像已构建但未推送"
            return
        fi
    fi
    
    # 推送镜像
    if docker push ${FULL_IMAGE_NAME}; then
        success "镜像推送成功: ${FULL_IMAGE_NAME}"
        
        # 如果标签不是 latest，询问是否也推送 latest
        if [ "${TAG}" != "latest" ]; then
            info "是否同时推送 latest 标签? (y/n)"
            read -p "" -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                info "标记为 latest..."
                docker tag ${FULL_IMAGE_NAME} ${IMAGE_NAME}:latest
                if docker push ${IMAGE_NAME}:latest; then
                    success "latest 标签推送成功"
                fi
            fi
        fi
    else
        error "镜像推送失败"
    fi
}

# 显示使用信息
show_usage() {
    echo "用法: $0 [tag]"
    echo ""
    echo "参数:"
    echo "  tag    镜像标签（可选，默认为时间戳格式：YYYYMMDDHHmm，例如：202512291337）"
    echo ""
    echo "示例:"
    echo "  $0                    # 构建并推送时间戳标签（自动生成，如：202512291337）"
    echo "  $0 latest             # 构建并推送 latest 标签"
    echo "  $0 v0.0.6             # 构建并推送 v0.0.6 标签"
    echo "  $0 202512291337       # 构建并推送指定时间戳标签"
    echo ""
    echo "注意:"
    echo "  - 不提供参数时，会自动生成当前时间的时间戳标签"
    echo "  - 时间戳格式：YYYYMMDDHHmm（年月日时分）"
    echo ""
    echo "环境变量:"
    echo "  HTTP_PROXY       HTTP 代理地址（默认: http://127.0.0.1:7897）"
    echo "  HTTPS_PROXY      HTTPS 代理地址（默认: http://127.0.0.1:7897）"
}

# 主函数
main() {
    # 检查参数（在设置 TAG 之前）
    if [ "$1" == "-h" ] || [ "$1" == "--help" ] || [ "$1" == "help" ]; then
        show_usage
        exit 0
    fi
    
    echo "=========================================="
    echo "  Docker 镜像构建和推送脚本"
    echo "=========================================="
    echo ""
    info "镜像名称: ${FULL_IMAGE_NAME}"
    info "代理地址: ${PROXY_URL}"
    echo ""
    
    # 执行步骤
    check_docker
    pull_base_image
    build_image
    push_image
    
    echo ""
    success "所有操作完成！"
    echo ""
    info "镜像信息:"
    docker images ${IMAGE_NAME} --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"
    echo ""
    info "使用镜像:"
    echo "  docker pull ${FULL_IMAGE_NAME}"
    echo "  docker run -d -p 17521:17521 ${FULL_IMAGE_NAME}"
}

# 运行主函数
main "$@"

