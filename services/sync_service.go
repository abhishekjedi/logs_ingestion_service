package services

import "context"

// SyncService reconciles the denormalized issue counters in MySQL with the
// authoritative rollups in ClickHouse.
type SyncService interface {
	// SyncOnce runs one reconciliation pass and returns the number of issues updated.
	SyncOnce(ctx context.Context) (int, error)
}
