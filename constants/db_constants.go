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

// OrgRole is a member's role within an organization.
type OrgRole string

const (
	RoleOwner  OrgRole = "owner"  // created the org; full control incl. delete
	RoleAdmin  OrgRole = "admin"  // manage members + projects
	RoleMember OrgRole = "member" // view + create projects
)

// CanManageMembers reports whether a role may invite/remove members.
func (r OrgRole) CanManageMembers() bool {
	return r == RoleOwner || r == RoleAdmin
}

// MemberStatus tracks whether an invite has been accepted (via first login).
type MemberStatus string

const (
	MemberActive  MemberStatus = "active"  // linked to a user
	MemberPending MemberStatus = "pending" // invited by email, not yet logged in
)
