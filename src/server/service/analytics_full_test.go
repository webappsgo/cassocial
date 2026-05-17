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

// ---------------------------------------------------------------------------
// TrackView — DB error path (analytics table dropped before insert)
// ---------------------------------------------------------------------------

func TestAnalyticsService_TrackView_DBError(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	// Drop the analytics table so the INSERT fails while the profiles SELECT succeeds.
	_, err := db.Exec(`DROP TABLE analytics`)
	if err != nil {
		t.Fatalf("drop analytics table: %v", err)
	}

	err = svc.TrackView(profileID, "1.2.3.4", "Mozilla/5.0", "")
	if err == nil {
		t.Error("TrackView with dropped analytics table should return an error")
	}
}

// ---------------------------------------------------------------------------
// TrackClick — DB error path (analytics table insert fails)
// ---------------------------------------------------------------------------

func TestAnalyticsService_TrackClick_DBError(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	_, err := db.Exec(`DROP TABLE analytics`)
	if err != nil {
		t.Fatalf("drop analytics table: %v", err)
	}

	err = svc.TrackClick(profileID, "some-link", "1.2.3.4", "Mozilla/5.0", "")
	if err == nil {
		t.Error("TrackClick with dropped analytics table should return an error")
	}
}

// ---------------------------------------------------------------------------
// TrackSession — validation error (invalid device type)
// ---------------------------------------------------------------------------

func TestAnalyticsService_TrackSession_ValidationError(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	// Invalid DeviceType triggers Validate() failure.
	session := &model.AnalyticsSession{
		ProfileID:  profileID,
		SessionID:  "bad-device-session",
		IPHash:     "hash123",
		DeviceType: "supercomputer", // not mobile/tablet/desktop
		CreatedAt:  time.Now(),
	}

	err := svc.TrackSession(session)
	if err == nil {
		t.Error("TrackSession with invalid device type should return an error")
	}
}

// ---------------------------------------------------------------------------
// TrackSession — zero CreatedAt auto-populated
// ---------------------------------------------------------------------------

func TestAnalyticsService_TrackSession_ZeroCreatedAt(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	session := &model.AnalyticsSession{
		ProfileID:  profileID,
		SessionID:  "zero-time-session",
		IPHash:     "hash123",
		DeviceType: model.DeviceTypeDesktop,
		// CreatedAt deliberately left as zero value
	}

	if err := svc.TrackSession(session); err != nil {
		t.Fatalf("TrackSession with zero CreatedAt: %v", err)
	}
	if session.CreatedAt.IsZero() {
		t.Error("TrackSession should have set CreatedAt when it was zero")
	}
}

// ---------------------------------------------------------------------------
// TrackSession — sessionExists DB error (analytics_sessions table dropped before check)
// ---------------------------------------------------------------------------

func TestAnalyticsService_TrackSession_SessionExistsError(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	_, err := db.Exec(`DROP TABLE analytics_sessions`)
	if err != nil {
		t.Fatalf("drop analytics_sessions: %v", err)
	}

	session := &model.AnalyticsSession{
		ProfileID:  profileID,
		SessionID:  "check-error-session",
		IPHash:     "hash123",
		DeviceType: model.DeviceTypeDesktop,
		CreatedAt:  time.Now(),
	}

	err = svc.TrackSession(session)
	if err == nil {
		t.Error("TrackSession with dropped sessions table should return an error from sessionExists")
	}
}

// ---------------------------------------------------------------------------
// TrackSession — DB error on insert (duplicate primary key)
// ---------------------------------------------------------------------------

func TestAnalyticsService_TrackSession_InsertError(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	const conflictID = "conflict-id-0001"
	const conflictSessionID = "conflict-session-unique"

	// Pre-insert a row using the ID we will force the service to use.
	_, err := db.Exec(`
		INSERT INTO analytics_sessions (id, profile_id, session_id, ip_hash, created_at)
		VALUES (?, ?, 'pre-existing-session', 'hash', CURRENT_TIMESTAMP)
	`, conflictID, profileID)
	if err != nil {
		t.Fatalf("pre-insert session: %v", err)
	}

	// Create a session with the same ID — INSERT will fail with PK conflict.
	session := &model.AnalyticsSession{
		ID:         conflictID, // forced duplicate PK
		ProfileID:  profileID,
		SessionID:  conflictSessionID, // new session_id so sessionExists returns false
		IPHash:     "hash123",
		DeviceType: model.DeviceTypeDesktop,
		CreatedAt:  time.Now(),
	}

	err = svc.TrackSession(session)
	if err == nil {
		t.Error("TrackSession with duplicate primary key should return an error")
	}
}

