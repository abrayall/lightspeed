package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"lightspeed/core/lib/ui"
	"lightspeed/core/lib/version"
	"lightspeed/platform/operator/api"
	"lightspeed/platform/operator/config"
	"lightspeed/platform/operator/proxy"
	"lightspeed/platform/operator/registry"
)

// Version is set by ldflags during build
var Version = "dev"

// CLI flags
var (
	port             string
	publicHost       string
	upstreamRegistry string
	defaultRegistry  string
	showVersion      bool
	tlsEnabled       bool
	tlsCert          string
	tlsKey           string
)

func init() {
	// Load defaults from config (which checks env vars)
	defaults := config.Load()

	flag.StringVar(&port, "port", defaults.Port, "Server port")
	flag.StringVar(&port, "p", defaults.Port, "Server port (shorthand)")
	flag.StringVar(&publicHost, "host", defaults.PublicHost, "Public hostname for registry")
	flag.StringVar(&upstreamRegistry, "upstream", defaults.UpstreamRegistry, "Upstream registry to proxy")
	flag.StringVar(&upstreamRegistry, "u", defaults.UpstreamRegistry, "Upstream registry (shorthand)")
	flag.StringVar(&defaultRegistry, "registry", defaults.DefaultRegistry, "Default container registry name")
	flag.StringVar(&defaultRegistry, "r", defaults.DefaultRegistry, "Default registry (shorthand)")
	flag.BoolVar(&showVersion, "version", false, "Show version and exit")
	flag.BoolVar(&showVersion, "v", false, "Show version (shorthand)")
	flag.BoolVar(&tlsEnabled, "tls", defaults.TLSEnabled, "Enable TLS/HTTPS")
	flag.StringVar(&tlsCert, "cert", defaults.TLSCert, "TLS certificate file (auto-generated if empty)")
	flag.StringVar(&tlsKey, "key", defaults.TLSKey, "TLS private key file (auto-generated if empty)")
}

func main() {
	// Get version from git if available
	if Version == "dev" {
		if v, err := version.GetFromGit("."); err == nil {
			Version = v.String()
		}
	}

	// Parse CLI flags
	flag.Parse()

	// Handle version flag
	if showVersion {
		fmt.Printf("Lightspeed Operator %s\n", Version)
		os.Exit(0)
	}

	// Print header
	ui.PrintHeader(Version)

	// Build config from CLI flags (which already have env/defaults applied)
	cfg := &config.Config{
		Port:             port,
		PublicHost:       publicHost,
		UpstreamRegistry: upstreamRegistry,
		DefaultRegistry:  defaultRegistry,
	}

	// Create router
	mux := http.NewServeMux()

	// Registry proxy for /v2/
	registryProxy, err := proxy.NewRegistryProxy(cfg.UpstreamRegistry, cfg.PublicHost)
	if err != nil {
		ui.PrintError("Failed to create registry proxy: %v", err)
		os.Exit(1)
	}
	registryProxy.SetAuthToken(config.GetDOToken())
	registryProxy.SetRegistryName(cfg.DefaultRegistry)
	mux.Handle("/v2/", registryProxy)

	// Sites API - uses built-in DO token
	sitesHandler := api.NewSitesHandler(config.GetDOToken(), cfg.DefaultRegistry)
	mux.Handle("/sites", sitesHandler)
	mux.Handle("/sites/", sitesHandler)

	// Health and version
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/version", handleVersion)

	// Root
	mux.HandleFunc("/", handleRoot)

	// Start server
	addr := ":" + cfg.Port
	ui.PrintSuccess("Operator started")
	fmt.Println()
	ui.PrintKeyValue("  Port", cfg.Port)
	if tlsEnabled {
		ui.PrintKeyValue("  TLS", "enabled")
	}
	ui.PrintKeyValue("  Upstream", cfg.UpstreamRegistry)
	fmt.Println()
	ui.PrintInfo("Endpoints:")
	fmt.Println("  • /v2/*                     - Registry proxy (push & pull)")
	fmt.Println("  • GET /sites                - List all sites")
	fmt.Println("  • POST /sites               - Create a site")
	fmt.Println("  • GET /sites/{name}         - Get site details")
	fmt.Println("  • DELETE /sites/{name}      - Delete a site")
	fmt.Println("  • POST /sites/{name}/deploy - Trigger deployment")
	fmt.Println("  • /health                   - Health check")
	fmt.Println("  • /version                  - Version info")
	fmt.Println()

	// Start image pruner (runs daily, after startup messages)
	pruner := registry.NewPruner(config.GetDOToken(), cfg.DefaultRegistry)
	pruner.Start()

	if tlsEnabled {
		// Generate or use provided certs
		certFile, keyFile, err := ensureTLSCerts(tlsCert, tlsKey)
		if err != nil {
			ui.PrintError("Failed to setup TLS: %v", err)
			os.Exit(1)
		}
		log.Fatal(http.ListenAndServeTLS(addr, certFile, keyFile, mux))
	} else {
		log.Fatal(http.ListenAndServe(addr, mux))
	}
}

// ensureTLSCerts returns cert and key paths, generating self-signed if needed
func ensureTLSCerts(certPath, keyPath string) (string, string, error) {
	// If both provided, use them
	if certPath != "" && keyPath != "" {
		return certPath, keyPath, nil
	}

	// Generate self-signed cert
	ui.PrintInfo("Generating self-signed certificate...")

	// Create temp directory for certs
	tmpDir := filepath.Join(os.TempDir(), "lightspeed-certs")
	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		return "", "", err
	}

	certFile := filepath.Join(tmpDir, "cert.pem")
	keyFile := filepath.Join(tmpDir, "key.pem")

	// Generate private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	// Create certificate template
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Lightspeed"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", "*.localhost", "host.docker.internal"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	// Write cert file
	certOut, err := os.Create(certFile)
	if err != nil {
		return "", "", err
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	certOut.Close()

	// Write key file
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return "", "", err
	}
	keyBytes, _ := x509.MarshalECPrivateKey(privateKey)
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	keyOut.Close()

	return certFile, keyFile, nil
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"name":"Lightspeed","status":"ok"}`))
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"version":"%s"}`, Version)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"name":"Lightspeed","version":"%s"}`, Version)
}
