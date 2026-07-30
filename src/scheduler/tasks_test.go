package scheduler

import (
	"fmt"
	"testing"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// newTestDB creates an in-memory SQLite DB for testing.
func newTestDB(t *testing.T) *store.DB {
	t.Helper()
	tmp := t.TempDir()
	db, err := store.Connect("sqlite", tmp+"/test.db")
	if err != nil {
		t.Fatalf("store.Connect() failed: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// newTestConfig creates a minimal config suitable for scheduler task tests.
func newTestConfig() *config.Config {
	cfg := &config.Config{}
	cfg.SSL.Enabled = false
	cfg.SSL.LetsEncrypt = false
	return cfg
}

// ---- NewTasks ----

func TestNewTasks_ReturnsNonNil(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()

	tasks := NewTasks(cfg, db)
	if tasks == nil {
		t.Fatal("NewTasks() returned nil")
	}
	if tasks.config != cfg {
		t.Error("NewTasks() did not store config reference")
	}
	if tasks.db != db {
		t.Error("NewTasks() did not store db reference")
	}
}

// ---- RegisterAllTasks ----

func TestRegisterAllTasks_RegistersSevenTasks(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)
	s := New()

	if err := tasks.RegisterAllTasks(s); err != nil {
		t.Fatalf("RegisterAllTasks() returned error: %v", err)
	}

	listed := s.ListTasks()
	if len(listed) != 7 {
		t.Errorf("RegisterAllTasks() registered %d tasks, want 7", len(listed))
	}
}

func TestRegisterAllTasks_TaskNamesPresent(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)
	s := New()

	if err := tasks.RegisterAllTasks(s); err != nil {
		t.Fatalf("RegisterAllTasks() returned error: %v", err)
	}

	wantNames := []string{
		"analytics_aggregation",
		"cert_renewal_check",
		"database_cleanup",
		"automated_backup",
		"email_queue",
		"session_cleanup",
		"geoip_update",
	}
	for _, name := range wantNames {
		if _, err := s.GetTaskStatus(name); err != nil {
			t.Errorf("task %q not registered: %v", name, err)
		}
	}
}

// ---- AggregateAnalytics ----

func TestAggregateAnalytics_NoError(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)

	if err := tasks.AggregateAnalytics(); err != nil {
		t.Errorf("AggregateAnalytics() returned error: %v", err)
	}
}

// ---- CheckCertificateRenewal ----

func TestCheckCertificateRenewal_SSLDisabled(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	cfg.SSL.Enabled = false
	tasks := NewTasks(cfg, db)

	if err := tasks.CheckCertificateRenewal(); err != nil {
		t.Errorf("CheckCertificateRenewal(disabled) returned error: %v", err)
	}
}

func TestCheckCertificateRenewal_SSLEnabled(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	cfg.SSL.Enabled = true
	tasks := NewTasks(cfg, db)

	// SSL enabled but no cert files configured — function should still return nil
	if err := tasks.CheckCertificateRenewal(); err != nil {
		t.Errorf("CheckCertificateRenewal(enabled) returned error: %v", err)
	}
}

// ---- CleanupDatabase ----

func TestCleanupDatabase_WithMemoryDB(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)

	// Should not panic; may log an error if table doesn't exist yet.
	_ = tasks.CleanupDatabase()
}

// ---- CreateBackup ----

func TestCreateBackup_NoError(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)

	if err := tasks.CreateBackup(); err != nil {
		t.Errorf("CreateBackup() returned error: %v", err)
	}
}

// ---- ProcessEmailQueue ----

func TestProcessEmailQueue_NoError(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)

	if err := tasks.ProcessEmailQueue(); err != nil {
		t.Errorf("ProcessEmailQueue() returned error: %v", err)
	}
}

// ---- CleanupSessions ----

func TestCleanupSessions_WithMemoryDB(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)

	// Should not panic; may return an error if table isn't migrated yet.
	_ = tasks.CleanupSessions()
}

