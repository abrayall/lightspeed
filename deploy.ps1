# Lightspeed Operator Deploy Script
# Builds and pushes operator Docker image to DigitalOcean registry

$ErrorActionPreference = "Stop"

# Registry configuration
$Registry = "registry.digitalocean.com"
$Repo = "lightspeed-images"
$Image = "lightspeed-operator"

# Colors
$Yellow = "`e[38;2;222;184;65m"
$Blue = "`e[38;2;59;130;246m"
$Green = "`e[38;2;39;201;63m"
$Gray = "`e[38;2;136;136;136m"
$NC = "`e[0m"

# Work directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$WorkDir = Join-Path $ScriptDir "build\work"
New-Item -ItemType Directory -Path $WorkDir -Force | Out-Null

Write-Host "=============================================="
Write-Host "${Yellow}Lightspeed Operator Deploy${NC}"
Write-Host "=============================================="
Write-Host ""

# Get version from git
try {
    $GitDescribe = git describe --tags --match "v*.*.*" 2>$null
    if (-not $GitDescribe) { $GitDescribe = "v0.1.0" }
} catch {
    $GitDescribe = "v0.1.0"
}

if ($GitDescribe -match '^v(\d+)\.(\d+)\.(\d+)(-(\d+)-g([0-9a-f]+))?$') {
    $Major = $Matches[1]
    $Minor = $Matches[2]
    $Maintenance = $Matches[3]
    $CommitCount = $Matches[5]

    if ($CommitCount) {
        $Version = "$Major.$Minor.$Maintenance-$CommitCount"
    } else {
        $Version = "$Major.$Minor.$Maintenance"
    }
} else {
    $Version = "0.1.0"
}

# Check for uncommitted changes
$Status = git status --porcelain 2>$null
if ($Status) {
    $Timestamp = Get-Date -Format "MMddHHmm"
    $Version = "$Version-$Timestamp"
    Write-Host "${Gray}Detected uncommitted changes, appending timestamp${NC}"
}

# Full image names
$VersionTag = "$Registry/$Repo/${Image}:$Version"
$LatestTag = "$Registry/$Repo/${Image}:latest"

Write-Host "${Blue}Version:${NC}  $Version"
Write-Host "${Blue}Registry:${NC} $Registry/$Repo"
Write-Host "${Blue}Image:${NC}    $Image"
Write-Host ""

# Generate Dockerfile
Write-Host "${Blue}Generating Dockerfile...${NC}"
$Dockerfile = @"
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build operator
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.Version=`${VERSION}" \
    -o /operator \
    ./platform/operator

# Runtime image
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

COPY --from=builder /operator /usr/local/bin/operator

EXPOSE 8080

ENTRYPOINT ["operator"]
"@
$Dockerfile | Out-File -FilePath "$WorkDir\Dockerfile" -Encoding UTF8

# Build the Docker image
Write-Host "${Yellow}Building Docker image...${NC}"
Write-Host ""

docker build `
    --platform linux/amd64 `
    --build-arg VERSION="$Version" `
    -t "$VersionTag" `
    -t "$LatestTag" `
    -f "$WorkDir\Dockerfile" `
    .

if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Docker build failed"
    exit 1
}

Write-Host ""
Write-Host "${Green}Built: $VersionTag${NC}"
Write-Host ""

# Login to registry
if ($env:TOKEN) {
    Write-Host "${Blue}Logging in to registry...${NC}"
    $env:TOKEN | docker login $Registry --username $env:TOKEN --password-stdin
    Write-Host ""
} else {
    Write-Host "${Gray}No TOKEN env var set, assuming already logged in${NC}"
}

# Push to registry
Write-Host "${Yellow}Pushing to registry...${NC}"
Write-Host ""

docker push "$VersionTag"
docker push "$LatestTag"

if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Docker push failed"
    exit 1
}

Write-Host ""
Write-Host "=============================================="
Write-Host "${Green}Deploy Complete!${NC}"
Write-Host "=============================================="
Write-Host ""
Write-Host "Pushed images:"
Write-Host "  * $VersionTag"
Write-Host "  * $LatestTag"
Write-Host ""
