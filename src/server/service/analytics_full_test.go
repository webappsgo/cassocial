package service

import (
	"testing"
	"time"

	"github.com/casapps/cassocial/src/server/model"
	"github.com/casapps/cassocial/src/server/store"
)

// newTestAnalyticsDB creates an in-memory DB ready for analytics testing.
// It also inserts a user and an analytics-enabled profile.
func newTestAnalyticsDB(t *testing.T) (*store.DB, string) {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	// Insert a user
	_, err = db.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, status, email_verified, two_factor_enabled, created_at, updated_at)
		 VALUES ('analytics-user-001', 'analyticsowner', 'analytics@example.com',
		         '$argon2id$v=19$m=65536,t=3,p=4$s$h', 'user', 'active', 1, 0,
		         CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
	)
	if err != nil {
		t.Fatalf("insert analytics user: %v", err)
	}

	// Insert a profile with analytics enabled
	if err := db.CreateProfile(&store.Profile{
		ID:       "analytics-profile-001",
		UserID:   "analytics-user-001",
		Slug:     "analytics-test-profile",
		IsPublic: true,
	}); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	// Enable analytics for the profile
	_, err = db.Exec(
		`UPDATE profiles SET analytics_enabled = 1 WHERE id = 'analytics-profile-001'`,
	)
	if err != nil {
		t.Fatalf("enable analytics: %v", err)
	}

	return db, "analytics-profile-001"
}

// ---------------------------------------------------------------------------
// NewAnalyticsService
// ---------------------------------------------------------------------------

func TestNewAnalyticsService(t *testing.T) {
	db, _ := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)
	if svc == nil {
		t.Fatal("NewAnalyticsService returned nil")
	}
}

// ---------------------------------------------------------------------------
// TrackView
// ---------------------------------------------------------------------------

func TestAnalyticsService_TrackView_Enabled(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	if err := svc.TrackView(profileID, "1.2.3.4", "Mozilla/5.0", "https://referrer.test"); err != nil {
		t.Fatalf("TrackView: %v", err)
	}

	// Verify row in DB
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM analytics WHERE profile_id = ? AND event_type = 'view'`, profileID).Scan(&count); err != nil {
		t.Fatalf("count analytics: %v", err)
	}
	if count != 1 {
		t.Errorf("TrackView: analytics count = %d, want 1", count)
	}
}

func TestAnalyticsService_TrackView_Disabled(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	// Disable analytics
	_, err := db.Exec(`UPDATE profiles SET analytics_enabled = 0 WHERE id = ?`, profileID)
	if err != nil {
		t.Fatalf("disable analytics: %v", err)
	}

	if err := svc.TrackView(profileID, "1.2.3.4", "Mozilla/5.0", ""); err != nil {
		t.Fatalf("TrackView with disabled analytics: %v", err)
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM analytics WHERE profile_id = ?`, profileID).Scan(&count) //nolint:errcheck
	if count != 0 {
		t.Errorf("TrackView with disabled analytics should not record event, got %d rows", count)
	}
}

