package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// newTestServer creates a minimal Server for testing handlers.
func newTestServer(t *testing.T) *Server {
	t.Helper()
	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	cfg := &config.Config{
		DataDir: t.TempDir(),
		Server: config.ServerConfig{
			Mode: "production",
		},
	}

	return &Server{
		config:    cfg,
		db:        db,
		startTime: time.Now(),
	}
}

func TestHandleHealthz_OK(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()
	s.handleHealthz(rr, req)

	if rr.Code != 200 {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}

func TestHandleHealthz_ShuttingDown(t *testing.T) {
	s := newTestServer(t)
	s.isShuttingDown = true

	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()
	s.handleHealthz(rr, req)

	if rr.Code != 503 {
		t.Errorf("status = %d, want 503 when shutting down", rr.Code)
	}
}

func TestHandleAPIHealthz_OK(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("GET", "/api/v1/healthz", nil)
	rr := httptest.NewRecorder()
	s.handleAPIHealthz(rr, req)

	if rr.Code != 200 {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var health HealthResponse
	if err := json.NewDecoder(rr.Body).Decode(&health); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if health.Status != "healthy" {
		t.Errorf("Status = %q, want healthy", health.Status)
	}
}

func TestHandleAPIHealthz_ShuttingDown(t *testing.T) {
	s := newTestServer(t)
	s.isShuttingDown = true

	req := httptest.NewRequest("GET", "/api/v1/healthz", nil)
	rr := httptest.NewRecorder()
	s.handleAPIHealthz(rr, req)

	if rr.Code != 503 {
		t.Errorf("status = %d, want 503 when shutting down", rr.Code)
	}

	var health HealthResponse
	json.NewDecoder(rr.Body).Decode(&health)
	if health.Status == "healthy" {
		t.Error("status should not be healthy when shutting down")
	}
}

func TestGetHealthStatus(t *testing.T) {
	s := newTestServer(t)
	health := s.getHealthStatus()

	if health.Status == "" {
		t.Error("Status is empty")
	}
	if health.Timestamp == "" {
		t.Error("Timestamp is empty")
	}
	if health.Uptime == "" {
		t.Error("Uptime is empty")
	}
	if _, ok := health.Checks["database"]; !ok {
		t.Error("Checks missing 'database' key")
	}
	if _, ok := health.Checks["disk"]; !ok {
		t.Error("Checks missing 'disk' key")
	}
}

func TestFormatUptime_Minutes(t *testing.T) {
	d := 45 * time.Minute
	got := formatUptime(d)
	if got != "45m" {
		t.Errorf("formatUptime(%v) = %q, want 45m", d, got)
	}
}

func TestFormatUptime_Hours(t *testing.T) {
	d := 2*time.Hour + 30*time.Minute
	got := formatUptime(d)
	if got != "2h 30m" {
		t.Errorf("formatUptime(%v) = %q, want '2h 30m'", d, got)
	}
}

func TestFormatUptime_Days(t *testing.T) {
	d := 3*24*time.Hour + 5*time.Hour + 15*time.Minute
	got := formatUptime(d)
	if got != "3d 5h 15m" {
		t.Errorf("formatUptime(%v) = %q, want '3d 5h 15m'", d, got)
	}
}

func TestFormatUptime_Zero(t *testing.T) {
	got := formatUptime(0)
	if got != "0m" {
		t.Errorf("formatUptime(0) = %q, want 0m", got)
	}
}

func TestIsDiskHealthy_ValidDir(t *testing.T) {
	dir := t.TempDir()
	if !isDiskHealthy(dir) {
		t.Error("isDiskHealthy returned false for writable temp dir")
	}
}

func TestIsDiskHealthy_InvalidDir(t *testing.T) {
	// Non-existent directory should return false
	if isDiskHealthy("/nonexistent/path/that/does/not/exist") {
		t.Error("isDiskHealthy returned true for non-existent dir")
	}
}

func TestGetVersion(t *testing.T) {
	v := getVersion()
	if v == "" {
		t.Error("getVersion returned empty string")
	}
}

func TestNew(t *testing.T) {
	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	cfg := &config.Config{
		DataDir: t.TempDir(),
		Server: config.ServerConfig{
			Address: "127.0.0.1",
			Port:    0,
			Mode:    "production",
		},
	}

	s, err := New(cfg, db, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if s == nil {
		t.Fatal("New returned nil server")
	}
	if s.config != cfg {
		t.Error("server config not set")
	}
	if s.db != db {
		t.Error("server db not set")
	}
	if s.httpServer == nil {
		t.Error("httpServer not initialized")
	}
}

func TestShutdown(t *testing.T) {
	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	cfg := &config.Config{
		DataDir: t.TempDir(),
		Server: config.ServerConfig{
			Address: "127.0.0.1",
			Port:    0,
			Mode:    "production",
		},
	}

	s, err := New(cfg, db, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Shutdown should not error even when server has not started (httpServer.Shutdown on idle server)
	if err := s.Shutdown(); err != nil {
		t.Errorf("Shutdown returned unexpected error: %v", err)
	}
	if !s.isShuttingDown {
		t.Error("isShuttingDown should be true after Shutdown")
	}
}

func TestGetHealthStatus_IsShuttingDown(t *testing.T) {
	s := newTestServer(t)
	s.isShuttingDown = true
	health := s.getHealthStatus()
	if health.Status != "shutting_down" {
		t.Errorf("status = %q, want shutting_down", health.Status)
	}
}

func TestGetHealthStatus_UnhealthyDisk(t *testing.T) {
	s := newTestServer(t)
	// Point DataDir at a non-writable path so disk check fails
	s.config.DataDir = "/nonexistent/path/that/does/not/exist"
	health := s.getHealthStatus()
	if health.Status == "healthy" {
		t.Error("status should not be healthy when disk is not writable")
	}
	if health.Checks["disk"] != "error" {
		t.Errorf("disk check = %q, want error", health.Checks["disk"])
	}
}
