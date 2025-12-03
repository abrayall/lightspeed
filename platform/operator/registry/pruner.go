package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Pruner handles automatic cleanup of old container images
type Pruner struct {
	apiToken     string
	registryName string
	client       *http.Client
	keepLatest   bool
	keepVersions int // Number of semver versions to keep
}

// SemVer represents a parsed semantic version
type SemVer struct {
	Major int
	Minor int
	Patch int
	Raw   string // Original tag string
}

// TagInfo represents a tag with its metadata
type TagInfo struct {
	Tag       string
	UpdatedAt time.Time
}

// NewPruner creates a new image pruner
func NewPruner(apiToken, registryName string) *Pruner {
	return &Pruner{
		apiToken:     apiToken,
		registryName: registryName,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		keepLatest:   true,
		keepVersions: 3,
	}
}

// Start begins the daily pruning schedule
func (p *Pruner) Start() {
	// Run immediately on start
	// Log startup message first
	log.Printf("[PRUNER] Started - will prune daily, keeping latest + %d most recent versions", p.keepVersions)

	// Run first prune after 30 seconds
	go func() {
		time.Sleep(30 * time.Second) // Wait for startup
		p.Prune()
	}()

	// Then run daily
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			p.Prune()
		}
	}()
}

// Prune removes old image tags from all repositories
func (p *Pruner) Prune() {
	log.Printf("[PRUNER] Starting image cleanup...")

	repos, err := p.listRepositories()
	if err != nil {
		log.Printf("[PRUNER] Failed to list repositories: %v", err)
		return
	}

	totalDeleted := 0
	for _, repo := range repos {
		deleted, err := p.pruneRepository(repo)
		if err != nil {
			log.Printf("[PRUNER] Failed to prune %s: %v", repo, err)
			continue
		}
		totalDeleted += deleted
	}

	if totalDeleted > 0 {
		log.Printf("[PRUNER] Cleanup complete - deleted %d old tags", totalDeleted)
		// Trigger garbage collection
		if err := p.startGarbageCollection(); err != nil {
			log.Printf("[PRUNER] Failed to start garbage collection: %v", err)
		}
	} else {
		log.Printf("[PRUNER] Cleanup complete - no tags to delete")
	}
}

