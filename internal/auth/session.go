package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// Session represents a user session
type Session struct {
	ID        string
	UserID    string
	Username  string
	Role      string
	CreatedAt time.Time
	ExpiresAt time.Time
	IPAddress string
	UserAgent string
	LastActivity time.Time
}

// SessionManager manages user sessions
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	timeout  time.Duration
	cleanup  time.Duration
}

// NewSessionManager creates a new SessionManager
func NewSessionManager(timeoutMinutes int) *SessionManager {
	if timeoutMinutes == 0 {
		timeoutMinutes = 1440 // Default 24 hours
	}

	sm := &SessionManager{
		sessions: make(map[string]*Session),
		timeout:  time.Duration(timeoutMinutes) * time.Minute,
		cleanup:  time.Minute * 15, // Cleanup every 15 minutes
	}

	// Start cleanup goroutine
	go sm.cleanupLoop()

	return sm
}

// CreateSession creates a new session for a user
func (sm *SessionManager) CreateSession(userID, username, role, ipAddress, userAgent string) (*Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	now := time.Now()
	session := &Session{
		ID:           sessionID,
		UserID:       userID,
		Username:     username,
		Role:         role,
		CreatedAt:    now,
		ExpiresAt:    now.Add(sm.timeout),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		LastActivity: now,
	}

	sm.mu.Lock()
	sm.sessions[sessionID] = session
	sm.mu.Unlock()

	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	sm.mu.RLock()
	session, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		sm.DestroySession(sessionID)
		return nil, fmt.Errorf("session expired")
	}

	// Update last activity
	sm.mu.Lock()
	session.LastActivity = time.Now()
	// Optionally extend expiration on activity (sliding expiration)
	session.ExpiresAt = time.Now().Add(sm.timeout)
	sm.mu.Unlock()

	return session, nil
}

// ValidateSession validates a session and returns the session if valid
func (sm *SessionManager) ValidateSession(sessionID string) (*Session, bool) {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return nil, false
	}
	return session, true
}

// DestroySession removes a session
func (sm *SessionManager) DestroySession(sessionID string) {
	sm.mu.Lock()
	delete(sm.sessions, sessionID)
	sm.mu.Unlock()
}

// DestroyUserSessions removes all sessions for a user
func (sm *SessionManager) DestroyUserSessions(userID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for id, session := range sm.sessions {
		if session.UserID == userID {
			delete(sm.sessions, id)
		}
	}
}

// GetUserSessions returns all active sessions for a user
func (sm *SessionManager) GetUserSessions(userID string) []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sessions []*Session
	now := time.Now()

	for _, session := range sm.sessions {
		if session.UserID == userID && now.Before(session.ExpiresAt) {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// UpdateSessionActivity updates the last activity time for a session
func (sm *SessionManager) UpdateSessionActivity(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		session.LastActivity = time.Now()
		// Sliding expiration - extend session on activity
		session.ExpiresAt = time.Now().Add(sm.timeout)
	}
}

// GetActiveSessions returns the number of active sessions
func (sm *SessionManager) GetActiveSessions() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	count := 0
	now := time.Now()

	for _, session := range sm.sessions {
		if now.Before(session.ExpiresAt) {
			count++
		}
	}

	return count
}

// GetActiveUsers returns the number of unique active users
func (sm *SessionManager) GetActiveUsers() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	users := make(map[string]bool)
	now := time.Now()

	for _, session := range sm.sessions {
		if now.Before(session.ExpiresAt) {
			users[session.UserID] = true
		}
	}

	return len(users)
}

// cleanupLoop periodically removes expired sessions
func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(sm.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		sm.cleanupExpired()
	}
}

// cleanupExpired removes all expired sessions
func (sm *SessionManager) cleanupExpired() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for id, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			delete(sm.sessions, id)
		}
	}
}

// ExtendSession extends the expiration time of a session
func (sm *SessionManager) ExtendSession(sessionID string, duration time.Duration) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	session.ExpiresAt = time.Now().Add(duration)
	return nil
}

// SetSessionTimeout updates the default session timeout
func (sm *SessionManager) SetSessionTimeout(minutes int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.timeout = time.Duration(minutes) * time.Minute
}

// GetSessionTimeout returns the current session timeout
func (sm *SessionManager) GetSessionTimeout() time.Duration {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.timeout
}

// generateSessionID generates a cryptographically secure session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// SessionInfo returns public session information (without sensitive data)
type SessionInfo struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	LastActivity time.Time `json:"last_activity"`
	IsCurrent    bool      `json:"is_current"`
}

// GetSessionInfo returns sanitized session information
func (s *Session) GetSessionInfo(currentSessionID string) *SessionInfo {
	return &SessionInfo{
		ID:           s.ID,
		CreatedAt:    s.CreatedAt,
		ExpiresAt:    s.ExpiresAt,
		IPAddress:    s.IPAddress,
		UserAgent:    s.UserAgent,
		LastActivity: s.LastActivity,
		IsCurrent:    s.ID == currentSessionID,
	}
}

// GetUserSessionsInfo returns sanitized session information for all user sessions
func (sm *SessionManager) GetUserSessionsInfo(userID, currentSessionID string) []*SessionInfo {
	sessions := sm.GetUserSessions(userID)
	var infos []*SessionInfo

	for _, session := range sessions {
		infos = append(infos, session.GetSessionInfo(currentSessionID))
	}

	return infos
}
