package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// newTestUserHandler creates a UserHandler backed by an in-memory SQLite DB and default config.
func newTestUserHandler(t *testing.T) (*UserHandler, *store.DB) {
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
	cfg.Cassocial.AllowRegistration = true

	return NewUserHandler(cfg, db), db
}

func TestNewUserHandler_NotNil(t *testing.T) {
	h, _ := newTestUserHandler(t)
	if h == nil {
		t.Fatal("NewUserHandler returned nil")
	}
}

// HandleRegister — wrong HTTP method must be rejected.
func TestUserHandler_HandleRegister_WrongMethod(t *testing.T) {
	h, _ := newTestUserHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	rr := httptest.NewRecorder()
	h.HandleRegister(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("HandleRegister GET returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

// HandleRegister — registration disabled must return 403.
func TestUserHandler_HandleRegister_Disabled(t *testing.T) {
	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect returned error: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	cfg := &config.Config{}
	cfg.Cassocial.AllowRegistration = false
	h := NewUserHandler(cfg, db)

	body := map[string]string{"username": "user1", "email": "u@e.com", "password": "SecurePass1!"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleRegister(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("HandleRegister (disabled) returned %d, want %d", rr.Code, http.StatusForbidden)
	}
}

// HandleRegister — invalid JSON body must return 400.
func TestUserHandler_HandleRegister_InvalidBody(t *testing.T) {
	h, _ := newTestUserHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleRegister(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleRegister (invalid JSON) returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// HandleRegister — username too short must return 400.
func TestUserHandler_HandleRegister_UsernameTooShort(t *testing.T) {
	h, _ := newTestUserHandler(t)

	body := map[string]string{"username": "ab", "email": "a@b.com", "password": "SecurePass1!"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleRegister(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleRegister (short username) returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// HandleRegister — password too short must return 400.
func TestUserHandler_HandleRegister_PasswordTooShort(t *testing.T) {
	h, _ := newTestUserHandler(t)

	body := map[string]string{"username": "validuser", "email": "a@b.com", "password": "short"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleRegister(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleRegister (short password) returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// HandleRegister — empty email must return 400.
func TestUserHandler_HandleRegister_EmptyEmail(t *testing.T) {
	h, _ := newTestUserHandler(t)

	body := map[string]string{"username": "validuser", "email": "", "password": "SecurePass1!"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleRegister(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleRegister (empty email) returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// HandleRegister — valid registration must return 201.
func TestUserHandler_HandleRegister_Success(t *testing.T) {
	h, _ := newTestUserHandler(t)

	body := map[string]string{
		"username": "newreguser",
		"email":    "newreg@example.com",
		"password": "SecurePass1!",
	}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleRegister(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("HandleRegister (valid) returned %d, want %d; body: %s",
			rr.Code, http.StatusCreated, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["status"] != "success" {
		t.Errorf("HandleRegister status = %v, want \"success\"", resp["status"])
	}
}

// HandleVerifyEmail — missing token must return 400.
func TestUserHandler_HandleVerifyEmail_MissingToken(t *testing.T) {
	h, _ := newTestUserHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/verify-email", nil)
	rr := httptest.NewRecorder()
	h.HandleVerifyEmail(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleVerifyEmail (no token) returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// HandleVerifyEmail — invalid token must return 400.
func TestUserHandler_HandleVerifyEmail_InvalidToken(t *testing.T) {
	h, _ := newTestUserHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/verify-email?token=bogustoken", nil)
	rr := httptest.NewRecorder()
	h.HandleVerifyEmail(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleVerifyEmail (invalid token) returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// HandleRequestPasswordReset — wrong method must return 405.
func TestUserHandler_HandleRequestPasswordReset_WrongMethod(t *testing.T) {
	h, _ := newTestUserHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/forgot-password", nil)
	rr := httptest.NewRecorder()
	h.HandleRequestPasswordReset(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("HandleRequestPasswordReset GET returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

// HandleRequestPasswordReset — invalid JSON must return 400.
func TestUserHandler_HandleRequestPasswordReset_InvalidBody(t *testing.T) {
	h, _ := newTestUserHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleRequestPasswordReset(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleRequestPasswordReset (invalid JSON) returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// HandleRequestPasswordReset — unknown email must still return 200 (enumeration mitigation).
func TestUserHandler_HandleRequestPasswordReset_UnknownEmail(t *testing.T) {
	h, _ := newTestUserHandler(t)

	body := map[string]string{"email": "nobody@example.com"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleRequestPasswordReset(rr, req)

	// Enumeration mitigation: must always respond with 200.
	if rr.Code != http.StatusOK {
		t.Errorf("HandleRequestPasswordReset (unknown email) returned %d, want %d", rr.Code, http.StatusOK)
	}
}

// HandleResetPassword — wrong method must return 405.
func TestUserHandler_HandleResetPassword_WrongMethod(t *testing.T) {
	h, _ := newTestUserHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/reset-password", nil)
	rr := httptest.NewRecorder()
	h.HandleResetPassword(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("HandleResetPassword GET returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

// HandleResetPassword — invalid JSON must return 400.
func TestUserHandler_HandleResetPassword_InvalidBody(t *testing.T) {
	h, _ := newTestUserHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleResetPassword(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleResetPassword (invalid JSON) returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// HandleResetPassword — short new password must return 400.
func TestUserHandler_HandleResetPassword_ShortPassword(t *testing.T) {
	h, _ := newTestUserHandler(t)

	body := map[string]string{"token": "sometoken", "new_password": "short"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleResetPassword(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleResetPassword (short password) returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// HandleResetPassword — invalid/unknown token must return 400.
func TestUserHandler_HandleResetPassword_InvalidToken(t *testing.T) {
	h, _ := newTestUserHandler(t)

	body := map[string]string{"token": "bogustoken", "new_password": "NewSecurePass1!"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleResetPassword(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleResetPassword (invalid token) returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// HandleAccountSettings — missing auth must return 401.
func TestUserHandler_HandleAccountSettings_NoAuth(t *testing.T) {
	h, _ := newTestUserHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	rr := httptest.NewRecorder()
	h.HandleAccountSettings(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("HandleAccountSettings (no auth) returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// HandleAccountSettings GET — valid auth returns settings.
func TestUserHandler_HandleAccountSettings_GET_WithAuth(t *testing.T) {
	h, db := newTestUserHandler(t)

	// Use store.CreateUser directly so two_factor_secret is "" (not NULL).
	user := &store.User{
		ID:           "settings-user-get-001",
		Username:     "settingsuser2",
		Email:        "settings2@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	req = withUserID(req, user.ID)
	rr := httptest.NewRecorder()
	h.HandleAccountSettings(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleAccountSettings GET returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["username"] != "settingsuser2" {
		t.Errorf("HandleAccountSettings username = %v, want \"settingsuser2\"", body["username"])
	}
}

// HandleAccountSettings PUT — updates email field.
func TestUserHandler_HandleAccountSettings_PUT_WithAuth(t *testing.T) {
	h, db := newTestUserHandler(t)

	user := &store.User{
		ID:           "settings-user-put-001",
		Username:     "settingsuser3",
		Email:        "settings3@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	updates := map[string]string{"email": "newemail@example.com"}
	data, _ := json.Marshal(updates)
	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, user.ID)
	rr := httptest.NewRecorder()
	h.HandleAccountSettings(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleAccountSettings PUT returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

// HandleAccountSettings — unsupported method must return 405.
func TestUserHandler_HandleAccountSettings_WrongMethod(t *testing.T) {
	h, db := newTestUserHandler(t)

	user := &store.User{
		ID:           "settings-user-del-001",
		Username:     "settingsuser4",
		Email:        "settings4@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/settings", nil)
	req = withUserID(req, user.ID)
	rr := httptest.NewRecorder()
	h.HandleAccountSettings(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("HandleAccountSettings DELETE returned %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

// Handle2FASetup must return 501.
func TestUserHandler_Handle2FASetup_NotImplemented(t *testing.T) {
	h, _ := newTestUserHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/2fa/setup", nil)
	rr := httptest.NewRecorder()
	h.Handle2FASetup(rr, req)

	if rr.Code != http.StatusNotImplemented {
		t.Errorf("Handle2FASetup returned %d, want %d", rr.Code, http.StatusNotImplemented)
	}
}
