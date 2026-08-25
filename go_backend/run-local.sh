#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ENV_FILE="$SCRIPT_DIR/.env.local"

if [ ! -f "$ENV_FILE" ]; then
  echo "缺少 $ENV_FILE，请先创建本地环境配置。" >&2
  exit 1
fi

set -a
. "$ENV_FILE"
set +a

cd "$SCRIPT_DIR"
exec go run ./cmd
