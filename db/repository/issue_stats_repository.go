package repository

import (
	"context"
	"time"

	dbdto "error-logging/db/dto"
)

type IssueStatsRepository interface {
	ActiveSince(ctx context.Context, since time.Time) ([]dbdto.IssueStatsAggregate, error)
}
