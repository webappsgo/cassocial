package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestMiddleware creates a Middleware backed by an in-memory SQLite database.
func newTestMiddleware(t *testing.T) *Middleware {
	t.Helper()
	a := newTestAuth(t)
	return NewMiddleware(a)
}

// makeTokenRequest creates an HTTP request with a Bearer token header.
func makeTokenRequest(token string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

// TestNewMiddleware verifies that NewMiddleware returns a non-nil instance.
func TestNewMiddleware(t *testing.T) {
	m := newTestMiddleware(t)
	if m == nil {
		t.Fatal("NewMiddleware returned nil")
	}
}

// TestRequireAuth_MissingToken verifies that a request without a token gets 401.
func TestRequireAuth_MissingToken(t *testing.T) {
	m := newTestMiddleware(t)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	m.RequireAuth(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("RequireAuth no token: status %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// TestRequireAuth_InvalidToken verifies that an invalid token gets 401.
func TestRequireAuth_InvalidToken(t *testing.T) {
	m := newTestMiddleware(t)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})

	req := makeTokenRequest("this-is-not-a-valid-jwt")
	rr := httptest.NewRecorder()
	m.RequireAuth(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("RequireAuth invalid token: status %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// TestRequireAuth_ValidToken verifies that a valid token passes through to next handler.
func TestRequireAuth_ValidToken(t *testing.T) {
	a := newTestAuth(t)
	m := NewMiddleware(a)
	user := registerTestUser(t, a, "mw-user", "mw@example.com", "ValidPass1")

	token, err := a.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		uid, ok := r.Context().Value(ContextKeyUserID).(string)
		if !ok || uid != user.ID {
			t.Errorf("context UserID = %q, want %q", uid, user.ID)
		}
	})

	req := makeTokenRequest(token)
	rr := httptest.NewRecorder()
	m.RequireAuth(next).ServeHTTP(rr, req)

	if !called {
		t.Error("RequireAuth with valid token should call next handler")
	}
}

// TestRequireAuth_CookieFallback verifies that a token in a cookie is accepted.
func TestRequireAuth_CookieFallback(t *testing.T) {
	a := newTestAuth(t)
	m := NewMiddleware(a)
	user := registerTestUser(t, a, "cookie-user", "cookie@example.com", "ValidPass1")

	token, err := a.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	rr := httptest.NewRecorder()
	m.RequireAuth(next).ServeHTTP(rr, req)

	if !called {
		t.Error("RequireAuth should accept token from cookie")
	}
}

// TestRequireRole_Allowed verifies that the correct role passes through.
func TestRequireRole_Allowed(t *testing.T) {
	m := newTestMiddleware(t)
	mwFn := m.RequireRole("admin")

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), ContextKeyRole, "admin")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	mwFn(next).ServeHTTP(rr, req)

	if !called {
		t.Error("RequireRole should call next for correct role")
	}
}

// TestRequireRole_WrongRole verifies that a mismatched role gets 403.
func TestRequireRole_WrongRole(t *testing.T) {
	m := newTestMiddleware(t)
	mwFn := m.RequireRole("admin")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called for wrong role")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), ContextKeyRole, "user")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	mwFn(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("RequireRole wrong role: status %d, want %d", rr.Code, http.StatusForbidden)
	}
}

// TestRequireRole_NoRole verifies that missing role context gets 401.
func TestRequireRole_NoRole(t *testing.T) {
	m := newTestMiddleware(t)
	mwFn := m.RequireRole("admin")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called with no role in context")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	mwFn(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("RequireRole no role: status %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// TestRequireAdmin_AdminPasses verifies that RequireAdmin allows an admin.
func TestRequireAdmin_AdminPasses(t *testing.T) {
	m := newTestMiddleware(t)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), ContextKeyRole, "admin")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	m.RequireAdmin(next).ServeHTTP(rr, req)

	if !called {
		t.Error("RequireAdmin should allow admin role")
	}
}

