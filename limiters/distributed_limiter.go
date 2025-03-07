package limiters

import (
	"log"
	"time"

	rate_limiter "github.com/krishpatel023/ratelimiter/internal/rate-limiter"
)

// DistributedRateLimiterConfig holds configuration for both implementations
type DistributedRateLimiterConfig struct {
	CleanupInterval           time.Duration // Time interval to clean up expired buckets
	ExpirationTime            time.Duration // Time after which a bucket expires
	Capacity                  int           // Capacity of each bucket
	RefillRate                int           // Refill rate of each bucket
	TargetURL                 string        // Target URL for reverse proxy
	UniqueHeaderNameInRequest string        // Unique header name in request
	RedisDBAddress            string        // Redis DB address
	RedisDBPassword           string        // Redis DB password
	StorageDB                 int           // Redis DB number
	KeyPrefix                 string        // Redis key prefix - used for multiple instances
}

// GetDistributedRateLimiterDefaultConfig returns the default configuration for the distributed rate limiter
func GetDistributedRateLimiterDefaultConfig() DistributedRateLimiterConfig {

	config := DistributedRateLimiterConfig{
		CleanupInterval: 5 * time.Minute,
		ExpirationTime:  30 * time.Minute,
		RedisDBAddress:  "localhost:6379",
		RedisDBPassword: "",
		StorageDB:       0,
		KeyPrefix:       "ratelimit",
		Capacity:        20,
		RefillRate:      1,
	}

	return config
}

// CreateDistributedRateLimiter creates the appropriate rate limiter based on the configuration
func CreateDistributedRateLimiter(config DistributedRateLimiterConfig) (*rate_limiter.DistributedRateLimiter, error) {

	rateLimiter, err := rate_limiter.NewDistributedRateLimiter(
		config.RedisDBAddress,
		config.RedisDBPassword,
		config.StorageDB,
		config.KeyPrefix,
		config.CleanupInterval,
		config.ExpirationTime,
	)
	if err != nil {
		log.Fatalf("Failed to initialize distributed rate limiter: %v", err)
	}

	return rateLimiter, nil
}

// StopDistributedRateLimiter stops the local rate limiter
func StopDistributedRateLimiter(rl *rate_limiter.DistributedRateLimiter) {
	rl.Stop()
}
