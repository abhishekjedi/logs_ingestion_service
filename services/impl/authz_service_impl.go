package impl

import (
	"context"

	"error-logging/constants"
	"error-logging/db/repository"
	"error-logging/services"
)

type authzService struct {
	members  repository.OrgMemberRepository
	projects repository.ProjectRepository
	svcs     repository.ServiceRepository
	issues   repository.IssueRepository
}

func NewAuthzService(
	members repository.OrgMemberRepository,
	projects repository.ProjectRepository,
	svcs repository.ServiceRepository,
	issues repository.IssueRepository,
) services.AuthzService {
	return &authzService{members: members, projects: projects, svcs: svcs, issues: issues}
}

func (a *authzService) OrgRole(ctx context.Context, userID, orgID uint64) (constants.OrgRole, error) {
	m, err := a.members.GetMembership(ctx, orgID, userID)
	if err != nil {
		return "", services.ErrForbidden
	}
	return m.Role, nil
}

func (a *authzService) RequireMember(ctx context.Context, userID, orgID uint64) error {
	if _, err := a.members.GetMembership(ctx, orgID, userID); err != nil {
		return services.ErrForbidden
	}
	return nil
}

func (a *authzService) RequireManage(ctx context.Context, userID, orgID uint64) error {
	m, err := a.members.GetMembership(ctx, orgID, userID)
	if err != nil {
		return services.ErrForbidden
	}
	if !m.Role.CanManageMembers() {
		return services.ErrForbidden
	}
	return nil
}

func (a *authzService) RequireProjectAccess(ctx context.Context, userID, projectID uint64) error {
	proj, err := a.projects.GetByID(ctx, projectID)
	if err != nil {
		return services.ErrNotFound
	}
	if proj.OrgID == nil {
		return services.ErrForbidden
	}
	return a.RequireMember(ctx, userID, *proj.OrgID)
}

func (a *authzService) RequireProjectManage(ctx context.Context, userID, projectID uint64) error {
	proj, err := a.projects.GetByID(ctx, projectID)
	if err != nil {
		return services.ErrNotFound
	}
	if proj.OrgID == nil {
		return services.ErrForbidden
	}
	return a.RequireManage(ctx, userID, *proj.OrgID)
}

func (a *authzService) RequireServiceAccess(ctx context.Context, userID, serviceID uint64) error {
	svc, err := a.svcs.GetByID(ctx, serviceID)
	if err != nil {
		return services.ErrNotFound
	}
	return a.RequireProjectAccess(ctx, userID, svc.ProjectID)
}

func (a *authzService) RequireIssueAccess(ctx context.Context, userID, issueID uint64) error {
	issue, err := a.issues.GetByID(ctx, issueID)
	if err != nil {
		return services.ErrNotFound
	}
	return a.RequireServiceAccess(ctx, userID, issue.ServiceID)
}
