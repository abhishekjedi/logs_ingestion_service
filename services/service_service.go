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
	// AuthenticateKey resolves a raw API key to its service. It checks a Redis cache
	// first (keyed by the key hash) and falls back to MySQL, so the ingest hot path
	// avoids a DB hit per request. Returns an error if the key is unknown.
	AuthenticateKey(ctx context.Context, rawKey string) (*dbdto.Service, error)
}