// TestRequireAdmin_UserBlocked verifies that RequireAdmin blocks a regular user.
func TestRequireAdmin_UserBlocked(t *testing.T) {
	m := newTestMiddleware(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called for non-admin")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), ContextKeyRole, "user")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	m.RequireAdmin(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("RequireAdmin for user: status %d, want %d", rr.Code, http.StatusForbidden)
	}
}

// TestRequireActiveUser_Active verifies that an active user passes through.
func TestRequireActiveUser_Active(t *testing.T) {
	a := newTestAuth(t)
	m := NewMiddleware(a)
	user := registerTestUser(t, a, "active-user", "active@example.com", "ValidPass1")

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), ContextKeyUserID, user.ID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	m.RequireActiveUser(next).ServeHTTP(rr, req)

	if !called {
		t.Error("RequireActiveUser should allow active user")
	}
}

// TestRequireActiveUser_NoUserID verifies that missing user ID in context gets 401.
func TestRequireActiveUser_NoUserID(t *testing.T) {
	m := newTestMiddleware(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called with no user ID")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	m.RequireActiveUser(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("RequireActiveUser no ID: status %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// TestRequireActiveUser_UnknownUser verifies that an unknown user ID gets 401.
func TestRequireActiveUser_UnknownUser(t *testing.T) {
	m := newTestMiddleware(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called for unknown user")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), ContextKeyUserID, "does-not-exist")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	m.RequireActiveUser(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("RequireActiveUser unknown user: status %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// TestOptionalAuth_NoToken verifies that OptionalAuth passes through without a token.
func TestOptionalAuth_NoToken(t *testing.T) {
	m := newTestMiddleware(t)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	m.OptionalAuth(next).ServeHTTP(rr, req)

	if !called {
		t.Error("OptionalAuth with no token should call next")
	}
}

// TestOptionalAuth_InvalidToken verifies that an invalid token is ignored and next is called.
func TestOptionalAuth_InvalidToken(t *testing.T) {
	m := newTestMiddleware(t)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := makeTokenRequest("invalid.token.string")
	rr := httptest.NewRecorder()
	m.OptionalAuth(next).ServeHTTP(rr, req)

	if !called {
		t.Error("OptionalAuth with invalid token should still call next")
	}
}

// TestOptionalAuth_ValidToken verifies that a valid token enriches context.
func TestOptionalAuth_ValidToken(t *testing.T) {
	a := newTestAuth(t)
	m := NewMiddleware(a)
	user := registerTestUser(t, a, "opt-user", "opt@example.com", "ValidPass1")

	token, err := a.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		uid, _ := r.Context().Value(ContextKeyUserID).(string)
		if uid != user.ID {
			t.Errorf("context UserID = %q, want %q", uid, user.ID)
		}
	})

	req := makeTokenRequest(token)
	rr := httptest.NewRecorder()
	m.OptionalAuth(next).ServeHTTP(rr, req)

	if !called {
		t.Error("OptionalAuth with valid token should call next")
	}
}

// TestRateLimitByUser_Allow verifies that requests within the limit pass through.
func TestRateLimitByUser_Allow(t *testing.T) {
	m := newTestMiddleware(t)
	limiter := NewRateLimiter(10, time.Minute)
	mwFn := m.RateLimitByUser(limiter)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()
	mwFn(next).ServeHTTP(rr, req)

	if !called {
		t.Error("RateLimitByUser should allow request within limit")
	}
}

// TestRateLimitByUser_Exceeded verifies that requests beyond limit get 429.
func TestRateLimitByUser_Exceeded(t *testing.T) {
	m := newTestMiddleware(t)
	limiter := NewRateLimiter(1, time.Minute)
	mwFn := m.RateLimitByUser(limiter)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:9999"

	// First request: should pass.
	rr1 := httptest.NewRecorder()
	mwFn(next).ServeHTTP(rr1, req)

	// Second request: should be rate-limited.
	rr2 := httptest.NewRecorder()
	mwFn(next).ServeHTTP(rr2, req)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("RateLimitByUser exceeded: status %d, want %d", rr2.Code, http.StatusTooManyRequests)
	}
}

// TestCORS_AllowedOrigin verifies that an allowed origin gets CORS headers.
func TestCORS_AllowedOrigin(t *testing.T) {
	m := newTestMiddleware(t)
	mwFn := m.CORS([]string{"https://example.com"})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	mwFn(next).ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Errorf("CORS allowed origin: Access-Control-Allow-Origin = %q, want %q", got, "https://example.com")
	}
}

