package openreplay

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"error-logging/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientValidateProjectAndFetchEvents(t *testing.T) {
	var pages atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer org-key", r.Header.Get("Authorization"))
		switch r.URL.Path {
		case "/v2/public/projects/project-key":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"projectKey": "project-key"}})
		case "/v2/public/project-key/sessions/session-1/events":
			pages.Add(1)
			var request struct {
				Page int `json:"page"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
			events := []map[string]any{}
			if request.Page == 1 {
				events = []map[string]any{
					{"event_id": "1", "$event_name": "$pageview", "created_at": 1000},
					{"event_id": "2", "$event_name": "$click", "created_at": 2000},
				}
			} else {
				events = []map[string]any{{"event_id": "3", "$event_name": "$error", "created_at": 3000}}
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"total": 3, "events": events},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(config.ReplayConfig{RequestTimeout: time.Second, AllowInsecure: true})
	require.NoError(t, client.ValidateProject(context.Background(), server.URL+"/v2", "project-key", "org-key"))

	events, total, err := client.FetchEvents(
		context.Background(),
		server.URL+"/v2",
		"project-key",
		"session-1",
		"org-key",
		time.UnixMilli(1),
		time.UnixMilli(5000),
		5,
		2,
	)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, events, 3)
	assert.Equal(t, int32(2), pages.Load())
}

func TestClientRejectsInsecureBaseURL(t *testing.T) {
	client := NewClient(config.ReplayConfig{RequestTimeout: time.Second})
	err := client.ValidateProject(context.Background(), "http://openreplay.example", "project", "key")
	assert.ErrorContains(t, err, "must use HTTPS")
}

func TestClientRetriesServerErrors(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 3 {
			http.Error(w, "temporary", http.StatusServiceUnavailable)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"projectKey": "key"}})
	}))
	defer server.Close()

	client := NewClient(config.ReplayConfig{RequestTimeout: 2 * time.Second, AllowInsecure: true})
	require.NoError(t, client.ValidateProject(context.Background(), server.URL, "key", "org-key"))
	assert.Equal(t, int32(3), calls.Load())
}

func TestClientHandlesProviderFailures(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		timeout    time.Duration
		errorMatch string
	}{
		{
			name: "unauthorized",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
			},
			timeout:    time.Second,
			errorMatch: "401",
		},
		{
			name: "not found",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				http.NotFound(w, nil)
			},
			timeout:    time.Second,
			errorMatch: "404",
		},
		{
			name: "malformed JSON",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"data":`))
			},
			timeout:    time.Second,
			errorMatch: "decode OpenReplay events",
		},
		{
			name: "timeout",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				time.Sleep(30 * time.Millisecond)
				_, _ = w.Write([]byte(`{"data":{"total":0,"events":[]}}`))
			},
			timeout:    5 * time.Millisecond,
			errorMatch: "request failed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()
			client := NewClient(config.ReplayConfig{RequestTimeout: test.timeout, AllowInsecure: true})

			_, _, err := client.FetchEvents(
				context.Background(),
				server.URL,
				"project",
				"session",
				"key",
				time.UnixMilli(1),
				time.UnixMilli(2),
				1,
				200,
			)
			assert.ErrorContains(t, err, test.errorMatch)
		})
	}
}

func TestClientRetriesRateLimit(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			http.Error(w, "slow down", http.StatusTooManyRequests)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"total": 0, "events": []any{}},
		})
	}))
	defer server.Close()
	client := NewClient(config.ReplayConfig{RequestTimeout: time.Second, AllowInsecure: true})

	_, _, err := client.FetchEvents(
		context.Background(),
		server.URL,
		"project",
		"session",
		"key",
		time.UnixMilli(1),
		time.UnixMilli(2),
		1,
		200,
	)
	require.NoError(t, err)
	assert.Equal(t, int32(2), calls.Load())
}
