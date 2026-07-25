package impl

import (
	"context"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	mysqlclient "error-logging/pkg/client/mysql"

	"gorm.io/gorm"
)

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(c *mysqlclient.Client) repository.ProjectRepository {
	return &projectRepository{db: c.DB}
}

func (r *projectRepository) Create(ctx context.Context, project *dbdto.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *projectRepository) GetByID(ctx context.Context, id uint64) (*dbdto.Project, error) {
	var project dbdto.Project
	if err := r.db.WithContext(ctx).First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) ListByOrg(ctx context.Context, orgID uint64) ([]dbdto.Project, error) {
	var projects []dbdto.Project
	if err := r.db.WithContext(ctx).Where("org_id = ?", orgID).Order("id DESC").Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}
