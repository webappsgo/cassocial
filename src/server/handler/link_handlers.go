package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/casapps/cassocial/src/server"
	"github.com/casapps/cassocial/src/server/store"
	"github.com/casapps/cassocial/src/server/model"
)

// LinkHandlers handles link-related HTTP requests
type LinkHandlers struct {
	db *store.DB
}

// NewLinkHandlers creates a new LinkHandlers instance
func NewLinkHandlers(db *store.DB) *LinkHandlers {
	return &LinkHandlers{db: db}
}

// CreateLinkRequest represents a link creation request
type CreateLinkRequest struct {
	ServiceID       string `json:"service_id,omitempty"`
	Title           string `json:"title"`
	Username        string `json:"username,omitempty"`
	URL             string `json:"url"`
	IconURL         string `json:"icon_url,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	TextColor       string `json:"text_color,omitempty"`
}

// UpdateLinkRequest represents a link update request
type UpdateLinkRequest struct {
	Title           string `json:"title,omitempty"`
	Username        string `json:"username,omitempty"`
	URL             string `json:"url,omitempty"`
	IconURL         string `json:"icon_url,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	TextColor       string `json:"text_color,omitempty"`
	IsActive        *bool  `json:"is_active,omitempty"`
}

// ReorderLinksRequest represents a link reordering request
type ReorderLinksRequest struct {
	LinkIDs []string `json:"link_ids"`
}

