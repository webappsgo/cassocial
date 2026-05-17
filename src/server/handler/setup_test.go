package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// newTestSetupHandler creates a SetupHandler backed by an in-memory SQLite database
// and a config that writes to a temp directory.
func newTestSetupHandler(t *testing.T) (*SetupHandler, *store.DB) {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect returned error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	// Use a temp dir so Config.Save() can write without touching the real system.
	tmpDir := t.TempDir()
	cfg, err := config.Load(tmpDir, filepath.Join(tmpDir, "data"), filepath.Join(tmpDir, "log"))
	if err != nil {
		t.Fatalf("config.Load returned error: %v", err)
	}

	return NewSetupHandler(cfg, db), db
}

// ---- NewSetupHandler ----

func TestNewSetupHandler(t *testing.T) {
	h, _ := newTestSetupHandler(t)
	if h == nil {
		t.Fatal("NewSetupHandler returned nil")
	}
}

// ---- HandleSetupStatus ----

func TestSetupHandler_HandleSetupStatus_NotInitialized(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/setup/status", nil)
	rr := httptest.NewRecorder()
	h.HandleSetupStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupStatus returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["initialized"] != false {
		t.Errorf("expected initialized=false, got %v", resp["initialized"])
	}
	if _, ok := resp["steps"]; !ok {
		t.Error("response missing 'steps' field")
	}
}

