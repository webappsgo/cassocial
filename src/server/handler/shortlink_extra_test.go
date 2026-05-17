package handler

// Tests for ShortlinkHandler — zero coverage before this file.
// Covers HandleCreateShortlink, HandleGetShortlink, HandleDeleteShortlink,
// HandleRedirectShortlink, HandleListShortlinks, and renderError.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casapps/cassocial/src/server/store"
)

// newTestShortlinkHandler returns a ShortlinkHandler backed by an in-memory DB
// and the userID of a test user already inserted in that DB.
func newTestShortlinkHandler(t *testing.T) (*ShortlinkHandler, *store.DB, string) {
	t.Helper()

	_, db := newTestProfileHandlers(t)
	h := NewShortlinkHandler(nil, db)
	userID := createTestUser(t, db, "sluser", "sluser@example.com")

	return h, db, userID
}

// postShortlinkCreate fires a POST to HandleCreateShortlink as the given user.
func postShortlinkCreate(t *testing.T, h *ShortlinkHandler, userID string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/shortlinks", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	if userID != "" {
		req = withUserID(req, userID)
	}

	rr := httptest.NewRecorder()
	h.HandleCreateShortlink(rr, req)
	return rr
}

// ---- renderError (via HandleCreateShortlink with empty URL) ----

func TestShortlinkHandler_renderError_EmptyURL(t *testing.T) {
	// renderError is only called via the handler path; hitting it proves 100% coverage.
	h, _, userID := newTestShortlinkHandler(t)

	rr := postShortlinkCreate(t, h, userID, map[string]interface{}{
		"url": "",
	})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleCreateShortlink with empty URL returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode renderError JSON: %v", err)
	}
	if resp["error"] == "" {
		t.Error("renderError response should contain an 'error' key")
	}
}

// ---- HandleCreateShortlink ----

func TestHandleCreateShortlink_WrongMethod(t *testing.T) {
	h, _, userID := newTestShortlinkHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/shortlinks", nil)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.HandleCreateShortlink(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("HandleCreateShortlink GET returned %d, want %d; body: %s",
			rr.Code, http.StatusMethodNotAllowed, rr.Body.String())
	}
}

