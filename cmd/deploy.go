package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"lightspeed/internal/ui"
	"lightspeed/internal/version"
)

var (
	deployToken   string
	deployAppName string
	deployRegion  string
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Publish and deploy to DigitalOcean App Platform",
	Long:  "Build, push to registry, and trigger deployment on DigitalOcean App Platform",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader(Version)

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

		// Get app name (default to project name)
		appName := deployAppName
		if appName == "" {
			appName = imageName
		}

		ui.PrintKeyValue("Project", projectName)
		ui.PrintKeyValue("Version", tag)
		ui.PrintKeyValue("App", appName)
		fmt.Println()

		// Check for API token
		token := deployToken
		if token == "" {
			token = os.Getenv("DIGITALOCEAN_TOKEN")
		}
		if token == "" {
			token = getDefaultToken()
		}

		// Step 1: Publish
		ui.PrintInfo("Publishing image...")
		publishCmd.Run(cmd, args)

		// Step 2: Find or create app
		ui.PrintInfo("Finding app '%s'...", appName)
		appID, err := findAppByName(token, appName)
		if err != nil {
			ui.PrintError("Failed to find app: %v", err)
			os.Exit(1)
		}

		if appID == "" {
			// Create new app
			ui.PrintInfo("App not found, creating '%s'...", appName)
			appID, err = createApp(token, appName, imageName, deployRegion)
			if err != nil {
				ui.PrintError("Failed to create app: %v", err)
				os.Exit(1)
			}
			ui.PrintSuccess("Created app '%s'", appName)
		} else {
			// Create deployment for existing app
			ui.PrintInfo("Creating deployment...")
			deploymentID, err := createDeployment(token, appID)
			if err != nil {
				ui.PrintError("Failed to create deployment: %v", err)
				os.Exit(1)
			}
			ui.PrintKeyValue("Deployment ID", deploymentID)
		}

		fmt.Println()
		ui.PrintSuccess("Deployed successfully!")
		fmt.Println()
		ui.PrintKeyValue("App ID", appID)
		ui.PrintKeyValue("App", appName)
		fmt.Println()
	},
}

func findAppByName(token, name string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.digitalocean.com/v2/apps", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Apps []struct {
			ID   string `json:"id"`
			Spec struct {
				Name string `json:"name"`
			} `json:"spec"`
		} `json:"apps"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	for _, app := range result.Apps {
		if app.Spec.Name == name {
			return app.ID, nil
		}
	}

	return "", nil
}

func createDeployment(token, appID string) (string, error) {
	url := fmt.Sprintf("https://api.digitalocean.com/v2/apps/%s/deployments", appID)

	// Force rebuild
	payload := map[string]interface{}{
		"force_build": true,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(respBody))
	}

	var result struct {
		Deployment struct {
			ID string `json:"id"`
		} `json:"deployment"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Deployment.ID, nil
}

func createApp(token, appName, imageName, region string) (string, error) {
	if region == "" {
		region = "nyc"
	}

	// Build app spec
	spec := map[string]interface{}{
		"name":   appName,
		"region": region,
		"features": []string{
			"buildpack-stack=ubuntu-22",
		},
		"alerts": []map[string]string{
			{"rule": "DEPLOYMENT_FAILED"},
			{"rule": "DOMAIN_FAILED"},
		},
		"ingress": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"component": map[string]string{
						"name": appName,
					},
					"match": map[string]interface{}{
						"path": map[string]string{
							"prefix": "/",
						},
					},
				},
			},
		},
		"services": []map[string]interface{}{
			{
				"name":      appName,
				"http_port": 80,
				"image": map[string]interface{}{
					"registry_type": "DOCR",
					"registry":      "lightspeed-images",
					"repository":    imageName,
					"tag":           "latest",
					"deploy_on_push": map[string]bool{
						"enabled": true,
					},
				},
				"instance_count":     1,
				"instance_size_slug": "apps-s-1vcpu-0.5gb",
			},
		},
	}

	payload := map[string]interface{}{
		"spec": spec,
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "https://api.digitalocean.com/v2/apps", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(respBody))
	}

	var result struct {
		App struct {
			ID string `json:"id"`
		} `json:"app"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.App.ID, nil
}

func init() {
	deployCmd.Flags().StringVarP(&deployToken, "token", "k", "", "DigitalOcean API token (or set DIGITALOCEAN_TOKEN)")
	deployCmd.Flags().StringVarP(&deployAppName, "app", "a", "", "App name (default: project directory name)")
	deployCmd.Flags().StringVarP(&deployRegion, "region", "r", "nyc", "DigitalOcean region")

	rootCmd.AddCommand(deployCmd)
}