// TestCORS_NotAllowedOrigin verifies that a disallowed origin does not get CORS headers.
func TestCORS_NotAllowedOrigin(t *testing.T) {
	m := newTestMiddleware(t)
	mwFn := m.CORS([]string{"https://example.com"})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()
	mwFn(next).ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("CORS disallowed origin: Access-Control-Allow-Origin = %q, want empty", got)
	}
}

// TestCORS_WildcardOrigin verifies that a wildcard allows any origin.
func TestCORS_WildcardOrigin(t *testing.T) {
	m := newTestMiddleware(t)
	mwFn := m.CORS([]string{"*"})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://any.com")
	rr := httptest.NewRecorder()
	mwFn(next).ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://any.com" {
		t.Errorf("CORS wildcard: Access-Control-Allow-Origin = %q, want %q", got, "https://any.com")
	}
}

// TestCORS_PreflightRequest verifies that OPTIONS requests get 200 immediately.
func TestCORS_PreflightRequest(t *testing.T) {
	m := newTestMiddleware(t)
	mwFn := m.CORS([]string{"https://example.com"})

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	mwFn(next).ServeHTTP(rr, req)

	if called {
		t.Error("CORS preflight should not call next handler")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("CORS preflight: status %d, want %d", rr.Code, http.StatusOK)
	}
}

// TestSecurityHeaders verifies that required security headers are set.
func TestSecurityHeaders(t *testing.T) {
	m := newTestMiddleware(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	m.SecurityHeaders(next).ServeHTTP(rr, req)

	headers := map[string]string{
		"X-Frame-Options":        "DENY",
		"X-Content-Type-Options": "nosniff",
		"X-XSS-Protection":       "1; mode=block",
	}
	for name, want := range headers {
		if got := rr.Header().Get(name); got != want {
			t.Errorf("SecurityHeaders %s = %q, want %q", name, got, want)
		}
	}
}

// TestExtractToken_BearerHeader verifies that Bearer token is extracted from header.
func TestExtractToken_BearerHeader(t *testing.T) {
	m := newTestMiddleware(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer mytoken123")

	token := m.extractToken(req)
	if token != "mytoken123" {
		t.Errorf("extractToken = %q, want %q", token, "mytoken123")
	}
}

// TestExtractToken_MalformedHeader verifies that a malformed Authorization header returns empty.
func TestExtractToken_MalformedHeader(t *testing.T) {
	m := newTestMiddleware(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "NotBearer mytoken123")

	token := m.extractToken(req)
	if token != "" {
		t.Errorf("extractToken malformed header = %q, want empty", token)
	}
}

// TestExtractToken_NoHeader verifies that no Authorization header returns empty string.
func TestExtractToken_NoHeader(t *testing.T) {
	m := newTestMiddleware(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	token := m.extractToken(req)
	if token != "" {
		t.Errorf("extractToken no header = %q, want empty", token)
	}
}

// TestGetUserIDFromContext verifies context extraction helpers.
func TestGetUserIDFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextKeyUserID, "user-123")
	uid, ok := GetUserIDFromContext(ctx)
	if !ok {
		t.Error("GetUserIDFromContext: ok = false, want true")
	}
	if uid != "user-123" {
		t.Errorf("GetUserIDFromContext = %q, want %q", uid, "user-123")
	}
}

// TestGetUserIDFromContext_Missing verifies that missing key returns empty string and false.
func TestGetUserIDFromContext_Missing(t *testing.T) {
	_, ok := GetUserIDFromContext(context.Background())
	if ok {
		t.Error("GetUserIDFromContext missing: ok = true, want false")
	}
}

// TestGetUsernameFromContext verifies username extraction.
func TestGetUsernameFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextKeyUsername, "alice")
	name, ok := GetUsernameFromContext(ctx)
	if !ok {
		t.Error("GetUsernameFromContext: ok = false, want true")
	}
	if name != "alice" {
		t.Errorf("GetUsernameFromContext = %q, want %q", name, "alice")
	}
}

