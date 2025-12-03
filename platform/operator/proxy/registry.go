package proxy

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// RegistryProxy proxies requests to an upstream Docker registry
type RegistryProxy struct {
	upstream       *url.URL
	registryClient *http.Client // For proxying registry requests
	apiClient      *http.Client // For calling DO API
	publicHost     string       // The public hostname of this proxy (for rewriting auth challenges)
	authToken      string       // DO API token for authentication
	registryName   string       // Registry namespace to prepend to paths (e.g., "lightspeed-images")

	// Cached docker credentials (base64 username:password)
	dockerCreds string
	credsExpiry time.Time
	credsMu     sync.RWMutex
}

// SetAuthToken sets the DO API token to use for upstream authentication
func (p *RegistryProxy) SetAuthToken(token string) {
	p.authToken = token
}

// SetRegistryName sets the registry namespace to prepend to paths
func (p *RegistryProxy) SetRegistryName(name string) {
	p.registryName = name
}

// getDockerCreds gets cached docker credentials, refreshing if needed
func (p *RegistryProxy) getDockerCreds() (string, error) {
	p.credsMu.RLock()
	if p.dockerCreds != "" && time.Now().Before(p.credsExpiry) {
		creds := p.dockerCreds
		p.credsMu.RUnlock()
		return creds, nil
	}
	p.credsMu.RUnlock()

	p.credsMu.Lock()
	defer p.credsMu.Unlock()

	if p.dockerCreds != "" && time.Now().Before(p.credsExpiry) {
		return p.dockerCreds, nil
	}

	creds, err := p.fetchDockerCreds()
	if err != nil {
		return "", err
	}

	p.dockerCreds = creds
	p.credsExpiry = time.Now().Add(30 * time.Minute)
	log.Printf("[PROXY] Refreshed docker credentials")

	return creds, nil
}

// getTokenForRepo gets a Bearer token for a specific repository
func (p *RegistryProxy) getTokenForRepo(repoPath string) (string, error) {
	log.Printf("[PROXY] [DEBUG] Getting token for repo: %s", repoPath)

	creds, err := p.getDockerCreds()
	if err != nil {
		log.Printf("[PROXY] [DEBUG] Failed to get docker creds: %v", err)
		return "", err
	}
	log.Printf("[PROXY] [DEBUG] Got docker creds (length: %d)", len(creds))

	// Request token with exact scope for this repo
	scope := fmt.Sprintf("repository:%s:push,pull", repoPath)
	authURL := fmt.Sprintf("https://api.digitalocean.com/v2/registry/auth?service=registry.digitalocean.com&scope=%s", url.QueryEscape(scope))

	log.Printf("[PROXY] [DEBUG] Token request URL: %s", authURL)
	log.Printf("[PROXY] [DEBUG] Scope: %s", scope)

	req, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		return "", err
	}

	authHeader := "Basic " + creds
	req.Header.Set("Authorization", authHeader)
	log.Printf("[PROXY] [DEBUG] Authorization header: %s", authHeader[:50]+"...") // Log first 50 chars
	log.Printf("[PROXY] [DEBUG] Authorization header length: %d", len(authHeader))
	log.Printf("[PROXY] [DEBUG] Full request URL: %s", req.URL.String())
	log.Printf("[PROXY] [DEBUG] Request headers: %v", req.Header)

	resp, err := p.apiClient.Do(req)
	if err != nil {
		log.Printf("[PROXY] [DEBUG] Request failed: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("[PROXY] [DEBUG] Token response status: %d", resp.StatusCode)
	log.Printf("[PROXY] [DEBUG] Token response body: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		log.Printf("[PROXY] Token fetch failed for %s: %s - %s", repoPath, resp.Status, string(body))
		return "", fmt.Errorf("token fetch failed: %s", resp.Status)
	}

	var result struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[PROXY] [DEBUG] Failed to unmarshal token response: %v", err)
		return "", err
	}

	token := result.Token
	if token == "" {
		token = result.AccessToken
	}

	if token == "" {
		log.Printf("[PROXY] [DEBUG] No token found in response")
		return "", fmt.Errorf("no token in response")
	}

	log.Printf("[PROXY] [DEBUG] Successfully got token (length: %d)", len(token))
	return token, nil
}

