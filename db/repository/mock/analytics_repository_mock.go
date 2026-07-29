package mock

import (
	"context"
	"time"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"

	"github.com/stretchr/testify/mock"
)

type AnalyticsRepository struct {
	mock.Mock
}

var _ repository.AnalyticsRepository = (*AnalyticsRepository)(nil)

func (m *AnalyticsRepository) IssueTimeseries(ctx context.Context, issueID uint64, from, to time.Time) ([]dbdto.TimePoint, error) {
	args := m.Called(ctx, issueID, from, to)
	var out []dbdto.TimePoint
	if v := args.Get(0); v != nil {
		out = v.([]dbdto.TimePoint)
	}
	return out, args.Error(1)
}

func (m *AnalyticsRepository) ServiceOverview(ctx context.Context, serviceID uint64, from, to time.Time) ([]dbdto.ServiceOverviewPoint, error) {
	args := m.Called(ctx, serviceID, from, to)
	var out []dbdto.ServiceOverviewPoint
	if v := args.Get(0); v != nil {
		out = v.([]dbdto.ServiceOverviewPoint)
	}
	return out, args.Error(1)
}

func (m *AnalyticsRepository) ReleaseHealth(ctx context.Context, serviceID uint64, from, to time.Time) ([]dbdto.ReleaseHealth, error) {
	args := m.Called(ctx, serviceID, from, to)
	var out []dbdto.ReleaseHealth
	if v := args.Get(0); v != nil {
		out = v.([]dbdto.ReleaseHealth)
	}
	return out, args.Error(1)
}

func (m *AnalyticsRepository) RecentErrorEvents(ctx context.Context, issueID uint64, limit int) ([]dbdto.ErrorEventDetail, error) {
	args := m.Called(ctx, issueID, limit)
	var out []dbdto.ErrorEventDetail
	if v := args.Get(0); v != nil {
		out = v.([]dbdto.ErrorEventDetail)
	}
	return out, args.Error(1)
}

func (m *AnalyticsRepository) GetErrorEvent(
	ctx context.Context,
	issueID uint64,
	eventID string,
) (*dbdto.ErrorEventDetail, error) {
	args := m.Called(ctx, issueID, eventID)
	var event *dbdto.ErrorEventDetail
	if value := args.Get(0); value != nil {
		event = value.(*dbdto.ErrorEventDetail)
	}
	return event, args.Error(1)
}

func (m *AnalyticsRepository) Breadcrumbs(ctx context.Context, serviceID uint64, sessionID string, before time.Time, limit int) ([]dbdto.Breadcrumb, error) {
	args := m.Called(ctx, serviceID, sessionID, before, limit)
	var out []dbdto.Breadcrumb
	if v := args.Get(0); v != nil {
		out = v.([]dbdto.Breadcrumb)
	}
	return out, args.Error(1)
}
