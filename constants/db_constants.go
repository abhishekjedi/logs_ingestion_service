package constants

type IssueLevel string

const (
	LevelDebug   IssueLevel = "debug"
	LevelInfo    IssueLevel = "info"
	LevelWarning IssueLevel = "warning"
	LevelError   IssueLevel = "error"
	LevelFatal   IssueLevel = "fatal"
)

type IssueStatus string

const (
	StatusUnresolved IssueStatus = "unresolved"
	StatusResolved   IssueStatus = "resolved"
	StatusIgnored    IssueStatus = "ignored"
	StatusRegressed  IssueStatus = "regressed"
)

type OrgRole string

const (
	RoleOwner  OrgRole = "owner"
	RoleAdmin  OrgRole = "admin"
	RoleMember OrgRole = "member"
)

func (r OrgRole) CanManageMembers() bool {
	return r == RoleOwner || r == RoleAdmin
}

type MemberStatus string

const (
	MemberActive  MemberStatus = "active"
	MemberPending MemberStatus = "pending"
)

const ReplayProviderOpenReplay = "openreplay"
