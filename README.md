# Lightspeed

A lightweight, rapid development tool for small PHP websites.

```
──────────────────────────────────────────────

 █   ▀█▀ █▀▀▀ █  █ ▀▀█▀▀ █▀▀ █▀▀█ █▀▀ █▀▀ █▀▀▄
 █    █  █ ▀█ █▀▀█   █   ▀▀█ █  █ █▀▀ █▀▀ █  █
 ▀▀▀ ▀▀▀ ▀▀▀▀ ▀  ▀   ▀   ▀▀▀ █▀▀▀ ▀▀▀ ▀▀▀ ▀▀▀

──────────────────────────────────────────────
```

## Installation

### From Source

```bash
git clone https://github.com/abrayall/lightspeed.git
cd lightspeed
./install.sh
```

### Manual Build

```bash
./build.sh
sudo cp build/lightspeed-*-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) /usr/local/bin/lightspeed
```

## Commands

### init

Initialize a new Lightspeed project with basic directory structure.

```bash
lightspeed init
```

Creates:
- `index.php` - Hello World starter page
- `assets/css/style.css` - Basic stylesheet
- `assets/js/` - JavaScript directory
- `includes/` - PHP includes directory

### start

Start a PHP development server using Docker.

```bash
lightspeed start
```

Options:
- `-p, --port` - Port to expose (default: auto-detect in 9000 range)
- `-i, --image` - Docker image to use (default: php:8.2-apache)

The server mounts your current directory and serves it at `http://localhost:<port>`.

### stop

Stop the running development server.

```bash
lightspeed stop
```

### build

Build a Docker container for the project.

```bash
lightspeed build
```

Options:
- `-t, --tag` - Version tag (default: git version or 'latest')
- `-i, --image` - Base Docker image (default: php:8.2-apache)

Builds for `linux/amd64` platform for production deployment.

### publish

Build and push Docker image to DigitalOcean Container Registry.

```bash
lightspeed publish
```

Options:
- `-t, --tag` - Version tag (default: git version or 'latest')
- `-r, --registry` - Container registry URL (default: registry.digitalocean.com/lightspeed-images)
- `-k, --token` - DigitalOcean API token

Pushes both versioned tag and `latest` tag.

### deploy

Build, push, and deploy to DigitalOcean App Platform.

```bash
lightspeed deploy
```

Options:
- `-a, --app` - App name (default: project directory name)
- `-r, --region` - DigitalOcean region (default: nyc)
- `-k, --token` - DigitalOcean API token

If the app doesn't exist, it will be created automatically.

## Workflow

### Development

```bash
# Create a new project
mkdir mysite && cd mysite
lightspeed init

# Start development server
lightspeed start

# Edit your PHP files...
# Changes are reflected immediately

# Stop when done
lightspeed stop
```

### Deployment

```bash
# Build and deploy to DigitalOcean
lightspeed deploy

# Or step by step:
lightspeed build      # Build Docker image
lightspeed publish    # Push to registry
lightspeed deploy     # Deploy to App Platform
```

## Project Structure

```
mysite/
├── index.php           # Main entry point
├── assets/
│   ├── css/
│   │   └── style.css   # Stylesheets
│   └── js/             # JavaScript files
├── includes/           # PHP includes
└── Dockerfile          # Generated on build
```

## Requirements

- Go 1.21+ (for building from source)
- Docker (for development server and builds)
- DigitalOcean account (for deployment)

## License

MIT
