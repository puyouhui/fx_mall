#!/bin/sh

set -eu

PROJECT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
PIDS=""

cleanup() {
  trap - INT TERM EXIT
  if [ -n "$PIDS" ]; then
    kill $PIDS 2>/dev/null || true
    wait $PIDS 2>/dev/null || true
  fi
}

trap cleanup INT TERM EXIT

for command_name in go node npm docker; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "缺少命令：$command_name" >&2
    exit 1
  fi
done

if ! nc -z 127.0.0.1 3306 >/dev/null 2>&1; then
  echo "MySQL 未监听 127.0.0.1:3306，请先启动 mysql-container。" >&2
  exit 1
fi

if ! nc -z 127.0.0.1 9000 >/dev/null 2>&1; then
  echo "启动本地 MinIO..."
  docker compose -f "$PROJECT_DIR/docker-compose.local.yml" up -d minio
  attempt=0
  while ! nc -z 127.0.0.1 9000 >/dev/null 2>&1; do
    attempt=$((attempt + 1))
    if [ "$attempt" -ge 30 ]; then
      echo "等待本地 MinIO 启动超时。" >&2
      exit 1
    fi
    sleep 1
  done
fi

if [ ! -d "$PROJECT_DIR/admin_console/node_modules" ] || [ ! -d "$PROJECT_DIR/supplier_console/node_modules" ]; then
  echo "前端依赖尚未安装，请先在 admin_console 和 supplier_console 中执行 npm install。" >&2
  exit 1
fi

echo "启动后端：http://localhost:8082"
(cd "$PROJECT_DIR/go_backend" && exec ./run-local.sh) &
PIDS="$PIDS $!"

echo "启动运营后台：http://localhost:5173/admin/"
(cd "$PROJECT_DIR/admin_console" && exec npm run dev -- --host 0.0.0.0) &
PIDS="$PIDS $!"

echo "启动供应商后台：http://localhost:21321/supplier/"
(cd "$PROJECT_DIR/supplier_console" && exec npm run dev -- --host 0.0.0.0) &
PIDS="$PIDS $!"

echo "按 Ctrl+C 停止全部本地服务。"
wait
