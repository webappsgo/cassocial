package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server"
	"github.com/casapps/cassocial/src/server/store"
)

// Router handles HTTP routing for the application
type Router struct {
	mux               *http.ServeMux
	middleware        *Middleware
	db                *store.DB
	cfg               *config.Config
	authHandlers      *AuthHandlers
	profileHandlers   *ProfileHandlers
	linkHandlers      *LinkHandlers
	serviceHandlers   *ServiceHandlers
	analyticsHandlers *AnalyticsHandlers
	adminHandlers     *AdminHandlers
	publicHandlers    *PublicHandlers
	setupHandler      *SetupHandler
	shortlinkHandler  *ShortlinkHandler
	i18n              *server.I18N
	startTime         time.Time
	routes            []string
}

// hf registers a pattern+handler function on the mux and records the pattern
// for introspection via the /debug/routes endpoint (AI.md PART 6).
func (rt *Router) hf(pattern string, handler http.HandlerFunc) {
	rt.routes = append(rt.routes, pattern)
	rt.mux.HandleFunc(pattern, handler)
}

// h registers a pattern+http.Handler on the mux and records the pattern for
// introspection via the /debug/routes endpoint (AI.md PART 6).
func (rt *Router) h(pattern string, handler http.Handler) {
	rt.routes = append(rt.routes, pattern)
	rt.mux.Handle(pattern, handler)
}

// NewRouter creates a new Router instance with all handlers wired up. lang
// is the resolved --lang/LANG preference (AI.md PART 8, PART 31); pass ""
// to fall back to the server's configured default language.
// Templates are parsed at startup; any failure is fatal.
func NewRouter(db *store.DB, authService *Auth, cfg *config.Config, lang string) *Router {
	mux := http.NewServeMux()
	middleware := server.NewMiddleware(authService)

	rt := &Router{
		mux:               mux,
		middleware:        middleware,
		db:                db,
		cfg:               cfg,
		authHandlers:      NewAuthHandlers(authService, db),
		profileHandlers:   NewProfileHandlers(db),
		linkHandlers:      NewLinkHandlers(db),
		serviceHandlers:   NewServiceHandlers(db),
		analyticsHandlers: NewAnalyticsHandlers(db),
		adminHandlers:     NewAdminHandlers(db, authService),
		publicHandlers:    NewPublicHandlers(db),
		setupHandler:      NewSetupHandler(cfg, db),
		shortlinkHandler:  NewShortlinkHandler(cfg, db),
		i18n:              server.NewI18N(cfg.ConfigDir, lang),
		startTime:         time.Now(),
	}

	if err := initTemplates(); err != nil {
		log.Printf("Warning: failed to parse templates: %v", err)
	}

	return rt
}

