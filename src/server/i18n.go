package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	commonI18n "github.com/casapps/cassocial/src/common/i18n"
)

// I18N provides internationalization support
type I18N struct {
	defaultLang string
	languages   map[string]map[string]string
	enabled     bool
}

// NewI18N creates a new I18N instance.
// Embedded locale files are always loaded first; runtime files in configDir/i18n/ override them.
func NewI18N(configDir string, defaultLang string) *I18N {
	i18n := &I18N{
		defaultLang: defaultLang,
		languages:   make(map[string]map[string]string),
		enabled:     false,
	}

	// Load embedded defaults first
	if err := i18n.loadEmbedded(); err != nil {
		fmt.Printf("Failed to load embedded translations: %v\n", err)
	}

	// Load runtime overrides (configDir/i18n/*.json) — these replace embedded entries
	if configDir != "" {
		if err := i18n.loadTranslations(configDir); err != nil {
			fmt.Printf("Failed to load translations: %v\n", err)
		}
	}

	return i18n
}

// loadEmbedded loads translations from the bundled locale files.
func (i *I18N) loadEmbedded() error {
	entries, err := commonI18n.Locales.ReadDir("locales")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		lang := strings.TrimSuffix(entry.Name(), ".json")

		data, err := commonI18n.Locales.ReadFile("locales/" + entry.Name())
		if err != nil {
			fmt.Printf("Failed to read embedded translation %s: %v\n", entry.Name(), err)
			continue
		}

		translations := make(map[string]string)
		if err := json.Unmarshal(data, &translations); err != nil {
			fmt.Printf("Failed to parse embedded translation %s: %v\n", entry.Name(), err)
			continue
		}

		i.languages[lang] = translations
		i.enabled = true
	}

	return nil
}

// loadTranslations loads translation files from configDir/i18n/, overriding embedded defaults.
func (i *I18N) loadTranslations(configDir string) error {
	i18nDir := filepath.Join(configDir, "i18n")

	// Check if i18n directory exists
	if _, err := os.Stat(i18nDir); os.IsNotExist(err) {
		return nil
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

// Translate translates a key to the specified language.
// If the key is missing from the requested language it falls back to the default language.
// If the key is still not found the key itself is returned.
func (i *I18N) Translate(key, lang string) string {
	if !i.enabled {
		return key
	}

	// Try the requested language first
	if translations, exists := i.languages[lang]; exists {
		if translated, ok := translations[key]; ok {
			return translated
		}
	}

	// Fall back to the default language
	if lang != i.defaultLang {
		if translations, exists := i.languages[i.defaultLang]; exists {
			if translated, ok := translations[key]; ok {
				return translated
			}
		}
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
		"ar": true,
		"he": true,
		"fa": true,
		"ur": true,
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
