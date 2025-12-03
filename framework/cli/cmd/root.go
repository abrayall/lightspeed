package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"lightspeed/core/lib/ui"
)

// Version is set by ldflags during build
var Version = "dev"

// Default hosts
const (
	defaultRegistryHost = "registry.lightspeed.ee"
	defaultAPIHost      = "api.lightspeed.ee"
)

// Shared hosts for deploy/publish commands
var (
	apiHostOverride string // Set by --api flag
	registryHost    string // Computed: override or default
	apiHost         string // Computed: override or default
)

var rootCmd = &cobra.Command{
	Use:   "lightspeed",
	Short: "Lightweight rapid development tool for PHP websites",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Long = ui.Divider() + "\n" + ui.Banner() + "\n" + ui.VersionLine(Version) + "\n\n" + ui.Divider() + "\n\nA lightweight, rapid development tool for small PHP websites"
	rootCmd.AddCommand(versionCmd)

	// Hidden --api flag for testing (shared by deploy/publish)
	// When set, overrides both registry and API hosts
	rootCmd.PersistentFlags().StringVar(&apiHostOverride, "api", "", "Override API and registry host:port")
	rootCmd.PersistentFlags().MarkHidden("api")

	// Set up pre-run to compute hosts after flags are parsed
	originalPreRun := rootCmd.PersistentPreRun
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Check env var first, then flag
		override := os.Getenv("LIGHTSPEED_API")
		if apiHostOverride != "" {
			override = apiHostOverride
		}

		if override != "" {
			// Use override for both
			registryHost = override
			apiHost = override
		} else {
			// Use separate defaults
			registryHost = defaultRegistryHost
			apiHost = defaultAPIHost
		}

		if originalPreRun != nil {
			originalPreRun(cmd, args)
		}
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("lightspeed %s\n", Version)
	},
}

// getDockerRegistryHost returns the registry host for Docker operations
// On macOS, localhost must be translated to host.docker.internal for Docker to reach the host
func getDockerRegistryHost() string {
	host := registryHost

	// Docker Desktop runs in a VM, so localhost doesn't work
	// Translate localhost to host.docker.internal
	if strings.HasPrefix(host, "localhost:") {
		return "host.docker.internal" + strings.TrimPrefix(host, "localhost")
	}
	if host == "localhost" {
		return "host.docker.internal"
	}
	if strings.HasPrefix(host, "127.0.0.1:") {
		return "host.docker.internal" + strings.TrimPrefix(host, "127.0.0.1")
	}
	if host == "127.0.0.1" {
		return "host.docker.internal"
	}

	return host
}

// getAPIURL returns the full API URL with correct scheme based on port
// Port 8443 or no port -> HTTPS, otherwise HTTP
func getAPIURL() string {
	host := apiHost

	// Check if host has a port
	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		port := parts[len(parts)-1]

		// Use HTTPS for 8443, HTTP for other explicit ports
		if port == "8443" {
			return "https://" + host
		}
		return "http://" + host
	}

	// No port specified, use HTTPS (default 443)
	return "https://" + host
}
