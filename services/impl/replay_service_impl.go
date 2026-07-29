package impl

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"error-logging/constants"
	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	"error-logging/dto"
	openreplayclient "error-logging/pkg/client/openreplay"
	redisclient "error-logging/pkg/client/redis"
	"error-logging/pkg/config"
	"error-logging/services"

	"gorm.io/gorm"
)

const (
	openReplayCloudAPI    = "https://api.openreplay.com/v2"
	openReplayCloudIngest = "https://api.openreplay.com/ingest"
	sessionContextLimit   = 80
)

var (
	emailPattern  = regexp.MustCompile(`(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`)
	jwtPattern    = regexp.MustCompile(`eyJ[a-zA-Z0-9_-]{8,}\.[a-zA-Z0-9_-]{8,}\.[a-zA-Z0-9_-]{8,}`)
	secretPattern = regexp.MustCompile(`(?i)(authorization|cookie|token|password|secret|api[_-]?key)\s*[:=]\s*[^\s,;]+`)
)

type replayService struct {
	integrations repository.ReplayIntegrationRepository
	issues       repository.IssueRepository
	services     repository.ServiceRepository
	analytics    repository.AnalyticsRepository
	openReplay   *openreplayclient.Client
	redis        *redisclient.Client
	cfg          config.ReplayConfig
}

func NewReplayService(
	integrations repository.ReplayIntegrationRepository,
	issues repository.IssueRepository,
	serviceRepo repository.ServiceRepository,
	analytics repository.AnalyticsRepository,
	openReplay *openreplayclient.Client,
	redisClient *redisclient.Client,
	cfg config.ReplayConfig,
) services.ReplayService {
	return &replayService{
		integrations: integrations,
		issues:       issues,
		services:     serviceRepo,
		analytics:    analytics,
		openReplay:   openReplay,
		redis:        redisClient,
		cfg:          cfg,
	}
}

func (s *replayService) ListIntegrations(
	ctx context.Context,
	projectID uint64,
) ([]dto.ReplayIntegrationResponse, error) {
	rows, err := s.integrations.ListByProject(ctx, projectID, constants.ReplayProviderOpenReplay)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ReplayIntegrationResponse, 0, len(rows))
	for i := range rows {
		out = append(out, replayIntegrationResponse(&rows[i]))
	}
	return out, nil
}

func (s *replayService) UpsertIntegration(
	ctx context.Context,
	projectID uint64,
	request dto.UpsertReplayIntegrationRequest,
) (*dto.ReplayIntegrationResponse, error) {
	projectKey := strings.TrimSpace(request.ExternalProjectKey)
	if projectKey == "" {
		return nil, fmt.Errorf("%w: external_project_key is required", services.ErrInvalidInput)
	}
	if _, err := s.encryptionKey(); err != nil {
		return nil, err
	}
	apiBaseURL := strings.TrimSpace(request.APIBaseURL)
	if apiBaseURL == "" {
		apiBaseURL = openReplayCloudAPI
	}
	ingestPoint := strings.TrimSpace(request.IngestPoint)
	if ingestPoint == "" {
		ingestPoint = openReplayCloudIngest
	}

	existing, err := s.integrations.GetByProjectAndKey(
		ctx,
		projectID,
		constants.ReplayProviderOpenReplay,
		projectKey,
	)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	apiKey := strings.TrimSpace(request.OrganizationAPIKey)
	if apiKey == "" && existing != nil {
		apiKey, err = s.decrypt(existing.APIKeyCiphertext)
		if err != nil {
			return nil, err
		}
	}
	if apiKey == "" {
		return nil, fmt.Errorf("%w: organization_api_key is required", services.ErrInvalidInput)
	}
	validationCtx, cancelValidation := context.WithTimeout(ctx, s.cfg.RequestTimeout)
	defer cancelValidation()
	if err := s.openReplay.ValidateProject(validationCtx, apiBaseURL, projectKey, apiKey); err != nil {
		if existing != nil {
			failed := *existing
			failed.LastError = truncate(err.Error(), 1024)
			failed.UpdatedAt = time.Now().UTC()
			_ = s.integrations.Upsert(ctx, &failed)
		}
		return nil, fmt.Errorf("%w: OpenReplay validation failed: %v", services.ErrUpstream, err)
	}

	ciphertext, err := s.encrypt(apiKey)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	enabled := true
	if request.Enabled != nil {
		enabled = *request.Enabled
	}
	row := &dbdto.ReplayIntegration{
		ProjectID:          projectID,
		Provider:           constants.ReplayProviderOpenReplay,
		ExternalProjectKey: projectKey,
		APIBaseURL:         apiBaseURL,
		IngestPoint:        ingestPoint,
		APIKeyCiphertext:   ciphertext,
		Enabled:            enabled,
		LastValidatedAt:    &now,
		LastError:          "",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if existing != nil {
		row.ID = existing.ID
		row.CreatedAt = existing.CreatedAt
	}
	if err := s.integrations.Upsert(ctx, row); err != nil {
		return nil, err
	}
	saved, err := s.integrations.GetByProjectAndKey(
		ctx,
		projectID,
		constants.ReplayProviderOpenReplay,
		projectKey,
	)
	if err != nil {
		return nil, err
	}
	response := replayIntegrationResponse(saved)
	return &response, nil
}

func (s *replayService) DeleteIntegration(
	ctx context.Context,
	projectID uint64,
	externalProjectKey string,
) error {
	if strings.TrimSpace(externalProjectKey) == "" {
		return fmt.Errorf("%w: project_key is required", services.ErrInvalidInput)
	}
	err := s.integrations.DeleteByProjectAndKey(
		ctx,
		projectID,
		constants.ReplayProviderOpenReplay,
		externalProjectKey,
	)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return services.ErrNotFound
	}
	return err
}

