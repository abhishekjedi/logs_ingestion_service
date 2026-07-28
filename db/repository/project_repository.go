package repository

import (
	"context"

	dbdto "error-logging/db/dto"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *dbdto.Project) error
	GetByID(ctx context.Context, id uint64) (*dbdto.Project, error)
	ListByOrg(ctx context.Context, orgID uint64) ([]dbdto.Project, error)
}
