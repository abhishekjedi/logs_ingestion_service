package impl

import (
	"context"

	"error-logging/constants"
	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	mysqlclient "error-logging/pkg/client/mysql"

	"gorm.io/gorm"
)

type orgMemberRepository struct {
	db *gorm.DB
}

func NewOrgMemberRepository(c *mysqlclient.Client) repository.OrgMemberRepository {
	return &orgMemberRepository{db: c.DB}
}

func (r *orgMemberRepository) Create(ctx context.Context, member *dbdto.OrganizationMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *orgMemberRepository) ActivateInvites(ctx context.Context, userID uint64, email string) error {
	return r.db.WithContext(ctx).
		Model(&dbdto.OrganizationMember{}).
		Where("email = ? AND (user_id IS NULL OR status = ?)", email, constants.MemberPending).
		Updates(map[string]any{"user_id": userID, "status": constants.MemberActive}).Error
}

func (r *orgMemberRepository) GetMembership(ctx context.Context, orgID, userID uint64) (*dbdto.OrganizationMember, error) {
	var m dbdto.OrganizationMember
	err := r.db.WithContext(ctx).
		Where("org_id = ? AND user_id = ? AND status = ?", orgID, userID, constants.MemberActive).
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *orgMemberRepository) ListOrgIDsByUser(ctx context.Context, userID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).
		Model(&dbdto.OrganizationMember{}).
		Where("user_id = ? AND status = ?", userID, constants.MemberActive).
		Pluck("org_id", &ids).Error
	return ids, err
}

func (r *orgMemberRepository) ListByOrg(ctx context.Context, orgID uint64) ([]dbdto.OrganizationMember, error) {
	var members []dbdto.OrganizationMember
	err := r.db.WithContext(ctx).Where("org_id = ?", orgID).Order("id ASC").Find(&members).Error
	return members, err
}

func (r *orgMemberRepository) GetByOrgEmail(ctx context.Context, orgID uint64, email string) (*dbdto.OrganizationMember, error) {
	var m dbdto.OrganizationMember
	err := r.db.WithContext(ctx).Where("org_id = ? AND email = ?", orgID, email).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *orgMemberRepository) RemoveByID(ctx context.Context, orgID, memberID uint64) error {
	return r.db.WithContext(ctx).
		Where("org_id = ? AND id = ?", orgID, memberID).
		Delete(&dbdto.OrganizationMember{}).Error
}
