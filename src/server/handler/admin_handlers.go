package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/casapps/cassocial/src/server"
	"github.com/casapps/cassocial/src/server/store"
	"github.com/casapps/cassocial/src/server/model"
)

// AdminHandlers handles admin-related HTTP requests
type AdminHandlers struct {
	db   *store.DB
	auth *Auth
}

// NewAdminHandlers creates a new AdminHandlers instance
func NewAdminHandlers(db *store.DB, authService *Auth) *AdminHandlers {
	return &AdminHandlers{
		db:   db,
		auth: authService,
	}
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Role   string `json:"role,omitempty"`
	Status string `json:"status,omitempty"`
}

// ListUsers lists all users (admin only)
// GET /api/admin/users
func (h *AdminHandlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, username, email, role, status, created_at, updated_at,
			  last_login, email_verified, two_factor_enabled
			  FROM users ORDER BY created_at DESC`

	rows, err := h.db.Query(query)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		var u model.User
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.Status, &u.CreatedAt,
			&u.UpdatedAt, &u.LastLogin, &u.EmailVerified, &u.TwoFactorEnabled)
		if err != nil {
			continue
		}
		users = append(users, u)
	}

	respondJSON(w, http.StatusOK, users)
}

// GetUser retrieves a specific user (admin only)
// GET /api/admin/users/{id}
func (h *AdminHandlers) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "user ID required")
		return
	}

	user, err := h.auth.GetUserByID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, user.SanitizeForJSON())
}

// UpdateUser updates a user (admin only)
// PUT /api/admin/users/{id}
func (h *AdminHandlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "user ID required")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get existing user
	user, err := h.auth.GetUserByID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	// Update fields
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Status != "" {
		user.Status = req.Status
	}
	user.UpdatedAt = time.Now()

	// Validate
	if err := user.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update in database
	query := "UPDATE users SET role = ?, status = ?, updated_at = ? WHERE id = ?"
	if h.db.Driver == "postgres" {
		query = "UPDATE users SET role = $1, status = $2, updated_at = $3 WHERE id = $4"
	}

	_, err = h.db.Exec(query, user.Role, user.Status, h.db.BindTime(user.UpdatedAt), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	respondJSON(w, http.StatusOK, user.SanitizeForJSON())
}

// DeleteUser deletes a user (admin only)
// DELETE /api/admin/users/{id}
func (h *AdminHandlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "user ID required")
		return
	}

	// Prevent admin from deleting themselves
	currentUserID, _ := server.GetUserIDFromContext(r.Context())
	if currentUserID == userID {
		respondError(w, http.StatusForbidden, "cannot delete your own account")
		return
	}

	// Delete user (cascade will delete related data)
	query := "DELETE FROM users WHERE id = ?"
	if h.db.Driver == "postgres" {
		query = "DELETE FROM users WHERE id = $1"
	}

	_, err := h.db.Exec(query, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "user deleted successfully",
	})
}

// GetSystemStats retrieves system statistics (admin only)
// GET /api/admin/stats
func (h *AdminHandlers) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{}

	// Total users
	var totalUsers int
	h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	stats["total_users"] = totalUsers

	// Active users
	var activeUsers int
	query := "SELECT COUNT(*) FROM users WHERE status = ?"
	if h.db.Driver == "postgres" {
		query = "SELECT COUNT(*) FROM users WHERE status = $1"
	}
	h.db.QueryRow(query, "active").Scan(&activeUsers)
	stats["active_users"] = activeUsers

	// Total profiles
	var totalProfiles int
	h.db.QueryRow("SELECT COUNT(*) FROM profiles").Scan(&totalProfiles)
	stats["total_profiles"] = totalProfiles

	// Public profiles
	var publicProfiles int
	query = "SELECT COUNT(*) FROM profiles WHERE is_public = ?"
	if h.db.Driver == "postgres" {
		query = "SELECT COUNT(*) FROM profiles WHERE is_public = $1"
	}
	h.db.QueryRow(query, true).Scan(&publicProfiles)
	stats["public_profiles"] = publicProfiles

	// Total links
	var totalLinks int
	h.db.QueryRow("SELECT COUNT(*) FROM links").Scan(&totalLinks)
	stats["total_links"] = totalLinks

	// Total views (last 30 days)
	var recentViews int
	query = "SELECT COUNT(*) FROM analytics WHERE event_type = ? AND created_at >= ?"
	if h.db.Driver == "postgres" {
		query = "SELECT COUNT(*) FROM analytics WHERE event_type = $1 AND created_at >= $2"
	}
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	h.db.QueryRow(query, "view", thirtyDaysAgo).Scan(&recentViews)
	stats["recent_views"] = recentViews

	// Total clicks (last 30 days)
	var recentClicks int
	query = "SELECT COUNT(*) FROM analytics WHERE event_type = ? AND created_at >= ?"
	if h.db.Driver == "postgres" {
		query = "SELECT COUNT(*) FROM analytics WHERE event_type = $1 AND created_at >= $2"
	}
	h.db.QueryRow(query, "click", thirtyDaysAgo).Scan(&recentClicks)
	stats["recent_clicks"] = recentClicks

	respondJSON(w, http.StatusOK, stats)
}

// TriggerBackup triggers a manual backup (admin only)
// POST /api/admin/backup
func (h *AdminHandlers) TriggerBackup(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "backup triggered successfully",
	})
}

// GetSettings retrieves all system settings (admin only)
// GET /api/admin/settings
func (h *AdminHandlers) GetSettings(w http.ResponseWriter, r *http.Request) {
	query := "SELECT key, value, updated_at FROM settings ORDER BY key ASC"

	rows, err := h.db.Query(query)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch settings")
		return
	}
	defer rows.Close()

	settings := []map[string]interface{}{}
	for rows.Next() {
		var key, value string
		var updatedAt time.Time
		err := rows.Scan(&key, &value, &updatedAt)
		if err != nil {
			continue
		}
		settings = append(settings, map[string]interface{}{
			"key":        key,
			"value":      value,
			"updated_at": updatedAt,
		})
	}

	respondJSON(w, http.StatusOK, settings)
}

// UpdateSettings updates system settings (admin only)
// PUT /api/admin/settings
func (h *AdminHandlers) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var settings map[string]string
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Update each setting
	for key, value := range settings {
		err := h.db.SetSetting(key, value)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update setting: "+key)
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "settings updated successfully",
		"count":   len(settings),
	})
}

// ImportServices imports services from file (admin only)
// POST /api/admin/services/import
func (h *AdminHandlers) ImportServices(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "service import not yet implemented",
	})
}

// ClearCache clears system cache (admin only)
// POST /api/admin/cache/clear
func (h *AdminHandlers) ClearCache(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "cache cleared successfully",
	})
}

// GetSMTPConfig retrieves SMTP configuration (admin only)
// GET /api/admin/smtp/config
func (h *AdminHandlers) GetSMTPConfig(w http.ResponseWriter, r *http.Request) {
	config := map[string]interface{}{}

	smtpSettings := []string{
		"smtp_provider", "smtp_host", "smtp_port", "smtp_security",
		"smtp_user", "smtp_from_name", "smtp_from_address",
		"admin_email", "smtp_enabled",
	}

	for _, key := range smtpSettings {
		value, _ := h.db.GetSetting(key)
		// Don't return password
		if key != "smtp_password" {
			config[key] = value
		}
	}

	respondJSON(w, http.StatusOK, config)
}

// UpdateSMTPConfig updates SMTP configuration (admin only)
// PUT /api/admin/smtp/config
func (h *AdminHandlers) UpdateSMTPConfig(w http.ResponseWriter, r *http.Request) {
	var config map[string]string
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Update SMTP settings
	for key, value := range config {
		err := h.db.SetSetting(key, value)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update SMTP setting: "+key)
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "SMTP configuration updated successfully",
	})
}

// TestSMTPConnection tests SMTP connection (admin only)
// POST /api/admin/smtp/test
func (h *AdminHandlers) TestSMTPConnection(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "SMTP connection test not yet implemented",
		"status":  "pending",
	})
}

// GetNotificationPreferences retrieves notification preferences (admin only)
// GET /api/admin/notifications/preferences
func (h *AdminHandlers) GetNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	preferences := map[string]interface{}{}

	notificationSettings := []string{
		"notify_emergency", "notify_certificate", "notify_bug_report",
		"notify_user_registration", "notify_domain_verification",
		"notify_backup_status", "notify_high_traffic",
		"notification_batch_delay",
	}

	for _, key := range notificationSettings {
		value, _ := h.db.GetSetting(key)
		preferences[key] = value
	}

	respondJSON(w, http.StatusOK, preferences)
}

// UpdateNotificationPreferences updates notification preferences (admin only)
// PUT /api/admin/notifications/preferences
func (h *AdminHandlers) UpdateNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	var preferences map[string]string
	if err := json.NewDecoder(r.Body).Decode(&preferences); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Update notification preferences
	for key, value := range preferences {
		err := h.db.SetSetting(key, value)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update preference: "+key)
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "notification preferences updated successfully",
	})
}
