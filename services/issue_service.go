package services

import (
	"context"
	"time"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
)

// IssueListResult is a page of issues plus the total matching count.
type IssueListResult struct {
	Issues []dbdto.Issue
	Total  int64
}

// IssueService serves the issue read API: list/detail from MySQL, trend and
// full-fidelity events from ClickHouse.
type IssueService interface {
	ListIssues(ctx context.Context, filter repository.IssueListFilter) (IssueListResult, error)
	GetIssue(ctx context.Context, id uint64) (*dbdto.Issue, error)
	GetTimeseries(ctx context.Context, issueID uint64, from, to time.Time) ([]repository.TimePoint, error)
	GetEvents(ctx context.Context, issueID uint64, limit int) ([]repository.ErrorEventDetail, error)
}
