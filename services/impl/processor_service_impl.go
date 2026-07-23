package impl

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"error-logging/constants"
	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	"error-logging/dto"
	redisclient "error-logging/pkg/client/redis"
	s3client "error-logging/pkg/client/s3"
	"error-logging/pkg/fingerprint"
	"error-logging/pkg/otlp"
	"error-logging/services"

	"github.com/google/uuid"
)

// rateLimitPerSecond bounds how many full-fidelity error_events rows we store per
// fingerprint per second. Counting (via logs → MVs) is never gated by this.
const rateLimitPerSecond = 10

// metadataFrames bounds how many frames are cached on the issue metadata sample.
const metadataFrames = 3

type processorService struct {
	s3          *s3client.Client
	issues      repository.IssueRepository
	logs        repository.LogRepository
	errorEvents repository.ErrorEventRepository
	redis       *redisclient.Client
}

func NewProcessorService(
	s3 *s3client.Client,
	issues repository.IssueRepository,
	logs repository.LogRepository,
	errorEvents repository.ErrorEventRepository,
	redis *redisclient.Client,
) services.ProcessorService {
	return &processorService{s3: s3, issues: issues, logs: logs, errorEvents: errorEvents, redis: redis}
}

func (p *processorService) Process(ctx context.Context, msg dto.LogIngestMessage) error {
	// 1. Archive the raw payload (best-effort — archival never blocks processing).
	s3Key := p.archive(ctx, msg)

	// 2. Parse OTLP into normalized records.
	records, err := otlp.Flatten(msg.Payload)
	if err != nil {
		return fmt.Errorf("flatten otlp: %w", err)
	}

	ingestedAt := time.Now().UTC()
	logRows := make([]repository.LogRow, 0, len(records))
	var errorRows []repository.ErrorEventRow

	for _, rec := range records {
		logRow := toLogRow(rec, msg)

		if rec.IsError() {
			frames := fingerprint.ParseStacktrace(rec.RawStacktrace)
			fp := fingerprint.Compute(rec.ExceptionType, rec.ExceptionMessage, frames, rec.FingerprintHint)

			issue, err := p.resolveIssue(ctx, msg.ServiceID, fp, rec, frames)
			if err != nil {
				log.Printf("resolve issue (fp=%s): %v", fp, err)
			} else {
				logRow.IssueID = issue.ID
				// Rate limit gates only the fat error_events row; the log row (and
				// therefore the counts) is always written.
				if p.allowErrorEvent(ctx, fp) {
					errorRows = append(errorRows, toErrorEventRow(rec, issue.ID, msg, frames, ingestedAt, s3Key))
				}
			}
		}

		logRows = append(logRows, logRow)
	}

	if err := p.logs.InsertBatch(ctx, logRows); err != nil {
		return fmt.Errorf("insert logs: %w", err)
	}
	if err := p.errorEvents.InsertBatch(ctx, errorRows); err != nil {
		return fmt.Errorf("insert error_events: %w", err)
	}
	return nil
}

func (p *processorService) archive(ctx context.Context, msg dto.LogIngestMessage) string {
	key := fmt.Sprintf("%d/%s/%s.json", msg.ServiceID, msg.ReceivedAt.Format("2006/01/02"), uuid.NewString())
	if err := p.s3.Put(ctx, key, msg.Payload, "application/json"); err != nil {
		log.Printf("archive to s3 failed (continuing): %v", err)
		return ""
	}
	return key
}

func (p *processorService) resolveIssue(ctx context.Context, serviceID uint64, fp string, rec otlp.NormalizedLog, frames []dbdto.StackFrame) (*dbdto.Issue, error) {
	top := frames
	if len(top) > metadataFrames {
		top = top[:metadataFrames]
	}

	issue := &dbdto.Issue{
		ServiceID:   serviceID,
		Fingerprint: fp,
		Title:       title(rec),
		Culprit:     culprit(frames),
		Level:       severityToLevel(rec.SeverityNumber),
		Status:      constants.StatusUnresolved,
		FirstSeen:   rec.Timestamp,
		LastSeen:    rec.Timestamp,
		Metadata: &dbdto.IssueMetadata{
			ExceptionType:    rec.ExceptionType,
			ExceptionMessage: rec.ExceptionMessage,
			TopFrames:        top,
			SampleSessionID:  rec.SessionID,
		},
	}

	resolved, created, err := p.issues.ResolveOrCreate(ctx, issue)
	if err != nil {
		return nil, err
	}
	if !created {
		// Reopen a resolved issue that has recurred.
		if regressed, err := p.issues.MarkRegressed(ctx, resolved.ID); err != nil {
			log.Printf("regression check (issue=%d): %v", resolved.ID, err)
		} else if regressed {
			log.Printf("issue %d regressed", resolved.ID)
		}
	}
	return resolved, nil
}

