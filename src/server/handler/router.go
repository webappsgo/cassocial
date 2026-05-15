package handler

import (
	"net/http"

	"github.com/casapps/cassocial/src/server"
	"github.com/casapps/cassocial/src/server/store"
)

// Router handles HTTP routing for the API
type Router struct {
	mux              *http.ServeMux
	middleware       *Middleware
	authHandlers     *AuthHandlers
	profileHandlers  *ProfileHandlers
	linkHandlers     *LinkHandlers
	serviceHandlers  *ServiceHandlers
	analyticsHandlers *AnalyticsHandlers
	adminHandlers    *AdminHandlers
	publicHandlers   *PublicHandlers
}

// NewRouter creates a new Router instance with all handlers
func NewRouter(db *store.DB, authService *Auth) *Router {
	mux := http.NewServeMux()
	middleware := server.NewMiddleware(authService)

	return &Router{
		mux:              mux,
		middleware:       middleware,
		authHandlers:     NewAuthHandlers(authService, db),
		profileHandlers:  NewProfileHandlers(db),
		linkHandlers:     NewLinkHandlers(db),
		serviceHandlers:  NewServiceHandlers(db),
		analyticsHandlers: NewAnalyticsHandlers(db),
		adminHandlers:    NewAdminHandlers(db, authService),
		publicHandlers:   NewPublicHandlers(db),
	}
}

