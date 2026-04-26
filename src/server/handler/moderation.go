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

	// TODO: Create report in database
	// TODO: Notify moderators

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Report submitted. Thank you for helping keep Cassocial safe.",
	})
}

// HandleGetModerationQueue returns content pending moderation
func (h *ModerationHandler) HandleGetModerationQueue(w http.ResponseWriter, r *http.Request) {
	// TODO: Require admin/moderator role
	// TODO: Get pending reports from database

	queue := []interface{}{}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"queue": queue,
		"total": len(queue),
	})
}

// HandleModerateContent moderates a piece of content
func (h *ModerationHandler) HandleModerateContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Require admin/moderator role

	var req struct {
		ReportID string `json:"report_id"`
		Action   string `json:"action"` // approve, remove, ban
		Notes    string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// TODO: Process moderation action
	// TODO: Update report status
	// TODO: Notify content owner if needed

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Moderation action applied",
	})
}

// HandleGetBlockedPatterns returns blocked URL/content patterns
func (h *ModerationHandler) HandleGetBlockedPatterns(w http.ResponseWriter, r *http.Request) {
	// TODO: Require admin role
	// TODO: Get blocked patterns from database

	patterns := []interface{}{}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"patterns": patterns,
		"total":    len(patterns),
	})
}

// HandleAddBlockedPattern adds a blocked pattern
func (h *ModerationHandler) HandleAddBlockedPattern(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Require admin role

	var req struct {
		Type    string `json:"type"` // url, domain, keyword
		Pattern string `json:"pattern"`
		Reason  string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// TODO: Add pattern to database
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
