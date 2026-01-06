@echo off
chcp 65001 >nul
echo ========================================
echo Admin App - Development Build Script
echo ========================================
echo.
echo Note: This script builds for device debugging
echo Uses local network IP to connect directly to backend
echo.
echo Please ensure:
echo 1. Phone and computer are on same local network
echo 2. Backend service is running at http://192.168.1.3:8082
echo 3. If IP differs, modify devBaseUrl in lib/utils/config.dart
echo.

REM Set environment variable for device debugging
set APP_ENV=device

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
echo [3/3] Building development APK (APP_ENV=device)...
call flutter build apk --release --dart-define=APP_ENV=device
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
echo Development API base: http://192.168.1.3:8082/api/mini
echo Example API path: http://192.168.1.3:8082/api/mini/admin/login
echo.
echo If phone cannot connect, check:
echo 1. Phone and computer are on same local network
echo 2. Computer firewall allows port 8082
echo 3. devBaseUrl IP in config.dart is correct
echo.
pause


