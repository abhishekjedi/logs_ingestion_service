package repository

import (
	"context"

	dbdto "error-logging/db/dto"
)

type IssueRepository interface {
	// ResolveOrCreate returns the issue for (service_id, fingerprint), creating it if
	// new. The unique constraint settles concurrent first-sightings without a lock.
	// The bool reports whether the issue was newly created.
	ResolveOrCreate(ctx context.Context, issue *dbdto.Issue) (*dbdto.Issue, bool, error)
	// MarkRegressed flips a resolved issue to regressed (no-op otherwise), returning
	// whether a transition happened.
	MarkRegressed(ctx context.Context, id uint64) (bool, error)
}
