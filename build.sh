#!/bin/bash

# Lightspeed Build Script
# Builds CLI and platform components for multiple platforms

set -e  # Exit on error

echo "=============================================="
echo "Lightspeed Build"
echo "=============================================="
echo ""

# Colors for output
GREEN='\033[38;2;39;201;63m'
YELLOW='\033[38;2;222;184;65m'
BLUE='\033[38;2;59;130;246m'
GRAY='\033[38;2;136;136;136m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Build directory
BUILD_DIR="$SCRIPT_DIR/build"

# Clean previous build
echo -e "${BLUE}Cleaning previous build...${NC}"
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Get version from latest git tag
echo -e "${BLUE}Reading version from git tags...${NC}"
GIT_DESCRIBE=$(git describe --tags --match "v*.*.*" 2>/dev/null || echo "v0.1.0")

# Parse git describe output
# Format: v0.1.0 or v0.1.0-5-g1a2b3c4 (if commits exist after tag)
if [[ "$GIT_DESCRIBE" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)(-([0-9]+)-g([0-9a-f]+))?$ ]]; then
    MAJOR="${BASH_REMATCH[1]}"
    MINOR="${BASH_REMATCH[2]}"
    MAINTENANCE="${BASH_REMATCH[3]}"
    COMMIT_COUNT="${BASH_REMATCH[5]}"

    # If there are commits after the tag, append commit count to maintenance
    if [[ -n "$COMMIT_COUNT" ]]; then
        MAINTENANCE="${MAINTENANCE}-${COMMIT_COUNT}"
        VERSION="${MAJOR}.${MINOR}.${MAINTENANCE}"
    else
        VERSION="${MAJOR}.${MINOR}.${MAINTENANCE}"
    fi
else
    # Fallback
    MAJOR=0
    MINOR=1
    MAINTENANCE=0
    VERSION="0.1.0"
fi

# Check for uncommitted local changes
if [[ -n $(git status --porcelain 2>/dev/null) ]]; then
    TIMESTAMP=$(date +"%m%d%H%M")
    MAINTENANCE="${MAINTENANCE}-${TIMESTAMP}"
    VERSION="${MAJOR}.${MINOR}.${MAINTENANCE}"
    echo -e "${GRAY}Detected uncommitted changes, appending timestamp${NC}"
fi

echo -e "${GREEN}Building version: ${VERSION}${NC}"
echo ""

# Build platforms
PLATFORMS=("darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64" "windows/amd64")

# Build CLI
echo -e "${YELLOW}=== Building CLI ===${NC}"
echo ""

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS="${PLATFORM%/*}"
    GOARCH="${PLATFORM#*/}"

    OUTPUT_NAME="lightspeed-cli-${VERSION}-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi

    echo -e "${BLUE}Building CLI ${GOOS}/${GOARCH}...${NC}"

    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-X lightspeed/framework/cli/cmd.Version=${VERSION}" \
        -o "$BUILD_DIR/$OUTPUT_NAME" \
        ./framework/cli

    echo -e "${GREEN}✓ Created: ${OUTPUT_NAME}${NC}"
done

echo ""

# Build Operator
echo -e "${YELLOW}=== Building Operator ===${NC}"
echo ""

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS="${PLATFORM%/*}"
    GOARCH="${PLATFORM#*/}"

    OUTPUT_NAME="lightspeed-operator-${VERSION}-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi

    echo -e "${BLUE}Building Operator ${GOOS}/${GOARCH}...${NC}"

    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-X main.Version=${VERSION}" \
        -o "$BUILD_DIR/$OUTPUT_NAME" \
        ./platform/operator

    echo -e "${GREEN}✓ Created: ${OUTPUT_NAME}${NC}"
done

echo ""

# Package library
echo -e "${YELLOW}=== Packaging Library ===${NC}"
echo ""

LIBRARY_ZIP="lightspeed-library-${VERSION}.zip"
echo -e "${BLUE}Creating ${LIBRARY_ZIP}...${NC}"

cd "$SCRIPT_DIR/framework/library"
zip -r "$BUILD_DIR/$LIBRARY_ZIP" . -x "*.DS_Store"
cd "$SCRIPT_DIR"

echo -e "${GREEN}✓ Created: ${LIBRARY_ZIP}${NC}"
echo ""

# Summary
echo "=============================================="
echo -e "${GREEN}Build Complete!${NC}"
echo "=============================================="
echo ""
echo "Artifacts created in build/:"
ls -1 "$BUILD_DIR"
echo ""
