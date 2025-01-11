package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter redis sliding window rate limiter
type RateLimiter struct {
	client      *redis.Client
	maxRequests int64
	duration    time.Duration
}

func NewRateLimiter(client *redis.Client, maxRequests int64, duration time.Duration) *RateLimiter {
	return &RateLimiter{
		client:      client,
		maxRequests: maxRequests,
		duration:    duration,
	}
}

// Allow check if request is allowed
func (l *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixNano()
	windowStart := now - l.duration.Nanoseconds()

	pipe := l.client.Pipeline()
	// remove requests before time window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	// add current request
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
	// get current time window requests count
	count := pipe.ZCard(ctx, key)
	// set key expiration time
	pipe.Expire(ctx, key, l.duration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	return count.Val() <= l.maxRequests, nil
}
