package services

import (
	"context"

	"error-logging/constants"
)

type AuthzService interface {
	OrgRole(ctx context.Context, userID, orgID uint64) (constants.OrgRole, error)

	RequireMember(ctx context.Context, userID, orgID uint64) error

	RequireManage(ctx context.Context, userID, orgID uint64) error

	RequireProjectAccess(ctx context.Context, userID, projectID uint64) error
	RequireProjectManage(ctx context.Context, userID, projectID uint64) error
	RequireServiceAccess(ctx context.Context, userID, serviceID uint64) error
	RequireIssueAccess(ctx context.Context, userID, issueID uint64) error
}
