package mock

import (
	"context"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"

	"github.com/stretchr/testify/mock"
)

type OrgMemberRepository struct {
	mock.Mock
}

var _ repository.OrgMemberRepository = (*OrgMemberRepository)(nil)

func (m *OrgMemberRepository) Create(ctx context.Context, member *dbdto.OrganizationMember) error {
	return m.Called(ctx, member).Error(0)
}

func (m *OrgMemberRepository) ActivateInvites(ctx context.Context, userID uint64, email string) error {
	return m.Called(ctx, userID, email).Error(0)
}

func (m *OrgMemberRepository) GetMembership(
	ctx context.Context,
	orgID, userID uint64,
) (*dbdto.OrganizationMember, error) {
	args := m.Called(ctx, orgID, userID)
	var member *dbdto.OrganizationMember
	if value := args.Get(0); value != nil {
		member = value.(*dbdto.OrganizationMember)
	}
	return member, args.Error(1)
}

func (m *OrgMemberRepository) ListOrgIDsByUser(ctx context.Context, userID uint64) ([]uint64, error) {
	args := m.Called(ctx, userID)
	var ids []uint64
	if value := args.Get(0); value != nil {
		ids = value.([]uint64)
	}
	return ids, args.Error(1)
}

func (m *OrgMemberRepository) ListByOrg(ctx context.Context, orgID uint64) ([]dbdto.OrganizationMember, error) {
	args := m.Called(ctx, orgID)
	var members []dbdto.OrganizationMember
	if value := args.Get(0); value != nil {
		members = value.([]dbdto.OrganizationMember)
	}
	return members, args.Error(1)
}

func (m *OrgMemberRepository) GetByOrgEmail(
	ctx context.Context,
	orgID uint64,
	email string,
) (*dbdto.OrganizationMember, error) {
	args := m.Called(ctx, orgID, email)
	var member *dbdto.OrganizationMember
	if value := args.Get(0); value != nil {
		member = value.(*dbdto.OrganizationMember)
	}
	return member, args.Error(1)
}

func (m *OrgMemberRepository) RemoveByID(ctx context.Context, orgID, memberID uint64) error {
	return m.Called(ctx, orgID, memberID).Error(0)
}
