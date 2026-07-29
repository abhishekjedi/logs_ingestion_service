package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"error-logging/constants"
	dbdto "error-logging/db/dto"
	repomock "error-logging/db/repository/mock"
	"error-logging/dto"
	openreplayclient "error-logging/pkg/client/openreplay"
	"error-logging/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestReplayCredentialEncryption(t *testing.T) {
	service := &replayService{cfg: config.ReplayConfig{EncryptionKey: "test-encryption-key"}}
	first, err := service.encrypt("organization-secret")
	require.NoError(t, err)
	second, err := service.encrypt("organization-secret")
	require.NoError(t, err)
	assert.NotEqual(t, first, second)

	plaintext, err := service.decrypt(first)
	require.NoError(t, err)
	assert.Equal(t, "organization-secret", plaintext)
}

func TestReplayServiceUpsertValidatesAndEncryptsCredential(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/public/projects/project-key", r.URL.Path)
		assert.Equal(t, "Bearer organization-secret", r.Header.Get("Authorization"))
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]string{"projectKey": "project-key"}})
	}))
	defer server.Close()

	integrations := new(repomock.ReplayIntegrationRepository)
	saved := new(dbdto.ReplayIntegration)
	integrations.On(
		"GetByProjectAndKey",
		mock.Anything,
		uint64(9),
		constants.ReplayProviderOpenReplay,
		"project-key",
	).Return(nil, gorm.ErrRecordNotFound).Once()
	integrations.On("Upsert", mock.Anything, mock.AnythingOfType("*dto.ReplayIntegration")).
		Run(func(arguments mock.Arguments) {
			*saved = *arguments.Get(1).(*dbdto.ReplayIntegration)
			saved.ID = 12
		}).
		Return(nil).
		Once()
	integrations.On(
		"GetByProjectAndKey",
		mock.Anything,
		uint64(9),
		constants.ReplayProviderOpenReplay,
		"project-key",
	).Return(saved, nil).Once()

	cfg := config.ReplayConfig{
		EncryptionKey:  "test-key",
		RequestTimeout: time.Second,
		AllowInsecure:  true,
	}
	implementation := &replayService{
		integrations: integrations,
		openReplay:   openreplayclient.NewClient(cfg),
		cfg:          cfg,
	}
	response, err := implementation.UpsertIntegration(context.Background(), 9, dto.UpsertReplayIntegrationRequest{
		ExternalProjectKey: "project-key",
		APIBaseURL:         server.URL + "/v2",
		OrganizationAPIKey: "organization-secret",
	})
	require.NoError(t, err)
	assert.True(t, response.HasAPIKey)
	assert.NotEqual(t, "organization-secret", saved.APIKeyCiphertext)
	plaintext, err := implementation.decrypt(saved.APIKeyCiphertext)
	require.NoError(t, err)
	assert.Equal(t, "organization-secret", plaintext)
}

func TestNormalizeReplayEventRedactsSensitiveData(t *testing.T) {
	event, ok := normalizeReplayEvent(openreplayclient.Event{
		EventID:   "network-1",
		EventName: "$network_request",
		CreatedAt: time.Now().UnixMilli(),
		Properties: map[string]any{
			"method": "post",
			"url":    "https://shop.example/charge?token=secret&email=user@example.com",
			"status": float64(500),
		},
	})
	require.True(t, ok)
	assert.Equal(t, "network_failure", event.Kind)
	assert.Equal(t, "https://shop.example/charge", event.URL)
	assert.NotContains(t, event.Label, "secret")
	assert.NotContains(t, event.Label, "user@example.com")

	input, ok := normalizeReplayEvent(openreplayclient.Event{
		EventID:   "input-1",
		EventName: "$input",
		CreatedAt: time.Now().UnixMilli(),
		Properties: map[string]any{
			"name":  "card-number",
			"value": "4242424242424242",
		},
	})
	require.True(t, ok)
	assert.NotContains(t, input.Label, "4242")
}

