#!/bin/bash

# Lightspeed Operator Deploy Script
# Builds and pushes operator Docker image to DigitalOcean registry

set -e

# Colors
GREEN='\033[38;2;39;201;63m'
YELLOW='\033[38;2;222;184;65m'
BLUE='\033[38;2;59;130;246m'
GRAY='\033[38;2;136;136;136m'
RED='\033[0;31m'
NC='\033[0m'

# Parse arguments
COMPONENTS="site,operator"
while [[ $# -gt 0 ]]; do
    case $1 in
        --components=*)
            COMPONENTS="${1#*=}"
            shift
            ;;
        --components)
            COMPONENTS="$2"
            shift 2
            ;;
        *)
            echo -e "${RED}Unknown argument: $1${NC}"
            echo "Usage: $0 [--components=site,operator]"
            exit 1
            ;;
    esac
done

# Registry configuration
REGISTRY="registry.digitalocean.com"
REPO="abrayall"
IMAGE="lightspeed-operator"

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Work directory
WORK_DIR="$SCRIPT_DIR/build/work"
mkdir -p "$WORK_DIR"

echo "=============================================="
echo -e "${YELLOW}Lightspeed Operator Deploy${NC}"
echo "=============================================="
echo ""

# Get version from git
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

# Check for uncommitted changes
if [[ -n $(git status --porcelain 2>/dev/null) ]]; then
    TIMESTAMP=$(date +"%m%d%H%M")
    VERSION="${VERSION}-${TIMESTAMP}"
    echo -e "${GRAY}Detected uncommitted changes, appending timestamp${NC}"
fi

# Full image names
VERSION_TAG="${REGISTRY}/${REPO}/${IMAGE}:${VERSION}"
LATEST_TAG="${REGISTRY}/${REPO}/${IMAGE}:latest"

echo -e "${BLUE}Version:${NC}  ${VERSION}"
echo -e "${BLUE}Registry:${NC} ${REGISTRY}/${REPO}"
echo -e "${BLUE}Image:${NC}    ${IMAGE}"
echo ""

# Deploy lightspeed website first (before operator deployment)
if [[ "$COMPONENTS" == *"site"* ]]; then
    WWW_DIR="$SCRIPT_DIR/platform/www"
    if [ -d "$WWW_DIR" ]; then
        echo -e "${YELLOW}Deploying lightspeed website...${NC}"

        # Determine which CLI binary to use based on platform
        PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')
        ARCH=$(uname -m)
        if [ "$ARCH" = "x86_64" ]; then
            ARCH="amd64"
        elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
            ARCH="arm64"
        fi

        CLI_BINARY="$SCRIPT_DIR/build/lightspeed-cli-${VERSION}-${PLATFORM}-${ARCH}"

        # Fallback to finding any CLI binary if version-specific doesn't exist
        if [ ! -f "$CLI_BINARY" ]; then
            CLI_BINARY=$(find "$SCRIPT_DIR/build" -name "lightspeed-cli-*-${PLATFORM}-${ARCH}" | head -1)
        fi

        if [ -f "$CLI_BINARY" ]; then
            echo -e "${GRAY}Using CLI: $(basename $CLI_BINARY)${NC}"
            cd "$WWW_DIR"
            "$CLI_BINARY" deploy --name lightspeed
            cd "$SCRIPT_DIR"
            echo ""
        else
            echo -e "${YELLOW}⚠ CLI binary not found, skipping website deployment${NC}"
            echo -e "${GRAY}  Expected: lightspeed-cli-${VERSION}-${PLATFORM}-${ARCH}${NC}"
            echo ""
        fi
    fi
fi

# Deploy operator
if [[ "$COMPONENTS" == *"operator"* ]]; then
    # Generate Dockerfile
    echo -e "${BLUE}Generating Dockerfile...${NC}"
    cat > "$WORK_DIR/Dockerfile" << 'DOCKERFILE'
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy only the source directories needed
COPY core/ core/
COPY platform/operator/ platform/operator/

