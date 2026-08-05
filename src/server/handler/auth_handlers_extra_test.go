package handler

// Tests for the zero-coverage auth handler paths:
// RefreshToken, ForgotPassword, ResetPassword, VerifyEmail,
// Enable2FA, Verify2FA, Disable2FA.

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casapps/cassocial/src/server"
	"github.com/casapps/cassocial/src/server/model"
	"github.com/casapps/cassocial/src/server/store"
	"github.com/casapps/cassocial/src/service"
)

// newTestMailer returns a disabled Mailer suitable for tests that don't
// exercise SMTP behavior directly.
func newTestMailer(t *testing.T) *service.Mailer {
	t.Helper()
	mailer, err := service.NewMailer(nil, "Test App", "https://test.example")
	if err != nil {
		t.Fatalf("service.NewMailer returned error: %v", err)
	}
	return mailer
}

// newEnabledTestMailer returns a Mailer with a syntactically-valid SMTP
// config so IsEnabled() reports true — used by tests that exercise the
// SMTP-configured code path (e.g. ForgotPassword enumeration-safety
// behavior) without requiring a live SMTP server; SendPasswordReset send
// errors are logged and swallowed by the handler, never asserted here.
func newEnabledTestMailer(t *testing.T) *service.Mailer {
	t.Helper()
	cfg := &model.SMTPConfig{
		Host:        "smtp.test.example",
		Port:        587,
		FromAddress: "no-reply@test.example",
		Enabled:     true,
	}
	mailer, err := service.NewMailer(cfg, "Test App", "https://test.example")
	if err != nil {
		t.Fatalf("service.NewMailer returned error: %v", err)
	}
	if !mailer.IsEnabled() {
		t.Fatal("expected test mailer to report IsEnabled() == true")
	}
	return mailer
}

// hashTestToken returns the SHA-256 hex of a raw token string for test injection.
func hashTestToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// ---- RefreshToken ----

