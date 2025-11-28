package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"lightspeed/internal/ui"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Lightspeed project",
	Long:  "Create a new PHP project with basic directory structure and index.php",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader(Version)

		dir, err := os.Getwd()
		if err != nil {
			ui.PrintError("Failed to get current directory: %v", err)
			os.Exit(1)
		}

		// Check if index.php already exists
		indexPath := filepath.Join(dir, "index.php")
		if _, err := os.Stat(indexPath); err == nil {
			ui.PrintWarning("index.php already exists")
			os.Exit(1)
		}

		// Create directories
		dirs := []string{"assets", "assets/css", "assets/js", "includes"}
		for _, d := range dirs {
			path := filepath.Join(dir, d)
			if err := os.MkdirAll(path, 0755); err != nil {
				ui.PrintError("Failed to create directory %s: %v", d, err)
				os.Exit(1)
			}
		}

		// Create index.php
		indexContent := `<?php
/**
 * Lightspeed Project
 */
?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Lightspeed</title>
    <link rel="stylesheet" href="assets/css/style.css">
</head>
<body>
    <h1>Hello, World!</h1>
</body>
</html>
`
		if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
			ui.PrintError("Failed to create index.php: %v", err)
			os.Exit(1)
		}

		// Create basic CSS file
		cssContent := `/* Lightspeed Styles */

* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
    line-height: 1.6;
    padding: 2rem;
}

h1 {
    color: #333;
}
`
		cssPath := filepath.Join(dir, "assets", "css", "style.css")
		if err := os.WriteFile(cssPath, []byte(cssContent), 0644); err != nil {
			ui.PrintWarning("Failed to create style.css: %v", err)
		}

		// Create .gitignore
		gitignoreContent := `.DS_Store
*.log
`
		gitignorePath := filepath.Join(dir, ".gitignore")
		os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)

		// Print success
		ui.PrintSuccess("Created Lightspeed project")
		fmt.Println()
		ui.PrintInfo("Files created:")
		fmt.Println("  • index.php")
		fmt.Println("  • assets/css/style.css")
		fmt.Println("  • assets/js/")
		fmt.Println("  • includes/")
		fmt.Println()
		ui.PrintInfo("Run 'lightspeed start' to start the development server")
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
