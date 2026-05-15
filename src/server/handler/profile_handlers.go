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

// ProfileHandlers handles profile-related HTTP requests
type ProfileHandlers struct {
	db *store.DB
}

// NewProfileHandlers creates a new ProfileHandlers instance
func NewProfileHandlers(db *store.DB) *ProfileHandlers {
	return &ProfileHandlers{db: db}
}

// CreateProfileRequest represents a profile creation request
type CreateProfileRequest struct {
	Slug            string `json:"slug"`
	DisplayName     string `json:"display_name"`
	Bio             string `json:"bio,omitempty"`
	AvatarURL       string `json:"avatar_url,omitempty"`
	HeaderImageURL  string `json:"header_image_url,omitempty"`
	ShowUsernames   bool   `json:"show_usernames"`
	IsPublic        bool   `json:"is_public"`
	MetaTitle       string `json:"meta_title,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	DisplayName        string `json:"display_name,omitempty"`
	Bio                string `json:"bio,omitempty"`
	AvatarURL          string `json:"avatar_url,omitempty"`
	HeaderImageURL     string `json:"header_image_url,omitempty"`
	ShowUsernames      *bool  `json:"show_usernames,omitempty"`
	IsPublic           *bool  `json:"is_public,omitempty"`
	PasswordProtected  *bool  `json:"password_protected,omitempty"`
	ProtectionPassword string `json:"protection_password,omitempty"`
	AnalyticsEnabled   *bool  `json:"analytics_enabled,omitempty"`
	QRCodeEnabled      *bool  `json:"qr_code_enabled,omitempty"`
	MetaTitle          string `json:"meta_title,omitempty"`
	MetaDescription    string `json:"meta_description,omitempty"`
	OgImageURL         string `json:"og_image_url,omitempty"`
	CustomCSS          string `json:"custom_css,omitempty"`
}

// ListProfiles lists all profiles for the authenticated user
// GET /api/profiles
func (h *ProfileHandlers) ListProfiles(w http.ResponseWriter, r *http.Request) {
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	query := `SELECT id, user_id, slug, display_name, bio, avatar_url, header_image_url,
			  theme_id, custom_css, show_usernames, is_public, password_protected,
			  custom_domain, domain_verified, analytics_enabled, meta_title,
			  meta_description, og_image_url, view_count, qr_code_enabled,
			  created_at, updated_at
			  FROM profiles WHERE user_id = ? ORDER BY created_at DESC`

	if h.db.Driver == "postgres" {
		query = replaceQuestionMarks(query, 1)
	}

	rows, err := h.db.Query(query, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch profiles")
		return
	}
	defer rows.Close()

	profiles := []model.Profile{}
	for rows.Next() {
		var p model.Profile
		err := rows.Scan(&p.ID, &p.UserID, &p.Slug, &p.DisplayName, &p.Bio, &p.AvatarURL,
			&p.HeaderImageURL, &p.ThemeID, &p.CustomCSS, &p.ShowUsernames, &p.IsPublic,
			&p.PasswordProtected, &p.CustomDomain, &p.DomainVerified, &p.AnalyticsEnabled,
			&p.MetaTitle, &p.MetaDescription, &p.OgImageURL, &p.ViewCount, &p.QRCodeEnabled,
			&p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}
		profiles = append(profiles, p)
	}

	respondJSON(w, http.StatusOK, profiles)
}

// GetProfile retrieves a specific profile
// GET /api/profiles/{id}
func (h *ProfileHandlers) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	profileID := r.PathValue("id")
	if profileID == "" {
		respondError(w, http.StatusBadRequest, "profile ID required")
		return
	}

	profile, err := h.getProfileByID(profileID)
	if err != nil {
		respondError(w, http.StatusNotFound, "profile not found")
		return
	}

	// Ensure user owns this profile
	if profile.UserID != userID {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	respondJSON(w, http.StatusOK, profile)
}

// CreateProfile creates a new profile
// POST /api/profiles
func (h *ProfileHandlers) CreateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Check max profiles per user
	maxProfilesStr, _ := h.db.GetSetting("max_profiles_per_user")
	maxProfiles, _ := strconv.Atoi(maxProfilesStr)
	if maxProfiles == 0 {
		maxProfiles = 5
	}

	currentCount, _ := h.getProfileCount(userID)
	if currentCount >= maxProfiles {
		respondError(w, http.StatusForbidden, "maximum number of profiles reached")
		return
	}

	// Check if slug exists
	if h.slugExists(req.Slug) {
		respondError(w, http.StatusConflict, "slug already exists")
		return
	}

	// Create profile
	profile := &model.Profile{
		UserID:           userID,
		Slug:             req.Slug,
		DisplayName:      req.DisplayName,
		Bio:              req.Bio,
		AvatarURL:        req.AvatarURL,
		HeaderImageURL:   req.HeaderImageURL,
		ThemeID:          "00000000-0000-0000-0000-000000000001",
		ShowUsernames:    req.ShowUsernames,
		IsPublic:         req.IsPublic,
		AnalyticsEnabled: true,
		QRCodeEnabled:    true,
		MetaTitle:        req.MetaTitle,
		MetaDescription:  req.MetaDescription,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Validate profile
	if err := profile.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Insert profile
	query := `INSERT INTO profiles (id, user_id, slug, display_name, bio, avatar_url,
			  header_image_url, theme_id, show_usernames, is_public, analytics_enabled,
			  qr_code_enabled, meta_title, meta_description, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	if h.db.Driver == "postgres" {
		query = `INSERT INTO profiles (user_id, slug, display_name, bio, avatar_url,
				 header_image_url, theme_id, show_usernames, is_public, analytics_enabled,
				 qr_code_enabled, meta_title, meta_description, created_at, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
				 RETURNING id`
		err := h.db.QueryRow(query, profile.UserID, profile.Slug, profile.DisplayName,
			profile.Bio, profile.AvatarURL, profile.HeaderImageURL, profile.ThemeID,
			profile.ShowUsernames, profile.IsPublic, profile.AnalyticsEnabled,
			profile.QRCodeEnabled, profile.MetaTitle, profile.MetaDescription,
			profile.CreatedAt, profile.UpdatedAt).Scan(&profile.ID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create profile")
			return
		}
	} else {
		profile.ID = generateUUID()
		_, err := h.db.Exec(query, profile.ID, profile.UserID, profile.Slug,
			profile.DisplayName, profile.Bio, profile.AvatarURL, profile.HeaderImageURL,
			profile.ThemeID, profile.ShowUsernames, profile.IsPublic, profile.AnalyticsEnabled,
			profile.QRCodeEnabled, profile.MetaTitle, profile.MetaDescription,
			profile.CreatedAt, profile.UpdatedAt)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create profile")
			return
		}
	}

	respondJSON(w, http.StatusCreated, profile)
}

