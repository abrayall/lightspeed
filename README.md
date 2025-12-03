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
lightspeed init --name mysite
lightspeed init --name mysite --domain example.com
lightspeed init -d example.com -d www.example.com
```

Options:
- `-n, --name` - Site name (default: directory name)
- `-d, --domain` - Domain(s) for the site (default: name.com). Can be specified multiple times.

Creates:
- `site.properties` - Site configuration
- `index.php` - Hello World starter page
- `assets/css/style.css` - Basic stylesheet
- `assets/js/` - JavaScript directory
- `includes/` - PHP includes directory
- `.idea/` - PhpStorm project configuration
- `.gitignore` - Git ignore file

Running `init` again is safe - it only creates files that don't exist and updates the PhpStorm configuration.

### start

Start a PHP development server using Docker.

```bash
lightspeed start
```

Options:
- `-p, --port` - Port to expose (default: auto-detect in 9000 range)
- `-i, --image` - Docker image to use (default: lightspeed-server)

The server mounts your current directory and serves it at `http://localhost:<port>`.

**Features:**
- Clean URLs - access `/about` instead of `/about.php`
- Automatic PHP library loading from `~/.lightspeed/library/`
- Hot reload - changes are reflected immediately

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
- `-i, --image` - Base Docker image (default: lightspeed-server)

Builds for `linux/amd64` platform for production deployment.

### publish

Build and push Docker image to the Lightspeed registry.

```bash
lightspeed publish
```

Options:
- `-t, --tag` - Version tag (default: git version or 'latest')
- `-n, --name` - Site name (default: from site.properties or directory name)

Pushes both versioned tag and `latest` tag.

### deploy

Build, push, and deploy to DigitalOcean App Platform.

```bash
lightspeed deploy
```

Options:
- `-n, --name` - Site name (default: project directory name)

If the app doesn't exist, it will be created automatically. Your site will be accessible at:
- `https://[name].lightspeed.ee` (automatically configured)

## Configuration

### site.properties

Create a `site.properties` file in your project root to configure your site:

```properties
# Site name (used for [name].lightspeed.ee)
name=mysite

# Custom domains
domain=example.com
domains=www.example.com,app.example.com

# Base image (pin to specific version)
image=0.5.4

# PHP libraries (for include path)
libraries=lightspeed
```

#### Properties

| Property | Description | Default |
|----------|-------------|---------|
| `name` | Site name, used for lightspeed.ee subdomain | Directory name |
| `domain` | Single custom domain | - |
| `domains` | Comma-separated list of custom domains | - |
| `image` | Base Docker image version | CLI version |
| `libraries` | Comma-separated PHP library paths | - |

#### Image Property

The `image` property controls which base image is used for `start` and `build`:

```properties
# Use specific version
image=0.5.4

# Use latest
image=latest

# Use custom image
image=ghcr.io/myorg/myimage:latest
```

#### Libraries Property

The `libraries` property specifies PHP libraries to include in the PHP include path:

```properties
# Use lightspeed library (matches CLI version)
libraries=lightspeed

# Use specific lightspeed version
libraries=lightspeed:v0.5.0

# Multiple libraries
libraries=lightspeed,/path/to/custom/lib
```

## PHP Library

Lightspeed includes a PHP library that's automatically available in the server image at `/opt/lightspeed/`. The PHP include path is configured to allow:

```php
<?php
require_once('lightspeed/version.php');

echo lightspeed_version(); // Returns the Lightspeed version
```

### IDE Support

When you run `lightspeed init` or any lightspeed command in a project with `.idea/` and `site.properties`, the PhpStorm include paths are automatically updated to point to the resolved library locations.

The library is downloaded to `~/.lightspeed/library/v[version]/` on first use.

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
# Access /about instead of /about.php

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
├── site.properties     # Site configuration
├── index.php           # Main entry point
├── assets/
│   ├── css/
│   │   └── style.css   # Stylesheets
│   └── js/             # JavaScript files
├── includes/           # PHP includes
├── .idea/              # PhpStorm configuration
│   └── php.xml         # PHP include paths
├── .gitignore          # Git ignore file
└── Dockerfile          # Generated on build
```

## Server Image

Lightspeed uses a custom server image (`ghcr.io/abrayall/lightspeed-server`) based on:
- PHP 8.2 FPM
- Nginx

**Features:**
- Clean URLs (no `.php` extension required)
- Pre-configured PHP include path for Lightspeed library
- Optimized for small PHP sites

## Requirements

- Docker (for development server and builds)
- DigitalOcean account (for deployment)

## License

MIT
