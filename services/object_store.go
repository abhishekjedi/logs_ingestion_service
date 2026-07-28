package services

import "context"

type ObjectStore interface {
	Put(ctx context.Context, key string, data []byte, contentType string) error
}
