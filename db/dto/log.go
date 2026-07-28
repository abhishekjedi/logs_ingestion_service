package dto

import "time"

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

type Breadcrumb struct {
	Timestamp     time.Time `json:"timestamp"`
	SeverityText  string    `json:"severity_text"`
	Body          string    `json:"body"`
	ExceptionType string    `json:"exception_type"`
}
