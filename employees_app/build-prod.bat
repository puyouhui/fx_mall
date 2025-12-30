@echo off
chcp 65001 >nul
echo ========================================
echo 员工端App - 生产环境打包脚本
echo ========================================
echo.

REM 设置环境变量为生产环境
set APP_ENV=prod

echo [1/3] 清理之前的构建文件...
call flutter clean

echo.
echo [2/3] 获取依赖包...
call flutter pub get
if errorlevel 1 (
    echo 错误: 获取依赖包失败！
    pause
    exit /b 1
)

echo.
echo [3/3] 开始构建生产环境APK (APP_ENV=prod)...
call flutter build apk --release --dart-define=APP_ENV=prod
if errorlevel 1 (
    echo 错误: APK构建失败！
    pause
    exit /b 1
)

echo.
echo ========================================
echo 构建完成！
echo ========================================
echo APK文件位置: build\app\outputs\flutter-apk\app-release.apk
echo 生产环境API地址: https://mall.sscchh.com/api/mini
echo.
pause


