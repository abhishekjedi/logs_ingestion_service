package services

import (
	"context"

	dbdto "error-logging/db/dto"
	"error-logging/dto"
)

type ServiceService interface {
	CreateService(ctx context.Context, projectID uint64, req dto.CreateServiceRequest) (*dto.CreateServiceResponse, error)
	ListServices(ctx context.Context, projectID uint64) ([]dbdto.Service, error)

	AuthenticateKey(ctx context.Context, rawKey string) (*dbdto.Service, error)
}
