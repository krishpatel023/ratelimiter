package token_bucket

import (
	"github.com/redis/go-redis/v9"
)

type DistributedTokenBucket struct {
	client     *redis.Client
	capacity   int
	refillRate int
}

func NewDistributedTokenBucket(client *redis.Client, capacity, refillRate int) *DistributedTokenBucket {
	return &DistributedTokenBucket{
		client:     client,
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func TokenBucketLuaScript() string {
	//Lua script for atomic operations
	// It takes care of token verification, refill and request verification
	script := `
	local bucket_key = KEYS[1]
	local tokens_requested = tonumber(ARGV[1])
	local total_tokens = tonumber(ARGV[2])
	local refill_rate = tonumber(ARGV[3])
	local expiration = tonumber(ARGV[4])
	
	-- Get current bucket state
	local current_tokens = redis.call('GET', bucket_key .. ':tokens')
	local last_refill_time = redis.call('GET', bucket_key .. ':last_refill')
	
	-- Initialize if not exists
	if not current_tokens then
		current_tokens = total_tokens
	else
		current_tokens = tonumber(current_tokens)
	end
	
	local now = redis.call('TIME')
	now = tonumber(now[1]) + (tonumber(now[2]) / 1000000)
	
	if not last_refill_time then
		last_refill_time = now
	else
		last_refill_time = tonumber(last_refill_time)
	end
	
	-- Calculate refill
	local elapsed = now - last_refill_time
	local refill_amount = elapsed * refill_rate
	current_tokens = math.min(total_tokens, current_tokens + refill_amount)
	
	-- Check if enough tokens
	local allowed = 0
	if current_tokens >= tokens_requested then
		current_tokens = current_tokens - tokens_requested
		allowed = 1
	end
	
	-- Update bucket state
	-- Only set expiration if the key is new
	if not redis.call('EXISTS', bucket_key .. ':tokens') then
		redis.call('SET', bucket_key .. ':tokens', current_tokens, 'EX', expiration)
		redis.call('SET', bucket_key .. ':last_refill', now, 'EX', expiration)
	else
		redis.call('SET', bucket_key .. ':tokens', current_tokens)
		redis.call('SET', bucket_key .. ':last_refill', now)
	end
	
	return allowed
	`
	return script
}
