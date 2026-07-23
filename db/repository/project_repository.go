// Package repository is the data-access layer: it owns all DB queries and is the
// only layer services call to read/write persisted state. Connection clients live
// in pkg/client; repositories wrap them and expose query methods.
package repository

import (
	"context"

	dbdto "error-logging/db/dto"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *dbdto.Project) error
	GetByID(ctx context.Context, id uint64) (*dbdto.Project, error)
	List(ctx context.Context) ([]dbdto.Project, error)
}
