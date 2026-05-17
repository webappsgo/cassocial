package service

import (
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
