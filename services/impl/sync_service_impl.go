package impl

import (
	"context"
	"time"

	"error-logging/db/repository"
	"error-logging/dto"
	"error-logging/pkg/config"
	"error-logging/services"
)

type syncService struct {
	stats  repository.IssueStatsRepository
	issues repository.IssueRepository
	cfg    config.SyncConfig
}

func NewSyncService(
	stats repository.IssueStatsRepository,
	issues repository.IssueRepository,
	cfg config.SyncConfig,
) services.SyncService {
	return &syncService{stats: stats, issues: issues, cfg: cfg}
}

func (s *syncService) SyncOnce(ctx context.Context) (int, error) {
	since := time.Now().Add(-s.cfg.ActiveWindow).UTC()

	aggregates, err := s.stats.ActiveSince(ctx, since)
	if err != nil {
		return 0, err
	}
	if len(aggregates) == 0 {
		return 0, nil
	}

	updates := make([]dto.IssueStatsUpdate, 0, len(aggregates))
	for _, a := range aggregates {
		updates = append(updates, dto.IssueStatsUpdate{
			IssueID:          a.IssueID,
			EventCount:       a.EventCount,
			AffectedUsers:    a.Users,
			AffectedSessions: a.Sessions,
			LastSeen:         a.LastSeen,
		})
	}

	batch := s.cfg.BatchSize
	if batch <= 0 {
		batch = len(updates)
	}
	synced := 0
	for lo := 0; lo < len(updates); lo += batch {
		hi := lo + batch
		if hi > len(updates) {
			hi = len(updates)
		}
		if err := s.issues.UpdateStatsBatch(ctx, updates[lo:hi]); err != nil {
			return synced, err
		}
		synced += hi - lo
	}
	return synced, nil
}
