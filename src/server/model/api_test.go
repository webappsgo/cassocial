package model

import (
	"database/sql"
	"testing"
	"time"
)

func TestAPIKey_Validate_Valid(t *testing.T) {
	ak := &APIKey{Name: "My Key"}
	if err := ak.Validate(); err != nil {
		t.Errorf("valid APIKey Validate() = %v, want nil", err)
	}
}

func TestAPIKey_Validate_EmptyName(t *testing.T) {
	ak := &APIKey{}
	if err := ak.Validate(); err != ErrAPIKeyNameEmpty {
		t.Errorf("empty name Validate() = %v, want ErrAPIKeyNameEmpty", err)
	}
}

func TestAPIKey_IsExpired_NoExpiry(t *testing.T) {
	ak := &APIKey{}
	if ak.IsExpired() {
		t.Error("key with no expiry should not be expired")
	}
}

func TestAPIKey_IsExpired_Future(t *testing.T) {
	ak := &APIKey{ExpiresAt: sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true}}
	if ak.IsExpired() {
		t.Error("future expiry key should not be expired")
	}
}

func TestAPIKey_IsExpired_Past(t *testing.T) {
	ak := &APIKey{ExpiresAt: sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true}}
	if !ak.IsExpired() {
		t.Error("past expiry key should be expired")
	}
}

func TestAPIKey_UpdateLastUsed(t *testing.T) {
	ak := &APIKey{}
	ak.UpdateLastUsed()
	if !ak.LastUsedAt.Valid {
		t.Error("UpdateLastUsed() should set LastUsedAt")
	}
}

func TestAPIWebhook_Validate_Valid(t *testing.T) {
	wh := &APIWebhook{Name: "MyHook", URL: "https://example.com/webhook"}
	if err := wh.Validate(); err != nil {
		t.Errorf("valid webhook Validate() = %v, want nil", err)
	}
}

func TestAPIWebhook_Validate_EmptyName(t *testing.T) {
	wh := &APIWebhook{URL: "https://example.com/hook"}
	if err := wh.Validate(); err == nil {
		t.Error("empty name should return error")
	}
}

func TestAPIWebhook_Validate_EmptyURL(t *testing.T) {
	wh := &APIWebhook{Name: "Hook"}
	if err := wh.Validate(); err != ErrWebhookURLEmpty {
		t.Errorf("empty URL Validate() = %v, want ErrWebhookURLEmpty", err)
	}
}

func TestAPIWebhook_IsActive(t *testing.T) {
	wh := &APIWebhook{Active: true}
	if !wh.IsActive() {
		t.Error("IsActive() should return true")
	}
	wh.Active = false
	if wh.IsActive() {
		t.Error("IsActive() should return false")
	}
}

func TestAPIWebhook_IncrementFailureCount(t *testing.T) {
	wh := &APIWebhook{Active: true, FailureCount: 0}
	for i := 1; i <= 4; i++ {
		wh.IncrementFailureCount()
		if wh.FailureCount != i {
			t.Errorf("after %d increments, FailureCount = %d, want %d", i, wh.FailureCount, i)
		}
		if !wh.Active {
			t.Errorf("webhook should still be active after %d failures", i)
		}
	}
	// 5th failure should disable
	wh.IncrementFailureCount()
	if wh.Active {
		t.Error("webhook should be disabled after 5 failures")
	}
}

func TestAPIWebhook_ResetFailureCount(t *testing.T) {
	wh := &APIWebhook{FailureCount: 10}
	wh.ResetFailureCount()
	if wh.FailureCount != 0 {
		t.Errorf("ResetFailureCount() = %d, want 0", wh.FailureCount)
	}
}

func TestAPIWebhook_UpdateLastTriggered(t *testing.T) {
	wh := &APIWebhook{}
	wh.UpdateLastTriggered()
	if !wh.LastTriggeredAt.Valid {
		t.Error("UpdateLastTriggered() should set LastTriggeredAt")
	}
}

func TestAPIKey_HasScope(t *testing.T) {
	ak := &APIKey{Scopes: "profile:read, link:write"}
	if !ak.HasScope("profile:read") {
		t.Error("HasScope() should return true for a granted scope")
	}
	if !ak.HasScope("link:write") {
		t.Error("HasScope() should return true for a granted scope (whitespace-trimmed)")
	}
	if ak.HasScope("user:write") {
		t.Error("HasScope() should return false for a scope that was not granted")
	}

	empty := &APIKey{}
	if empty.HasScope("profile:read") {
		t.Error("HasScope() should return false when no scopes are granted")
	}

	wildcard := &APIKey{Scopes: "*"}
	if !wildcard.HasScope("anything") {
		t.Error("HasScope() should return true for any scope when wildcard is granted")
	}
}

func TestAPIWebhook_HasEvent(t *testing.T) {
	wh := &APIWebhook{Events: "profile.updated,link.clicked"}
	if !wh.HasEvent("profile.updated") {
		t.Error("HasEvent() should return true for a subscribed event")
	}
	if wh.HasEvent("user.created") {
		t.Error("HasEvent() should return false for an event that was not subscribed")
	}

	empty := &APIWebhook{}
	if empty.HasEvent("profile.updated") {
		t.Error("HasEvent() should return false when no events are subscribed")
	}

	wildcard := &APIWebhook{Events: "*"}
	if !wildcard.HasEvent("anything") {
		t.Error("HasEvent() should return true for any event when wildcard is subscribed")
	}
}

func TestAPIWebhook_Validate_InvalidURL(t *testing.T) {
	wh := &APIWebhook{Name: "Test", URL: "not-a-url"}
	if err := wh.Validate(); err != ErrInvalidURL {
		t.Errorf("invalid URL Validate() = %v, want ErrInvalidURL", err)
	}
}
