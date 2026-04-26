package config

import "strings"

// ParseBool parses a boolean value from a string
// Accepts all truthy/falsy values per TEMPLATE.md specification
func ParseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))

	// Truthy values
	truthy := map[string]bool{
		"1":           true,
		"yes":         true,
		"true":        true,
		"on":          true,
		"enable":      true,
		"enabled":     true,
		"y":           true,
		"t":           true,
		"yep":         true,
		"yup":         true,
		"yeah":        true,
		"aye":         true,
		"si":          true,
		"oui":         true,
		"da":          true,
		"hai":         true,
		"affirmative": true,
		"accept":      true,
		"allow":       true,
		"totally":     true,
	}

	return truthy[s]
}
