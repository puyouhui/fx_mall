#!/bin/bash
# 管理员App - 开发环境打包脚本（真机调试）(macOS/Linux)

set -e  # Exit on error

echo "========================================"
echo "管理员App - 开发环境打包脚本（真机调试）"
echo "========================================"
echo ""
echo "提示: 此脚本用于真机调试，使用局域网IP直接连接后端"
echo "请确保："
echo "1. 手机和电脑在同一局域网"
echo "2. 后端服务在 http://192.168.1.3:8082 运行"
echo "3. 如果IP不同，请修改 lib/utils/config.dart 中的 devBaseUrl"
echo ""

# 检测 Flutter 命令
FLUTTER_CMD="flutter"

# 如果 flutter 不在 PATH 中，尝试常见路径
if ! command -v flutter &> /dev/null; then
    # 检查常见的 Flutter 安装位置
    if [ -f "$HOME/flutter/bin/flutter" ]; then
        FLUTTER_CMD="$HOME/flutter/bin/flutter"
        export PATH="$HOME/flutter/bin:$PATH"
    elif [ -f "/usr/local/bin/flutter" ]; then
        FLUTTER_CMD="/usr/local/bin/flutter"
    elif [ -f "/opt/flutter/bin/flutter" ]; then
        FLUTTER_CMD="/opt/flutter/bin/flutter"
        export PATH="/opt/flutter/bin:$PATH"
    else
        echo "错误: 找不到 Flutter 命令！"
        echo "请确保 Flutter 已安装并在 PATH 中"
        exit 1
    fi
    echo "使用 Flutter: $FLUTTER_CMD"
    echo ""
fi

# 设置环境变量为真机调试环境
export APP_ENV=device

echo "[1/3] 清理之前的构建文件..."
$FLUTTER_CMD clean

echo ""
echo "[2/3] 获取依赖包..."
$FLUTTER_CMD pub get
if [ $? -ne 0 ]; then
    echo "错误: 获取依赖包失败！"
    exit 1
fi

echo ""
echo "[3/3] 开始构建开发环境APK (APP_ENV=device)..."
$FLUTTER_CMD build apk --release --dart-define=APP_ENV=device
if [ $? -ne 0 ]; then
    echo "错误: APK构建失败！"
    exit 1
fi

echo ""
echo "========================================"
echo "构建完成！"
echo "========================================"
echo "APK文件位置: build/app/outputs/flutter-apk/app-release.apk"
echo "开发环境API基础地址: http://192.168.1.3:8082/api/mini"
echo "管理员API路径示例: http://192.168.1.3:8082/api/mini/admin/login"
echo ""
echo "注意: 如果手机无法连接，请检查："
echo "1. 手机和电脑是否在同一局域网"
echo "2. 电脑防火墙是否允许8082端口"
echo "3. config.dart 中的 devBaseUrl IP是否正确"
echo ""

