package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishErrorPropagatesReplayContextAsOTelAttributes(t *testing.T) {
	var payload map[string]any
	ingest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "elk_test_key", r.Header.Get("X-API-Key"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ingest.Close()

	source := httptest.NewRequest(http.MethodPost, "/checkout", nil)
	source.Header.Set("X-OpenReplay-Session-ID", "session-4813")
	source.Header.Set("X-OpenReplay-Project-Key", "project-key")
	source.Header.Set("X-OpenReplay-Session-URL", "https://app.openreplay.com/session/4813")
	source.Header.Set("Cookie", "secret-cookie")

	err := publishError(source, demoConfig{IngestURL: ingest.URL, APIKey: "elk_test_key"})
	require.NoError(t, err)
	encoded, err := json.Marshal(payload)
	require.NoError(t, err)
	body := string(encoded)
	assert.Contains(t, body, `"session.id"`)
	assert.Contains(t, body, `"session-4813"`)
	assert.Contains(t, body, `"openreplay.project.key"`)
	assert.Contains(t, body, `"openreplay.session.url"`)
	assert.NotContains(t, body, "secret-cookie")
}