func (s *replayService) GetSessionContext(
	ctx context.Context,
	issueID uint64,
	eventID string,
) (*dto.SessionContextResponse, error) {
	event, err := s.analytics.GetErrorEvent(ctx, issueID, eventID)
	if err != nil {
		return nil, services.ErrNotFound
	}
	response := emptySessionContext(event.Timestamp)
	if event.SessionID == "" {
		response.Status = "missing_session"
		return response, nil
	}
	response.SessionID = event.SessionID
	response.ReplayURL = event.Attributes["openreplay.session.url"]

	issue, err := s.issues.GetByID(ctx, issueID)
	if err != nil {
		return nil, services.ErrNotFound
	}
	service, err := s.services.GetByID(ctx, issue.ServiceID)
	if err != nil {
		return nil, services.ErrNotFound
	}
	integration, err := s.resolveIntegration(ctx, service.ProjectID, event.Attributes["openreplay.project.key"])
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Status = "not_configured"
			return response, nil
		}
		return nil, err
	}
	if !integration.Enabled {
		response.Status = "not_configured"
		return response, nil
	}

	cacheKey := fmt.Sprintf(
		"replay-context:%d:%s:%s:%d",
		service.ProjectID,
		integration.ExternalProjectKey,
		event.SessionID,
		event.Timestamp.Unix(),
	)
	if cached := s.getCachedContext(ctx, cacheKey); cached != nil {
		return cached, nil
	}

	apiKey, err := s.decrypt(integration.APIKeyCiphertext)
	if err != nil {
		return nil, err
	}
	fetchCtx, cancelFetch := context.WithTimeout(ctx, s.cfg.RequestTimeout)
	defer cancelFetch()
	sourceEvents, total, err := s.openReplay.FetchEvents(
		fetchCtx,
		integration.APIBaseURL,
		integration.ExternalProjectKey,
		event.SessionID,
		apiKey,
		event.Timestamp.Add(-5*time.Minute),
		event.Timestamp.Add(30*time.Second),
		5,
		200,
	)
	if err != nil {
		response.Status = "temporarily_unavailable"
		return response, nil
	}
	if len(sourceEvents) == 0 {
		response.Status = "recording_pending"
		return response, nil
	}

	compiled := compileSessionContext(event.Timestamp, event.SessionID, response.ReplayURL, sourceEvents, total)
	s.cacheContext(ctx, cacheKey, compiled)
	return compiled, nil
}

func (s *replayService) resolveIntegration(
	ctx context.Context,
	projectID uint64,
	externalProjectKey string,
) (*dbdto.ReplayIntegration, error) {
	if externalProjectKey != "" {
		return s.integrations.GetByProjectAndKey(
			ctx,
			projectID,
			constants.ReplayProviderOpenReplay,
			externalProjectKey,
		)
	}
	rows, err := s.integrations.ListByProject(ctx, projectID, constants.ReplayProviderOpenReplay)
	if err != nil {
		return nil, err
	}
	active := make([]dbdto.ReplayIntegration, 0, len(rows))
	for i := range rows {
		if rows[i].Enabled {
			active = append(active, rows[i])
		}
	}
	if len(active) != 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return &active[0], nil
}

func replayIntegrationResponse(row *dbdto.ReplayIntegration) dto.ReplayIntegrationResponse {
	return dto.ReplayIntegrationResponse{
		ID:                 row.ID,
		ProjectID:          row.ProjectID,
		Provider:           row.Provider,
		ExternalProjectKey: row.ExternalProjectKey,
		APIBaseURL:         row.APIBaseURL,
		IngestPoint:        row.IngestPoint,
		Enabled:            row.Enabled,
		HasAPIKey:          row.APIKeyCiphertext != "",
		LastValidatedAt:    row.LastValidatedAt,
		LastError:          row.LastError,
	}
}

