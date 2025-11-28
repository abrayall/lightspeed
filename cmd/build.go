package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"lightspeed/internal/ui"
	"lightspeed/internal/version"
)

var (
	buildTag   string
	buildImage string
)

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

		fullImageName := fmt.Sprintf("%s:%s", imageName, tag)

		ui.PrintKeyValue("Project", projectName)
		ui.PrintKeyValue("Version", tag)
		fmt.Println()

		// Check if Dockerfile exists, create if not
		dockerfilePath := filepath.Join(dir, "Dockerfile")
		if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
			ui.PrintInfo("Creating Dockerfile...")
			if err := createDockerfile(dockerfilePath, buildImage); err != nil {
				ui.PrintError("Failed to create Dockerfile: %v", err)
				os.Exit(1)
			}
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

		if err := dockerCmd.Run(); err != nil {
			ui.PrintError("Failed to build image: %v", err)
			os.Exit(1)
		}

		fmt.Println()
		ui.PrintSuccess("Built image: %s", fullImageName)
		fmt.Println()
		ui.PrintInfo("Run with: docker run -p 8080:80 %s", fullImageName)
		fmt.Println()
	},
}

func createDockerfile(path string, baseImage string) error {
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

func init() {
	buildCmd.Flags().StringVarP(&buildTag, "tag", "t", "", "Tag for the image (default: git version or 'latest')")
	buildCmd.Flags().StringVarP(&buildImage, "image", "i", "php:8.2-apache", "Base Docker image to use")

	rootCmd.AddCommand(buildCmd)
}
