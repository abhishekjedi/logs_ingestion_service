package mock

import (
	"context"
	"time"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"

	"github.com/stretchr/testify/mock"
)

type IssueStatsRepository struct {
	mock.Mock
}

var _ repository.IssueStatsRepository = (*IssueStatsRepository)(nil)

func (m *IssueStatsRepository) ActiveSince(ctx context.Context, since time.Time) ([]dbdto.IssueStatsAggregate, error) {
	args := m.Called(ctx, since)
	var out []dbdto.IssueStatsAggregate
	if v := args.Get(0); v != nil {
		out = v.([]dbdto.IssueStatsAggregate)
	}
	return out, args.Error(1)
}
