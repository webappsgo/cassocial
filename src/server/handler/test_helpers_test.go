package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casapps/cassocial/src/server"
	"github.com/casapps/cassocial/src/server/store"
)

// newTestProfileHandlers creates a ProfileHandlers backed by an in-memory SQLite database.
func newTestProfileHandlers(t *testing.T) (*ProfileHandlers, *store.DB) {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect returned error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	return NewProfileHandlers(db), db
}

// withUserID adds a user ID to the request context to simulate an authenticated request.
func withUserID(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), server.ContextKeyUserID, userID)
	return r.WithContext(ctx)
}

// createTestUser inserts a user directly into the DB via SQL and returns its ID.
func createTestUser(t *testing.T, db *store.DB, username, email string) string {
	t.Helper()

	userID := generateUUID()
	_, err := db.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, status, email_verified, two_factor_enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, 'user', 'active', 1, 0, datetime('now'), datetime('now'))`,
		userID, username, email, "argon2id-placeholder",
	)
	if err != nil {
		t.Fatalf("createTestUser: INSERT returned error: %v", err)
	}

	return userID
}

// createTestProfile inserts a profile for the given user and returns the profile ID.
func createTestProfile(t *testing.T, h *ProfileHandlers, userID, slug string) string {
	t.Helper()

	body := map[string]interface{}{
		"slug":         slug,
		"display_name": "Test Profile",
		"is_public":    true,
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/profiles", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	h.CreateProfile(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("createTestProfile: CreateProfile returned %d, want %d; body: %s",
			rr.Code, http.StatusCreated, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("createTestProfile: failed to decode response: %v", err)
	}
	id, _ := resp["id"].(string)
	return id
}
