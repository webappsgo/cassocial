package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casapps/cassocial/src/server/store"
)

func newTestPublicHandlers(t *testing.T) (*PublicHandlers, *ProfileHandlers, *LinkHandlers, *store.DB) {
	t.Helper()

	ph, db := newTestProfileHandlers(t)
	lh := NewLinkHandlers(db)
	pub := NewPublicHandlers(db)

	return pub, ph, lh, db
}

func TestGetPublicProfile_NotFound(t *testing.T) {
	pub, _, _, _ := newTestPublicHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles/nonexistent", nil)
	req.SetPathValue("username", "nonexistent")

	rr := httptest.NewRecorder()
	pub.GetPublicProfile(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("GetPublicProfile with nonexistent slug returned %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestGetPublicProfile_PublicProfile(t *testing.T) {
	pub, ph, _, db := newTestPublicHandlers(t)
	userID := createTestUser(t, db, "pubprofileuser", "pubprofile@example.com")
	createTestProfile(t, ph, userID, "pubslug")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles/pubslug", nil)
	req.SetPathValue("username", "pubslug")

	rr := httptest.NewRecorder()
	pub.GetPublicProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetPublicProfile with public profile returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if slug, _ := resp["slug"].(string); slug != "pubslug" {
		t.Errorf("expected slug=pubslug, got %q", slug)
	}
}

func TestGetPublicProfile_PrivateProfile(t *testing.T) {
	pub, ph, _, db := newTestPublicHandlers(t)
	userID := createTestUser(t, db, "privprofileuser", "privprofile@example.com")

	// Create a private profile (is_public = false).
	body := map[string]interface{}{
		"slug":         "privateslug",
		"display_name": "Private Profile",
		"is_public":    false,
	}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/profiles", new(stringer))
	req.Body = toReadCloser(data)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	ph.CreateProfile(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("CreateProfile (private) returned %d; body: %s", rr.Code, rr.Body.String())
	}

	// Public access must be denied.
	pubReq := httptest.NewRequest(http.MethodGet, "/api/v1/profiles/privateslug", nil)
	pubReq.SetPathValue("username", "privateslug")

	pubRR := httptest.NewRecorder()
	pub.GetPublicProfile(pubRR, pubReq)

	if pubRR.Code != http.StatusForbidden {
		t.Errorf("GetPublicProfile for private profile returned %d, want %d", pubRR.Code, http.StatusForbidden)
	}
}

func TestGetPublicProfile_EmptyUsername(t *testing.T) {
	pub, _, _, _ := newTestPublicHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles/", nil)
	req.SetPathValue("username", "")

	rr := httptest.NewRecorder()
	pub.GetPublicProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("GetPublicProfile with empty username returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGetPublicProfileLinks_NotFound(t *testing.T) {
	pub, _, _, _ := newTestPublicHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles/noone/links", nil)
	req.SetPathValue("username", "noone")

	rr := httptest.NewRecorder()
	pub.GetPublicProfileLinks(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("GetPublicProfileLinks with nonexistent slug returned %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestGetPublicProfileLinks_PublicProfile(t *testing.T) {
	pub, ph, lh, db := newTestPublicHandlers(t)
	userID := createTestUser(t, db, "publinksuser", "publinks@example.com")
	profileID := createTestProfile(t, ph, userID, "publiclinksslug")
	createTestLink(t, lh, userID, profileID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles/publiclinksslug/links", nil)
	req.SetPathValue("username", "publiclinksslug")

	rr := httptest.NewRecorder()
	pub.GetPublicProfileLinks(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetPublicProfileLinks returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var links []interface{}
	json.NewDecoder(rr.Body).Decode(&links)
	if len(links) != 1 {
		t.Errorf("expected 1 link, got %d", len(links))
	}
}

func TestGetPublicProfileQR_Valid(t *testing.T) {
	pub, ph, _, db := newTestPublicHandlers(t)
	userID := createTestUser(t, db, "qrpublicuser", "qrpublic@example.com")
	createTestProfile(t, ph, userID, "qrpublicslug")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles/qrpublicslug/qr", nil)
	req.SetPathValue("username", "qrpublicslug")

	rr := httptest.NewRecorder()
	pub.GetPublicProfileQR(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetPublicProfileQR returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestGetPublicProfileQR_NotFound(t *testing.T) {
	pub, _, _, _ := newTestPublicHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles/nobody/qr", nil)
	req.SetPathValue("username", "nobody")

	rr := httptest.NewRecorder()
	pub.GetPublicProfileQR(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("GetPublicProfileQR for nonexistent slug returned %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestTrackLinkClick_NotFound(t *testing.T) {
	pub, _, _, _ := newTestPublicHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/link/badid/click", nil)
	req.SetPathValue("id", "badid")

	rr := httptest.NewRecorder()
	pub.TrackLinkClick(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("TrackLinkClick for nonexistent link returned %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestTrackLinkClick_ActiveLink(t *testing.T) {
	pub, ph, lh, db := newTestPublicHandlers(t)
	userID := createTestUser(t, db, "trackclickuser", "trackclick@example.com")
	profileID := createTestProfile(t, ph, userID, "trackclickslug")
	linkID := createTestLink(t, lh, userID, profileID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/link/"+linkID+"/click", nil)
	req.SetPathValue("id", linkID)

	rr := httptest.NewRecorder()
	pub.TrackLinkClick(rr, req)

	// Should redirect (307).
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("TrackLinkClick returned %d, want %d; body: %s",
			rr.Code, http.StatusTemporaryRedirect, rr.Body.String())
	}
	location := rr.Header().Get("Location")
	if location != "https://example.com" {
		t.Errorf("TrackLinkClick redirect location = %q, want https://example.com", location)
	}
}

// stringer is a dummy type to allow constructing requests with a replaceable body.
type stringer struct{}

func (s *stringer) Read(p []byte) (int, error) { return 0, nil }

// toReadCloser wraps JSON bytes in an io.ReadCloser.
func toReadCloser(data []byte) interface{ Read([]byte) (int, error); Close() error } {
	return noopCloser{data: data, pos: 0}
}

type noopCloser struct {
	data []byte
	pos  int
}

func (n noopCloser) Read(p []byte) (int, error) {
	if n.pos >= len(n.data) {
		return 0, nil
	}
	copied := copy(p, n.data[n.pos:])
	n.pos += copied
	return copied, nil
}

func (n noopCloser) Close() error { return nil }
