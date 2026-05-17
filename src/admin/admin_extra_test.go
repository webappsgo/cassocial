package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

func TestHandleUserCreate_Post(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	body := `{"username":"newuser","email":"newuser@example.com","password":"ValidPass1"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/users/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("POST /admin/users/create returned %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["status"] != "success" {
		t.Errorf("expected status=success, got %q", resp["status"])
	}
}

func TestHandleUserCreate_WrongMethod(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/create", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /admin/users/create returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleUserEdit_Post(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	body := `{"role":"admin"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/users/edit", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("POST /admin/users/edit returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestHandleUserEdit_WrongMethod(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/edit", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /admin/users/edit returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleUserDelete_Post(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	body := `{"user_id":"some-id"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/users/delete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("POST /admin/users/delete returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestHandleUserDelete_WrongMethod(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/delete", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /admin/users/delete returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleProfiles(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/profiles", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /admin/profiles returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := resp["profiles"]; !ok {
		t.Error("profiles response missing 'profiles' field")
	}
}

func TestHandleProfileView(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/profiles/view", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /admin/profiles/view returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestHandleServices(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/services", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /admin/services returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := resp["services"]; !ok {
		t.Error("services response missing 'services' field")
	}
}

func TestHandleThemes(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/themes", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /admin/themes returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := resp["themes"]; !ok {
		t.Error("themes response missing 'themes' field")
	}
}

func TestHandleBackupCreate_Post(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/admin/backup/create", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("POST /admin/backup/create returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestHandleBackupCreate_WrongMethod(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/backup/create", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /admin/backup/create returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleBackupRestore_Post(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/admin/backup/restore", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("POST /admin/backup/restore returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestHandleBackupRestore_WrongMethod(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/backup/restore", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /admin/backup/restore returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleMaintenance(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/maintenance", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /admin/maintenance returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := resp["enabled"]; !ok {
		t.Error("maintenance response missing 'enabled' field")
	}
}

func TestHandleMaintenanceToggle_Post(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/admin/maintenance/toggle", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("POST /admin/maintenance/toggle returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestHandleMaintenanceToggle_WrongMethod(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/maintenance/toggle", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /admin/maintenance/toggle returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleSecurity(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/security", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /admin/security returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := resp["ssl_enabled"]; !ok {
		t.Error("security response missing 'ssl_enabled' field")
	}
}

func TestHandleSettingsSave_InvalidJSON(t *testing.T) {
	a := newTestAdmin(t)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/admin/settings/save", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("POST /admin/settings/save with invalid JSON returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestRenderError(t *testing.T) {
	a := newTestAdmin(t)

	rr := httptest.NewRecorder()
	a.renderError(rr, http.StatusNotFound, "resource not found")

	if rr.Code != http.StatusNotFound {
		t.Errorf("renderError returned status %d, want %d", rr.Code, http.StatusNotFound)
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode renderError response: %v", err)
	}
	if resp["error"] != "resource not found" {
		t.Errorf("renderError body error=%q, want %q", resp["error"], "resource not found")
	}
}

func TestRenderError_ContentType(t *testing.T) {
	a := newTestAdmin(t)

	rr := httptest.NewRecorder()
	a.renderError(rr, http.StatusInternalServerError, "something went wrong")

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("renderError Content-Type=%q, want application/json", ct)
	}
}

// ---------------------------------------------------------------------------
// renderJSON — encode error path
// ---------------------------------------------------------------------------

// failWriter is an http.ResponseWriter whose Write method always fails.
type failWriter struct {
	header     http.Header
	statusCode int
}

func newFailWriter() *failWriter {
	return &failWriter{header: make(http.Header)}
}

func (f *failWriter) Header() http.Header         { return f.header }
func (f *failWriter) WriteHeader(code int)         { f.statusCode = code }
func (f *failWriter) Write(_ []byte) (int, error) { return 0, errors.New("write error") }

func TestRenderJSON_EncodeError(t *testing.T) {
	a := newTestAdmin(t)

	fw := newFailWriter()
	// json.Encoder.Encode writes to the writer; if Write fails the error is logged
	// and Encode returns an error. renderJSON must not panic.
	a.renderJSON(fw, http.StatusOK, map[string]string{"key": "value"})
	// We just verify it does not panic; the log output is not checked.
}

// ---------------------------------------------------------------------------
// RequireAuth — pass-through when CheckAdminSession returns true
// ---------------------------------------------------------------------------

func TestRequireAuth_PassThrough(t *testing.T) {
	a := newTestAdmin(t)

	// Inject a session checker that always grants access.
	orig := checkSessionFn
	checkSessionFn = func(_ *http.Request) (bool, string) { return true, "admin-id" }
	defer func() { checkSessionFn = orig }()

	called := false
	handler := a.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/protected", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if !called {
		t.Error("RequireAuth should call next handler when session is valid")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("RequireAuth pass-through returned %d, want 200", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// GenerateSetupToken — SetSetting error path
// ---------------------------------------------------------------------------

func TestGenerateSetupToken_DBError(t *testing.T) {
	// Use a DB without migrations so the settings table does not exist.
	tmp := t.TempDir()
	db, err := store.Connect("sqlite", tmp+"/noschema.db")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	defer db.Close()
	// Deliberately skip RunMigrations — settings table absent.

	cfg := &config.Config{}
	cfg.Server.Mode = "development"
	cfg.Database.Driver = "sqlite"
	a := New(cfg, db)

	_, tokenErr := a.GenerateSetupToken()
	if tokenErr == nil {
		t.Error("GenerateSetupToken with missing settings table should return error")
	}
}

// ---------------------------------------------------------------------------
// generateRandomBytes — error path via randReadFn injection
// ---------------------------------------------------------------------------

func TestGenerateRandomBytes_ReadError(t *testing.T) {
	// Inject a failing rand reader.
	orig := randReadFn
	randReadFn = func(b []byte) (int, error) { return 0, fmt.Errorf("injected rand failure") }
	defer func() { randReadFn = orig }()

	// generateRandomBytes must not panic even when the read fails.
	b := generateRandomBytes(16)
	// The bytes may be zeroed since the read failed, but the slice should still
	// have the requested length.
	if len(b) != 16 {
		t.Errorf("generateRandomBytes(16) returned %d bytes after error, want 16", len(b))
	}
}