// listRepositories gets all repositories in the registry
func (p *Pruner) listRepositories() ([]string, error) {
	url := fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositoriesV2", p.registryName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Repositories []struct {
			Name string `json:"name"`
		} `json:"repositories"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	repos := make([]string, len(result.Repositories))
	for i, r := range result.Repositories {
		repos[i] = r.Name
	}

	return repos, nil
}

// pruneRepository removes old tags from a single repository
func (p *Pruner) pruneRepository(repoName string) (int, error) {
	tags, err := p.listTags(repoName)
	if err != nil {
		return 0, err
	}

	// If no tags, delete the entire repository
	if len(tags) == 0 {
		log.Printf("[PRUNER] %s: no tags, deleting repository", repoName)
		if err := p.deleteRepository(repoName); err != nil {
			return 0, err
		}
		return 1, nil
	}

	// Separate tags into categories
	var latestTag *TagInfo
	var versionTags []SemVer
	var otherTags []TagInfo

	semverRegex := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)

	for _, tagInfo := range tags {
		if tagInfo.Tag == "latest" {
			t := tagInfo // Copy to avoid pointer issues
			latestTag = &t
			continue
		}

		matches := semverRegex.FindStringSubmatch(tagInfo.Tag)
		if matches != nil {
			major, _ := strconv.Atoi(matches[1])
			minor, _ := strconv.Atoi(matches[2])
			patch, _ := strconv.Atoi(matches[3])
			versionTags = append(versionTags, SemVer{
				Major: major,
				Minor: minor,
				Patch: patch,
				Raw:   tagInfo.Tag,
			})
		} else {
			otherTags = append(otherTags, tagInfo)
		}
	}

	// Sort semver versions descending (highest first)
	sort.Slice(versionTags, func(i, j int) bool {
		if versionTags[i].Major != versionTags[j].Major {
			return versionTags[i].Major > versionTags[j].Major
		}
		if versionTags[i].Minor != versionTags[j].Minor {
			return versionTags[i].Minor > versionTags[j].Minor
		}
		return versionTags[i].Patch > versionTags[j].Patch
	})

	// Sort other tags by update date descending (most recent first)
	sort.Slice(otherTags, func(i, j int) bool {
		return otherTags[i].UpdatedAt.After(otherTags[j].UpdatedAt)
	})

	// Determine which tags to keep
	keepTags := make(map[string]bool)

	// Keep latest
	if latestTag != nil && p.keepLatest {
		keepTags[latestTag.Tag] = true
	}

	// Keep top N semver versions
	for i := 0; i < len(versionTags) && i < p.keepVersions; i++ {
		keepTags[versionTags[i].Raw] = true
	}

	// Keep top N non-semver tags by date (if no semver versions exist)
	if len(versionTags) == 0 {
		for i := 0; i < len(otherTags) && i < p.keepVersions; i++ {
			keepTags[otherTags[i].Tag] = true
		}
	}

	// Determine which to delete
	var tagsToDelete []string

	// Old semver versions beyond the keep limit
	for i := p.keepVersions; i < len(versionTags); i++ {
		tagsToDelete = append(tagsToDelete, versionTags[i].Raw)
	}

	// If no semver versions, delete old non-semver tags beyond the keep limit
	if len(versionTags) == 0 {
		for i := p.keepVersions; i < len(otherTags); i++ {
			tagsToDelete = append(tagsToDelete, otherTags[i].Tag)
		}
	}

	if len(tagsToDelete) == 0 {
		return 0, nil
	}

	log.Printf("[PRUNER] %s: keeping %v, deleting %v", repoName, keysFromMap(keepTags), tagsToDelete)

	// Delete old tags
	deleted := 0
	for _, tag := range tagsToDelete {
		if err := p.deleteTag(repoName, tag); err != nil {
			log.Printf("[PRUNER] Failed to delete %s:%s: %v", repoName, tag, err)
			continue
		}
		deleted++
	}

	return deleted, nil
}

// deleteRepository deletes an entire repository (when it has no tags)
func (p *Pruner) deleteRepository(repoName string) error {
	encodedRepo := strings.ReplaceAll(repoName, "/", "%2F")
	url := fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositories/%s", p.registryName, encodedRepo)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	log.Printf("[PRUNER] Deleted repository %s", repoName)
	return nil
}

// listTags gets all tags for a repository with their metadata
func (p *Pruner) listTags(repoName string) ([]TagInfo, error) {
	// URL encode the repo name (it may contain slashes)
	encodedRepo := strings.ReplaceAll(repoName, "/", "%2F")
	url := fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositories/%s/tags", p.registryName, encodedRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Tags []struct {
			Tag       string    `json:"tag"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"tags"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	tags := make([]TagInfo, len(result.Tags))
	for i, t := range result.Tags {
		tags[i] = TagInfo{
			Tag:       t.Tag,
			UpdatedAt: t.UpdatedAt,
		}
	}

	return tags, nil
}

// deleteTag deletes a specific tag from a repository
func (p *Pruner) deleteTag(repoName, tag string) error {
	encodedRepo := strings.ReplaceAll(repoName, "/", "%2F")
	url := fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositories/%s/tags/%s", p.registryName, encodedRepo, tag)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	log.Printf("[PRUNER] Deleted %s:%s", repoName, tag)
	return nil
}

// startGarbageCollection triggers DO's garbage collection to reclaim space
func (p *Pruner) startGarbageCollection() error {
	url := fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/garbage-collection", p.registryName)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 201 Created or 409 Conflict (already running) are both OK
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	log.Printf("[PRUNER] Garbage collection started")
	return nil
}

func keysFromMap(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
