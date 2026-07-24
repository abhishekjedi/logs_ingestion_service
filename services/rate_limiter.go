package services

import "context"

// RateLimiter gates full-fidelity storage (never counting). AllowN reserves up to
// n occurrences of a key in the current window and returns how many are permitted,
// so a whole cycle's worth of one fingerprint costs a single Redis round trip
// instead of one per event.
type RateLimiter interface {
	AllowN(ctx context.Context, key string, n int) int
}
