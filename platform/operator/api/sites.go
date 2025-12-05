package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const digitalOceanAPI = "https://api.digitalocean.com/v2"

// SitesHandler handles /sites endpoints
type SitesHandler struct {
	defaultToken    string
	defaultRegistry string
	cfClient        *CloudflareClient
	operatorURL     string
	operatorToken   string
}

// NewSitesHandler creates a new sites handler
func NewSitesHandler(defaultToken, defaultRegistry, cfToken, operatorURL, operatorToken string) *SitesHandler {
	return &SitesHandler{
		defaultToken:    defaultToken,
		defaultRegistry: defaultRegistry,
		cfClient:        NewCloudflareClient(cfToken),
		operatorURL:     operatorURL,
		operatorToken:   operatorToken,
	}
}

// Site represents a site/app configuration (public API)
type Site struct {
	Name    string   `json:"name"`
	Image   string   `json:"image,omitempty"`
	Tag     string   `json:"tag,omitempty"`
	Domains []string `json:"domains,omitempty"`
}

// Internal defaults (not exposed via API)
const (
	defaultRegion    = "nyc"
	defaultPort      = 80
	defaultInstances = 1
	defaultSize      = "apps-s-1vcpu-0.5gb"
)

// SiteResponse represents a site in responses
type SiteResponse struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Region    string   `json:"region,omitempty"`
	URLs      []string `json:"urls,omitempty"`
	Status    string   `json:"status,omitempty"`
	UpdatedAt string   `json:"updated_at,omitempty"`
}

