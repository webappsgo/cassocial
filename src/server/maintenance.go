package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/casapps/cassocial/src/server/store"
)

// MaintenanceMode handles maintenance mode operations
type MaintenanceMode struct {
	db *store.DB
}

// NewMaintenanceMode creates a new maintenance mode handler
func NewMaintenanceMode(db *store.DB) *MaintenanceMode {
	return &MaintenanceMode{
		db: db,
	}
}

// IsEnabled checks if maintenance mode is enabled
func (m *MaintenanceMode) IsEnabled() bool {
	enabled, err := m.db.GetSetting("maintenance_mode")
	if err != nil {
		return false
	}
	return enabled == "true"
}

// Enable enables maintenance mode
func (m *MaintenanceMode) Enable(message string) error {
	if err := m.db.SetSetting("maintenance_mode", "true"); err != nil {
		return err
	}

	if message != "" {
		if err := m.db.SetSetting("maintenance_message", message); err != nil {
			return err
		}
	}

	return nil
}

// Disable disables maintenance mode
func (m *MaintenanceMode) Disable() error {
	return m.db.SetSetting("maintenance_mode", "false")
}

// GetMessage returns the maintenance mode message
func (m *MaintenanceMode) GetMessage() string {
	message, err := m.db.GetSetting("maintenance_message")
	if err != nil || message == "" {
		return "We are performing scheduled maintenance. Please check back soon."
	}
	return message
}

// IsIPBypassed checks if an IP is in the bypass list
func (m *MaintenanceMode) IsIPBypassed(ip string) bool {
	// Localhost always bypasses
	if ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		return true
	}

	return false
}

// MaintenanceMiddleware returns middleware that enforces maintenance mode
func MaintenanceMiddleware(mm *MaintenanceMode) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip maintenance check for certain paths
			exemptPaths := []string{
				"/healthz",
				"/api/v1/healthz",
				"/admin",
				"/api/v1/admin",
			}

			for _, path := range exemptPaths {
				if r.URL.Path == path || len(r.URL.Path) > len(path) && r.URL.Path[:len(path)+1] == path+"/" {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Check if maintenance mode is enabled
			if !mm.IsEnabled() {
				next.ServeHTTP(w, r)
				return
			}

			// Check if IP is bypassed
			ip := getClientIPFromRequest(r)
			if mm.IsIPBypassed(ip) {
				next.ServeHTTP(w, r)
				return
			}

			// Return maintenance mode response
			message := mm.GetMessage()

			// API requests get JSON response
			if r.Header.Get("Accept") == "application/json" ||
			   len(r.URL.Path) > 4 && r.URL.Path[:5] == "/api/" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "Service temporarily unavailable",
					"message": message,
					"retry_after": 3600,
				})
				return
			}

			// Web requests get HTML response
			html := generateMaintenanceHTML(message)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(html))
		})
	}
}

// generateMaintenanceHTML generates maintenance mode HTML page
func generateMaintenanceHTML(message string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Maintenance Mode - Cassocial</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #fff;
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
            margin: 0;
            padding: 20px;
        }
        .container {
            text-align: center;
            max-width: 600px;
        }
        h1 { font-size: 3em; margin: 0 0 20px 0; }
        p { font-size: 1.2em; line-height: 1.6; opacity: 0.9; }
        .icon { font-size: 5em; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">🔧</div>
        <h1>Maintenance Mode</h1>
        <p>` + message + `</p>
        <p style="font-size: 0.9em; margin-top: 40px;">
            <a href="/" style="color: #fff;">Refresh Page</a>
        </p>
    </div>
</body>
</html>`
}

// SelfHealingCheck performs self-healing checks and fixes common issues
func SelfHealingCheck(db *store.DB, dataDir string) error {
	log.Println("Running self-healing checks...")

	// Check database connection
	if err := db.Ping(); err != nil {
		log.Printf("Database connection issue detected: %v", err)
	}

	// Check critical directories exist
	dirs := []string{
		filepath.Join(dataDir, "db"),
		filepath.Join(dataDir, "backup"),
		filepath.Join(dataDir, "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Failed to create directory %s: %v", dir, err)
		}
	}

	log.Println("Self-healing check complete")
	return nil
}
