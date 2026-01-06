@echo off
chcp 65001 >nul
echo ========================================
echo Admin App - Production Build Script
echo ========================================
echo.
echo Important: This script builds for production
echo Production server: https://mall.sscchh.com
echo Device must have internet access
echo.
echo For local development, use build-dev.bat
echo.

REM Set environment variable for production
set APP_ENV=prod

echo [1/3] Cleaning previous build files...
call flutter clean

echo.
echo [2/3] Getting dependencies...
call flutter pub get
if errorlevel 1 (
    echo Error: Failed to get dependencies!
    pause
    exit /b 1
)

echo.
echo [3/3] Building production APK (APP_ENV=prod)...
call flutter build apk --release --dart-define=APP_ENV=prod
if errorlevel 1 (
    echo Error: APK build failed!
    pause
    exit /b 1
)

echo.
echo ========================================
echo Build completed!
echo ========================================
echo APK location: build\app\outputs\flutter-apk\app-release.apk
echo Production API base: https://mall.sscchh.com/api_mall/mini
echo Example API path: https://mall.sscchh.com/api_mall/mini/admin/login
echo Note: Production uses /api_mall/mini (via Nginx proxy)
echo.
echo If API requests fail at runtime, check:
echo 1. APP_ENV=prod is correctly set (currently set)
echo 2. Backend service is running normally
echo 3. Network connection and HTTPS certificate
echo.
pause


