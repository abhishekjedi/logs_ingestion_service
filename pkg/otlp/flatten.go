package otlp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

const (
	attrEnvironment      = "deployment.environment"
	attrRelease          = "service.version"
	attrReleaseAlt       = "release"
	attrUserID           = "user.id"
	attrSessionID        = "session.id"
	attrExceptionType    = "exception.type"
	attrExceptionMessage = "exception.message"
	attrExceptionStack   = "exception.stacktrace"
	attrFingerprintHint  = "log.fingerprint"
)

const severityError = 17

type NormalizedLog struct {
	Timestamp        time.Time
	ObservedAt       time.Time
	SeverityNumber   uint8
	SeverityText     string
	Body             string
	TraceID          string
	SpanID           string
	Environment      string
	Release          string
	UserID           string
	SessionID        string
	ExceptionType    string
	ExceptionMessage string
	RawStacktrace    string
	FingerprintHint  string

	Attributes         map[string]string
	ResourceAttributes map[string]string
}

func (r NormalizedLog) IsError() bool {
	return r.ExceptionType != ""
}

func Flatten(payload []byte) ([]NormalizedLog, error) {
	var req ExportLogsRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, fmt.Errorf("parse otlp logs: %w", err)
	}

	var out []NormalizedLog
	for _, rl := range req.ResourceLogs {
		resAttrs := asMap(rl.Resource.Attributes)
		for _, sl := range rl.ScopeLogs {
			for _, lr := range sl.LogRecords {
				out = append(out, normalize(lr, resAttrs))
			}
		}
	}
	return out, nil
}

func normalize(lr LogRecord, resAttrs map[string]string) NormalizedLog {
	logAttrs := asMap(lr.Attributes)

	pick := func(key string) string {
		if v, ok := logAttrs[key]; ok && v != "" {
			return v
		}
		return resAttrs[key]
	}

	ts := unixNano(lr.TimeUnixNano)
	if ts.IsZero() {
		ts = unixNano(lr.ObservedTimeUnixNano)
	}
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	observed := unixNano(lr.ObservedTimeUnixNano)
	if observed.IsZero() {
		observed = ts
	}

	release := pick(attrRelease)
	if release == "" {
		release = pick(attrReleaseAlt)
	}

	return NormalizedLog{
		Timestamp:          ts,
		ObservedAt:         observed,
		SeverityNumber:     clampSeverity(lr.SeverityNumber),
		SeverityText:       lr.SeverityText,
		Body:               lr.Body.AsString(),
		TraceID:            lr.TraceID,
		SpanID:             lr.SpanID,
		Environment:        pick(attrEnvironment),
		Release:            release,
		UserID:             pick(attrUserID),
		SessionID:          pick(attrSessionID),
		ExceptionType:      pick(attrExceptionType),
		ExceptionMessage:   pick(attrExceptionMessage),
		RawStacktrace:      pick(attrExceptionStack),
		FingerprintHint:    pick(attrFingerprintHint),
		Attributes:         logAttrs,
		ResourceAttributes: resAttrs,
	}
}

func unixNano(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n == 0 {
		return time.Time{}
	}
	return time.Unix(0, n).UTC()
}

func clampSeverity(n int) uint8 {
	if n < 0 {
		return 0
	}
	if n > 24 {
		return 24
	}
	return uint8(n)
}
