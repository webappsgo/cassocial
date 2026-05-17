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
