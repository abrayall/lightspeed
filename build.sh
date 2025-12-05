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

# Get version using vermouth
echo -e "${BLUE}Reading version from git tags...${NC}"
VERSION=$(vermouth 2>/dev/null || curl -sfL https://raw.githubusercontent.com/abrayall/vermouth/refs/heads/main/vermouth.sh | sh -)

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

# Create temp directory with lightspeed/ subfolder structure
LIBRARY_TMP=$(mktemp -d)
mkdir -p "$LIBRARY_TMP/lightspeed"
cp -r "$SCRIPT_DIR/framework/library/"* "$LIBRARY_TMP/lightspeed/"
echo "version=${VERSION}" > "$LIBRARY_TMP/lightspeed/version.properties"

cd "$LIBRARY_TMP"
zip -r "$BUILD_DIR/$LIBRARY_ZIP" lightspeed -x "*.DS_Store" -x "*/test.php" -x "*/tests/*"
cd "$SCRIPT_DIR"
rm -rf "$LIBRARY_TMP"

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
