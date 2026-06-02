package server

import "embed"

// TemplateFS holds all HTML templates embedded into the binary at build time
//
//go:embed template
var TemplateFS embed.FS

// StaticFS holds all static assets (CSS, JS, images) embedded at build time
//
//go:embed static
var StaticFS embed.FS
