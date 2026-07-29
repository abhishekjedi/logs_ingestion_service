package openreplay

type Event struct {
	EventID       string         `json:"event_id"`
	EventName     string         `json:"$event_name"`
	CreatedAt     int64          `json:"created_at"`
	SessionID     string         `json:"session_id"`
	Properties    map[string]any `json:"properties"`
	AltProperties map[string]any `json:"$properties"`
}
