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

// ---------------------------------------------------------------------------
// ListProfiles – additional coverage
// ---------------------------------------------------------------------------

func TestListProfiles_IsolatedByUser(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	ownerID := createTestUser(t, db, "listowner", "listowner@example.com")
	otherID := createTestUser(t, db, "listother", "listother@example.com")
	createTestProfile(t, h, ownerID, "owner-slug")

	// otherID should see zero profiles even though ownerID has one.
	req := httptest.NewRequest(http.MethodGet, "/api/profiles", nil)
	req = withUserID(req, otherID)
	rr := httptest.NewRecorder()
	h.ListProfiles(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ListProfiles returned %d, want %d", rr.Code, http.StatusOK)
	}
	var profiles []interface{}
	json.NewDecoder(rr.Body).Decode(&profiles)
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles for other user, got %d", len(profiles))
	}
}

// ---------------------------------------------------------------------------
// CreateProfile – additional coverage
// ---------------------------------------------------------------------------

func TestCreateProfile_MaxProfilesReached(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "maxprofuser", "maxprof@example.com")

	// Set max_profiles_per_user to 1.
	db.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES ('max_profiles_per_user', '1')")

	createTestProfile(t, h, userID, "firstslug")

	body := map[string]interface{}{"slug": "secondslug", "display_name": "Second"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/profiles", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.CreateProfile(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("CreateProfile over max returned %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestCreateProfile_ValidationFailure(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "valuser", "val@example.com")

	// Slug with invalid characters triggers Validate().
	body := map[string]interface{}{"slug": "bad slug!", "display_name": "Bad"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/profiles", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.CreateProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("CreateProfile with invalid slug returned %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// UpdateProfile – additional coverage
// ---------------------------------------------------------------------------

func TestUpdateProfile_NoProfileID(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "updnoid", "updnoid@example.com")

	data, _ := json.Marshal(map[string]interface{}{"display_name": "x"})
	req := httptest.NewRequest(http.MethodPut, "/api/profiles/", bytes.NewReader(data))
	// No SetPathValue → path value is empty string.
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("UpdateProfile with empty profile ID returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpdateProfile_ProfileNotFound(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "updnotfound", "updnotfound@example.com")

	data, _ := json.Marshal(map[string]interface{}{"display_name": "x"})
	req := httptest.NewRequest(http.MethodPut, "/api/profiles/ghost", bytes.NewReader(data))
	req.SetPathValue("id", "ghost-id")
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("UpdateProfile with nonexistent ID returned %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestUpdateProfile_WrongUser(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	ownerID := createTestUser(t, db, "updowner", "updowner@example.com")
	otherID := createTestUser(t, db, "updother", "updother@example.com")
	profileID := createTestProfile(t, h, ownerID, "updowned")

	data, _ := json.Marshal(map[string]interface{}{"display_name": "Hacked"})
	req := httptest.NewRequest(http.MethodPut, "/api/profiles/"+profileID, bytes.NewReader(data))
	req.SetPathValue("id", profileID)
	req = withUserID(req, otherID)
	rr := httptest.NewRecorder()
	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("UpdateProfile by wrong user returned %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestUpdateProfile_InvalidBody(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "updbadjson", "updbadjson@example.com")
	profileID := createTestProfile(t, h, userID, "updbadslug")

	req := httptest.NewRequest(http.MethodPut, "/api/profiles/"+profileID, bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("UpdateProfile with bad JSON returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpdateProfile_ValidationFailure(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "updvaluser", "updval@example.com")
	profileID := createTestProfile(t, h, userID, "updvalslug")

	// bio over 500 chars triggers Validate().
	longBio := make([]byte, 501)
	for i := range longBio {
		longBio[i] = 'x'
	}
	body := map[string]interface{}{"bio": string(longBio)}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/api/profiles/"+profileID, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("UpdateProfile with invalid bio returned %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestUpdateProfile_AllOptionalFields(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "updallfields", "updallfields@example.com")
	profileID := createTestProfile(t, h, userID, "updallfslug")

	trueVal := true
	falseVal := false
	body := map[string]interface{}{
		"avatar_url":          "https://example.com/avatar.png",
		"header_image_url":    "https://example.com/header.png",
		"show_usernames":      &trueVal,
		"is_public":           &falseVal,
		"password_protected":  &trueVal,
		"protection_password": "secret",
		"analytics_enabled":   &falseVal,
		"qr_code_enabled":     &falseVal,
		"meta_title":          "Meta",
		"meta_description":    "Desc",
		"og_image_url":        "https://example.com/og.png",
		"custom_css":          "body { color: red; }",
	}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/api/profiles/"+profileID, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("UpdateProfile all fields returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// DeleteProfile – additional coverage
// ---------------------------------------------------------------------------

func TestDeleteProfile_NoProfileID(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "delnoid", "delnoid@example.com")

	req := httptest.NewRequest(http.MethodDelete, "/api/profiles/", nil)
	// No SetPathValue → empty path value.
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.DeleteProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("DeleteProfile with empty profile ID returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestDeleteProfile_NotFound(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "delnotfound", "delnotfound@example.com")

	req := httptest.NewRequest(http.MethodDelete, "/api/profiles/ghost", nil)
	req.SetPathValue("id", "ghost-id")
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.DeleteProfile(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("DeleteProfile with nonexistent ID returned %d, want %d", rr.Code, http.StatusNotFound)
	}
}

// ---------------------------------------------------------------------------
// DuplicateProfile – additional coverage
// ---------------------------------------------------------------------------

func TestDuplicateProfile_NoAuth(t *testing.T) {
	h, _ := newTestProfileHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/someid/duplicate", nil)
	req.SetPathValue("id", "someid")
	rr := httptest.NewRecorder()
	h.DuplicateProfile(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("DuplicateProfile without auth returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestDuplicateProfile_NoProfileID(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "dupnoid", "dupnoid@example.com")

	req := httptest.NewRequest(http.MethodPost, "/api/profiles//duplicate", nil)
	// No SetPathValue → empty path value.
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.DuplicateProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("DuplicateProfile with empty profile ID returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestDuplicateProfile_NotFound(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "dupnotfound", "dupnf@example.com")

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/ghost/duplicate", nil)
	req.SetPathValue("id", "ghost-id")
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.DuplicateProfile(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("DuplicateProfile with nonexistent ID returned %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestDuplicateProfile_WrongUser(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	ownerID := createTestUser(t, db, "dupowner", "dupowner@example.com")
	otherID := createTestUser(t, db, "dupother", "dupother@example.com")
	profileID := createTestProfile(t, h, ownerID, "dupownedslug")

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/duplicate", nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, otherID)
	rr := httptest.NewRecorder()
	h.DuplicateProfile(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("DuplicateProfile by wrong user returned %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestDuplicateProfile_MaxProfilesReached(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "dupmaxuser", "dupmax@example.com")

	db.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES ('max_profiles_per_user', '1')")

	profileID := createTestProfile(t, h, userID, "dupmaxslug")

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/duplicate", nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.DuplicateProfile(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("DuplicateProfile over max returned %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestDuplicateProfile_SlugCollision(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "dupslugcoll", "dupslugcoll@example.com")

	db.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES ('max_profiles_per_user', '10')")

	profileID := createTestProfile(t, h, userID, "collslug")
	// Pre-create the first -copy slug so the handler must try -copy-1.
	createTestProfile(t, h, userID, "collslug-copy")

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/duplicate", nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.DuplicateProfile(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("DuplicateProfile slug collision returned %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	slug, _ := resp["slug"].(string)
	if slug != "collslug-copy-1" {
		t.Errorf("expected slug 'collslug-copy-1', got %q", slug)
	}
}

// ---------------------------------------------------------------------------
// GenerateQRCode – additional coverage
// ---------------------------------------------------------------------------

func TestGenerateQRCode_NoAuth(t *testing.T) {
	h, _ := newTestProfileHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/someid/qr", nil)
	req.SetPathValue("id", "someid")
	rr := httptest.NewRecorder()
	h.GenerateQRCode(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("GenerateQRCode without auth returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestGenerateQRCode_NoProfileID(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "qrnoid", "qrnoid@example.com")

	req := httptest.NewRequest(http.MethodGet, "/api/profiles//qr", nil)
	// No SetPathValue → empty path value.
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.GenerateQRCode(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("GenerateQRCode with empty profile ID returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGenerateQRCode_NotFound(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "qrnotfound", "qrnf@example.com")

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/ghost/qr", nil)
	req.SetPathValue("id", "ghost-id")
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.GenerateQRCode(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("GenerateQRCode with nonexistent ID returned %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestGenerateQRCode_WrongUser(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	ownerID := createTestUser(t, db, "qrowner", "qrowner@example.com")
	otherID := createTestUser(t, db, "qrother", "qrother@example.com")
	profileID := createTestProfile(t, h, ownerID, "qrownedslug")

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/"+profileID+"/qr", nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, otherID)
	rr := httptest.NewRecorder()
	h.GenerateQRCode(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("GenerateQRCode by wrong user returned %d, want %d", rr.Code, http.StatusForbidden)
	}
}

// ---------------------------------------------------------------------------
// VerifyDomain – full coverage (was 0%)
// ---------------------------------------------------------------------------

func TestVerifyDomain_NoAuth(t *testing.T) {
	h, _ := newTestProfileHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/someid/verify-domain", nil)
	req.SetPathValue("id", "someid")
	rr := httptest.NewRecorder()
	h.VerifyDomain(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("VerifyDomain without auth returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestVerifyDomain_NoProfileID(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "vdnoid", "vdnoid@example.com")

	req := httptest.NewRequest(http.MethodPost, "/api/profiles//verify-domain", nil)
	// No SetPathValue → empty path value.
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.VerifyDomain(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("VerifyDomain with empty profile ID returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestVerifyDomain_NotFound(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "vdnotfound", "vdnf@example.com")

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/ghost/verify-domain", nil)
	req.SetPathValue("id", "ghost-id")
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.VerifyDomain(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("VerifyDomain with nonexistent ID returned %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestVerifyDomain_WrongUser(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	ownerID := createTestUser(t, db, "vdowner", "vdowner@example.com")
	otherID := createTestUser(t, db, "vdother", "vdother@example.com")
	profileID := createTestProfile(t, h, ownerID, "vdownedslug")

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/verify-domain", nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, otherID)
	rr := httptest.NewRecorder()
	h.VerifyDomain(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("VerifyDomain by wrong user returned %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestVerifyDomain_Valid(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "vdvalid", "vdvalid@example.com")
	profileID := createTestProfile(t, h, userID, "vdvalidslug")

	// Set a real public domain so DNS resolution succeeds.
	// TXT record will not be present, so the response is 200 verified=false.
	_, err := db.ExecR("UPDATE profiles SET custom_domain = ? WHERE id = ?", "example.com", profileID)
	if err != nil {
		t.Fatalf("set custom_domain: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/verify-domain", nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.VerifyDomain(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("VerifyDomain valid returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// GetProfile – missing no-auth and no-ID paths
// ---------------------------------------------------------------------------

func TestGetProfile_NoAuth(t *testing.T) {
	h, _ := newTestProfileHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/someid", nil)
	req.SetPathValue("id", "someid")
	rr := httptest.NewRecorder()
	h.GetProfile(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("GetProfile without auth returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestGetProfile_NoProfileID(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "getnoid", "getnoid@example.com")

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/", nil)
	// No SetPathValue → empty path value.
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.GetProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("GetProfile with empty profile ID returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// ---------------------------------------------------------------------------
// CreateProfile – max_profiles default (when setting row missing)
// ---------------------------------------------------------------------------

func TestCreateProfile_MaxProfilesDefault(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "defmaxuser", "defmax@example.com")

	// Remove max_profiles_per_user so the handler falls back to the default of 5.
	db.Exec("DELETE FROM settings WHERE key = 'max_profiles_per_user'")

	// Creating one profile should still succeed with the default limit of 5.
	body := map[string]interface{}{"slug": "defmaxslug", "display_name": "Default"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/profiles", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.CreateProfile(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("CreateProfile with no max setting returned %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// DuplicateProfile – max_profiles default (when setting row missing)
// ---------------------------------------------------------------------------

func TestDuplicateProfile_MaxProfilesDefault(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "dupdefmax", "dupdefmax@example.com")

	// Remove the setting so the handler uses the coded default of 5.
	db.Exec("DELETE FROM settings WHERE key = 'max_profiles_per_user'")

	profileID := createTestProfile(t, h, userID, "dupdefmaxslug")

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/duplicate", nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.DuplicateProfile(rr, req)

	// Should succeed because current count (1) < default max (5).
	if rr.Code != http.StatusCreated {
		t.Errorf("DuplicateProfile with no max setting returned %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// ListProfiles – DB error path
// ---------------------------------------------------------------------------

func TestListProfiles_DBError(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "listerr", "listerr@example.com")

	// Close the DB to force a query error.
	db.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/profiles", nil)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.ListProfiles(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("ListProfiles with closed DB returned %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

// ---------------------------------------------------------------------------
// UpdateProfile – DB exec error path (trigger forces UPDATE to fail)
// ---------------------------------------------------------------------------

func TestUpdateProfile_DBError(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "upddberr", "upddberr@example.com")
	profileID := createTestProfile(t, h, userID, "upddberrslug")

	// Install a BEFORE UPDATE trigger that always raises an error.
	db.Exec(`CREATE TRIGGER block_profile_update BEFORE UPDATE ON profiles
		BEGIN SELECT RAISE(ABORT,'update blocked by test trigger'); END`)

	data, _ := json.Marshal(map[string]interface{}{"display_name": "New Name"})
	req := httptest.NewRequest(http.MethodPut, "/api/profiles/"+profileID, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("UpdateProfile with blocked exec returned %d, want %d; body: %s", rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// DeleteProfile – DB exec error path (trigger forces DELETE to fail)
// ---------------------------------------------------------------------------

func TestDeleteProfile_DBError(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "deldberr", "deldberr@example.com")
	profileID := createTestProfile(t, h, userID, "deldberrslug")

	// Install a BEFORE DELETE trigger that always raises an error.
	db.Exec(`CREATE TRIGGER block_profile_delete BEFORE DELETE ON profiles
		BEGIN SELECT RAISE(ABORT,'delete blocked by test trigger'); END`)

	req := httptest.NewRequest(http.MethodDelete, "/api/profiles/"+profileID, nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.DeleteProfile(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("DeleteProfile with blocked exec returned %d, want %d; body: %s", rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// DuplicateProfile – DB exec error path (trigger forces INSERT to fail)
// ---------------------------------------------------------------------------

func TestDuplicateProfile_DBError(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "dupdberr", "dupdberr@example.com")
	profileID := createTestProfile(t, h, userID, "dupdberrslug")

	// Install a BEFORE INSERT trigger that always raises an error.
	// We first need to remove the trigger after the initial profile is created.
	// The createTestProfile helper already ran its INSERT so we add it now.
	db.Exec(`CREATE TRIGGER block_profile_insert BEFORE INSERT ON profiles
		BEGIN SELECT RAISE(ABORT,'insert blocked by test trigger'); END`)

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+profileID+"/duplicate", nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.DuplicateProfile(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("DuplicateProfile with blocked exec returned %d, want %d; body: %s", rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// CreateProfile – DB exec error path (trigger forces INSERT to fail)
// ---------------------------------------------------------------------------

func TestCreateProfile_DBError(t *testing.T) {
	h, db := newTestProfileHandlers(t)
	userID := createTestUser(t, db, "createdbuser", "createdb@example.com")

	// Install a BEFORE INSERT trigger that always raises an error.
	db.Exec(`CREATE TRIGGER block_new_profile_insert BEFORE INSERT ON profiles
		BEGIN SELECT RAISE(ABORT,'insert blocked by test trigger'); END`)

	body := map[string]interface{}{"slug": "dbfailslug", "display_name": "DB Fail"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/profiles", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)
	rr := httptest.NewRecorder()
	h.CreateProfile(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("CreateProfile with blocked exec returned %d, want %d; body: %s", rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
}
