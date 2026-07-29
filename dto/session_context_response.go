package dto

import "time"

type SessionContextResponse struct {
	Status          string                `json:"status"`
	SessionID       string                `json:"session_id,omitempty"`
	ReplayURL       string                `json:"replay_url,omitempty"`
	FocusedAt       time.Time             `json:"focused_at"`
	Journey         []SessionContextEvent `json:"journey"`
	NetworkFailures []SessionContextEvent `json:"network_failures"`
	ConsoleErrors   []SessionContextEvent `json:"console_errors"`
	Exceptions      []SessionContextEvent `json:"exceptions"`
	Counts          map[string]int        `json:"counts"`
	Truncated       bool                  `json:"truncated"`
}
