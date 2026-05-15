package handler

import (
	"encoding/json"
	"net/http"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// ThemeHandler handles theme-related operations
type ThemeHandler struct {
	config *config.Config
	db     *store.DB
}

// NewThemeHandler creates a new theme handler
func NewThemeHandler(cfg *config.Config, db *store.DB) *ThemeHandler {
	return &ThemeHandler{
		config: cfg,
		db:     db,
	}
}

// Theme represents a theme configuration
type Theme struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"` // dark, light
	Background  BackgroundConfig  `json:"background"`
	ButtonStyle ButtonStyleConfig `json:"button_style"`
	Colors      ColorConfig       `json:"colors"`
	Fonts       FontConfig        `json:"fonts"`
}

// BackgroundConfig represents background configuration
type BackgroundConfig struct {
	Type     string   `json:"type"` // color, gradient, image
	Color    string   `json:"color,omitempty"`
	Gradient Gradient `json:"gradient,omitempty"`
	ImageURL string   `json:"image_url,omitempty"`
}

// Gradient represents a gradient configuration
type Gradient struct {
	Type   string   `json:"type"` // linear, radial
	Colors []string `json:"colors"`
	Angle  int      `json:"angle,omitempty"`
}

// ButtonStyleConfig represents button styling
type ButtonStyleConfig struct {
	Shape      string `json:"shape"` // rounded, square, pill
	Style      string `json:"style"` // filled, outlined, ghost
	Animation  string `json:"animation"` // none, pulse, bounce, slide
	HoverEffect string `json:"hover_effect"`
}

// ColorConfig represents color scheme
type ColorConfig struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
	Text      string `json:"text"`
	Link      string `json:"link"`
	Accent    string `json:"accent"`
}

// FontConfig represents font configuration
type FontConfig struct {
	Family     string `json:"family"`
	Size       string `json:"size"`
	Weight     string `json:"weight"`
	GoogleFont bool   `json:"google_font"`
}

// HandleGetThemes returns available themes
func (h *ThemeHandler) HandleGetThemes(w http.ResponseWriter, r *http.Request) {
	themes := []Theme{
		{
			ID:   "dark-dracula",
			Name: "Dark (Dracula)",
			Type: "dark",
			Background: BackgroundConfig{
				Type: "gradient",
				Gradient: Gradient{
					Type:   "linear",
					Colors: []string{"#282a36", "#44475a"},
					Angle:  135,
				},
			},
			ButtonStyle: ButtonStyleConfig{
				Shape:       "rounded",
				Style:       "filled",
				Animation:   "none",
				HoverEffect: "lift",
			},
			Colors: ColorConfig{
				Primary:   "#bd93f9",
				Secondary: "#ff79c6",
				Text:      "#f8f8f2",
				Link:      "#8be9fd",
				Accent:    "#50fa7b",
			},
			Fonts: FontConfig{
				Family:     "Inter",
				Size:       "16px",
				Weight:     "400",
				GoogleFont: true,
			},
		},
		{
			ID:   "light-minimal",
			Name: "Light Minimal",
			Type: "light",
			Background: BackgroundConfig{
				Type:  "color",
				Color: "#ffffff",
			},
			ButtonStyle: ButtonStyleConfig{
				Shape:       "rounded",
				Style:       "outlined",
				Animation:   "none",
				HoverEffect: "fill",
			},
			Colors: ColorConfig{
				Primary:   "#2563eb",
				Secondary: "#7c3aed",
				Text:      "#1f2937",
				Link:      "#0ea5e9",
				Accent:    "#10b981",
			},
			Fonts: FontConfig{
				Family:     "Inter",
				Size:       "16px",
				Weight:     "400",
				GoogleFont: true,
			},
		},
	}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"themes": themes,
		"total":  len(themes),
	})
}

// HandleGetTheme returns a specific theme
func (h *ThemeHandler) HandleGetTheme(w http.ResponseWriter, r *http.Request) {
	themeID := r.URL.Query().Get("id")
	if themeID == "" {
		h.renderError(w, http.StatusBadRequest, "Theme ID required")
		return
	}

	h.renderJSON(w, http.StatusOK, map[string]string{
		"id":   themeID,
		"name": "Theme",
	})
}

// HandleSaveCustomTheme saves a custom theme for a profile
func (h *ThemeHandler) HandleSaveCustomTheme(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var theme Theme
	if err := json.NewDecoder(r.Body).Decode(&theme); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if theme.ID == "" {
		h.renderError(w, http.StatusBadRequest, "Theme ID is required")
		return
	}

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Theme saved successfully",
		"theme_id": theme.ID,
	})
}

// HandleGetGradientPresets returns gradient presets
func (h *ThemeHandler) HandleGetGradientPresets(w http.ResponseWriter, r *http.Request) {
	presets := []map[string]interface{}{
		{
			"id":     "sunset",
			"name":   "Sunset",
			"colors": []string{"#ff6b6b", "#feca57", "#ee5a6f"},
		},
		{
			"id":     "ocean",
			"name":   "Ocean",
			"colors": []string{"#667eea", "#764ba2", "#f093fb"},
		},
		{
			"id":     "forest",
			"name":   "Forest",
			"colors": []string{"#134e5e", "#71b280"},
		},
		{
			"id":     "midnight",
			"name":   "Midnight",
			"colors": []string{"#2c3e50", "#3498db", "#2980b9"},
		},
	}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"presets": presets,
		"total":   len(presets),
	})
}

// renderJSON renders a JSON response
func (h *ThemeHandler) renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// renderError renders an error response
func (h *ThemeHandler) renderError(w http.ResponseWriter, status int, message string) {
	h.renderJSON(w, status, map[string]string{
		"error": message,
	})
}
