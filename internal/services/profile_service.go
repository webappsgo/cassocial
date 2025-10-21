package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/casapps/cassocial/internal/database"
	"github.com/casapps/cassocial/internal/models"
)

var (
	ErrProfileNotFound      = errors.New("profile not found")
	ErrSlugAlreadyExists    = errors.New("slug already exists")
	ErrMaxProfilesReached   = errors.New("maximum number of profiles reached")
	ErrInvalidSlugFormat    = errors.New("invalid slug format")
	ErrDomainNotVerified    = errors.New("domain not verified")
	ErrProfileLimitExceeded = errors.New("profile limit exceeded for user")
)

// ProfileService handles profile business logic
type ProfileService struct {
	db *database.DB
}

// NewProfileService creates a new profile service
func NewProfileService(db *database.DB) *ProfileService {
	return &ProfileService{db: db}
}

// Create creates a new profile
func (s *ProfileService) Create(userID string, profile *models.Profile) error {
	// Check profile limit for user
	maxProfiles, err := s.getMaxProfilesPerUser()
	if err != nil {
		return fmt.Errorf("failed to get max profiles setting: %w", err)
	}

	count, err := s.CountByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to count user profiles: %w", err)
	}

	if count >= maxProfiles {
		return ErrMaxProfilesReached
	}

	// Generate slug if not provided
	if profile.Slug == "" {
		profile.Slug, err = s.generateSlug(profile.DisplayName)
		if err != nil {
			return fmt.Errorf("failed to generate slug: %w", err)
		}
	}

	// Validate slug format
	if !s.isValidSlug(profile.Slug) {
		return ErrInvalidSlugFormat
	}

	// Check if slug already exists
	exists, err := s.SlugExists(profile.Slug)
	if err != nil {
		return fmt.Errorf("failed to check slug existence: %w", err)
	}
	if exists {
		return ErrSlugAlreadyExists
	}

	// Validate profile
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("profile validation failed: %w", err)
	}

	// Set defaults
	profile.UserID = userID
	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()
	if profile.ThemeID == "" {
		profile.ThemeID = "00000000-0000-0000-0000-000000000001" // Default theme
	}

	// Insert into database
	query := `
		INSERT INTO profiles (
			id, user_id, slug, display_name, bio, avatar_url, header_image_url,
			theme_id, custom_css, show_usernames, is_public, password_protected,
			protection_password, custom_domain, domain_verified, analytics_enabled,
			meta_title, meta_description, og_image_url, view_count, qr_code_enabled,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	profile.ID = s.generateID()
	_, err = s.db.Exec(query,
		profile.ID, profile.UserID, profile.Slug, profile.DisplayName, profile.Bio,
		profile.AvatarURL, profile.HeaderImageURL, profile.ThemeID, profile.CustomCSS,
		profile.ShowUsernames, profile.IsPublic, profile.PasswordProtected,
		profile.ProtectionPassword, profile.CustomDomain, profile.DomainVerified,
		profile.AnalyticsEnabled, profile.MetaTitle, profile.MetaDescription,
		profile.OgImageURL, profile.ViewCount, profile.QRCodeEnabled,
		profile.CreatedAt, profile.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert profile: %w", err)
	}

	return nil
}

// GetByID retrieves a profile by ID
func (s *ProfileService) GetByID(id string) (*models.Profile, error) {
	query := `SELECT * FROM profiles WHERE id = ?`

	var profile models.Profile
	err := s.db.QueryRow(query, id).Scan(
		&profile.ID, &profile.UserID, &profile.Slug, &profile.DisplayName,
		&profile.Bio, &profile.AvatarURL, &profile.HeaderImageURL, &profile.ThemeID,
		&profile.CustomCSS, &profile.ShowUsernames, &profile.IsPublic,
		&profile.PasswordProtected, &profile.ProtectionPassword, &profile.CustomDomain,
		&profile.DomainVerified, &profile.AnalyticsEnabled, &profile.MetaTitle,
		&profile.MetaDescription, &profile.OgImageURL, &profile.ViewCount,
		&profile.QRCodeEnabled, &profile.CreatedAt, &profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrProfileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return &profile, nil
}

// GetBySlug retrieves a profile by slug
func (s *ProfileService) GetBySlug(slug string) (*models.Profile, error) {
	query := `SELECT * FROM profiles WHERE slug = ?`

	var profile models.Profile
	err := s.db.QueryRow(query, slug).Scan(
		&profile.ID, &profile.UserID, &profile.Slug, &profile.DisplayName,
		&profile.Bio, &profile.AvatarURL, &profile.HeaderImageURL, &profile.ThemeID,
		&profile.CustomCSS, &profile.ShowUsernames, &profile.IsPublic,
		&profile.PasswordProtected, &profile.ProtectionPassword, &profile.CustomDomain,
		&profile.DomainVerified, &profile.AnalyticsEnabled, &profile.MetaTitle,
		&profile.MetaDescription, &profile.OgImageURL, &profile.ViewCount,
		&profile.QRCodeEnabled, &profile.CreatedAt, &profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrProfileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return &profile, nil
}

// GetByUserID retrieves all profiles for a user
func (s *ProfileService) GetByUserID(userID string) ([]*models.Profile, error) {
	query := `SELECT * FROM profiles WHERE user_id = ? ORDER BY created_at DESC`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query profiles: %w", err)
	}
	defer rows.Close()

	var profiles []*models.Profile
	for rows.Next() {
		var profile models.Profile
		err := rows.Scan(
			&profile.ID, &profile.UserID, &profile.Slug, &profile.DisplayName,
			&profile.Bio, &profile.AvatarURL, &profile.HeaderImageURL, &profile.ThemeID,
			&profile.CustomCSS, &profile.ShowUsernames, &profile.IsPublic,
			&profile.PasswordProtected, &profile.ProtectionPassword, &profile.CustomDomain,
			&profile.DomainVerified, &profile.AnalyticsEnabled, &profile.MetaTitle,
			&profile.MetaDescription, &profile.OgImageURL, &profile.ViewCount,
			&profile.QRCodeEnabled, &profile.CreatedAt, &profile.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan profile: %w", err)
		}
		profiles = append(profiles, &profile)
	}

	return profiles, nil
}

// Update updates a profile
func (s *ProfileService) Update(profile *models.Profile) error {
	// Validate profile
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("profile validation failed: %w", err)
	}

	// Check if slug changed and if new slug exists
	existing, err := s.GetByID(profile.ID)
	if err != nil {
		return err
	}

	if existing.Slug != profile.Slug {
		exists, err := s.SlugExists(profile.Slug)
		if err != nil {
			return fmt.Errorf("failed to check slug existence: %w", err)
		}
		if exists {
			return ErrSlugAlreadyExists
		}
	}

	profile.UpdatedAt = time.Now()

	query := `
		UPDATE profiles SET
			slug = ?, display_name = ?, bio = ?, avatar_url = ?, header_image_url = ?,
			theme_id = ?, custom_css = ?, show_usernames = ?, is_public = ?,
			password_protected = ?, protection_password = ?, custom_domain = ?,
			domain_verified = ?, analytics_enabled = ?, meta_title = ?,
			meta_description = ?, og_image_url = ?, qr_code_enabled = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = s.db.Exec(query,
		profile.Slug, profile.DisplayName, profile.Bio, profile.AvatarURL,
		profile.HeaderImageURL, profile.ThemeID, profile.CustomCSS,
		profile.ShowUsernames, profile.IsPublic, profile.PasswordProtected,
		profile.ProtectionPassword, profile.CustomDomain, profile.DomainVerified,
		profile.AnalyticsEnabled, profile.MetaTitle, profile.MetaDescription,
		profile.OgImageURL, profile.QRCodeEnabled, profile.UpdatedAt, profile.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	return nil
}

