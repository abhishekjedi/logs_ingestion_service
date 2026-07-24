package services

import "context"

// ObjectStore is the minimal object-storage surface the pipeline needs (the S3
// client satisfies it). Kept narrow so the processor can be tested with a fake.
type ObjectStore interface {
	Put(ctx context.Context, key string, data []byte, contentType string) error
}
