package services

import (
	"context"
	"time"

	"error-logging/db/repository"
)

// ReleaseHealthPoint is per-release session health with the crash-free rate derived.
type ReleaseHealthPoint struct {
	Release         string  `json:"release"`
	SessionsTotal   uint64  `json:"sessions_total"`
	SessionsErrored uint64  `json:"sessions_errored"`
	CrashFreeRate   float64 `json:"crash_free_rate"`
}

// AnalyticsService serves service-level read APIs backed by ClickHouse rollups.
type AnalyticsService interface {
	GetServiceOverview(ctx context.Context, serviceID uint64, from, to time.Time) ([]repository.ServiceOverviewPoint, error)
	GetReleaseHealth(ctx context.Context, serviceID uint64, from, to time.Time) ([]ReleaseHealthPoint, error)
}