// SetupRoutes configures all routes and returns the final http.Handler.
func (rt *Router) SetupRoutes() http.Handler {
	// ── Static assets ──────────────────────────────────────────────────────────
	rt.h("GET /static/", staticHandler())

	// ── Infrastructure endpoints ───────────────────────────────────────────────
	// Health checks (spec: /server/healthz HTML, /api/v1/server/healthz JSON)
	rt.hf("GET /server/healthz", rt.healthzHTML)
	rt.hf("GET /api/v1/server/healthz", rt.healthzJSON)
	// Legacy aliases kept for backward compatibility
	rt.hf("GET /health", rt.healthzJSON)
	rt.hf("GET /health/ready", rt.readinessCheck)
	rt.hf("GET /health/live", rt.livenessCheck)

	// Prometheus metrics (internal only — no public exposure)
	rt.hf("GET /metrics", rt.metrics)

	// PWA
	rt.hf("GET /manifest.json", func(w http.ResponseWriter, r *http.Request) {
		rt.pwaInstance().ServeManifest(w, r)
	})
	rt.hf("GET /service-worker.js", func(w http.ResponseWriter, r *http.Request) {
		rt.pwaInstance().ServeServiceWorker(w, r)
	})

	// Crawlers / discovery
	rt.hf("GET /robots.txt", rt.robotsTxt)
	rt.hf("GET /sitemap.xml", rt.sitemapXML)
	rt.hf("GET /.well-known/security.txt", rt.securityTxt)

	// ── Setup wizard API ───────────────────────────────────────────────────────
	rt.hf("GET /api/setup/status", rt.setupHandler.HandleSetupStatus)
	rt.hf("GET /api/setup/welcome", rt.setupHandler.HandleSetupWelcome)
	rt.hf("GET /api/setup/basic", rt.setupHandler.HandleSetupBasic)
	rt.hf("POST /api/setup/basic", rt.setupHandler.HandleSetupBasic)
	rt.hf("GET /api/setup/domain", rt.setupHandler.HandleSetupDomain)
	rt.hf("POST /api/setup/domain", rt.setupHandler.HandleSetupDomain)
	rt.hf("GET /api/setup/email", rt.setupHandler.HandleSetupEmail)
	rt.hf("POST /api/setup/email", rt.setupHandler.HandleSetupEmail)
	rt.hf("GET /api/setup/features", rt.setupHandler.HandleSetupFeatures)
	rt.hf("POST /api/setup/features", rt.setupHandler.HandleSetupFeatures)
	rt.hf("GET /api/setup/database", rt.setupHandler.HandleSetupDatabase)
	rt.hf("POST /api/setup/database", rt.setupHandler.HandleSetupDatabase)
	rt.hf("POST /api/setup/complete", rt.setupHandler.HandleSetupComplete)

	// ── Authentication (/api/auth/ AND /api/v1/auth/ — both accepted) ─────────
	for _, prefix := range []string{"/api/auth", "/api/v1/auth"} {
		rt.hf("POST "+prefix+"/register", rt.authHandlers.Register)
		rt.hf("POST "+prefix+"/login", rt.authHandlers.Login)
		rt.hf("POST "+prefix+"/login/2fa", rt.authHandlers.LoginWith2FA)
		rt.hf("POST "+prefix+"/forgot-password", rt.authHandlers.ForgotPassword)
		rt.hf("POST "+prefix+"/reset-password", rt.authHandlers.ResetPassword)
		rt.hf("GET "+prefix+"/verify-email/{token}", rt.authHandlers.VerifyEmail)
		rt.h("POST "+prefix+"/logout",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.authHandlers.Logout)))
		rt.h("POST "+prefix+"/refresh",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.authHandlers.RefreshToken)))
		rt.h("POST "+prefix+"/2fa/enable",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.authHandlers.Enable2FA)))
		rt.h("POST "+prefix+"/2fa/verify",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.authHandlers.Verify2FA)))
		rt.h("POST "+prefix+"/2fa/disable",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.authHandlers.Disable2FA)))
	}

	// ── Profiles (authenticated CRUD — /api/profiles only) ───────────────────
	// NOTE: /api/v1/profiles/{username} is the public read endpoint registered below.
	// Registering /{id} wildcard under /api/v1/profiles too would conflict with
	// the public /api/v1/profiles/{username} pattern.
	for _, prefix := range []string{"/api/profiles"} {
		rt.h("GET "+prefix,
			rt.middleware.RequireAuth(http.HandlerFunc(rt.profileHandlers.ListProfiles)))
		rt.h("POST "+prefix,
			rt.middleware.RequireAuth(http.HandlerFunc(rt.profileHandlers.CreateProfile)))
		rt.h("GET "+prefix+"/{id}",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.profileHandlers.GetProfile)))
		rt.h("PUT "+prefix+"/{id}",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.profileHandlers.UpdateProfile)))
		rt.h("DELETE "+prefix+"/{id}",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.profileHandlers.DeleteProfile)))
		rt.h("POST "+prefix+"/{id}/duplicate",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.profileHandlers.DuplicateProfile)))
		rt.h("GET "+prefix+"/{id}/qr",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.profileHandlers.GenerateQRCode)))
		rt.h("POST "+prefix+"/{id}/verify-domain",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.profileHandlers.VerifyDomain)))
		rt.h("GET "+prefix+"/{id}/links",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.linkHandlers.ListLinks)))
		rt.h("POST "+prefix+"/{id}/links",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.linkHandlers.CreateLink)))
	}

	// ── Links (/api/links AND /api/v1/links) ──────────────────────────────────
	for _, prefix := range []string{"/api/links", "/api/v1/links"} {
		rt.h("PUT "+prefix+"/{id}",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.linkHandlers.UpdateLink)))
		rt.h("DELETE "+prefix+"/{id}",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.linkHandlers.DeleteLink)))
		rt.h("POST "+prefix+"/reorder",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.linkHandlers.ReorderLinks)))
		rt.h("POST "+prefix+"/{id}/toggle",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.linkHandlers.ToggleLink)))
	}

	// ── Services (public read, /api/services AND /api/v1/services) ────────────
	for _, prefix := range []string{"/api/services", "/api/v1/services"} {
		rt.hf("GET "+prefix, rt.serviceHandlers.ListServices)
		rt.hf("GET "+prefix+"/search", rt.serviceHandlers.SearchServices)
		rt.hf("GET "+prefix+"/categories", rt.serviceHandlers.ListCategories)
		rt.hf("GET "+prefix+"/popular", rt.serviceHandlers.ListPopularServices)
		rt.hf("GET "+prefix+"/{id}", rt.serviceHandlers.GetService)
	}

	// ── Analytics (/api/analytics AND /api/v1/analytics) ─────────────────────
	for _, prefix := range []string{"/api/analytics", "/api/v1/analytics"} {
		rt.h("GET "+prefix+"/profile/{id}",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.analyticsHandlers.GetProfileAnalytics)))
		rt.h("GET "+prefix+"/links/{profile_id}",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.analyticsHandlers.GetLinkAnalytics)))
		rt.h("GET "+prefix+"/export/{profile_id}",
			rt.middleware.RequireAuth(http.HandlerFunc(rt.analyticsHandlers.ExportAnalytics)))
	}

	// ── Shortlinks (/api/v1/shortlinks) ───────────────────────────────────────
	rt.h("GET /api/v1/shortlinks",
		rt.middleware.RequireAuth(http.HandlerFunc(rt.shortlinkHandler.HandleListShortlinks)))
	rt.h("POST /api/v1/shortlinks",
		rt.middleware.RequireAuth(http.HandlerFunc(rt.shortlinkHandler.HandleCreateShortlink)))
	rt.h("DELETE /api/v1/shortlinks",
		rt.middleware.RequireAuth(http.HandlerFunc(rt.shortlinkHandler.HandleDeleteShortlink)))
	rt.hf("GET /api/v1/shortlinks/{code}", rt.shortlinkHandler.HandleGetShortlink)

	// Shortlink public redirect: /s/{code}
	rt.hf("GET /s/", rt.shortlinkHandler.HandleRedirectShortlink)

	// ── Public API v1 ─────────────────────────────────────────────────────────
	rt.hf("GET /api/v1/profiles/{username}", rt.publicHandlers.GetPublicProfile)
	rt.hf("GET /api/v1/profiles/{username}/links", rt.publicHandlers.GetPublicProfileLinks)
	rt.hf("GET /api/v1/profiles/{username}/qr", rt.publicHandlers.GetPublicProfileQR)
	rt.hf("GET /api/v1/link/{id}/click", rt.publicHandlers.TrackLinkClick)

	// ── Admin API (/api/admin AND /api/v1/admin) ──────────────────────────────
	for _, prefix := range []string{"/api/admin", "/api/v1/admin"} {
		rt.h("GET "+prefix+"/users",
			rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.ListUsers)))
		rt.h("GET "+prefix+"/users/{id}",
			rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.GetUser)))
		rt.h("PUT "+prefix+"/users/{id}",
			rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.UpdateUser)))
		rt.h("DELETE "+prefix+"/users/{id}",
			rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.DeleteUser)))
		rt.h("GET "+prefix+"/stats",
			rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.GetSystemStats)))
		rt.h("POST "+prefix+"/backup",
			rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.TriggerBackup)))
		rt.h("GET "+prefix+"/settings",
			rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.GetSettings)))
		rt.h("PUT "+prefix+"/settings",
			rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.UpdateSettings)))
		rt.h("POST "+prefix+"/services/import",
			rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.ImportServices)))
		rt.h("POST "+prefix+"/cache/clear",
			rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.ClearCache)))
		rt.h("GET "+prefix+"/smtp/config",
			rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.GetSMTPConfig)))
		rt.h("PUT "+prefix+"/smtp/config",
			rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.UpdateSMTPConfig)))
		rt.h("POST "+prefix+"/smtp/test",
			rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.TestSMTPConnection)))
		rt.h("GET "+prefix+"/notifications/preferences",
			rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.GetNotificationPreferences)))
		rt.h("PUT "+prefix+"/notifications/preferences",
			rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.UpdateNotificationPreferences)))
	}

	// ── Link click redirect: /l/{id} (used by profile link cards) ────────────
	rt.hf("GET /l/{id}", rt.publicHandlers.TrackLinkClick)

	// ── HTML pages (server-side rendered templates) ────────────────────────────
	rt.hf("GET /setup", rt.serveSetupPage)
	rt.hf("GET /auth/login", rt.serveLoginPage)
	rt.hf("GET /auth/register", rt.serveRegisterPage)
	rt.h("GET /dashboard",
		rt.middleware.RequireAuth(http.HandlerFunc(rt.serveDashboardPage)))
	rt.h("GET /admin",
		rt.middleware.RequireAdmin(http.HandlerFunc(rt.serveAdminPage)))

	// Home page — must be last fixed path so /{slug} can catch everything else
	rt.hf("GET /", rt.serveHomePage)

	// ── Debug endpoints (--debug/DEBUG=true only; 404 otherwise) ──────────────
	rt.registerDebugRoutes()

	// Apply security headers to all routes
	return rt.middleware.SecurityHeaders(rt.mux)
}

