@echo off
REM Docker image build script for Windows
REM Build frontend locally, then package into Docker image, optionally push to registry
REM Default tag uses current datetime format (YYYYMMDDHHmm)

setlocal enabledelayedexpansion

set "IMAGE_NAME=mall_admin"
set "PROXY_PORT=7897"
set "PUSH_IMAGE=false"
set "REGISTRY="
set "TAG="

REM Parse arguments
:parse_args
if "%~1"=="" goto args_done
set "arg=%~1"
if /i "!arg!"=="--push" (
    set "PUSH_IMAGE=true"
    shift
    goto parse_args
)
REM Check if --registry=value format
set "arg_prefix=!arg:~0,11!"
if /i "!arg_prefix!"=="--registry=" (
    set "REGISTRY=!arg:~11!"
    shift
    goto parse_args
)
REM Check for other -- options
if "!arg:~0,2!"=="--" (
    echo ==^> Unknown option: %~1
    exit /b 1
)
if "!TAG!"=="" (
    set "TAG=%~1"
)
shift
goto parse_args

:args_done

REM If no tag specified, use current datetime format (YYYYMMDDHHmm)
if "!TAG!"=="" (
    REM Get current datetime and format as YYYYMMDDHHmm
    for /f "tokens=2 delims==" %%I in ('wmic os get localdatetime /value') do set datetime=%%I
    set "TAG=!datetime:~0,12!"
)

REM Build full image name
if not "!REGISTRY!"=="" (
    set "FULL_IMAGE_NAME=!REGISTRY!/!IMAGE_NAME!:!TAG!"
) else if "!PUSH_IMAGE!"=="true" (
    REM Prompt for Docker Hub username
    echo ==^> Please enter your Docker Hub username:
    set /p DOCKER_USERNAME=
    if "!DOCKER_USERNAME!"=="" (
        echo ==^> Error: Docker Hub username is required for pushing
        echo ==^> Usage: docker-build.bat --push
        echo ==^> Or specify registry: docker-build.bat --push --registry=your-registry/namespace
        exit /b 1
    )
    set "FULL_IMAGE_NAME=!DOCKER_USERNAME!/!IMAGE_NAME!:!TAG!"
    echo ==^> Using Docker Hub username: !DOCKER_USERNAME!
) else (
    set "FULL_IMAGE_NAME=!IMAGE_NAME!:!TAG!"
)

echo.
echo ==^> Step 1: Building frontend locally...

REM Check node_modules
if not exist "node_modules" (
    echo ==^> Installing dependencies...
    call npm install
    if errorlevel 1 (
        echo ==^> npm install failed!
        exit /b 1
    )
)

REM Build frontend
call npm run build

if errorlevel 1 (
    echo ==^> Build failed!
    exit /b 1
)

REM Check if dist directory exists
if not exist "dist" (
    echo ==^> Error: dist directory not found after build!
    echo ==^> Please check the build output above
    exit /b 1
)

echo ==^> Frontend build successful, dist directory created

echo.
echo ==^> Step 2: Building Docker image from dist directory...

REM Check if Docker is running
docker info >nul 2>&1
if errorlevel 1 (
    echo ==^> Error: Docker daemon is not running!
    echo ==^> Please start Docker Desktop and try again
    echo.
    echo ==^> Or you can:
    echo    1. Keep the dist directory
    echo    2. Start Docker Desktop
    echo    3. Run: docker build -f Dockerfile -t !FULL_IMAGE_NAME! .
    exit /b 1
)

REM Pull base image with proxy
echo ==^> Pulling base image with proxy (port !PROXY_PORT!)...
set "HTTP_PROXY=http://127.0.0.1:!PROXY_PORT!"
set "HTTPS_PROXY=http://127.0.0.1:!PROXY_PORT!"
docker pull nginx:alpine

REM Build Docker image
echo ==^> Building Docker image...
docker build --build-arg HTTP_PROXY=http://127.0.0.1:!PROXY_PORT! --build-arg HTTPS_PROXY=http://127.0.0.1:!PROXY_PORT! -f Dockerfile -t "!FULL_IMAGE_NAME!" .

if errorlevel 1 (
    echo ==^> Build failed!
    exit /b 1
)

echo.
echo ==^> Build successful!
echo ==^> Image: !FULL_IMAGE_NAME!

REM Push if needed
if "!PUSH_IMAGE!"=="true" (
    echo.
    echo ==^> Step 3: Pushing image to registry...
    
    REM Check if logged in
    docker info | findstr /i "username" >nul 2>&1
    if errorlevel 1 (
        echo ==^> Warning: Not logged in to Docker registry
        echo ==^> Please login first:
        if not "!REGISTRY!"=="" (
            echo    docker login !REGISTRY!
        ) else (
            echo    docker login
        )
        echo.
        set /p LOGIN_NOW="Do you want to login now? (y/n): "
        if /i "!LOGIN_NOW!"=="y" (
            if not "!REGISTRY!"=="" (
                docker login "!REGISTRY!"
            ) else (
                docker login
            )
        ) else (
            echo ==^> Skipping push. You can push manually later:
            echo    docker push !FULL_IMAGE_NAME!
            exit /b 0
        )
    )
    
    REM Push image
    docker push "!FULL_IMAGE_NAME!"
    
    if errorlevel 1 (
        echo.
        echo ==^> Push failed!
        exit /b 1
    )
    
    echo.
    echo ==^> Push successful!
    echo ==^> Image available at: !FULL_IMAGE_NAME!
) else (
    echo.
    echo ==^> Next steps:
    echo    1. Test locally:
    echo       docker run -d -p 15173:5173 --name admin-console-test !FULL_IMAGE_NAME!
    echo       Then visit: http://localhost:15173
    echo.
    echo    2. Push to registry:
    echo       docker-build.bat !TAG! --push
    if not "!REGISTRY!"=="" (
        echo       Or: docker-build.bat !TAG! --push --registry=!REGISTRY!
    )
    echo.
    echo    3. Save image to file:
    echo       docker save !FULL_IMAGE_NAME! ^> !IMAGE_NAME!-!TAG!.tar
    echo.
    echo    4. Load image on server:
    echo       docker load ^< !IMAGE_NAME!-!TAG!.tar
)

endlocal
