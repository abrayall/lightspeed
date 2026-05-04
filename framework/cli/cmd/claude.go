package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"lightspeed/core/lib/ui"
)

const skillVersionPrefix = "<!-- lightspeed:"
const skillVersionSuffix = " -->"

// addClaudeSupport generates .claude/skills/lightspeed/SKILL.md
// Returns list of created file paths (relative to dir)
func addClaudeSupport(dir string) []string {
	var created []string

	skillDir := filepath.Join(dir, ".claude", "skills", "lightspeed")
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			ui.PrintWarning("Failed to create .claude/skills/lightspeed directory: %v", err)
			return created
		}
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		content := generateSkillMD()
		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			ui.PrintWarning("Failed to create SKILL.md: %v", err)
		} else {
			created = append(created, ".claude/skills/lightspeed/SKILL.md")
		}
	}

	return created
}

// upgradeClaudeSkill checks if the skill file exists and is outdated, and upgrades it.
// Returns true if an upgrade was performed.
func upgradeClaudeSkill(dir string) bool {
	skillPath := filepath.Join(dir, ".claude", "skills", "lightspeed", "SKILL.md")

	f, err := os.Open(skillPath)
	if err != nil {
		return false
	}
	defer f.Close()

	// Read first line to check version
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return false
	}
	firstLine := scanner.Text()

	// Parse embedded version
	if !strings.HasPrefix(firstLine, skillVersionPrefix) {
		// Old file without version marker — upgrade it
		content := generateSkillMD()
		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			return false
		}
		return true
	}

	embeddedVersion := strings.TrimPrefix(firstLine, skillVersionPrefix)
	embeddedVersion = strings.TrimSuffix(embeddedVersion, skillVersionSuffix)

	if embeddedVersion == Version {
		return false
	}

	// Version mismatch — regenerate
	content := generateSkillMD()
	if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
		return false
	}
	return true
}

// parseSiteName reads site.properties and extracts the name= value.
// Falls back to the directory basename if not found.
func parseSiteName(dir string) string {
	propsPath := filepath.Join(dir, "site.properties")
	f, err := os.Open(propsPath)
	if err != nil {
		return filepath.Base(dir)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "name=") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "name="))
			if val != "" {
				return val
			}
		}
		if strings.HasPrefix(line, "name:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			if val != "" {
				return val
			}
		}
	}

	return filepath.Base(dir)
}

