package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// ProfileHandler handles profile-related HTTP requests
type ProfileHandler struct {
	config *config.Config
	db     *store.DB
}

// NewProfileHandler creates a new profile handler
func NewProfileHandler(cfg *config.Config, db *store.DB) *ProfileHandler {
	return &ProfileHandler{
		config: cfg,
		db:     db,
	}
}

// CreateProfileRequest represents a profile creation request
type CreateProfileRequest struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Theme       string `json:"theme"`
	Public      bool   `json:"public"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Theme       string `json:"theme"`
	CustomCSS   string `json:"custom_css"`
	Public      bool   `json:"public"`
	Password    string `json:"password"` // For password-protected profiles
}

// HandleCreateProfile creates a new profile
func (h *ProfileHandler) HandleCreateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Get user ID from session
	userID := "temp-user-id"

	// Parse request
	var req CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Validate slug
	if !isValidSlug(req.Slug) {
		h.renderError(w, http.StatusBadRequest, "Invalid slug. Use only lowercase letters, numbers, and hyphens.")
		return
	}

	// Check max profiles per user
	// TODO: Get user's profile count
	profileCount := 0
	if profileCount >= h.config.Cassocial.MaxProfilesPerUser {
		h.renderError(w, http.StatusForbidden, fmt.Sprintf("Maximum %d profiles per user", h.config.Cassocial.MaxProfilesPerUser))
		return
	}

	// Check if slug is available
	// TODO: Check database for existing slug

	// Create profile
	profile := &store.Profile{
		UserID:      userID,
		Slug:        req.Slug,
		Title:       req.Title,
		Description: req.Description,
		Theme:       req.Theme,
		Public:      req.Public,
	}

	// TODO: Save to database
	_ = profile

	h.renderJSON(w, http.StatusCreated, map[string]interface{}{
		"status":  "success",
		"message": "Profile created successfully",
		"profile": map[string]string{
			"id":   "temp-profile-id",
			"slug": req.Slug,
			"url":  fmt.Sprintf("/%s", req.Slug),
		},
	})
}

// HandleGetProfile retrieves a profile
func (h *ProfileHandler) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	// Extract slug from URL
	slug := strings.TrimPrefix(r.URL.Path, "/api/v1/profiles/")
	if slug == "" {
		h.renderError(w, http.StatusBadRequest, "Profile slug required")
		return
	}

	// TODO: Get profile from database
	profile := map[string]interface{}{
		"id":          "temp-id",
		"slug":        slug,
		"title":       "Profile Title",
		"description": "Profile description",
		"theme":       "dark",
		"public":      true,
		"links":       []interface{}{},
	}

	h.renderJSON(w, http.StatusOK, profile)
}

// HandleUpdateProfile updates a profile
func (h *ProfileHandler) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Get user ID from session
	// TODO: Verify user owns this profile

	// Parse request
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// TODO: Update profile in database

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Profile updated successfully",
	})
}

// HandleDeleteProfile deletes a profile
func (h *ProfileHandler) HandleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Get user ID from session
	// TODO: Verify user owns this profile
	// TODO: Delete profile and all associated links

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Profile deleted successfully",
	})
}

// HandleListProfiles lists user's profiles
func (h *ProfileHandler) HandleListProfiles(w http.ResponseWriter, r *http.Request) {
	// TODO: Get user ID from session
	// TODO: Get user's profiles from database

	profiles := []interface{}{}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"profiles": profiles,
		"total":    len(profiles),
	})
}

// HandlePublicProfile serves a public profile page
func (h *ProfileHandler) HandlePublicProfile(w http.ResponseWriter, r *http.Request) {
	// Extract slug from URL
	slug := strings.TrimPrefix(r.URL.Path, "/")
	if slug == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// TODO: Get profile from database
	// TODO: Track page view
	// TODO: Render profile template with links

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"slug":  slug,
		"title": "Profile",
		"links": []interface{}{},
	})
}

// Helper functions

// isValidSlug validates a profile slug
func isValidSlug(slug string) bool {
	if len(slug) < 3 || len(slug) > 50 {
		return false
	}

	// Only lowercase letters, numbers, and hyphens
	for _, c := range slug {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}

	// Cannot start or end with hyphen
	if slug[0] == '-' || slug[len(slug)-1] == '-' {
		return false
	}

	return true
}

// renderJSON renders a JSON response
func (h *ProfileHandler) renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// renderError renders an error response
func (h *ProfileHandler) renderError(w http.ResponseWriter, status int, message string) {
	h.renderJSON(w, status, map[string]string{
		"error": message,
	})
}
