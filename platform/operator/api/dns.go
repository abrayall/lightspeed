package api

import (
	"encoding/json"
	"log"
	"time"
)

// DNSSyncWorker periodically checks apps and ensures DNS records exist
type DNSSyncWorker struct {
	handler  *SitesHandler
	interval time.Duration
}

// NewDNSSyncWorker creates a new DNS sync worker
func NewDNSSyncWorker(handler *SitesHandler, interval time.Duration) *DNSSyncWorker {
	return &DNSSyncWorker{
		handler:  handler,
		interval: interval,
	}
}

// Start begins the DNS sync worker
func (w *DNSSyncWorker) Start() {
	// Sync all sites on startup
	log.Printf("[DNS Sync] Initial sync of all sites")
	w.syncAllDNS()

	// Then start periodic sync for new sites only
	go w.run()
}

func (w *DNSSyncWorker) run() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Printf("[DNS Sync] Worker started, checking new sites every %v", w.interval)

	for range ticker.C {
		w.syncNewSitesDNS()
	}
}

// syncAllDNS syncs DNS for all apps (used on startup)
func (w *DNSSyncWorker) syncAllDNS() {
	// Get all apps
	resp, err := w.handler.doRequest("GET", "/apps", "Bearer "+w.handler.defaultToken, nil)
	if err != nil {
		log.Printf("[DNS Sync] Failed to list apps: %v", err)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Apps []struct {
			Spec struct {
				Name string `json:"name"`
			} `json:"spec"`
			DefaultIngress string `json:"default_ingress"`
		} `json:"apps"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[DNS Sync] Failed to parse apps: %v", err)
		return
	}

	// For each app with a default_ingress, ensure DNS exists
	count := 0
	for _, app := range result.Apps {
		if app.DefaultIngress != "" {
			appName := app.Spec.Name
			if err := w.handler.cfClient.EnsureCNAME(appName, app.DefaultIngress); err != nil {
				log.Printf("[DNS Sync] Failed to sync DNS for %s: %v", appName, err)
			} else {
				count++
			}
		}
	}
	log.Printf("[DNS Sync] Initial sync complete (%d apps checked)", count)
}

// syncNewSitesDNS only syncs DNS for recently created apps (last 10 minutes)
func (w *DNSSyncWorker) syncNewSitesDNS() {
	// Get all apps
	resp, err := w.handler.doRequest("GET", "/apps", "Bearer "+w.handler.defaultToken, nil)
	if err != nil {
		log.Printf("[DNS Sync] Failed to list apps: %v", err)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Apps []struct {
			Spec struct {
				Name string `json:"name"`
			} `json:"spec"`
			DefaultIngress string    `json:"default_ingress"`
			CreatedAt      time.Time `json:"created_at"`
		} `json:"apps"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[DNS Sync] Failed to parse apps: %v", err)
		return
	}

	// Only check apps created in the last 10 minutes
	cutoff := time.Now().Add(-10 * time.Minute)
	for _, app := range result.Apps {
		if app.CreatedAt.After(cutoff) && app.DefaultIngress != "" {
			appName := app.Spec.Name
			if err := w.handler.cfClient.EnsureCNAME(appName, app.DefaultIngress); err != nil {
				log.Printf("[DNS Sync] Failed to sync DNS for %s: %v", appName, err)
			}
		}
	}
}