// TestGetRoleFromContext verifies role extraction.
func TestGetRoleFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextKeyRole, "admin")
	role, ok := GetRoleFromContext(ctx)
	if !ok {
		t.Error("GetRoleFromContext: ok = false, want true")
	}
	if role != "admin" {
		t.Errorf("GetRoleFromContext = %q, want %q", role, "admin")
	}
}

// TestGetClaimsFromContext verifies JWT claims extraction.
func TestGetClaimsFromContext(t *testing.T) {
	claims := &JWTClaims{UserID: "user-abc", Username: "bob", Role: "user"}
	ctx := context.WithValue(context.Background(), ContextKeyClaims, claims)

	got, ok := GetClaimsFromContext(ctx)
	if !ok {
		t.Error("GetClaimsFromContext: ok = false, want true")
	}
	if got.UserID != "user-abc" {
		t.Errorf("GetClaimsFromContext UserID = %q, want %q", got.UserID, "user-abc")
	}
}

// TestIsAdmin_True verifies that IsAdmin returns true for admin role.
func TestIsAdmin_True(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextKeyRole, "admin")
	if !IsAdmin(ctx) {
		t.Error("IsAdmin should return true for admin role")
	}
}

// TestIsAdmin_False verifies that IsAdmin returns false for non-admin role.
func TestIsAdmin_False(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextKeyRole, "user")
	if IsAdmin(ctx) {
		t.Error("IsAdmin should return false for user role")
	}
}

// TestIsAdmin_NoRole verifies that IsAdmin returns false when role is absent.
func TestIsAdmin_NoRole(t *testing.T) {
	if IsAdmin(context.Background()) {
		t.Error("IsAdmin should return false when no role in context")
	}
}

// TestGetIPAddress_ForwardedFor verifies X-Forwarded-For extraction.
func TestGetIPAddress_ForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1")

	ip := getIPAddress(req)
	if ip != "203.0.113.1" {
		t.Errorf("getIPAddress X-Forwarded-For = %q, want %q", ip, "203.0.113.1")
	}
}

// TestGetIPAddress_RealIP verifies X-Real-IP extraction.
func TestGetIPAddress_RealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "198.51.100.5")

	ip := getIPAddress(req)
	if ip != "198.51.100.5" {
		t.Errorf("getIPAddress X-Real-IP = %q, want %q", ip, "198.51.100.5")
	}
}

// TestGetIPAddress_RemoteAddr verifies fallback to RemoteAddr.
func TestGetIPAddress_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:4567"

	ip := getIPAddress(req)
	if ip != "192.0.2.1" {
		t.Errorf("getIPAddress RemoteAddr = %q, want %q", ip, "192.0.2.1")
	}
}

// TestRequireActiveUser_InactiveUser verifies that an inactive (suspended) user gets 403.
func TestRequireActiveUser_InactiveUser(t *testing.T) {
	a := newTestAuth(t)
	m := NewMiddleware(a)
	user := registerTestUser(t, a, "inactive-user", "inactive@example.com", "ValidPass1")

	// Suspend the user directly via DB so IsActive() returns false.
	_, err := a.db.Exec(`UPDATE users SET status = 'suspended' WHERE id = ?`, user.ID)
	if err != nil {
		t.Fatalf("suspending user: %v", err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called for inactive user")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), ContextKeyUserID, user.ID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	m.RequireActiveUser(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("RequireActiveUser inactive: status %d, want %d", rr.Code, http.StatusForbidden)
	}
}
