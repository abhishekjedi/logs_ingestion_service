package services

import (
	"context"

	dbdto "error-logging/db/dto"
	"error-logging/dto"
)

// ServiceService manages registered services (the ingest endpoints under a project).
type ServiceService interface {
	// CreateService registers a service under a project, minting its public id and
	// API key. The returned response carries the raw key exactly once.
	CreateService(ctx context.Context, projectID uint64, req dto.CreateServiceRequest) (*dto.CreateServiceResponse, error)
	ListServices(ctx context.Context, projectID uint64) ([]dbdto.Service, error)
}