// SetupRoutes configures all API routes
func (rt *Router) SetupRoutes() http.Handler {
	// Health check endpoint (no auth required)
	rt.mux.HandleFunc("GET /health", rt.healthCheck)
	rt.mux.HandleFunc("GET /health/ready", rt.readinessCheck)
	rt.mux.HandleFunc("GET /health/live", rt.livenessCheck)

	// Authentication endpoints (no auth required)
	rt.mux.HandleFunc("POST /api/auth/register", rt.authHandlers.Register)
	rt.mux.HandleFunc("POST /api/auth/login", rt.authHandlers.Login)
	rt.mux.HandleFunc("POST /api/auth/login/2fa", rt.authHandlers.LoginWith2FA)
	rt.mux.HandleFunc("POST /api/auth/forgot-password", rt.authHandlers.ForgotPassword)
	rt.mux.HandleFunc("POST /api/auth/reset-password", rt.authHandlers.ResetPassword)
	rt.mux.HandleFunc("GET /api/auth/verify-email/{token}", rt.authHandlers.VerifyEmail)

	// Authenticated auth endpoints
	rt.mux.Handle("POST /api/auth/logout", rt.middleware.RequireAuth(http.HandlerFunc(rt.authHandlers.Logout)))
	rt.mux.Handle("POST /api/auth/refresh", rt.middleware.RequireAuth(http.HandlerFunc(rt.authHandlers.RefreshToken)))
	rt.mux.Handle("POST /api/auth/2fa/enable", rt.middleware.RequireAuth(http.HandlerFunc(rt.authHandlers.Enable2FA)))
	rt.mux.Handle("POST /api/auth/2fa/verify", rt.middleware.RequireAuth(http.HandlerFunc(rt.authHandlers.Verify2FA)))
	rt.mux.Handle("POST /api/auth/2fa/disable", rt.middleware.RequireAuth(http.HandlerFunc(rt.authHandlers.Disable2FA)))

	// Profile endpoints (authenticated)
	rt.mux.Handle("GET /api/profiles", rt.middleware.RequireAuth(http.HandlerFunc(rt.profileHandlers.ListProfiles)))
	rt.mux.Handle("POST /api/profiles", rt.middleware.RequireAuth(http.HandlerFunc(rt.profileHandlers.CreateProfile)))
	rt.mux.Handle("GET /api/profiles/{id}", rt.middleware.RequireAuth(http.HandlerFunc(rt.profileHandlers.GetProfile)))
	rt.mux.Handle("PUT /api/profiles/{id}", rt.middleware.RequireAuth(http.HandlerFunc(rt.profileHandlers.UpdateProfile)))
	rt.mux.Handle("DELETE /api/profiles/{id}", rt.middleware.RequireAuth(http.HandlerFunc(rt.profileHandlers.DeleteProfile)))
	rt.mux.Handle("POST /api/profiles/{id}/duplicate", rt.middleware.RequireAuth(http.HandlerFunc(rt.profileHandlers.DuplicateProfile)))
	rt.mux.Handle("GET /api/profiles/{id}/qr", rt.middleware.RequireAuth(http.HandlerFunc(rt.profileHandlers.GenerateQRCode)))
	rt.mux.Handle("POST /api/profiles/{id}/verify-domain", rt.middleware.RequireAuth(http.HandlerFunc(rt.profileHandlers.VerifyDomain)))

	// Link endpoints (authenticated)
	rt.mux.Handle("GET /api/profiles/{id}/links", rt.middleware.RequireAuth(http.HandlerFunc(rt.linkHandlers.ListLinks)))
	rt.mux.Handle("POST /api/profiles/{id}/links", rt.middleware.RequireAuth(http.HandlerFunc(rt.linkHandlers.CreateLink)))
	rt.mux.Handle("PUT /api/links/{id}", rt.middleware.RequireAuth(http.HandlerFunc(rt.linkHandlers.UpdateLink)))
	rt.mux.Handle("DELETE /api/links/{id}", rt.middleware.RequireAuth(http.HandlerFunc(rt.linkHandlers.DeleteLink)))
	rt.mux.Handle("POST /api/links/reorder", rt.middleware.RequireAuth(http.HandlerFunc(rt.linkHandlers.ReorderLinks)))
	rt.mux.Handle("POST /api/links/{id}/toggle", rt.middleware.RequireAuth(http.HandlerFunc(rt.linkHandlers.ToggleLink)))

	// Service endpoints (no auth required for public access)
	rt.mux.HandleFunc("GET /api/services", rt.serviceHandlers.ListServices)
	rt.mux.HandleFunc("GET /api/services/search", rt.serviceHandlers.SearchServices)
	rt.mux.HandleFunc("GET /api/services/categories", rt.serviceHandlers.ListCategories)
	rt.mux.HandleFunc("GET /api/services/popular", rt.serviceHandlers.ListPopularServices)
	rt.mux.HandleFunc("GET /api/services/{id}", rt.serviceHandlers.GetService)

	// Analytics endpoints (authenticated)
	rt.mux.Handle("GET /api/analytics/profile/{id}", rt.middleware.RequireAuth(http.HandlerFunc(rt.analyticsHandlers.GetProfileAnalytics)))
	rt.mux.Handle("GET /api/analytics/links/{profile_id}", rt.middleware.RequireAuth(http.HandlerFunc(rt.analyticsHandlers.GetLinkAnalytics)))
	rt.mux.Handle("GET /api/analytics/export/{profile_id}", rt.middleware.RequireAuth(http.HandlerFunc(rt.analyticsHandlers.ExportAnalytics)))

	// Admin endpoints (authenticated + admin role)
	rt.mux.Handle("GET /api/admin/users", rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.ListUsers)))
	rt.mux.Handle("GET /api/admin/users/{id}", rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.GetUser)))
	rt.mux.Handle("PUT /api/admin/users/{id}", rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.UpdateUser)))
	rt.mux.Handle("DELETE /api/admin/users/{id}", rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.DeleteUser)))
	rt.mux.Handle("GET /api/admin/stats", rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.GetSystemStats)))
	rt.mux.Handle("POST /api/admin/backup", rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.TriggerBackup)))
	rt.mux.Handle("GET /api/admin/settings", rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.GetSettings)))
	rt.mux.Handle("PUT /api/admin/settings", rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.UpdateSettings)))
	rt.mux.Handle("POST /api/admin/services/import", rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.ImportServices)))
	rt.mux.Handle("POST /api/admin/cache/clear", rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.ClearCache)))
	rt.mux.Handle("GET /api/admin/smtp/config", rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.GetSMTPConfig)))
	rt.mux.Handle("PUT /api/admin/smtp/config", rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.UpdateSMTPConfig)))
	rt.mux.Handle("POST /api/admin/smtp/test", rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.TestSMTPConnection)))
	rt.mux.Handle("GET /api/admin/notifications/preferences", rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.GetNotificationPreferences)))
	rt.mux.Handle("PUT /api/admin/notifications/preferences", rt.middleware.RequireAdmin(http.HandlerFunc(rt.adminHandlers.UpdateNotificationPreferences)))

	// Public API v1 endpoints (no auth required)
	rt.mux.HandleFunc("GET /api/v1/profiles/{username}", rt.publicHandlers.GetPublicProfile)
	rt.mux.HandleFunc("GET /api/v1/profiles/{username}/links", rt.publicHandlers.GetPublicProfileLinks)
	rt.mux.HandleFunc("GET /api/v1/profiles/{username}/qr", rt.publicHandlers.GetPublicProfileQR)
	rt.mux.HandleFunc("GET /api/v1/link/{id}/click", rt.publicHandlers.TrackLinkClick)

	// Apply security headers middleware to all routes
	handler := rt.middleware.SecurityHeaders(rt.mux)

	return handler
}

// Handler returns the HTTP handler
func (rt *Router) Handler() http.Handler {
	return rt.SetupRoutes()
}

// Health check handlers

func (rt *Router) healthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"time":   getCurrentTimestamp(),
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

// ServeHTTP implements the http.Handler interface
func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.Handler().ServeHTTP(w, r)
}
