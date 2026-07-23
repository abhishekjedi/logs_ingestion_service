package dto

import (
	"time"

	"error-logging/constants"
)

// Issue is a group of events sharing a fingerprint (Sentry-style). Counts and
// affected-* estimates are denormalized copies synced from ClickHouse issue_stats;
// the source of truth for those numbers is ClickHouse, not this row.
type Issue struct {
	ID                       uint64                `gorm:"primaryKey;autoIncrement" json:"id"`
	ServiceID                uint64                `gorm:"column:service_id" json:"service_id"`
	Fingerprint              string                `gorm:"column:fingerprint" json:"fingerprint"`
	Title                    string                `gorm:"column:title" json:"title"`
	Culprit                  string                `gorm:"column:culprit" json:"culprit"`
	Level                    constants.IssueLevel  `gorm:"column:level" json:"level"`
	Status                   constants.IssueStatus `gorm:"column:status" json:"status"`
	FirstSeen                time.Time             `gorm:"column:first_seen" json:"first_seen"`
	LastSeen                 time.Time             `gorm:"column:last_seen" json:"last_seen"`
	EventCount               uint64                `gorm:"column:event_count" json:"event_count"`
	AffectedUsersEstimate    uint64                `gorm:"column:affected_users_estimate" json:"affected_users_estimate"`
	AffectedSessionsEstimate uint64                `gorm:"column:affected_sessions_estimate" json:"affected_sessions_estimate"`
	RegressedAt              *time.Time            `gorm:"column:regressed_at" json:"regressed_at,omitempty"`
	Metadata                 *IssueMetadata        `gorm:"column:metadata;serializer:json" json:"metadata,omitempty"`
	CreatedAt                time.Time             `json:"created_at"`
	UpdatedAt                time.Time             `json:"updated_at"`
}

func (Issue) TableName() string { return "issues" }

// IssueMetadata caches a representative sample of the group (stored as JSON on the
// issues row) so list/detail views can render "what this issue is" without a
// ClickHouse round trip.
type IssueMetadata struct {
	ExceptionType    string       `json:"exception_type,omitempty"`
	ExceptionMessage string       `json:"exception_message,omitempty"`
	TopFrames        []StackFrame `json:"top_frames,omitempty"`
	SampleSessionID  string       `json:"sample_session_id,omitempty"`
}

// StackFrame is a single parsed stack frame; mirrors the ClickHouse stack_frames
// tuple shape used by error_events.
type StackFrame struct {
	File     string `json:"file"`
	Function string `json:"function"`
	Line     uint32 `json:"line"`
	Col      uint32 `json:"col"`
	InApp    bool   `json:"in_app"`
}
