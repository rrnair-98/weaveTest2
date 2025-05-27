package client

import (
	"context"
	"go.uber.org/zap"
	"sync"
	"time"
)

var (
	rateLimiterInstance *RateLimiter
	rateLimiterOnce     sync.Once
	rateLimiterMu       sync.RWMutex
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu            sync.Mutex
	maxTokens     int           // Maximum number of tokens in the bucket
	tokens        int           // Current number of tokens
	refillRate    time.Duration // How often a token is added
	lastRefill    time.Time     // When tokens were last refilled
	nextResetTime time.Time     // When the rate limit will fully reset (from X-RateLimit-Reset)
	logger        *zap.Logger
}

// GetRateLimiter returns the singleton instance of the rate limiter
func GetRateLimiter() *RateLimiter {
	rateLimiterMu.RLock()
	if rateLimiterInstance != nil {
		defer rateLimiterMu.RUnlock()
		return rateLimiterInstance
	}
	rateLimiterMu.RUnlock()

	return rateLimiterInstance
}

// InitRateLimiter initializes the rate limiter with custom parameters
// This should be called early in your application startup
func InitRateLimiter(maxRequests int, timeWindow time.Duration, logger *zap.Logger) {
	rateLimiterOnce.Do(func() {
		rateLimiterMu.Lock()
		defer rateLimiterMu.Unlock()

		rateLimiterInstance = newRateLimiter(maxRequests, timeWindow, logger)
	})
}

// newRateLimiter creates a new token bucket rate limiter instance
func newRateLimiter(maxTokens int, timeWindow time.Duration, logger *zap.Logger) *RateLimiter {
	// Calculate token refill rate (how often a single token is added)
	refillRate := timeWindow / time.Duration(maxTokens)

	return &RateLimiter{
		maxTokens:  maxTokens,
		tokens:     maxTokens, // Start with a full bucket
		refillRate: refillRate,
		lastRefill: time.Now(),
		logger:     logger,
	}
}

// refillTokens calculates how many tokens should be added based on time elapsed
// and adds them to the bucket (up to maxTokens)
func (rl *RateLimiter) refillTokens() {
	now := time.Now()

	// If we have a reset time and it has passed, fully refill the bucket
	if !rl.nextResetTime.IsZero() && now.After(rl.nextResetTime) {
		rl.tokens = rl.maxTokens
		rl.lastRefill = now
		rl.nextResetTime = time.Time{} // Clear the reset time
		rl.logger.Debug("rate limit reset time reached, bucket fully refilled",
			zap.Int("tokens", rl.tokens))
		return
	}

	// Otherwise calculate tokens to add based on time elapsed
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int(elapsed / rl.refillRate)

	if tokensToAdd > 0 {
		rl.tokens = min(rl.tokens+tokensToAdd, rl.maxTokens)
		// Update last refill time, but only for the tokens we actually added
		rl.lastRefill = rl.lastRefill.Add(time.Duration(tokensToAdd) * rl.refillRate)
		rl.logger.Debug("refilled tokens",
			zap.Int("added", tokensToAdd),
			zap.Int("current", rl.tokens))
	}
}

// UpdateResetTime updates the rate limiter with information from X-RateLimit-Reset header
func (rl *RateLimiter) UpdateResetTime(resetTime time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.nextResetTime = resetTime
	rl.logger.Debug("updated rate limit reset time", zap.Time("resetTime", resetTime))
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait() {
	for {
		waitTime := rl.reserveToken()
		if waitTime == 0 {
			return
		}
		time.Sleep(waitTime)
	}
}

// WaitWithContext blocks until a token is available or context is canceled
func (rl *RateLimiter) WaitWithContext(ctx context.Context) error {
	for {
		waitTime := rl.reserveToken()
		if waitTime == 0 {
			return nil
		}

		select {
		case <-time.After(waitTime):
			// Continue the loop to try again
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// reserveToken attempts to take a token from the bucket
// Returns the wait time needed if no token is available (0 if token was taken)
func (rl *RateLimiter) reserveToken() time.Duration {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()

	if rl.tokens > 0 {
		rl.tokens--
		rl.logger.Debug("token consumed", zap.Int("remaining", rl.tokens))
		return 0
	}

	// Calculate wait time until next token is available
	waitTime := rl.refillRate

	// If we have a reset time and waiting for it is shorter, use that instead
	if !rl.nextResetTime.IsZero() {
		resetWait := time.Until(rl.nextResetTime)
		if resetWait < waitTime {
			waitTime = resetWait
		}
	}

	rl.logger.Debug("no tokens available", zap.Duration("waitTime", waitTime))
	return waitTime
}

// ExecuteWithRateLimit executes the given function while respecting rate limits
func (rl *RateLimiter) ExecuteWithRateLimit(ctx context.Context, fn func() error) error {
	err := rl.WaitWithContext(ctx)
	if err != nil {
		return err
	}

	return fn()
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ShutdownRateLimiter resets the singleton rate limiter instance
func ShutdownRateLimiter() {
	rateLimiterMu.Lock()
	defer rateLimiterMu.Unlock()

	rateLimiterInstance = nil

	// Reset the once so it can be initialized again if needed
	rateLimiterOnce = sync.Once{}
}
