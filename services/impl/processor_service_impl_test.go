package impl

import (
	"context"
	"testing"
	"time"

	"error-logging/constants"
	dbdto "error-logging/db/dto"
	repomock "error-logging/db/repository/mock"
	"error-logging/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type fakeStore struct{ puts int }

func (f *fakeStore) Put(context.Context, string, []byte, string) error { f.puts++; return nil }

type fakeLimiter struct{ allow bool }

func (f *fakeLimiter) AllowN(_ context.Context, _ string, n int) int {
	if f.allow {
		return n
	}
	return 0
}

// countingLimiter records how many times AllowN is called and caps at `allow`.
type countingLimiter struct {
	calls int
	allow int
}

func (c *countingLimiter) AllowN(_ context.Context, _ string, n int) int {
	c.calls++
	if c.allow < n {
		return c.allow
	}
	return n
}

// fakeCache always misses, so tests exercise the ResolveOrCreate (MySQL) path.
type fakeCache struct{}

func (fakeCache) Get(context.Context, uint64, string) (uint64, bool) { return 0, false }
func (fakeCache) Set(context.Context, uint64, string, uint64)        {}

const errorOTLP = `{"resourceLogs":[{"scopeLogs":[{"logRecords":[
  {"severityNumber":17,"body":{"stringValue":"boom"},"attributes":[
    {"key":"exception.type","value":{"stringValue":"NullPointerException"}},
    {"key":"exception.message","value":{"stringValue":"null id 42"}},
    {"key":"user.id","value":{"stringValue":"u-1"}}
  ]}
]}]}]}`

const infoOTLP = `{"resourceLogs":[{"scopeLogs":[{"logRecords":[
  {"severityNumber":9,"body":{"stringValue":"just info"}}
]}]}]}`

func testMsg(payload string) dto.LogIngestMessage {
	return dto.LogIngestMessage{ServiceID: 3, ProjectID: 1, ReceivedAt: time.Unix(0, 0).UTC(), Payload: []byte(payload)}
}

func newProcessor(store *fakeStore, issues *repomock.IssueRepository, allow bool) *processorService {
	return &processorService{
		store:    store,
		issues:   issues,
		cache:    fakeCache{},
		limiter:  &fakeLimiter{allow: allow},
		poolSize: 4,
	}
}

func TestProcessor_ErrorRecord_GroupsAndStores(t *testing.T) {
	issues := new(repomock.IssueRepository)
	issues.On("ResolveOrCreate", mock.Anything, mock.AnythingOfType("*dto.Issue")).
		Return(&dbdto.Issue{ID: 5, Status: constants.StatusUnresolved}, true, nil)

	p := newProcessor(&fakeStore{}, issues, true)
	res, err := p.TransformBatch(context.Background(), []dto.LogIngestMessage{testMsg(errorOTLP)})
	require.NoError(t, err)

	require.Len(t, res.Logs, 1)
	assert.Equal(t, uint64(5), res.Logs[0].IssueID, "log row tagged with resolved issue id")
	require.Len(t, res.ErrorEvents, 1, "full-fidelity row stored when under the limit")
	assert.Equal(t, "NullPointerException", res.ErrorEvents[0].ExceptionType)
	issues.AssertExpectations(t)
}

func TestProcessor_NonErrorRecord_NoIssueNoErrorEvent(t *testing.T) {
	issues := new(repomock.IssueRepository)

	p := newProcessor(&fakeStore{}, issues, true)
	res, err := p.TransformBatch(context.Background(), []dto.LogIngestMessage{testMsg(infoOTLP)})
	require.NoError(t, err)

	require.Len(t, res.Logs, 1)
	assert.Equal(t, uint64(0), res.Logs[0].IssueID)
	assert.Empty(t, res.ErrorEvents)
	issues.AssertNotCalled(t, "ResolveOrCreate", mock.Anything, mock.Anything)
}

func TestProcessor_RateLimit_GatesFidelityNotCounting(t *testing.T) {
	// Limiter denies: the log row (→ counts) is still written, but no error_events row.
	issues := new(repomock.IssueRepository)
	issues.On("ResolveOrCreate", mock.Anything, mock.Anything).
		Return(&dbdto.Issue{ID: 5, Status: constants.StatusUnresolved}, true, nil)

	p := newProcessor(&fakeStore{}, issues, false)
	res, err := p.TransformBatch(context.Background(), []dto.LogIngestMessage{testMsg(errorOTLP)})
	require.NoError(t, err)

	require.Len(t, res.Logs, 1, "log row always written — counting is never gated")
	assert.Equal(t, uint64(5), res.Logs[0].IssueID)
	assert.Empty(t, res.ErrorEvents, "fidelity row dropped by the rate limiter")
}

func TestProcessor_Regression_ReopensResolvedIssue(t *testing.T) {
	issues := new(repomock.IssueRepository)
	issues.On("ResolveOrCreate", mock.Anything, mock.Anything).
		Return(&dbdto.Issue{ID: 9, Status: constants.StatusResolved}, false, nil)
	issues.On("MarkRegressed", mock.Anything, uint64(9)).Return(true, nil)

	p := newProcessor(&fakeStore{}, issues, true)
	_, err := p.TransformBatch(context.Background(), []dto.LogIngestMessage{testMsg(errorOTLP)})
	require.NoError(t, err)
	issues.AssertCalled(t, "MarkRegressed", mock.Anything, uint64(9))
}

func TestProcessor_MemoizesFingerprintAcrossCycle(t *testing.T) {
	// Two messages with the SAME fingerprint → resolved once for the whole cycle.
	issues := new(repomock.IssueRepository)
	issues.On("ResolveOrCreate", mock.Anything, mock.Anything).
		Return(&dbdto.Issue{ID: 5, Status: constants.StatusUnresolved}, true, nil).Once()

	p := newProcessor(&fakeStore{}, issues, true)
	res, err := p.TransformBatch(context.Background(), []dto.LogIngestMessage{testMsg(errorOTLP), testMsg(errorOTLP)})
	require.NoError(t, err)

	assert.Len(t, res.Logs, 2)
	issues.AssertNumberOfCalls(t, "ResolveOrCreate", 1)
}

func TestProcessor_ArchivesOncePerCycle(t *testing.T) {
	// Three messages in one cycle → a single archive object (not one per message).
	store := &fakeStore{}
	p := newProcessor(store, new(repomock.IssueRepository), true)
	_, err := p.TransformBatch(context.Background(), []dto.LogIngestMessage{
		testMsg(infoOTLP), testMsg(infoOTLP), testMsg(infoOTLP),
	})
	require.NoError(t, err)
	assert.Equal(t, 1, store.puts, "one batched archive object per cycle")
}

func TestProcessor_BatchedRateLimit_OneCallPerFingerprint(t *testing.T) {
	// Five error records sharing a fingerprint → ONE rate-limit call, capped to the
	// returned allowance; all five still counted in logs.
	issues := new(repomock.IssueRepository)
	issues.On("ResolveOrCreate", mock.Anything, mock.Anything).
		Return(&dbdto.Issue{ID: 5, Status: constants.StatusUnresolved}, true, nil)
	lim := &countingLimiter{allow: 2}

	p := &processorService{store: &fakeStore{}, issues: issues, cache: fakeCache{}, limiter: lim, poolSize: 4}
	msgs := []dto.LogIngestMessage{
		testMsg(errorOTLP), testMsg(errorOTLP), testMsg(errorOTLP), testMsg(errorOTLP), testMsg(errorOTLP),
	}

	res, err := p.TransformBatch(context.Background(), msgs)
	require.NoError(t, err)

	assert.Len(t, res.Logs, 5, "every event counted")
	assert.Len(t, res.ErrorEvents, 2, "fidelity capped to the batched allowance")
	assert.Equal(t, 1, lim.calls, "one Redis op for the fingerprint across the whole cycle")
}