# Build operator
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.Version=${VERSION}" \
    -o /operator \
    ./platform/operator

# Runtime image
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

COPY --from=builder /operator /usr/local/bin/operator

EXPOSE 8080

ENTRYPOINT ["operator"]
DOCKERFILE

# Build the Docker image
echo -e "${YELLOW}Building Docker image...${NC}"
echo ""

docker build \
    --platform linux/amd64 \
    --build-arg VERSION="${VERSION}" \
    -t "${VERSION_TAG}" \
    -t "${LATEST_TAG}" \
    -f "$WORK_DIR/Dockerfile" \
    .

echo ""
echo -e "${GREEN}✓ Built: ${VERSION_TAG}${NC}"
echo ""

# Login to registry
TOKEN="${DIGITALOCEAN_TOKEN:-$TOKEN}"  # Support both names for backwards compatibility
if [ -n "$TOKEN" ]; then
    echo -e "${BLUE}Logging in to registry...${NC}"
    echo "$TOKEN" | docker login "$REGISTRY" --username "$TOKEN" --password-stdin
    echo ""
else
    echo -e "${GRAY}No DIGITALOCEAN_TOKEN env var set, assuming already logged in${NC}"
fi

# Push to registry
echo -e "${YELLOW}Pushing to registry...${NC}"
echo ""

docker push "${VERSION_TAG}"
docker push "${LATEST_TAG}"

echo ""
echo -e "${GREEN}✓ Pushed images${NC}"
echo ""

# Check/create DigitalOcean App
APP_NAME="lightspeed-operator"

if [ -z "$TOKEN" ]; then
    echo -e "${GRAY}No DIGITALOCEAN_TOKEN set, skipping app deployment${NC}"
