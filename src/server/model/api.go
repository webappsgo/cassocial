package model

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

// APIKey represents an API key for authentication
type APIKey struct {
	ID          string       `json:"id" db:"id"`
	UserID      string       `json:"user_id" db:"user_id"`
	Name        string       `json:"name" db:"name"`
	KeyHash     string       `json:"-" db:"key_hash"`
	LastUsedAt  sql.NullTime `json:"last_used_at,omitempty" db:"last_used_at"`
	ExpiresAt   sql.NullTime `json:"expires_at,omitempty" db:"expires_at"`
	Scopes      string       `json:"scopes" db:"scopes"` // Array stored as string
	RateLimit   int          `json:"rate_limit" db:"rate_limit"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
}

// APIWebhook represents a webhook configuration
type APIWebhook struct {
	ID              string       `json:"id" db:"id"`
	UserID          string       `json:"user_id" db:"user_id"`
	Name            string       `json:"name" db:"name"`
	URL             string       `json:"url" db:"url"`
	Secret          string       `json:"-" db:"secret"`
	Events          string       `json:"events" db:"events"` // Array stored as string
	Active          bool         `json:"active" db:"active"`
	FailureCount    int          `json:"failure_count" db:"failure_count"`
	LastTriggeredAt sql.NullTime `json:"last_triggered_at,omitempty" db:"last_triggered_at"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
}

// Valid API scopes
const (
	ScopeProfileRead   = "profile:read"
	ScopeProfileWrite  = "profile:write"
	ScopeLinkRead      = "link:read"
	ScopeLinkWrite     = "link:write"
	ScopeAnalyticsRead = "analytics:read"
	ScopeUserRead      = "user:read"
	ScopeUserWrite     = "user:write"
)

// Valid webhook events
const (
	WebhookEventProfileCreated = "profile.created"
	WebhookEventProfileUpdated = "profile.updated"
	WebhookEventProfileDeleted = "profile.deleted"
	WebhookEventLinkCreated    = "link.created"
	WebhookEventLinkUpdated    = "link.updated"
	WebhookEventLinkDeleted    = "link.deleted"
	WebhookEventLinkClicked    = "link.clicked"
)

var (
	ErrAPIKeyNameEmpty = errors.New("API key name cannot be empty")
	ErrAPIKeyExpired   = errors.New("API key has expired")
	ErrWebhookURLEmpty = errors.New("webhook URL cannot be empty")
	ErrWebhookInactive = errors.New("webhook is inactive")
)

// Validate validates the API key model
func (ak *APIKey) Validate() error {
	if ak.Name == "" {
		return ErrAPIKeyNameEmpty
	}
	return nil
}

// IsExpired checks if the API key has expired
func (ak *APIKey) IsExpired() bool {
	if !ak.ExpiresAt.Valid {
		return false
	}
	return time.Now().After(ak.ExpiresAt.Time)
}

// UpdateLastUsed updates the last used timestamp
func (ak *APIKey) UpdateLastUsed() {
	ak.LastUsedAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
}

// HasScope checks if the API key grants a specific scope. Scopes are stored as
// a comma-separated list; the "*" wildcard grants every scope (global tokens).
func (ak *APIKey) HasScope(scope string) bool {
	return listContains(ak.Scopes, scope)
}

// Validate validates the webhook model
func (wh *APIWebhook) Validate() error {
	if wh.Name == "" {
		return errors.New("webhook name cannot be empty")
	}
	if wh.URL == "" {
		return ErrWebhookURLEmpty
	}
	if !isValidURL(wh.URL) {
		return ErrInvalidURL
	}
	return nil
}

// IsActive checks if the webhook is active
func (wh *APIWebhook) IsActive() bool {
	return wh.Active
}

// IncrementFailureCount increments the failure counter
func (wh *APIWebhook) IncrementFailureCount() {
	wh.FailureCount++
	// Disable webhook after 5 consecutive failures
	if wh.FailureCount >= 5 {
		wh.Active = false
	}
}

// ResetFailureCount resets the failure counter
func (wh *APIWebhook) ResetFailureCount() {
	wh.FailureCount = 0
}

// UpdateLastTriggered updates the last triggered timestamp
func (wh *APIWebhook) UpdateLastTriggered() {
	wh.LastTriggeredAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
}

// HasEvent checks if the webhook is subscribed to a specific event. Events are
// stored as a comma-separated list; the "*" wildcard subscribes to every event.
func (wh *APIWebhook) HasEvent(event string) bool {
	return listContains(wh.Events, event)
}

// listContains reports whether a comma-separated list grants the given item.
// Entries are trimmed of surrounding whitespace and "*" matches anything.
func listContains(list, item string) bool {
	for _, entry := range strings.Split(list, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "*" || entry == item {
			return true
		}
	}
	return false
}
