#!/bin/bash

echo "========================================"
echo "员工端App - 生产环境打包脚本"
echo "========================================"
echo ""

# 检测 Flutter 命令
FLUTTER_CMD="flutter"
if ! command -v flutter &> /dev/null; then
    # 尝试常见的 Flutter 安装路径
    if [ -f "$HOME/flutter/bin/flutter" ]; then
        FLUTTER_CMD="$HOME/flutter/bin/flutter"
    elif [ -f "/usr/local/bin/flutter" ]; then
        FLUTTER_CMD="/usr/local/bin/flutter"
    elif [ -f "/opt/flutter/bin/flutter" ]; then
        FLUTTER_CMD="/opt/flutter/bin/flutter"
    else
        echo "错误: 未找到 Flutter 命令！"
        echo "请确保 Flutter 已安装并添加到 PATH 环境变量中。"
        echo "或者修改此脚本，设置 FLUTTER_CMD 变量为 Flutter 的完整路径。"
        exit 1
    fi
fi

echo "使用 Flutter: $FLUTTER_CMD"
echo ""

# 设置环境变量为生产环境
export APP_ENV=prod

echo "[1/3] 清理之前的构建文件..."
$FLUTTER_CMD clean
if [ $? -ne 0 ]; then
    echo "错误: 清理构建文件失败！"
    exit 1
fi

echo ""
echo "[2/3] 获取依赖包..."
$FLUTTER_CMD pub get
if [ $? -ne 0 ]; then
    echo "错误: 获取依赖包失败！"
    exit 1
fi

echo ""
echo "[3/3] 开始构建生产环境APK (APP_ENV=prod)..."
$FLUTTER_CMD build apk --release --dart-define=APP_ENV=prod
if [ $? -ne 0 ]; then
    echo "错误: APK构建失败！"
    exit 1
fi

echo ""
echo "========================================"
echo "构建完成！"
echo "========================================"
echo "APK文件位置: build/app/outputs/flutter-apk/app-release.apk"
echo "生产环境API地址: https://mall.sscchh.com/api/mini"
echo ""