// fetchDockerCreds gets docker credentials from DO API
func (p *RegistryProxy) fetchDockerCreds() (string, error) {
	credsURL := "https://api.digitalocean.com/v2/registry/docker-credentials?read_write=true"
	log.Printf("[PROXY] [DEBUG] Fetching docker credentials from DO API")
	log.Printf("[PROXY] [DEBUG] API token length: %d", len(p.authToken))

	req, err := http.NewRequest("GET", credsURL, nil)
	if err != nil {
		log.Printf("[PROXY] [DEBUG] Failed to create request: %v", err)
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+p.authToken)
	log.Printf("[PROXY] [DEBUG] Set Bearer token in Authorization header")

	resp, err := p.apiClient.Do(req)
	if err != nil {
		log.Printf("[PROXY] [DEBUG] Request failed: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("[PROXY] [DEBUG] Credentials response status: %d", resp.StatusCode)
	log.Printf("[PROXY] [DEBUG] Credentials response body: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		log.Printf("[PROXY] Credentials fetch failed: %s - %s", resp.Status, string(body))
		return "", fmt.Errorf("credentials fetch failed: %s", resp.Status)
	}

	var credsResult struct {
		Auths map[string]struct {
			Auth string `json:"auth"`
		} `json:"auths"`
	}

	if err := json.Unmarshal(body, &credsResult); err != nil {
		log.Printf("[PROXY] [DEBUG] Failed to unmarshal credentials: %v", err)
		return "", fmt.Errorf("failed to decode credentials: %v", err)
	}

	log.Printf("[PROXY] [DEBUG] Found %d auth entries in response", len(credsResult.Auths))
	for host := range credsResult.Auths {
		log.Printf("[PROXY] [DEBUG] Auth entry for host: %s", host)
	}

	registryAuth, ok := credsResult.Auths["registry.digitalocean.com"]
	if !ok || registryAuth.Auth == "" {
		log.Printf("[PROXY] [DEBUG] No registry.digitalocean.com auth found")
		return "", fmt.Errorf("no auth credentials in response")
	}

	log.Printf("[PROXY] [DEBUG] Successfully fetched docker credentials (length: %d)", len(registryAuth.Auth))
	return registryAuth.Auth, nil
}

// extractRepoFromPath extracts the repository path from a registry API path
// Handles both /v2/myimage/... and /v2/lightspeed-images/myimage/...
func (p *RegistryProxy) extractRepoFromPath(path string) string {
	if !strings.HasPrefix(path, "/v2/") {
		return ""
	}
	rest := strings.TrimPrefix(path, "/v2/")

	// If path already starts with registry name, use it as-is
	if strings.HasPrefix(rest, p.registryName+"/") {
		// Extract registryName/imageName from registryName/imageName/blobs/...
		afterRegistry := strings.TrimPrefix(rest, p.registryName+"/")
		parts := strings.SplitN(afterRegistry, "/", 2)
		if len(parts) == 0 || parts[0] == "" {
			return ""
		}
		return p.registryName + "/" + parts[0]
	}

	// Otherwise add registry name prefix
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	return p.registryName + "/" + parts[0]
}

// NewRegistryProxy creates a new registry proxy
func NewRegistryProxy(upstreamURL, publicHost string) (*RegistryProxy, error) {
	// Ensure https
	if !strings.HasPrefix(upstreamURL, "http://") && !strings.HasPrefix(upstreamURL, "https://") {
		upstreamURL = "https://" + upstreamURL
	}

	upstream, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, err
	}

	// Create HTTP client for registry operations
	registryClient := &http.Client{
		// Follow redirects automatically (for CDN URLs from DigitalOcean)
		// This ensures Docker clients don't need to handle redirect auth
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't copy Authorization header to CDN URLs
			// CDN uses pre-signed URLs in query params, auth header will cause 400
			return nil
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
			// Important: Don't limit idle connections for streaming
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
			// Disable compression to preserve Content-Length for uploads
			DisableCompression: true,
		},
		// No timeout - uploads can be large
		Timeout: 0,
	}

	// Create standard HTTP client for API calls
	apiClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
		Timeout: 30 * time.Second,
	}

	return &RegistryProxy{
		upstream:       upstream,
		registryClient: registryClient,
		apiClient:      apiClient,
		publicHost:     publicHost,
	}, nil
}

