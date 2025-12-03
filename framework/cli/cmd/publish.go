package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"lightspeed/core/lib/ui"
	"lightspeed/core/lib/version"
)

var (
	publishTag  string
	publishName string
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Build and push Docker image to registry",
	Long:  "Build the Docker image and push to the Lightspeed registry",
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

		// Get site name (--name flag takes precedence, then site.properties, then directory name)
		siteName := publishName
		var domains []string
		if siteName == "" {
			siteName = imageName
			if siteInfo != nil && siteInfo.Name != "" {
				siteName = siteInfo.Name
			}
		}
		if siteInfo != nil {
			domains = siteInfo.Domains
		}

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

		// Registry image names (use Docker-specific host for Docker operations)
		// Use siteName for the image name (respects --name flag)
		dockerRegistry := getDockerRegistryHost()
		registryBase := fmt.Sprintf("%s/%s", dockerRegistry, siteName)
		versionImage := fmt.Sprintf("%s:%s", registryBase, tag)
		latestImage := fmt.Sprintf("%s:latest", registryBase)

		printSiteInfo(siteName, tag, domains)
		ui.PrintKeyValue("Registry", dockerRegistry)
		ui.PrintKeyValue("Platform", apiHost)
		fmt.Println()

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

		// Build the image
		ui.PrintInfo("Building Docker image...")
		fmt.Println()

		buildArgs := []string{
			"build",
			"--platform", "linux/amd64",
			"-t", versionImage,
			"-t", latestImage,
			".",
		}

		dockerBuildCmd := exec.Command("docker", buildArgs...)
		dockerBuildCmd.Dir = dir
		dockerBuildCmd.Stdout = os.Stdout
		dockerBuildCmd.Stderr = os.Stderr

		buildErr := dockerBuildCmd.Run()

		// Clean up Dockerfile if we created it
		if createdDockerfile {
			os.Remove(dockerfilePath)
		}

		if buildErr != nil {
			ui.PrintError("Failed to build image: %v", buildErr)
			os.Exit(1)
		}

		fmt.Println()
		ui.PrintSuccess("Built image: %s", versionImage)
		fmt.Println()

		// Auto-login to registry
		ui.PrintInfo("Logging in to registry...")
		if err := dockerLogin(dockerRegistry); err != nil {
			ui.PrintError("Failed to login to registry: %v", err)
			os.Exit(1)
		}

		// Push specific tags we just built
		ui.PrintInfo("Pushing images...")
		if err := pushImage(versionImage); err != nil {
			ui.PrintError("Failed to push image: %v", err)
			os.Exit(1)
		}
		if tag != "latest" {
			if err := pushImage(latestImage); err != nil {
				ui.PrintError("Failed to push image: %v", err)
				os.Exit(1)
			}
		}

		fmt.Println()
		ui.PrintSuccess("Published successfully!")
		fmt.Println()
		ui.PrintInfo("Published tags:")
		fmt.Printf("  • %s\n", versionImage)
		if tag != "latest" {
			fmt.Printf("  • %s\n", latestImage)
		}
		fmt.Println()
	},
}

func dockerLogin(registry string) error {
	cmd := exec.Command("docker", "login", registry, "-u", "lightspeed", "--password-stdin")
	cmd.Stdin = strings.NewReader("lightspeed")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func pushImage(image string) error {
	fmt.Printf("• Pushing %s...\n", image)
	cmd := exec.Command("docker", "push", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() {
	publishCmd.Flags().StringVarP(&publishTag, "tag", "t", "", "Version tag (default: git version or 'latest')")
	publishCmd.Flags().StringVarP(&publishName, "name", "n", "", "Site name (default: project directory name)")

	rootCmd.AddCommand(publishCmd)
}
