package server

import (
	"testing"
	"time"
)

func newTestSessionManager(t *testing.T) *SessionManager {
	t.Helper()
	return NewSessionManager(60) // 60 minute timeout
}

func TestNewSessionManager_Default(t *testing.T) {
	sm := NewSessionManager(0)
	if sm == nil {
		t.Fatal("NewSessionManager returned nil")
	}
	// Default is 1440 minutes
	if sm.timeout != 1440*time.Minute {
		t.Errorf("timeout = %v, want %v", sm.timeout, 1440*time.Minute)
	}
}

func TestNewSessionManager_Custom(t *testing.T) {
	sm := NewSessionManager(30)
	if sm.timeout != 30*time.Minute {
		t.Errorf("timeout = %v, want 30m", sm.timeout)
	}
}

func TestCreateSession(t *testing.T) {
	sm := newTestSessionManager(t)
	session, err := sm.CreateSession("user-1", "alice", "user", "127.0.0.1", "TestBrowser/1.0")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if session == nil {
		t.Fatal("CreateSession returned nil session")
	}
	if session.ID == "" {
		t.Error("session ID is empty")
	}
	if session.UserID != "user-1" {
		t.Errorf("UserID = %q, want user-1", session.UserID)
	}
	if session.Username != "alice" {
		t.Errorf("Username = %q, want alice", session.Username)
	}
	if session.Role != "user" {
		t.Errorf("Role = %q, want user", session.Role)
	}
}

func TestGetSession_Found(t *testing.T) {
	sm := newTestSessionManager(t)
	created, _ := sm.CreateSession("user-1", "alice", "user", "127.0.0.1", "")

	got, err := sm.GetSession(created.ID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}
}

func TestGetSession_NotFound(t *testing.T) {
	sm := newTestSessionManager(t)
	_, err := sm.GetSession("nonexistent-session-id")
	if err == nil {
		t.Error("GetSession should return error for nonexistent session")
	}
}

func TestGetSession_Expired(t *testing.T) {
	sm := NewSessionManager(0)
	created, _ := sm.CreateSession("user-1", "alice", "user", "127.0.0.1", "")

	// Manually expire the session
	sm.mu.Lock()
	sm.sessions[created.ID].ExpiresAt = time.Now().Add(-1 * time.Hour)
	sm.mu.Unlock()

	_, err := sm.GetSession(created.ID)
	if err == nil {
		t.Error("GetSession should return error for expired session")
	}
}

func TestValidateSession_Valid(t *testing.T) {
	sm := newTestSessionManager(t)
	created, _ := sm.CreateSession("user-1", "alice", "user", "127.0.0.1", "")

	session, ok := sm.ValidateSession(created.ID)
	if !ok {
		t.Error("ValidateSession returned false for valid session")
	}
	if session == nil {
		t.Error("ValidateSession returned nil session")
	}
}

func TestValidateSession_Invalid(t *testing.T) {
	sm := newTestSessionManager(t)
	_, ok := sm.ValidateSession("bad-session-id")
	if ok {
		t.Error("ValidateSession returned true for invalid session")
	}
}

func TestDestroySession(t *testing.T) {
	sm := newTestSessionManager(t)
	created, _ := sm.CreateSession("user-1", "alice", "user", "127.0.0.1", "")

	sm.DestroySession(created.ID)

	_, err := sm.GetSession(created.ID)
	if err == nil {
		t.Error("session should be gone after DestroySession")
	}
}

func TestDestroySession_NonExistent(t *testing.T) {
	sm := newTestSessionManager(t)
	// Should not panic
	sm.DestroySession("does-not-exist")
}

func TestDestroyUserSessions(t *testing.T) {
	sm := newTestSessionManager(t)
	s1, _ := sm.CreateSession("user-1", "alice", "user", "127.0.0.1", "browser1")
	s2, _ := sm.CreateSession("user-1", "alice", "user", "127.0.0.2", "browser2")
	s3, _ := sm.CreateSession("user-2", "bob", "user", "127.0.0.1", "browser1")

	sm.DestroyUserSessions("user-1")

	if _, err := sm.GetSession(s1.ID); err == nil {
		t.Error("s1 should be destroyed")
	}
	if _, err := sm.GetSession(s2.ID); err == nil {
		t.Error("s2 should be destroyed")
	}
	// user-2's session should survive
	if _, err := sm.GetSession(s3.ID); err != nil {
		t.Errorf("s3 (user-2) should still exist: %v", err)
	}
}

func TestGetUserSessions(t *testing.T) {
	sm := newTestSessionManager(t)
	sm.CreateSession("user-1", "alice", "user", "127.0.0.1", "b1")
	sm.CreateSession("user-1", "alice", "user", "127.0.0.2", "b2")
	sm.CreateSession("user-2", "bob", "user", "127.0.0.1", "b1")

	sessions := sm.GetUserSessions("user-1")
	if len(sessions) != 2 {
		t.Errorf("GetUserSessions(user-1) = %d, want 2", len(sessions))
	}
}

func TestGetUserSessions_Empty(t *testing.T) {
	sm := newTestSessionManager(t)
	sessions := sm.GetUserSessions("no-such-user")
	if len(sessions) != 0 {
		t.Errorf("GetUserSessions for unknown user = %d, want 0", len(sessions))
	}
}

