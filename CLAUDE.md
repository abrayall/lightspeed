# Lightspeed

A lightweight, rapid development CLI tool for small PHP websites built in Go.

## Project Structure

- `main.go` - Entry point
- `cmd/` - Cobra command implementations
  - `root.go` - Root command with banner and version
  - `init.go` - Initialize new project
  - `run.go` - Start/stop development server
  - `build.go` - Build Docker container
  - `publish.go` - Push to DigitalOcean registry
  - `deploy.go` - Deploy to DigitalOcean App Platform
- `internal/ui/` - Terminal styling (colors, banner, output formatting)
- `internal/version/` - Git tag version parsing
- `build.sh` - Multi-platform build script
- `install.sh` - Installation script

## Build & Run

```bash
go run .                    # Run directly
go build -o lightspeed .    # Build binary
./build.sh                  # Build for all platforms
./install.sh                # Install to /usr/local/bin
```

## Dependencies

- github.com/spf13/cobra - CLI framework
- github.com/charmbracelet/lipgloss - Terminal styling

## Commit Messages

- Single line only
- No Claude references or co-author tags
- No emojis
- Use imperative mood (e.g., "Add deploy command" not "Added deploy command")

## Code Style

- Follow standard Go conventions
- Commands print header with `ui.PrintHeader(Version)`
- Use `ui.PrintSuccess`, `ui.PrintError`, `ui.PrintInfo` for output
- Use `ui.PrintKeyValue` for key-value pairs

## Platform Components

### Operator (platform/operator)
- Registry proxy at `/v2/*` - accepts any credentials, authenticates to DO registry
- Sites API at `/sites/*` - CRUD for DO App Platform deployments
- Image pruner - runs daily, keeps latest + 3 highest semver versions per repo
- TLS support with auto-generated self-signed certs

### CLI (framework/cli)
- `lightspeed init` - Initialize project
- `lightspeed run` - Local dev server
- `lightspeed build` - Build Docker image
- `lightspeed publish` - Push to registry
- `lightspeed deploy` - Deploy to DO App Platform

---

## Future Work

### Authentication & User Accounts

**Architecture:**
```
┌──────────┐      ┌──────────────┐      ┌──────────┐
│   CLI    │─────▶│   Operator   │─────▶│ Storage  │
│          │ API  │              │      │ (S3/DO   │
│          │ Key  │              │      │  Spaces) │
└──────────┘      └──────────────┘      └──────────┘
                         │
                         ▼
                  ┌──────────────┐
                  │    Stripe    │
                  └──────────────┘
```

**Data Model:**
```
users/
  {user_id}.json
    - id, email, password_hash, stripe_customer_id, plan, created_at

keys/
  {key_prefix}.json
    - key_hash -> user_id mapping

sites/
  {site_name}.json
    - id, user_id, name, app_id, created_at

indexes/
  email_to_user.json
```

**Flow:**
1. User signs up via CLI (`lightspeed register`) or web
2. Creates API key (`lightspeed login`)
3. CLI stores key in `~/.lightspeed/config`
4. All API calls include `Authorization: Bearer <key>`
5. Operator validates key, checks user's plan limits

### Storage: S3/Spaces with JSON

**Approach:**
- Store all persistent data as JSON files in DO Spaces (S3-compatible)
- Simple, no database to manage, cheap
- Works well for < 1000 users
- Easy to inspect/debug

**Concerns & Mitigations:**
- API key lookup latency: Cache keys in memory
- Race conditions: Use optimistic locking or accept eventual consistency
- No complex queries: Maintain manual indexes

### Caching

**Recommended: eko/gocache**
- Supports in-memory (go-cache, ristretto) and Redis backends
- Same interface, easy to swap

**Implementation:**
```go
// config flags
--cache=memory|redis
--redis=localhost:6379

// Usage
cache.Set(ctx, "key", value, store.WithExpiration(5*time.Minute))
val, _ := cache.Get(ctx, "key")
```

**Strategy:**
- Start with in-memory caching
- Add Redis when scaling to multiple operator instances
- Cache API keys (1hr TTL) and user data (5min TTL)

### Stripe Integration

**Features:**
- Subscription billing (monthly/yearly)
- Free tier: 1 site, limited resources
- Paid tier: more sites, custom domains, etc.

**Webhooks to handle:**
- `invoice.payment_succeeded`
- `invoice.payment_failed`
- `customer.subscription.updated`
- `customer.subscription.deleted`

**Flow:**
1. On signup, create Stripe customer
2. User adds payment via Stripe Checkout/Portal
3. Webhook updates user's plan in storage
4. Operator checks plan limits before allowing actions

### CLI Authentication Commands

```bash
lightspeed register              # Create account
lightspeed login                 # Authenticate and store API key
lightspeed logout                # Remove stored credentials
lightspeed account               # Show account info
lightspeed keys                  # List API keys
lightspeed keys create           # Create new API key
lightspeed keys revoke <id>      # Revoke an API key
```
