package model

import (
	"errors"
	"time"
)

// Theme represents a visual theme for profiles
type Theme struct {
	ID             string    `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	BackgroundColor string   `json:"background_color" db:"background_color"`
	TextColor      string    `json:"text_color" db:"text_color"`
	LinkBackground string    `json:"link_background" db:"link_background"`
	LinkHover      string    `json:"link_hover" db:"link_hover"`
	LinkText       string    `json:"link_text" db:"link_text"`
	BorderRadius   string    `json:"border_radius" db:"border_radius"`
	FontFamily     string    `json:"font_family" db:"font_family"`
	IsPremium      bool      `json:"is_premium" db:"is_premium"`
	PreviewImage   string    `json:"preview_image,omitempty" db:"preview_image"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

var (
	ErrThemeNameEmpty = errors.New("theme name cannot be empty")
	ErrInvalidColor   = errors.New("invalid color format")
)

// Default theme ID (as per spec)
const DefaultThemeID = "00000000-0000-0000-0000-000000000001"

// Validate validates the theme model
func (t *Theme) Validate() error {
	// Validate name
	if t.Name == "" {
		return ErrThemeNameEmpty
	}

	// Basic color validation (hex colors)
	colors := []string{
		t.BackgroundColor,
		t.TextColor,
		t.LinkBackground,
		t.LinkHover,
		t.LinkText,
	}

	for _, color := range colors {
		if color != "" && !isValidHexColor(color) {
			return ErrInvalidColor
		}
	}

	return nil
}

// GetDefaultTheme returns the default dark theme
func GetDefaultTheme() *Theme {
	return &Theme{
		ID:              DefaultThemeID,
		Name:            "Default Dark",
		BackgroundColor: "#1a1a1a",
		TextColor:       "#ffffff",
		LinkBackground:  "#2a2a2a",
		LinkHover:       "#3a3a3a",
		LinkText:        "#ffffff",
		BorderRadius:    "12px",
		FontFamily:      "Inter, system-ui, sans-serif",
		IsPremium:       false,
		CreatedAt:       time.Now(),
	}
}

// GetLightTheme returns a light theme preset
func GetLightTheme() *Theme {
	return &Theme{
		Name:            "Light",
		BackgroundColor: "#ffffff",
		TextColor:       "#212529",
		LinkBackground:  "#ffffff",
		LinkHover:       "#f8f9fa",
		LinkText:        "#212529",
		BorderRadius:    "12px",
		FontFamily:      "Inter, system-ui, sans-serif",
		IsPremium:       false,
	}
}

// GetDraculaTheme returns a Dracula-inspired theme
func GetDraculaTheme() *Theme {
	return &Theme{
		Name:            "Dracula",
		BackgroundColor: "#282a36",
		TextColor:       "#f8f8f2",
		LinkBackground:  "#44475a",
		LinkHover:       "#6272a4",
		LinkText:        "#f8f8f2",
		BorderRadius:    "12px",
		FontFamily:      "Inter, system-ui, sans-serif",
		IsPremium:       false,
	}
}

// GetOceanTheme returns an ocean-inspired theme
func GetOceanTheme() *Theme {
	return &Theme{
		Name:            "Ocean",
		BackgroundColor: "#0a192f",
		TextColor:       "#8892b0",
		LinkBackground:  "#112240",
		LinkHover:       "#233554",
		LinkText:        "#ccd6f6",
		BorderRadius:    "12px",
		FontFamily:      "Inter, system-ui, sans-serif",
		IsPremium:       true,
	}
}

// GetSunsetTheme returns a sunset-inspired theme
func GetSunsetTheme() *Theme {
	return &Theme{
		Name:            "Sunset",
		BackgroundColor: "#1a1a2e",
		TextColor:       "#eaeaea",
		LinkBackground:  "#16213e",
		LinkHover:       "#0f3460",
		LinkText:        "#e94560",
		BorderRadius:    "12px",
		FontFamily:      "Inter, system-ui, sans-serif",
		IsPremium:       true,
	}
}

// isValidHexColor validates hex color format
func isValidHexColor(color string) bool {
	if len(color) != 4 && len(color) != 7 {
		return false
	}
	if color[0] != '#' {
		return false
	}
	for i := 1; i < len(color); i++ {
		c := color[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// ApplyToProfile applies theme settings to a profile theme
func (t *Theme) ApplyToProfile() map[string]string {
	return map[string]string{
		"--bg-primary":     t.BackgroundColor,
		"--text-primary":   t.TextColor,
		"--link-bg":        t.LinkBackground,
		"--link-hover":     t.LinkHover,
		"--link-text":      t.LinkText,
		"--radius":         t.BorderRadius,
		"--font-family":    t.FontFamily,
	}
}
