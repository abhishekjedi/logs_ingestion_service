// Package constants holds shared enum-like constants used across the persistence
// layer and services.
package constants

// IssueLevel mirrors OTel severity buckets, collapsed to the levels we group on.
type IssueLevel string

const (
	LevelDebug   IssueLevel = "debug"
	LevelInfo    IssueLevel = "info"
	LevelWarning IssueLevel = "warning"
	LevelError   IssueLevel = "error"
	LevelFatal   IssueLevel = "fatal"
)

// IssueStatus is the lifecycle of a grouped issue.
type IssueStatus string

const (
	StatusUnresolved IssueStatus = "unresolved"
	StatusResolved   IssueStatus = "resolved"
	StatusIgnored    IssueStatus = "ignored"
	StatusRegressed  IssueStatus = "regressed"
)
