package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestNewGeoIP_NoDatabase verifies that GeoIP is disabled when no database file exists.
func TestNewGeoIP_NoDatabase(t *testing.T) {
	g := NewGeoIP(t.TempDir())
	if g == nil {
		t.Fatal("NewGeoIP returned nil")
	}
	if g.IsEnabled() {
		t.Error("GeoIP should be disabled when database file is absent")
	}
}

// TestNewGeoIP_WithDatabase verifies that GeoIP is enabled when the database file exists.
func TestNewGeoIP_WithDatabase(t *testing.T) {
	dir := t.TempDir()
	dbDir := filepath.Join(dir, "security", "geoip")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Create a placeholder file to simulate the database.
	dbFile := filepath.Join(dbDir, "GeoLite2-Country.mmdb")
	if err := os.WriteFile(dbFile, []byte("placeholder"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	g := NewGeoIP(dir)
	if !g.IsEnabled() {
		t.Error("GeoIP should be enabled when database file exists")
	}
}

// TestIsEnabled_False verifies that IsEnabled returns false for a GeoIP with no DB.
func TestIsEnabled_False(t *testing.T) {
	g := NewGeoIP(t.TempDir())
	if g.IsEnabled() {
		t.Error("IsEnabled should be false when no database is present")
	}
}

// TestLookup_NotEnabled verifies that Lookup returns an error when GeoIP is disabled.
func TestLookup_NotEnabled(t *testing.T) {
	g := NewGeoIP(t.TempDir())
	_, _, err := g.Lookup("1.2.3.4")
	if err == nil {
		t.Error("Lookup with GeoIP disabled should return error")
	}
}

// TestCheckCountryBlocked_Disabled verifies that no IP is blocked when GeoIP is disabled.
func TestCheckCountryBlocked_Disabled(t *testing.T) {
	g := NewGeoIP(t.TempDir())
	blocked, err := g.CheckCountryBlocked("1.2.3.4", []string{"US", "CA"})
	if err != nil {
		t.Errorf("CheckCountryBlocked disabled: unexpected error: %v", err)
	}
	if blocked {
		t.Error("CheckCountryBlocked should return false when GeoIP is disabled")
	}
}

// TestCheckCountryBlocked_EmptyList verifies that no IP is blocked when blockedCountries is empty.
func TestCheckCountryBlocked_EmptyList(t *testing.T) {
	dir := t.TempDir()
	dbDir := filepath.Join(dir, "security", "geoip")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	dbFile := filepath.Join(dbDir, "GeoLite2-Country.mmdb")
	if err := os.WriteFile(dbFile, []byte("placeholder"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	g := NewGeoIP(dir)
	blocked, err := g.CheckCountryBlocked("1.2.3.4", []string{})
	if err != nil {
		t.Errorf("CheckCountryBlocked empty list: unexpected error: %v", err)
	}
	if blocked {
		t.Error("CheckCountryBlocked should return false for empty blocked list")
	}
}

// TestCheckCountryBlocked_Enabled_NotBlocked verifies that "XX" (stub country) is not blocked
// unless it's in the blocked list.
func TestCheckCountryBlocked_Enabled_NotBlocked(t *testing.T) {
	dir := t.TempDir()
	dbDir := filepath.Join(dir, "security", "geoip")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "GeoLite2-Country.mmdb"), []byte("placeholder"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	g := NewGeoIP(dir)
	// Lookup returns "XX" — not in blocked list.
	blocked, err := g.CheckCountryBlocked("1.2.3.4", []string{"US", "CA"})
	if err != nil {
		t.Errorf("CheckCountryBlocked: unexpected error: %v", err)
	}
	if blocked {
		t.Error("CheckCountryBlocked should return false when country code is not in blocked list")
	}
}

// TestCheckCountryBlocked_Enabled_Blocked verifies that "XX" is blocked when included in the list.
func TestCheckCountryBlocked_Enabled_Blocked(t *testing.T) {
	dir := t.TempDir()
	dbDir := filepath.Join(dir, "security", "geoip")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "GeoLite2-Country.mmdb"), []byte("placeholder"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	g := NewGeoIP(dir)
	// Stub Lookup always returns "XX".
	blocked, err := g.CheckCountryBlocked("1.2.3.4", []string{"XX", "US"})
	if err != nil {
		t.Errorf("CheckCountryBlocked: unexpected error: %v", err)
	}
	if !blocked {
		t.Error("CheckCountryBlocked should return true when stub country 'XX' is in the blocked list")
	}
}

// TestGeoIPMiddleware_PassThrough verifies that the middleware passes through when GeoIP is disabled.
func TestGeoIPMiddleware_PassThrough(t *testing.T) {
	g := NewGeoIP(t.TempDir())
	mw := GeoIPMiddleware(g, []string{"US"})

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	mw(next).ServeHTTP(rr, req)

	if !called {
		t.Error("GeoIPMiddleware with disabled GeoIP should call next handler")
	}
}

// TestGeoIPMiddleware_EmptyBlockedList verifies that no blocking occurs with an empty blocked list.
func TestGeoIPMiddleware_EmptyBlockedList(t *testing.T) {
	g := NewGeoIP(t.TempDir())
	mw := GeoIPMiddleware(g, []string{})

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	mw(next).ServeHTTP(rr, req)

	if !called {
		t.Error("GeoIPMiddleware with empty blocked list should call next handler")
	}
}

// TestGeoIPMiddleware_BlockedCountry verifies that a blocked country gets a 403.
func TestGeoIPMiddleware_BlockedCountry(t *testing.T) {
	dir := t.TempDir()
	dbDir := filepath.Join(dir, "security", "geoip")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "GeoLite2-Country.mmdb"), []byte("placeholder"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	g := NewGeoIP(dir)
	// Stub Lookup returns "XX" — block "XX".
	mw := GeoIPMiddleware(g, []string{"XX"})

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rr := httptest.NewRecorder()
	mw(next).ServeHTTP(rr, req)

	if called {
		t.Error("GeoIPMiddleware should not call next handler for blocked country")
	}
	if rr.Code != http.StatusForbidden {
		t.Errorf("GeoIPMiddleware blocked country: status %d, want %d", rr.Code, http.StatusForbidden)
	}
}

// TestGeoIPMiddleware_NotBlockedCountry verifies that a request from a non-blocked country
// passes through when GeoIP is enabled. The stub Lookup returns "XX", so blocking "US" only
// exercises the pass-through branch.
func TestGeoIPMiddleware_NotBlockedCountry(t *testing.T) {
	dir := t.TempDir()
	dbDir := filepath.Join(dir, "security", "geoip")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "GeoLite2-Country.mmdb"), []byte("placeholder"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	g := NewGeoIP(dir)
	// Block "US" only; stub Lookup returns "XX", so the request should pass through.
	mw := GeoIPMiddleware(g, []string{"US"})

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rr := httptest.NewRecorder()
	mw(next).ServeHTTP(rr, req)

	if !called {
		t.Error("GeoIPMiddleware should call next handler when country is not in the blocked list")
	}
	if rr.Code == http.StatusForbidden {
		t.Error("GeoIPMiddleware should not return 403 for a non-blocked country")
	}
}


// TestGetStats_Disabled verifies that GetStats returns correct fields when disabled.
func TestGetStats_Disabled(t *testing.T) {
	g := NewGeoIP(t.TempDir())
	stats := g.GetStats()
	if stats["enabled"] != false {
		t.Errorf("GetStats enabled = %v, want false", stats["enabled"])
	}
}

// TestGetStats_Enabled verifies that GetStats returns enabled=true when DB exists.
func TestGetStats_Enabled(t *testing.T) {
	dir := t.TempDir()
	dbDir := filepath.Join(dir, "security", "geoip")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "GeoLite2-Country.mmdb"), []byte("placeholder"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	g := NewGeoIP(dir)
	stats := g.GetStats()
	if stats["enabled"] != true {
		t.Errorf("GetStats enabled = %v, want true", stats["enabled"])
	}
}