// Handler returns the configured http.Handler (calls SetupRoutes).
func (rt *Router) Handler() http.Handler {
	return rt.SetupRoutes()
}

// ServeHTTP implements http.Handler.
func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.Handler().ServeHTTP(w, r)
}

// pwaInstance builds a server.PWA from current DB settings (re-reads on each
// call so settings changes take effect without restart).
func (rt *Router) pwaInstance() *server.PWA {
	siteName := "Cassocial"
	siteDesc := "Self-hosted link aggregator"
	if v, err := rt.db.GetSetting("site_name"); err == nil && v != "" {
		siteName = v
	}
	if v, err := rt.db.GetSetting("site_description"); err == nil && v != "" {
		siteDesc = v
	}
	return server.NewPWA(siteName, siteDesc)
}

// ── Infrastructure handlers ────────────────────────────────────────────────

func (rt *Router) healthzHTML(w http.ResponseWriter, r *http.Request) {
	status := "healthy"
	code := http.StatusOK
	if err := rt.db.Ping(); err != nil {
		status = "unhealthy"
		code = http.StatusServiceUnavailable
	}

	uptime := formatUptime(time.Since(rt.startTime))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8">
<title>Health Status - Cassocial</title>
<style>body{font-family:sans-serif;max-width:600px;margin:40px auto;padding:20px}
.ok{color:#28a745}.error{color:#dc3545}</style></head><body>
<h1>Cassocial Health Status</h1>
<p>Status: <strong class="%s">%s</strong></p>
<p>Uptime: %s</p>
<p>Version: %s</p>
</body></html>`, status, status, uptime, appVersionString)
}

func (rt *Router) healthzJSON(w http.ResponseWriter, r *http.Request) {
	status := "healthy"
	code := http.StatusOK
	if err := rt.db.Ping(); err != nil {
		status = "unhealthy"
		code = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    status,
		"uptime":    formatUptime(time.Since(rt.startTime)),
		"version":   appVersionString,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (rt *Router) readinessCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ready",
		"time":   getCurrentTimestamp(),
	})
}

func (rt *Router) livenessCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "alive",
		"time":   getCurrentTimestamp(),
	})
}

func (rt *Router) metrics(w http.ResponseWriter, r *http.Request) {
	// Only accept connections from localhost to keep metrics internal
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	if ip != "127.0.0.1" && ip != "::1" && ip != "localhost" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# HELP cassocial_up Whether the server is up\n")
	fmt.Fprintf(w, "# TYPE cassocial_up gauge\n")
	fmt.Fprintf(w, "cassocial_up 1\n")
	fmt.Fprintf(w, "# HELP cassocial_uptime_seconds Server uptime in seconds\n")
	fmt.Fprintf(w, "# TYPE cassocial_uptime_seconds gauge\n")
	fmt.Fprintf(w, "cassocial_uptime_seconds %d\n", int(time.Since(rt.startTime).Seconds()))
}

func (rt *Router) robotsTxt(w http.ResponseWriter, r *http.Request) {
	siteURL := "https://" + r.Host
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "User-agent: *\nAllow: /\nDisallow: /api/\nDisallow: /admin\n\nSitemap: %s/sitemap.xml\n", siteURL)
}

func (rt *Router) sitemapXML(w http.ResponseWriter, r *http.Request) {
	siteURL := "https://" + r.Host
	if v, err := rt.db.GetSetting("site_url"); err == nil && v != "" {
		siteURL = strings.TrimRight(v, "/")
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>%s/</loc><changefreq>daily</changefreq><priority>1.0</priority></url>
</urlset>`, siteURL)
}

