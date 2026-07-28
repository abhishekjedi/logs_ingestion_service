package services

import "context"

type RateLimiter interface {
	AllowN(ctx context.Context, key string, n int) int
}
