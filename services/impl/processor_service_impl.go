package impl

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"error-logging/constants"
	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	"error-logging/dto"
	"error-logging/pkg/config"
	"error-logging/pkg/fingerprint"
	"error-logging/pkg/otlp"
	"error-logging/services"

	"github.com/google/uuid"
)

const metadataFrames = 3

type processorService struct {
	store    services.ObjectStore
	issues   repository.IssueRepository
	cache    services.IssueCache
	limiter  services.RateLimiter
	poolSize int
}

func NewProcessorService(
	store services.ObjectStore,
	issues repository.IssueRepository,
	cache services.IssueCache,
	limiter services.RateLimiter,
	cfg config.WorkerConfig,
) services.ProcessorService {
	size := cfg.PoolSize
	if size < 1 {
		size = 1
	}
	return &processorService{store: store, issues: issues, cache: cache, limiter: limiter, poolSize: size}
}

type msgResult struct {
	logs       []dbdto.LogRow
	candidates []errorCandidate
}

type errorCandidate struct {
	fingerprint string
	row         dbdto.ErrorEventRow
}

// TransformBatch archives one NDJSON object per cycle, transforms messages in a
// bounded worker pool, memoizes fingerprint lookups across the cycle, then applies
// rate limits once per fingerprint instead of once per event.
func (p *processorService) TransformBatch(ctx context.Context, msgs []dto.LogIngestMessage) (dto.TransformResult, error) {
	s3Key := p.archiveBatch(ctx, msgs)

	perMsg := make([]msgResult, len(msgs))
	var memo sync.Map

	var wg sync.WaitGroup
	sem := make(chan struct{}, p.poolSize)
	for i := range msgs {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()
			perMsg[i] = p.transformMessage(ctx, msgs[i], s3Key, &memo)
		}(i)
	}
	wg.Wait()

	var logs []dbdto.LogRow
	var candidates []errorCandidate
	for _, r := range perMsg {
		logs = append(logs, r.logs...)
		candidates = append(candidates, r.candidates...)
	}

	return dto.TransformResult{
		Logs:        logs,
		ErrorEvents: p.applyRateLimit(ctx, candidates),
	}, nil
}

// applyRateLimit does one Redis reservation per fingerprint per cycle, not one per
// event, while still preserving all rows in logs for analytics counts.
func (p *processorService) applyRateLimit(ctx context.Context, candidates []errorCandidate) []dbdto.ErrorEventRow {
	if len(candidates) == 0 {
		return nil
	}

	byFP := make(map[string][]dbdto.ErrorEventRow)
	order := make([]string, 0)
	for _, c := range candidates {
		if _, ok := byFP[c.fingerprint]; !ok {
			order = append(order, c.fingerprint)
		}
		byFP[c.fingerprint] = append(byFP[c.fingerprint], c.row)
	}

	var kept []dbdto.ErrorEventRow
	for _, fp := range order {
		rows := byFP[fp]
		allowed := p.limiter.AllowN(ctx, fp, len(rows))
		if allowed > len(rows) {
			allowed = len(rows)
		}
		// Mint event IDs only for survivors. uuid.NewString reads crypto/rand (one
		// syscall per call), so generating them for every candidate burned most of
		// the worker's CPU on rows the limiter was about to discard.
		survivors := rows[:allowed]
		for i := range survivors {
			survivors[i].EventID = uuid.NewString()
		}
		kept = append(kept, survivors...)
	}
	return kept
}

func (p *processorService) transformMessage(ctx context.Context, msg dto.LogIngestMessage, s3Key string, memo *sync.Map) msgResult {
	records, err := otlp.Flatten(msg.Payload)
	if err != nil {
		log.Printf("flatten otlp (service=%d): %v", msg.ServiceID, err)
		return msgResult{}
	}

	ingestedAt := time.Now().UTC()
	res := msgResult{logs: make([]dbdto.LogRow, 0, len(records))}

	for _, rec := range records {
		logRow := toLogRow(rec, msg)

		if rec.IsError() {
			frames := fingerprint.ParseStacktrace(rec.RawStacktrace)
			fp := fingerprint.Compute(rec.ExceptionType, rec.ExceptionMessage, frames, rec.FingerprintHint)

			issueID := p.resolveIssueMemo(ctx, msg.ServiceID, fp, rec, frames, memo)
			if issueID != 0 {
				logRow.IssueID = issueID
				res.candidates = append(res.candidates, errorCandidate{
					fingerprint: fp,
					row:         toErrorEventRow(rec, issueID, msg, frames, ingestedAt, s3Key),
				})
			}
		}

		res.logs = append(res.logs, logRow)
	}
	return res
}

// archiveBatch uses one S3/MinIO write per worker cycle instead of one write per
// Kafka message.
func (p *processorService) archiveBatch(ctx context.Context, msgs []dto.LogIngestMessage) string {
	if len(msgs) == 0 {
		return ""
	}
	var buf bytes.Buffer
	for _, m := range msgs {
		buf.Write(m.Payload)
		buf.WriteByte('\n')
	}
	key := fmt.Sprintf("raw/%s/%s.ndjson", msgs[0].ReceivedAt.Format("2006/01/02"), uuid.NewString())
	if err := p.store.Put(ctx, key, buf.Bytes(), "application/x-ndjson"); err != nil {
		log.Printf("archive batch failed (continuing): %v", err)
		return ""
	}
	return key
}

func (p *processorService) resolveIssueMemo(ctx context.Context, serviceID uint64, fp string, rec otlp.NormalizedLog, frames []dbdto.StackFrame, memo *sync.Map) uint64 {
	key := strconv.FormatUint(serviceID, 10) + ":" + fp
	if v, ok := memo.Load(key); ok {
		return v.(uint64)
	}

	id, err := p.resolveIssue(ctx, serviceID, fp, rec, frames)
	if err != nil {
		log.Printf("resolve issue (fp=%s): %v", fp, err)
		return 0
	}
	memo.Store(key, id)
	return id
}

func (p *processorService) resolveIssue(ctx context.Context, serviceID uint64, fp string, rec otlp.NormalizedLog, frames []dbdto.StackFrame) (uint64, error) {
	if id, ok := p.cache.Get(ctx, serviceID, fp); ok {
		return id, nil
	}

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
		return 0, err
	}
	if !created {
		if regressed, err := p.issues.MarkRegressed(ctx, resolved.ID); err != nil {
			log.Printf("regression check (issue=%d): %v", resolved.ID, err)
		} else if regressed {
			log.Printf("issue %d regressed", resolved.ID)
		}
	}

	p.cache.Set(ctx, serviceID, fp, resolved.ID)
	return resolved.ID, nil
}

func toLogRow(rec otlp.NormalizedLog, msg dto.LogIngestMessage) dbdto.LogRow {
	return dbdto.LogRow{
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

func toErrorEventRow(rec otlp.NormalizedLog, issueID uint64, msg dto.LogIngestMessage, frames []dbdto.StackFrame, ingestedAt time.Time, s3Key string) dbdto.ErrorEventRow {
	chFrames := make([]dbdto.ErrorEventFrame, len(frames))
	for i, f := range frames {
		inApp := uint8(0)
		if f.InApp {
			inApp = 1
		}
		chFrames[i] = dbdto.ErrorEventFrame{
			File: f.File, Function: f.Function, Line: f.Line, Col: f.Col, InApp: inApp,
		}
	}

	return dbdto.ErrorEventRow{
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
