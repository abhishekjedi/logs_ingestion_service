package mock

import (
	"context"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"

	"github.com/stretchr/testify/mock"
)

type LogRepository struct {
	mock.Mock
}

var _ repository.LogRepository = (*LogRepository)(nil)

func (m *LogRepository) InsertBatch(ctx context.Context, rows []dbdto.LogRow) error {
	return m.Called(ctx, rows).Error(0)
}

type ErrorEventRepository struct {
	mock.Mock
}

var _ repository.ErrorEventRepository = (*ErrorEventRepository)(nil)

func (m *ErrorEventRepository) InsertBatch(ctx context.Context, rows []dbdto.ErrorEventRow) error {
	return m.Called(ctx, rows).Error(0)
}
