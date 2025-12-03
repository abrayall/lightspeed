package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"lightspeed/core/lib/properties"
	"lightspeed/core/lib/ui"
	"lightspeed/core/lib/version"
)

var (
	buildTag   string
	buildImage string
)

// getBaseImage returns the appropriate base image for building
func getBaseImage() string {
	if buildImage != "" && buildImage != "php:8.2-apache" {
		return buildImage // User specified a custom image
	}

	// If version is "dev" or contains timestamp/commit info, use latest
	if Version == "dev" || strings.Contains(Version, "-") {
		return defaultServerImage + ":latest"
	}

	// Otherwise use the matching version tag
	return defaultServerImage + ":" + Version
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a Docker container for the project",
	Long:  "Build and tag a Docker container with the PHP project",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader(Version)

		dir, err := os.Getwd()
		if err != nil {
			ui.PrintError("Failed to get current directory: %v", err)
			os.Exit(1)
		}

		projectName := filepath.Base(dir)
		imageName := sanitizeContainerName(projectName)

		// Load site info from site.properties
		siteInfo, err := loadSiteInfo(dir)
		if err != nil {
			ui.PrintError("Failed to load site.properties: %v", err)
			os.Exit(1)
		}

		// Get site name
		siteName := imageName
		var domains []string
		if siteInfo != nil {
			if siteInfo.Name != "" {
				siteName = siteInfo.Name
			}
			domains = siteInfo.Domains
		}

		// Determine tag
		tag := buildTag
		if tag == "" {
			// Try to get version from git
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

		fullImageName := fmt.Sprintf("%s:%s", siteName, tag)

		printSiteInfo(siteName, tag, domains)

		// Check if Dockerfile exists, create if not
		dockerfilePath := filepath.Join(dir, "Dockerfile")
		createdDockerfile := false
		if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
			ui.PrintInfo("Creating Dockerfile...")
			if err := createDockerfile(dockerfilePath); err != nil {
				ui.PrintError("Failed to create Dockerfile: %v", err)
				os.Exit(1)
			}
			createdDockerfile = true
		}

		ui.PrintInfo("Building Docker image...")
		fmt.Println()

		// Build the image for linux/amd64 platform
		dockerArgs := []string{
			"build",
			"--platform", "linux/amd64",
			"-t", fullImageName,
			".",
		}

		dockerCmd := exec.Command("docker", dockerArgs...)
		dockerCmd.Dir = dir
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr

		buildErr := dockerCmd.Run()

		// Clean up Dockerfile if we created it
		if createdDockerfile {
			os.Remove(dockerfilePath)
		}

		if buildErr != nil {
			ui.PrintError("Failed to build image: %v", buildErr)
			os.Exit(1)
		}

		fmt.Println()
		ui.PrintSuccess("Built image: %s", fullImageName)
		fmt.Println()
		ui.PrintInfo("Run with: docker run -p 8080:80 %s", fullImageName)
		fmt.Println()
	},
}

func createDockerfile(path string) error {
	baseImage := getBaseImage()
	content := fmt.Sprintf(`FROM %s

# Copy project files
COPY . /var/www/html/

# Set proper permissions
RUN chown -R www-data:www-data /var/www/html

# Expose port 80
EXPOSE 80
`, baseImage)

	return os.WriteFile(path, []byte(content), 0644)
}

// SiteInfo holds information about a site from site.properties
type SiteInfo struct {
	Name    string
	Domains []string
}

// loadSiteInfo loads site information from site.properties
func loadSiteInfo(dir string) (*SiteInfo, error) {
	propsPath := filepath.Join(dir, "site.properties")
	if !properties.FileExists(propsPath) {
		return nil, nil
	}

	props, err := properties.ParseProperties(propsPath)
	if err != nil {
		return nil, err
	}

	info := &SiteInfo{}

	// Get site name
	name := props.Get("name")
	if name != "" {
		info.Name = sanitizeContainerName(name)
	}

	// Get domains
	domain := props.Get("domain")
	if domain != "" {
		info.Domains = append(info.Domains, domain)
	}
	domainsList := props.GetList("domains")
	info.Domains = append(info.Domains, domainsList...)

	return info, nil
}

// printSiteInfo prints site information
func printSiteInfo(siteName string, version string, domains []string) {
	ui.PrintKeyValue("Site", siteName)
	ui.PrintKeyValue("Version", version)
	if len(domains) == 1 {
		ui.PrintKeyValue("Domain", domains[0])
		fmt.Println()
	} else if len(domains) > 1 {
		ui.PrintKeyValue("Domains", strings.Join(domains, ", "))
		fmt.Println()
	}
}

func init() {
	buildCmd.Flags().StringVarP(&buildTag, "tag", "t", "", "Tag for the image (default: git version or 'latest')")
	buildCmd.Flags().StringVarP(&buildImage, "image", "i", "", "Base Docker image to use (default: lightspeed-server)")

	rootCmd.AddCommand(buildCmd)
}
