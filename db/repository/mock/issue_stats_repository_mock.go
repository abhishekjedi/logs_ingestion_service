package mock

import (
	"context"
	"time"

	"error-logging/db/repository"

	"github.com/stretchr/testify/mock"
)

// IssueStatsRepository is a mock of repository.IssueStatsRepository.
type IssueStatsRepository struct {
	mock.Mock
}

var _ repository.IssueStatsRepository = (*IssueStatsRepository)(nil)

func (m *IssueStatsRepository) ActiveSince(ctx context.Context, since time.Time) ([]repository.IssueStatsAggregate, error) {
	args := m.Called(ctx, since)
	var out []repository.IssueStatsAggregate
	if v := args.Get(0); v != nil {
		out = v.([]repository.IssueStatsAggregate)
	}
	return out, args.Error(1)
}