func generateSkillMD() string {
	return fmt.Sprintf(`%s%s%s
---
name: lightspeed
description: "Use this skill when working on a Lightspeed PHP site, including CLI commands, PHP library usage, site.properties configuration, deployment, and project structure."
---

# Lightspeed Development Skill

## CLI Commands

### lightspeed init
Initialize a new project in the current directory.

Flags:
- `+"`--name, -n`"+` — Site name (default: directory name)
- `+"`--domain, -d`"+` — Domain(s) for the site (default: name.com)
- `+"`--git, -g`"+` — Generate GitHub Actions deploy workflow and .gitignore
- `+"`--claude, -c`"+` — Generate Claude Code support files

Creates: `+"`index.php`"+`, `+"`assets/`"+`, `+"`includes/`"+`, `+"`site.properties`"+`, `+"`.idea/`"+`

### lightspeed start
Start a local development server using Docker.

Flags:
- `+"`--port, -p`"+` — Port to expose (default: auto-detect in 9000-9099 range)
- `+"`--image, -i`"+` — Docker image to use

Mounts the current directory into the container. Opens browser automatically.

### lightspeed stop
Stop the running development server container.

### lightspeed build
Build a Docker image for the project.

Flags:
- `+"`--tag, -t`"+` — Image tag (default: git version or "latest")
- `+"`--image, -i`"+` — Base Docker image

Auto-generates a Dockerfile if one doesn't exist (removed after build).

### lightspeed publish
Build and push Docker image to the Lightspeed registry.

Flags:
- `+"`--tag, -t`"+` — Version tag (default: git version or "latest")
- `+"`--name, -n`"+` — Site name for registry

### lightspeed deploy
Build, push, and deploy to Lightspeed platform. Combines publish + site creation/update.

Flags:
- `+"`--name, -n`"+` — Site name

Deploys to `+"`https://{name}.lightspeed.ee`"+`.

### lightspeed add [feature]
Add features to an existing project.

Available features:
- `+"`git`"+` — GitHub Actions deploy workflow and .gitignore
- `+"`claude`"+` — Claude Code support files

### lightspeed encrypt [value]
Encrypt a value for use in site.properties environment variables. Uses AES-256-GCM.

The encrypted output can be used directly as an environment variable value in site.properties — no special prefix or syntax needed. At runtime, encrypted values are automatically detected and decrypted.

Key resolution order:
1. `+"`LIGHTSPEED_KEY`"+` environment variable
2. `+"`~/.lightspeed/key`"+` file
3. `+"`key`"+` property in site.properties
4. `+"`domain`"+` or `+"`name`"+` property in site.properties
5. Derived from directory name

Must be run from the project directory to use the correct key.

## PHP Library

The `+"`lightspeed`"+` library is auto-installed and available via `+"`libraries=lightspeed`"+` in site.properties.

### site.php — Site Configuration

`+"```php"+`
// Get the site instance (singleton)
$site = site();

// Read values from site.properties
$site->get('key', 'default');          // String value
$site->getBoolean('key', false);       // Boolean value
$site->getList('key', []);             // Array (comma-separated or YAML list)
$site->getInt('key', 0);              // Integer value
$site->getEncrypted('key', '');       // Auto-decrypted value
$site->has('key');                     // Check existence
$site->all();                          // All properties

// Convenience methods
$site->name();                         // Site name
$site->domain();                       // Primary domain
$site->domains();                      // All domains
`+"```"+`

### cms.php — Velocity CMS Client

`+"```php"+`
// Get CMS instance
$cms = cms();

// Fetch content
$html = cms()->get('page-name', 'pages');
$html = content('page-name', 'pages');         // Shorthand

// Get asset URL (for images in src attributes)
$url = cms()->getUrl('hero', 'images');
$url = contentUrl('hero', 'images');           // Shorthand

// Get metadata (alt text, dimensions)
$meta = cms()->getMetadata('hero', 'images');
$meta = contentMetadata('hero', 'images');     // Shorthand

// Get URL + metadata together
$asset = cms()->getAsset('hero', 'images');
$asset = contentAsset('hero', 'images');       // Shorthand
// $asset['url'], $asset['metadata']['alt']

// Bulk fetch multiple items
$results = cms()->getAll([
    ['type' => 'pages', 'id' => 'header'],
    ['type' => 'pages', 'id' => 'footer'],
    ['type' => 'images', 'id' => 'hero', 'attributes' => ['url', 'metadata']],
]);
// Access: $results['pages/header']['content']

$results = contents([...]); // Shorthand for getAll

// List and inspect content
$items = cms()->list('pages');            // List items in a type
$types = cms()->types();                   // List all content types
$versions = cms()->versions('id', 'type'); // Get version history

// Cache management
cms()->clearCache();                                    // Clear all
cms()->clearCacheFor('id', 'type');                    // Clear specific
cms()->clearCacheForAll('id', 'type');                 // Clear all attributes
`+"```"+`

CMS configuration in site.properties:
`+"```properties"+`
velocity.endpoint=https://velocity.ee
velocity.tenant=my-site
velocity.cache=true
velocity.cache.soft_ttl=300
velocity.cache.hard_ttl=3600
`+"```"+`

### mail.php — Email via SMTP

`+"```php"+`
// Fluent interface
mailer()
    ->to('user@example.com', 'Name')
    ->subject('Hello')
    ->html('<h1>Hi</h1>', 'Plain text alt')
    ->attach('/path/to/file.pdf')
    ->send();

// Quick send
send_mail('user@example.com', 'Subject', 'Body text');
send_mail('user@example.com', 'Subject', '<h1>HTML</h1>', true);
`+"```"+`

SMTP configuration in site.properties:
`+"```properties"+`
smtp.host=smtp.example.com
smtp.port=587
smtp.secure=tls
smtp.username=<encrypted>
smtp.password=<encrypted>
smtp.from=noreply@example.com
smtp.from.name=My Site
`+"```"+`

### crypto.php — Encryption

`+"```php"+`
// Encrypt/decrypt values
$encrypted = crypto()->encrypt('secret value');
$decrypted = crypto()->decrypt($encrypted);

// Auto-decrypt from site.properties
$password = site()->getEncrypted('smtp.password');
`+"```"+`

## Environment Variables

Environment variables are defined in the `+"`environment:`"+` section of site.properties:

`+"```properties"+`
environment:
  API_URL: https://api.example.com
  DB_HOST: mydb.example.com
  DB_PASSWORD: <encrypted value from lightspeed encrypt>
`+"```"+`

- **Local dev** (`+"`lightspeed start`"+`): injected as `+"`-e`"+` flags into the Docker container
- **Production** (`+"`lightspeed build`"+`): baked as `+"`ENV`"+` lines in the Dockerfile
- Values can be plain text or encrypted — encrypted values are automatically decrypted at build/start time
- To encrypt a value: `+"`lightspeed encrypt \"my-secret\"`"+`

## Composer Packages

Composer PHP packages are configured via the `+"`libraries`"+` property with the `+"`composer:`"+` prefix:

`+"```properties"+`
libraries=lightspeed,composer:stripe/stripe-php:^16.0,composer:monolog/monolog:^3.0
`+"```"+`

Syntax:
- `+"`composer`"+` or `+"`composer:<version>`"+` — the Composer tool itself (no `+"`/`"+` = tool version)
- `+"`composer:<vendor>/<package>:<version>`"+` — a package (has `+"`/`"+` = package spec)
- Version defaults to `+"`*`"+` if omitted

How it works:
- `+"`composer.json`"+` is auto-generated from site.properties (gitignored)
- `+"`vendor/`"+` is auto-installed on `+"`lightspeed start`"+` (gitignored)
- `+"`vendor/autoload.php`"+` is automatically loaded via `+"`auto_prepend_file`"+`
- Production builds run `+"`composer install --no-dev --optimize-autoloader`"+` in the Docker image
- Composer phar is cached at `+"`~/.lightspeed/composer/v<version>/composer.phar`"+`

## site.properties Reference

`+"```properties"+`
# Basic
name=my-site
domain=example.com
domains=example.com,www.example.com

# Libraries and Composer packages
libraries=lightspeed,composer:stripe/stripe-php:^16.0

# Base image (optional)
image=0.5.3

# Environment variables (plain text or encrypted)
environment:
  API_URL: https://api.example.com
  DB_PASSWORD: <encrypted value>

# Encryption key (optional, otherwise derived from domain/name)
key=my-secret-key

# CMS
velocity.endpoint=https://velocity.ee
velocity.tenant=my-site
velocity.cache=true
velocity.cache.soft_ttl=300
velocity.cache.hard_ttl=3600

# SMTP
smtp.host=smtp.example.com
smtp.port=587
smtp.secure=tls
smtp.username=<encrypted>
smtp.password=<encrypted>
smtp.from=noreply@example.com
smtp.from.name=My Site
`+"```"+`

Format supports both `+"`key=value`"+` and `+"`key: value`"+` syntax. Lists can be comma-separated or YAML-style with `+"`- item`"+` on following lines.

## Project Structure

`+"```"+`
my-site/
├── index.php              # Main entry / router
├── site.properties        # Configuration
├── assets/
│   ├── css/style.css
│   └── js/
├── includes/              # PHP partials
├── about.php              # Additional pages (clean URLs)
└── .github/
    └── workflows/
        └── deploy.yml     # CI/CD (optional, via lightspeed add git)
`+"```"+`

## Clean URL Routing

The server routes requests to PHP files without the `+"`.php`"+` extension:
- `+"`/about`"+` → `+"`about.php`"+`
- `+"`/contact`"+` → `+"`contact.php`"+`
- `+"`/blog/post`"+` → `+"`blog/post.php`"+`

Static assets in `+"`assets/`"+` are served directly.

## Deployment

1. Code is built into a Docker image based on `+"`lightspeed-server`"+`
2. Image is pushed to the Lightspeed registry
3. Platform deploys to DigitalOcean App Platform
4. Site is available at `+"`https://{name}.lightspeed.ee`"+`
5. Custom domains can be configured via the `+"`domain`"+`/`+"`domains`"+` property

Git tag versions are used for image tags automatically. Falls back to "latest".
`, skillVersionPrefix, Version, skillVersionSuffix)
}
