package impl

import (
	"context"
	"errors"
	"strings"
	"testing"

	dbdto "error-logging/db/dto"
	repomock "error-logging/db/repository/mock"
	"error-logging/dto"
	"error-logging/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServiceService_CreateService(t *testing.T) {
	svcRepo := new(repomock.ServiceRepository)
	projRepo := new(repomock.ProjectRepository)

	projRepo.On("GetByID", mock.Anything, uint64(5)).Return(&dbdto.Project{ID: 5}, nil)
	svcRepo.On("Create", mock.Anything, mock.AnythingOfType("*dto.Service")).
		Return(nil).
		Run(func(args mock.Arguments) {
			args.Get(1).(*dbdto.Service).ID = 99
		})

	svc := NewServiceService(svcRepo, projRepo, nil, config.AppConfig{BaseURL: "http://localhost:8080"})

	resp, err := svc.CreateService(context.Background(), 5, dto.CreateServiceRequest{Name: "web"})

	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, uint64(99), resp.Service.ID)
	assert.Equal(t, uint64(5), resp.Service.ProjectID)
	assert.Equal(t, "web", resp.Service.Name)

	assert.True(t, strings.HasPrefix(resp.APIKey, "elk_live_"), "raw key uses scheme prefix")
	assert.Equal(t, hashAPIKey(resp.APIKey), resp.Service.APIKeyHash, "stored hash matches raw key")
	assert.NotEqual(t, resp.APIKey, resp.Service.APIKeyHash, "raw key is never stored")
	assert.True(t, strings.HasPrefix(resp.Service.APIKeyPrefix, "elk_live_"))
	assert.NotEmpty(t, resp.Service.PublicID)

	assert.Equal(t, "http://localhost:8080/api/v1/logs/"+resp.Service.PublicID, resp.IngestURL)

	svcRepo.AssertExpectations(t)
	projRepo.AssertExpectations(t)
}

func TestServiceService_CreateService_ProjectNotFound(t *testing.T) {
	svcRepo := new(repomock.ServiceRepository)
	projRepo := new(repomock.ProjectRepository)

	projRepo.On("GetByID", mock.Anything, uint64(404)).Return(nil, errors.New("not found"))

	svc := NewServiceService(svcRepo, projRepo, nil, config.AppConfig{BaseURL: "http://localhost:8080"})

	resp, err := svc.CreateService(context.Background(), 404, dto.CreateServiceRequest{Name: "web"})

	require.Error(t, err)
	assert.Nil(t, resp)

	svcRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	projRepo.AssertExpectations(t)
}

func TestServiceService_CreateService_CreateError(t *testing.T) {
	svcRepo := new(repomock.ServiceRepository)
	projRepo := new(repomock.ProjectRepository)

	projRepo.On("GetByID", mock.Anything, uint64(5)).Return(&dbdto.Project{ID: 5}, nil)
	svcRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("duplicate public_id"))

	svc := NewServiceService(svcRepo, projRepo, nil, config.AppConfig{BaseURL: "http://localhost:8080"})

	resp, err := svc.CreateService(context.Background(), 5, dto.CreateServiceRequest{Name: "web"})

	require.Error(t, err)
	assert.Nil(t, resp)
	svcRepo.AssertExpectations(t)
	projRepo.AssertExpectations(t)
}

func TestServiceService_ListServices(t *testing.T) {
	svcRepo := new(repomock.ServiceRepository)
	projRepo := new(repomock.ProjectRepository)

	want := []dbdto.Service{{ID: 1}, {ID: 2}}
	svcRepo.On("ListByProject", mock.Anything, uint64(5)).Return(want, nil)

	svc := NewServiceService(svcRepo, projRepo, nil, config.AppConfig{})

	got, err := svc.ListServices(context.Background(), 5)

	require.NoError(t, err)
	assert.Len(t, got, 2)
	svcRepo.AssertExpectations(t)
}

func TestServiceService_AuthenticateKey(t *testing.T) {
	svcRepo := new(repomock.ServiceRepository)
	rawKey := "elk_live_secret"
	want := &dbdto.Service{ID: 1, PublicID: "pub", APIKeyHash: hashAPIKey(rawKey)}

	svcRepo.On("GetByAPIKeyHash", mock.Anything, hashAPIKey(rawKey)).Return(want, nil)

	svc := NewServiceService(svcRepo, nil, nil, config.AppConfig{})

	got, err := svc.AuthenticateKey(context.Background(), rawKey)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	svcRepo.AssertExpectations(t)
}

func TestServiceService_AuthenticateKey_Unknown(t *testing.T) {
	svcRepo := new(repomock.ServiceRepository)
	svcRepo.On("GetByAPIKeyHash", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))

	svc := NewServiceService(svcRepo, nil, nil, config.AppConfig{})

	got, err := svc.AuthenticateKey(context.Background(), "elk_live_bogus")
	require.Error(t, err)
	assert.Nil(t, got)
}
