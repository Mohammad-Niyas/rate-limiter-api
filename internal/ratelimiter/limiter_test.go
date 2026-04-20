package ratelimiter

import (
	"testing"
	"time"

	"github.com/mohammad-niyas/rate-limiter-api/internal/store"
)

// TestAllowUnderLimit checks that requests under the limit are allowed.
func TestAllowUnderLimit(t *testing.T) {
	memStore := store.NewMemoryStore()

	cfg := Config{MaxRequests: 5, WindowSize: 1 * time.Minute}
	limiter := NewRateLimiter(memStore, cfg)

	for i := 0; i < 5; i++ {
		result := limiter.Allow("user1")
		if !result.Allowed {
			t.Errorf("Request %d should be allowed but was blocked", i+1)
		}
		expectedRemaining := 4 - i
		if result.Remaining != expectedRemaining {
			t.Errorf("Request %d: expected remaining %d, got %d", i+1, expectedRemaining, result.Remaining)
		}
	}
}

// TestBlockOverLimit checks that the 6th request is blocked.
func TestBlockOverLimit(t *testing.T) {
	memStore := store.NewMemoryStore()

	cfg := Config{MaxRequests: 5, WindowSize: 1 * time.Minute}
	limiter := NewRateLimiter(memStore, cfg)

	for i := 0; i < 5; i++ {
		limiter.Allow("user1")
	}

	result := limiter.Allow("user1")
	
	if result.Allowed {
		t.Error("6th request should be blocked but was allowed")
	}
	if result.Remaining != 0 {
		t.Errorf("Expected remaining 0, got %d", result.Remaining)
	}
}

// TestDifferentUsersIndependent checks that rate limits are per-user.
// user1 being rate-limited should NOT affect user2.
func TestDifferentUsersIndependent(t *testing.T) {
	memStore := store.NewMemoryStore()

	cfg := Config{MaxRequests: 5, WindowSize: 1 * time.Minute}
	limiter := NewRateLimiter(memStore, cfg)

	for i := 0; i < 5; i++ {
		limiter.Allow("user1")
	}

	result1 := limiter.Allow("user1")
	if result1.Allowed {
		t.Error("user1 should be blocked")
	}

	result2 := limiter.Allow("user2")
	if !result2.Allowed {
		t.Error("user2 should be allowed — rate limits are per-user")
	}
}

// TestWindowExpiry checks that old requests expire after the window.
func TestWindowExpiry(t *testing.T) {
	memStore := store.NewMemoryStore()

	cfg := Config{MaxRequests: 2, WindowSize: 1 * time.Second}
	limiter := NewRateLimiter(memStore, cfg)

	limiter.Allow("user1")
	limiter.Allow("user1")

	result := limiter.Allow("user1")
	if result.Allowed {
		t.Error("Should be blocked — limit reached")
	}

	time.Sleep(1100 * time.Millisecond)

	result = limiter.Allow("user1")
	if !result.Allowed {
		t.Error("Should be allowed after window expiry")
	}
}

// TestConcurrentRequests checks that rate limiting is accurate
// when many goroutines send requests simultaneously.
func TestConcurrentRequests(t *testing.T) {
	memStore := store.NewMemoryStore()

	cfg := Config{MaxRequests: 5, WindowSize: 1 * time.Minute}
	limiter := NewRateLimiter(memStore, cfg)

	results := make(chan bool, 20)

	for i := 0; i < 20; i++ {
		go func() {
			result := limiter.Allow("concurrent-user")
			results <- result.Allowed
		}()
	}

	allowed := 0
	blocked := 0
	for i := 0; i < 20; i++ {
		if <-results {
			allowed++
		} else {
			blocked++
		}
	}

	if allowed != 5 {
		t.Errorf("Expected exactly 5 allowed, got %d (blocked: %d)", allowed, blocked)
	}
}

// TestGetStats checks that stats are returned correctly.
func TestGetStats(t *testing.T) {
	memStore := store.NewMemoryStore()

	cfg := Config{MaxRequests: 5, WindowSize: 1 * time.Minute}
	limiter := NewRateLimiter(memStore, cfg)

	limiter.Allow("user1")
	limiter.Allow("user1")
	limiter.Allow("user2")

	stats := limiter.GetStats()
	if len(stats) != 2 {
		t.Errorf("Expected 2 users in stats, got %d", len(stats))
	}
}