// UpdateProfile updates an existing profile
// PUT /api/profiles/{id}
func (h *ProfileHandlers) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	profileID := r.PathValue("id")
	if profileID == "" {
		respondError(w, http.StatusBadRequest, "profile ID required")
		return
	}

	// Get existing profile
	profile, err := h.getProfileByID(profileID)
	if err != nil {
		respondError(w, http.StatusNotFound, "profile not found")
		return
	}

	// Ensure user owns this profile
	if profile.UserID != userID {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Update fields
	if req.DisplayName != "" {
		profile.DisplayName = req.DisplayName
	}
	if req.Bio != "" {
		profile.Bio = req.Bio
	}
	if req.AvatarURL != "" {
		profile.AvatarURL = req.AvatarURL
	}
	if req.HeaderImageURL != "" {
		profile.HeaderImageURL = req.HeaderImageURL
	}
	if req.ShowUsernames != nil {
		profile.ShowUsernames = *req.ShowUsernames
	}
	if req.IsPublic != nil {
		profile.IsPublic = *req.IsPublic
	}
	if req.PasswordProtected != nil {
		profile.PasswordProtected = *req.PasswordProtected
	}
	if req.ProtectionPassword != "" {
		profile.ProtectionPassword = req.ProtectionPassword
	}
	if req.AnalyticsEnabled != nil {
		profile.AnalyticsEnabled = *req.AnalyticsEnabled
	}
	if req.QRCodeEnabled != nil {
		profile.QRCodeEnabled = *req.QRCodeEnabled
	}
	if req.MetaTitle != "" {
		profile.MetaTitle = req.MetaTitle
	}
	if req.MetaDescription != "" {
		profile.MetaDescription = req.MetaDescription
	}
	if req.OgImageURL != "" {
		profile.OgImageURL = req.OgImageURL
	}
	if req.CustomCSS != "" {
		profile.CustomCSS = req.CustomCSS
	}

	profile.UpdatedAt = time.Now()

	// Validate profile
	if err := profile.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update in database
	query := `UPDATE profiles SET display_name = ?, bio = ?, avatar_url = ?,
			  header_image_url = ?, show_usernames = ?, is_public = ?,
			  password_protected = ?, protection_password = ?, analytics_enabled = ?,
			  qr_code_enabled = ?, meta_title = ?, meta_description = ?, og_image_url = ?,
			  custom_css = ?, updated_at = ? WHERE id = ?`

	if h.db.Driver == "postgres" {
		query = replaceQuestionMarks(query, 16)
	}

	_, err = h.db.Exec(query, profile.DisplayName, profile.Bio, profile.AvatarURL,
		profile.HeaderImageURL, profile.ShowUsernames, profile.IsPublic,
		profile.PasswordProtected, profile.ProtectionPassword, profile.AnalyticsEnabled,
		profile.QRCodeEnabled, profile.MetaTitle, profile.MetaDescription, profile.OgImageURL,
		profile.CustomCSS, profile.UpdatedAt, profileID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	respondJSON(w, http.StatusOK, profile)
}

// DeleteProfile deletes a profile
// DELETE /api/profiles/{id}
func (h *ProfileHandlers) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	profileID := r.PathValue("id")
	if profileID == "" {
		respondError(w, http.StatusBadRequest, "profile ID required")
		return
	}

	// Get profile
	profile, err := h.getProfileByID(profileID)
	if err != nil {
		respondError(w, http.StatusNotFound, "profile not found")
		return
	}

	// Ensure user owns this profile
	if profile.UserID != userID {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	// Delete profile (cascade will delete related data)
	query := "DELETE FROM profiles WHERE id = ?"
	if h.db.Driver == "postgres" {
		query = "DELETE FROM profiles WHERE id = $1"
	}

	_, err = h.db.Exec(query, profileID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete profile")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "profile deleted successfully",
	})
}

