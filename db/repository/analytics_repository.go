package repository

import (
	"context"
	"time"
)

// TimePoint is one hourly bucket of an issue's trend.
type TimePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Events    uint64    `json:"events"`
	Users     uint64    `json:"users"`
}

// ServiceOverviewPoint is one hourly bucket of a service's activity.
type ServiceOverviewPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Events    uint64    `json:"events"`
	Issues    uint64    `json:"issues"`
	Users     uint64    `json:"users"`
}

// ReleaseHealth is crash-free session data for one release.
type ReleaseHealth struct {
	Release         string `json:"release"`
	SessionsTotal   uint64 `json:"sessions_total"`
	SessionsErrored uint64 `json:"sessions_errored"`
}

// ErrorEventDetail is a full-fidelity event for issue drill-down.
type ErrorEventDetail struct {
	EventID            string            `json:"event_id"`
	Timestamp          time.Time         `json:"timestamp"`
	SeverityText       string            `json:"severity_text"`
	ExceptionType      string            `json:"exception_type"`
	ExceptionMessage   string            `json:"exception_message"`
	UserID             string            `json:"user_id"`
	SessionID          string            `json:"session_id"`
	Environment        string            `json:"environment"`
	Release            string            `json:"release"`
	TraceID            string            `json:"trace_id"`
	SpanID             string            `json:"span_id"`
	StackFrames        []ErrorEventFrame `json:"stack_frames"`
	Attributes         map[string]string `json:"attributes"`
	ResourceAttributes map[string]string `json:"resource_attributes"`
}

// Breadcrumb is one preceding log event in the same session (the trail before an error).
type Breadcrumb struct {
	Timestamp     time.Time `json:"timestamp"`
	SeverityText  string    `json:"severity_text"`
	Body          string    `json:"body"`
	ExceptionType string    `json:"exception_type"`
}

// AnalyticsRepository reads the ClickHouse analytics tables for the read API.
type AnalyticsRepository interface {
	IssueTimeseries(ctx context.Context, issueID uint64, from, to time.Time) ([]TimePoint, error)
	ServiceOverview(ctx context.Context, serviceID uint64, from, to time.Time) ([]ServiceOverviewPoint, error)
	ReleaseHealth(ctx context.Context, serviceID uint64, from, to time.Time) ([]ReleaseHealth, error)
	RecentErrorEvents(ctx context.Context, issueID uint64, limit int) ([]ErrorEventDetail, error)
	// Breadcrumbs returns the preceding log events in a session (the trail before an error).
	Breadcrumbs(ctx context.Context, serviceID uint64, sessionID string, before time.Time, limit int) ([]Breadcrumb, error)
}
