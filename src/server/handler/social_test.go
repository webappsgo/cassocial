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

func newTestSocialHandler(t *testing.T) (*SocialHandler, *store.DB) {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	return NewSocialHandler(&config.Config{}, db), db
}

// TestNewSocialHandler verifies the constructor returns a non-nil handler.
func TestNewSocialHandler(t *testing.T) {
	h, _ := newTestSocialHandler(t)
	if h == nil {
		t.Fatal("NewSocialHandler returned nil")
	}
}

// TestHandleProfileDirectory covers default pagination and optional filters.
func TestHandleProfileDirectory(t *testing.T) {
	h, _ := newTestSocialHandler(t)

	tests := []struct {
		name  string
		query string
	}{
		{"no params", ""},
		{"with page and per_page", "?page=2&per_page=10"},
		{"with tag filter", "?tag=developer"},
		{"with verified filter", "?verified=true"},
		{"invalid page falls back", "?page=notanumber"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/directory"+tt.query, nil)
			rr := httptest.NewRecorder()
			h.HandleProfileDirectory(rr, req)
			if rr.Code != http.StatusOK {
				t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
			}
			var resp map[string]interface{}
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("response not valid JSON: %v", err)
			}
			if _, ok := resp["profiles"]; !ok {
				t.Error("response missing 'profiles' field")
			}
		})
	}
}

// TestHandleSearchProfiles covers: missing query, valid query.
func TestHandleSearchProfiles(t *testing.T) {
	h, _ := newTestSocialHandler(t)

	t.Run("missing q returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/search", nil)
		rr := httptest.NewRecorder()
		h.HandleSearchProfiles(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("valid query returns 200 with results", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/search?q=alice", nil)
		rr := httptest.NewRecorder()
		h.HandleSearchProfiles(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		var resp map[string]interface{}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("response not valid JSON: %v", err)
		}
		if resp["query"] != "alice" {
			t.Errorf("query = %q, want alice", resp["query"])
		}
	})
}

// TestHandleFeaturedProfiles verifies the endpoint returns a featured list.
func TestHandleFeaturedProfiles(t *testing.T) {
	h, _ := newTestSocialHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/featured", nil)
	rr := httptest.NewRecorder()
	h.HandleFeaturedProfiles(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("response not valid JSON: %v", err)
	}
	if _, ok := resp["featured"]; !ok {
		t.Error("response missing 'featured' field")
	}
}

// TestHandleVerifyProfile covers: wrong method, missing auth, bad JSON, valid request.
func TestHandleVerifyProfile(t *testing.T) {
	h, db := newTestSocialHandler(t)
	userID := createTestUser(t, db, "verifyuser", "verifyuser@example.com")

	t.Run("wrong method returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/verify", nil)
		rr := httptest.NewRecorder()
		h.HandleVerifyProfile(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("no auth returns 401", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"profile_id": "p1"})
		req := httptest.NewRequest(http.MethodPost, "/api/verify", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		h.HandleVerifyProfile(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})

	t.Run("bad JSON returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/verify", bytes.NewBufferString("{bad"))
		req = withUserID(req, userID)
		rr := httptest.NewRecorder()
		h.HandleVerifyProfile(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("valid request returns 200", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"profile_id": "p1",
			"proof":      "https://example.com/proof",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = withUserID(req, userID)
		rr := httptest.NewRecorder()
		h.HandleVerifyProfile(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		var resp map[string]string
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("response not valid JSON: %v", err)
		}
		if resp["status"] != "success" {
			t.Errorf("status = %q, want success", resp["status"])
		}
	})
}

// TestHandleGetTags verifies the endpoint returns static tag list.
func TestHandleGetTags(t *testing.T) {
	h, _ := newTestSocialHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
	rr := httptest.NewRecorder()
	h.HandleGetTags(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("response not valid JSON: %v", err)
	}
	if _, ok := resp["tags"]; !ok {
		t.Error("response missing 'tags' field")
	}
}

// TestHandleAddTag covers: wrong method, bad JSON, valid request.
func TestHandleAddTag(t *testing.T) {
	h, _ := newTestSocialHandler(t)

	t.Run("wrong method returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
		rr := httptest.NewRecorder()
		h.HandleAddTag(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("bad JSON returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/tags", bytes.NewBufferString("{bad"))
		rr := httptest.NewRecorder()
		h.HandleAddTag(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("valid tag returns 200", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"profile_id": "p1",
			"tag":        "developer",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/tags", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		h.HandleAddTag(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		var resp map[string]string
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("response not valid JSON: %v", err)
		}
		if resp["status"] != "success" {
			t.Errorf("status = %q, want success", resp["status"])
		}
	})
}

// TestGetIntParam verifies boundary conditions: missing, valid, invalid, zero, negative.
func TestGetIntParam(t *testing.T) {
	tests := []struct {
		query    string
		param    string
		defVal   int
		expected int
	}{
		{"", "page", 1, 1},          // missing → default
		{"page=5", "page", 1, 5},    // valid integer
		{"page=abc", "page", 1, 1},  // non-integer → default
		{"page=0", "page", 1, 0},    // zero is a valid int
		{"page=-1", "page", 1, -1},  // negative is a valid int
		{"page=", "page", 3, 3},     // empty string → default
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.query, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test?"+tt.query, nil)
			got := getIntParam(req, tt.param, tt.defVal)
			if got != tt.expected {
				t.Errorf("getIntParam(%q, %q, %d) = %d, want %d",
					tt.query, tt.param, tt.defVal, got, tt.expected)
			}
		})
	}
}