// ---------------------------------------------------------------------------
// TrackSession — DB error on update (BEFORE UPDATE trigger raises an error)
// ---------------------------------------------------------------------------

func TestAnalyticsService_TrackSession_UpdateError(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	// Create session first.
	session := &model.AnalyticsSession{
		ProfileID:  profileID,
		SessionID:  "update-error-001",
		IPHash:     "hash456",
		DeviceType: model.DeviceTypeDesktop,
		CreatedAt:  time.Now(),
	}
	if err := svc.TrackSession(session); err != nil {
		t.Fatalf("TrackSession (create): %v", err)
	}

	// Install a trigger that blocks UPDATE on analytics_sessions.
	_, err := db.Exec(`
		CREATE TRIGGER block_sessions_update
		BEFORE UPDATE ON analytics_sessions
		BEGIN
			SELECT RAISE(ABORT, 'updates blocked by test trigger');
		END
	`)
	if err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	session.DurationSeconds = 60
	err = svc.TrackSession(session)
	if err == nil {
		t.Error("TrackSession update with blocking trigger should return an error")
	}
}

// ---------------------------------------------------------------------------
// AggregateHourly — DB error path
// ---------------------------------------------------------------------------

func TestAnalyticsService_AggregateHourly_DBError(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	_, err := db.Exec(`DROP TABLE analytics_hourly`)
	if err != nil {
		t.Fatalf("drop analytics_hourly: %v", err)
	}

	err = svc.AggregateHourly(profileID, time.Now())
	if err == nil {
		t.Error("AggregateHourly with dropped table should return an error")
	}
}

// ---------------------------------------------------------------------------
// GetSummary — DB error paths (various query failures)
// ---------------------------------------------------------------------------

func TestAnalyticsService_GetSummary_AnalyticsTableDropped(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	_, err := db.Exec(`DROP TABLE analytics`)
	if err != nil {
		t.Fatalf("drop analytics: %v", err)
	}

	start := time.Now().Add(-time.Hour)
	end := time.Now().Add(time.Hour)
	_, err = svc.GetSummary(profileID, start, end)
	if err == nil {
		t.Error("GetSummary with dropped analytics table should return an error")
	}
}

func TestAnalyticsService_GetSummary_SessionsTableDropped(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	_, err := db.Exec(`DROP TABLE analytics_sessions`)
	if err != nil {
		t.Fatalf("drop analytics_sessions: %v", err)
	}

	start := time.Now().Add(-time.Hour)
	end := time.Now().Add(time.Hour)
	_, err = svc.GetSummary(profileID, start, end)
	if err == nil {
		t.Error("GetSummary with dropped analytics_sessions table should return an error")
	}
}

func TestAnalyticsService_GetSummary_HourlyTableDropped(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	// Need analytics and analytics_sessions intact, only drop analytics_hourly.
	_, err := db.Exec(`DROP TABLE analytics_hourly`)
	if err != nil {
		t.Fatalf("drop analytics_hourly: %v", err)
	}

	start := time.Now().Add(-time.Hour)
	end := time.Now().Add(time.Hour)
	_, err = svc.GetSummary(profileID, start, end)
	if err == nil {
		t.Error("GetSummary with dropped analytics_hourly table should return an error")
	}
}

func TestAnalyticsService_GetSummary_LinksTableDropped(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	// Drop only links so that analytics/sessions/device/referrer/country queries succeed
	// but the link click stats query (JOIN with links) fails.
	_, err := db.Exec(`DROP TABLE links`)
	if err != nil {
		t.Fatalf("drop links: %v", err)
	}

	start := time.Now().Add(-time.Hour)
	end := time.Now().Add(time.Hour)
	_, err = svc.GetSummary(profileID, start, end)
	if err == nil {
		t.Error("GetSummary with dropped links table should return an error")
	}
}