func emptySessionContext(focusedAt time.Time) *dto.SessionContextResponse {
	return &dto.SessionContextResponse{
		FocusedAt:       focusedAt,
		Journey:         []dto.SessionContextEvent{},
		NetworkFailures: []dto.SessionContextEvent{},
		ConsoleErrors:   []dto.SessionContextEvent{},
		Exceptions:      []dto.SessionContextEvent{},
		Counts:          map[string]int{},
	}
}

func compileSessionContext(
	focusedAt time.Time,
	sessionID, replayURL string,
	source []openreplayclient.Event,
	total int,
) *dto.SessionContextResponse {
	response := emptySessionContext(focusedAt)
	response.Status = "ready"
	response.SessionID = sessionID
	response.ReplayURL = replayURL

	events := make([]dto.SessionContextEvent, 0, len(source))
	for i := range source {
		if normalized, ok := normalizeReplayEvent(source[i]); ok {
			events = append(events, normalized)
			response.Counts[normalized.Kind]++
		}
	}
	sort.Slice(events, func(i, j int) bool { return events[i].Timestamp.Before(events[j].Timestamp) })
	events = collapseReplayEvents(events)
	response.Truncated = total > len(source) || len(events) > sessionContextLimit
	events = prioritizeReplayEvents(events, sessionContextLimit)

	for _, event := range events {
		switch event.Kind {
		case "network_failure":
			response.NetworkFailures = append(response.NetworkFailures, event)
		case "console_error":
			response.ConsoleErrors = append(response.ConsoleErrors, event)
		case "exception":
			response.Exceptions = append(response.Exceptions, event)
		default:
			response.Journey = append(response.Journey, event)
		}
	}
	return response
}

func normalizeReplayEvent(source openreplayclient.Event) (dto.SessionContextEvent, bool) {
	properties := make(map[string]any, len(source.Properties)+len(source.AltProperties))
	for key, value := range source.Properties {
		properties[key] = value
	}
	for key, value := range source.AltProperties {
		properties[key] = value
	}
	name := strings.ToLower(strings.TrimSpace(source.EventName))
	if name == "" {
		return dto.SessionContextEvent{}, false
	}
	event := dto.SessionContextEvent{
		Timestamp:     time.UnixMilli(source.CreatedAt).UTC(),
		SourceEventID: source.EventID,
		Untrusted:     true,
	}

	switch {
	case containsAny(name, "network", "request", "fetch", "xhr"):
		status := intProperty(properties, "status", "status_code", "response_status")
		if status < 400 {
			return dto.SessionContextEvent{}, false
		}
		event.Kind = "network_failure"
		event.Method = strings.ToUpper(stringProperty(properties, "method", "request_method", "networkrequest_method"))
		event.URL = sanitizeURL(stringProperty(properties, "url", "request_url", "networkrequest_url"))
		event.StatusCode = status
		event.DurationMS = floatProperty(properties, "duration", "duration_ms", "response_time")
		event.Label = strings.TrimSpace(event.Method + " " + event.URL + " returned " + strconv.Itoa(status))
	case containsAny(name, "console"):
		level := strings.ToLower(stringProperty(properties, "level", "console_level", "consolelog_level"))
		if level != "error" && level != "warn" && level != "warning" && level != "assert" {
			return dto.SessionContextEvent{}, false
		}
		event.Kind = "console_error"
		event.Label = redactText(stringProperty(properties, "message", "value", "consolelog_value"))
	case containsAny(name, "exception", "jsexception", "$error"):
		event.Kind = "exception"
		event.Label = redactText(firstNonEmpty(
			stringProperty(properties, "message", "exception_message", "jsexception_message"),
			source.EventName,
		))
	case containsAny(name, "pageview", "navigation", "location"):
		event.Kind = "navigation"
		path := sanitizeURL(stringProperty(properties, "path", "url", "href"))
		event.URL = path
		event.Label = "Opened " + path
	case containsAny(name, "click", "tap"):
		event.Kind = "click"
		event.Label = "Clicked " + redactText(firstNonEmpty(
			stringProperty(properties, "label", "text", "selector", "target"),
			"element",
		))
	case containsAny(name, "input", "change"):
		event.Kind = "input"
		event.Label = "Entered a value in " + redactText(firstNonEmpty(
			stringProperty(properties, "label", "name", "selector", "target"),
			"field",
		))
	default:
		event.Kind = "custom"
		event.Label = redactText(source.EventName)
	}
	event.Label = truncate(event.Label, 256)
	return event, event.Label != ""
}