// DuplicateProfile duplicates an existing profile
// POST /api/profiles/{id}/duplicate
func (h *ProfileHandlers) DuplicateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	profileID := r.PathValue("id")
	if profileID == "" {
		respondError(w, http.StatusBadRequest, "profile ID required")
		return
	}

	// Get original profile
	original, err := h.getProfileByID(profileID)
	if err != nil {
		respondError(w, http.StatusNotFound, "profile not found")
		return
	}

	// Ensure user owns this profile
	if original.UserID != userID {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	// Check max profiles limit
	maxProfilesStr, _ := h.db.GetSetting("max_profiles_per_user")
	maxProfiles, _ := strconv.Atoi(maxProfilesStr)
	if maxProfiles == 0 {
		maxProfiles = 5
	}

	currentCount, _ := h.getProfileCount(userID)
	if currentCount >= maxProfiles {
		respondError(w, http.StatusForbidden, "maximum number of profiles reached")
		return
	}

	// Create new slug
	newSlug := original.Slug + "-copy"
	counter := 1
	for h.slugExists(newSlug) {
		newSlug = original.Slug + "-copy-" + strconv.Itoa(counter)
		counter++
	}

	// Create duplicate profile
	duplicate := *original
	duplicate.Slug = newSlug
	duplicate.CreatedAt = time.Now()
	duplicate.UpdatedAt = time.Now()
	duplicate.ViewCount = 0

	// Insert duplicate
	query := `INSERT INTO profiles (id, user_id, slug, display_name, bio, avatar_url,
			  header_image_url, theme_id, custom_css, show_usernames, is_public,
			  analytics_enabled, qr_code_enabled, meta_title, meta_description, og_image_url,
			  created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	if h.db.Driver == "postgres" {
		query = `INSERT INTO profiles (user_id, slug, display_name, bio, avatar_url,
				 header_image_url, theme_id, custom_css, show_usernames, is_public,
				 analytics_enabled, qr_code_enabled, meta_title, meta_description, og_image_url,
				 created_at, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
				 RETURNING id`
		err = h.db.QueryRow(query, duplicate.UserID, duplicate.Slug, duplicate.DisplayName,
			duplicate.Bio, duplicate.AvatarURL, duplicate.HeaderImageURL, duplicate.ThemeID,
			duplicate.CustomCSS, duplicate.ShowUsernames, duplicate.IsPublic,
			duplicate.AnalyticsEnabled, duplicate.QRCodeEnabled, duplicate.MetaTitle,
			duplicate.MetaDescription, duplicate.OgImageURL, duplicate.CreatedAt,
			duplicate.UpdatedAt).Scan(&duplicate.ID)
	} else {
		duplicate.ID = generateUUID()
		_, err = h.db.Exec(query, duplicate.ID, duplicate.UserID, duplicate.Slug,
			duplicate.DisplayName, duplicate.Bio, duplicate.AvatarURL, duplicate.HeaderImageURL,
			duplicate.ThemeID, duplicate.CustomCSS, duplicate.ShowUsernames, duplicate.IsPublic,
			duplicate.AnalyticsEnabled, duplicate.QRCodeEnabled, duplicate.MetaTitle,
			duplicate.MetaDescription, duplicate.OgImageURL, duplicate.CreatedAt,
			duplicate.UpdatedAt)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to duplicate profile")
		return
	}

	respondJSON(w, http.StatusCreated, duplicate)
}

// GenerateQRCode generates a QR code for the profile
// GET /api/profiles/{id}/qr
func (h *ProfileHandlers) GenerateQRCode(w http.ResponseWriter, r *http.Request) {
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	profileID := r.PathValue("id")
	if profileID == "" {
		respondError(w, http.StatusBadRequest, "profile ID required")
		return
	}

	// Get profile
	profile, err := h.getProfileByID(profileID)
	if err != nil {
		respondError(w, http.StatusNotFound, "profile not found")
		return
	}

	// Ensure user owns this profile
	if profile.UserID != userID {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	siteURL, _ := h.db.GetSetting("site_url")
	profileURL := profile.GetPublicURL(siteURL)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"profile_url": profileURL,
		"message":     "QR code generation to be implemented",
	})
}

// VerifyDomain verifies a custom domain for the profile
// POST /api/profiles/{id}/verify-domain
func (h *ProfileHandlers) VerifyDomain(w http.ResponseWriter, r *http.Request) {
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	profileID := r.PathValue("id")
	if profileID == "" {
		respondError(w, http.StatusBadRequest, "profile ID required")
		return
	}

	// Get profile
	profile, err := h.getProfileByID(profileID)
	if err != nil {
		respondError(w, http.StatusNotFound, "profile not found")
		return
	}

	// Ensure user owns this profile
	if profile.UserID != userID {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "domain verification not yet implemented",
	})
}

// Helper functions

func (h *ProfileHandlers) getProfileByID(id string) (*model.Profile, error) {
	profile := &model.Profile{}
	query := `SELECT id, user_id, slug, display_name, bio, avatar_url, header_image_url,
			  theme_id, custom_css, show_usernames, is_public, password_protected,
			  protection_password, custom_domain, domain_verified, analytics_enabled,
			  meta_title, meta_description, og_image_url, view_count, qr_code_enabled,
			  created_at, updated_at FROM profiles WHERE id = ?`

	if h.db.Driver == "postgres" {
		query = replaceQuestionMarks(query, 1)
	}

	err := h.db.QueryRow(query, id).Scan(&profile.ID, &profile.UserID, &profile.Slug,
		&profile.DisplayName, &profile.Bio, &profile.AvatarURL, &profile.HeaderImageURL,
		&profile.ThemeID, &profile.CustomCSS, &profile.ShowUsernames, &profile.IsPublic,
		&profile.PasswordProtected, &profile.ProtectionPassword, &profile.CustomDomain,
		&profile.DomainVerified, &profile.AnalyticsEnabled, &profile.MetaTitle,
		&profile.MetaDescription, &profile.OgImageURL, &profile.ViewCount, &profile.QRCodeEnabled,
		&profile.CreatedAt, &profile.UpdatedAt)

	return profile, err
}

func (h *ProfileHandlers) getProfileCount(userID string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM profiles WHERE user_id = ?"
	if h.db.Driver == "postgres" {
		query = "SELECT COUNT(*) FROM profiles WHERE user_id = $1"
	}
	err := h.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

func (h *ProfileHandlers) slugExists(slug string) bool {
	var count int
	query := "SELECT COUNT(*) FROM profiles WHERE slug = ?"
	if h.db.Driver == "postgres" {
		query = "SELECT COUNT(*) FROM profiles WHERE slug = $1"
	}
	h.db.QueryRow(query, slug).Scan(&count)
	return count > 0
}
