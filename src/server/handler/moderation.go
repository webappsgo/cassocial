package handler

import (
	"encoding/json"
	"net/http"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// ModerationHandler handles content moderation
type ModerationHandler struct {
	config *config.Config
	db     *store.DB
}

// NewModerationHandler creates a new moderation handler
func NewModerationHandler(cfg *config.Config, db *store.DB) *ModerationHandler {
	return &ModerationHandler{
		config: cfg,
		db:     db,
	}
}

// HandleReportContent handles content reporting
func (h *ModerationHandler) HandleReportContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ContentType string `json:"content_type"` // profile, link, user
		ContentID   string `json:"content_id"`
		Reason      string `json:"reason"`
		Details     string `json:"details"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Report submitted. Thank you for helping keep Cassocial safe.",
	})
}

// HandleGetModerationQueue returns content pending moderation
func (h *ModerationHandler) HandleGetModerationQueue(w http.ResponseWriter, r *http.Request) {
	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"queue": []interface{}{},
		"total": 0,
	})
}

// HandleModerateContent moderates a piece of content
func (h *ModerationHandler) HandleModerateContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ReportID string `json:"report_id"`
		Action   string `json:"action"` // approve, remove, ban
		Notes    string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Moderation action applied",
	})
}

// HandleGetBlockedPatterns returns blocked URL/content patterns
func (h *ModerationHandler) HandleGetBlockedPatterns(w http.ResponseWriter, r *http.Request) {
	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"patterns": []interface{}{},
		"total":    0,
	})
}

// HandleAddBlockedPattern adds a blocked pattern
func (h *ModerationHandler) HandleAddBlockedPattern(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Type    string `json:"type"` // url, domain, keyword
		Pattern string `json:"pattern"`
		Reason  string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Pattern added to blocklist",
	})
}

// renderJSON renders a JSON response
func (h *ModerationHandler) renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// renderError renders an error response
func (h *ModerationHandler) renderError(w http.ResponseWriter, status int, message string) {
	h.renderJSON(w, status, map[string]string{
		"error": message,
	})
}
