package dto

import (
	"encoding/json"
	"time"
)

type LogIngestMessage struct {
	ServiceID  uint64          `json:"service_id"`
	ProjectID  uint64          `json:"project_id"`
	ReceivedAt time.Time       `json:"received_at"`
	Payload    json.RawMessage `json:"payload"`
}
