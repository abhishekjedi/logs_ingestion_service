package impl

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	dbdto "error-logging/db/dto"
	repomock "error-logging/db/repository/mock"
	"error-logging/dto"
	"error-logging/pkg/config"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type fakeReader struct {
	queue     []kafka.Message
	idx       int
	committed []kafka.Message
}

func (f *fakeReader) FetchMessage(context.Context) (kafka.Message, error) {
	if f.idx < len(f.queue) {
		m := f.queue[f.idx]
		f.idx++
		return m, nil
	}
	return kafka.Message{}, context.DeadlineExceeded
}

func (f *fakeReader) CommitMessages(_ context.Context, msgs ...kafka.Message) error {
	f.committed = append(f.committed, msgs...)
	return nil
}

type fakeProcessor struct {
	result dto.TransformResult
	err    error
}

func (f *fakeProcessor) TransformBatch(context.Context, []dto.LogIngestMessage) (dto.TransformResult, error) {
	return f.result, f.err
}

func testCfg() config.WorkerConfig {
	return config.WorkerConfig{
		PoolSize:         2,
		FetchMaxMessages: 100,
		FetchMaxBytes:    1 << 30,
		FetchMaxWait:     10 * time.Millisecond,
		FlushChunkRows:   1000,
		FlushRetries:     3,
	}
}

func kmsg(value string) kafka.Message { return kafka.Message{Value: []byte(value)} }

func parseMsgs(t *testing.T, kmsgs []kafka.Message) []dto.LogIngestMessage {
	t.Helper()
	var out []dto.LogIngestMessage
	for _, km := range kmsgs {
		var m dto.LogIngestMessage
		if json.Unmarshal(km.Value, &m) == nil {
			out = append(out, m)
		}
	}
	return out
}

const validMsg = `{"service_id":3,"project_id":1,"payload":{"resourceLogs":[]}}`

func TestFetchCycle_SkipsPoisonButCommitsIt(t *testing.T) {
	reader := &fakeReader{queue: []kafka.Message{
		kmsg(validMsg),
		kmsg(`{ this is not json`),
		kmsg(validMsg),
	}}
	c := &batchConsumer{reader: reader, cfg: testCfg()}

	msgs, kmsgs, shutdown := c.fetchCycle(context.Background())
	assert.False(t, shutdown)
	assert.Len(t, msgs, 2, "two valid messages parsed")
	assert.Len(t, kmsgs, 3, "all three (incl. poison) queued for commit")
}

func TestProcessCycle_FlushesThenCommits(t *testing.T) {
	reader := &fakeReader{}
	logs := new(repomock.LogRepository)
	errs := new(repomock.ErrorEventRepository)
	logs.On("InsertBatch", mock.Anything, mock.Anything).Return(nil)
	errs.On("InsertBatch", mock.Anything, mock.Anything).Return(nil)

	c := &batchConsumer{
		reader:      reader,
		processor:   &fakeProcessor{result: dto.TransformResult{Logs: []dbdto.LogRow{{ServiceID: 3}}}},
		logs:        logs,
		errorEvents: errs,
		cfg:         testCfg(),
	}

	kmsgs := []kafka.Message{kmsg(validMsg)}
	err := c.processCycle(parseMsgs(t, kmsgs), kmsgs)
	require.NoError(t, err)

	logs.AssertCalled(t, "InsertBatch", mock.Anything, mock.Anything)
	assert.Len(t, reader.committed, 1, "committed only after a successful flush")
}

func TestFlush_ChunksLargeBatch(t *testing.T) {
	logs := new(repomock.LogRepository)
	errs := new(repomock.ErrorEventRepository)
	logs.On("InsertBatch", mock.Anything, mock.Anything).Return(nil)
	errs.On("InsertBatch", mock.Anything, mock.Anything).Return(nil)

	c := &batchConsumer{
		logs:        logs,
		errorEvents: errs,
		cfg:         config.WorkerConfig{FlushChunkRows: 2, FlushRetries: 1},
	}

	res := dto.TransformResult{Logs: make([]dbdto.LogRow, 5)}
	require.NoError(t, c.flush(context.Background(), res))
	logs.AssertNumberOfCalls(t, "InsertBatch", 3)
}

func TestProcessCycle_FlushFailure_DoesNotCommit(t *testing.T) {
	reader := &fakeReader{}
	logs := new(repomock.LogRepository)
	errs := new(repomock.ErrorEventRepository)
	logs.On("InsertBatch", mock.Anything, mock.Anything).Return(errors.New("clickhouse down"))

	c := &batchConsumer{
		reader:      reader,
		processor:   &fakeProcessor{result: dto.TransformResult{Logs: []dbdto.LogRow{{ServiceID: 3}}}},
		logs:        logs,
		errorEvents: errs,
		cfg:         config.WorkerConfig{FetchMaxMessages: 100, FlushRetries: 2},
	}

	kmsgs := []kafka.Message{kmsg(validMsg)}
	err := c.processCycle(parseMsgs(t, kmsgs), kmsgs)
	require.Error(t, err)
	assert.Empty(t, reader.committed, "no commit when the flush ultimately fails")
	errs.AssertNotCalled(t, "InsertBatch", mock.Anything, mock.Anything)
}
