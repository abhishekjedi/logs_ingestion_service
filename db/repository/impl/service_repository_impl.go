package impl

import (
	"context"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	mysqlclient "error-logging/pkg/client/mysql"

	"gorm.io/gorm"
)

type serviceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(c *mysqlclient.Client) repository.ServiceRepository {
	return &serviceRepository{db: c.DB}
}

func (r *serviceRepository) Create(ctx context.Context, service *dbdto.Service) error {
	return r.db.WithContext(ctx).Create(service).Error
}

func (r *serviceRepository) GetByID(ctx context.Context, id uint64) (*dbdto.Service, error) {
	var service dbdto.Service
	if err := r.db.WithContext(ctx).First(&service, id).Error; err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *serviceRepository) GetByPublicID(ctx context.Context, publicID string) (*dbdto.Service, error) {
	var service dbdto.Service
	if err := r.db.WithContext(ctx).Where("public_id = ?", publicID).First(&service).Error; err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *serviceRepository) GetByAPIKeyHash(ctx context.Context, apiKeyHash string) (*dbdto.Service, error) {
	var service dbdto.Service
	if err := r.db.WithContext(ctx).Where("api_key_hash = ?", apiKeyHash).First(&service).Error; err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *serviceRepository) ListByProject(ctx context.Context, projectID uint64) ([]dbdto.Service, error) {
	var services []dbdto.Service
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("id DESC").Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}
