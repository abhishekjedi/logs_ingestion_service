package repository

import (
	"context"

	dbdto "error-logging/db/dto"
	"error-logging/dto"
)

type IssueRepository interface {
	ResolveOrCreate(ctx context.Context, issue *dbdto.Issue) (*dbdto.Issue, bool, error)
	MarkRegressed(ctx context.Context, id uint64) (bool, error)
	UpdateStatsBatch(ctx context.Context, updates []dto.IssueStatsUpdate) error
	List(ctx context.Context, filter dto.IssueListFilter) ([]dbdto.Issue, int64, error)
	GetByID(ctx context.Context, id uint64) (*dbdto.Issue, error)
}
