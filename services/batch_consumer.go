package services

import "context"

type BatchConsumer interface {
	Run(ctx context.Context) error
}
