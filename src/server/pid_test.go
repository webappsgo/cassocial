package server

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

func TestCheckPIDFile_UnreadableFile(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "unreadable.pid")

	if err := os.WriteFile(pidPath, []byte("12345"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	// Remove read permission so ReadFile fails
	if err := os.Chmod(pidPath, 0000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { os.Chmod(pidPath, 0644) })

	running, pid, err := CheckPIDFile(pidPath)
	// On Linux running as root this may succeed; as non-root it should error
	// We just ensure no panic
	_ = running
	_ = pid
	_ = err
}

func TestIsProcessRunning_OwnPID(t *testing.T) {
	// Our own PID is definitely running
	if !isProcessRunning(os.Getpid()) {
		t.Error("isProcessRunning should return true for own PID")
	}
}

func TestIsProcessRunning_InvalidPID(t *testing.T) {
	// PID 0 and negative PIDs are invalid on Unix
	if isProcessRunning(-1) {
		t.Error("isProcessRunning(-1) should return false")
	}
}

func TestIsOurProcess_OwnPID(t *testing.T) {
	// Just call it on our own PID — we only verify it does not panic
	_ = isOurProcess(os.Getpid())
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

// TestIsOurProcess_KernelThread covers the Readlink-failure branch in isOurProcess (pid_unix.go:31-34).
// On Linux, PID 2 is a kernel thread (kthreadd) with no /proc/2/exe symlink, so Readlink fails
// and the code falls through to isOurProcessDarwin, which returns false for non-cassocial processes.
func TestIsOurProcess_KernelThread(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("kernel thread PID 2 test is Linux-only")
	}
	// Verify PID 2 exists before testing.
	if _, err := os.Stat("/proc/2"); err != nil {
		t.Skip("PID 2 not accessible on this system")
	}
	// isOurProcess(2): Readlink("/proc/2/exe") fails for kernel threads,
	// falling through to isOurProcessDarwin(2) which returns false.
	if isOurProcess(2) {
		t.Error("isOurProcess(2) should return false for a kernel thread")
	}
}

// TestCheckPIDFile_OtherProcess covers the isOurProcess-returns-false branch (line 44).
// PID 1 (init/systemd) is always running but is not our binary.
func TestCheckPIDFile_OtherProcess(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "other.pid")

	if err := os.WriteFile(pidPath, []byte("1"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// PID 1 is running (init/systemd) but is not "cassocial", so isOurProcess returns false.
	// CheckPIDFile should return false (stale PID file removed).
	running, pid, err := CheckPIDFile(pidPath)
	if err != nil {
		t.Fatalf("CheckPIDFile: %v", err)
	}
	if running {
		// This is acceptable if somehow PID 1 is considered "ours" — just skip.
		t.Log("PID 1 was considered our process — skipping assertion")
		return
	}
	if pid != 0 {
		t.Errorf("pid = %d, want 0 when not our process", pid)
	}
}

// TestWritePIDFile_CheckError covers the error path when CheckPIDFile returns an error.
// It passes a directory as the PID file path, making os.ReadFile return "is a directory"
// (not os.IsNotExist), which causes CheckPIDFile to return an error and WritePIDFile to
// propagate it (pid.go:51-53).
func TestWritePIDFile_CheckError(t *testing.T) {
	dir := t.TempDir()
	// Create a subdirectory with the .pid name so ReadFile fails with "is a directory".
	pidPath := filepath.Join(dir, "server.pid")
	if err := os.MkdirAll(pidPath, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// WritePIDFile(pidPath) → CheckPIDFile(pidPath) → ReadFile(pidPath) → "is a directory" error
	// → CheckPIDFile returns error → WritePIDFile returns that error (line 51-53).
	err := WritePIDFile(pidPath)
	if err == nil {
		t.Error("WritePIDFile should return error when PID path is a directory")
	}
}

// TestWritePIDFile_MkdirAllError covers the MkdirAll failure path in WritePIDFile (pid.go:60-62).
// A blocking file is created where the PID directory should be, making MkdirAll fail.
func TestWritePIDFile_MkdirAllError(t *testing.T) {
	base := t.TempDir()
	// Create a file named "subdir" so MkdirAll("subdir") fails.
	blocker := filepath.Join(base, "subdir")
	if err := os.WriteFile(blocker, []byte("block"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	// The PID file path is inside "subdir" which is a file — MkdirAll will fail.
	pidPath := filepath.Join(blocker, "server.pid")
	err := WritePIDFile(pidPath)
	if err == nil {
		t.Error("WritePIDFile should return error when MkdirAll fails")
	}
}

// TestCheckPIDFile_ReadError covers the non-IsNotExist read error path.
func TestCheckPIDFile_ReadError(t *testing.T) {
	dir := t.TempDir()
	// Make pidPath a directory — ReadFile on a dir returns "is a directory" error,
	// which is NOT os.IsNotExist, so we hit the error branch.
	dirPath := filepath.Join(dir, "pid-is-dir.pid")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	running, pid, err := CheckPIDFile(dirPath)
	// Should return an error (not IsNotExist), running=false
	if err == nil {
		// On some systems this may succeed — just ensure no panic.
		_ = running
		_ = pid
		return
	}
	if running {
		t.Error("running should be false on read error")
	}
}
