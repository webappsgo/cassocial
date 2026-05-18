package server

import (
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)
	if rl == nil {
		t.Fatal("NewRateLimiter() returned nil")
	}
}

func TestRateLimiter_Allow_WithinLimit(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	for i := 0; i < 5; i++ {
		if !rl.Allow("client1") {
			t.Errorf("request %d should be allowed (within limit)", i+1)
		}
	}
}

func TestRateLimiter_Allow_ExceedsLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		rl.Allow("client2")
	}
	if rl.Allow("client2") {
		t.Error("4th request should be blocked (limit=3)")
	}
}

func TestRateLimiter_Allow_SeparateClients(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	// Fill client3 to the limit
	rl.Allow("client3")
	rl.Allow("client3")
	// client4 should still be allowed
	if !rl.Allow("client4") {
		t.Error("client4 first request should be allowed")
	}
}

func TestRateLimiter_Allow_WindowReset(t *testing.T) {
	rl := NewRateLimiter(1, 50*time.Millisecond)
	if !rl.Allow("reset-client") {
		t.Fatal("first request should be allowed")
	}
	if rl.Allow("reset-client") {
		t.Fatal("second request should be blocked")
	}

	// Wait for window to expire
	time.Sleep(100 * time.Millisecond)

	if !rl.Allow("reset-client") {
		t.Error("request after window reset should be allowed")
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	rl.Allow("reset-test")
	rl.Allow("reset-test")
	if rl.Allow("reset-test") {
		t.Fatal("should be blocked before reset")
	}

	rl.Reset("reset-test")

	if !rl.Allow("reset-test") {
		t.Error("should be allowed after reset")
	}
}

func TestRateLimiter_GetRemaining_Initial(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)
	remaining := rl.GetRemaining("new-client")
	if remaining != 10 {
		t.Errorf("GetRemaining for new client = %d, want 10", remaining)
	}
}

func TestRateLimiter_GetRemaining_AfterRequests(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)
	rl.Allow("rem-client")
	rl.Allow("rem-client")
	rl.Allow("rem-client")

	remaining := rl.GetRemaining("rem-client")
	if remaining != 7 {
		t.Errorf("GetRemaining after 3 requests = %d, want 7", remaining)
	}
}

func TestRateLimiter_GetRemaining_AtZero(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	rl.Allow("zero-client")
	rl.Allow("zero-client")
	rl.Allow("zero-client") // blocked but still increments or stays at limit

	remaining := rl.GetRemaining("zero-client")
	if remaining != 0 {
		t.Errorf("GetRemaining when exhausted = %d, want 0", remaining)
	}
}

func TestRateLimiter_GetResetTime(t *testing.T) {
	window := time.Minute
	rl := NewRateLimiter(5, window)

	before := time.Now()
	rl.Allow("time-client")
	after := time.Now()

	resetTime := rl.GetResetTime("time-client")
	expectedMin := before.Add(window)
	expectedMax := after.Add(window)

	if resetTime.Before(expectedMin) || resetTime.After(expectedMax) {
		t.Errorf("GetResetTime = %v, expected in range [%v, %v]", resetTime, expectedMin, expectedMax)
	}
}

func TestRateLimiter_GetResetTime_UnknownClient(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	before := time.Now()
	resetTime := rl.GetResetTime("unknown")
	after := time.Now()

	// Should return a future time
	if resetTime.Before(before) || resetTime.After(after.Add(time.Minute+time.Second)) {
		t.Errorf("GetResetTime for unknown client = %v, expected ~1 minute from now", resetTime)
	}
}

func TestRateLimiter_GetRemaining_WindowExpired(t *testing.T) {
	rl := NewRateLimiter(5, 50*time.Millisecond)
	rl.Allow("expire-client")
	rl.Allow("expire-client")

	// Wait for window to expire
	time.Sleep(100 * time.Millisecond)

	remaining := rl.GetRemaining("expire-client")
	if remaining != 5 {
		t.Errorf("GetRemaining after window expiry = %d, want 5 (full limit)", remaining)
	}
}

func TestRateLimiter_CleanupLoop_Fires(t *testing.T) {
	rl := &RateLimiter{
		requests:        make(map[string]*rateLimitEntry),
		limit:           10,
		window:          time.Minute,
		cleanupInterval: 5 * time.Millisecond,
	}

	// Add a stale entry (last accessed over 2x cleanupInterval ago)
	staleAccess := time.Now().Add(-1 * time.Hour)
	rl.requests["stale-loop"] = &rateLimitEntry{
		count:      1,
		resetTime:  staleAccess,
		lastAccess: staleAccess,
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	// Wait for cleanup to fire
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		rl.mu.Lock()
		_, exists := rl.requests["stale-loop"]
		rl.mu.Unlock()
		if !exists {
			return // cleaned up — success
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Error("cleanupLoop did not remove stale entry within 200ms")
}

func TestRateLimiter_GetRemaining_NegativeCount(t *testing.T) {
	rl := &RateLimiter{
		requests:        make(map[string]*rateLimitEntry),
		limit:           2,
		window:          time.Minute,
		cleanupInterval: time.Minute,
	}

	// Directly inject an entry where count exceeds limit to exercise the remaining < 0 branch.
	rl.requests["overflow-client"] = &rateLimitEntry{
		count:      10, // exceeds limit of 2
		resetTime:  time.Now().Add(time.Minute),
		lastAccess: time.Now(),
	}

	remaining := rl.GetRemaining("overflow-client")
	if remaining != 0 {
		t.Errorf("GetRemaining when count > limit = %d, want 0", remaining)
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := &RateLimiter{
		requests:        make(map[string]*rateLimitEntry),
		limit:           10,
		window:          time.Minute,
		cleanupInterval: 10 * time.Millisecond,
	}

	// Add a stale entry
	staleTime := time.Now().Add(-1 * time.Hour)
	rl.requests["stale"] = &rateLimitEntry{
		count:      1,
		resetTime:  staleTime,
		lastAccess: staleTime,
	}

	// Add a fresh entry
	rl.Allow("fresh")

	rl.cleanup()

	if _, ok := rl.requests["stale"]; ok {
		t.Error("cleanup() should have removed stale entry")
	}
	if _, ok := rl.requests["fresh"]; !ok {
		t.Error("cleanup() should not have removed fresh entry")
	}
}
