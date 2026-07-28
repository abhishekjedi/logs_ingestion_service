package dto

import "time"

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

type ErrorEventFrame struct {
	File     string `ch:"file" json:"file"`
	Function string `ch:"function" json:"function"`
	Line     uint32 `ch:"line" json:"line"`
	Col      uint32 `ch:"col" json:"col"`
	InApp    uint8  `ch:"in_app" json:"in_app"`
}

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
