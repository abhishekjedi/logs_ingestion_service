package services

import "context"

type SyncService interface {
	SyncOnce(ctx context.Context) (int, error)
}
