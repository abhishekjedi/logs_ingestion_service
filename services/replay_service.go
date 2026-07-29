package services

import (
	"context"

	"error-logging/dto"
)

type ReplayService interface {
	ListIntegrations(ctx context.Context, projectID uint64) ([]dto.ReplayIntegrationResponse, error)
	UpsertIntegration(ctx context.Context, projectID uint64, request dto.UpsertReplayIntegrationRequest) (*dto.ReplayIntegrationResponse, error)
	DeleteIntegration(ctx context.Context, projectID uint64, externalProjectKey string) error
	GetSessionContext(ctx context.Context, issueID uint64, eventID string) (*dto.SessionContextResponse, error)
}
