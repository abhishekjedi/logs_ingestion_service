package impl

import (
	"context"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	mysqlclient "error-logging/pkg/client/mysql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type replayIntegrationRepository struct {
	db *gorm.DB
}

func NewReplayIntegrationRepository(c *mysqlclient.Client) repository.ReplayIntegrationRepository {
	return &replayIntegrationRepository{db: c.DB}
}

func (r *replayIntegrationRepository) Upsert(ctx context.Context, integration *dbdto.ReplayIntegration) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "project_id"},
				{Name: "provider"},
				{Name: "external_project_key"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"api_base_url",
				"ingest_point",
				"api_key_ciphertext",
				"enabled",
				"last_validated_at",
				"last_error",
				"updated_at",
			}),
		}).
		Create(integration).Error
}

func (r *replayIntegrationRepository) GetByProjectAndKey(
	ctx context.Context,
	projectID uint64,
	provider, externalProjectKey string,
) (*dbdto.ReplayIntegration, error) {
	var integration dbdto.ReplayIntegration
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND provider = ? AND external_project_key = ?", projectID, provider, externalProjectKey).
		First(&integration).Error
	if err != nil {
		return nil, err
	}
	return &integration, nil
}

func (r *replayIntegrationRepository) ListByProject(
	ctx context.Context,
	projectID uint64,
	provider string,
) ([]dbdto.ReplayIntegration, error) {
	var integrations []dbdto.ReplayIntegration
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND provider = ?", projectID, provider).
		Order("id ASC").
		Find(&integrations).Error
	return integrations, err
}

func (r *replayIntegrationRepository) DeleteByProjectAndKey(
	ctx context.Context,
	projectID uint64,
	provider, externalProjectKey string,
) error {
	result := r.db.WithContext(ctx).
		Where("project_id = ? AND provider = ? AND external_project_key = ?", projectID, provider, externalProjectKey).
		Delete(&dbdto.ReplayIntegration{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
