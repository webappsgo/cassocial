package web

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/casapps/cassocial/internal/auth"
	"github.com/casapps/cassocial/internal/config"
	"github.com/casapps/cassocial/internal/database"
	"github.com/casapps/cassocial/internal/models"
)

// Note: embed paths are relative to the package directory
// These will be populated when building from the root

var staticFiles embed.FS
var templateFiles embed.FS

// Server represents the HTTP server
type Server struct {
	config         *config.Config
	db             *database.DB
	auth           *auth.Auth
	middleware     *auth.Middleware
	sessionManager *auth.SessionManager
	templates      *TemplateManager
	router         *http.ServeMux
	httpServer     *http.Server
}

// New creates a new web server instance
func New(cfg *config.Config, db *database.DB, authService *auth.Auth) (*Server, error) {
	// Create session manager
	sessionTimeoutStr, _ := db.GetSetting(models.SettingSessionTimeoutMinutes)
	sessionTimeout := 1440 // Default 24 hours
	if sessionTimeoutStr != "" {
		fmt.Sscanf(sessionTimeoutStr, "%d", &sessionTimeout)
	}
	sessionManager := auth.NewSessionManager(sessionTimeout)

	// Create middleware
	middleware := auth.NewMiddleware(authService)

	// Create template manager
	templates, err := NewTemplateManager(templateFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize templates: %w", err)
	}

	s := &Server{
		config:         cfg,
		db:             db,
		auth:           authService,
		middleware:     middleware,
		sessionManager: sessionManager,
		templates:      templates,
		router:         http.NewServeMux(),
	}

	// Setup routes
	s.setupRoutes()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s, nil
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Health endpoints
	s.router.HandleFunc("/health", s.handleHealth)
	s.router.HandleFunc("/health/ready", s.handleHealthReady)
	s.router.HandleFunc("/health/live", s.handleHealthLive)

	// Static files - will be served from embedded FS when available
	// For now, we'll skip static file serving until proper embedding is set up
	// s.router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Apply global middleware
	secureRouter := s.middleware.SecurityHeaders(s.router)
	maintenanceRouter := s.maintenanceMiddleware(secureRouter)

	// Setup wizard and public routes
	s.router.HandleFunc("/setup", s.handleSetup)
	s.router.HandleFunc("/setup/welcome", s.handleSetupWelcome)
	s.router.HandleFunc("/setup/basic", s.handleSetupBasic)
	s.router.HandleFunc("/setup/domain", s.handleSetupDomain)
	s.router.HandleFunc("/setup/email", s.handleSetupEmail)
	s.router.HandleFunc("/setup/features", s.handleSetupFeatures)
	s.router.HandleFunc("/setup/database", s.handleSetupDatabase)
	s.router.HandleFunc("/setup/complete", s.handleSetupComplete)

	// Authentication routes
	s.router.HandleFunc("/login", s.handleLogin)
	s.router.HandleFunc("/register", s.handleRegister)
	s.router.HandleFunc("/logout", s.handleLogout)
	s.router.HandleFunc("/forgot-password", s.handleForgotPassword)
	s.router.HandleFunc("/reset-password", s.handleResetPassword)

	// Dashboard (requires authentication)
	s.router.Handle("/dashboard", s.requireAuth(http.HandlerFunc(s.handleDashboard)))
	s.router.Handle("/dashboard/profiles", s.requireAuth(http.HandlerFunc(s.handleProfilesList)))
	s.router.Handle("/dashboard/profile/new", s.requireAuth(http.HandlerFunc(s.handleProfileNew)))
	s.router.Handle("/dashboard/profile/edit", s.requireAuth(http.HandlerFunc(s.handleProfileEdit)))
	s.router.Handle("/dashboard/analytics", s.requireAuth(http.HandlerFunc(s.handleAnalytics)))
	s.router.Handle("/dashboard/settings", s.requireAuth(http.HandlerFunc(s.handleUserSettings)))

	// Admin routes (requires admin role)
	s.router.Handle("/admin", s.requireAdmin(http.HandlerFunc(s.handleAdmin)))
	s.router.Handle("/admin/users", s.requireAdmin(http.HandlerFunc(s.handleAdminUsers)))
	s.router.Handle("/admin/settings", s.requireAdmin(http.HandlerFunc(s.handleAdminSettings)))
	s.router.Handle("/admin/smtp", s.requireAdmin(http.HandlerFunc(s.handleAdminSMTP)))
	s.router.Handle("/admin/notifications", s.requireAdmin(http.HandlerFunc(s.handleAdminNotifications)))
	s.router.Handle("/admin/backup", s.requireAdmin(http.HandlerFunc(s.handleAdminBackup)))

	// Public profile pages
	s.router.HandleFunc("/", s.handleIndex)
	s.router.HandleFunc("/p/", s.handleProfile)

	// Update main handler to use middleware chain
	s.httpServer.Handler = maintenanceRouter
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting Cassocial server on %s", s.httpServer.Addr)

	// Check if first run
	initialized, err := s.isInitialized()
	if err != nil {
		log.Printf("Warning: Could not check initialization status: %v", err)
	} else if !initialized {
		log.Printf("First run detected - setup wizard will be available at /setup")
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for interrupt signal or error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case sig := <-sigChan:
		log.Printf("Received signal: %v. Shutting down gracefully...", sig)
		return s.Shutdown()
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Shutting down HTTP server...")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("Server shutdown complete")
	return nil
}

// Health endpoint handlers
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"cassocial"}`))
}

func (s *Server) handleHealthReady(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := s.db.Ping(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"error","message":"database not ready"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}

func (s *Server) handleHealthLive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"alive"}`))
}