func TestRefreshToken_MissingHeader(t *testing.T) {
	h := newTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	rr := httptest.NewRecorder()
	h.RefreshToken(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("RefreshToken with no Authorization header returned %d, want %d; body: %s",
			rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
}

func TestRefreshToken_TooShortHeader(t *testing.T) {
	// Header present but shorter than 8 chars (i.e. no real token after "Bearer ")
	h := newTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.Header.Set("Authorization", "Bearer")
	rr := httptest.NewRecorder()
	h.RefreshToken(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("RefreshToken with short Authorization header returned %d, want %d; body: %s",
			rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	h := newTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.Header.Set("Authorization", "Bearer totally.invalid.jwt")
	rr := httptest.NewRecorder()
	h.RefreshToken(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("RefreshToken with invalid JWT returned %d, want %d; body: %s",
			rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
}

func TestRefreshToken_ValidToken(t *testing.T) {
	h := newTestAuthHandlers(t)

	// Disable email verification so login works.
	if _, err := h.db.Exec(`UPDATE settings SET value = 'false' WHERE key = 'email_verification_required'`); err != nil {
		t.Fatalf("disabling email verification: %v", err)
	}

	// Register and login to get a real token.
	rr := postJSON(t, h.Register, "/api/auth/register", map[string]string{
		"username": "refreshuser",
		"email":    "refreshuser@example.com",
		"password": "ValidPass1",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Register returned %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = postJSON(t, h.Login, "/api/auth/login", map[string]string{
		"username": "refreshuser",
		"password": "ValidPass1",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("Login returned %d; body: %s", rr.Code, rr.Body.String())
	}

	var loginResp map[string]interface{}
	if err := jsonDecode(rr.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	token, _ := loginResp["token"].(string)
	if token == "" {
		t.Fatal("login response missing token")
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	h.RefreshToken(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("RefreshToken with valid token returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

// ---- ForgotPassword ----

func TestForgotPassword_InvalidBody(t *testing.T) {
	h := newTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/forgot-password", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ForgotPassword(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("ForgotPassword with invalid JSON returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestForgotPassword_UnknownEmail(t *testing.T) {
	// Unknown email must still return 200 (enumeration mitigation).
	h := newTestAuthHandlersEnabled(t)

	rr := postJSON(t, h.ForgotPassword, "/api/auth/forgot-password", map[string]string{
		"email": "nobody@nowhere.example",
	})

	if rr.Code != http.StatusOK {
		t.Errorf("ForgotPassword unknown email returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestForgotPassword_KnownEmail(t *testing.T) {
	h := newTestAuthHandlersEnabled(t)

	// Register a user so the email is known.
	rr := postJSON(t, h.Register, "/api/auth/register", map[string]string{
		"username": "forgotuser",
		"email":    "forgot@example.com",
		"password": "ValidPass1",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Register returned %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = postJSON(t, h.ForgotPassword, "/api/auth/forgot-password", map[string]string{
		"email": "forgot@example.com",
	})

	// Whether email exists or not, the status must be 200 (no enumeration).
	if rr.Code != http.StatusOK {
		t.Errorf("ForgotPassword known email returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

// ---- ResetPassword ----

func TestResetPassword_InvalidBody(t *testing.T) {
	h := newTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ResetPassword(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("ResetPassword with invalid JSON returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestResetPassword_InvalidToken(t *testing.T) {
	h := newTestAuthHandlers(t)

	rr := postJSON(t, h.ResetPassword, "/api/auth/reset-password", map[string]string{
		"token":    "completely-bogus-token",
		"password": "NewValidPass1",
	})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("ResetPassword with invalid token returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestResetPassword_WeakPassword(t *testing.T) {
	// Even with a plausible token format, a weak password must be rejected.
	h := newTestAuthHandlers(t)

	rr := postJSON(t, h.ResetPassword, "/api/auth/reset-password", map[string]string{
		"token":    "bogus-token",
		"password": "weak",
	})

	// Must fail — either bad token (400) or weak password (400).
	if rr.Code == http.StatusOK {
		t.Errorf("ResetPassword with weak password returned 200, want 4xx; body: %s", rr.Body.String())
	}
}

// ---- VerifyEmail ----

func TestVerifyEmail_MissingToken(t *testing.T) {
	h := newTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/verify-email/", nil)
	// PathValue("token") returns "" when path value is absent.
	rr := httptest.NewRecorder()
	h.VerifyEmail(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("VerifyEmail with missing token returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	h := newTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/verify-email/bogus", nil)
	req.SetPathValue("token", "bogus-invalid-token")
	rr := httptest.NewRecorder()
	h.VerifyEmail(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("VerifyEmail with invalid token returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

// ---- Enable2FA ----

func TestEnable2FA_Unauthenticated(t *testing.T) {
	h := newTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/2fa/enable", nil)
	rr := httptest.NewRecorder()
	h.Enable2FA(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Enable2FA unauthenticated returned %d, want %d; body: %s",
			rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
}

func TestEnable2FA_UserNotFound(t *testing.T) {
	h := newTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/2fa/enable", nil)
	req = req.WithContext(context.WithValue(req.Context(), server.ContextKeyUserID, "nonexistent-user-id"))
	rr := httptest.NewRecorder()
	h.Enable2FA(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Enable2FA with nonexistent user returned %d, want %d; body: %s",
			rr.Code, http.StatusNotFound, rr.Body.String())
	}
}

func TestEnable2FA_AlreadyEnabled(t *testing.T) {
	h := newTestAuthHandlers(t)

	// Create a user with 2FA already enabled.
	userID := generateUUID()
	_, err := h.db.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, status, email_verified, two_factor_enabled, two_factor_secret, created_at, updated_at)
		 VALUES (?, ?, ?, ?, 'user', 'active', 1, 1, 'SOMESECRET', datetime('now'), datetime('now'))`,
		userID, "twofa_already", "twofa_already@example.com", "argon2id-placeholder",
	)
	if err != nil {
		t.Fatalf("insert user with 2FA enabled: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/2fa/enable", nil)
	req = req.WithContext(context.WithValue(req.Context(), server.ContextKeyUserID, userID))
	rr := httptest.NewRecorder()
	h.Enable2FA(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Enable2FA when already enabled returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestEnable2FA_Success(t *testing.T) {
	h := newTestAuthHandlers(t)

	userID := generateUUID()
	_, err := h.db.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, status, email_verified, two_factor_enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, 'user', 'active', 1, 0, datetime('now'), datetime('now'))`,
		userID, "twofa_new", "twofa_new@example.com", "argon2id-placeholder",
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/2fa/enable", nil)
	req = req.WithContext(context.WithValue(req.Context(), server.ContextKeyUserID, userID))
	rr := httptest.NewRecorder()
	h.Enable2FA(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Enable2FA success returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

// ---- Verify2FA ----

func TestVerify2FA_Unauthenticated(t *testing.T) {
	h := newTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/2fa/verify", nil)
	rr := httptest.NewRecorder()
	h.Verify2FA(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Verify2FA unauthenticated returned %d, want %d; body: %s",
			rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
}

func TestVerify2FA_InvalidBody(t *testing.T) {
	h := newTestAuthHandlers(t)

	userID := generateUUID()
	_, _ = h.db.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, status, email_verified, two_factor_enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, 'user', 'active', 1, 0, datetime('now'), datetime('now'))`,
		userID, "v2fa_badjson", "v2fa_badjson@example.com", "argon2id-placeholder",
	)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/2fa/verify", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), server.ContextKeyUserID, userID))
	rr := httptest.NewRecorder()
	h.Verify2FA(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Verify2FA with invalid JSON returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestVerify2FA_InvalidCode(t *testing.T) {
	h := newTestAuthHandlers(t)

	userID := generateUUID()
	_, _ = h.db.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, status, email_verified, two_factor_enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, 'user', 'active', 1, 0, datetime('now'), datetime('now'))`,
		userID, "v2fa_badcode", "v2fa_badcode@example.com", "argon2id-placeholder",
	)

	rr := postJSON(t, func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), server.ContextKeyUserID, userID))
		h.Verify2FA(w, r)
	}, "/api/auth/2fa/verify", map[string]string{
		"code":   "000000",
		"secret": "INVALIDSECRET",
	})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Verify2FA with invalid code returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

// ---- Disable2FA ----

func TestDisable2FA_Unauthenticated(t *testing.T) {
	h := newTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/2fa/disable", nil)
	rr := httptest.NewRecorder()
	h.Disable2FA(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Disable2FA unauthenticated returned %d, want %d; body: %s",
			rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
}

func TestDisable2FA_InvalidBody(t *testing.T) {
	h := newTestAuthHandlers(t)

	userID := generateUUID()
	_, _ = h.db.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, status, email_verified, two_factor_enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, 'user', 'active', 1, 0, datetime('now'), datetime('now'))`,
		userID, "d2fa_badjson", "d2fa_badjson@example.com", "argon2id-placeholder",
	)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/2fa/disable", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), server.ContextKeyUserID, userID))
	rr := httptest.NewRecorder()
	h.Disable2FA(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Disable2FA with invalid JSON returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestDisable2FA_UserNotFound(t *testing.T) {
	h := newTestAuthHandlers(t)

	rr := postJSON(t, func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), server.ContextKeyUserID, "nonexistent-id"))
		h.Disable2FA(w, r)
	}, "/api/auth/2fa/disable", map[string]string{
		"code": "000000",
	})

	if rr.Code != http.StatusNotFound {
		t.Errorf("Disable2FA user not found returned %d, want %d; body: %s",
			rr.Code, http.StatusNotFound, rr.Body.String())
	}
}

func TestDisable2FA_InvalidCode(t *testing.T) {
	h := newTestAuthHandlers(t)

	// Insert a user with 2FA enabled so the handler reaches Verify2FACode.
	userID := generateUUID()
	_, err := h.db.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, status, email_verified, two_factor_enabled, two_factor_secret, created_at, updated_at)
		 VALUES (?, ?, ?, ?, 'user', 'active', 1, 1, 'INVALIDSECRET', datetime('now'), datetime('now'))`,
		userID, "d2fa_badcode", "d2fa_badcode@example.com", "argon2id-placeholder",
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	rr := postJSON(t, func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), server.ContextKeyUserID, userID))
		h.Disable2FA(w, r)
	}, "/api/auth/2fa/disable", map[string]string{
		"code": "000000",
	})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Disable2FA with invalid code returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

// ---- Register additional branches ----

func TestRegister_RegistrationDisabled(t *testing.T) {
	h := newTestAuthHandlers(t)

	if _, err := h.db.Exec(`UPDATE settings SET value = 'false' WHERE key = 'registration_enabled'`); err != nil {
		t.Fatalf("disabling registration: %v", err)
	}

	rr := postJSON(t, h.Register, "/api/auth/register", map[string]string{
		"username": "shouldnotexist",
		"email":    "shouldnotexist@example.com",
		"password": "ValidPass1",
	})

	if rr.Code != http.StatusForbidden {
		t.Errorf("Register with registration disabled returned %d, want %d; body: %s",
			rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	h := newTestAuthHandlers(t)

	rr := postJSON(t, h.Register, "/api/auth/register", map[string]string{
		"username": "dupuser",
		"email":    "dupuser1@example.com",
		"password": "ValidPass1",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("first Register returned %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = postJSON(t, h.Register, "/api/auth/register", map[string]string{
		"username": "dupuser",
		"email":    "dupuser2@example.com",
		"password": "ValidPass1",
	})
	if rr.Code != http.StatusConflict {
		t.Errorf("Register with duplicate username returned %d, want %d; body: %s",
			rr.Code, http.StatusConflict, rr.Body.String())
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	h := newTestAuthHandlers(t)

	rr := postJSON(t, h.Register, "/api/auth/register", map[string]string{
		"username": "dupemailuser1",
		"email":    "dupemail@example.com",
		"password": "ValidPass1",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("first Register returned %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = postJSON(t, h.Register, "/api/auth/register", map[string]string{
		"username": "dupemailuser2",
		"email":    "dupemail@example.com",
		"password": "ValidPass1",
	})
	if rr.Code != http.StatusConflict {
		t.Errorf("Register with duplicate email returned %d, want %d; body: %s",
			rr.Code, http.StatusConflict, rr.Body.String())
	}
}

// ---- Login additional branches ----

func TestLogin_Success(t *testing.T) {
	h := newTestAuthHandlers(t)

	// Disable email verification so login works.
	if _, err := h.db.Exec(`UPDATE settings SET value = 'false' WHERE key = 'email_verification_required'`); err != nil {
		t.Fatalf("disabling email verification: %v", err)
	}

	rr := postJSON(t, h.Register, "/api/auth/register", map[string]string{
		"username": "loginuser",
		"email":    "loginuser@example.com",
		"password": "ValidPass1",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Register returned %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = postJSON(t, h.Login, "/api/auth/login", map[string]string{
		"username": "loginuser",
		"password": "ValidPass1",
	})
	if rr.Code != http.StatusOK {
		t.Errorf("Login returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := jsonDecode(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if token, _ := resp["token"].(string); token == "" {
		t.Error("Login response missing token field")
	}
}

func TestLogin_UserNotActive(t *testing.T) {
	h := newTestAuthHandlers(t)

	// Insert a suspended user directly.
	userID := generateUUID()
	_, err := h.db.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, status, email_verified, two_factor_enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, 'user', 'suspended', 1, 0, datetime('now'), datetime('now'))`,
		userID, "suspendeduser", "suspended@example.com", "argon2id-placeholder",
	)
	if err != nil {
		t.Fatalf("insert suspended user: %v", err)
	}

	rr := postJSON(t, h.Login, "/api/auth/login", map[string]string{
		"username": "suspendeduser",
		"password": "anything",
	})
	// Suspended user or wrong password both return 401; we just verify it is not 200.
	if rr.Code == http.StatusOK {
		t.Errorf("Login for suspended user returned 200, want 4xx; body: %s", rr.Body.String())
	}
}

func TestLogin_EmailNotVerified(t *testing.T) {
	h := newTestAuthHandlers(t)

	// Enable email verification requirement.
	if _, err := h.db.Exec(`UPDATE settings SET value = 'true' WHERE key = 'email_verification_required'`); err != nil {
		t.Fatalf("enabling email verification: %v", err)
	}

	// Register a user (email_verified defaults to 0 when verification required).
	rr := postJSON(t, h.Register, "/api/auth/register", map[string]string{
		"username": "unverifuser",
		"email":    "unverif@example.com",
		"password": "ValidPass1",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Register returned %d; body: %s", rr.Code, rr.Body.String())
	}

	// Make sure the user's email is not marked verified.
	if _, err := h.db.Exec(`UPDATE users SET email_verified = 0 WHERE username = 'unverifuser'`); err != nil {
		t.Fatalf("clearing email_verified: %v", err)
	}

	rr = postJSON(t, h.Login, "/api/auth/login", map[string]string{
		"username": "unverifuser",
		"password": "ValidPass1",
	})
	if rr.Code != http.StatusForbidden {
		t.Errorf("Login for unverified email returned %d, want %d; body: %s",
			rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestLogin_2FARequired(t *testing.T) {
	h := newTestAuthHandlers(t)

	// Disable email verification.
	if _, err := h.db.Exec(`UPDATE settings SET value = 'false' WHERE key = 'email_verification_required'`); err != nil {
		t.Fatalf("disabling email verification: %v", err)
	}

	// Insert a user with 2FA enabled.
	userID := generateUUID()
	_, err := h.db.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, status, email_verified, two_factor_enabled, two_factor_secret, created_at, updated_at)
		 VALUES (?, ?, ?, ?, 'user', 'active', 1, 1, 'AAAAAAAAAAAAAAAA', datetime('now'), datetime('now'))`,
		userID, "twofauser", "twofa@example.com", "placeholder",
	)
	if err != nil {
		t.Fatalf("insert 2FA user: %v", err)
	}

	// We can't verify the password with a placeholder hash, so test the
	// 2FA-required path via direct DB setup with a real hashed password instead.
	// Register the user properly then enable 2FA.
	user, err := h.auth.Register("twofaloginuser", "twofalogin@example.com", "ValidPass1")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	_, err = h.db.Exec(
		`UPDATE users SET two_factor_enabled = 1, two_factor_secret = 'AAAAAAAAAAAAAAAA', email_verified = 1 WHERE id = ?`,
		user.ID,
	)
	if err != nil {
		t.Fatalf("enable 2FA: %v", err)
	}

	// Disable email verification so the unverified check is bypassed.
	rr := postJSON(t, h.Login, "/api/auth/login", map[string]string{
		"username": "twofaloginuser",
		"password": "ValidPass1",
	})
	if rr.Code != http.StatusOK {
		t.Errorf("Login with 2FA enabled returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := jsonDecode(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if v, _ := resp["requires_2fa"].(bool); !v {
		t.Errorf("Login with 2FA enabled: expected requires_2fa=true, got %v", resp)
	}
}

// ---- ResetPassword success path ----

func TestResetPassword_Success(t *testing.T) {
	h := newTestAuthHandlers(t)

	// Register a user.
	rr := postJSON(t, h.Register, "/api/auth/register", map[string]string{
		"username": "resetsuccessuser",
		"email":    "resetsuccess@example.com",
		"password": "ValidPass1",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Register returned %d; body: %s", rr.Code, rr.Body.String())
	}

	// Obtain a real reset token via the auth service.
	token, err := h.auth.RequestPasswordReset("resetsuccess@example.com")
	if err != nil {
		t.Fatalf("RequestPasswordReset: %v", err)
	}
	if token == "" {
		t.Fatal("RequestPasswordReset returned empty token")
	}

	rr = postJSON(t, h.ResetPassword, "/api/auth/reset-password", map[string]string{
		"token":    token,
		"password": "NewValidPass1",
	})
	if rr.Code != http.StatusOK {
		t.Errorf("ResetPassword with valid token returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

// ---- VerifyEmail success path ----

func TestVerifyEmail_Success(t *testing.T) {
	h := newTestAuthHandlers(t)

	// Register a user.
	rr := postJSON(t, h.Register, "/api/auth/register", map[string]string{
		"username": "verifyemailuser",
		"email":    "verifyemail@example.com",
		"password": "ValidPass1",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Register returned %d; body: %s", rr.Code, rr.Body.String())
	}

	// Inject a verification token directly into the email_verification_tokens table.
	// Store the SHA-256 hash — never the raw token.
	rawToken := "testverifytoken123"
	tokenHash := hashTestToken(rawToken)
	expiry := "2099-01-01 00:00:00"

	// Retrieve the user ID to link the token.
	var userID string
	if err := h.db.QueryRow(`SELECT id FROM users WHERE username = 'verifyemailuser'`).Scan(&userID); err != nil {
		t.Fatalf("get user id: %v", err)
	}
	if _, err := h.db.Exec(
		`INSERT INTO email_verification_tokens (user_id, token_hash, expires_at) VALUES (?, ?, ?)`,
		userID, tokenHash, expiry,
	); err != nil {
		t.Fatalf("inject verification token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/verify-email/"+rawToken, nil)
	req.SetPathValue("token", rawToken)
	rr = httptest.NewRecorder()
	h.VerifyEmail(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("VerifyEmail with valid token returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	// Confirm email is now marked verified.
	var verified bool
	if err := h.db.QueryRow(`SELECT email_verified FROM users WHERE username = 'verifyemailuser'`).Scan(&verified); err != nil {
		t.Fatalf("query email_verified: %v", err)
	}
	if !verified {
		t.Error("email_verified should be true after VerifyEmail success")
	}
}

// TestForgotPassword_DBError verifies that when RequestPasswordReset returns a non-ErrUserNotFound
// error (triggered here by closing the DB), ForgotPassword returns 200 with a safe message
// (no enumeration) rather than exposing the error.
func TestForgotPassword_DBError(t *testing.T) {
	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	if err := db.RunMigrations(); err != nil {
		db.Close()
		t.Fatalf("RunMigrations: %v", err)
	}

	authSvc := server.NewAuth(db, "test-jwt-secret-for-tests")

	// Insert a user so GetUserByEmail finds it (before we close the DB).
	user, err := authSvc.Register("dberroruser", "dberror@example.com", "ValidPass1")
	if err != nil {
		db.Close()
		t.Fatalf("Register: %v", err)
	}
	_ = user

	h := NewAuthHandlers(authSvc, db, newEnabledTestMailer(t), "https://test.example")

	// Close the DB — GetUserByEmail will now return a DB error (not ErrUserNotFound),
	// causing RequestPasswordReset to propagate a non-nil error.
	db.Close()

	rr := postJSON(t, h.ForgotPassword, "/api/auth/forgot-password", map[string]string{
		"email": "dberror@example.com",
	})

	// The handler must return 200 regardless (no enumeration).
	if rr.Code != http.StatusOK {
		t.Errorf("ForgotPassword (DB error) returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

// TestResetPassword_DBError triggers the default error branch in ResetPassword by
// closing the DB so that auth.ResetPassword returns an unexpected error.
func TestResetPassword_DBError(t *testing.T) {
	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	if err := db.RunMigrations(); err != nil {
		db.Close()
		t.Fatalf("RunMigrations: %v", err)
	}

	authSvc := server.NewAuth(db, "test-jwt-secret-for-tests")

	// Register a user and get a valid reset token.
	if _, err := authSvc.Register("resetdberroruser", "resetdberror@example.com", "ValidPass1"); err != nil {
		db.Close()
		t.Fatalf("Register: %v", err)
	}
	token, err := authSvc.RequestPasswordReset("resetdberror@example.com")
	if err != nil || token == "" {
		db.Close()
		t.Fatalf("RequestPasswordReset: err=%v token=%q", err, token)
	}

	h := NewAuthHandlers(authSvc, db, newTestMailer(t), "https://test.example")

	// Close the DB so that the UPDATE inside auth.ResetPassword fails.
	db.Close()

	rr := postJSON(t, h.ResetPassword, "/api/auth/reset-password", map[string]string{
		"token":    token,
		"password": "NewValidPass1",
	})

	// Expect 500 Internal Server Error (the default case in the switch).
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("ResetPassword (DB error) returned %d, want %d; body: %s",
			rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
}

// jsonDecode decodes a JSON byte slice into v.
func jsonDecode(data []byte, v interface{}) error {
	return json.NewDecoder(bytes.NewReader(data)).Decode(v)
}
