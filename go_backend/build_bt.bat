@echo off
REM Build script for Baota/Linux server (Windows version)
REM Output directory: go_backend/dist/
REM
REM Usage:
REM   cd go_backend
REM   build_bt.bat
REM
REM Optional:
REM   build_bt.bat arm64   # Build linux/arm64 (for ARM cloud servers)

setlocal

set APP_NAME=go_backend
set OUT_DIR=dist
if not exist "%OUT_DIR%" mkdir "%OUT_DIR%"

set GOOS=linux
set GOARCH=amd64

REM If first parameter is arm64, build ARM64 version
if "%1"=="arm64" (
    set GOARCH=arm64
)

set OUT_BIN=%OUT_DIR%\%APP_NAME%_%GOOS%_%GOARCH%

echo ==^> Building %APP_NAME% for %GOOS%/%GOARCH% ...

REM Notes:
REM - CGO_ENABLED=0: Generate static binary, reduce server runtime dependencies
REM - -trimpath/-s -w: Reduce binary size
set CGO_ENABLED=0
go build -trimpath -ldflags="-s -w" -o "%OUT_BIN%" ./cmd

if %ERRORLEVEL% EQU 0 (
    echo ==^> Done: %OUT_BIN%
    echo ==^> Tip: After uploading to server, set executable permission: chmod +x %APP_NAME%
) else (
    echo ==^> Build failed!
    exit /b 1
)

endlocal

