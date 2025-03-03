package token_bucket

import (
	"sync"
	"time"
)

type TokenBucket struct {
	mu             sync.Mutex
	capacity       int // Maximum number of tokens in the bucket
	refillRate     int // Number of tokens to add per second
	currentFill    int // Current number of tokens in the bucket
	lastRefillTime time.Time
}

func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		capacity:       capacity,
		refillRate:     refillRate,
		currentFill:    capacity,
		lastRefillTime: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()

	// Tokens according to the elapsed time
	tokenCount := tb.currentFill + int(elapsed*float64(tb.refillRate))

	// If the token count is greater than the capacity, set it to the capacity
	tb.currentFill = min(tb.capacity, tokenCount)
	tb.lastRefillTime = now
}

// It will check if the token is available or not
// If the token is available, it will return true - allowing the request
// Else, it will return false - denying the request
func (tb *TokenBucket) AllowRequest(tokens int) bool {
	tb.refill()
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.currentFill >= tokens {
		tb.currentFill -= tokens
		return true
	}
	return false
}
