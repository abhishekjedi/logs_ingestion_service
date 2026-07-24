package mock

import (
	"context"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"

	"github.com/stretchr/testify/mock"
)

// IssueRepository is a mock of repository.IssueRepository.
type IssueRepository struct {
	mock.Mock
}

var _ repository.IssueRepository = (*IssueRepository)(nil)

func (m *IssueRepository) ResolveOrCreate(ctx context.Context, issue *dbdto.Issue) (*dbdto.Issue, bool, error) {
	args := m.Called(ctx, issue)
	var out *dbdto.Issue
	if v := args.Get(0); v != nil {
		out = v.(*dbdto.Issue)
	}
	return out, args.Bool(1), args.Error(2)
}

func (m *IssueRepository) MarkRegressed(ctx context.Context, id uint64) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}
