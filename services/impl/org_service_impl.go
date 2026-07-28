package impl

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"error-logging/constants"
	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	"error-logging/services"
)

type orgService struct {
	orgs    repository.OrgRepository
	members repository.OrgMemberRepository
	users   repository.UserRepository
	authz   services.AuthzService
}

func NewOrgService(
	orgs repository.OrgRepository,
	members repository.OrgMemberRepository,
	users repository.UserRepository,
	authz services.AuthzService,
) services.OrgService {
	return &orgService{orgs: orgs, members: members, users: users, authz: authz}
}

func (s *orgService) CreateOrg(ctx context.Context, userID uint64, name string) (*dbdto.Organization, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("resolve user: %w", err)
	}

	suffix, err := randomHex(3)
	if err != nil {
		return nil, err
	}
	org := &dbdto.Organization{
		Name:      name,
		Slug:      slugify(name) + "-" + suffix,
		CreatedBy: userID,
	}
	if err := s.orgs.Create(ctx, org); err != nil {
		return nil, fmt.Errorf("create org: %w", err)
	}

	owner := &dbdto.OrganizationMember{
		OrgID:  org.ID,
		UserID: &userID,
		Email:  user.Email,
		Role:   constants.RoleOwner,
		Status: constants.MemberActive,
	}
	if err := s.members.Create(ctx, owner); err != nil {
		return nil, fmt.Errorf("create owner membership: %w", err)
	}
	return org, nil
}

func (s *orgService) ListMyOrgs(ctx context.Context, userID uint64) ([]dbdto.Organization, error) {
	ids, err := s.members.ListOrgIDsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.orgs.ListByIDs(ctx, ids)
}

func (s *orgService) InviteMember(ctx context.Context, actorID, orgID uint64, email string, role constants.OrgRole) (*dbdto.OrganizationMember, error) {
	if err := s.authz.RequireManage(ctx, actorID, orgID); err != nil {
		return nil, err
	}
	if role != constants.RoleAdmin && role != constants.RoleMember {
		role = constants.RoleMember
	}
	if existing, err := s.members.GetByOrgEmail(ctx, orgID, email); err == nil && existing != nil {
		return nil, fmt.Errorf("%s is already a member or invited", email)
	}

	member := &dbdto.OrganizationMember{
		OrgID:  orgID,
		Email:  email,
		Role:   role,
		Status: constants.MemberPending,
	}
	if u, err := s.users.GetByEmail(ctx, email); err == nil && u != nil {
		member.UserID = &u.ID
		member.Status = constants.MemberActive
	}

	if err := s.members.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("create membership: %w", err)
	}
	return member, nil
}

func (s *orgService) ListMembers(ctx context.Context, actorID, orgID uint64) ([]dbdto.OrganizationMember, error) {
	if err := s.authz.RequireMember(ctx, actorID, orgID); err != nil {
		return nil, err
	}
	return s.members.ListByOrg(ctx, orgID)
}

func (s *orgService) RemoveMember(ctx context.Context, actorID, orgID, memberID uint64) error {
	if err := s.authz.RequireManage(ctx, actorID, orgID); err != nil {
		return err
	}
	return s.members.RemoveByID(ctx, orgID, memberID)
}

var slugPattern = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	s := slugPattern.ReplaceAllString(strings.ToLower(name), "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "org"
	}
	return s
}
