package rate_limiter

import (
	"context"
	"log"
	"strconv"
	"time"

	token_bucket "github.com/krishpatel023/ratelimiter/internal/token-bucket"
	"github.com/redis/go-redis/v9"
)

type DistributedRateLimiter struct {
	client          *redis.Client
	keyPrefix       string
	expirationTime  time.Duration
	cleanupInterval time.Duration
}

func NewDistributedRateLimiter(redisAddr, password string, db int, keyPrefix string, cleanupInterval, expirationTime time.Duration) (*DistributedRateLimiter, error) {

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Wait for Redis to be ready
	// Requesting the ping command for at least 3 times
	// with a 2 second interval
	for i := 0; i < 3; i++ {
		if err := client.Ping(ctx).Err(); err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}

	return &DistributedRateLimiter{
		client:          client,
		keyPrefix:       keyPrefix,
		expirationTime:  expirationTime,
		cleanupInterval: cleanupInterval,
	}, nil
}

// Stop the rate limiter
// It closes the Redis client internally
func (rl *DistributedRateLimiter) Stop() {
	_ = rl.client.Close()
}

// AllowRequest checks if the request is allowed
func (rl *DistributedRateLimiter) AllowRequest(id string, tokens int, totalTokens int, refillRate int) bool {

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	bucketKey := rl.keyPrefix + ":" + id

	// Use Lua script for atomic operations
	script := token_bucket.TokenBucketLuaScript()

	// Execute the Lua script
	keys := []string{bucketKey}
	args := []interface{}{
		strconv.FormatFloat(float64(tokens), 'f', -1, 64),
		strconv.FormatFloat(float64(totalTokens), 'f', -1, 64),
		strconv.FormatFloat(float64(refillRate), 'f', -1, 64),
		int(rl.expirationTime.Seconds()),
	}

	result, err := rl.client.Eval(ctx, script, keys, args...).Int()
	if err != nil {
		log.Printf("Error executing Redis Lua script: %v", err)
		return false
	}

	return result == 1
}
