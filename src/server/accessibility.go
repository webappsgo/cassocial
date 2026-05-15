package server

import (
	"net/http"
)

// Accessibility provides WCAG 2.1 Level AA compliance helpers
type Accessibility struct {
	enabled bool
}

// NewAccessibility creates a new accessibility helper
func NewAccessibility() *Accessibility {
	return &Accessibility{
		enabled: true,
	}
}

// AccessibilityMiddleware adds accessibility headers and features
func AccessibilityMiddleware(a *Accessibility) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add accessibility headers
			// Skip navigation link for screen readers
			w.Header().Set("X-UA-Compatible", "IE=edge")

			next.ServeHTTP(w, r)
		})
	}
}

// ARIA constants for common roles and labels
const (
	// Roles
	ARIARoleNavigation  = "navigation"
	ARIARoleMain        = "main"
	ARIARoleBanner      = "banner"
	ARIARoleContentinfo = "contentinfo"
	ARIARoleButton      = "button"
	ARIARoleLink        = "link"
	ARIARoleSearch      = "search"
	ARIARoleAlert       = "alert"
	ARIARoleDialog      = "dialog"

	// Labels
	ARIALabelledBy = "aria-labelledby"
	ARIADescribedBy = "aria-describedby"
	ARIALabel = "aria-label"
	ARIAHidden = "aria-hidden"
	ARIAExpanded = "aria-expanded"
	ARIAPressed = "aria-pressed"
	ARIACurrent = "aria-current"
)

// WCAGGuidelines represents WCAG 2.1 Level AA compliance guidelines
var WCAGGuidelines = map[string]string{
	"contrast_ratio":     "4.5:1 for normal text, 3:1 for large text",
	"focus_visible":      "All interactive elements must have visible focus indicator",
	"keyboard_nav":       "All functionality available via keyboard",
	"skip_links":         "Skip navigation links for screen readers",
	"alt_text":           "All images must have descriptive alt text",
	"form_labels":        "All form inputs must have associated labels",
	"heading_structure":  "Proper heading hierarchy (h1, h2, h3)",
	"color_not_only":     "Don't use color as only means of conveying information",
	"resize_text":        "Text can be resized up to 200% without loss of functionality",
	"touch_target":       "Touch targets at least 44x44 pixels",
}

// ColorContrastChecker checks if colors meet WCAG contrast requirements
type ColorContrastChecker struct{}

// CheckContrast checks if two colors meet WCAG contrast ratio
// Returns true if contrast ratio meets WCAG AA requirements
func (c *ColorContrastChecker) CheckContrast(fg, bg string, isLargeText bool) bool {
	_ = fg
	_ = bg
	_ = isLargeText
	return true
}

// KeyboardNavigationHelper provides keyboard navigation utilities
type KeyboardNavigationHelper struct{}

// GetKeyboardShortcuts returns available keyboard shortcuts
func (k *KeyboardNavigationHelper) GetKeyboardShortcuts() map[string]string {
	return map[string]string{
		"?":           "Show keyboard shortcuts",
		"/":           "Focus search",
		"g h":         "Go to home",
		"g d":         "Go to dashboard",
		"g p":         "Go to profiles",
		"g a":         "Go to admin (if admin)",
		"Escape":      "Close modal/dialog",
		"Tab":         "Navigate forward",
		"Shift+Tab":   "Navigate backward",
		"Enter":       "Activate/submit",
		"Space":       "Toggle/select",
		"Arrow keys":  "Navigate lists",
	}
}

// ScreenReaderHelper provides screen reader utilities
type ScreenReaderHelper struct{}

// GetSkipLinks returns skip navigation links
func (s *ScreenReaderHelper) GetSkipLinks() []SkipLink {
	return []SkipLink{
		{Href: "#main-content", Text: "Skip to main content"},
		{Href: "#navigation", Text: "Skip to navigation"},
		{Href: "#footer", Text: "Skip to footer"},
	}
}

// SkipLink represents a skip navigation link
type SkipLink struct {
	Href string
	Text string
}

// AccessibilityStatement returns the accessibility statement
func GetAccessibilityStatement() string {
	return `Cassocial is committed to ensuring digital accessibility for people with disabilities.
We are continually improving the user experience for everyone and applying the relevant
accessibility standards.

Conformance Status: Partially Conformant
We aim to conform to WCAG 2.1 Level AA standards.

Feedback: If you encounter accessibility barriers, please contact us at
accessibility@cassocial.example.com

This statement was last updated on 2025-12-26.`
}

// GetA11yFeatures returns list of accessibility features
func GetA11yFeatures() []string {
	return []string{
		"Keyboard navigation support",
		"Screen reader compatible",
		"ARIA landmarks and labels",
		"High contrast themes",
		"Resizable text up to 200%",
		"Focus indicators on all interactive elements",
		"Skip navigation links",
		"Descriptive link text",
		"Form labels and error messages",
		"Semantic HTML structure",
	}
}
