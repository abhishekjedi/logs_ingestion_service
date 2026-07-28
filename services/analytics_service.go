package services

import (
	"context"
	"time"

	dbdto "error-logging/db/dto"
	"error-logging/dto"
)

type AnalyticsService interface {
	GetServiceOverview(ctx context.Context, serviceID uint64, from, to time.Time) ([]dbdto.ServiceOverviewPoint, error)
	GetReleaseHealth(ctx context.Context, serviceID uint64, from, to time.Time) ([]dto.ReleaseHealthPoint, error)
	GetBreadcrumbs(ctx context.Context, serviceID uint64, sessionID string, before time.Time, limit int) ([]dbdto.Breadcrumb, error)
}
