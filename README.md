# Rate Limiter
[![Go](https://img.shields.io/badge/Go-1.23.2-blue)](https://golang.org/)

> A high-performance Go rate limiting package that provides both local and distributed rate limiting implementations using token bucket algorithm. Supports middleware integration and reverse proxy functionality.

## Overview
This rate limiter package offers a flexible solution for controlling request rates in Go applications with the following features:
- Dual implementation support: Local (in-memory) and Distributed (Redis-based)
- Token bucket algorithm with configurable rates and burst capacity
- Built-in HTTP middleware with reverse proxy support
- Thread-safe operations
- Automatic cleanup of expired rate limiters
- Request identification through customizable headers
- Easy integration with existing Go applications

## Architecture
### Local
The local rate limiter implementation includes:
- In-memory token bucket algorithm using sync.Mutex for thread safety
- LRU cache implementation for storing user buckets
- Automatic cleanup of expired buckets
- Configurable parameters:
  - Bucket capacity
  - Refill rate
  - Maximum entries
  - Cleanup interval
  - Expiration time

### Distributed
The distributed rate limiter leverages Redis for cross-service rate limiting:
- Redis-based token bucket implementation
- Lua scripting for atomic operations
- Configurable parameters:
  - Redis connection settings
  - Key prefixing for multi-tenant support
  - Expiration time
  - Cleanup intervals
- Automatic key expiration and cleanup

## Installation
```bash
go get github.com/krishpatel023/ratelimiter
```

## Implementation

### Local
```go
import "github.com/krishpatel023/ratelimiter"

// Local Rate Limiter Example
func main() {
	// Load default config
	config := ratelimiter.Local.Config

    // Configure local rate limiter
	config.Capacity = 100
	config.RefillRate = 1
	config.TargetURL = "http://localhost:8081" // URL of the server we need to forward the request to

	// Header Name that will have an identification variable
	// Example:  userID
	// Thus, all the request from the same userID will be batched together and ratelimited accordingly
	config.UniqueHeaderNameInRequest = "X-USER-ID"

	// Create New Ratelimiter
	rl, err := ratelimiter.Local.New(config)
	if err != nil {
		log.Fatalf("Failed to initialize rate limiter: %v", err)
	}
	defer ratelimiter.Local.Stop(rl)

    // Setup the middleware
	rateLimitedHandler := ratelimiter.Local.Middleware(rl, config)
	if rateLimitedHandler == nil {
		log.Println("Ratelimiter shutting down due to error")
		return
	}
    http.ListenAndServe(":8080", rateLimitedHandler)
}
```
### Distributed 
```go
// Distributed Rate Limiter Example
func main() {
    // Load default config
    config := ratelimiter.Distributed.Config
    
    // Configure distributed rate limiter
	config.Capacity = 100
	config.RefillRate = 1
	config.RedisDBAddress = "your-redis-server-url"
	config.RedisDBPassword = "password"     // Leave blank is not required

	// URL of the server we need to forward the request to
	config.TargetURL = "http://localhost:8081"

	// Header Name that will have an identification variable
	// Example:  userID
	// Thus, all the request from the same userID will be batched together and ratelimited accordingly
	config.UniqueHeaderNameInRequest = "X-ID"


    // Create New Ratelimiter
	rl, err := ratelimiter.Distributed.New(config)
	if err != nil {
		log.Fatalf("Failed to initialize rate limiter: %v", err)
	}
	defer ratelimiter.Distributed.Stop(rl)

    // Setup the middleware
	rateLimitedHandler := ratelimiter.Distributed.Middleware(rl, config)
	if rateLimitedHandler == nil {
		log.Println("Ratelimiter shutting down due to error")
		return
	}
    http.ListenAndServe(":8080", rateLimitedHandler)
}
```

## Config
### Local Rate Limiter Configuration
```go
    Capacity                  int           // Total tokens in bucket
    RefillRate                int           // Tokens added per second
    TargetURL                 string        // Reverse proxy target URL
    UniqueHeaderNameInRequest string        // Header for request identification
    MaxEntries                int           // Maximum cache entries
    CleanupInterval           time.Duration // Cache cleanup interval
    ExpirationTime            time.Duration // Entry expiration time
```

### Distributed Rate Limiter Configuration
```go
    CleanupInterval           time.Duration // Cache cleanup interval
    ExpirationTime            time.Duration // Entry expiration time
    Capacity                  int           // Total tokens in bucket
    RefillRate                int           // Tokens added per second
    TargetURL                 string        // Reverse proxy target URL
    UniqueHeaderNameInRequest string        // Header for request identification
    RedisDBAddress            string        // Redis DB Address
    RedisDBPassword          string         // Redis DB Password
	StorageDB                 int           // Redis DB number
	KeyPrefix                 string        // Redis key prefix - used for multiple instances
```
