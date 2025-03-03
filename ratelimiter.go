package ratelimiter

import (
	"net/http"

	rate_limiter "github.com/krishpatel023/ratelimiter/internal/rate-limiter"
	"github.com/krishpatel023/ratelimiter/limiters"
)

type LocalWrapper struct {
	Config     limiters.LocalRateLimiterConfig
	New        func(config limiters.LocalRateLimiterConfig) (*rate_limiter.LocalRateLimiter, error)
	Stop       func(rl *rate_limiter.LocalRateLimiter)
	Middleware func(rl *rate_limiter.LocalRateLimiter, config limiters.LocalRateLimiterConfig) http.Handler
}

var Local = LocalWrapper{
	Config:     limiters.GetLocalRateLimiterDefaultConfig(),
	New:        limiters.CreateLocalRateLimiter,
	Stop:       limiters.StopLocalRateLimiter,
	Middleware: limiters.LocalRateLimitingMiddleware,
}
