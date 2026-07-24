package repository

import (
	"context"
	"time"

	dbdto "error-logging/db/dto"
)

// IssueStatsUpdate carries denormalized counts synced from ClickHouse onto an issue.
type IssueStatsUpdate struct {
	IssueID          uint64
	EventCount       uint64
	AffectedUsers    uint64
	AffectedSessions uint64
	LastSeen         time.Time
}

type IssueRepository interface {
	// ResolveOrCreate returns the issue for (service_id, fingerprint), creating it if
	// new. The unique constraint settles concurrent first-sightings without a lock.
	// The bool reports whether the issue was newly created.
	ResolveOrCreate(ctx context.Context, issue *dbdto.Issue) (*dbdto.Issue, bool, error)
	// MarkRegressed flips a resolved issue to regressed (no-op otherwise), returning
	// whether a transition happened.
	MarkRegressed(ctx context.Context, id uint64) (bool, error)
	// UpdateStatsBatch writes denormalized counts (from ClickHouse) onto issues in
	// one transaction. Idempotent: values are absolute, not deltas.
	UpdateStatsBatch(ctx context.Context, updates []IssueStatsUpdate) error
}
