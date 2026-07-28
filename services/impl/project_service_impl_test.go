package impl

import (
	"context"
	"errors"
	"testing"

	dbdto "error-logging/db/dto"
	repomock "error-logging/db/repository/mock"
	"error-logging/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProjectService_CreateProject(t *testing.T) {
	repo := new(repomock.ProjectRepository)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*dto.Project")).
		Return(nil).
		Run(func(args mock.Arguments) {
			args.Get(1).(*dbdto.Project).ID = 42
		})

	svc := NewProjectService(repo)

	got, err := svc.CreateProject(context.Background(), 3, 7, dto.CreateProjectRequest{Name: "checkout"})

	require.NoError(t, err)
	assert.Equal(t, uint64(42), got.ID)
	assert.Equal(t, "checkout", got.Name)
	require.NotNil(t, got.OrgID)
	assert.Equal(t, uint64(3), *got.OrgID)
	require.NotNil(t, got.OwnerID)
	assert.Equal(t, uint64(7), *got.OwnerID)
	repo.AssertExpectations(t)
}

func TestProjectService_CreateProject_RepoError(t *testing.T) {
	repo := new(repomock.ProjectRepository)
	repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db down"))

	svc := NewProjectService(repo)

	got, err := svc.CreateProject(context.Background(), 3, 7, dto.CreateProjectRequest{Name: "checkout"})

	require.Error(t, err)
	assert.Nil(t, got)
	repo.AssertExpectations(t)
}

func TestProjectService_GetProject(t *testing.T) {
	repo := new(repomock.ProjectRepository)
	want := &dbdto.Project{ID: 1, Name: "api"}
	repo.On("GetByID", mock.Anything, uint64(1)).Return(want, nil)

	svc := NewProjectService(repo)

	got, err := svc.GetProject(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, want, got)
	repo.AssertExpectations(t)
}

func TestProjectService_GetProject_NotFound(t *testing.T) {
	repo := new(repomock.ProjectRepository)
	repo.On("GetByID", mock.Anything, uint64(99)).Return(nil, errors.New("not found"))

	svc := NewProjectService(repo)

	got, err := svc.GetProject(context.Background(), 99)

	require.Error(t, err)
	assert.Nil(t, got)
	repo.AssertExpectations(t)
}

func TestProjectService_ListProjects(t *testing.T) {
	repo := new(repomock.ProjectRepository)
	want := []dbdto.Project{{ID: 1}, {ID: 2}}
	repo.On("ListByOrg", mock.Anything, uint64(3)).Return(want, nil)

	svc := NewProjectService(repo)

	got, err := svc.ListProjects(context.Background(), 3)

	require.NoError(t, err)
	assert.Len(t, got, 2)
	repo.AssertExpectations(t)
}