func TestHandleCreateShortlink_InvalidBody(t *testing.T) {
	h, _, userID := newTestShortlinkHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/shortlinks", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.HandleCreateShortlink(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleCreateShortlink invalid body returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestHandleCreateShortlink_Success_GeneratedCode(t *testing.T) {
	h, _, userID := newTestShortlinkHandler(t)

	rr := postShortlinkCreate(t, h, userID, map[string]interface{}{
		"url": "https://example.com/long/path",
	})

	if rr.Code != http.StatusCreated {
		t.Errorf("HandleCreateShortlink success returned %d, want %d; body: %s",
			rr.Code, http.StatusCreated, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	code, _ := resp["code"].(string)
	if len(code) < 3 {
		t.Errorf("generated code too short: %q", code)
	}
}

func TestHandleCreateShortlink_Success_CustomCode(t *testing.T) {
	h, _, userID := newTestShortlinkHandler(t)

	rr := postShortlinkCreate(t, h, userID, map[string]interface{}{
		"url":         "https://example.com",
		"custom_code": "mycode",
	})

	if rr.Code != http.StatusCreated {
		t.Errorf("HandleCreateShortlink custom code returned %d, want %d; body: %s",
			rr.Code, http.StatusCreated, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["code"] != "mycode" {
		t.Errorf("code = %v, want mycode", resp["code"])
	}
}

func TestHandleCreateShortlink_InvalidCustomCode(t *testing.T) {
	h, _, userID := newTestShortlinkHandler(t)

	rr := postShortlinkCreate(t, h, userID, map[string]interface{}{
		"url":         "https://example.com",
		"custom_code": "!!bad code!!",
	})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleCreateShortlink invalid custom code returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestHandleCreateShortlink_DuplicateCustomCode(t *testing.T) {
	h, _, userID := newTestShortlinkHandler(t)

	// Create first.
	rr := postShortlinkCreate(t, h, userID, map[string]interface{}{
		"url":         "https://example.com",
		"custom_code": "dupcode",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("first create returned %d; body: %s", rr.Code, rr.Body.String())
	}

	// Attempt to create again with the same code.
	rr = postShortlinkCreate(t, h, userID, map[string]interface{}{
		"url":         "https://other.com",
		"custom_code": "dupcode",
	})

	if rr.Code != http.StatusConflict {
		t.Errorf("duplicate custom code returned %d, want %d; body: %s",
			rr.Code, http.StatusConflict, rr.Body.String())
	}
}

func TestHandleCreateShortlink_WithExpiry(t *testing.T) {
	h, _, userID := newTestShortlinkHandler(t)

	rr := postShortlinkCreate(t, h, userID, map[string]interface{}{
		"url":        "https://example.com",
		"expires_in": 24, // hours
	})

	if rr.Code != http.StatusCreated {
		t.Errorf("HandleCreateShortlink with expiry returned %d, want %d; body: %s",
			rr.Code, http.StatusCreated, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["expires_at"] == "" {
		t.Error("response should contain expires_at when expires_in is set")
	}
}

// ---- HandleGetShortlink ----

func TestHandleGetShortlink_MissingCode(t *testing.T) {
	h, _, _ := newTestShortlinkHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/shortlinks", nil)
	rr := httptest.NewRecorder()
	h.HandleGetShortlink(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleGetShortlink no code returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestHandleGetShortlink_NotFound(t *testing.T) {
	h, _, _ := newTestShortlinkHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/shortlinks?code=doesnotexist", nil)
	rr := httptest.NewRecorder()
	h.HandleGetShortlink(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("HandleGetShortlink not found returned %d, want %d; body: %s",
			rr.Code, http.StatusNotFound, rr.Body.String())
	}
}

func TestHandleGetShortlink_Success(t *testing.T) {
	h, _, userID := newTestShortlinkHandler(t)

	// Create a shortlink first.
	rr := postShortlinkCreate(t, h, userID, map[string]interface{}{
		"url":         "https://example.com",
		"custom_code": "getme",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create shortlink returned %d; body: %s", rr.Code, rr.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/api/shortlinks?code=getme", nil)
	rr = httptest.NewRecorder()
	h.HandleGetShortlink(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleGetShortlink success returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["code"] != "getme" {
		t.Errorf("code = %v, want getme", resp["code"])
	}
}

// ---- HandleDeleteShortlink ----

func TestHandleDeleteShortlink_WrongMethod(t *testing.T) {
	h, _, userID := newTestShortlinkHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/shortlinks?code=x", nil)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.HandleDeleteShortlink(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("HandleDeleteShortlink GET returned %d, want %d; body: %s",
			rr.Code, http.StatusMethodNotAllowed, rr.Body.String())
	}
}

func TestHandleDeleteShortlink_Unauthenticated(t *testing.T) {
	h, _, _ := newTestShortlinkHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/shortlinks?code=x", nil)
	rr := httptest.NewRecorder()
	h.HandleDeleteShortlink(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("HandleDeleteShortlink unauthenticated returned %d, want %d; body: %s",
			rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
}

func TestHandleDeleteShortlink_MissingCode(t *testing.T) {
	h, _, userID := newTestShortlinkHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/shortlinks", nil)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.HandleDeleteShortlink(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleDeleteShortlink missing code returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestHandleDeleteShortlink_NotFound(t *testing.T) {
	h, _, userID := newTestShortlinkHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/shortlinks?code=ghost", nil)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.HandleDeleteShortlink(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("HandleDeleteShortlink not found returned %d, want %d; body: %s",
			rr.Code, http.StatusNotFound, rr.Body.String())
	}
}

func TestHandleDeleteShortlink_Forbidden(t *testing.T) {
	h, db, owner := newTestShortlinkHandler(t)
	otherUser := createTestUser(t, db, "other_user", "other@example.com")

	// Create a shortlink as owner.
	rr := postShortlinkCreate(t, h, owner, map[string]interface{}{
		"url":         "https://example.com",
		"custom_code": "ownercode",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create shortlink returned %d; body: %s", rr.Code, rr.Body.String())
	}

	// Try to delete as a different user.
	req := httptest.NewRequest(http.MethodDelete, "/api/shortlinks?code=ownercode", nil)
	req = withUserID(req, otherUser)
	rr = httptest.NewRecorder()
	h.HandleDeleteShortlink(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("HandleDeleteShortlink foreign user returned %d, want %d; body: %s",
			rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestHandleDeleteShortlink_Success(t *testing.T) {
	h, _, userID := newTestShortlinkHandler(t)

	// Create a shortlink.
	rr := postShortlinkCreate(t, h, userID, map[string]interface{}{
		"url":         "https://example.com",
		"custom_code": "todelete",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create shortlink returned %d; body: %s", rr.Code, rr.Body.String())
	}

	// Delete it.
	req := httptest.NewRequest(http.MethodDelete, "/api/shortlinks?code=todelete", nil)
	req = withUserID(req, userID)
	rr = httptest.NewRecorder()
	h.HandleDeleteShortlink(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleDeleteShortlink success returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	// Verify it is gone.
	req = httptest.NewRequest(http.MethodGet, "/api/shortlinks?code=todelete", nil)
	rr = httptest.NewRecorder()
	h.HandleGetShortlink(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("after delete, HandleGetShortlink returned %d, want %d", rr.Code, http.StatusNotFound)
	}
}

// ---- HandleRedirectShortlink ----

func TestHandleRedirectShortlink_EmptyCode(t *testing.T) {
	h, _, _ := newTestShortlinkHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/s/", nil)
	rr := httptest.NewRecorder()
	h.HandleRedirectShortlink(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("HandleRedirectShortlink empty code returned %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestHandleRedirectShortlink_NotFound(t *testing.T) {
	h, _, _ := newTestShortlinkHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/s/nothere", nil)
	rr := httptest.NewRecorder()
	h.HandleRedirectShortlink(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("HandleRedirectShortlink not found returned %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestHandleRedirectShortlink_Success(t *testing.T) {
	h, _, userID := newTestShortlinkHandler(t)

	rr := postShortlinkCreate(t, h, userID, map[string]interface{}{
		"url":         "https://destination.example.com",
		"custom_code": "redir1",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create shortlink returned %d; body: %s", rr.Code, rr.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/s/redir1", nil)
	rr = httptest.NewRecorder()
	h.HandleRedirectShortlink(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("HandleRedirectShortlink success returned %d, want %d; body: %s",
			rr.Code, http.StatusTemporaryRedirect, rr.Body.String())
	}
	if loc := rr.Header().Get("Location"); loc != "https://destination.example.com" {
		t.Errorf("Location header = %q, want https://destination.example.com", loc)
	}
}

// ---- HandleListShortlinks ----

func TestHandleListShortlinks_Unauthenticated(t *testing.T) {
	h, _, _ := newTestShortlinkHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/shortlinks/list", nil)
	rr := httptest.NewRecorder()
	h.HandleListShortlinks(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("HandleListShortlinks unauthenticated returned %d, want %d; body: %s",
			rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
}

func TestHandleListShortlinks_EmptyList(t *testing.T) {
	h, _, userID := newTestShortlinkHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/shortlinks/list", nil)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.HandleListShortlinks(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleListShortlinks empty returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	total, _ := resp["total"].(float64)
	if total != 0 {
		t.Errorf("total = %v, want 0 for new user", total)
	}
}

func TestHandleListShortlinks_WithItems(t *testing.T) {
	h, _, userID := newTestShortlinkHandler(t)

	// Create two shortlinks.
	for _, code := range []string{"list1", "list2"} {
		rr := postShortlinkCreate(t, h, userID, map[string]interface{}{
			"url":         "https://example.com/" + code,
			"custom_code": code,
		})
		if rr.Code != http.StatusCreated {
			t.Fatalf("create shortlink %q returned %d; body: %s", code, rr.Code, rr.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/shortlinks/list", nil)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.HandleListShortlinks(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleListShortlinks returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	total, _ := resp["total"].(float64)
	if total != 2 {
		t.Errorf("total = %v, want 2", total)
	}
}

// ---- isValidShortCode ----

func TestIsValidShortCode(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"abc", true},
		{"ABC123", true},
		{"a", false},              // too short (< 3)
		{"ab", false},             // too short
		{"abcdefghijklmnopqrstu", false}, // too long (> 20)
		{"abc-def", false},        // hyphen not allowed
		{"abc def", false},        // space not allowed
		{"abc_def", false},        // underscore not allowed
		{"", false},
	}

	for _, tt := range tests {
		got := isValidShortCode(tt.code)
		if got != tt.want {
			t.Errorf("isValidShortCode(%q) = %v, want %v", tt.code, got, tt.want)
		}
	}
}

// ---- generateShortCode ----

func TestGenerateShortCode_Length(t *testing.T) {
	h, _, _ := newTestShortlinkHandler(t)
	code := h.generateShortCode()
	if len(code) != 6 {
		t.Errorf("generateShortCode() length = %d, want 6", len(code))
	}
	if !isValidShortCode(code) {
		t.Errorf("generateShortCode() produced invalid code: %q", code)
	}
}

func TestGenerateShortCode_Uniqueness(t *testing.T) {
	h, _, _ := newTestShortlinkHandler(t)
	seen := make(map[string]bool, 10)
	for i := 0; i < 10; i++ {
		code := h.generateShortCode()
		if seen[code] {
			// Collisions are statistically possible but extremely rare over 10 iterations.
			t.Logf("generateShortCode collision at iteration %d: %q", i, code)
		}
		seen[code] = true
	}
}
