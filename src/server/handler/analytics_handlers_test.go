package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casapps/cassocial/src/server/store"
)

func newTestAnalyticsHandlers(t *testing.T) (*AnalyticsHandlers, *ProfileHandlers, *store.DB) {
	t.Helper()

	ph, db := newTestProfileHandlers(t)
	ah := NewAnalyticsHandlers(db)

	return ah, ph, db
}

func TestGetProfileAnalytics_NoAuth(t *testing.T) {
	ah, _, _ := newTestAnalyticsHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/profile/someid", nil)
	req.SetPathValue("id", "someid")

	rr := httptest.NewRecorder()
	ah.GetProfileAnalytics(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("GetProfileAnalytics without auth returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestGetProfileAnalytics_NotOwner(t *testing.T) {
	ah, ph, db := newTestAnalyticsHandlers(t)

	ownerID := createTestUser(t, db, "analyticsowner", "analyticsowner@example.com")
	otherID := createTestUser(t, db, "analyticsother", "analyticsother@example.com")
	profileID := createTestProfile(t, ph, ownerID, "analyticsownedslug")

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/profile/"+profileID, nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, otherID)

	rr := httptest.NewRecorder()
	ah.GetProfileAnalytics(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("GetProfileAnalytics by non-owner returned %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestGetProfileAnalytics_Valid(t *testing.T) {
	ah, ph, db := newTestAnalyticsHandlers(t)

	userID := createTestUser(t, db, "analyticsuser", "analyticsuser@example.com")
	profileID := createTestProfile(t, ph, userID, "analyticsslug")

	tests := []struct {
		name   string
		period string
	}{
		{"default period", ""},
		{"day", "day"},
		{"week", "week"},
		{"month", "month"},
		{"year", "year"},
		{"all", "all"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/analytics/profile/" + profileID
			if tt.period != "" {
				url += "?period=" + tt.period
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.SetPathValue("id", profileID)
			req = withUserID(req, userID)

			rr := httptest.NewRecorder()
			ah.GetProfileAnalytics(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("GetProfileAnalytics(period=%q) returned %d, want %d; body: %s",
					tt.period, rr.Code, http.StatusOK, rr.Body.String())
			}

			var resp map[string]interface{}
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			for _, field := range []string{"total_views", "unique_visitors", "total_clicks"} {
				if _, ok := resp[field]; !ok {
					t.Errorf("response missing field %q", field)
				}
			}
		})
	}
}

func TestGetLinkAnalytics_NoAuth(t *testing.T) {
	ah, _, _ := newTestAnalyticsHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/links/someid", nil)
	req.SetPathValue("profile_id", "someid")

	rr := httptest.NewRecorder()
	ah.GetLinkAnalytics(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("GetLinkAnalytics without auth returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestGetLinkAnalytics_Valid(t *testing.T) {
	ah, ph, db := newTestAnalyticsHandlers(t)

	userID := createTestUser(t, db, "linkanalyticsuser", "linkanalytics@example.com")
	profileID := createTestProfile(t, ph, userID, "linkanalyticsslug")

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/links/"+profileID, nil)
	req.SetPathValue("profile_id", profileID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	ah.GetLinkAnalytics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetLinkAnalytics returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := resp["links"]; !ok {
		t.Errorf("response missing 'links' field")
	}
}

// ---- GetLinkAnalytics missing branches ----

func TestGetLinkAnalytics_MissingID(t *testing.T) {
	ah, _, _ := newTestAnalyticsHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/links/", nil)
	// PathValue("profile_id") returns "" when not set.
	req = withUserID(req, "any-user-id")

	rr := httptest.NewRecorder()
	ah.GetLinkAnalytics(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("GetLinkAnalytics with missing profile_id returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestGetLinkAnalytics_NotOwner(t *testing.T) {
	ah, ph, db := newTestAnalyticsHandlers(t)

	ownerID := createTestUser(t, db, "linkanalyticsowner", "linkanalyticsowner@example.com")
	otherID := createTestUser(t, db, "linkanalyticsother", "linkanalyticsother@example.com")
	profileID := createTestProfile(t, ph, ownerID, "linkanalyticsownedslug")

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/links/"+profileID, nil)
	req.SetPathValue("profile_id", profileID)
	req = withUserID(req, otherID)

	rr := httptest.NewRecorder()
	ah.GetLinkAnalytics(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("GetLinkAnalytics by non-owner returned %d, want %d; body: %s",
			rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

// ---- GetProfileAnalytics with actual analytics rows exercises getViewsByDay/etc inner loops ----

func insertAnalyticsRow(t *testing.T, db interface {
	Exec(query string, args ...interface{}) (interface{ LastInsertId() (int64, error) }, error)
}, profileID, eventType, referrer, country, device string) {
	t.Helper()
}

// TestGetProfileAnalytics_WithData inserts analytics rows so the helper
// query functions (getViewsByDay, getTopReferrers, getDeviceBreakdown,
// getCountryBreakdown) scan actual rows and exercise their inner loop bodies.
func TestGetProfileAnalytics_WithData(t *testing.T) {
	ah, ph, db := newTestAnalyticsHandlers(t)

	userID := createTestUser(t, db, "analyticsdatauser", "analyticsdata@example.com")
	profileID := createTestProfile(t, ph, userID, "analyticsdataslug")

	// Insert analytics events so inner loops execute.
	for _, row := range []struct {
		eventType string
		referrer  string
		country   string
		device    string
		ipHash    string
	}{
		{"view", "https://example.com", "US", "mobile", "hash1"},
		{"view", "https://example.com", "US", "desktop", "hash2"},
		{"view", "https://other.com", "CA", "tablet", "hash3"},
		{"click", "", "US", "mobile", "hash4"},
	} {
		_, err := db.Exec(
			`INSERT INTO analytics (id, profile_id, event_type, referrer, country, device_type, ip_hash, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
			generateUUID(), profileID, row.eventType, row.referrer, row.country, row.device, row.ipHash,
		)
		if err != nil {
			t.Fatalf("insert analytics row: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/profile/"+profileID+"?period=all", nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	ah.GetProfileAnalytics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetProfileAnalytics with data returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// With actual view rows, total_views must be > 0.
	if views, _ := resp["total_views"].(float64); views == 0 {
		t.Errorf("expected total_views > 0, got %v", resp["total_views"])
	}
}

// TestGetProfileAnalytics_WeekData exercises the "week" period path (most common)
// with actual data so the sub-queries return rows.
func TestGetProfileAnalytics_WeekData(t *testing.T) {
	ah, ph, db := newTestAnalyticsHandlers(t)

	userID := createTestUser(t, db, "analyticsweekuser", "analyticsweek@example.com")
	profileID := createTestProfile(t, ph, userID, "analyticsweekslug")

	// Insert a recent view event.
	_, err := db.Exec(
		`INSERT INTO analytics (id, profile_id, event_type, referrer, country, device_type, ip_hash, created_at)
		 VALUES (?, ?, 'view', 'https://ref.example.com', 'DE', 'mobile', 'hashweek', datetime('now'))`,
		generateUUID(), profileID,
	)
	if err != nil {
		t.Fatalf("insert analytics row: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/profile/"+profileID+"?period=week", nil)
	req.SetPathValue("id", profileID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	ah.GetProfileAnalytics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetProfileAnalytics week with data returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

// TestGetLinkAnalytics_WithData inserts links so the rows.Next() body is exercised.
func TestGetLinkAnalytics_WithData(t *testing.T) {
	ah, ph, db := newTestAnalyticsHandlers(t)

	lh := NewLinkHandlers(db)
	userID := createTestUser(t, db, "linkanalyticsdatauser", "linkanalyticsdata@example.com")
	profileID := createTestProfile(t, ph, userID, "linkanalyticsdataslug")

	// Create a link so the query returns rows.
	createTestLink(t, lh, userID, profileID)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/links/"+profileID, nil)
	req.SetPathValue("profile_id", profileID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	ah.GetLinkAnalytics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetLinkAnalytics with data returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	links, _ := resp["links"].([]interface{})
	if len(links) == 0 {
		t.Errorf("expected at least one link in analytics response, got none")
	}
}

// ---- ExportAnalytics missing branches ----

func TestExportAnalytics_MissingID(t *testing.T) {
	ah, _, _ := newTestAnalyticsHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/export/", nil)
	req = withUserID(req, "any-user-id")

	rr := httptest.NewRecorder()
	ah.ExportAnalytics(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("ExportAnalytics with missing profile_id returned %d, want %d; body: %s",
			rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestExportAnalytics_NotOwner(t *testing.T) {
	ah, ph, db := newTestAnalyticsHandlers(t)

	ownerID := createTestUser(t, db, "exportowner", "exportowner@example.com")
	otherID := createTestUser(t, db, "exportother", "exportother@example.com")
	profileID := createTestProfile(t, ph, ownerID, "exportownedslug")

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/export/"+profileID+"?format=csv", nil)
	req.SetPathValue("profile_id", profileID)
	req = withUserID(req, otherID)

	rr := httptest.NewRecorder()
	ah.ExportAnalytics(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("ExportAnalytics by non-owner returned %d, want %d; body: %s",
			rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestExportAnalytics_CSV(t *testing.T) {
	ah, ph, db := newTestAnalyticsHandlers(t)

	userID := createTestUser(t, db, "exportcsvuser", "exportcsv@example.com")
	profileID := createTestProfile(t, ph, userID, "exportcsvslug")

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/export/"+profileID+"?format=csv", nil)
	req.SetPathValue("profile_id", profileID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	ah.ExportAnalytics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ExportAnalytics csv returned %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestExportAnalytics_NoAuth(t *testing.T) {
	ah, _, _ := newTestAnalyticsHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/export/someid", nil)
	req.SetPathValue("profile_id", "someid")

	rr := httptest.NewRecorder()
	ah.ExportAnalytics(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("ExportAnalytics without auth returned %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestExportAnalytics_Valid(t *testing.T) {
	ah, ph, db := newTestAnalyticsHandlers(t)

	userID := createTestUser(t, db, "exportanalyticsuser", "exportanalytics@example.com")
	profileID := createTestProfile(t, ph, userID, "exportanalyticsslug")

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/export/"+profileID+"?format=json", nil)
	req.SetPathValue("profile_id", profileID)
	req = withUserID(req, userID)

	rr := httptest.NewRecorder()
	ah.ExportAnalytics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ExportAnalytics returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}
