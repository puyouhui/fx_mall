#!/bin/bash

echo "========================================"
echo "获取Android应用SHA1值"
echo "========================================"
echo ""

# 尝试多个可能的debug keystore路径
KEYSTORE_PATH1="$HOME/.android/debug.keystore"
KEYSTORE_PATH2="$HOME/Library/Android/sdk/.android/debug.keystore"

echo "正在查找debug keystore文件..."
echo ""

if [ -f "$KEYSTORE_PATH1" ]; then
    echo "找到keystore: $KEYSTORE_PATH1"
    echo ""
    echo "正在获取SHA1值..."
    echo "========================================"
    keytool -list -v -keystore "$KEYSTORE_PATH1" -storepass android -keypass android | grep -i "SHA1"
    echo "========================================"
    exit 0
fi

if [ -f "$KEYSTORE_PATH2" ]; then
    echo "找到keystore: $KEYSTORE_PATH2"
    echo ""
    echo "正在获取SHA1值..."
    echo "========================================"
    keytool -list -v -keystore "$KEYSTORE_PATH2" -storepass android -keypass android | grep -i "SHA1"
    echo "========================================"
    exit 0
fi

echo "未找到debug keystore文件！"
echo ""
echo "请尝试以下方法："
echo "1. 先运行一次 flutter run 或 flutter build apk --debug 来生成keystore"
echo "2. 或者手动运行以下命令（替换路径）："
echo "   keytool -list -v -keystore \"你的keystore路径\" -storepass android -keypass android"
echo ""
echo "如果使用自定义keystore，请替换命令中的路径和密码"
echo ""

