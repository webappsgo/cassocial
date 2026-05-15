package handler

import (
	"encoding/json"
	"net/http"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server"
	"github.com/casapps/cassocial/src/server/store"
)

// SetupHandler handles first-run setup wizard
type SetupHandler struct {
	config *config.Config
	db     *store.DB
}

// NewSetupHandler creates a new setup handler
func NewSetupHandler(cfg *config.Config, db *store.DB) *SetupHandler {
	return &SetupHandler{
		config: cfg,
		db:     db,
	}
}

// SetupStep represents a step in the setup wizard
type SetupStep struct {
	Step        int    `json:"step"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

// HandleSetupStatus checks if setup is needed
func (h *SetupHandler) HandleSetupStatus(w http.ResponseWriter, r *http.Request) {
	// Check if already initialized
	initialized, _ := h.db.GetSetting("initialized")

	if initialized == "true" {
		h.renderJSON(w, http.StatusOK, map[string]interface{}{
			"initialized": true,
			"message":     "System already configured",
		})
		return
	}

	// Return setup steps
	steps := []SetupStep{
		{Step: 1, Title: "Welcome", Description: "Welcome to Cassocial"},
		{Step: 2, Title: "Basic Configuration", Description: "Configure basic settings"},
		{Step: 3, Title: "Domain & Access", Description: "Configure domain and access settings"},
		{Step: 4, Title: "Email Configuration", Description: "Configure SMTP for email notifications"},
		{Step: 5, Title: "Features & Limits", Description: "Configure features and limits"},
		{Step: 6, Title: "Database", Description: "Configure database settings"},
		{Step: 7, Title: "Review & Complete", Description: "Review and complete setup"},
	}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"initialized": false,
		"steps":       steps,
	})
}

// HandleSetupWelcome handles the welcome step
func (h *SetupHandler) HandleSetupWelcome(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"step":         1,
		"title":        "Welcome to Cassocial",
		"description":  "Let's get your link aggregator set up",
		"site_name":    h.config.Cassocial.SiteName,
	}

	h.renderJSON(w, http.StatusOK, data)
}

// HandleSetupBasic handles basic configuration step
func (h *SetupHandler) HandleSetupBasic(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := map[string]interface{}{
			"step":             2,
			"title":            "Basic Configuration",
			"site_name":        h.config.Cassocial.SiteName,
			"site_description": h.config.Cassocial.SiteDescription,
		}
		h.renderJSON(w, http.StatusOK, data)
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			SiteName        string `json:"site_name"`
			SiteDescription string `json:"site_description"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.renderError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		h.db.SetSetting("site_name", req.SiteName)
		h.db.SetSetting("site_description", req.SiteDescription)
		h.renderJSON(w, http.StatusOK, map[string]string{
			"status": "success",
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HandleSetupDomain handles domain and access configuration
func (h *SetupHandler) HandleSetupDomain(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := map[string]interface{}{
			"step":    3,
			"title":   "Domain & Access",
			"port":    h.config.Server.Port,
			"address": h.config.Server.Address,
		}
		h.renderJSON(w, http.StatusOK, data)
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			Domain string `json:"domain"`
			Port   int    `json:"port"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.renderError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		if req.Domain != "" {
			h.db.SetSetting("site_url", req.Domain)
		}
		h.renderJSON(w, http.StatusOK, map[string]string{
			"status": "success",
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HandleSetupEmail handles email configuration
func (h *SetupHandler) HandleSetupEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := map[string]interface{}{
			"step":     4,
			"title":    "Email Configuration",
			"enabled":  h.config.Email.Enabled,
			"host":     h.config.Email.Host,
			"port":     h.config.Email.Port,
			"username": h.config.Email.Username,
			"from":     h.config.Email.From,
			"tls":      h.config.Email.TLS,
		}
		h.renderJSON(w, http.StatusOK, data)
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			Enabled  bool   `json:"enabled"`
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Username string `json:"username"`
			Password string `json:"password"`
			From     string `json:"from"`
			TLS      bool   `json:"tls"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.renderError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		if req.Enabled {
			h.db.SetSetting("email_enabled", "true")
		} else {
			h.db.SetSetting("email_enabled", "false")
		}
		h.db.SetSetting("smtp_host", req.Host)
		h.db.SetSetting("smtp_user", req.Username)
		h.db.SetSetting("smtp_from_address", req.From)
		h.renderJSON(w, http.StatusOK, map[string]string{
			"status": "success",
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HandleSetupFeatures handles features and limits configuration
func (h *SetupHandler) HandleSetupFeatures(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := map[string]interface{}{
			"step":                   5,
			"title":                  "Features & Limits",
			"allow_registration":     h.config.Cassocial.AllowRegistration,
			"max_profiles_per_user":  h.config.Cassocial.MaxProfilesPerUser,
			"max_links_per_profile":  h.config.Cassocial.MaxLinksPerProfile,
		}
		h.renderJSON(w, http.StatusOK, data)
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			AllowRegistration    bool `json:"allow_registration"`
			MaxProfilesPerUser   int  `json:"max_profiles_per_user"`
			MaxLinksPerProfile   int  `json:"max_links_per_profile"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.renderError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		if req.AllowRegistration {
			h.db.SetSetting("allow_registration", "true")
		} else {
			h.db.SetSetting("allow_registration", "false")
		}
		h.renderJSON(w, http.StatusOK, map[string]string{
			"status": "success",
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HandleSetupDatabase handles database configuration
func (h *SetupHandler) HandleSetupDatabase(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := map[string]interface{}{
			"step":   6,
			"title":  "Database",
			"driver": h.config.Database.Driver,
		}
		h.renderJSON(w, http.StatusOK, data)
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			Driver   string `json:"driver"` // sqlite, postgres, mysql
			Host     string `json:"host,omitempty"`
			Port     int    `json:"port,omitempty"`
			Name     string `json:"name"`
			User     string `json:"user,omitempty"`
			Password string `json:"password,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.renderError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		if err := h.db.Ping(); err != nil {
			h.renderError(w, http.StatusBadRequest, "Database connection failed: "+err.Error())
			return
		}
		h.renderJSON(w, http.StatusOK, map[string]string{
			"status": "success",
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HandleSetupComplete completes the setup wizard
func (h *SetupHandler) HandleSetupComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create admin user
	var req struct {
		AdminUsername string `json:"admin_username"`
		AdminEmail    string `json:"admin_email"`
		AdminPassword string `json:"admin_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Validate admin credentials
	if len(req.AdminUsername) < 3 {
		h.renderError(w, http.StatusBadRequest, "Username must be at least 3 characters")
		return
	}

	if len(req.AdminPassword) < 8 {
		h.renderError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	// Hash password
	passwordHash, err := server.HashPassword(req.AdminPassword)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create admin user
	admin := &store.User{
		ID:            generateUUID(),
		Username:      req.AdminUsername,
		Email:         req.AdminEmail,
		PasswordHash:  passwordHash,
		Role:          "admin",
		Status:        "active",
		EmailVerified: true,
	}

	if err := h.db.CreateUser(admin); err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to create admin user")
		return
	}

	// Mark as initialized
	if err := h.db.SetSetting("initialized", "true"); err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to complete setup")
		return
	}

	// Save config
	if err := h.config.Save(); err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to save configuration")
		return
	}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Setup completed successfully",
		"admin": map[string]string{
			"username": req.AdminUsername,
			"email":    req.AdminEmail,
		},
	})
}

// renderJSON renders a JSON response
func (h *SetupHandler) renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// renderError renders an error response
func (h *SetupHandler) renderError(w http.ResponseWriter, status int, message string) {
	h.renderJSON(w, status, map[string]string{
		"error": message,
	})
}
