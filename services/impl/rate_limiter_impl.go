package impl

import (
	"context"
	"fmt"
	"time"

	redisclient "error-logging/pkg/client/redis"
	"error-logging/services"
)

type rateLimiter struct {
	redis *redisclient.Client
	limit int64
}

func NewRateLimiter(redis *redisclient.Client, limitPerSecond int) services.RateLimiter {
	return &rateLimiter{redis: redis, limit: int64(limitPerSecond)}
}

// AllowN reserves a whole cycle's occurrences with one Redis INCRBY instead of
// one operation per event.
func (r *rateLimiter) AllowN(ctx context.Context, key string, n int) int {
	if n <= 0 {
		return 0
	}
	if r.redis == nil || r.redis.RDB == nil {
		return n
	}

	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, time.Now().Unix())
	total, err := r.redis.RDB.IncrBy(ctx, windowKey, int64(n)).Result()
	if err != nil {
		return n
	}
	if total == int64(n) {
		r.redis.RDB.Expire(ctx, windowKey, 2*time.Second)
	}

	allowed := r.limit - (total - int64(n))
	if allowed < 0 {
		allowed = 0
	}
	if allowed > int64(n) {
		allowed = int64(n)
	}
	return int(allowed)
}
