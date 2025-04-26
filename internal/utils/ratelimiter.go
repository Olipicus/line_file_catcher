package utils

import (
	"sync"
	"time"
)

// RateLimiter provides a simple rate limiting mechanism
type RateLimiter struct {
	rate       int           // Maximum number of requests per time window
	interval   time.Duration // Time window
	tokens     int           // Current number of available tokens
	lastRefill time.Time     // Last time the bucket was refilled
	mu         sync.Mutex    // Mutex for thread safety
}

// NewRateLimiter creates a new rate limiter
// rate: maximum number of requests
// interval: time window for rate (e.g. 1 minute)
func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		interval:   interval,
		tokens:     rate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request should be allowed based on the current rate limit
// Returns true if the request is allowed, false otherwise
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Refill tokens if interval has passed
	elapsed := now.Sub(rl.lastRefill)
	if elapsed >= rl.interval {
		// Refill the bucket completely
		rl.tokens = rl.rate
		rl.lastRefill = now
	} else if elapsed > 0 {
		// Refill tokens proportionally to time elapsed
		newTokens := int(float64(elapsed) / float64(rl.interval) * float64(rl.rate))
		if newTokens > 0 {
			rl.tokens = min(rl.tokens+newTokens, rl.rate)
			rl.lastRefill = now
		}
	}

	// Check if we have tokens available
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// min returns the smaller of a and b
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RemainingTokens returns the number of remaining tokens
func (rl *RateLimiter) RemainingTokens() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Refill tokens if interval has passed
	elapsed := now.Sub(rl.lastRefill)
	if elapsed >= rl.interval {
		// Refill the bucket completely
		rl.tokens = rl.rate
		rl.lastRefill = now
	} else if elapsed > 0 {
		// Refill tokens proportionally to time elapsed
		newTokens := int(float64(elapsed) / float64(rl.interval) * float64(rl.rate))
		if newTokens > 0 {
			rl.tokens = min(rl.tokens+newTokens, rl.rate)
			rl.lastRefill = now
		}
	}

	return rl.tokens
}

// ResetInterval returns the amount of time until tokens are fully replenished
func (rl *RateLimiter) ResetInterval() time.Duration {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	return rl.interval - time.Since(rl.lastRefill)
}
