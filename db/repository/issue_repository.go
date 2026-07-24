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

// IssueListFilter parameterizes a paginated issue list query.
type IssueListFilter struct {
	ServiceID uint64
	Status    string // optional exact-match filter
	Sort      string // event_count | last_seen | first_seen (default last_seen)
	Order     string // asc | desc (default desc)
	Limit     int
	Offset    int
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
	// List returns a page of issues for a service plus the total matching count.
	List(ctx context.Context, filter IssueListFilter) ([]dbdto.Issue, int64, error)
	// GetByID returns a single issue.
	GetByID(ctx context.Context, id uint64) (*dbdto.Issue, error)
}