// Delete deletes a profile
func (s *ProfileService) Delete(id string) error {
	query := `DELETE FROM profiles WHERE id = ?`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrProfileNotFound
	}

	return nil
}

// Duplicate creates a copy of a profile
func (s *ProfileService) Duplicate(profileID, userID string) (*models.Profile, error) {
	// Get original profile
	original, err := s.GetByID(profileID)
	if err != nil {
		return nil, err
	}

	// Check if user owns the profile
	if original.UserID != userID {
		return nil, errors.New("user does not own this profile")
	}

	// Create duplicate
	duplicate := *original
	duplicate.ID = ""
	duplicate.Slug = s.generateUniqueSlug(original.Slug)
	duplicate.DisplayName = original.DisplayName + " (Copy)"
	duplicate.CustomDomain = ""
	duplicate.DomainVerified = false
	duplicate.ViewCount = 0

	if err := s.Create(userID, &duplicate); err != nil {
		return nil, err
	}

	return &duplicate, nil
}

// IncrementViewCount increments the view count for a profile
func (s *ProfileService) IncrementViewCount(id string) error {
	query := `UPDATE profiles SET view_count = view_count + 1 WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	return nil
}

// SlugExists checks if a slug already exists
func (s *ProfileService) SlugExists(slug string) (bool, error) {
	query := `SELECT COUNT(*) FROM profiles WHERE slug = ?`

	var count int
	err := s.db.QueryRow(query, slug).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check slug existence: %w", err)
	}

	return count > 0, nil
}

// CountByUserID counts the number of profiles for a user
func (s *ProfileService) CountByUserID(userID string) (int, error) {
	query := `SELECT COUNT(*) FROM profiles WHERE user_id = ?`

	var count int
	err := s.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count profiles: %w", err)
	}

	return count, nil
}

// VerifyDomain verifies a custom domain for a profile
func (s *ProfileService) VerifyDomain(profileID, domain string) error {
	// TODO: Implement DNS verification logic
	// Check for CNAME or TXT record

	query := `UPDATE profiles SET domain_verified = 1 WHERE id = ? AND custom_domain = ?`

	result, err := s.db.Exec(query, profileID, domain)
	if err != nil {
		return fmt.Errorf("failed to verify domain: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrProfileNotFound
	}

	return nil
}

// GetQRCodeSettings retrieves QR code settings for a profile
func (s *ProfileService) GetQRCodeSettings(profileID string) (*models.QRCodeSettings, error) {
	query := `SELECT * FROM qr_code_settings WHERE profile_id = ?`

	var settings models.QRCodeSettings
	err := s.db.QueryRow(query, profileID).Scan(
		&settings.ProfileID, &settings.Size, &settings.ErrorCorrection,
		&settings.Style, &settings.DarkColor, &settings.LightColor,
		&settings.LogoEnabled, &settings.LogoSize, &settings.Format,
		&settings.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Return default settings
		return &models.QRCodeSettings{
			ProfileID:       profileID,
			Size:            256,
			ErrorCorrection: "M",
			Style:           "square",
			DarkColor:       "#000000",
			LightColor:      "#ffffff",
			LogoEnabled:     false,
			LogoSize:        30,
			Format:          "png",
			UpdatedAt:       time.Now(),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get QR code settings: %w", err)
	}

	return &settings, nil
}

// UpdateQRCodeSettings updates QR code settings for a profile
func (s *ProfileService) UpdateQRCodeSettings(settings *models.QRCodeSettings) error {
	// Validate settings
	if err := settings.Validate(); err != nil {
		return fmt.Errorf("QR code settings validation failed: %w", err)
	}

	settings.UpdatedAt = time.Now()

	query := `
		INSERT INTO qr_code_settings (
			profile_id, size, error_correction, style, dark_color, light_color,
			logo_enabled, logo_size, format, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(profile_id) DO UPDATE SET
			size = EXCLUDED.size,
			error_correction = EXCLUDED.error_correction,
			style = EXCLUDED.style,
			dark_color = EXCLUDED.dark_color,
			light_color = EXCLUDED.light_color,
			logo_enabled = EXCLUDED.logo_enabled,
			logo_size = EXCLUDED.logo_size,
			format = EXCLUDED.format,
			updated_at = EXCLUDED.updated_at
	`

	_, err := s.db.Exec(query,
		settings.ProfileID, settings.Size, settings.ErrorCorrection,
		settings.Style, settings.DarkColor, settings.LightColor,
		settings.LogoEnabled, settings.LogoSize, settings.Format,
		settings.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update QR code settings: %w", err)
	}

	return nil
}

// Helper functions

func (s *ProfileService) generateSlug(displayName string) (string, error) {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(displayName)
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove non-alphanumeric characters except hyphens and underscores
	reg := regexp.MustCompile(`[^a-z0-9_-]+`)
	slug = reg.ReplaceAllString(slug, "")

	// Ensure slug is unique
	originalSlug := slug
	counter := 1
	for {
		exists, err := s.SlugExists(slug)
		if err != nil {
			return "", err
		}
		if !exists {
			break
		}
		slug = fmt.Sprintf("%s-%d", originalSlug, counter)
		counter++
	}

	return slug, nil
}

func (s *ProfileService) generateUniqueSlug(baseSlug string) string {
	counter := 1
	slug := fmt.Sprintf("%s-copy", baseSlug)
	for {
		exists, _ := s.SlugExists(slug)
		if !exists {
			break
		}
		slug = fmt.Sprintf("%s-copy-%d", baseSlug, counter)
		counter++
	}
	return slug
}

func (s *ProfileService) isValidSlug(slug string) bool {
	// Slug must be alphanumeric with hyphens and underscores only
	match, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, slug)
	return match && len(slug) >= 3 && len(slug) <= 100
}

func (s *ProfileService) generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *ProfileService) getMaxProfilesPerUser() (int, error) {
	maxProfilesStr, err := s.db.GetSetting("max_profiles_per_user")
	if err != nil {
		return 5, nil // Default value
	}

	var maxProfiles int
	_, err = fmt.Sscanf(maxProfilesStr, "%d", &maxProfiles)
	if err != nil {
		return 5, nil // Default value
	}

	return maxProfiles, nil
}
