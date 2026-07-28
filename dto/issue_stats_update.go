package dto

import "time"

type IssueStatsUpdate struct {
	IssueID          uint64
	EventCount       uint64
	AffectedUsers    uint64
	AffectedSessions uint64
	LastSeen         time.Time
}
