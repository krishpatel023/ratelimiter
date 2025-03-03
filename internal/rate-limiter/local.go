package rate_limiter

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	token_bucket "github.com/krishpatel023/ratelimiter/internal/token-bucket"
)

type BucketWrapper struct {
	Bucket   *token_bucket.TokenBucket
	LastUsed time.Time
}

type LocalRateLimiter struct {
	buckets       *lru.Cache
	mu            sync.RWMutex
	cleanupTicker *time.Ticker  // Ticker for cleanup routine - to remove expired buckets
	stopCleanup   chan struct{} // Channel to stop the cleanup routine
	expiration    time.Duration // Expiration time for buckets
}

func NewLocalRateLimiter(totalEntries int, cleanupInterval, expiration time.Duration) (*LocalRateLimiter, error) {
	cache, err := lru.New(totalEntries)
	if err != nil {
		return nil, err
	}

	limiter := &LocalRateLimiter{
		buckets:       cache,
		cleanupTicker: time.NewTicker(cleanupInterval),
		stopCleanup:   make(chan struct{}),
		expiration:    expiration,
	}

	// Start the cleanup routine
	go limiter.startCleanupRoutine()

	return limiter, nil
}

func (rl *LocalRateLimiter) startCleanupRoutine() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.cleanupExpiredBuckets()
		case <-rl.stopCleanup:
			rl.cleanupTicker.Stop()
			return
		}
	}
}

func (rl *LocalRateLimiter) cleanupExpiredBuckets() {
	now := time.Now()
	expiredKeys := []string{}

	// First pass: collect expired keys
	rl.mu.RLock()
	for _, key := range rl.buckets.Keys() {
		keyStr := key.(string)
		if val, ok := rl.buckets.Peek(keyStr); ok {
			wrapper := val.(*BucketWrapper)
			if now.Sub(wrapper.LastUsed) > rl.expiration {
				expiredKeys = append(expiredKeys, keyStr)
			}
		}
	}
	rl.mu.RUnlock()

	// Second pass: remove expired keys
	if len(expiredKeys) > 0 {
		rl.mu.Lock()
		for _, key := range expiredKeys {
			rl.buckets.Remove(key)
		}
		rl.mu.Unlock()
	}
}

func (rl *LocalRateLimiter) Stop() {
	close(rl.stopCleanup)
}

func (rl *LocalRateLimiter) GetBucket(id string, capacity int, refillRate int) *token_bucket.TokenBucket {
	// First check with read lock
	rl.mu.RLock()
	if val, ok := rl.buckets.Get(id); ok {
		wrapper := val.(*BucketWrapper)
		wrapper.LastUsed = time.Now() // Update last used time
		rl.mu.RUnlock()
		return wrapper.Bucket
	}
	rl.mu.RUnlock()

	// If not found, acquire write lock
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if val, ok := rl.buckets.Get(id); ok {
		wrapper := val.(*BucketWrapper)
		wrapper.LastUsed = time.Now() // Update last used time
		return wrapper.Bucket
	}

	// Create new bucket if still not found
	tb := token_bucket.NewTokenBucket(capacity, refillRate)
	wrapper := &BucketWrapper{
		Bucket:   tb,
		LastUsed: time.Now(),
	}
	rl.buckets.Add(id, wrapper)
	return tb
}

func (rl *LocalRateLimiter) AllowRequest(id string, tokens int, capacity int, refillRate int) bool {
	bucket := rl.GetBucket(id, capacity, refillRate)

	// Update LastUsed time after the bucket is actually used
	rl.mu.Lock()
	if val, ok := rl.buckets.Get(id); ok {
		wrapper := val.(*BucketWrapper)
		wrapper.LastUsed = time.Now()
	}
	rl.mu.Unlock()

	return bucket.AllowRequest(tokens)
}
