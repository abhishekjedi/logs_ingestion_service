package impl

import (
	"context"
	"time"

	"error-logging/db/repository"
	"error-logging/services"
)

type analyticsService struct {
	analytics repository.AnalyticsRepository
}

func NewAnalyticsService(analytics repository.AnalyticsRepository) services.AnalyticsService {
	return &analyticsService{analytics: analytics}
}

func (s *analyticsService) GetServiceOverview(ctx context.Context, serviceID uint64, from, to time.Time) ([]repository.ServiceOverviewPoint, error) {
	return s.analytics.ServiceOverview(ctx, serviceID, from, to)
}

func (s *analyticsService) GetReleaseHealth(ctx context.Context, serviceID uint64, from, to time.Time) ([]services.ReleaseHealthPoint, error) {
	rows, err := s.analytics.ReleaseHealth(ctx, serviceID, from, to)
	if err != nil {
		return nil, err
	}

	out := make([]services.ReleaseHealthPoint, 0, len(rows))
	for _, r := range rows {
		rate := 1.0
		if r.SessionsTotal > 0 {
			rate = 1 - float64(r.SessionsErrored)/float64(r.SessionsTotal)
		}
		out = append(out, services.ReleaseHealthPoint{
			Release:         r.Release,
			SessionsTotal:   r.SessionsTotal,
			SessionsErrored: r.SessionsErrored,
			CrashFreeRate:   rate,
		})
	}
	return out, nil
}
