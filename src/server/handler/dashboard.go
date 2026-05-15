package handler

import (
	"encoding/json"
	"net/http"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server"
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
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		h.renderError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	dashboard := map[string]interface{}{
		"user": map[string]interface{}{
			"id": userID,
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
	_, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		h.renderError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"profiles": []interface{}{},
		"total":    0,
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
	_, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		h.renderError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

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
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		h.renderError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	settings := map[string]interface{}{
		"username":           user.Username,
		"email":              user.Email,
		"email_verified":     user.EmailVerified,
		"two_factor_enabled": user.TwoFactorEnabled,
		"created_at":         user.CreatedAt.Format("2006-01-02"),
	}

	h.renderJSON(w, http.StatusOK, settings)
}

// HandleNotifications shows user notifications
func (h *DashboardHandler) HandleNotifications(w http.ResponseWriter, r *http.Request) {
	_, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		h.renderError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": []interface{}{},
		"unread":        0,
	})
}

// HandleRecentActivity shows recent activity
func (h *DashboardHandler) HandleRecentActivity(w http.ResponseWriter, r *http.Request) {
	_, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		h.renderError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"activity": []interface{}{},
		"total":    0,
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