func collapseReplayEvents(events []dto.SessionContextEvent) []dto.SessionContextEvent {
	out := make([]dto.SessionContextEvent, 0, len(events))
	for _, event := range events {
		if len(out) > 0 {
			previous := &out[len(out)-1]
			if previous.Kind == event.Kind &&
				previous.Label == event.Label &&
				event.Timestamp.Sub(previous.Timestamp) <= 2*time.Second {
				if previous.Count == 0 {
					previous.Count = 2
				} else {
					previous.Count++
				}
				continue
			}
		}
		out = append(out, event)
	}
	return out
}

func prioritizeReplayEvents(events []dto.SessionContextEvent, limit int) []dto.SessionContextEvent {
	if len(events) <= limit {
		return events
	}
	selected := make([]dto.SessionContextEvent, 0, limit)
	for _, event := range events {
		if event.Kind == "network_failure" || event.Kind == "console_error" || event.Kind == "exception" {
			selected = append(selected, event)
			if len(selected) == limit {
				break
			}
		}
	}
	if len(selected) < limit {
		for _, event := range events {
			if event.Kind != "network_failure" && event.Kind != "console_error" && event.Kind != "exception" {
				selected = append(selected, event)
				if len(selected) == limit {
					break
				}
			}
		}
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].Timestamp.Before(selected[j].Timestamp) })
	return selected
}

func (s *replayService) getCachedContext(ctx context.Context, key string) *dto.SessionContextResponse {
	if s.redis == nil || s.redis.RDB == nil {
		return nil
	}
	value, err := s.redis.RDB.Get(ctx, key).Bytes()
	if err != nil {
		return nil
	}
	var response dto.SessionContextResponse
	if json.Unmarshal(value, &response) != nil {
		return nil
	}
	return &response
}

func (s *replayService) cacheContext(ctx context.Context, key string, response *dto.SessionContextResponse) {
	if s.redis == nil || s.redis.RDB == nil {
		return
	}
	value, err := json.Marshal(response)
	if err != nil {
		return
	}
	_ = s.redis.RDB.Set(ctx, key, value, s.cfg.CacheTTL).Err()
}

func (s *replayService) encrypt(plaintext string) (string, error) {
	key, err := s.encryptionKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), []byte("openreplay:v1"))
	return base64.RawStdEncoding.EncodeToString(sealed), nil
}

func (s *replayService) decrypt(encoded string) (string, error) {
	sealed, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode replay credential: %w", err)
	}
	key, err := s.encryptionKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(sealed) < gcm.NonceSize() {
		return "", errors.New("invalid replay credential")
	}
	plaintext, err := gcm.Open(nil, sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():], []byte("openreplay:v1"))
	if err != nil {
		return "", fmt.Errorf("decrypt replay credential: %w", err)
	}
	return string(plaintext), nil
}

func (s *replayService) encryptionKey() ([]byte, error) {
	if strings.TrimSpace(s.cfg.EncryptionKey) == "" {
		return nil, fmt.Errorf("%w: replay encryption key is not configured", services.ErrInvalidInput)
	}
	sum := sha256.Sum256([]byte(s.cfg.EncryptionKey))
	return sum[:], nil
}

func stringProperty(properties map[string]any, keys ...string) string {
	for _, key := range keys {
		for candidate, value := range properties {
			if strings.EqualFold(candidate, key) {
				switch typed := value.(type) {
				case string:
					return typed
				case json.Number:
					return typed.String()
				case int:
					return strconv.Itoa(typed)
				case int64:
					return strconv.FormatInt(typed, 10)
				case uint64:
					return strconv.FormatUint(typed, 10)
				case float64:
					return strconv.FormatFloat(typed, 'f', -1, 64)
				}
			}
		}
	}
	return ""
}

func intProperty(properties map[string]any, keys ...string) int {
	value := stringProperty(properties, keys...)
	number, _ := strconv.Atoi(value)
	return number
}

func floatProperty(properties map[string]any, keys ...string) float64 {
	value := stringProperty(properties, keys...)
	number, _ := strconv.ParseFloat(value, 64)
	return number
}

func sanitizeURL(value string) string {
	value = redactText(value)
	parsed, err := url.Parse(value)
	if err != nil {
		return truncate(strings.SplitN(value, "?", 2)[0], 256)
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	if parsed.IsAbs() {
		return truncate(parsed.Scheme+"://"+parsed.Host+parsed.EscapedPath(), 256)
	}
	return truncate(parsed.Path, 256)
}

func redactText(value string) string {
	value = emailPattern.ReplaceAllString(value, "[REDACTED_EMAIL]")
	value = jwtPattern.ReplaceAllString(value, "[REDACTED_TOKEN]")
	value = secretPattern.ReplaceAllString(value, "$1=[REDACTED]")
	return truncate(value, 512)
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

func containsAny(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
