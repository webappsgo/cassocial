package mode

import (
	"os"
	"strings"

	"github.com/casapps/cassocial/src/config"
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
// Priority: MODE env var > default (production)
// "debug" is an alias for development (see IsDebug for the debug-flag side of the alias)
func Current() Mode {
	return FromString(os.Getenv("MODE"))
}

// IsProduction returns true if running in production mode
func IsProduction() bool {
	return Current() == Production
}

// IsDevelopment returns true if running in development mode
func IsDevelopment() bool {
	return Current() == Development
}

// IsDebug returns true if debug mode is enabled.
// Priority: DEBUG env var (truthy/falsy, explicit) > MODE=debug alias > default (false).
// An explicit DEBUG value always wins over the MODE=debug alias.
func IsDebug() bool {
	debugEnv := os.Getenv("DEBUG")
	if config.IsTruthy(debugEnv) {
		return true
	}
	if config.IsFalsy(debugEnv) {
		return false
	}

	// MODE=debug alias: expands to development + debug on
	if strings.ToLower(strings.TrimSpace(os.Getenv("MODE"))) == "debug" {
		return true
	}

	return false
}

// String returns the mode as a string
func (m Mode) String() string {
	return string(m)
}

// FromString converts a string to a Mode
func FromString(s string) Mode {
	s = strings.ToLower(strings.TrimSpace(s))

	switch s {
	case "development", "dev", "devel", "debug":
		return Development
	case "production", "prod":
		return Production
	default:
		return Production
	}
}
