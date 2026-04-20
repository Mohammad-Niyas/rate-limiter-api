package ratelimiter

import (
	"sync"
	"time"

	"github.com/mohammad-niyas/rate-limiter-api/internal/store"
)

type Config struct {
	MaxRequests int
	WindowSize  time.Duration
}

type Result struct {
	Allowed        bool
	Remaining      int
	RetryAfter     time.Duration
	TotalRequests  int
	WindowRequests int
}

type UserStats struct {
	UserID         string `json:"user_id"`
	TotalRequests  int    `json:"total_requests"`
	WindowRequests int    `json:"requests_in_current_window"`
	Remaining      int    `json:"remaining_requests"`
}

type RateLimiter struct {
	store  store.Store
	config Config
	mu     sync.Mutex
}

func NewRateLimiter(s store.Store, cfg Config) *RateLimiter {
	return &RateLimiter{
		store:  s,
		config: cfg,
	}
}

func (rl *RateLimiter) Allow(userID string) Result {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.config.WindowSize)

	rl.store.CleanExpired(userID, windowStart)

	timestamps := rl.store.GetTimestamps(userID, windowStart)
	currentCount := len(timestamps)

	if currentCount >= rl.config.MaxRequests {
		oldestInWindow := timestamps[0]
		retryAfter := oldestInWindow.Add(rl.config.WindowSize).Sub(now)
		if retryAfter < 0 {
			retryAfter = 0
		}

		return Result{
			Allowed:        false,
			Remaining:      0,
			RetryAfter:     retryAfter,
			WindowRequests: currentCount,
		}
	}

	rl.store.AddRequest(userID, now)

	return Result{
		Allowed:        true,
		Remaining:      rl.config.MaxRequests - currentCount - 1,
		WindowRequests: currentCount + 1,
	}
}

func (rl *RateLimiter) GetStats() []UserStats {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.config.WindowSize)

	allUsers := rl.store.GetAllUsers()
	stats := make([]UserStats, 0, len(allUsers))

	for userID, userData := range allUsers {
		windowCount := 0
		for _, ts := range userData.Timestamps {
			if ts.After(windowStart) {
				windowCount++
			}
		}

		remaining := rl.config.MaxRequests - windowCount
		if remaining < 0 {
			remaining = 0
		}

		stats = append(stats, UserStats{
			UserID:         userID,
			TotalRequests:  userData.TotalCount,
			WindowRequests: windowCount,
			Remaining:      remaining,
		})
	}

	return stats
}
