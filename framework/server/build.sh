#!/bin/bash

# Lightspeed Server Build Script
# Builds the lightspeed-server Docker image

set -e

echo "=============================================="
echo "Lightspeed Server Build"
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

# Generate Dockerfile
echo -e "${YELLOW}=== Generating Dockerfile ===${NC}"
echo ""

cat > "$BUILD_DIR/Dockerfile" << EOF
FROM php:8.2-fpm

# Install nginx and APCu
RUN apt-get update && apt-get install -y nginx && rm -rf /var/lib/apt/lists/*
RUN pecl install apcu && docker-php-ext-enable apcu
RUN echo 'apc.enable_cli=1' >> /usr/local/etc/php/conf.d/docker-php-ext-apcu.ini

# Configure nginx
RUN echo 'server {\n\
    listen 80;\n\
    server_name _;\n\
    root /var/www/html;\n\
    index index.php index.html;\n\
\n\
    location / {\n\
        try_files \$uri \$uri/ \$uri.php?\$query_string;\n\
    }\n\
\n\
    location ~ \.php\$ {\n\
        fastcgi_pass 127.0.0.1:9000;\n\
        fastcgi_param SCRIPT_FILENAME \$document_root\$fastcgi_script_name;\n\
        include fastcgi_params;\n\
    }\n\
}' > /etc/nginx/sites-available/default

# Create lightspeed directory and store version
RUN mkdir -p /opt/lightspeed
COPY library/ /opt/lightspeed/
RUN echo 'version=${VERSION}' > /opt/lightspeed/version.properties && \
    chmod -R 755 /opt/lightspeed

# Add /opt to PHP include path
RUN echo 'include_path = ".:/opt"' > /usr/local/etc/php/conf.d/lightspeed.ini

# Start script to run both nginx and php-fpm
RUN echo '#!/bin/bash\n\
php-fpm -D\n\
nginx -g "daemon off;"' > /start.sh && chmod +x /start.sh

EXPOSE 80

CMD ["/start.sh"]
EOF

echo -e "${GREEN}✓ Created: build/Dockerfile${NC}"

# Copy library files
echo -e "${BLUE}Copying library files...${NC}"
cp -r "$SCRIPT_DIR/../library" "$BUILD_DIR/library"
echo -e "${GREEN}✓ Copied library files${NC}"
echo ""

# Build the image
echo -e "${YELLOW}=== Building Docker Image ===${NC}"
echo ""

cd "$BUILD_DIR"
echo -e "${BLUE}Building lightspeed-server:${VERSION}...${NC}"
docker build -t lightspeed-server:${VERSION} .
docker tag lightspeed-server:${VERSION} lightspeed-server:latest

echo ""
echo "=============================================="
echo -e "${GREEN}Build Complete!${NC}"
echo "=============================================="
echo ""
echo -e "Images created:"
echo -e "  ${GREEN}✓${NC} lightspeed-server:${VERSION}"
echo -e "  ${GREEN}✓${NC} lightspeed-server:latest"
echo ""
