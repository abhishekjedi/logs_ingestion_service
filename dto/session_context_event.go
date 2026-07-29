package dto

import "time"

type SessionContextEvent struct {
	Timestamp     time.Time `json:"timestamp"`
	Kind          string    `json:"kind"`
	Label         string    `json:"label"`
	Method        string    `json:"method,omitempty"`
	URL           string    `json:"url,omitempty"`
	StatusCode    int       `json:"status_code,omitempty"`
	DurationMS    float64   `json:"duration_ms,omitempty"`
	SourceEventID string    `json:"source_event_id"`
	Count         int       `json:"count,omitempty"`
	Untrusted     bool      `json:"untrusted"`
}