func (rt *Router) securityTxt(w http.ResponseWriter, r *http.Request) {
	siteURL := "https://" + r.Host
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Contact: https://github.com/casapps/cassocial/security\nPreferred-Languages: en\nCanonical: %s/.well-known/security.txt\n", siteURL)
}

// ── HTML page handlers ─────────────────────────────────────────────────────

// serveSetupPage redirects to home if already initialized, otherwise renders setup wizard.
func (rt *Router) serveSetupPage(w http.ResponseWriter, r *http.Request) {
	if v, err := rt.db.GetSetting("initialized"); err == nil && v == "true" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	base := newBaseData(rt, r)
	renderTemplate(w, "setup", base)
}

// serveLoginPage renders the login page.
func (rt *Router) serveLoginPage(w http.ResponseWriter, r *http.Request) {
	// Already logged in → dashboard
	if tok := rt.middleware.ExtractToken(r); tok != "" {
		if _, err := rt.authHandlers.auth.ValidateToken(tok); err == nil {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
	}
	base := newBaseData(rt, r)
	base.MetaTitle = "Login"
	renderTemplate(w, "login", base)
}

// serveRegisterPage renders the registration page.
func (rt *Router) serveRegisterPage(w http.ResponseWriter, r *http.Request) {
	base := newBaseData(rt, r)
	base.MetaTitle = "Create Account"
	renderTemplate(w, "register", base)
}

// serveDashboardPage renders the authenticated dashboard.
func (rt *Router) serveDashboardPage(w http.ResponseWriter, r *http.Request) {
	base := newBaseData(rt, r)
	base.MetaTitle = "Dashboard"
	renderTemplate(w, "dashboard", base)
}

// serveAdminPage renders the admin panel.
func (rt *Router) serveAdminPage(w http.ResponseWriter, r *http.Request) {
	base := newBaseData(rt, r)
	base.MetaTitle = "Admin"
	renderTemplate(w, "admin", base)
}

// ProfilePageData is the template data for the public profile page.
type ProfilePageData struct {
	BaseTemplateData
	Profile       *store.Profile
	Links         []*store.Link
	ShowUsernames bool
}

// serveHomePage serves GET /. If the path is "/" it shows the home page;
// any other path is treated as a profile slug and routed to serveProfilePage.
func (rt *Router) serveHomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		// Check if setup is needed and redirect
		if v, err := rt.db.GetSetting("initialized"); err != nil || v != "true" {
			http.Redirect(w, r, "/setup", http.StatusSeeOther)
			return
		}
		base := newBaseData(rt, r)
		renderTemplate(w, "home", base)
		return
	}

	// Any path that doesn't match a registered route falls through to here
	// and is treated as a profile slug (e.g. /casjay)
	rt.serveProfilePage(w, r)
}