else
    echo -e "${YELLOW}Checking DigitalOcean App Platform...${NC}"

    # Get all apps and search for our app by name
    APPS_RESPONSE=$(curl -s -X GET \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        "https://api.digitalocean.com/v2/apps")

    # Check if our app exists (look for the name in spec)
    if echo "$APPS_RESPONSE" | grep -q "\"name\":\"$APP_NAME\""; then
        echo -e "${GREEN}✓ App '$APP_NAME' already exists${NC}"

        # Get the app URL
        # Extract the app ID for our app
        APP_URL=$(echo "$APPS_RESPONSE" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for app in data.get('apps', []):
    if app.get('spec', {}).get('name') == '$APP_NAME':
        print(app.get('live_url', ''))
        break
" 2>/dev/null)

        if [ -n "$APP_URL" ]; then
            echo -e "${BLUE}  URL:${NC} $APP_URL"
        fi
        echo -e "${GRAY}  Deployment will be triggered automatically by deploy_on_push${NC}"
    else
        echo -e "${BLUE}Creating app '$APP_NAME'...${NC}"

        # Create app spec (single-line JSON for reliable parsing)
        APP_SPEC='{"spec":{"name":"lightspeed-operator","region":"nyc","features":["buildpack-stack=ubuntu-22"],"alerts":[{"rule":"DEPLOYMENT_FAILED"},{"rule":"DOMAIN_FAILED"}],"domains":[{"domain":"api.lightspeed.ee","type":"PRIMARY"},{"domain":"registry.lightspeed.ee","type":"ALIAS"}],"ingress":{"rules":[{"component":{"name":"lightspeed-operator"},"match":{"path":{"prefix":"/"}}}]},"services":[{"name":"lightspeed-operator","http_port":80,"image":{"registry_type":"DOCR","registry":"abrayall","repository":"lightspeed-operator","tag":"latest","deploy_on_push":{"enabled":true}},"health_check":{"http_path":"/health","initial_delay_seconds":5,"period_seconds":10,"timeout_seconds":3,"success_threshold":1,"failure_threshold":3},"instance_count":1,"instance_size_slug":"apps-s-1vcpu-0.5gb"}]}}'

        RESPONSE=$(curl -s -X POST \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "$APP_SPEC" \
            "https://api.digitalocean.com/v2/apps")

        if echo "$RESPONSE" | grep -q '"app"'; then
            echo -e "${GREEN}✓ App '$APP_NAME' created${NC}"

            # Extract app ID from response
            APP_ID=$(echo "$RESPONSE" | python3 -c "
import sys, json
data = json.load(sys.stdin)
print(data.get('app', {}).get('id', ''))
" 2>/dev/null)

            echo -e "${GRAY}  Waiting for deployment...${NC}"

            # Poll for deployment status
            LAST_STATUS=""
            TIMEOUT=300  # 5 minutes
            ELAPSED=0

            while [ $ELAPSED -lt $TIMEOUT ]; do
                STATUS_RESPONSE=$(curl -s -X GET \
                    -H "Authorization: Bearer $TOKEN" \
                    -H "Content-Type: application/json" \
                    "https://api.digitalocean.com/v2/apps/$APP_ID")

                CURRENT_STATUS=$(echo "$STATUS_RESPONSE" | python3 -c "
import sys, json
data = json.load(sys.stdin)
app = data.get('app', {})
deployment = app.get('active_deployment') or app.get('pending_deployment') or {}
print(deployment.get('phase', 'UNKNOWN'))
" 2>/dev/null)

                # Show status changes
                if [ "$CURRENT_STATUS" != "$LAST_STATUS" ]; then
                    case "$CURRENT_STATUS" in
                        "PENDING_BUILD") echo -e "${GRAY}  Status: Pending build...${NC}" ;;
                        "BUILDING") echo -e "${YELLOW}  Status: Building...${NC}" ;;
                        "PENDING_DEPLOY") echo -e "${GRAY}  Status: Pending deploy...${NC}" ;;
                        "DEPLOYING") echo -e "${YELLOW}  Status: Deploying...${NC}" ;;
                        "ACTIVE") echo -e "${GREEN}  Status: Active${NC}" ;;
                        "ERROR"|"FAILED") echo -e "${RED}  Status: Failed${NC}" ;;
                        *) echo -e "${GRAY}  Status: $CURRENT_STATUS${NC}" ;;
                    esac
                    LAST_STATUS="$CURRENT_STATUS"
                fi

                # Check for terminal states
                if [ "$CURRENT_STATUS" = "ACTIVE" ]; then
                    APP_URL=$(echo "$STATUS_RESPONSE" | python3 -c "
import sys, json
data = json.load(sys.stdin)
print(data.get('app', {}).get('live_url', ''))
" 2>/dev/null)
                    echo ""
                    echo -e "${GREEN}✓ Deployment successful!${NC}"
                    if [ -n "$APP_URL" ]; then
                        echo -e "${BLUE}  URL:${NC} $APP_URL"
                    fi
                    break
                fi

                if [ "$CURRENT_STATUS" = "ERROR" ] || [ "$CURRENT_STATUS" = "FAILED" ]; then
                    echo ""
                    echo -e "${RED}✗ Deployment failed${NC}"
                    break
                fi

                sleep 5
                ELAPSED=$((ELAPSED + 5))
            done

            if [ $ELAPSED -ge $TIMEOUT ]; then
                echo -e "${YELLOW}⚠ Deployment still in progress (timed out waiting)${NC}"
            fi
        else
            echo -e "${RED}✗ Failed to create app${NC}"
            ERROR_MSG=$(echo "$RESPONSE" | python3 -c "
import sys, json
data = json.load(sys.stdin)
print(data.get('message', data.get('id', 'Unknown error')))
" 2>/dev/null || echo "$RESPONSE")
            echo -e "${RED}  Error: $ERROR_MSG${NC}"
        fi
    fi
fi
fi

echo ""
echo "=============================================="
echo -e "${GREEN}Deploy Complete!${NC}"
echo "=============================================="
echo ""

if [[ "$COMPONENTS" == *"operator"* ]]; then
    echo "Pushed images:"
    echo "  • ${VERSION_TAG}"
    echo "  • ${LATEST_TAG}"
    echo ""
fi
