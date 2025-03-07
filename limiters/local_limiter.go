package limiters

import (
	"log"
	"time"

	rate_limiter "github.com/krishpatel023/ratelimiter/internal/rate-limiter"
)

// RateLimiterConfig holds configuration for both implementations
type LocalRateLimiterConfig struct {
	Capacity                  int           // Total number of tokens in the bucket
	RefillRate                int           // Number of tokens to add per second
	TargetURL                 string        // Target URL for reverse proxy - to be used in the middleware
	UniqueHeaderNameInRequest string        // Unique header name in the request
	MaxEntries                int           // Maximum number of entries in the cache
	CleanupInterval           time.Duration // Cleanup interval for the cache,
	ExpirationTime            time.Duration // Cleanup interval and expiration time for the cache
}

// GetLocalRateLimiterDefaultConfig returns the default configuration for the local rate limiter
func GetLocalRateLimiterDefaultConfig() LocalRateLimiterConfig {
	return LocalRateLimiterConfig{
		Capacity:                  20,
		RefillRate:                1,
		TargetURL:                 "",
		UniqueHeaderNameInRequest: "",
		MaxEntries:                1000,
		CleanupInterval:           1 * time.Minute,
		ExpirationTime:            5 * time.Minute,
	}
}

// LocalNewRateLimiter creates the appropriate rate limiter based on the configuration
func CreateLocalRateLimiter(config LocalRateLimiterConfig) (*rate_limiter.LocalRateLimiter, error) {
	// Initialize the local rate limiter
	rateLimiter, err := rate_limiter.NewLocalRateLimiter(
		config.MaxEntries,
		config.CleanupInterval,
		config.ExpirationTime,
	)
	if err != nil {
		log.Fatalf("Failed to initialize local rate limiter: %v", err)
	}

	return rateLimiter, nil
}

// StopLocalRateLimiter stops the local rate limiter
func StopLocalRateLimiter(rl *rate_limiter.LocalRateLimiter) {
	rl.Stop()
}
