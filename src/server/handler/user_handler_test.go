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

// HandleVerifyEmail — valid token must mark email verified and return 200.
func TestUserHandler_HandleVerifyEmail_Success(t *testing.T) {
	h, db := newTestUserHandler(t)

	// Create a user first.
	user := &store.User{
		ID:            "verify-user-001",
		Username:      "verifyuser",
		Email:         "verify@example.com",
		PasswordHash:  "x",
		Role:          "user",
		Status:        "pending",
		EmailVerified: false,
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	// Generate a verification token using the helper.
	token, err := h.generateVerificationToken(user.Email)
	if err != nil {
		t.Fatalf("generateVerificationToken returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/verify-email?token="+token, nil)
	rr := httptest.NewRecorder()
	h.HandleVerifyEmail(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleVerifyEmail (valid token) returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["status"] != "success" {
		t.Errorf("HandleVerifyEmail status = %q, want success", resp["status"])
	}
}

// HandleVerifyEmail — getUserByID failure after token lookup returns 500.
// We simulate this by deleting the user after the token is created.
func TestUserHandler_HandleVerifyEmail_UserNotFound(t *testing.T) {
	h, db := newTestUserHandler(t)

	user := &store.User{
		ID:           "verify-user-002",
		Username:     "verifyuser2",
		Email:        "verify2@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "pending",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	token, err := h.generateVerificationToken(user.Email)
	if err != nil {
		t.Fatalf("generateVerificationToken returned error: %v", err)
	}

	// Delete the user so GetUserByID fails.
	if _, err := db.Exec(`DELETE FROM users WHERE id = ?`, user.ID); err != nil {
		t.Fatalf("DELETE user: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/verify-email?token="+token, nil)
	rr := httptest.NewRecorder()
	h.HandleVerifyEmail(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("HandleVerifyEmail (user deleted) returned %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

// generateVerificationToken — unknown email returns error.
func TestUserHandler_GenerateVerificationToken_UnknownEmail(t *testing.T) {
	h, _ := newTestUserHandler(t)

	_, err := h.generateVerificationToken("nobody@nowhere.invalid")
	if err == nil {
		t.Error("generateVerificationToken with unknown email: expected error, got nil")
	}
}

// generatePasswordResetToken — success path creates a token.
func TestUserHandler_GeneratePasswordResetToken_Success(t *testing.T) {
	h, db := newTestUserHandler(t)

	user := &store.User{
		ID:           "reset-token-user-001",
		Username:     "resetuser",
		Email:        "reset@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	token, err := h.generatePasswordResetToken(user.Email)
	if err != nil {
		t.Fatalf("generatePasswordResetToken returned error: %v", err)
	}
	if token == "" {
		t.Error("generatePasswordResetToken returned empty token")
	}
}

// generatePasswordResetToken — unknown email returns error.
func TestUserHandler_GeneratePasswordResetToken_UnknownEmail(t *testing.T) {
	h, _ := newTestUserHandler(t)

	_, err := h.generatePasswordResetToken("nobody@nowhere.invalid")
	if err == nil {
		t.Error("generatePasswordResetToken with unknown email: expected error, got nil")
	}
}

// HandleResetPassword — valid token resets password and returns 200.
func TestUserHandler_HandleResetPassword_Success(t *testing.T) {
	h, db := newTestUserHandler(t)

	user := &store.User{
		ID:           "reset-user-001",
		Username:     "resetpassuser",
		Email:        "resetpass@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	token, err := h.generatePasswordResetToken(user.Email)
	if err != nil {
		t.Fatalf("generatePasswordResetToken returned error: %v", err)
	}

	body := map[string]string{"token": token, "new_password": "NewSecurePass1!"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleResetPassword(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleResetPassword (valid token) returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["status"] != "success" {
		t.Errorf("HandleResetPassword status = %q, want success", resp["status"])
	}
}

// HandleResetPassword — user deleted after token created returns 500.
func TestUserHandler_HandleResetPassword_UserNotFound(t *testing.T) {
	h, db := newTestUserHandler(t)

	user := &store.User{
		ID:           "reset-user-002",
		Username:     "resetpassuser2",
		Email:        "resetpass2@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	token, err := h.generatePasswordResetToken(user.Email)
	if err != nil {
		t.Fatalf("generatePasswordResetToken returned error: %v", err)
	}

	// Delete user to trigger GetUserByID failure.
	if _, err := db.Exec(`DELETE FROM users WHERE id = ?`, user.ID); err != nil {
		t.Fatalf("DELETE user: %v", err)
	}

	body := map[string]string{"token": token, "new_password": "NewSecurePass1!"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleResetPassword(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("HandleResetPassword (user deleted) returned %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

// HandleAccountSettings POST — updates email field.
func TestUserHandler_HandleAccountSettings_POST_WithAuth(t *testing.T) {
	h, db := newTestUserHandler(t)

	user := &store.User{
		ID:           "settings-user-post-001",
		Username:     "settingsuserpost",
		Email:        "settingspost@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	updates := map[string]string{"email": "updated@example.com"}
	data, _ := json.Marshal(updates)
	req := httptest.NewRequest(http.MethodPost, "/settings", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, user.ID)
	rr := httptest.NewRecorder()
	h.HandleAccountSettings(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleAccountSettings POST returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

// HandleAccountSettings GET — invalid user ID returns 500.
func TestUserHandler_HandleAccountSettings_GET_InvalidUser(t *testing.T) {
	h, _ := newTestUserHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	req = withUserID(req, "nonexistent-user-id-xyz")
	rr := httptest.NewRecorder()
	h.HandleAccountSettings(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("HandleAccountSettings GET (bad user) returned %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

// HandleAccountSettings POST — invalid JSON returns 400.
func TestUserHandler_HandleAccountSettings_POST_InvalidBody(t *testing.T) {
	h, db := newTestUserHandler(t)

	user := &store.User{
		ID:           "settings-user-badjson-001",
		Username:     "settingsuserj",
		Email:        "settingsj@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/settings", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, user.ID)
	rr := httptest.NewRecorder()
	h.HandleAccountSettings(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleAccountSettings POST (bad JSON) returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// HandleAccountSettings POST — user deleted between auth and update returns 500.
func TestUserHandler_HandleAccountSettings_POST_UserNotFound(t *testing.T) {
	h, db := newTestUserHandler(t)

	user := &store.User{
		ID:           "settings-user-del2-001",
		Username:     "settingsdel2",
		Email:        "settingsdel2@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	// Delete user to trigger GetUserByID failure during POST.
	if _, err := db.Exec(`DELETE FROM users WHERE id = ?`, user.ID); err != nil {
		t.Fatalf("DELETE user: %v", err)
	}

	updates := map[string]string{"email": "new@example.com"}
	data, _ := json.Marshal(updates)
	req := httptest.NewRequest(http.MethodPost, "/settings", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, user.ID)
	rr := httptest.NewRecorder()
	h.HandleAccountSettings(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("HandleAccountSettings POST (user deleted) returned %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

// HandleRegister — duplicate username/email causes CreateUser to fail and returns 500.
func TestUserHandler_HandleRegister_DuplicateUser(t *testing.T) {
	h, _ := newTestUserHandler(t)

	body := map[string]string{
		"username": "dupuser",
		"email":    "dup@example.com",
		"password": "SecurePass1!",
	}
	data, _ := json.Marshal(body)

	// First registration should succeed.
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleRegister(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("first register returned %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	// Second registration with same email should fail at CreateUser.
	data2, _ := json.Marshal(map[string]string{
		"username": "dupuser2",
		"email":    "dup@example.com",
		"password": "SecurePass1!",
	})
	req2 := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(data2))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	h.HandleRegister(rr2, req2)
	if rr2.Code != http.StatusInternalServerError {
		t.Errorf("duplicate register returned %d, want %d; body: %s", rr2.Code, http.StatusInternalServerError, rr2.Body.String())
	}
}

// HandleRegister — username too long (>30 chars) must return 400.
func TestUserHandler_HandleRegister_UsernameTooLong(t *testing.T) {
	h, _ := newTestUserHandler(t)

	body := map[string]string{
		"username": "this_username_is_way_too_long_x",
		"email":    "a@b.com",
		"password": "SecurePass1!",
	}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleRegister(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleRegister (long username) returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// HandleVerifyEmail — UpdateUser failure returns 500.
// A trigger blocks UPDATE on the users table after the user is loaded.
func TestUserHandler_HandleVerifyEmail_UpdateUserFails(t *testing.T) {
	h, db := newTestUserHandler(t)

	user := &store.User{
		ID:           "verify-user-003",
		Username:     "verifyuser3",
		Email:        "verify3@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "pending",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	token, err := h.generateVerificationToken(user.Email)
	if err != nil {
		t.Fatalf("generateVerificationToken returned error: %v", err)
	}

	// Install a BEFORE UPDATE trigger that raises an error.
	if _, err := db.Exec(`CREATE TRIGGER block_user_update BEFORE UPDATE ON users BEGIN SELECT RAISE(ABORT,'update blocked'); END`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DROP TRIGGER IF EXISTS block_user_update`) })

	req := httptest.NewRequest(http.MethodGet, "/verify-email?token="+token, nil)
	rr := httptest.NewRecorder()
	h.HandleVerifyEmail(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("HandleVerifyEmail (UpdateUser fails) returned %d, want %d; body: %s",
			rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
}

// HandleResetPassword — UpdateUser failure returns 500.
func TestUserHandler_HandleResetPassword_UpdateUserFails(t *testing.T) {
	h, db := newTestUserHandler(t)

	user := &store.User{
		ID:           "reset-user-003",
		Username:     "resetpassuser3",
		Email:        "resetpass3@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	token, err := h.generatePasswordResetToken(user.Email)
	if err != nil {
		t.Fatalf("generatePasswordResetToken returned error: %v", err)
	}

	// Install a BEFORE UPDATE trigger that raises an error.
	if _, err := db.Exec(`CREATE TRIGGER block_user_update2 BEFORE UPDATE ON users BEGIN SELECT RAISE(ABORT,'update blocked'); END`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DROP TRIGGER IF EXISTS block_user_update2`) })

	body := map[string]string{"token": token, "new_password": "NewSecurePass1!"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleResetPassword(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("HandleResetPassword (UpdateUser fails) returned %d, want %d; body: %s",
			rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
}

// HandleAccountSettings POST — UpdateUser failure returns 500.
func TestUserHandler_HandleAccountSettings_POST_UpdateUserFails(t *testing.T) {
	h, db := newTestUserHandler(t)

	user := &store.User{
		ID:           "settings-user-updfail-001",
		Username:     "settingsupdfail",
		Email:        "settingsupdfail@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	// Install a BEFORE UPDATE trigger that raises an error.
	if _, err := db.Exec(`CREATE TRIGGER block_user_update3 BEFORE UPDATE ON users BEGIN SELECT RAISE(ABORT,'update blocked'); END`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DROP TRIGGER IF EXISTS block_user_update3`) })

	updates := map[string]string{"email": "new@example.com"}
	data, _ := json.Marshal(updates)
	req := httptest.NewRequest(http.MethodPost, "/settings", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, user.ID)
	rr := httptest.NewRecorder()
	h.HandleAccountSettings(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("HandleAccountSettings POST (UpdateUser fails) returned %d, want %d; body: %s",
			rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
}

// generateVerificationToken — CreateEmailVerificationToken failure returns error.
func TestUserHandler_GenerateVerificationToken_TokenCreateFails(t *testing.T) {
	h, db := newTestUserHandler(t)

	user := &store.User{
		ID:           "verify-token-fail-001",
		Username:     "tokenfailuser",
		Email:        "tokenfail@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "pending",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	// Rename the token table so CreateEmailVerificationToken fails.
	if _, err := db.Exec(`ALTER TABLE email_verification_tokens RENAME TO email_verification_tokens_bak`); err != nil {
		t.Fatalf("rename table: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`ALTER TABLE email_verification_tokens_bak RENAME TO email_verification_tokens`)
	})

	_, err := h.generateVerificationToken(user.Email)
	if err == nil {
		t.Error("generateVerificationToken with broken token table: expected error, got nil")
	}
}

// generatePasswordResetToken — CreatePasswordResetToken failure returns error.
func TestUserHandler_GeneratePasswordResetToken_TokenCreateFails(t *testing.T) {
	h, db := newTestUserHandler(t)

	user := &store.User{
		ID:           "reset-token-fail-001",
		Username:     "resettokenfailuser",
		Email:        "resettokenfail@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	// Rename the token table so CreatePasswordResetToken fails.
	if _, err := db.Exec(`ALTER TABLE password_reset_tokens RENAME TO password_reset_tokens_bak`); err != nil {
		t.Fatalf("rename table: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`ALTER TABLE password_reset_tokens_bak RENAME TO password_reset_tokens`)
	})

	_, err := h.generatePasswordResetToken(user.Email)
	if err == nil {
		t.Error("generatePasswordResetToken with broken token table: expected error, got nil")
	}
}

// HandleRequestPasswordReset — known email returns 200 (same message for enumeration safety).
func TestUserHandler_HandleRequestPasswordReset_KnownEmail(t *testing.T) {
	h, db := newTestUserHandler(t)

	user := &store.User{
		ID:           "pwreset-req-user-001",
		Username:     "pwresetuser",
		Email:        "pwreset@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	body := map[string]string{"email": user.Email}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleRequestPasswordReset(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleRequestPasswordReset (known email) returned %d, want %d", rr.Code, http.StatusOK)
	}
}
