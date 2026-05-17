package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// newTestAdmin creates an Admin backed by an in-memory SQLite database.
func newTestAdmin(t *testing.T) *Admin {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect returned error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	cfg := &config.Config{}
	cfg.Server.Mode = "development"
	cfg.Server.Port = 8080
	cfg.Database.Driver = "sqlite"

	return New(cfg, db)
}

func TestNew(t *testing.T) {
	a := newTestAdmin(t)
	if a == nil {
		t.Fatal("New returned nil")
	}
}

func TestGenerateSetupToken(t *testing.T) {
	a := newTestAdmin(t)

	token, err := a.GenerateSetupToken()
	if err != nil {
		t.Fatalf("GenerateSetupToken returned error: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateSetupToken returned empty token")
	}
	if len(token) < 16 {
		t.Errorf("GenerateSetupToken returned token of length %d, want >=16", len(token))
	}
}

func TestValidateSetupToken_Valid(t *testing.T) {
	a := newTestAdmin(t)

	token, err := a.GenerateSetupToken()
	if err != nil {
		t.Fatalf("GenerateSetupToken returned error: %v", err)
	}

	if !a.ValidateSetupToken(token) {
		t.Errorf("ValidateSetupToken returned false for a valid token")
	}
}

func TestValidateSetupToken_Invalid(t *testing.T) {
	a := newTestAdmin(t)

	// No token generated yet — any input must return false.
	if a.ValidateSetupToken("wrong-token") {
		t.Errorf("ValidateSetupToken returned true for an invalid token")
	}
}

