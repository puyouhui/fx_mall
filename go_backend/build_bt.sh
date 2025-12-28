#!/bin/bash
# Build script for Baota/Linux server (Mac version)
# Output directory: go_backend/dist/
#
# Usage:
#   cd go_backend
#   ./build_bt.sh          # Build linux/amd64 (default)
#   ./build_bt.sh arm64     # Build linux/arm64 (for ARM cloud servers)
#
# Troubleshooting:
#   If you get "cannot execute binary file" error:
#   1. Check server architecture: uname -m
#      - x86_64 or amd64 -> use default (amd64)
#      - aarch64 or arm64 -> use: ./build_bt.sh arm64
#   2. After uploading, set permission: chmod +x go_backend_linux_amd64

set -e

APP_NAME="go_backend"
OUT_DIR="dist"
mkdir -p "$OUT_DIR"

GOOS="linux"
GOARCH="amd64"

# If first parameter is arm64, build ARM64 version
if [ "$1" == "arm64" ]; then
    GOARCH="arm64"
    echo "==> Building for ARM64 architecture..."
elif [ "$1" == "help" ] || [ "$1" == "-h" ] || [ "$1" == "--help" ]; then
    echo "Usage: $0 [arm64]"
    echo ""
    echo "Options:"
    echo "  (no args)  Build for linux/amd64 (default)"
    echo "  arm64      Build for linux/arm64 (for ARM servers)"
    echo ""
    echo "To check your server architecture, run on server:"
    echo "  uname -m"
    echo ""
    echo "Common architectures:"
    echo "  x86_64, amd64 -> use default (amd64)"
    echo "  aarch64, arm64 -> use: $0 arm64"
    exit 0
fi

# 生成日期编号（年月日格式：YYYYMMDD）
DATE_TAG=$(date +"%Y%m%d")
OUT_BIN="${OUT_DIR}/${APP_NAME}_${GOOS}_${GOARCH}_${DATE_TAG}"

echo "==> Building ${APP_NAME} for ${GOOS}/${GOARCH} (date: ${DATE_TAG}) ..."

# Notes:
# - CGO_ENABLED=0: Generate static binary, reduce server runtime dependencies
# - -trimpath/-s -w: Reduce binary size
# - Must set GOOS and GOARCH as environment variables before go build
export CGO_ENABLED=0
export GOOS
export GOARCH
go build -trimpath -ldflags="-s -w" -o "$OUT_BIN" ./cmd

if [ $? -eq 0 ]; then
    # Get file size
    FILE_SIZE=$(ls -lh "$OUT_BIN" | awk '{print $5}')
    echo "==> Build successful!"
    echo "==> Output: ${OUT_BIN} (${FILE_SIZE})"
    echo ""
    echo "==> Next steps:"
    echo "   1. Upload ${OUT_BIN} to your server"
    echo "   2. Set executable permission: chmod +x ${OUT_BIN}"
    echo "   3. If you get 'cannot execute binary file' error:"
    echo "      - Check server architecture: uname -m"
    echo "      - If server is ARM64, rebuild with: ./build_bt.sh arm64"
    echo ""
    echo "==> File naming:"
    echo "   Format: ${APP_NAME}_${GOOS}_${GOARCH}_YYYYMMDD"
    echo "   Example: ${OUT_BIN}"
    echo "   This allows easy rollback by date"
else
    echo "==> Build failed!"
    exit 1
fi

