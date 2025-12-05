@echo off
setlocal enabledelayedexpansion

:: Lightspeed Operator Deploy Script
:: Builds and pushes operator Docker image to DigitalOcean registry

:: Registry configuration
set "REGISTRY=registry.digitalocean.com"
set "REPO=abrayall"
set "IMAGE=lightspeed-operator"

:: Work directory
set "SCRIPT_DIR=%~dp0"
set "WORK_DIR=%SCRIPT_DIR%build\work"
if not exist "%WORK_DIR%" mkdir "%WORK_DIR%"

echo ==============================================
echo [38;2;222;184;65mLightspeed Operator Deploy[0m
echo ==============================================
echo.

:: Get version from git
for /f "tokens=*" %%i in ('git describe --tags --match "v*.*.*" 2^>nul') do set "GIT_DESCRIBE=%%i"
if not defined GIT_DESCRIBE set "GIT_DESCRIBE=v0.1.0"

:: Parse version (simplified - just use git describe output minus 'v')
set "VERSION=%GIT_DESCRIBE:~1%"

:: Replace - with . for any commit count suffix, then back
:: This is simplified - just use the raw version
for /f "tokens=1-3 delims=-" %%a in ("%VERSION%") do (
    set "BASE=%%a"
    set "COMMITS=%%b"
    set "HASH=%%c"
)

if defined HASH (
    set "VERSION=%BASE%-%COMMITS%"
) else if defined COMMITS (
    :: Check if COMMITS is a number (commit count) or something else
    echo %COMMITS%| findstr /r "^[0-9]*$" >nul
    if !errorlevel! equ 0 (
        set "VERSION=%BASE%-%COMMITS%"
    ) else (
        set "VERSION=%BASE%"
    )
) else (
    set "VERSION=%BASE%"
)

:: Check for uncommitted changes
for /f "tokens=*" %%i in ('git status --porcelain 2^>nul') do set "HAS_CHANGES=1"
if defined HAS_CHANGES (
    for /f "tokens=*" %%i in ('powershell -NoProfile -Command "Get-Date -Format 'MMddHHmm'"') do set "TIMESTAMP=%%i"
    set "VERSION=!VERSION!-!TIMESTAMP!"
    echo [38;2;136;136;136mDetected uncommitted changes, appending timestamp[0m
)

:: Full image names
set "VERSION_TAG=%REGISTRY%/%REPO%/%IMAGE%:%VERSION%"
set "LATEST_TAG=%REGISTRY%/%REPO%/%IMAGE%:latest"

echo [38;2;59;130;246mVersion:[0m  %VERSION%
echo [38;2;59;130;246mRegistry:[0m %REGISTRY%/%REPO%
echo [38;2;59;130;246mImage:[0m    %IMAGE%
echo.

:: Generate Dockerfile
echo [38;2;59;130;246mGenerating Dockerfile...[0m
(
echo FROM golang:1.21-alpine AS builder
echo.
echo WORKDIR /app
echo.
echo # Copy go mod files
echo COPY go.mod go.sum ./
echo RUN go mod download
echo.
echo # Copy source
echo COPY . .
echo.
echo # Build operator
echo ARG VERSION=dev
echo RUN CGO_ENABLED=0 GOOS=linux go build \
echo     -ldflags "-X main.Version=${VERSION}" \
echo     -o /operator \
echo     ./platform/operator
echo.
echo # Runtime image
echo FROM alpine:3.19
echo.
echo RUN apk --no-cache add ca-certificates
echo.
echo COPY --from=builder /operator /usr/local/bin/operator
echo.
echo EXPOSE 8080
echo.
echo ENTRYPOINT ["operator"]
) > "%WORK_DIR%\Dockerfile"

:: Build the Docker image
echo [38;2;222;184;65mBuilding Docker image...[0m
echo.

docker build ^
    --platform linux/amd64 ^
    --build-arg VERSION="%VERSION%" ^
    -t "%VERSION_TAG%" ^
    -t "%LATEST_TAG%" ^
    -f "%WORK_DIR%\Dockerfile" ^
    .

if %errorlevel% neq 0 (
    echo Error: Docker build failed
    exit /b 1
)

echo.
echo [38;2;39;201;63mBuilt: %VERSION_TAG%[0m
echo.

:: Login to registry
if defined TOKEN (
    echo [38;2;59;130;246mLogging in to registry...[0m
    echo %TOKEN%| docker login %REGISTRY% --username %TOKEN% --password-stdin
    echo.
) else (
    echo [38;2;136;136;136mNo TOKEN env var set, assuming already logged in[0m
)

:: Push to registry
echo [38;2;222;184;65mPushing to registry...[0m
echo.

docker push "%VERSION_TAG%"
docker push "%LATEST_TAG%"

if %errorlevel% neq 0 (
    echo Error: Docker push failed
    exit /b 1
)

echo.
echo ==============================================
echo [38;2;39;201;63mDeploy Complete![0m
echo ==============================================
echo.
echo Pushed images:
echo   * %VERSION_TAG%
echo   * %LATEST_TAG%
echo.

endlocal
