package services

import (
	"context"

	dbdto "error-logging/db/dto"
)

type IngestService interface {
	Ingest(ctx context.Context, service *dbdto.Service, payload []byte) error
}
