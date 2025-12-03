package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"lightspeed/core/lib/properties"
)

const (
	libraryBaseURL = "https://github.com/abrayall/lightspeed/releases/download"
	libraryBaseDir = ".lightspeed/library"
)

// getBaseVersion strips dev suffixes from version (e.g., "0.5.3-12031417" -> "0.5.3")
func getBaseVersion() string {
	v := Version
	// Match semantic version pattern and strip anything after
	re := regexp.MustCompile(`^(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(v)
	if len(matches) > 1 {
		return matches[1]
	}
	return v
}

// getLibraryDir returns the path to the library directory for the current version
func getLibraryDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, libraryBaseDir, "v"+getBaseVersion())
}

// isLibraryInstalled checks if the library is installed for the current version
func isLibraryInstalled() bool {
	libDir := getLibraryDir()
	if libDir == "" {
		return false
	}
	// Check if directory exists and has files
	info, err := os.Stat(libDir)
	if err != nil || !info.IsDir() {
		return false
	}
	// Check for at least one file
	files, err := os.ReadDir(libDir)
	if err != nil || len(files) == 0 {
		return false
	}
	return true
}

// ensureLibrary checks if library is installed, downloads if not
func ensureLibrary() error {
	if isLibraryInstalled() {
		return nil
	}

	baseVersion := getBaseVersion()
	if baseVersion == "dev" {
		// Don't try to download for dev builds
		return nil
	}

	return downloadLibrary(baseVersion)
}

// downloadLibrary downloads and extracts the library for the given version
func downloadLibrary(version string) error {
	libDir := getLibraryDir()
	if libDir == "" {
		return fmt.Errorf("could not determine library directory")
	}

	// Create directory
	if err := os.MkdirAll(libDir, 0755); err != nil {
		return fmt.Errorf("failed to create library directory: %w", err)
	}

	// Download URL
	zipURL := fmt.Sprintf("%s/v%s/lightspeed-library-%s.zip", libraryBaseURL, version, version)

	// Download to temp file
	resp, err := http.Get(zipURL)
	if err != nil {
		return fmt.Errorf("failed to download library: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download library: HTTP %d", resp.StatusCode)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "lightspeed-library-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write to temp file
	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		return fmt.Errorf("failed to save library: %w", err)
	}

	// Extract zip
	if err := extractZip(tmpPath, libDir); err != nil {
		return fmt.Errorf("failed to extract library: %w", err)
	}

	return nil
}

// extractZip extracts a zip file to the destination directory
func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		// Security check
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// getLibraryDirForVersion returns the library directory for a specific version
func getLibraryDirForVersion(version string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	version = strings.TrimPrefix(version, "v")
	return filepath.Join(homeDir, libraryBaseDir, "v"+version)
}

// ensureLibraryVersion ensures a specific version of the library is installed
func ensureLibraryVersion(version string) error {
	version = strings.TrimPrefix(version, "v")
	libDir := getLibraryDirForVersion(version)

	// Check if already installed
	if info, err := os.Stat(libDir); err == nil && info.IsDir() {
		files, err := os.ReadDir(libDir)
		if err == nil && len(files) > 0 {
			return nil
		}
	}

	return downloadLibrary(version)
}

// resolveLibraryPath resolves a library specification to an absolute path
func resolveLibraryPath(spec string) (string, error) {
	if spec == "lightspeed" {
		if err := ensureLibrary(); err != nil {
			return "", err
		}
		return getLibraryDir(), nil
	}

	if strings.HasPrefix(spec, "lightspeed:") {
		version := strings.TrimPrefix(spec, "lightspeed:")
		version = strings.TrimPrefix(version, "v")
		if err := ensureLibraryVersion(version); err != nil {
			return "", err
		}
		return getLibraryDirForVersion(version), nil
	}

	return spec, nil
}

// loadLibraries loads and resolves library paths from site.properties
func loadLibraries(dir string) ([]string, error) {
	propsPath := filepath.Join(dir, "site.properties")
	if _, err := os.Stat(propsPath); os.IsNotExist(err) {
		return nil, nil
	}

	props, err := properties.ParseProperties(propsPath)
	if err != nil {
		return nil, err
	}

	librariesStr := props.Get("libraries")
	if librariesStr == "" {
		return nil, nil
	}

	specs := strings.Split(librariesStr, ",")
	var resolved []string

	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}

		path, err := resolveLibraryPath(spec)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve library '%s': %w", spec, err)
		}
		resolved = append(resolved, path)
	}

	return resolved, nil
}

// updateIdeaConfig updates .idea/php.xml and run configurations with resolved library paths
func updateIdeaConfig(dir string) error {
	ideaDir := filepath.Join(dir, ".idea")
	propsPath := filepath.Join(dir, "site.properties")

	// Only proceed if both .idea and site.properties exist
	if _, err := os.Stat(ideaDir); os.IsNotExist(err) {
		return nil
	}
	if _, err := os.Stat(propsPath); os.IsNotExist(err) {
		return nil
	}

	// Load and resolve libraries
	libraries, err := loadLibraries(dir)
	if err != nil {
		return err
	}

	if len(libraries) == 0 {
		return nil
	}

	// Generate php.xml content
	var paths string
	for _, lib := range libraries {
		paths += fmt.Sprintf("      <path value=\"%s\" />\n", lib)
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<project version="4">
  <component name="PhpIncludePathManager">
    <include_path>
%s    </include_path>
  </component>
</project>
`, paths)

	phpXmlPath := filepath.Join(ideaDir, "php.xml")

	// Only write if content is different
	existing, err := os.ReadFile(phpXmlPath)
	if err == nil && string(existing) == content {
		// php.xml unchanged, but still update run config
		updateRunConfig(dir, libraries)
		return nil
	}

	if err := os.WriteFile(phpXmlPath, []byte(content), 0644); err != nil {
		return err
	}

	// Update run configuration
	return updateRunConfig(dir, libraries)
}

