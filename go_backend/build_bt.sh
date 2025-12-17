#!/usr/bin/env bash
set -euo pipefail

# 用于“宝塔/Linux服务器”运行的可执行文件编译脚本
# 默认产物输出到：go_backend/dist/
#
# 用法（在本机/macOS 或 Linux 都可交叉编译）：
#   cd go_backend
#   bash build_bt.sh
#
# 可选：
#   GOARCH=arm64 bash build_bt.sh   # 编译 linux/arm64（例如部分 ARM 云服务器）
#

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

APP_NAME="go_backend"
OUT_DIR="${ROOT_DIR}/dist"
mkdir -p "$OUT_DIR"

GOOS="${GOOS:-linux}"
GOARCH="${GOARCH:-amd64}"

OUT_BIN="${OUT_DIR}/${APP_NAME}_${GOOS}_${GOARCH}"

echo "==> Building ${APP_NAME} for ${GOOS}/${GOARCH} ..."

# 说明：
# - CGO_ENABLED=0：尽量生成静态二进制，降低服务器运行环境依赖
# - -trimpath/-s -w：减小体积
CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" \
  go build -trimpath -ldflags="-s -w" -o "$OUT_BIN" ./cmd

echo "==> Done: ${OUT_BIN}"
echo "==> Tip: 上传到服务器后，赋予可执行权限：chmod +x ${APP_NAME}"


