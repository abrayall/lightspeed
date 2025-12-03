#!/bin/bash

# Lightspeed Server Publish Script
# Publishes the lightspeed-server Docker image to GitHub Container Registry

set -e

echo "=============================================="
echo "Lightspeed Server Publish"
echo "=============================================="
echo ""

# Colors for output
GREEN='\033[38;2;39;201;63m'
YELLOW='\033[38;2;222;184;65m'
BLUE='\033[38;2;59;130;246m'
RED='\033[38;2;239;68;68m'
GRAY='\033[38;2;136;136;136m'
NC='\033[0m' # No Color

# GitHub Container Registry
REGISTRY="ghcr.io"
IMAGE_NAME="lightspeed-server"

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Get GitHub org/user from git remote
REMOTE_URL=$(git remote get-url origin 2>/dev/null || echo "")
if [[ "$REMOTE_URL" =~ github\.com[:/]([^/]+)/ ]]; then
    GITHUB_ORG="${BASH_REMATCH[1]}"
else
    echo -e "${RED}Error: Could not determine GitHub org from git remote${NC}"
    echo "Make sure you're in a git repo with a GitHub remote"
    exit 1
fi

FULL_IMAGE_NAME="${REGISTRY}/${GITHUB_ORG}/${IMAGE_NAME}"

# Get version from latest git tag
echo -e "${BLUE}Reading version from git tags...${NC}"
GIT_DESCRIBE=$(git describe --tags --match "v*.*.*" 2>/dev/null || echo "v0.1.0")

if [[ "$GIT_DESCRIBE" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)(-([0-9]+)-g([0-9a-f]+))?$ ]]; then
    MAJOR="${BASH_REMATCH[1]}"
    MINOR="${BASH_REMATCH[2]}"
    MAINTENANCE="${BASH_REMATCH[3]}"
    COMMIT_COUNT="${BASH_REMATCH[5]}"

    if [[ -n "$COMMIT_COUNT" ]]; then
        VERSION="${MAJOR}.${MINOR}.${MAINTENANCE}-${COMMIT_COUNT}"
    else
        VERSION="${MAJOR}.${MINOR}.${MAINTENANCE}"
    fi
else
    VERSION="0.1.0"
fi

echo -e "${GREEN}Publishing version: ${VERSION}${NC}"
echo -e "${GRAY}Registry: ${FULL_IMAGE_NAME}${NC}"
echo ""

# Check if local image exists
if ! docker image inspect "lightspeed-server:${VERSION}" &>/dev/null; then
    echo -e "${YELLOW}Local image not found, running build first...${NC}"
    echo ""
    "$SCRIPT_DIR/build.sh"
    echo ""
fi

# Login to GitHub Container Registry
echo -e "${YELLOW}=== Authenticating to GitHub Container Registry ===${NC}"
echo ""

# Check for token in order: env var, config file, gh cli
TOKEN_FILE="$HOME/.config/lightspeed/registry-token"

if [ -n "$REGISTRY_TOKEN" ]; then
    echo -e "${BLUE}Logging in with REGISTRY_TOKEN env var...${NC}"
    echo "$REGISTRY_TOKEN" | docker login ghcr.io -u "$GITHUB_ORG" --password-stdin
elif [ -f "$TOKEN_FILE" ]; then
    echo -e "${BLUE}Logging in with token from ~/.config/lightspeed/registry-token...${NC}"
    cat "$TOKEN_FILE" | docker login ghcr.io -u "$GITHUB_ORG" --password-stdin
elif command -v gh &>/dev/null; then
    echo -e "${BLUE}Logging in with GitHub CLI...${NC}"
    GH_USER=$(gh api user --jq '.login' 2>/dev/null || echo "$GITHUB_ORG")
    gh auth token | docker login ghcr.io -u "$GH_USER" --password-stdin
else
    echo -e "${RED}Error: No GitHub authentication found${NC}"
    echo ""
    echo "Options:"
    echo "  1. Set REGISTRY_TOKEN environment variable"
    echo "  2. Create ~/.config/lightspeed/registry-token with your token"
    echo "  3. Install and authenticate GitHub CLI: gh auth login"
    exit 1
fi

echo ""

# Tag and push images
echo -e "${YELLOW}=== Pushing to GitHub Container Registry ===${NC}"
echo ""

echo -e "${BLUE}Tagging ${FULL_IMAGE_NAME}:${VERSION}...${NC}"
docker tag lightspeed-server:${VERSION} ${FULL_IMAGE_NAME}:${VERSION}

echo -e "${BLUE}Tagging ${FULL_IMAGE_NAME}:latest...${NC}"
docker tag lightspeed-server:latest ${FULL_IMAGE_NAME}:latest

echo -e "${BLUE}Pushing ${FULL_IMAGE_NAME}:${VERSION}...${NC}"
docker push ${FULL_IMAGE_NAME}:${VERSION}

echo -e "${BLUE}Pushing ${FULL_IMAGE_NAME}:latest...${NC}"
docker push ${FULL_IMAGE_NAME}:latest

echo ""
echo "=============================================="
echo -e "${GREEN}Publish Complete!${NC}"
echo "=============================================="
echo ""
echo -e "Images published:"
echo -e "  ${GREEN}✓${NC} ${FULL_IMAGE_NAME}:${VERSION}"
echo -e "  ${GREEN}✓${NC} ${FULL_IMAGE_NAME}:latest"
echo ""