// updateRunConfig creates/updates .idea/runConfigurations/{sitename}.xml
func updateRunConfig(dir string, libraries []string) error {
	runConfigDir := filepath.Join(dir, ".idea", "runConfigurations")

	// Create runConfigurations directory if it doesn't exist
	if err := os.MkdirAll(runConfigDir, 0755); err != nil {
		return err
	}

	// Get site name from site.properties
	siteName := filepath.Base(dir) // default to directory name
	propsPath := filepath.Join(dir, "site.properties")
	if props, err := properties.ParseProperties(propsPath); err == nil {
		if name := props.Get("name"); name != "" {
			siteName = name
		}
	}
	siteName = sanitizeContainerName(siteName)

	// Build include_path from libraries using $USER_HOME$ variable
	includePath := "."
	homeDir, _ := os.UserHomeDir()
	for _, lib := range libraries {
		// Convert absolute home path to $USER_HOME$ variable
		if strings.HasPrefix(lib, homeDir) {
			lib = "$USER_HOME$" + strings.TrimPrefix(lib, homeDir)
		}
		includePath += ":" + lib
	}

	content := fmt.Sprintf(`<component name="ProjectRunConfigurationManager">
  <configuration default="false" name="%s" type="PhpBuiltInWebServerConfigurationType" factoryName="PHP Built-in Web Server" document_root="$PROJECT_DIR$" port="8888">
    <CommandLine parameters="-d include_path=&quot;%s&quot;" />
    <method v="2" />
  </configuration>
</component>
`, siteName, includePath)

	runConfigPath := filepath.Join(runConfigDir, siteName+".xml")

	// Only write if content is different
	existing, err := os.ReadFile(runConfigPath)
	if err == nil && string(existing) == content {
		return nil
	}

	return os.WriteFile(runConfigPath, []byte(content), 0644)
}
