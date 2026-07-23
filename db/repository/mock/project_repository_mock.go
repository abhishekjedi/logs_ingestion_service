// Package mock provides testify-based mock implementations of the repository
// interfaces for use in service-layer unit tests.
package mock

import (
	"context"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"

	"github.com/stretchr/testify/mock"
)

// ProjectRepository is a mock of repository.ProjectRepository.
type ProjectRepository struct {
	mock.Mock
}

var _ repository.ProjectRepository = (*ProjectRepository)(nil)

func (m *ProjectRepository) Create(ctx context.Context, project *dbdto.Project) error {
	return m.Called(ctx, project).Error(0)
}

func (m *ProjectRepository) GetByID(ctx context.Context, id uint64) (*dbdto.Project, error) {
	args := m.Called(ctx, id)
	var project *dbdto.Project
	if v := args.Get(0); v != nil {
		project = v.(*dbdto.Project)
	}
	return project, args.Error(1)
}

func (m *ProjectRepository) List(ctx context.Context) ([]dbdto.Project, error) {
	args := m.Called(ctx)
	var projects []dbdto.Project
	if v := args.Get(0); v != nil {
		projects = v.([]dbdto.Project)
	}
	return projects, args.Error(1)
}
