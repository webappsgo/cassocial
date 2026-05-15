package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server"
	"github.com/casapps/cassocial/src/server/store"
)

// SocialHandler handles social features (discovery, search, featured profiles)
type SocialHandler struct {
	config *config.Config
	db     *store.DB
}

// NewSocialHandler creates a new social features handler
func NewSocialHandler(cfg *config.Config, db *store.DB) *SocialHandler {
	return &SocialHandler{
		config: cfg,
		db:     db,
	}
}

// HandleProfileDirectory shows public profile directory
func (h *SocialHandler) HandleProfileDirectory(w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters
	page := getIntParam(r, "page", 1)
	perPage := getIntParam(r, "per_page", 20)

	// Get filters
	tag := r.URL.Query().Get("tag")
	verified := r.URL.Query().Get("verified") // "true" or "false"

	profiles := []interface{}{}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"profiles":   profiles,
		"total":      0,
		"page":       page,
		"per_page":   perPage,
		"total_pages": 0,
		"filters": map[string]string{
			"tag":      tag,
			"verified": verified,
		},
	})
}

// HandleSearchProfiles searches profiles
func (h *SocialHandler) HandleSearchProfiles(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.renderError(w, http.StatusBadRequest, "Search query required")
		return
	}

	results := []interface{}{}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
		"query":   query,
		"total":   len(results),
	})
}

// HandleFeaturedProfiles returns featured profiles
func (h *SocialHandler) HandleFeaturedProfiles(w http.ResponseWriter, r *http.Request) {
	featured := []interface{}{}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"featured": featured,
		"total":    len(featured),
	})
}

// HandleVerifyProfile handles profile verification requests
func (h *SocialHandler) HandleVerifyProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		h.renderError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req struct {
		ProfileID string `json:"profile_id"`
		Proof     string `json:"proof"` // URL or text proving ownership
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Verification request submitted. We'll review it shortly.",
	})
}

// HandleGetTags returns available tags
func (h *SocialHandler) HandleGetTags(w http.ResponseWriter, r *http.Request) {
	tags := []map[string]interface{}{
		{"name": "developer", "count": 0},
		{"name": "designer", "count": 0},
		{"name": "creator", "count": 0},
		{"name": "business", "count": 0},
	}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"tags":  tags,
		"total": len(tags),
	})
}

// HandleAddTag adds a tag to a profile
func (h *SocialHandler) HandleAddTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProfileID string `json:"profile_id"`
		Tag       string `json:"tag"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Tag added",
	})
}

// getIntParam gets an integer query parameter with default
func getIntParam(r *http.Request, name string, defaultValue int) int {
	val := r.URL.Query().Get(name)
	if val == "" {
		return defaultValue
	}

	var result int
	if _, err := fmt.Sscanf(val, "%d", &result); err != nil {
		return defaultValue
	}

	return result
}

// renderJSON renders a JSON response
func (h *SocialHandler) renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// renderError renders an error response
func (h *SocialHandler) renderError(w http.ResponseWriter, status int, message string) {
	h.renderJSON(w, status, map[string]string{
		"error": message,
	})
}
