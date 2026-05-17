package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// makeI18NDir creates a temporary config dir with i18n translation files.
func makeI18NDir(t *testing.T, translations map[string]map[string]string) string {
	t.Helper()
	base := t.TempDir()
	i18nDir := filepath.Join(base, "i18n")
	if err := os.MkdirAll(i18nDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	for lang, trans := range translations {
		data, err := json.Marshal(trans)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		if err := os.WriteFile(filepath.Join(i18nDir, lang+".json"), data, 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}
	return base
}

func TestNewI18N_NoDir(t *testing.T) {
	dir := t.TempDir()
	i := NewI18N(dir, "en")
	if i == nil {
		t.Fatal("NewI18N returned nil")
	}
	// No i18n dir -> not enabled
	if i.enabled {
		t.Error("i18n should not be enabled when no i18n dir exists")
	}
}

func TestNewI18N_WithTranslations(t *testing.T) {
	configDir := makeI18NDir(t, map[string]map[string]string{
		"en": {"hello": "Hello", "bye": "Goodbye"},
		"es": {"hello": "Hola", "bye": "Adios"},
	})
	i := NewI18N(configDir, "en")
	if !i.enabled {
		t.Error("i18n should be enabled when translation files are present")
	}
}

func TestTranslate_Disabled(t *testing.T) {
	i := NewI18N(t.TempDir(), "en")
	result := i.Translate("hello", "en")
	if result != "hello" {
		t.Errorf("Translate (disabled) = %q, want key %q", result, "hello")
	}
}

func TestTranslate_Found(t *testing.T) {
	configDir := makeI18NDir(t, map[string]map[string]string{
		"en": {"hello": "Hello", "bye": "Goodbye"},
	})
	i := NewI18N(configDir, "en")

	if got := i.Translate("hello", "en"); got != "Hello" {
		t.Errorf("Translate(hello, en) = %q, want Hello", got)
	}
}

func TestTranslate_FallbackToDefault(t *testing.T) {
	configDir := makeI18NDir(t, map[string]map[string]string{
		"en": {"hello": "Hello"},
		"es": {"hello": "Hola"},
	})
	i := NewI18N(configDir, "en")

	// fr not present, should fall back to en
	if got := i.Translate("hello", "fr"); got != "Hello" {
		t.Errorf("Translate fallback = %q, want Hello", got)
	}
}

func TestTranslate_KeyNotFound(t *testing.T) {
	configDir := makeI18NDir(t, map[string]map[string]string{
		"en": {"hello": "Hello"},
	})
	i := NewI18N(configDir, "en")

	if got := i.Translate("missing_key", "en"); got != "missing_key" {
		t.Errorf("Translate missing key = %q, want key itself", got)
	}
}

func TestTranslate_NoDefaultLang(t *testing.T) {
	configDir := makeI18NDir(t, map[string]map[string]string{
		"en": {"hello": "Hello"},
	})
	i := NewI18N(configDir, "fr") // default lang not in files

	if got := i.Translate("hello", "de"); got != "hello" {
		t.Errorf("Translate no-default = %q, want key", got)
	}
}

func TestDetectLanguage_QueryParam(t *testing.T) {
	configDir := makeI18NDir(t, map[string]map[string]string{
		"en": {"k": "v"},
		"es": {"k": "v"},
	})
	i := NewI18N(configDir, "en")

	req := httptest.NewRequest("GET", "/?lang=es", nil)
	if got := i.DetectLanguage(req); got != "es" {
		t.Errorf("DetectLanguage (query) = %q, want es", got)
	}
}

func TestDetectLanguage_QueryParamUnknown(t *testing.T) {
	configDir := makeI18NDir(t, map[string]map[string]string{
		"en": {"k": "v"},
	})
	i := NewI18N(configDir, "en")

	req := httptest.NewRequest("GET", "/?lang=unknown", nil)
	if got := i.DetectLanguage(req); got != "en" {
		t.Errorf("DetectLanguage (unknown query) = %q, want en (default)", got)
	}
}

func TestDetectLanguage_Cookie(t *testing.T) {
	configDir := makeI18NDir(t, map[string]map[string]string{
		"en": {"k": "v"},
		"fr": {"k": "v"},
	})
	i := NewI18N(configDir, "en")

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "language", Value: "fr"})
	if got := i.DetectLanguage(req); got != "fr" {
		t.Errorf("DetectLanguage (cookie) = %q, want fr", got)
	}
}

