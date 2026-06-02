package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server"
	"github.com/casapps/cassocial/src/server/store"
)

// newTestRouter creates a Router backed by an in-memory SQLite database.
func newTestRouter(t *testing.T) *Router {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect returned error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	authSvc := server.NewAuth(db, "test-jwt-secret-router")
	return NewRouter(db, authSvc, &config.Config{})
}

func TestNewRouter_NotNil(t *testing.T) {
	rt := newTestRouter(t)
	if rt == nil {
		t.Fatal("NewRouter returned nil")
	}
}

func TestRouter_Handler_NotNil(t *testing.T) {
	rt := newTestRouter(t)
	h := rt.Handler()
	if h == nil {
		t.Fatal("Router.Handler() returned nil")
	}
}

func TestRouter_SetupRoutes_NotNil(t *testing.T) {
	rt := newTestRouter(t)
	h := rt.SetupRoutes()
	if h == nil {
		t.Fatal("Router.SetupRoutes() returned nil")
	}
}

// healthCheck must return 200 with status "ok".
func TestRouter_HealthCheck(t *testing.T) {
	rt := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	rt.healthzJSON(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("healthCheck returned %d, want %d", rr.Code, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode healthCheck response: %v", err)
	}

	if body["status"] != "healthy" {
		t.Errorf("healthzJSON status = %v, want \"healthy\"", body["status"])
	}
}

// readinessCheck must return 200 with status "ready".
func TestRouter_ReadinessCheck(t *testing.T) {
	rt := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rr := httptest.NewRecorder()
	rt.readinessCheck(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("readinessCheck returned %d, want %d", rr.Code, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode readinessCheck response: %v", err)
	}

	if body["status"] != "ready" {
		t.Errorf("readinessCheck status = %v, want \"ready\"", body["status"])
	}
}

// livenessCheck must return 200 with status "alive".
func TestRouter_LivenessCheck(t *testing.T) {
	rt := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rr := httptest.NewRecorder()
	rt.livenessCheck(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("livenessCheck returned %d, want %d", rr.Code, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode livenessCheck response: %v", err)
	}

	if body["status"] != "alive" {
		t.Errorf("livenessCheck status = %v, want \"alive\"", body["status"])
	}
}

// Health endpoints through the full mux must honour the method pattern.
func TestRouter_HealthEndpoints_ViaHandler(t *testing.T) {
	rt := newTestRouter(t)
	h := rt.Handler()

	endpoints := []struct {
		path       string
		wantStatus string
	}{
		{"/health", "healthy"},
		{"/health/ready", "ready"},
		{"/health/live", "alive"},
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(http.MethodGet, ep.path, nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("GET %s returned %d, want %d", ep.path, rr.Code, http.StatusOK)
			continue
		}

		var body map[string]interface{}
		if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response for %s: %v", ep.path, err)
		}

		if body["status"] != ep.wantStatus {
			t.Errorf("GET %s status = %v, want %q", ep.path, body["status"], ep.wantStatus)
		}
	}
}

// ServeHTTP must delegate correctly (smoke test via a health route).
func TestRouter_ServeHTTP(t *testing.T) {
	rt := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	rt.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ServeHTTP /health returned %d, want %d", rr.Code, http.StatusOK)
	}
}

// Response must include Content-Type application/json for health endpoints.
func TestRouter_HealthCheck_ContentType(t *testing.T) {
	rt := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	rt.healthzJSON(rr, req)

	ct := rr.Header().Get("Content-Type")
	if ct == "" {
		t.Error("healthCheck did not set Content-Type header")
	}
}
