package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewAccessibility verifies that NewAccessibility returns a non-nil instance.
func TestNewAccessibility(t *testing.T) {
	a := NewAccessibility()
	if a == nil {
		t.Fatal("NewAccessibility returned nil")
	}
	if !a.enabled {
		t.Error("NewAccessibility should have enabled=true")
	}
}

// TestAccessibilityMiddleware_SetsHeader verifies that the middleware sets the X-UA-Compatible header.
func TestAccessibilityMiddleware_SetsHeader(t *testing.T) {
	a := NewAccessibility()
	mw := AccessibilityMiddleware(a)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	mw(next).ServeHTTP(rr, req)

	if !called {
		t.Error("AccessibilityMiddleware did not call the next handler")
	}
	if got := rr.Header().Get("X-UA-Compatible"); got != "IE=edge" {
		t.Errorf("X-UA-Compatible = %q, want %q", got, "IE=edge")
	}
}

// TestCheckContrast_ReturnsTrue verifies that CheckContrast always returns true (stub implementation).
func TestCheckContrast_ReturnsTrue(t *testing.T) {
	c := &ColorContrastChecker{}
	if !c.CheckContrast("#000000", "#ffffff", false) {
		t.Error("CheckContrast should return true")
	}
	if !c.CheckContrast("#333333", "#eeeeee", true) {
		t.Error("CheckContrast should return true for large text")
	}
}

// TestGetKeyboardShortcuts_NonEmpty verifies that keyboard shortcuts are returned.
func TestGetKeyboardShortcuts_NonEmpty(t *testing.T) {
	k := &KeyboardNavigationHelper{}
	shortcuts := k.GetKeyboardShortcuts()
	if len(shortcuts) == 0 {
		t.Error("GetKeyboardShortcuts should return non-empty map")
	}
	if _, ok := shortcuts["?"]; !ok {
		t.Error("GetKeyboardShortcuts should include '?' shortcut")
	}
}

// TestGetSkipLinks_NonEmpty verifies that skip links are returned with correct fields.
func TestGetSkipLinks_NonEmpty(t *testing.T) {
	s := &ScreenReaderHelper{}
	links := s.GetSkipLinks()
	if len(links) == 0 {
		t.Error("GetSkipLinks should return non-empty slice")
	}
	for _, l := range links {
		if l.Href == "" {
			t.Error("SkipLink.Href is empty")
		}
		if l.Text == "" {
			t.Error("SkipLink.Text is empty")
		}
	}
}

// TestGetSkipLinks_ContainsMainContent verifies that there is a skip-to-main-content link.
func TestGetSkipLinks_ContainsMainContent(t *testing.T) {
	s := &ScreenReaderHelper{}
	links := s.GetSkipLinks()

	found := false
	for _, l := range links {
		if l.Href == "#main-content" {
			found = true
			break
		}
	}
	if !found {
		t.Error("GetSkipLinks should include a #main-content link")
	}
}

// TestGetAccessibilityStatement_NonEmpty verifies that the statement is non-empty.
func TestGetAccessibilityStatement_NonEmpty(t *testing.T) {
	stmt := GetAccessibilityStatement()
	if stmt == "" {
		t.Error("GetAccessibilityStatement returned empty string")
	}
}

// TestGetA11yFeatures_NonEmpty verifies that the features list is non-empty.
func TestGetA11yFeatures_NonEmpty(t *testing.T) {
	features := GetA11yFeatures()
	if len(features) == 0 {
		t.Error("GetA11yFeatures should return non-empty slice")
	}
}

// TestGetA11yFeatures_ContainsKeyboard verifies that keyboard navigation is listed.
func TestGetA11yFeatures_ContainsKeyboard(t *testing.T) {
	features := GetA11yFeatures()
	found := false
	for _, f := range features {
		if f == "Keyboard navigation support" {
			found = true
			break
		}
	}
	if !found {
		t.Error("GetA11yFeatures should include 'Keyboard navigation support'")
	}
}

// TestWCAGGuidelines_NonEmpty verifies that the WCAG guidelines map has entries.
func TestWCAGGuidelines_NonEmpty(t *testing.T) {
	if len(WCAGGuidelines) == 0 {
		t.Error("WCAGGuidelines should be non-empty")
	}
}

// TestARIAConstants verifies that ARIA role constants are non-empty.
func TestARIAConstants(t *testing.T) {
	constants := []struct {
		name  string
		value string
	}{
		{"ARIARoleNavigation", ARIARoleNavigation},
		{"ARIARoleMain", ARIARoleMain},
		{"ARIARoleBanner", ARIARoleBanner},
		{"ARIARoleContentinfo", ARIARoleContentinfo},
		{"ARIARoleButton", ARIARoleButton},
		{"ARIARoleLink", ARIARoleLink},
		{"ARIARoleSearch", ARIARoleSearch},
		{"ARIARoleAlert", ARIARoleAlert},
		{"ARIARoleDialog", ARIARoleDialog},
		{"ARIALabelledBy", ARIALabelledBy},
		{"ARIADescribedBy", ARIADescribedBy},
		{"ARIALabel", ARIALabel},
		{"ARIAHidden", ARIAHidden},
		{"ARIAExpanded", ARIAExpanded},
		{"ARIAPressed", ARIAPressed},
		{"ARIACurrent", ARIACurrent},
	}

	for _, c := range constants {
		if c.value == "" {
			t.Errorf("ARIA constant %s is empty", c.name)
		}
	}
}
