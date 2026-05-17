package handler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/casapps/cassocial/src/server"
)

// computeTOTP generates the current TOTP code for a given base32-encoded secret.
// This mirrors the logic in server.Auth.verifyTOTP / generateTOTP.
func computeTOTP(secret string) string {
	// Add padding if needed.
	if len(secret)%8 != 0 {
		secret += "========"[:8-len(secret)%8]
	}
	decoded, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "000000"
	}
	counter := time.Now().Unix() / 30
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))
	h := hmac.New(sha1.New, decoded)
	h.Write(buf)
	hash := h.Sum(nil)
	offset := hash[len(hash)-1] & 0x0F
	truncatedHash := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7FFFFFFF
	otp := truncatedHash % uint32(math.Pow10(6))
	return fmt.Sprintf("%06d", otp)
}

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

func TestLoginWith2FA_Success(t *testing.T) {
	h := newTestAuthHandlers(t)

	// Disable email verification.
	if _, err := h.db.Exec(`UPDATE settings SET value = 'false' WHERE key = 'email_verification_required'`); err != nil {
		t.Fatalf("disabling email verification: %v", err)
	}

	// Register a user with a properly hashed password.
	user, err := h.auth.Register("loginwith2fauser", "loginwith2fa@example.com", "ValidPass1")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	const testSecret = "JBSWY3DPEHPK3PXP"

	// Enable 2FA with a known secret.
	if _, err := h.db.Exec(
		`UPDATE users SET two_factor_enabled = 1, two_factor_secret = ?, email_verified = 1 WHERE id = ?`,
		testSecret, user.ID,
	); err != nil {
		t.Fatalf("enabling 2FA in DB: %v", err)
	}

	// Compute the live TOTP code.
	code := computeTOTP(testSecret)

	rr := postJSON(t, h.LoginWith2FA, "/api/auth/login/2fa", map[string]string{
		"user_id": user.ID,
		"code":    code,
	})

	if rr.Code != http.StatusOK {
		t.Errorf("LoginWith2FA success returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestLogin_InvalidBody(t *testing.T) {
	h := newTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Login with invalid JSON returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestDisable2FA_Success(t *testing.T) {
	h := newTestAuthHandlers(t)

	// Register a user.
	user, err := h.auth.Register("disable2fasuccessuser", "disable2fasuccess@example.com", "ValidPass1")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Use a well-known base32 secret (16 chars = valid base32 with padding).
	const testSecret = "JBSWY3DPEHPK3PXP"

	// Enable 2FA directly in the DB.
	if _, err := h.db.Exec(
		`UPDATE users SET two_factor_enabled = 1, two_factor_secret = ? WHERE id = ?`,
		testSecret, user.ID,
	); err != nil {
		t.Fatalf("enabling 2FA in DB: %v", err)
	}

	// Compute the live TOTP code for this secret.
	code := computeTOTP(testSecret)

	rr := postJSON(t, func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), server.ContextKeyUserID, user.ID))
		h.Disable2FA(w, r)
	}, "/api/auth/2fa/disable", map[string]string{
		"code": code,
	})

	if rr.Code != http.StatusOK {
		t.Errorf("Disable2FA success returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}
