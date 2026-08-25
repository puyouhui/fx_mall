#!/bin/sh

set -eu

PROJECT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
APP_NAME=${1:-}

case "$APP_NAME" in
  employees_app|distribution_app|super_app)
    ;;
  *)
    echo "用法：$0 employees_app|distribution_app|super_app" >&2
    exit 1
    ;;
esac

if ! command -v flutter >/dev/null 2>&1; then
  echo "缺少 flutter 命令。" >&2
  exit 1
fi

LAN_IP=${DEV_HOST_IP:-$(ipconfig getifaddr en0 2>/dev/null || true)}
if [ -z "$LAN_IP" ]; then
  echo "无法自动获取局域网 IP，请设置 DEV_HOST_IP 后重试。" >&2
  exit 1
fi

DEV_BASE_URL="http://$LAN_IP:8082"
echo "启动 $APP_NAME，API：$DEV_BASE_URL/api/mini"

cd "$PROJECT_DIR/$APP_NAME"
exec flutter run \
  --dart-define=APP_ENV=device \
  --dart-define=DEV_BASE_URL="$DEV_BASE_URL"
