package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRoutes configures all HTTP routes
// Per PART 20: API Structure
func (s *Server) SetupRoutes() {
	// Health endpoints (NON-NEGOTIABLE per AI.md)
	s.router.HandleFunc("/healthz", s.handleHealthz).Methods("GET")
	s.router.HandleFunc("/api/v1/healthz", s.handleAPIHealthz).Methods("GET")

	// Setup subrouters
	api := s.router.PathPrefix("/api/v1").Subrouter()
	admin := s.router.PathPrefix("/admin").Subrouter()

	// Authentication routes (public)
	s.router.HandleFunc("/auth/register", s.authHandler.HandleRegister).Methods("POST")
	s.router.HandleFunc("/auth/login", s.authHandler.HandleLogin).Methods("POST")
	s.router.HandleFunc("/auth/logout", s.authHandler.HandleLogout).Methods("POST")
	s.router.HandleFunc("/auth/forgot-password", s.authHandler.HandleForgotPassword).Methods("POST")
	s.router.HandleFunc("/auth/reset-password", s.authHandler.HandleResetPassword).Methods("POST")
	s.router.HandleFunc("/auth/verify-email", s.authHandler.HandleVerifyEmail).Methods("GET")

	// Public routes
	s.router.HandleFunc("/", s.publicHandler.HandleHome).Methods("GET")
	s.router.HandleFunc("/{slug}", s.publicHandler.HandleProfileView).Methods("GET")
	s.router.HandleFunc("/qr/{slug}", s.qrHandler.HandleQRCode).Methods("GET")
	s.router.HandleFunc("/s/{code}", s.shortlinkHandler.HandleRedirect).Methods("GET")

	// API Routes - Services (public)
	api.HandleFunc("/services", s.serviceHandler.ListServices).Methods("GET")
	api.HandleFunc("/services/{id}", s.serviceHandler.GetService).Methods("GET")
	api.HandleFunc("/services/search", s.serviceHandler.SearchServices).Methods("GET")

	// API Routes - Profiles (requires auth)
	api.HandleFunc("/profiles", s.middleware.RequireAuth(http.HandlerFunc(s.profileHandler.ListProfiles))).Methods("GET")
	api.HandleFunc("/profiles", s.middleware.RequireAuth(http.HandlerFunc(s.profileHandler.CreateProfile))).Methods("POST")
	api.HandleFunc("/profiles/{id}", s.middleware.RequireAuth(http.HandlerFunc(s.profileHandler.GetProfile))).Methods("GET")
	api.HandleFunc("/profiles/{id}", s.middleware.RequireAuth(http.HandlerFunc(s.profileHandler.UpdateProfile))).Methods("PUT")
	api.HandleFunc("/profiles/{id}", s.middleware.RequireAuth(http.HandlerFunc(s.profileHandler.DeleteProfile))).Methods("DELETE")

	// API Routes - Links (requires auth)
	api.HandleFunc("/profiles/{id}/links", s.middleware.RequireAuth(http.HandlerFunc(s.linkHandler.ListLinks))).Methods("GET")
	api.HandleFunc("/profiles/{id}/links", s.middleware.RequireAuth(http.HandlerFunc(s.linkHandler.CreateLink))).Methods("POST")
	api.HandleFunc("/links/{id}", s.middleware.RequireAuth(http.HandlerFunc(s.linkHandler.UpdateLink))).Methods("PUT")
	api.HandleFunc("/links/{id}", s.middleware.RequireAuth(http.HandlerFunc(s.linkHandler.DeleteLink))).Methods("DELETE")
	api.HandleFunc("/links/reorder", s.middleware.RequireAuth(http.HandlerFunc(s.linkHandler.ReorderLinks))).Methods("POST")

	// API Routes - Themes (requires auth)
	api.HandleFunc("/profiles/{id}/theme", s.middleware.RequireAuth(http.HandlerFunc(s.themeHandler.GetTheme))).Methods("GET")
	api.HandleFunc("/profiles/{id}/theme", s.middleware.RequireAuth(http.HandlerFunc(s.themeHandler.UpdateTheme))).Methods("PUT")

	// API Routes - Analytics (requires auth)
	api.HandleFunc("/profiles/{id}/analytics", s.middleware.RequireAuth(http.HandlerFunc(s.analyticsHandler.GetProfileAnalytics))).Methods("GET")
	api.HandleFunc("/links/{id}/analytics", s.middleware.RequireAuth(http.HandlerFunc(s.analyticsHandler.GetLinkAnalytics))).Methods("GET")
	api.HandleFunc("/analytics/export", s.middleware.RequireAuth(http.HandlerFunc(s.analyticsHandler.ExportAnalytics))).Methods("GET")

	// API Routes - Shortlinks (requires auth)
	api.HandleFunc("/shortlinks", s.middleware.RequireAuth(http.HandlerFunc(s.shortlinkHandler.ListShortlinks))).Methods("GET")
	api.HandleFunc("/shortlinks", s.middleware.RequireAuth(http.HandlerFunc(s.shortlinkHandler.CreateShortlink))).Methods("POST")
	api.HandleFunc("/shortlinks/{id}", s.middleware.RequireAuth(http.HandlerFunc(s.shortlinkHandler.DeleteShortlink))).Methods("DELETE")

	// API Routes - Import/Export (requires auth)
	api.HandleFunc("/import/linktree", s.middleware.RequireAuth(http.HandlerFunc(s.importExportHandler.ImportLinktree))).Methods("POST")
	api.HandleFunc("/import/linkstack", s.middleware.RequireAuth(http.HandlerFunc(s.importExportHandler.ImportLinkstack))).Methods("POST")
	api.HandleFunc("/export/profile/{id}", s.middleware.RequireAuth(http.HandlerFunc(s.importExportHandler.ExportProfile))).Methods("GET")

	// User Dashboard routes (requires auth)
	s.router.HandleFunc("/dashboard", s.middleware.RequireAuth(http.HandlerFunc(s.dashboardHandler.HandleDashboard))).Methods("GET")
	s.router.HandleFunc("/user/profile", s.middleware.RequireAuth(http.HandlerFunc(s.userHandler.HandleProfile))).Methods("GET")
	s.router.HandleFunc("/user/settings", s.middleware.RequireAuth(http.HandlerFunc(s.userHandler.HandleSettings))).Methods("GET", "POST")

	// Admin routes (requires admin auth)
	admin.Use(s.middleware.RequireAdminAuth)
	admin.HandleFunc("", s.adminHandler.Dashboard).Methods("GET")
	admin.HandleFunc("/users", s.adminHandler.ListUsers).Methods("GET")
	admin.HandleFunc("/users/{id}", s.adminHandler.GetUser).Methods("GET")
	admin.HandleFunc("/users/{id}", s.adminHandler.UpdateUser).Methods("PUT")
	admin.HandleFunc("/users/{id}", s.adminHandler.DeleteUser).Methods("DELETE")
	admin.HandleFunc("/settings", s.adminHandler.GetSettings).Methods("GET")
	admin.HandleFunc("/settings", s.adminHandler.UpdateSettings).Methods("POST")
	admin.HandleFunc("/backup", s.adminHandler.TriggerBackup).Methods("POST")
	admin.HandleFunc("/stats", s.adminHandler.GetSystemStats).Methods("GET")

	// API Documentation routes
	s.router.HandleFunc("/openapi", s.swaggerHandler.ServeUI).Methods("GET")
	s.router.HandleFunc("/openapi.json", s.swaggerHandler.ServeSpec).Methods("GET")
	s.router.HandleFunc("/graphql", s.graphqlHandler.ServePlayground).Methods("GET")
	s.router.HandleFunc("/graphql", s.graphqlHandler.ServeQuery).Methods("POST")

	// Static files
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
}