func TestDetectLanguage_AcceptLanguageHeader(t *testing.T) {
	configDir := makeI18NDir(t, map[string]map[string]string{
		"en": {"k": "v"},
		"de": {"k": "v"},
	})
	i := NewI18N(configDir, "en")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en;q=0.8")
	if got := i.DetectLanguage(req); got != "de" {
		t.Errorf("DetectLanguage (accept-language) = %q, want de", got)
	}
}

func TestDetectLanguage_Default(t *testing.T) {
	configDir := makeI18NDir(t, map[string]map[string]string{
		"en": {"k": "v"},
	})
	i := NewI18N(configDir, "en")

	req := httptest.NewRequest("GET", "/", nil)
	if got := i.DetectLanguage(req); got != "en" {
		t.Errorf("DetectLanguage (default) = %q, want en", got)
	}
}

func TestGetAvailableLanguages(t *testing.T) {
	configDir := makeI18NDir(t, map[string]map[string]string{
		"en": {"k": "v"},
		"es": {"k": "v"},
		"fr": {"k": "v"},
	})
	i := NewI18N(configDir, "en")

	langs := i.GetAvailableLanguages()
	if len(langs) != 3 {
		t.Errorf("GetAvailableLanguages = %v, want 3 entries", langs)
	}
}

func TestIsRTL(t *testing.T) {
	i := NewI18N(t.TempDir(), "en")

	rtl := []string{"ar", "he", "fa", "ur"}
	for _, lang := range rtl {
		if !i.IsRTL(lang) {
			t.Errorf("IsRTL(%q) = false, want true", lang)
		}
	}

	ltr := []string{"en", "es", "fr", "de", "zh", "ja"}
	for _, lang := range ltr {
		if i.IsRTL(lang) {
			t.Errorf("IsRTL(%q) = true, want false", lang)
		}
	}
}

func TestGetDirection(t *testing.T) {
	i := NewI18N(t.TempDir(), "en")

	if got := i.GetDirection("ar"); got != "rtl" {
		t.Errorf("GetDirection(ar) = %q, want rtl", got)
	}
	if got := i.GetDirection("en"); got != "ltr" {
		t.Errorf("GetDirection(en) = %q, want ltr", got)
	}
}

func TestLoadTranslations_InvalidJSON(t *testing.T) {
	base := t.TempDir()
	i18nDir := filepath.Join(base, "i18n")
	if err := os.MkdirAll(i18nDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Write invalid JSON
	if err := os.WriteFile(filepath.Join(i18nDir, "bad.json"), []byte("not json{{{"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	// Should not panic, just skip the bad file
	i := NewI18N(base, "en")
	if i == nil {
		t.Fatal("NewI18N returned nil")
	}
}

func TestLoadTranslations_UnreadableFile(t *testing.T) {
	base := t.TempDir()
	i18nDir := filepath.Join(base, "i18n")
	if err := os.MkdirAll(i18nDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Write a JSON file then chmod it unreadable
	filePath := filepath.Join(i18nDir, "de.json")
	if err := os.WriteFile(filePath, []byte(`{"key":"val"}`), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.Chmod(filePath, 0000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { os.Chmod(filePath, 0644) })

	// Should not panic — just skip unreadable file
	i := NewI18N(base, "en")
	if i == nil {
		t.Fatal("NewI18N returned nil")
	}
}

func TestDetectLanguage_CookieUnknownLang(t *testing.T) {
	configDir := makeI18NDir(t, map[string]map[string]string{
		"en": {"k": "v"},
	})
	i := NewI18N(configDir, "en")

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "language", Value: "zz"}) // unknown lang
	if got := i.DetectLanguage(req); got != "en" {
		t.Errorf("DetectLanguage (unknown cookie) = %q, want en (default)", got)
	}
}

func TestLoadTranslations_Directory(t *testing.T) {
	base := t.TempDir()
	i18nDir := filepath.Join(base, "i18n")
	if err := os.MkdirAll(i18nDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Create a subdirectory — should be skipped, not cause panic
	if err := os.MkdirAll(filepath.Join(i18nDir, "subdir"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	i := NewI18N(base, "en")
	if i == nil {
		t.Fatal("NewI18N returned nil")
	}
}
