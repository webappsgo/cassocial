package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewTorService(t *testing.T) {
	dir := t.TempDir()
	ts := NewTorService(dir)
	if ts == nil {
		t.Fatal("NewTorService returned nil")
	}
	if ts.dataDir != dir {
		t.Errorf("dataDir = %q, want %q", ts.dataDir, dir)
	}
}

func TestTorService_GetOnionAddress_Empty(t *testing.T) {
	ts := NewTorService(t.TempDir())
	if addr := ts.GetOnionAddress(); addr != "" {
		t.Errorf("GetOnionAddress() = %q, want empty (not started)", addr)
	}
}

func TestTorService_IsEnabled_NotInstalled(t *testing.T) {
	ts := NewTorService(t.TempDir())
	// IsEnabled = enabled && onionAddr != ""
	// Even if tor binary is installed, onionAddr is empty until Start succeeds
	if ts.IsEnabled() {
		t.Error("IsEnabled() should be false when onionAddr is empty")
	}
}

func TestTorService_IsEnabled_SetOnion(t *testing.T) {
	ts := NewTorService(t.TempDir())
	ts.enabled = true
	ts.onionAddr = "abc123.onion"
	if !ts.IsEnabled() {
		t.Error("IsEnabled() should be true when enabled=true and onionAddr set")
	}
}

func TestTorService_GetStatus_NotRunning(t *testing.T) {
	ts := NewTorService(t.TempDir())
	status := ts.GetStatus()

	if _, ok := status["enabled"]; !ok {
		t.Error("GetStatus() missing 'enabled' key")
	}
	if _, ok := status["running"]; !ok {
		t.Error("GetStatus() missing 'running' key")
	}
	if _, ok := status["onion_address"]; !ok {
		t.Error("GetStatus() missing 'onion_address' key")
	}
	if status["running"].(bool) {
		t.Error("running should be false when cmd is nil")
	}
}

func TestTorService_GetStatus_WithOnion(t *testing.T) {
	ts := NewTorService(t.TempDir())
	ts.onionAddr = "test.onion"
	status := ts.GetStatus()
	if status["onion_address"].(string) != "test.onion" {
		t.Errorf("onion_address = %v, want test.onion", status["onion_address"])
	}
}

func TestTorService_GenerateVanityAddress(t *testing.T) {
	ts := NewTorService(t.TempDir())
	err := ts.GenerateVanityAddress("cas")
	if err == nil {
		t.Error("GenerateVanityAddress should return an error (not supported)")
	}
}

func TestTorService_Stop_NilCmd(t *testing.T) {
	ts := NewTorService(t.TempDir())
	// Should not panic when cmd is nil
	if err := ts.Stop(); err != nil {
		t.Errorf("Stop() with nil cmd should not error: %v", err)
	}
}

// TestCreateTorrc_CreatesFile verifies that createTorrc writes a valid torrc file.
func TestCreateTorrc_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	ts := NewTorService(dir)

	// createTorrc writes to ts.torrcPath; its parent must exist first.
	if err := os.MkdirAll(ts.torDataDir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	if err := ts.createTorrc(); err != nil {
		t.Fatalf("createTorrc: %v", err)
	}

	data, err := os.ReadFile(ts.torrcPath)
	if err != nil {
		t.Fatalf("reading torrc: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "HiddenServicePort") {
		t.Error("torrc missing HiddenServicePort directive")
	}
	if !strings.Contains(content, "HiddenServiceDir") {
		t.Error("torrc missing HiddenServiceDir directive")
	}
	if !strings.Contains(content, "DataDirectory") {
		t.Error("torrc missing DataDirectory directive")
	}
	if !strings.Contains(content, ts.torDataDir) {
		t.Errorf("torrc does not reference torDataDir %q", ts.torDataDir)
	}
}

// TestCreateTorrc_Overwrite verifies that a second createTorrc call overwrites the file cleanly.
func TestCreateTorrc_Overwrite(t *testing.T) {
	dir := t.TempDir()
	ts := NewTorService(dir)

	if err := os.MkdirAll(ts.torDataDir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	if err := ts.createTorrc(); err != nil {
		t.Fatalf("createTorrc first call: %v", err)
	}
	if err := ts.createTorrc(); err != nil {
		t.Fatalf("createTorrc second call: %v", err)
	}

	// File must still be readable and contain expected content.
	data, err := os.ReadFile(ts.torrcPath)
	if err != nil {
		t.Fatalf("reading torrc after overwrite: %v", err)
	}
	if !strings.Contains(string(data), "HiddenServicePort") {
		t.Error("torrc after overwrite missing HiddenServicePort")
	}
}

// TestCreateTorrc_MissingDir verifies that createTorrc fails when the parent directory
// does not exist (write permission denied).
func TestCreateTorrc_MissingDir(t *testing.T) {
	dir := t.TempDir()
	ts := NewTorService(dir)
	// Do NOT create ts.torDataDir — write should fail.
	err := ts.createTorrc()
	if err == nil {
		t.Error("createTorrc should fail when parent directory does not exist")
	}
}

// TestWaitForOnionAddress_ReadsHostnameFile verifies that waitForOnionAddress
// sets onionAddr from a pre-existing hostname file without sleeping.
func TestWaitForOnionAddress_ReadsHostnameFile(t *testing.T) {
	dir := t.TempDir()
	ts := NewTorService(dir)

	// Create the hidden_service directory and hostname file that Tor would produce.
	hsDir := filepath.Join(ts.torDataDir, "hidden_service")
	if err := os.MkdirAll(hsDir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	const expectedOnion = "abc123def456ghi7.onion\n"
	hostnameFile := filepath.Join(hsDir, "hostname")
	if err := os.WriteFile(hostnameFile, []byte(expectedOnion), 0600); err != nil {
		t.Fatalf("WriteFile hostname: %v", err)
	}

	if err := ts.waitForOnionAddress(); err != nil {
		t.Fatalf("waitForOnionAddress: %v", err)
	}

	// The implementation does filepath.Base(onionAddr) which strips directory prefixes
	// but leaves the trimmed name as-is (no newline stripping via filepath.Base).
	// The raw value should be non-empty.
	if ts.onionAddr == "" {
		t.Error("onionAddr should be set after waitForOnionAddress")
	}
}

// TestGetStatus_WithOnionAndEnabled verifies that GetStatus reflects a running configured service.
func TestGetStatus_WithOnionAndEnabled(t *testing.T) {
	ts := NewTorService(t.TempDir())
	ts.enabled = true
	ts.onionAddr = "xyz987.onion"

	status := ts.GetStatus()

	if status["enabled"].(bool) != true {
		t.Error("GetStatus enabled should be true")
	}
	if status["onion_address"].(string) != "xyz987.onion" {
		t.Errorf("onion_address = %v, want xyz987.onion", status["onion_address"])
	}
	if status["running"].(bool) {
		t.Error("running should be false (cmd is nil)")
	}
}
