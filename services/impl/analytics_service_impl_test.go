package impl

import (
	"context"
	"testing"
	"time"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	repomock "error-logging/db/repository/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIssueService_ListIssues(t *testing.T) {
	issues := new(repomock.IssueRepository)
	filter := repository.IssueListFilter{ServiceID: 3, Limit: 50}
	issues.On("List", mock.Anything, filter).
		Return([]dbdto.Issue{{ID: 1}, {ID: 2}}, 2, nil)

	svc := NewIssueService(issues, new(repomock.AnalyticsRepository))
	res, err := svc.ListIssues(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, res.Issues, 2)
	assert.Equal(t, int64(2), res.Total)
}

func TestIssueService_GetTimeseries(t *testing.T) {
	analytics := new(repomock.AnalyticsRepository)
	analytics.On("IssueTimeseries", mock.Anything, uint64(5), mock.Anything, mock.Anything).
		Return([]repository.TimePoint{{Events: 100, Users: 10}}, nil)

	svc := NewIssueService(new(repomock.IssueRepository), analytics)
	pts, err := svc.GetTimeseries(context.Background(), 5, time.Now().Add(-time.Hour), time.Now())
	require.NoError(t, err)
	require.Len(t, pts, 1)
	assert.Equal(t, uint64(100), pts[0].Events)
}

func TestIssueService_GetEvents_ClampsLimit(t *testing.T) {
	analytics := new(repomock.AnalyticsRepository)
	// limit 0 → clamped to 50 before hitting the repo
	analytics.On("RecentErrorEvents", mock.Anything, uint64(5), 50).
		Return([]repository.ErrorEventDetail{{EventID: "e1"}}, nil)

	svc := NewIssueService(new(repomock.IssueRepository), analytics)
	events, err := svc.GetEvents(context.Background(), 5, 0)
	require.NoError(t, err)
	assert.Len(t, events, 1)
	analytics.AssertExpectations(t)
}

func TestAnalyticsService_ReleaseHealth_ComputesCrashFree(t *testing.T) {
	analytics := new(repomock.AnalyticsRepository)
	analytics.On("ReleaseHealth", mock.Anything, uint64(3), mock.Anything, mock.Anything).
		Return([]repository.ReleaseHealth{
			{Release: "v1", SessionsTotal: 100, SessionsErrored: 5},
			{Release: "v2", SessionsTotal: 0, SessionsErrored: 0},
		}, nil)

	svc := NewAnalyticsService(analytics)
	pts, err := svc.GetReleaseHealth(context.Background(), 3, time.Now().Add(-time.Hour), time.Now())
	require.NoError(t, err)
	require.Len(t, pts, 2)
	assert.InDelta(t, 0.95, pts[0].CrashFreeRate, 0.0001, "1 - 5/100")
	assert.Equal(t, 1.0, pts[1].CrashFreeRate, "no sessions → treated as fully crash-free")
}
