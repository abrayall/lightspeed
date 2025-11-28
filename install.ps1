# Lightspeed Install Script for Windows
# Builds locally if in git repo, otherwise downloads from GitHub releases

$ErrorActionPreference = "Stop"

$Repo = "abrayall/lightspeed"
$InstallDir = "$env:USERPROFILE\bin"

# Colors
$Blue = "`e[38;2;59;130;246m"
$White = "`e[97m"
$Green = "`e[92m"
$Red = "`e[91m"
$NC = "`e[0m"

# Detect architecture
$Arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }

# Check if we're in the lightspeed repo
function Test-InRepo {
    if (Test-Path "go.mod") {
        if ((Get-Content "go.mod" -Raw) -match "lightspeed") {
            if (Test-Path "build.bat") {
                return $true
            }
        }
    }
    return $false
}

# Build from source
function Build-Local {
    & .\build.bat

    $Binary = Get-ChildItem "build\lightspeed-*-windows-$Arch.exe" | Select-Object -First 1
    if (-not $Binary) {
        Write-Host "${Red}Error: No binary found for windows-$Arch${NC}"
        exit 1
    }

    return $Binary.FullName
}

# Download from GitHub releases
function Download-Release {
    Write-Host "${Blue}Fetching latest release...${NC}"

    try {
        $Latest = (Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest").tag_name
    } catch {
        Write-Host "${Red}Error: Failed to fetch latest release${NC}"
        exit 1
    }

    Write-Host "${Blue}Latest version: ${White}$Latest${NC}"
    Write-Host ""

    # Remove 'v' prefix for filename
    $Version = $Latest.TrimStart('v')
    $Filename = "lightspeed-$Version-windows-$Arch.exe"
    $Url = "https://github.com/$Repo/releases/download/$Latest/$Filename"

    # Create temp directory
    $TmpDir = Join-Path $env:TEMP "lightspeed-install-$([System.Guid]::NewGuid().ToString('N').Substring(0,8))"
    New-Item -ItemType Directory -Path $TmpDir | Out-Null
    $Binary = Join-Path $TmpDir "lightspeed.exe"

    Write-Host "${Blue}Downloading $Filename...${NC}"

    try {
        Invoke-WebRequest -Uri $Url -OutFile $Binary -UseBasicParsing
    } catch {
        Write-Host "${Red}Error: Failed to download from $Url${NC}"
        Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue
        exit 1
    }

    return $Binary
}

# Install binary
function Install-Binary {
    param([string]$Binary)

    Write-Host "${Blue}Installing to $InstallDir...${NC}"

    # Create install directory if it doesn't exist
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir | Out-Null
    }

    # Copy binary
    Copy-Item $Binary "$InstallDir\lightspeed.exe" -Force

    # Check if install dir is in PATH
    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        Write-Host ""
        Write-Host "NOTE: $InstallDir is not in your PATH."
        Write-Host "Add it to your PATH to run lightspeed from anywhere:"
        Write-Host "  [Environment]::SetEnvironmentVariable('Path', `$env:Path + ';$InstallDir', 'User')"
        Write-Host ""
    }
}

# Main
if (Test-InRepo) {
    $Binary = Build-Local
} else {
    $Binary = Download-Release
}

Install-Binary -Binary $Binary

# Cleanup temp files if downloaded
if ($Binary -like "$env:TEMP*") {
    Remove-Item -Recurse -Force (Split-Path $Binary) -ErrorAction SilentlyContinue
}

Write-Host ""
Write-Host "${Green}Successfully installed lightspeed to $InstallDir\lightspeed.exe${NC}"
Write-Host ""
Write-Host "Run 'lightspeed --help' to get started"
Write-Host ""
