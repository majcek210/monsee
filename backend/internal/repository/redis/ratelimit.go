package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter implements a Redis sliding-window rate limiter.
type RateLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
}

func NewRateLimiter(rdb *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{rdb: rdb, limit: limit, window: window}
}

// Allow returns true if the key is within the rate limit, false if exceeded.
// Key is typically "rl:<ip>".
func (r *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-r.window).UnixMilli()
	nowMs := now.UnixMilli()

	pipe := r.rdb.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(nowMs), Member: fmt.Sprintf("%d", nowMs)})
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	countCmd := pipe.ZCard(ctx, key)
	pipe.Expire(ctx, key, r.window+time.Second)

	if _, err := pipe.Exec(ctx); err != nil {
		// Fail open — don't block traffic on Redis error
		return true, err
	}

	return countCmd.Val() <= int64(r.limit), nil
}

// NewRedisClient creates a go-redis client from a redis:// URL.
func NewRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr})
}
