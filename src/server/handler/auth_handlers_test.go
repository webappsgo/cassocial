package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casapps/cassocial/src/server"
	"github.com/casapps/cassocial/src/server/store"
	"github.com/casapps/cassocial/src/service"
)

// newTestAuthHandlers creates an AuthHandlers backed by an in-memory SQLite database.
func newTestAuthHandlers(t *testing.T) *AuthHandlers {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect returned error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	authService := server.NewAuth(db, "test-jwt-secret-for-tests")
	mailer, err := service.NewMailer(nil, "Test App", "https://test.example")
	if err != nil {
		t.Fatalf("service.NewMailer returned error: %v", err)
	}
	return NewAuthHandlers(authService, db, mailer, "https://test.example")
}

// newTestAuthHandlersEnabled creates an AuthHandlers backed by an in-memory
// SQLite database, with a Mailer that reports IsEnabled() == true — for
// tests exercising the SMTP-configured code path (see newEnabledTestMailer).
func newTestAuthHandlersEnabled(t *testing.T) *AuthHandlers {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect returned error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	authService := server.NewAuth(db, "test-jwt-secret-for-tests")
	return NewAuthHandlers(authService, db, newEnabledTestMailer(t), "https://test.example")
}

// postJSON fires a POST request with a JSON body and returns the recorder.
func postJSON(t *testing.T, h http.HandlerFunc, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h(rr, req)
	return rr
}

func TestRegisterValidUser(t *testing.T) {
	h := newTestAuthHandlers(t)

	rr := postJSON(t, h.Register, "/api/auth/register", map[string]string{
		"username": "newuser",
		"email":    "newuser@example.com",
		"password": "ValidPass1",
	})

	if rr.Code != http.StatusCreated {
		t.Errorf("Register with valid body returned %d, want %d; body: %s",
			rr.Code, http.StatusCreated, rr.Body.String())
	}
}

func TestRegisterMissingFields(t *testing.T) {
	tests := []struct {
		name string
		body map[string]string
	}{
		{
			name: "empty username",
			body: map[string]string{"username": "", "email": "a@b.com", "password": "ValidPass1"},
		},
		{
			name: "empty email",
			body: map[string]string{"username": "validname", "email": "", "password": "ValidPass1"},
		},
		{
			name: "empty password",
			body: map[string]string{"username": "validname", "email": "a@b.com", "password": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestAuthHandlers(t)

			rr := postJSON(t, h.Register, "/api/auth/register", tt.body)

			// Registration with missing/empty required fields must not succeed.
			if rr.Code == http.StatusCreated {
				t.Errorf("Register with %s returned %d (created), want a non-2xx status; body: %s",
					tt.name, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestRegisterInvalidBody(t *testing.T) {
	h := newTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Register with invalid JSON returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	h := newTestAuthHandlers(t)

	body := map[string]string{
		"username": "dupuser",
		"email":    "dup@example.com",
		"password": "ValidPass1",
	}

	// First registration must succeed.
	rr1 := postJSON(t, h.Register, "/api/auth/register", body)
	if rr1.Code != http.StatusCreated {
		t.Fatalf("first Register returned %d, want %d; body: %s",
			rr1.Code, http.StatusCreated, rr1.Body.String())
	}

	// Second registration with the same username must fail.
	body["email"] = "different@example.com"
	rr2 := postJSON(t, h.Register, "/api/auth/register", body)

	if rr2.Code != http.StatusConflict && rr2.Code != http.StatusBadRequest {
		t.Errorf("duplicate username Register returned %d, want 409 or 400; body: %s",
			rr2.Code, rr2.Body.String())
	}
}

func TestLoginCorrectCredentials(t *testing.T) {
	h := newTestAuthHandlers(t)

	// Disable email verification so registered users can log in immediately.
	if _, err := h.db.Exec(
		`UPDATE settings SET value = 'false' WHERE key = 'email_verification_required'`,
	); err != nil {
		t.Fatalf("updating email_verification_required returned error: %v", err)
	}

	const username = "loginuser"
	const password = "ValidPass1"

	// Register the user.
	rr := postJSON(t, h.Register, "/api/auth/register", map[string]string{
		"username": username,
		"email":    "loginuser@example.com",
		"password": password,
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Register returned %d, want %d; body: %s",
			rr.Code, http.StatusCreated, rr.Body.String())
	}

	// Now log in.
	rr = postJSON(t, h.Login, "/api/auth/login", map[string]string{
		"username": username,
		"password": password,
	})

	if rr.Code != http.StatusOK {
		t.Errorf("Login with correct credentials returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
		return
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}

	token, ok := resp["token"].(string)
	if !ok || token == "" {
		t.Errorf("login response missing non-empty token field; got %v", resp)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	h := newTestAuthHandlers(t)

	// Disable email verification so the login attempt reaches the password check.
	if _, err := h.db.Exec(
		`UPDATE settings SET value = 'false' WHERE key = 'email_verification_required'`,
	); err != nil {
		t.Fatalf("updating email_verification_required returned error: %v", err)
	}

	// Register a user.
	rr := postJSON(t, h.Register, "/api/auth/register", map[string]string{
		"username": "wrongpassuser",
		"email":    "wrongpass@example.com",
		"password": "ValidPass1",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Register returned %d, want %d; body: %s",
			rr.Code, http.StatusCreated, rr.Body.String())
	}

	// Attempt login with the wrong password.
	rr = postJSON(t, h.Login, "/api/auth/login", map[string]string{
		"username": "wrongpassuser",
		"password": "WrongPass999",
	})

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Login with wrong password returned %d, want %d; body: %s",
			rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
}
