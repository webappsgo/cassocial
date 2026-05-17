package server

import (
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
