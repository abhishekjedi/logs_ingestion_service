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

func (m *IssueRepository) UpdateStatsBatch(ctx context.Context, updates []repository.IssueStatsUpdate) error {
	return m.Called(ctx, updates).Error(0)
}

func (m *IssueRepository) List(ctx context.Context, filter repository.IssueListFilter) ([]dbdto.Issue, int64, error) {
	args := m.Called(ctx, filter)
	var issues []dbdto.Issue
	if v := args.Get(0); v != nil {
		issues = v.([]dbdto.Issue)
	}
	return issues, int64(args.Int(1)), args.Error(2)
}

func (m *IssueRepository) GetByID(ctx context.Context, id uint64) (*dbdto.Issue, error) {
	args := m.Called(ctx, id)
	var issue *dbdto.Issue
	if v := args.Get(0); v != nil {
		issue = v.(*dbdto.Issue)
	}
	return issue, args.Error(1)
}
