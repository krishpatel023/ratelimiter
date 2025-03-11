package limiters

import (
	"net/http"

	"github.com/krishpatel023/ratelimiter/internal/helper"
	rate_limiter "github.com/krishpatel023/ratelimiter/internal/rate-limiter"
)

// Local Rate Limiter Middleware
// It will handle the verification of the UniqueHeaderNameInRequest Header and also will check
// the allowed requests. Based on the response will either decline or accept the request.

func LocalNonProxyRateLimitingMiddleware(rl *rate_limiter.LocalRateLimiter, config LocalRateLimiterConfig) http.Handler {

	// Check unique header name in request
	if config.UniqueHeaderNameInRequest == "" {
		helper.Log("Request rejected: Set UniqueHeaderNameInRequest header in config", "warning")
		return nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check if the X-ID header is present
		// RequestID is used to identify the request group - all requests with the same X-ID header
		// are considered as a single group of requests and are rate limited together
		requestID := r.Header.Get(config.UniqueHeaderNameInRequest)
		if requestID == "" {
			http.Error(w, "Missing UniqueHeaderNameInRequest header", http.StatusBadRequest)
			helper.Log("Request rejected: Missing UniqueHeaderNameInRequest header", "warning")
			return
		}

		// Can add a function to get dynamic rate limit config
		// token_per_req, total_token, refill_rate := getRateLimiterConfig(requestID)

		// Get rate limit config
		token_per_req, total_token, refill_rate := 1, config.Capacity, config.RefillRate

		// Check if the request is allowed
		allowed := rl.AllowRequest(requestID, token_per_req, total_token, refill_rate)
		if !allowed {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			helper.Log("Request blocked - RequestID: "+requestID, "warning")
			return
		}

		helper.Log("Request allowed - RequestID: "+requestID, "info")
		w.WriteHeader(http.StatusOK)
	})
}

// Distributed Rate Limiter Middleware
// It will handle the verification of the UniqueHeaderNameInRequest Header and also will check
// the allowed requests. Based on the response will either decline or accept the request.

func DistributedNonProxyRateLimitingMiddleware(rl *rate_limiter.DistributedRateLimiter, config DistributedRateLimiterConfig) http.Handler {
	// Check if the UniqueHeaderNameInRequest header is present
	if config.UniqueHeaderNameInRequest == "" {
		helper.Log("Request rejected: Set UniqueHeaderNameInRequest header in config", "warning")
		http.Error(nil, "Set UniqueHeaderNameInRequest header in config", http.StatusBadRequest)
		return nil
	}

	// Check if the redis connection is working
	connection, err := RedisCheck(config.RedisDBAddress, config.RedisDBPassword, config.StorageDB)
	if !connection || err != nil {
		helper.Log("Request rejected: Redis connection failed", "warning")
		return nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(config.UniqueHeaderNameInRequest)
		if requestID == "" {
			http.Error(w, "Missing "+config.UniqueHeaderNameInRequest+" header", http.StatusBadRequest)
			helper.Log("Request rejected: Missing "+config.UniqueHeaderNameInRequest+" header", "warning")
			return
		}

		// Can add a function to get dynamic rate limit config
		// totalTokenPerUser, refillRate := getRateLimiterConfig(requestID)

		// Get rate limit config
		token_per_req, total_token, refill_rate := 1, config.Capacity, config.RefillRate

		allowed := rl.AllowRequest(requestID, token_per_req, total_token, refill_rate)
		if !allowed {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			helper.Log("Request blocked - RequestID: "+requestID, "warning")
			return
		}

		helper.Log("Request allowed - RequestID: "+requestID, "info")
		w.WriteHeader(http.StatusOK)
	})
}