// Middleware
func (s *Server) maintenanceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip maintenance check for health endpoints and setup
		if r.URL.Path == "/health" || r.URL.Path == "/health/ready" ||
		   r.URL.Path == "/health/live" || r.URL.Path == "/setup" ||
		   strings.HasPrefix(r.URL.Path, "/setup/") {
			next.ServeHTTP(w, r)
			return
		}

		// Check if in maintenance mode
		maintenanceMode, _ := s.db.GetSetting(models.SettingMaintenanceMode)
		if maintenanceMode == "true" {
			// Check if IP is in bypass list
			if !s.isBypassIP(r) {
				s.renderMaintenancePage(w, r)
				return
			}
		}

		// Check if initialized
		initialized, err := s.isInitialized()
		if err == nil && !initialized {
			// Redirect to setup unless already there
			if !strings.HasPrefix(r.URL.Path, "/setup") {
				http.Redirect(w, r, "/setup", http.StatusTemporaryRedirect)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for session cookie
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		// Validate session
		session, valid := s.sessionManager.ValidateSession(cookie.Value)
		if !valid {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), auth.ContextKeyUserID, session.UserID)
		ctx = context.WithValue(ctx, auth.ContextKeyUsername, session.Username)
		ctx = context.WithValue(ctx, auth.ContextKeyRole, session.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) requireAdmin(next http.Handler) http.Handler {
	return s.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(auth.ContextKeyRole).(string)
		if !ok || role != models.RoleAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}))
}

// Helper functions
func (s *Server) isInitialized() (bool, error) {
	initialized, err := s.db.GetSetting(models.SettingInitialized)
	if err != nil {
		return false, err
	}
	return initialized == "true", nil
}

func (s *Server) isBypassIP(r *http.Request) bool {
	ip := getIPAddress(r)
	bypassIPsJSON, err := s.db.GetSetting(models.SettingMaintenanceBypassIPs)
	if err != nil {
		return false
	}

	// Simple check for localhost
	if ip == "127.0.0.1" || ip == "::1" {
		return true
	}

	// TODO: Parse JSON array and check if IP is in bypass list
	_ = bypassIPsJSON
	return false
}

func (s *Server) renderMaintenancePage(w http.ResponseWriter, r *http.Request) {
	message, _ := s.db.GetSetting(models.SettingMaintenanceMessage)
	if message == "" {
		message = "We are performing scheduled maintenance. Please check back soon."
	}

	data := map[string]interface{}{
		"Title":   "Maintenance Mode",
		"Message": message,
	}

	s.templates.Render(w, "maintenance.html", data)
}

func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Use RemoteAddr
	return r.RemoteAddr
}
