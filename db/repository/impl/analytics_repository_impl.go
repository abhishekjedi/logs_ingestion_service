package impl

import (
	"context"
	"fmt"
	"time"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	chclient "error-logging/pkg/client/clickhouse"
)

type analyticsRepository struct {
	conn *chclient.NativeClient
}

func NewAnalyticsRepository(c *chclient.NativeClient) repository.AnalyticsRepository {
	return &analyticsRepository{conn: c}
}

func (r *analyticsRepository) IssueTimeseries(ctx context.Context, issueID uint64, from, to time.Time) ([]dbdto.TimePoint, error) {
	const q = `
SELECT hour, countMerge(event_count) AS events, uniqMerge(users) AS users
FROM issue_stats
WHERE issue_id = ? AND hour >= ? AND hour <= ?
GROUP BY hour ORDER BY hour`

	rows, err := r.conn.Conn.Query(ctx, q, issueID, from, to)
	if err != nil {
		return nil, fmt.Errorf("issue timeseries: %w", err)
	}
	defer rows.Close()

	var out []dbdto.TimePoint
	for rows.Next() {
		var p dbdto.TimePoint
		if err := rows.Scan(&p.Timestamp, &p.Events, &p.Users); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *analyticsRepository) ServiceOverview(ctx context.Context, serviceID uint64, from, to time.Time) ([]dbdto.ServiceOverviewPoint, error) {
	const q = `
SELECT hour,
       countMerge(event_count) AS events,
       uniqMerge(issues)       AS issues,
       uniqMerge(users)        AS users
FROM service_stats
WHERE service_id = ? AND hour >= ? AND hour <= ?
GROUP BY hour ORDER BY hour`

	rows, err := r.conn.Conn.Query(ctx, q, serviceID, from, to)
	if err != nil {
		return nil, fmt.Errorf("service overview: %w", err)
	}
	defer rows.Close()

	var out []dbdto.ServiceOverviewPoint
	for rows.Next() {
		var p dbdto.ServiceOverviewPoint
		if err := rows.Scan(&p.Timestamp, &p.Events, &p.Issues, &p.Users); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *analyticsRepository) ReleaseHealth(ctx context.Context, serviceID uint64, from, to time.Time) ([]dbdto.ReleaseHealth, error) {
	const q = `
SELECT release,
       uniqMerge(sessions_total)   AS total,
       uniqMerge(sessions_errored) AS errored
FROM release_health
WHERE service_id = ? AND hour >= ? AND hour <= ?
GROUP BY release ORDER BY release`

	rows, err := r.conn.Conn.Query(ctx, q, serviceID, from, to)
	if err != nil {
		return nil, fmt.Errorf("release health: %w", err)
	}
	defer rows.Close()

	var out []dbdto.ReleaseHealth
	for rows.Next() {
		var h dbdto.ReleaseHealth
		if err := rows.Scan(&h.Release, &h.SessionsTotal, &h.SessionsErrored); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (r *analyticsRepository) RecentErrorEvents(ctx context.Context, issueID uint64, limit int) ([]dbdto.ErrorEventDetail, error) {
	const q = `
SELECT toString(event_id), timestamp, severity_text, exception_type, exception_message,
       user_id, session_id, environment, release, trace_id, span_id, stack_frames,
       attributes, resource_attributes
FROM error_events
WHERE issue_id = ?
ORDER BY timestamp DESC
LIMIT ?`

	rows, err := r.conn.Conn.Query(ctx, q, issueID, limit)
	if err != nil {
		return nil, fmt.Errorf("recent error events: %w", err)
	}
	defer rows.Close()

	var out []dbdto.ErrorEventDetail
	for rows.Next() {
		var e dbdto.ErrorEventDetail
		if err := rows.Scan(
			&e.EventID, &e.Timestamp, &e.SeverityText, &e.ExceptionType, &e.ExceptionMessage,
			&e.UserID, &e.SessionID, &e.Environment, &e.Release, &e.TraceID, &e.SpanID, &e.StackFrames,
			&e.Attributes, &e.ResourceAttributes,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *analyticsRepository) Breadcrumbs(ctx context.Context, serviceID uint64, sessionID string, before time.Time, limit int) ([]dbdto.Breadcrumb, error) {
	const q = `
SELECT timestamp, severity_text, body, exception_type
FROM logs
WHERE service_id = ? AND session_id = ? AND timestamp <= ?
ORDER BY timestamp DESC
LIMIT ?`

	rows, err := r.conn.Conn.Query(ctx, q, serviceID, sessionID, before, limit)
	if err != nil {
		return nil, fmt.Errorf("breadcrumbs: %w", err)
	}
	defer rows.Close()

	var out []dbdto.Breadcrumb
	for rows.Next() {
		var b dbdto.Breadcrumb
		if err := rows.Scan(&b.Timestamp, &b.SeverityText, &b.Body, &b.ExceptionType); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
