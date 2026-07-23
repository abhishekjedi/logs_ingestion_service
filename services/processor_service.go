package services

import (
	"context"

	"error-logging/dto"
)

// ProcessorService is the worker pipeline: it turns one ingested OTLP batch into
// archived raw payload, ClickHouse log rows, grouped issues, and rate-limited
// full-fidelity error rows.
type ProcessorService interface {
	Process(ctx context.Context, msg dto.LogIngestMessage) error
}
