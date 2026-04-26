package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// I18N provides internationalization support
type I18N struct {
	defaultLang string
	languages   map[string]map[string]string
	enabled     bool
}

// NewI18N creates a new I18N instance
func NewI18N(configDir string, defaultLang string) *I18N {
	i18n := &I18N{
		defaultLang: defaultLang,
		languages:   make(map[string]map[string]string),
		enabled:     false,
	}

	// Load translations
	if err := i18n.loadTranslations(configDir); err != nil {
		// Log but don't fail - app works without translations
		fmt.Printf("Failed to load translations: %v\n", err)
	}

	return i18n
}

// loadTranslations loads translation files
func (i *I18N) loadTranslations(configDir string) error {
	i18nDir := filepath.Join(configDir, "i18n")

	// Check if i18n directory exists
	if _, err := os.Stat(i18nDir); os.IsNotExist(err) {
		return nil // No translations configured
	}

	// Read translation files
	entries, err := os.ReadDir(i18nDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// Extract language code from filename (e.g., en.json -> en)
		lang := strings.TrimSuffix(entry.Name(), ".json")

		// Load translation file
		filePath := filepath.Join(i18nDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Failed to read translation file %s: %v\n", entry.Name(), err)
			continue
		}

		// Parse JSON
		translations := make(map[string]string)
		if err := json.Unmarshal(data, &translations); err != nil {
			fmt.Printf("Failed to parse translation file %s: %v\n", entry.Name(), err)
			continue
		}

		i.languages[lang] = translations
		i.enabled = true
	}

	return nil
}

// Translate translates a key to the specified language
func (i *I18N) Translate(key, lang string) string {
	if !i.enabled {
		return key
	}

	// Get translations for language
	translations, exists := i.languages[lang]
	if !exists {
		// Fallback to default language
		translations, exists = i.languages[i.defaultLang]
		if !exists {
			return key
		}
	}

	// Get translation
	if translated, exists := translations[key]; exists {
		return translated
	}

	return key
}

// DetectLanguage detects user's preferred language from request
func (i *I18N) DetectLanguage(r *http.Request) string {
	// Check query parameter
	if lang := r.URL.Query().Get("lang"); lang != "" {
		if _, exists := i.languages[lang]; exists {
			return lang
		}
	}

	// Check cookie
	if cookie, err := r.Cookie("language"); err == nil {
		if _, exists := i.languages[cookie.Value]; exists {
			return cookie.Value
		}
	}

	// Check Accept-Language header
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang != "" {
		// Parse Accept-Language header (simplified)
		langs := strings.Split(acceptLang, ",")
		for _, lang := range langs {
			// Extract language code (ignore quality values)
			code := strings.Split(lang, ";")[0]
			code = strings.TrimSpace(code)
			code = strings.Split(code, "-")[0] // en-US -> en

			if _, exists := i.languages[code]; exists {
				return code
			}
		}
	}

	return i.defaultLang
}

// GetAvailableLanguages returns list of available languages
func (i *I18N) GetAvailableLanguages() []string {
	langs := make([]string, 0, len(i.languages))
	for lang := range i.languages {
		langs = append(langs, lang)
	}
	return langs
}

// IsRTL checks if a language is right-to-left
func (i *I18N) IsRTL(lang string) bool {
	rtlLanguages := map[string]bool{
		"ar": true, // Arabic
		"he": true, // Hebrew
		"fa": true, // Persian
		"ur": true, // Urdu
	}

	return rtlLanguages[lang]
}

// GetDirection returns text direction for a language
func (i *I18N) GetDirection(lang string) string {
	if i.IsRTL(lang) {
		return "rtl"
	}
	return "ltr"
}
