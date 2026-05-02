package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"lightspeed/core/lib/ui"
)

var addCmd = &cobra.Command{
	Use:   "add [feature]",
	Short: "Add features to an existing Lightspeed project",
	Long:  "Add optional features like GitHub Actions deploy workflow to an existing project",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader(Version)

		if len(args) == 0 {
			ui.PrintInfo("Usage: lightspeed add [feature]")
			fmt.Println()
			ui.PrintInfo("Available features:")
			fmt.Println("  git      GitHub Actions deploy workflow and .gitignore")
			fmt.Println("  claude   Claude Code support files (CLAUDE.md and skill)")
			fmt.Println()
			return
		}

		dir, err := os.Getwd()
		if err != nil {
			ui.PrintError("Failed to get current directory: %v", err)
			os.Exit(1)
		}

		// Verify this is a Lightspeed project
		propsPath := filepath.Join(dir, "site.properties")
		if _, err := os.Stat(propsPath); os.IsNotExist(err) {
			ui.PrintError("Not a Lightspeed project (missing site.properties)")
			fmt.Println()
			ui.PrintInfo("Run 'lightspeed init' to create a new project")
			fmt.Println()
			os.Exit(1)
		}

		switch args[0] {
		case "git":
			addGitSupport(dir)
		case "claude":
			created := addClaudeSupport(dir)
			if len(created) > 0 {
				ui.PrintSuccess("Added Claude Code support")
				fmt.Println()
				ui.PrintInfo("Files created:")
				for _, f := range created {
					fmt.Printf("  • %s\n", f)
				}
			} else {
				ui.PrintSuccess("Claude Code support already configured")
			}
			fmt.Println()
		default:
			ui.PrintError("Unknown feature: %s", args[0])
			fmt.Println()
			ui.PrintInfo("Available features:")
			fmt.Println("  git      GitHub Actions deploy workflow and .gitignore")
			fmt.Println("  claude   Claude Code support files (CLAUDE.md and skill)")
			fmt.Println()
		}
	},
}

func addGitSupport(dir string) {
	var created []string

	// Create GitHub Actions deploy workflow
	workflowDir := filepath.Join(dir, ".github", "workflows")
	if _, err := os.Stat(workflowDir); os.IsNotExist(err) {
		if err := os.MkdirAll(workflowDir, 0755); err != nil {
			ui.PrintWarning("Failed to create .github/workflows directory: %v", err)
		}
	}
	workflowPath := filepath.Join(workflowDir, "deploy.yml")
	if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
		workflowContent := `name: deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Login to GHCR
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.REGISTRY_TOKEN }}

      - name: Install Lightspeed
        run: curl -sfL https://raw.githubusercontent.com/abrayall/lightspeed/refs/heads/main/install.sh | sh -

      - name: Deploy
        run: lightspeed deploy
`
		if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
			ui.PrintWarning("Failed to create deploy.yml: %v", err)
		} else {
			created = append(created, ".github/workflows/deploy.yml")
		}
	}

	// Create .gitignore
	gitignorePath := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		gitignoreContent := `.DS_Store
*.log
lightspeed
`
		if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
			ui.PrintWarning("Failed to create .gitignore: %v", err)
		} else {
			created = append(created, ".gitignore")
		}
	}

	if len(created) > 0 {
		ui.PrintSuccess("Added git support")
		fmt.Println()
		ui.PrintInfo("Files created:")
		for _, f := range created {
			fmt.Printf("  • %s\n", f)
		}
	} else {
		ui.PrintSuccess("Git support already configured")
	}
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(addCmd)
}
