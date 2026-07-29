package repository

import (
	"context"
	"time"

	dbdto "error-logging/db/dto"
)

type AnalyticsRepository interface {
	IssueTimeseries(ctx context.Context, issueID uint64, from, to time.Time) ([]dbdto.TimePoint, error)
	ServiceOverview(ctx context.Context, serviceID uint64, from, to time.Time) ([]dbdto.ServiceOverviewPoint, error)
	ReleaseHealth(ctx context.Context, serviceID uint64, from, to time.Time) ([]dbdto.ReleaseHealth, error)
	RecentErrorEvents(ctx context.Context, issueID uint64, limit int) ([]dbdto.ErrorEventDetail, error)
	GetErrorEvent(ctx context.Context, issueID uint64, eventID string) (*dbdto.ErrorEventDetail, error)
	Breadcrumbs(ctx context.Context, serviceID uint64, sessionID string, before time.Time, limit int) ([]dbdto.Breadcrumb, error)
}