// ListLinks lists all links for a profile
// GET /api/profiles/{id}/links
func (h *LinkHandlers) ListLinks(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	profileID := r.PathValue("id")
	if profileID == "" {
		respondError(w, http.StatusBadRequest, "profile ID required")
		return
	}

	// Verify profile ownership
	if !h.userOwnsProfile(userID, profileID) {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	query := `SELECT id, profile_id, service_id, title, username, url, icon_url,
			  background_color, text_color, position, is_active, click_count,
			  created_at, updated_at
			  FROM links WHERE profile_id = ? ORDER BY position ASC`

	if h.db.Driver == "postgres" {
		query = replaceQuestionMarks(query, 1)
	}

	rows, err := h.db.Query(query, profileID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch links")
		return
	}
	defer rows.Close()

	links := []model.Link{}
	for rows.Next() {
		var l model.Link
		err := rows.Scan(&l.ID, &l.ProfileID, &l.ServiceID, &l.Title, &l.Username, &l.URL,
			&l.IconURL, &l.BackgroundColor, &l.TextColor, &l.Position, &l.IsActive,
			&l.ClickCount, &l.CreatedAt, &l.UpdatedAt)
		if err != nil {
			continue
		}
		links = append(links, l)
	}

	respondJSON(w, http.StatusOK, links)
}

// CreateLink creates a new link
// POST /api/profiles/{id}/links
func (h *LinkHandlers) CreateLink(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	profileID := r.PathValue("id")
	if profileID == "" {
		respondError(w, http.StatusBadRequest, "profile ID required")
		return
	}

	// Verify profile ownership
	if !h.userOwnsProfile(userID, profileID) {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	var req CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Check max links per profile
	maxLinksStr, _ := h.db.GetSetting("max_links_per_profile")
	maxLinks, _ := strconv.Atoi(maxLinksStr)
	if maxLinks == 0 {
		maxLinks = 100
	}

	currentCount, _ := h.getLinkCount(profileID)
	if currentCount >= maxLinks {
		respondError(w, http.StatusForbidden, "maximum number of links reached")
		return
	}

	// Get next position
	nextPosition := h.getNextLinkPosition(profileID)

	// Create link
	link := &model.Link{
		ProfileID:       profileID,
		ServiceID:       req.ServiceID,
		Title:           req.Title,
		Username:        req.Username,
		URL:             req.URL,
		IconURL:         req.IconURL,
		BackgroundColor: req.BackgroundColor,
		TextColor:       req.TextColor,
		Position:        nextPosition,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Validate link
	if err := link.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Insert link
	query := `INSERT INTO links (id, profile_id, service_id, title, username, url, icon_url,
			  background_color, text_color, position, is_active, click_count, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	if h.db.Driver == "postgres" {
		query = `INSERT INTO links (profile_id, service_id, title, username, url, icon_url,
				 background_color, text_color, position, is_active, click_count, created_at, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
				 RETURNING id`
		err := h.db.QueryRow(query, link.ProfileID, link.ServiceID, link.Title, link.Username,
			link.URL, link.IconURL, link.BackgroundColor, link.TextColor, link.Position,
			link.IsActive, 0, link.CreatedAt, link.UpdatedAt).Scan(&link.ID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create link")
			return
		}
	} else {
		link.ID = generateUUID()
		_, err := h.db.Exec(query, link.ID, link.ProfileID, link.ServiceID, link.Title,
			link.Username, link.URL, link.IconURL, link.BackgroundColor, link.TextColor,
			link.Position, link.IsActive, 0, link.CreatedAt, link.UpdatedAt)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create link")
			return
		}
	}

	respondJSON(w, http.StatusCreated, link)
}

// UpdateLink updates an existing link
// PUT /api/links/{id}
func (h *LinkHandlers) UpdateLink(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	linkID := r.PathValue("id")
	if linkID == "" {
		respondError(w, http.StatusBadRequest, "link ID required")
		return
	}

	// Get existing link
	link, err := h.getLinkByID(linkID)
	if err != nil {
		respondError(w, http.StatusNotFound, "link not found")
		return
	}

	// Verify profile ownership
	if !h.userOwnsProfile(userID, link.ProfileID) {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	var req UpdateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Update fields
	if req.Title != "" {
		link.Title = req.Title
	}
	if req.Username != "" {
		link.Username = req.Username
	}
	if req.URL != "" {
		link.URL = req.URL
	}
	if req.IconURL != "" {
		link.IconURL = req.IconURL
	}
	if req.BackgroundColor != "" {
		link.BackgroundColor = req.BackgroundColor
	}
	if req.TextColor != "" {
		link.TextColor = req.TextColor
	}
	if req.IsActive != nil {
		link.IsActive = *req.IsActive
	}

	link.UpdatedAt = time.Now()

	// Validate link
	if err := link.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update in database
	query := `UPDATE links SET title = ?, username = ?, url = ?, icon_url = ?,
			  background_color = ?, text_color = ?, is_active = ?, updated_at = ?
			  WHERE id = ?`

	if h.db.Driver == "postgres" {
		query = replaceQuestionMarks(query, 9)
	}

	_, err = h.db.Exec(query, link.Title, link.Username, link.URL, link.IconURL,
		link.BackgroundColor, link.TextColor, link.IsActive, link.UpdatedAt, linkID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update link")
		return
	}

	respondJSON(w, http.StatusOK, link)
}

// DeleteLink deletes a link
// DELETE /api/links/{id}
func (h *LinkHandlers) DeleteLink(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	linkID := r.PathValue("id")
	if linkID == "" {
		respondError(w, http.StatusBadRequest, "link ID required")
		return
	}

	// Get link
	link, err := h.getLinkByID(linkID)
	if err != nil {
		respondError(w, http.StatusNotFound, "link not found")
		return
	}

	// Verify profile ownership
	if !h.userOwnsProfile(userID, link.ProfileID) {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	// Delete link
	query := "DELETE FROM links WHERE id = ?"
	if h.db.Driver == "postgres" {
		query = "DELETE FROM links WHERE id = $1"
	}

	_, err = h.db.Exec(query, linkID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete link")
		return
	}

	// Reorder remaining links
	h.reorderLinks(link.ProfileID)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "link deleted successfully",
	})
}

// ReorderLinks reorders links for a profile
// POST /api/links/reorder
func (h *LinkHandlers) ReorderLinks(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req ReorderLinksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.LinkIDs) == 0 {
		respondError(w, http.StatusBadRequest, "link IDs required")
		return
	}

	// Get first link to determine profile
	firstLink, err := h.getLinkByID(req.LinkIDs[0])
	if err != nil {
		respondError(w, http.StatusNotFound, "link not found")
		return
	}

	// Verify profile ownership
	if !h.userOwnsProfile(userID, firstLink.ProfileID) {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	// Update positions
	query := "UPDATE links SET position = ? WHERE id = ?"
	if h.db.Driver == "postgres" {
		query = "UPDATE links SET position = $1 WHERE id = $2"
	}

	for i, linkID := range req.LinkIDs {
		_, err := h.db.Exec(query, i, linkID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to reorder links")
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "links reordered successfully",
	})
}

// ToggleLink toggles the active status of a link
// POST /api/links/{id}/toggle
func (h *LinkHandlers) ToggleLink(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	linkID := r.PathValue("id")
	if linkID == "" {
		respondError(w, http.StatusBadRequest, "link ID required")
		return
	}

	// Get link
	link, err := h.getLinkByID(linkID)
	if err != nil {
		respondError(w, http.StatusNotFound, "link not found")
		return
	}

	// Verify profile ownership
	if !h.userOwnsProfile(userID, link.ProfileID) {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	// Toggle active status
	link.Toggle()
	link.UpdatedAt = time.Now()

	// Update in database
	query := "UPDATE links SET is_active = ?, updated_at = ? WHERE id = ?"
	if h.db.Driver == "postgres" {
		query = "UPDATE links SET is_active = $1, updated_at = $2 WHERE id = $3"
	}

	_, err = h.db.Exec(query, link.IsActive, link.UpdatedAt, linkID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to toggle link")
		return
	}

	respondJSON(w, http.StatusOK, link)
}

// Helper functions

func (h *LinkHandlers) getLinkByID(id string) (*model.Link, error) {
	link := &model.Link{}
	query := `SELECT id, profile_id, service_id, title, username, url, icon_url,
			  background_color, text_color, position, is_active, click_count,
			  created_at, updated_at FROM links WHERE id = ?`

	if h.db.Driver == "postgres" {
		query = replaceQuestionMarks(query, 1)
	}

	err := h.db.QueryRow(query, id).Scan(&link.ID, &link.ProfileID, &link.ServiceID,
		&link.Title, &link.Username, &link.URL, &link.IconURL, &link.BackgroundColor,
		&link.TextColor, &link.Position, &link.IsActive, &link.ClickCount, &link.CreatedAt,
		&link.UpdatedAt)

	return link, err
}

func (h *LinkHandlers) getLinkCount(profileID string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM links WHERE profile_id = ?"
	if h.db.Driver == "postgres" {
		query = "SELECT COUNT(*) FROM links WHERE profile_id = $1"
	}
	err := h.db.QueryRow(query, profileID).Scan(&count)
	return count, err
}

func (h *LinkHandlers) getNextLinkPosition(profileID string) int {
	var maxPosition int
	query := "SELECT COALESCE(MAX(position), -1) FROM links WHERE profile_id = ?"
	if h.db.Driver == "postgres" {
		query = "SELECT COALESCE(MAX(position), -1) FROM links WHERE profile_id = $1"
	}
	h.db.QueryRow(query, profileID).Scan(&maxPosition)
	return maxPosition + 1
}

func (h *LinkHandlers) userOwnsProfile(userID, profileID string) bool {
	var count int
	query := "SELECT COUNT(*) FROM profiles WHERE id = ? AND user_id = ?"
	if h.db.Driver == "postgres" {
		query = "SELECT COUNT(*) FROM profiles WHERE id = $1 AND user_id = $2"
	}
	h.db.QueryRow(query, profileID, userID).Scan(&count)
	return count > 0
}

func (h *LinkHandlers) reorderLinks(profileID string) {
	// Get all links for profile ordered by position
	query := "SELECT id FROM links WHERE profile_id = ? ORDER BY position ASC"
	if h.db.Driver == "postgres" {
		query = "SELECT id FROM links WHERE profile_id = $1 ORDER BY position ASC"
	}

	rows, err := h.db.Query(query, profileID)
	if err != nil {
		return
	}
	defer rows.Close()

	linkIDs := []string{}
	for rows.Next() {
		var id string
		rows.Scan(&id)
		linkIDs = append(linkIDs, id)
	}

	// Update positions
	updateQuery := "UPDATE links SET position = ? WHERE id = ?"
	if h.db.Driver == "postgres" {
		updateQuery = "UPDATE links SET position = $1 WHERE id = $2"
	}

	for i, linkID := range linkIDs {
		h.db.Exec(updateQuery, i, linkID)
	}
}