// ServeHTTP handles proxied requests
func (p *RegistryProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Handle /v2/ base endpoint - accept any auth and return OK
	// This allows docker login to succeed with any credentials
	if r.URL.Path == "/v2/" || r.URL.Path == "/v2" {
		w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
		log.Printf("[PROXY] %s %s -> 200 (auth accepted)", r.Method, r.URL.Path)
		return
	}

	// Create upstream request
	upstreamURL := *p.upstream

	// Rewrite path to include registry namespace
	// /v2/myimage/... -> /v2/lightspeed-images/myimage/...
	path := r.URL.Path
	if p.registryName != "" && strings.HasPrefix(path, "/v2/") {
		rest := strings.TrimPrefix(path, "/v2/")
		if rest != "" && !strings.HasPrefix(rest, p.registryName+"/") {
			path = "/v2/" + p.registryName + "/" + rest
		}
	}
	upstreamURL.Path = path
	upstreamURL.RawQuery = r.URL.RawQuery

	// Create new request with the same method and body
	// IMPORTANT: Don't buffer the body - stream it directly
	upstreamReq, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL.String(), r.Body)
	if err != nil {
		log.Printf("[PROXY] Error creating request: %v", err)
		http.Error(w, "Proxy error", http.StatusBadGateway)
		return
	}

	// CRITICAL: Preserve the original Content-Length for uploads
	// Without this, large uploads may fail
	if r.ContentLength > 0 {
		upstreamReq.ContentLength = r.ContentLength
	}

	// Copy headers from original request
	p.copyRequestHeaders(r, upstreamReq)

	// Set Host header to upstream
	upstreamReq.Host = p.upstream.Host

	// Get Bearer token for this specific repository
	bearerToken := ""
	if p.authToken != "" {
		repoPath := p.extractRepoFromPath(r.URL.Path)
		if repoPath != "" {
			token, err := p.getTokenForRepo(repoPath)
			if err != nil {
				log.Printf("[PROXY] Failed to get token for %s: %v", repoPath, err)
				http.Error(w, "Authentication error", http.StatusBadGateway)
				return
			}
			bearerToken = token
			upstreamReq.Header.Set("Authorization", "Bearer "+token)
		}
	}

	// Log the request (with special marking for manifest operations)
	if strings.Contains(r.URL.Path, "/manifests/") {
		log.Printf("[PROXY] [MANIFEST] %s %s -> %s", r.Method, r.URL.Path, upstreamURL.String())
		if len(bearerToken) > 20 {
			log.Printf("[PROXY] [MANIFEST] Auth header: Bearer %s...", bearerToken[:20])
		}
		log.Printf("[PROXY] [MANIFEST] Content-Length: %d", r.ContentLength)
		log.Printf("[PROXY] [MANIFEST] Content-Type: %s", r.Header.Get("Content-Type"))
	} else {
		log.Printf("[PROXY] %s %s -> %s", r.Method, r.URL.Path, upstreamURL.String())
	}

	// Execute request
	resp, err := p.registryClient.Do(upstreamReq)
	if err != nil {
		log.Printf("[PROXY] Error forwarding request: %v", err)
		http.Error(w, "Upstream error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	p.copyResponseHeaders(resp, w)

	// Handle WWW-Authenticate header for 401 responses
	if resp.StatusCode == http.StatusUnauthorized {
		p.handleAuthChallenge(resp, w)
	}

	// Write status code
	w.WriteHeader(resp.StatusCode)

	// Log failed requests
	if resp.StatusCode >= 400 {
		log.Printf("[PROXY] [ERROR] %s %s -> %d", r.Method, r.URL.Path, resp.StatusCode)
	}

	// Stream response body with flushing for real-time streaming
	var bytesCopied int64
	if flusher, ok := w.(http.Flusher); ok {
		// Use a buffer for chunked streaming
		buf := make([]byte, 32*1024) // 32KB buffer
		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				written, writeErr := w.Write(buf[:n])
				bytesCopied += int64(written)
				if writeErr != nil {
					log.Printf("[PROXY] Error writing response: %v", writeErr)
					return
				}
				flusher.Flush()
			}
			if readErr != nil {
				if readErr != io.EOF {
					log.Printf("[PROXY] Error reading response: %v", readErr)
				}
				break
			}
		}
	} else {
		var err error
		bytesCopied, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Printf("[PROXY] Error copying response body: %v", err)
			return
		}
	}

	duration := time.Since(startTime)

	// Log with more detail for errors and manifests
	if resp.StatusCode >= 400 {
		if strings.Contains(r.URL.Path, "/manifests/") {
			log.Printf("[PROXY] %s %s -> %d ERROR [MANIFEST] (%d bytes, %v)", r.Method, r.URL.Path, resp.StatusCode, bytesCopied, duration)
		} else {
			log.Printf("[PROXY] %s %s -> %d ERROR (%d bytes, %v)", r.Method, r.URL.Path, resp.StatusCode, bytesCopied, duration)
		}
	} else {
		if strings.Contains(r.URL.Path, "/manifests/") {
			log.Printf("[PROXY] %s %s -> %d [MANIFEST] (%d bytes, %v)", r.Method, r.URL.Path, resp.StatusCode, bytesCopied, duration)
		} else {
			log.Printf("[PROXY] %s %s -> %d (%d bytes, %v)", r.Method, r.URL.Path, resp.StatusCode, bytesCopied, duration)
		}
	}
}

