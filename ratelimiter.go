package ratelimiter

import (
	"net/http"

	rate_limiter "github.com/krishpatel023/ratelimiter/internal/rate-limiter"
	"github.com/krishpatel023/ratelimiter/limiters"
)

type LocalWrapper struct {
	Config              limiters.LocalRateLimiterConfig
	New                 func(config limiters.LocalRateLimiterConfig) (*rate_limiter.LocalRateLimiter, error)
	Stop                func(rl *rate_limiter.LocalRateLimiter)
	MiddlewareWithProxy func(rl *rate_limiter.LocalRateLimiter, config limiters.LocalRateLimiterConfig) http.Handler
	Middleware          func(rl *rate_limiter.LocalRateLimiter, config limiters.LocalRateLimiterConfig) http.Handler
}

var Local = LocalWrapper{
	Config:              limiters.GetLocalRateLimiterDefaultConfig(),
	New:                 limiters.CreateLocalRateLimiter,
	Stop:                limiters.StopLocalRateLimiter,
	MiddlewareWithProxy: limiters.LocalRateLimitingMiddleware,
	Middleware:          limiters.LocalNonProxyRateLimitingMiddleware,
}

type DistributedWrapper struct {
	Config              limiters.DistributedRateLimiterConfig
	New                 func(config limiters.DistributedRateLimiterConfig) (*rate_limiter.DistributedRateLimiter, error)
	Stop                func(rl *rate_limiter.DistributedRateLimiter)
	Middleware          func(rl *rate_limiter.DistributedRateLimiter, config limiters.DistributedRateLimiterConfig) http.Handler
	MiddlewareWithProxy func(rl *rate_limiter.DistributedRateLimiter, config limiters.DistributedRateLimiterConfig) http.Handler
}

var Distributed = DistributedWrapper{
	Config:              limiters.GetDistributedRateLimiterDefaultConfig(),
	New:                 limiters.CreateDistributedRateLimiter,
	Stop:                limiters.StopDistributedRateLimiter,
	MiddlewareWithProxy: limiters.DistributedRateLimitingMiddleware,
	Middleware:          limiters.DistributedNonProxyRateLimitingMiddleware,
}
