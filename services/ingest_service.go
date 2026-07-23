package services

import (
	"context"

	dbdto "error-logging/db/dto"
)

// IngestService hands a raw OTLP payload off to the durable pipeline (Kafka). It
// does no parsing — that happens in the worker — so the ingest hot path stays lean.
type IngestService interface {
	Ingest(ctx context.Context, service *dbdto.Service, payload []byte) error
}
