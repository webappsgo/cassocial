package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListProfiles_NoAuth(t *testing.T) {
	h, _ := newTestProfileHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/profiles", nil)
	rr := httptest.NewRecorder()
	h.ListProfiles(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("ListProfiles without auth returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestListProfiles_EmptyForNewUser(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "listprofileuser", "listprofile@example.com")

	req := httptest.NewRequest(http.MethodGet, "/api/profiles", nil)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.ListProfiles(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ListProfiles returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var profiles []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&profiles); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}
}

func TestCreateProfile_Valid(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "createprofuser", "createprof@example.com")

	body := map[string]interface{}{
		"slug":         "myprofile",
		"display_name": "My Profile",
		"is_public":    true,
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/profiles", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	h.CreateProfile(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("CreateProfile returned %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
}

func TestCreateProfile_NoAuth(t *testing.T) {
	h, _ := newTestProfileHandlers(t)

	data, _ := json.Marshal(map[string]interface{}{"slug": "test", "display_name": "Test"})
	req := httptest.NewRequest(http.MethodPost, "/api/profiles", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateProfile(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("CreateProfile without auth returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestCreateProfile_InvalidBody(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "invbodyuser", "invbody@example.com")

	req := httptest.NewRequest(http.MethodPost, "/api/profiles", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	h.CreateProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("CreateProfile with bad JSON returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestCreateProfile_DuplicateSlug(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "dupsluguser", "dupslug@example.com")

	createTestProfile(t, h, userID, "dupslug")

	// Try creating a second profile with the same slug.
	body := map[string]interface{}{"slug": "dupslug", "display_name": "Dup"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/profiles", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.CreateProfile(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("CreateProfile with duplicate slug returned %d, want %d; body: %s", rr.Code, http.StatusConflict, rr.Body.String())
	}
}

func TestGetProfile_Valid(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "getprofuser", "getprof@example.com")
	profileID := createTestProfile(t, h, userID, "getprofslug")

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/"+profileID, nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	h.GetProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetProfile returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "getprofnotfound", "getprofnf@example.com")

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/nonexistent", nil)
	req.SetPathValue("id", "nonexistent-id")
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	h.GetProfile(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("GetProfile with nonexistent ID returned %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestGetProfile_WrongUser(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	ownerID := createTestUser(t, db, "owner1", "owner1@example.com")
	otherID := createTestUser(t, db, "other1", "other1@example.com")
	profileID := createTestProfile(t, h, ownerID, "ownedprofile")

	// Other user tries to access owner's profile.
	req := httptest.NewRequest(http.MethodGet, "/api/profiles/"+profileID, nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, otherID)

	rr := httptest.NewRecorder()
	h.GetProfile(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("GetProfile by wrong user returned %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestUpdateProfile_Valid(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "updateprofuser", "updateprof@example.com")
	profileID := createTestProfile(t, h, userID, "updateslug")

	update := map[string]interface{}{
		"display_name": "Updated Name",
		"bio":          "Updated bio",
	}
	data, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/api/profiles/"+profileID, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("UpdateProfile returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestUpdateProfile_NoAuth(t *testing.T) {
	h, _ := newTestProfileHandlers(t)

	data, _ := json.Marshal(map[string]interface{}{"display_name": "x"})
	req := httptest.NewRequest(http.MethodPut, "/api/profiles/someid", bytes.NewReader(data))
	req.SetPathValue("id", "someid")

	rr := httptest.NewRecorder()
	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("UpdateProfile without auth returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestDeleteProfile_Valid(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "delprofuser", "delprof@example.com")
	profileID := createTestProfile(t, h, userID, "delslug")

	req := httptest.NewRequest(http.MethodDelete, "/api/profiles/"+profileID, nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	h.DeleteProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("DeleteProfile returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestDeleteProfile_NoAuth(t *testing.T) {
	h, _ := newTestProfileHandlers(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/profiles/someid", nil)
	req.SetPathValue("id", "someid")

	rr := httptest.NewRecorder()
	h.DeleteProfile(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("DeleteProfile without auth returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestDeleteProfile_WrongUser(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	ownerID := createTestUser(t, db, "delowner2", "delowner2@example.com")
	otherID := createTestUser(t, db, "delother2", "delother2@example.com")
	profileID := createTestProfile(t, h, ownerID, "delownedprofile")

	req := httptest.NewRequest(http.MethodDelete, "/api/profiles/"+profileID, nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, otherID)

	rr := httptest.NewRecorder()
	h.DeleteProfile(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("DeleteProfile by wrong user returned %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestListProfiles_ShowsCreatedProfile(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "listshowuser", "listshow@example.com")
	createTestProfile(t, h, userID, "listshowslug")

	req := httptest.NewRequest(http.MethodGet, "/api/profiles", nil)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.ListProfiles(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ListProfiles returned %d, want %d", rr.Code, http.StatusOK)
	}

	var profiles []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&profiles); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(profiles))
	}
}

func TestDuplicateProfile_Valid(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "duprofuser", "duprof@example.com")
	profileID := createTestProfile(t, h, userID, "origslug")

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/duplicate", nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	h.DuplicateProfile(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("DuplicateProfile returned %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	slug, _ := resp["slug"].(string)
	if slug == "" || slug == "origslug" {
		t.Errorf("DuplicateProfile should return a new slug, got %q", slug)
	}
}

func TestGenerateQRCode_Valid(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "qruser", "qr@example.com")
	profileID := createTestProfile(t, h, userID, "qrslug")

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/"+profileID+"/qr", nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	h.GenerateQRCode(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GenerateQRCode returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}
