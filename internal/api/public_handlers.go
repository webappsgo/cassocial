package api

import (
	"net/http"

	"github.com/casapps/cassocial/internal/database"
	"github.com/casapps/cassocial/internal/models"
)

// PublicHandlers handles public API requests (no authentication required)
type PublicHandlers struct {
	db *database.DB
}

// NewPublicHandlers creates a new PublicHandlers instance
func NewPublicHandlers(db *database.DB) *PublicHandlers {
	return &PublicHandlers{db: db}
}

// GetPublicProfile retrieves public profile data by username/slug
// GET /api/v1/profiles/{username}
func (h *PublicHandlers) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		respondError(w, http.StatusBadRequest, "username required")
		return
	}

	// Get profile by slug
	profile := &models.Profile{}
	query := `SELECT id, user_id, slug, display_name, bio, avatar_url, header_image_url,
			  theme_id, show_usernames, is_public, password_protected, custom_domain,
			  domain_verified, meta_title, meta_description, og_image_url, view_count,
			  qr_code_enabled, created_at, updated_at
			  FROM profiles WHERE slug = ?`

	if h.db.Driver == "postgres" {
		query = replaceQuestionMarks(query, 1)
	}

	err := h.db.QueryRow(query, username).Scan(&profile.ID, &profile.UserID, &profile.Slug,
		&profile.DisplayName, &profile.Bio, &profile.AvatarURL, &profile.HeaderImageURL,
		&profile.ThemeID, &profile.ShowUsernames, &profile.IsPublic, &profile.PasswordProtected,
		&profile.CustomDomain, &profile.DomainVerified, &profile.MetaTitle,
		&profile.MetaDescription, &profile.OgImageURL, &profile.ViewCount, &profile.QRCodeEnabled,
		&profile.CreatedAt, &profile.UpdatedAt)

	if err != nil {
		respondError(w, http.StatusNotFound, "profile not found")
		return
	}

	// Check if profile is public
	if !profile.IsPublic {
		respondError(w, http.StatusForbidden, "profile is private")
		return
	}

	// Increment view count
	h.incrementViewCount(profile.ID)

	// Track analytics
	h.trackView(r, profile.ID)

	respondJSON(w, http.StatusOK, profile)
}

// GetPublicProfileLinks retrieves links for a public profile
// GET /api/v1/profiles/{username}/links
func (h *PublicHandlers) GetPublicProfileLinks(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		respondError(w, http.StatusBadRequest, "username required")
		return
	}

	// Get profile by slug
	var profileID string
	var isPublic bool
	query := "SELECT id, is_public FROM profiles WHERE slug = ?"
	if h.db.Driver == "postgres" {
		query = "SELECT id, is_public FROM profiles WHERE slug = $1"
	}

	err := h.db.QueryRow(query, username).Scan(&profileID, &isPublic)
	if err != nil {
		respondError(w, http.StatusNotFound, "profile not found")
		return
	}

	// Check if profile is public
	if !isPublic {
		respondError(w, http.StatusForbidden, "profile is private")
		return
	}

	// Get active links
	query = `SELECT id, profile_id, service_id, title, username, url, icon_url,
			 background_color, text_color, position, is_active, click_count,
			 created_at, updated_at
			 FROM links WHERE profile_id = ? AND is_active = ?
			 ORDER BY position ASC`

	if h.db.Driver == "postgres" {
		query = replaceQuestionMarks(query, 2)
	}

	rows, err := h.db.Query(query, profileID, true)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch links")
		return
	}
	defer rows.Close()

	links := []models.Link{}
	for rows.Next() {
		var l models.Link
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

// GetPublicProfileQR generates QR code for a public profile
// GET /api/v1/profiles/{username}/qr
func (h *PublicHandlers) GetPublicProfileQR(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		respondError(w, http.StatusBadRequest, "username required")
		return
	}

	// Get profile by slug
	var profileID string
	var isPublic, qrEnabled bool
	query := "SELECT id, is_public, qr_code_enabled FROM profiles WHERE slug = ?"
	if h.db.Driver == "postgres" {
		query = "SELECT id, is_public, qr_code_enabled FROM profiles WHERE slug = $1"
	}

	err := h.db.QueryRow(query, username).Scan(&profileID, &isPublic, &qrEnabled)
	if err != nil {
		respondError(w, http.StatusNotFound, "profile not found")
		return
	}

	// Check if profile is public and QR is enabled
	if !isPublic {
		respondError(w, http.StatusForbidden, "profile is private")
		return
	}

	if !qrEnabled {
		respondError(w, http.StatusForbidden, "QR code is disabled for this profile")
		return
	}

	// TODO: Implement QR code generation
	siteURL, _ := h.db.GetSetting("site_url")
	profileURL := siteURL + "/" + username

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"profile_url": profileURL,
		"message":     "QR code generation to be implemented",
	})
}

