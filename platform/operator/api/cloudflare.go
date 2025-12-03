package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const cloudflareAPI = "https://api.cloudflare.com/client/v4"

// CloudflareClient handles Cloudflare API interactions
type CloudflareClient struct {
	token  string
	zoneID string
}

// NewCloudflareClient creates a new Cloudflare client
func NewCloudflareClient(token string) *CloudflareClient {
	return &CloudflareClient{
		token: token,
	}
}

// CloudflareResponse is the standard CF API response
type CloudflareResponse struct {
	Success bool              `json:"success"`
	Errors  []CloudflareError `json:"errors"`
	Result  json.RawMessage   `json:"result"`
}

type CloudflareError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CloudflareZone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CloudflareDNSRecord struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

// getZoneID finds the zone ID for lightspeed.ee
func (c *CloudflareClient) getZoneID() (string, error) {
	if c.zoneID != "" {
		return c.zoneID, nil
	}

	req, err := http.NewRequest("GET", cloudflareAPI+"/zones?name=lightspeed.ee", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var cfResp CloudflareResponse
	if err := json.Unmarshal(body, &cfResp); err != nil {
		return "", err
	}

	if !cfResp.Success {
		if len(cfResp.Errors) > 0 {
			return "", fmt.Errorf("cloudflare error: %s", cfResp.Errors[0].Message)
		}
		return "", fmt.Errorf("cloudflare API failed")
	}

	var zones []CloudflareZone
	if err := json.Unmarshal(cfResp.Result, &zones); err != nil {
		return "", err
	}

	if len(zones) == 0 {
		return "", fmt.Errorf("zone lightspeed.ee not found")
	}

	c.zoneID = zones[0].ID
	return c.zoneID, nil
}

// findDNSRecord finds a DNS record by name
func (c *CloudflareClient) findDNSRecord(name string) (*CloudflareDNSRecord, error) {
	zoneID, err := c.getZoneID()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/zones/%s/dns_records?type=CNAME&name=%s", cloudflareAPI, zoneID, name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var cfResp CloudflareResponse
	if err := json.Unmarshal(body, &cfResp); err != nil {
		return nil, err
	}

	if !cfResp.Success {
		if len(cfResp.Errors) > 0 {
			return nil, fmt.Errorf("cloudflare error: %s", cfResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("cloudflare API failed")
	}

	var records []CloudflareDNSRecord
	if err := json.Unmarshal(cfResp.Result, &records); err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	return &records[0], nil
}

// EnsureCNAME creates or updates a CNAME record
func (c *CloudflareClient) EnsureCNAME(subdomain, target string) error {
	// Ensure full domain name
	fullName := subdomain
	if !strings.HasSuffix(subdomain, ".lightspeed.ee") {
		fullName = subdomain + ".lightspeed.ee"
	}

	// Remove https:// prefix if present
	target = strings.TrimPrefix(target, "https://")
	target = strings.TrimPrefix(target, "http://")

	// Check if record exists
	existing, err := c.findDNSRecord(fullName)
	if err != nil {
		return err
	}

	record := CloudflareDNSRecord{
		Type:    "CNAME",
		Name:    fullName,
		Content: target,
		TTL:     1, // Auto
		Proxied: false,
	}

	zoneID, err := c.getZoneID()
	if err != nil {
		return err
	}

	var req *http.Request
	if existing != nil {
		// Update existing record
		if existing.Content == target {
			log.Printf("DNS record %s already points to %s", fullName, target)
			return nil
		}

		record.ID = existing.ID
		body, _ := json.Marshal(record)
		url := fmt.Sprintf("%s/zones/%s/dns_records/%s", cloudflareAPI, zoneID, existing.ID)
		req, err = http.NewRequest("PUT", url, bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		log.Printf("Updating DNS record %s -> %s", fullName, target)
	} else {
		// Create new record
		body, _ := json.Marshal(record)
		url := fmt.Sprintf("%s/zones/%s/dns_records", cloudflareAPI, zoneID)
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		log.Printf("Creating DNS record %s -> %s", fullName, target)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var cfResp CloudflareResponse
	if err := json.Unmarshal(body, &cfResp); err != nil {
		return err
	}

	if !cfResp.Success {
		if len(cfResp.Errors) > 0 {
			return fmt.Errorf("cloudflare error: %s", cfResp.Errors[0].Message)
		}
		return fmt.Errorf("cloudflare API failed")
	}

	log.Printf("DNS record %s successfully configured", fullName)
	return nil
}
