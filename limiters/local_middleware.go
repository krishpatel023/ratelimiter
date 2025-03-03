package limiters

import (
	"net/http"

	"github.com/krishpatel023/ratelimiter/internal/helper"
	rate_limiter "github.com/krishpatel023/ratelimiter/internal/rate-limiter"
)

// Local Rate Limiter Middleware
// It will handle the reverse proxy and the verification of the X-ID Header and also will check
// the allowed requests. Based on the response will either decline or forward the request.
// It is also responsible for the reverse proxy setup and the forwarding of the request if the
// request is allowed

func LocalRateLimitingMiddleware(rl *rate_limiter.LocalRateLimiter, config LocalRateLimiterConfig) http.Handler {

	// Create a reverse proxy
	handler := reverseProxy(config.TargetURL)
	if handler == nil {
		return nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check if the X-ID header is present
		// RequestID is used to identify the request group - all requests with the same X-ID header
		// are considered as a single group of requests and are rate limited together
		requestID := r.Header.Get("X-ID")
		if requestID == "" {
			http.Error(w, "Missing X-ID header", http.StatusBadRequest)
			helper.Log("Request rejected: Missing X-ID header", "warning")
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
		handler.ServeHTTP(w, r)
	})
}
