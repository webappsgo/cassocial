package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// newTestDashboardHandler creates a DashboardHandler with an in-memory DB and default config.
func newTestDashboardHandler(t *testing.T) (*DashboardHandler, *store.DB) {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect returned error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	cfg := &config.Config{}
	cfg.Cassocial.MaxProfilesPerUser = 5
	cfg.Cassocial.MaxLinksPerProfile = 100

	return NewDashboardHandler(cfg, db), db
}

func TestNewDashboardHandler_NotNil(t *testing.T) {
	h, _ := newTestDashboardHandler(t)
	if h == nil {
		t.Fatal("NewDashboardHandler returned nil")
	}
}

// HandleDashboard without auth must return 401.
func TestDashboardHandler_HandleDashboard_NoAuth(t *testing.T) {
	h, _ := newTestDashboardHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.HandleDashboard(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("HandleDashboard (no auth) returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// HandleDashboard with a valid user ID in context must return 200.
func TestDashboardHandler_HandleDashboard_WithAuth(t *testing.T) {
	h, db := newTestDashboardHandler(t)

	// Use store.CreateUser directly so two_factor_secret is "" (not NULL),
	// avoiding a store-level scan error on NULL string columns.
	user := &store.User{
		ID:           "dash-user-id-001",
		Username:     "dashuser",
		Email:        "dash@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req = withUserID(req, user.ID)
	rr := httptest.NewRecorder()
	h.HandleDashboard(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleDashboard (with auth) returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["stats"] == nil {
		t.Error("HandleDashboard response missing 'stats' field")
	}
}

// HandleProfileList without auth must return 401.
func TestDashboardHandler_HandleProfileList_NoAuth(t *testing.T) {
	h, _ := newTestDashboardHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/profiles", nil)
	rr := httptest.NewRecorder()
	h.HandleProfileList(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("HandleProfileList (no auth) returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// HandleProfileList with auth must return 200 with profiles and limit fields.
func TestDashboardHandler_HandleProfileList_WithAuth(t *testing.T) {
	h, _ := newTestDashboardHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/profiles", nil)
	req = withUserID(req, "test-user-id")
	rr := httptest.NewRecorder()
	h.HandleProfileList(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleProfileList (with auth) returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["profiles"] == nil {
		t.Error("HandleProfileList response missing 'profiles' field")
	}
	if body["limit"] == nil {
		t.Error("HandleProfileList response missing 'limit' field")
	}
}

// HandleProfileCreate GET must return 200 with themes and max_links.
func TestDashboardHandler_HandleProfileCreate_GET(t *testing.T) {
	h, _ := newTestDashboardHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/profiles/create", nil)
	rr := httptest.NewRecorder()
	h.HandleProfileCreate(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleProfileCreate GET returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["themes"] == nil {
		t.Error("HandleProfileCreate GET response missing 'themes' field")
	}
	if body["max_links"] == nil {
		t.Error("HandleProfileCreate GET response missing 'max_links' field")
	}
}

// HandleProfileCreate POST must redirect (303) away to the API endpoint.
func TestDashboardHandler_HandleProfileCreate_POST(t *testing.T) {
	h, _ := newTestDashboardHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/dashboard/profiles/create", nil)
	rr := httptest.NewRecorder()
	h.HandleProfileCreate(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("HandleProfileCreate POST returned %d, want %d", rr.Code, http.StatusSeeOther)
	}
}

// HandleProfileEdit without profile ID must return 400.
func TestDashboardHandler_HandleProfileEdit_MissingID(t *testing.T) {
	h, _ := newTestDashboardHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/profiles/edit", nil)
	rr := httptest.NewRecorder()
	h.HandleProfileEdit(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleProfileEdit (no id) returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// HandleProfileEdit with profile ID must return 200.
func TestDashboardHandler_HandleProfileEdit_WithID(t *testing.T) {
	h, _ := newTestDashboardHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/profiles/edit?id=123", nil)
	rr := httptest.NewRecorder()
	h.HandleProfileEdit(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleProfileEdit (with id) returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["id"] != "123" {
		t.Errorf("HandleProfileEdit id = %v, want \"123\"", body["id"])
	}
}

// HandleAnalyticsOverview without auth must return 401.
func TestDashboardHandler_HandleAnalyticsOverview_NoAuth(t *testing.T) {
	h, _ := newTestDashboardHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/analytics", nil)
	rr := httptest.NewRecorder()
	h.HandleAnalyticsOverview(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("HandleAnalyticsOverview (no auth) returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// HandleAnalyticsOverview with auth must return 200 with analytics fields.
func TestDashboardHandler_HandleAnalyticsOverview_WithAuth(t *testing.T) {
	h, _ := newTestDashboardHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/analytics", nil)
	req = withUserID(req, "test-user-id")
	rr := httptest.NewRecorder()
	h.HandleAnalyticsOverview(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleAnalyticsOverview (with auth) returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["total_views"] == nil {
		t.Error("HandleAnalyticsOverview response missing 'total_views' field")
	}
}

// HandleAccountSettings without auth must return 401.
func TestDashboardHandler_HandleAccountSettings_NoAuth(t *testing.T) {
	h, _ := newTestDashboardHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/settings", nil)
	rr := httptest.NewRecorder()
	h.HandleAccountSettings(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("HandleAccountSettings (no auth) returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// HandleAccountSettings with valid auth must return 200 with user fields.
func TestDashboardHandler_HandleAccountSettings_WithAuth(t *testing.T) {
	h, db := newTestDashboardHandler(t)

	user := &store.User{
		ID:           "settings-user-id-001",
		Username:     "settingsuser",
		Email:        "settings@example.com",
		PasswordHash: "x",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/dashboard/settings", nil)
	req = withUserID(req, user.ID)
	rr := httptest.NewRecorder()
	h.HandleAccountSettings(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleAccountSettings (with auth) returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["username"] != "settingsuser" {
		t.Errorf("HandleAccountSettings username = %v, want \"settingsuser\"", body["username"])
	}
}

// HandleAccountSettings returns 500 when the authenticated user ID is not in the DB.
func TestDashboardHandler_HandleAccountSettings_UserNotFound(t *testing.T) {
	h, _ := newTestDashboardHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/settings", nil)
	req = withUserID(req, "nonexistent-user-id")
	rr := httptest.NewRecorder()
	h.HandleAccountSettings(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("HandleAccountSettings (user not found) returned %d, want %d; body: %s",
			rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
}

// HandleNotifications without auth must return 401.
func TestDashboardHandler_HandleNotifications_NoAuth(t *testing.T) {
	h, _ := newTestDashboardHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/notifications", nil)
	rr := httptest.NewRecorder()
	h.HandleNotifications(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("HandleNotifications (no auth) returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// HandleNotifications with auth must return 200.
func TestDashboardHandler_HandleNotifications_WithAuth(t *testing.T) {
	h, _ := newTestDashboardHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/notifications", nil)
	req = withUserID(req, "test-user-id")
	rr := httptest.NewRecorder()
	h.HandleNotifications(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleNotifications (with auth) returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

// HandleRecentActivity without auth must return 401.
func TestDashboardHandler_HandleRecentActivity_NoAuth(t *testing.T) {
	h, _ := newTestDashboardHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/activity", nil)
	rr := httptest.NewRecorder()
	h.HandleRecentActivity(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("HandleRecentActivity (no auth) returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// HandleRecentActivity with auth must return 200.
func TestDashboardHandler_HandleRecentActivity_WithAuth(t *testing.T) {
	h, _ := newTestDashboardHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/activity", nil)
	req = withUserID(req, "test-user-id")
	rr := httptest.NewRecorder()
	h.HandleRecentActivity(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleRecentActivity (with auth) returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["activity"] == nil {
		t.Error("HandleRecentActivity response missing 'activity' field")
	}
}
