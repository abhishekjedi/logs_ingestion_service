package repository

import (
	"context"

	dbdto "error-logging/db/dto"
)

type ReplayIntegrationRepository interface {
	Upsert(ctx context.Context, integration *dbdto.ReplayIntegration) error
	GetByProjectAndKey(ctx context.Context, projectID uint64, provider, externalProjectKey string) (*dbdto.ReplayIntegration, error)
	ListByProject(ctx context.Context, projectID uint64, provider string) ([]dbdto.ReplayIntegration, error)
	DeleteByProjectAndKey(ctx context.Context, projectID uint64, provider, externalProjectKey string) error
}
