package repository

import (
	"context"
	"time"
)

// IssueStatsAggregate is the rolled-up state for one issue, read from ClickHouse
// issue_stats (the source of truth for counts).
type IssueStatsAggregate struct {
	ServiceID  uint64
	IssueID    uint64
	EventCount uint64
	Users      uint64
	Sessions   uint64
	// LastSeen is hour-granular (max bucket in issue_stats).
	LastSeen time.Time
}

// IssueStatsRepository reads the ClickHouse issue rollups.
type IssueStatsRepository interface {
	// ActiveSince returns totals for issues with activity at/after `since`.
	ActiveSince(ctx context.Context, since time.Time) ([]IssueStatsAggregate, error)
}
