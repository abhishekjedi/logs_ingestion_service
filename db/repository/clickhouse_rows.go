package repository

import (
	"context"

	dbdto "error-logging/db/dto"
)

type LogRepository interface {
	InsertBatch(ctx context.Context, rows []dbdto.LogRow) error
}

type ErrorEventRepository interface {
	InsertBatch(ctx context.Context, rows []dbdto.ErrorEventRow) error
}
