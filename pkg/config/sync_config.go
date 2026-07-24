package config

import (
	"time"

	"github.com/knadh/koanf/v2"
)

// SyncConfig tunes the ClickHouse→MySQL issue-stats sync. Env overrides via
// ERRLOG_SYNC_* win over config.yaml.
type SyncConfig struct {
	// Interval between sync passes.
	Interval time.Duration
	// ActiveWindow bounds which issues are re-synced: those with activity newer
	// than now-ActiveWindow. Dormant issues keep their last-synced totals.
	ActiveWindow time.Duration
	// BatchSize chunks the MySQL updates per transaction.
	BatchSize int
}

func NewSyncConfig(k *koanf.Koanf) SyncConfig {
	return SyncConfig{
		Interval:     envDuration("ERRLOG_SYNC_INTERVAL", kDuration(k, "sync.interval", 10*time.Second)),
		ActiveWindow: envDuration("ERRLOG_SYNC_ACTIVE_WINDOW", kDuration(k, "sync.active_window", 2*time.Hour)),
		BatchSize:    envInt("ERRLOG_SYNC_BATCH_SIZE", kInt(k, "sync.batch_size", 500)),
	}
}
