package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// LinkHandler handles link-related HTTP requests
type LinkHandler struct {
	config *config.Config
	db     *store.DB
}

// NewLinkHandler creates a new link handler
func NewLinkHandler(cfg *config.Config, db *store.DB) *LinkHandler {
	return &LinkHandler{
		config: cfg,
		db:     db,
	}
}

// CreateLinkRequest represents a link creation request
type CreateLinkRequest struct {
	ProfileID string `json:"profile_id"`
	Service   string `json:"service"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	Order     int    `json:"order"`
}

// UpdateLinkRequest represents a link update request
type UpdateLinkRequest struct {
	Service string `json:"service"`
	URL     string `json:"url"`
	Title   string `json:"title"`
	Order   int    `json:"order"`
	Enabled bool   `json:"enabled"`
}

// HandleCreateLink creates a new link
func (h *LinkHandler) HandleCreateLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Get user ID from session
	// TODO: Verify user owns the profile

	// Parse request
	var req CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Validate URL
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		h.renderError(w, http.StatusBadRequest, "URL must start with http:// or https://")
		return
	}

	// Check max links per profile
	// TODO: Get profile's link count
	linkCount := 0
	if linkCount >= h.config.Cassocial.MaxLinksPerProfile {
		h.renderError(w, http.StatusForbidden, "Maximum links per profile reached")
		return
	}

	// Create link
	link := &store.Link{
		ProfileID: req.ProfileID,
		Service:   req.Service,
		URL:       req.URL,
		Title:     req.Title,
		Order:     req.Order,
		Enabled:   true,
	}

	// TODO: Save to database
	_ = link

	h.renderJSON(w, http.StatusCreated, map[string]interface{}{
		"status":  "success",
		"message": "Link created successfully",
		"link_id": "temp-link-id",
	})
}

// HandleGetLinks retrieves all links for a profile
func (h *LinkHandler) HandleGetLinks(w http.ResponseWriter, r *http.Request) {
	profileID := r.URL.Query().Get("profile_id")
	if profileID == "" {
		h.renderError(w, http.StatusBadRequest, "Profile ID required")
		return
	}

	// TODO: Get links from database
	links := []interface{}{}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"links": links,
		"total": len(links),
	})
}

// HandleUpdateLink updates a link
func (h *LinkHandler) HandleUpdateLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Get user ID from session
	// TODO: Get link ID from URL
	// TODO: Verify user owns the profile

	// Parse request
	var req UpdateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// TODO: Update link in database

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Link updated successfully",
	})
}

// HandleDeleteLink deletes a link
func (h *LinkHandler) HandleDeleteLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Get user ID from session
	// TODO: Get link ID from URL
	// TODO: Verify user owns the profile
	// TODO: Delete link from database

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Link deleted successfully",
	})
}

// HandleReorderLinks updates link order
func (h *LinkHandler) HandleReorderLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse new order
	var req struct {
		ProfileID string   `json:"profile_id"`
		LinkIDs   []string `json:"link_ids"` // Array of link IDs in new order
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// TODO: Verify user owns profile
	// TODO: Update link order in database

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Links reordered successfully",
	})
}

// HandleLinkClick tracks a link click
func (h *LinkHandler) HandleLinkClick(w http.ResponseWriter, r *http.Request) {
	linkID := strings.TrimPrefix(r.URL.Path, "/click/")
	if linkID == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Get client info for analytics
	ip := getClientIP(r)
	userAgent := r.UserAgent()
	referer := r.Referer()

	// TODO: Record click in analytics
	_ = ip
	_ = userAgent
	_ = referer

	// TODO: Get link URL from database
	targetURL := "https://example.com"

	// Redirect to target URL
	http.Redirect(w, r, targetURL, http.StatusTemporaryRedirect)
}

// HandleGetServices returns the list of supported services
func (h *LinkHandler) HandleGetServices(w http.ResponseWriter, r *http.Request) {
	// TODO: Load services from database or embedded JSON
	// For now, return a sample list
	services := []map[string]string{
		{"id": "twitter", "name": "Twitter", "icon": "twitter.svg"},
		{"id": "github", "name": "GitHub", "icon": "github.svg"},
		{"id": "linkedin", "name": "LinkedIn", "icon": "linkedin.svg"},
		{"id": "instagram", "name": "Instagram", "icon": "instagram.svg"},
		{"id": "youtube", "name": "YouTube", "icon": "youtube.svg"},
		{"id": "website", "name": "Website", "icon": "globe.svg"},
	}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"services": services,
		"total":    len(services),
	})
}

// HandleSearchServices searches for services
func (h *LinkHandler) HandleSearchServices(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.renderError(w, http.StatusBadRequest, "Search query required")
		return
	}

	// TODO: Search services database (5000+ services)
	results := []interface{}{}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
		"query":   query,
		"total":   len(results),
	})
}

// Helper functions

// getClientIP extracts the real client IP
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the list
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Use RemoteAddr as fallback
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

// renderJSON renders a JSON response
func (h *LinkHandler) renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// renderError renders an error response
func (h *LinkHandler) renderError(w http.ResponseWriter, status int, message string) {
	h.renderJSON(w, status, map[string]string{
		"error": message,
	})
}
