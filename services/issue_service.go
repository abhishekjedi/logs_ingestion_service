package services

import (
	"context"
	"time"

	dbdto "error-logging/db/dto"
	"error-logging/dto"
)

type IssueService interface {
	ListIssues(ctx context.Context, filter dto.IssueListFilter) (dto.IssueListResult, error)
	GetIssue(ctx context.Context, id uint64) (*dbdto.Issue, error)
	GetTimeseries(ctx context.Context, issueID uint64, from, to time.Time) ([]dbdto.TimePoint, error)
	GetEvents(ctx context.Context, issueID uint64, limit int) ([]dbdto.ErrorEventDetail, error)
}
