package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter Redis 实现的滑动窗口速率限制器
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

// Allow 检查是否允许请求
func (l *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixNano()
	windowStart := now - l.duration.Nanoseconds()

	pipe := l.client.Pipeline()
	// 移除时间窗口之前的请求
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	// 添加当前请求
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
	// 获取当前时间窗口内的请求数
	count := pipe.ZCard(ctx, key)
	// 设置 key 的过期时间
	pipe.Expire(ctx, key, l.duration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	return count.Val() <= l.maxRequests, nil
}