func TestCompileSessionContextPrioritizesFailuresAndStaysBounded(t *testing.T) {
	now := time.Now().UTC()
	events := make([]openreplayclient.Event, 0, 120)
	for i := 0; i < 110; i++ {
		events = append(events, openreplayclient.Event{
			EventID:    fmt.Sprintf("click-%d", i),
			EventName:  "$click",
			CreatedAt:  now.Add(time.Duration(i) * time.Second).UnixMilli(),
			Properties: map[string]any{"label": fmt.Sprintf("button-%d", i)},
		})
	}
	for i := 0; i < 10; i++ {
		events = append(events, openreplayclient.Event{
			EventID:   "failure",
			EventName: "$network_request",
			CreatedAt: now.Add(time.Duration(110+i) * time.Second).UnixMilli(),
			Properties: map[string]any{
				"method": "POST",
				"url":    "https://shop.example/charge?token=secret",
				"status": 500,
			},
		})
	}

	compiled := compileSessionContext(now, "session", "https://replay", events, 120)
	assert.True(t, compiled.Truncated)
	assert.NotEmpty(t, compiled.NetworkFailures)
	encoded, err := json.Marshal(compiled)
	require.NoError(t, err)
	assert.Less(t, len(encoded), 50_000)
	assert.LessOrEqual(
		t,
		len(compiled.Journey)+len(compiled.NetworkFailures)+len(compiled.ConsoleErrors)+len(compiled.Exceptions),
		sessionContextLimit,
	)
}

func TestReplayServiceReturnsUpstreamStatusWithoutFailingIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	integrations := new(repomock.ReplayIntegrationRepository)
	issues := new(repomock.IssueRepository)
	serviceRepo := new(repomock.ServiceRepository)
	analytics := new(repomock.AnalyticsRepository)
	cfg := config.ReplayConfig{
		EncryptionKey:  "test-key",
		RequestTimeout: 2 * time.Second,
		CacheTTL:       time.Minute,
		AllowInsecure:  true,
	}
	implementation := &replayService{
		integrations: integrations,
		issues:       issues,
		services:     serviceRepo,
		analytics:    analytics,
		openReplay:   openreplayclient.NewClient(cfg),
		cfg:          cfg,
	}
	ciphertext, err := implementation.encrypt("org-key")
	require.NoError(t, err)
	focusedAt := time.Now().UTC()

	analytics.On("GetErrorEvent", mock.Anything, uint64(4), "event-1").Return(&dbdto.ErrorEventDetail{
		EventID:   "event-1",
		Timestamp: focusedAt,
		SessionID: "session-1",
		Attributes: map[string]string{
			"openreplay.project.key": "project-key",
			"openreplay.session.url": "https://app.openreplay.com/replay",
		},
	}, nil)
	issues.On("GetByID", mock.Anything, uint64(4)).Return(&dbdto.Issue{ID: 4, ServiceID: 7}, nil)
	serviceRepo.On("GetByID", mock.Anything, uint64(7)).Return(&dbdto.Service{ID: 7, ProjectID: 9}, nil)
	integrations.On(
		"GetByProjectAndKey",
		mock.Anything,
		uint64(9),
		constants.ReplayProviderOpenReplay,
		"project-key",
	).Return(&dbdto.ReplayIntegration{
		ProjectID:          9,
		Provider:           constants.ReplayProviderOpenReplay,
		ExternalProjectKey: "project-key",
		APIBaseURL:         server.URL,
		APIKeyCiphertext:   ciphertext,
		Enabled:            true,
	}, nil)

	response, err := implementation.GetSessionContext(context.Background(), 4, "event-1")
	require.NoError(t, err)
	assert.Equal(t, "temporarily_unavailable", response.Status)
	assert.Equal(t, "session-1", response.SessionID)
}

func TestReplayServiceMissingSessionSkipsProviderLookup(t *testing.T) {
	analytics := new(repomock.AnalyticsRepository)
	analytics.On("GetErrorEvent", mock.Anything, uint64(4), "event-1").Return(&dbdto.ErrorEventDetail{
		EventID:    "event-1",
		Timestamp:  time.Now().UTC(),
		Attributes: map[string]string{},
	}, nil)
	implementation := &replayService{analytics: analytics}

	response, err := implementation.GetSessionContext(context.Background(), 4, "event-1")
	require.NoError(t, err)
	assert.Equal(t, "missing_session", response.Status)
}
