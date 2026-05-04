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
// Priority: --image flag > site.properties image > CLI version default
func getBaseImage(siteImage string) string {
	if buildImage != "" {
		return resolveImage(buildImage)
	}
	return resolveImage(siteImage)
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

		// Get site image for Dockerfile
		siteImage := ""
		if siteInfo != nil {
			siteImage = siteInfo.Image
		}

		// Check if Dockerfile exists, create if not
		dockerfilePath := filepath.Join(dir, "Dockerfile")
		createdDockerfile := false
		if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
			ui.PrintInfo("Creating Dockerfile...")
			if err := createDockerfile(dockerfilePath, siteImage, tag, dir); err != nil {
				ui.PrintError("Failed to create Dockerfile: %v", err)
				os.Exit(1)
			}
			createdDockerfile = true
		}

		ui.PrintInfo("Building Docker image...")
		fmt.Println()

		// Build the image for linux/amd64 platform
		// Use --pull to always get the latest base image
		dockerArgs := []string{
			"build",
			"--pull",
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

func createDockerfile(path string, siteImage string, version string, dir string) error {
	baseImage := getBaseImage(siteImage)
	content := fmt.Sprintf(`FROM %s

# Copy project files
COPY . /var/www/html/

# Inject site version
RUN echo 'version=%s' > /var/www/html/version.properties
`, baseImage, version)

	// Add composer support if packages are configured
	if hasComposerPackages(dir) {
		content += `
# Install Composer dependencies
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer
RUN cd /var/www/html && composer install --no-dev --optimize-autoloader --no-interaction
RUN echo 'auto_prepend_file = /var/www/html/vendor/autoload.php' >> /usr/local/etc/php/conf.d/lightspeed.ini
`
	}

	// Add environment variables
	envVars := loadEnvironment(dir)
	if len(envVars) > 0 {
		content += "\n# Environment variables\n"
		for key, value := range envVars {
			content += fmt.Sprintf("ENV %s=%q\n", key, value)
		}
	}

	content += `
# Set proper permissions
RUN chown -R www-data:www-data /var/www/html

# Expose port 80
EXPOSE 80
`

	return os.WriteFile(path, []byte(content), 0644)
}

// SiteInfo holds information about a site from site.properties
type SiteInfo struct {
	Name        string
	Domains     []string
	Image       string
	Environment map[string]string
}

// resolveImage normalizes an image specification
// - empty string -> default based on CLI version
// - version only (e.g., "0.5.3" or "v0.5.3") -> ghcr.io/abrayall/lightspeed-server:version
// - full image name (contains "/" or ":") -> used as-is
func resolveImage(image string) string {
	if image == "" {
		// Default based on CLI version
		if Version == "dev" || strings.Contains(Version, "-") {
			return defaultServerImage + ":latest"
		}
		return defaultServerImage + ":" + Version
	}

	// If it contains "/" or ":", it's a full image reference - use as-is
	if strings.Contains(image, "/") || strings.Contains(image, ":") {
		return image
	}

	// Just a version like "0.5.3" or "v0.5.3" - prepend default image
	image = strings.TrimPrefix(image, "v")
	return defaultServerImage + ":" + image
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

	// Get base image
	info.Image = props.Get("image")

	// Get environment variables
	info.Environment = props.GetMap("environment")

	return info, nil
}

// loadEnvironment loads environment variables from site.properties, decrypting secrets
func loadEnvironment(dir string) map[string]string {
	siteInfo, err := loadSiteInfo(dir)
	if err != nil || siteInfo == nil || len(siteInfo.Environment) == 0 {
		return nil
	}

	salt := loadEncryptionSalt()
	resolved := make(map[string]string, len(siteInfo.Environment))
	for key, value := range siteInfo.Environment {
		resolved[key] = resolveEnvironmentValue(value, salt)
	}
	return resolved
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
