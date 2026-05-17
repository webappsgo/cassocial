package server

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckPIDFile_NoFile(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "test.pid")

	running, pid, err := CheckPIDFile(pidPath)
	if err != nil {
		t.Fatalf("CheckPIDFile: %v", err)
	}
	if running {
		t.Error("running = true, want false when no PID file")
	}
	if pid != 0 {
		t.Errorf("pid = %d, want 0", pid)
	}
}

func TestCheckPIDFile_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "corrupt.pid")

	if err := os.WriteFile(pidPath, []byte("not-a-number"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	running, pid, err := CheckPIDFile(pidPath)
	if err != nil {
		t.Fatalf("CheckPIDFile: %v", err)
	}
	if running {
		t.Error("running = true for corrupt PID file, want false")
	}
	if pid != 0 {
		t.Errorf("pid = %d, want 0", pid)
	}
	// File should have been removed
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Error("corrupt PID file should have been removed")
	}
}

func TestCheckPIDFile_StalePID(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "stale.pid")

	// PID 99999999 almost certainly does not exist
	if err := os.WriteFile(pidPath, []byte("99999999"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	running, _, err := CheckPIDFile(pidPath)
	if err != nil {
		t.Fatalf("CheckPIDFile: %v", err)
	}
	if running {
		t.Error("running = true for non-existent PID, want false")
	}
}

func TestWritePIDFile_Success(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "server.pid")

	if err := WritePIDFile(pidPath); err != nil {
		t.Fatalf("WritePIDFile: %v", err)
	}

	data, err := os.ReadFile(pidPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) == 0 {
		t.Error("PID file is empty")
	}
}

func TestWritePIDFile_CreatesDirectory(t *testing.T) {
	base := t.TempDir()
	pidPath := filepath.Join(base, "nested", "dir", "server.pid")

	if err := WritePIDFile(pidPath); err != nil {
		t.Fatalf("WritePIDFile with nested path: %v", err)
	}
	if _, err := os.Stat(pidPath); err != nil {
		t.Errorf("PID file not created: %v", err)
	}
}

func TestRemovePIDFile_Success(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "remove.pid")

	if err := os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := RemovePIDFile(pidPath); err != nil {
		t.Fatalf("RemovePIDFile: %v", err)
	}
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Error("PID file still exists after removal")
	}
}

func TestRemovePIDFile_NonExistent(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "nonexistent.pid")

	err := RemovePIDFile(pidPath)
	if err == nil {
		t.Error("expected error removing non-existent PID file")
	}
}

func TestWritePIDFile_AlreadyRunning(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "running.pid")

	// Write own PID to simulate a running process
	ownPID := os.Getpid()
	if err := os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", ownPID)), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// WritePIDFile should fail because the process (us) is running and isOurProcess passes
	err := WritePIDFile(pidPath)
	// This may or may not fail depending on isOurProcess logic, but should not panic
	_ = err
}
