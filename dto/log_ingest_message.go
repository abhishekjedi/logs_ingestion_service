package dto

import (
	"encoding/json"
	"time"
)

// LogIngestMessage is the envelope produced to Kafka by the ingest API and
// consumed by the worker. It carries the raw OTLP logs payload plus the resolved
// tenant context, so the worker never re-authenticates. One message per ingest
// request (a per-service OTLP batch); Payload is parsed into records by the worker.
type LogIngestMessage struct {
	ServiceID  uint64          `json:"service_id"`
	ProjectID  uint64          `json:"project_id"`
	ReceivedAt time.Time       `json:"received_at"`
	Payload    json.RawMessage `json:"payload"`
}
