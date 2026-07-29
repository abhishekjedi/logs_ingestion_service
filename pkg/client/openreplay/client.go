package openreplay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"error-logging/pkg/config"
)

const maxResponseBytes = 4 << 20

type Client struct {
	http          *http.Client
	allowInsecure bool
}

func NewClient(cfg config.ReplayConfig) *Client {
	return &Client{
		http:          &http.Client{Timeout: cfg.RequestTimeout},
		allowInsecure: cfg.AllowInsecure,
	}
}

func (c *Client) ValidateProject(
	ctx context.Context,
	apiBaseURL, projectKey, organizationAPIKey string,
) error {
	base, err := c.validateBaseURL(apiBaseURL)
	if err != nil {
		return err
	}
	endpoint := strings.TrimRight(base.String(), "/") + "/public/projects/" + url.PathEscape(projectKey)
	_, err = c.doJSON(ctx, http.MethodGet, endpoint, organizationAPIKey, nil)
	return err
}

func (c *Client) FetchEvents(
	ctx context.Context,
	apiBaseURL, projectKey, sessionID, organizationAPIKey string,
	from, to time.Time,
	maxPages, pageSize int,
) ([]Event, int, error) {
	base, err := c.validateBaseURL(apiBaseURL)
	if err != nil {
		return nil, 0, err
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 200
	}
	if maxPages <= 0 {
		maxPages = 1
	}
	endpoint := strings.TrimRight(base.String(), "/") + "/public/" +
		url.PathEscape(projectKey) + "/sessions/" + url.PathEscape(sessionID) + "/events"

	events := make([]Event, 0, min(pageSize*maxPages, 1000))
	total := 0
	for page := 1; page <= maxPages; page++ {
		payload := map[string]any{
			"startTimestamp": from.UnixMilli(),
			"endTimestamp":   to.UnixMilli(),
			"limit":          pageSize,
			"page":           page,
			"sortOrder":      "asc",
		}
		body, err := c.doJSON(ctx, http.MethodPost, endpoint, organizationAPIKey, payload)
		if err != nil {
			return nil, total, err
		}

		var response struct {
			Data struct {
				Total  int     `json:"total"`
				Events []Event `json:"events"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, total, fmt.Errorf("decode OpenReplay events: %w", err)
		}
		total = response.Data.Total
		events = append(events, response.Data.Events...)
		if len(response.Data.Events) < pageSize || len(events) >= total {
			break
		}
	}
	return events, total, nil
}

func (c *Client) validateBaseURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, fmt.Errorf("invalid OpenReplay API URL: %w", err)
	}
	if parsed.Host == "" || parsed.User != nil {
		return nil, errors.New("invalid OpenReplay API URL")
	}
	if parsed.Scheme != "https" && !(c.allowInsecure && parsed.Scheme == "http") {
		return nil, errors.New("OpenReplay API URL must use HTTPS")
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed, nil
}

func (c *Client) doJSON(
	ctx context.Context,
	method, endpoint, organizationAPIKey string,
	payload any,
) ([]byte, error) {
	var encodedPayload []byte
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		encodedPayload = encoded
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		var body io.Reader
		if encodedPayload != nil {
			body = bytes.NewReader(encodedPayload)
		}
		request, err := http.NewRequestWithContext(ctx, method, endpoint, body)
		if err != nil {
			return nil, err
		}
		request.Header.Set("Authorization", "Bearer "+organizationAPIKey)
		request.Header.Set("Accept", "application/json")
		if payload != nil {
			request.Header.Set("Content-Type", "application/json")
		}

		response, err := c.http.Do(request)
		if err != nil {
			lastErr = err
		} else {
			responseBody, readErr := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
			response.Body.Close()
			if readErr != nil {
				return nil, readErr
			}
			if len(responseBody) > maxResponseBytes {
				return nil, errors.New("OpenReplay response exceeds 4 MiB")
			}
			if response.StatusCode >= 200 && response.StatusCode < 300 {
				return responseBody, nil
			}
			lastErr = fmt.Errorf("OpenReplay API returned %d: %s", response.StatusCode, limitedError(responseBody))
			if response.StatusCode != http.StatusTooManyRequests && response.StatusCode < 500 {
				return nil, lastErr
			}
			if retryAfter := response.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, parseErr := strconv.Atoi(retryAfter); parseErr == nil && seconds > 0 && seconds <= 2 {
					select {
					case <-time.After(time.Duration(seconds) * time.Second):
					case <-ctx.Done():
						return nil, ctx.Err()
					}
					continue
				}
			}
		}

		if attempt < 2 {
			select {
			case <-time.After(time.Duration(attempt+1) * 100 * time.Millisecond):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
	return nil, fmt.Errorf("OpenReplay request failed: %w", lastErr)
}

func limitedError(body []byte) string {
	const limit = 256
	value := strings.TrimSpace(string(body))
	if len(value) > limit {
		return value[:limit]
	}
	return value
}
