package repository

import (
	"context"
	"time"
)

// LogRow is one row destined for the ClickHouse `logs` table. Column order in the
// insert is defined by the repository, matching the table DDL.
type LogRow struct {
	Timestamp          time.Time
	ObservedAt         time.Time
	ProjectID          uint64
	ServiceID          uint64
	IssueID            uint64
	SeverityNumber     uint8
	SeverityText       string
	Body               string
	TraceID            string
	SpanID             string
	Environment        string
	Release            string
	UserID             string
	SessionID          string
	ExceptionType      string
	ExceptionMessage   string
	Attributes         map[string]string
	ResourceAttributes map[string]string
}

// ErrorEventFrame mirrors the ClickHouse stack_frames Tuple element. The `ch` tags
// let clickhouse-go scan the named tuple back into this struct on reads (writes use
// positional tuples, so the tags are inert there).
type ErrorEventFrame struct {
	File     string `ch:"file"`
	Function string `ch:"function"`
	Line     uint32 `ch:"line"`
	Col      uint32 `ch:"col"`
	InApp    uint8  `ch:"in_app"`
}

// ErrorEventRow is one full-fidelity row for the `error_events` table.
type ErrorEventRow struct {
	EventID            string
	IssueID            uint64
	ServiceID          uint64
	ProjectID          uint64
	Timestamp          time.Time
	IngestedAt         time.Time
	SeverityNumber     uint8
	SeverityText       string
	Environment        string
	Release            string
	ExceptionType      string
	ExceptionMessage   string
	UserID             string
	SessionID          string
	StackFrames        []ErrorEventFrame
	RawStacktrace      string
	TraceID            string
	SpanID             string
	Attributes         map[string]string
	ResourceAttributes map[string]string
	S3Key              string
}

// LogRepository writes to the ClickHouse logs table (every record).
type LogRepository interface {
	InsertBatch(ctx context.Context, rows []LogRow) error
}

// ErrorEventRepository writes to the rate-limited ClickHouse error_events table.
type ErrorEventRepository interface {
	InsertBatch(ctx context.Context, rows []ErrorEventRow) error
}
