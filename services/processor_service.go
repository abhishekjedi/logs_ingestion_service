package services

import (
	"context"

	"error-logging/dto"
)

type ProcessorService interface {
	TransformBatch(ctx context.Context, msgs []dto.LogIngestMessage) (dto.TransformResult, error)
}
