package repository

import (
	"context"

	dbdto "error-logging/db/dto"
)

type OrgRepository interface {
	Create(ctx context.Context, org *dbdto.Organization) error
	GetByID(ctx context.Context, id uint64) (*dbdto.Organization, error)
	ListByIDs(ctx context.Context, ids []uint64) ([]dbdto.Organization, error)
}
