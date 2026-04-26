package model

import (
	"errors"
	"time"
)

// Service represents a predefined service (social media platform, etc.)
type Service struct {
	ID                string    `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	Category          string    `json:"category" db:"category"`
	IconURL           string    `json:"icon_url,omitempty" db:"icon_url"`
	IconSVG           string    `json:"icon_svg,omitempty" db:"icon_svg"`
	URLPattern        string    `json:"url_pattern,omitempty" db:"url_pattern"`
	BackgroundColor   string    `json:"background_color,omitempty" db:"background_color"`
	TextColor         string    `json:"text_color,omitempty" db:"text_color"`
	Popularity        int       `json:"popularity" db:"popularity"`
	IsActive          bool      `json:"is_active" db:"is_active"`
	RequiresUsername  bool      `json:"requires_username" db:"requires_username"`
	PlaceholderText   string    `json:"placeholder_text,omitempty" db:"placeholder_text"`
	ValidationPattern string    `json:"validation_pattern,omitempty" db:"validation_pattern"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// Valid service categories
const (
	CategorySocial        = "social"
	CategoryProfessional  = "professional"
	CategoryDevelopment   = "development"
	CategoryContent       = "content"
	CategoryPayment       = "payment"
	CategoryGaming        = "gaming"
	CategoryCommunication = "communication"
	CategoryPortfolio     = "portfolio"
	CategoryOther         = "other"
)

var (
	ErrInvalidCategory = errors.New("invalid service category")
	ErrServiceNameEmpty = errors.New("service name cannot be empty")
)

// AllCategories returns all valid service categories
func AllCategories() []string {
	return []string{
		CategorySocial,
		CategoryProfessional,
		CategoryDevelopment,
		CategoryContent,
		CategoryPayment,
		CategoryGaming,
		CategoryCommunication,
		CategoryPortfolio,
		CategoryOther,
	}
}

// Validate validates the service model
func (s *Service) Validate() error {
	// Validate name
	if s.Name == "" {
		return ErrServiceNameEmpty
	}

	// Validate category
	validCategories := map[string]bool{
		CategorySocial:        true,
		CategoryProfessional:  true,
		CategoryDevelopment:   true,
		CategoryContent:       true,
		CategoryPayment:       true,
		CategoryGaming:        true,
		CategoryCommunication: true,
		CategoryPortfolio:     true,
		CategoryOther:         true,
	}

	if !validCategories[s.Category] {
		return ErrInvalidCategory
	}

	return nil
}

// IncrementPopularity increments the popularity counter
func (s *Service) IncrementPopularity() {
	s.Popularity++
}

// BuildURL builds a URL from the URL pattern and username
func (s *Service) BuildURL(username string) string {
	if s.URLPattern == "" {
		return ""
	}

	// Simple replacement - in production, use a proper template engine
	// Pattern example: "https://github.com/{username}"
	url := s.URLPattern
	if s.RequiresUsername && username != "" {
		// Replace {username} placeholder
		for i := 0; i < len(url)-9; i++ {
			if url[i:i+10] == "{username}" {
				url = url[:i] + username + url[i+10:]
				break
			}
		}
	}
	return url
}

// GetPlaceholder returns the placeholder text or a default
func (s *Service) GetPlaceholder() string {
	if s.PlaceholderText != "" {
		return s.PlaceholderText
	}
	if s.RequiresUsername {
		return "Enter your " + s.Name + " username"
	}
	return "Enter URL"
}
