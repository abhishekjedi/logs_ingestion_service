package dto

import "time"

type IssueStatsAggregate struct {
	ServiceID  uint64
	IssueID    uint64
	EventCount uint64
	Users      uint64
	Sessions   uint64
	LastSeen   time.Time
}

type TimePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Events    uint64    `json:"events"`
	Users     uint64    `json:"users"`
}
