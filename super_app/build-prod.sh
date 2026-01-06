#!/bin/bash
# 管理员App - 生产环境打包脚本 (macOS/Linux)

set -e  # Exit on error

echo "========================================"
echo "管理员App - 生产环境打包脚本"
echo "========================================"
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
        echo "请确保 Flutter 已安装并在 PATH 中，或者设置 FLUTTER_HOME 环境变量"
        echo ""
        echo "您可以："
        echo "1. 将 Flutter 添加到 PATH: export PATH=\"\$PATH:\$HOME/flutter/bin\""
        echo "2. 或者在脚本中设置 FLUTTER_HOME 环境变量"
        exit 1
    fi
    echo "使用 Flutter: $FLUTTER_CMD"
    echo ""
fi

# 设置环境变量为生产环境
export APP_ENV=prod

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
echo "生产环境API基础地址: https://mall.sscchh.com/api_mall/mini"
echo "管理员API路径示例: https://mall.sscchh.com/api_mall/mini/admin/login"
echo "注意: 生产环境使用 /api_mall/mini (通过 Nginx 代理)"
echo ""
echo "提示: 如果运行时API请求失败，请检查："
echo "1. 确保APP_ENV=prod已正确传递（当前已设置）"
echo "2. 检查后端服务是否正常运行"
echo "3. 检查网络连接和HTTPS证书"
echo ""

