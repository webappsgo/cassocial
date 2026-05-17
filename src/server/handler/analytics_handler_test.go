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

// newTestAnalyticsHandler creates an AnalyticsHandler backed by an in-memory SQLite database.
func newTestAnalyticsHandler(t *testing.T) *AnalyticsHandler {
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
	cfg.Server.Mode = "development"
	return NewAnalyticsHandler(cfg, db)
}

func TestNewAnalyticsHandler(t *testing.T) {
	h := newTestAnalyticsHandler(t)
	if h == nil {
		t.Fatal("NewAnalyticsHandler returned nil")
	}
}

func TestAnalyticsHandler_HandleGetProfileAnalytics_MissingID(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/profile", nil)
	rr := httptest.NewRecorder()
	h.HandleGetProfileAnalytics(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleGetProfileAnalytics without profile_id returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAnalyticsHandler_HandleGetProfileAnalytics_WithID(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/profile?profile_id=test-profile-123", nil)
	rr := httptest.NewRecorder()
	h.HandleGetProfileAnalytics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleGetProfileAnalytics returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := resp["profile_id"]; !ok {
		t.Error("response missing 'profile_id' field")
	}
}

func TestAnalyticsHandler_HandleGetProfileAnalytics_WithDays(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/profile?profile_id=pid&days=7", nil)
	rr := httptest.NewRecorder()
	h.HandleGetProfileAnalytics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleGetProfileAnalytics with days param returned status %d, want %d", rr.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if days, _ := resp["period_days"].(float64); days != 7 {
		t.Errorf("expected period_days=7, got %v", resp["period_days"])
	}
}

func TestAnalyticsHandler_HandleGetLinkAnalytics_MissingID(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/link", nil)
	rr := httptest.NewRecorder()
	h.HandleGetLinkAnalytics(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleGetLinkAnalytics without link_id returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAnalyticsHandler_HandleGetLinkAnalytics_WithID(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/link?link_id=link-abc", nil)
	rr := httptest.NewRecorder()
	h.HandleGetLinkAnalytics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleGetLinkAnalytics returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := resp["link_id"]; !ok {
		t.Error("response missing 'link_id' field")
	}
}

func TestAnalyticsHandler_HandleTrackView_WrongMethod(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/view", nil)
	rr := httptest.NewRecorder()
	h.HandleTrackView(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("HandleTrackView with GET returned status %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestAnalyticsHandler_HandleTrackView_InvalidBody(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/analytics/view", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleTrackView(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleTrackView with invalid body returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAnalyticsHandler_HandleTrackView_Valid(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	body, _ := json.Marshal(map[string]string{"profile_id": "test-profile"})
	req := httptest.NewRequest(http.MethodPost, "/api/analytics/view", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleTrackView(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleTrackView returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestAnalyticsHandler_HandleTrackClick_WrongMethod(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/click", nil)
	rr := httptest.NewRecorder()
	h.HandleTrackClick(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("HandleTrackClick with GET returned status %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestAnalyticsHandler_HandleTrackClick_InvalidBody(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/analytics/click", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleTrackClick(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleTrackClick with invalid body returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAnalyticsHandler_HandleTrackClick_Valid(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	body, _ := json.Marshal(map[string]string{"link_id": "link-123"})
	req := httptest.NewRequest(http.MethodPost, "/api/analytics/click", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.HandleTrackClick(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleTrackClick returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

// ---- HandleExportAnalytics (was 0% covered) ----
// Tests cover: missing profile_id (400), csv format, json format (default),
// pdf format (501), unknown format (400), and explicit days param on link analytics.

func TestAnalyticsHandler_HandleExportAnalytics_MissingID(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/export", nil)
	rr := httptest.NewRecorder()
	h.HandleExportAnalytics(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleExportAnalytics without profile_id returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestAnalyticsHandler_HandleExportAnalytics_JSON(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/export?profile_id=pid123&format=json", nil)
	rr := httptest.NewRecorder()
	h.HandleExportAnalytics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleExportAnalytics json returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode json export response: %v", err)
	}
	if _, ok := resp["profile_id"]; !ok {
		t.Error("json export response missing 'profile_id' field")
	}
}

func TestAnalyticsHandler_HandleExportAnalytics_DefaultFormat(t *testing.T) {
	// Omitting format should default to json.
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/export?profile_id=pid456", nil)
	rr := httptest.NewRecorder()
	h.HandleExportAnalytics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleExportAnalytics default format returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestAnalyticsHandler_HandleExportAnalytics_CSV(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/export?profile_id=pid789&format=csv", nil)
	rr := httptest.NewRecorder()
	h.HandleExportAnalytics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleExportAnalytics csv returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	ct := rr.Header().Get("Content-Type")
	if ct != "text/csv" {
		t.Errorf("HandleExportAnalytics csv Content-Type = %q, want %q", ct, "text/csv")
	}
}

func TestAnalyticsHandler_HandleExportAnalytics_PDF(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/export?profile_id=pid000&format=pdf", nil)
	rr := httptest.NewRecorder()
	h.HandleExportAnalytics(rr, req)

	if rr.Code != http.StatusNotImplemented {
		t.Errorf("HandleExportAnalytics pdf returned %d, want %d; body: %s",
			rr.Code, http.StatusNotImplemented, rr.Body.String())
	}
}

func TestAnalyticsHandler_HandleExportAnalytics_UnknownFormat(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/export?profile_id=pid111&format=xml", nil)
	rr := httptest.NewRecorder()
	h.HandleExportAnalytics(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("HandleExportAnalytics unknown format returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

// HandleGetLinkAnalytics: missing link_id with custom days param exercises the days branch.
func TestAnalyticsHandler_HandleGetLinkAnalytics_InvalidDays(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/link?link_id=linkX&days=notanumber", nil)
	rr := httptest.NewRecorder()
	h.HandleGetLinkAnalytics(rr, req)

	// Invalid days value falls back to 30; response still 200.
	if rr.Code != http.StatusOK {
		t.Errorf("HandleGetLinkAnalytics with invalid days returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestAnalyticsHandler_HandleGetLinkAnalytics_WithDays(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/link?link_id=linkY&days=14", nil)
	rr := httptest.NewRecorder()
	h.HandleGetLinkAnalytics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleGetLinkAnalytics with days returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if days, _ := resp["period_days"].(float64); days != 14 {
		t.Errorf("expected period_days=14, got %v", resp["period_days"])
	}
}

func TestAnalyticsHandler_HandleGetDashboard(t *testing.T) {
	h := newTestAnalyticsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/dashboard", nil)
	rr := httptest.NewRecorder()
	h.HandleGetDashboard(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleGetDashboard returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	for _, field := range []string{"total_views", "total_clicks", "total_profiles"} {
		if _, ok := resp[field]; !ok {
			t.Errorf("dashboard response missing field %q", field)
		}
	}
}