// allowErrorEvent applies a per-fingerprint fixed-window rate limit. Redis is
// degradable: if unavailable we allow the write rather than drop fidelity.
func (p *processorService) allowErrorEvent(ctx context.Context, fp string) bool {
	if p.redis == nil || p.redis.RDB == nil {
		return true
	}
	key := fmt.Sprintf("ratelimit:%s:%d", fp, time.Now().Unix())
	n, err := p.redis.RDB.Incr(ctx, key).Result()
	if err != nil {
		return true
	}
	if n == 1 {
		p.redis.RDB.Expire(ctx, key, 2*time.Second)
	}
	return n <= rateLimitPerSecond
}

func toLogRow(rec otlp.NormalizedLog, msg dto.LogIngestMessage) repository.LogRow {
	return repository.LogRow{
		Timestamp:          rec.Timestamp,
		ObservedAt:         rec.ObservedAt,
		ProjectID:          msg.ProjectID,
		ServiceID:          msg.ServiceID,
		IssueID:            0,
		SeverityNumber:     rec.SeverityNumber,
		SeverityText:       rec.SeverityText,
		Body:               rec.Body,
		TraceID:            rec.TraceID,
		SpanID:             rec.SpanID,
		Environment:        rec.Environment,
		Release:            rec.Release,
		UserID:             rec.UserID,
		SessionID:          rec.SessionID,
		ExceptionType:      rec.ExceptionType,
		ExceptionMessage:   rec.ExceptionMessage,
		Attributes:         rec.Attributes,
		ResourceAttributes: rec.ResourceAttributes,
	}
}

func toErrorEventRow(rec otlp.NormalizedLog, issueID uint64, msg dto.LogIngestMessage, frames []dbdto.StackFrame, ingestedAt time.Time, s3Key string) repository.ErrorEventRow {
	chFrames := make([]repository.ErrorEventFrame, len(frames))
	for i, f := range frames {
		inApp := uint8(0)
		if f.InApp {
			inApp = 1
		}
		chFrames[i] = repository.ErrorEventFrame{
			File: f.File, Function: f.Function, Line: f.Line, Col: f.Col, InApp: inApp,
		}
	}

	return repository.ErrorEventRow{
		EventID:            uuid.NewString(),
		IssueID:            issueID,
		ServiceID:          msg.ServiceID,
		ProjectID:          msg.ProjectID,
		Timestamp:          rec.Timestamp,
		IngestedAt:         ingestedAt,
		SeverityNumber:     rec.SeverityNumber,
		SeverityText:       rec.SeverityText,
		Environment:        rec.Environment,
		Release:            rec.Release,
		ExceptionType:      rec.ExceptionType,
		ExceptionMessage:   rec.ExceptionMessage,
		UserID:             rec.UserID,
		SessionID:          rec.SessionID,
		StackFrames:        chFrames,
		RawStacktrace:      rec.RawStacktrace,
		TraceID:            rec.TraceID,
		SpanID:             rec.SpanID,
		Attributes:         rec.Attributes,
		ResourceAttributes: rec.ResourceAttributes,
		S3Key:              s3Key,
	}
}

func title(rec otlp.NormalizedLog) string {
	msg := firstLine(rec.ExceptionMessage)
	if msg == "" {
		return rec.ExceptionType
	}
	return rec.ExceptionType + ": " + msg
}

func culprit(frames []dbdto.StackFrame) string {
	for _, f := range frames {
		if f.InApp {
			if f.Function != "" {
				return f.File + " in " + f.Function
			}
			return f.File
		}
	}
	return ""
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}

func severityToLevel(n uint8) constants.IssueLevel {
	switch {
	case n >= 21:
		return constants.LevelFatal
	case n >= 17:
		return constants.LevelError
	case n >= 13:
		return constants.LevelWarning
	case n >= 9:
		return constants.LevelInfo
	default:
		return constants.LevelDebug
	}
}
