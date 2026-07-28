package impl

import (
	"context"
	"time"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	"error-logging/dto"
	"error-logging/services"
)

type analyticsService struct {
	analytics repository.AnalyticsRepository
}

func NewAnalyticsService(analytics repository.AnalyticsRepository) services.AnalyticsService {
	return &analyticsService{analytics: analytics}
}

func (s *analyticsService) GetServiceOverview(ctx context.Context, serviceID uint64, from, to time.Time) ([]dbdto.ServiceOverviewPoint, error) {
	return s.analytics.ServiceOverview(ctx, serviceID, from, to)
}

func (s *analyticsService) GetReleaseHealth(ctx context.Context, serviceID uint64, from, to time.Time) ([]dto.ReleaseHealthPoint, error) {
	rows, err := s.analytics.ReleaseHealth(ctx, serviceID, from, to)
	if err != nil {
		return nil, err
	}

	out := make([]dto.ReleaseHealthPoint, 0, len(rows))
	for _, r := range rows {
		rate := 1.0
		if r.SessionsTotal > 0 {
			rate = 1 - float64(r.SessionsErrored)/float64(r.SessionsTotal)
		}
		out = append(out, dto.ReleaseHealthPoint{
			Release:         r.Release,
			SessionsTotal:   r.SessionsTotal,
			SessionsErrored: r.SessionsErrored,
			CrashFreeRate:   rate,
		})
	}
	return out, nil
}

func (s *analyticsService) GetBreadcrumbs(ctx context.Context, serviceID uint64, sessionID string, before time.Time, limit int) ([]dbdto.Breadcrumb, error) {
	if sessionID == "" {
		return []dbdto.Breadcrumb{}, nil
	}
	rows, err := s.analytics.Breadcrumbs(ctx, serviceID, sessionID, before, limit)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []dbdto.Breadcrumb{}
	}
	return rows, nil
}
