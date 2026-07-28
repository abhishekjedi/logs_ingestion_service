package repository

import (
	"context"

	dbdto "error-logging/db/dto"
)

type OrgMemberRepository interface {
	Create(ctx context.Context, member *dbdto.OrganizationMember) error

	ActivateInvites(ctx context.Context, userID uint64, email string) error

	GetMembership(ctx context.Context, orgID, userID uint64) (*dbdto.OrganizationMember, error)

	ListOrgIDsByUser(ctx context.Context, userID uint64) ([]uint64, error)
	ListByOrg(ctx context.Context, orgID uint64) ([]dbdto.OrganizationMember, error)
	GetByOrgEmail(ctx context.Context, orgID uint64, email string) (*dbdto.OrganizationMember, error)
	RemoveByID(ctx context.Context, orgID, memberID uint64) error
}
