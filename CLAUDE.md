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
