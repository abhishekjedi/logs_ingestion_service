package impl

import (
	"context"
	"fmt"
	"time"

	redisclient "error-logging/pkg/client/redis"
	"error-logging/services"
)

// rateLimiter is a Redis fixed-window limiter. Redis is degradable: if it is
// unavailable the limiter allows the events rather than dropping fidelity.
type rateLimiter struct {
	redis *redisclient.Client
	limit int64
}

func NewRateLimiter(redis *redisclient.Client, limitPerSecond int) services.RateLimiter {
	return &rateLimiter{redis: redis, limit: int64(limitPerSecond)}
}

// AllowN reserves n occurrences of key in the current 1s window with a single
// INCRBY and returns how many fall within the limit. One Redis op per key per
// cycle instead of one per event.
func (r *rateLimiter) AllowN(ctx context.Context, key string, n int) int {
	if n <= 0 {
		return 0
	}
	if r.redis == nil || r.redis.RDB == nil {
		return n // degrade: allow all
	}

	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, time.Now().Unix())
	total, err := r.redis.RDB.IncrBy(ctx, windowKey, int64(n)).Result()
	if err != nil {
		return n // degrade: allow all
	}
	if total == int64(n) { // key was newly created this window
		r.redis.RDB.Expire(ctx, windowKey, 2*time.Second)
	}

	// Occurrences already used before this batch = total - n.
	allowed := r.limit - (total - int64(n))
	if allowed < 0 {
		allowed = 0
	}
	if allowed > int64(n) {
		allowed = int64(n)
	}
	return int(allowed)
}
