package service

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// newTestBackupService creates a BackupService wired to temp directories.
func newTestBackupService(t *testing.T) *BackupService {
	t.Helper()

	tmpDir := t.TempDir()

	// Backup service expects these directories to exist under DataDir.
	for _, sub := range []string{"db", "backup", "config"} {
		if err := os.MkdirAll(filepath.Join(tmpDir, sub), 0o755); err != nil {
			t.Fatalf("MkdirAll %s: %v", sub, err)
		}
	}

	cfg := &config.Config{
		DataDir:   tmpDir,
		ConfigDir: filepath.Join(tmpDir, "config"),
	}

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	return NewBackupService(cfg, db)
}

// ---------------------------------------------------------------------------
// NewBackupService
// ---------------------------------------------------------------------------

func TestNewBackupService(t *testing.T) {
	svc := newTestBackupService(t)
	if svc == nil {
		t.Fatal("NewBackupService returned nil")
	}
}

// ---------------------------------------------------------------------------
// CreateBackup
// ---------------------------------------------------------------------------

func TestBackupService_CreateBackup(t *testing.T) {
	svc := newTestBackupService(t)

	backup, err := svc.CreateBackup("manual")
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}
	if backup == nil {
		t.Fatal("CreateBackup returned nil Backup")
	}
	if backup.Filename == "" {
		t.Error("Backup filename is empty")
	}
	if backup.Type != "manual" {
		t.Errorf("Backup type = %q, want manual", backup.Type)
	}
	if backup.Size < 0 {
		t.Errorf("Backup size = %d, want >= 0", backup.Size)
	}

	// File should exist
	backupPath := filepath.Join(svc.config.DataDir, "backup", backup.Filename)
	if _, err := os.Stat(backupPath); err != nil {
		t.Errorf("backup file not found: %v", err)
	}
}

func TestBackupService_CreateBackup_Auto(t *testing.T) {
	svc := newTestBackupService(t)

	backup, err := svc.CreateBackup("auto")
	if err != nil {
		t.Fatalf("CreateBackup auto: %v", err)
	}
	if !strings.Contains(backup.Filename, "auto") {
		t.Errorf("auto backup filename should contain 'auto', got %q", backup.Filename)
	}
}

// ---------------------------------------------------------------------------
// ListBackups
// ---------------------------------------------------------------------------

func TestBackupService_ListBackups_Empty(t *testing.T) {
	svc := newTestBackupService(t)

	backups, err := svc.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups (empty): %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("ListBackups empty dir = %d, want 0", len(backups))
	}
}

func TestBackupService_ListBackups_AfterCreate(t *testing.T) {
	svc := newTestBackupService(t)

	if _, err := svc.CreateBackup("manual"); err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}

	backups, err := svc.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups: %v", err)
	}
	if len(backups) != 1 {
		t.Errorf("ListBackups = %d, want 1", len(backups))
	}
}

// ---------------------------------------------------------------------------
// RestoreBackup
// ---------------------------------------------------------------------------

func TestBackupService_RestoreBackup(t *testing.T) {
	svc := newTestBackupService(t)

	backup, err := svc.CreateBackup("manual")
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}

	if err := svc.RestoreBackup(backup.Filename); err != nil {
		t.Fatalf("RestoreBackup: %v", err)
	}
}

func TestBackupService_RestoreBackup_NotFound(t *testing.T) {
	svc := newTestBackupService(t)

	if err := svc.RestoreBackup("no-such-backup.tar.gz"); err == nil {
		t.Error("RestoreBackup(nonexistent) should return error")
	}
}

// ---------------------------------------------------------------------------
// Rotation
// ---------------------------------------------------------------------------

func TestBackupService_RotatesOldBackups(t *testing.T) {
	svc := newTestBackupService(t)

	// Create 6 backups (more than the 4 max)
	for i := 0; i < 6; i++ {
		if _, err := svc.CreateBackup("manual"); err != nil {
			t.Fatalf("CreateBackup[%d]: %v", i, err)
		}
	}

	backups, err := svc.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups: %v", err)
	}
	if len(backups) > 4 {
		t.Errorf("after rotation: %d backups remain, want <= 4", len(backups))
	}
}

