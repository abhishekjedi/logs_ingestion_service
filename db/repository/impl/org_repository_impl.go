package impl

import (
	"context"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	mysqlclient "error-logging/pkg/client/mysql"

	"gorm.io/gorm"
)

type orgRepository struct {
	db *gorm.DB
}

func NewOrgRepository(c *mysqlclient.Client) repository.OrgRepository {
	return &orgRepository{db: c.DB}
}

func (r *orgRepository) Create(ctx context.Context, org *dbdto.Organization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

func (r *orgRepository) GetByID(ctx context.Context, id uint64) (*dbdto.Organization, error) {
	var org dbdto.Organization
	if err := r.db.WithContext(ctx).First(&org, id).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *orgRepository) ListByIDs(ctx context.Context, ids []uint64) ([]dbdto.Organization, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var orgs []dbdto.Organization
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Order("id DESC").Find(&orgs).Error
	return orgs, err
}
