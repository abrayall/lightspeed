package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"lightspeed/core/lib/ui"
)

var (
	initName    string
	initDomains []string
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

		// Determine site name
		siteName := initName
		if siteName == "" {
			siteName = filepath.Base(dir)
		}
		siteName = sanitizeContainerName(siteName)

		// Determine domains
		domains := initDomains
		if len(domains) == 0 {
			domains = []string{siteName + ".com"}
		}

		// Track what we create
		var created []string

		// Create directories
		dirs := []string{"assets", "assets/css", "assets/js", "includes"}
		for _, d := range dirs {
			path := filepath.Join(dir, d)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				if err := os.MkdirAll(path, 0755); err != nil {
					ui.PrintError("Failed to create directory %s: %v", d, err)
					os.Exit(1)
				}
				created = append(created, d+"/")
			}
		}

		// Create index.php if it doesn't exist
		indexPath := filepath.Join(dir, "index.php")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
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
			created = append(created, "index.php")
		}

		// Create basic CSS file if it doesn't exist
		cssPath := filepath.Join(dir, "assets", "css", "style.css")
		if _, err := os.Stat(cssPath); os.IsNotExist(err) {
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
			if err := os.WriteFile(cssPath, []byte(cssContent), 0644); err != nil {
				ui.PrintWarning("Failed to create style.css: %v", err)
			} else {
				created = append(created, "assets/css/style.css")
			}
		}

		// Create site.properties if it doesn't exist
		propsPath := filepath.Join(dir, "site.properties")
		if _, err := os.Stat(propsPath); os.IsNotExist(err) {
			var propsContent string
			propsContent = fmt.Sprintf("name=%s\n", siteName)
			if len(domains) == 1 {
				propsContent += fmt.Sprintf("domain=%s\n", domains[0])
			} else {
				propsContent += fmt.Sprintf("domains=%s\n", strings.Join(domains, ","))
			}
			propsContent += "libraries=lightspeed\n"
			if err := os.WriteFile(propsPath, []byte(propsContent), 0644); err != nil {
				ui.PrintWarning("Failed to create site.properties: %v", err)
			} else {
				created = append(created, "site.properties")
			}
		}

		// Create .idea directory for PhpStorm
		ideaDir := filepath.Join(dir, ".idea")
		ideaCreated := false
		if _, err := os.Stat(ideaDir); os.IsNotExist(err) {
			if err := os.MkdirAll(ideaDir, 0755); err != nil {
				ui.PrintWarning("Failed to create .idea directory: %v", err)
			} else {
				ideaCreated = true
				created = append(created, ".idea/")
			}
		}

		// Resolve libraries and create/update php.xml
		if err := updateIdeaConfig(dir); err != nil {
			ui.PrintWarning("Failed to update .idea/php.xml: %v", err)
		} else if !ideaCreated {
			// .idea existed but we may have updated php.xml
		}

		// Create .gitignore if it doesn't exist
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

		// Print success
		if len(created) > 0 {
			ui.PrintSuccess("Initialized Lightspeed project")
			fmt.Println()
			ui.PrintKeyValue("Name", siteName)
			ui.PrintKeyValue("Domain", strings.Join(domains, ", "))
			fmt.Println()
			ui.PrintInfo("Files created:")
			for _, f := range created {
				fmt.Printf("  • %s\n", f)
			}
		} else {
			ui.PrintSuccess("Lightspeed project up to date")
			fmt.Println()
			ui.PrintKeyValue("Name", siteName)
			ui.PrintKeyValue("Domain", strings.Join(domains, ", "))
		}
		fmt.Println()
		ui.PrintInfo("Run 'lightspeed start' to start the development server")
		fmt.Println()
	},
}

func init() {
	initCmd.Flags().StringVarP(&initName, "name", "n", "", "Site name (default: directory name)")
	initCmd.Flags().StringSliceVarP(&initDomains, "domain", "d", nil, "Domain(s) for the site (default: name.com)")

	rootCmd.AddCommand(initCmd)
}