// TrackLinkClick tracks a link click and redirects
// GET /api/v1/link/{id}/click
func (h *PublicHandlers) TrackLinkClick(w http.ResponseWriter, r *http.Request) {
	linkID := r.PathValue("id")
	if linkID == "" {
		respondError(w, http.StatusBadRequest, "link ID required")
		return
	}

	// Get link
	var link models.Link
	query := `SELECT id, profile_id, url, is_active FROM links WHERE id = ?`
	if h.db.Driver == "postgres" {
		query = replaceQuestionMarks(query, 1)
	}

	err := h.db.QueryRow(query, linkID).Scan(&link.ID, &link.ProfileID, &link.URL, &link.IsActive)
	if err != nil {
		respondError(w, http.StatusNotFound, "link not found")
		return
	}

	// Check if link is active
	if !link.IsActive {
		respondError(w, http.StatusForbidden, "link is inactive")
		return
	}

	// Increment click count
	h.incrementClickCount(linkID)

	// Track analytics
	h.trackClick(r, link.ProfileID, linkID)

	// Redirect to URL
	http.Redirect(w, r, link.URL, http.StatusTemporaryRedirect)
}

// Helper functions

func (h *PublicHandlers) incrementViewCount(profileID string) {
	query := "UPDATE profiles SET view_count = view_count + 1 WHERE id = ?"
	if h.db.Driver == "postgres" {
		query = "UPDATE profiles SET view_count = view_count + 1 WHERE id = $1"
	}
	h.db.Exec(query, profileID)
}

func (h *PublicHandlers) incrementClickCount(linkID string) {
	query := "UPDATE links SET click_count = click_count + 1 WHERE id = ?"
	if h.db.Driver == "postgres" {
		query = "UPDATE links SET click_count = click_count + 1 WHERE id = $1"
	}
	h.db.Exec(query, linkID)
}

func (h *PublicHandlers) trackView(r *http.Request, profileID string) {
	// Get IP address
	ip := getIPAddress(r)
	ipHash := hashIP(ip)

	// Get device type
	deviceType := detectDeviceType(r.UserAgent())

	// Insert analytics record
	query := `INSERT INTO analytics (id, profile_id, event_type, ip_hash, user_agent,
			  referrer, device_type, created_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	if h.db.Driver == "postgres" {
		query = `INSERT INTO analytics (profile_id, event_type, ip_hash, user_agent,
				 referrer, device_type, created_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7)`
		h.db.Exec(query, profileID, "view", ipHash, r.UserAgent(),
			r.Referer(), deviceType, getCurrentTimestamp())
	} else {
		h.db.Exec(query, generateUUID(), profileID, "view", ipHash, r.UserAgent(),
			r.Referer(), deviceType, getCurrentTimestamp())
	}
}

func (h *PublicHandlers) trackClick(r *http.Request, profileID, linkID string) {
	// Get IP address
	ip := getIPAddress(r)
	ipHash := hashIP(ip)

	// Get device type
	deviceType := detectDeviceType(r.UserAgent())

	// Insert analytics record
	query := `INSERT INTO analytics (id, profile_id, link_id, event_type, ip_hash,
			  user_agent, referrer, device_type, created_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	if h.db.Driver == "postgres" {
		query = `INSERT INTO analytics (profile_id, link_id, event_type, ip_hash,
				 user_agent, referrer, device_type, created_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
		h.db.Exec(query, profileID, linkID, "click", ipHash, r.UserAgent(),
			r.Referer(), deviceType, getCurrentTimestamp())
	} else {
		h.db.Exec(query, generateUUID(), profileID, linkID, "click", ipHash,
			r.UserAgent(), r.Referer(), deviceType, getCurrentTimestamp())
	}
}