// ServeHTTP routes requests to appropriate handlers
func (h *SitesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get token from header or use default
	token := r.Header.Get("Authorization")
	if token == "" && h.defaultToken != "" {
		token = "Bearer " + h.defaultToken
	}
	if token != "" && !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	path := strings.TrimPrefix(r.URL.Path, "/sites")
	path = strings.TrimPrefix(path, "/")

	log.Printf("[API] %s /sites/%s", r.Method, path)

	switch {
	case path == "" && r.Method == http.MethodGet:
		h.listSites(w, r, token)
	case path == "" && r.Method == http.MethodPost:
		h.createSite(w, r, token)
	case r.Method == http.MethodGet:
		h.getSite(w, r, token, path)
	case r.Method == http.MethodDelete:
		h.deleteSite(w, r, token, path)
	case strings.HasSuffix(path, "/deploy") && r.Method == http.MethodPost:
		name := strings.TrimSuffix(path, "/deploy")
		h.deploySite(w, r, token, name)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listSites returns all apps from DigitalOcean
func (h *SitesHandler) listSites(w http.ResponseWriter, r *http.Request, token string) {
	resp, err := h.doRequest("GET", "/apps", token, nil)
	if err != nil {
		h.writeError(w, "Failed to list sites", err, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.forwardError(w, resp)
		return
	}

	var result struct {
		Apps []struct {
			ID              string `json:"id"`
			OwnerUUID       string `json:"owner_uuid"`
			Spec            struct {
				Name   string `json:"name"`
				Region string `json:"region"`
			} `json:"spec"`
			DefaultIngress  string `json:"default_ingress"`
			LiveURL         string `json:"live_url"`
			ActiveDeployment struct {
				Phase string `json:"phase"`
			} `json:"active_deployment"`
			UpdatedAt string `json:"updated_at"`
		} `json:"apps"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		h.writeError(w, "Failed to parse response", err, http.StatusInternalServerError)
		return
	}

	// Transform to our format
	sites := make([]SiteResponse, 0, len(result.Apps))
	for _, app := range result.Apps {
		urls := []string{}
		if app.LiveURL != "" {
			urls = append(urls, app.LiveURL)
		}
		if app.DefaultIngress != "" {
			urls = append(urls, app.DefaultIngress)
		}

		sites = append(sites, SiteResponse{
			ID:        app.ID,
			Name:      app.Spec.Name,
			Region:    app.Spec.Region,
			URLs:      urls,
			Status:    app.ActiveDeployment.Phase,
			UpdatedAt: app.UpdatedAt,
		})
	}

	h.writeJSON(w, map[string]interface{}{"sites": sites})
}

// createSite creates a new app on DigitalOcean
func (h *SitesHandler) createSite(w http.ResponseWriter, r *http.Request, token string) {
	var site Site
	if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
		h.writeError(w, "Invalid request body", err, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if site.Name == "" {
		h.writeError(w, "name is required", nil, http.StatusBadRequest)
		return
	}

	// Set defaults for optional fields
	image := site.Image
	if image == "" {
		image = site.Name
	}
	tag := site.Tag
	if tag == "" {
		tag = "latest"
	}

	// Wait for the tag to be available in the registry
	log.Printf("[API] Verifying tag %s:%s exists in registry...", image, tag)
	if err := h.waitForTag(image, tag, token); err != nil {
		h.writeError(w, "Image tag not available", err, http.StatusNotFound)
		return
	}

	// Build domains list - start with default lightspeed.ee domain as PRIMARY
	domains := []map[string]string{
		{
			"domain": site.Name + ".lightspeed.ee",
			"type":   "PRIMARY",
		},
	}
	// Add any custom domains from the request as ALIAS domains
	for _, domain := range site.Domains {
		domains = append(domains, map[string]string{
			"domain": domain,
			"type":   "ALIAS",
		})
	}

	// Build app spec using internal defaults
	spec := map[string]interface{}{
		"name":   site.Name,
		"region": defaultRegion,
		"features": []string{
			"buildpack-stack=ubuntu-22",
		},
		"alerts": []map[string]string{
			{"rule": "DEPLOYMENT_FAILED"},
			{"rule": "DOMAIN_FAILED"},
		},
		"domains": domains,
		"ingress": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"component": map[string]string{
						"name": site.Name,
					},
					"match": map[string]interface{}{
						"path": map[string]string{
							"prefix": "/",
						},
					},
				},
			},
		},
		"services": []map[string]interface{}{
			{
				"name":      site.Name,
				"http_port": defaultPort,
				"image": map[string]interface{}{
					"registry_type": "DOCR",
					"registry":      h.defaultRegistry,
					"repository":    image,
					"tag":           tag,
					"deploy_on_push": map[string]bool{
						"enabled": true,
					},
				},
				"instance_count":     defaultInstances,
				"instance_size_slug": defaultSize,
				"envs": []map[string]interface{}{
					{
						"key":   "OPERATOR_URL",
						"value": h.operatorURL,
						"type":  "GENERAL",
					},
					{
						"key":   "OPERATOR_TOKEN",
						"value": h.operatorToken,
						"type":  "SECRET",
					},
				},
			},
		},
	}

	payload := map[string]interface{}{
		"spec": spec,
	}

	body, _ := json.Marshal(payload)

	resp, err := h.doRequest("POST", "/apps", token, body)
	if err != nil {
		h.writeError(w, "Failed to create site", err, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		h.forwardError(w, resp)
		return
	}

	var result struct {
		App struct {
			ID             string `json:"id"`
			DefaultIngress string `json:"default_ingress"`
			Spec           struct {
				Name   string `json:"name"`
				Region string `json:"region"`
			} `json:"spec"`
		} `json:"app"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		h.writeError(w, "Failed to parse response", err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	h.writeJSON(w, SiteResponse{
		ID:     result.App.ID,
		Name:   result.App.Spec.Name,
		Region: result.App.Spec.Region,
	})
}

// getSite gets a specific app by name
func (h *SitesHandler) getSite(w http.ResponseWriter, r *http.Request, token string, name string) {
	// First, find the app ID by name
	appID, err := h.findAppByName(token, name)
	if err != nil {
		h.writeError(w, "Failed to find site", err, http.StatusBadGateway)
		return
	}
	if appID == "" {
		http.Error(w, `{"error":"Site not found"}`, http.StatusNotFound)
		return
	}

	// Get the app details
	resp, err := h.doRequest("GET", "/apps/"+appID, token, nil)
	if err != nil {
		h.writeError(w, "Failed to get site", err, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.forwardError(w, resp)
		return
	}

	var result struct {
		App struct {
			ID              string `json:"id"`
			Spec            struct {
				Name   string `json:"name"`
				Region string `json:"region"`
			} `json:"spec"`
			LiveURL         string `json:"live_url"`
			DefaultIngress  string `json:"default_ingress"`
			ActiveDeployment struct {
				Phase string `json:"phase"`
			} `json:"active_deployment"`
			UpdatedAt string `json:"updated_at"`
		} `json:"app"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		h.writeError(w, "Failed to parse response", err, http.StatusInternalServerError)
		return
	}

	urls := []string{}
	if result.App.LiveURL != "" {
		urls = append(urls, result.App.LiveURL)
	}
	if result.App.DefaultIngress != "" {
		urls = append(urls, result.App.DefaultIngress)
	}

	h.writeJSON(w, SiteResponse{
		ID:        result.App.ID,
		Name:      result.App.Spec.Name,
		Region:    result.App.Spec.Region,
		URLs:      urls,
		Status:    result.App.ActiveDeployment.Phase,
		UpdatedAt: result.App.UpdatedAt,
	})
}

// deleteSite deletes an app
func (h *SitesHandler) deleteSite(w http.ResponseWriter, r *http.Request, token string, name string) {
	appID, err := h.findAppByName(token, name)
	if err != nil {
		h.writeError(w, "Failed to find site", err, http.StatusBadGateway)
		return
	}
	if appID == "" {
		http.Error(w, `{"error":"Site not found"}`, http.StatusNotFound)
		return
	}

	resp, err := h.doRequest("DELETE", "/apps/"+appID, token, nil)
	if err != nil {
		h.writeError(w, "Failed to delete site", err, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		h.forwardError(w, resp)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// deploySite triggers a deployment
func (h *SitesHandler) deploySite(w http.ResponseWriter, r *http.Request, token string, name string) {
	appID, err := h.findAppByName(token, name)
	if err != nil {
		h.writeError(w, "Failed to find site", err, http.StatusBadGateway)
		return
	}
	if appID == "" {
		http.Error(w, `{"error":"Site not found"}`, http.StatusNotFound)
		return
	}

	payload := map[string]interface{}{
		"force_build": true,
	}
	body, _ := json.Marshal(payload)

	resp, err := h.doRequest("POST", "/apps/"+appID+"/deployments", token, body)
	if err != nil {
		h.writeError(w, "Failed to create deployment", err, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		h.forwardError(w, resp)
		return
	}

	var result struct {
		Deployment struct {
			ID    string `json:"id"`
			Phase string `json:"phase"`
		} `json:"deployment"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		h.writeError(w, "Failed to parse response", err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	h.writeJSON(w, map[string]interface{}{
		"deployment_id": result.Deployment.ID,
		"status":        result.Deployment.Phase,
	})
}

// findAppByName finds an app ID by name
func (h *SitesHandler) findAppByName(token, name string) (string, error) {
	resp, err := h.doRequest("GET", "/apps", token, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Apps []struct {
			ID   string `json:"id"`
			Spec struct {
				Name string `json:"name"`
			} `json:"spec"`
		} `json:"apps"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	for _, app := range result.Apps {
		if app.Spec.Name == name {
			return app.ID, nil
		}
	}

	return "", nil
}

// doRequest makes a request to DigitalOcean API
func (h *SitesHandler) doRequest(method, path, token string, body []byte) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(method, digitalOceanAPI+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	return client.Do(req)
}

// writeJSON writes a JSON response
func (h *SitesHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func (h *SitesHandler) writeError(w http.ResponseWriter, message string, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	errMsg := message
	if err != nil {
		errMsg = fmt.Sprintf("%s: %v", message, err)
		log.Printf("[API] Error: %s", errMsg)
	}
	json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
}

// forwardError forwards an error response from DigitalOcean
func (h *SitesHandler) forwardError(w http.ResponseWriter, resp *http.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// tagExists checks if an image tag exists in the registry
func (h *SitesHandler) tagExists(repository, tag, token string) (bool, error) {
	url := fmt.Sprintf("/registry/%s/repositories/%s/tags", h.defaultRegistry, repository)

	resp, err := h.doRequest("GET", url, token, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	// Parse response to check if our tag is in the list
	var result struct {
		Tags []struct {
			Tag string `json:"tag"`
		} `json:"tags"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	// Check if our tag is in the list
	for _, t := range result.Tags {
		if t.Tag == tag {
			return true, nil
		}
	}

	return false, nil
}

// waitForTag waits for a tag to appear in the registry (with retries)
func (h *SitesHandler) waitForTag(repository, tag, token string) error {
	maxRetries := 5
	retryDelay := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		exists, err := h.tagExists(repository, tag, token)
		if err != nil {
			log.Printf("[API] Error checking tag existence (attempt %d/%d): %v", attempt, maxRetries, err)
		} else if exists {
			log.Printf("[API] Tag %s:%s verified in registry", repository, tag)
			return nil
		}

		if attempt < maxRetries {
			log.Printf("[API] Tag %s:%s not yet indexed, retrying in %v (attempt %d/%d)",
				repository, tag, retryDelay, attempt, maxRetries)
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("tag %s:%s not found in registry after %d attempts", repository, tag, maxRetries)
}
