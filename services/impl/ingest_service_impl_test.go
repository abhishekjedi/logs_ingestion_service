package impl

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	dbdto "error-logging/db/dto"
	"error-logging/dto"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeWriter struct {
	msgs []kafka.Message
	err  error
}

func (f *fakeWriter) WriteMessages(_ context.Context, msgs ...kafka.Message) error {
	if f.err != nil {
		return f.err
	}
	f.msgs = append(f.msgs, msgs...)
	return nil
}

func TestIngestService_Ingest(t *testing.T) {
	fw := &fakeWriter{}
	svc := &ingestService{writer: fw}

	payload := []byte(`{"resourceLogs":[]}`)
	err := svc.Ingest(context.Background(), &dbdto.Service{ID: 7, ProjectID: 3}, payload)
	require.NoError(t, err)

	require.Len(t, fw.msgs, 1)
	assert.Equal(t, "7", string(fw.msgs[0].Key), "keyed by service id for stable partitioning")

	var msg dto.LogIngestMessage
	require.NoError(t, json.Unmarshal(fw.msgs[0].Value, &msg))
	assert.Equal(t, uint64(7), msg.ServiceID)
	assert.Equal(t, uint64(3), msg.ProjectID)
	assert.JSONEq(t, string(payload), string(msg.Payload))
	assert.False(t, msg.ReceivedAt.IsZero())
}

func TestIngestService_Ingest_WriterError(t *testing.T) {
	fw := &fakeWriter{err: errors.New("broker down")}
	svc := &ingestService{writer: fw}

	err := svc.Ingest(context.Background(), &dbdto.Service{ID: 1}, []byte(`{}`))
	assert.Error(t, err)
}
