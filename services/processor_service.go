package services

import (
	"context"

	"error-logging/db/repository"
	"error-logging/dto"
)

// TransformResult is the ClickHouse-bound output of transforming a cycle of ingest
// messages: every record becomes a log row; error records under the rate limit also
// become a full-fidelity error_events row.
type TransformResult struct {
	Logs        []repository.LogRow
	ErrorEvents []repository.ErrorEventRow
}

// ProcessorService turns a batch of ingested OTLP messages into ClickHouse rows. It
// archives raw payloads, parses, fingerprints, resolves issues, and applies the
// rate limit — but performs no ClickHouse writes (the batch consumer flushes the
// returned rows in one insert per cycle).
type ProcessorService interface {
	TransformBatch(ctx context.Context, msgs []dto.LogIngestMessage) (TransformResult, error)
}
