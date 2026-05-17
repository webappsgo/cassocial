package handler

// Tests for the zero-coverage auth handler paths:
// RefreshToken, ForgotPassword, ResetPassword, VerifyEmail,
// Enable2FA, Verify2FA, Disable2FA.

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casapps/cassocial/src/server"
)

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
	h := newTestAuthHandlers(t)

	rr := postJSON(t, h.ForgotPassword, "/api/auth/forgot-password", map[string]string{
		"email": "nobody@nowhere.example",
	})

	if rr.Code != http.StatusOK {
		t.Errorf("ForgotPassword unknown email returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestForgotPassword_KnownEmail(t *testing.T) {
	h := newTestAuthHandlers(t)

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

// jsonDecode decodes a JSON byte slice into v.
func jsonDecode(data []byte, v interface{}) error {
	return json.NewDecoder(bytes.NewReader(data)).Decode(v)
}