// serveProfilePage fetches a profile by slug and renders the HTML profile page.
func (rt *Router) serveProfilePage(w http.ResponseWriter, r *http.Request) {
	// Extract slug: strip leading "/" and any trailing path segments
	slug := strings.TrimPrefix(r.URL.Path, "/")
	if idx := strings.Index(slug, "/"); idx != -1 {
		slug = slug[:idx]
	}
	slug = strings.ToLower(strings.TrimSpace(slug))

	// Guard: reject slugs that collide with reserved paths
	reserved := map[string]bool{
		"api": true, "admin": true, "auth": true, "setup": true,
		"dashboard": true, "static": true, "health": true,
		"metrics": true, "manifest.json": true, "service-worker.js": true,
		"robots.txt": true, "sitemap.xml": true, "s": true, "l": true,
		".well-known": true, "server": true,
	}
	if slug == "" || reserved[slug] {
		http.NotFound(w, r)
		return
	}

	profile, err := rt.db.GetProfileBySlug(slug)
	if err != nil || profile == nil {
		http.NotFound(w, r)
		return
	}

	if !profile.IsPublic {
		http.Error(w, "This profile is private", http.StatusForbidden)
		return
	}

	// Fire-and-forget view count and analytics
	go func() {
		_ = rt.db.IncrementProfileViewCount(profile.ID)
		rt.trackProfileView(r, profile.ID)
	}()

	// Fetch active links ordered by position
	links, err := rt.db.GetLinksByProfileID(profile.ID)
	if err != nil {
		links = nil
	}

	// Filter to active links only
	activeLinks := make([]*store.Link, 0, len(links))
	for _, l := range links {
		if l.IsActive {
			activeLinks = append(activeLinks, l)
		}
	}

	base := newBaseData(rt, r)
	if profile.MetaTitle != "" {
		base.MetaTitle = profile.MetaTitle
	} else if profile.DisplayName != "" {
		base.MetaTitle = profile.DisplayName
	} else {
		base.MetaTitle = "@" + profile.Slug
	}
	base.MetaDescription = profile.MetaDescription
	base.OgImageURL = profile.OgImageURL

	data := ProfilePageData{
		BaseTemplateData: base,
		Profile:          profile,
		Links:            activeLinks,
		ShowUsernames:    profile.ShowUsernames,
	}

	renderTemplate(w, "profile", data)
}

// trackProfileView records a profile view in the analytics table.
func (rt *Router) trackProfileView(r *http.Request, profileID string) {
	ip := getIPAddress(r)
	ipHash := hashIP(ip)
	deviceType := detectDeviceType(r.UserAgent())

	query := `INSERT INTO analytics (id, profile_id, event_type, ip_hash, user_agent, referrer, device_type, created_at)
			  VALUES (?, ?, 'view', ?, ?, ?, ?, ?)`
	if rt.db.Driver == "postgres" {
		query = `INSERT INTO analytics (id, profile_id, event_type, ip_hash, user_agent, referrer, device_type, created_at)
				 VALUES ($1, $2, 'view', $3, $4, $5, $6, $7)`
	}
	rt.db.ExecR(query, generateUUID(), profileID, ipHash, r.UserAgent(), r.Referer(), deviceType, rt.db.BindTime(getCurrentTimestamp()))
}

// ── Helpers ────────────────────────────────────────────────────────────────

func formatUptime(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
