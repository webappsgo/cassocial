package model

import (
	"database/sql"
	"errors"
	"regexp"
	"time"
)

// Profile represents a user's public profile/landing page
type Profile struct {
	ID                  string       `json:"id" db:"id"`
	UserID              string       `json:"user_id" db:"user_id"`
	Slug                string       `json:"slug" db:"slug"`
	DisplayName         string       `json:"display_name" db:"display_name"`
	Bio                 string       `json:"bio" db:"bio"`
	AvatarURL           string       `json:"avatar_url" db:"avatar_url"`
	HeaderImageURL      string       `json:"header_image_url" db:"header_image_url"`
	ThemeID             string       `json:"theme_id" db:"theme_id"`
	CustomCSS           string       `json:"custom_css,omitempty" db:"custom_css"`
	ShowUsernames       bool         `json:"show_usernames" db:"show_usernames"`
	IsPublic            bool         `json:"is_public" db:"is_public"`
	PasswordProtected   bool         `json:"password_protected" db:"password_protected"`
	ProtectionPassword  string       `json:"-" db:"protection_password"`
	CustomDomain        string       `json:"custom_domain" db:"custom_domain"`
	DomainVerified      bool         `json:"domain_verified" db:"domain_verified"`
	AnalyticsEnabled    bool         `json:"analytics_enabled" db:"analytics_enabled"`
	MetaTitle           string       `json:"meta_title" db:"meta_title"`
	MetaDescription     string       `json:"meta_description" db:"meta_description"`
	OgImageURL          string       `json:"og_image_url" db:"og_image_url"`
	ViewCount           int          `json:"view_count" db:"view_count"`
	QRCodeEnabled       bool         `json:"qr_code_enabled" db:"qr_code_enabled"`
	CreatedAt           time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at" db:"updated_at"`
}

// ProfileTheme represents custom theme settings for a profile
type ProfileTheme struct {
	ProfileID              string    `json:"profile_id" db:"profile_id"`
	BackgroundType         string    `json:"background_type" db:"background_type"`
	BackgroundValue        string    `json:"background_value" db:"background_value"`
	ButtonStyle            string    `json:"button_style" db:"button_style"`
	ButtonAnimation        string    `json:"button_animation" db:"button_animation"`
	ButtonShadow           string    `json:"button_shadow" db:"button_shadow"`
	FontOverride           string    `json:"font_override" db:"font_override"`
	CustomCSS              string    `json:"custom_css" db:"custom_css"`
	LinkThumbnailPosition  string    `json:"link_thumbnail_position" db:"link_thumbnail_position"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

// ProfileMaintenance represents maintenance mode settings for a profile
type ProfileMaintenance struct {
	ProfileID    string         `json:"profile_id" db:"profile_id"`
	Status       string         `json:"status" db:"status"`
	Message      string         `json:"message" db:"message"`
	BypassToken  string         `json:"bypass_token,omitempty" db:"bypass_token"`
	StartedAt    time.Time      `json:"started_at" db:"started_at"`
	EstimatedEnd sql.NullTime   `json:"estimated_end,omitempty" db:"estimated_end"`
}

// QRCodeSettings represents QR code generation settings for a profile
type QRCodeSettings struct {
	ProfileID        string    `json:"profile_id" db:"profile_id"`
	Size             int       `json:"size" db:"size"`
	ErrorCorrection  string    `json:"error_correction" db:"error_correction"`
	Style            string    `json:"style" db:"style"`
	DarkColor        string    `json:"dark_color" db:"dark_color"`
	LightColor       string    `json:"light_color" db:"light_color"`
	LogoEnabled      bool      `json:"logo_enabled" db:"logo_enabled"`
	LogoSize         int       `json:"logo_size" db:"logo_size"`
	Format           string    `json:"format" db:"format"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// ProfileTag represents a tag associated with a profile
type ProfileTag struct {
	ProfileID string `json:"profile_id" db:"profile_id"`
	Tag       string `json:"tag" db:"tag"`
}

// FeaturedProfile represents a featured profile
type FeaturedProfile struct {
	ProfileID    string       `json:"profile_id" db:"profile_id"`
	FeaturedAt   time.Time    `json:"featured_at" db:"featured_at"`
	FeaturedUntil sql.NullTime `json:"featured_until,omitempty" db:"featured_until"`
	Reason       string       `json:"reason" db:"reason"`
}

// ProfileVerification represents verification status of a profile
type ProfileVerification struct {
	ProfileID        string         `json:"profile_id" db:"profile_id"`
	Verified         bool           `json:"verified" db:"verified"`
	VerificationType string         `json:"verification_type" db:"verification_type"`
	VerificationData string         `json:"verification_data,omitempty" db:"verification_data"`
	VerifiedAt       sql.NullTime   `json:"verified_at,omitempty" db:"verified_at"`
	VerifiedBy       string         `json:"verified_by,omitempty" db:"verified_by"`
}

var (
	ErrInvalidSlug            = errors.New("slug must be alphanumeric with hyphens and underscores only")
	ErrDisplayNameTooLong     = errors.New("display name must be 100 characters or less")
	ErrBioTooLong             = errors.New("bio must be 500 characters or less")
	ErrMetaTitleTooLong       = errors.New("meta title must be 60 characters or less")
	ErrMetaDescriptionTooLong = errors.New("meta description must be 160 characters or less")
	ErrInvalidBackgroundType  = errors.New("invalid background type")
	ErrInvalidButtonStyle     = errors.New("invalid button style")
	ErrInvalidQRCodeSize      = errors.New("invalid QR code size")
)

// Validate validates the profile model
func (p *Profile) Validate() error {
	// Validate slug
	slugRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !slugRegex.MatchString(p.Slug) {
		return ErrInvalidSlug
	}

	// Validate display name length
	if len(p.DisplayName) > 100 {
		return ErrDisplayNameTooLong
	}

	// Validate bio length
	if len(p.Bio) > 500 {
		return ErrBioTooLong
	}

	// Validate meta title length
	if len(p.MetaTitle) > 60 {
		return ErrMetaTitleTooLong
	}

	// Validate meta description length
	if len(p.MetaDescription) > 160 {
		return ErrMetaDescriptionTooLong
	}

	return nil
}

// GetPublicURL returns the public URL for the profile
func (p *Profile) GetPublicURL(baseURL string) string {
	if p.CustomDomain != "" && p.DomainVerified {
		return "https://" + p.CustomDomain
	}
	return baseURL + "/" + p.Slug
}

// IsAccessible checks if the profile is accessible without password
func (p *Profile) IsAccessible() bool {
	return p.IsPublic && !p.PasswordProtected
}

// IncrementViewCount increments the view count
func (p *Profile) IncrementViewCount() {
	p.ViewCount++
}

// Validate validates the profile theme model
func (pt *ProfileTheme) Validate() error {
	// Validate background type
	validBackgroundTypes := map[string]bool{"color": true, "gradient": true, "image": true}
	if !validBackgroundTypes[pt.BackgroundType] {
		return ErrInvalidBackgroundType
	}

	// Validate button style
	validButtonStyles := map[string]bool{"rounded": true, "square": true, "pill": true}
	if !validButtonStyles[pt.ButtonStyle] {
		return ErrInvalidButtonStyle
	}

	return nil
}

// Validate validates the QR code settings model
func (q *QRCodeSettings) Validate() error {
	// Validate size
	validSizes := map[int]bool{128: true, 256: true, 512: true, 1024: true}
	if !validSizes[q.Size] {
		return ErrInvalidQRCodeSize
	}

	return nil
}
