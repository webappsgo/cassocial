package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// GeoIP provides geographic IP lookup functionality
type GeoIP struct {
	dataDir    string
	dbPath     string
	enabled    bool
}

// NewGeoIP creates a new GeoIP instance
func NewGeoIP(dataDir string) *GeoIP {
	dbPath := filepath.Join(dataDir, "security", "geoip", "GeoLite2-Country.mmdb")

	// Check if database exists
	enabled := false
	if _, err := os.Stat(dbPath); err == nil {
		enabled = true
	}

	return &GeoIP{
		dataDir: dataDir,
		dbPath:  dbPath,
		enabled: enabled,
	}
}

// Lookup looks up country information for an IP address
func (g *GeoIP) Lookup(ip string) (string, string, error) {
	if !g.enabled {
		return "", "", fmt.Errorf("GeoIP database not available")
	}

	// TODO: Implement actual GeoIP lookup using maxminddb-golang
	// For now, return unknown
	return "Unknown", "XX", nil
}

// IsEnabled returns whether GeoIP is enabled
func (g *GeoIP) IsEnabled() bool {
	return g.enabled
}

// DownloadDatabase downloads the latest GeoIP database
func (g *GeoIP) DownloadDatabase() error {
	log.Println("Downloading GeoIP database...")

	// Ensure directory exists
	geoipDir := filepath.Join(g.dataDir, "security", "geoip")
	if err := os.MkdirAll(geoipDir, 0755); err != nil {
		return fmt.Errorf("failed to create geoip directory: %w", err)
	}

	// Download from ip-location-db (free, no API key required)
	// https://github.com/sapics/ip-location-db
	url := "https://github.com/sapics/ip-location-db/raw/main/geolite2-country/geolite2-country-ipv4.csv.gz"

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download GeoIP database: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download GeoIP database: status %d", resp.StatusCode)
	}

	// Save to file
	tempFile := filepath.Join(geoipDir, "geoip.csv.gz")
	out, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to save GeoIP database: %w", err)
	}

	log.Println("GeoIP database downloaded successfully")

	// TODO: Extract and process CSV
	// TODO: Convert to efficient lookup format

	g.enabled = true
	return nil
}

// CheckCountryBlocked checks if a country is blocked
func (g *GeoIP) CheckCountryBlocked(ip string, blockedCountries []string) (bool, error) {
	if !g.enabled || len(blockedCountries) == 0 {
		return false, nil
	}

	// Lookup country
	_, countryCode, err := g.Lookup(ip)
	if err != nil {
		return false, err
	}

	// Check if country is in blocked list
	for _, blocked := range blockedCountries {
		if blocked == countryCode {
			return true, nil
		}
	}

	return false, nil
}

// GeoIPMiddleware returns middleware that blocks requests from certain countries
func GeoIPMiddleware(geoip *GeoIP, blockedCountries []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if GeoIP not enabled or no countries blocked
			if !geoip.IsEnabled() || len(blockedCountries) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Get client IP
			ip := getClientIPFromRequest(r)

			// Check if country is blocked
			blocked, err := geoip.CheckCountryBlocked(ip, blockedCountries)
			if err != nil {
				// On error, allow request (fail open)
				next.ServeHTTP(w, r)
				return
			}

			if blocked {
				http.Error(w, "Access denied from your location", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetStats returns GeoIP statistics
func (g *GeoIP) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled": g.enabled,
		"db_path": g.dbPath,
		"db_exists": g.enabled,
	}
}