func TestAnalyticsService_TrackView_LocalIP(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	// Loopback IP should still record (privacy policy doesn't prevent it)
	if err := svc.TrackView(profileID, "127.0.0.1", "Mozilla/5.0", ""); err != nil {
		t.Fatalf("TrackView loopback: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TrackClick
// ---------------------------------------------------------------------------

func TestAnalyticsService_TrackClick(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	// Insert a link for FK satisfaction
	_, err := db.Exec(
		`INSERT INTO links (id, profile_id, title, url, position, is_active, click_count, created_at, updated_at)
		 VALUES ('analytics-link-001', ?, 'Test', 'https://example.com', 1, 1, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		profileID,
	)
	if err != nil {
		t.Fatalf("insert link: %v", err)
	}

	if err := svc.TrackClick(profileID, "analytics-link-001", "1.2.3.4", "Mozilla/5.0", ""); err != nil {
		t.Fatalf("TrackClick: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM analytics WHERE profile_id = ? AND event_type = 'click'`, profileID).Scan(&count); err != nil {
		t.Fatalf("count click analytics: %v", err)
	}
	if count != 1 {
		t.Errorf("TrackClick: analytics count = %d, want 1", count)
	}
}

func TestAnalyticsService_TrackClick_Disabled(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	_, err := db.Exec(`UPDATE profiles SET analytics_enabled = 0 WHERE id = ?`, profileID)
	if err != nil {
		t.Fatalf("disable analytics: %v", err)
	}

	if err := svc.TrackClick(profileID, "some-link-id", "1.2.3.4", "Mozilla/5.0", ""); err != nil {
		t.Fatalf("TrackClick with disabled analytics: %v", err)
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM analytics WHERE profile_id = ?`, profileID).Scan(&count) //nolint:errcheck
	if count != 0 {
		t.Errorf("TrackClick disabled: expected 0 rows, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// TrackSession
// ---------------------------------------------------------------------------

func TestAnalyticsService_TrackSession_New(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	session := &model.AnalyticsSession{
		ProfileID:  profileID,
		SessionID:  "test-session-001",
		IPHash:     "abc123hash",
		DeviceType: model.DeviceTypeDesktop,
		CreatedAt:  time.Now(),
	}

	if err := svc.TrackSession(session); err != nil {
		t.Fatalf("TrackSession (new): %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM analytics_sessions WHERE session_id = ?`, session.SessionID).Scan(&count); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if count != 1 {
		t.Errorf("TrackSession: session count = %d, want 1", count)
	}
}

func TestAnalyticsService_TrackSession_Update(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	session := &model.AnalyticsSession{
		ProfileID:       profileID,
		SessionID:       "update-session-001",
		IPHash:          "abc123hash",
		DeviceType:      model.DeviceTypeDesktop,
		DurationSeconds: 10,
		CreatedAt:       time.Now(),
	}

	if err := svc.TrackSession(session); err != nil {
		t.Fatalf("TrackSession (create): %v", err)
	}

	// Update the same session
	session.DurationSeconds = 30
	if err := svc.TrackSession(session); err != nil {
		t.Fatalf("TrackSession (update): %v", err)
	}

	// Should still be 1 row
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM analytics_sessions WHERE session_id = ?`, session.SessionID).Scan(&count) //nolint:errcheck
	if count != 1 {
		t.Errorf("TrackSession update: expected 1 row, got %d", count)
	}
}

func TestAnalyticsService_TrackSession_Disabled(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	_, err := db.Exec(`UPDATE profiles SET analytics_enabled = 0 WHERE id = ?`, profileID)
	if err != nil {
		t.Fatalf("disable analytics: %v", err)
	}

	session := &model.AnalyticsSession{
		ProfileID: profileID,
		SessionID: "disabled-session",
		IPHash:    "hash",
		CreatedAt: time.Now(),
	}

	if err := svc.TrackSession(session); err != nil {
		t.Fatalf("TrackSession disabled: %v", err)
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM analytics_sessions WHERE session_id = ?`, session.SessionID).Scan(&count) //nolint:errcheck
	if count != 0 {
		t.Errorf("TrackSession disabled: expected 0, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// AggregateHourly
// ---------------------------------------------------------------------------

func TestAnalyticsService_AggregateHourly(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	// Track a view first
	if err := svc.TrackView(profileID, "1.2.3.4", "Mozilla/5.0", ""); err != nil {
		t.Fatalf("TrackView: %v", err)
	}

	hour := time.Now().Truncate(time.Hour)
	if err := svc.AggregateHourly(profileID, hour); err != nil {
		t.Fatalf("AggregateHourly: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM analytics_hourly WHERE profile_id = ?`, profileID).Scan(&count); err != nil {
		t.Fatalf("count analytics_hourly: %v", err)
	}
	if count != 1 {
		t.Errorf("AggregateHourly: expected 1 row, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// GetSummary
// ---------------------------------------------------------------------------

func TestAnalyticsService_GetSummary_Empty(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(time.Hour)

	summary, err := svc.GetSummary(profileID, start, end)
	if err != nil {
		t.Fatalf("GetSummary (empty): %v", err)
	}
	if summary == nil {
		t.Fatal("GetSummary returned nil")
	}
	if summary.TotalViews != 0 {
		t.Errorf("TotalViews = %d, want 0", summary.TotalViews)
	}
}

func TestAnalyticsService_GetSummary_WithData(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	// Track some views
	for i := 0; i < 3; i++ {
		svc.TrackView(profileID, "1.2.3.4", "Mozilla/5.0 Chrome/91", "https://ref.test") //nolint:errcheck
	}

	start := time.Now().Add(-time.Hour)
	end := time.Now().Add(time.Hour)

	summary, err := svc.GetSummary(profileID, start, end)
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if summary.TotalViews != 3 {
		t.Errorf("TotalViews = %d, want 3", summary.TotalViews)
	}
}

// ---------------------------------------------------------------------------
// CleanupOldData
// ---------------------------------------------------------------------------

func TestAnalyticsService_CleanupOldData(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	// Insert an old analytics event directly
	_, err := db.Exec(
		`INSERT INTO analytics (id, profile_id, event_type, ip_hash, created_at)
		 VALUES ('old-event-001', ?, 'view', 'hash123', ?)`,
		profileID,
		time.Now().AddDate(-2, 0, 0), // 2 years old
	)
	if err != nil {
		t.Fatalf("insert old event: %v", err)
	}

	if err := svc.CleanupOldData(); err != nil {
		t.Fatalf("CleanupOldData: %v", err)
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM analytics WHERE id = 'old-event-001'`).Scan(&count) //nolint:errcheck
	if count != 0 {
		t.Error("CleanupOldData: old event should have been deleted")
	}
}

// ---------------------------------------------------------------------------
// getCountryFromIP
// ---------------------------------------------------------------------------

func TestAnalyticsService_GetCountryFromIP(t *testing.T) {
	db, _ := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	tests := []struct {
		ip   string
		want string
	}{
		{"127.0.0.1", "Local"},
		{"192.168.1.1", "Local"},
		{"", "Unknown"},     // invalid IP
		{"1.2.3.4", "Unknown"}, // real IP, no GeoIP DB
	}

	for _, tt := range tests {
		got := svc.getCountryFromIP(tt.ip)
		if got != tt.want {
			t.Errorf("getCountryFromIP(%q) = %q, want %q", tt.ip, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// isAnalyticsEnabled
// ---------------------------------------------------------------------------

func TestAnalyticsService_IsAnalyticsEnabled(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	enabled, err := svc.isAnalyticsEnabled(profileID)
	if err != nil {
		t.Fatalf("isAnalyticsEnabled: %v", err)
	}
	if !enabled {
		t.Error("isAnalyticsEnabled should be true")
	}

	_, err = db.Exec(`UPDATE profiles SET analytics_enabled = 0 WHERE id = ?`, profileID)
	if err != nil {
		t.Fatalf("disable: %v", err)
	}

	enabled, err = svc.isAnalyticsEnabled(profileID)
	if err != nil {
		t.Fatalf("isAnalyticsEnabled (disabled): %v", err)
	}
	if enabled {
		t.Error("isAnalyticsEnabled should be false after disabling")
	}
}

// ---------------------------------------------------------------------------
// sessionExists
// ---------------------------------------------------------------------------

func TestAnalyticsService_SessionExists(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	exists, err := svc.sessionExists("nonexistent-session")
	if err != nil {
		t.Fatalf("sessionExists(nonexistent): %v", err)
	}
	if exists {
		t.Error("sessionExists should be false for nonexistent session")
	}

	// Create a session
	session := &model.AnalyticsSession{
		ProfileID:  profileID,
		SessionID:  "exists-session-001",
		IPHash:     "hash123",
		DeviceType: model.DeviceTypeDesktop,
		CreatedAt:  time.Now(),
	}
	if err := svc.TrackSession(session); err != nil {
		t.Fatalf("TrackSession: %v", err)
	}

	exists, err = svc.sessionExists("exists-session-001")
	if err != nil {
		t.Fatalf("sessionExists(existing): %v", err)
	}
	if !exists {
		t.Error("sessionExists should be true after TrackSession")
	}
}

// ---------------------------------------------------------------------------
// generateID (analytics service)
// ---------------------------------------------------------------------------

func TestAnalyticsService_GenerateID_Unique(t *testing.T) {
	db, _ := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	ids := make(map[string]bool)
	for i := 0; i < 20; i++ {
		id := svc.generateID()
		if id == "" {
			t.Fatal("generateID returned empty string")
		}
		if ids[id] {
			t.Errorf("generateID returned duplicate: %s", id)
		}
		ids[id] = true
	}
}
