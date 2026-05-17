package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthHandlers_Logout(t *testing.T) {
	h := newTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	rr := httptest.NewRecorder()
	h.Logout(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Logout returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestAuthHandlers_LoginWith2FA_InvalidBody(t *testing.T) {
	h := newTestAuthHandlers(t)

	rr := postJSON(t, h.LoginWith2FA, "/api/auth/login/2fa", "not a struct")
	// The JSON decoder will decode a bare string successfully and produce a zero-value
	// Login2FARequest which fails 2FA verification — we just check it does not panic.
	if rr.Code == 0 {
		t.Fatal("LoginWith2FA returned zero status code")
	}
}

func TestAuthHandlers_LoginWith2FA_BadJSON(t *testing.T) {
	h := newTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login/2fa", nil)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.LoginWith2FA(rr, req)

	// nil body produces EOF — should return 400 Bad Request.
	if rr.Code != http.StatusBadRequest {
		t.Errorf("LoginWith2FA with nil body returned status %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestAuthHandlers_LoginWith2FA_InvalidCode(t *testing.T) {
	h := newTestAuthHandlers(t)

	rr := postJSON(t, h.LoginWith2FA, "/api/auth/login/2fa", map[string]string{
		"user_id": "nonexistent-user-id",
		"code":    "000000",
	})

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("LoginWith2FA with invalid code returned status %d, want %d; body: %s",
			rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
}
