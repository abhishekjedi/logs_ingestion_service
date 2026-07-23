package mock

import (
	"context"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"

	"github.com/stretchr/testify/mock"
)

// ServiceRepository is a mock of repository.ServiceRepository.
type ServiceRepository struct {
	mock.Mock
}

var _ repository.ServiceRepository = (*ServiceRepository)(nil)

func (m *ServiceRepository) Create(ctx context.Context, service *dbdto.Service) error {
	return m.Called(ctx, service).Error(0)
}

func (m *ServiceRepository) GetByID(ctx context.Context, id uint64) (*dbdto.Service, error) {
	return m.getService(m.Called(ctx, id))
}

func (m *ServiceRepository) GetByPublicID(ctx context.Context, publicID string) (*dbdto.Service, error) {
	return m.getService(m.Called(ctx, publicID))
}

func (m *ServiceRepository) GetByAPIKeyHash(ctx context.Context, apiKeyHash string) (*dbdto.Service, error) {
	return m.getService(m.Called(ctx, apiKeyHash))
}

func (m *ServiceRepository) ListByProject(ctx context.Context, projectID uint64) ([]dbdto.Service, error) {
	args := m.Called(ctx, projectID)
	var services []dbdto.Service
	if v := args.Get(0); v != nil {
		services = v.([]dbdto.Service)
	}
	return services, args.Error(1)
}

func (m *ServiceRepository) getService(args mock.Arguments) (*dbdto.Service, error) {
	var service *dbdto.Service
	if v := args.Get(0); v != nil {
		service = v.(*dbdto.Service)
	}
	return service, args.Error(1)
}
