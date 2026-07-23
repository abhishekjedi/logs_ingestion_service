package otlp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlatten_ExtractsPromotedFields(t *testing.T) {
	payload := []byte(`{
      "resourceLogs": [{
        "resource": { "attributes": [
          {"key":"deployment.environment","value":{"stringValue":"production"}},
          {"key":"service.version","value":{"stringValue":"v1.2.3"}}
        ]},
        "scopeLogs": [{
          "logRecords": [{
            "timeUnixNano":"1700000000000000000",
            "severityNumber":17,
            "severityText":"ERROR",
            "body":{"stringValue":"boom"},
            "traceId":"abc","spanId":"def",
            "attributes":[
              {"key":"user.id","value":{"stringValue":"user-1"}},
              {"key":"session.id","value":{"stringValue":"sess-1"}},
              {"key":"exception.type","value":{"stringValue":"RuntimeError"}},
              {"key":"exception.message","value":{"stringValue":"kaboom"}}
            ]
          }]
        }]
      }]
    }`)

	recs, err := Flatten(payload)
	require.NoError(t, err)
	require.Len(t, recs, 1)

	r := recs[0]
	assert.Equal(t, "production", r.Environment)
	assert.Equal(t, "v1.2.3", r.Release)
	assert.Equal(t, "user-1", r.UserID)
	assert.Equal(t, "sess-1", r.SessionID)
	assert.Equal(t, "RuntimeError", r.ExceptionType)
	assert.Equal(t, "kaboom", r.ExceptionMessage)
	assert.Equal(t, uint8(17), r.SeverityNumber)
	assert.Equal(t, "boom", r.Body)
	assert.Equal(t, "abc", r.TraceID)
	assert.True(t, r.IsError())
	assert.False(t, r.Timestamp.IsZero())
}

func TestFlatten_NonErrorAndMultipleRecords(t *testing.T) {
	payload := []byte(`{"resourceLogs":[{"scopeLogs":[{"logRecords":[
      {"severityNumber":9,"body":{"stringValue":"info log"}},
      {"severityNumber":9,"body":{"stringValue":"another"}}
    ]}]}]}`)

	recs, err := Flatten(payload)
	require.NoError(t, err)
	require.Len(t, recs, 2)
	assert.False(t, recs[0].IsError(), "no exception → not grouped as an issue")
	assert.False(t, recs[0].Timestamp.IsZero(), "missing timestamp falls back to now")
}

func TestFlatten_InvalidJSON(t *testing.T) {
	_, err := Flatten([]byte(`not json`))
	assert.Error(t, err)
}
