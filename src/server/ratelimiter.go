package server

import (
	"sync"
	"time"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	mu              sync.Mutex
	requests        map[string]*rateLimitEntry
	limit           int           // Maximum requests allowed
	window          time.Duration // Time window for rate limiting
	cleanupInterval time.Duration // Cleanup interval
}

type rateLimitEntry struct {
	count      int
	resetTime  time.Time
	lastAccess time.Time
}

// NewRateLimiter creates a new RateLimiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests:        make(map[string]*rateLimitEntry),
		limit:           limit,
		window:          window,
		cleanupInterval: time.Minute * 5,
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Allow checks if a request from the given identifier is allowed
func (rl *RateLimiter) Allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	entry, exists := rl.requests[identifier]
	if !exists {
		// First request from this identifier
		rl.requests[identifier] = &rateLimitEntry{
			count:      1,
			resetTime:  now.Add(rl.window),
			lastAccess: now,
		}
		return true
	}

	// Check if window has expired
	if now.After(entry.resetTime) {
		// Reset the counter
		entry.count = 1
		entry.resetTime = now.Add(rl.window)
		entry.lastAccess = now
		return true
	}

	// Update last access time
	entry.lastAccess = now

	// Check if limit exceeded
	if entry.count >= rl.limit {
		return false
	}

	// Increment counter
	entry.count++
	return true
}

// Reset resets the rate limit for a given identifier
func (rl *RateLimiter) Reset(identifier string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.requests, identifier)
}

// cleanupLoop periodically removes old entries
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes expired entries from the rate limiter
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for identifier, entry := range rl.requests {
		// Remove entries that haven't been accessed in 2x the cleanup interval
		if now.Sub(entry.lastAccess) > rl.cleanupInterval*2 {
			delete(rl.requests, identifier)
		}
	}
}

// GetRemaining returns the remaining requests for an identifier
func (rl *RateLimiter) GetRemaining(identifier string) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.requests[identifier]
	if !exists {
		return rl.limit
	}

	now := time.Now()
	if now.After(entry.resetTime) {
		return rl.limit
	}

	remaining := rl.limit - entry.count
	if remaining < 0 {
		return 0
	}

	return remaining
}

// GetResetTime returns the reset time for an identifier
func (rl *RateLimiter) GetResetTime(identifier string) time.Time {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.requests[identifier]
	if !exists {
		return time.Now().Add(rl.window)
	}

	return entry.resetTime
}
