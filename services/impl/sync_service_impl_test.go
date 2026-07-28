package impl

import (
	"context"
	"errors"
	"testing"
	"time"

	dbdto "error-logging/db/dto"
	repomock "error-logging/db/repository/mock"
	"error-logging/dto"
	"error-logging/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSyncService_MapsAggregatesToUpdates(t *testing.T) {
	stats := new(repomock.IssueStatsRepository)
	issues := new(repomock.IssueRepository)

	stats.On("ActiveSince", mock.Anything, mock.Anything).Return([]dbdto.IssueStatsAggregate{
		{ServiceID: 1, IssueID: 5, EventCount: 100, Users: 7, Sessions: 4, LastSeen: time.Unix(1700000000, 0)},
		{ServiceID: 1, IssueID: 6, EventCount: 3, Users: 1, Sessions: 1},
	}, nil)

	var captured []dto.IssueStatsUpdate
	issues.On("UpdateStatsBatch", mock.Anything, mock.Anything).Return(nil).Run(func(a mock.Arguments) {
		captured = a.Get(1).([]dto.IssueStatsUpdate)
	})

	svc := NewSyncService(stats, issues, config.SyncConfig{ActiveWindow: time.Hour, BatchSize: 500})
	n, err := svc.SyncOnce(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 2, n)
	require.Len(t, captured, 2)
	assert.Equal(t, uint64(5), captured[0].IssueID)
	assert.Equal(t, uint64(100), captured[0].EventCount)
	assert.Equal(t, uint64(7), captured[0].AffectedUsers)
	assert.Equal(t, uint64(4), captured[0].AffectedSessions)
}

func TestSyncService_Empty_NoWrite(t *testing.T) {
	stats := new(repomock.IssueStatsRepository)
	issues := new(repomock.IssueRepository)
	stats.On("ActiveSince", mock.Anything, mock.Anything).Return(nil, nil)

	svc := NewSyncService(stats, issues, config.SyncConfig{ActiveWindow: time.Hour, BatchSize: 500})
	n, err := svc.SyncOnce(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	issues.AssertNotCalled(t, "UpdateStatsBatch", mock.Anything, mock.Anything)
}

func TestSyncService_ChunksUpdates(t *testing.T) {
	stats := new(repomock.IssueStatsRepository)
	issues := new(repomock.IssueRepository)
	stats.On("ActiveSince", mock.Anything, mock.Anything).Return([]dbdto.IssueStatsAggregate{
		{IssueID: 1}, {IssueID: 2}, {IssueID: 3},
	}, nil)
	issues.On("UpdateStatsBatch", mock.Anything, mock.Anything).Return(nil)

	svc := NewSyncService(stats, issues, config.SyncConfig{ActiveWindow: time.Hour, BatchSize: 1})
	n, err := svc.SyncOnce(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 3, n)
	issues.AssertNumberOfCalls(t, "UpdateStatsBatch", 3)
}

func TestSyncService_ReadError(t *testing.T) {
	stats := new(repomock.IssueStatsRepository)
	issues := new(repomock.IssueRepository)
	stats.On("ActiveSince", mock.Anything, mock.Anything).Return(nil, errors.New("clickhouse down"))

	svc := NewSyncService(stats, issues, config.SyncConfig{ActiveWindow: time.Hour, BatchSize: 500})
	_, err := svc.SyncOnce(context.Background())
	assert.Error(t, err)
}
