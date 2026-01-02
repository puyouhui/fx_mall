#!/bin/bash
# Docker image build script for macOS/Linux
# Build frontend locally, then package into Docker image, optionally push to registry
# Default tag uses current datetime format (YYYYMMDDHHmm)

set -e  # Exit on error

IMAGE_NAME="mall_supplier"
PROXY_PORT="7897"
PUSH_IMAGE=false
REGISTRY=""
TAG=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --push)
            PUSH_IMAGE=true
            shift
            ;;
        --registry=*)
            REGISTRY="${1#*=}"
            shift
            ;;
        --*)
            echo "==> Unknown option: $1"
            exit 1
            ;;
        *)
            if [ -z "$TAG" ]; then
                TAG="$1"
            fi
            shift
            ;;
    esac
done

# If no tag specified, use current datetime format (YYYYMMDDHHmm)
if [ -z "$TAG" ]; then
    TAG=$(date +"%Y%m%d%H%M")
fi

# Build full image name
if [ -n "$REGISTRY" ]; then
    FULL_IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}:${TAG}"
elif [ "$PUSH_IMAGE" = true ]; then
    # Prompt for Docker Hub username
    echo "==> Please enter your Docker Hub username:"
    read -r DOCKER_USERNAME
    if [ -z "$DOCKER_USERNAME" ]; then
        echo "==> Error: Docker Hub username is required for pushing"
        echo "==> Usage: ./docker-build.sh --push"
        echo "==> Or specify registry: ./docker-build.sh --push --registry=your-registry/namespace"
        exit 1
    fi
    FULL_IMAGE_NAME="${DOCKER_USERNAME}/${IMAGE_NAME}:${TAG}"
    echo "==> Using Docker Hub username: $DOCKER_USERNAME"
else
    FULL_IMAGE_NAME="${IMAGE_NAME}:${TAG}"
fi

echo ""
echo "==> Step 1: Building frontend locally..."

# Check node_modules
if [ ! -d "node_modules" ]; then
    echo "==> Installing dependencies..."
    npm install
    if [ $? -ne 0 ]; then
        echo "==> npm install failed!"
        exit 1
    fi
fi

# Build frontend
npm run build

if [ $? -ne 0 ]; then
    echo "==> Build failed!"
    exit 1
fi

# Check if dist directory exists
if [ ! -d "dist" ]; then
    echo "==> Error: dist directory not found after build!"
    echo "==> Please check the build output above"
    exit 1
fi

echo "==> Frontend build successful, dist directory created"

echo ""
echo "==> Step 2: Building Docker image from dist directory..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "==> Error: Docker daemon is not running!"
    echo "==> Please start Docker Desktop and try again"
    echo ""
    echo "==> Or you can:"
    echo "    1. Keep the dist directory"
    echo "    2. Start Docker Desktop"
    echo "    3. Run: docker build -f Dockerfile -t $FULL_IMAGE_NAME ."
    exit 1
fi

# Pull base image with proxy
echo "==> Pulling base image with proxy (port $PROXY_PORT)..."
HTTP_PROXY="http://127.0.0.1:$PROXY_PORT" \
HTTPS_PROXY="http://127.0.0.1:$PROXY_PORT" \
docker pull nginx:alpine

# Build Docker image
echo "==> Building Docker image..."
docker build \
    --build-arg HTTP_PROXY="http://127.0.0.1:$PROXY_PORT" \
    --build-arg HTTPS_PROXY="http://127.0.0.1:$PROXY_PORT" \
    -f Dockerfile \
    -t "$FULL_IMAGE_NAME" \
    .

if [ $? -ne 0 ]; then
    echo "==> Build failed!"
    exit 1
fi

echo ""
echo "==> Build successful!"
echo "==> Image: $FULL_IMAGE_NAME"

# Push if needed
if [ "$PUSH_IMAGE" = true ]; then
    echo ""
    echo "==> Step 3: Pushing image to registry..."
    
    # Check if logged in (simple check - try to get docker info)
    if ! docker info | grep -q "Username"; then
        echo "==> Warning: Not logged in to Docker registry"
        echo "==> Please login first:"
        if [ -n "$REGISTRY" ]; then
            echo "    docker login $REGISTRY"
        else
            echo "    docker login"
        fi
        echo ""
        read -p "Do you want to login now? (y/n): " LOGIN_NOW
        if [ "$LOGIN_NOW" = "y" ] || [ "$LOGIN_NOW" = "Y" ]; then
            if [ -n "$REGISTRY" ]; then
                docker login "$REGISTRY"
            else
                docker login
            fi
        else
            echo "==> Skipping push. You can push manually later:"
            echo "    docker push $FULL_IMAGE_NAME"
            exit 0
        fi
    fi
    
    # Push image
    docker push "$FULL_IMAGE_NAME"
    
    if [ $? -ne 0 ]; then
        echo ""
        echo "==> Push failed!"
        exit 1
    fi
    
    echo ""
    echo "==> Push successful!"
    echo "==> Image available at: $FULL_IMAGE_NAME"
else
    echo ""
    echo "==> Next steps:"
    echo "    1. Test locally:"
    echo "       docker run -d -p 15174:5174 --name supplier-console-test $FULL_IMAGE_NAME"
    echo "       Then visit: http://localhost:15174"
    echo ""
    echo "    2. Push to registry:"
    echo "       ./docker-build.sh $TAG --push"
    if [ -n "$REGISTRY" ]; then
        echo "       Or: ./docker-build.sh $TAG --push --registry=$REGISTRY"
    fi
    echo ""
    echo "    3. Save image to file:"
    echo "       docker save $FULL_IMAGE_NAME > ${IMAGE_NAME}-${TAG}.tar"
    echo ""
    echo "    4. Load image on server:"
    echo "       docker load < ${IMAGE_NAME}-${TAG}.tar"
fi

