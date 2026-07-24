package impl

import (
	"context"
	"time"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	"error-logging/services"
)

type issueService struct {
	issues    repository.IssueRepository
	analytics repository.AnalyticsRepository
}

func NewIssueService(
	issues repository.IssueRepository,
	analytics repository.AnalyticsRepository,
) services.IssueService {
	return &issueService{issues: issues, analytics: analytics}
}

func (s *issueService) ListIssues(ctx context.Context, filter repository.IssueListFilter) (services.IssueListResult, error) {
	issues, total, err := s.issues.List(ctx, filter)
	if err != nil {
		return services.IssueListResult{}, err
	}
	return services.IssueListResult{Issues: issues, Total: total}, nil
}

func (s *issueService) GetIssue(ctx context.Context, id uint64) (*dbdto.Issue, error) {
	return s.issues.GetByID(ctx, id)
}

func (s *issueService) GetTimeseries(ctx context.Context, issueID uint64, from, to time.Time) ([]repository.TimePoint, error) {
	return s.analytics.IssueTimeseries(ctx, issueID, from, to)
}

func (s *issueService) GetEvents(ctx context.Context, issueID uint64, limit int) ([]repository.ErrorEventDetail, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.analytics.RecentErrorEvents(ctx, issueID, limit)
}