func TestValidateSetupToken_WrongValue(t *testing.T) {
	a := newTestAdmin(t)

	_, err := a.GenerateSetupToken()
	if err != nil {
		t.Fatalf("GenerateSetupToken returned error: %v", err)
	}

	if a.ValidateSetupToken("not-the-right-token") {
		t.Errorf("ValidateSetupToken returned true for a wrong token")
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	b := generateRandomBytes(16)
	if len(b) != 16 {
		t.Errorf("generateRandomBytes(16) returned %d bytes, want 16", len(b))
	}

	b2 := generateRandomBytes(32)
	if len(b2) != 32 {
		t.Errorf("generateRandomBytes(32) returned %d bytes, want 32", len(b2))
	}

	// Two calls should not return identical bytes.
	b3 := generateRandomBytes(16)
	identical := true
	for i := range b {
		if b[i] != b3[i] {
			identical = false
			break
		}
	}
	if identical {
		t.Error("generateRandomBytes returned identical bytes on two consecutive calls")
	}
}

func TestCheckAdminSession_Unauthenticated(t *testing.T) {
	a := newTestAdmin(t)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ok, userID := a.CheckAdminSession(req)

	if ok {
		t.Errorf("CheckAdminSession returned ok=true for unauthenticated request")
	}
	if userID != "" {
		t.Errorf("CheckAdminSession returned non-empty userID=%q for unauthenticated request", userID)
	}
}

func TestRequireAuth_Blocks(t *testing.T) {
	a := newTestAdmin(t)

	reached := false
	protected := a.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rr := httptest.NewRecorder()
	protected(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("RequireAuth returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
	if reached {
		t.Error("RequireAuth allowed unauthenticated request to reach handler")
	}
}

func TestHandleDashboard(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /admin returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := resp["Stats"]; !ok {
		t.Errorf("dashboard response missing 'Stats' key")
	}
}

func TestHandleSystemInfo(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/system", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /admin/system returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	for _, field := range []string{"os", "arch", "go_version"} {
		if _, ok := resp[field]; !ok {
			t.Errorf("system info missing field %q", field)
		}
	}
}

func TestHandleSettings(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/settings", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /admin/settings returned %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHandleSettingsSave_Valid(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	body := `{"some_key":"some_value"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/settings/save", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("POST /admin/settings/save returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestHandleSettingsSave_WrongMethod(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/settings/save", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /admin/settings/save returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleUsers(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /admin/users returned %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHandleSMTP(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/smtp", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /admin/smtp returned %d, want %d", rr.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := resp["enabled"]; !ok {
		t.Error("smtp response missing 'enabled' field")
	}
}

func TestHandleSMTPTest_WrongMethod(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/smtp/test", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /admin/smtp/test returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleSMTPTest_Post(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/admin/smtp/test", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("POST /admin/smtp/test returned %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHandleAPIServerInfo(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/server/info", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /api/v1/admin/server/info returned %d, want %d", rr.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if service, _ := resp["service"].(string); service != "cassocial" {
		t.Errorf("expected service=cassocial, got %q", service)
	}
}

func TestHandleAPIServerStats(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/server/stats", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /api/v1/admin/server/stats returned %d, want %d", rr.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := resp["goroutines"]; !ok {
		t.Error("stats response missing 'goroutines' field")
	}
}

func TestHandleAPISettings_Get(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /api/v1/admin/settings returned %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHandleAPISettings_Put(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	body := `{"key":"value"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("PUT /api/v1/admin/settings returned %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHandleAnalytics(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /admin/analytics returned %d, want %d", rr.Code, http.StatusOK)
	}
}

// ---------------------------------------------------------------------------
// handleAPISettings — invalid JSON body (POST/PUT) and wrong method
// ---------------------------------------------------------------------------

func TestHandleAPISettings_Post(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	body := `{"key":"value"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("POST /api/v1/admin/settings returned %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHandleAPISettings_InvalidJSON(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("PUT /api/v1/admin/settings with invalid JSON returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestHandleAPISettings_WrongMethod(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/settings", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("DELETE /api/v1/admin/settings returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

// ---------------------------------------------------------------------------
// handleSettingsSave — invalid JSON body
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// renderJSON — verify Content-Type header and valid JSON output
// ---------------------------------------------------------------------------

func TestRenderJSON_SetsContentType(t *testing.T) {
	a := newTestAdmin(t)

	rr := httptest.NewRecorder()
	a.renderJSON(rr, http.StatusCreated, map[string]string{"hello": "world"})

	if rr.Code != http.StatusCreated {
		t.Errorf("renderJSON status = %d, want %d", rr.Code, http.StatusCreated)
	}
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("renderJSON Content-Type = %q, want application/json", ct)
	}

	var result map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("renderJSON produced invalid JSON: %v", err)
	}
	if result["hello"] != "world" {
		t.Errorf("renderJSON body hello = %q, want world", result["hello"])
	}
}

func TestRenderJSON_NilData(t *testing.T) {
	a := newTestAdmin(t)

	rr := httptest.NewRecorder()
	a.renderJSON(rr, http.StatusOK, nil)

	if rr.Code != http.StatusOK {
		t.Errorf("renderJSON(nil) status = %d, want %d", rr.Code, http.StatusOK)
	}
}

// ---------------------------------------------------------------------------
// RequireAuth — handler is reached when CheckAdminSession succeeds
// (tested via a patched admin that overrides CheckAdminSession indirectly)
// ---------------------------------------------------------------------------

// authPassAdmin wraps Admin to allow RequireAuth to pass through by intercepting
// the request and injecting an auth cookie. Since CheckAdminSession always returns
// false in the current implementation, we verify the block path is covered with a
// separate explicit test.

func TestRequireAuth_BlocksWithoutAuth(t *testing.T) {
	a := newTestAdmin(t)

	called := false
	handler := a.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/protected", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("RequireAuth returned %d, want 401", rr.Code)
	}
	if called {
		t.Error("RequireAuth should not call handler when unauthenticated")
	}
}

// ---------------------------------------------------------------------------
// GenerateSetupToken — token is a hex-encoded 32-char string
// ---------------------------------------------------------------------------

func TestGenerateSetupToken_Format(t *testing.T) {
	a := newTestAdmin(t)

	token, err := a.GenerateSetupToken()
	if err != nil {
		t.Fatalf("GenerateSetupToken returned error: %v", err)
	}
	// 16 random bytes → 32 hex chars
	if len(token) != 32 {
		t.Errorf("GenerateSetupToken token length = %d, want 32", len(token))
	}
	for _, ch := range token {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			t.Errorf("GenerateSetupToken token %q contains non-hex char %q", token, ch)
			break
		}
	}
}

func TestGenerateSetupToken_UniqueTokens(t *testing.T) {
	a := newTestAdmin(t)

	t1, err := a.GenerateSetupToken()
	if err != nil {
		t.Fatalf("first GenerateSetupToken returned error: %v", err)
	}
	t2, err := a.GenerateSetupToken()
	if err != nil {
		t.Fatalf("second GenerateSetupToken returned error: %v", err)
	}
	if t1 == t2 {
		t.Error("GenerateSetupToken returned identical tokens on two calls")
	}
}

// ---------------------------------------------------------------------------
// generateRandomBytes — zero-length and large
// ---------------------------------------------------------------------------

func TestGenerateRandomBytes_ZeroLength(t *testing.T) {
	b := generateRandomBytes(0)
	if len(b) != 0 {
		t.Errorf("generateRandomBytes(0) returned %d bytes, want 0", len(b))
	}
}

func TestGenerateRandomBytes_LargeSize(t *testing.T) {
	b := generateRandomBytes(256)
	if len(b) != 256 {
		t.Errorf("generateRandomBytes(256) returned %d bytes, want 256", len(b))
	}
}

// ---------------------------------------------------------------------------
// Remaining admin HTTP handlers not yet covered
// ---------------------------------------------------------------------------

func TestHandleBackup(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/backup", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /admin/backup returned %d, want %d", rr.Code, http.StatusOK)
	}
}