// ---------------------------------------------------------------------------
// addFileToTar (private — tested directly in-package)
// ---------------------------------------------------------------------------

// TestAddFileToTar_BasicFile verifies that addFileToTar writes the correct
// header and content into the tar archive.
func TestAddFileToTar_BasicFile(t *testing.T) {
	content := []byte("hello from addFileToTar test")
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(srcFile, content, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	if err := addFileToTar(tw, srcFile, "archive/testfile.txt"); err != nil {
		t.Fatalf("addFileToTar: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar.Writer.Close: %v", err)
	}

	// Read back the tar and verify
	tr := tar.NewReader(&buf)
	hdr, err := tr.Next()
	if err != nil {
		t.Fatalf("tar.Reader.Next: %v", err)
	}
	if hdr.Name != "archive/testfile.txt" {
		t.Errorf("tar header Name = %q, want %q", hdr.Name, "archive/testfile.txt")
	}
	if hdr.Size != int64(len(content)) {
		t.Errorf("tar header Size = %d, want %d", hdr.Size, len(content))
	}

	var out bytes.Buffer
	if _, err := out.ReadFrom(tr); err != nil {
		t.Fatalf("reading tar entry: %v", err)
	}
	if !bytes.Equal(out.Bytes(), content) {
		t.Errorf("tar entry content = %q, want %q", out.Bytes(), content)
	}
}

// TestAddFileToTar_NonExistentFile verifies that a missing source file
// returns an error rather than silently producing an empty entry.
func TestAddFileToTar_NonExistentFile(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	err := addFileToTar(tw, "/no/such/file/exists.txt", "archive/missing.txt")
	if err == nil {
		t.Error("addFileToTar(nonexistent) should return error")
	}
}

// TestAddFileToTar_EmptyFile verifies zero-byte files are handled correctly.
func TestAddFileToTar_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(srcFile, []byte{}, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	if err := addFileToTar(tw, srcFile, "empty.txt"); err != nil {
		t.Fatalf("addFileToTar(empty): %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar.Writer.Close: %v", err)
	}

	tr := tar.NewReader(&buf)
	hdr, err := tr.Next()
	if err != nil {
		t.Fatalf("tar.Reader.Next: %v", err)
	}
	if hdr.Size != 0 {
		t.Errorf("empty file tar header Size = %d, want 0", hdr.Size)
	}
}

// TestAddFileToTar_LargerFile verifies multi-kilobyte content round-trips correctly.
func TestAddFileToTar_LargerFile(t *testing.T) {
	// 8 KB of data
	content := bytes.Repeat([]byte("abcdefgh"), 1024)
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "large.bin")
	if err := os.WriteFile(srcFile, content, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	if err := addFileToTar(tw, srcFile, "large.bin"); err != nil {
		t.Fatalf("addFileToTar(large): %v", err)
	}
	tw.Close()

	tr := tar.NewReader(&buf)
	if _, err := tr.Next(); err != nil {
		t.Fatalf("tar.Reader.Next: %v", err)
	}

	var out bytes.Buffer
	if _, err := out.ReadFrom(tr); err != nil {
		t.Fatalf("reading tar entry: %v", err)
	}
	if !bytes.Equal(out.Bytes(), content) {
		t.Error("large file content did not round-trip correctly through addFileToTar")
	}
}

// TestCreateBackup_ExercisesAddFileToTar confirms that CreateBackup exercises
// the addFileToTar code path by seeding real database files before backing up.
func TestCreateBackup_ExercisesAddFileToTar(t *testing.T) {
	svc := newTestBackupService(t)

	// Seed a real file in the db dir so addDatabaseToBackup has something to tar.
	dbFile := filepath.Join(svc.config.DataDir, "db", "server.db")
	if err := os.WriteFile(dbFile, []byte("SQLite stub"), 0o644); err != nil {
		t.Fatalf("WriteFile db seed: %v", err)
	}

	backup, err := svc.CreateBackup("manual")
	if err != nil {
		t.Fatalf("CreateBackup with seeded db: %v", err)
	}
	// The archive must be non-empty (it contains at least the db file).
	if backup.Size == 0 {
		t.Error("backup size should be > 0 when db file is present")
	}
}
