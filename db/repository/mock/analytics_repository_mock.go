package mock

import (
	"context"
	"time"

	"error-logging/db/repository"

	"github.com/stretchr/testify/mock"
)

// AnalyticsRepository is a mock of repository.AnalyticsRepository.
type AnalyticsRepository struct {
	mock.Mock
}

var _ repository.AnalyticsRepository = (*AnalyticsRepository)(nil)

func (m *AnalyticsRepository) IssueTimeseries(ctx context.Context, issueID uint64, from, to time.Time) ([]repository.TimePoint, error) {
	args := m.Called(ctx, issueID, from, to)
	var out []repository.TimePoint
	if v := args.Get(0); v != nil {
		out = v.([]repository.TimePoint)
	}
	return out, args.Error(1)
}

func (m *AnalyticsRepository) ServiceOverview(ctx context.Context, serviceID uint64, from, to time.Time) ([]repository.ServiceOverviewPoint, error) {
	args := m.Called(ctx, serviceID, from, to)
	var out []repository.ServiceOverviewPoint
	if v := args.Get(0); v != nil {
		out = v.([]repository.ServiceOverviewPoint)
	}
	return out, args.Error(1)
}

func (m *AnalyticsRepository) ReleaseHealth(ctx context.Context, serviceID uint64, from, to time.Time) ([]repository.ReleaseHealth, error) {
	args := m.Called(ctx, serviceID, from, to)
	var out []repository.ReleaseHealth
	if v := args.Get(0); v != nil {
		out = v.([]repository.ReleaseHealth)
	}
	return out, args.Error(1)
}

func (m *AnalyticsRepository) RecentErrorEvents(ctx context.Context, issueID uint64, limit int) ([]repository.ErrorEventDetail, error) {
	args := m.Called(ctx, issueID, limit)
	var out []repository.ErrorEventDetail
	if v := args.Get(0); v != nil {
		out = v.([]repository.ErrorEventDetail)
	}
	return out, args.Error(1)
}

func (m *AnalyticsRepository) Breadcrumbs(ctx context.Context, serviceID uint64, sessionID string, before time.Time, limit int) ([]repository.Breadcrumb, error) {
	args := m.Called(ctx, serviceID, sessionID, before, limit)
	var out []repository.Breadcrumb
	if v := args.Get(0); v != nil {
		out = v.([]repository.Breadcrumb)
	}
	return out, args.Error(1)
}
