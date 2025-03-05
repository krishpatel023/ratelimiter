package limiters

import (
	"net/http"

	"github.com/krishpatel023/ratelimiter/internal/helper"
	rate_limiter "github.com/krishpatel023/ratelimiter/internal/rate-limiter"
)

// Distributed Rate Limiter Middleware
// It will handle the reverse proxy and the verification of the UniqueHeaderNameInRequest Header and also will check
// the allowed requests. Based on the response will either decline or forward the request.
// It is also responsible for the reverse proxy setup and the forwarding of the request if the
// request is allowed
func DistributedRateLimitingMiddleware(rl *rate_limiter.DistributedRateLimiter, config DistributedRateLimiterConfig) http.Handler {
	// Create a reverse proxy
	handler := reverseProxy(config.TargetURL)
	if handler == nil {
		return nil
	}

	// Check if the UniqueHeaderNameInRequest header is present
	if config.UniqueHeaderNameInRequest == "" {
		helper.Log("Request rejected: Set UniqueHeaderNameInRequest header in config", "warning")
		http.Error(nil, "Set UniqueHeaderNameInRequest header in config", http.StatusBadRequest)
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
		handler.ServeHTTP(w, r)
	})
}