// copyRequestHeaders copies relevant headers from client request to upstream request
func (p *RegistryProxy) copyRequestHeaders(src *http.Request, dst *http.Request) {
	// Headers to forward (NOT Authorization - we use our own token)
	headersToForward := []string{
		"Accept",
		"Accept-Encoding",
		"Content-Type",
		"Content-Length",
		"Content-Range",
		"Range",
		"If-None-Match",
		"If-Match",
		"Docker-Content-Digest",
		"Docker-Distribution-API-Version",
		"User-Agent",
	}

	for _, h := range headersToForward {
		if v := src.Header.Get(h); v != "" {
			dst.Header.Set(h, v)
		}
	}

	// Handle chunked transfer encoding
	if src.TransferEncoding != nil {
		dst.TransferEncoding = src.TransferEncoding
	}

	// Preserve Content-Length if set
	if src.ContentLength > 0 {
		dst.ContentLength = src.ContentLength
	}
}

// copyResponseHeaders copies response headers from upstream to client
func (p *RegistryProxy) copyResponseHeaders(resp *http.Response, w http.ResponseWriter) {
	// Headers to forward back
	headersToForward := []string{
		"Content-Type",
		"Content-Length",
		"Content-Range",
		"Docker-Content-Digest",
		"Docker-Distribution-API-Version",
		"Docker-Upload-UUID",
		"ETag",
		"Location",
		"Range",
		"WWW-Authenticate",
		"X-Content-Type-Options",
	}

	for _, h := range headersToForward {
		if v := resp.Header.Get(h); v != "" {
			w.Header().Set(h, v)
		}
	}

	// Handle Location header for redirects and upload URLs
	if location := resp.Header.Get("Location"); location != "" {
		// If location is relative, it stays as-is
		// If it's absolute pointing to upstream, rewrite to our host
		if strings.HasPrefix(location, p.upstream.String()) {
			location = strings.Replace(location, p.upstream.String(), "", 1)
			w.Header().Set("Location", location)
		}
	}
}

// handleAuthChallenge handles WWW-Authenticate headers
// We pass through the auth challenge as-is - the client will authenticate
// directly with the upstream auth server (token server)
func (p *RegistryProxy) handleAuthChallenge(resp *http.Response, w http.ResponseWriter) {
	// Get the WWW-Authenticate header
	authHeader := resp.Header.Get("WWW-Authenticate")
	if authHeader == "" {
		return
	}

	// Log the auth challenge
	log.Printf("[PROXY] Auth challenge: %s", authHeader)

	// Pass through as-is - client authenticates with upstream's auth server
	// The token they receive will work for requests through our proxy
	// because we forward the Authorization header to upstream
	w.Header().Set("WWW-Authenticate", authHeader)
}
