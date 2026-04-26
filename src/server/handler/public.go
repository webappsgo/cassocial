package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// PublicHandler handles public-facing pages
type PublicHandler struct {
	config *config.Config
	db     *store.DB
}

// NewPublicHandler creates a new public handler
func NewPublicHandler(cfg *config.Config, db *store.DB) *PublicHandler {
	return &PublicHandler{
		config: cfg,
		db:     db,
	}
}

// HandleHomepage handles the homepage
func (h *PublicHandler) HandleHomepage(w http.ResponseWriter, r *http.Request) {
	// If path is not root, might be a profile slug
	if r.URL.Path != "/" {
		h.HandleProfilePage(w, r)
		return
	}

	// Render homepage
	data := map[string]interface{}{
		"site_name":        h.config.Cassocial.SiteName,
		"site_description": h.config.Cassocial.SiteDescription,
		"allow_registration": h.config.Cassocial.AllowRegistration,
	}

	// TODO: Render homepage template
	h.renderJSON(w, http.StatusOK, data)
}

// HandleProfilePage renders a public profile page
func (h *PublicHandler) HandleProfilePage(w http.ResponseWriter, r *http.Request) {
	// Extract slug from URL path
	slug := strings.TrimPrefix(r.URL.Path, "/")
	slug = strings.Split(slug, "/")[0] // Get first part of path

	if slug == "" || slug == "api" || slug == "admin" || slug == "healthz" {
		http.NotFound(w, r)
		return
	}

	// TODO: Get profile from database by slug
	// TODO: Check if profile is public or password-protected
	// TODO: Check if profile exists

	// Get client info for analytics
	ip := getClientIP(r)
	userAgent := r.UserAgent()
	referer := r.Referer()

	// TODO: Track page view
	_ = ip
	_ = userAgent
	_ = referer

	// TODO: Get links for this profile
	links := []interface{}{}

	// Profile data for rendering
	profileData := map[string]interface{}{
		"slug":        slug,
		"title":       "Profile Title",
		"description": "Profile description",
		"avatar":      "",
		"theme":       "dark",
		"links":       links,
		"seo": map[string]string{
			"title":       "Profile Title",
			"description": "Profile description",
			"image":       "",
			"url":         r.URL.String(),
		},
	}

	// TODO: Render profile template with theme
	h.renderJSON(w, http.StatusOK, profileData)
}

// HandleProfilePreview renders a profile preview (for editing)
func (h *PublicHandler) HandleProfilePreview(w http.ResponseWriter, r *http.Request) {
	// TODO: Get user from session
	// TODO: Get profile data from request
	// TODO: Render preview without saving

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status": "preview",
	})
}

// HandleSitemap generates a sitemap.xml
func (h *PublicHandler) HandleSitemap(w http.ResponseWriter, r *http.Request) {
	// TODO: Get all public profiles
	// TODO: Generate XML sitemap

	sitemap := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <url>
        <loc>https://example.com/</loc>
        <changefreq>daily</changefreq>
        <priority>1.0</priority>
    </url>
</urlset>`

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(sitemap))
}

// HandleRobotsTxt generates robots.txt
func (h *PublicHandler) HandleRobotsTxt(w http.ResponseWriter, r *http.Request) {
	robots := `User-agent: *
Allow: /
Disallow: /admin
Disallow: /api

Sitemap: https://` + r.Host + `/sitemap.xml`

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(robots))
}

// HandleSecurityTxt generates security.txt
func (h *PublicHandler) HandleSecurityTxt(w http.ResponseWriter, r *http.Request) {
	security := `Contact: https://github.com/casapps/cassocial/security
Preferred-Languages: en
Canonical: https://` + r.Host + `/.well-known/security.txt`

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(security))
}

// renderJSON renders a JSON response
func (h *PublicHandler) renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
