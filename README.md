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

### Quick Install

**macOS/Linux:**
```bash
curl -sfL https://raw.githubusercontent.com/abrayall/lightspeed/refs/heads/main/install.sh | sh -
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/abrayall/lightspeed/refs/heads/main/install.ps1 | iex
```

### From Source

```bash
git clone https://github.com/abrayall/lightspeed.git
cd lightspeed
./install.sh
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
- `-n, --name` - Site name (default: project directory name)
- `-r, --region` - DigitalOcean region (default: nyc)
- `-k, --token` - DigitalOcean API token

If the app doesn't exist, it will be created automatically. Your site will be accessible at:
- `https://[name].lightspeed.ee` (automatically configured)
- `https://[name]-[hash].ondigitalocean.app` (DigitalOcean URL)

## Configuration

### site.properties

Create a `site.properties` file in your project root to configure deployment:

```properties
# Site name (used for [name].lightspeed.ee)
name=mysite

# Custom domains (optional)
domain=example.com
domains=www.example.com,app.example.com
```

Properties:
- `name` - Site name, used for lightspeed.ee subdomain
- `domain` - Single custom domain
- `domains` - Comma-separated list of custom domains

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
