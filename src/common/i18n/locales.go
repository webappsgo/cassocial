// Package i18n exposes the embedded locale files for use by other packages.
package i18n

import "embed"

// Locales holds all bundled JSON translation files from the locales/ directory.
//
//go:embed locales
var Locales embed.FS
