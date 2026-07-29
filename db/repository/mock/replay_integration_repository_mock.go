package mock

import (
	"context"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"

	"github.com/stretchr/testify/mock"
)

type ReplayIntegrationRepository struct {
	mock.Mock
}

var _ repository.ReplayIntegrationRepository = (*ReplayIntegrationRepository)(nil)

func (m *ReplayIntegrationRepository) Upsert(ctx context.Context, integration *dbdto.ReplayIntegration) error {
	return m.Called(ctx, integration).Error(0)
}

func (m *ReplayIntegrationRepository) GetByProjectAndKey(
	ctx context.Context,
	projectID uint64,
	provider, externalProjectKey string,
) (*dbdto.ReplayIntegration, error) {
	args := m.Called(ctx, projectID, provider, externalProjectKey)
	var integration *dbdto.ReplayIntegration
	if value := args.Get(0); value != nil {
		integration = value.(*dbdto.ReplayIntegration)
	}
	return integration, args.Error(1)
}

func (m *ReplayIntegrationRepository) ListByProject(
	ctx context.Context,
	projectID uint64,
	provider string,
) ([]dbdto.ReplayIntegration, error) {
	args := m.Called(ctx, projectID, provider)
	var integrations []dbdto.ReplayIntegration
	if value := args.Get(0); value != nil {
		integrations = value.([]dbdto.ReplayIntegration)
	}
	return integrations, args.Error(1)
}

func (m *ReplayIntegrationRepository) DeleteByProjectAndKey(
	ctx context.Context,
	projectID uint64,
	provider, externalProjectKey string,
) error {
	return m.Called(ctx, projectID, provider, externalProjectKey).Error(0)
}
