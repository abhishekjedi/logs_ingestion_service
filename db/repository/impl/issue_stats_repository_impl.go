package impl

import (
	"context"
	"fmt"
	"time"

	"error-logging/db/repository"
	chclient "error-logging/pkg/client/clickhouse"
)

type issueStatsRepository struct {
	conn *chclient.NativeClient
}

func NewIssueStatsRepository(c *chclient.NativeClient) repository.IssueStatsRepository {
	return &issueStatsRepository{conn: c}
}

// ActiveSince merges the aggregate states into per-issue totals across all hours,
// keeping only issues whose latest hour is at/after `since`.
func (r *issueStatsRepository) ActiveSince(ctx context.Context, since time.Time) ([]repository.IssueStatsAggregate, error) {
	const q = `
SELECT service_id,
       issue_id,
       countMerge(event_count) AS events,
       uniqMerge(users)        AS users,
       uniqMerge(sessions)     AS sessions,
       max(hour)               AS last_hour
FROM issue_stats
GROUP BY service_id, issue_id
HAVING last_hour >= ?`

	rows, err := r.conn.Conn.Query(ctx, q, since)
	if err != nil {
		return nil, fmt.Errorf("query issue_stats: %w", err)
	}
	defer rows.Close()

	var out []repository.IssueStatsAggregate
	for rows.Next() {
		var a repository.IssueStatsAggregate
		if err := rows.Scan(&a.ServiceID, &a.IssueID, &a.EventCount, &a.Users, &a.Sessions, &a.LastSeen); err != nil {
			return nil, fmt.Errorf("scan issue_stats: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
