package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestLinkHandlers creates a LinkHandlers backed by an in-memory SQLite database.
func newTestLinkHandlers(t *testing.T) (*LinkHandlers, *ProfileHandlers, string) {
	t.Helper()

	ph, db := newTestProfileHandlers(t)
	lh := NewLinkHandlers(db)

	userID := createTestUser(t, db, "linkuser", "linkuser@example.com")

	return lh, ph, userID
}

// createTestLink creates a link in the given profile and returns its ID.
func createTestLink(t *testing.T, h *LinkHandlers, userID, profileID string) string {
	t.Helper()

	body := map[string]interface{}{
		"title": "Test Link",
		"url":   "https://example.com",
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/links", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	h.CreateLink(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("createTestLink: CreateLink returned %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	id, _ := resp["id"].(string)
	if id == "" {
		t.Fatalf("createTestLink: response missing id; got %v", resp)
	}
	return id
}

func TestListLinks_NoAuth(t *testing.T) {
	lh, _, _ := newTestLinkHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/someid/links", nil)
	req.SetPathValue("id", "someid")
	rr := httptest.NewRecorder()
	lh.ListLinks(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("ListLinks without auth returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestListLinks_Empty(t *testing.T) {
	lh, ph, userID := newTestLinkHandlers(t)
	profileID := createTestProfile(t, ph, userID, "linklistslug")

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/"+profileID+"/links", nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	lh.ListLinks(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ListLinks returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var links []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&links); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("expected 0 links, got %d", len(links))
	}
}

func TestCreateLink_Valid(t *testing.T) {
	lh, ph, userID := newTestLinkHandlers(t)
	profileID := createTestProfile(t, ph, userID, "createlink")

	body := map[string]interface{}{
		"title": "GitHub",
		"url":   "https://github.com",
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/links", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	lh.CreateLink(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("CreateLink returned %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
}

func TestCreateLink_NoAuth(t *testing.T) {
	lh, ph, userID := newTestLinkHandlers(t)
	profileID := createTestProfile(t, ph, userID, "createlinknoa")

	data, _ := json.Marshal(map[string]interface{}{"title": "x", "url": "https://x.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/links", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", profileID)

	rr := httptest.NewRecorder()
	lh.CreateLink(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("CreateLink without auth returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestCreateLink_InvalidURL(t *testing.T) {
	lh, ph, userID := newTestLinkHandlers(t)
	profileID := createTestProfile(t, ph, userID, "invalidurlslug")

	body := map[string]interface{}{
		"title": "Bad Link",
		"url":   "not-a-url",
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/links", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	lh.CreateLink(rr, req)

	// Invalid URL must not be accepted.
	if rr.Code == http.StatusCreated {
		t.Errorf("CreateLink with invalid URL returned %d (created), want non-2xx", rr.Code)
	}
}

func TestUpdateLink_Valid(t *testing.T) {
	lh, ph, userID := newTestLinkHandlers(t)
	profileID := createTestProfile(t, ph, userID, "updatelink")
	linkID := createTestLink(t, lh, userID, profileID)

	update := map[string]interface{}{
		"title": "Updated Title",
		"url":   "https://updated.example.com",
	}
	data, _ := json.Marshal(update)

	req := httptest.NewRequest(http.MethodPut, "/api/links/"+linkID, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", linkID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	lh.UpdateLink(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("UpdateLink returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestUpdateLink_NotFound(t *testing.T) {
	lh, _, userID := newTestLinkHandlers(t)

	data, _ := json.Marshal(map[string]interface{}{"title": "x"})
	req := httptest.NewRequest(http.MethodPut, "/api/links/nonexistent", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "nonexistent-id")
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	lh.UpdateLink(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("UpdateLink with nonexistent ID returned %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestDeleteLink_Valid(t *testing.T) {
	lh, ph, userID := newTestLinkHandlers(t)
	profileID := createTestProfile(t, ph, userID, "deletelink")
	linkID := createTestLink(t, lh, userID, profileID)

	req := httptest.NewRequest(http.MethodDelete, "/api/links/"+linkID, nil)
	req.SetPathValue("id", linkID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	lh.DeleteLink(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("DeleteLink returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestDeleteLink_NoAuth(t *testing.T) {
	lh, _, _ := newTestLinkHandlers(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/links/someid", nil)
	req.SetPathValue("id", "someid")

	rr := httptest.NewRecorder()
	lh.DeleteLink(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("DeleteLink without auth returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestToggleLink_Valid(t *testing.T) {
	lh, ph, userID := newTestLinkHandlers(t)
	profileID := createTestProfile(t, ph, userID, "togglelink")
	linkID := createTestLink(t, lh, userID, profileID)

	req := httptest.NewRequest(http.MethodPost, "/api/links/"+linkID+"/toggle", nil)
	req.SetPathValue("id", linkID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	lh.ToggleLink(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ToggleLink returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	// The link should now be inactive (was active by default).
	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	isActive, _ := resp["is_active"].(bool)
	if isActive {
		t.Errorf("ToggleLink: expected is_active=false after toggle, got true")
	}
}

func TestReorderLinks_Valid(t *testing.T) {
	lh, ph, userID := newTestLinkHandlers(t)
	profileID := createTestProfile(t, ph, userID, "reorderlink")

	id1 := createTestLink(t, lh, userID, profileID)
	id2 := createTestLink(t, lh, userID, profileID)

	body := map[string]interface{}{
		"link_ids": []string{id2, id1},
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/links/reorder", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	lh.ReorderLinks(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ReorderLinks returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestReorderLinks_Empty(t *testing.T) {
	lh, _, userID := newTestLinkHandlers(t)

	body := map[string]interface{}{"link_ids": []string{}}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/links/reorder", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	lh.ReorderLinks(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("ReorderLinks with empty IDs returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestListLinks_ShowsCreatedLink(t *testing.T) {
	lh, ph, userID := newTestLinkHandlers(t)
	profileID := createTestProfile(t, ph, userID, "listlinksshowslug")
	createTestLink(t, lh, userID, profileID)

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/"+profileID+"/links", nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	lh.ListLinks(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ListLinks returned %d, want %d", rr.Code, http.StatusOK)
	}

	var links []interface{}
	json.NewDecoder(rr.Body).Decode(&links)
	if len(links) != 1 {
		t.Errorf("expected 1 link, got %d", len(links))
	}
}