func TestSetupHandler_HandleSetupStatus_Initialized(t *testing.T) {
	h, db := newTestSetupHandler(t)

	if err := db.SetSetting("initialized", "true"); err != nil {
		t.Fatalf("SetSetting returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/setup/status", nil)
	rr := httptest.NewRecorder()
	h.HandleSetupStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupStatus returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["initialized"] != true {
		t.Errorf("expected initialized=true, got %v", resp["initialized"])
	}
}

// ---- HandleSetupWelcome ----

func TestSetupHandler_HandleSetupWelcome(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/setup/welcome", nil)
	rr := httptest.NewRecorder()
	h.HandleSetupWelcome(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupWelcome returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if step, _ := resp["step"].(float64); step != 1 {
		t.Errorf("expected step=1, got %v", resp["step"])
	}
}

// ---- HandleSetupBasic ----

func TestSetupHandler_HandleSetupBasic_GET(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/setup/basic", nil)
	rr := httptest.NewRecorder()
	h.HandleSetupBasic(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupBasic GET returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if step, _ := resp["step"].(float64); step != 2 {
		t.Errorf("expected step=2, got %v", resp["step"])
	}
}

func TestSetupHandler_HandleSetupBasic_POST(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	body, _ := json.Marshal(map[string]string{
		"site_name":        "My Test Site",
		"site_description": "A test description",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/setup/basic", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupBasic(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupBasic POST returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["status"] != "success" {
		t.Errorf("expected status=success, got %v", resp["status"])
	}
}

func TestSetupHandler_HandleSetupBasic_POST_InvalidBody(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/setup/basic", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupBasic(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleSetupBasic POST with invalid body returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSetupHandler_HandleSetupBasic_MethodNotAllowed(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/setup/basic", nil)
	rr := httptest.NewRecorder()
	h.HandleSetupBasic(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("HandleSetupBasic DELETE returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

// ---- HandleSetupDomain ----

func TestSetupHandler_HandleSetupDomain_GET(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/setup/domain", nil)
	rr := httptest.NewRecorder()
	h.HandleSetupDomain(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupDomain GET returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if step, _ := resp["step"].(float64); step != 3 {
		t.Errorf("expected step=3, got %v", resp["step"])
	}
}

func TestSetupHandler_HandleSetupDomain_POST(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	body, _ := json.Marshal(map[string]interface{}{
		"domain": "https://example.com",
		"port":   8080,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/setup/domain", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupDomain(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupDomain POST returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestSetupHandler_HandleSetupDomain_POST_EmptyDomain(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	// Empty domain should still succeed (domain is optional in the handler)
	body, _ := json.Marshal(map[string]interface{}{
		"domain": "",
		"port":   8080,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/setup/domain", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupDomain(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupDomain POST with empty domain returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestSetupHandler_HandleSetupDomain_POST_InvalidBody(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/setup/domain", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupDomain(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleSetupDomain POST with invalid body returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSetupHandler_HandleSetupDomain_MethodNotAllowed(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/setup/domain", nil)
	rr := httptest.NewRecorder()
	h.HandleSetupDomain(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("HandleSetupDomain PATCH returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

// ---- HandleSetupEmail ----

func TestSetupHandler_HandleSetupEmail_GET(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/setup/email", nil)
	rr := httptest.NewRecorder()
	h.HandleSetupEmail(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupEmail GET returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if step, _ := resp["step"].(float64); step != 4 {
		t.Errorf("expected step=4, got %v", resp["step"])
	}
}

func TestSetupHandler_HandleSetupEmail_POST_Enabled(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	body, _ := json.Marshal(map[string]interface{}{
		"enabled":  true,
		"host":     "smtp.example.com",
		"port":     587,
		"username": "user@example.com",
		"password": "secret",
		"from":     "noreply@example.com",
		"tls":      true,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/setup/email", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupEmail(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupEmail POST enabled returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestSetupHandler_HandleSetupEmail_POST_Disabled(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	body, _ := json.Marshal(map[string]interface{}{
		"enabled": false,
		"host":    "",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/setup/email", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupEmail(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupEmail POST disabled returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestSetupHandler_HandleSetupEmail_POST_InvalidBody(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/setup/email", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupEmail(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleSetupEmail POST with invalid body returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSetupHandler_HandleSetupEmail_MethodNotAllowed(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodPut, "/api/setup/email", nil)
	rr := httptest.NewRecorder()
	h.HandleSetupEmail(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("HandleSetupEmail PUT returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

// ---- HandleSetupFeatures ----

func TestSetupHandler_HandleSetupFeatures_GET(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/setup/features", nil)
	rr := httptest.NewRecorder()
	h.HandleSetupFeatures(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupFeatures GET returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if step, _ := resp["step"].(float64); step != 5 {
		t.Errorf("expected step=5, got %v", resp["step"])
	}
}

func TestSetupHandler_HandleSetupFeatures_POST_AllowRegistration(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	body, _ := json.Marshal(map[string]interface{}{
		"allow_registration":    true,
		"max_profiles_per_user": 5,
		"max_links_per_profile": 100,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/setup/features", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupFeatures(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupFeatures POST allow returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestSetupHandler_HandleSetupFeatures_POST_DenyRegistration(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	body, _ := json.Marshal(map[string]interface{}{
		"allow_registration": false,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/setup/features", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupFeatures(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupFeatures POST deny returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestSetupHandler_HandleSetupFeatures_POST_InvalidBody(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/setup/features", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupFeatures(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleSetupFeatures POST with invalid body returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSetupHandler_HandleSetupFeatures_MethodNotAllowed(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/setup/features", nil)
	rr := httptest.NewRecorder()
	h.HandleSetupFeatures(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("HandleSetupFeatures DELETE returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

// ---- HandleSetupDatabase ----

func TestSetupHandler_HandleSetupDatabase_GET(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/setup/database", nil)
	rr := httptest.NewRecorder()
	h.HandleSetupDatabase(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupDatabase GET returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if step, _ := resp["step"].(float64); step != 6 {
		t.Errorf("expected step=6, got %v", resp["step"])
	}
}

func TestSetupHandler_HandleSetupDatabase_POST_Valid(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	body, _ := json.Marshal(map[string]interface{}{
		"driver": "sqlite",
		"name":   ":memory:",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/setup/database", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupDatabase(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupDatabase POST returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestSetupHandler_HandleSetupDatabase_POST_InvalidBody(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/setup/database", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupDatabase(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleSetupDatabase POST with invalid body returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSetupHandler_HandleSetupDatabase_MethodNotAllowed(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/setup/database", nil)
	rr := httptest.NewRecorder()
	h.HandleSetupDatabase(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("HandleSetupDatabase PATCH returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

// ---- HandleSetupComplete ----

func TestSetupHandler_HandleSetupComplete_MethodNotAllowed(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/setup/complete", nil)
	rr := httptest.NewRecorder()
	h.HandleSetupComplete(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("HandleSetupComplete GET returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestSetupHandler_HandleSetupComplete_InvalidBody(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/setup/complete", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupComplete(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleSetupComplete POST with invalid body returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSetupHandler_HandleSetupComplete_UsernameTooShort(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	body, _ := json.Marshal(map[string]string{
		"admin_username": "ab",
		"admin_email":    "admin@example.com",
		"admin_password": "ValidPass1",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/setup/complete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupComplete(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleSetupComplete with short username returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSetupHandler_HandleSetupComplete_PasswordTooShort(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	body, _ := json.Marshal(map[string]string{
		"admin_username": "adminuser",
		"admin_email":    "admin@example.com",
		"admin_password": "short",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/setup/complete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupComplete(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleSetupComplete with short password returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSetupHandler_HandleSetupComplete_Valid(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	// Config.Save() writes to a temp dir (already configured in newTestSetupHandler).
	body, _ := json.Marshal(map[string]string{
		"admin_username": "adminuser",
		"admin_email":    "admin@example.com",
		"admin_password": "SecurePass1",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/setup/complete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupComplete(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupComplete valid returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["status"] != "success" {
		t.Errorf("expected status=success, got %v", resp["status"])
	}
	admin, ok := resp["admin"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected admin object in response, got %T: %v", resp["admin"], resp["admin"])
	}
	if admin["username"] != "adminuser" {
		t.Errorf("expected admin.username=adminuser, got %v", admin["username"])
	}
}

func TestSetupHandler_HandleSetupComplete_SavesConfigFile(t *testing.T) {
	h, _ := newTestSetupHandler(t)

	body, _ := json.Marshal(map[string]string{
		"admin_username": "configtest",
		"admin_email":    "configtest@example.com",
		"admin_password": "SecurePass1",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/setup/complete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleSetupComplete(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("HandleSetupComplete returned %d; body: %s", rr.Code, rr.Body.String())
	}

	// Confirm the config file was written.
	configFile := filepath.Join(h.config.ConfigDir, "server.yml")
	if _, err := os.Stat(configFile); err != nil {
		t.Errorf("config file not written to %s: %v", configFile, err)
	}
}
