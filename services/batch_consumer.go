package services

import "context"

// BatchConsumer drives the worker's fetch → transform → flush → commit cycle.
type BatchConsumer interface {
	Run(ctx context.Context) error
}
