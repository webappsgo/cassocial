package service

import (
	"testing"

	"github.com/casapps/cassocial/src/server/store"
)

func newTestAnalyticsService(t *testing.T) *AnalyticsService {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	return NewAnalyticsService(db)
}

func TestHashIP_Deterministic(t *testing.T) {
	s := newTestAnalyticsService(t)

	h1 := s.hashIP("192.168.1.1")
	h2 := s.hashIP("192.168.1.1")
	if h1 != h2 {
		t.Errorf("hashIP is not deterministic: %q != %q", h1, h2)
	}
}

func TestHashIP_DifferentInputs(t *testing.T) {
	s := newTestAnalyticsService(t)

	h1 := s.hashIP("1.2.3.4")
	h2 := s.hashIP("5.6.7.8")
	if h1 == h2 {
		t.Errorf("hashIP returned same hash for different IPs")
	}
}

func TestHashIP_NonEmpty(t *testing.T) {
	s := newTestAnalyticsService(t)

	h := s.hashIP("10.0.0.1")
	if h == "" {
		t.Error("hashIP returned empty string")
	}
}

func TestParseDeviceType_Mobile(t *testing.T) {
	s := newTestAnalyticsService(t)

	tests := []struct {
		ua string
	}{
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/537.36"},
		{"Mozilla/5.0 (Linux; Android 11; Pixel 5) AppleWebKit/537.36 Mobile"},
		{"BlackBerry Bold 9900"},
	}

	for _, tt := range tests {
		t.Run(tt.ua[:20], func(t *testing.T) {
			got := s.parseDeviceType(tt.ua)
			if got != "mobile" {
				t.Errorf("parseDeviceType(%q) = %q, want mobile", tt.ua, got)
			}
		})
	}
}

func TestParseDeviceType_Tablet(t *testing.T) {
	s := newTestAnalyticsService(t)

	tests := []string{
		"Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X) AppleWebKit/537.36",
		"Mozilla/5.0 (Linux; Android 4.4.2; Tablet Build/KOT49H)",
	}

	for _, ua := range tests {
		t.Run(ua[:20], func(t *testing.T) {
			got := s.parseDeviceType(ua)
			if got != "tablet" {
				t.Errorf("parseDeviceType(%q) = %q, want tablet", ua, got)
			}
		})
	}
}

func TestParseDeviceType_Desktop(t *testing.T) {
	s := newTestAnalyticsService(t)

	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/91.0"
	got := s.parseDeviceType(ua)
	if got != "desktop" {
		t.Errorf("parseDeviceType(%q) = %q, want desktop", ua, got)
	}
}

func TestParseDeviceType_EmptyUA(t *testing.T) {
	s := newTestAnalyticsService(t)

	got := s.parseDeviceType("")
	if got != "desktop" {
		t.Errorf("parseDeviceType('') = %q, want desktop", got)
	}
}

func TestGetRetentionDays_Default(t *testing.T) {
	s := newTestAnalyticsService(t)

	days, err := s.getRetentionDays()
	if err != nil {
		t.Fatalf("getRetentionDays returned error: %v", err)
	}
	if days <= 0 {
		t.Errorf("getRetentionDays returned %d, want > 0", days)
	}
}

func TestGetRetentionDays_CustomValue(t *testing.T) {
	s := newTestAnalyticsService(t)

	// Set a custom retention period.
	if err := s.db.SetSetting("analytics_retention_days", "30"); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}

	days, err := s.getRetentionDays()
	if err != nil {
		t.Fatalf("getRetentionDays returned error: %v", err)
	}
	if days != 30 {
		t.Errorf("getRetentionDays returned %d, want 30", days)
	}
}

func TestGetRetentionDays_SettingMissing(t *testing.T) {
	s := newTestAnalyticsService(t)

	// Delete the setting so GetSetting returns sql.ErrNoRows.
	if _, err := s.db.Exec(`DELETE FROM settings WHERE key = 'analytics_retention_days'`); err != nil {
		t.Fatalf("delete setting: %v", err)
	}

	days, err := s.getRetentionDays()
	if err != nil {
		t.Fatalf("getRetentionDays with missing setting returned error: %v", err)
	}
	if days != 90 {
		t.Errorf("getRetentionDays with missing setting = %d, want 90 (default)", days)
	}
}
