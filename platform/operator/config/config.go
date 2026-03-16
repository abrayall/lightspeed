package config

import (
	"log"
	"os"
)

// GetDOToken returns the DigitalOcean API token from environment
func GetDOToken() string {
	return requireEnv("DIGITALOCEAN_TOKEN")
}

// GetCFToken returns the Cloudflare API token from environment
func GetCFToken() string {
	return requireEnv("CLOUDFLARE_TOKEN")
}

// GetOperatorToken returns the operator API token from environment
func GetOperatorToken() string {
	return requireEnv("OPERATOR_TOKEN")
}

// Config holds operator configuration
type Config struct {
	Port             string
	PublicHost       string
	UpstreamRegistry string
	DefaultRegistry  string
	TLSEnabled       bool
	TLSCert          string
	TLSKey           string
	OperatorURL      string
	OperatorToken    string
}

// Load loads configuration from environment
func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "8080"),
		PublicHost:       getEnv("PUBLIC_HOST", "localhost:8080"),
		UpstreamRegistry: getEnv("UPSTREAM_REGISTRY", "registry.digitalocean.com"),
		DefaultRegistry:  getEnv("DEFAULT_REGISTRY", "lightspeed-images"),
		TLSEnabled:       getEnv("TLS_ENABLED", "") != "",
		TLSCert:          getEnv("TLS_CERT", ""),
		TLSKey:           getEnv("TLS_KEY", ""),
		OperatorURL:      getEnv("OPERATOR_URL", "https://operator.lightspeed.ee"),
		OperatorToken:    GetOperatorToken(),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func requireEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("[CONFIG] WARNING: %s environment variable is not set", key)
	}
	return value
}