func TestUpdateSessionActivity(t *testing.T) {
	sm := newTestSessionManager(t)
	created, _ := sm.CreateSession("user-1", "alice", "user", "127.0.0.1", "")

	before := created.LastActivity

	time.Sleep(2 * time.Millisecond)
	sm.UpdateSessionActivity(created.ID)

	sm.mu.RLock()
	updated := sm.sessions[created.ID]
	sm.mu.RUnlock()

	if !updated.LastActivity.After(before) {
		t.Error("LastActivity was not updated")
	}
}

func TestUpdateSessionActivity_NonExistent(t *testing.T) {
	sm := newTestSessionManager(t)
	// Should not panic
	sm.UpdateSessionActivity("does-not-exist")
}

func TestGetActiveSessions(t *testing.T) {
	sm := newTestSessionManager(t)
	sm.CreateSession("u1", "a", "user", "127.0.0.1", "")
	sm.CreateSession("u2", "b", "user", "127.0.0.1", "")

	count := sm.GetActiveSessions()
	if count != 2 {
		t.Errorf("GetActiveSessions = %d, want 2", count)
	}
}

func TestGetActiveUsers(t *testing.T) {
	sm := newTestSessionManager(t)
	sm.CreateSession("u1", "alice", "user", "127.0.0.1", "b1")
	sm.CreateSession("u1", "alice", "user", "127.0.0.2", "b2") // same user, two sessions
	sm.CreateSession("u2", "bob", "user", "127.0.0.1", "b1")

	users := sm.GetActiveUsers()
	if users != 2 {
		t.Errorf("GetActiveUsers = %d, want 2", users)
	}
}

func TestCleanupExpired(t *testing.T) {
	sm := newTestSessionManager(t)
	created, _ := sm.CreateSession("u1", "alice", "user", "127.0.0.1", "")

	// Manually expire
	sm.mu.Lock()
	sm.sessions[created.ID].ExpiresAt = time.Now().Add(-1 * time.Hour)
	sm.mu.Unlock()

	sm.cleanupExpired()

	if sm.GetActiveSessions() != 0 {
		t.Error("expired session should have been cleaned up")
	}
}

func TestExtendSession(t *testing.T) {
	sm := newTestSessionManager(t)
	created, _ := sm.CreateSession("u1", "alice", "user", "127.0.0.1", "")
	original := created.ExpiresAt

	time.Sleep(2 * time.Millisecond)
	if err := sm.ExtendSession(created.ID, 2*time.Hour); err != nil {
		t.Fatalf("ExtendSession: %v", err)
	}

	sm.mu.RLock()
	extended := sm.sessions[created.ID]
	sm.mu.RUnlock()

	if !extended.ExpiresAt.After(original) {
		t.Error("ExpiresAt was not extended")
	}
}

func TestExtendSession_NotFound(t *testing.T) {
	sm := newTestSessionManager(t)
	err := sm.ExtendSession("nonexistent", time.Hour)
	if err == nil {
		t.Error("ExtendSession should return error for nonexistent session")
	}
}

func TestSetGetSessionTimeout(t *testing.T) {
	sm := newTestSessionManager(t)
	sm.SetSessionTimeout(90)
	timeout := sm.GetSessionTimeout()
	if timeout != 90*time.Minute {
		t.Errorf("GetSessionTimeout = %v, want 90m", timeout)
	}
}

func TestGenerateSessionID_Unique(t *testing.T) {
	id1, err := generateSessionID()
	if err != nil {
		t.Fatalf("generateSessionID: %v", err)
	}
	id2, err := generateSessionID()
	if err != nil {
		t.Fatalf("generateSessionID: %v", err)
	}
	if id1 == id2 {
		t.Error("generateSessionID returned identical IDs")
	}
}

func TestGenerateSessionID_NotEmpty(t *testing.T) {
	id, err := generateSessionID()
	if err != nil {
		t.Fatalf("generateSessionID: %v", err)
	}
	if id == "" {
		t.Error("generateSessionID returned empty string")
	}
}

func TestGetSessionInfo(t *testing.T) {
	sm := newTestSessionManager(t)
	session, _ := sm.CreateSession("u1", "alice", "user", "192.168.1.1", "Firefox")

	info := session.GetSessionInfo(session.ID)
	if info == nil {
		t.Fatal("GetSessionInfo returned nil")
	}
	if !info.IsCurrent {
		t.Error("IsCurrent should be true for own session")
	}
	if info.IPAddress != "192.168.1.1" {
		t.Errorf("IPAddress = %q, want 192.168.1.1", info.IPAddress)
	}
}

func TestGetSessionInfo_OtherSession(t *testing.T) {
	sm := newTestSessionManager(t)
	session, _ := sm.CreateSession("u1", "alice", "user", "127.0.0.1", "")

	info := session.GetSessionInfo("different-session-id")
	if info.IsCurrent {
		t.Error("IsCurrent should be false for a different session")
	}
}

func TestGetUserSessionsInfo(t *testing.T) {
	sm := newTestSessionManager(t)
	s1, _ := sm.CreateSession("u1", "alice", "user", "127.0.0.1", "")
	sm.CreateSession("u1", "alice", "user", "127.0.0.2", "")

	infos := sm.GetUserSessionsInfo("u1", s1.ID)
	if len(infos) != 2 {
		t.Errorf("GetUserSessionsInfo = %d entries, want 2", len(infos))
	}

	// Exactly one should be IsCurrent
	currentCount := 0
	for _, info := range infos {
		if info.IsCurrent {
			currentCount++
		}
	}
	if currentCount != 1 {
		t.Errorf("currentCount = %d, want 1", currentCount)
	}
}
