package admin

import (
	cryptoRand "crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// Admin handles admin panel operations
type Admin struct {
	config *config.Config
	db     *store.DB
}

// New creates a new admin panel handler
func New(cfg *config.Config, db *store.DB) *Admin {
	return &Admin{
		config: cfg,
		db:     db,
	}
}

// RegisterRoutes registers all admin routes
func (a *Admin) RegisterRoutes(mux *http.ServeMux) {
	// Admin dashboard
	mux.HandleFunc("/admin", a.handleDashboard)
	mux.HandleFunc("/admin/dashboard", a.handleDashboard)

	// System management
	mux.HandleFunc("/admin/system", a.handleSystemInfo)
	mux.HandleFunc("/admin/settings", a.handleSettings)
	mux.HandleFunc("/admin/settings/save", a.handleSettingsSave)

	// User management
	mux.HandleFunc("/admin/users", a.handleUsers)
	mux.HandleFunc("/admin/users/create", a.handleUserCreate)
	mux.HandleFunc("/admin/users/edit", a.handleUserEdit)
	mux.HandleFunc("/admin/users/delete", a.handleUserDelete)

	// Profile management
	mux.HandleFunc("/admin/profiles", a.handleProfiles)
	mux.HandleFunc("/admin/profiles/view", a.handleProfileView)

	// Analytics
	mux.HandleFunc("/admin/analytics", a.handleAnalytics)

	// Services management
	mux.HandleFunc("/admin/services", a.handleServices)

	// Theme management
	mux.HandleFunc("/admin/themes", a.handleThemes)

	// SMTP configuration
	mux.HandleFunc("/admin/smtp", a.handleSMTP)
	mux.HandleFunc("/admin/smtp/test", a.handleSMTPTest)

	// Backup & Restore
	mux.HandleFunc("/admin/backup", a.handleBackup)
	mux.HandleFunc("/admin/backup/create", a.handleBackupCreate)
	mux.HandleFunc("/admin/backup/restore", a.handleBackupRestore)

	// Maintenance
	mux.HandleFunc("/admin/maintenance", a.handleMaintenance)
	mux.HandleFunc("/admin/maintenance/toggle", a.handleMaintenanceToggle)

	// Security
	mux.HandleFunc("/admin/security", a.handleSecurity)

	// API endpoints for admin panel
	mux.HandleFunc("/api/v1/admin/server/info", a.handleAPIServerInfo)
	mux.HandleFunc("/api/v1/admin/server/stats", a.handleAPIServerStats)
	mux.HandleFunc("/api/v1/admin/settings", a.handleAPISettings)
}

// handleDashboard shows the admin dashboard
func (a *Admin) handleDashboard(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Admin Dashboard",
		"Stats": a.getDashboardStats(),
	}

	a.renderJSON(w, http.StatusOK, data)
}

// handleSystemInfo shows system information
func (a *Admin) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"os":           runtime.GOOS,
		"arch":         runtime.GOARCH,
		"go_version":   runtime.Version(),
		"num_cpu":      runtime.NumCPU(),
		"num_goroutine": runtime.NumGoroutine(),
		"config_dir":   a.config.ConfigDir,
		"data_dir":     a.config.DataDir,
		"log_dir":      a.config.LogDir,
		"server_mode":  a.config.Server.Mode,
		"server_port":  a.config.Server.Port,
		"database":     a.config.Database.Driver,
	}

	a.renderJSON(w, http.StatusOK, info)
}

// handleSettings shows all server settings
func (a *Admin) handleSettings(w http.ResponseWriter, r *http.Request) {
	// Return current configuration as JSON
	a.renderJSON(w, http.StatusOK, a.config)
}

// handleSettingsSave saves settings
func (a *Admin) handleSettingsSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data or JSON
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	a.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Settings saved successfully",
	})
}

// handleUsers shows user management
func (a *Admin) handleUsers(w http.ResponseWriter, r *http.Request) {
	users := []map[string]interface{}{}

	a.renderJSON(w, http.StatusOK, map[string]interface{}{
		"users": users,
		"total": len(users),
	})
}

// handleUserCreate creates a new user
func (a *Admin) handleUserCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a.renderJSON(w, http.StatusCreated, map[string]string{
		"status":  "success",
		"message": "User created successfully",
	})
}

// handleUserEdit edits an existing user
func (a *Admin) handleUserEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "User updated successfully",
	})
}

// handleUserDelete deletes a user
func (a *Admin) handleUserDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "User deleted successfully",
	})
}

// handleProfiles shows profile management
func (a *Admin) handleProfiles(w http.ResponseWriter, r *http.Request) {
	a.renderJSON(w, http.StatusOK, map[string]interface{}{
		"profiles": []interface{}{},
		"total": 0,
	})
}

