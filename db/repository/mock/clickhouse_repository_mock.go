package mock

import (
	"context"

	"error-logging/db/repository"

	"github.com/stretchr/testify/mock"
)

// LogRepository is a mock of repository.LogRepository.
type LogRepository struct {
	mock.Mock
}

var _ repository.LogRepository = (*LogRepository)(nil)

func (m *LogRepository) InsertBatch(ctx context.Context, rows []repository.LogRow) error {
	return m.Called(ctx, rows).Error(0)
}

// ErrorEventRepository is a mock of repository.ErrorEventRepository.
type ErrorEventRepository struct {
	mock.Mock
}

var _ repository.ErrorEventRepository = (*ErrorEventRepository)(nil)

func (m *ErrorEventRepository) InsertBatch(ctx context.Context, rows []repository.ErrorEventRow) error {
	return m.Called(ctx, rows).Error(0)
}
