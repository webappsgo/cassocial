package handler

import (
	"encoding/json"
	"net/http"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// DashboardHandler handles user dashboard operations
type DashboardHandler struct {
	config *config.Config
	db     *store.DB
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(cfg *config.Config, db *store.DB) *DashboardHandler {
	return &DashboardHandler{
		config: cfg,
		db:     db,
	}
}

// HandleDashboard serves the main user dashboard
func (h *DashboardHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	// TODO: Get user ID from session
	userID := "temp-user-id"

	// TODO: Get user's profiles from database
	// TODO: Get analytics summary

	dashboard := map[string]interface{}{
		"user": map[string]interface{}{
			"id":       userID,
			"username": "user",
			"email":    "user@example.com",
		},
		"stats": map[string]interface{}{
			"total_profiles": 0,
			"total_links":    0,
			"total_views":    0,
			"total_clicks":   0,
		},
		"recent_profiles": []interface{}{},
		"recent_activity": []interface{}{},
	}

	h.renderJSON(w, http.StatusOK, dashboard)
}

// HandleProfileList shows user's profile list
func (h *DashboardHandler) HandleProfileList(w http.ResponseWriter, r *http.Request) {
	// TODO: Get user ID from session
	// TODO: Get user's profiles from database

	profiles := []interface{}{}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"profiles": profiles,
		"total":    len(profiles),
		"limit":    h.config.Cassocial.MaxProfilesPerUser,
	})
}

// HandleProfileCreate shows profile creation form
func (h *DashboardHandler) HandleProfileCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := map[string]interface{}{
			"themes":    []string{"dark", "light"},
			"max_links": h.config.Cassocial.MaxLinksPerProfile,
		}
		h.renderJSON(w, http.StatusOK, data)
		return
	}

	// POST handled by ProfileHandler.HandleCreateProfile
	http.Error(w, "Use /api/v1/profiles", http.StatusSeeOther)
}

// HandleProfileEdit shows profile editing form
func (h *DashboardHandler) HandleProfileEdit(w http.ResponseWriter, r *http.Request) {
	profileID := r.URL.Query().Get("id")
	if profileID == "" {
		h.renderError(w, http.StatusBadRequest, "Profile ID required")
		return
	}

	// TODO: Get profile from database
	// TODO: Verify user owns profile

	profile := map[string]interface{}{
		"id":          profileID,
		"slug":        "example",
		"title":       "Profile Title",
		"description": "Description",
		"theme":       "dark",
		"public":      true,
	}

	h.renderJSON(w, http.StatusOK, profile)
}

// HandleAnalyticsOverview shows analytics overview
func (h *DashboardHandler) HandleAnalyticsOverview(w http.ResponseWriter, r *http.Request) {
	// TODO: Get user ID from session
	// TODO: Get aggregated analytics for all user's profiles

	analytics := map[string]interface{}{
		"total_views":      0,
		"total_clicks":     0,
		"unique_visitors":  0,
		"top_profiles":     []interface{}{},
		"views_chart":      []interface{}{},
		"clicks_chart":     []interface{}{},
		"geographic_data":  []interface{}{},
	}

	h.renderJSON(w, http.StatusOK, analytics)
}

// HandleAccountSettings shows account settings page
func (h *DashboardHandler) HandleAccountSettings(w http.ResponseWriter, r *http.Request) {
	// TODO: Get user ID from session
	// TODO: Get user info from database

	settings := map[string]interface{}{
		"username":           "user",
		"email":              "user@example.com",
		"email_verified":     true,
		"two_factor_enabled": false,
		"created_at":         "2025-01-01",
	}

	h.renderJSON(w, http.StatusOK, settings)
}

// HandleNotifications shows user notifications
func (h *DashboardHandler) HandleNotifications(w http.ResponseWriter, r *http.Request) {
	// TODO: Get user ID from session
	// TODO: Get notifications from database

	notifications := []interface{}{}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": notifications,
		"unread":        0,
	})
}

// HandleRecentActivity shows recent activity
func (h *DashboardHandler) HandleRecentActivity(w http.ResponseWriter, r *http.Request) {
	// TODO: Get user ID from session
	// TODO: Get recent activity (views, clicks, profile changes)

	activity := []interface{}{}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"activity": activity,
		"total":    len(activity),
	})
}

// renderJSON renders a JSON response
func (h *DashboardHandler) renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// renderError renders an error response
func (h *DashboardHandler) renderError(w http.ResponseWriter, status int, message string) {
	h.renderJSON(w, status, map[string]string{
		"error": message,
	})
}