func TestCleanupSessions_Success(t *testing.T) {
	// Create a DB and manually create the sessions table so CleanupSessions
	// can succeed (return nil path).
	tmp := t.TempDir()
	db, err := store.Connect("sqlite", tmp+"/sessions.db")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	defer db.Close()

	// Create the sessions table that CleanupExpiredSessions expects.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL
	)`); err != nil {
		t.Fatalf("create sessions table: %v", err)
	}

	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)

	if taskErr := tasks.CleanupSessions(); taskErr != nil {
		t.Errorf("CleanupSessions() with sessions table should return nil, got: %v", taskErr)
	}
}

// ---- UpdateGeoIPDatabase ----

func TestUpdateGeoIPDatabase_NoError(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)

	if err := tasks.UpdateGeoIPDatabase(); err != nil {
		t.Errorf("UpdateGeoIPDatabase() returned error: %v", err)
	}
}

// ---- RegisterAllTasks — error propagation ----

// errAfterN is a TaskRegistrar that succeeds for the first n calls then fails.
type errAfterN struct {
	n       int
	calls   int
	realSch *Scheduler
}

func (e *errAfterN) RegisterTask(name, schedule string, handler func() error) error {
	e.calls++
	if e.calls > e.n {
		return fmt.Errorf("injected error on call %d", e.calls)
	}
	return e.realSch.RegisterTask(name, schedule, handler)
}

// TestRegisterAllTasks_ErrorOnFirstTask verifies error propagation when the
// first RegisterTask call fails.
func TestRegisterAllTasks_ErrorOnFirstTask(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)

	mock := &errAfterN{n: 0, realSch: New()}
	err := tasks.RegisterAllTasks(mock)
	if err == nil {
		t.Error("RegisterAllTasks should return error when first RegisterTask fails")
	}
}

// TestRegisterAllTasks_ErrorOnSecondTask verifies error propagation on the
// second RegisterTask call.
func TestRegisterAllTasks_ErrorOnSecondTask(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)

	mock := &errAfterN{n: 1, realSch: New()}
	err := tasks.RegisterAllTasks(mock)
	if err == nil {
		t.Error("RegisterAllTasks should return error when second RegisterTask fails")
	}
}

// TestRegisterAllTasks_ErrorOnThirdTask verifies error propagation on call 3.
func TestRegisterAllTasks_ErrorOnThirdTask(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)

	mock := &errAfterN{n: 2, realSch: New()}
	err := tasks.RegisterAllTasks(mock)
	if err == nil {
		t.Error("RegisterAllTasks should return error when third RegisterTask fails")
	}
}

// TestRegisterAllTasks_ErrorOnFourthTask verifies error propagation on call 4.
func TestRegisterAllTasks_ErrorOnFourthTask(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)

	mock := &errAfterN{n: 3, realSch: New()}
	err := tasks.RegisterAllTasks(mock)
	if err == nil {
		t.Error("RegisterAllTasks should return error when fourth RegisterTask fails")
	}
}

// TestRegisterAllTasks_ErrorOnFifthTask verifies error propagation on call 5.
func TestRegisterAllTasks_ErrorOnFifthTask(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)

	mock := &errAfterN{n: 4, realSch: New()}
	err := tasks.RegisterAllTasks(mock)
	if err == nil {
		t.Error("RegisterAllTasks should return error when fifth RegisterTask fails")
	}
}

// TestRegisterAllTasks_ErrorOnSixthTask verifies error propagation on call 6.
func TestRegisterAllTasks_ErrorOnSixthTask(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)

	mock := &errAfterN{n: 5, realSch: New()}
	err := tasks.RegisterAllTasks(mock)
	if err == nil {
		t.Error("RegisterAllTasks should return error when sixth RegisterTask fails")
	}
}

// TestRegisterAllTasks_ErrorOnSeventhTask verifies error propagation on call 7.
func TestRegisterAllTasks_ErrorOnSeventhTask(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)

	mock := &errAfterN{n: 6, realSch: New()}
	err := tasks.RegisterAllTasks(mock)
	if err == nil {
		t.Error("RegisterAllTasks should return error when seventh RegisterTask fails")
	}
}

// ---- CleanupSessions — error path ----

func TestCleanupSessions_ErrorPath(t *testing.T) {
	// Use a fresh DB without running migrations so the sessions table does not
	// exist, causing CleanupExpiredSessions to return a "no such table" error.
	tmp := t.TempDir()
	db, err := store.Connect("sqlite", tmp+"/noschema.db")
	if err != nil {
		t.Fatalf("store.Connect() failed: %v", err)
	}
	defer db.Close()
	// Deliberately do NOT call db.RunMigrations() — sessions table absent.

	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)

	taskErr := tasks.CleanupSessions()
	if taskErr == nil {
		t.Error("CleanupSessions() with missing sessions table should return an error")
	}
}

// ---- CleanupDatabase — logs error but returns nil ----

func TestCleanupDatabase_ClosedDB_ReturnsNil(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)
	db.Close()

	// CleanupDatabase logs errors but always returns nil.
	err := tasks.CleanupDatabase()
	if err != nil {
		t.Errorf("CleanupDatabase() with closed DB should return nil (errors are only logged), got: %v", err)
	}
}

// ---- GetTaskStatistics ----

func TestGetTaskStatistics_EmptyScheduler(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)
	s := New()

	stats := tasks.GetTaskStatistics(s)
	if stats == nil {
		t.Fatal("GetTaskStatistics() returned nil")
	}
	total, ok := stats["total_tasks"]
	if !ok {
		t.Fatal("GetTaskStatistics() missing 'total_tasks' key")
	}
	if total != 0 {
		t.Errorf("total_tasks = %v, want 0 for empty scheduler", total)
	}
}

func TestGetTaskStatistics_WithTasks(t *testing.T) {
	db := newTestDB(t)
	cfg := newTestConfig()
	tasks := NewTasks(cfg, db)
	s := New()

	if err := tasks.RegisterAllTasks(s); err != nil {
		t.Fatalf("RegisterAllTasks() failed: %v", err)
	}

	stats := tasks.GetTaskStatistics(s)
	if stats == nil {
		t.Fatal("GetTaskStatistics() returned nil")
	}

	total, ok := stats["total_tasks"]
	if !ok {
		t.Fatal("GetTaskStatistics() missing 'total_tasks' key")
	}
	if total.(int) != 7 {
		t.Errorf("total_tasks = %v, want 7", total)
	}

	taskList, ok := stats["tasks"]
	if !ok {
		t.Fatal("GetTaskStatistics() missing 'tasks' key")
	}
	taskSlice, ok := taskList.([]map[string]interface{})
	if !ok {
		t.Fatal("GetTaskStatistics() 'tasks' value is not []map[string]interface{}")
	}
	if len(taskSlice) != 7 {
		t.Errorf("tasks slice length = %d, want 7", len(taskSlice))
	}

	// Verify required fields are present in each task stat
	for i, task := range taskSlice {
		for _, field := range []string{"name", "schedule", "enabled", "last_run", "next_run", "run_count", "error_count", "last_error"} {
			if _, ok := task[field]; !ok {
				t.Errorf("task[%d] missing field %q", i, field)
			}
		}
	}
}
