package impl

import (
	"context"
	"testing"

	"error-logging/constants"
	dbdto "error-logging/db/dto"
	repomock "error-logging/db/repository/mock"
	"error-logging/services"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthzServiceRequireProjectManage(t *testing.T) {
	orgID := uint64(3)
	projects := new(repomock.ProjectRepository)
	members := new(repomock.OrgMemberRepository)
	projects.On("GetByID", mock.Anything, uint64(9)).Return(&dbdto.Project{ID: 9, OrgID: &orgID}, nil)
	members.On("GetMembership", mock.Anything, orgID, uint64(5)).Return(&dbdto.OrganizationMember{
		OrgID:  orgID,
		UserID: pointer(uint64(5)),
		Role:   constants.RoleAdmin,
	}, nil)

	authz := NewAuthzService(members, projects, nil, nil)
	require.NoError(t, authz.RequireProjectManage(context.Background(), 5, 9))
}

func TestAuthzServiceRequireProjectManageRejectsMember(t *testing.T) {
	orgID := uint64(3)
	projects := new(repomock.ProjectRepository)
	members := new(repomock.OrgMemberRepository)
	projects.On("GetByID", mock.Anything, uint64(9)).Return(&dbdto.Project{ID: 9, OrgID: &orgID}, nil)
	members.On("GetMembership", mock.Anything, orgID, uint64(5)).Return(&dbdto.OrganizationMember{
		OrgID:  orgID,
		UserID: pointer(uint64(5)),
		Role:   constants.RoleMember,
	}, nil)

	authz := NewAuthzService(members, projects, nil, nil)
	require.ErrorIs(t, authz.RequireProjectManage(context.Background(), 5, 9), services.ErrForbidden)
}

func pointer[T any](value T) *T {
	return &value
}
