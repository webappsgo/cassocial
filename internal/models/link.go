package models

import (
	"errors"
	"net/url"
	"time"
)

// Link represents a link on a user's profile
type Link struct {
	ID              string    `json:"id" db:"id"`
	ProfileID       string    `json:"profile_id" db:"profile_id"`
	ServiceID       string    `json:"service_id,omitempty" db:"service_id"`
	Title           string    `json:"title" db:"title"`
	Username        string    `json:"username,omitempty" db:"username"`
	URL             string    `json:"url" db:"url"`
	IconURL         string    `json:"icon_url,omitempty" db:"icon_url"`
	BackgroundColor string    `json:"background_color,omitempty" db:"background_color"`
	TextColor       string    `json:"text_color,omitempty" db:"text_color"`
	Position        int       `json:"position" db:"position"`
	IsActive        bool      `json:"is_active" db:"is_active"`
	ClickCount      int       `json:"click_count" db:"click_count"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// FooterItem represents a footer element on a profile
type FooterItem struct {
	ID        string    `json:"id" db:"id"`
	ProfileID string    `json:"profile_id" db:"profile_id"`
	ItemType  string    `json:"item_type" db:"item_type"`
	Content   string    `json:"content" db:"content"` // JSONB stored as string
	Position  int       `json:"position" db:"position"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Shortlink represents a shortened URL
type Shortlink struct {
	ID         string       `json:"id" db:"id"`
	ShortCode  string       `json:"short_code" db:"short_code"`
	TargetURL  string       `json:"target_url" db:"target_url"`
	ProfileID  string       `json:"profile_id,omitempty" db:"profile_id"`
	Title      string       `json:"title,omitempty" db:"title"`
	ClickCount int          `json:"click_count" db:"click_count"`
	ExpiresAt  *time.Time   `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt  time.Time    `json:"created_at" db:"created_at"`
}

// Valid footer item types
const (
	FooterTypeText      = "text"
	FooterTypeLink      = "link"
	FooterTypeSocialRow = "social_row"
	FooterTypeBadge     = "badge"
	FooterTypeHTML      = "html"
)

var (
	ErrTitleTooLong     = errors.New("title must be 100 characters or less")
	ErrInvalidURL       = errors.New("invalid URL format")
	ErrInvalidFooterType = errors.New("invalid footer item type")
)

// Validate validates the link model
func (l *Link) Validate() error {
	// Validate title length
	if len(l.Title) > 100 {
		return ErrTitleTooLong
	}

	// Validate URL format
	if !isValidURL(l.URL) {
		return ErrInvalidURL
	}

	return nil
}

// IncrementClickCount increments the click counter
func (l *Link) IncrementClickCount() {
	l.ClickCount++
}

// Toggle toggles the active status
func (l *Link) Toggle() {
	l.IsActive = !l.IsActive
}

// GetDisplayText returns the display text for the link
// Format: {username}@{Service} or just {Service} if username is empty
func (l *Link) GetDisplayText(serviceName string, showUsernames bool) string {
	if !showUsernames || l.Username == "" {
		return serviceName
	}
	return l.Username + "@" + serviceName
}

// Validate validates the footer item model
func (f *FooterItem) Validate() error {
	// Validate item type
	validTypes := map[string]bool{
		FooterTypeText:      true,
		FooterTypeLink:      true,
		FooterTypeSocialRow: true,
		FooterTypeBadge:     true,
		FooterTypeHTML:      true,
	}

	if !validTypes[f.ItemType] {
		return ErrInvalidFooterType
	}

	return nil
}

// Validate validates the shortlink model
func (s *Shortlink) Validate() error {
	// Validate target URL
	if !isValidURL(s.TargetURL) {
		return ErrInvalidURL
	}

	return nil
}

// IncrementClickCount increments the click counter for shortlink
func (s *Shortlink) IncrementClickCount() {
	s.ClickCount++
}

// IsExpired checks if the shortlink has expired
func (s *Shortlink) IsExpired() bool {
	if s.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.ExpiresAt)
}

// isValidURL validates URL format
func isValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}
