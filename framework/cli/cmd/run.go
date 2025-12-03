package cmd

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"lightspeed/core/lib/ui"
)

var (
	runPort  int
	runImage string
)

// Default server image from GitHub Container Registry
const defaultServerImage = "ghcr.io/abrayall/lightspeed-server"

// getServerImage returns the appropriate server image based on CLI version
func getServerImage() string {
	if runImage != "" && runImage != "php:8.2-apache" {
		return runImage // User specified a custom image
	}

	// If version is "dev" or contains timestamp/commit info, use latest
	if Version == "dev" || strings.Contains(Version, "-") {
		return defaultServerImage + ":latest"
	}

	// Otherwise use the matching version tag
	return defaultServerImage + ":" + Version
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a PHP development server",
	Long:  "Start a PHP container with the current directory mounted as a volume",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader(Version)

		dir, err := os.Getwd()
		if err != nil {
			ui.PrintError("Failed to get current directory: %v", err)
			os.Exit(1)
		}

		// Get project name from directory
		projectName := filepath.Base(dir)
		containerName := fmt.Sprintf("lightspeed-%s", sanitizeContainerName(projectName))

		// Check if container is already running
		if isContainerRunning(containerName) {
			ui.PrintWarning("Container %s is already running", containerName)
			ui.PrintInfo("Stop it with: lightspeed stop")
			os.Exit(1)
		}

		// Remove any existing stopped container with same name
		stopContainer(containerName)

		// Use specified port or find an available one
		port := runPort
		if port == 0 {
			port = findAvailablePort()
			if port == 0 {
				ui.PrintError("No available ports found in range 9000-9099")
				os.Exit(1)
			}
		}

		ui.PrintInfo("Starting development server...")
		fmt.Println()

		// Run PHP container with Apache
		serverImage := getServerImage()
		dockerArgs := []string{
			"run",
			"-d",
			"--name", containerName,
			"-p", fmt.Sprintf("%d:80", port),
			"-v", fmt.Sprintf("%s:/var/www/html", dir),
			serverImage,
		}

		dockerCmd := exec.Command("docker", dockerArgs...)
		output, err := dockerCmd.CombinedOutput()
		if err != nil {
			ui.PrintError("Failed to start container: %v", err)
			ui.PrintError("%s", string(output))
			os.Exit(1)
		}

		url := fmt.Sprintf("http://localhost:%d", port)

		ui.PrintSuccess("Development server started")
		fmt.Println()
		ui.PrintKeyValue("  URL", url)
		ui.PrintKeyValue("  Container", containerName)
		fmt.Println()

		// Wait for server to be ready and open browser
		if waitForServer(url, 30) {
			openBrowser(url)
		}

		ui.PrintInfo("Run 'lightspeed stop' to stop the server")
		fmt.Println()
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the PHP development server",
	Long:  "Stop and remove the running PHP container",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader(Version)

		dir, err := os.Getwd()
		if err != nil {
			ui.PrintError("Failed to get current directory: %v", err)
			os.Exit(1)
		}

		projectName := filepath.Base(dir)
		containerName := fmt.Sprintf("lightspeed-%s", sanitizeContainerName(projectName))

		if !isContainerRunning(containerName) {
			ui.PrintWarning("No running container found for this project")
			os.Exit(0)
		}

		ui.PrintInfo("Stopping development server...")

		if stopContainer(containerName) {
			ui.PrintSuccess("Development server stopped")
		} else {
			ui.PrintError("Failed to stop container")
			os.Exit(1)
		}
	},
}

func isContainerRunning(name string) bool {
	cmd := exec.Command("docker", "ps", "-q", "-f", fmt.Sprintf("name=%s", name))
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}

func stopContainer(name string) bool {
	// Stop container if running
	exec.Command("docker", "stop", name).Run()
	// Remove container
	err := exec.Command("docker", "rm", name).Run()
	return err == nil || !containerExists(name)
}

func containerExists(name string) bool {
	cmd := exec.Command("docker", "ps", "-aq", "-f", fmt.Sprintf("name=%s", name))
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}

func sanitizeContainerName(name string) string {
	// Docker container names can only contain [a-zA-Z0-9_.-]
	result := strings.ToLower(name)
	result = strings.ReplaceAll(result, " ", "-")
	var sanitized []rune
	for _, r := range result {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			sanitized = append(sanitized, r)
		}
	}
	return string(sanitized)
}

func findAvailablePort() int {
	for port := 9000; port < 9100; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return port
		}
	}
	return 0
}

func waitForServer(url string, timeoutSeconds int) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch {
	case isCommandAvailable("open"):
		cmd = exec.Command("open", url)
	case isCommandAvailable("xdg-open"):
		cmd = exec.Command("xdg-open", url)
	case isCommandAvailable("start"):
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	cmd.Run()
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func init() {
	startCmd.Flags().IntVarP(&runPort, "port", "p", 0, "Port to expose (default: auto-detect in 9000 range)")
	startCmd.Flags().StringVarP(&runImage, "image", "i", "", "Docker image to use (default: lightspeed-server)")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
}
