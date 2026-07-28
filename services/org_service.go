package services

import (
	"context"

	"error-logging/constants"
	dbdto "error-logging/db/dto"
)

type OrgService interface {
	CreateOrg(ctx context.Context, userID uint64, name string) (*dbdto.Organization, error)
	ListMyOrgs(ctx context.Context, userID uint64) ([]dbdto.Organization, error)
	InviteMember(ctx context.Context, actorID, orgID uint64, email string, role constants.OrgRole) (*dbdto.OrganizationMember, error)
	ListMembers(ctx context.Context, actorID, orgID uint64) ([]dbdto.OrganizationMember, error)
	RemoveMember(ctx context.Context, actorID, orgID, memberID uint64) error
}
