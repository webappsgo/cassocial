package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// ShortlinkHandler handles shortlink operations
type ShortlinkHandler struct {
	config *config.Config
	db     *store.DB
}

// NewShortlinkHandler creates a new shortlink handler
func NewShortlinkHandler(cfg *config.Config, db *store.DB) *ShortlinkHandler {
	return &ShortlinkHandler{
		config: cfg,
		db:     db,
	}
}

// Shortlink represents a short URL
type Shortlink struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	UserID    string    `json:"user_id"`
	Clicks    int       `json:"clicks"`
	ExpiresAt string    `json:"expires_at,omitempty"`
	CreatedAt string    `json:"created_at"`
}

// CreateShortlinkRequest represents a shortlink creation request
type CreateShortlinkRequest struct {
	URL        string `json:"url"`
	CustomCode string `json:"custom_code,omitempty"` // Optional custom short code
	ExpiresIn  int    `json:"expires_in,omitempty"`  // Optional expiry in hours
}

// HandleCreateShortlink creates a new shortlink
func (h *ShortlinkHandler) HandleCreateShortlink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Get user ID from session (or allow anonymous with rate limiting)
	userID := "temp-user-id"

	var req CreateShortlinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Validate URL
	if req.URL == "" {
		h.renderError(w, http.StatusBadRequest, "URL is required")
		return
	}

	// Generate or validate custom code
	code := req.CustomCode
	if code == "" {
		// Generate random code
		code = h.generateShortCode()
	} else {
		// Validate custom code
		if !isValidShortCode(code) {
			h.renderError(w, http.StatusBadRequest, "Invalid custom code. Use 3-20 alphanumeric characters.")
			return
		}

		// TODO: Check if code already exists
	}

	// Calculate expiry
	var expiresAt string
	if req.ExpiresIn > 0 {
		expiry := time.Now().Add(time.Duration(req.ExpiresIn) * time.Hour)
		expiresAt = expiry.Format(time.RFC3339)
	}

	// Create shortlink
	shortlink := &Shortlink{
		Code:      code,
		URL:       req.URL,
		UserID:    userID,
		Clicks:    0,
		ExpiresAt: expiresAt,
	}

	// TODO: Save to database
	_ = shortlink

	// Build short URL
	shortURL := fmt.Sprintf("https://%s/s/%s", r.Host, code)

	h.renderJSON(w, http.StatusCreated, map[string]interface{}{
		"status":    "success",
		"message":   "Shortlink created successfully",
		"code":      code,
		"short_url": shortURL,
		"target_url": req.URL,
		"expires_at": expiresAt,
	})
}

// HandleRedirectShortlink redirects a shortlink and tracks click
func (h *ShortlinkHandler) HandleRedirectShortlink(w http.ResponseWriter, r *http.Request) {
	// Extract code from URL
	code := r.URL.Path[len("/s/"):]
	if code == "" {
		http.NotFound(w, r)
		return
	}

	// TODO: Get shortlink from database
	// TODO: Check if expired
	// TODO: Track click with IP, user agent, referer

	// For now, redirect to a placeholder
	targetURL := "https://github.com/casapps/cassocial"

	// Increment click count
	// TODO: Update database

	http.Redirect(w, r, targetURL, http.StatusTemporaryRedirect)
}

// HandleGetShortlink returns shortlink information
func (h *ShortlinkHandler) HandleGetShortlink(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		h.renderError(w, http.StatusBadRequest, "Shortlink code required")
		return
	}

	// TODO: Get shortlink from database
	shortlink := map[string]interface{}{
		"code":       code,
		"url":        "https://example.com",
		"clicks":     0,
		"created_at": time.Now().Format(time.RFC3339),
	}

	h.renderJSON(w, http.StatusOK, shortlink)
}

// HandleListShortlinks lists user's shortlinks
func (h *ShortlinkHandler) HandleListShortlinks(w http.ResponseWriter, r *http.Request) {
	// TODO: Get user ID from session
	// TODO: Get user's shortlinks from database

	shortlinks := []interface{}{}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"shortlinks": shortlinks,
		"total":      len(shortlinks),
	})
}

// HandleDeleteShortlink deletes a shortlink
func (h *ShortlinkHandler) HandleDeleteShortlink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		h.renderError(w, http.StatusBadRequest, "Shortlink code required")
		return
	}

	// TODO: Get user ID from session
	// TODO: Verify user owns shortlink
	// TODO: Delete shortlink

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Shortlink deleted",
	})
}

// HandleShortlinkAnalytics returns analytics for a shortlink
func (h *ShortlinkHandler) HandleShortlinkAnalytics(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		h.renderError(w, http.StatusBadRequest, "Shortlink code required")
		return
	}

	// TODO: Get user ID from session
	// TODO: Verify user owns shortlink
	// TODO: Get click analytics

	analytics := map[string]interface{}{
		"code":         code,
		"total_clicks": 0,
		"clicks_by_day": []interface{}{},
		"top_referers": []interface{}{},
		"top_countries": []interface{}{},
	}

	h.renderJSON(w, http.StatusOK, analytics)
}

// HandleShortlinkQR generates QR code for a shortlink
func (h *ShortlinkHandler) HandleShortlinkQR(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		h.renderError(w, http.StatusBadRequest, "Shortlink code required")
		return
	}

	// Build short URL
	shortURL := fmt.Sprintf("https://%s/s/%s", r.Host, code)

	// Redirect to QR handler
	http.Redirect(w, r, fmt.Sprintf("/api/v1/qr?url=%s", shortURL), http.StatusTemporaryRedirect)
}

// Helper functions

// generateShortCode generates a random short code
func (h *ShortlinkHandler) generateShortCode() string {
	// Generate 6-character alphanumeric code
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6

	b := make([]byte, length)
	rand.Read(b)

	code := make([]byte, length)
	for i := 0; i < length; i++ {
		code[i] = chars[int(b[i])%len(chars)]
	}

	return string(code)
}

// isValidShortCode validates a custom short code
func isValidShortCode(code string) bool {
	if len(code) < 3 || len(code) > 20 {
		return false
	}

	// Only alphanumeric characters
	for _, c := range code {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}

	return true
}

// renderJSON renders a JSON response
func (h *ShortlinkHandler) renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// renderError renders an error response
func (h *ShortlinkHandler) renderError(w http.ResponseWriter, status int, message string) {
	h.renderJSON(w, status, map[string]string{
		"error": message,
	})
}
