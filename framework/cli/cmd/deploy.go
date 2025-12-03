package cmd

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"lightspeed/core/lib/ui"
	"lightspeed/core/lib/version"
)

// SiteStatus represents the status response from the API
type SiteStatus struct {
	Name   string   `json:"name"`
	Status string   `json:"status"`
	URLs   []string `json:"urls"`
}

var (
	deploySiteName string
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Build and deploy to Lightspeed",
	Long:  "Build, push to registry, and deploy via Lightspeed operator",
	Run: func(cmd *cobra.Command, args []string) {
		// Note: buildCmd.Run prints the header

		dir, err := os.Getwd()
		if err != nil {
			ui.PrintError("Failed to get current directory: %v", err)
			os.Exit(1)
		}

		projectName := filepath.Base(dir)
		imageName := sanitizeContainerName(projectName)

		// Determine version tag
		tag := publishTag
		if tag == "" {
			if version.IsGitRepo(dir) {
				v, err := version.GetFromGit(dir)
				if err == nil {
					tag = v.String()
				}
			}
			if tag == "" {
				tag = "latest"
			}
		}

		// Get site name (default to project name)
		siteName := deploySiteName
		if siteName == "" {
			siteName = imageName
		}

		// Set the publish name flag so publish command uses it
		publishName = deploySiteName

		// Step 1: Build and push the image (prints header and initial info including site and platform)
		publishCmd.Run(cmd, args)

		// Step 2: Check if site exists
		apiURL := getAPIURL()
		ui.PrintInfo("Checking site '%s'...", siteName)
		exists, err := siteExists(apiURL, siteName)
		if err != nil {
			ui.PrintError("Failed to check site: %v", err)
			os.Exit(1)
		}

		if !exists {
			// Create new site
			ui.PrintInfo("Creating site '%s'...", siteName)
			// Use siteName for image because that's what publish command uses
			err = createSite(apiURL, siteName, siteName, tag)
			if err != nil {
				ui.PrintError("Failed to create site: %v", err)
				os.Exit(1)
			}
			ui.PrintSuccess("Created site '%s'", siteName)

			// Wait for deployment to complete (new sites need to wait)
			fmt.Println()
			siteURL, err := waitForDeployment(apiURL, siteName)
			if err != nil {
				ui.PrintError("Deployment failed: %v", err)
				os.Exit(1)
			}

			// Wait for site to respond
			if siteURL != "" {
				fmt.Println()
				if err := waitForURLReady(siteURL); err != nil{
					ui.PrintError("Site deployment completed but URL not responding: %v", err)
					fmt.Println()
					ui.PrintKeyValue("URL", siteURL)
					os.Exit(1)
				}

				// Open browser
				fmt.Println()
				ui.PrintInfo("Opening browser...")
				openBrowser(siteURL)

				// Final success message
				fmt.Println()
				ui.PrintSuccess("Deployed successfully!")
				fmt.Printf("  %s\n", siteURL)
			}
		} else {
			// Existing site - deploy_on_push triggers deployment automatically
			// Wait for deployment to complete
			ui.PrintInfo("Deployment triggered by image push")

			fmt.Println()
			siteURL, err := waitForRedeployment(apiURL, siteName)
			if err != nil {
				ui.PrintError("Deployment failed: %v", err)
				os.Exit(1)
			}

			// Wait for site to respond
			if siteURL != "" {
				fmt.Println()
				if err := waitForURLReady(siteURL); err != nil {
					ui.PrintError("Site deployment completed but URL not responding: %v", err)
					fmt.Println()
					ui.PrintKeyValue("URL", siteURL)
					os.Exit(1)
				}

				// Open browser
				fmt.Println()
				ui.PrintInfo("Opening browser...")
				openBrowser(siteURL)

				fmt.Println()
				ui.PrintSuccess("Deployed successfully!")
				fmt.Printf("  %s\n", siteURL)
			}
		}
		fmt.Println()
	},
}

// siteExists checks if a site exists via the operator API
func siteExists(operatorURL, name string) (bool, error) {
	url := fmt.Sprintf("%s/sites/%s", operatorURL, name)
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// createSite creates a new site via the operator API
func createSite(operatorURL, name, image, tag string) error {
	url := fmt.Sprintf("%s/sites", operatorURL)

	payload := map[string]string{
		"name":  name,
		"image": image,
		"tag":   tag,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(respBody))
	}

	return nil
}

// triggerDeploy triggers a deployment via the operator API
func triggerDeploy(operatorURL, name string) error {
	url := fmt.Sprintf("%s/sites/%s/deploy", operatorURL, name)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(respBody))
	}

	return nil
}

// getSiteStatus gets the current status of a site
func getSiteStatus(operatorURL, name string) (*SiteStatus, error) {
	url := fmt.Sprintf("%s/sites/%s", operatorURL, name)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var status SiteStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}

// getDigitalOceanURL extracts the .ondigitalocean.app URL from a list of URLs
func getDigitalOceanURL(urls []string) string {
	for _, url := range urls {
		if strings.Contains(url, ".ondigitalocean.app") {
			return url
		}
	}
	// Fallback to first URL if no DO URL found
	if len(urls) > 0 {
		return urls[0]
	}
	return ""
}