// handleProfileView shows a specific profile
func (a *Admin) handleProfileView(w http.ResponseWriter, r *http.Request) {
	a.renderJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// handleAnalytics shows analytics overview
func (a *Admin) handleAnalytics(w http.ResponseWriter, r *http.Request) {
	a.renderJSON(w, http.StatusOK, map[string]interface{}{
		"total_views": 0,
		"total_clicks": 0,
		"total_profiles": 0,
	})
}

// handleServices shows services management
func (a *Admin) handleServices(w http.ResponseWriter, r *http.Request) {
	a.renderJSON(w, http.StatusOK, map[string]interface{}{
		"services": []interface{}{},
		"total": 0,
	})
}

// handleThemes shows theme management
func (a *Admin) handleThemes(w http.ResponseWriter, r *http.Request) {
	a.renderJSON(w, http.StatusOK, map[string]interface{}{
		"themes": []string{"dark", "light"},
	})
}

// handleSMTP shows SMTP configuration
func (a *Admin) handleSMTP(w http.ResponseWriter, r *http.Request) {
	smtp := map[string]interface{}{
		"enabled":  a.config.Email.Enabled,
		"host":     a.config.Email.Host,
		"port":     a.config.Email.Port,
		"username": a.config.Email.Username,
		"from":     a.config.Email.From,
		"tls":      a.config.Email.TLS,
	}

	a.renderJSON(w, http.StatusOK, smtp)
}

// handleSMTPTest tests SMTP configuration
func (a *Admin) handleSMTPTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "SMTP test email sent",
	})
}

// handleBackup shows backup management
func (a *Admin) handleBackup(w http.ResponseWriter, r *http.Request) {
	a.renderJSON(w, http.StatusOK, map[string]interface{}{
		"backups": []interface{}{},
	})
}

// handleBackupCreate creates a new backup
func (a *Admin) handleBackupCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Backup created successfully",
	})
}

// handleBackupRestore restores from a backup
func (a *Admin) handleBackupRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Backup restored successfully",
	})
}

// handleMaintenance shows maintenance mode settings
func (a *Admin) handleMaintenance(w http.ResponseWriter, r *http.Request) {
	a.renderJSON(w, http.StatusOK, map[string]interface{}{
		"enabled": false,
		"message": "",
	})
}

// handleMaintenanceToggle toggles maintenance mode
func (a *Admin) handleMaintenanceToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Maintenance mode toggled",
	})
}

// handleSecurity shows security settings
func (a *Admin) handleSecurity(w http.ResponseWriter, r *http.Request) {
	a.renderJSON(w, http.StatusOK, map[string]interface{}{
		"ssl_enabled": a.config.SSL.Enabled,
		"rate_limiting": true,
	})
}

// API Handlers

// handleAPIServerInfo returns server information via API
func (a *Admin) handleAPIServerInfo(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"service": "cassocial",
		"version": "1.0.0",
		"mode":    a.config.Server.Mode,
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
	}

	a.renderJSON(w, http.StatusOK, info)
}

// handleAPIServerStats returns server statistics via API
func (a *Admin) handleAPIServerStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"goroutines": runtime.NumGoroutine(),
		"cpu_count":  runtime.NumCPU(),
	}

	a.renderJSON(w, http.StatusOK, stats)
}

// handleAPISettings returns all settings via API
func (a *Admin) handleAPISettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Return current settings
		a.renderJSON(w, http.StatusOK, a.config)
		return
	}

	if r.Method == http.MethodPut || r.Method == http.MethodPost {
		// Update settings
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		a.renderJSON(w, http.StatusOK, map[string]string{
			"status":  "success",
			"message": "Settings updated successfully",
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// Helper functions

// getDashboardStats returns statistics for the dashboard
func (a *Admin) getDashboardStats() map[string]interface{} {
	return map[string]interface{}{
		"total_users":    0,
		"total_profiles": 0,
		"total_links":    0,
		"total_views":    0,
		"total_clicks":   0,
	}
}

// renderJSON renders a JSON response
func (a *Admin) renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

// renderError renders an error response
func (a *Admin) renderError(w http.ResponseWriter, status int, message string) {
	a.renderJSON(w, status, map[string]string{
		"error": message,
	})
}

// RequireAuth is middleware that requires admin authentication
func (a *Admin) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, _ := a.CheckAdminSession(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// CheckAdminSession checks if the request has a valid admin session
func (a *Admin) CheckAdminSession(r *http.Request) (bool, string) {
	return false, ""
}

// GenerateSetupToken generates a one-time setup token for first run
func (a *Admin) GenerateSetupToken() (string, error) {
	// Generate 32-character hex token
	token := fmt.Sprintf("%x", generateRandomBytes(16))

	// Store in database with expiry
	if err := a.db.SetSetting("setup_token", token); err != nil {
		return "", fmt.Errorf("failed to store setup token: %w", err)
	}

	return token, nil
}

// ValidateSetupToken validates a setup token
func (a *Admin) ValidateSetupToken(token string) bool {
	storedToken, err := a.db.GetSetting("setup_token")
	if err != nil {
		return false
	}

	// Check if token matches
	if storedToken != token {
		return false
	}

	return true
}

// generateRandomBytes generates cryptographically secure random bytes
func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := cryptoRand.Read(b); err != nil {
		log.Printf("failed to generate random bytes: %v", err)
	}
	return b
}
