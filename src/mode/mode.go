package mode

import (
	"os"
	"strings"
)

// Mode represents the application mode
type Mode string

const (
	// Production mode - strict security, minimal logging, caching enabled
	Production Mode = "production"
	// Development mode - relaxed security, verbose logging, no caching
	Development Mode = "development"
)

// Current returns the current application mode
// Priority: MODE env var > config file setting > default (production)
func Current() Mode {
	modeStr := strings.ToLower(strings.TrimSpace(os.Getenv("MODE")))

	switch modeStr {
	case "development", "dev", "debug":
		return Development
	case "production", "prod":
		return Production
	default:
		// Default to production for safety
		return Production
	}
}

// IsProduction returns true if running in production mode
func IsProduction() bool {
	return Current() == Production
}

// IsDevelopment returns true if running in development mode
func IsDevelopment() bool {
	return Current() == Development
}

// IsDebug returns true if debug mode is enabled
// Debug mode can be enabled in any mode via DEBUG env var or --debug flag
func IsDebug() bool {
	debug := strings.ToLower(strings.TrimSpace(os.Getenv("DEBUG")))

	// Truthy values
	truthy := map[string]bool{
		"1":      true,
		"yes":    true,
		"true":   true,
		"on":     true,
		"enable": true,
		"enabled": true,
		"y":      true,
		"t":      true,
	}

	return truthy[debug]
}

// String returns the mode as a string
func (m Mode) String() string {
	return string(m)
}

// FromString converts a string to a Mode
func FromString(s string) Mode {
	s = strings.ToLower(strings.TrimSpace(s))

	switch s {
	case "development", "dev", "debug":
		return Development
	case "production", "prod":
		return Production
	default:
		return Production
	}
}
