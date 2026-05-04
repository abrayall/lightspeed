package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"lightspeed/core/lib/properties"
	"lightspeed/core/lib/ui"
)

const (
	composerBaseDir    = ".lightspeed/composer"
	composerLatestURL  = "https://getcomposer.org/download/latest-stable/composer.phar"
	composerVersionURL = "https://getcomposer.org/download/%s/composer.phar"
)

// isComposerSpec checks if a library spec is a composer specification
func isComposerSpec(spec string) bool {
	return spec == "composer" || strings.HasPrefix(spec, "composer:")
}

// isComposerToolSpec checks if a spec refers to the composer tool itself (no "/" in the value)
func isComposerToolSpec(spec string) bool {
	if !isComposerSpec(spec) {
		return false
	}
	if spec == "composer" {
		return true
	}
	value := strings.TrimPrefix(spec, "composer:")
	return !strings.Contains(value, "/")
}

// isComposerPackageSpec checks if a spec refers to a composer package (has "/" in the value)
func isComposerPackageSpec(spec string) bool {
	if !isComposerSpec(spec) {
		return false
	}
	if spec == "composer" {
		return false
	}
	value := strings.TrimPrefix(spec, "composer:")
	return strings.Contains(value, "/")
}

// parseComposerToolVersion extracts the version from a composer tool spec
func parseComposerToolVersion(spec string) string {
	if spec == "composer" {
		return "latest"
	}
	version := strings.TrimPrefix(spec, "composer:")
	if version == "" {
		return "latest"
	}
	return version
}

// parseComposerPackage parses a composer package spec into name and version
func parseComposerPackage(spec string) (string, string, error) {
	if !isComposerPackageSpec(spec) {
		return "", "", fmt.Errorf("not a composer package spec: %s", spec)
	}

	value := strings.TrimPrefix(spec, "composer:")

	// Find the package name (vendor/package) and optional version
	// Format: vendor/package or vendor/package:version
	parts := strings.SplitN(value, ":", 2)
	name := parts[0]

	if !strings.Contains(name, "/") {
		return "", "", fmt.Errorf("invalid package name (missing vendor/): %s", name)
	}

	version := "*"
	if len(parts) > 1 && parts[1] != "" {
		version = parts[1]
	}

	return name, version, nil
}

// loadComposerConfig reads site.properties and separates composer tool version from package specs
func loadComposerConfig(dir string) (string, map[string]string, error) {
	propsPath := filepath.Join(dir, "site.properties")
	if _, err := os.Stat(propsPath); os.IsNotExist(err) {
		return "", nil, nil
	}

	props, err := properties.ParseProperties(propsPath)
	if err != nil {
		return "", nil, err
	}

	librariesStr := props.Get("libraries")
	if librariesStr == "" {
		return "", nil, nil
	}

	toolVersion := ""
	packages := make(map[string]string)

	specs := strings.Split(librariesStr, ",")
	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		if spec == "" || !isComposerSpec(spec) {
			continue
		}

		if isComposerToolSpec(spec) {
			toolVersion = parseComposerToolVersion(spec)
		} else if isComposerPackageSpec(spec) {
			name, version, err := parseComposerPackage(spec)
			if err != nil {
				return "", nil, fmt.Errorf("invalid composer package spec '%s': %w", spec, err)
			}
			packages[name] = version
			// If we have packages but no explicit tool spec, default to latest
			if toolVersion == "" {
				toolVersion = "latest"
			}
		}
	}

	if toolVersion == "" && len(packages) == 0 {
		return "", nil, nil
	}

	return toolVersion, packages, nil
}

// ensureComposerPhar downloads composer.phar if not already present
func ensureComposerPhar(version string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	pharDir := filepath.Join(homeDir, composerBaseDir, "v"+version)
	pharPath := filepath.Join(pharDir, "composer.phar")

	// Check if already downloaded
	if _, err := os.Stat(pharPath); err == nil {
		return pharPath, nil
	}

	// Create directory
	if err := os.MkdirAll(pharDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create composer directory: %w", err)
	}

	// Determine download URL
	var downloadURL string
	if version == "latest" {
		downloadURL = composerLatestURL
	} else {
		downloadURL = fmt.Sprintf(composerVersionURL, version)
	}

	ui.PrintInfo("Downloading Composer %s...", version)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download composer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to download composer: HTTP %d", resp.StatusCode)
	}

	// Write to temp file first, then move
	tmpFile, err := os.CreateTemp(pharDir, "composer-*.phar.tmp")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to save composer: %w", err)
	}

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to set permissions: %w", err)
	}

	// Move to final location
	if err := os.Rename(tmpPath, pharPath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to install composer: %w", err)
	}

	return pharPath, nil
}

