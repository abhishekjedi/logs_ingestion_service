package repository

import (
	"context"

	dbdto "error-logging/db/dto"
)

type ServiceRepository interface {
	Create(ctx context.Context, service *dbdto.Service) error
	GetByID(ctx context.Context, id uint64) (*dbdto.Service, error)
	GetByPublicID(ctx context.Context, publicID string) (*dbdto.Service, error)
	GetByAPIKeyHash(ctx context.Context, apiKeyHash string) (*dbdto.Service, error)
	ListByProject(ctx context.Context, projectID uint64) ([]dbdto.Service, error)
}
