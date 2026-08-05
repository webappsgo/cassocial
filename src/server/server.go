package server

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/scheduler"
	"github.com/casapps/cassocial/src/server/store"
)

// Server represents the HTTP server
type Server struct {
	config         *config.Config
	db             *store.DB
	httpServer     *http.Server
	isShuttingDown bool
	startTime      time.Time
	scheduler      *scheduler.Scheduler
	tor            *TorService
	version        string
}

// New creates a new server instance. The caller is responsible for building the
// http.Handler (e.g. via handler.NewRouter) and passing it in. version is the
// canonical build-info Version declared once in src/main.go (AI.md PART 26 —
// build info is embedded via -ldflags in main.go, never redeclared elsewhere).
func New(cfg *config.Config, db *store.DB, h http.Handler, version string) (*Server, error) {
	s := &Server{
		config:    cfg,
		db:        db,
		startTime: time.Now(),
		scheduler: scheduler.New(),
		tor:       NewTorService(cfg.DataDir),
		version:   version,
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s, nil
}

// Start starts the HTTP server with graceful shutdown
func (s *Server) Start() error {
	log.Printf("Starting Cassocial on %s", s.httpServer.Addr)

	// Register and start the built-in scheduler (AI.md PART 19 - the ONLY
	// mechanism for background tasks; no external cron under any circumstance).
	tasks := scheduler.NewTasks(s.config, s.db)
	if err := tasks.RegisterAllTasks(s.scheduler); err != nil {
		return fmt.Errorf("failed to register scheduled tasks: %w", err)
	}
	s.scheduler.Start()

	// Auto-enable the Tor hidden service if the tor binary is present
	// (AI.md PART 32 - no enable/disable toggle, app owns the process lifecycle).
	if err := s.tor.Start(); err != nil {
		log.Printf("Warning: failed to start Tor hidden service: %v", err)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	log.Println("Server started successfully")

	// Wait for shutdown signal or error
	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case sig := <-sigChan:
		log.Printf("Received signal %v, starting graceful shutdown...", sig)
		return s.Shutdown()
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	s.isShuttingDown = true

	// Stop background services before the HTTP server so in-flight scheduled
	// work and the Tor process wind down cleanly ahead of the listener.
	s.scheduler.Stop()
	if err := s.tor.Stop(); err != nil {
		log.Printf("Warning: error stopping Tor hidden service: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Shutting down HTTP server...")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("Server shutdown complete")
	return nil
}

// HealthResponse represents the complete health check response (AI.md PART 16)
type HealthResponse struct {
	// healthy, unhealthy, shutting_down
	Status string `json:"status"`
	// Semantic version
	Version string `json:"version"`
	// production or development
	Mode string `json:"mode"`
	// Formatted: "2d 5h 30m"
	Uptime string `json:"uptime"`
	// ISO 8601: "2024-01-15T10:30:00Z"
	Timestamp string `json:"timestamp"`
	Node      NodeInfo         `json:"node"`
	Cluster   ClusterInfo      `json:"cluster"`
	// ok, degraded, error
	Checks map[string]string `json:"checks"`
}

// NodeInfo contains information about this node
type NodeInfo struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
}

// ClusterInfo contains cluster status information
type ClusterInfo struct {
	Enabled bool `json:"enabled"`
	// connected, degraded, disconnected
	Status string `json:"status,omitempty"`
	Nodes  int    `json:"nodes,omitempty"`
	// member
	Role string `json:"role,omitempty"`
}

// handleHealthz handles the /healthz endpoint (HTML version)
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	health := s.getHealthStatus()

	// Return 503 if shutting down or unhealthy
	if health.Status != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Simple HTML template for health status
	tmpl := `<!DOCTYPE html>
<html>
<head>
	<title>Cassocial Health Status</title>
	<style>
		body { font-family: sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
		.status { padding: 20px; border-radius: 5px; margin: 20px 0; }
		.healthy { background: #d4edda; color: #155724; }
		.unhealthy { background: #f8d7da; color: #721c24; }
		.shutting_down { background: #fff3cd; color: #856404; }
		table { width: 100%; border-collapse: collapse; margin: 20px 0; }
		th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
		th { background: #f8f9fa; }
		.ok { color: #28a745; }
		.error { color: #dc3545; }
		.degraded { color: #ffc107; }
	</style>
</head>
<body>
	<h1>Cassocial Health Status</h1>
	<div class="status {{.Status}}">
		<h2>Status: {{.Status}}</h2>
	</div>
	<table>
		<tr><th>Version</th><td>{{.Version}}</td></tr>
		<tr><th>Mode</th><td>{{.Mode}}</td></tr>
		<tr><th>Uptime</th><td>{{.Uptime}}</td></tr>
		<tr><th>Node ID</th><td>{{.Node.ID}}</td></tr>
		<tr><th>Hostname</th><td>{{.Node.Hostname}}</td></tr>
		<tr><th>Cluster</th><td>{{if .Cluster.Enabled}}Enabled ({{.Cluster.Status}}){{else}}Disabled{{end}}</td></tr>
	</table>
	<h3>System Checks</h3>
	<table>
		{{range $key, $value := .Checks}}
		<tr><th>{{$key}}</th><td class="{{$value}}">{{$value}}</td></tr>
		{{end}}
	</table>
</body>
</html>`

	t := template.Must(template.New("health").Parse(tmpl))
	t.Execute(w, health)
}

// handleAPIHealthz handles the /api/v1/healthz endpoint (JSON version)
func (s *Server) handleAPIHealthz(w http.ResponseWriter, r *http.Request) {
	health := s.getHealthStatus()

	// Return 503 if shutting down or unhealthy
	if health.Status != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// getHealthStatus returns the current health status matching AI.md PART 16 spec
func (s *Server) getHealthStatus() HealthResponse {
	hostname, _ := os.Hostname()

	// Calculate uptime
	uptime := formatUptime(time.Since(s.startTime))

	// Determine status
	status := "healthy"
	if s.isShuttingDown {
		status = "shutting_down"
	}

	// Perform health checks
	checks := make(map[string]string)

	// Database check
	if err := s.db.Ping(); err != nil {
		checks["database"] = "error"
		status = "unhealthy"
	} else {
		checks["database"] = "ok"
	}

	// Cache check (always ok for now - no cache implemented yet)
	checks["cache"] = "ok"

	// Disk check (basic check - verifies data dir is writable)
	checks["disk"] = "ok"
	if !isDiskHealthy(s.config.DataDir) {
		checks["disk"] = "error"
		status = "unhealthy"
	}

	// Build response
	return HealthResponse{
		Status:    status,
		Version:   s.getVersion(),
		Mode:      s.config.Server.Mode,
		Uptime:    uptime,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Node: NodeInfo{
			ID:       "standalone",
			Hostname: hostname,
		},
		Cluster: ClusterInfo{
			Enabled: false,
		},
		Checks: checks,
	}
}

// formatUptime formats duration as "2d 5h 30m"
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

// isDiskHealthy checks if the data directory is writable
func isDiskHealthy(dataDir string) bool {
	testFile := fmt.Sprintf("%s/.health-check-%d", dataDir, time.Now().UnixNano())
	if err := os.WriteFile(testFile, []byte("ok"), 0600); err != nil {
		return false
	}
	os.Remove(testFile)
	return true
}

// getVersion returns the version string embedded in main.Version at build
// time and threaded through New() — never redeclared here (AI.md PART 26).
func (s *Server) getVersion() string {
	if s.version == "" {
		return "dev"
	}
	return s.version
}
