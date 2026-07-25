package services

import (
	"context"

	"error-logging/constants"
)

// AuthzService answers "may this user do this?" by resolving a resource up to its
// owning org and checking the user's membership/role. Returns ErrForbidden (not a
// member / insufficient role) or ErrNotFound (resource missing).
type AuthzService interface {
	// OrgRole returns the user's role in an org, or ErrForbidden if not a member.
	OrgRole(ctx context.Context, userID, orgID uint64) (constants.OrgRole, error)
	// RequireMember passes if the user belongs to the org.
	RequireMember(ctx context.Context, userID, orgID uint64) error
	// RequireManage passes if the user is owner/admin of the org.
	RequireManage(ctx context.Context, userID, orgID uint64) error
	// RequireProjectAccess/ServiceAccess/IssueAccess resolve the resource to its org
	// and require membership.
	RequireProjectAccess(ctx context.Context, userID, projectID uint64) error
	RequireServiceAccess(ctx context.Context, userID, serviceID uint64) error
	RequireIssueAccess(ctx context.Context, userID, issueID uint64) error
}