// composerJSON represents the structure of a composer.json file
type composerJSON struct {
	Require map[string]string `json:"require,omitempty"`
}

// syncComposerJSON generates composer.json in the project directory from package entries
func syncComposerJSON(dir string, packages map[string]string) (bool, error) {
	if len(packages) == 0 {
		return false, nil
	}

	composerPath := filepath.Join(dir, "composer.json")

	newContent := composerJSON{
		Require: packages,
	}

	newJSON, err := json.MarshalIndent(newContent, "", "    ")
	if err != nil {
		return false, fmt.Errorf("failed to marshal composer.json: %w", err)
	}
	newJSON = append(newJSON, '\n')

	// Check if content changed
	existing, err := os.ReadFile(composerPath)
	if err == nil {
		existingHash := sha256.Sum256(existing)
		newHash := sha256.Sum256(newJSON)
		if hex.EncodeToString(existingHash[:]) == hex.EncodeToString(newHash[:]) {
			return false, nil
		}
	}

	if err := os.WriteFile(composerPath, newJSON, 0644); err != nil {
		return false, fmt.Errorf("failed to write composer.json: %w", err)
	}

	return true, nil
}

// runComposerInstall runs composer install in the project directory
func runComposerInstall(dir string, pharPath string) error {
	var cmd *exec.Cmd

	if isCommandAvailable("php") {
		cmd = exec.Command("php", pharPath, "install", "--no-interaction")
	} else if isCommandAvailable("composer") {
		cmd = exec.Command("composer", "install", "--no-interaction")
	} else {
		return fmt.Errorf("php or composer command required but not found")
	}

	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("composer install failed: %w", err)
	}

	return nil
}

// ensureComposerDependencies orchestrates the full composer setup
func ensureComposerDependencies(dir string) error {
	toolVersion, packages, err := loadComposerConfig(dir)
	if err != nil {
		return err
	}

	// No composer configuration found
	if toolVersion == "" && len(packages) == 0 {
		return nil
	}

	// Ensure composer.phar is available
	pharPath, err := ensureComposerPhar(toolVersion)
	if err != nil {
		return err
	}

	// No packages to install
	if len(packages) == 0 {
		return nil
	}

	// Sync composer.json
	changed, err := syncComposerJSON(dir, packages)
	if err != nil {
		return err
	}

	// Run composer install if vendor/ is missing or composer.json changed
	vendorDir := filepath.Join(dir, "vendor")
	needsInstall := changed
	if _, err := os.Stat(vendorDir); os.IsNotExist(err) {
		needsInstall = true
	}

	if needsInstall {
		ui.PrintInfo("Installing Composer dependencies...")
		if err := runComposerInstall(dir, pharPath); err != nil {
			return err
		}
		ui.PrintSuccess("Composer dependencies installed")
		fmt.Println()
	}

	// Update .gitignore
	if err := addComposerToGitignore(dir); err != nil {
		return err
	}

	return nil
}

// addComposerToGitignore appends vendor/ and composer.json to .gitignore if not already present
func addComposerToGitignore(dir string) error {
	gitignorePath := filepath.Join(dir, ".gitignore")

	var content string
	existing, err := os.ReadFile(gitignorePath)
	if err == nil {
		content = string(existing)
	}

	entries := []string{"vendor/", "composer.json"}
	var toAdd []string

	for _, entry := range entries {
		if !containsGitignoreEntry(content, entry) {
			toAdd = append(toAdd, entry)
		}
	}

	if len(toAdd) == 0 {
		return nil
	}

	// Ensure content ends with newline before appending
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	content += strings.Join(toAdd, "\n") + "\n"

	return os.WriteFile(gitignorePath, []byte(content), 0644)
}

// containsGitignoreEntry checks if a .gitignore entry already exists
func containsGitignoreEntry(content string, entry string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == entry {
			return true
		}
	}
	return false
}

// hasComposerPackages checks if the project has any composer package specs in site.properties
func hasComposerPackages(dir string) bool {
	_, packages, err := loadComposerConfig(dir)
	if err != nil {
		return false
	}
	return len(packages) > 0
}
