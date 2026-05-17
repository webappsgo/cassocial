package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/casapps/cassocial/src/server/store"
)

func newTestMaintenanceMode(t *testing.T) *MaintenanceMode {
	t.Helper()
	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}
	return NewMaintenanceMode(db)
}

func TestNewMaintenanceMode(t *testing.T) {
	mm := newTestMaintenanceMode(t)
	if mm == nil {
		t.Fatal("NewMaintenanceMode returned nil")
	}
}

func TestMaintenanceMode_IsEnabled_Default(t *testing.T) {
	mm := newTestMaintenanceMode(t)
	if mm.IsEnabled() {
		t.Error("IsEnabled() = true by default, want false")
	}
}

func TestMaintenanceMode_Enable_Disable(t *testing.T) {
	mm := newTestMaintenanceMode(t)

	if err := mm.Enable("Down for updates"); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if !mm.IsEnabled() {
		t.Error("IsEnabled() = false after Enable, want true")
	}

	if err := mm.Disable(); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	if mm.IsEnabled() {
		t.Error("IsEnabled() = true after Disable, want false")
	}
}

func TestMaintenanceMode_Enable_EmptyMessage(t *testing.T) {
	mm := newTestMaintenanceMode(t)
	if err := mm.Enable(""); err != nil {
		t.Fatalf("Enable with empty message: %v", err)
	}
	if !mm.IsEnabled() {
		t.Error("IsEnabled() = false after Enable, want true")
	}
}

func TestMaintenanceMode_GetMessage_Default(t *testing.T) {
	mm := newTestMaintenanceMode(t)
	msg := mm.GetMessage()
	if msg == "" {
		t.Error("GetMessage() returned empty string, want default message")
	}
}

func TestMaintenanceMode_GetMessage_Custom(t *testing.T) {
	mm := newTestMaintenanceMode(t)
	custom := "Back in 10 minutes"
	if err := mm.Enable(custom); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	got := mm.GetMessage()
	if got != custom {
		t.Errorf("GetMessage() = %q, want %q", got, custom)
	}
}

func TestMaintenanceMode_IsIPBypassed(t *testing.T) {
	mm := newTestMaintenanceMode(t)

	tests := []struct {
		ip   string
		want bool
	}{
		{"127.0.0.1", true},
		{"::1", true},
		{"localhost", true},
		{"1.2.3.4", false},
		{"192.168.1.1", false},
	}

	for _, tt := range tests {
		if got := mm.IsIPBypassed(tt.ip); got != tt.want {
			t.Errorf("IsIPBypassed(%q) = %v, want %v", tt.ip, got, tt.want)
		}
	}
}

func TestMaintenanceMiddleware_Disabled(t *testing.T) {
	mm := newTestMaintenanceMode(t)
	// maintenance off by default — handler should pass through

	called := false
	handler := MaintenanceMiddleware(mm)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("inner handler not called when maintenance is disabled")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestMaintenanceMiddleware_Enabled_HTMLResponse(t *testing.T) {
	mm := newTestMaintenanceMode(t)
	if err := mm.Enable("Scheduled maintenance"); err != nil {
		t.Fatalf("Enable: %v", err)
	}

	handler := MaintenanceMiddleware(mm)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called during maintenance")
	}))

	req := httptest.NewRequest("GET", "/profile", nil)
	req.RemoteAddr = "5.6.7.8:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}

func TestMaintenanceMiddleware_Enabled_JSONResponse(t *testing.T) {
	mm := newTestMaintenanceMode(t)
	if err := mm.Enable("API maintenance"); err != nil {
		t.Fatalf("Enable: %v", err)
	}

	handler := MaintenanceMiddleware(mm)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called during maintenance")
	}))

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	req.Header.Set("Accept", "application/json")
	req.RemoteAddr = "5.6.7.8:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rr.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Error("JSON response missing 'error' key")
	}
}

func TestMaintenanceMiddleware_ExemptPaths(t *testing.T) {
	mm := newTestMaintenanceMode(t)
	if err := mm.Enable("maintenance"); err != nil {
		t.Fatalf("Enable: %v", err)
	}

	exemptPaths := []string{"/healthz", "/api/v1/healthz", "/admin", "/admin/users"}

	for _, path := range exemptPaths {
		called := false
		handler := MaintenanceMiddleware(mm)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest("GET", path, nil)
		req.RemoteAddr = "5.6.7.8:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if !called {
			t.Errorf("exempt path %q: inner handler not called", path)
		}
	}
}

func TestMaintenanceMiddleware_BypassedIP(t *testing.T) {
	mm := newTestMaintenanceMode(t)
	if err := mm.Enable("maintenance"); err != nil {
		t.Fatalf("Enable: %v", err)
	}

	called := false
	handler := MaintenanceMiddleware(mm)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest("GET", "/profile", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("localhost IP should bypass maintenance mode")
	}
}

func TestGenerateMaintenanceHTML(t *testing.T) {
	html := generateMaintenanceHTML("Test message")
	if !strings.Contains(html, "Maintenance Mode") {
		t.Error("HTML missing 'Maintenance Mode' heading")
	}
	if !strings.Contains(html, "Test message") {
		t.Error("HTML missing the message text")
	}
}

func TestSelfHealingCheck(t *testing.T) {
	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	defer db.Close()
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	dataDir := t.TempDir()
	if err := SelfHealingCheck(db, dataDir); err != nil {
		t.Errorf("SelfHealingCheck returned unexpected error: %v", err)
	}
}