func TestAnalyticsService_GetSummary_WithReferrersAndCountries(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	// Insert a link for the click event.
	_, err := db.Exec(`
		INSERT INTO links (id, profile_id, title, url, position, is_active, click_count, created_at, updated_at)
		VALUES ('sum-link-001', ?, 'Example', 'https://example.com', 1, 1, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, profileID)
	if err != nil {
		t.Fatalf("insert link: %v", err)
	}

	// Insert views with referrer and country, and a click with a link_id.
	_, err = db.Exec(`
		INSERT INTO analytics (id, profile_id, link_id, event_type, ip_hash, user_agent, referrer, country, device_type, created_at)
		VALUES ('sum-001', ?, NULL,          'view',  'hash1', 'Mozilla', 'https://google.com',  'US', 'desktop', CURRENT_TIMESTAMP),
		       ('sum-002', ?, NULL,          'view',  'hash2', 'Mozilla', 'https://twitter.com', 'GB', 'mobile',  CURRENT_TIMESTAMP),
		       ('sum-003', ?, 'sum-link-001','click', 'hash3', 'Mozilla', '',                    'US', 'desktop', CURRENT_TIMESTAMP)
	`, profileID, profileID, profileID)
	if err != nil {
		t.Fatalf("insert analytics rows: %v", err)
	}

	// Aggregate hourly data so hourly stats are present.
	hour := time.Now().Truncate(time.Hour)
	if err := svc.AggregateHourly(profileID, hour); err != nil {
		t.Fatalf("AggregateHourly: %v", err)
	}

	// Also insert a session so avg duration is non-zero.
	_, err = db.Exec(`
		INSERT INTO analytics_sessions (id, profile_id, session_id, ip_hash, device_type, duration_seconds, created_at)
		VALUES ('sum-sess-001', ?, 'sum-sess-id', 'hash9', 'desktop', 120, CURRENT_TIMESTAMP)
	`, profileID)
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}

	start := time.Now().Add(-time.Hour)
	end := time.Now().Add(time.Hour)
	summary, err := svc.GetSummary(profileID, start, end)
	if err != nil {
		t.Fatalf("GetSummary with data: %v", err)
	}
	if summary.TotalViews != 2 {
		t.Errorf("TotalViews = %d, want 2", summary.TotalViews)
	}
	if summary.TotalClicks != 1 {
		t.Errorf("TotalClicks = %d, want 1", summary.TotalClicks)
	}
	if len(summary.TopReferrers) == 0 {
		t.Error("TopReferrers should have at least one entry")
	}
	if len(summary.TopCountries) == 0 {
		t.Error("TopCountries should have at least one entry")
	}
	if len(summary.DeviceBreakdown) == 0 {
		t.Error("DeviceBreakdown should have at least one entry")
	}
	if len(summary.LinkClickStats) == 0 {
		t.Error("LinkClickStats should have at least one entry")
	}
	if len(summary.HourlyStats) == 0 {
		t.Error("HourlyStats should have at least one entry")
	}
}

// ---------------------------------------------------------------------------
// GetSummary — scan error in hourly loop (analytics_hourly has wrong schema)
// ---------------------------------------------------------------------------

func TestAnalyticsService_GetSummary_HourlyScanError(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	// Replace analytics_hourly with a table that has fewer columns — scan will fail.
	if _, err := db.Exec(`DROP TABLE analytics_hourly`); err != nil {
		t.Fatalf("drop analytics_hourly: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE analytics_hourly (profile_id TEXT, hour TIMESTAMP)`); err != nil {
		t.Fatalf("create narrow analytics_hourly: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO analytics_hourly VALUES (?, CURRENT_TIMESTAMP)`, profileID); err != nil {
		t.Fatalf("insert narrow hourly row: %v", err)
	}

	start := time.Now().Add(-time.Hour)
	end := time.Now().Add(time.Hour)
	_, err := svc.GetSummary(profileID, start, end)
	if err == nil {
		t.Error("GetSummary with malformed analytics_hourly schema should return an error")
	}
}

// ---------------------------------------------------------------------------
// GetSummary — scan error in link click stats loop (links table has wrong schema)
// ---------------------------------------------------------------------------

func TestAnalyticsService_GetSummary_LinkScanError(t *testing.T) {
	db, profileID := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	// Insert a link and a click event so the linkQuery returns rows.
	if _, err := db.Exec(`
		INSERT INTO links (id, profile_id, title, url, position, is_active, click_count, created_at, updated_at)
		VALUES ('scan-link-001', ?, 'Test', 'https://example.com', 1, 1, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, profileID); err != nil {
		t.Fatalf("insert link: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO analytics (id, profile_id, link_id, event_type, ip_hash, created_at)
		VALUES ('scan-click-001', ?, 'scan-link-001', 'click', 'hash', CURRENT_TIMESTAMP)
	`, profileID); err != nil {
		t.Fatalf("insert click: %v", err)
	}

	// Rename links to links_backup, create a narrow fake links table with only one column
	// so the 3-column scan in GetSummary fails.
	if _, err := db.Exec(`ALTER TABLE links RENAME TO links_backup`); err != nil {
		t.Fatalf("rename links: %v", err)
	}
	// Create a fake links with only one column (id) — the query selects l.id, l.title, COUNT(*)
	// but l.title won't exist → query itself fails with "no such column".
	if _, err := db.Exec(`CREATE TABLE links (id TEXT PRIMARY KEY)`); err != nil {
		t.Fatalf("create narrow links: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO links (id) VALUES ('scan-link-001')`); err != nil {
		t.Fatalf("insert into narrow links: %v", err)
	}

	start := time.Now().Add(-time.Hour)
	end := time.Now().Add(time.Hour)
	_, err := svc.GetSummary(profileID, start, end)
	if err == nil {
		t.Error("GetSummary with narrow links table should return an error")
	}
}

// ---------------------------------------------------------------------------
// CleanupOldData — second DELETE error (sessions table dropped after first)
// ---------------------------------------------------------------------------

func TestAnalyticsService_CleanupOldData_SessionDeleteError(t *testing.T) {
	db, _ := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	// Drop only analytics_sessions so the first DELETE succeeds but second fails.
	_, err := db.Exec(`DROP TABLE analytics_sessions`)
	if err != nil {
		t.Fatalf("drop analytics_sessions: %v", err)
	}

	err = svc.CleanupOldData()
	if err == nil {
		t.Error("CleanupOldData with dropped sessions table should return an error")
	}
}

func TestAnalyticsService_CleanupOldData_AnalyticsDeleteError(t *testing.T) {
	db, _ := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	_, err := db.Exec(`DROP TABLE analytics`)
	if err != nil {
		t.Fatalf("drop analytics: %v", err)
	}

	err = svc.CleanupOldData()
	if err == nil {
		t.Error("CleanupOldData with dropped analytics table should return an error")
	}
}

// ---------------------------------------------------------------------------
// isAnalyticsEnabled — profile not found (DB error)
// ---------------------------------------------------------------------------

func TestAnalyticsService_IsAnalyticsEnabled_ProfileNotFound(t *testing.T) {
	db, _ := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	_, err := svc.isAnalyticsEnabled("nonexistent-profile-id")
	if err == nil {
		t.Error("isAnalyticsEnabled for non-existent profile should return an error")
	}
}

// ---------------------------------------------------------------------------
// sessionExists — DB error path
// ---------------------------------------------------------------------------

func TestAnalyticsService_SessionExists_DBError(t *testing.T) {
	db, _ := newTestAnalyticsDB(t)
	svc := NewAnalyticsService(db)

	_, err := db.Exec(`DROP TABLE analytics_sessions`)
	if err != nil {
		t.Fatalf("drop analytics_sessions: %v", err)
	}

	_, err = svc.sessionExists("any-session")
	if err == nil {
		t.Error("sessionExists with dropped table should return an error")
	}
}

// ---------------------------------------------------------------------------
// getRetentionDays — non-numeric setting value falls back to default
// ---------------------------------------------------------------------------

func TestGetRetentionDays_NonNumericFallback(t *testing.T) {
	s := newTestAnalyticsService(t)

	if err := s.db.SetSetting("analytics_retention_days", "not-a-number"); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}

	days, err := s.getRetentionDays()
	if err != nil {
		t.Fatalf("getRetentionDays returned error: %v", err)
	}
	if days != 90 {
		t.Errorf("getRetentionDays with non-numeric value = %d, want 90 (default)", days)
	}
}
