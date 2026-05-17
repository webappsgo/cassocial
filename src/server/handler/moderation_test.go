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

func newTestModerationHandler(t *testing.T) *ModerationHandler {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	return NewModerationHandler(&config.Config{}, db)
}

// TestNewModerationHandler verifies the constructor returns a non-nil handler.
func TestNewModerationHandler(t *testing.T) {
	h := newTestModerationHandler(t)
	if h == nil {
		t.Fatal("NewModerationHandler returned nil")
	}
}

// TestHandleReportContent covers: wrong method, bad JSON, happy path.
func TestHandleReportContent(t *testing.T) {
	h := newTestModerationHandler(t)

	t.Run("wrong method returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/moderation/report", nil)
		rr := httptest.NewRecorder()
		h.HandleReportContent(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("bad JSON returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/moderation/report", bytes.NewBufferString("{bad"))
		rr := httptest.NewRecorder()
		h.HandleReportContent(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("valid report returns 200", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"content_type": "profile",
			"content_id":   "abc123",
			"reason":       "spam",
			"details":      "lots of spam",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/moderation/report", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		h.HandleReportContent(rr, req)
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

// TestHandleGetModerationQueue verifies GET returns an empty queue.
func TestHandleGetModerationQueue(t *testing.T) {
	h := newTestModerationHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/moderation/queue", nil)
	rr := httptest.NewRecorder()
	h.HandleGetModerationQueue(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("response not valid JSON: %v", err)
	}
	if _, ok := resp["queue"]; !ok {
		t.Error("response missing 'queue' field")
	}
}

// TestHandleModerateContent covers: wrong method, bad JSON, happy path.
func TestHandleModerateContent(t *testing.T) {
	h := newTestModerationHandler(t)

	t.Run("wrong method returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/moderation/action", nil)
		rr := httptest.NewRecorder()
		h.HandleModerateContent(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("bad JSON returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/moderation/action", bytes.NewBufferString("{bad"))
		rr := httptest.NewRecorder()
		h.HandleModerateContent(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("valid action returns 200", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"report_id": "r1",
			"action":    "remove",
			"notes":     "policy violation",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/moderation/action", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		h.HandleModerateContent(rr, req)
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

// TestHandleGetBlockedPatterns verifies GET returns empty patterns list.
func TestHandleGetBlockedPatterns(t *testing.T) {
	h := newTestModerationHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/moderation/patterns", nil)
	rr := httptest.NewRecorder()
	h.HandleGetBlockedPatterns(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("response not valid JSON: %v", err)
	}
	if _, ok := resp["patterns"]; !ok {
		t.Error("response missing 'patterns' field")
	}
}

// TestHandleAddBlockedPattern covers: wrong method, bad JSON, happy path.
func TestHandleAddBlockedPattern(t *testing.T) {
	h := newTestModerationHandler(t)

	t.Run("wrong method returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/moderation/patterns", nil)
		rr := httptest.NewRecorder()
		h.HandleAddBlockedPattern(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("bad JSON returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/moderation/patterns", bytes.NewBufferString("{bad"))
		rr := httptest.NewRecorder()
		h.HandleAddBlockedPattern(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("valid pattern returns 200", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"type":    "domain",
			"pattern": "spam.example.com",
			"reason":  "known spam",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/moderation/patterns", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		h.HandleAddBlockedPattern(rr, req)
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
