#!/bin/bash
# Docker 镜像构建脚本
# 先本地构建前端，再打包成 Docker 镜像，并可选择推送到镜像仓库
# 默认 tag 使用当前时间的年月日时分格式 (YYYYMMDDHHmm)
#
# Usage:
#   cd admin_console
#   ./docker-build.sh [tag] [--push] [--registry=registry_url]
#
# Examples:
#   ./docker-build.sh                    # 只构建，tag 自动使用当前时间 (如: 202412282345)
#   ./docker-build.sh --push             # 构建并推送到 Docker Hub，tag 使用当前时间（需要先 docker login）
#   ./docker-build.sh latest --push       # 构建并推送到 Docker Hub，使用指定 tag
#   ./docker-build.sh --push --registry=registry.cn-hangzhou.aliyuncs.com/your-namespace/mall_admin

set -e

IMAGE_NAME="mall_admin"
PROXY_PORT="7897"
PUSH_IMAGE=false
REGISTRY=""

# 解析参数
TAG=""
for arg in "$@"; do
    case $arg in
        --push)
            PUSH_IMAGE=true
            ;;
        --registry=*)
            REGISTRY="${arg#*=}"
            ;;
        *)
            if [ -z "$TAG" ] && [[ ! "$arg" =~ ^-- ]]; then
                TAG="$arg"
            fi
            ;;
    esac
done

# 如果没有指定 tag，使用当前时间的年月日时分格式 (YYYYMMDDHHmm)
if [ -z "$TAG" ]; then
    TAG=$(date +"%Y%m%d%H%M")
fi

# 构建完整镜像名
# 如果指定了仓库，使用仓库地址
# 如果没有指定仓库但需要推送，尝试获取 Docker Hub 用户名
if [ -n "$REGISTRY" ]; then
    FULL_IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}:${TAG}"
elif [ "$PUSH_IMAGE" = true ]; then
    # 尝试从 Docker 配置中获取用户名
    DOCKER_USERNAME=$(docker info 2>/dev/null | grep -i "username" | awk '{print $2}' | head -1)
    
    # 如果无法从配置获取，尝试从已登录的账户获取
    if [ -z "$DOCKER_USERNAME" ]; then
        # 尝试从 ~/.docker/config.json 获取
        if [ -f ~/.docker/config.json ]; then
            DOCKER_USERNAME=$(cat ~/.docker/config.json | grep -o '"auths":{[^}]*"[^"]*":{[^}]*}' | grep -o '"[^"]*":' | head -1 | tr -d '":' | grep -v "docker.io" | head -1)
        fi
    fi
    
    # 如果还是无法获取，提示用户输入
    if [ -z "$DOCKER_USERNAME" ]; then
        echo "==> Docker Hub username not found automatically"
        read -p "==> Please enter your Docker Hub username: " DOCKER_USERNAME
        if [ -z "$DOCKER_USERNAME" ]; then
            echo "==> Error: Docker Hub username is required for pushing"
            echo "==> Usage: ./docker-build.sh --push"
            echo "==> Or specify registry: ./docker-build.sh --push --registry=your-registry/namespace"
            exit 1
        fi
    fi
    
    FULL_IMAGE_NAME="${DOCKER_USERNAME}/${IMAGE_NAME}:${TAG}"
    echo "==> Using Docker Hub username: ${DOCKER_USERNAME}"
else
    FULL_IMAGE_NAME="${IMAGE_NAME}:${TAG}"
fi

echo "==> Step 1: Building frontend locally..."

# 检查 node_modules
if [ ! -d "node_modules" ]; then
    echo "==> Installing dependencies..."
    npm install
fi

# 构建前端
npm run build

if [ $? -ne 0 ]; then
    echo "==> Build failed!"
    exit 1
fi

# 检查 dist 目录是否存在
if [ ! -d "dist" ]; then
    echo "==> Error: dist directory not found after build!"
    echo "==> Please check the build output above"
    exit 1
fi

echo "==> Frontend build successful, dist directory created"

echo ""
echo "==> Step 2: Building Docker image from dist directory..."

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo "==> Error: Docker daemon is not running!"
    echo "==> Please start Docker Desktop and try again"
    echo ""
    echo "==> Or you can:"
    echo "   1. Keep the dist directory"
    echo "   2. Start Docker Desktop"
    echo "   3. Run: docker build -f Dockerfile -t ${FULL_IMAGE_NAME} ."
    exit 1
fi

# 先拉取基础镜像（使用代理环境变量）
echo "==> Pulling base image with proxy (port ${PROXY_PORT})..."
export HTTP_PROXY=http://127.0.0.1:${PROXY_PORT}
export HTTPS_PROXY=http://127.0.0.1:${PROXY_PORT}
docker pull nginx:alpine

# 使用 Dockerfile 构建
echo "==> Building Docker image..."
docker build \
    --build-arg HTTP_PROXY=http://127.0.0.1:${PROXY_PORT} \
    --build-arg HTTPS_PROXY=http://127.0.0.1:${PROXY_PORT} \
    -f Dockerfile -t "${FULL_IMAGE_NAME}" .

if [ $? -eq 0 ]; then
    # 获取镜像大小
    IMAGE_SIZE=$(docker images "${FULL_IMAGE_NAME}" --format "{{.Size}}" 2>/dev/null || echo "unknown")
    echo ""
    echo "==> Build successful!"
    echo "==> Image: ${FULL_IMAGE_NAME} (${IMAGE_SIZE})"
    
    # 如果需要推送
    if [ "$PUSH_IMAGE" = true ]; then
        echo ""
        echo "==> Step 3: Pushing image to registry..."
        
        # 检查是否已登录
        if ! docker info | grep -q "Username"; then
            echo "==> Warning: Not logged in to Docker registry"
            echo "==> Please login first:"
            if [ -n "$REGISTRY" ]; then
                echo "   docker login ${REGISTRY}"
            else
                echo "   docker login"
            fi
            read -p "==> Do you want to login now? (y/n) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                if [ -n "$REGISTRY" ]; then
                    docker login "${REGISTRY}"
                else
                    docker login
                fi
            else
                echo "==> Skipping push. You can push manually later:"
                echo "   docker push ${FULL_IMAGE_NAME}"
                exit 0
            fi
        fi
        
        # 推送镜像
        docker push "${FULL_IMAGE_NAME}"
        
        if [ $? -eq 0 ]; then
            echo ""
            echo "==> Push successful!"
            echo "==> Image available at: ${FULL_IMAGE_NAME}"
        else
            echo ""
            echo "==> Push failed!"
            exit 1
        fi
    else
        echo ""
        echo "==> Next steps:"
        echo "   1. Test locally:"
        echo "      docker run -d -p 15173:5173 --name admin-console-test ${FULL_IMAGE_NAME}"
        echo "      Then visit: http://localhost:15173"
        echo ""
        echo "   2. Push to registry:"
        echo "      ./docker-build.sh ${TAG} --push"
        if [ -n "$REGISTRY" ]; then
            echo "      Or: ./docker-build.sh ${TAG} --push --registry=${REGISTRY}"
        fi
        echo ""
        echo "   3. Save image to file:"
        echo "      docker save ${FULL_IMAGE_NAME} | gzip > ${IMAGE_NAME}-${TAG}.tar.gz"
        echo ""
        echo "   4. Load image on server:"
        echo "      gunzip -c ${IMAGE_NAME}-${TAG}.tar.gz | docker load"
    fi
else
    echo "==> Build failed!"
    exit 1
fi

