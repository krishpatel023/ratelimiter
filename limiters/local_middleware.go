package limiters

import (
	"net/http"

	"github.com/krishpatel023/ratelimiter/internal/helper"
	rate_limiter "github.com/krishpatel023/ratelimiter/internal/rate-limiter"
)

// Local Rate Limiter Middleware
func LocalRateLimitingMiddleware(rl *rate_limiter.LocalRateLimiter, config LocalRateLimiterConfig) http.Handler {
	handler := reverseProxy(config.TargetURL)
	if handler == nil {
		return nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-ID")
		if userID == "" {
			http.Error(w, "Missing X-ID header", http.StatusBadRequest)
			helper.Log("Request rejected: Missing X-ID header", "warning")
			return
		}

		// Can add a function to get dynamic rate limit config
		// totalTokenPerUser, refillRate := getRateLimiterConfig(userID)

		// Get rate limit config
		token_per_req, total_token, refill_rate := 1, config.Capacity, config.RefillRate

		allowed := rl.AllowRequest(userID, token_per_req, total_token, refill_rate)
		if !allowed {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			helper.Log("Rate limit exceeded for user "+userID, "warning")
			return
		}

		helper.Log("Request allowed for user "+userID, "info")
		handler.ServeHTTP(w, r)
	})
}