// waitForRedeployment waits for an existing app to redeploy (DEPLOYING → ACTIVE)
func waitForRedeployment(operatorURL, name string) (string, error) {
	ui.PrintInfo("Waiting for deployment...")

	lastStatus := ""
	sawDeploying := false
	firstActiveTime := time.Time{}
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("deployment timed out after 5 minutes")
		case <-ticker.C:
			status, err := getSiteStatus(operatorURL, name)
			if err != nil {
				continue
			}

			// Show status change
			if status.Status != lastStatus {
				statusDisplay := formatStatus(status.Status)
				ui.PrintKeyValue("  Status", statusDisplay)
				lastStatus = status.Status
			}

			// Track if we've seen deploying state
			// SUPERSEDED means old deployment was replaced by new one
			if status.Status == "DEPLOYING" || status.Status == "PENDING_DEPLOY" || status.Status == "BUILDING" || status.Status == "PENDING_BUILD" || status.Status == "SUPERSEDED" {
				sawDeploying = true
				firstActiveTime = time.Time{} // Reset active timer
			}

			// If ACTIVE and we saw deploying, deployment is complete
			if status.Status == "ACTIVE" && sawDeploying {
				return getDigitalOceanURL(status.URLs), nil
			}

			// If ACTIVE but no deploying state seen yet, track how long it's been ACTIVE
			// After 30 seconds of ACTIVE without seeing deploying, assume no deployment needed
			if status.Status == "ACTIVE" && !sawDeploying {
				if firstActiveTime.IsZero() {
					firstActiveTime = time.Now()
				} else if time.Since(firstActiveTime) > 30*time.Second {
					ui.PrintInfo("No new deployment detected (already up to date)")
					return getDigitalOceanURL(status.URLs), nil
				}
			}

			// Handle failures
			if status.Status == "ERROR" || status.Status == "FAILED" {
				return "", fmt.Errorf("deployment failed with status: %s", status.Status)
			}
		}
	}
}

// waitForDeployment polls for deployment status and shows progress (new sites)
func waitForDeployment(operatorURL, name string) (string, error) {
	ui.PrintInfo("Waiting for deployment...")

	lastStatus := ""
	timeout := time.After(10 * time.Minute)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("deployment timed out after 10 minutes")
		case <-ticker.C:
			status, err := getSiteStatus(operatorURL, name)
			if err != nil {
				// Might not be ready yet, continue polling
				continue
			}

			// Show status change
			if status.Status != lastStatus {
				statusDisplay := formatStatus(status.Status)
				ui.PrintKeyValue("  Status", statusDisplay)
				lastStatus = status.Status
			}

			// Check for terminal states
			switch status.Status {
			case "ACTIVE":
				return getDigitalOceanURL(status.URLs), nil
			case "ERROR", "FAILED":
				return "", fmt.Errorf("deployment failed with status: %s", status.Status)
			case "CANCELED":
				return "", fmt.Errorf("deployment was canceled")
			}
		}
	}
}

// waitForURLReady does a quick check to see if the URL is responding
func waitForURLReady(siteURL string) error {
	ui.PrintInfo("Waiting for site to respond...")
	maxAttempts := 60 // 60 attempts * 5 seconds = 5 minutes
	retryDelay := 5 * time.Second

	// Parse hostname from URL
	var hostname string
	if strings.HasPrefix(siteURL, "https://") {
		hostname = strings.TrimPrefix(siteURL, "https://")
	} else if strings.HasPrefix(siteURL, "http://") {
		hostname = strings.TrimPrefix(siteURL, "http://")
	}
	hostname = strings.Split(hostname, "/")[0]

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Use Google's DNS (8.8.8.8) to resolve hostname and get IP
		resolver := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 10 * time.Second}
				return d.DialContext(ctx, network, "8.8.8.8:53")
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		ips, err := resolver.LookupHost(ctx, hostname)
		cancel()

		if err != nil || len(ips) == 0 {
			// DNS not propagated yet
			if attempt%6 == 0 {
				ui.PrintInfo("DNS not yet propagated, retrying...")
			}
			if attempt < maxAttempts {
				time.Sleep(retryDelay)
			}
			continue
		}

		// Got IP! Now check if site responds
		ip := ips[0]

		// Create HTTP request to IP with Host header set to hostname
		req, _ := http.NewRequest("GET", "https://"+ip+"/", nil)
		req.Host = hostname

		client := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					ServerName:         hostname, // For SNI
				},
			},
		}

		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 400 {
				return nil
			}
			// Show status code if not in success range
			if attempt%6 == 0 { // Log every 30 seconds
				ui.PrintInfo("Site returned status %d, still waiting...", resp.StatusCode)
			}
		} else {
			// Log errors occasionally
			if attempt%6 == 0 { // Log every 30 seconds
				ui.PrintInfo("Connection error: %v, retrying...", err)
			}
		}

		if attempt < maxAttempts {
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("site did not respond with 200 after %d attempts (5 minutes)", maxAttempts)
}

// formatStatus returns a human-readable status
func formatStatus(status string) string {
	switch status {
	case "PENDING_BUILD":
		return "Pending build..."
	case "BUILDING":
		return "Building..."
	case "PENDING_DEPLOY":
		return "Pending deploy..."
	case "DEPLOYING":
		return "Deploying..."
	case "SUPERSEDED":
		return "Redeploying..."
	case "ACTIVE":
		return "Active"
	case "ERROR", "FAILED":
		return "Failed"
	case "CANCELED":
		return "Canceled"
	default:
		return status
	}
}

func init() {
	deployCmd.Flags().StringVarP(&deploySiteName, "name", "n", "", "Site name (default: project directory name)")

	rootCmd.AddCommand(deployCmd)
}
